# Phase 0 re-rank: shaman_ret (Enhancement, Elemental, Retribution)

Brief `briefs/phase0_rerank_group_1125.md`. `RANK_ONLY=1 SLOTS=mainhand,offhand,ranged,trinket` on master 34fe148.
Request DPS is `verify.py` at 180s, 20000 iterations, seed 1. "Before" is the meta's published number (2026-11-25).

## Slot changes

| Spec | Slot | Old | New | List dps |
|---|---|---|---|---|
| Enhancement | trinket | Adaptive Combat Assistant | Heart of Wyrmthalak | 993.2 -> 972.57 |
| Elemental | trinket | Eye of the Beast | Talisman of Ascendance | 849.04 -> 868.91 |
| Elemental Alliance copy | trinket | Eye of the Beast | Talisman of Ascendance | 842.68 -> 866.17 |
| Ret | main hand | Warblade of Caer Darrow | Blackblade of Shahram | 944.01 -> 941.01 |
| Ret | trinket | Blackhand's Breadth | Heart of Wyrmthalak | |

- Enhancement's 993.2 predates the sim batches (the prices round didn't re-measure it). The list was ranked on
  the old Imp GW talents (below), which sim 2.0 lower than the fixed ones.
- Enhancement main hand: The Blackrock Slicer came out 1.5 (0.6) over Brain Hacker at 20000 iterations, not
  confirmed, so Brain Hacker stays.
- Talisman of Ascendance is `doubtful` (Insignia quest, live or not unknown). Standard 42 keeps it. Eye of the
  Beast is 0.48 behind (3.68 on the Alliance copy).
- Ret vs Undead (`level60_phase0_ret_undead`): Blackblade of Shahram for Warblade, 1025.55 against 1024.51.
  Heart of Wyrmthalak for Rune of the Dawn loses 11.4, Rune stays.
- Shock Ret: Blackblade loses 6.1 on Shock (898.1 against 904.3). A RANK_ONLY of main hand and trinkets on the
  Shock request puts Warblade first among the feasible weapons (Destiny is over the line), 1.2 over The Blackrock
  Slicer. Heart of Wyrmthalak and Blackhand's Breadth tie (0.04). The Shock builds get their own set,
  `level60_phase0_ret_shock` (the ret set with Warblade).
- Twist and Sub Holy check: Warblade loses 3.8 and 3.3, Blackhand's Breadth 4.0 and 3.4, Hand of Justice 1.4 and 1.5.

## Requests

| Build | Before | After |
|---|---|---|
| Enh Level 60 | 969.08 | 971.28 |
| Enh Imp GW | 960.23 | 965.39 (talent fix too) |
| Enh Mana Tide | 877.08 | 880.44 |
| Enh Water Shield | 919.15 | 923.12 |
| Ele Level 60 | 853.09 | 856.74 |
| Ele Mana Tide | 841.71 | 844.78 |
| Ret Twist | 923.62 | 930.29 |
| Twist, Crusader duty | 911.32 | 920.62 |
| Twist, Wisdom duty | 913.37 | 919.67 |
| Sub Holy | 895.99 | 901.77 |
| Sub Holy, Crusader duty | 883.48 | 891.00 |
| Sub Holy, Wisdom duty | 885.59 | 890.28 |
| Sub Holy, vs Undead | 1025.15 | 1025.55 |
| Shock | 900.96 | 904.27 |
| Shock, Crusader duty | 895.13 | 898.09 |
| Shock, Wisdom duty | 895.44 | 898.39 |

Each request wears its gear set exactly. List minus request is 1.3 (Enh, list on the Level 60 build's fight),
12.2 (Ele), 10.7 (Ret), the same gaps as the prices round (the list scores uses on 90s). Elemental Dwarf (Alliance
copy) 854.01 and 841.74 against its list's 866.17. No pick changed hit, so the race_gear rows stand.

## Set trap

`verify.py` on the Level 60 / Twist requests with the pieces swapped for each slot's best non-set item:
- Enhancement Five Thunders 6pc 971.3. 4pc with Belt of the Archmage and Timbermaw Tunic 966.9, with the Kilt,
  Bands or Face swapped instead 966.7, 966.7, 966.2. 6pc stands.
- Elemental 6pc 856.7. 4pc with Skyfury Gauntlets and Belt of the Archmage 849.8. 6pc stands.
- Ret Soulforge 4pc 930.3. 6pc (Soulforge Belt and Legplates) 925.6, 2pc (Furbolg Berserker's Boots with
  Vambraces of the Sadist or Heartflame Gauntlets) 926.4 and 926.1, none 920.2. 4pc stands.

## Open items

- `enhancement-imp-gw` talents: `053303102-054030110051212141-03` had 14 points above Flurry's row (Ancestral
  Knowledge 4). Now `053303102-055030110051202141-03`, Spirit Weapons to Ancestral Knowledge 5: 962.5 against
  960.5 (two seeds, 20000 iterations, before the trinket change). Spirit Weapons and Improved Ghost Wolf add
  nothing in the sim. Other legal fixes: Improved Ghost Wolf 1 962.5, Maelstrom Weapon 3 956.3, Improved
  Stormstrike 0 953.6, Trial by Fire 1 949.8. The Enhancement list's talents (phase0_specs, bis_phase0.sh,
  docs) moved too. If the user plays the old string in game, our tree's row rule or the string is wrong.
- Windshaper rotations on the Horde gear (Darkspear Insurgent's Spellblade): rotopt `-eval` at 120/180/300s,
  10000 iterations. Level 60 downrank at 30 mana a second 872.5 / 853.0 / 795.0, 24 ties at 180s and loses 2.1
  at 300s, 36 loses 5.2 at 180s, rank 8 in place of 9 loses 4.2 at 300s. Mana Tide at 20 841.5 / 815.6 (180 / 300s),
  16 ties at 120 and 180s and gains 1.6 at 300s only, 26 loses 14.3 at 300s, rank 9 ties. Both presets stay.
