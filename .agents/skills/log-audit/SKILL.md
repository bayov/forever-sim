---
name: log-audit
description: Audit the sim against the user's WoW Forever combat logs and BayovCore history, now and then hourly. Use on /log-audit (or "/log-audit stop"), "audit the logs", "check my new logs", after a play session, or when asked whether the sim still matches the game.
---

# Audit the sim against our combat logs

The audit checks each new stretch of the user's Forever play (combat logs plus BayovCore's
character history) against the sim. We triage each flag and record the answer so the next
run doesn't ask again.

Run everything from `/home/bayov/projects/forever-sim`.

## Modes

- **`/log-audit`**: run the audit, go through the findings with the user (sections 1 to 4),
  then start the hourly runs (section 5) unless already scheduled.
- **`/log-audit hourly`** (the scheduled prompt): section 6. Changes only
  `tools/logcheck/findings.md`.
- **`/log-audit daily`**: like `/log-audit`, but schedule the unattended runs daily (cron like
  `13 9 * * *`, still prompt `/log-audit hourly`), replacing any hourly job.
- **`/log-audit stop`**: delete the job (CronList, then CronDelete) and say so.

## 1. Run the audit

```sh
go run --tags=with_db ./tools/logcheck -audit
```

It first copies the game's logs and history into `~/forever-logs` (`tools/combat_log/archive.sh`;
nothing else runs the archive). Then it splits logs into windows where the character holds
still, runs logcheck on each new window, and prints a digest:

- **Each new window**, with its flags: `NEW` (no earlier window flagged that ability and
  kind) or `known` (in `tools/logcheck/known.json`).
- **Waiting**: windows the history doesn't cover yet. WoW saves history only at logout and
  `/reload`, so tell the user to `/reload` to check the current session.
- **Flags across the ledger**, by how many windows flagged each. Recurring flags are the
  strong signal; one flag in one window can be chance.

Reports are in `~/forever-logs/audit/reports/`, the ledger in
`~/forever-logs/audit/ledger.jsonl`. Read the "Reading the report" section of
`tools/logcheck/README.md` before the first triage of a session.

Options: `-redo` reruns every window on the current sim (after a fix, to see what it cleared),
`-log <path>` audits one log, `-flavor <folder>` one archive folder (like `_classic_beta_`),
`-iters` sets replays per mob level (default 200).

## 2. Go through the findings and the new flags

Read `tools/logcheck/findings.md` first. Go through its open findings with the user, "Fix in
the sim" first: read the reports each names, then do what we agree on (fix the sim or logcheck,
add a Need to Verify item, add to `known.json`, or tell the user what to do in game). Move it
to "Done" with what we did and the date.

Then triage the flags of the windows just run that aren't `known` and `findings.md` doesn't
cover, new ones first. Read the flag's section of the report, not just the digest line. Go in
this order, since character flags explain the ability flags:

1. **`sheet`**: the export's sheet doesn't match the sim's character. A buff at export time
   explains a higher sheet. Otherwise logcheck built the character wrong (missing enchant, item
   the sim lacks), which skews every damage row: fix `export.go` or the item data first.
2. **`change`**: the character changed mid-window. Expected only in windows from before
   BayovCore's history (before 2026-10-14); the digest leaves such windows out of the ledger
   counts. In a window with history, it's an audit bug.
3. **`skill`**: a weapon swung below full skill. That window's melee outcome rows (miss,
   dodge, parry, glance, crit) depend on the sim's low-skill handling, still a Need to Verify
   item. Damage rows don't.
4. **Ability flags**: an outcome rate (`hit`, `crit`, `glance`, `miss`, `dodge`, `parry`,
   `block`, `immune`...), `raw` damage, `partial` resists, `procs` per 100 swings, `amount`
   rounding, or `never` (the sim never did it).

Decide which each flag is:

- **logcheck built the wrong character or fight**: wrong export, missing enchant or imbue, a
  buff from another player (the sim runs with no buffs, no consumes and full mana; the report
  lists other players' buffs), a mob immune to a school (Thundering Boulderkin: Nature), an
  ability the sim keys under another spell ID. Fix logcheck when we can, else add to
  `known.json` with the reason.
- **The sim is wrong.** First check `notes.md` "Need to Verify" and `beta_handoff.md` "Recheck
  with every new log": it may be an open question or a settled answer breaking. Pin it down
  with `tools/combat_log/` scripts (`recheck.py`, `glance.py`, `ftt.py`, `crit.py`...). Tell
  the user what the log shows and what the sim does, with counts. Don't change a sim mechanic
  without the user's go, since every beta answer is a hypothesis; matching a client value we
  already trust is fine.
- **Noise**: p just under 0.001 in one window, few attempts, no other window has it. Keep it
  out of `known.json` so it shows again if it recurs.

Never assume a character's name, realm or game folder. Read `reference.md` when a window or
script can't find our character, a window is skipped or waits unexpectedly, or the audit finds
no new windows though the user played.

## 3. Record what we found

- **`tools/logcheck/known.json`**: one entry per explained ability and kind, optionally limited
  to some mob levels:

  ```json
  [
    {"kind": "immune", "ability": "Lightning Shield (26366)",
     "why": "mobs immune to Nature (Thundering Boulderkin), the sim's mob isn't"}
  ]
  ```

  `kind` and `ability` match the ledger exactly (the digest prints `ability: kind`); character
  flags have no ability. `why` says what we found and where it's tracked (Need to Verify item,
  sim fix, logcheck limit). Remove an entry when its reason no longer holds.
- **`notes.md` "Need to Verify"**: a mechanic the logs put in doubt, under its class, with the
  counts and log window. Mark it "Beta" when testable at the beta's level cap.
- **`beta_handoff.md` "Recheck with every new log"**: when a settled answer breaks, tell the
  user and update its line.
- **`tools/logcheck/findings.md`**: what's still open (format in section 6).
- Commit what changed with `/cm`, without asking.

## 4. Tell the user

Briefly: what the hourly runs found since we last looked and what we did about each; windows
checked, skipped as too small, and waiting; each new or unexplained flag with our verdict and
what we did or propose; any sim change that needs their go.

Publish a written report for others as an artifact only when asked.

## 5. Start the hourly runs

Call CronList. When no job has the prompt `/log-audit hourly`, create one with CronCreate:
recurring, prompt `/log-audit hourly`, at a minute other than 0 or 30 (like `17 * * * *`).
Tell the user the next run time (from CronCreate), that the runs stop when this session
closes, and that a recurring job expires after 7 days. Tell them to start `/log-audit` again
within the week: BayovCore keeps only 7 days of history in game and the archive
(`history.py`) is the only full copy, so older records are lost without a run.

## 6. An hourly run

The user may be away and other sessions share the working copy. So an hourly run:

- doesn't change the sim, logcheck, `known.json`, `notes.md` or `beta_handoff.md`
- writes and commits only `tools/logcheck/findings.md`:
  `jj commit -m "Log audit findings" tools/logcheck/findings.md`, then
  `jj bookmark set master -r @-` and `jj git push --remote origin --bookmark master`, with the
  session's attribution trailer in the message
- stays cheap: dig (logcheck on one window, the `tools/combat_log/` scripts) only when a flag
  looks like a real sim error. A compressed log
  (`.txt.gz`) needs `zcat` into a scratch file before a Python script reads it.

Run the audit (section 1). With no new windows, say so in one line and stop. Otherwise, for
each new window, read the report of each flag that isn't `known`, triage it as in section 2,
and update `findings.md`:

- When an open finding covers the flag (same ability, kind and cause), add the window to it and
  update its counts. No duplicates.
- Otherwise add a new finding in the right section.
- Plain noise (one window, p barely under 0.001, few attempts) goes under "Watching".

`findings.md` sections, in order:

```markdown
# Log audit findings

## Fix in the sim
## Collect more data
## logcheck issues
## Watching
## Done
```

Each finding is a `###` heading with a short name, then:

- **Flag**: ability and kind as the ledger has them (like `Melee: parry`), and the mob levels.
- **Windows**: each as `log from-to`, with the log's and the sim's numbers.
- **What we think**: the cause, and what in the sim or logcheck it points at (file and
  function when known), plus the Need to Verify item or recheck list line it touches.
- **Next step**: the sim change to propose, the test that would settle it (what to do in game,
  how many swings or casts, against what), or the logcheck fix.
- **Status**: `open`, the date first seen and the date of the last update.

Never remove a finding in an hourly run; only `/log-audit` with the user moves one to "Done".
End with a few lines: the windows run, and each finding added or updated.
