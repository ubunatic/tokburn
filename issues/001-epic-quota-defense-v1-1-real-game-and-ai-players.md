# 001 — Epic: Quota Defense v1.1 — a real game, a joker, and AI players

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Epic
**Related**: 002, 003, ubunatic.com 120 (talk use case designed later)

---

## 1. Goal

/goal Turn Quota Defense from a six-question quiz into a game worth replaying: about 1000 embedded tasks
with golden tiers, an "all m" easter egg that always survives but never scores high, a once-per-game
joker `j` that can also be earned by fast answers and reveals the right tier, and an AI mode where
deciders (Jev, Clef, neus, harnez agents) play batches of tasks over simulated weeks or months, scored
against the golden baseline (the averaged tier choice of argon, opus and sol), so vendors can be
compared; or stop and report when blocked.

### Exit Clause

1. The binary embeds about 1000 tasks with golden tiers; a round draws a fresh mix.
2. A test plays every seed in a range with "m" only: each run survives the month, and each score is
   below the "good" threshold. The end screen names the easter egg.
3. `j` reveals the right tier once per game; fast answers (under a set time) earn extra jokers.
   Covered by tests with a fake clock.
4. `tokburn play --ai <decider>` plays a batch-based game over N simulated weeks without a human,
   and `tokburn bench` prints a per-decider comparison table (score, survival, cost, wrong-tier rate).
5. AI players only ever see a batch of tasks without tiers (decision 2).
6. `make check` and `make install` pass; the runtime canary (one human round, one AI round) is recorded.

### Circuit Breaker

Clear the goal and write a failure report here (what was tried, why it failed, resume point) if:
- a worker loops twice on the same review finding,
- a decider cannot be called without file, internet or skill access,
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
  and round. Generated once, stored in `spec/`, embedded in the binary.
- **Deciders.** A `Decider` interface with one method: batch in, tiers out. Adapters:
  `human` (keyboard), `all-m`, `jev` (Jev decider API via `harnez decide`), `clef` (the open-source
  Clef decider model, run locally), `neus` (`neus ask`), `agent:<model>` (`harnez agent`).

## 3. Decisions (owner, 2026-10-26)

1. **Clef** is an open-source decider model; **Jev** is a decider API. Both get adapters (T6).
2. **Hiding the answers.** tokburn itself calls the AI players, so they never see the binary. Each call
   runs without file access, internet access, skills or tools: the batch text is the only input.
   Human players run the binary, which embeds the answers.
3. **More tasks by download** is later work: ticket 003.
4. **Talk.** Not decided. Focus on a good, fun v1.1 and v2 first; design the talk use case (a simulated
   new feature or a planted bug) afterwards.

## 4. Open Decisions

1. **Cost of AI mode and of the golden run.** 1000 tasks x 3 models is about 3000 model answers once;
   cloud deciders over many simulated months add more. Recommendation: send tasks in batches of 50,
   default AI games to 4 weeks, cap `--weeks`, and print the expected number of calls before starting.
2. **Where the 1000 tasks come from.** Recommendation: generate them with a cheap model from about 30
   hand-written templates (task kind x repo size x run count), then the owner spot-checks 50.

## 5. Tasks

Placeholders until the code fanout (epic skill step 3) and refinement (step 4) have run.
Prior-art web fanout skipped: this is game logic and adapters for in-house tools.

| Ticket | What | Sprint Group | Searcher → Architect → Developer | Reviewer | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| T1 | ~1000 tasks in `spec/`, embedded, generated from templates (open decision 2) | A Game | haiku → — → codex:sol:low | owner spot-check | Open |
| T2 | All-m easter egg: tuning + every-seed test + end-screen line | A Game | — → — → codex:sol:low | agy:argon:low | Open |
| T3 | Joker `j` once per game, extra jokers for fast-answer combos (fake clock) | A Game | — → — → codex:sol:low | orchestrator | Open |
| T4 | Golden baseline: argon/opus/sol tier per task, averaged, written to `spec/` | B Golden | — → opus → codex:sol:low | agy:argon:low | Open |
| T5 | `Decider` interface, batch loop over simulated weeks, `human` + `all-m` deciders | C AI | — → opus → codex:sol:low | agy:argon:low | Open |
| T6 | Isolated deciders `jev`, `clef`, `neus`, `agent:<model>`: no files, net, skills | C AI | haiku → — → codex:sol:low | agy:argon:low | Open |
| T7 | `tokburn bench`: per-decider comparison table | C AI | — → — → codex:sol:low | agy:argon:low | Open |
| T8 | `/evergreen` | Final | host | owner | Open |
| T9 | `/wrap` | Final | host | owner | Open |

Phases: A and B are independent (B needs T1's task list). C needs A and B.

## 6. Out of Scope

- The real-time arcade interface (epic 002).
- Downloading more tasks (ticket 003).
- The talk use case (designed after v1.1 and v2).
- A shared online leaderboard.
