# 011 — Measure whether the all-high-busts redraw biases short rounds

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Investigation
**Related**: 001 (game rules), study `docs/studies/2026-10-26-quota-defense-v1-screens-and-design-rounds.md`

---

## 1. Problem & Motivation

Since bb5b0bf, `DrawTasks` redraws until each tier has at least `min_per_tier` cards **and** the
all-high route busts the round budget. The reviewer predicted at most 0.2% rejected draws; the
measured rate is 0.54% at size 20 but 87.31% at size 6. Rounds stay random, but short rounds may
lean towards task mixes where high is expensive, which could make them more guessable — the owner
explicitly wants rounds that are not guessable.

/goal Measure the tier and task distribution of accepted short rounds (sizes 6–9) against
unfiltered draws, and either show the bias is negligible or propose a spec change (e.g. a
different deck mix or budget rule) that keeps all four strategy checks green. Report to the
owner before changing rules. Circuit breaker: on a loop or no progress, clear the goal and write a
failure report here.
