---
kind: sprint
title: Quota Defense v1, screens and design rounds
---

# Quota Defense v1, screens and design rounds (2026-10-26)

## Context

One day on `tokburn`, starting from the scaffold (19524b8, cost engine with a stubbed surcharge).
Goal set by the owner: a playable v1 they can try themselves, then polish of its screens.
Working mode: the orchestrator (Claude Opus) routes work to cheaper agents via `harnez agent` —
codex:sol:low/med for development and design, opus:low/med for review and design, a resumable
sol:med advisor (`tokburn-advisor`) for decisions instead of asking the owner. Owner rules: no
push or publish, big golden/AI runs need approval, scratch only in `.harnez/scratch/`, edit with
Edit/Write rather than shell scripts.

## Outcome

34 commits, no pushes. Delivered:

- **Game (epic 001, T1–T4):** 50-task deck from 30 templates (edffb61), per-round budget and
  all-m easter egg with a 100k-seed strategy test (18fd042), joker and combos (cc4c928), golden
  tier pilot (f602809). Owner played a round and scored 83.07; decision 6 keeps the 50 tasks.
- **Round sizes 6–20 (bb5b0bf):** one rule set (penalty 0.27, budget 1.02 × all-m, good 86)
  plus a draw rule that rejects rounds where all-high survives. `make check` went from 56 s to 3 s.
- **Screens:** opus title and game screen from a three-designer competition (004 → 006), number
  keys (09ed41d), animated sol flames from a two-designer competition (007 → 008), result screen
  that waits for Enter and returns to the title (009), menu highlight fix (e7f55d0).
- **Game screen redesign (010):** round 1 (six designs), owner kept sol-c and opus-a; round 2
  (four merges); a final merge in `design/game-final/` (uncommitted, build pending owner go).
- **Discovered workflow:** the `.ansi` → `.tui` → `.ansi` round trip with `loom markup` and
  `loom diff`, filed upstream as loom 386.

## What Went Wrong

- **T2 tuning overfit.** The first strategy test played the whole deck instead of real rounds, and
  always-low won; the rework overfit seeds 0–999. Fixed with 20-card rounds, a per-round budget,
  a redo-based penalty and 100k seeds.
- **Per-size tuning was a fit, not a design.** Thresholds of 95–97 and penalties jumping between
  0.01 and 0.27 per size. Caught by the opus:low review; replaced by one rule set plus a draw rule.
- **Reviewer's prediction was wrong.** The review promised ≤0.2% rejected draws; the developer
  measured 0.54%–87.31% and documented the real numbers instead of the promised one. Open
  question filed as 011.
- **Score "disappeared".** The owner thought the game crashed: `foot -e tokburn play` closes the
  window when the program exits. Not a bug in the game, but the cause of a real UX need → 009.
- **Stale highlight on `]`.** The menu row is 36 cells but was written with width 35, so the last
  cell showed the static sketch underneath. Tests compared only the default selection. Fixed and
  covered by a regression test that fails without the fix.
- **"$-0.00 over budget".** Found by the 009 review; the test had locked the bug in.
- **Agent process:** a harnez role mis-inference (developer prompt read as reviewer); the
  leaf-worker guard stopped a developer from starting rater agents (orchestrator ran them); an
  agent wrote to `/tmp` outside the repo (moved to scratch); a design launch loop used zsh-only
  `${=p}` in bash and started nothing; a designer generated `.ansi` with a Python heredoc against
  the owner's rule; my own direct `.ansi` edit lost three trailing spaces.

## Learnings

- Strategy tests must play the real game shape (round size, budget per round) and many seeds;
  small seed sets invite fitted constants.
- Prefer one explainable rule plus a filter over per-case tables of tuned numbers.
- Treat reviewer numbers as hypotheses; developers should measure and report the real value.
- Cell-by-cell tests against a sketch need every interactive state (each menu selection, edge
  values like exactly $0), not only the sketch's own state.
- Cheap designers in parallel, owner picks, then merge rounds works well: owner choices took
  seconds and each round narrowed the design.
- Agents draft `.ansi` well but edit it badly; the `.tui` round trip plus `loom diff` makes edits
  exact and provable. Help screens with concrete examples were how the agent discovered it.
- Launch background agents with plain POSIX/bash syntax; check the job started before reporting.

## Follow-Up Tickets

- tokburn **011** — measure whether the all-high-busts redraw biases short rounds (new).
- tokburn **010** — build `design/game-final/` (open, waiting for owner go).
- tokburn **005** — long-context surcharge stub (open, existing).
- loom **386** — document the `.tui` round trip for agents; add `loom markup -o` / `--in-place`.
- harnez **961** (scratch dir default) and **962** (isolated agent mode, blocks epic 001 T6).
- Non-goals: no ticket for the foot window closing (solved by 009) or the zsh launch slip
  (orchestrator habit, recorded with `harnez rate`).
