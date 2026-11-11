# Fairness pass 35 (2026-11-11, master 707ae92)

Result: quiet (no new finding). Watched files changed since 27d0d59 (61, 40 for this auditor): standards 25 amended,
`sim/core` (natural flasks, weapon imbue + weapon totem, dot expiry), shaman Nov 10 tuning and Enhancement re-optimization,
warlock hotfix 113766, mage Arcane Concentration split, hunter mana costs and hawk, Idol of Ferocity, Sweeping Strikes,
item database for client 70888. Full pass: `audit1b.py` and a full 5k race sweep on the loop's 707ae92 CLI and tree.
New baseline: 707ae92.

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Natural flasks: 30 Stamina, zone bonus 5/4/5/5 -> 2 (client 70862) | 032692409 | The 28 physical presets drop: hunters -21, Feral -18.2, Survival -28.7, rogues -16 to -19, Fury -38.4, Fury 2H -35.6, Arms -17.5. Ret and casters (Supreme Power) unchanged. A game change applied to every preset that carries the flask: fair |
| Weapon imbue and weapon totem both work (Nov 10 hotfix), standards 25 amended | cb651085c, 02dee9e1d | Enhancement places its own Windfury Totem (10614) next to Windfury Weapon, so every melee spec has the totem. No asymmetry |
| Shaman Nov 10 tuning, Enhancement re-optimized and re-simmed | cb651085c, 54a15ba24, f8a37c8e8 | Enhancement -116.8 (1147.5 -> 1030.7), 4th -> 10th; metas match (+0.1). Elemental unchanged |
| Warlock hotfix 113766, Wrack; no dot tick after expiry | 546d0dcaa, 99c8ce2b1, f20d5938a | Affliction +36.9 (11th -> 5th); Destruction Fire +0.0 |
| Arcane Concentration splits across a Missiles channel | 5bfae1163 | Arcane -6.6, Fire -4.0 (binaries fc065aae3 970.2 -> a34d22451 963.6, Arcane, 5k seed 1) |
| Hunter mana costs, hawk; Idol of Ferocity; Sweeping Strikes charges; Bladewhisper's Namesake | 30e67a529, 6feec7094, 333643570, a34d22451, b1291e61a | No separate move beyond the flask in the swept builds |

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Movement against the ca5e35c baseline: every moved build attributed above. Ret, Destruction, Demonology, Frostfire,
  Frost, Balance, Shadow, Smite, Elemental identical. No shared change leaked.
- Published against current: Affliction +37.0; the natural-flask presets -15 to -22 on top of the open Windfury Totem
  gaps (Fury -136.3, Fury 2H -224.8, Arms -217.3, Survival -124.2, Ret -83.2, rogues -31 to -76); Fire -4.1.
  All are stale metas waiting on the re-sim under results-20261109-1 (its 2026-11-11 update lists them). Independent
  confirmation, not refiled. One correction to that update: Arcane -6.6 and Fire -4.3 do not predate 27d0d59; both
  move on 5bfae1163 (27d0d59 1058.2 / 970.2 -> f20d5938a 1053.8 / 963.6, Fire / Arcane). Arcane's meta (962.3) already
  carries it; Fire's (1057.9) does not.
- Best race: the report race is best everywhere within noise (MM Night Elf +0.3, Fire Gnome +0.4, Arms Undead tie).
- fairness-20261109-1 still open: the totem's proc mask (buffs.go) and Raptor Strike are unchanged.

## Ranking now (report race, 5k)

Lone Wolf 1291.8, MM 1232.7, BM 1174.7, Feral 1124.4, Affliction 1071.1, Frostfire 1065.7, Destruction 1063.2,
Demonology 1055.3, Fire 1053.8, Enhancement 1030.7, Ret 1008.9, Fury DW 999.7, Destruction Fire 997.9, Elemental 977.4,
Arcane 963.6, Balance 952.1, Shadow 948.6, Frost 943.6, Fury 2H 926.2, Survival 917.0, Combat Daggers 915.4,
Combat Swords 899.6, SF Swords 881.1, Arms 870.7, SF Daggers 869.6, Subtlety 839.3, Mutilate 827.6, Smite 781.0.
Top over bottom 1.65x. Full table: `orchestration/audit/ref/fairness/table_baseline.md`.

## Open for the next pass

- After the re-sim of results-20261109-1, check that published matches current for every moved spec (now 45 metas).
- The ruling on fairness-20261109-1 (Option A: Survival +25; Option B: Fury -16.6, Fury 2H -3.5, Arms -4.8; measured
  before the flask change, so re-measure when it is ruled).
