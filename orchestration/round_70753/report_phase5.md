# Round 70753 phase 5: full re-sim, landed and deployed (2026-11-04)

- **Master:** 1a54a518 (fast-forward from branch `phase5-full-run`; three merges of origin/master on the way).
- **Deploy:** `5aa05cc0` (https://5aa05cc0.forever-dps.pages.dev), production https://forever-dps.pages.dev, every file checked live.
- **Data:** all 68 builds re-simmed in full. After each master merge, a probe found which builds moved and only those were re-simmed:
  - first merge: 29 builds, threat only, no DPS moved;
  - second merge: feral cat, from the Goblin Sapper Charge in Cat Form;
  - third merge: feral cat again, from hotfix 113403's Ferocious Bite.
- **Tanking and Healing modes:** re-run on the final engine. stale.mjs reports 0 of 68 builds and 0 of 2 modes stale at master.
- **Checks:**
  - Full Go suite green on the final merge (25 packages).
  - `report_hashes.txt` holds.
  - `changes.py check`: 0 uncovered.
  - Playwright passes at 390 and 1440 px, in light and dark, across the DPS, Tank and Heal modes and every tab: no console errors and no horizontal scroll.

## Rulings applied

1. **Ret default:** the twist build on the hit-3 step (`paladin__level_60`) is now the paladin class default in place of sub-Holy.
   - Changed: the metas' `isDefault`, `ui/retribution_paladin/presets.ts` (default gear, APL and talents) and the order of the `sim.ts` builds.
   - DISPLAY in `build_data.py` needed no change.
   - Changes tab: a method line.
2. **Survival talents:** now 502-0050051-500230230250022151 in the 3 requests, their metas, `ui/hunter/presets.ts`, the wiki status page and the hunter test preset.
   - Only the talents moved.
   - Re-simmed with races. 180 s: 1041.2 → 1040.4 (flat, as expected).
   - Its stat weights row in `orchestration/stat_weights.md` is regenerated.
   - Changes tab: a talents line.
3. **Tests:** the AllItems-<Class>Synthetic(Phase1) tests skip the hit-step items (`core.IsHitTradeStepItem`).
   - Goldens were regenerated, including the 9 that master's merges touched.
   - The Mage, Warlock and Rogue synthetic rows are byte-identical to 75e1e9a13 (`git diff 75e1e9a13 -- sim/mage/TestP1Mage.results` is empty).
4. **No change:** Demonology keeps Suppression 5/5, and Fury DW stays on hit+2.

## Other changes

- **Fury Sunder duty meta:** the rotation text now matches its APL (Bloodthirst before Whirlwind; Heroic Strike at 70+ Rage).
- **Changes tab, new public lines:**
  - hit-trade gear (21 builds);
  - Dwarf hit steps;
  - warrior hotfix 113539;
  - the warrior rotation;
  - Unbridled Wrath on white swings only;
  - Divine Favor without Benediction;
  - labeled random streams;
  - adaptive fight counts;
  - `warr_rs_landed`;
  - the Ret default;
  - Survival talents;
  - feral Goblin Sapper Charge;
  - Ferocious Bite hotfix 113403.
- **Changes tab, other edits:**
  - The shaman 113380 line now gives its 180 s delta.
  - Tank, healer, leveling and test commits are marked internal.
- **compsearch `drops.json` regenerated:** TestDropsAreUnread failed on mage__frost, because its entries were stale since 45111f01.
- **Tank page text:** two LEAK rules in `tankheal_text/tank.md`, for "(a scoring run reads no health)" and "(docs/…)" citations.

## 180 s ranking, old (live before this round) and new

"Old" is the report data on master before this run (c40ede3), which the live site showed.

| # | Build | Old (live) | New | Δ | Best race |
|---|---|---|---|---|---|
| 1 | hunter__lone-wolf | 1313.7 | 1313.8 | +0.1 | Night Elf |
| 2 | hunter__marksmanship | 1254.3 | 1254.0 | -0.3 | Orc |
| 3 | hunter__marksmanship-wolf | 1228.3 | 1228.8 | +0.5 | Night Elf |
| 4 | paladin__level_60_sub_holy_undead | 1221.4 | 1221.4 | +0.0 | Undead |
| 5 | hunter__marksmanship-tallstrider | 1220.9 | 1221.1 | +0.2 | Night Elf |
| 6 | hunter__beast-mastery (default) | 1195.6 | 1195.4 | -0.2 | Orc |
| 7 | druid__feral-cat (default) | 1139.0 | 1168.8 | +29.8 | Night Elf |
| 8 | shaman__enhancement-level60 (default) | 1159.4 | 1166.0 | +6.6 | Orc |
| 9 | druid__feral-cat-duty-ff | 1134.1 | 1164.4 | +30.3 | Night Elf |
| 10 | shaman__enhancement-imp-gw | 1149.8 | 1150.5 | +0.7 | Orc |
| 11 | hunter__beast-mastery-wolf | 1150.5 | 1150.4 | -0.1 | Orc |
| 12 | warrior__fury_2h (default) | 1173.3 | 1150.2 | -23.1 | Orc |
| 13 | warrior__fury_2h-duty-battle-shout | 1168.8 | 1146.7 | -22.1 | Orc |
| 14 | warrior__arms_dragonkin | 1142.2 | 1146.0 | +3.8 | Orc |
| 15 | hunter__beast-mastery-tallstrider | 1136.9 | 1136.9 | +0.0 | Orc |
| 16 | warrior__fury | 1153.0 | 1136.4 | -16.6 | Orc |
| 17 | warrior__fury-duty-battle-shout | 1150.5 | 1125.3 | -25.2 | Orc |
| 18 | shaman__enhancement-water-shield | 1111.3 | 1118.1 | +6.8 | Orc |
| 19 | warrior__fury_2h-duty-sunder | 1134.6 | 1114.7 | -19.9 | Orc |
| 20 | paladin__level_60 (default) | 1092.1 | 1103.4 | +11.3 | Undead |
| 21 | warrior__fury-duty-sunder | 1117.7 | 1098.3 | -19.4 | Orc |
| 22 | paladin__level_60-duty-crusader | 1082.9 | 1094.9 | +12.0 | Undead |
| 23 | paladin__level_60_sub_holy | 1091.2 | 1091.6 | +0.4 | Undead |
| 24 | warrior__arms | 1097.7 | 1087.1 | -10.6 | Orc |
| 25 | paladin__level_60-duty-wisdom | 1075.3 | 1086.6 | +11.3 | Undead |
| 26 | paladin__level_60_sub_holy-duty-crusader | 1082.8 | 1083.1 | +0.3 | Undead |
| 27 | warrior__arms-duty-battle-shout | 1094.0 | 1083.0 | -11.0 | Orc |
| 28 | paladin__level_60_sub_holy-duty-wisdom | 1075.0 | 1074.9 | -0.1 | Undead |
| 29 | mage__frostfire (default) | 1064.3 | 1064.3 | +0.0 | Human |
| 30 | warlock__destruction (default) | 1063.4 | 1062.8 | -0.6 | Orc |
| 31 | shaman__enhancement-mana-tide | 1055.1 | 1060.8 | +5.7 | Orc |
| 32 | mage__fire | 1057.9 | 1057.3 | -0.6 | Gnome |
| 33 | warlock__demonology | 1053.9 | 1054.0 | +0.1 | Orc |
| 34 | warlock__destruction-duty-curse | 1052.6 | 1052.7 | +0.1 | Orc |
| 35 | warlock__destruction-duty-recklessness | 1046.4 | 1046.2 | -0.2 | Orc |
| 36 | warlock__demonology-duty-curse | 1043.6 | 1043.9 | +0.3 | Orc |
| 37 | warrior__arms-duty-sunder | 1047.1 | 1040.6 | -6.5 | Orc |
| 38 | hunter__survival | 1041.2 | 1040.4 | -0.8 | Orc |
| 39 | warlock__demonology-duty-recklessness | 1036.6 | 1036.8 | +0.2 | Orc |
| 40 | warlock__affliction | 1034.1 | 1034.4 | +0.3 | Undead |
| 41 | paladin__level_60_shock_ret | 1030.2 | 1030.2 | +0.0 | Undead |
| 42 | paladin__level_60_shock_ret-duty-crusader | 1029.3 | 1029.1 | -0.2 | Undead |
| 43 | warlock__affliction-duty-curse | 1025.1 | 1025.1 | +0.0 | Undead |
| 44 | paladin__level_60_shock_ret-duty-wisdom | 1020.0 | 1019.9 | -0.1 | Undead |
| 45 | warlock__affliction-duty-recklessness | 1019.9 | 1019.7 | -0.2 | Undead |
| 46 | warlock__affliction_shadow_bolt | 1015.6 | 1015.1 | -0.5 | Orc |
| 47 | hunter__survival-wolf | 1014.5 | 1014.0 | -0.5 | Orc |
| 48 | hunter__survival-tallstrider | 1007.4 | 1006.3 | -1.1 | Orc |
| 49 | warlock__destruction_fire | 996.7 | 997.3 | +0.6 | Undead |
| 50 | warlock__destruction_fire-duty-curse | 986.9 | 987.7 | +0.8 | Undead |
| 51 | warlock__destruction_fire-duty-recklessness | 980.7 | 981.5 | +0.8 | Undead |
| 52 | shaman__elemental-level60 (default) | 970.1 | 976.9 | +6.8 | Windshaper |
| 53 | mage__arcane | 962.3 | 969.3 | +7.0 | Gnome |
| 54 | rogue__combat-swords (default) | 958.6 | 964.5 | +5.9 | Undead |
| 55 | shaman__elemental-level60-mana-tide | 952.7 | 959.6 | +6.9 | Windshaper |
| 56 | rogue__combat-daggers | 949.8 | 954.6 | +4.8 | Undead |
| 57 | druid__balance (default) | 952.1 | 951.7 | -0.4 | Night Elf |
| 58 | priest__shadow (default) | 948.7 | 948.5 | -0.2 | Undead |
| 59 | rogue__seal-fate-swords | 940.4 | 945.2 | +4.8 | Undead |
| 60 | mage__frost | 944.7 | 943.9 | -0.8 | High Order |
| 61 | druid__balance-duty-ff | 915.5 | 915.5 | +0.0 | Night Elf |
| 62 | rogue__seal-fate-daggers | 903.4 | 909.0 | +5.6 | Undead |
| 63 | priest__shadow-holy | 906.6 | 906.4 | -0.2 | Undead |
| 64 | rogue__subtlety (default) | 881.3 | 885.9 | +4.6 | Night Elf |
| 65 | rogue__mutilate (default) | 858.3 | 863.4 | +5.1 | Undead |
| 66 | druid__balance-wrath | 861.5 | 861.5 | +0.0 | Night Elf |
| 67 | druid__balance-wrath-duty-ff | 829.0 | 828.8 | -0.2 | Night Elf |
| 68 | priest__smite | 781.0 | 780.9 | -0.1 | Undead |

The 21 regeared builds match `hit_trade/numbers.md`, apart from two known gaps:
- the warriors, by -1.1 to -1.8, from the later change that limits Unbridled Wrath to white swings;
- Enhancement, by +1.3 to +1.4, because numbers.md predates hotfix 113380.

The warriors are within -1.4 to +0.6 of `hotfix_113539_warrior.md`. Every other build is within ±1.1 of live, with these exceptions:
- Elemental: +6.8 / +6.9, from hotfix 113380.
- Feral cat: +29.8 / +30.3, which adds up as:
  - +26 from hit-trade gear;
  - +2.5 / +3.3 from the sapper in Cat Form;
  - +1.0 from Ferocious Bite.

## Race-table changes

Best race, Horde and Alliance picks that changed:

- warrior__arms-duty-sunder, bestRace: Undead (1047.2) -> Orc (1040.6)
- warrior__arms-duty-sunder, hordeRace: Undead (1047.2) -> Orc (1040.6)
- rogue__seal-fate-swords, allianceRace: Gnome (940.4) -> Night Elf (944.5)
- mage__frostfire, hordeRace: Troll (1058.5) -> Undead (1059.9)
- mage__frost, hordeRace: Troll (938.1) -> Undead (937.6)
- warlock__demonology-duty-curse, allianceRace: Human (1042.0) -> Gnome (1042.0)

Race rows moved per race (count, range): Dwarf 26 (-22.4 to +12.8), Gnome 22 (-23.1 to +7.2), High Order 19 (-20.8 to
+30.0), Human 30 (-22.5 to +11.9), Night Elf 19 (-22.3 to +29.6), Orc 28 (-23.1 to +7.0), Tauren 16 (-22.9 to +30.5), Troll
32 (-23.4 to +7.3), Undead 25 (-22.6 to +11.7), Windshaper 22 (-20.8 to +30.0). The big moves are the warrior hotfix
(down) and feral cat (up).

## Open questions

- **`warr_uw_queued`:** in `assumptions_def.py` it still says a queued Heroic Strike or Cleave procs Unbridled Wrath, which has been stale since 9173d08 (white swings only).
- **compsearch drop table:** `furious_howl` on hunter__survival-wolf was removed by hand. The generator drops it at its own seed, but TestDropsAreUnread fails with it at the test seed, so the next `-drops` run may add it back.
- **Six rogue "humanoid" builds:** their entries left the drop table because those builds are no longer in the comp-search pool. Why they are gone is not confirmed.
- **Pipeline timing:**
  - The full run hit the 2 h background limit during tankheal.
  - `pipeline.sh resume` re-ran the whole probe.
  - Healing alone takes 40 to 75 min a pass.
  - Landing took three laps because master kept moving.
- **compsearch -stale:** m1_10_class_buffs, m1_20, m1_20_class_buffs and m1_40_class_buffs are still unstamped (searched before the stamps).
