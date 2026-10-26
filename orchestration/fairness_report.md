# Fairness report: Level 60 DPS presets (2026-10-12)

From the fairness auditor. The question this answers: is every DPS spec simmed under equally strong and equally representative conditions, so their numbers compare? Six passes over all nine class branches, 09:50 to 15:10 UTC. Final heads (rogue re-swept at edd0ff7 after its Murder fix): warrior 82610f8, paladin e855d22, hunter d4554f1, rogue 6df047e, priest 3da596a, shaman 6bc4225, mage c2a9f7a, warlock 820c83b, druid 672e891.

Method: each class's shipped Level 60 build (gear, talents, APL), built with mksettings and run with wowsimcli. I read code and logs and did not use rotopt. Measurements change one input at a time against the same seed. Two switches existed only in my local worktrees and were never pushed: one turns weapon racials off, one skips set-bonus tiers.

## 1. Cross-spec table (master f62cb79: the seven wowsims/forever mechanics as one-value assumptions, standards 27 and 28 (Multi-Shot 0.5 sec cast, live spell DoT and channel ticks, hunter base RAP 110, failed Bite keeps its energy, casting regen to a paid channel's end); full re-sweep 2026-10-26)

Same method: the report's base requests, every race, 5000 iterations, seed 1. All 19 Level 60 gear sets pass 1b and 5a. "Report's published" is now the deployed `ui/dps_report/data/builds.json`, not the req metas' `published.dps`, which the regeneration doesn't refresh.

| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Gap | Before (74b3a5f) |
|---|---|---|---|---|---|---|---|---|
| 1 | Beast Mastery | Human | 1143.6 | Orc 1139.7 / Human | Human (1143.6) | 1143.7 | -0.1 | 1141.4 (+0.2%) |
| 2 | Marksmanship | Human | 1136.3 | Orc 1133.4 / Human | Human (1136.3) | 1136.1 | +0.2 | 1139.6 (-0.3%) |
| 3 | Lone Wolf (hunter) | High Order | 1123.9 | Windshaper 1123.9 / High Order | Windshaper (1123.9) | 1122.9 | +1.0 | 1127.9 (-0.4%) |
| 4 | Fire | Human | 1100.4 | Orc 1095.6 / Human | Human (1100.4) | 1100.6 | -0.2 | 1100.8 (-0.0%) |
| 5 | Frostfire | Human | 1097.6 | Orc 1085.3 / Human | Human (1097.6) | 1096.1 | +1.5 | 1098.0 (-0.0%) |
| 6 | Demonology | Human | 1060.1 | Orc 1054.4 / Human | Human (1060.1) | 1060.4 | -0.3 | 1064.9 (-0.5%) |
| 7 | Affliction | Human | 1030.4 | Orc 1024.5 / Human | Human (1030.4) | 1029.4 | +1.0 | 1036.4 (-0.6%) |
| 8 | Survival | Human | 1025.7 | Orc 1025.0 / Human | Human (1025.7) | 1025.4 | +0.3 | 1025.5 (+0.0%) |
| 9 | Destruction | Human | 1012.3 | Orc 1006.7 / Human | Human (1012.3) | 1011.9 | +0.4 | 1012.2 (+0.0%) |
| 10 | Destruction Fire | Human | 988.1 | Orc 986.6 / Human | Human (988.1) | 988.1 | +0.0 | 987.5 (+0.1%) |
| 11 | Fury (dual wield) | Orc | 987.3 | Orc / Human 987.1 | Orc (987.3) | 986.5 | +0.8 | 987.3 (+0.0%) |
| 12 | Arcane | Human | 983.2 | Orc 983.0 / Human | Human (983.2) | 982.5 | +0.7 | 983.2 (+0.0%) |
| 13 | Fury (two-hand) | Orc | 965.1 | Orc / Human 962.6 | Orc (965.1) | 965.6 | -0.5 | 965.1 (+0.0%) |
| 14 | Shadow | Human | 956.8 | Troll 949.6 / Human | Human (956.8) | 956.8 | +0.0 | 961.7 (-0.5%) |
| 15 | Frost | High Order | 947.1 | Orc 944.8 / High Order | High Order (947.1) | 946.9 | +0.2 | 947.1 (+0.0%) |
| 16 | Retribution | Human | 947.0 | Undead 940.1 / Human | Human (947.0) | 945.1 | +1.9 | 947.0 (+0.0%) |
| 17 | Arms | Orc | 906.7 | Orc / Human 904.4 | Orc (906.7) | 906.6 | +0.1 | 906.7 (+0.0%) |
| 18 | Enhancement | Orc | 887.6 | Orc / Dwarf 879.4 | Orc (887.6) | 888.0 | -0.4 | 887.7 (-0.0%) |
| 19 | Feral | Night Elf | 867.2 | Tauren 862.8 / Night Elf | Night Elf (867.2) | 866.8 | +0.4 | 867.1 (+0.0%) |
| 20 | Balance | Tauren | 852.5 | Tauren / Night Elf 850.0 | Tauren (852.5) | 852.1 | +0.4 | 856.7 (-0.5%) |
| 21 | Elemental | Windshaper | 776.6 | Windshaper / Dwarf 772.6 | Windshaper (776.6) | 776.9 | -0.3 | 777.1 (-0.1%) |
| 22 | Rogue Combat Daggers | High Order | 770.5 | Windshaper 770.5 / High Order | Windshaper (770.5) | 770.3 | +0.2 | 770.5 (+0.0%) |
| 23 | Rogue Combat Swords | Orc | 764.4 | Orc / High Order 763.9 | Windshaper (763.9) | 764.0 | -0.1 | 764.4 (+0.0%) |
| 24 | Smite | Undead | 762.9 | Undead / Gnome 742.8 | Undead (762.9) | 763.5 | -0.6 | 765.5 (-0.3%) |
| 25 | Rogue Seal Fate Daggers | Orc | 732.1 | Orc / Human 731.7 | Orc (732.1) | 731.5 | +0.6 | 732.1 (+0.0%) |
| 26 | Rogue Seal Fate Swords | Human | 729.4 | Orc 729.4 / Human | Orc (729.4) | 728.9 | +0.5 | 729.4 (+0.0%) |
| 27 | Rogue Mutilate | Orc | 719.4 | Orc / High Order 719.1 | Windshaper (719.1) | 718.9 | +0.2 | 719.4 (+0.0%) |
| 28 | Rogue Subtlety | Orc | 715.8 | Orc / Night Elf 714.6 | Night Elf (714.6) | 714.9 | -0.3 | 715.8 (+0.0%) |

| Variant | Best race | DPS | Report's published (at its race) | Before |
|---|---|---|---|---|
| druid__balance-wrath | High Order | 811.9 | 810.9 (at Tauren 811.3) | 814.0 |
| paladin__level_60_shock | Human | 869.2 | 868.8 (at Human 869.2) | 869.4 |
| paladin__level_60_sub_holy | Human | 949.8 | 949.1 (at Human 949.8) | 949.8 |
| priest__shadow-holy | Human | 914.2 | 914.2 (at Human 914.2) | 918.9 |
| shaman__elemental-level60-mana-tide | Windshaper | 761.8 | 761.6 (at Windshaper 761.8) | 762.2 |
| shaman__enhancement-imp-gw | Orc | 883.7 | 884.2 (at Orc 883.7) | 883.8 |
| shaman__enhancement-mana-tide | Orc | 798.8 | 800.0 (at Orc 798.8) | 798.9 |
| shaman__enhancement-water-shield | Windshaper | 839.7 | 839.8 (at Orc 839.6) | 839.7 |

Top over bottom: 1.60x

What moved / findings:

- DoT and channel specs, from live ticks (32a6848f9, `dot_live_ticks`): Affliction -6.0, Demonology -4.8, Shadow -4.9, Shadow + Holy -4.7, Balance -4.2, Balance Wrath -2.1, Smite -2.6 (#23 to #24, below Combat Swords), Elemental -0.5, Fire / Frostfire -0.4. Ranks otherwise unchanged.
- Hunters, from the Multi-Shot cast and base RAP 110: BM +2.2, Marksmanship -3.3, Lone Wolf -4.0, Survival +0.2. They stay #1 to #3 and #8.
- Feral +0.1 with the Bite change. Warriors, rogues, Arcane, Frost, Ret and Enhancement unchanged to the decimal, so no shared change leaked. The new core mechanics are gated only by spell school and Forever mode, never by class.
- Published against current: all within 1.9 of builds.json (0 of 57 over 0.5 in the shared 20k run). The req metas are 3 to 6 stale for the moved builds; the dashboard is current.
- Best race: report race everywhere, Lone Wolf High Order ties Windshaper (1123.9).
- No new findings. Top over bottom: 1.60x (was 1.59x).

<details><summary>Previous §1 (74b3a5f, 2026-10-26)</summary>

Same method: the report's base requests, every race, 5000 iterations, seed 1. All 19 Level 60 gear sets pass 1b and 5a.

| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Gap | Before (016daa6) |
|---|---|---|---|---|---|---|---|---|
| 1 | Beast Mastery | Human | 1141.4 | Orc 1137.6 / Human | Human (1141.4) | 1141.6 | -0.2 | 1171.9 (-2.6%) |
| 2 | Marksmanship | Human | 1139.6 | Orc 1136.4 / Human | Human (1139.6) | 1139.4 | +0.2 | 1168.4 (-2.5%) |
| 3 | Lone Wolf (hunter) | Human | 1127.9 | Windshaper 1127.4 / Human | Windshaper (1127.4) | 1127.0 | +0.4 | 1156.6 (-2.5%) |
| 4 | Fire | Human | 1100.8 | Orc 1096.2 / Human | Human (1100.8) | 1101.0 | -0.2 | 1100.8 (+0.0%) |
| 5 | Frostfire | Human | 1098.0 | Orc 1085.8 / Human | Human (1098.0) | 1096.4 | +1.6 | 1098.0 (+0.0%) |
| 6 | Demonology | Human | 1064.9 | Orc 1062.1 / Human | Human (1064.9) | 1065.1 | -0.2 | 1064.9 (+0.0%) |
| 7 | Affliction | Human | 1036.4 | Orc 1033.4 / Human | Human (1036.4) | 1035.3 | +1.1 | 1036.4 (+0.0%) |
| 8 | Survival | Human | 1025.5 | Orc 1024.8 / Human | Human (1025.5) | 1025.2 | +0.3 | 1055.7 (-2.9%) |
| 9 | Destruction | Human | 1012.2 | Orc 1008.9 / Human | Human (1012.2) | 1011.8 | +0.4 | 1012.2 (+0.0%) |
| 10 | Destruction Fire | Human | 987.5 | Orc 987.1 / Human | Human (987.5) | 987.4 | +0.1 | 987.5 (+0.0%) |
| 11 | Fury (dual wield) | Orc | 987.3 | Orc / Human 987.1 | Orc (987.3) | 986.6 | +0.7 | 1069.1 (-7.7%) |
| 12 | Arcane | Human | 983.2 | Orc 983.0 / Human | Human (983.2) | 982.5 | +0.7 | 983.2 (+0.0%) |
| 13 | Fury (two-hand) | Orc | 965.1 | Orc / Human 962.6 | Orc (965.1) | 965.6 | -0.5 | 1034.5 (-6.7%) |
| 14 | Shadow | Human | 961.7 | Troll 954.5 / Human | Human (961.7) | 961.8 | -0.1 | 961.7 (+0.0%) |
| 15 | Frost | High Order | 947.1 | Orc 944.8 / High Order | High Order (947.1) | 946.6 | +0.5 | 946.8 (+0.0%) |
| 16 | Retribution | Human | 947.0 | Undead 940.1 / Human | Human (947.0) | 945.1 | +1.9 | 947.0 (+0.0%) |
| 17 | Arms | Orc | 906.7 | Orc / Human 904.4 | Orc (906.7) | 906.6 | +0.1 | 956.5 (-5.2%) |
| 18 | Enhancement | Orc | 887.7 | Orc / Dwarf 879.4 | Orc (887.7) | 888.1 | -0.4 | 887.7 (+0.0%) |
| 19 | Feral | Night Elf | 867.1 | Tauren 862.7 / Night Elf | Night Elf (867.1) | 868.1 | -1.0 | 868.5 (-0.2%) |
| 20 | Balance | Tauren | 856.7 | Tauren / Night Elf 854.1 | Tauren (856.7) | 856.3 | +0.4 | 856.7 (+0.0%) |
| 21 | Elemental | Windshaper | 777.1 | Windshaper / Dwarf 773.0 | Windshaper (777.1) | 777.4 | -0.3 | 777.1 (+0.0%) |
| 22 | Rogue Combat Daggers | High Order | 770.5 | Windshaper 770.5 / High Order | Windshaper (770.5) | 770.3 | +0.2 | 770.5 (+0.0%) |
| 23 | Smite | Undead | 765.5 | Undead / Gnome 745.6 | Undead (765.5) | 766.1 | -0.6 | 765.5 (+0.0%) |
| 24 | Rogue Combat Swords | Orc | 764.4 | Orc / High Order 763.9 | Windshaper (763.9) | 764.0 | -0.1 | 764.4 (+0.0%) |
| 25 | Rogue Seal Fate Daggers | Orc | 732.1 | Orc / Human 731.7 | Orc (732.1) | 731.5 | +0.6 | 732.1 (+0.0%) |
| 26 | Rogue Seal Fate Swords | Human | 729.4 | Orc 729.4 / Human | Orc (729.4) | 728.9 | +0.5 | 729.4 (+0.0%) |
| 27 | Rogue Mutilate | Orc | 719.4 | Orc / High Order 719.1 | Windshaper (719.1) | 718.9 | +0.2 | 719.4 (+0.0%) |
| 28 | Rogue Subtlety | Orc | 715.8 | Orc / Night Elf 714.6 | Night Elf (714.6) | 715.9 | -1.3 | 715.8 (+0.0%) |

| Variant | Best race | DPS | Report's published (at its race) | Before |
|---|---|---|---|---|
| druid__balance-wrath | High Order | 814.0 | 813.3 (at Tauren 813.7) | 814.0 |
| paladin__level_60_shock | Human | 869.4 | 869.0 (at Human 869.4) | 869.4 |
| paladin__level_60_sub_holy | Human | 949.8 | 949.1 (at Human 949.8) | 949.8 |
| priest__shadow-holy | Human | 918.9 | 918.9 (at Human 918.9) | 918.9 |
| shaman__elemental-level60-mana-tide | Windshaper | 762.2 | 762.1 (at Windshaper 762.2) | 762.2 |
| shaman__enhancement-imp-gw | Orc | 883.8 | 884.3 (at Orc 883.8) | 883.8 |
| shaman__enhancement-mana-tide | Orc | 798.9 | 800.0 (at Orc 798.9) | 798.9 |
| shaman__enhancement-water-shield | Windshaper | 839.7 | 839.9 (at Orc 839.7) | 839.7 |

Top over bottom: 1.59x

What moved / findings:

- Warriors: Fury -82 (-7.7%), Fury 2H -69 (-6.7%), Arms -50 (-5.2%). That is the flat Rage ruling (107874b7, rage.go s = 0) and the re-searched warrior rotations (44fdd859, 41665489). Fury drops from #6 to #11, Arms from #15 to #17.
- Hunters: BM -30.5, Marksmanship -28.8, Lone Wolf -28.7, Survival -30.2 (-2.5 to -2.9%). That is Classic's crit per Agility at 60 (01afcb1d, 0f47f49e) plus the Multi-Shot rotations and the Human Marksmanship (cdceb56c). BM, Marksmanship and Lone Wolf stay #1 to #3 and Survival stays #8.
- Race base stats (92484f3c): Feral (Night Elf) -1.4, Subtlety (Night Elf) -1.2, Frost (High Order) +0.3. No report build plays a Dwarf or Gnome. Every other class is unchanged to the decimal, so no shared change leaked.
- Published against current: every build within 1.9 at 5000 iterations (Ret +1.9 and Frostfire +1.6 are the usual seed noise). The req metas of Feral and Subtlety are 1.0 to 1.5 stale (868.1 / 715.9 against 866.7 / 714.9), but the dashboard's builds.json carries the re-simmed numbers (117d18af), so the published ranking is current.
- Best race: fairness-20261025-1 resolved (Marksmanship now Human, the best race). Subtlety: Orc and Night Elf tie at 20k (715.2 / 714.7 vs 714.9 / 714.7, seeds 1 / 1000003), keep. Lone Wolf: Human and Windshaper tie at 20k (1127.1 / 1127.0 vs 1127.0 / 1127.6), keep.
- No new findings. Top over bottom: 1.59x (was 1.64x).

</details>

<details><summary>Previous §1 (016daa6, 2026-10-25 19:03)</summary>

Same method: the report's base requests, every race, 5000 iterations, seed 1. All 19 Level 60 gear sets pass 1b and 5a.

| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Gap | Before (54878fa) |
|---|---|---|---|---|---|---|---|---|
| 1 | Beast Mastery | Human | 1171.9 | Orc 1167.8 / Human | Human (1171.9) | 1171.5 | +0.4 | 1172.1 (-0.0%) |
| 2 | Marksmanship | Human | 1168.4 | Windshaper 1166.7 / Human | Windshaper (1166.7) | 1166.1 | +0.6 | 1169.4 (-0.1%) |
| 3 | Lone Wolf (hunter) | Windshaper | 1156.6 | Windshaper / Human 1156.1 | Windshaper (1156.6) | 1156.8 | -0.2 | 1159.7 (-0.3%) |
| 4 | Fire | Human | 1100.8 | Orc 1096.2 / Human | Human (1100.8) | 1101.0 | -0.2 | 1100.8 (+0.0%) |
| 5 | Frostfire | Human | 1098.0 | Orc 1085.8 / Human | Human (1098.0) | 1096.4 | +1.6 | 1098.0 (+0.0%) |
| 6 | Fury (dual wield) | Orc | 1069.1 | Orc / Human 1064.9 | Orc (1069.1) | 1068.7 | +0.4 | 1069.1 (+0.0%) |
| 7 | Demonology | Human | 1064.9 | Orc 1062.1 / Human | Human (1064.9) | 1065.1 | -0.2 | 1064.9 (+0.0%) |
| 8 | Survival | Human | 1055.7 | Orc 1054.7 / Human | Human (1055.7) | 1055.0 | +0.7 | 1057.1 (-0.1%) |
| 9 | Affliction | Human | 1036.4 | Orc 1033.4 / Human | Human (1036.4) | 1035.3 | +1.1 | 1036.4 (+0.0%) |
| 10 | Fury (two-hand) | Orc | 1034.5 | Orc / Human 1030.9 | Orc (1034.5) | 1033.7 | +0.8 | 1034.5 (+0.0%) |
| 11 | Destruction | Human | 1012.2 | Orc 1008.9 / Human | Human (1012.2) | 1011.8 | +0.4 | 1012.2 (+0.0%) |
| 12 | Destruction Fire | Human | 987.5 | Orc 987.1 / Human | Human (987.5) | 987.4 | +0.1 | 987.5 (+0.0%) |
| 13 | Arcane | Human | 983.2 | Orc 983.0 / Human | Human (983.2) | 982.5 | +0.7 | 983.2 (+0.0%) |
| 14 | Shadow | Human | 961.7 | Troll 954.5 / Human | Human (961.7) | 961.8 | -0.1 | 961.7 (+0.0%) |
| 15 | Arms | Orc | 956.5 | Orc / Human 955.4 | Orc (956.5) | 957.4 | -0.9 | 956.5 (+0.0%) |
| 16 | Retribution | Human | 947.0 | Undead 940.1 / Human | Human (947.0) | 945.1 | +1.9 | 947.0 (+0.0%) |
| 17 | Frost | High Order | 946.8 | Orc 944.8 / High Order | High Order (946.8) | 946.6 | +0.2 | 946.8 (+0.0%) |
| 18 | Enhancement | Orc | 887.7 | Orc / Dwarf 878.9 | Orc (887.7) | 888.1 | -0.4 | 887.7 (+0.0%) |
| 19 | Feral | Night Elf | 868.5 | Tauren 862.7 / Night Elf | Night Elf (868.5) | 868.1 | +0.4 | 868.5 (+0.0%) |
| 20 | Balance | Tauren | 856.7 | Tauren / Night Elf 854.1 | Tauren (856.7) | 856.3 | +0.4 | 856.7 (+0.0%) |
| 21 | Elemental | Windshaper | 777.1 | Windshaper / Dwarf 773.0 | Windshaper (777.1) | 777.4 | -0.3 | 777.1 (+0.0%) |
| 22 | Rogue Combat Daggers | Windshaper | 770.5 | Windshaper / High Order 769.9 | Windshaper (770.5) | 770.3 | +0.2 | 770.5 (+0.0%) |
| 23 | Smite | Undead | 765.5 | Undead / Gnome 745.6 | Undead (765.5) | 766.1 | -0.6 | 765.5 (+0.0%) |
| 24 | Rogue Combat Swords | Orc | 764.4 | Orc / High Order 763.7 | Windshaper (763.9) | 764.0 | -0.1 | 764.4 (+0.0%) |
| 25 | Rogue Seal Fate Daggers | Orc | 732.1 | Orc / Human 731.7 | Orc (732.1) | 731.5 | +0.6 | 732.1 (+0.0%) |
| 26 | Rogue Seal Fate Swords | Human | 729.4 | Orc 729.4 / Human | Orc (729.4) | 728.9 | +0.5 | 729.4 (+0.0%) |
| 27 | Rogue Mutilate | Orc | 719.4 | Orc / Human 718.7 | Windshaper (719.1) | 718.9 | +0.2 | 719.4 (+0.0%) |
| 28 | Rogue Subtlety | Orc | 715.8 | Orc / Night Elf 715.7 | Night Elf (715.7) | 715.9 | -0.2 | 715.8 (+0.0%) |

| Variant | Best race | DPS | Report's published (at its race) | Before |
|---|---|---|---|---|
| druid__balance-wrath | Windshaper | 814.0 | 813.3 (at Tauren 813.7) | 814.0 |
| paladin__level_60_shock | Human | 869.4 | 869.0 (at Human 869.4) | 869.4 |
| paladin__level_60_sub_holy | Human | 949.8 | 949.1 (at Human 949.8) | 949.8 |
| priest__shadow-holy | Human | 918.9 | 918.9 (at Human 918.9) | 918.9 |
| shaman__elemental-level60-mana-tide | Windshaper | 762.2 | 762.1 (at Windshaper 762.2) | 762.2 |
| shaman__enhancement-imp-gw | Orc | 883.8 | 884.3 (at Orc 883.8) | 883.8 |
| shaman__enhancement-mana-tide | Orc | 798.9 | 800.0 (at Orc 798.9) | 798.9 |
| shaman__enhancement-water-shield | Windshaper | 839.7 | 839.9 (at Orc 839.7) | 839.7 |

Top over bottom: 1.64x

What moved / findings:

- Only the hunters moved: Beast Mastery -0.2, Marksmanship -1.0, Survival -1.4, Lone Wolf -3.1 (best race against 54878fa). That is e71b90f5: the base presets are now the old Mark duty presets (own Hunter's Mark, one GCD and 60 mana each cast, the raid's mark off), which the ruling put at 0.5 to 2.5 DPS. Every other build and variant is identical to the 54878fa sweep, so the database changes (Wall of the Dead, Changuk Smasher, Axe of Inevitable Ruin) did not reach any preset.
- Hunter's Mark is ranged attack power only (core huntersMarkAura), and no request carries the raid's mark, so dropping it from the raid changes no non-hunter. The raid's Sunder stays on for all 57 requests; the hunter/warrior asymmetry is the user's ruling (standards 23), not re-flagged.
- Published against current: every build within 1.9 at 5000 iterations; the shared 20k run (vs_published) has 0 of 57 off by more than 0.5. Ret +1.9 and Frostfire +1.6 are seed noise, as last pass.
- Best race: fairness-20261025-1 still open. On the new Marksmanship request Human beats its Windshaper by 2.6 / 2.3 (20k, seeds 1 / 1000003: 1168.8 / 1168.0 vs 1166.1 / 1165.6). Lone Wolf stays Windshaper (+0.5 here). Other near-ties are noise.
- Top over bottom: 1.64x (unchanged).

</details>

<details><summary>Previous §1 (54878fa, 2026-10-25 13:18)</summary>

Same method: the report's base requests, every race, 5000 iterations, seed 1. All 19 Level 60 gear sets pass 1b and 5a.

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

What moved / findings:

- Every class moved against the 1e2f7ec sweep, all attributed: the alignment round (sim/core Rage on misses and crits, natural flasks, open explosives, crit elixir, wands, item effects; standards 4, 10, 11, 21 updates), each class's regear and rotation round (8% yellow cap, merged via claude/keen-brown-o2g32n), and standards 26 must-take talents. The largest: Demonology -112.9 (Suppression 5/5, measured 68.1 by its commit 0b566e87e), Beast Mastery +201.1, Survival +122.2, Marksmanship +124.6, Feral +115.6, Frost +102.4 (regear and rotation rounds). Hunters now hold the top three.
- Published against current: every build within 1.9 at 5000 iterations; the shared 20k run (vs_published) has 0 of 64 off by more than 0.5. Ret +1.9 and Frostfire +1.6 here are seed noise.
- Best race: Marksmanship loses 2.5 on its Windshaper to the Human (20k, seeds 1 and 1000003; the duty mark build 2.6). Finding fairness-20261025-1. Lone Wolf stays Windshaper (+1.0 at 20k). Rogue Combat Swords and Mutilate (Orc +0.5 / +0.3), Seal Fate Swords (Human = Orc), Subtlety (+0.1), Balance Wrath (+0.3) and Enhancement Water Shield (tie) are within noise.
- Feral has Engineering but no Goblin Sapper: correct, the sapper's spell row is not usable shapeshifted (consumes.go, measured 0.0 with it added).
- Top over bottom: 1.64x (was 1.79x).

</details>

<details><summary>Previous §1 (1e2f7ec, 2026-10-17 18:55)</summary>


Same method: the report's base requests, every race, 5000 iterations, seed 1. All 19 Level 60 gear sets pass 1b and 5a.

| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Gap | Before (0a204d4) |
|---|---|---|---|---|---|---|---|---|
| 1 | Demonology | Human | 1177.8 | Orc 1173.6 / Human | Human (1177.8) | 1177.7 | +0.1 | 1177.8 (+0.0%) |
| 2 | Lone Wolf (hunter) | Windshaper | 1094.6 | Windshaper / High Order 1093.2 | Windshaper (1094.6) | 1094.1 | +0.5 | 1094.6 (+0.0%) |
| 3 | Frostfire | Human | 1083.7 | Orc 1078.2 / Human | Human (1083.7) | 1082.5 | +1.2 | 1083.7 (+0.0%) |
| 4 | Fire | Human | 1067.5 | Orc 1062.2 / Human | Human (1067.5) | 1067.0 | +0.5 | 1067.5 (+0.0%) |
| 5 | Affliction | Human | 1066.4 | Orc 1063.2 / Human | Human (1066.4) | 1066.6 | -0.2 | 1066.4 (+0.0%) |
| 6 | Destruction | Human | 1048.1 | Orc 1045.7 / Human | Human (1048.1) | 1047.4 | +0.7 | 1048.1 (+0.0%) |
| 7 | Marksmanship | Windshaper | 1044.8 | Windshaper / High Order 1043.6 | Windshaper (1044.8) | 1045.3 | -0.5 | 1044.8 (+0.0%) |
| 8 | Destruction Fire | Human | 1019.7 | Orc 1018.5 / Human | Human (1019.7) | 1019.2 | +0.5 | 1019.7 (+0.0%) |
| 9 | Fury (dual wield) | Orc | 989.7 | Orc / High Order 985.2 | Orc (989.7) | 989.8 | -0.1 | 989.7 (+0.0%) |
| 10 | Arms | Orc | 976.8 | Orc / Human 975.1 | Orc (976.8) | 977.2 | -0.4 | 976.8 (+0.0%) |
| 11 | Beast Mastery | Windshaper | 971.0 | Windshaper / High Order 969.9 | Windshaper (971.0) | 970.7 | +0.3 | 971.0 (+0.0%) |
| 12 | Fury (two-hand) | Orc | 971.0 | Orc / Human 969.4 | Orc (971.0) | 970.1 | +0.9 | 971.0 (+0.0%) |
| 13 | Arcane | Orc | 964.8 | Orc / Human 964.8 | Human (964.8) | 964.8 | +0.0 | 964.8 (+0.0%) |
| 14 | Shadow | Human | 960.6 | Undead 952.5 / Human | Human (960.6) | 960.7 | -0.1 | 960.6 (+0.0%) |
| 15 | Survival | Orc | 934.9 | Orc / Human 934.9 | Human (934.9) | 935.4 | -0.5 | 934.9 (+0.0%) |
| 16 | Enhancement | Orc | 892.0 | Orc / Dwarf 882.7 | Orc (892.0) | 892.2 | -0.2 | 892.0 (+0.0%) |
| 17 | Retribution | Human | 889.1 | Undead 885.6 / Human | Human (889.1) | 889.2 | -0.1 | 889.1 (+0.0%) |
| 18 | Balance | Tauren | 873.1 | Tauren / Night Elf 871.1 | Tauren (873.1) | 872.8 | +0.3 | 873.1 (+0.0%) |
| 19 | Frost | High Order | 844.4 | Orc 837.6 / High Order | High Order (844.4) | 843.3 | +1.1 | 844.4 (+0.0%) |
| 20 | Elemental | Windshaper | 772.0 | Windshaper / Dwarf 766.5 | Windshaper (772.0) | 771.8 | +0.2 | 772.0 (+0.0%) |
| 21 | Smite | Undead | 765.8 | Undead / Gnome 744.1 | Undead (765.8) | 765.4 | +0.4 | 765.8 (+0.0%) |
| 22 | Feral | Tauren | 752.9 | Tauren / High Order 749.6 | Tauren (752.9) | 752.9 | +0.0 | 752.9 (+0.0%) |
| 23 | Rogue Combat Swords | Windshaper | 726.5 | Windshaper / High Order 726.4 | Windshaper (726.5) | 726.4 | +0.1 | 726.5 (+0.0%) |
| 24 | Rogue Combat Daggers | Windshaper | 725.5 | Windshaper / High Order 725.2 | Windshaper (725.5) | 725.9 | -0.4 | 725.5 (+0.0%) |
| 25 | Rogue Seal Fate Daggers | Windshaper | 687.2 | Windshaper / High Order 686.8 | Windshaper (687.2) | 687.1 | +0.1 | 687.2 (+0.0%) |
| 26 | Rogue Mutilate | Windshaper | 683.9 | Windshaper / High Order 683.5 | Windshaper (683.9) | 684.2 | -0.3 | 683.9 (+0.0%) |
| 27 | Rogue Seal Fate Swords | Windshaper | 677.8 | Windshaper / High Order 677.7 | Windshaper (677.8) | 678.1 | -0.3 | 677.8 (+0.0%) |
| 28 | Rogue Subtlety | Orc | 656.7 | Orc / High Order 655.6 | Orc (656.7) | 656.3 | +0.4 | 656.7 (+0.0%) |

| Variant | Best race | DPS | Report's published (at its race) | Before |
|---|---|---|---|---|
| druid__balance-wrath | Tauren | 826.2 | 825.9 (at Tauren 826.2) | 826.2 |
| paladin__level_60_shock | Human | 901.0 | 901.7 (at Human 901.0) | 901.0 |
| paladin__level_60_sub_holy | Human | 883.6 | 883.1 (at Human 883.6) | 883.6 |
| priest__shadow-holy | Human | 917.1 | 917.1 (at Human 917.1) | 917.1 |
| shaman__elemental-level60-mana-tide | Windshaper | 761.3 | 761.4 (at Windshaper 761.3) | 761.3 |
| shaman__enhancement-imp-gw | Orc | 887.9 | 887.8 (at Orc 887.9) | 887.9 |
| shaman__enhancement-mana-tide | Orc | 802.5 | 802.8 (at Orc 802.5) | 802.5 |
| shaman__enhancement-water-shield | Orc | 842.2 | 842.2 (at Orc 842.2) | 842.2 |

Top over bottom: 1.79x

What moved / findings:

- Nothing. The sweep log is byte-identical to the 0a204d4 sweep (every base build and variant, every race). The item database regeneration (cb3f64fa enchant values and item sets) changed no item in the report requests; the db.json diff touches set names and random-suffix entries only. Racials (1e2f7ec5) regenerate only the UI text; the sim's racial code is unchanged. The `highest_rank_id` / top-rank changes (bdf075cc, spell_ranks.go, unit.go, shaman healing.go, Enhancement `rotationsNameTopRanks`) affect leveling and the rotation editor only.
- Every build is within 1.2 DPS of its published number; the report race is the best race everywhere. No new findings.
- Top over bottom is still 1.79×.

</details>

<details><summary>Previous §1 (0a204d4, 15:04)</summary>

### 1 (superseded 18:55) Cross-spec table (master 0a204d4: Wowhead item database refresh, shaman heal rank families; full re-sweep 2026-10-17 15:04 UTC)

Same method: the report's base requests, every race, 5000 iterations, seed 1. All 19 Level 60 gear sets pass 1b and 5a.

| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Gap | Before (1735382) |
|---|---|---|---|---|---|---|---|---|
| 1 | Demonology | Human | 1177.8 | Orc 1173.6 / Human | Human (1177.8) | 1177.7 | +0.1 | 1177.8 (+0.0%) |
| 2 | Lone Wolf (hunter) | Windshaper | 1094.6 | Windshaper / High Order 1093.2 | Windshaper (1094.6) | 1094.1 | +0.5 | 1094.6 (+0.0%) |
| 3 | Frostfire | Human | 1083.7 | Orc 1078.2 / Human | Human (1083.7) | 1082.5 | +1.2 | 1083.7 (+0.0%) |
| 4 | Fire | Human | 1067.5 | Orc 1062.2 / Human | Human (1067.5) | 1067.0 | +0.5 | 1067.5 (+0.0%) |
| 5 | Affliction | Human | 1066.4 | Orc 1063.2 / Human | Human (1066.4) | 1066.6 | -0.2 | 1066.4 (+0.0%) |
| 6 | Destruction | Human | 1048.1 | Orc 1045.7 / Human | Human (1048.1) | 1047.4 | +0.7 | 1048.1 (+0.0%) |
| 7 | Marksmanship | Windshaper | 1044.8 | Windshaper / High Order 1043.6 | Windshaper (1044.8) | 1045.3 | -0.5 | 1044.8 (+0.0%) |
| 8 | Destruction Fire | Human | 1019.7 | Orc 1018.5 / Human | Human (1019.7) | 1019.2 | +0.5 | 1019.7 (+0.0%) |
| 9 | Fury (dual wield) | Orc | 989.7 | Orc / High Order 985.2 | Orc (989.7) | 989.8 | -0.1 | 989.7 (+0.0%) |
| 10 | Arms | Orc | 976.8 | Orc / Human 975.1 | Orc (976.8) | 977.2 | -0.4 | 976.8 (+0.0%) |
| 11 | Beast Mastery | Windshaper | 971.0 | Windshaper / High Order 969.9 | Windshaper (971.0) | 970.7 | +0.3 | 971.0 (+0.0%) |
| 12 | Fury (two-hand) | Orc | 971.0 | Orc / Human 969.4 | Orc (971.0) | 970.1 | +0.9 | 971.0 (+0.0%) |
| 13 | Arcane | Orc | 964.8 | Orc / Human 964.8 | Human (964.8) | 964.8 | +0.0 | 964.8 (+0.0%) |
| 14 | Shadow | Human | 960.6 | Undead 952.5 / Human | Human (960.6) | 960.7 | -0.1 | 960.6 (+0.0%) |
| 15 | Survival | Orc | 934.9 | Orc / Human 934.9 | Human (934.9) | 935.4 | -0.5 | 934.9 (+0.0%) |
| 16 | Enhancement | Orc | 892.0 | Orc / Dwarf 882.7 | Orc (892.0) | 892.2 | -0.2 | 892.0 (+0.0%) |
| 17 | Retribution | Human | 889.1 | Undead 885.6 / Human | Human (889.1) | 889.2 | -0.1 | 889.1 (+0.0%) |
| 18 | Balance | Tauren | 873.1 | Tauren / Night Elf 871.1 | Tauren (873.1) | 872.8 | +0.3 | 873.1 (+0.0%) |
| 19 | Frost | High Order | 844.4 | Orc 837.6 / High Order | High Order (844.4) | 843.3 | +1.1 | 844.4 (+0.0%) |
| 20 | Elemental | Windshaper | 772.0 | Windshaper / Dwarf 766.5 | Windshaper (772.0) | 771.8 | +0.2 | 772.0 (+0.0%) |
| 21 | Smite | Undead | 765.8 | Undead / Gnome 744.1 | Undead (765.8) | 765.4 | +0.4 | 765.8 (+0.0%) |
| 22 | Feral | Tauren | 752.9 | Tauren / High Order 749.6 | Tauren (752.9) | 752.9 | +0.0 | 752.9 (+0.0%) |
| 23 | Rogue Combat Swords | Windshaper | 726.5 | Windshaper / High Order 726.4 | Windshaper (726.5) | 726.4 | +0.1 | 726.5 (+0.0%) |
| 24 | Rogue Combat Daggers | Windshaper | 725.5 | Windshaper / High Order 725.2 | Windshaper (725.5) | 725.9 | -0.4 | 725.5 (+0.0%) |
| 25 | Rogue Seal Fate Daggers | Windshaper | 687.2 | Windshaper / High Order 686.8 | Windshaper (687.2) | 687.1 | +0.1 | 687.2 (+0.0%) |
| 26 | Rogue Mutilate | Windshaper | 683.9 | Windshaper / High Order 683.5 | Windshaper (683.9) | 684.2 | -0.3 | 683.9 (+0.0%) |
| 27 | Rogue Seal Fate Swords | Windshaper | 677.8 | Windshaper / High Order 677.7 | Windshaper (677.8) | 678.1 | -0.3 | 677.8 (+0.0%) |
| 28 | Rogue Subtlety | Orc | 656.7 | Orc / High Order 655.6 | Orc (656.7) | 656.3 | +0.4 | 656.7 (+0.0%) |

| Variant | Best race | DPS | Report's published (at its race) | Before |
|---|---|---|---|---|
| druid__balance-wrath | Tauren | 826.2 | 825.9 (at Tauren 826.2) | 826.2 |
| paladin__level_60_shock | Human | 901.0 | 901.7 (at Human 901.0) | 901.0 |
| paladin__level_60_sub_holy | Human | 883.6 | 883.1 (at Human 883.6) | 883.6 |
| priest__shadow-holy | Human | 917.1 | 917.1 (at Human 917.1) | 917.1 |
| shaman__elemental-level60-mana-tide | Windshaper | 761.3 | 761.4 (at Windshaper 761.3) | 761.3 |
| shaman__enhancement-imp-gw | Orc | 887.9 | 887.8 (at Orc 887.9) | 887.9 |
| shaman__enhancement-mana-tide | Orc | 802.5 | 802.8 (at Orc 802.5) | 802.5 |
| shaman__enhancement-water-shield | Orc | 842.2 | 842.2 (at Orc 842.2) | 842.2 |

Top over bottom: 1.79x

What moved / findings:

- Nothing. Every base build and variant, at every race, is identical to the 1735382 sweep. The db regeneration (b1b07631) changed no synthetic item (ids 99xxxx) and no item in the report requests; the shaman heal rank families (b803e856) touch heals only.
- Every build is within 1.2 DPS of its published number (report version 35); the report race is the best race everywhere. No new findings.
- Top over bottom is still 1.79×.

</details>

<details><summary>Previous §1 (1735382, 00:50)</summary>

### 1 (superseded 15:04) Cross-spec table (master 1735382: raid Flametongue Totem for Enhancement, rotation pass 24; full re-sweep 2026-10-17 00:50 UTC)

Same method: the report's base requests, every race, 5000 iterations, seed 1. All 19 Level 60 gear sets pass 1b and 5a.

| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Gap | Before (850988f) |
|---|---|---|---|---|---|---|---|---|
| 1 | Demonology | Human | 1177.8 | Orc 1173.6 / Human | Human (1177.8) | 1177.7 | +0.1 | 1177.8 (+0.0%) |
| 2 | Lone Wolf (hunter) | Windshaper | 1094.6 | Windshaper / High Order 1093.2 | Windshaper (1094.6) | 1094.1 | +0.5 | 1094.6 (+0.0%) |
| 3 | Frostfire | Human | 1083.7 | Orc 1078.2 / Human | Human (1083.7) | 1082.5 | +1.2 | 1083.7 (+0.0%) |
| 4 | Fire | Human | 1067.5 | Orc 1062.2 / Human | Human (1067.5) | 1067.0 | +0.5 | 1067.5 (+0.0%) |
| 5 | Affliction | Human | 1066.4 | Orc 1063.2 / Human | Human (1066.4) | 1066.6 | -0.2 | 1066.4 (+0.0%) |
| 6 | Destruction | Human | 1048.1 | Orc 1045.7 / Human | Human (1048.1) | 1047.4 | +0.7 | 1048.1 (+0.0%) |
| 7 | Marksmanship | Windshaper | 1044.8 | Windshaper / High Order 1043.6 | Windshaper (1044.8) | 1045.3 | -0.5 | 1044.8 (+0.0%) |
| 8 | Destruction Fire | Human | 1019.7 | Orc 1018.5 / Human | Human (1019.7) | 1019.2 | +0.5 | 1019.7 (+0.0%) |
| 9 | Fury (dual wield) | Orc | 989.7 | Orc / High Order 985.2 | Orc (989.7) | 989.8 | -0.1 | 989.7 (+0.0%) |
| 10 | Arms | Orc | 976.8 | Orc / Human 975.1 | Orc (976.8) | 977.2 | -0.4 | 976.8 (+0.0%) |
| 11 | Beast Mastery | Windshaper | 971.0 | Windshaper / High Order 969.9 | Windshaper (971.0) | 970.7 | +0.3 | 969.4 (+0.2%) |
| 12 | Fury (two-hand) | Orc | 971.0 | Orc / Human 969.4 | Orc (971.0) | 970.1 | +0.9 | 971.0 (+0.0%) |
| 13 | Arcane | Orc | 964.8 | Orc / Human 964.8 | Human (964.8) | 964.8 | +0.0 | 964.8 (+0.0%) |
| 14 | Shadow | Human | 960.6 | Undead 952.5 / Human | Human (960.6) | 960.7 | -0.1 | 960.6 (+0.0%) |
| 15 | Survival | Orc | 934.9 | Orc / Human 934.9 | Human (934.9) | 935.4 | -0.5 | 934.9 (+0.0%) |
| 16 | Enhancement | Orc | 892.0 | Orc / Dwarf 882.7 | Orc (892.0) | 892.2 | -0.2 | 847.6 (+5.2%) |
| 17 | Retribution | Human | 889.1 | Undead 885.6 / Human | Human (889.1) | 889.2 | -0.1 | 889.1 (+0.0%) |
| 18 | Balance | Tauren | 873.1 | Tauren / Night Elf 871.1 | Tauren (873.1) | 872.8 | +0.3 | 873.1 (+0.0%) |
| 19 | Frost | High Order | 844.4 | Orc 837.6 / High Order | High Order (844.4) | 843.3 | +1.1 | 844.4 (+0.0%) |
| 20 | Elemental | Windshaper | 772.0 | Windshaper / Dwarf 766.5 | Windshaper (772.0) | 771.8 | +0.2 | 772.0 (+0.0%) |
| 21 | Smite | Undead | 765.8 | Undead / Gnome 744.1 | Undead (765.8) | 765.4 | +0.4 | 765.8 (+0.0%) |
| 22 | Feral | Tauren | 752.9 | Tauren / High Order 749.6 | Tauren (752.9) | 752.9 | +0.0 | 752.9 (+0.0%) |
| 23 | Rogue Combat Swords | Windshaper | 726.5 | Windshaper / High Order 726.4 | Windshaper (726.5) | 726.4 | +0.1 | 726.5 (+0.0%) |
| 24 | Rogue Combat Daggers | Windshaper | 725.5 | Windshaper / High Order 725.2 | Windshaper (725.5) | 725.9 | -0.4 | 725.5 (+0.0%) |
| 25 | Rogue Seal Fate Daggers | Windshaper | 687.2 | Windshaper / High Order 686.8 | Windshaper (687.2) | 687.1 | +0.1 | 687.2 (+0.0%) |
| 26 | Rogue Mutilate | Windshaper | 683.9 | Windshaper / High Order 683.5 | Windshaper (683.9) | 684.2 | -0.3 | 683.9 (+0.0%) |
| 27 | Rogue Seal Fate Swords | Windshaper | 677.8 | Windshaper / High Order 677.7 | Windshaper (677.8) | 678.1 | -0.3 | 677.8 (+0.0%) |
| 28 | Rogue Subtlety | Orc | 656.7 | Orc / High Order 655.6 | Orc (656.7) | 656.3 | +0.4 | 656.7 (+0.0%) |

**Findings:**
- **The published report is current.** Every base build and variant is within 1.2 DPS of its published number. The report's race is the best race everywhere; Arcane and Survival are exact Human/Orc ties.
- **What moved:**
  - Enhancement 847.6 to 892.0 (+5.2%): the raid's Flametongue Totem (standards 25, user ruling) plus rotation pass 24, with Elemental Focus in place of a Tidal Focus point. It passes Ret, Balance and Frost for 16th. Its variants gained +21 to +45.
  - BM +1.6 (Serpent Sting last).
  - Everything else is identical to the decimal.
- **Flametongue on top of the Windfury imbue** is now report assumption `sham_ftt_stack` with one value, per the user's ruling. It's still To Verify in game.
- Top over bottom is 1.79×.

</details>

<details><summary>Previous §1 (850988f, 18:40)</summary>

### 1 (superseded 00:50) Cross-spec table (master 850988f: no level partial resist, Salvation, rotation pass 23, the melee and rogue republishes; full re-sweep 2026-10-16 18:40 UTC)

Same method as 12:20: the report's own base requests, every race, 5000 iterations, seed 1. All 19 Level 60 gear sets pass 1b and 5a.

| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Gap | Before (8f0db21) |
|---|---|---|---|---|---|---|---|---|
| 1 | Demonology | Human | 1177.8 | Orc 1173.6 / Human | Human (1177.8) | 1177.7 | +0.1 | 1118.6 (+5.3%) |
| 2 | Lone Wolf (hunter) | Windshaper | 1094.6 | Windshaper / High Order 1093.2 | Windshaper (1094.6) | 1094.1 | +0.5 | 1076.1 (+1.7%) |
| 3 | Frostfire | Human | 1083.7 | Orc 1078.2 / Human | Human (1083.7) | 1082.5 | +1.2 | 1011.4 (+7.1%) |
| 4 | Fire | Human | 1067.5 | Orc 1062.2 / Human | Human (1067.5) | 1067.0 | +0.5 | 993.8 (+7.4%) |
| 5 | Affliction | Human | 1066.4 | Orc 1063.2 / Human | Human (1066.4) | 1066.6 | -0.2 | 1012.9 (+5.3%) |
| 6 | Destruction | Human | 1048.1 | Orc 1045.7 / Human | Human (1048.1) | 1047.4 | +0.7 | 997.1 (+5.1%) |
| 7 | Marksmanship | Windshaper | 1044.8 | Windshaper / High Order 1043.6 | Windshaper (1044.8) | 1045.3 | -0.5 | 1029.2 (+1.5%) |
| 8 | Destruction Fire | Human | 1019.7 | Orc 1018.5 / Human | Human (1019.7) | 1019.2 | +0.5 | 970.9 (+5.0%) |
| 9 | Fury (dual wield) | Orc | 989.7 | Orc / High Order 985.2 | Orc (989.7) | 989.8 | -0.1 | 986.0 (+0.4%) |
| 10 | Arms | Orc | 976.8 | Orc / Human 975.1 | Orc (976.8) | 977.2 | -0.4 | 976.1 (+0.1%) |
| 11 | Fury (two-hand) | Orc | 971.0 | Orc / Human 969.4 | Orc (971.0) | 970.1 | +0.9 | 970.2 (+0.1%) |
| 12 | Beast Mastery | Windshaper | 969.4 | Windshaper / Human 968.7 | Windshaper (969.4) | 969.3 | +0.1 | 955.2 (+1.5%) |
| 13 | Arcane | Orc | 964.8 | Orc / Human 964.8 | Human (964.8) | 964.8 | +0.0 | 882.8 (+9.3%) |
| 14 | Shadow | Human | 960.6 | Undead 952.5 / Human | Human (960.6) | 960.7 | -0.1 | 926.0 (+3.7%) |
| 15 | Survival | Orc | 934.9 | Orc / Human 934.9 | Human (934.9) | 935.4 | -0.5 | 926.1 (+1.0%) |
| 16 | Retribution | Human | 889.1 | Undead 885.6 / Human | Human (889.1) | 889.2 | -0.1 | 852.6 (+4.3%) |
| 17 | Balance | Tauren | 873.1 | Tauren / Night Elf 871.1 | Tauren (873.1) | 872.8 | +0.3 | 828.6 (+5.4%) |
| 18 | Enhancement | Orc | 847.6 | Orc / Dwarf 839.6 | Orc (847.6) | 848.1 | -0.5 | 826.0 (+2.6%) |
| 19 | Frost | High Order | 844.4 | Orc 837.6 / High Order | High Order (844.4) | 843.3 | +1.1 | 827.1 (+2.1%) |
| 20 | Elemental | Windshaper | 772.0 | Windshaper / Dwarf 766.5 | Windshaper (772.0) | 771.8 | +0.2 | 725.0 (+6.5%) |
| 21 | Smite | Undead | 765.8 | Undead / Gnome 744.1 | Undead (765.8) | 765.4 | +0.4 | 724.1 (+5.8%) |
| 22 | Feral | Tauren | 752.9 | Tauren / High Order 749.6 | Tauren (752.9) | 752.9 | +0.0 | 752.7 (+0.0%) |
| 23 | Rogue Combat Swords | Windshaper | 726.5 | Windshaper / High Order 726.4 | Windshaper (726.5) | 726.4 | +0.1 | 719.4 (+1.0%) |
| 24 | Rogue Combat Daggers | Windshaper | 725.5 | Windshaper / High Order 725.2 | Windshaper (725.5) | 725.9 | -0.4 | 714.8 (+1.5%) |
| 25 | Rogue Seal Fate Daggers | Windshaper | 687.2 | Windshaper / High Order 686.8 | Windshaper (687.2) | 687.1 | +0.1 | 683.1 (+0.6%) |
| 26 | Rogue Mutilate | Windshaper | 683.9 | Windshaper / High Order 683.5 | Windshaper (683.9) | 684.2 | -0.3 | 679.5 (+0.6%) |
| 27 | Rogue Seal Fate Swords | Windshaper | 677.8 | Windshaper / High Order 677.7 | Windshaper (677.8) | 678.1 | -0.3 | 672.8 (+0.7%) |
| 28 | Rogue Subtlety | Orc | 656.7 | Orc / High Order 655.6 | Orc (656.7) | 656.3 | +0.4 | 652.4 (+0.7%) |

**Update 22:30 (master 23d0b8f, results auditor pass 39):** the second shaman's fire slot now holds Flametongue Totem (user ruling, standards 25), so Enhancement goes from 848.1 to 886.4 and passes Balance and Frost, from 18th to 17th. The raid stays symmetric: every melee spec gets its best weapon totem, Windfury for the others and Flametongue for Enhancement, whose own Windfury Weapon turns off the totem's buff for it. FTT doesn't reach ranged attacks or casters, so no one else is affected. The +38 depends on Flametongue stacking with the Windfury imbue, which is still To Verify; the results auditor proposed it as a dashboard assumption row, which I support.

**Findings:**
- **The published report is current.** Every base build and variant is within 1.2 DPS of its published number at the report's race. The 12:20 melee gap and the 16:05 partial-resist gap are both closed.
- **The report's race is the best race for every build,** or within noise of it (Arcane and Survival: Human and Orc tie).
- **The partial-resist ruling moved the casters up 3.7 to 9.3%:** Arcane +9.3%, Fire +7.4%, Frostfire +7.1%, Elemental +6.5%, Smite +5.8%, Balance +5.4%, warlocks +5.0 to +5.3%, Ret +4.3%, Shadow +3.7%. Melee moved 0 to 1.5%.
  - Top over bottom is now 1.79× (was 1.71×).
  - Every caster is now above every rogue and Feral; the eight spell-casting specs above 1000 DPS are all mage or warlock.
  - This applies the user's ruling evenly, so it is a balance finding, not a fairness defect.
- Salvation (threat only) and Ret's own Concentration Aura change no other build.

</details>

<details><summary>Previous §1 (8f0db21, 12:20)</summary>

### 1 (superseded 18:40) Cross-spec table (master 8f0db21: the class ports of client 70425, weapon-free racials, raid hits on every DPS, Grace of Air, main-hand Windfury; full re-sweep 2026-10-16 12:20 UTC)

**Method change.** The sweep now runs the DPS report's own base requests (tools/dps_report/req/*.json) and changes only the race, at 5000 iterations, seed 1. My hand-kept talent strings had gone stale after the class ports. Checks at 20k: Fury 986.5 (published 987.9), Balance 828.2 (published 828.2). The weapon-racials-off column is retired with standards 16 (no racial depends on the weapon any more). All 19 Level 60 gear sets pass 1b and 5a.

| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Report understated by |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Human | 1118.6 | Orc 1114.4 / Human | Human (1118.6) | 1118.3 | +0.3 |
| 2 | Lone Wolf (hunter) | Windshaper | 1076.1 | Windshaper / High Order 1074.7 | Windshaper (1076.1) | 1075.7 | +0.4 |
| 3 | Marksmanship | Windshaper | 1029.2 | Windshaper / High Order 1028.0 | Windshaper (1029.2) | 1029.6 | -0.4 |
| 4 | Affliction | Human | 1012.9 | Orc 1009.7 / Human | Human (1012.9) | 1013.3 | -0.4 |
| 5 | Frostfire | Human | 1011.4 | Orc 1006.5 / Human | Human (1011.4) | 1010.8 | +0.6 |
| 6 | Destruction | Human | 997.1 | Orc 994.6 / Human | Human (997.1) | 995.9 | +1.2 |
| 7 | Fire | Human | 993.8 | Orc 989.5 / Human | Human (993.8) | 993.2 | +0.6 |
| 8 | Fury (dual wield) | Orc | 986.0 | Orc / High Order 980.9 | Orc (986.0) | 987.9 | -1.9 |
| 9 | Arms | Orc | 976.1 | Orc / Human 974.3 | Orc (976.1) | 944.2 | +31.9 |
| 10 | Destruction Fire | Human | 970.9 | Orc 969.7 / Human | Human (970.9) | 970.2 | +0.7 |
| 11 | Fury (two-hand) | Orc | 970.2 | Orc / Human 968.6 | Orc (970.2) | 932.3 | +37.9 |
| 12 | Beast Mastery | Windshaper | 955.2 | Windshaper / Human 954.4 | Windshaper (955.2) | 955.1 | +0.1 |
| 13 | Survival | Orc | 926.1 | Orc / Human 925.8 | Human (925.8) | 904.6 | +21.2 |
| 14 | Shadow | Human | 926.0 | Undead 918.2 / Human | Human (926.0) | 928.6 | -2.6 |
| 15 | Arcane | Human | 882.8 | Orc 881.9 / Human | Human (882.8) | 884.9 | -2.1 |
| 16 | Retribution | Human | 852.6 | Undead 849.4 / Human | Human (852.6) | 828.9 | +23.7 |
| 17 | Balance | Tauren | 828.6 | Tauren / Night Elf 826.4 | Tauren (828.6) | 828.2 | +0.4 |
| 18 | Frost | High Order | 827.1 | Orc 820.1 / High Order | High Order (827.1) | 827.3 | -0.2 |
| 19 | Enhancement | Orc | 826.0 | Orc / Dwarf 817.7 | Orc (826.0) | 825.3 | +0.7 |
| 20 | Feral | Tauren | 752.7 | Tauren / High Order 749.4 | Tauren (752.7) | 713.7 | +39.0 |
| 21 | Elemental | Windshaper | 725.0 | Windshaper / Dwarf 719.9 | Windshaper (725.0) | 724.9 | +0.1 |
| 22 | Smite | Undead | 724.1 | Undead / Gnome 703.2 | Undead (724.1) | 724.6 | -0.5 |
| 23 | Rogue Combat Swords | Windshaper | 719.4 | Windshaper / High Order 719.4 | Orc (716.3) | 704.7 | +11.6 |
| 24 | Rogue Combat Daggers | Windshaper | 714.8 | Windshaper / High Order 714.5 | Undead (712.3) | 698.7 | +13.6 |
| 25 | Rogue Seal Fate Daggers | Windshaper | 683.1 | Windshaper / High Order 682.7 | Undead (681.7) | 665.5 | +16.2 |
| 26 | Rogue Mutilate | Windshaper | 679.5 | Windshaper / High Order 679.1 | Undead (679.0) | 669.9 | +9.1 |
| 27 | Rogue Seal Fate Swords | Windshaper | 672.8 | Windshaper / High Order 672.6 | Windshaper (672.8) | 666.5 | +6.3 |
| 28 | Rogue Subtlety | Orc | 652.4 | Orc / High Order 651.3 | Undead (651.5) | 634.3 | +17.2 |

| Variant | Best race | DPS | Report's published |
|---|---|---|---|
| druid__balance-wrath | Tauren | 784.8 | 784.4 (at Tauren 784.8) |
| paladin__level_60_shock | Human | 864.8 | 851.0 (at Human 864.8) |
| paladin__level_60_sub_holy | Human | 847.3 | 822.9 (at Human 847.3) |
| priest__shadow-holy | Human | 883.9 | 886.4 (at Human 883.9) |
| shaman__elemental-level60-mana-tide | Windshaper | 714.3 | 714.6 (at Windshaper 714.3) |
| shaman__enhancement-imp-gw | Orc | 820.1 | 818.4 (at Orc 820.1) |
| shaman__enhancement-mana-tide | Orc | 761.0 | 760.5 (at Orc 761.0) |
| shaman__enhancement-water-shield | Orc | 795.1 | 795.2 (at Orc 795.1) |

**Findings:**
- **The published numbers understate nine melee builds** against their current requests:
  - Feral +39.0, Fury 2H +37.9, Arms +31.9, Ret +23.7, Survival +21.2
  - rogues +6 to +17
  - Ret Shock +13.8 and Ret Sub Holy +24.4 (variants)
  
  This matches the results auditor's pass 34, which I reproduced independently: they predate Grace of Air in the melee raid and main-hand-only Windfury. The table above ranks the current numbers. The report's published ranking puts every one of these builds too low, for example Fury 2H 11th here against 13th in the report. A full report re-run is needed and has been requested.
- **Resolved 13:57 (rogue republish #97, master 9cc6436):** the rogue presets now use Windshaper (Subtlety Orc), the best race in this sweep, and their published numbers are current: Combat Daggers 722.8, Combat Swords 723.4, Mutilate 679.7, Seal Fate Daggers 683.0, Seal Fate Swords 673.8, Subtlety 652.0. The results auditor verified them at 20k. The original finding was: **The report's race is not the best race for five rogue builds:** Windshaper's 3% haste now leads, by 0.5 to 3.1 DPS. Combat Swords Orc 716.3 against Windshaper 719.4; Combat Daggers Undead +2.5; Seal Fate Daggers +1.4; Subtlety Undead against Orc +0.9; Mutilate +0.5. Survival's Human is 0.3 under Orc, which is noise. All are small, but the rogue worker should take the best race when it re-runs.
- **Human is now the best race for every warlock, mage (except Frost), Shadow and Ret,** from Ruthless's 2% crit on spells and attacks (client 70425 text, sourced). Orc is within 1 to 4 everywhere, so the Horde/Alliance gap stays small. Smite is the exception: Undead leads Gnome by 20.9.
- **Pushback talents resolved:** Nature's Focus (Balance) and Improved Channeling (Arcane) are modelled now. The shaman raid hits are real spells, as the warriors' are, so the Lightning Bolt pushback asymmetry is gone too.
- **Big moves from the class ports**, which are class-specific and owned by the workers:
  - Frost 949.1 to 827.1 (mage 70425 port)
  - Ret 977.2 to 852.6 (70425 paladin values)
  - Fury 1058.5 to 986.0 (off-hand Rage ×0.5 ruling)
- Top over bottom is 1.71×.

</details>

<details><summary>Previous §1 (8f718b3, 08:00)</summary>

### 1 (superseded 12:20) Cross-spec table (master 8f718b3: client 70425 racials, Moonkin Aura and Leader of the Pack; full re-sweep 2026-10-16 08:00 UTC)

Every class was rebuilt from master 8f718b3 and fully re-swept with the usual settings. All 19 Level 60 gear sets pass 1b and 5a. "Racials off" turns off only the weapon racials and shows the best race; "same" means that race doesn't change. "Before" is the e1d2db0 table (01:40).

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (e1d2db0) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1117.9 | Undead 1107.0 / Gnome | same | 1116.0 | +0.2% |
| 2 | Fury (dual wield) | Orc | 1058.5 | Orc / Human 1054.5 | Undead 1047.5 | 1065.7 | -0.7% |
| 3 | Marksmanship | Night Elf | 1020.7 | Windshaper 1012.5 / Night Elf | same | 1020.7 | +0.0% |
| 4 | Frostfire | Gnome | 1012.7 | Undead 1008.7 / Gnome | same | 1011.9 | +0.1% |
| 5 | Destruction | Gnome | 990.4 | Undead 984.0 / Gnome | same | 988.7 | +0.2% |
| 6 | Affliction | Gnome | 984.7 | Undead 975.1 / Gnome | same | 983.4 | +0.1% |
| 7 | Retribution | Undead | 977.2 | Undead / Human 967.2 | same | 981.5 | -0.4% |
| 8 | Destruction Fire | Gnome | 974.2 | Undead 969.2 / Gnome | same | 973.2 | +0.1% |
| 9 | Beast Mastery | Night Elf | 952.0 | Windshaper 945.5 / Night Elf | same | 952.0 | +0.0% |
| 10 | Frost | High Order | 949.0 | Undead 946.8 / High Order | same | 949.1 | -0.0% |
| 11 | Arms | Orc | 948.9 | Orc / Night Elf 942.1 | Undead 947.1 | 954.2 | -0.6% |
| 12 | Fury (two-hand) | Orc | 936.2 | Orc / Night Elf 923.6 | Undead 930.6 | 943.0 | -0.7% |
| 13 | Survival | Orc | 888.1 | Orc / Night Elf 881.8 | Night Elf 881.8 | 892.6 | -0.5% |
| 14 | Fire | Gnome | 886.3 | Undead 883.8 / Gnome | same | 889.6 | -0.4% |
| 15 | Shadow | Gnome | 852.7 | Undead 851.3 / Gnome | same | 855.0 | -0.3% |
| 16 | Arcane | Undead | 834.1 | Undead / Gnome 831.5 | same | 843.9 | -1.2% |
| 17 | Balance | Night Elf | 832.0 | Tauren 822.4 / Night Elf | same | 832.0 | +0.0% |
| 18 | Enhancement | Orc | 831.0 | Orc / Dwarf 816.0 | same | 834.6 | -0.4% |
| 19 | Feral | Night Elf | 717.2 | Windshaper 713.7 / Night Elf | same | 717.3 | -0.0% |
| 20 | Smite | Undead | 713.9 | Undead / Gnome 707.9 | same | 716.4 | -0.3% |
| 21 | Rogue Combat Swords | Orc | 704.7 | Orc / Gnome 698.0 | Undead 702.3 | 710.5 | -0.8% |
| 22 | Rogue Combat Daggers | Undead | 698.2 | Undead / Night Elf 693.1 | same | 708.9 | -1.5% |
| 23 | Elemental | Windshaper | 693.1 | Windshaper / Dwarf 687.7 | same | 694.8 | -0.2% |
| 24 | Rogue Mutilate | Undead | 670.0 | Undead / Gnome 667.0 | same | 680.8 | -1.6% |
| 25 | Rogue Seal Fate Daggers | Undead | 666.4 | Undead / Gnome 662.6 | same | 674.9 | -1.3% |
| 26 | Rogue Seal Fate Swords | Undead | 666.0 | Undead / Gnome 662.7 | same | 673.3 | -1.1% |
| 27 | Rogue Subtlety | Undead | 634.0 | Undead / Gnome 632.3 | same | 644.4 | -1.6% |

**What moved (client 70425's racial changes only, so every change is race-specific):**
- **Undead: Touch of the Grave** fires 20% / 30% a hit, but at most once every 25 sec (it was 5% / 10% with a 1 sec lockout). That's a large loss for fast hitters:
  - Fury Undead -18.2
  - rogues Undead -7 to -11
  - Arcane Undead -9.8
  - Ret -4.3
  - casters -1 to -6
- **Orc: Blood Fury** has a 3 min cooldown instead of 2, so it's used once in a 180 sec fight instead of twice: -3 to -7 for every Orc.
- **Gnome: Eureka!** lasts 8 sec with no charge limit, instead of 3 charges in 15 sec: +1 to +2.5 for most specs, slightly negative for the daggers rogues.
- **Moonkin Aura / Leader of the Pack** split back to spell crit / melee and ranged crit. No preset moved: every non-warlock preset has both, and the warlock demons don't take player raid buffs, so Demonology Human is unchanged to the decimal.
- **Best-race changes:**
  - Fury: Undead to Orc
  - Rogue Combat Swords: Undead to Orc
  - Frostfire, Fire, Shadow, Destruction Fire: Undead or Orc to Gnome
  - Frost: Undead to High Order (0.1 over Undead, which is noise)
  - Elemental: Orc to Windshaper
- **Rank changes:** Frost passes Arms for 10th, and Balance passes Enhancement for 17th.
- **Top over bottom is 1.76×** (was 1.73×). The rogues lost the most, because Undead was their best race and they hit fastest.
- Touch of the Grave is now a flat once-per-25-sec drain whatever the spec's hit rate, so Undead no longer favours fast hitters. That's a game change, not a fairness issue.

</details>

<details><summary>Previous §1 (e1d2db0, 01:40)</summary>

### 1 (superseded 08:00) Cross-spec table (master e1d2db0: pushback fix, Fury 2H Rage thresholds, shaman rotation pass 21; full re-sweep 2026-10-16 01:40 UTC)

Every class was rebuilt from master e1d2db0 and fully re-swept with the usual settings. All 19 Level 60 gear sets pass 1b and 5a. Most of the 40 commits are the Level 30 leveling mode, which doesn't touch the Level 60 presets. "Before" is the 0798c11 table (22:10).

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (0798c11) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1116.0 | Undead 1110.0 / Gnome | same | 1116.0 | +0.0% |
| 2 | Fury (dual wield) | Undead | 1065.7 | Undead / Human 1054.5 | same | 1065.5 | +0.0% |
| 3 | Marksmanship | Night Elf | 1020.7 | Orc 1015.5 / Night Elf | same | 1020.7 | +0.0% |
| 4 | Frostfire | Orc | 1011.9 | Orc / Gnome 1011.0 | same | 1011.9 | +0.0% |
| 5 | Destruction | Gnome | 988.7 | Orc 988.4 / Gnome | same | 988.7 | +0.0% |
| 6 | Affliction | Gnome | 983.4 | Orc 979.6 / Gnome | same | 983.4 | +0.0% |
| 7 | Retribution | Undead | 981.5 | Undead / Human 967.2 | same | 981.5 | +0.0% |
| 8 | Destruction Fire | Undead | 973.2 | Undead / Gnome 972.7 | same | 973.2 | +0.0% |
| 9 | Arms | Orc | 954.2 | Orc / Night Elf 942.1 | Undead 950.2 | 953.4 | +0.1% |
| 10 | Beast Mastery | Night Elf | 952.0 | Orc 948.6 / Night Elf | same | 952.0 | +0.0% |
| 11 | Frost | Undead | 949.1 | Undead / High Order 949.0 | same | 949.1 | +0.0% |
| 12 | Fury (two-hand) | Orc | 943.0 | Orc / Night Elf 923.6 | Undead 936.2 | 940.9 | +0.2% |
| 13 | Survival | Orc | 892.6 | Orc / Night Elf 881.8 | same | 892.6 | +0.0% |
| 14 | Fire | Undead | 889.6 | Undead / Gnome 885.2 | same | 889.6 | +0.0% |
| 15 | Shadow | Undead | 855.0 | Undead / Gnome 849.5 | same | 855.0 | +0.0% |
| 16 | Arcane | Undead | 843.9 | Undead / Gnome 833.0 | same | 843.9 | +0.0% |
| 17 | Enhancement | Orc | 834.6 | Orc / Dwarf 816.0 | same | 833.8 | +0.1% |
| 18 | Balance | Night Elf | 832.0 | Tauren 822.4 / Night Elf | same | 832.0 | +0.0% |
| 19 | Feral | Night Elf | 717.3 | Windshaper 713.7 / Night Elf | same | 717.3 | +0.0% |
| 20 | Smite | Undead | 716.4 | Undead / Gnome 707.4 | same | 716.4 | +0.0% |
| 21 | Rogue Combat Swords | Undead | 710.5 | Undead / Gnome 698.1 | same | 710.5 | +0.0% |
| 22 | Rogue Combat Daggers | Undead | 708.9 | Undead / Gnome 693.2 | same | 708.9 | +0.0% |
| 23 | Elemental | Orc | 694.8 | Orc / Dwarf 687.7 | same | 693.5 | +0.2% |
| 24 | Rogue Mutilate | Undead | 680.8 | Undead / Gnome 667.4 | same | 680.8 | +0.0% |
| 25 | Rogue Seal Fate Daggers | Undead | 674.9 | Undead / Gnome 662.8 | same | 674.9 | +0.0% |
| 26 | Rogue Seal Fate Swords | Undead | 673.3 | Undead / Gnome 662.0 | same | 673.3 | +0.0% |
| 27 | Rogue Subtlety | Undead | 644.4 | Undead / Night Elf 629.9 | same | 644.4 | +0.0% |

**What moved:**
- **Fury 2H +0.2%:** Bloodrage under 20 Rage and Heroic Strike from 85 (PR #48).
- **Ele +0.2% and Enh +0.1%:** shaman rotation pass 21, which pre-pulls Searing Totem, gates the shield and drops late totems.
- **Arms +0.1%:** pushback can no longer take Slam back past its start (9f8d7ec).
- Fury +0.0%.
- Every other spec is identical to the decimal, which confirms the core changes didn't leak into them.
- Ranks are unchanged.

Top over bottom is 1.73×.

**Raid damage hits (standards 24):** see the 0798c11 notes below. Slam pushback from the hits is correct per the client (SpellInterrupts 70338, InterruptFlags 15 on every Slam rank). The shaman hits deal no damage, so Lightning Bolt isn't pushed back. That's a small open asymmetry, under 0.5%, proposed to the orchestrator.

</details>

<details><summary>Previous §1 (0798c11, 22:10)</summary>

### 1 (superseded 01:40) Cross-spec table (master 0798c11: standards 24 raid damage hits; re-swept 2026-10-15 22:10 UTC)

Same settings as before: 180 sec, 5000 iterations, seed 1, every race, a racials-off column, the generic boss, Sunder Armor, standards 21 at the midpoint. Only the shamans and warriors were re-swept; every other build's code and request are unchanged since 5738729. "Before" is the 5738729 sweep (19:40).

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (5738729) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1116.0 | Undead 1110.0 / Gnome | same | 1116.0 | +0.0% |
| 2 | Fury (dual wield) | Undead | 1065.5 | Undead / Human 1054.6 | same | 1058.5 | +0.7% |
| 3 | Marksmanship | Night Elf | 1020.7 | Orc 1015.5 / Night Elf | same | 1020.7 | +0.0% |
| 4 | Frostfire | Orc | 1011.9 | Orc / Gnome 1011.0 | same | 1011.9 | +0.0% |
| 5 | Destruction | Gnome | 988.7 | Orc 988.4 / Gnome | same | 988.7 | +0.0% |
| 6 | Affliction | Gnome | 983.4 | Orc 979.6 / Gnome | same | 983.4 | +0.0% |
| 7 | Retribution | Undead | 981.5 | Undead / Human 967.2 | same | 981.5 | +0.0% |
| 8 | Destruction Fire | Undead | 973.2 | Undead / Gnome 972.7 | same | 973.2 | +0.0% |
| 9 | Arms | Orc | 953.4 | Orc / Night Elf 941.1 | Undead 949.4 | 950.8 | +0.3% |
| 10 | Beast Mastery | Night Elf | 952.0 | Orc 948.6 / Night Elf | same | 952.0 | +0.0% |
| 11 | Frost | Undead | 949.1 | Undead / High Order 949.0 | same | 949.1 | +0.0% |
| 12 | Fury (two-hand) | Orc | 940.9 | Orc / Night Elf 921.9 | Undead 934.6 | 925.7 | +1.6% |
| 13 | Survival | Orc | 892.6 | Orc / Night Elf 881.8 | same | 892.6 | +0.0% |
| 14 | Fire | Undead | 889.6 | Undead / Gnome 885.2 | same | 889.6 | +0.0% |
| 15 | Shadow | Undead | 855.0 | Undead / Gnome 849.5 | same | 855.0 | +0.0% |
| 16 | Arcane | Undead | 843.9 | Undead / Gnome 833.0 | same | 843.9 | +0.0% |
| 17 | Enhancement | Orc | 833.8 | Orc / Dwarf 815.4 | same | 825.9 | +1.0% |
| 18 | Balance | Night Elf | 832.0 | Tauren 822.4 / Night Elf | same | 832.0 | +0.0% |
| 19 | Feral | Night Elf | 717.3 | Windshaper 713.7 / Night Elf | same | 717.3 | +0.0% |
| 20 | Smite | Undead | 716.4 | Undead / Gnome 707.4 | same | 716.4 | +0.0% |
| 21 | Rogue Combat Swords | Undead | 710.5 | Undead / Gnome 698.1 | same | 710.5 | +0.0% |
| 22 | Rogue Combat Daggers | Undead | 708.9 | Undead / Gnome 693.2 | same | 708.9 | +0.0% |
| 23 | Elemental | Orc | 693.5 | Orc / Dwarf 686.4 | same | 691.2 | +0.3% |
| 24 | Rogue Mutilate | Undead | 680.8 | Undead / Gnome 667.4 | same | 680.8 | +0.0% |
| 25 | Rogue Seal Fate Daggers | Undead | 674.9 | Undead / Gnome 662.8 | same | 674.9 | +0.0% |
| 26 | Rogue Seal Fate Swords | Undead | 673.3 | Undead / Gnome 662.0 | same | 673.3 | +0.0% |
| 27 | Rogue Subtlety | Undead | 644.4 | Undead / Night Elf 629.9 | same | 644.4 | +0.0% |

**Standards 24 (raid damage hits).** Enhancement takes 2 raid hits a minute and Elemental 1, which keep Lightning Shield firing. Every DPS warrior takes 2 a minute at 1500 Fire damage each, through the normal damage-taken path: Rage, plus Forever's Enrage, which procs on any damage taken.
- Fury 2H gains the most (+1.6%), because Enrage uptime rises from about 3.5 to 23 sec.
- Fury +0.7%, Enhancement +1.0%, Arms +0.3%, Elemental +0.3%. These match the orchestrator's and the results auditor's numbers.
- **Rank changes:** Arms passes BM for 9th (a 1.4 DPS gap, at the edge of the noise) and Enhancement passes Balance for 17th.
- No other base preset has an effect that triggers on damage taken.
- **Arms gap explained (orchestrator, 22:00):** for Arms, real hits gain less than the same Rage added directly (+2.1 against +4.8 at 20k iterations). The cause is spell pushback on Slam. In the client (SpellInterrupts, build 70338), Slam ranks 1464, 8820, 11604 and 11605 have InterruptFlags 15, which includes the pushback bit, as Frostbolt, Shadow Bolt and Lightning Bolt do. So the pushback is correct. Since 6d7de7c, pushback can't take a cast back past its start, and Arms is at +3.1.
- **Open, small:** the shaman raid hits deal no damage (shields.go only fires the shield), so Elemental's Lightning Bolt is never pushed back, though the client allows it. This is under 0.5% at 1 hit a minute. I proposed that the shamans use the warriors' real spell hit.

Top over bottom is 1.73×.

</details>

<details><summary>Previous §1 (5738729, 19:40)</summary>

### 1 (superseded 22:10) Cross-spec table (master 5738729: standards 23 raid duties on top of 96c210d's rotation pass 19 and hunter ranged pass; re-swept 2026-10-15 19:40 UTC)

Built from master 96c210d for every class. Settings as before: 180 sec, 5000 iterations, seed 1, every race, a racials-off column, the generic boss, Sunder Armor, standards 21 at the midpoint. All 19 Level 60 gear sets pass 1b and 5a. "Before" is the b5b1f44 sweep (09:40). The numbers match the results auditor's pass 28.

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (96c210d) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1116.0 | Undead 1110.0 / Gnome | same | 1116.0 | +0.0% |
| 2 | Fury (dual wield) | Undead | 1058.5 | Undead / Human 1047.7 | same | 1061.7 | -0.3% |
| 3 | Marksmanship | Night Elf | 1020.7 | Orc 1015.5 / Night Elf | same | 1020.7 | +0.0% |
| 4 | Frostfire | Orc | 1011.9 | Orc / Gnome 1011.0 | same | 1011.9 | +0.0% |
| 5 | Destruction | Gnome | 988.7 | Orc 988.4 / Gnome | same | 988.7 | +0.0% |
| 6 | Affliction | Gnome | 983.4 | Orc 979.6 / Gnome | same | 983.4 | +0.0% |
| 7 | Retribution | Undead | 981.5 | Undead / Human 967.2 | same | 981.5 | +0.0% |
| 8 | Destruction Fire | Undead | 973.2 | Undead / Gnome 972.7 | same | 973.2 | +0.0% |
| 9 | Beast Mastery | Night Elf | 952.0 | Orc 948.6 / Night Elf | same | 952.0 | +0.0% |
| 10 | Arms | Orc | 950.8 | Orc / Night Elf 938.0 | Undead 947.0 | 952.0 | -0.1% |
| 11 | Frost | Undead | 949.1 | Undead / High Order 949.0 | same | 949.1 | +0.0% |
| 12 | Fury (two-hand) | Orc | 925.7 | Orc / Night Elf 906.4 | Undead 919.0 | 926.2 | -0.1% |
| 13 | Survival | Orc | 892.6 | Orc / Night Elf 881.8 | same | 892.6 | +0.0% |
| 14 | Fire | Undead | 889.6 | Undead / Gnome 885.2 | same | 889.6 | +0.0% |
| 15 | Shadow | Undead | 855.0 | Undead / Gnome 849.5 | same | 855.0 | +0.0% |
| 16 | Arcane | Undead | 843.9 | Undead / Gnome 833.0 | same | 843.9 | +0.0% |
| 17 | Balance | Night Elf | 832.0 | Tauren 822.4 / Night Elf | same | 832.0 | +0.0% |
| 18 | Enhancement | Orc | 825.9 | Orc / Dwarf 807.4 | same | 825.9 | +0.0% |
| 19 | Feral | Night Elf | 717.3 | Windshaper 713.7 / Night Elf | same | 717.3 | +0.0% |
| 20 | Smite | Undead | 716.4 | Undead / Gnome 707.4 | same | 716.4 | +0.0% |
| 21 | Rogue Combat Swords | Undead | 710.5 | Undead / Gnome 698.1 | same | 710.5 | +0.0% |
| 22 | Rogue Combat Daggers | Undead | 708.9 | Undead / Gnome 693.2 | same | 708.9 | +0.0% |
| 23 | Elemental | Orc | 691.2 | Orc / Dwarf 684.0 | same | 691.2 | +0.0% |
| 24 | Rogue Mutilate | Undead | 680.8 | Undead / Gnome 667.4 | same | 680.8 | +0.0% |
| 25 | Rogue Seal Fate Daggers | Undead | 674.9 | Undead / Gnome 662.8 | same | 674.9 | +0.0% |
| 26 | Rogue Seal Fate Swords | Undead | 673.3 | Undead / Gnome 662.0 | same | 673.3 | +0.0% |
| 27 | Rogue Subtlety | Undead | 644.4 | Undead / Night Elf 629.9 | same | 644.4 | +0.0% |

**Standards 23 (raid actions that cost a GCD, 19:40 re-sweep).** Base presets cast no raid duty; the raid's buffs and debuffs cover them, and separate duty presets carry the cost.
- Every warrior spends one Sunder Armor GCD at the pull: Fury -0.3%, Arms -0.1%, Fury 2H -0.1%.
- Every other base preset is unchanged.
- Melee raid buffs are consistent: Windfury and Strength of Earth for every melee build except Enhancement, which drops its own totems. Every build gets the full debuff set. Hunter's Mark is in the hunters' requests only, because it raises ranged AP.

**What moved at 13:20 (96c210d against b5b1f44):**
- **Frostfire +2.4%** (pass 19: Ice Lance ahead of Scorch, end-of-fight instants). Its best race is now Orc, 0.9 over Gnome, which is noise.
- **Arms +2.1%** (Rend scales with attack power, warrior assumption default).
- **BM -2.9%** (hunter assumption default: pet hit scales with speed). MM +0.1% and SV -0.4% after the ranged weapon pass, which is the standards 22 method on bows, guns and crossbows.
- **The other casters +0.3 to +0.8%** from the end-of-fight instants.

Top over bottom is 1.73×.

**Coverage of rotation pass 19 (the end-of-fight instants):** it changed Balance, Elemental, Enhancement, all four mages, all four warlocks, Smite and Fury 2H. It left Shadow, Ret, Feral, the rogues, Fury, Arms and the hunters untouched.
- For melee and energy users there is little to gain: no cast is lost at the pull's end.
- **Shadow** was tested in the rotation auditor's pass 19b. No end-of-fight instant gains 1 DPS or more (best +0.5, a Mind Blast time-left gate), so Shadow's +0.0% is real, not an omission.
- **The hunters** (Aimed / Sniper Shot past the end) are being tested now, at the orchestrator's request.

**Creature-type talents on the generic boss (user ruling 14:48, standards 19).** Hunter Improved Tracking stays on, because a hunter can track every boss type except Mechanical. That is worth +4.1% to MM and +4.0% to SV. Rogue Murder (Humanoid and Giant) and Ret's Demon and Undead bonuses stay off; those builds get them through their "vs <type>" presets. Measured by switching the boss type: rogues +4.00% vs Humanoid, MM/SV -4.08/-3.99% vs Mechanical, Ret +5.98% vs Demon. This is an inherent asymmetry by ruling (see §4), not an open item.

</details>

<details><summary>Previous §1 (b5b1f44, 09:40)</summary>

### 1 (superseded 13:20) Cross-spec table (master b5b1f44: Phase 1 weapon speeds, caster defaults; re-swept 2026-10-15 09:40 UTC)

Built from master b5b1f44 for every class. Settings as before: 180 sec, 5000 iterations, seed 1, every race, a racials-off column, the generic boss, Sunder Armor, standards 21 at the midpoint.
- **Weapons (standards 22, revised).** Each preset takes the best weapon type among Era Phase 1 BiS speeds. Fury 2H shares the Arms set again (3.7 s axe). Mutilate and Seal Fate Swords keep their own sets.
- **Caster defaults (2026-10-15).** The caster assumptions' expected values are now the sim defaults:
  - Ice Lance at 0.429 on every Fingers of Frost charge.
  - The mage staff imbue scrolls.
  - Lightning Bolt ×1.12.
- **Audit.** All 19 Level 60 gear sets pass 1b and 5a.
- **"Before"** is the 4dfaa63 sweep (07:40).

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (4dfaa63) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1111.7 | Undead 1105.7 / Gnome | same | 1111.7 | +0.0% |
| 2 | Fury (dual wield) | Undead | 1061.7 | Undead / Human 1051.6 | same | 1067.0 | -0.5% |
| 3 | Marksmanship | Night Elf | 1019.7 | Orc 1014.6 / Night Elf | same | 1019.7 | +0.0% |
| 4 | Frostfire | Undead | 988.2 | Undead / Gnome 987.7 | same | 904.2 | +9.3% |
| 5 | Destruction | Gnome | 985.2 | Orc 984.9 / Gnome | same | 985.2 | +0.0% |
| 6 | Retribution | Undead | 981.5 | Undead / Human 967.2 | same | 981.5 | +0.0% |
| 7 | Beast Mastery | Night Elf | 980.8 | Orc 975.8 / Night Elf | same | 980.8 | +0.0% |
| 8 | Affliction | Gnome | 979.5 | Orc 975.7 / Gnome | same | 979.5 | +0.0% |
| 9 | Destruction Fire | Undead | 970.2 | Undead / Gnome 969.7 | same | 970.2 | +0.0% |
| 10 | Frost | High Order | 942.2 | Undead 942.2 / High Order | same | 745.7 | +26.4% |
| 11 | Arms | Orc | 932.5 | Orc / Night Elf 920.9 | Undead 930.4 | 951.6 | -2.0% |
| 12 | Fury (two-hand) | Orc | 924.2 | Orc / Night Elf 905.5 | Undead 917.7 | 924.2 | +0.0% |
| 13 | Survival | Orc | 896.0 | Orc / Night Elf 886.0 | same | 900.0 | -0.4% |
| 14 | Fire | Undead | 884.9 | Undead / Gnome 880.9 | same | 874.0 | +1.2% |
| 15 | Shadow | Undead | 855.0 | Undead / Gnome 849.5 | same | 855.0 | +0.0% |
| 16 | Arcane | Undead | 836.9 | Undead / Gnome 826.0 | same | 830.6 | +0.8% |
| 17 | Balance | Night Elf | 829.1 | Tauren 819.6 / Night Elf | same | 827.4 | +0.2% |
| 18 | Enhancement | Orc | 823.8 | Orc / Dwarf 805.4 | same | 825.1 | -0.2% |
| 19 | Feral | Night Elf | 717.3 | Windshaper 713.7 / Night Elf | same | 717.3 | +0.0% |
| 20 | Smite | Undead | 713.4 | Undead / Gnome 704.9 | same | 706.7 | +0.9% |
| 21 | Rogue Combat Swords | Undead | 710.5 | Undead / Gnome 698.1 | same | 724.4 | -1.9% |
| 22 | Rogue Combat Daggers | Undead | 708.9 | Undead / Gnome 693.2 | same | 728.7 | -2.7% |
| 23 | Elemental | Orc | 687.7 | Orc / Dwarf 680.6 | same | 674.6 | +1.9% |
| 24 | Rogue Mutilate | Undead | 680.8 | Undead / Gnome 667.4 | same | 696.3 | -2.2% |
| 25 | Rogue Seal Fate Daggers | Undead | 674.9 | Undead / Gnome 662.8 | same | 695.7 | -3.0% |
| 26 | Rogue Seal Fate Swords | Undead | 673.3 | Undead / Gnome 662.0 | same | 678.6 | -0.8% |
| 27 | Rogue Subtlety | Undead | 644.4 | Undead / Night Elf 629.9 | same | 668.1 | -3.5% |

**What moved:**
- **Frost +26.4% and Frostfire +9.3%.** Ice Lance is now cast on every Fingers of Frost charge at a 0.429 coefficient, plus the staff scrolls. Frost goes from 18th to 10th, and Frostfire to 4th.
- **Smaller caster gains.** Fire +1.2% and Arcane +0.8% (scrolls), Elemental +1.9% (Lightning Bolt ×1.12), Smite +0.9%.
- **Melee lose some of the 4dfaa63 gain to the Phase 1 speed cap.**
  - Rogues -0.8 to -3.5%; the dagger rogues are back on Perdition's Blade and Core Hound Tooth.
  - Arms -2.0%, Fury -0.5%.
  - Rogues are now the lowest melee class and hold 6 of the bottom 8 places.

Top over bottom is 1.73× (Demonology 1112 over Subtlety 644).

**Weapon racials:** the Orc axe still decides Arms (worth 2.1 now; Undead 930.4 with racials off) and Fury 2H (worth 6.5; Undead 917.7). Survival is Orc either way.

**Standing gap from 07:40:** the ranged hunters' bows and guns were not searched under standards 22. Raised with the orchestrator.


</details>

<details><summary>Previous §1 (4dfaa63, 07:40)</summary>

### 1 (superseded 09:40) Cross-spec table (master 4dfaa63: standards 22 weapons by best type and speed, raid-buff ranks by level; re-swept 2026-10-15 07:40 UTC)

The presets now live on master, and the harness builds every class from master 4dfaa63. Settings as before: 180 sec, 5000 iterations, seed 1, every race, a racials-off column, the generic boss and Sunder Armor, and standards 21 at the midpoint.
- Three builds now carry their own gear sets with their own weapons: Fury 2H, Mutilate and Seal Fate Swords.
- All 20 Level 60 gear sets pass 1b and 5a. Standards 22 keeps every weapon's stats and DPS; only type, speed and spread change.
- "Before" is the Oct 14 table, with the warrior at 55979ee.

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (Oct 14) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1111.7 | Undead 1105.7 / Gnome | same | 1111.7 | +0.0% |
| 2 | Fury (dual wield) | Undead | 1067.0 | Undead / Human 1054.2 | same | 1038.3 | +2.8% |
| 3 | Marksmanship | Night Elf | 1019.7 | Orc 1014.6 / Night Elf | same | 1019.7 | +0.0% |
| 4 | Destruction | Gnome | 985.2 | Orc 984.9 / Gnome | same | 985.2 | +0.0% |
| 5 | Retribution | Undead | 981.5 | Undead / Human 967.2 | same | 981.1 | +0.0% |
| 6 | Beast Mastery | Night Elf | 980.8 | Orc 975.8 / Night Elf | same | 980.8 | +0.0% |
| 7 | Affliction | Gnome | 979.5 | Orc 975.7 / Gnome | same | 979.5 | +0.0% |
| 8 | Destruction Fire | Undead | 970.2 | Undead / Gnome 969.7 | same | 969.8 | +0.0% |
| 9 | Arms | Orc | 951.6 | Orc / Night Elf 939.6 | Undead 946.5 | 915.2 | +4.0% |
| 10 | Fury (two-hand) | Orc | 924.2 | Orc / Night Elf 905.5 | Undead 917.7 | 900.8 | +2.6% |
| 11 | Frostfire | Undead | 904.2 | Undead / High Order 902.4 | same | 903.9 | +0.0% |
| 12 | Survival | Orc | 900.0 | Orc / Night Elf 891.0 | same | 859.0 | +4.8% |
| 13 | Fire | Undead | 874.0 | Undead / Gnome 870.0 | same | 873.6 | +0.0% |
| 14 | Shadow | Undead | 855.0 | Undead / Gnome 849.5 | same | 854.6 | +0.0% |
| 15 | Arcane | Undead | 830.6 | Undead / Gnome 819.8 | same | 830.1 | +0.1% |
| 16 | Balance | Night Elf | 827.4 | Tauren 817.9 / Night Elf | same | 827.4 | +0.0% |
| 17 | Enhancement | Orc | 825.1 | Orc / Dwarf 805.8 | same | 813.2 | +1.5% |
| 18 | Frost | High Order | 745.7 | Undead 744.4 / High Order | same | 745.7 | +0.0% |
| 19 | Rogue Combat Daggers | Undead | 728.7 | Undead / Gnome 713.5 | same | 708.4 | +2.9% |
| 20 | Rogue Combat Swords | Undead | 724.4 | Undead / Human 716.0 | same | 679.5 | +6.6% |
| 21 | Feral | Night Elf | 717.3 | Windshaper 713.7 / Night Elf | same | 717.3 | +0.0% |
| 22 | Smite | Undead | 706.7 | Undead / Gnome 699.3 | same | 706.4 | +0.0% |
| 23 | Rogue Mutilate | Undead | 696.3 | Undead / Gnome 682.2 | same | 669.8 | +4.0% |
| 24 | Rogue Seal Fate Daggers | Undead | 695.7 | Undead / Gnome 683.5 | same | 674.5 | +3.1% |
| 25 | Rogue Seal Fate Swords | Undead | 678.6 | Undead / Gnome 666.2 | same | 644.3 | +5.3% |
| 26 | Elemental | Orc | 674.6 | Orc / Dwarf 667.4 | same | 674.6 | +0.0% |
| 27 | Rogue Subtlety | Undead | 668.1 | Undead / Gnome 653.3 | same | 643.9 | +3.8% |

**What moved (standards 22, every melee build on the weapon type and speed it sims best with):**

| Build | Change |
|---|---|
| Rogue Combat Swords (2.9 s axe + 1.5 s sword) | +6.6% |
| Rogue Seal Fate Swords | +5.3% |
| Survival (Orc axe) | +4.8% |
| Arms (4.0 s axe) | +4.0% |
| Rogue Mutilate | +4.0% |
| Rogue Subtlety | +3.8% |
| Rogue Seal Fate Daggers | +3.1% |
| Rogue Combat Daggers | +2.9% |
| Fury (2.9 s axe + 1.5 s sword) | +2.8% |
| Fury 2H (3.7 s axe) | +2.6% |
| Enhancement | +1.5% |

Casters, ranged hunters, Feral and Ret move under 0.1%; the raid-buff ranks by level change almost nothing. Top over bottom narrowed to 1.66× (Demonology 1112 over Subtlety 668), from 1.73×.

**Weapon racials (fairness P6, the racials-off column):**
- The Orc axe now decides Arms (Orc 951.6, or Undead 946.5 with racials off, worth 5.1) and Fury 2H (Orc 924.2, or Undead 917.7, worth 6.5).
- Survival is Orc either way.
- Everything else is unchanged.

**Open fairness gap (sent to the orchestrator):** standards 22 searched melee weapons only ("ranged weapon types and speeds were not part of this pass"). Every melee build gained 2.6 to 6.6% from its best weapon type and speed. The ranged hunters (BM, MM) still carry their Era Phase 1 bow or gun speed, and their ranged attacks scale with weapon speed the same way (Aimed Shot and Multi-Shot are per hit; auto-shots scale with swing damage). Until the ranged pass runs, BM and MM are the only weapon-scaling builds not on their best weapon.


</details>

<details><summary>Previous §1 (Oct 13-14, class branches)</summary>

### 1 (superseded 2026-10-15) Cross-spec table (standards 21 normalised Rage at the midpoint s = 0.5, master merges; re-swept 2026-10-13 20:05 UTC, warrior 55979ee at 06:40 on 2026-10-14, rotation audit 14)

Settings as before: 180 sec, 5000 iterations, seed 1, every race, a racials-off column, the generic boss and Sunder Armor. The heads are warrior cb78cf9, paladin e62dc38, hunter 5a9a22f, rogue 353fb47, priest 3b0c87f, shaman 8f15071, mage 00272b3, warlock bbbad89, druid cbfb9ae, every one rebuilt. Every set passes 1b and 5a. "Before" is the 18:45 sweep.

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (18:45, Classic Rage) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1111.7 | Undead 1105.2 / Gnome | same | 1111.7 | +0.0% |
| 2 | Fury (dual wield) | Undead | 1038.3 | Undead / Human 1031.0 | same | 1052.9 | -1.4% |
| 3 | Marksmanship | Night Elf | 1019.7 | Orc 1014.6 / Night Elf | same | 1019.7 | +0.0% |
| 4 | Destruction | Gnome | 985.2 | Orc 984.9 / Gnome | same | 985.2 | +0.0% |
| 5 | Retribution | Undead | 981.1 | Undead / Human 967.2 | same | 981.1 | +0.0% |
| 6 | Beast Mastery | Night Elf | 980.8 | Orc 975.8 / Night Elf | same | 980.8 | +0.0% |
| 7 | Affliction | Gnome | 979.5 | Orc 975.7 / Gnome | same | 979.5 | +0.0% |
| 8 | Destruction Fire | Undead | 969.8 | Undead / Gnome 969.7 | same | 969.8 | +0.0% |
| 9 | Arms | Undead | 915.2 | Undead / Human 912.3 | same | 922.4 | -0.8% |
| 10 | Frostfire | Undead | 903.9 | Undead / High Order 902.4 | same | 903.9 | +0.0% |
| 11 | Fury (two-hand) | Undead | 900.8 | Undead / Human 898.7 | same | 919.3 | -2.0% |
| 12 | Fire | Undead | 873.6 | Undead / Gnome 870.0 | same | 873.6 | +0.0% |
| 13 | Survival | Human | 859.0 | Orc 853.0 / Human | Orc 853.0 | 859.0 | +0.0% |
| 14 | Shadow | Undead | 854.6 | Undead / Gnome 849.5 | same | 854.6 | +0.0% |
| 15 | Arcane | Undead | 830.1 | Undead / Gnome 819.8 | same | 830.1 | +0.0% |
| 16 | Balance | Night Elf | 827.4 | Tauren 817.9 / Night Elf | same | 827.4 | +0.0% |
| 17 | Enhancement | Orc | 813.2 | Orc / Dwarf 794.3 | same | 813.2 | +0.0% |
| 18 | Frost | High Order | 745.7 | Undead 744.1 / High Order | same | 745.7 | +0.0% |
| 19 | Feral | Night Elf | 717.3 | Windshaper 713.7 / Night Elf | same | 717.3 | +0.0% |
| 20 | Rogue Combat Daggers | Undead | 708.4 | Undead / Gnome 693.2 | same | 708.4 | +0.0% |
| 21 | Smite | Undead | 706.4 | Undead / Gnome 699.3 | same | 706.4 | +0.0% |
| 22 | Rogue Combat Swords | Undead | 679.5 | Undead / Human 674.4 | same | 679.5 | +0.0% |
| 23 | Elemental | Orc | 674.6 | Orc / Dwarf 667.4 | same | 674.6 | +0.0% |
| 24 | Rogue Seal Fate Daggers | Undead | 674.5 | Undead / Gnome 662.8 | same | 674.5 | +0.0% |
| 25 | Rogue Mutilate | Undead | 669.8 | Undead / Gnome 655.3 | same | 669.8 | +0.0% |
| 26 | Rogue Seal Fate Swords | Undead | 644.3 | Undead / Night Elf 635.9 | same | 644.3 | +0.0% |
| 27 | Rogue Subtlety | Undead | 643.9 | Undead / Night Elf 629.9 | same | 643.9 | +0.0% |

**What moved:**
- **Only the warriors.** At standards 21's midpoint (s = 0.5, k × 1.5 at level 60, the presets' default since 38b94b4): Fury -1.4% (1053 → 1038, still #2), Arms -0.8% (922 → 915), Fury 2H -2.0% (919 → 901, 14 below Arms). That includes the rotation re-search under the new Rage model (55979ee, +1 / +6 / +11).
- **Every other build is unchanged.** The master merges (item DB update, UI) don't touch the presets.

Top over bottom is still 1.73× (Demonology 1112 over Subtlety 644).

**The warrior rows are the least settled in the table.** k comes from beta testers' level 30 measurements, and no level 60 figure exists (see §3). The results auditor's sensitivity, at 20k with the current rotations, so lower bounds:

| Build | k as ruled | k × 1.5 | k × 2.07 (Classic's level 60 income) |
|---|---|---|---|
| Fury | 950.6 | 1030.4 | 1070.5 |
| Fury 2H | 835.0 | 891.2 | 927.8 |
| Arms | 861.4 | 908.8 | 937.8 |

The worker's settings: s = 0 (Low) is k as measured, s = 0.5 (Expected, shipped) is ×1.5, and s = 1 (High, about ×2) gives 1076.1 / 924.2 / 935.0. Fury's rank runs from 8th (Low) to 2nd (Expected and High).


</details>

<details><summary>Previous §1 (18:45, shared consumes merge)</summary>

### 1 (superseded 20:05) Cross-spec table (shared consumes merge; re-swept 2026-10-13 18:45 UTC)

Settings as before: 180 sec, 5000 iterations, seed 1, every race, a racials-off column, the generic boss and Sunder Armor. The heads are warrior aa11a4f, paladin d24d523, hunter 4283a29, rogue 1e012bb, priest 81d751f, shaman 8529c90, mage fd1ba1f, warlock bfb8961, druid 43cf0dd, every one rebuilt. Every set passes 1b and 5a. "Before" is the 17:15 sweep.

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (17:15) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1111.7 | Undead 1105.2 / Gnome | same | 1111.7 | +0.0% |
| 2 | Fury (dual wield) | Undead | 1052.9 | Undead / Human 1050.4 | same | 1055.3 | -0.2% |
| 3 | Marksmanship | Night Elf | 1019.7 | Orc 1014.6 / Night Elf | same | 1007.3 | +1.2% |
| 4 | Destruction | Gnome | 985.2 | Orc 984.9 / Gnome | same | 985.2 | +0.0% |
| 5 | Retribution | Undead | 981.1 | Undead / Human 967.2 | same | 980.9 | +0.0% |
| 6 | Beast Mastery | Night Elf | 980.8 | Orc 975.8 / Night Elf | same | 970.2 | +1.1% |
| 7 | Affliction | Gnome | 979.5 | Orc 975.7 / Gnome | same | 979.5 | +0.0% |
| 8 | Destruction Fire | Undead | 969.8 | Undead / Gnome 969.7 | same | 969.8 | +0.0% |
| 9 | Arms | Undead | 922.4 | Undead / Human 921.8 | same | 922.9 | -0.1% |
| 10 | Fury (two-hand) | Undead | 919.3 | Undead / Human 917.0 | same | 918.0 | +0.1% |
| 11 | Frostfire | Undead | 903.9 | Undead / High Order 902.4 | same | 903.9 | +0.0% |
| 12 | Fire | Undead | 873.6 | Undead / Gnome 870.0 | same | 873.6 | +0.0% |
| 13 | Survival | Human | 859.0 | Orc 853.0 / Human | Orc 853.0 | 858.4 | +0.1% |
| 14 | Shadow | Undead | 854.6 | Undead / Gnome 849.5 | same | 854.6 | +0.0% |
| 15 | Arcane | Undead | 830.1 | Undead / Gnome 819.8 | same | 830.1 | +0.0% |
| 16 | Balance | Night Elf | 827.4 | Tauren 817.9 / Night Elf | same | 827.4 | +0.0% |
| 17 | Enhancement | Orc | 813.2 | Orc / Dwarf 794.3 | same | 814.1 | -0.1% |
| 18 | Frost | High Order | 745.7 | Undead 744.1 / High Order | same | 745.7 | +0.0% |
| 19 | Feral | Night Elf | 717.3 | Windshaper 713.7 / Night Elf | same | 709.7 | +1.1% |
| 20 | Rogue Combat Daggers | Undead | 708.4 | Undead / Gnome 693.2 | same | 708.2 | +0.0% |
| 21 | Smite | Undead | 706.4 | Undead / Gnome 699.3 | same | 706.4 | +0.0% |
| 22 | Rogue Combat Swords | Undead | 679.5 | Undead / Human 674.4 | same | 679.2 | +0.0% |
| 23 | Elemental | Orc | 674.6 | Orc / Dwarf 667.4 | same | 674.6 | +0.0% |
| 24 | Rogue Seal Fate Daggers | Undead | 674.5 | Undead / Gnome 662.8 | same | 674.1 | +0.1% |
| 25 | Rogue Mutilate | Undead | 669.8 | Undead / Gnome 655.3 | same | 669.5 | +0.0% |
| 26 | Rogue Seal Fate Swords | Undead | 644.3 | Undead / Night Elf 635.9 | same | 644.2 | +0.0% |
| 27 | Rogue Subtlety | Undead | 643.9 | Undead / Night Elf 629.9 | same | 643.6 | +0.0% |

**What moved:**
- **Hunters +1.1 to +1.2%:** Trueshot Aura and Hunter's Mark now use the strongest learned rank (8a).
- **Feral +1.1%:** Windfury Totem in place of Grace of Air, and a Dense Weightstone.
- **Everything else is within ±0.2%.** Forever's oil and Dragonbreath Chili proc damage moved almost nothing.

Top over bottom is still 1.73×.

**Cross-class checks this round, all consistent:**
- **Windfury Totem** (totemWeaponBuff) goes to every melee build in a shaman group: warriors, Ret, rogues, Feral and Survival. Ranged hunters get Grace of Air, and Enh drops its own.
- **Consumables:**
  - Every melee build runs Dragonbreath Chili.
  - Casters run Wizard Oil.
  - Ret and Enh run Forever's Shadow Oil.
  - Physical builds run sharpening stones or a weightstone.
- **New enchants:** Forever's Bracer Superior Agility (+9 Agi) and 2H Strength (+25 Str) are in the DB but in no preset, because their launch availability is unconfirmed (c613ab1). That holds the same for every class. If they're confirmed, rogues and hunters should move to Superior Agility. They are on Superior Strength (+9) and Bracer Agility (+5) today.


</details>

<details><summary>Previous §1 (17:15, class-changes round)</summary>

### 1 (superseded 18:45) Cross-spec table (Forever class-changes round; re-swept 2026-10-13 17:15 UTC)

Settings as before: 180 sec, 5000 iterations, seed 1, every race, plus a racials-off column. The boss is the generic level 63 boss (standards 19), with Sunder Armor (standards 18). The heads are warrior 41e1492, paladin 57c04a8, hunter 302fe1f, rogue eb66a70, priest 333cbeb, shaman 651dcb6, mage aa4e5c4, warlock 55f3d3f, druid 72f3927, every one rebuilt. Every set still passes 1b and 5a. "Before" is the 14:50 sweep.

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (14:50) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1111.7 | Undead 1105.2 / Gnome | same | 1111.7 | +0.0% |
| 2 | Fury (dual wield) | Undead | 1055.3 | Undead / Human 1049.0 | same | 1023.6 | +3.1% |
| 3 | Marksmanship | Night Elf | 1007.3 | Orc 1002.3 / Night Elf | same | 1006.9 | +0.0% |
| 4 | Destruction | Gnome | 985.2 | Orc 984.9 / Gnome | same | 985.2 | +0.0% |
| 5 | Retribution | Undead | 980.9 | Undead / Human 966.0 | same | 970.8 | +1.0% |
| 6 | Affliction | Gnome | 979.5 | Orc 975.7 / Gnome | same | 979.5 | +0.0% |
| 7 | Beast Mastery | Night Elf | 970.2 | Orc 965.2 / Night Elf | same | 969.5 | +0.1% |
| 8 | Destruction Fire | Undead | 969.8 | Undead / Gnome 969.7 | same | 969.8 | +0.0% |
| 9 | Arms | Undead | 922.9 | Undead / Human 920.7 | same | 890.1 | +3.7% |
| 10 | Fury (two-hand) | Human | 918.0 | Undead 918.0 / Human | Undead 918.0 | 886.8 | +3.5% |
| 11 | Frostfire | Undead | 903.9 | Undead / High Order 902.4 | same | 903.9 | +0.0% |
| 12 | Fire | Undead | 873.6 | Undead / Gnome 870.0 | same | 873.6 | +0.0% |
| 13 | Survival | Human | 858.4 | Orc 851.7 / Human | Night Elf 851.7 | 857.9 | +0.1% |
| 14 | Shadow | Undead | 854.6 | Undead / Gnome 849.5 | same | 853.9 | +0.1% |
| 15 | Arcane | Undead | 830.1 | Undead / Gnome 819.8 | same | 830.1 | +0.0% |
| 16 | Balance | Night Elf | 827.4 | Tauren 817.9 / Night Elf | same | 826.0 | +0.2% |
| 17 | Enhancement | Orc | 814.1 | Orc / Dwarf 795.5 | same | 802.3 | +1.5% |
| 18 | Frost | High Order | 745.7 | Undead 744.1 / High Order | same | 737.5 | +1.1% |
| 19 | Feral | Night Elf | 709.7 | Windshaper 707.1 / Night Elf | same | 688.5 | +3.1% |
| 20 | Rogue Combat Daggers | Undead | 708.2 | Undead / Night Elf 693.0 | same | 686.8 | +3.1% |
| 21 | Smite | Undead | 706.4 | Undead / Gnome 699.3 | same | 676.9 | +4.4% |
| 22 | Rogue Combat Swords | Undead | 679.2 | Undead / Human 674.7 | same | 664.9 | +2.2% |
| 23 | Elemental | Orc | 674.6 | Orc / Dwarf 667.4 | same | 674.6 | +0.0% |
| 24 | Rogue Seal Fate Daggers | Undead | 674.1 | Undead / Gnome 662.0 | same | 657.0 | +2.6% |
| 25 | Rogue Mutilate | Undead | 669.5 | Undead / Gnome 655.9 | same | 654.4 | +2.3% |
| 26 | Rogue Seal Fate Swords | Undead | 644.2 | Undead / Gnome 636.0 | same | 631.6 | +2.0% |
| 27 | Rogue Subtlety | Undead | 643.6 | Undead / Gnome 629.4 | same | 627.7 | +2.5% |

**What moved, all modelling or preset changes in the class branches:**
- **Warriors +3.1 to +3.7%:** Deep Wounds now rolls and no longer scales with AP.
- **Feral +3.1%:** Tiger's Fury removed, under Forever's shapeshift rules.
- **Rogues +2.0 to +3.1%:** generic-boss talents (Lethality in place of Murder).
- **Enh +1.5%:** Elixir of Cunning + Elixir of Ferocity, and Rage of the Farseer held for the trinket.
- **Ret +1.0%:** Elixir of Cunning.
- **Frost +1.1%.**
- **Smite +4.4%:** see the harness note below.

**Spread:** top over bottom is 1.73× (Demonology 1112 over Subtlety 644). It was 1.77×.

**Consumables audit for this round:**
- **Elixir of Cunning** (+25 Agi +25 Int, agility slot) and **Elixir of Ferocity** (+18 Str +18 Agi, strength slot) are Forever elixirs, on every class branch.
  - Every physical and hybrid build runs Ferocity.
  - Ret and Enh run Cunning in place of Mongoose. That keeps them at zero or one crit elixir, which item 10 allows.
  - The warriors, rogues, hunters and Feral keep Mongoose. That is their single crit elixir, and it sims ahead for them.
- **Elixir of Greater Firepower** is Forever's Elixir of Holy Power (+40 Holy). Ret and Smite use it, and the fire casters correctly use the plain Elixir of Firepower.

**Harness note:** until this sweep my harness gave Smite the Shadow consumes (mksettings has no Smite consumes). The Smite preset ships ConsumesSmite: Major Mana Potion, the Holy Power elixir, Greater Mageblood and Greater Spirit. Earlier Smite rows were about 2 to 4% low. With the fix, Smite matches the priest's published 706.5.

**Races:** Fury 2H is a tie between Human and Undead (918.0 each). Frost is High Order by 1.6.


</details>

<details><summary>Previous §1 (14:50, generic boss)</summary>

### 1 (superseded 17:15) Cross-spec table (standards 18 Sunder Armor + 19 generic boss; re-swept 2026-10-13 14:50 UTC)

Settings as before: 180 sec, 5000 iterations, seed 1, every race, plus a racials-off column. The boss is a level 63 creature of type Unknown, with Sunder Armor in place of Expose Armor. The heads are warrior 9d3bc19, paladin bdcd086, hunter 0547d52, rogue 06b242f, priest 5000439, shaman 30c25c7, mage fcd1d79, warlock 02fbb87, druid 7600a33, every one rebuilt. The audit is unchanged: every set still passes 1b and 5a. "Before" is the 10:30 bundled-round sweep (Humanoid boss, Expose Armor), with Enh from its 12:25 re-sweep.

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (Humanoid, Expose) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1111.7 | Undead 1105.2 / Gnome | same | 1110.3 | +0.1% |
| 2 | Fury (dual wield) | Undead | 1023.6 | Undead / Human 1016.0 | same | 1020.2 | +0.3% |
| 3 | Marksmanship | Night Elf | 1006.9 | Orc 1001.9 / Night Elf | same | 1006.9 | +0.0% |
| 4 | Destruction | Gnome | 985.2 | Orc 984.9 / Gnome | same | 984.1 | +0.1% |
| 5 | Affliction | Gnome | 979.5 | Orc 975.7 / Gnome | same | 978.4 | +0.1% |
| 6 | Retribution | Undead | 970.8 | Undead / Human 956.2 | same | 967.3 | +0.4% |
| 7 | Destruction Fire | Undead | 969.8 | Undead / Gnome 969.7 | same | 968.6 | +0.1% |
| 8 | Beast Mastery | Night Elf | 969.5 | Orc 964.5 / Night Elf | same | 969.5 | +0.0% |
| 9 | Frostfire | Undead | 903.9 | Undead / High Order 902.4 | same | 900.3 | +0.4% |
| 10 | Arms | Undead | 890.1 | Undead / Human 885.3 | same | 890.1 | +0.0% |
| 11 | Fury (two-hand) | Undead | 886.8 | Undead / Human 882.7 | same | 886.8 | +0.0% |
| 12 | Fire | Undead | 873.6 | Undead / Gnome 870.0 | same | 872.1 | +0.2% |
| 13 | Survival | Human | 857.9 | Orc 851.3 / Human | Night Elf 851.3 | 857.9 | +0.0% |
| 14 | Shadow | Undead | 853.9 | Undead / Gnome 848.8 | same | 853.9 | +0.0% |
| 15 | Arcane | Undead | 830.1 | Undead / Gnome 819.8 | same | 830.1 | +0.0% |
| 16 | Balance | Night Elf | 826.0 | Tauren 817.6 / Night Elf | same | 824.8 | +0.1% |
| 17 | Enhancement | Orc | 802.3 | Orc / Dwarf 783.9 | same | 802.3 | +0.0% |
| 18 | Frost | High Order | 737.5 | Undead 736.4 / High Order | same | 737.5 | +0.0% |
| 19 | Feral | Night Elf | 688.5 | Windshaper 686.7 / Night Elf | same | 688.5 | +0.0% |
| 20 | Rogue Combat Daggers | Undead | 686.8 | Undead / Night Elf 671.8 | same | 714.3 | -3.8% |
| 21 | Smite | Undead | 676.9 | Undead / Gnome 671.1 | same | 676.9 | +0.0% |
| 22 | Elemental | Orc | 674.6 | Orc / Dwarf 667.4 | same | 674.6 | +0.0% |
| 23 | Rogue Combat Swords | Undead | 664.9 | Undead / Human 660.4 | same | 691.5 | -3.8% |
| 24 | Rogue Seal Fate Daggers | Undead | 657.0 | Undead / Gnome 645.0 | same | 683.3 | -3.8% |
| 25 | Rogue Mutilate | Undead | 654.4 | Undead / Gnome 639.3 | same | 680.6 | -3.8% |
| 26 | Rogue Seal Fate Swords | Undead | 631.6 | Undead / Gnome 623.3 | same | 656.8 | -3.8% |
| 27 | Rogue Subtlety | Undead | 627.7 | Undead / Gnome 614.2 | same | 652.8 | -3.8% |

**Readings:**
- **Rogues are the only real movers, -3.8% each.** Murder (+4% against Humanoid, Giant, Beast and Dragonkin) is dead on the generic boss. Every rogue build now sits in the bottom 9 of 27, with Combat Daggers (686.8) just under Feral (688.5).
- **Everything else moves -0.0 to +0.4%.** Sunder and Expose remove the same 2250 armor.
- **Spread:** top over bottom is 1.77× (Demonology 1112 over Subtlety 628). It was 1.70×.


</details>

<details><summary>Previous §1 (bundled round, Humanoid boss)</summary>

### 1 (superseded) Cross-spec table (bundled round, standards 1b + 5a + 5b; re-swept 2026-10-13 10:30 UTC)

Settings: 180 sec, 5000 iterations, seed 1, standard error about 0.5 to 1. Each spec was swept over every race its class can play in Forever (Skyborne included; Windshaper is not a mage). The racials-off column is the best race and DPS with the weapon-type racials (Human sword, Orc axe, Dwarf mace) switched off. The heads are warrior 7348a81, paladin 6bf522f, hunter 23a266f, rogue d880922, priest 10bae4e, shaman 148542d, mage 4230073, warlock 1750d98, druid 2444b9f, every one rebuilt. The "Before" column is the 09:00 sweep on the previous standards (5a + 1a; Ret, Enh and the hunters on their original sets).

**What fairness means now** (user ruling for 1b: predict Forever Phase 1 BiS as accurately as possible). Set totals differ by archetype on purpose, as real epics do:
- Casters and hunters sit above physical sets, and hybrids at physical × 1.15.
- Per-spec Stamina comes from each spec's SoD BiS.
- Every spec gets two synthetic trinkets.
A set is fair when it follows 1b and 5a for its archetype.

**Audit:**
- Every set passes 5a and 1b: per-piece linear totals, the 1.7-norm shape check, weapons, the trinket stat totals of 45.5 to 46.3 (against 45.9 ±1%), and the on-use trinket sizes (140 SP, 335 AP, 415 RAP, 165 Str, 20 s per 2 min).
- Survival dual-wields one-handers at 0.42 × L(71) = 34.4, the hunter one-hand rate added to 1b in a04b43a.

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off | Before (09:00) | Change |
|---|---|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1110.3 | Undead 1103.8 / Gnome | same | 1053.7 (Gnome) | +5.4% |
| 2 | Fury (dual wield) | Undead | 1020.2 | Undead / Human 1012.2 | same | 1006.3 (Undead) | +1.4% |
| 3 | Marksmanship | Night Elf | 1006.9 | Orc 1001.9 / Night Elf | same | 911.9 (Night Elf) | +10.4% |
| 4 | Destruction | Gnome | 984.1 | Orc 983.9 / Gnome | same | 949.0 (Undead) | +3.7% |
| 5 | Affliction | Gnome | 978.4 | Orc 974.6 / Gnome | same | 941.2 (Gnome) | +4.0% |
| 6 | Beast Mastery | Night Elf | 969.5 | Orc 964.5 / Night Elf | same | 881.1 (Night Elf) | +10.0% |
| 7 | Destruction Fire | Undead | 968.6 | Undead / Gnome 968.5 | same | 935.7 (Undead) | +3.5% |
| 8 | Retribution | Undead | 967.3 | Undead / Human 951.5 | same | 961.9 (Undead) | +0.6% |
| 9 | Frostfire | Undead | 900.3 | Undead / High Order 899.6 | same | 859.3 (Undead) | +4.8% |
| 10 | Arms | Undead | 890.1 | Undead / Human 885.3 | same | 883.5 (Undead) | +0.7% |
| 11 | Fury (two-hand) | Undead | 886.8 | Undead / Human 882.7 | same | 878.8 (Undead) | +0.9% |
| 12 | Fire | Undead | 872.1 | Undead / Gnome 868.9 | same | 836.6 (Undead) | +4.2% |
| 13 | Survival | Human | 857.9 | Orc 851.3 / Human | Night Elf 851.3 | 778.3 (Human) | +10.2% |
| 14 | Shadow | Undead | 853.9 | Undead / Gnome 848.8 | same | 845.0 (Undead) | +1.1% |
| 15 | Arcane | Undead | 830.1 | Undead / Gnome 819.8 | same | 807.6 (Undead) | +2.8% |
| 16 | Balance | Night Elf | 824.8 | Tauren 816.4 / Night Elf | same | 789.6 (Night Elf) | +4.5% |
| 17 | Enhancement | Orc | 802.3 | Orc / Dwarf 783.9 | same (Orc 794.9) | 791.9 (Orc) | +1.3% (Searing Totem, shaman 0024236) |
| 18 | Frost | High Order | 737.5 | Undead 736.4 / High Order | same | 704.3 (Undead) | +4.7% |
| 19 | Rogue Combat Daggers | Undead | 714.3 | Undead / Night Elf 698.7 | same | 700.8 (Undead) | +1.9% |
| 20 | Rogue Combat Swords | Undead | 691.5 | Undead / Human 686.8 | same | 675.9 (Undead) | +2.3% |
| 21 | Feral | Night Elf | 688.5 | Windshaper 686.7 / Night Elf | same | 673.6 (Night Elf) | +2.2% |
| 22 | Rogue Seal Fate Daggers | Undead | 683.3 | Undead / Gnome 670.8 | same | 673.2 (Undead) | +1.5% |
| 23 | Rogue Mutilate | Undead | 680.6 | Undead / Gnome 664.8 | same | 670.4 (Undead) | +1.5% |
| 24 | Smite | Undead | 676.9 | Undead / Gnome 671.1 | same | 681.0 (Undead) | -0.6% |
| 25 | Elemental | Orc | 674.6 | Orc / Dwarf 667.4 | same | 646.7 (Orc) | +4.3% |
| 26 | Rogue Seal Fate Swords | Undead | 656.8 | Undead / Gnome 648.2 | same | 643.5 (Undead) | +2.1% |
| 27 | Rogue Subtlety | Undead | 652.8 | Undead / Gnome 638.7 | same | 640.2 (Undead) | +2.0% |

**Readings:**
- **Spread.** Top over bottom is 1.70× (Demonology 1110 over Subtlety 653). It was 1.65× before.
- **Hunters gain the most** (+10.0 to +10.4%). They get both their first 5a rebudget (ranged AP and Int moved into Agility and crit) and the hunter archetype factor.
- **Casters +2.8 to +5.4%** (Shadow +1.1, Smite -0.6). Two things drive it:
  - Forever's caster weapon rate (a staff at 2.21 × L) and the caster archetype factor outweigh spell power's move to 0.86.
  - Priests gain less because their per-spec Stamina is higher: Shadow and Smite 16.6% against Mage 13.9% and Warlock 14.2%, about 2.5 points of the set that the others spend on damage. Their linear totals match the other cloth casters to within 0.3%.
  - Elemental and Balance carry a totem or idol instead of a stat wand. The relic's effect is its value under 1b, so their weapon-slot totals are 193 against 215.
- **Physical +0.6 to +2.3%.** The physical archetype factor (0.91 to 0.96) offsets the trinkets and the higher per-spec Stamina (19.6 to 20.6%).
- **Ret +0.6%, Enh +1.3%** (Enh +0.7 of that from moving to Searing Totem at shaman 0024236, 12:25 re-sweep). The hybrid factor (physical × 1.15) and the trinkets roughly offset spell power's move to 0.86 on their held SP share.
- **Races.**
  - Frost's best race is now High Order, 1.1 ahead of Undead, which is within noise.
  - Destruction's is now Gnome, 0.2 ahead of Orc, also within noise.
  - Enhancement stays Orc with racials off (794.9 against Windshaper 788.2); the axe racial is worth 7.4.
  - Survival is the only build a weapon racial still moves: Human's sword is worth 6.6, and Night Elf is best with racials off.


</details>

## 2. What is on equal footing (the standards in orchestration/standards.md)

| Area | Standard | Status |
|---|---|---|
| Armor budget | One item-level table: tier slots 66; neck, back, wrist, waist and rings 71 (402.5) | Every set between 397.2 (Feral) and 403.6 (rogue daggers); gear auditor pass 4 |
| Weapon budget | Two-handers 54.8 (item level 76). Physical one-handers about 20.5. Caster main hand plus off hand, or staff, about 73 | All sets |
| Ranged slot | A stat wand or bow (about 15.6), or a coded Forever relic | Libram of Law, Burning Totem, Swarming and Howling Idols |
| Hit | Each set optimised at budget prices, in practice at its spec's cap | Casters 10% gear plus 2-piece (Demo 14%, Balance 11%, Smite 0% via Holy talents); melee 9%; SV 6% + Surefooted |
| Set bonuses | 2-piece only. 4- and 5-piece values reported separately | All |
| Trinkets | None (Forever removed the dungeon ones; Talisman of Ephemeral Power is Molten Core) | All |
| Enchants | Best Forever Phase 1 enchant per slot, Enchanting 300 included | All. Head is Lesser Arcanum of Voracity for everyone, and nobody uses the full Dire Maul Arcanums |
| Crit elixirs | At most one of Mongoose, Grizzly, Owl, Sages, Whale | All |
| Consumables | Forever Phase 1 sources only, best per slot | All. The hunter merged the consumes branch in pass 6 |
| Raid | Shared main.go raid. Melee groups get Windfury, Strength of Earth and Battle Shout. Casters lose Might and gain Mana Spring | Consistent. Under Forever, Moonkin and Leader of the Pack are a shared 3% crit; Trueshot is ranged only; improved totems are neutralised |
| Pre-pull | Any class may pre-buff or pre-cast; damage counts from 0 | Every pre-cast lands at 0 or later (Frostfire Bolt +0.5, Fireball +1.0, Shadow Bolt +0.2, Lightning Bolt +0.5, Starfire 0, Arcane Blast 0, Mind Blast and Smite at 0 or later) |
| Stealth openers | No start in Stealth (user ruling) and no Prowl opener | Rogue opens without Stealth; Feral's Prowl opener was reverted (8bc55c6) |
| Pets | No player buffs and no pet consumables, for hunter pets and demons alike. Both inherit their owner's hit | Hunter 0d2cf5e; warlock 4742e1a |
| Encounter | Level 63, 3731 armor, Humanoid; 120/180/300 sec; same execute phases; reaction time 200 ms | Shared main.go, unchanged on every branch |
| Professions | Engineering for explosives; the second profession has no DPS effect except Tailoring for crafted sets | Feral has no explosives (cat form) |

## 3. Sensitivities: placeholders that move the ranking

| Item | Effect | Why it stays |
|---|---|---|
| Warlock demon scaling | Demonology 1022 to 1098 across the documented range (no owner scaling up to the Season of Discovery placeholder). The placeholder adds about 50 to 60 on demon-out builds | Forever's figures are server side and unmeasured. Sourced Season of Discovery figures: 0.5 x spell power as attack power, 25% as spell damage, hit 100% spell / 80% melee, no crit. **The warlock lead (Demo about 13% over Ret) rests partly on this.** |
| Hunter pet scaling | On Season of Discovery's 22% of ranged attack power instead of Forever's measured 10%: BM +58, MM +32, SV +28 | The 10% is a Forever beta measurement and ranks above Season of Discovery in the source order |
| Pet consumables | +17 BM if Forever lets owners use Juju or scrolls on pets | Ruled off (item 14): consumables on a pet count as a player buff |
| 4- and 5-piece set bonuses | Off in presets. If on: Frostfire 5-piece +6.0%, Shadow 5-piece +5.0%, BM +2.0%, MM +1.7%, Feral +1.7%, others 0 to 1.3%. Warrior and rogue 4-piece (+36 AP vs Humanoids) would be live on the default boss (about +1.1%) | Forever Tier 1 pieces have no stats and no known source yet |
| Demon in melee (warlock) | Reported by the warlock with halved melee as its own range | Placeholder |

- **Ice Lance coefficient (mage_ice_lance_coef, 2026-10-15).** The sim uses 0.429, SoD's server value. The client's value is 0. The dashboard's Low is 0.2145 (Frost -10.7%, Frostfire -6.0%). The client's literal 0 would cost Frost about 18.5% (results auditor), taking it from 10th to 18th. The Low end should include 0, because it is the only value with a Forever source.

- **Warrior Rage normalisation constant k (standards 21; the largest open sensitivity, 2026-10-13 20:05).** k = 3.46 / 4.5 Rage per weapon-second (one-hander / two-hander) is a level 30 beta measurement, scaled to level 60 by s (dashboard assumption warr_rage_k). The presets ship s = 0.5 (Expected: Fury 1037, Arms 909, Fury 2H 890). The Low end (s = 0) costs Fury -8.4%, Arms -5.3% and Fury 2H -6.2%, taking Fury from 2nd to 8th. The High end (s = 1) adds +3.8%, +2.8% and +3.7%. The fix is a level 60 Forever measurement. Until then the warrior rows should carry this range wherever the ranking is shown (dashboard assumption warr_rage_norm).

- **Raid damage hits (ruled (a) and merged as standards 24, PR #46/#47; the §1 table at 0798c11 has the shipped numbers, which include Enrage).** Original measurement (master f3dc866, 2026-10-15 21:50; raised by the results auditor): The shaman presets now assume raid damage hits Enhancement 2 times a minute and Elemental once a minute, which keeps Lightning Shield firing (Enh +7.9, Ele +2.3). Warriors gain Rage from damage taken, but their sims take no hits. I measured this with a local-only hook that adds damage-taken Rage at the classic formula (damage x 2.5 / the level 63 conversion, times the stance multiplier), with exponential gaps like the shaman option, at 10k iterations:

  | Build (race) | Base | 1/min x 1500 | 2/min x 1000 | 2/min x 1500 | 2/min x 2500 |
  |---|---|---|---|---|---|
  | Fury (Orc) | 1056.2 | +3.0 | +3.8 | +6.4 | +8.6 |
  | Fury (Undead) | 1058.4 | +3.1 | +2.8 | +4.7 | +6.8 |
  | Arms (Orc) | 951.3 | +1.9 | +2.5 | +5.0 | +7.4 |
  | Arms (Undead) | 947.2 | +2.4 | +3.0 | +4.9 | +7.3 |
  | Fury 2H (Orc) | 924.8 | +2.0 | +3.9 | +5.6 | +8.4 |
  | Fury 2H (Undead) | 918.6 | +3.3 | +4.0 | +6.2 | +8.4 |

  At the shamans' melee rate (2 a minute) and a 1500 hit, warriors gain about 0.5 to 0.6%, about the same in DPS as Enhancement's +7.9 (+1.0%). Arms (Orc, 956.3) would pass BM (952.0) for 9th. Fury stays 2nd and Fury 2H stays 12th. The hit size is my assumption, because the shaman option has no damage figure. No other base preset gains from being hit. Whether normalised Rage (standards 21) should scale damage-taken Rage as well is open.

- **Generic raid hits and pushback (#55, 2026-10-16 08:15; not yet swept).** Every DPS now takes 2 real spell hits a minute, which push back casts: casters lose 1.5 to 2.5% (orchestrator figure). Pushback talents in the base presets:
  - Balance has Nature's Focus 5/5 (70% avoid), which is **unmodeled**, so Balance is over-penalised.
  - Arcane has Improved Channeling 1/5, also unmodeled but small.
  - No other caster preset takes one: no Burning Soul, Twilight Focus, Intensity, Fel Concentration or Eye of the Storm.
  - I reported both to the orchestrator for the druid and mage workers.

- **No level-based partial resist (#102, user ruling 2026-10-16).** Forever's client resist formula has no level term, and 520 beta hits on mobs 3 to 5 levels up showed no partial resists. The rule applies to every non-binary spell from every class, so it is consistent. Casters gain 2 to 7%, rogues about 1% (poisons) and warriors about 0.2%, which widens the caster/melee gap. At 16:05 only the warlock metas carried it, so the live ranking had the warlocks about 5% ahead of the other casters (results auditor pass 36; re-run requested). My 18:17 pass re-sweeps it. A boss log is still on the to-verify list.

## 4. Inherent Forever asymmetries (not fixed, by design)

- **Gloves school power:** Forever Phase 1 has +20 gloves only for Holy, Arcane and Nature. Shadow, Fire and Frost casters (Shadow, every warlock, Frostfire, Fire, Frost) get only the +6 spell power kit, worth about 0.3 to 0.5%.
- **Feral in cat form** can't use weapon oils, stones, explosives or Windfury. This is structural.
- **Rogue deficit** (results auditor, client-confirmed): Forever's rogue nerfs (-59), the warrior's off-hand rage (+44 on the comparison side) and no Execute or Heroic Strike equivalent.
- **Elemental and Smite are low** from client coefficients: Lightning Bolt 196 base at 0.714, Smite 126 at 0.714, against Fireball 483 at 1.0.
- Creature-type talents on the generic boss (standards 19): hunter Improved Tracking applies (+4% MM/SV); rogue Murder and Ret's Demon/Undead bonuses do not, and live in the type-specific presets.
- Mage staff imbue scrolls (Forever's mage-only Comprehension items, 277500-277502; user ruling 2026-10-15, assumption mage_imbue_scrolls). They stack with Wizard Oil and add 10.8 to 11.4 DPS (Greater Flame / Greater Frost) or 6.2 (Precision) at 180 sec. No other caster has a class imbue of its own. foreverchanges.pro lists the scrolls as hidden.

## 5. What is not on equal footing

Nothing that the standards cover. Every class applies every standard on its final head. The only remaining differences are the sensitivities in section 3 and the inherent asymmetries in section 4.

## 6. History (what the passes changed, largest first)

- Crit elixirs: casters and hybrids drank Owl and Sages (+4% crit) while pure melee did not. Settled by the one-crit-elixir rule.
- Caster hit: casters were 4 to 6% under the spell hit cap. Trading spell power for hit was worth +0.4 to +5.2%. All now at cap.
- Caster weapon slots: a staff was 25 points behind main hand plus off hand, about 2.8% for Ele and Balance. Equalised at about 73.
- Armor item levels: the spread was 382 to 413 from each class's Era best-in-slot item levels. One table now.
- Set bonuses: 0 to 7.3% spread, with the warrior 4-piece live on the Humanoid boss. Now 2-piece only.
- Weapon racials: the Human sword gave 1.4 to 2.3% on casters with Azuresong. Casters on staves now; a racials-off column was added.
- Missing items:
  - Ret's Libram of Fervor did nothing; Libram of Law is now coded.
  - Druids had empty relic slots; Swarming and Howling Idols are now coded.
  - Neck enchants were missing on five classes.
  - Arcanum of Focus (Dire Maul, phase 2 in the database) was on Ele and Warlock.
  - Several pre-casts were missing: Arcane, Balance, Ele, Shadow, Smite.
- Hunter: hadn't merged the shared consumables until pass 6. Pet hit inheritance is now added.
- Race columns: corrected from preset defaults to real sweeps. Warlock moved to Gnome; mage moved to Undead.

## 7. Stat mix (standards item 5a, audited 2026-10-13 06:50 UTC)

Method: each set's synthetic items, without enchants or stat-less relics. Weapons and the ranged slot are included. Each stat's budget share comes from the item 1 formula: an item's cost is (sum p^1.5)^(2/3) with p = weight x amount. Each stat gets p^1.5 / sum p^1.5 of its item's cost, and the set share is the cost-weighted sum. AP and RAP in equal amounts count once, as AP. Hit and crit are rating at weight 1. DPS effects are 180 sec, 8-10k iterations, on the set's shipped build, applied as bonus stats. They are a linear approximation and upper bounds, because the 1.5 norm pays back some budget when a set diversifies.

### 7.1 Budget share per set

| Set | Sta | Str | Agi | Int | Spi | AP | RAP | SP | School SP | Hit | Crit | Largest non-Sta stat (share of non-Sta) |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Warrior Fury | 17.3 | 24.4 | 16.0 | - | - | 7.8 | - | - | - | 12.4 | 22.0 | Str 29.6% |
| Warrior Arms / Fury 2H | 17.9 | 25.9 | 16.0 | - | - | 7.5 | - | - | - | 12.0 | 20.7 | Str 31.6% |
| Paladin Ret | 15.0 | 20.3 | 3.9 | 10.1 | 8.9 | - | - | 21.7 | - | 10.4 | 9.8 | SP 25.5% |
| Hunter ranged | 18.4 | - | 32.6 | 5.9 | 8.5 | - | 14.3 | - | - | 10.8 | 9.5 | Agi 40.0% |
| Hunter Survival | 18.4 | 5.3 | 35.3 | 4.2 | 8.7 | 10.6 | - | - | - | 7.8 | 9.7 | Agi 43.3% |
| Rogue daggers | 21.7 | 2.3 | 38.7 | - | - | 13.1 | - | - | - | 6.2 | 18.0 | Agi 49.5% |
| Rogue swords | 21.7 | 10.7 | 30.4 | - | - | 17.1 | - | - | - | 6.2 | 14.0 | Agi 38.8% |
| Priest Shadow | 11.4 | - | - | 14.9 | 8.0 | - | - | 49.0 | - | 10.4 | 6.3 | SP 55.3% |
| Priest Smite | 12.9 | - | - | 16.6 | 8.9 | - | - | 55.3 | - | 0.0 | 6.3 | SP 63.4% |
| Shaman Enh | 17.1 | 14.1 | 6.3 | 10.9 | 9.3 | 2.1 | - | 24.1 | - | 10.4 | 5.7 | SP 29.1% |
| Shaman Ele | 12.9 | - | - | 14.3 | 7.9 | - | - | 45.1 | - | 11.2 | 8.6 | SP 51.8% |
| Mage (and Arcane) | 8.8 | - | - | 15.3 | 8.4 | - | - | 46.2 | - | 11.4 | 9.9 | SP 50.7% |
| Warlock | 13.5 | - | - | 13.5 | 7.1 | - | - | 46.5 | - | 10.8 | 8.6 | SP 53.7% |
| Warlock Demo | 13.5 | - | - | 13.5 | 7.1 | - | - | 41.4 | - | 15.9 | 8.6 | SP 47.9% |
| Druid Balance (and Wrath) | 15.0 | - | - | 17.2 | 8.1 | - | - | 40.4 | - | 12.8 | 6.5 | SP 47.5% |
| Druid Feral | 19.2 | 25.0 | 29.4 | 1.7 | - | 5.0 | - | - | - | 11.8 | 7.7 | Agi 36.5% |

No set carries school-only spell power. The Arcane mage and Balance Wrath sets match their main sets item for item and differ only in enchants.

### 7.2 Stamina

The current sets span 8.8% (Mage) to 21.7% (Rogue) of the item budget, and raw Stamina runs 94 (Mage) to 164 (Rogue). Casters sit low and melee high. **Proposed common share: 15% of each set's item budget, ±1.5 points.** That is the median of the current sets, and Ret and Balance already sit on it.

Effect of moving each set to 15%: its non-Stamina stats scale by 0.85 / (1 - current Stamina share), with hit left at cap.

| Set | Sta now | DPS change at 15% |
|---|---|---|
| Mage Frostfire / Fire / Arcane | 8.8% | -3.75% / -3.70% / -3.16% |
| Priest Shadow | 11.4% | -1.55% |
| Priest Smite | 12.9% | -1.22% |
| Shaman Ele | 12.9% | -1.21% |
| Warlock (Destro, Demo) | 13.5% | -0.92% |
| Paladin Ret | 15.0% | 0 |
| Druid Balance | 15.0% | 0 |
| Shaman Enh | 17.1% | +0.78% |
| Warrior Fury | 17.3% | +0.99% |
| Warrior Arms | 17.9% | +1.28% |
| Hunter BM / MM / SV | 18.4% | +1.35% / +1.41% / +1.26% |
| Druid Feral | 19.2% | +1.74% |
| Rogue daggers / swords | 21.7% | +1.97% / +3.22% |

At present the casters gain up to about 4% over the rogue from Stamina alone, so this matters more than any other remaining asymmetry.

Spirit, for reference: the mana users carry about 7 to 9% Spirit (the original "about 10% Spirit" note), and the pure physical sets carry none. Spirit is worth close to nothing for most of these specs (stat_weights.md), so it is a handicap the mana users carry, not an advantage.

### 7.3 Item 5a check per set

| Set | Verdict | Reason | Fair mix in Phase 1 terms | Cost of the fix |
|---|---|---|---|---|
| Priest Smite | Fail | SP 63.4% of non-Sta. No hit at all: Holy talents cover it, so the budget went into SP | SP about 45%, the rest Int, crit and Spirit, like real Phase 1 Holy/Disc caster gear | -4.95% at a 50% cap, -6.87% at 45% |
| Priest Shadow | Fail | SP 55.3% | SP about 45 to 50%, more Int and crit | -0.80% at 50%, -1.64% at 45% |
| Warlock (Destro, Affli, DestroFire) | Fail | SP 53.7% | Same | -0.82% at 50%, -2.10% at 45% |
| Shaman Ele | Borderline fail | SP 51.8% | Same | -0.58% at 50%, -2.29% at 45% |
| Mage, Mage Arcane | Borderline | SP 50.7% | Same | -0.12% at 50%, -1.3% at 45% |
| Warlock Demo | Pass | SP 47.9% | - | (-0.71% at 45%) |
| Druid Balance | Pass | SP 47.5% | - | (-0.82% at 45%) |
| Rogue daggers | Borderline pass | Agi 49.5%, Str only 2.3% | Move about 5 points of Agi into Str and AP, like real Phase 1 rogue leather | Small; Str and AP are worth about 1.1 against Agi 1.44 on Combat Daggers |
| All others (Warrior, Ret, Hunter, Rogue swords, Enh, Feral) | Pass | Largest stat 25 to 43%, with at least two useful secondary stats | - | - |

Costs move a set's excess spell-power budget into Int and crit half and half, at linear prices. They are upper bounds.

### 7.4 Linear shares (the gear auditor's metric)

These are shares of weight × amount, which the gear auditor reproduces exactly.

| Set | Stamina | Largest non-Sta stat |
|---|---|---|
| Warrior Fury | 19.8% | Str 28.4% |
| Warrior Arms | 20.2% | Str 29.9% |
| Paladin Ret | 15.9% | SP 24.0% |
| Hunter ranged | 19.6% | Agi 36.0% |
| Hunter Survival | 20.0% | Agi 38.5% |
| Rogue daggers | 23.9% | Agi 45.4% |
| Rogue swords | 23.0% | Agi 37.0% |
| Priest Shadow | 14.4% | SP 45.0% |
| Priest Smite | 16.7% | SP 53.4% |
| Shaman Enh | 17.6% | SP 26.8% |
| Shaman Ele | 15.6% | SP 42.4% |
| Mage | 12.0% | SP 41.9% |
| Warlock | 16.2% | SP 44.2% |
| Warlock Demo | 15.9% | SP 39.7% |
| Druid Balance | 17.1% | SP 40.0% |
| Druid Feral | 21.0% | Agi 33.8% |

My recommendation to the orchestrator is to rule in these terms: a common Stamina share of 18% ±1.5, with the 5a cap at 50% or 45% of the non-Sta budget. The ruling decides which.

Ruled (bfccd2e68): metric (L), Stamina 18% ±1.5, guardrail 45% of non-Sta, Spirit under item 5. §7.2's 15% table is superseded by the 18% column in §7b.3.

## 7b. Target stat mix per set, from Season of Discovery level 60 BiS (audited 2026-10-13 07:15 UTC; replaces the 07:00 version)

### 7b.1 Source, and checking the orchestrator's method

The SoD input is orchestration/sod_stat_mix.md (orchestrator, coordination branch): wowsims/sod master, every level 60 BiS set from phase 4 on, averaged per spec. I reproduced it from the same files.

- **Hit and crit (the orchestrator is right).** A SoD item that gives "+1% hit" or "+1% crit" sets the melee and the spell stat both to 1. They must be counted once (10 points per 1% hit, 14 per 1% crit). My 07:00 version counted both, which doubled SoD hit and crit. Those targets and DPS numbers are withdrawn.
- **Where the orchestrator's table departs from the ruled 5a metric.** With these departures switched on, I reproduce 49 of its 70 SoD rows to 0.1 points and the rest to within 4.
  1. Trinkets are counted. Our sets carry none (item 8a), so trinket AP, crit and haste inflate SoD's AP and crit shares (e.g. Speedstone, AP trinkets).
  2. AP and RAP are both counted when an item carries the same amount of each. The ruling counts them once. This also inflates the table's "Ours" rows for Rogue daggers (AP 21.3, ruled metric 13.0), Rogue swords (25.0 vs 15.6) and Hunter Survival (16.4 vs 9.8).
  3. Feral: SoD's weapon-derived feral AP is counted as AP (Rae'lar's 338 feral AP gives the SoD feral 24% AP). It's a SoD-only mechanic derived from weapon DPS, and our feral weapon carries none. Counted, feral AP comes out at 2%.
  4. School-only SP seems to be weighted at about 0.55. Item 1 says school-only SP is lower, and I use 0.45. This only shifts Shadow priest shares by a few points.
  5. Haste is in the denominator (small: it only shows on p6 trinkets and the p7/p8 sets).
- **Dropped.** These are not priced by item 1 and are left out: haste, expertise / weapon skill, Defense, block value, weapon feral AP, resistances, set bonuses, runes, procs, use effects and enchants.
- **Target.** Each set's target is built in four steps:
  1. Take the SoD average proportions.
  2. Set Stamina to 18%.
  3. Set Spirit to 10% for mana users (item 5); rage and energy users get none.
  4. Hold hit at the current set's share (item 4). The rest is shared out in SoD proportions.
  
  Guardrail clip: any stat over 45% of non-Sta is clipped and the excess shared pro rata. Only Rogue swords needed it: Agi 47.7% clipped to 45.0%.
- **Smite.** Proposal A (recommended): Shadow priest SoD proportions, school SP as generic SP (Smite is Holy), no hit. Proposal B: Balance proportions, no hit.
- **Mage.** "Mage" averages the SoD fire and frost sets. "Mage Arcane" averages the arcane sets.

### 7b.2 Target table (% of the set's weighted stats, (L))

| Set | Sta | Str | Agi | Int | Spi | AP | SP | School SP | Hit | Crit | Largest non-Sta | Distance now (½ L1, points) | SoD sets averaged |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Warrior Fury | 18.0 | 34.0 | 16.1 | - | - | 6.4 | - | - | 13.0 | 12.5 | Str 41.5% | 11.2 | 4 |
| Warrior Arms / Fury 2H | 18.0 | 35.0 | 17.2 | - | - | 7.3 | - | - | 12.5 | 9.9 | Str 42.7% | 11.8 | 4 |
| Paladin Ret | 18.0 | 25.9 | 5.0 | 7.4 | 10.0 | 3.0 | 1.8 | 2.8 (Holy) | 11.2 | 14.8 | Str 31.6% | 22.0 | 7 |
| Hunter ranged | 18.0 | 1.1 | 33.8 | 3.4 | 10.0 | 4.5 | - | - | 11.7 | 17.4 | Agi 41.2% | 19.3 | 6 |
| Hunter Survival | 18.0 | 10.7 | 29.5 | 0.2 | 10.0 | 9.2 | - | - | 8.2 | 14.2 | Agi 36.0% | 9.0 | 4 |
| Rogue daggers | 18.0 | 18.0 | 35.9 | - | - | 7.4 | - | - | 7.3 | 13.3 | Agi 43.8% | 16.5 | 5 |
| Rogue swords | 18.0 | 18.9 | 36.9 | - | - | 5.6 | - | - | 7.2 | 13.5 | Agi 45.0% (clipped) | 15.7 | 3 |
| Priest Shadow | 18.0 | - | - | 18.2 | 10.0 | - | 12.5 | 16.1 (Shadow) | 12.8 | 12.4 | Int 22.3% | 26.0 | 4 |
| Priest Smite (A, recommended) | 18.0 | - | - | 22.2 | 10.0 | - | 34.8 | - | - | 15.0 | SP 42.4% | 11.2 | 4 (Shadow) |
| Priest Smite (B) | 18.0 | - | - | 23.5 | 10.0 | - | 33.0 | - | - | 15.5 | SP 40.2% | 13.0 | 3 (Balance) |
| Shaman Enh | 18.0 | 24.9 | 5.5 | 10.6 | 10.0 | 7.6 | 0.6 | - | 11.3 | 11.6 | Str 30.3% | 23.6 | 6 |
| Shaman Ele | 18.0 | - | - | 20.7 | 10.0 | - | 25.8 | - | 12.9 | 12.6 | SP 31.4% | 10.3 | 3 |
| Mage (Frostfire, Frost, Fire) | 18.0 | - | - | 18.6 | 10.0 | - | 24.4 | 2.9 | 12.8 | 13.4 | SP 29.8% | 12.6 | 4 |
| Mage Arcane | 18.0 | - | - | 18.7 | 10.0 | - | 24.1 | 6.6 (Arcane) | 12.8 | 9.4 | SP 29.4% | 14.3 | 2 (plus 0.5 MP5) |
| Warlock (Destro, Affli, DestroFire) | 18.0 | - | - | 18.8 | 10.0 | - | 26.2 | 1.5 | 12.6 | 12.9 | SP 31.9% | 10.8 | 6 |
| Warlock Demo | 18.0 | - | - | 15.4 | 10.0 | - | 24.6 | - | 17.3 | 14.7 | SP 30.0% | 8.8 | 2 |
| Druid Balance (and Wrath) | 18.0 | - | - | 18.8 | 10.0 | - | 26.4 | - | 14.3 | 12.5 | SP 32.2% | 6.7 | 3 |
| Druid Feral | 18.0 | 24.9 | 26.4 | 6.2 | - | 2.2 | - | - | 12.7 | 9.6 | Agi 32.1% | 6.4 | 2 |

### 7b.3 DPS effect (180 s, 10k iterations, the shipped build, bonus stats at the set's current (L) total)

| Build | Stamina to 18% only | Full SoD target | Of which mix |
|---|---|---|---|
| Warrior Fury | +0.91% | -3.55% | -4.46 |
| Warrior Arms / Fury 2H | +1.10% / +1.12% | -2.52% / -3.93% | -3.62 / -5.05 |
| Paladin Ret | -0.98% | -3.86% (SP share held: +0.98%) | -2.88 (+1.96) |
| Hunter BM / MM | +0.89% / +0.94% | +5.42% / +5.48% | +4.53 / +4.54 |
| Hunter Survival | +0.96% | +3.84% | +2.88 |
| Rogue Combat / Seal Fate daggers | +1.71% / +2.41% | -2.38% / -2.92% | -4.09 / -5.33 |
| Rogue Mutilate / Subtlety | +2.26% / +2.01% | -2.93% / -1.90% | -5.19 / -3.91 |
| Rogue Combat / Seal Fate swords | +2.50% / +2.68% | +0.46% / +0.70% | -2.04 / -1.98 |
| Priest Shadow | -2.06% | -0.26% | +1.80 |
| Priest Smite | -0.87% | A -5.67% / B -7.01% | -4.80 / -6.14 |
| Shaman Enh | -0.10% | +1.06% (SP share held: +1.26%) | +1.16 (+1.36) |
| Shaman Ele | -1.93% | -7.90% | -5.97 |
| Mage Frostfire / Frost / Fire | -5.17% / -4.44% / -4.95% | -7.36% / -7.16% / -6.06% | -2.19 / -2.72 / -1.11 |
| Mage Arcane | -4.03% | -5.11% | -1.08 |
| Warlock Destro / Affli / DestroFire | -1.48% / -1.54% / -1.43% | -5.39% / -5.23% / -6.48% | -3.91 / -3.69 / -5.05 |
| Warlock Demo | -1.90% | -3.64% | -1.74 |
| Druid Balance | -0.74% | -3.09% | -2.35 |
| Druid Feral | +1.51% | +0.65% | -0.86 |

Notes:
- **Method.** "Stamina to 18% only" scales every other stat to fill the gap, holding hit and Spirit. The effect is linear at a fixed (L) total. The item 1 cost norm pays back a little budget on the more spread-out targets, so real rebudgets come out slightly better.
- **Why most builds lose.** Today's sets were tuned to each build's stat weights, so a realistic mix costs most builds DPS: 2 to 6% on the mix for melee, 1 to 6% for casters. The exceptions:
  - Hunters gain about 4.5% on the mix: today's 13.7% ranged-only AP moves into Agility and crit.
  - Shadow and Enh gain about 1 to 2%.
- **Cross-spec effect.** The changes are uneven. Relative to the median build, hunters move up about 8 to 9 points, and Ele, Mage and DestroFire move down about 3 to 5. The §1 ranking would reshuffle, so it needs a re-sweep after the rebudgets.
- **Ret and Enh.** Forever's Ret and Enh scale with SP (stat_weights.md), and SoD melee hybrid sets wear little SP (Ret's shockadin sets aside). The "SP share held" variant keeps today's SP share and fills the rest in SoD proportions. I recommend it for these two.

### 7b.4 Open points for the ruling

1. **Ret and Enh:** the SoD target as written, or "SP share held" (recommended).
2. **Tolerance:** each non-Sta stat within ±3 points (L) of target, or ½ L1 ≤ 5 overall.
3. **Smite:** A (recommended) or B.
4. **Mage:** one shared set for Frostfire, Frost and Fire, or a Fire / Frost split.
5. **The SoD input file.** sod_stat_mix.md should be corrected on items 1 to 3 of 7b.1 (trinkets, AP/RAP double count, feral weapon AP) so workers who read it see the ruled metric.

### 7b.5 Under standards 1a (linear spend, 1.73 × budget per piece)

- **Set budgets.** V(ilvl) × slot mod, summed over the pieces that carry stats.
  - Caster weapon slots are taken at item 3's ~73 (46.2 main hand + 26.8 off-hand, or 73 for a staff). 1a's "V × slot mod" would give a caster main hand only 0.42 × V(76) (about 23), but today's caster main hands carry 2.2 to 2.45× that. The two rules conflict, and 1a should say which one wins for caster weapons.
  - Sets whose ranged / relic slot carries no stats (Ret's libram, Enh and Ele totems, Balance and Feral idols) get no linear budget there; the relic's effect is the value (item 3). That is why Ret and Enh land at about today's total, while sets with a stat wand, bow or thrown gain the extra 1.73 × 15.6 = 27 points.
- **1a total / today's total.**
  - Warrior Fury 1.16, Arms 1.14
  - Rogue 1.16 to 1.17
  - Hunter ranged 1.07, SV 1.10
  - Feral 1.12
  - Smite 1.16, Shadow 1.09
  - Mage 1.08, Warlock 1.07, Demo 1.05
  - Ele 1.06, Balance 1.07
  - Ret 0.98, Enh 1.00
- **Per-piece check.** Every current piece is off 1a by more than 1%, as expected before the rebudgets. Physical pieces mostly sit at 1.4 to 1.6×, caster pieces at 1.5 to 1.8×, and physical off-hands and ranged at about 1.0 to 1.2×.

DPS effect at the 1a total (180 s, 10k iterations, hit held at today's capped amount, the rest in the stated proportions):

| Build | 1a total, today's mix | 1a total, SoD target |
|---|---|---|
| Warrior Fury / Arms / Fury 2H | +7.21% / +5.68% / +6.12% | +4.22% / +2.64% / +1.72% |
| Paladin Ret | -0.62% | -4.44% (SP share held: +0.36%) |
| Hunter BM / MM / SV | +2.40% / +2.54% / +3.10% | +8.17% / +8.49% / +7.49% |
| Rogue CDag / SFDag / Mut / Sub | +4.39% / +5.33% / +5.11% / +5.31% | +4.09% / +3.85% / +3.78% / +4.56% |
| Rogue CSwd / SFSwd | +6.63% / +6.93% | +7.22% / +7.12% |
| Priest Shadow / Smite | +3.83% / +8.13% | +3.63% / +1.70% |
| Shaman Enh / Ele | -0.08% / +3.57% | +0.96% (SP held +1.10%) / -4.90% |
| Mage Frostfire / Frost / Fire / Arcane | +5.43% / +4.57% / +5.04% / +4.57% | -2.59% / -3.28% / -1.64% / -1.10% |
| Warlock Destro / Affli / DestroFire / Demo | +4.52% / +4.71% / +4.26% / +3.12% | -1.38% / -1.09% / -2.74% / -0.67% |
| Druid Balance / Feral | +3.86% / +4.34% | +0.61% / +4.94% |

Readings:
- **Under 1a with the SoD targets,** physical builds gain 2 to 8%, while casters land between -5% (Ele) and +4% (Shadow).
- **Ret and Enh are the only builds that don't gain from 1a.** They have no stat ranged slot, and they keep their SP only if the user rules "SP held". This widens the gap between them and the other melee by about 4 to 7 points. The relic slot's linear value under 1a (27 points) should be compared with what the libram or totem gives, which is a separate question for the ruling.

### 7b.6 Standards 5b targets (Ret, Enh, hunters) and the pending spell power price

- **Ret and Enh** (orchestrator ruling on 5b): SP and Int are held at today's shares, Sta 18, Spi 10, hit held, and the rest in SoD physical proportions.
  - Ret: Sta 18 / Str 15.7 / Agi 3.1 / Int 11.1 / Spi 10 / AP 1.8 / SP 20.2 / Hit 11.2 / Crit 9.0.
  - Enh: Sta 18 / Str 13.5 / Agi 3.0 / Int 11.7 / Spi 10 / AP 4.1 / SP 22.1 / Hit 11.3 / Crit 6.3.
  - Today's sets miss only Ret's Stamina (15.9) and Enh's Agility (6.3 against 3.0). Both still need the 1a per-piece spend.
- **Hunters** keep 10% Spirit, so their §7b.2 rows stand.

**Pending next-round change: spell power priced at 0.86 instead of 0.55** (orchestration/forever_item_budget.md). At the same linear share, every SP user's gear SP drops to 0.55 / 0.86 = 64%. Measured at 180 s, 10k iterations, on the shipped builds; Ret and Enh are on their pre-rebudget sets:

| Build | Gear SP now | DPS change |
|---|---|---|
| Ret / Enh | 295 / 319 | -5.1% / -3.9% |
| Shadow / Smite | 445 / 537 | -9.2% / -14.1% |
| Elemental | 389 | -12.2% |
| Mage Frostfire / Frost / Fire / Arcane | 425 / 425 / 425 / 485 | -12.8% / -11.9% / -11.3% / -11.9% |
| Warlock Demo / Destro / Affli / DestroFire | 403 / 434 / 434 / 434 | -11.5% / -12.5% / -12.7% / -12.5% |
| Balance | 400 | -13.1% |

Physical builds are unaffected by the SP price. So under this change casters fall 9 to 14% relative to the physical builds, and the hybrids 4 to 5%. The proposal's other half, the 1.28 × ilvl - 10.9 linear total, cuts every set by about 1% (ilvl 66 slots), 6% (ilvl 71) and 10% (two-handers), so it falls a little harder on two-hand builds.

Harness note: until 08:35 my run cache keyed on the request only. Item stats live in the simulator's compiled DB, so a rebudget that kept item IDs could reuse an old result. A full re-sweep with the cache keyed on the binary build reproduced every §1 number exactly. The only stale values were the bases of the first SP 0.86 run, which are corrected above.

### 7c. Synthetic trinket worth per build (2026-10-13 12:35 UTC; raised by the results auditor)

The shipped builds, 180 s, 10k iterations, the harness's default race. Each trinket slot was emptied in turn, then both together. Shown as DPS (% of the build without that trinket):

| Build | Stat trinket | On-use | Both |
|---|---|---|---|
| Fury / Arms / Fury 2H | +32.3 (3.3%) / +23.2 (2.7%) / +23.7 (2.7%) | +22.4 (2.2%) / +17.0 (1.9%) / +17.8 (2.1%) | 5.7% / 4.8% / 5.0% |
| Rogue CDag / SFDag / Mut / Sub | 2.6 / 3.0 / 3.0 / 2.0% | 2.7 / 2.4 / 2.4 / 2.8% | 5.4 / 5.4 / 5.4 / 4.9% |
| Rogue CSwd / SFSwd | 3.1 / 3.2% | 3.0 / 2.7% | 6.2 / 6.1% |
| Feral | 3.1% | 2.7% | 5.9% |
| Hunter BM / MM / SV | 3.0 / 3.2 / 2.8% | 2.3 / 2.5 / 2.1% | 5.5 / 5.8 / 5.0% |
| Ret / Enh | 2.6 / 2.3% | 1.4 / 1.6% | 4.1 / 4.0% |
| Shadow / Smite | 2.9 / 3.2% | 2.0 / 2.4% | 5.1 / 5.7% |
| Ele / Balance | 4.0 / 4.2% | 2.8 / 3.1% | 7.0 / 7.5% |
| Mage FF / Frost / Fire / Arcane | 4.3 / 3.7 / 3.8 / 3.4% | 2.1 / 2.0 / 2.2 / 2.2% | 6.6 / 5.9 / 6.1 / 5.7% |
| Warlock Demo / Destro / Affli / DestroFire | 3.8 / 3.9 / 4.0 / 3.7% | 2.8 / 3.0 / 2.9 / 2.8% | 6.8 / 7.0 / 7.1 / 6.6% |

**Reading:**
- **Same budget, different returns.** Both trinkets are priced the same for everyone by 1b: the stat trinket at 45.9 linear, and the on-use at the Talisman of Ephemeral Power / Earthstrike averages.
- **Spread.** In percent they range from 4.0% (Enh, Ret 4.1) to 7.5% (Balance), and in absolute DPS from 30 to 70.
- **Why the stat trinket varies.** It is the same 45.9 points for every set, but a caster set's total weighted stats are larger. So its spread is about how many DPS a budget point buys per spec, which is the same thing that sets the armor's value. It is not a pricing defect.
- **Hybrids get the least from both:**
  - Their stat trinket spreads over many lower-yield stats.
  - Ret's on-use is spell power, which Ret converts at a low rate. 1b allows either kind, and Ret chose the one that sims best.
- **Casters gain about 1.5 points more than physical builds overall.** That is because the 1.7-norm and the SP price leave casters' marginal spell power more valuable at their stat levels than physical builds' marginal AP and Agility.

**Options for a ruling:**
- **(A) Keep it.** Trinkets are budget-equal, like the armor. This is the current rule.
- **(B) Apply the 1b ring-tier archetype factor to the stat trinket** (caster 0.995, physical 0.911, hunter 0.977, hybrid 1.048), for consistency with the armor. Estimated effect from the measured worth:
  - physical builds -0.25 to -0.30 points
  - hunters -0.07
  - hybrids +0.12
  - casters about 0
  
  That slightly widens the caster lead.
- **(C) Size the trinkets by DPS worth instead of budget, so every build gains the same percentage.** This would make trinkets the only slot not priced by budget, which runs against 1b's "predict real BiS" premise.

I recommend (A). Option (B) is defensible for consistency, but it moves almost nothing and widens the gap. (C) conflicts with the user's 1b ruling.
