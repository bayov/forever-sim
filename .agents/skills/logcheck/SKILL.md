---
name: logcheck
description: Audit the sim against the user's WoW Forever combat logs and BayovCore character history. Use it when the user runs /logcheck, says "audit the logs", "check my new logs", "check the sim against my play", comes back from a play session, or asks whether the sim still matches the game.
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
- `-flavor <folder>` picks the game folder. The default is `_classic_beta_`. At release, add
  the live game's folder to `FLAVORS` in `tools/combat_log/archive.sh` and audit it with
  `-flavor`.
- `-iters` is the number of sim replays per mob level (default 200).

## 2. Triage each flag

Go through the flags that aren't `known`, the new ones first. Read the flag's section of the
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
- Commit what changed with `/cm` (memory: commit without asking).

## 4. Tell the user

Keep it short:

- how many windows were checked, how many were skipped as too small, and what's waiting
- each new or still unexplained flag: what it is, our verdict, and what we did or propose
- any sim change that needs the user's go

When the user asks for a written report for others, publish it as an artifact.

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

## Known limits

- The logs from 2026-10-10 to 10-12 came before the history. Their windows use the pasted
  exports, and most of them have `change` flags. They're in the ledger as a baseline.
- The history hasn't been seen in a real game session yet (as of 2026-10-14). On the first
  audit with real history, check that its records look right: a login record at each login,
  change records for a weapon swap and skill ups, a logout record at each `/reload`. Then
  check that the audit cut the windows at those times.
- The sim plays the window with no buffs, no consumes and full mana. Rows that depend on
  those (like a Strength buff from another player) can flag. The report lists other players'
  buffs.
