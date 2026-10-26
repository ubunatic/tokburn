# tokburn

Check prompt sizes against the long-context price cliff, and play Quota Defense,
a game about routing agent tasks to the cheapest model tier that does the job.

```sh
go install ubunatic.com/tokburn@latest
tokburn check .      # tokens and cost per tier for this directory
tokburn play         # route 6–20 tasks to low/mid/high, keep the round budget alive
```

Prices live in `spec/pricing.yaml`. They are round example numbers, not any
vendor's real price list. At or above `cliff_tokens` the whole prompt should cost
`surcharge_multiplier` times more; the engine does not apply this yet (issue 005).

Game rules: `spec/game-rules.md`. Project docs: `docs/README.md`.

Built live on stage for the DecompileD 2027 30-minute talk
(`https://ubunatic.com/talk/live/`). The tags `stage-start`, `stage-2-buggy` and
`stage-4-fixed` are the stage checkpoints.
