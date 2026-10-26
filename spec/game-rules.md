# Quota Defense game rules

Each round draws 20 cards uniformly by shuffling the embedded deck, then taking the first 20. A
round is accepted only when it contains at least two cards of each required tier (`min_per_tier: 2`);
the draw does not force an exact tier mix. Round size and minimums live in `pricing.yaml`.

The budget is the cost of routing every drawn card to mid, including the normal redo cost and any
break penalty, multiplied by `budget_all_m_factor: 1.02`. This gives the all-m route a 2% cushion.
If a choice is below the card's needed tier, the task breaks and is redone at its needed tier. The
break adds `break_penalty_redo_factor: 0.27` times that card's redo cost. Prices and both factors
are configured in `pricing.yaml`.

The golden cost is the cost of routing each drawn card at its needed tier. Score is
`golden cost / actual cost x 100`; good is `good_score: 80`, and right-tier play scores 100. The
all-m easter egg is named on the end screen when every answer is mid.

`TestSeededStrategies` runs seeds 0 through 99,999 using real 20-card draws and guarantees that
right-tier play scores 100 and survives, all-m survives and scores below 80, all-low scores below
80, and all-high goes over budget. Measured score ranges across those seeds:

| Strategy | Score range |
| --- | ---: |
| Right-tier | 100.00–100.00 |
| All-m | 41.33–78.20 |
| All-low | 71.62–79.50 |
| All-high | 13.37–60.69 |

Any change to the embedded deck or model prices must keep the 100,000-seed test green; update the
measured ranges here after verifying it.

Each round starts with `joker_start: 1` joker. Press `j` to reveal the current task's tier
(provisional until T4 calibrates the golden baseline), without answering or advancing the task.
The first reveal consumes one joker; repeating a reveal on that same task is free. With no jokers
left, an unrevealed task stays hidden. Reveals do not change costs or scores.

A combo of `joker_combo_answers: 3` consecutive correct, unassisted answers, each strictly under
`joker_fast_seconds: 3` seconds, earns `joker_combo_reward: 1` joker. Timing starts when the task
is displayed and includes invalid input. Exactly three seconds is slow. A slow or wrong answer,
a reveal request, or earning a reward resets the combo. Assisted answers cannot build a combo.
Jokers, reveals and combos never carry into another round. The game accepts an injected clock
for deterministic threshold tests.

Deck loading validates that the configured round can meet its tier minimums; invalid draws also
return an error instead of retrying forever. `go test -short ./...` samples the first 100 seeds;
`make check` retains the full 100,000-seed strategy regression.
