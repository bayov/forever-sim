# Results audit, pass 69 (master 5a6e9c6a6)

Shared run (77 requests, 20k, seed 1) against builds.json: 47 of 77 within 0.15. The 30 that differ are all stale report data
(the October 30 hotfix re-searches changed presets; the DPS report is not re-simmed yet). Hunters now match builds.json (11de3318a).

| Builds | builds.json -> run | cause |
|---|---|---|
| Druid Balance, Wrath, Feral (6, with FF duties) | 846.6 -> 891.1 (+44), 806.5 -> 855.7, 885.6 -> 942.5 (+57) | 175e3fdb3 hotfix tree, 4176687d3 re-search |
| Mage Arcane, Frost, Frostfire | 910.9 -> 939.0, 907.8 -> 944.9, 1046.3 -> 1076.6 | 5a6e9c6a6 mage re-search on the Oct 30 hotfix and Arcane Blast change |
| Paladin Ret twist / sub-Holy (6) | 974.5 -> 1039.3, 977.9 -> 1038.1 (+60 to +65) | 306f78678: Unsealed Wrath twisting; matches the commit message numbers |
| Paladin sub-Holy Undead | 1069.6 -> 1151.4 | same |
| Paladin Ret Shock (shock_ret 960.6 -> 974.9; deep Holy shock 908.4 -> 855.6; shock_undead 995.1 -> 973.5) | | same; run equals the commit's numbers to 0.1 |
| Warrior Arms, Arms dragonkin, Fury 2H and duty runs (7) | +11 to +12 | 5b2eec6de two-hand white crit doubles Rage; Fury dual wield unmoved, as expected |

Checks: the 30 gaps equal the paladin/druid/mage/warrior commit messages. Drift report equals the reference (no mksettings or
request-shape drift). Unread fields: none. Metas: 62 of 77 off published (results-20261101-1, updated).

Ranking notes (for the republish): Ret twist/sub-Holy (~1039) rise above Fire and Marksmanship; deep-Holy Ret Shock is 17.7% behind
twist (855.6) by the author's own measurement; the hunter pet builds fell 4 to 8% with the pet swing change, now in builds.json.
Not re-flagged: pending re-sim. Next pass should compare the republished builds.json against this run.
