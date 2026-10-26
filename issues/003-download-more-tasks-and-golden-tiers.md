# 003 — Download more tasks and golden tiers

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 001

---

## Problem

The binary embeds about 1000 tasks with golden tiers (epic 001). New task packs should reach players
without a new release.

## Idea

Publish task packs (tasks plus golden tiers) as release assets; `tokburn` downloads a pack into
`~/.config/tokburn/` on request and merges it with the embedded deck. AI players keep getting only
task batches without tiers (epic 001, decision 2).

## Acceptance

- `tokburn tasks update` downloads and verifies a pack; `play` uses it; an offline run still works
  with the embedded deck.
