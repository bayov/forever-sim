# Results audit, pass 68 (master c23ef0440)

Used the shared run (77 requests, 20k, seed 1) against builds.json. 66 of 77 within 0.15.

| Build | builds.json | run | gap | cause |
|---|---|---|---|---|
| mage__frostfire | 1046.3 | 1076.6 | +30.2 | 98e819d39 Frost hotfix values, not yet re-simmed |
| mage__frost | 907.8 | 929.3 | +21.5 | same |
| mage__arcane | 910.9 | 893.6 | -17.3 | f619fb648 Arcane Blast 12% Int / 100% mana / 18 s, not yet re-simmed |
| warrior arms, arms_dragonkin, fury_2h and their duty runs (6) | | | +11 to +12 | 5b2eec6de crit Rage fix; eaced12a8 re-simmed them (Arms 1008.7, Fury 2H 1038.9, matching the run) but the c23ef0440 merge held the report data, so builds.json lacks it |

The 9 pet hunters now match builds.json (11de3318a). Warlock Improved Drains change moves no build beyond 0.15.
All gaps are stale report data waiting on a re-sim or republish, not bugs. The open metas finding (results-20261101-1) stands.
Drift and unread-field checks were not rerun (no mksettings or request-shape change).
