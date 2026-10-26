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
