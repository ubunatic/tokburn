# 004 — Title screen and first game screen: three competing .ansi designs, owner picks

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 001 (v1 game), 002 (v2 arcade view)

---

## 1. Problem & Motivation

`tokburn play` starts with two plain text lines (title, budget and key help) that wrap in a normal
terminal window, then the first task. It does not look like a game. The owner wants a title screen
and a nicer first game screen, chosen from competing designs before any code changes.

/goal Three designers each deliver a title screen and a first game screen as `.ansi` files that
pass loom validation; a sol:low fixer cleans all three; the owner edits them with `loom edit` and
picks one. Stop there: building the chosen design is a follow-up ticket. Stop and report when
blocked on an owner decision. Circuit breaker: on a loop or no progress, clear the goal and write a
failure report here.

## 2. Requirements (owner, 2026-10-26)

- **Title screen** menu, navigated with the arrow keys:
  - `Start (<num> tasks)`: `+` and `-` change `<num>` (the short mode, today 7 tasks).
  - `Start (20 tasks)`: the normal round.
  - `Exit`.
- **Keyboard hints** at the bottom of the title screen.
- **First game screen**: the first two or three lines (title, budget, key help) look good and do not
  wrap; then the first task and the input prompt.
- Show real content: game name, real budget figures, a real task from `spec/tasks.yaml`.

## 3. Process

1. **Designers**, one design each, in `design/<model>/title.ansi` and `design/<model>/game.ansi`:
   agy:argon:med, opus:med, astra:low. Their only job is these first designs; they do not touch code.
2. **Validate** each file with loom: `loom measure`, `loom eval`, `loom check-box` (if boxes are
   used); fix with `loom format` / `loom fmt`. Target width: 80 columns, at most 24 rows.
3. **Fixer**: one codex:sol:low agent fixes issues in all three designs with the loom CLI and
   manual edits, without changing the design ideas.
4. **Owner** edits with `loom edit design/<model>/<file>.ansi`, views with `loom view`, and picks.
5. Follow-up ticket: implement the chosen design (with a `/tui-check` against the `.ansi` sketch).

## 4. Result (owner, 2026-10-26)

All three designs are kept in `design/` for their good points. The owner picked **opus**
(`design/opus/`) to build first: the simplest one, still nice, with good colours. Opus's `.tui`
markup sources are in `.harnez/scratch/design/opus/` (not committed). Implementation: ticket 006.
