# Mechanics pass 44 (2026-10-28, master 5d27b0015)

Range audited: 37a8a8ccd..5d27b0015. One sim commit: ae8522c14 (racials, via the PR #117 merge 297226d14). TASKS.md / ledger commit 5d27b0015 records post 8 landing.

## Get current

- Client: newest 1.60 build is still 1.60.1.70634 (`wago.py latest --probe 300`). No verdict to re-check.
- Dev notes thread 2360696: still 8 posts; post 7 (edited 2026-10-27 16:47:14) and post 8 (edited 16:46:55) unchanged since pass 43.
- Blue tracker since 2026-10-27 15:37: Forever category posts are a name-reservation issue (Oceanic) and the Portal Room announcement; nothing mechanical. Other posts are TBC / Midnight hotfixes.
- Server hotfixes: `tools/audit/mechanics/hotfixes.py` is still missing from the tree; not re-read.

## Checks (ae8522c14)

| Item | Master now | Tier 1 (post 8) | Verdict |
|---|---|---|---|
| Blood Fury CD | 2 min (racials_forever.go) | 2 min | Matches |
| Berserking duration | 15 sec when `IsForever()`, 10 sec otherwise; CD 3 min, flat 10% on Forever | 15 sec | Matches |
| Touch of the Grave ICD | 15 sec | 15 sec | Matches |

Knock-on: `git grep` finds no other user of these durations outside sim/core/racials*.go (hunter summon_hawk.go only mentions Berserking haste in a comment). Goldens moved only on Orc, Troll and Undead characters, as the commit says.

## Findings

- mechanics-20261027-3 resolved (ae8522c14).
- mechanics-20261027-2 (Amplify Curse vs Bane of Agony on 70630) stays open; no warlock commit in range.
- No new findings.
