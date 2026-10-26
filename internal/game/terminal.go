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
	"time"

	"ubunatic.com/tokburn/internal/budget"
)

type terminalInput struct {
	keys      *bufio.Reader
	out       io.Writer
	pending   []byte
	quit      bool
	titleMu   sync.Mutex
	titleStop func()
}

func (t *terminalInput) draw(c canvas)       { fmt.Fprint(t.out, "\x1b[H\x1b[2J", c.ansi()) }
func (t *terminalInput) message(text string) { t.statusLine(text, false) }
func (t *terminalInput) clearBelow()         { t.statusLine("", true) }
func (t *terminalInput) statusLine(text string, clearBelow bool) {
	if clearBelow {
		fmt.Fprint(t.out, "\x1b[22;5H\x1b[J")
		return
	}
	// Two reserved rows retain the complete verdict and combo reward without
	// touching the side panel, tiles, or terminal's bottom-right cell.
	runes := []rune(text)
	for row := 0; row < 2; row++ {
		start := min(row*70, len(runes))
		end := min(start+70, len(runes))
		fmt.Fprintf(t.out, "\x1b[%d;5H\x1b[38;5;245m\x1b[48;5;16m%-70s", 22+row, string(runes[start:end]))
	}
	fmt.Fprint(t.out, "\x1b[0m")
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
			if err == io.EOF {
				t.quit = true
			}
			return 0, err
		}
		if k == 3 || k == 4 || k == 'q' {
			t.quit = true
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

// animateTitle owns all title output. Stopping joins the renderer before game
// output or terminal restoration, while key reads remain on the calling thread.
func (t *terminalInput) animateTitle(s budget.Spec, selected, count int) (chan<- [2]int, func()) {
	updates := make(chan [2]int)
	stop, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(time.Duration(s.TitleFrameSeconds * float64(time.Second)))
		defer ticker.Stop()
		frame := 0
		t.draw(titleScreenFrame(s, selected, count, frame))
		for {
			select {
			case <-stop:
				return
			case state := <-updates:
				selected, count = state[0], state[1]
			case <-ticker.C:
				frame = (frame + 1) % len(titleFrames)
			}
			t.draw(titleScreenFrame(s, selected, count, frame))
		}
	}()
	var once sync.Once
	return updates, func() { once.Do(func() { close(stop); <-done }) }
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
	t := &terminalInput{keys: bufio.NewReader(in), out: out}

	fmt.Fprint(out, "\x1b[?25l")

	go func() {
		select {
		case sig := <-signals:
			t.stopTitle()
			restoreInput()
			fmt.Fprint(out, "\x1b[0m\x1b[?25h\r\n")
			os.Exit(128 + int(sig.(syscall.Signal)))
		case <-done:
		}
	}()
	defer fmt.Fprint(out, "\x1b[0m\x1b[?25h\r\n")
	defer t.stopTitle()
	return t.playRounds(c, rng)
}

func (t *terminalInput) stopTitle() {
	t.titleMu.Lock()
	defer t.titleMu.Unlock()
	if t.titleStop != nil {
		t.titleStop()
		t.titleStop = nil
	}
}

func (t *terminalInput) playRounds(c budget.Calculator, rng *rand.Rand) (float64, error) {
	selected, count := 0, c.Spec.GameRoundTasksShort
	left := 0.0
	for {
		t.titleMu.Lock()
		updates, stop := t.animateTitle(c.Spec, selected, count)
		t.titleStop = stop
		t.titleMu.Unlock()
		round, err := t.chooseRound(c, updates, &selected, &count)
		t.stopTitle()
		if err != nil || t.quit {
			return left, err
		}
		// Advance the caller's RNG to a fresh seed for every round, including
		// deterministic sessions started with -seed.
		left, err = PlayWithClock(round, rand.New(rand.NewSource(rng.Int63())), t, t.out, wallClock{})
		if err != nil || t.quit {
			return left, err
		}
		for {
			key, err := readKey(t.keys)
			if err == io.EOF {
				return left, nil
			}
			if err != nil {
				return left, err
			}
			if key == 'q' || key == 3 || key == 4 {
				return left, nil
			}
			if key == '\r' || key == '\n' {
				break
			}
		}
	}
}

func (t *terminalInput) chooseRound(c budget.Calculator, updates chan<- [2]int, selected, count *int) (budget.Calculator, error) {
	for {
		key, err := readKey(t.keys)
		if err != nil {
			if err == io.EOF {
				t.quit = true
				err = nil
			}
			return c, err
		}
		switch key {
		case 'q', 3, 4:
			t.quit = true
			return c, nil
		case 'U':
			*selected = (*selected + 2) % 3
		case 'D':
			*selected = (*selected + 1) % 3
		case '-':
			if *selected == 0 {
				*count = max(*count-1, c.Spec.GameRoundTasksMin)
			}
		case '+', '=':
			if *selected == 0 {
				*count = min(*count+1, c.Spec.GameRoundTasksMax)
			}
		case '\r', '\n':
			if *selected == 2 {
				t.quit = true
				return c, nil
			}
			if *selected == 0 {
				c.Spec = c.Spec.ShortRound()
				c.Spec.GameRoundTasks = *count
			}
			return c, nil
		}
		updates <- [2]int{*selected, *count}
	}
}
