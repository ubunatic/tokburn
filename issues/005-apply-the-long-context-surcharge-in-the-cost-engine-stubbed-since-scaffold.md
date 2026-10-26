# 005 — Apply the long-context surcharge in the cost engine (stubbed since scaffold)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 001 (game scoring), 004 (design figures assume the surcharge)

---

## 1. Problem & Motivation

`spec/pricing.yaml` defines `cliff_tokens`, `cliff_inclusive` and `surcharge_multiplier`, but the
cost engine ignores them: `internal/budget/engine.go:18` still has
`TODO(stage): apply spec.surcharge_multiplier at the long-context cliff.` So `tokburn check` and the
game under-price prompts at or above 100k tokens (for example `classic-12`, 400k tokens). Found by
the 004 design fixer, whose screen mock-ups use spec-correct budgets that the game does not show.

/goal The engine applies the surcharge as the spec says, `tokburn check` and the game use it, and
the 100k-seed strategy test (both round modes, see `spec/game-rules.md`) stays green, retuning only
spec values if needed; or stop and report which rule cannot hold. Circuit breaker: on a loop or no
progress, clear the goal and write a failure report here.

## 2. Verification

Unit tests at, just below and just above the cliff (inclusive edge), `make check`, `make install`,
and updated measured ranges in `spec/game-rules.md`.
