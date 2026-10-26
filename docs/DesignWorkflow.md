# Design Workflow

How screens get designed and built. Used for the title screen (004 → 006), the flames
(007 → 008) and the game screen (010).

## Competition, pick, merge, build

1. **File a design ticket** with the requirements and the designers. No code in this step.
2. **Two or three designers in parallel** on different models (e.g. codex:sol:med, opus:med),
   each writing `.ansi` files to `design/<name>/`. Asking each for two or three *different
   ideas* gives the owner more to pick from than one design per model.
3. **Validate** every file with `loom measure`, `loom eval` (80 columns, at most 24 rows) and
   `loom check-box` when boxes are used.
4. **Owner views and picks** — or names parts to merge ("layout from A, cards from B"). Run a
   merge round on the favourites. Picks take seconds; each round narrows the design.
5. **Commit the kept designs as references**; park the rest in `.harnez/scratch/`.
6. **Build ticket:** developer (codex:sol:low) builds from the sketch, reviewer (opus:low)
   checks. Use `/tui-check`: cell-by-cell tests against the sketch, bounds tests for every
   live value, and a live tmux run.

## Editing .ansi files: the .tui round trip

Agents draft `.ansi` well but edit it badly: the ESC byte is invisible to the read tool and
cannot be typed into an exact-match edit, identical rows cannot be told apart, and trailing
spaces get lost. Instead:

```sh
loom markup design.ansi > .harnez/scratch/design.tui   # readable tags, visible spaces
# edit the .tui with the Edit tool
loom markup .harnez/scratch/design.tui > .harnez/scratch/design.ansi
loom eval .harnez/scratch/design.ansi                   # widths
loom diff design.ansi .harnez/scratch/design.ansi       # only the intended cells changed
```

Prefer this over generator scripts. `.tui` is the owner's own markup format, defined by loom;
`.ansi` stays the committed format. Upstream improvements (`markup -o`, `--in-place`, agent
docs) are loom issue 386.

## Pitfalls

- A sketch shows one state; the build must handle all of them (every menu selection, joker
  revealed, two-digit task numbers, large and negative budgets).
- Font-dependent glyphs (block digits, flames) still need the owner's eyes in the real terminal.
