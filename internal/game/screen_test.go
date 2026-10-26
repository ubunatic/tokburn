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
	"time"

	final "ubunatic.com/tokburn/design/game-final"
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
	t.Run("animated title", func(t *testing.T) {
		frame := titleFrames[0]
		want := parseCanvas(strings.Repeat(strings.Repeat(" ", 80)+"\n", 24))
		if len(opusTitle) != 24 || len(frame) != 12 {
			t.Fatalf("title/frame heights: %d/%d", len(opusTitle), len(frame))
		}
		for y := 0; y < 12; y++ {
			if len(opusTitle[y]) != 80 || len(frame[y]) > 48 {
				t.Fatalf("row %d dimensions: title=%d frame=%d", y+1, len(opusTitle[y]), len(frame[y]))
			}
			copy(want[y][16:64], frame[y])
		}
		want[12], want[13] = opusTitle[9], opusTitle[10]
		for y := 12; y <= 18; y++ {
			want[y+2] = opusTitle[y]
		}
		want[21], want[23] = opusTitle[20], opusTitle[23]
		got := titleScreen(calc(t).Spec, 0, 7)
		assertCanvas(t, got, want)
	})

	t.Run("game", func(t *testing.T) {
		card := Card{Task: "Reproduce and fix a race between cache invalidation and concurrent writes in a Python background worker with six packages.", Tokens: 30000, Runs: 80, Need: "high"}
		assertCanvas(t, gameScreen(calc(t), card, 1, 20, 176.60, 176.60, 1, 0, false), parseCanvas(final.Game))
	})
}
func TestDynamicScreenBounds(t *testing.T) {
	c := calc(t)
	for i := 0; i < 3; i++ {
		s := titleScreen(c.Spec, i, 7)
		if s[16+i][22].bg != 214 || s[16+i][22].fg != 16 {
			t.Fatal("selection colours")
		}
		// Only the selected row keeps the highlight, up to the closing "]".
		for y := 16; y <= 18; y++ {
			for x := 22; x < 58; x++ {
				if (s[y][x].bg == 214) != (y == 16+i) {
					t.Fatalf("selected %d: cell %d,%d bg %d", i, y, x, s[y][x].bg)
				}
			}
		}
	}
}

func TestComboMeterUsesConfiguredAnswerCount(t *testing.T) {
	c := calc(t)
	c.Spec.JokerComboAnswers = 4
	got := gameScreen(c, Deck[0], 1, 20, 10, 20, 1, 2, false)
	for i, want := range []rune{'●', '●', '○', '○'} {
		if got[19][10+2*i].glyph != want {
			t.Fatalf("combo cell %d: got %q, want %q", i, got[19][10+2*i].glyph, want)
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
		cards, err := DrawTasks(rand.New(rand.NewSource(rand.New(rand.NewSource(42)).Int63())), n, round)
		if err != nil {
			t.Fatal(err)
		}
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
		revealed := gameScreen(round, cards[0], 1, n, limit, limit, 0, 0, true)
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("revealed-%d.ansi", n)), []byte(revealed.ansi()), 0600); err != nil {
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

func TestSolFramesEveryCell(t *testing.T) {
	paths, err := filepath.Glob("../../design/flames-sol/frame-*.ansi")
	if err != nil || len(paths) != len(titleFrames) {
		t.Fatalf("frame files: %v %v; loaded %d", paths, err, len(titleFrames))
	}
	for i, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		want := parseCanvas(string(data))
		got := titleScreenFrame(calc(t).Spec, 0, 7, i)
		if len(got) != 24 || len(want) != 12 {
			t.Fatal("height")
		}
		region := make(canvas, 12)
		for y := range got {
			if len(got[y]) != 80 {
				t.Fatal("width")
			}
		}
		for y := range region {
			region[y] = got[y][16:64]
			if len(want[y]) != 48 {
				t.Fatal("frame width")
			}
		}
		assertCanvas(t, region, want)
		assertCanvas(t, titleScreenFrame(calc(t).Spec, 0, 7, i+8), got)
	}
}

// The renderer must keep ticking without input and join before game output.
func TestTitleAnimationStops(t *testing.T) {
	c := calc(t)
	c.Spec.TitleFrameSeconds = .001
	frames := make(chan string, 128)
	in := &terminalInput{out: frameWriter{frames}}
	updates, stop := in.animateTitle(c.Spec, 0, 7)
	defer stop()
	next := func() string {
		t.Helper()
		select {
		case frame := <-frames:
			return frame
		case <-time.After(time.Second):
			t.Fatal("title renderer stalled")
			return ""
		}
	}
	first := next()
	if second := next(); first == second {
		t.Fatal("animation did not advance without input")
	}
	updates <- [2]int{1, 6}
	for !strings.Contains(next(), "Start (6 tasks)") {
	}
	stop()
	stop()
	for len(frames) > 0 {
		<-frames
	}
	select {
	case <-frames:
		t.Fatal("title output after stop")
	case <-time.After(10 * time.Millisecond):
	}
}

type frameWriter struct{ frames chan string }

func (w frameWriter) Write(p []byte) (int, error) {
	w.frames <- string(p)
	return len(p), nil
}

// Check every immutable cell as well as full field contents: clipping text must
// fail even when the canvas dimensions themselves remain unchanged.
func TestFinalGameAllDeckBounds(t *testing.T) {
	c := calc(t)
	sketch := parseCanvas(final.Game)
	text := func(s canvas, y, x, w int) string {
		var b strings.Builder
		for _, v := range s[y][x : x+w] {
			b.WriteRune(v.glyph)
		}
		return strings.TrimSpace(b.String())
	}
	mutable := func(y, x int) bool {
		return (y >= 5 && y <= 9 && x >= 4 && x < 16) || (y >= 5 && y <= 9 && x >= 25 && x < 75) ||
			((y == 11 || y == 14 || y == 17 || y == 19) && x >= 4 && x < 22) ||
			(y == 15 && ((x >= 25 && x < 36) || (x >= 38 && x < 49) || (x >= 51 && x < 62) || (x >= 64 && x < 75))) ||
			(y == 16 && x >= 64 && x < 75)
	}
	for _, card := range Deck {
		for n := c.Spec.GameRoundTasksMin; n <= c.Spec.GameRoundTasksMax; n++ {
			for _, revealed := range []bool{false, true} {
				for _, amount := range []float64{-1234.56, 0, 0.01, 176.60, 1234567.89} {
					for combo := 0; combo <= c.Spec.JokerComboAnswers; combo++ {
						s := gameScreen(c, card, n, n, amount, 176.60, n, combo, revealed)
						if len(s) != 24 {
							t.Fatal("height")
						}
						for y, row := range s {
							if len(row) != 80 {
								t.Fatalf("%s row %d width %d", card.ID, y, len(row))
							}
							for x, v := range row {
								if !mutable(y, x) && v != sketch[y][x] {
									t.Fatalf("%s immutable cell %d,%d changed", card.ID, x, y)
								}
							}
						}
						var task []string
						for y := 5; y < 9; y++ {
							if line := text(s, y, 25, 50); line != "" {
								task = append(task, line)
							}
						}
						if strings.Join(task, " ") != card.Task {
							t.Fatalf("%s task clipped: %q", card.ID, task)
						}
						if text(s, 14, 4, 18) != fmt.Sprintf("$%.2f", amount) {
							t.Fatal("budget clipped")
						}
						if text(s, 11, 4, 18) != fmt.Sprintf("TASK / %d", n) {
							t.Fatal("total clipped")
						}
						if text(s, 17, 4, 18) != fmt.Sprintf("◆ %d JOKER", n) {
							t.Fatal("jokers clipped")
						}
						if strings.Count(text(s, 19, 4, 18), "●") != combo {
							t.Fatal("combo clipped")
						}
						if revealed && text(s, 16, 64, 11) != card.Need {
							t.Fatal("reveal clipped")
						}
						for i, tier := range []string{"low", "mid", "high"} {
							e, _ := c.Estimate(card.Tokens, tier)
							if text(s, 15, 25+13*i, 11) != fmt.Sprintf("$%.2f", e.USD*float64(card.Runs)) {
								t.Fatal("price clipped")
							}
						}
						digits := fmt.Sprintf("%02d", n)
						for i, d := range digits {
							for y, line := range taskDigits[d-'0'] {
								if stringRunes(s[5+y][4+7*i:9+7*i]) != line {
									t.Fatal("big number clipped")
								}
							}
						}
					}
				}
			}
		}
	}
}

func stringRunes(row []cell) string {
	var b strings.Builder
	for _, v := range row {
		b.WriteRune(v.glyph)
	}
	return b.String()
}

func TestGameStatusBounds(t *testing.T) {
	var out strings.Builder
	terminal := &terminalInput{out: &out}
	for _, message := range []string{"No jokers left.", "l/1, m/2 or h/3? (j/0 to reveal; q to quit)", "Combo: earned 1 joker (21 left) · BROKE, redone on high: -$1234.56 · $1234567.89 left"} {
		out.Reset()
		terminal.message(message)
		raw := out.String()
		if !strings.HasPrefix(raw, "\x1b[22;5H") {
			t.Fatal("status position")
		}
		row := parseCanvas(strings.ReplaceAll(strings.TrimPrefix(raw, "\x1b[22;5H"), "\x1b[23;5H", "\n"))
		if len(row) != 2 || len(row[0]) != 70 || len(row[1]) != 70 {
			t.Fatal("status escapes its region")
		}
		if strings.TrimRight(stringRunes(row[0])+stringRunes(row[1]), " ") != message {
			t.Fatal("status clipped")
		}
		for _, v := range append(row[0], row[1]...) {
			if v.fg != 245 || v.bg != 16 {
				t.Fatal("status colour")
			}
		}
	}
}
