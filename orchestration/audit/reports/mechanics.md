# Mechanics pass 39 (2026-10-26T00:51:53Z, master 74b3a5f7c)

Range audited: 016daa654..74b3a5f7c (3 non-merge sim commits, all user rulings of 2026-10-25). Result: quiet.

## Get current

| Source | State |
|---|---|
| Client build | 1.60.1.70532, no build answered up to +300 |
| Dev notes thread t/2360696 | newest post /6 (10-16 00:15, edited 10-16 20:50), unchanged |
| Blue tracker | nothing newer than 10-23 (beta ended; Linxy posts are retail Midnight) |
| Daily update | no forever-update merges, no TASKS.md / ledger changes in range |
| Server hotfixes | not re-read (tools/audit/mechanics/hotfixes.py is not in the repo) |

## Commits

| Commit | Change | Check |
|---|---|---|
| 107874b75 | Warrior white-hit Rage level slope s 0.5 -> 0 (flat k) | user ruling, standards 21; only user is rage.go:76; verdicts row already updated |
| 01afcb1d3 | Hunter crit per Agility from Forever's PlayerExpectedStat by level (Classic's 0.0189 at 60) | all 60 table values match 70532 PlayerExpectedStat ClassID 3 CritPerAgility x100; at 60 no stat dependency is added; no other users |
| 92484f3c8 | Dwarf / Night Elf / Gnome / High Order base stat offsets to Forever's level 1 sheets | user ruling; each offset sums to 0 with Spirit 0; Skyborne lower levels built from Undead rows like Windshaper; verdicts row added |

## Conclusions

No shared mechanic moved away from tier 1 data. Open: mechanics-20261025-1 (Spellstone school), unchanged on master.
The race offsets at 60 and the reverted wowsims/forever offsets (6fc22306b1) stay open in notes.md.
