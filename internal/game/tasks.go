package game

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
	"ubunatic.com/tokburn/internal/budget"
	"ubunatic.com/tokburn/spec"
)

// LoadTasks parses a task list and validates the fields used by the game.
func LoadTasks(data []byte) ([]Card, error) {
	var cards []Card
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&cards); err != nil {
		return nil, fmt.Errorf("parse tasks: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("invalid tasks: expected one YAML document")
	}
	if len(cards) == 0 {
		return nil, fmt.Errorf("invalid tasks: empty deck")
	}
	seen := make(map[string]bool, len(cards))
	for i, card := range cards {
		if strings.TrimSpace(card.ID) == "" || seen[card.ID] {
			return nil, fmt.Errorf("invalid task %d: missing or duplicate id %q", i+1, card.ID)
		}
		seen[card.ID] = true
		if strings.TrimSpace(card.Task) == "" || card.Tokens <= 0 || card.Runs <= 0 {
			return nil, fmt.Errorf("invalid task %q: need text, positive tokens and runs", card.ID)
		}
		if _, ok := rank[card.Need]; !ok {
			return nil, fmt.Errorf("invalid task %q: unknown tier %q", card.ID, card.Need)
		}
	}
	return cards, nil
}

func mustLoadTasks() []Card {
	cards, err := LoadTasks(spec.Tasks)
	if err != nil {
		panic(err)
	}
	pricing, err := budget.LoadSpec(spec.Pricing)
	if err != nil {
		panic(err)
	}
	if err := validateDraw(cards, pricing.GameRoundTasks, pricing.MinPerTier); err != nil {
		panic(err)
	}
	if err := validateDraw(cards, pricing.GameRoundTasksShort, pricing.MinPerTier); err != nil {
		panic(err)
	}
	return cards
}

// validateDraw rejects impossible rejection-sampling constraints before drawing.
func validateDraw(cards []Card, n, minPerTier int) error {
	if n > len(cards) {
		n = len(cards)
	}
	if n < 0 || minPerTier < 0 || minPerTier > n/3 {
		return fmt.Errorf("invalid draw: round size %d cannot meet tier minimum %d", n, minPerTier)
	}
	counts := map[string]int{}
	for _, card := range cards {
		counts[card.Need]++
	}
	for _, tier := range []string{"low", "mid", "high"} {
		if counts[tier] < minPerTier {
			return fmt.Errorf("invalid deck: %s has %d cards, need %d", tier, counts[tier], minPerTier)
		}
	}
	return nil
}

// BatchTask is the input a decider may see, with no answer tier.
type BatchTask struct {
	ID     string `json:"id" yaml:"id"`
	Text   string `json:"text" yaml:"text"`
	Tokens int    `json:"tokens" yaml:"tokens"`
	Runs   int    `json:"runs" yaml:"runs"`
}

// ProjectBatch copies only the public task fields for a decider.
func ProjectBatch(cards []Card) []BatchTask {
	batch := make([]BatchTask, len(cards))
	for i, card := range cards {
		batch[i] = BatchTask{ID: card.ID, Text: card.Task, Tokens: card.Tokens, Runs: card.Runs}
	}
	return batch
}
