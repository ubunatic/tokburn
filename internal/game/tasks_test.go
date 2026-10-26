package game

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"ubunatic.com/tokburn/spec"
)

func TestEmbeddedTasks(t *testing.T) {
	cards, err := LoadTasks(spec.Tasks)
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 50 {
		t.Fatalf("got %d tasks, want 50", len(cards))
	}
	seen := make(map[string]bool)
	for _, card := range cards {
		if seen[card.ID] {
			t.Fatalf("duplicate id %q", card.ID)
		}
		seen[card.ID] = true
	}
	if !reflect.DeepEqual(cards, Deck) {
		t.Fatal("game deck differs from embedded tasks")
	}
}

func TestLoadTasksValidation(t *testing.T) {
	valid := "- id: task-1\n  text: Fix a typo\n  tokens: 1000\n  runs: 2\n  tier: low\n"
	if _, err := LoadTasks([]byte(valid)); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"malformed":      "[",
		"empty":          "[]",
		"missing id":     strings.Replace(valid, "task-1", "", 1),
		"duplicate id":   valid + valid,
		"blank text":     strings.Replace(valid, "Fix a typo", "'  '", 1),
		"zero tokens":    strings.Replace(valid, "1000", "0", 1),
		"negative runs":  strings.Replace(valid, "runs: 2", "runs: -1", 1),
		"invalid tier":   strings.Replace(valid, "tier: low", "tier: gold", 1),
		"missing tier":   strings.Replace(valid, "  tier: low\n", "", 1),
		"unknown field":  valid + "  secret: high\n",
		"extra document": valid + "---\n[]\n",
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadTasks([]byte(data)); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestBatchProjection(t *testing.T) {
	batch := ProjectBatch(Deck)
	if len(batch) != len(Deck) {
		t.Fatal("batch length differs")
	}
	for i, task := range batch {
		card := Deck[i]
		if task != (BatchTask{ID: card.ID, Text: card.Task, Tokens: card.Tokens, Runs: card.Runs}) {
			t.Fatalf("incorrect projection at %d", i)
		}
	}
	typ := reflect.TypeOf(BatchTask{})
	if typ.NumField() != 4 {
		t.Fatal("batch must contain only id, text, tokens and runs")
	}
	for _, marshal := range []func(any) ([]byte, error){json.Marshal, yaml.Marshal} {
		data, err := marshal(batch)
		if err != nil {
			t.Fatal(err)
		}
		var rows []map[string]any
		if err := yaml.Unmarshal(data, &rows); err != nil {
			t.Fatal(err)
		}
		for _, row := range rows {
			if len(row) != 4 || row["tier"] != nil || row["Need"] != nil {
				t.Fatalf("batch leaked answer fields: %v", row)
			}
		}
	}
}

func TestPlayEmbeddedDeck(t *testing.T) {
	var out bytes.Buffer
	_, err := Play(calc(t), rand.New(rand.NewSource(42)), len(Deck)+1,
		strings.NewReader(strings.Repeat("h\n", len(Deck))), &out)
	if err != nil {
		t.Fatal(err)
	}
	for _, card := range Deck {
		if !strings.Contains(out.String(), card.Task) {
			t.Fatalf("game did not play embedded task %q", card.ID)
		}
	}
	if strings.Count(out.String(), "\n> ") != len(Deck) {
		t.Fatal("game did not cap round at deck size")
	}
}
