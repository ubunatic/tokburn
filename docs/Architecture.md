# Architecture

How tokburn is put together, and the decisions behind it. Game rules and numbers are not
repeated here: they live in `spec/`.

## Layout

| Path | Role |
|---|---|
| `main.go`, `version.go` | CLI: `tokburn check`, `tokburn play`, `tokburn golden` |
| `spec/` | Single source of truth, embedded with `go:embed` (`spec/spec.go`) |
| `spec/pricing.yaml` | Every tunable value: prices, round sizes, budget factor, penalty, good score, joker, frame delay |
| `spec/game-rules.md` | The game rules in prose, with measured ranges |
| `spec/tasks.yaml`, `templates.yaml` | The 50-task deck (20 low / 17 mid / 13 high) and the 30 templates it was generated from |
| `spec/golden.yaml` | Golden-tier pilot ratings (5 tasks, 3 models) |
| `internal/budget/` | Cost engine (`Calculator`, `Estimate`), spec loading and validation |
| `internal/game/` | Draw, score, play loop, screens and raw terminal |
| `design/` | `.ansi` screen sketches; `design/*/screens.go` embeds the ones the game renders |

## Decisions

- **All numbers in spec yaml, all rules in `spec/game-rules.md`.** Code reads values; it never
  hard-codes a tunable. Spec validation rejects missing or out-of-range values at load time.
- **One rule set for every round size (6–20)** instead of per-size tables (bb5b0bf). Per-size
  tuning was fitted to seeds and unexplainable. Holes are closed by a draw rule instead: redraw
  until each tier has `min_per_tier` cards and the all-high route busts. Open question: whether
  this biases short rounds (issue 011).
- **Strategy tests guard the rules** (`internal/budget/round_test.go`, `internal/game`): right
  tier scores 100 and survives, all-mid survives below good, all-low below good, all-high busts.
  10k seeds per size, 100k for sizes 7 and 20; keeps `make check` around 3 s.
- **Screens are canvases built from embedded `.ansi` sketches.** Static cells come from the
  sketch; live values are written into fixed fields. Tests compare every cell (glyph, FG, BG,
  bold) against the sketch and check bounds for every deck task and every interactive state.
- **Raw terminal via the standard library** on Linux and macOS (`terminal_*.go`), no TUI
  dependency. Restores the terminal on exit, SIGTERM and SIGHUP. Piped stdin skips the title
  screen and raw mode, so tests and scripts get line-based play.
- **Interactive loop:** title (animated flames) → round → result screen (waits for Enter) →
  title. Menu choice is kept; each round draws a fresh seed from the session RNG.

## Pitfalls

- A field written narrower than its text leaves sketch cells showing through (the `]`
  highlight bug, e7f55d0). Test every selection state, not only the sketch's own.
- Formatting `-left` at exactly 0 prints `$-0.00`; clamp with `max(-left, 0)`.
- `foot -e tokburn play` closes the window when the program exits; that is why the result
  screen waits for Enter.
- The long-context surcharge is still a stub in the cost engine (issue 005).

## Release

`harnez release` builds with `make dist`, which must produce versioned archives
(`tokburn-<version>-<os>-<arch>.tar.gz`) and `SHA256SUMS`; bare binaries are not attached.
v0.1.1 first went out empty for this reason and was completed with `harnez release --continue`.
