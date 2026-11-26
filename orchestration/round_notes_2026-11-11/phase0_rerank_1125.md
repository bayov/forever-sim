# Phase 0 re-rank on the new sim (2026-11-25/26)

Weapon, relic and trinket slots re-ranked for every DPS list on the sim after db86a18dc, aed8856d1, efeb06b95 and
22d2cf952, the set trap checked, then priced (phase0_prices.py --write, AH 2026-11-24, line 30g per DPS) and the
slots over the line re-ranked. Group notes: `phase0_rerank_<group>.md`.

## Per build (request dps at the start of the round -> final, 180s)

- Arcane: staff enchant Spellblasting -> Revelation. 830.65 -> 834.05
- Fire, Frost, Frostfire, Demonology, Shadow, Smite, Beast Mastery, Marksmanship, Fury 2H: no pick change
- Affliction: main hand enchant Mighty Intellect -> Revelation. 959.6 -> 962.5, Shadow Bolt Affliction 793.4 -> 820.5
- Destruction: trinket Weakness Analyzer -> Eye of the Beast. 868.2 -> 868.0, Fire Destruction (same set) 828.9 -> 825.5
- Balance: trinket Chained Essence -> Talisman of Ascendance, off hand Obsidian Bauble -> Spirit of Aquementas (Bauble
  over the line). 796.08 -> 801.74
- Feral: Dark Bomb (does nothing to level 63 bosses) -> Goblin Land Mine in the APL. 1058.41 -> 1058.89
- Enhancement: trinket Adaptive Combat Assistant -> Blackhand's Breadth, Imp GW talents fixed
  (053303102-055030110051202141-03). L60 969.1 -> 967.0, Imp GW 960.2 -> 960.4
- Elemental: trinket Eye of the Beast -> Talisman of Ascendance. 853.1 -> 856.7
- Retribution: main hand Warblade of Caer Darrow -> Destiny. Twist 923.6 -> 932.4, Sub Holy 896.0 -> 904.8. Shock
  keeps Warblade on its own set level60_phase0_ret_shock, 901.0 -> 900.9. vs Undead 1025.2 -> 1024.5
- Survival: no net change, 950.39
- Lone Wolf: ranged Eaglehorn Long Bow (over the line) -> Hawkeye Bow of Agility. 1140.13 -> 1135.46
- Arms: no net change, 1011.08 -> 1010.49
- Fury DW: main hand Ironwood Blade -> Thrash Blade, trinkets Maelstrom + Blackhand's Breadth, ranged Satyr's Bow
  (over the line) -> Bloodseeker. 983.62 -> 986.50
- Combat Swords: 840.96 -> 840.39
- Seal Fate Swords: ranged Deep Jungle Rifle (over the line) -> Ancient Bone Bow, trinket Eranikus -> Maelstrom after the audit (+8.9). 848.83 -> 848.3
- Combat Daggers, Mutilate, Subtlety, Seal Fate Daggers: Devilsaur Gauntlets and Leggings for the Darkmantle pair
  (Devilsaur 2 beats Darkmantle 6). 828.81 -> 829.20, 785.09 -> 787.14, 786.46 -> 787.33, 818.47 -> 820.02

Heart of Wyrmthalak (211g, LBRS) topped the trinket ranks of 12 lists but is over the line on every one (46.6 to
251.2 g/DPS), so none wears it. Blackblade of Shahram is 424.4 g/DPS on ret (set by hand, phase0_prices.py gave it
none).

## Open items

- Dark Bomb: does nothing, Goblin Land Mine +0.5 over none, kept (feral APLs).
- enhancement-imp-gw: 14 points above Flurry's row, fixed. If the old string works in game, our tree rule is wrong.
- Windshaper rotations on the Horde gear: both downrank presets hold.
- Hunter hit 6.6%: no hit item under the line is missing.
- Virtuous Cowl, Hyper Deluxe Sniper Rifle, Bloodseeker: include under standard 42 (only availability unsure).
- Faction copies and build variants (ret_shock, ret_undead) are now priced by phase0_prices.py --sets.
- Crystal-Infused Cleaver: quest reward (95925, 95926), 0g, not a pick.
