# The audit loop

The DPS report's auditors (results, rotations, gear, consumables, mechanics, rank tuning, fairness)
run as one scheduled routine. Every run is a **fresh cloud session** that reads everything it needs
from this directory, so no auditor carries a growing context from pass to pass.

Why: on 2026-10-17 the seven auditors lived in their own long-running sessions with 150k to 700k
tokens of context each. Every wake re-read that context, so a quiet "nothing changed" pass cost
$2 to $6, and the auditors cost about $85 a day while nothing moved (token audit:
`.agents/skills/token-audit`). A fresh pass starts from about 40k tokens.

## Pieces

| Path | What it is |
|---|---|
| `auditors.json` | The registry: each auditor's watched paths, outside sources, model, whether it sims, and the docs it owns. |
| `PASS.md` | The protocol every pass follows: what it may touch, how it reports, how it keeps its context small. |
| `passes/<id>.md` | Each auditor's pass, step by step: what to check, the commands, the tolerances, what's settled. |
| `state/<id>.json` | What each auditor last audited: master commit, client build, pass number, result, baselines. |
| `findings/<id>.md` | Each auditor's findings: open ones on top, resolved ones below. This replaces messages to workers and the orchestrator. |
| `reports/<id>.md` | The latest pass report per auditor, overwritten each pass. |
| `ref/<id>/` | Small reference outputs a pass compares against (sim summaries, budget tables). |
| `handoff/<id>.md` | What the long-running auditor sessions handed over when they were retired (historical). |
| `log.md` | One line per pass, newest last. |
| `ROUTINE.md` | The routine's prompt and settings. |
| `tools/audit/` | The scripts: `gate.py` (who is due), `setup.sh`, `build_cli.sh`, `run_reqs.py`, `compare_runs.py`, `wago_latest.py`, `commit_state.sh`, and per-auditor helpers in `tools/audit/<id>/`. |

The role briefs (what each auditor is for, at length) stay where they were:
`orchestration/auditor_*.md` and `auditors_common.md`. A pass reads them only when its
`passes/<id>.md` says to; the sections there about messaging workers, sessions and 45-minute passes
no longer apply.

## A run

This is what the routine session does, start to finish. Run every command from the repo root.

1. **Set up.** `git checkout -q master && git pull -q origin master`, then
   `source tools/audit/setup.sh`. If the repo isn't checked out, or `git push --dry-run origin HEAD:master`
   is refused, reply in one line saying which and stop.
2. **Gate.** `git fetch -q origin --prune`, then `python3 tools/audit/gate.py` (a short table) and
   `python3 tools/audit/gate.py --json > $AUDIT_WORK/gate.json`. It needs no model: due means a
   watched file changed on master since that auditor's last pass, a newer client build appeared on
   wago.tools, or a periodic check is due (see `auditors.json`).
3. **Nothing due.** Run `python3 tools/audit/gate.py --advance`, then
   `tools/audit/commit_state.sh "Audit loop: quiet run"`, and reply with one line:
   `Audit loop <UTC time>: nothing due (master <sha>).` Stop.
4. **Shared sims.** If any due auditor has `"sims": true`:
   - `CLI=$(tools/audit/build_cli.sh origin/master)` (a few minutes; cached per commit);
   - `python3 tools/audit/run_reqs.py $CLI $AUDIT_WORK/runs/master` (all report requests, 20k
     iterations, seed 1);
   - `python3 tools/audit/compare_runs.py --published $AUDIT_WORK/runs/master/summary.tsv > $AUDIT_WORK/runs/vs_published.txt`.
   Every sim pass starts from this run instead of simming the same requests again.
5. **Passes.** For each due auditor, start a subagent (the Agent tool) with the auditor's `model`
   from the gate output and this prompt, filled in:

   > You are the <title> auditor, running one pass of the forever-sim audit loop. Read
   > orchestration/audit/PASS.md, then orchestration/audit/passes/<id>.md, and do the pass.
   > Why you're due: <reasons>. Changed files: <files>. Master: <sha>. Shared sims (if any):
   > CLI <path>, run $AUDIT_WORK/runs/master, comparison $AUDIT_WORK/runs/vs_published.txt.
   > Reply in at most 10 lines.

   Passes without sims can run side by side in the background. Run sim passes one at a time:
   they share the CPUs.
6. **Record.** For each pass, append one line to `log.md`:
   `<UTC date time> | <id> | pass <n> | master <sha> | <quiet / N new findings / error: ...> | <one-line summary>`.
   Then `python3 tools/audit/gate.py --advance` (moves the auditors that weren't due) and
   `tools/audit/commit_state.sh "Audit loop: <ids> (<N> new findings)" <owned docs of the auditors that ran>`,
   ending the commit message with your session's attribution trailer. Never commit anything else.
7. **Reply.** Your last message is the phone notification, 2 to 4 lines:
   - which auditors ran and why, in a few words;
   - the new findings, one line each with severity, or "no new findings";
   - when there are new findings, a link to them:
     `https://github.com/bayov/forever-sim/blob/master/orchestration/audit/findings/<id>.md`.

   Don't message other sessions, don't schedule anything, and don't act on findings: fixing them is
   the orchestrator's job (`.agents/skills/dps-report`, block B13), started by the user.

## Running a pass by hand

From any session: `source tools/audit/setup.sh`, then follow steps 2 to 6 with
`gate.py --force <id>`. To try a new pass doc without committing, skip `commit_state.sh` and
`git checkout -- orchestration/audit` afterwards.

## Changing the loop

- **A new auditor:** add it to `auditors.json`, write `passes/<id>.md` (copy an existing one's shape),
  add `state/<id>.json` with `{"last": {"master": "<sha to start from>"}}`, and `findings/<id>.md` with
  the two headings. The next run picks it up.
- **Watched paths:** edit `watch` / `exclude` (git globs). `python3 tools/audit/gate.py` shows the
  effect right away.
- **Cadence:** the routine runs every 6 hours (`ROUTINE.md`). A quiet run costs well under a dollar:
  the session reads this file, runs the gate and stops.
- **Pausing one auditor:** set `"enabled": false` in `auditors.json`.

## Keeping passes cheap

These rules are in `PASS.md` for the passes; they apply to any agent session in this repo:

- **Start fresh.** A pass, a fix or an investigation starts in a new session or subagent with a
  short written brief. Don't keep a session alive to "remember" things: write them to the repo.
- **Keep big output out of the context.** Sim results, logs, diffs and tables go to files under
  `$AUDIT_WORK`. Read summaries (`compare_runs.py`, `head`, `grep`), never whole result JSON or
  combat logs.
- **Message only when it matters.** Findings go to `findings/<id>.md`. Nothing is sent to other
  sessions, and quiet passes write one log line.
- **Pick the model for the job.** Mechanical passes (results, gear, consumables) run on Sonnet. Passes
  that judge mechanics or balance (rotations, mechanics, rank tuning, fairness) run on Opus. The
  run session itself is Sonnet.
