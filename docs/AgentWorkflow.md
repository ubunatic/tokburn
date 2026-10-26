# Agent Workflow

How work on tokburn is split between the orchestrator and `harnez agent` workers.

## Model routing (owner, recorded in epics 001 and 002)

| Work | Model |
|---|---|
| Search and explore | luna:med |
| Development | codex:sol:low or codex:sol:med |
| Designs | codex:sol:med, opus:med (in competition) |
| Key reviews | agy:argon:low |
| Normal reviews | opus:low, or the orchestrator |
| Decisions instead of asking the owner | resumable advisor `tokburn-advisor` (codex:sol:med) |

The owner still approves big golden or AI runs, pushes and releases.

## Running agents

- Always in the background with `HTO=0`. Launch loops must be plain bash (the tool shell is
  bash, not zsh): a zsh-only `${=p}` once started no agent at all.
- Give the role explicitly (`--role developer|reviewer`). harnez once inferred "reviewer" from a
  developer prompt; resuming with a different role is refused, so a new session was needed.
- A developer cannot start other agents (leaf-worker guard). Runs that need agents (e.g.
  `tokburn golden`) are done by the orchestrator.
- Workers never commit; the orchestrator checks, commits, runs `make install` and closes the
  ticket with `harnez issues close --commit`.
- Scratch goes to `.harnez/scratch/` (gitignored), never `/tmp`.

## Review practice

- Developer → reviewer → fixes → orchestrator re-runs tests uncached before committing.
- Treat reviewer numbers as hypotheses. One review predicted ≤0.2% redraws; the developer
  measured up to 87% and reported the real figure.
- Ask reviewers to check every state, not just the sketch state; two bugs (`$-0.00`, stale `]`
  highlight) slipped past tests that only covered the default state.

## Studies

- `docs/studies/2026-10-26-quota-defense-v1-screens-and-design-rounds.md`
