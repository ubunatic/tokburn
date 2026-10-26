# 006 — Build the title screen and first game screen from the opus design

**Status**: Closed — opus screens shipped in 5ff8604
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 004 (design competition, owner picked opus), 005 (surcharge)

---

## 1. Goal

/goal `tokburn play` opens the title screen from `design/opus/title.ansi` (arrow-key menu:
`Start (<num> tasks)` with `-`/`+`, `Start (20 tasks)`, `Exit`; keyboard hints at the bottom), and
the game screen matches `design/opus/game.ansi` (header with task count, budget gauge, jokers and
combo; l/m/h/j keys in tier colours; task card; `ROUTE ›` prompt). Verified cell by cell with
`/tui-check` against the `.ansi` files and live in tmux; `make check` and `make install` pass. Stop
and report when blocked on an owner decision. Circuit breaker: on a loop or no progress, clear the
goal and write a failure report here.

## 2. Notes

- The design shows a per-tier cost line on the task card; keep it.
- `<num>` replaces the fixed short mode: default from `game_round_tasks_short` (7); limits in spec.
  The 100k-seed rules in `spec/game-rules.md` must hold for every reachable round size, or the
  limits are narrowed to sizes where they hold.
- Non-interactive input (piped stdin, tests) keeps working: skip the title screen and the raw-key
  mode when stdin is not a terminal.
- Design budgets assume the surcharge from 005; use whatever the engine returns.

## 3. Implementation and verification (2026-10-26)

- Embedded opus sketches supply glyphs and colours; dynamic fields use the real
  calculator. Linux/macOS raw input uses the standard library, with no new dependency.
  Other platforms retain the line-input game. Piped stdin skips the title and raw mode.
- Menu limits are 6–7 in pricing.yaml; the normal option remains 20 tasks. Both
  adjustable sizes pass all strategy assertions over 100,000 seeds. A trial range
  7–20 found all-high surviving at size 8 (seed 16565), so the range was narrowed.
- Canvas tests compare every glyph, FG, BG and bold attribute against both opus
  sketches. All embedded cards retain the 80-column geometry. Loom eval, measure,
  check-box and live loom diff pass.
- Private 80×24 tmux: 1,920 title cells and 1,360 seeded game cells match.
  PTY checks also compare 6- and 20-task screens, deliver Shift+Down, exercise
  menu bounds, reveal, routing, invalid input, q/Ctrl-C, and verify termios restoration.
- Verification scripts, captures, fixtures and logs are under `.harnez/scratch/`.
  Font-dependent appearance still needs the owner's final visual inspection.
