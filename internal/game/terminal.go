package game

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"ubunatic.com/tokburn/internal/budget"
)

type terminalInput struct {
	keys    *bufio.Reader
	out     io.Writer
	pending []byte
}

func (t *terminalInput) draw(c canvas) { fmt.Fprint(t.out, "\x1b[H\x1b[2J", c.ansi()) }
func (t *terminalInput) message(text string) {
	runes := []rune(text)
	if len(runes) > 78 {
		runes = runes[:78]
	}
	t.statusLine(string(runes), false)
}
func (t *terminalInput) clearBelow() { t.statusLine("", true) }
func (t *terminalInput) statusLine(text string, clearBelow bool) {
	fmt.Fprint(t.out, "\x1b[19;1H")
	if clearBelow {
		fmt.Fprint(t.out, "\x1b[J")
	} else {
		fmt.Fprint(t.out, "\x1b[2K", text)
	}
}

func readKey(r *bufio.Reader) (byte, error) {
	k, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	if k == 27 {
		a, e := r.ReadByte()
		if e != nil {
			return 0, e
		}
		if a == '[' || a == 'O' {
			for {
				b, e := r.ReadByte()
				if e != nil {
					return 0, e
				}
				if b >= 0x40 && b <= 0x7e {
					if b == 'A' {
						return 'U', nil
					}
					if b == 'B' {
						return 'D', nil
					}
					return 0, nil
				}
			}
		}
		return 0, nil
	}
	return k, nil
}
func (t *terminalInput) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if len(t.pending) == 0 {
		k, err := readKey(t.keys)
		if err != nil {
			return 0, err
		}
		if k == 3 || k == 4 || k == 'q' {
			return 0, io.EOF
		}
		if k == '\r' || k == '\n' {
			k = 0
		}
		t.pending = []byte{[]byte(strings.ToLower(string(k)))[0], '\n'}
	}
	n := copy(p, t.pending)
	t.pending = t.pending[n:]
	return n, nil
}

// PlayTerminal preserves the line-oriented protocol when stdin is a pipe.
func PlayTerminal(c budget.Calculator, rng *rand.Rand, in, out *os.File, short bool) (float64, error) {
	restore, terminal, err := rawTerminal(in)
	if err != nil {
		return 0, err
	}
	if !terminal {
		if short {
			c.Spec = c.Spec.ShortRound()
		}
		return Play(c, rng, in, out)
	}
	var restoreOnce sync.Once
	restoreInput := func() { restoreOnce.Do(restore) }
	defer restoreInput()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGHUP)
	done := make(chan struct{})
	defer close(done)
	defer signal.Stop(signals)
	go func() {
		select {
		case sig := <-signals:
			restoreInput()
			fmt.Fprint(out, "\x1b[0m\x1b[?25h\r\n")
			os.Exit(128 + int(sig.(syscall.Signal)))
		case <-done:
		}
	}()
	fmt.Fprint(out, "\x1b[?25l")
	defer fmt.Fprint(out, "\x1b[0m\x1b[?25h\r\n")
	t := &terminalInput{keys: bufio.NewReader(in), out: out}
	selected, count := 0, c.Spec.GameRoundTasksShort
	for {
		t.draw(titleScreen(c.Spec, selected, count))
		key, err := readKey(t.keys)
		if err == io.EOF {
			return 0, nil
		}
		if err != nil {
			return 0, err
		}
		switch key {
		case 'q', 3, 4:
			return 0, nil
		case 'U':
			selected = (selected + 2) % 3
		case 'D':
			selected = (selected + 1) % 3
		case '-':
			if selected == 0 && count > c.Spec.GameRoundTasksMin {
				count--
			}
		case '+', '=':
			if selected == 0 && count < c.Spec.GameRoundTasksMax {
				count++
			}
		case '\r', '\n':
			if selected == 2 {
				return 0, nil
			}
			if selected == 0 {
				c.Spec = c.Spec.ShortRound()
				c.Spec.GameRoundTasks = count
			}
			return PlayWithClock(c, rng, t, out, wallClock{})
		}
	}
}
