# Notes round 2026-11-11: full DPS, Tanking and Healing re-sim and deploy

- **Master:** ad68842 (fast-forward from branch `full-run-2026-11-11`).
- **Deploy:** `d4fee2d5` (https://d4fee2d5.forever-dps.pages.dev), production https://forever-dps.pages.dev. All 44 files were checked live.
- **What was re-simmed:**
  - All 68 DPS builds in full, at 1516752.
  - Mid-run, master took the Nov 5-10 daily-update merge (ca07df6), which brought the natural flasks of client 70862 (0326924), item folds and the report hashes of 28 builds. Those 28 builds and both modes were re-simmed at 707ae92. The other 40 builds' hashes did not move.
  - No sim input differs between 707ae92 and the deployed master. `stale.mjs` reports 0 of 68 builds and 0 of 2 modes stale.
- **Client build:** `state.json` `foreverClientBuild` is now 1.60.1.70888, the build the classes round's data pull used. ledger.md G4 says nothing newer exists.

## 180 s ranking (base builds)

1. Hunter · Lone Wolf · Lone Wolf (Night Elf): 1291.2
2. Hunter · Marksmanship · Marksmanship (Orc): 1232.3
3. Hunter · Beast Mastery · Beast Mastery (Orc): 1174.8
4. Druid · Feral · Feral cat (Night Elf): 1123.6
5. Warlock · Affliction · Affliction (Undead): 1070.8
6. Mage · Frostfire · Frostfire (Human): 1064.3
7. Warlock · Destruction · Destruction (Orc): 1062.8
8. Warlock · Demonology · Demonology (Orc): 1054.0
9. Mage · Fire · Fire (Human): 1052.8
10. Shaman · Enhancement · Level 60 (Orc): 1030.6
11. Shaman · Enh Imp GW · Imp GW (Orc): 1016.2
12. Warlock · Affliction Shadow Bolt · Affliction Shadow Bolt (Orc): 1015.1
13. Paladin · Ret twist · Level 60 (twist) (Undead): 1009.3
14. Paladin · Ret sub-Holy · Level 60 (sub Holy) (Undead): 999.1
15. Warrior · Fury · Fury Dual Wield (Orc): 998.9
16. Warlock · Destruction Fire · Destruction Fire (Undead): 997.3
17. Shaman · Elemental · Level 60 (Windshaper): 976.9
18. Shaman · Enh Water Shield · Water Shield (Orc): 976.3
19. Paladin · Ret Shock · Level 60 Shock (Undead): 971.1
20. Mage · Arcane · Arcane (Gnome): 962.1
21. Shaman · Elemental Mana Tide · Level 60 + Mana Tide (Windshaper): 959.6
22. Druid · Balance · Balance (Night Elf): 951.7
23. Priest · Shadow · Level 60 (Undead): 948.5
24. Mage · Frost · Frost (High Order): 943.9
25. Shaman · Enh Mana Tide · Mana Tide (Orc): 930.0
26. Warrior · Fury 2H · Fury Two-Hand (Orc): 925.6
27. Hunter · Survival · Survival (melee) (Orc): 916.9
28. Rogue · Combat Daggers · Combat Daggers (Undead): 914.6
29. Priest · Shadow + Holy · Level 60 + Holy (Undead): 906.4
30. Rogue · Combat Swords · Combat Swords (Undead): 899.7
31. Rogue · Seal Fate Swords · Seal Fate Swords (Undead): 881.7
32. Warrior · Arms · Arms (Orc): 871.2
33. Rogue · Seal Fate Daggers · Seal Fate Daggers (Undead): 869.8
34. Druid · Balance Wrath · Balance Wrath (Night Elf): 861.5
35. Rogue · Subtlety · Subtlety (Night Elf): 839.7
36. Rogue · Mutilate · Mutilate (Undead): 827.3
37. Priest · Smite · Level 60 Smite (Undead): 780.9

## Every build: live and new 180 s DPS, with the cause

Causes were split by sim at each commit: 3,000 iterations, seed 1, paired.
- Windfury Totem is 9ba00c2 -> 0ce9dbf.
- Natural flasks are 1516752 plus 0326924 cherry-picked, against 1516752.
- Nothing else between the live data (8dc4985) and master moves any of these builds.
- A move of 0.0 means the build didn't change; the run is deterministic.

| Build | Live | New | Move | Cause |
|---|---|---|---|---|
| druid__balance | 951.7 | 951.7 | +0.0 |  |
| druid__balance-duty-ff | 915.5 | 915.5 | +0.0 |  |
| druid__balance-wrath | 861.5 | 861.5 | +0.0 |  |
| druid__balance-wrath-duty-ff | 828.8 | 828.8 | +0.0 |  |
| druid__feral-cat | 1168.8 | 1123.6 | -45.2 | Windfury Totem (0ce9dbf) -28, natural flasks (0326924) -17 |
| druid__feral-cat-duty-ff | 1164.4 | 1118.7 | -45.7 | Windfury Totem (0ce9dbf) -28, natural flasks (0326924) -17 |
| hunter__beast-mastery | 1195.4 | 1174.8 | -20.6 | natural flasks (0326924): the Aggression flask zone crit 4% -> 2% |
| hunter__beast-mastery-tallstrider | 1136.9 | 1117.1 | -19.8 | natural flasks (0326924): the Aggression flask zone crit 4% -> 2% |
| hunter__beast-mastery-wolf | 1150.4 | 1130.5 | -19.9 | natural flasks (0326924): the Aggression flask zone crit 4% -> 2% |
| hunter__lone-wolf | 1313.8 | 1291.2 | -22.6 | natural flasks (0326924): the Aggression flask zone crit 4% -> 2% |
| hunter__marksmanship | 1254.0 | 1232.3 | -21.7 | natural flasks (0326924): the Aggression flask zone crit 4% -> 2% |
| hunter__marksmanship-tallstrider | 1221.1 | 1200.0 | -21.1 | natural flasks (0326924): the Aggression flask zone crit 4% -> 2% |
| hunter__marksmanship-wolf | 1228.8 | 1207.4 | -21.4 | natural flasks (0326924): the Aggression flask zone crit 4% -> 2% |
| hunter__survival | 1040.4 | 916.9 | -123.5 | Windfury Totem (0ce9dbf) -96, natural flasks (0326924) -28 |
| hunter__survival-tallstrider | 1006.3 | 883.8 | -122.5 | Windfury Totem (0ce9dbf) -96, natural flasks (0326924) -28 |
| hunter__survival-wolf | 1014.0 | 891.5 | -122.5 | Windfury Totem (0ce9dbf) -96, natural flasks (0326924) -28 |
| mage__arcane | 969.3 | 962.1 | -7.2 | Arcane Concentration (classes round M1); matches classes.md exactly |
| mage__fire | 1057.3 | 1052.8 | -4.5 | Arcane Concentration (classes round M1); matches classes.md exactly |
| mage__frost | 943.9 | 943.9 | +0.0 |  |
| mage__frostfire | 1064.3 | 1064.3 | +0.0 |  |
| paladin__level_60 | 1103.4 | 1009.3 | -94.1 | Windfury Totem (0ce9dbf); no flask |
| paladin__level_60-duty-crusader | 1094.9 | 1002.8 | -92.1 | Windfury Totem (0ce9dbf); no flask |
| paladin__level_60-duty-wisdom | 1086.6 | 995.6 | -91.0 | Windfury Totem (0ce9dbf); no flask |
| paladin__level_60_shock_ret | 1030.2 | 971.1 | -59.1 | Windfury Totem (0ce9dbf); no flask |
| paladin__level_60_shock_ret-duty-crusader | 1029.1 | 972.4 | -56.7 | Windfury Totem (0ce9dbf); no flask |
| paladin__level_60_shock_ret-duty-wisdom | 1019.9 | 963.3 | -56.6 | Windfury Totem (0ce9dbf); no flask |
| paladin__level_60_sub_holy | 1091.6 | 999.1 | -92.5 | Windfury Totem (0ce9dbf); no flask |
| paladin__level_60_sub_holy-duty-crusader | 1083.1 | 992.9 | -90.2 | Windfury Totem (0ce9dbf); no flask |
| paladin__level_60_sub_holy-duty-wisdom | 1074.9 | 985.6 | -89.3 | Windfury Totem (0ce9dbf); no flask |
| paladin__level_60_sub_holy_undead | 1221.4 | 1130.7 | -90.7 | Windfury Totem (0ce9dbf); no flask |
| priest__shadow | 948.5 | 948.5 | +0.0 |  |
| priest__shadow-holy | 906.4 | 906.4 | +0.0 |  |
| priest__smite | 780.9 | 780.9 | +0.0 |  |
| rogue__combat-daggers | 954.6 | 914.6 | -40.0 | Windfury Totem (0ce9dbf) -21 to -47, plus natural flasks (0326924) about -18 |
| rogue__combat-swords | 964.5 | 899.7 | -64.8 | Windfury Totem (0ce9dbf) -21 to -47, plus natural flasks (0326924) about -18 |
| rogue__mutilate | 863.4 | 827.3 | -36.1 | Windfury Totem (0ce9dbf) -21 to -47, plus natural flasks (0326924) about -18 |
| rogue__seal-fate-daggers | 909.0 | 869.8 | -39.2 | Windfury Totem (0ce9dbf) -21 to -47, plus natural flasks (0326924) about -18 |
| rogue__seal-fate-swords | 945.2 | 881.7 | -63.5 | Windfury Totem (0ce9dbf) -21 to -47, plus natural flasks (0326924) about -18 |
| rogue__subtlety | 885.9 | 839.7 | -46.2 | Windfury Totem (0ce9dbf) -21 to -47, plus natural flasks (0326924) about -18 |
| shaman__elemental-level60 | 976.9 | 976.9 | +0.0 |  |
| shaman__elemental-level60-mana-tide | 959.6 | 959.6 | +0.0 |  |
| shaman__enhancement-imp-gw | 1016.2 | 1016.2 | +0.0 |  |
| shaman__enhancement-level60 | 1030.6 | 1030.6 | +0.0 |  |
| shaman__enhancement-mana-tide | 930.0 | 930.0 | +0.0 |  |
| shaman__enhancement-water-shield | 976.3 | 976.3 | +0.0 |  |
| warlock__affliction | 1034.4 | 1070.8 | +36.4 | classes round (Wrack, Soul Siphon, Bane of Agony rotation) and the dot fix f20d593; matches classes.md (1076.9 before the fix; 1061.4 / 1051.5 for the duties after it) |
| warlock__affliction-duty-curse | 1025.1 | 1060.3 | +35.2 | classes round (Wrack, Soul Siphon, Bane of Agony rotation) and the dot fix f20d593; matches classes.md (1076.9 before the fix; 1061.4 / 1051.5 for the duties after it) |
| warlock__affliction-duty-recklessness | 1019.7 | 1051.8 | +32.1 | classes round (Wrack, Soul Siphon, Bane of Agony rotation) and the dot fix f20d593; matches classes.md (1076.9 before the fix; 1061.4 / 1051.5 for the duties after it) |
| warlock__affliction_shadow_bolt | 1015.1 | 1015.1 | +0.0 |  |
| warlock__demonology | 1054.0 | 1054.0 | +0.0 |  |
| warlock__demonology-duty-curse | 1043.9 | 1043.9 | +0.0 |  |
| warlock__demonology-duty-recklessness | 1036.8 | 1036.8 | +0.0 |  |
| warlock__destruction | 1062.8 | 1062.8 | +0.0 |  |
| warlock__destruction-duty-curse | 1052.7 | 1052.7 | +0.0 |  |
| warlock__destruction-duty-recklessness | 1046.2 | 1046.2 | +0.0 |  |
| warlock__destruction_fire | 997.3 | 997.3 | +0.0 |  |
| warlock__destruction_fire-duty-curse | 987.7 | 987.7 | +0.0 |  |
| warlock__destruction_fire-duty-recklessness | 981.5 | 981.5 | +0.0 |  |
| warrior__arms | 1087.1 | 871.2 | -215.9 | Windfury Totem white-swing rule (0ce9dbf) plus natural flasks (0326924): Arms -200 / -16, Fury DW -100 / -38, Fury 2H -190 / -35 |
| warrior__arms-duty-battle-shout | 1083.0 | 868.3 | -214.7 | Windfury Totem white-swing rule (0ce9dbf) plus natural flasks (0326924): Arms -200 / -16, Fury DW -100 / -38, Fury 2H -190 / -35 |
| warrior__arms-duty-sunder | 1040.6 | 829.3 | -211.3 | Windfury Totem white-swing rule (0ce9dbf) plus natural flasks (0326924): Arms -200 / -16, Fury DW -100 / -38, Fury 2H -190 / -35 |
| warrior__arms_dragonkin | 1146.0 | 930.6 | -215.4 | Windfury Totem white-swing rule (0ce9dbf) plus natural flasks (0326924): Arms -200 / -16, Fury DW -100 / -38, Fury 2H -190 / -35 |
| warrior__fury | 1136.4 | 998.9 | -137.5 | Windfury Totem white-swing rule (0ce9dbf) plus natural flasks (0326924): Arms -200 / -16, Fury DW -100 / -38, Fury 2H -190 / -35 |
| warrior__fury-duty-battle-shout | 1125.3 | 983.5 | -141.8 | Windfury Totem white-swing rule (0ce9dbf) plus natural flasks (0326924): Arms -200 / -16, Fury DW -100 / -38, Fury 2H -190 / -35 |
| warrior__fury-duty-sunder | 1098.3 | 952.8 | -145.5 | Windfury Totem white-swing rule (0ce9dbf) plus natural flasks (0326924): Arms -200 / -16, Fury DW -100 / -38, Fury 2H -190 / -35 |
| warrior__fury_2h | 1150.2 | 925.6 | -224.6 | Windfury Totem white-swing rule (0ce9dbf) plus natural flasks (0326924): Arms -200 / -16, Fury DW -100 / -38, Fury 2H -190 / -35 |
| warrior__fury_2h-duty-battle-shout | 1146.7 | 922.9 | -223.8 | Windfury Totem white-swing rule (0ce9dbf) plus natural flasks (0326924): Arms -200 / -16, Fury DW -100 / -38, Fury 2H -190 / -35 |
| warrior__fury_2h-duty-sunder | 1114.7 | 899.7 | -215.0 | Windfury Totem white-swing rule (0ce9dbf) plus natural flasks (0326924): Arms -200 / -16, Fury DW -100 / -38, Fury 2H -190 / -35 |

### Survival hunter

The brief expected about 945; the report now has 916.9.
- The unnamed cause is **0ce9dbf** "Windfury Totem procs from white main hand swings only under Forever", 2026-11-08 (1041.2 -> 944.7).
  - Raptor Strike and Mongoose Bite no longer roll the totem: 31.1 -> 8.4 extra attacks a fight.
  - It is intended, a user ruling backed by the Nov 5-8 logs: 882 totem procs and none from about 2,770 Stormstrikes.
  - The commit also rewrote the Survival report hashes.
- The natural flasks then take another -28.
- **Open:** `notes.md` line 15 still lists whether on-next-swing attacks (Heroic Strike, Cleave, Raptor Strike, Maul) roll the totem as untested, and the sim says no. That assumption drives most of the warrior and Survival drop:
  - Fury 2H: 33.8 -> 12.3 totem procs a fight, and Heroic Strike casts fall from 7.7 to 2.1 for lack of Rage.
  - One in-game test, a warrior with the totem up doing about 100 Heroic Strikes, would settle it.

## Race tables

No preset's best race changed. The orders below the best race shift on the builds hit by Windfury Totem and the flasks:

  hunter__survival order: Orc > High Order > Windshaper > Night Elf > Human (was Orc > Night Elf > High Order > Windshaper > Human)
  rogue__combat-daggers order: Undead > Gnome > Orc > High Order > Windshaper (was Undead > Gnome > Orc > Night Elf > High Order)
  rogue__combat-swords order: Undead > Orc > Gnome > High Order > Windshaper (was Undead > Orc > Gnome > Night Elf > High Order)
  rogue__mutilate order: Undead > Gnome > Orc > High Order > Windshaper (was Undead > Gnome > Orc > Night Elf > High Order)
  rogue__seal-fate-daggers order: Undead > Gnome > Orc > High Order > Windshaper (was Undead > Gnome > Orc > Night Elf > High Order)
  rogue__seal-fate-swords order: Undead > Gnome > Night Elf > Orc > High Order (was Undead > Night Elf > Gnome > Orc > High Order)
  warrior__arms order: Orc > Undead > High Order > Windshaper > Night Elf (was Orc > Undead > Night Elf > High Order > Windshaper)
  warrior__arms_dragonkin order: Orc > Undead > High Order > Windshaper > Night Elf (was Orc > Undead > Night Elf > High Order > Windshaper)
  warrior__fury order: Orc > Undead > High Order > Windshaper > Night Elf (was Orc > Undead > Night Elf > High Order > Windshaper)
  warrior__fury_2h order: Orc > Night Elf > Undead > High Order > Windshaper (was Orc > Night Elf > Undead > Troll > Human)

## Tanking and Healing

Tanking, every score lower. The causes come from the mechanism, not a per-commit sim split: the same Windfury Totem rule, and the halved natural flasks, which are on in every bench request.
- Warrior, Bear and Warden tank threat falls 13-25%.
- Paladin falls about 2-7%, and moves up to rank 2 for max threat.

  warrior__protection_max_threat     scoreMaxThreat 4651.67->3957.67 scoreBalanced 3814.67->3088.83 scoreMaxSurvival 1303.17->482.83 tps 2316.67->2028.5 dps 651.67->555.17 dtps 415.0->415.17
  warrior__protection_balanced       scoreMaxThreat 4475.0->3876.0 scoreBalanced 3880.33->3226.5 scoreMaxSurvival 2096.0->1277.67 tps 2204.5->1952.33 dps 619.5->534.5 dtps 375.67->379.33
  warrior__protection_max_survival   scoreMaxThreat 3897.67->3413.0 scoreBalanced 3457.67->2949.83 scoreMaxSurvival 2138.67->1561.83 tps 1996.5->1778.17 dps 565.33->491.67 dtps 336.17->336.67
  paladin__protection_max_threat     scoreMaxThreat 3249.17->3171.83 scoreBalanced 2823.33->2742.0 scoreMaxSurvival 1546.17->1452.83 tps 1542.67->1506.5 dps 679.67->660.17 dtps 305.0->306.17 rank 3->2
  paladin__protection_balanced       scoreMaxThreat 3243.5->3169.17 scoreBalanced 2835.0->2753.17 scoreMaxSurvival 1610.0->1505.67 tps 1534.33->1501.33 dps 676.33->658.0 dtps 314.17->314.33
  paladin__protection_max_survival   scoreMaxThreat 3075.67->2846.83 scoreBalanced 2784.5->2523.67 scoreMaxSurvival 1910.5->1554.0 tps 1430.17->1337.5 dps 645.17->595.5 dtps 295.0->298.33
  druid__bear_max_threat             scoreMaxThreat 3491.83->2654.83 scoreBalanced 2380.5->1499.83 scoreMaxSurvival -953.83->-1966.5 tps 1873.67->1505.0 dps 720.33->560.33 dtps 610.0->612.17 rank 2->3
  druid__bear_balanced               scoreMaxThreat 3343.17->2597.83 scoreBalanced 2498.5->1717.0 scoreMaxSurvival -35.33->-925.5 tps 1759.33->1428.17 dps 673.67->530.33 dtps 583.67->586.83
  druid__bear_max_survival           scoreMaxThreat 2958.0->2323.33 scoreBalanced 2277.83->1597.83 scoreMaxSurvival 237.5->-578.33 tps 1527.5->1262.33 dps 582.5->470.0 dtps 544.33->545.17
  shaman__warden_max_threat          scoreMaxThreat 3169.83->2495.67 scoreBalanced 1712.17->982.67 scoreMaxSurvival -2661.17->-3556.0 tps 1721.5->1437.67 dps 984.5->793.67 dtps 720.67->718.0
  shaman__warden_balanced            scoreMaxThreat 3071.5->2450.5 scoreBalanced 1786.0->1124.67 scoreMaxSurvival -2071.33->-2853.17 tps 1645.0->1379.0 dps 925.83->749.5 dtps 677.33->675.0
  shaman__warden_max_survival        scoreMaxThreat 2956.83->2367.67 scoreBalanced 1720.67->1091.5 scoreMaxSurvival -1988.17->-2736.0 tps 1602.5->1348.83 dps 902.67->735.17 dtps 686.5->684.67

Healing, flat except for one preset:
- Restoration druid in the raid: HPS 863 -> 920 and carry 1.437 -> 1.391, rank 1 -> 4. The cause is Nature's Swiftness on Wild Growth (9628e85, classes round).
- The rest move by under 0.3 HPS (the healbench raid's tank and damage changed).

  paladin__holy                      hps 721.71->721.57 tankDied 21.56->21.53 oomAt 575.29->574.86
  druid__restoration                 oomAt 452.14->451.71
  shaman__restoration                tankDied 17.54->17.53 oomAt 405.29->405.57
  priest__discipline                 hps 675.14->675.29 ehps 564.0->564.14 tankDied 19.4->19.43
  priest__holy                       hps 816.57->816.86 oomAt 336.14->336.0
  paladin__holy_raid                 ehps 249.0->248.86 carry 1.4129->1.4127 oomAt 2759.86->2760.14 rank 3->2
  druid__restoration_raid            hps 863.43->919.57 ehps 295.14->292.57 carry 1.4371->1.3909 oomAt 1671.86->1599.86 rank 1->4
  shaman__restoration_raid           hps 463.0->463.14 ehps 275.86->276.0 carry 1.3956->1.3954 oomAt 2484.0->2484.14 rank 4->3
  priest__holy_raid                  ehps 296.57->296.71 carry 1.4371->1.4367 oomAt 1604.14->1603.43

## Tests

All on master as of 1a474ba.
- **3335001**, goldens regenerated, each moved by intended data or fixes:
  - sim/mage TestP1Mage, sim/priest/healing TestHoly and TestDisc, sim/priest/shadow TestP1Shadow: only the Ironweave Battlesuit row, from the item fold 4aae529.
  - sim/rogue/dps_rogue (Combat Sinister Strike, Combat Daggers, Mutilate): pre-BiS items re-stated by the folds 8051c43, 4aae529 and 2eca7be and the scrape 5d9ece7. Mutilate also moved in the 5th digit from the dot fix f20d593.
  - sim/warrior/dps_warrior TestP1DPSWarrior: item folds 8051c43, 4aae529, a778e52 and 5d9ece7.
- **6bdc320**, tools/rotopt TestWarriorDpsHoldUnderBuff:
  - Moved by 0ce9dbf: with less Windfury Rage, `dwEnd=15` gains Death Wish at 167.05 s against 165.
  - The hold still works. The test now lets a use come up to 3 s late; early stays strict at 1.5 s.
- **Still failing, not regenerated:** sim/warrior/dps_warrior TestP1DPSWarrior again, on master after ca07df6 and b1291e6.
  - b1291e6's in-game record of Edgemaster's Handguards (14551) has Armor only. It replaced the Wowhead and client data, so the gloves lost their 1% expertise (`forever_wowhead_items.json` has exprtng 10).
  - The BayovCore item sweep doesn't read expertise, so this looks like a data loss, not a game change.
  - No report build wears the gloves, and the report hashes hold.
  - The full suite at 1a474ba otherwise passed: 25 packages ok, TestRaidForeverReportHashesStable included.

## Changes tab

New lines on the live update `2026-11-10-hotfixes`:
- Windfury Totem white main-hand swings only (0ce9dbf).
- Natural flasks 30 Stamina and a 2% zone bonus (0326924, fc065aa).
- Farraki Enforcement (2e32052).
- Item data folds (15 commits).
- Downranking penalty on heals only (0193721).
- Orc shaman base attributes at levels 2-43 (3b0983b).

Marked internal: fb41592, 6bdc320, 3335001. The dot fix was already covered. `changes.py check`: 0 uncovered.

## Checks

- Playwright on the site passed at 390 and 1440 px, light and dark, in DPS, #tank and #heal across every tab:
  - no console errors;
  - no horizontal scroll;
  - each mode fetches only its own data file and common (DPS also lazily loads one events file).
- `dist/dps_report.html` checked by file:// in each mode.
- Cosmetic, not fixed: the 390 px tank Spec page's by-boss table clips its third column inside its own scroll box.
- `compsearch -stale` lists no stale searches beyond four never stamped (m1_10_class_buffs, m1_20, m1_20_class_buffs, m1_40_class_buffs).

## Timing

session_01TXnjyPE5JfDBjP3yqKyxXE has been idle since 2026-11-04, so the timing is recorded here.

- **Full run** (timing/20261111T163528Z): 68 builds, 80.3 min wall, 248.5 core-min. The tank and heal step was stopped when master moved.

  | Step | Wall min | Core-min |
  |---|---|---|
  | run_build | 34.7 | 94.5 |
  | weights | 12.7 | 44.0 |
  | races | 3.5 | 13.2 |
  | mana | 7.7 | 23.5 |
  | ablate | 3.4 | 12.0 |
  | assumptions | 5.5 | 21.2 |
  | probe | 12.7 | 40.2 |

  Core-min per base build: hunter 10.0, warrior 7.8, rogue 5.1, warlock 4.5, mage 3.7, druid 3.4, shaman 3.3, paladin 3.3, priest 2.0.
- **Flask re-sim** (timing/20261111T185904Z): 28 builds plus both modes, 251.6 min wall, of which tankheal took 212.5 min (heal races about 3 h). The DPS steps took 39 min wall and 145.9 core-min.

## Needs a ruling (picks left as master has them)

1. Do Heroic Strike, Cleave, Raptor Strike and Maul roll Windfury Totem? This drives about -100 to -200 on warriors and -96 on Survival. Ruled 2026-11-12: "for now only white hits proc it until we verify"; the sim stays as 0ce9dbf has it (standards 17, TASKS.md).
2. Warrior default: Fury 2H (925.6) now trails Fury Dual Wield (998.9) by 73. Ruled 2026-11-12: switch to the best build.
3. Rogue default: Combat Swords (899.7) now trails Combat Daggers (914.6). Ruled 2026-11-12: switch to the best build.
4. Already open: the warlock default (Affliction 1070.8 beats Destruction 1062.8), Arcane Concentration's reading, Bane of Agony's ramp, the Windfury cooldowns. Ruled 2026-11-12: the warlock default switches to the best build; Arcane Concentration, Bane of Agony's ramp and the Windfury cooldowns stay as modelled, to verify after the patch lands (TASKS.md).
5. Edgemaster's Handguards' lost expertise (b1291e6) and the dps_warrior golden it breaks. Ruled 2026-11-12: "use my data, it's what forever changed it to be"; the in-game record stands (docs/data_sources.md) and the golden is regenerated.
