package game

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
	"time"
)

type fakeClock struct {
	calls     int
	durations []time.Duration
}

func (c *fakeClock) Now() time.Time {
	i := c.calls / 2
	end := c.calls%2 == 1
	c.calls++
	now := time.Unix(int64(i*10), 0)
	if end && i < len(c.durations) {
		now = now.Add(c.durations[i])
	}
	return now
}

func TestJokerCombos(t *testing.T) {
	for _, tc := range []struct {
		name     string
		slowAt   int
		duration time.Duration
		wrongAt  int
		revealAt int
		want     int
	}{
		{"under threshold", -1, 0, -1, -1, 2},
		{"at threshold", 1, 3 * time.Second, -1, -1, 1},
		{"above threshold", 1, 3*time.Second + time.Nanosecond, -1, -1, 1},
		{"just under threshold", 1, 3*time.Second - time.Nanosecond, -1, -1, 2},
		{"wrong resets", -1, 0, 1, -1, 1},
		{"reveal resets", -1, 0, -1, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := calc(t)
			cards := drawTasks(t, rand.New(rand.NewSource(42)), c.Spec.GameRoundTasks, c.Spec.MinPerTier)
			var input strings.Builder
			clock := &fakeClock{durations: make([]time.Duration, len(cards))}
			for i, card := range cards {
				if i == tc.revealAt {
					input.WriteString("j\n")
				}
				pick := card.Need[:1]
				if i == tc.wrongAt || i >= 6 {
					if pick == "l" {
						pick = "h"
					} else {
						pick = "l"
					}
				}
				input.WriteString(pick + "\n")
				if i == tc.slowAt {
					clock.durations[i] = tc.duration
				}
			}
			var out bytes.Buffer
			if _, err := PlayWithClock(c, rand.New(rand.NewSource(42)), strings.NewReader(input.String()), &out, clock); err != nil {
				t.Fatal(err)
			}
			if got := strings.Count(out.String(), "Combo: earned"); got != tc.want {
				t.Fatalf("earned %d, want %d:\n%s", got, tc.want, out.String())
			}
		})
	}
}

func TestJokerConsumptionRepeatAndRoundIsolation(t *testing.T) {
	c := calc(t)
	for round := 0; round < 2; round++ {
		var out bytes.Buffer
		// Repeated reveals on the first task consume only one joker; the next task has none.
		_, err := PlayWithClock(c, rand.New(rand.NewSource(42)), strings.NewReader("j\nj\nh\nj\nh\n"), &out, &fakeClock{})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(out.String(), "Provisional tier:") != 2 || strings.Count(out.String(), "No jokers left.") != 1 {
			t.Fatalf("unexpected reveal output:\n%s", out.String())
		}
		if strings.Count(out.String(), ": -$") != 2 {
			t.Fatalf("reveals advanced or answered tasks:\n%s", out.String())
		}
	}
}

func TestDrawValidation(t *testing.T) {
	for _, tc := range []struct {
		name       string
		cards      []Card
		n, minimum int
	}{
		{"missing tier", []Card{{Need: "low"}, {Need: "mid"}, {Need: "mid"}}, 3, 1},
		{"too few in tier", []Card{{Need: "low"}, {Need: "mid"}, {Need: "high"}, {Need: "high"}, {Need: "high"}, {Need: "high"}}, 6, 2},
		{"round too small", Deck, 5, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateDraw(tc.cards, tc.n, tc.minimum); err == nil {
				t.Fatal("expected invalid draw")
			}
		})
	}
}

func TestEarnedJokerCanBeSpent(t *testing.T) {
	c := calc(t)
	cards := drawTasks(t, rand.New(rand.NewSource(42)), c.Spec.GameRoundTasks, c.Spec.MinPerTier)
	var input strings.Builder
	for i := 0; i < 3; i++ {
		input.WriteString(cards[i].Need[:1] + "\n")
	}
	input.WriteString("j\nj\n" + cards[3].Need[:1] + "\nj\n" + cards[4].Need[:1] + "\nj\n")
	var out bytes.Buffer
	if _, err := PlayWithClock(c, rand.New(rand.NewSource(42)), strings.NewReader(input.String()), &out, &fakeClock{}); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "Combo: earned") != 1 || strings.Count(out.String(), "Provisional tier:") != 3 || strings.Count(out.String(), "No jokers left.") != 1 {
		t.Fatalf("earned joker was not available exactly once:\n%s", out.String())
	}
}
