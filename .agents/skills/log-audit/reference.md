# Log audit reference

## Characters

The tools know each character by GUID (`Player-4620-01872808`). Names in the digest and
reports (`Name-Realm (GUID)`) are only for reading: the log's name format changes and two
characters can share a name. When a window or script can't find us, look at the "mine" unit
flags and GUIDs in the log, not names. BayovCore writes the GUID into history records since
2026-10-16. Details: `tools/logcheck/README.md`, "The character".

## How windows are cut

- At gear, talent and level changes in the history (not skill ups; the report warns about
  those), item level changes on our casts the history missed, each logout and `/reload`, and
  breaks over 30 minutes between fights.
- Windows under 60 of our attacks on mobs are skipped, so a leveling character's first windows
  mostly are. A character with no history waits until its log has been quiet 30 minutes.
- Logs from 2026-10-10 to 10-12 predate the history: they use pasted exports, mostly have
  `change` flags, and stay in the ledger as a baseline.
- A log unchanged for a day with no waiting windows is compressed to `WoWCombatLog-*.txt.gz`;
  logcheck reads it the same and the ledger keeps the plain name.
- Details: `tools/logcheck/audit.go`.

## Game folders

The archive copies the folders listed in `FLAVORS` in `tools/combat_log/archive.sh` (`history.py`
takes the same list). The beta is `_classic_beta_`; the release folder's name is unknown
(`_forever_` is our guess). When the release is out, or the audit finds no new windows though
the user played, check the folders in `/mnt/d/Gaming/World of Warcraft` and put the real one in
`FLAVORS`. logcheck reads every archive folder, so it needs no change.
