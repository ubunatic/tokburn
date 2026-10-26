# 009 — Return to the title screen after a round instead of exiting

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 006 (title screen), 008 (flame animation)

---

## 1. Problem & Motivation

After the last task of a round, `tokburn play` prints the result (survived or over budget, score,
good-score line) and exits. The owner wants to keep playing: the result stays on screen until the
player presses Enter, then the game goes back to the title screen (flames, menu, last chosen round
size kept). `Exit` in the menu and `q`/Ctrl-C remain the ways to quit.

/goal In an interactive terminal, the round result waits for Enter and then returns to the title
screen, repeatedly, without leaking terminal state or animation goroutines; piped stdin keeps
today's behaviour (one round, then exit). Verified by tests and a live tmux/PTY run, `make check`
and `make install` pass. Stop and report when blocked on an owner decision. Circuit breaker: on a
loop or no progress, clear the goal and write a failure report here.

## 2. Notes

- Show a key hint on the result screen, e.g. `Enter: back to title · q: quit`.
- Each round draws new random tasks (new seed per round, not a replay).
