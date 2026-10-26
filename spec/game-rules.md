# Quota Defense game rules

`tokburn play` draws 20 cards; `tokburn play -short` draws 7. Both modes shuffle the embedded
deck uniformly and take the configured number of cards. A round is accepted only when it
contains at least two cards of each required tier; the draw does not force an exact tier mix.
Round sizes (`game_round_tasks: 20`, `game_round_tasks_short: 7`) and minimums live in
`pricing.yaml`. Both modes keep `min_per_tier: 2`.

On Linux and macOS terminals, play opens the opus title menu. Arrow keys select the
custom short round, the normal 20-task round, or Exit; `-` and `+` adjust the short
round between `game_round_tasks_min: 6` and `game_round_tasks_max: 7`. The initial
count is `game_round_tasks_short: 7`. Every custom short size uses the short penalty
and score threshold. Piped stdin skips the menu and accepts one l/m/h/j answer per
line; `-short` still selects seven tasks. Other platforms retain line input.
The screens target 80 columns by 24 rows; `q`, Ctrl-C or Ctrl-D exits raw-key play
and restores terminal input. No TUI dependency is required.

The budget is the cost of routing every drawn card to mid, including the normal redo cost and any
break penalty, multiplied by `budget_all_m_factor: 1.02`. This gives the all-m route a 2% cushion.
If a choice is below the card's needed tier, the task breaks and is redone at its needed tier. The
break adds `break_penalty_redo_factor: 0.27` times that card's redo cost. Prices and both factors
are configured in `pricing.yaml`. Short mode uses its own
`break_penalty_redo_factor_short: 0.10`; the budget factor remains 1.02 in both modes.

The golden cost is the cost of routing each drawn card at its needed tier. Score is
`golden cost / actual cost x 100`; good is `good_score: 80`, and right-tier play scores 100. The
short-mode good threshold is `good_score_short: 90`. The all-m easter egg is named on the end
screen when every answer is mid.

`TestMenuSeededStrategies` also runs seeds 0 through 99,999 for every integer in
the menu range, retaining the same survival and score assertions. Six-task ranges
are right-tier 100.00–100.00, all-m 53.35–78.43, all-low 82.96–87.32, and
all-high 21.04–66.67. The range deliberately stops at seven: the same probe found
all-high survives at eight tasks (seed 16565), so a continuous range through the
larger sizes would violate the rules.

`TestSeededStrategies` runs seeds 0 through 99,999 for each mode using real configured draws
and guarantees that right-tier play scores 100 and survives, all-m survives and scores below
that mode's good threshold, all-low scores below that threshold, and all-high goes over budget. Measured score ranges across those seeds:

| Strategy | Default (20 tasks) | Short (7 tasks) |
| --- | ---: | ---: |
| Right-tier | 100.00–100.00 | 100.00–100.00 |
| All-m | 41.33–78.20 | 49.22–81.31 |
| All-low | 71.62–79.50 | 82.25–87.56 |
| All-high | 13.37–60.69 | 17.63–69.97 |

Short rounds have more variable cost mixes. At the default penalty 0.27, all-high can survive
(seed 6622 leaves $6.73848). Lowering only the short penalty to 0.10 makes all-high bust across
the tested seeds; the smallest overspend is $2.313. With that lower penalty, all-low can score
87.56, so the short good threshold is 90. These short-only settings preserve the default
20-task rules and every strategy assertion.

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

Deck loading validates that both configured rounds can meet their tier minimums; invalid draws also
return an error instead of retrying forever. `go test -short ./...` samples the first 100 seeds;
`make check` retains the full 100,000-seed strategy regression.
