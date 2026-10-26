# 008 — Animate the sol flames above the title on the title screen

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 006 (opus title screen), 007 (flame designs, owner picked sol)

---

## 1. Goal

/goal The `tokburn play` title screen loops the frames from `design/flames-sol/` in place of the
static title and dots, while the menu keeps responding to keys without lag. Verified with
`/tui-check` (each frame cell by cell against its `.ansi` file) and live in tmux; `make check` and
`make install` pass. Stop and report when blocked on an owner decision. Circuit breaker: on a loop
or no progress, clear the goal and write a failure report here.

## 2. Notes

- The sol frames are 48×12, taller than the title area in `design/opus/title.ansi`; fit the rest
  of the screen into 80×24 without wrapping, or report the conflict.
- Frame delay goes in spec yaml (preview used 0.15 s).
- Animation stops when the game screen starts; piped stdin keeps skipping the title screen.
- Terminal restore on exit and signals must still work.
