# Round 70753, phase 3: DPS report regenerated and deployed

- Deployed to https://forever-dps.pages.dev on 2026-11-03, deploy `f435e8b7`, from master `0a988c9`.
- Full re-sim of all 68 builds, because the race base stats in sim/core changed (20k iterations, 120/180/300 s,
  generic boss). The follow-ups that landed during the run were merged and re-simmed: warlock Affliction (Wrack, Imp
  sacrificed) and Affliction Shadow Bolt, Enhancement Earth Shock hold, Smite, druid Feral and Balance, paladin Holy
  Strike first. The client 70755 restamp (same tables as 70753) moved nothing: three spot re-sims came out identical.
- Shared cleanup: Blessing of Salvation is gone from the raid buffs (0 DPS, checked on Fury and Fire). Assumption
  rows: shaman Lightning Bolt 31+ and Earth Shock 7 rows removed, Maelstrom Weapon PPM 20, Ice Lance one value,
  Trueshot Aura Rank 5 at 100, the hunter shot AP shares note at 15% / 5%, Two-Handed Weapon Specialization covers all
  Holy damage. Six rogue vs Humanoid builds removed; Affliction Shadow Bolt added. Defaults: warrior Fury 2H, rogue
  Combat Swords, warlock Destruction.
- Every build's 180 s number matches its class round file (within 0.2). `stale.mjs`: 0 of 68 stale.
  `changes.py check`: 91 commits, 0 uncovered.
- Browser checks (phone and desktop, every tab): no console errors. The page has a single dark theme.

## Race changes (baseline, all presets)

Undead best 15 -> 11 times (average gap 0.14 -> 0.18), Orc 12 -> 14 (0.25 -> 0.22), Gnome 1 -> 3, Night Elf 4 -> 5
(0.79 -> 0.56), Human average gap 0.84 -> 0.67, Tauren still last (1.57 -> 1.67).

## Open items

- The six shaman `req/*.meta.json` files still carry the pre-round `published.dps` (for example Enhancement 883.5);
  their requests and the report data are right.
- Patch 0002 (DASH_RACIALS_OFF) was rebased onto the new racials code.
- `compsearch -stale` lists no changed searches; 4 searches are unstamped (m1_10_class_buffs, m1_20,
  m1_20_class_buffs, m1_40_class_buffs).
- Left from the class files: Wildstalker caveat numbers not re-measured; stale druid_rip_31plus text in
  spec_assumptions.md; the rogue Assassination default and Reckless Swings GCD questions.

## Ranking at 180 s (base builds)

| # | Build | Race | DPS |
|---|---|---|---|
| 1 | Hunter · Lone Wolf | Night Elf | 1313.7 |
| 2 | Hunter · Marksmanship | Orc | 1254.3 |
| 3 | Hunter · Beast Mastery | Orc | 1195.6 |
| 4 | Warrior · Fury 2H | Orc | 1169.9 |
| 5 | Shaman · Enhancement | Orc | 1159.4 |
| 6 | Warrior · Fury | Orc | 1152.6 |
| 7 | Shaman · Enh Imp GW | Orc | 1149.8 |
| 8 | Druid · Feral | Night Elf | 1139.0 |
| 9 | Shaman · Enh Water Shield | Orc | 1111.3 |
| 10 | Warrior · Arms | Orc | 1097.7 |
| 11 | Paladin · Ret twist | Undead | 1092.1 |
| 12 | Paladin · Ret sub-Holy | Undead | 1091.2 |
| 13 | Mage · Frostfire | Human | 1064.3 |
| 14 | Warlock · Destruction | Orc | 1063.4 |
| 15 | Mage · Fire | Human | 1057.9 |
| 16 | Shaman · Enh Mana Tide | Orc | 1055.1 |
| 17 | Warlock · Demonology | Orc | 1053.9 |
| 18 | Hunter · Survival | Orc | 1041.2 |
| 19 | Warlock · Affliction | Undead | 1034.1 |
| 20 | Paladin · Ret Shock | Undead | 1030.2 |
| 21 | Warlock · Affliction Shadow Bolt | Orc | 1015.6 |
| 22 | Warlock · Destruction Fire | Undead | 996.7 |
| 23 | Shaman · Elemental | Windshaper | 970.1 |
| 24 | Rogue · Combat Swords | Undead | 963.5 |
| 25 | Mage · Arcane | Gnome | 962.3 |
| 26 | Rogue · Seal Fate Swords | Undead | 957.0 |
| 27 | Shaman · Elemental Mana Tide | Windshaper | 952.7 |
| 28 | Druid · Balance | Night Elf | 952.1 |
| 29 | Priest · Shadow | Undead | 948.7 |
| 30 | Rogue · Combat Daggers | Undead | 946.7 |
| 31 | Mage · Frost | High Order | 944.7 |
| 32 | Rogue · Seal Fate Daggers | Undead | 920.7 |
| 33 | Priest · Shadow + Holy | Undead | 906.6 |
| 34 | Rogue · Subtlety | Night Elf | 881.7 |
| 35 | Rogue · Mutilate | Undead | 866.1 |
| 36 | Druid · Balance Wrath | Night Elf | 861.5 |
| 37 | Priest · Smite | Undead | 781.0 |

## Every build, old (report before this round) -> new, 180 s

| Build | Old | New | Change |
|---|---|---|---|
| druid__balance | 892.3 | 952.1 | +6.7% |
| druid__balance-duty-ff | 858.6 | 915.5 | +6.6% |
| druid__balance-wrath | 853.0 | 861.5 | +1.0% |
| druid__balance-wrath-duty-ff | 820.8 | 829.0 | +1.0% |
| druid__feral-cat | 942.5 | 1139.0 | +20.8% |
| druid__feral-cat-duty-ff | 939.3 | 1134.1 | +20.7% |
| hunter__beast-mastery | 1046.5 | 1195.6 | +14.2% |
| hunter__beast-mastery-tallstrider | 988.5 | 1136.9 | +15.0% |
| hunter__beast-mastery-wolf | 1001.8 | 1150.5 | +14.8% |
| hunter__lone-wolf | 1089.9 | 1313.7 | +20.5% |
| hunter__marksmanship | 1066.3 | 1254.3 | +17.6% |
| hunter__marksmanship-tallstrider | 1033.2 | 1220.9 | +18.2% |
| hunter__marksmanship-wolf | 1040.9 | 1228.3 | +18.0% |
| hunter__survival | 1024.6 | 1041.2 | +1.6% |
| hunter__survival-tallstrider | 990.7 | 1007.4 | +1.7% |
| hunter__survival-wolf | 998.6 | 1014.5 | +1.6% |
| mage__arcane | 937.5 | 962.3 | +2.6% |
| mage__fire | 1057.1 | 1057.9 | +0.1% |
| mage__frost | 944.7 | 944.7 | +0.0% |
| mage__frostfire | 1064.3 | 1064.3 | +0.0% |
| paladin__level_60 | 1039.1 | 1092.1 | +5.1% |
| paladin__level_60-duty-crusader | 1031.7 | 1082.9 | +5.0% |
| paladin__level_60-duty-wisdom | 1024.2 | 1075.3 | +5.0% |
| paladin__level_60_shock_ret | 973.5 | 1030.2 | +5.8% |
| paladin__level_60_shock_ret-duty-crusader | 973.1 | 1029.3 | +5.8% |
| paladin__level_60_shock_ret-duty-wisdom | 963.8 | 1020.0 | +5.8% |
| paladin__level_60_sub_holy | 1037.4 | 1091.2 | +5.2% |
| paladin__level_60_sub_holy-duty-crusader | 1031.4 | 1082.8 | +5.0% |
| paladin__level_60_sub_holy-duty-wisdom | 1023.6 | 1075.0 | +5.0% |
| paladin__level_60_sub_holy_undead | 1150.3 | 1221.4 | +6.2% |
| priest__shadow | 948.7 | 948.7 | +0.0% |
| priest__shadow-holy | 906.6 | 906.6 | +0.0% |
| priest__smite | 761.4 | 781.0 | +2.6% |
| rogue__combat-daggers | 790.8 | 946.7 | +19.7% |
| rogue__combat-daggers-humanoid | 814.5 | - | removed |
| rogue__combat-swords | 782.1 | 963.5 | +23.2% |
| rogue__combat-swords-humanoid | 808.6 | - | removed |
| rogue__mutilate | 743.8 | 866.1 | +16.4% |
| rogue__mutilate-humanoid | 769.1 | - | removed |
| rogue__seal-fate-daggers | 752.9 | 920.7 | +22.3% |
| rogue__seal-fate-daggers-humanoid | 777.9 | - | removed |
| rogue__seal-fate-swords | 747.1 | 957.0 | +28.1% |
| rogue__seal-fate-swords-humanoid | 773.0 | - | removed |
| rogue__subtlety | 732.9 | 881.7 | +20.3% |
| rogue__subtlety-humanoid | 758.4 | - | removed |
| shaman__elemental-level60 | 767.6 | 970.1 | +26.4% |
| shaman__elemental-level60-mana-tide | 752.8 | 952.7 | +26.6% |
| shaman__enhancement-imp-gw | 879.3 | 1149.8 | +30.8% |
| shaman__enhancement-level60 | 883.5 | 1159.4 | +31.2% |
| shaman__enhancement-mana-tide | 795.8 | 1055.1 | +32.6% |
| shaman__enhancement-water-shield | 834.8 | 1111.3 | +33.1% |
| warlock__affliction | 1015.8 | 1034.1 | +1.8% |
| warlock__affliction-duty-curse | 1005.2 | 1025.1 | +2.0% |
| warlock__affliction-duty-recklessness | 999.6 | 1019.9 | +2.0% |
| warlock__affliction_shadow_bolt | - | 1015.6 | new |
| warlock__demonology | 1054.5 | 1053.9 | -0.1% |
| warlock__demonology-duty-curse | 1043.8 | 1043.6 | -0.0% |
| warlock__demonology-duty-recklessness | 1037.4 | 1036.6 | -0.1% |
| warlock__destruction | 997.5 | 1063.4 | +6.6% |
| warlock__destruction-duty-curse | 987.2 | 1052.6 | +6.6% |
| warlock__destruction-duty-recklessness | 981.7 | 1046.4 | +6.6% |
| warlock__destruction_fire | 977.1 | 996.7 | +2.0% |
| warlock__destruction_fire-duty-curse | 967.5 | 986.9 | +2.0% |
| warlock__destruction_fire-duty-recklessness | 961.3 | 980.7 | +2.0% |
| warrior__arms | 1008.7 | 1097.7 | +8.8% |
| warrior__arms-duty-battle-shout | 1004.9 | 1094.0 | +8.9% |
| warrior__arms-duty-sunder | 970.4 | 1047.1 | +7.9% |
| warrior__arms_dragonkin | 1050.3 | 1142.2 | +8.7% |
| warrior__fury | 1041.5 | 1152.6 | +10.7% |
| warrior__fury-duty-battle-shout | 1039.5 | 1150.5 | +10.7% |
| warrior__fury-duty-sunder | 1012.5 | 1117.7 | +10.4% |
| warrior__fury_2h | 1044.4 | 1169.9 | +12.0% |
| warrior__fury_2h-duty-battle-shout | 1041.3 | 1165.4 | +11.9% |
| warrior__fury_2h-duty-sunder | 1009.0 | 1132.6 | +12.2% |
