package game

import (
	"bufio"
	"math/rand"
	"strings"
	"testing"

	"ubunatic.com/tokburn/design/opus"
)

func TestResultScreenEveryCell(t *testing.T) {
	assertCanvas(t, resultScreen(12.34, 100, 80, true), parseCanvas(opus.Result))
	for _, left := range []float64{-12.34, 0, 12.34} {
		got := resultScreen(left, 42, 75, false)
		if len(got) != 24 {
			t.Fatal("height")
		}
		for _, row := range got {
			if len(row) != 80 {
				t.Fatal("width")
			}
		}
		want := "Month survived with $12.34 left. Score: 42.00"
		color := 71
		if left <= 0 {
			want = "Quota on fire: $12.34 over budget. Score: 42.00"
			if left == 0 {
				want = "Quota on fire: $0.00 over budget. Score: 42.00"
			}
			color = 167
		}
		for x, r := range []rune(want) {
			if got[9][4+x] != (cell{glyph: r, fg: color, bg: -1, bold: true}) {
				t.Fatalf("outcome cell %d", x)
			}
		}
		for _, c := range got[13] {
			if c.glyph != ' ' || c.fg != -1 || c.bg != -1 {
				t.Fatal("stale easter egg")
			}
		}
	}
}

func TestRoundCompletionAndQuit(t *testing.T) {
	for _, keys := range []string{"q", "\x03", "\x04", ""} {
		var out strings.Builder
		in := &terminalInput{keys: bufio.NewReader(strings.NewReader(keys)), out: &out}
		if _, err := PlayWithClock(calc(t), rand.New(rand.NewSource(42)), in, &out, &fakeClock{}); err != nil {
			t.Fatal(err)
		}
		if !in.quit || strings.Contains(out.String(), "ROUND COMPLETE") {
			t.Fatal("aborted round treated as completed")
		}
	}
	var out strings.Builder
	in := &terminalInput{keys: bufio.NewReader(strings.NewReader(strings.Repeat("m", 20))), out: &out}
	if _, err := PlayWithClock(calc(t), rand.New(rand.NewSource(42)), in, &out, &fakeClock{}); err != nil {
		t.Fatal(err)
	}
	if in.quit || !strings.Contains(out.String(), "ROUND COMPLETE") || !strings.Contains(out.String(), "Enter: back to title") {
		t.Fatal("completed round missing result")
	}
}

func TestRepeatedTerminalRounds(t *testing.T) {
	// Six tasks twice, then normal twenty twice (the last round is aborted).
	// Enter alone returns from results; irrelevant keys leave results in place.
	keys := "-\r" + strings.Repeat("m", 6) + "x\x1b[1;2B\r\r" +
		strings.Repeat("h", 6) + "\r\x1b[B\r" + strings.Repeat("m", 20) + "\r\rq"
	frames := make(chan string, 256)
	in := &terminalInput{keys: bufio.NewReader(strings.NewReader(keys)), out: frameWriter{frames}}
	if _, err := in.playRounds(calc(t), rand.New(rand.NewSource(42))); err != nil {
		t.Fatal(err)
	}
	if in.titleStop != nil {
		t.Fatal("title renderer still active")
	}
	close(frames)
	results := 0
	var starts []string
	for frame := range frames {
		if strings.Contains(frame, "ROUND COMPLETE") {
			results++
		}
		if strings.Contains(frame, "TASK 1 / ") {
			starts = append(starts, frame)
		}
	}
	if results != 3 || len(starts) != 4 {
		t.Fatalf("results=%d round starts=%d", results, len(starts))
	}
	if !strings.Contains(starts[0], "TASK 1 / 6") || !strings.Contains(starts[1], "TASK 1 / 6") ||
		!strings.Contains(starts[2], "TASK 1 / 20") || !strings.Contains(starts[3], "TASK 1 / 20") {
		t.Fatal("menu count or normal selection was lost")
	}
	if starts[0] == starts[1] {
		t.Fatal("round replayed the previous draw")
	}
}
