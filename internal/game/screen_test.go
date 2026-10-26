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
	cards, err := DrawTasks(rand.New(rand.NewSource(42)), c.Spec.GameRoundTasks, c)
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

// Every menu size receives the four strategy guarantees on seeded conditioned draws.
func TestMenuSeededStrategies(t *testing.T) {
	c := calc(t)
	for n := c.Spec.GameRoundTasksMin; n <= c.Spec.GameRoundTasksMax; n++ {
		n := n
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			t.Parallel()
			round := c
			round.Spec.GameRoundTasks = n
			seeds := int64(10000)
			if n == 7 || n == 20 {
				seeds = 100000
			}
			bounds := map[string][2]float64{"right": {1e9, 0}, "low": {1e9, 0}, "mid": {1e9, 0}}
			draws, rejected := 0, 0
			prices := make(map[string][4]float64, len(Deck))
			for _, card := range Deck {
				var costs [4]float64
				for i, pick := range []string{card.Need, "low", "mid", "high"} {
					result, err := Score(round, card, pick)
					if err != nil {
						t.Fatal(err)
					}
					costs[i] = result.USD
				}
				prices[card.ID] = costs
			}
			for seed := int64(0); seed < seeds; seed++ {
				cards, attempts, err := conditionedDraw(rand.New(rand.NewSource(seed)), n, round)
				if err != nil {
					t.Fatal(err)
				}
				draws++
				rejected += attempts - 1
				costs := [4]float64{}
				for _, card := range cards {
					for i, amount := range prices[card.ID] {
						costs[i] += amount
					}
				}
				budget := costs[2] * round.Spec.BudgetAllMFactor
				if costs[0] >= budget || fmt.Sprintf("%.2f", RoutingScore(costs[0], costs[0])) != "100.00" {
					t.Fatalf("seed %d right-tier failed: costs=%v budget=%.5f", seed, costs, budget)
				}
				if costs[2] >= budget {
					t.Fatalf("seed %d all-m did not survive", seed)
				}
				if costs[3] <= budget {
					t.Fatalf("seed %d all-high did not bust", seed)
				}
				for i, name := range []string{"right", "low", "mid"} {
					score := RoutingScore(costs[0], costs[i])
					if name == "right" && math.Round(score*100)/100 != 100 {
						t.Fatalf("seed %d right-tier score %.5f", seed, score)
					}
					if name != "right" && math.Round(score*100)/100 >= round.Spec.GoodScore {
						t.Fatalf("seed %d %s score %.5f >= %.0f", seed, name, score, round.Spec.GoodScore)
					}
					b := bounds[name]
					b[0] = min(b[0], score)
					b[1] = max(b[1], score)
					bounds[name] = b
				}
			}
			rate := float64(rejected) / float64(draws+rejected)
			t.Logf("rejection rate %.4f%% (%d rejected of %d candidates)", 100*rate, rejected, draws+rejected)
			for name, b := range bounds {
				t.Logf("%s score range %.2f–%.2f", name, b[0], b[1])
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
