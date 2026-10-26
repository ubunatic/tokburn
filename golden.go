package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"ubunatic.com/tokburn/internal/game"
)

var tierWord = regexp.MustCompile(`(?i)\b(low|mid|high)\b`)

var goldenModels = []string{"agy:argon:low", "opus:low", "codex:sol:low"}

type goldenVote struct {
	Model        string   `yaml:"model" json:"model"`
	Tier         string   `yaml:"tier" json:"tier"`
	InputTokens  int      `yaml:"input_tokens,omitempty" json:"input_tokens,omitempty"`
	OutputTokens int      `yaml:"output_tokens,omitempty" json:"output_tokens,omitempty"`
	CostUSD      *float64 `yaml:"cost_usd,omitempty" json:"cost_usd,omitempty"`
}

type goldenTask struct {
	ID          string       `yaml:"id" json:"id"`
	Votes       []goldenVote `yaml:"votes" json:"votes"`
	Golden      string       `yaml:"golden" json:"golden"`
	Provisional string       `yaml:"provisional" json:"provisional"`
}

type goldenFile struct {
	Date  string       `yaml:"date" json:"date"`
	Tasks []goldenTask `yaml:"tasks" json:"tasks"`
}

func parseVote(output string) (string, error) {
	var obj struct {
		Tier string `json:"tier"`
		Text string `json:"text"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(output)), &obj) == nil && obj.Tier != "" {
		output = obj.Tier
	}
	matches := tierWord.FindAllString(strings.TrimSpace(output), -1)
	if len(matches) != 1 {
		return "", fmt.Errorf("expected one tier (low, mid, high), found %d", len(matches))
	}
	return strings.ToLower(matches[0]), nil
}

func averageTier(votes []string) (string, error) {
	if len(votes) != 3 {
		return "", fmt.Errorf("expected three votes, got %d", len(votes))
	}
	sum := 0
	for _, v := range votes {
		switch v {
		case "low":
		case "mid":
			sum++
		case "high":
			sum += 2
		default:
			return "", fmt.Errorf("invalid tier %q", v)
		}
	}
	// Round the mean to the nearest integer, with halves rounded upward.
	return []string{"low", "mid", "high"}[(sum+1)/3], nil
}

type goldenRunner func(context.Context, string, string) (string, int, int, error)

func runGolden(ctx context.Context, tasks []game.BatchTask, provisional map[string]string, runner goldenRunner, date string) (goldenFile, error) {
	result := goldenFile{Date: date}
	for _, task := range tasks {
		row := goldenTask{ID: task.ID, Provisional: provisional[task.ID]}
		tiers := make([]string, 0, len(goldenModels))
		for _, model := range goldenModels {
			prompt, _ := json.Marshal(task)
			out, inTokens, outTokens, err := runner(ctx, model, "Choose the lowest tier that can complete this task. Input is one task object with id, text, tokens, and runs only. Reply with exactly one word: low, mid, or high.\n"+string(prompt))
			if err != nil {
				return result, fmt.Errorf("task %s model %s: %w", task.ID, model, err)
			}
			tier, err := parseVote(out)
			if err != nil {
				return result, fmt.Errorf("task %s model %s: %w", task.ID, model, err)
			}
			row.Votes = append(row.Votes, goldenVote{Model: model, Tier: tier, InputTokens: inTokens, OutputTokens: outTokens})
			tiers = append(tiers, tier)
		}
		row.Golden, _ = averageTier(tiers)
		result.Tasks = append(result.Tasks, row)
	}
	return result, nil
}

func realGoldenRunner(ctx context.Context, dir, model, prompt string) (string, int, int, error) {
	cmd := exec.CommandContext(ctx, "harnez", "agent", "--json", "--model", model, "-d", dir, "-p", prompt)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return string(out), 0, 0, err
	}
	var data struct {
		Response string `json:"response"`
		Input    int    `json:"input_tokens_total"`
		Output   int    `json:"output_tokens_total"`
	}
	if json.Unmarshal(out, &data) == nil && data.Response != "" {
		return data.Response, data.Input, data.Output, nil
	}
	return string(out), 0, 0, nil
}

func goldenCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("golden requires --ids id,id,...")
	}
	ids := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "--ids" && i+1 < len(args) {
			ids = args[i+1]
			i++
		} else {
			return fmt.Errorf("unknown golden argument %q", args[i])
		}
	}
	if ids == "" {
		return errors.New("golden requires --ids id,id,...")
	}
	cards, err := game.LoadTasks(mustReadTasks())
	if err != nil {
		return err
	}
	wanted := map[string]bool{}
	for _, id := range strings.Split(ids, ",") {
		wanted[strings.TrimSpace(id)] = true
	}
	var tasks []game.BatchTask
	provisional := map[string]string{}
	for _, card := range cards {
		if wanted[card.ID] {
			tasks = append(tasks, game.ProjectBatch([]game.Card{card})[0])
			provisional[card.ID] = card.Need
			delete(wanted, card.ID)
		}
	}
	if len(wanted) > 0 {
		return fmt.Errorf("unknown task ids: %v", keys(wanted))
	}
	if len(tasks) > 5 {
		return errors.New("pilot is limited to five tasks")
	}
	if len(tasks) == 0 {
		return errors.New("no tasks selected")
	}
	var result goldenFile
	for _, task := range tasks {
		// Each call gets a new empty directory; remove it once the call has completed.
		var votes []string
		row := goldenTask{ID: task.ID, Provisional: provisional[task.ID]}
		for _, model := range goldenModels {
			dir, err := os.MkdirTemp("", "tokburn-golden-")
			if err != nil {
				return err
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 0 {
				os.RemoveAll(dir)
				return errors.New("rater directory is not empty")
			}
			body, _ := json.Marshal(task)
			prompt := "Choose the lowest tier that can complete this task. Input is one task object with id, text, tokens, and runs only. Reply with exactly one word: low, mid, or high.\n" + string(body)
			out, inTokens, outTokens, runErr := realGoldenRunner(context.Background(), dir, model, prompt)
			os.RemoveAll(dir)
			if runErr != nil {
				return fmt.Errorf("task %s model %s: %w", task.ID, model, runErr)
			}
			tier, err := parseVote(out)
			if err != nil {
				return fmt.Errorf("task %s model %s: %w (raw %q)", task.ID, model, err, strings.TrimSpace(out))
			}
			votes = append(votes, tier)
			row.Votes = append(row.Votes, goldenVote{Model: model, Tier: tier, InputTokens: inTokens, OutputTokens: outTokens})
		}
		row.Golden, err = averageTier(votes)
		if err != nil {
			return err
		}
		result.Tasks = append(result.Tasks, row)
	}
	result.Date = time.Now().Format("2006-01-02")
	b, err := yaml.Marshal(result)
	if err != nil {
		return err
	}
	if err = os.WriteFile("spec/golden.yaml", b, 0644); err != nil {
		return err
	}
	return nil
}

func mustReadTasks() []byte {
	b, err := os.ReadFile("spec/tasks.yaml")
	if err != nil {
		panic(err)
	}
	return b
}
func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
