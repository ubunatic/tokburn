package main

import (
	"context"
	"testing"

	"ubunatic.com/tokburn/internal/game"
)

func TestParseVote(t *testing.T) {
	for _, tc := range []struct {
		input, want string
		bad         bool
	}{
		{"low", "low", false}, {"MID", "mid", false}, {"high\n", "high", false},
		{"I choose low or mid", "", true}, {"unknown", "", true},
	} {
		got, err := parseVote(tc.input)
		if (err != nil) != tc.bad || got != tc.want {
			t.Errorf("parseVote(%q) = %q, %v", tc.input, got, err)
		}
	}
}

func TestAverageTier(t *testing.T) {
	for _, tc := range []struct {
		votes []string
		want  string
	}{
		{[]string{"low", "low", "high"}, "mid"},
		{[]string{"low", "mid", "mid"}, "mid"},
		{[]string{"mid", "high", "high"}, "high"},
	} {
		got, err := averageTier(tc.votes)
		if err != nil || got != tc.want {
			t.Errorf("averageTier(%v) = %q, %v", tc.votes, got, err)
		}
	}
	if _, err := averageTier([]string{"low", "mid"}); err == nil {
		t.Fatal("expected vote count error")
	}
}

func TestRunGoldenWithFake(t *testing.T) {
	calls := 0
	tasks := []game.BatchTask{{ID: "x", Text: "task", Tokens: 100, Runs: 2}}
	result, err := runGolden(context.Background(), tasks, map[string]string{"x": "mid"}, func(_ context.Context, model, prompt string) (string, int, int, error) {
		calls++
		if model == "" || prompt == "" {
			t.Fatal("missing model or prompt")
		}
		return []string{"low", "mid", "high"}[calls-1], 12, 2, nil
	}, "2026-10-26")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 || len(result.Tasks) != 1 || result.Tasks[0].Golden != "mid" || result.Tasks[0].Provisional != "mid" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
