# Quota Defense game rules

`tokburn play` draws 20 cards; `tokburn play -short` draws 7. Both modes draw from
the configured deck. The menu supports every round size from 6 through 20, inclusive;
`-` and `+` change the custom size by one, without gaps. `min_per_tier: 2` remains
unchanged.

A candidate draw is accepted only when it contains at least two cards of each tier
and the cost of routing every card to high exceeds the round budget. Otherwise the
deck is shuffled and drawn again. The all-m route defines the round budget: its cost,
including redo costs and the `break_penalty_redo_factor: 0.27` break penalty, times
`budget_all_m_factor: 1.02`. The shared `good_score` is 86 for all sizes, including
the 20-task default. The draw rule ensures every accepted round makes all-high bust.

The observed rejection rate is above the requested 0.2% for this deck and budget.
The table shows rejected shuffled candidates divided by all candidates for seeds
0–9,999 at each size, except 7 and 20, which use seeds 0–99,999. Thus the stated
0.2% rejection-rate target is not met by the shared 0.27 penalty and this uniform
shuffle sampler; changing the good-score threshold cannot change that rate.

| Tasks | Penalty | Good | All-m score min–max | All-low score min–max | Rejection rate |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 6 | 0.27 | 86 | 53.66–73.01 | 73.07–76.61 | 87.3138% |
| 7 | 0.27 | 86 | 47.03–77.26 | 72.37–77.62 | 70.2150% |
| 8 | 0.27 | 86 | 44.58–79.82 | 72.01–78.14 | 53.6672% |
| 9 | 0.27 | 86 | 43.08–81.03 | 71.91–78.49 | 39.7300% |
| 10 | 0.27 | 86 | 41.23–81.70 | 71.70–78.83 | 28.6377% |
| 11 | 0.27 | 86 | 40.47–82.54 | 71.57–78.83 | 20.4771% |
| 12 | 0.27 | 86 | 39.09–80.44 | 71.55–79.12 | 14.4861% |
| 13 | 0.27 | 86 | 38.46–84.70 | 71.27–79.40 | 9.7961% |
| 14 | 0.27 | 86 | 36.82–79.74 | 71.69–79.96 | 6.7077% |
| 15 | 0.27 | 86 | 39.23–79.38 | 71.73–79.33 | 4.3611% |
| 16 | 0.27 | 86 | 38.95–79.17 | 71.60–79.51 | 3.0256% |
| 17 | 0.27 | 86 | 42.13–78.72 | 71.82–78.89 | 1.8935% |
| 18 | 0.27 | 86 | 40.30–78.53 | 71.71–79.45 | 1.2151% |
| 19 | 0.27 | 86 | 42.75–76.01 | 71.76–78.39 | 0.7149% |
| 20 | 0.27 | 86 | 41.33–78.20 | 71.62–79.50 | 0.5430% |

`TestMenuSeededStrategies` checks right-tier survival and a displayed score of 100,
all-m survival below good, all-low below good, and all-high busting. It uses 10,000
seeds for sizes 6, 8–19 and 100,000 seeds for sizes 7 and 20. Menu bounds and shared
rules are checked from `pricing.yaml`.

Each round starts with `joker_start: 1` joker. Press `j` or `0` to reveal the current task's tier
(provisional until T4 calibrates the golden baseline), without answering or advancing the task.
The first reveal consumes one joker; repeating a reveal on that same task is free. With no jokers
left, an unrevealed task stays hidden. Reveals do not change costs or scores.

A combo of `joker_combo_answers: 3` consecutive correct, unassisted answers, each strictly under
`joker_fast_seconds: 3` seconds, earns `joker_combo_reward: 1` joker. Timing starts when the task
is displayed and includes invalid input. Exactly three seconds is slow. A slow or wrong answer,
a reveal request, or earning a reward resets the combo. Assisted answers cannot build a combo.
Jokers, reveals and combos never carry into another round. The game accepts an injected clock
for deterministic threshold tests.

Deck loading validates that configured rounds can meet their tier minimums. An invalid draw
returns an error. `go test -short ./...` uses the first 100 seeds; `make check` runs the
size-specific seed regression described above.
