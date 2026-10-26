---
kind: story
title: tokburn from scaffold to v0.1.1
---

# tokburn from scaffold to v0.1.1 (2026-10-26)

Review: second-vendor review by codex gpt-6.1-sol (med); five findings fixed (parked 010 review
findings, open epic 001 tasks, test-timing claim, commit counts with baseline, file list and tag).

## 1. Header & Context

- **Previous story:** none. This is the first `kind: story` study; the baseline is the first
  commit, 19524b8 (scaffold: `tokburn check` with a stubbed long-context surcharge, a plain
  six-question `tokburn play`).
- **Scope:** one day, 2026-10-26, from scaffold to release v0.1.1. Summarises the sprint study
  [2026-10-26 Quota Defense v1, screens and design rounds](2026-10-26-quota-defense-v1-screens-and-design-rounds.md).
- **Goals (owner):** a playable v1 the owner can try (epic 001), then make it look and feel like a
  game; keep every rule and number in `spec/`; no push or release without the owner.

## 2. Executive Summary

Quota Defense went from a six-line prompt to a released terminal game: a 50-task deck, a
per-round budget with a seed-tested rule set for round sizes 6–20, jokers and combos, an animated
flame title, a redesigned game screen and a result screen that loops back to the title. Three
design competitions (title, flames, game screen) ran on cheap parallel designers with owner picks
in seconds. v0.1.1 is released on Codeberg with signed archives. Evergreen docs and `harnez init`
now give the repo an agent entry point (`AGENTS.md`, `docs/README.md`).

## 3. What Worked Well

- **Spec as single source of truth.** Every tunable lives in `spec/pricing.yaml`, rules in
  `spec/game-rules.md`; retuning never touched code constants.
- **Seed-based strategy tests.** Four checks (right tier = 100, all-mid survives below good,
  all-low below good, all-high busts) over 10k–100k seeds caught overfitting twice.
- **Design competitions.** 2–3 designers in parallel, owner picks or names a merge, then a build
  ticket. Six designs → two kept → four merges → one final, all within a short session.
- **Developer → reviewer pairs** (codex:sol:low builds, opus:low reviews) found a real issue in
  every review: per-size overfit, a deleted full-screen test, `$-0.00`, a missing embed file.
- **Cell-by-cell canvas tests** against `.ansi` sketches, plus live tmux/PTY runs.
- **The `.tui` round trip** (`loom markup`, edit, back, `loom diff`) made a hand edit of an `.ansi`
  design exact and provable; discovered from `--help` examples (loom 386).
- **Advisor instead of owner questions** (`tokburn-advisor`), once the owner asked for it.

## 4. Honest Post-Mortem (Failures, Bugs & Near-Misses)

| What happened | Caught by | Fix |
|---|---|---|
| T2 strategy test played the whole deck; always-low won | Seed test result | Real 20-card rounds + break penalty |
| T2 rework overfit seeds 0–999 | Wider seed range | Per-round budget (1.02 × all-mid), redo-based penalty, 100k seeds |
| Per-size tuning table fitted to seeds (good 95–97, gaps at 12 and 14) | opus:low review | One rule set + all-high-busts draw rule (bb5b0bf) |
| Reviewer predicted ≤0.2% redraws; real rate 0.54%–87.31% | Developer measured and refused to report the promised number | Documented; bias question filed as 011 |
| Owner thought the game crashed: score vanished | Owner report | Cause: `foot -e` closes on exit. Result screen waits for Enter (009) |
| Stale highlight on the menu's `]` | Owner, live | Width 35 → 36, regression test that fails without the fix (e7f55d0) |
| `$-0.00 over budget`, locked in by a test | 009 review | `max(-left, 0)` |
| Full title-screen cell test deleted during 008 | 008 review | Restored full 80×24 check |
| Missing `design/game-final/screens.go` in commit plan | 010 review | Committed with the change |
| v0.1.1 released without assets | Release log warning | `make dist` builds archives + SHA256SUMS; `harnez release --continue` (12a9932) |
| Agent wrote outside the repo (`/tmp`) | Owner | Moved to `.harnez/scratch/`; harnez 961 |
| harnez inferred reviewer role for a developer prompt | Session start | New session; `harnez rate` |
| Leaf-worker guard blocked a developer from starting raters | Agent stop | Orchestrator ran `tokburn golden` |
| Orchestrator launch loop used zsh `${=p}` in bash; nothing started | Job exit code | Relaunched with bash syntax; rule in AGENTS.md |
| Orchestrator edits via Bash (`printf >>`, heredoc), a designer's Python heredoc | User hook | Edit/Write; `.tui` round trip; loom 386 |
| Orchestrator `.ansi` edit lost three spaces | Rendered check | `.tui` round trip + `loom diff` |

No data loss, no reverts, no broken pushes.

## 5. Quality & Invariants Audit

| Area | Verdict | Evidence |
|---|---|---|
| Architecture & module separation | Good | `internal/budget` (cost, spec), `internal/game` (draw, play, screens, terminal), `spec/` embedded, `design/*/screens.go` embeds sketches |
| Idempotency | Good | `make dist` rebuilds from scratch; `harnez release --continue` reused the tag and attached assets without re-tagging; `harnez index --stage` stable |
| Backward compatibility | Kept | Piped stdin still plays one line-based round and exits (tested) |
| Tests & verification | Improved | ~1,030 test lines added; `make check` green before each code commit in the session (no per-commit log kept); ~3 s at the end, 56 s before the seed cut; regression tests for both live-found bugs |
| Open risks | Known | Surcharge stub (005), short-round draw bias (011), long status text cut at 140 chars, big digits assume < 100 tasks |

Quality verdict: **improved**. Every review found something. Most findings were fixed before
commit; ticket 010 parks five accepted limitations (among them status-text truncation, a
background mismatch, and the verdict lost after a joker reveal). The two bugs that reached the owner were state-coverage gaps, now tested.

## 6. Efficiency & Velocity Assessment

- 43 commits in one day including the baseline (42 in `19524b8..HEAD`): 21 touching
  code/spec/design/build (20 after the baseline), 31 touching `issues/` or `docs/` (some touch both). 11 tickets filed, 6 closed (004, 006–010), 5 open (001, 002, 003, 005, 011).
  One release (v0.1.1).
- Outside `issues/`: +6,501 / −57 lines. Of that: ~2,290 bundled harnez practice docs
  (generated copies), ~1,030 tests, ~1,020 spec (deck, templates, rules), ~550 design sketches,
  ~1,140 production Go.
- Typical cycles: designer rounds 1.5–3 minutes per model in parallel; sol:low builds 2.5–5
  minutes; opus:low reviews about 1 minute.
- Waste: two T2 rework loops, one 56-second test suite before the seed cut, one dead launch, one
  empty release. Owner corrections: scratch location, "stop asking me", the per-size question
  (+/- range), all early; later owner turns were picks and "go".

## 7. Key Learnings & Evergreen Upstream

- Test the real game shape with many seeds; prefer one explainable rule plus a filter over tuned
  tables → `docs/Architecture.md`.
- Reviewer numbers are hypotheses; developers measure → `docs/AgentWorkflow.md`.
- Screen tests must cover every interactive state → `AGENTS.md`, `docs/Architecture.md`.
- Design via competition, pick, merge, build; edit `.ansi` via `.tui` → `docs/DesignWorkflow.md`.
- Background agent launches in plain bash with explicit `--role` → `AGENTS.md`.
- `make dist` must produce versioned archives for `harnez release` → `docs/Architecture.md`.
- Help screens with examples and `-o`/`--in-place` flags are what make CLIs agent-friendly →
  loom 386.

## 8. File & Diff Summary

- **Game:** `internal/game/{game,tasks,screen,terminal*}.go`, `internal/budget/spec.go`,
  `golden.go`, `main.go`.
- **Spec:** `spec/{pricing.yaml,game-rules.md,tasks.yaml,templates.yaml,golden.yaml}`.
- **Design:** `design/{opus,argon,astra}` (004), `design/flames-{sol,opus}` (007),
  `design/game-{sol-c,opus-a,*-r2-*,final}` (010).
- **Tests:** `internal/game/{screen,result,joker,tasks,game}_test.go`, `internal/budget/round_test.go`.
- **Docs/harness:** `docs/{Architecture,DesignWorkflow,AgentWorkflow,README}.md`, two studies,
  `AGENTS.md`, `.harnez/`, bundled practice docs.
- **Build:** `Makefile` (`dist` archives, `test-q1`). Release v0.1.1 (annotated tag d7ba290 on commit 655ce6c).

## 9. Follow-Up Tickets & Review

- Open: tokburn 001 (epic: T4, T5–T7, T8, T9 open; the golden generator and a five-task pilot
  are delivered, full calibration is not, so task tiers stay provisional), 002 (v2 epic), 003,
  005 (surcharge), 011 (draw bias); loom 386; harnez 961, 962 (blocks 001 T6). No new tickets
  from this story: the parked 010 findings are recorded in 010.
- Second-vendor review: codex gpt-6.1-sol (med), all five findings applied. A first run under the
  name `story-review` returned another repo's review (rated via `harnez rate`).
