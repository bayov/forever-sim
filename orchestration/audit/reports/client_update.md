# Client update pass: 1.60.1.70634 -> 1.60.1.70658

- Pass 4, 2026-10-28, master 224b8a4e00994c8ede45e50bdf5172c98b3a7592. `wago.py latest --probe 300` returns 1.60.1.70658.
- For `.agents/skills/dps-report/state.json`: `foreverClientBuild` → 1.60.1.70658 (the orchestrator updates it).

## Client diff

`table_summary.py 70634 70658`: all 39 tables in the `client_update` set identical, no added rows (foreverchanges: only 2 of 137 tables differ, both creature models). No class, racial, talent, enchant or item change in the client.

## Launch-realm hotfix (foreverchanges, 2026-10-28)

FC read a hotfix from the launch realms' cache that Blizzard's notes don't mention. FC tier only: the cloud has no client, so `tools/sources/dbcache.py` can't confirm it.

| Change | L60 DPS | Sim | Effect if taken (180 s, 20k it., seeds 1 / 1000003) | Ledger | Finding |
|---|---|---|---|---|---|
| Aspect of the Hawk r1-5 15/22/30/38/47 (was 20/35/50/70/90); best at 60 becomes r6, 55 | yes | 90 (r5) | BM −0.66%, LW −0.90%, MM −0.78%, SV 0 | HUN-HF-01 | 20261028-1 (medium, verify) |
| Trueshot Aura r1-4 15/22/30/40 (was 30/40/50/75); r5 stays 50 in data | yes | 100 by ruling hunter_tsa_hm_rank | at 50: BM −0.94%, LW −1.28%, MM −1.12%, SV −0.09% | HUN-HF-02 | 20261028-2 (high, user ruling) |
| Blood Fury 2 min in server data | yes | landed (#117) | 0 | SHR-P8-01 | none |
| Talent-root shared DR group; Righteous Fury at 10; zombie-plague event spells | n/a | n/a | n/a | paragraph only | none |

Raw tables: `$AUDIT_WORK/client_update/hawk.md`, `tsa.md` (scratch). Sizing builds patched `sim/hunter/aspects.go` and `sim/core/buffs.go` in a scratch copy of the master tree; nothing in the repo's code changed.

## Sources

| Source | State |
|---|---|
| Blizzard dev notes (topic 2360696) | Not read: forum rate limit again. Next pass: diff against `sources/devnotes_post7_2026-10-26.txt` (post /8 of 10-27 is in the ledger as SHR-P8-01) |
| foreverchanges.pro | Lists 70658 as current (10-28, model files only, no Blizzard notes) and the 10-28 hotfix above |
| Wowhead | Not checked (blocked) |

## Ledger

Added `## Build 1.60.1.70658 (first seen 2026-10-28)`: one paragraph plus rows HUN-HF-01 and HUN-HF-02 (both ambiguous, needs verify); Summary counts (Hunter 84 rows, 3 ambiguous; total 742); Rulings needed item 7 (Trueshot Aura rank 5).

Open findings: client_update-20261028-1, -2 (new); client_update-20261027-3 (Amplify Curse, parked for an in-game test) unchanged.
