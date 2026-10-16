---
name: log-audit
description: Audit the sim against the user's WoW Forever combat logs and BayovCore character history, now and then every hour in this session. Use it when the user runs /log-audit (or "/log-audit stop"), says "audit the logs", "check my new logs", "check the sim against my play", comes back from a play session, or asks whether the sim still matches the game.
---

# Audit the sim against our combat logs

As the user plays Forever (the beta now, the live game after release), the client writes combat
logs and BayovCore records the character's gear, talents, level and skills as they change. This
skill checks every new stretch of play against the sim and finds where the sim and the game
disagree. Each run rechecks everything we've settled so far on new data.

The tools do the counting. Our job is the triage: for each flag, decide whether the sim is
wrong, logcheck built the wrong character, or it's noise, and then record the answer so the
next run doesn't ask again.

Run everything from `/home/bayov/projects/forever-sim`.

## Modes

The user keeps a Claude Code session open for this, and the skill keeps auditing while it's
open:

- **`/log-audit`**: run the audit now, go through the findings with the user (sections 1 to 4),
  then start the hourly runs (section 5) unless they're already scheduled.
- **An hourly run** (the scheduled prompt `/log-audit hourly`): run the audit, triage only the
  new windows, and collect what we find in `tools/logcheck/findings.md` (section 6). The user
  may be away, so it changes nothing else.
- **`/log-audit stop`**: delete the hourly job (CronList, then CronDelete) and say so.

## 1. Run the audit

```sh
go run --tags=with_db ./tools/logcheck -audit
```

It takes a few seconds. It copies the game's logs and history into `~/forever-logs` first
(`tools/combat_log/archive.sh`). Then it splits every log into windows where the character holds
still, runs logcheck on each window it hasn't run before, and prints a digest:

- **Each new window**: the log, the character, the times, why the window ends, the export the
  sim's character came from, the report file, and the window's flags. A flag is marked `NEW`
  when no earlier window had that ability and kind flagged, and `known` when
  `tools/logcheck/known.json` has it.
- **Waiting**: windows the history doesn't cover yet. WoW saves the history only at logout and
  `/reload`, so a window from the session the user is still playing waits. Tell the user to
  `/reload` if they want it checked now.
- **Flags across the ledger**: each ability and kind, by how many windows flagged it. A flag
  that comes back in window after window is the strong signal. One flag in one window can be
  chance, because each window has dozens of rows at a 1 in 1000 cutoff.

The full report of a window is in `~/forever-logs/audit/reports/`. The ledger of every window
we've run, with its flags, is `~/forever-logs/audit/ledger.jsonl`. Both live in the archive, not
in the repo. `tools/logcheck/README.md` explains every row of a report. Read its "Reading the
report" section before the first triage of a session.

Options:
- `-redo` runs every window again with the sim as it is now. Use it after a sim or logcheck
  fix, to see which flags the fix cleared. The digest counts the newest run of each window.
- `-log <path>` audits one log.
- `-flavor <folder>` audits only one game folder of the archive, like `_classic_beta_`. By
  default the audit takes every folder in `~/forever-logs/`.
- `-iters` is the number of sim replays per mob level (default 200).

## 2. Go through the findings and the new flags

Read `tools/logcheck/findings.md` first. The hourly runs fill it in while the user plays. Go
through its open findings with the user, the "Fix in the sim" ones first. For each one, look at
the reports it names, then do what we agree on: fix the sim or logcheck, add a Need to Verify
item, add it to `known.json`, or tell the user what to do in game to settle it. Then move it to
"Done" with what we did and the date.

Then triage the flags of the windows we just ran that `findings.md` doesn't cover yet. Go
through the flags that aren't `known`, the new ones first. Read the flag's section of the
report, not just the digest line. Work in this order, because the character's flags explain
the ability flags below them:

1. **`sheet`**: the export's sheet doesn't match the sim's character. A buff that was up at
   export time explains a sheet that's higher. Anything else is a gap in how logcheck builds
   the character (a missing enchant, an item the sim doesn't have), and every damage row in
   the window depends on it. Fix logcheck (`export.go`) or the item data first.
2. **`change`**: the character changed in the window in a way the sim can't follow. Only
   windows from before BayovCore's history (before 2026-10-14) should have these, because the
   audit cuts windows at every change the history records. The digest says the window's other
   flags may only show that change, and leaves them out of the counts across the ledger. In a
   window with history, a `change` flag is an audit bug.
3. **`skill`**: we swung a weapon below full skill. Melee outcome rows (miss, dodge, parry,
   glance, crit) in that window depend on how the sim handles low skill, which is still a Need
   to Verify item. Damage rows don't.
4. **Ability flags**: an outcome rate (`hit`, `crit`, `glance`, `miss`, `dodge`, `parry`,
   `block`, `immune`...), `raw` damage, `partial` resists, `procs` per 100 swings, `amount`
   rounding, or `never` (the sim never did it).

For each flag, decide which of these it is:

- **logcheck built the wrong character or fight.** Wrong export, missing enchant or imbue, a
  buff from another player, a mob that's immune to a school (Thundering Boulderkin is immune to
  Nature), an ability the sim keys under another spell ID. Fix logcheck when we can. When we
  can't, add it to `known.json` with the reason.
- **The sim is wrong.** Check `notes.md` "Need to Verify" and the "Recheck with every new log"
  list in `beta_handoff.md` first, since the flag may be an open question or a settled answer
  breaking. Then dig with the scripts in `tools/combat_log/` (`recheck.py`, `glance.py`,
  `ftt.py`, `crit.py`...) to pin the mechanic down. Tell the user what the log shows and what
  the sim does, with the counts. Don't change a sim mechanic without the user's go, because
  every beta answer is a hypothesis (memory: recheck settled beta answers). A change that only
  makes the sim match a client value we already trust is fine to make.
- **Noise.** A p just under 0.001 in one window, on a row with few attempts, that no other
  window has. Leave it out of `known.json`, so it shows again if it comes back.

## 3. Record what we found

- **`tools/logcheck/known.json`** holds the flags we've explained. One entry per ability and
  kind, optionally limited to some mob levels:

  ```json
  [
    {"kind": "immune", "ability": "Lightning Shield (26366)",
     "why": "mobs immune to Nature (Thundering Boulderkin), the sim's mob isn't"}
  ]
  ```

  `kind` and `ability` match the finding exactly as the ledger has them (the digest's table
  prints them as `ability: kind`). Character flags have no ability. `why` says what we found
  and where it's tracked: a Need to Verify item, a sim fix, or a limit of logcheck. Remove an
  entry when its reason no longer holds (the sim was fixed, the question was settled).
- **`notes.md` "Need to Verify"**: a mechanic the logs put in doubt goes there, under its
  class, with the counts and the log window (memory: need-to-verify-list). Mark it "Beta" when
  we can test it at the beta's level cap.
- **`beta_handoff.md` "Recheck with every new log"**: when a settled answer breaks, tell the
  user and update its line.
- **`tools/logcheck/findings.md`**: what's still open (section 6 has its format).
- Commit what changed with `/cm` (memory: commit without asking).

## 4. Tell the user

Keep it short:

- what the hourly runs found since we last looked, and what we did about each finding
- how many windows were checked, how many were skipped as too small, and what's waiting
- each new or still unexplained flag: what it is, our verdict, and what we did or propose
- any sim change that needs the user's go

When the user asks for a written report for others, publish it as an artifact.

## 5. Start the hourly runs

Call CronList. When no job has the prompt `/log-audit hourly`, create one with CronCreate:
recurring, prompt `/log-audit hourly`, at a minute that isn't 0 or 30 (like `17 * * * *`).
Tell the user when the next run is (the time CronCreate gives back), that the runs stop when
they close this session, and that a recurring job expires after 7 days.

## 6. An hourly run

The user may be away, and other sessions share the working copy. So an hourly run:

- doesn't change the sim, logcheck, `known.json`, `notes.md` or `beta_handoff.md`
- writes only `tools/logcheck/findings.md`, and commits only that file:
  `jj commit -m "Log audit findings" tools/logcheck/findings.md`, then
  `jj bookmark set master -r @-` and `jj git push --remote origin --bookmark master`, with the
  session's attribution trailer in the message
- stays cheap: most flags take one look at the report. Dig with logcheck on a single window or
  the `tools/combat_log/` scripts only when a flag looks like a real sim error. A compressed log
  needs `zcat` into a scratch file before a Python script reads it.

Run the audit (section 1). When it ran no new windows, say so in one line and stop. Otherwise,
for each new window, read the report of each flag that isn't `known`, triage it as in section 2,
and update `findings.md`:

- When an open finding already covers the flag (same ability, kind and cause), add the window
  to its list and update its counts. Don't add a duplicate.
- Otherwise add a new finding in the right section.
- A flag that looks like plain noise (one window, p barely under 0.001, few attempts) goes
  under "Watching", so we see if it comes back.

`findings.md` has these sections, in this order:

```markdown
# Log audit findings

## Fix in the sim
## Collect more data
## logcheck issues
## Watching
## Done
```

Each finding is a `###` heading with a short name, then:

- **Flag**: the ability and kind as the ledger has them (like `Melee: parry`), and the mob levels.
- **Windows**: each window as `log from-to`, with the log's and the sim's numbers.
- **What we think**: the cause, and what in the sim or in logcheck it points at (the file and
  function when we know them). Name the Need to Verify item or recheck list line it touches,
  from `notes.md` or `beta_handoff.md`.
- **Next step**: the sim change to propose, the test that would settle it (what to do in game,
  how many swings or casts, against what), or the logcheck fix.
- **Status**: `open`, the date we first saw it and the date of the last update.

Never remove a finding in an hourly run. Only a `/log-audit` run with the user moves one to
"Done". End with a few lines: the windows run, and each finding added or updated.

## How the windows are cut

The audit cuts each log at:

- each gear, talent or level change in the history (skill ups don't cut, because a low skill
  goes up every few swings, and the report warns about them instead)
- each change of the item level on our casts, when the history has no change there (logs from
  before the history, or gear swaps the history missed)
- each logout and `/reload`
- each break of more than 30 minutes between fights

A window with fewer than 60 of our attacks on mobs is listed as skipped. A character with no
history at all waits until the log has been quiet for 30 minutes. `tools/logcheck/audit.go`
has the details.

Once a log hasn't changed for a day and no window of it waits, the audit compresses it to
`WoWCombatLog-*.txt.gz` (about a tenth of the size). logcheck reads those the same way, the
ledger keeps the plain name, and `archive.sh` stops copying the game's plain copy.

## Characters and game folders

The user plays more than one character and will keep making new ones, so nothing here may
assume a character's name or realm. The tools know each character by its GUID
(`Player-4620-01872808`). The digest and reports print `Name-Realm (GUID)`, but the name is
only for reading. The game changes how the log writes names (`Bayov-ClassicBetaPvE2-` became
`Bayov-Oom-ClassicBetaPvE2-` for the same character on 2026-10-16), and two characters can
share a name. A log says which character was ours by the "mine" unit flags, and BayovCore
writes the GUID into each history record since 2026-10-16. `tools/logcheck/README.md`
("The character") has the details. When a window or a script can't find us, look at the
flags and GUIDs in the log, not at the names.

The game folder will change too. The beta's logs and history are in `_classic_beta_`. The
release client gets a folder of its own, and we don't know its name yet (`_forever_` is our
guess). `tools/combat_log/archive.sh` lists the folders it copies in `FLAVORS`, and
`history.py` takes the same list. When the release is out, or when the audit has no new
windows while the user says they played, check the folders in the WoW folder
(`/mnt/d/Gaming/World of Warcraft`) and put the real one in `FLAVORS`. logcheck reads every
folder in the archive, so it needs no change.

## Known limits

- The logs from 2026-10-10 to 10-12 came before the history. Their windows use the pasted
  exports, and most of them have `change` flags. They're in the ledger as a baseline.
- The first real sessions with history (2026-10-15 and 10-16) cut the windows where they
  should: at each level up, each gear pickup and each logout or `/reload`. A new character
  changes gear every few minutes while leveling, so its first windows are mostly too small
  to run.
- The sim plays the window with no buffs, no consumes and full mana. Rows that depend on
  those (like a Strength buff from another player) can flag. The report lists other players'
  buffs.
