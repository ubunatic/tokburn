# 001 — Epic: Quota Defense v1.1 — a real game, a joker, and AI players

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Epic
**Related**: ubunatic.com 120 (30-minute live-build talk)

---

## 1. Goal

/goal Turn Quota Defense from a six-question quiz into a game worth replaying: a larger task deck,
an "all m" easter egg that always survives but never scores high, a once-per-game joker `j` that can
also be earned by fast answers and reveals the right tier, and an AI mode where local or cloud deciders
(Jev via `harnez decide`, neus, Clef, harnez agents) play batches of tasks over simulated weeks or
months, scored against a hidden golden baseline (the averaged tier choice of argon, opus and sol),
so vendors can be compared; or stop and report when blocked.

### Exit Clause

1. The deck has at least 40 tasks across low, mid and high, and a round draws a fresh mix.
2. A test plays every seed in a range with "m" only: each run survives the month, and each score is
   below the "good" threshold. The end screen names the easter egg.
3. `j` reveals the right tier once per game; fast answers (under a set time) earn extra jokers.
   Covered by tests with a fake clock.
4. `tokburn play --ai <decider>` plays a batch-based game over N simulated weeks without a human,
   and `tokburn bench` prints a per-decider comparison table (score, survival, cost, wrong-tier rate).
5. Golden tiers live outside the repo; the game, tests and AI prompts never contain them.
6. `make check` and `make install` pass; the runtime canary (one human round, one AI round) is recorded.

### Circuit Breaker

Clear the goal and write a failure report here (what was tried, why it failed, resume point) if:
- a worker loops twice on the same review finding,
- the hidden-golden requirement cannot be met without breaking audience installs,
- no decider backend answers reliably enough to finish a benchmark run.

---

## 2. Design Notes

- **Scoring that punishes "all m" without killing it.** Mid on a low task works but costs more; mid on
  a high task breaks and is redone. Tune the deck and budget so all-m always ends above $0 but low.
  Score = budget saved versus the golden route, so only right-tier play scores high.
- **Joker.** One per game. Earned extras: answer within N seconds on K tasks in a row (combo).
  Reveal shows the golden tier for the current task only.
- **Batches.** The game hands a decider a batch (one simulated week) as a list of tasks with task text,
  tokens and runs, never the tier. The decider returns one tier per task. The game scores the batch,
  updates the budget and hands over the next week.
- **Golden baseline.** For each task, ask argon, opus and sol for a tier (low=0, mid=1, high=2), average
  and round. Store the result once in a file outside the repo, e.g. `~/.config/tokburn/golden.yaml`.
  The repo keeps only a small test fixture with made-up tasks.
- **Deciders.** A `Decider` interface with one method: batch in, tiers out. Adapters:
  `human` (keyboard), `all-m`, `jev` (`harnez decide --choice`), `neus` (`neus ask`),
  `agent:<model>` (`harnez agent --role advisor`), `clef` (see open decisions).

## 3. Open Decisions

1. **Where players get the golden file.** Hidden from agents, but every human player needs it to score.
   Recommendation: publish it as a release asset; tokburn downloads it into `~/.config/tokburn/` on
   first play. AI mode runs the decider in a temporary directory without that path.
   Limit: an agent with full disk access could still read it; this hides it, it does not secure it.
2. **What "Clef" is.** No repo or doc mentions it. Owner to name the tool or drop the adapter.
3. **Talk timing.** The live-build tags (`stage-start`, `stage-2-buggy`, `stage-4-fixed`) sit on the
   current game. If v1.1 should be what the audience installs, rebase the tags onto it and re-run the
   rehearsal. Recommendation: land v1.1 before the talk and re-tag.
4. **Cost of AI mode.** Cloud deciders over many simulated months cost money. Recommendation: default
   to 4 weeks, cap `--weeks`, and print the expected number of model calls before starting.

## 4. Tasks

Placeholders until the code fanout (epic skill step 3) and refinement (step 4) have run.
Prior-art web fanout skipped: this is game logic and adapters for in-house tools.

| Ticket | What | Sprint Group | Searcher → Architect → Developer | Reviewer | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| T1 | Deck to 40+ tasks in `spec/`, without tiers | A Game | haiku → — → codex:sol:low | orchestrator | Open |
| T2 | All-m easter egg: tuning + every-seed test + end-screen line | A Game | — → — → codex:sol:low | agy:argon:low | Open |
| T3 | Joker `j` once per game, extra jokers for fast-answer combos (fake clock) | A Game | — → — → codex:sol:low | orchestrator | Open |
| T4 | Golden baseline: ask argon/opus/sol per task, write file outside repo, loader + fixture | B Golden | — → opus → codex:sol:low | agy:argon:low | Open |
| T5 | `Decider` interface, batch loop over simulated weeks, `human` + `all-m` deciders | C AI | — → opus → codex:sol:low | agy:argon:low | Open |
| T6 | Deciders `jev` (`harnez decide`), `neus`, `agent:<model>`; `clef` after decision 2 | C AI | haiku → — → codex:sol:low | orchestrator | Open |
| T7 | `tokburn bench`: per-decider comparison table | C AI | — → — → codex:sol:low | agy:argon:low | Open |
| T8 | Re-tag stage checkpoints and update the ubunatic.com runbook (after decision 3) | D Talk | — → — → host | owner | Open |
| T9 | `/evergreen` | Final | host | owner | Open |
| T10 | `/wrap` | Final | host | owner | Open |

Phases: A and B are independent. C needs A (deck) and B (golden). D needs A–C.

## 5. Out of Scope

- The real-time arcade interface (epic 002).
- A shared online leaderboard.
