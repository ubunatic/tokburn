package game

import (
	"bufio"
	"io"
	"math/rand"
	"strings"
	"testing"
)

// Compare complete rounds so aliases must preserve routing, joker consumption,
// repeated reveals, exhausted jokers, and the final score in both input modes.
func TestNumberKeyAliases(t *testing.T) {
	for _, raw := range []bool{false, true} {
		name := "line"
		if raw {
			name = "raw"
		}
		t.Run(name, func(t *testing.T) {
			run := func(numeric bool) (float64, string) {
				t.Helper()
				keys := "jjhjlh" + strings.Repeat("lmh", 7)
				if numeric {
					keys = strings.NewReplacer("l", "1", "m", "2", "h", "3", "j", "0").Replace(keys)
				}
				var out strings.Builder
				var in io.Reader
				if raw {
					in = &terminalInput{keys: bufio.NewReader(strings.NewReader(keys)), out: &out}
				} else {
					var lines strings.Builder
					for _, key := range keys {
						lines.WriteString(" \t" + string(key) + " \n")
					}
					in = strings.NewReader(lines.String())
				}
				left, err := PlayWithClock(calc(t), rand.New(rand.NewSource(42)), in, &out, &fakeClock{})
				if err != nil {
					t.Fatal(err)
				}
				return left, out.String()
			}
			wantLeft, want := run(false)
			gotLeft, got := run(true)
			if gotLeft != wantLeft || got != want {
				t.Fatalf("numeric aliases differ from letter keys: budget %v, want %v", gotLeft, wantLeft)
			}
			if !strings.Contains(got, "Score:") || !strings.Contains(got, "No jokers left.") {
				t.Fatal("round or joker paths were not exercised")
			}
			if !raw {
				for _, help := range []string{"l/1 (low)", "m/2 (mid)", "h/3 (high)", "j/0 reveals"} {
					if !strings.Contains(got, help) {
						t.Errorf("missing help %q", help)
					}
				}
			}
		})
	}
}
