# 007 — Burning flame animation above the tokburn title: two frame-set designs

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 004 (design competition), 006 (opus title screen)

---

## 1. Problem & Motivation

The opus title screen (`design/opus/title.ansi`, built in 006) shows the TOKBURN title with small
dots above it that already hint at fire. The owner wants these to become a burning animation:
flames on top of the title text.

/goal Two designers each deliver 4–10 `.ansi` frames of the TOKBURN title text with flames on top,
validated with loom; the owner views them and picks one. Stop there: building the animation into
the title screen is a follow-up ticket. Stop and report when blocked on an owner decision. Circuit
breaker: on a loop or no progress, clear the goal and write a failure report here.

## 2. Requirements (owner, 2026-10-26)

- Designers: codex:sol:med and opus:med, one design each, independently.
- Each delivers 4–10 frames in `design/flames-<model>/frame-NN.ansi`.
- Frames contain **only the TOKBURN title text and the flames above it**, not the full title
  screen (no menu, no hints, no box). Start from the title rows and colours in
  `design/opus/title.ansi`; the title text itself stays the same in every frame.
- All frames of a design share the same width and height, so they can loop in place.
- Validate each frame with `loom measure` and `loom eval`; fix with `loom format` / `loom fmt`.
- Work files go in `.harnez/scratch/`, not `/tmp`.

## 3. Owner review

View the frames in sequence (e.g. `loom view` per frame, or a small loop script in scratch) and
pick a design.
