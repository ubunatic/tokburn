package game

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ubunatic.com/tokburn/design/opus"
)

func assertCanvas(t *testing.T, got, want canvas) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("height %d, want %d", len(got), len(want))
	}
	errors := 0
	for y, row := range want {
		if len(got[y]) != len(row) {
			t.Fatalf("row %d width %d, want %d", y+1, len(got[y]), len(row))
		}
		for x, c := range row {
			if got[y][x] != c {
				if errors < 12 {
					t.Errorf("cell (%d,%d): got %+v, want %+v", x+1, y+1, got[y][x], c)
				}
				errors++
			}
		}
	}
	if errors > 0 {
		t.Fatalf("%d cells differ (glyph, foreground, background, bold)", errors)
	}
}
func TestOpusScreensEveryCell(t *testing.T) {
	t.Run("title", func(t *testing.T) { assertCanvas(t, titleScreen(calc(t).Spec, 0, 7), parseCanvas(opus.Title)) })
	t.Run("game", func(t *testing.T) {
		card := Card{Task: "Reproduce and fix a race between cache invalidation and concurrent writes in a Python background worker with six packages.", Tokens: 30000, Runs: 80, Need: "high"}
		assertCanvas(t, gameScreen(calc(t), card, 1, 20, 176.60, 176.60, 1, 0, false), parseCanvas(opus.Game))
	})
}
func TestDynamicScreenBounds(t *testing.T) {
	c := calc(t)
	for _, card := range Deck {
		for _, i := range []int{1, 20} {
			s := gameScreen(c, card, i, 20, 1234.56, 1234.56, 2, 2, true)
			for y, row := range s {
				if len(row) != 80 {
					t.Fatalf("%s row %d width %d", card.ID, y, len(row))
				}
			}
			if s[5][2].glyph != '╭' || s[5][77].glyph != '╮' || s[11][77].glyph != '╯' {
				t.Fatal("card border moved")
			}
		}
	}
	for i := 0; i < 3; i++ {
		s := titleScreen(c.Spec, i, 7)
		if s[14+i][22].bg != 214 || s[14+i][22].fg != 16 {
			t.Fatal("selection colours")
		}
	}
}

func TestComboMeterUsesConfiguredAnswerCount(t *testing.T) {
	c := calc(t)
	c.Spec.JokerComboAnswers = 4
	got := gameScreen(c, Deck[0], 1, 20, 10, 20, 1, 2, false)
	for i, want := range []rune{'●', '●', '○', '○'} {
		if got[1][72+i].glyph != want {
			t.Fatalf("combo cell %d: got %q, want %q", i, got[1][72+i].glyph, want)
		}
	}
}

func TestTerminalComboRewardIsShown(t *testing.T) {
	c := calc(t)
	c.Spec.JokerComboAnswers = 1
	cards, err := DrawTasks(rand.New(rand.NewSource(42)), c.Spec.GameRoundTasks, c.Spec.MinPerTier)
	if err != nil {
		t.Fatal(err)
	}
	var keys strings.Builder
	for _, card := range cards {
		keys.WriteString(card.Need[:1])
	}
	var out strings.Builder
	in := &terminalInput{keys: bufio.NewReader(strings.NewReader(keys.String())), out: &out}
	if _, err := PlayWithClock(c, rand.New(rand.NewSource(42)), in, &out, &fakeClock{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Combo: earned 1 joker") {
		t.Fatal("interactive status omitted the combo reward")
	}
}
func TestRawKeyAdapter(t *testing.T) {
	// Arrow sequences, including Shift+arrow, are consumed as one event.
	for _, key := range []string{"\x1b[A", "\x1b[1;2A", "\x1bOA"} {
		r := bufio.NewReader(strings.NewReader(key + "m"))
		k, e := readKey(r)
		if k != 'U' || e != nil {
			t.Fatalf("%q: %q %v", key, k, e)
		}
		k, e = readKey(r)
		if k != 'm' || e != nil {
			t.Fatal("sequence leaked into answer")
		}
	}
}

// Every size reachable through +/- receives the same 100k-seed guarantees.
func TestMenuSeededStrategies(t *testing.T) {
	c := calc(t)
	c.Spec = c.Spec.ShortRound()
	seeds := int64(100000)
	if testing.Short() {
		seeds = 100
	}
	for n := c.Spec.GameRoundTasksMin; n <= c.Spec.GameRoundTasksMax; n++ {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			bounds := map[string][2]float64{"low": {1e9, 0}, "mid": {1e9, 0}, "high": {1e9, 0}}
			for seed := int64(0); seed < seeds; seed++ {
				cards, err := DrawTasks(rand.New(rand.NewSource(seed)), n, c.Spec.MinPerTier)
				if err != nil {
					t.Fatal(err)
				}
				costs := map[string]float64{"right": 0, "low": 0, "mid": 0, "high": 0}
				for _, card := range cards {
					for _, tier := range []string{"right", "low", "mid", "high"} {
						pick := tier
						if tier == "right" {
							pick = card.Need
						}
						r, e := Score(c, card, pick)
						if e != nil {
							t.Fatal(e)
						}
						costs[tier] += r.USD
					}
				}
				limit := costs["mid"] * c.Spec.BudgetAllMFactor
				if costs["right"] >= limit || fmt.Sprintf("%.2f", RoutingScore(costs["right"], costs["right"])) != "100.00" || costs["mid"] >= limit || costs["high"] <= limit {
					t.Fatalf("size %d seed %d: survival rules: %v budget %.5f", n, seed, costs, limit)
				}
				for _, tier := range []string{"low", "mid", "high"} {
					score := RoutingScore(costs["right"], costs[tier])
					b := bounds[tier]
					b[0] = min(b[0], score)
					b[1] = max(b[1], score)
					bounds[tier] = b
				}
				for _, tier := range []string{"low", "mid"} {
					if math.Round(RoutingScore(costs["right"], costs[tier])*100)/100 >= c.Spec.GoodScore {
						t.Fatalf("size %d seed %d: %s score %.5f", n, seed, tier, RoutingScore(costs["right"], costs[tier]))
					}
				}
			}
			for tier, b := range bounds {
				t.Logf("%s score range %.2f–%.2f", tier, b[0], b[1])
			}
		})
	}
}

// Optional fixtures allow the live tmux comparison to use the exact same
// calculator and seeded draw as the application, without fixed demo prices.
func TestLiveScreenFixtures(t *testing.T) {
	dir := os.Getenv("TOKBURN_TUI_FIXTURES")
	if dir == "" {
		t.Skip("live fixture export not requested")
	}
	c := calc(t)
	for _, n := range []int{6, 7, 20} {
		round := c
		if n != 20 {
			round.Spec = round.Spec.ShortRound()
			round.Spec.GameRoundTasks = n
		}
		cards := drawTasks(t, rand.New(rand.NewSource(42)), n, round.Spec.MinPerTier)
		limit := 0.0
		for _, card := range cards {
			r, _ := Score(round, card, "mid")
			limit += r.USD
		}
		limit *= round.Spec.BudgetAllMFactor
		s := gameScreen(round, cards[0], 1, n, limit, limit, 1, 0, false)
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("game-%d.ansi", n)), []byte(s.ansi()), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTerminalPipedInput(t *testing.T) {
	in, feed, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	output, out, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	fmt.Fprint(feed, strings.Repeat("m\n", 7))
	feed.Close()
	_, err = PlayTerminal(calc(t), rand.New(rand.NewSource(42)), in, out, true)
	out.Close()
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "route 7 tasks") || !strings.Contains(text, "Month survived") || strings.Contains(text, "Start (") || strings.Contains(text, "\x1b[H") {
		t.Fatalf("piped short game changed: %q", text)
	}
}
