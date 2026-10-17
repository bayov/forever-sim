# Results audit, pass 44 (master 1e2f7ec5b)

Full pass on 20 watched files (db regen, proto, spell_ranks.go, unit.go, shaman healing.go, tools/database, enhancement presets.ts).

| Check | Result |
|---|---|
| 66 requests, 20k, seed 1 (shared run) vs published | 0 of 66 off by more than 0.15 |
| vs reference req_summary.txt | 0 moved |
| drift_report.txt (mksettings rebuilt at 1e2f7ec5b) vs reference | identical |
| unread fields | only `notes` |

Why nothing moved: the changes are leveling and healing plumbing (TopHealRankID, rotationsNameTopRanks in the Enhancement leveling plan, core.TopRankID), the regenerated item db and the client-table generators; none reach a level 60 DPS request. No findings. Ranking explanation unchanged from pass 40.
