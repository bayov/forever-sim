# Fairness pass 4 (2026-10-25, master 54878fa)

Result: findings. One new, low: fairness-20261025-1 (Marksmanship's report race). Full pass.

## Why it ran

382 watched files changed since daba9d1 (104 commits): the 2026-10-23 alignment round with wowsims/forever (sim/core, standards 4/10/11/21), every class's regear and rotation round on the 8% physical hit cap, standards 26 (must-take talents) and report version 36 (regenerated at e8a8b50d).

## What I checked

- Gear: `audit1b.py`, all 19 sets "5a PASS | 1b PASS".
- Race sweep: 36 base requests and variants, every race, 5000 iterations, seed 1 (`sweep.py` on the 54878fa tree). Baseline refreshed (`ref/fairness/`, sha 54878fa).
- Published against current: all within 1.9 here; the shared 20k run says 0 of 64 off by more than 0.5. So the report matches its requests. Ret +1.9 and Frostfire +1.6 are 5000-iteration noise.
- Best race: Marksmanship Human beats Windshaper by 2.5 (20k, paired seeds 1 and 1000003; duty mark 2.6), so I filed fairness-20261025-1. Lone Wolf stays Windshaper (+1.0 at 20k). Other near-ties (rogue Combat Swords and Mutilate Orc +0.5 / +0.3, Seal Fate Swords tie, Subtlety +0.1, Balance Wrath +0.3, Enh Water Shield tie) are noise.
- Consumables across the 45 base requests: every build has Engineering and takes the Goblin Sapper except Feral. That one is correct: the sapper's row is not usable shapeshifted (sim/core/consumes.go registerExplosivesCD), and adding it measured 0.0 (20k, two seeds, base and FF duty). The filler explosives and potions differ by spec per standards 11 (by sim). results-20261025-1 (mksettings vs requests consumes) is the results auditor's; I didn't duplicate it.
- Movement: every class moved, all attributed to its regear round, the alignment round or standards 26. Demonology -112.9 is Suppression 5/5 (0b566e87e measured 68.1) plus the regear. Hunters gained the most (BM +201, MM +125, Survival +122). Top over bottom is now 1.64x (was 1.79x).

## Ranking (5000 iterations, seed 1)

| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Gap | Before (1e2f7ec) |
|---|---|---|---|---|---|---|---|---|
| 1 | Beast Mastery | Human | 1172.1 | Orc 1168.2 / Human | Human (1172.1) | 1172.0 | +0.1 | 971.0 (+20.7%) |
| 2 | Marksmanship | Human | 1169.4 | Windshaper 1168.0 / Human | Windshaper (1168.0) | 1167.6 | +0.4 | 1044.8 (+11.9%) |
| 3 | Lone Wolf (hunter) | Windshaper | 1159.7 | Windshaper / Human 1158.3 | Windshaper (1159.7) | 1159.3 | +0.4 | 1094.6 (+5.9%) |
| 4 | Fire | Human | 1100.8 | Orc 1096.2 / Human | Human (1100.8) | 1101.0 | -0.2 | 1067.5 (+3.1%) |
| 5 | Frostfire | Human | 1098.0 | Orc 1085.8 / Human | Human (1098.0) | 1096.4 | +1.6 | 1083.7 (+1.3%) |
| 6 | Fury (dual wield) | Orc | 1069.1 | Orc / Human 1064.9 | Orc (1069.1) | 1068.7 | +0.4 | 989.7 (+8.0%) |
| 7 | Demonology | Human | 1064.9 | Orc 1062.1 / Human | Human (1064.9) | 1065.1 | -0.2 | 1177.8 (-9.6%) |
| 8 | Survival | Human | 1057.1 | Orc 1056.1 / Human | Human (1057.1) | 1056.6 | +0.5 | 934.9 (+13.1%) |
| 9 | Affliction | Human | 1036.4 | Orc 1033.4 / Human | Human (1036.4) | 1035.3 | +1.1 | 1066.4 (-2.8%) |
| 10 | Fury (two-hand) | Orc | 1034.5 | Orc / Human 1030.9 | Orc (1034.5) | 1033.7 | +0.8 | 971.0 (+6.5%) |
| 11 | Destruction | Human | 1012.2 | Orc 1008.9 / Human | Human (1012.2) | 1011.8 | +0.4 | 1048.1 (-3.4%) |
| 12 | Destruction Fire | Human | 987.5 | Orc 987.1 / Human | Human (987.5) | 987.4 | +0.1 | 1019.7 (-3.2%) |
| 13 | Arcane | Human | 983.2 | Orc 983.0 / Human | Human (983.2) | 982.5 | +0.7 | 964.8 (+1.9%) |
| 14 | Shadow | Human | 961.7 | Troll 954.5 / Human | Human (961.7) | 961.8 | -0.1 | 960.6 (+0.1%) |
| 15 | Arms | Orc | 956.5 | Orc / Human 955.4 | Orc (956.5) | 957.4 | -0.9 | 976.8 (-2.1%) |
| 16 | Retribution | Human | 947.0 | Undead 940.1 / Human | Human (947.0) | 945.1 | +1.9 | 889.1 (+6.5%) |
| 17 | Frost | High Order | 946.8 | Orc 944.8 / High Order | High Order (946.8) | 946.6 | +0.2 | 844.4 (+12.1%) |
| 18 | Enhancement | Orc | 887.7 | Orc / Dwarf 878.9 | Orc (887.7) | 888.1 | -0.4 | 892.0 (-0.5%) |
| 19 | Feral | Night Elf | 868.5 | Tauren 862.7 / Night Elf | Night Elf (868.5) | 868.1 | +0.4 | 752.9 (+15.4%) |
| 20 | Balance | Tauren | 856.7 | Tauren / Night Elf 854.1 | Tauren (856.7) | 856.3 | +0.4 | 873.1 (-1.9%) |
| 21 | Elemental | Windshaper | 777.1 | Windshaper / Dwarf 773.0 | Windshaper (777.1) | 777.4 | -0.3 | 772.0 (+0.7%) |
| 22 | Rogue Combat Daggers | Windshaper | 770.5 | Windshaper / High Order 769.9 | Windshaper (770.5) | 770.3 | +0.2 | 725.5 (+6.2%) |
| 23 | Smite | Undead | 765.5 | Undead / Gnome 745.6 | Undead (765.5) | 766.1 | -0.6 | 765.8 (-0.0%) |
| 24 | Rogue Combat Swords | Orc | 764.4 | Orc / High Order 763.7 | Windshaper (763.9) | 764.0 | -0.1 | 726.5 (+5.2%) |
| 25 | Rogue Seal Fate Daggers | Orc | 732.1 | Orc / Human 731.7 | Orc (732.1) | 731.5 | +0.6 | 687.2 (+6.5%) |
| 26 | Rogue Seal Fate Swords | Human | 729.4 | Orc 729.4 / Human | Orc (729.4) | 728.9 | +0.5 | 677.8 (+7.6%) |
| 27 | Rogue Mutilate | Orc | 719.4 | Orc / Human 718.7 | Windshaper (719.1) | 718.9 | +0.2 | 683.9 (+5.2%) |
| 28 | Rogue Subtlety | Orc | 715.8 | Orc / Night Elf 715.7 | Night Elf (715.7) | 715.9 | -0.2 | 656.7 (+9.0%) |

| Variant | Best race | DPS | Report's published (at its race) | Before |
|---|---|---|---|---|
| druid__balance-wrath | Windshaper | 814.0 | 813.3 (at Tauren 813.7) | 826.2 |
| paladin__level_60_shock | Human | 869.4 | 869.0 (at Human 869.4) | 901.0 |
| paladin__level_60_sub_holy | Human | 949.8 | 949.1 (at Human 949.8) | 883.6 |
| priest__shadow-holy | Human | 918.9 | 918.9 (at Human 918.9) | 917.1 |
| shaman__elemental-level60-mana-tide | Windshaper | 762.2 | 762.1 (at Windshaper 762.2) | 761.3 |
| shaman__enhancement-imp-gw | Orc | 883.8 | 884.3 (at Orc 883.8) | 887.9 |
| shaman__enhancement-mana-tide | Orc | 798.9 | 800.0 (at Orc 798.9) | 802.5 |
| shaman__enhancement-water-shield | Windshaper | 839.7 | 839.9 (at Orc 839.7) | 842.2 |

Top over bottom: 1.64x

## Notes for the next pass

- `druid__feral-cat.meta.json` has race "Night Elf" (with a space) where every other meta has "NightElf". `table.py` crashes on it; I ran a scratch copy that strips spaces. The tool needs that one-line fix (`m["race"].replace(" ", "")`), or the meta needs normalising.
- Open items carried: hunter end-of-fight test, Ice Lance Low range, FTT + Windfury stacking, boss partial resist log, Rage k and pet scaling.
