package game

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"ubunatic.com/tokburn/internal/budget"
	"ubunatic.com/tokburn/spec"
)

func calc(t *testing.T) budget.Calculator {
	t.Helper()
	s, err := budget.LoadSpec(spec.Pricing)
	if err != nil {
		t.Fatal(err)
	}
	return budget.Calculator{Spec: s}
}

func TestScoreUnderpoweredBreaks(t *testing.T) {
	c := calc(t)
	card := Card{Task: "x", Need: "high", Tokens: 10_000, Runs: 1}
	weak, err := Score(c, card, "low")
	if err != nil {
		t.Fatal(err)
	}
	right, _ := Score(c, card, "high")
	if !weak.Broken || weak.USD <= right.USD {
		t.Fatalf("underpowered pick must break and cost more: weak=%+v right=%+v", weak, right)
	}
}

func TestPlayRound(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("x\n" + strings.Repeat("h\n", calc(t).Spec.GameRoundTasks))
	_, err := Play(calc(t), rand.New(rand.NewSource(1)), in, &out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "l, m or h?") {
		t.Fatalf("invalid input was not rejected:\n%s", out.String())
	}
}

// Exercise real configured rounds on identical shuffled cards for every strategy.
func TestSeededStrategies(t *testing.T) {
	c := calc(t)
	modes := []struct {
		name, input string
		outcome     string
		good        bool
	}{
		{"right-tier", "right", "survive", true},
		{"all-m", "m", "survive", false},
		{"all-low", "l", "any", false},
		{"all-high", "h", "bust", false},
	}
	bounds := make(map[string][2]float64)
	for _, mode := range modes {
		bounds[mode.name] = [2]float64{1e9, -1e9}
	}
	for seed := int64(0); seed < 100000; seed++ {
		for _, mode := range modes {
			var input strings.Builder
			for _, card := range DrawTasks(rand.New(rand.NewSource(seed)), c.Spec.GameRoundTasks, c.Spec.MinPerTier) {
				pick := mode.input
				if pick == "right" {
					pick = card.Need[:1]
				}
				fmt.Fprintln(&input, pick)
			}
			var out bytes.Buffer
			left, err := Play(c, rand.New(rand.NewSource(seed)), strings.NewReader(input.String()), &out)
			if err != nil {
				t.Fatalf("seed %d %s: %v", seed, mode.name, err)
			}
			if (mode.outcome == "survive" && left <= 0) || (mode.outcome == "bust" && left > 0) {
				t.Fatalf("seed %d %s: left %.2f", seed, mode.name, left)
			}
			score := 0.0
			if _, scanErr := fmt.Sscanf(strings.Split(strings.Split(out.String(), "Score: ")[1], "\n")[0], "%f", &score); scanErr != nil {
				t.Fatalf("seed %d %s: parse score: %v", seed, mode.name, scanErr)
			}
			if (mode.good && score != 100) || (!mode.good && score >= c.Spec.GoodScore) {
				t.Fatalf("seed %d %s: score %.2f, threshold %.2f", seed, mode.name, score, c.Spec.GoodScore)
			}
			if mode.name == "all-high" && !strings.Contains(out.String(), "Quota on fire:") {
				t.Fatalf("seed %d: missing over-budget end screen", seed)
			}
			if strings.Contains(out.String(), "Easter egg: all-m") != (mode.name == "all-m") {
				t.Fatalf("seed %d %s: incorrect easter egg output", seed, mode.name)
			}
			b := bounds[mode.name]
			if score < b[0] {
				b[0] = score
			}
			if score > b[1] {
				b[1] = score
			}
			bounds[mode.name] = b
		}
	}
	for name, b := range bounds {
		t.Logf("%s score range: %.2f..%.2f", name, b[0], b[1])
	}
}
