# How a pass works

You are one auditor running one pass, in a fresh subagent started by the audit loop
(`README.md`, "A run"). Your instructions are `passes/<id>.md`; this file is the protocol around them.

## Before you start

- Read `passes/<id>.md`, `state/<id>.json` and the open section of `findings/<id>.md`. Read
  `orchestration/standards.md` sections only as your pass needs them; it overrides older docs.
- The role brief (`brief` in `auditors.json`) is background. Its parts about messaging workers or
  the orchestrator, sessions, PRs and scheduling passes no longer apply.
- Work in `$AUDIT_WORK/<id>/` (scratch; it disappears with the session). `source tools/audit/setup.sh`
  first if `$AUDIT_WORK` isn't set.

## What you may change

Only these, all under `orchestration/audit/` unless your registry entry `owns` a doc:

- `state/<id>.json`: when you finish, set `last.master` to the master commit you audited,
  `last_pass_at` (UTC, `YYYY-MM-DDTHH:MM:SSZ`), `pass` (+1), `result` (`quiet`, `findings` or `error`),
  `client_build` when your pass checked it, and any baseline fields your pass doc names.
- `findings/<id>.md`: see below.
- `reports/<id>.md`: this pass's report, replacing the last one. What you checked, compact tables,
  conclusions. At most about 200 lines.
- `ref/<id>/`: new baselines, when your pass doc says to refresh them. Keep it small (well under
  2 MB). Never commit binaries or downloaded client tables.
- The docs your registry entry `owns` (e.g. `orchestration/rank_tuning.md`).

Never commit or push (the run commits once at the end), never edit code, presets or other docs,
never message other sessions, never schedule anything, never open or comment on PRs.

## Findings

A finding is something wrong or missing that someone should fix or rule on: a wrong number, a
broken or non-Forever mechanic, missed DPS, an unfair asymmetry. Not a nit, not a restatement of a
settled ruling (your pass doc lists those), and not something already open or resolved in your
findings file. New evidence on an existing finding goes under that finding.

Add each new one at the top of `## Open` in `findings/<id>.md`:

```
### <id>-<YYYYMMDD>-<n>: <one-line title>
- severity: high (moves DPS by 1% or more, or a spec's place in the ranking) | medium | low
- affects: <classes / specs / builds>
- found: <UTC date>, master <sha>
- evidence: <file:line, numbers, the command that reproduces it, source and tier for Forever facts>
- proposed fix: <exact change, or the options with their measured effect>
- needs: user ruling | orchestrator fix | worker fix
```

Say how sure you are. Measure before you claim DPS ("+3.2 DPS, paired seeds, 20k iterations"),
and estimate only where measuring isn't possible, saying so.

When a pass sees that an open finding was fixed on master, move it under `## Resolved` with
`- resolved: <date>, <commit>`.

## Keeping your context small

- Never print a whole sim result, combat log, db2 table or large diff. Write it to a file under
  `$AUDIT_WORK/<id>/` and read what you need with `compare_runs.py`, `jq`, `grep`, `head` or a short
  script that prints a summary.
- Reuse the shared run in `$AUDIT_WORK/runs/master` when the loop gave you one; sim only what it
  doesn't cover (other seeds, debug logs, hand-edited requests, an older build for a regression via
  `tools/audit/build_cli.sh <sha>`).
- Use your own helpers in `tools/audit/<id>/` and add to them when a step is worth keeping: the next
  pass is a stranger who has only the repo.
- If the pass would need far more work than its doc expects (a whole new client build, a broad
  regression), do the most important part, record what's left as a finding or an open item in
  your report, and stop. Don't run for hours.

## Your reply

At most 10 lines, for the run's log and notification: pass number, master sha, quiet or not, each
new finding's id, severity and title, and anything that blocked you.
