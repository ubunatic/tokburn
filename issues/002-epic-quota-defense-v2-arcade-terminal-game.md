# 002 — Epic: Quota Defense v2 — real-time arcade game in the terminal

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Epic
**Related**: 001

---

## 1. Goal

/goal Ship Quota Defense v2 as a real-time, colourful terminal arcade game: tasks fall down the screen
in their tier's colour, the player routes each with `l`/`m`/`h` before it lands, speed rises over time,
fast correct routing builds combos, and a "monorepo paste" bomb punishes crossing the long-context cliff;
installable with `go install` and as prebuilt binaries for Linux, macOS and Windows; or stop and report
when blocked.

### Exit Clause

1. `tokburn play` opens the arcade view (Bubble Tea or a lighter library, see decision 1); `--classic`
   keeps the v1.1 text mode, and AI mode (001) keeps working.
2. Falling tasks, tier colours, speed-up, combos, the monorepo bomb and a score screen work, checked
   cell by cell against an `.ansi` sketch (`/tui-check`).
3. Game logic (spawning, timing, scoring) is tested without a terminal, using a fake clock.
4. The release ships prebuilt binaries for 5 targets, and the README shows the download link next to
   `go install`.
5. Playable on an 80x24 terminal; degrades to plain colours when the terminal has no 256-colour support.
6. `make check` and `make install` pass; a recorded playtest with the owner.

### Circuit Breaker

Clear the goal and write a failure report here if a worker loops twice on the same finding, the chosen
TUI library cannot render reliably in common terminals, or the game loop cannot be tested without a TTY.

---

## 2. Open Decisions

1. **TUI library.** Recommendation: Bubble Tea plus Lip Gloss (maintained, widely used). Check weight
   and alternatives in the prior-art fanout first.
2. **Balance.** Start speed, speed-up curve, combo multiplier and bomb frequency. Recommendation: tune
   in owner playtests, with values in `spec/` rather than code.
3. **Audience use.** Whether the talk should end with v2 instead of v1.1. Recommendation: no, keep the
   talk on v1.1 and announce v2 as "next".

## 3. Tasks

Placeholders until the prior-art fanout (epic skill step 2, default on: new UI dependency) and the
code fanout have run.

| Ticket | What | Sprint Group | Searcher → Architect → Developer | Reviewer | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| T1 | Prior-art fanout: TUI libraries, terminal game examples, colour fallbacks | A Discovery | haiku + luna → — → — | host | Open |
| T2 | Terminal-free game loop: spawn, fall, route, combo, bomb, fake clock, tests | B Engine | — → opus → codex:sol:low | agy:argon:low | Open |
| T3 | Arcade view with tier colours and score screen, `.ansi` sketch + `/tui-check` | C View | — → — → codex:sol:low | agy:argon:low | Open |
| T4 | Balance values in `spec/`, owner playtest | C View | — → — → host | owner | Open |
| T5 | Prebuilt binaries in the release, README download link | D Release | — → — → codex:sol:low | orchestrator | Open |
| T6 | `/evergreen` | Final | host | owner | Open |
| T7 | `/wrap` | Final | host | owner | Open |

Phases: A, then B, then C; D can run after B.

## 4. Out of Scope

- A browser version and a live audience scoreboard (both need hosting; file separately if wanted).
