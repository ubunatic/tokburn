# 010 — Game screen redesign: less cluttered, more like a game (sol and opus, three designs each)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 004 (first design round), 006 (current opus game screen), 008 (sol flames)

---

## 1. Problem & Motivation

The current game screen (`design/opus/game.ansi`, built in 006) works but looks cluttered and more
like a business tool than a game: a dense header with budget, joker, combo and per-tier prices, a
task card with token maths, and three grey help lines above the prompt.

/goal Two designers, codex:sol:med and opus:med, each deliver three clearly different game screen
designs as `.ansi` files that pass loom validation; the owner views them and picks one (or mixes).
Stop there: building it is a follow-up ticket. Stop and report when blocked on an owner decision.
Circuit breaker: on a loop or no progress, clear the goal and write a failure report here.

## 2. Requirements (owner, 2026-10-26)

- Feel like a game, not a dashboard: fewer lines, clear focus on the current task and the choice.
- Keep the game's information available: task text, task number, budget left, jokers, combo, the
  l/m/h/j keys (1/2/3/0). Designers decide what to show always, what to shrink, and what to drop
  or move (e.g. help text, per-tier prices, token maths).
- Use real content: a real task from `spec/tasks.yaml` and real budget figures.
- May reuse the TOKBURN look (title colours, sol flames in `design/flames-sol/`), but the three
  designs of each designer must differ in idea, not only in colour.
- 80 columns, at most 24 rows. Validate with `loom measure` and `loom eval` (and `loom check-box`
  if boxes are used).
- Files: `design/game-<model>-<a|b|c>/game.ansi` plus a 2–3 line `README.md` naming the idea.
  Work files go in `.harnez/scratch/`.

## 3. Owner review

View with `loom view design/game-*/game.ansi` and pick.

## 4. Round 1 result (owner, 2026-10-26)

Kept: `game-sol-c` (Mission Poster: overall layout and the big task number) and `game-opus-a`
(Quiz show: the coloured tier answer tiles). The other four are parked in
`.harnez/scratch/game-round1/` (not committed).

## 5. Round 2

Both designers make two merged designs each in `design/game-<model>-r2-<a|b>/`: the sol-c layout
and big task number as the base, with opus-a's coloured tier tiles for the choices. Same rules as
section 2. Owner picks one to build.

## 6. Final design and build (owner, 2026-10-26)

Owner merged: opus-r2-a layout (side panel with task, budget, joker, combo) with sol-r2-b's
rectangular coloured tier cards, aligned with the task text. Result: `design/game-final/game.ansi`
(joker card text split into "reveal" / "tier"). Build it as the `tokburn play` game screen, verified
cell by cell with `/tui-check` and live in tmux; `make check` and `make install` pass. All live
values (task number, big number, task text wrap, budget, jokers, combo, per-tier prices, joker
revealed state, status messages) must stay inside the layout for every deck task and round size.

## 7. Built (2026-10-26)

Built by codex:sol:low, approved by opus:low. Tests cover every deck task × round sizes 6–20 ×
budgets −1234.56…1234567.89 × jokers up to 20 × every combo level, hidden and revealed.
Non-blocking review notes, parked here: status text over 140 characters is cut silently
(`terminal.go`); big digits assume rounds below 100 but spec validation does not cap
`game_round_tasks_max`; `gameScreen`'s `limit` parameter is unused; status rows paint BG 16
where the sketch has none; revealing a joker redraws and drops the previous verdict message
(pre-existing).
