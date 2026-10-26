# 006 — Build the title screen and first game screen from the opus design

**Status**: Open
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
