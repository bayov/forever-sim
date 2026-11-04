# Gear pass 53 (master f5c99ad018, client 1.60.1.70753)

Result: quiet. Changed since 9da90fcde1: the Restoration shaman set (new, commit 57061b563) and its standards 35 /
gear_targets.md "Healers" row; standards 36/37 (tanks, notes values) are not gear-budget rules for the DPS sets and the
tank sets belong to the tank_gear pass. forever_enchants.json: build bump to 70753, one new Speed Increase enchant
(1328948), one name fix; no enchant a preset uses changed. No other synthetic item or gear.json changed.

## Restoration shaman (990201-990213, ui/restoration_shaman/gear_sets/level60_phase1_synthetic.gear.json)
Recomputed by hand with the 1b weights (SP 0.86, healing 0.45, MP5 2.4); lin2.py/std.py do not know healing power.

| Check | Result |
|---|---|
| Linear per piece vs slot x L x caster f | all within +/-0.9% (finger II +0.9%, rest <= 0.4%) |
| 1.7-norm vs shape target | -2.8% to +0.7% (limit 6%) |
| Staff 990213 | 193.2 = 2.208 x L(76); damage/speed equal to the Elemental staff (2.9 s) |
| Item levels | 66 / 71 / 76 as item 2 |
| Mix vs row (Sta 10.8, Int 17.5, Spi 10, SP 22.2, Heal 27.1, Crit 7, MP5 5.4) | 10.6 / 17.5 / 10.1 / 22.4 / 26.9 / 7.0 / 5.6; top stat Healing 30% |
| Hit, Strength | none, as the row says |
| Set bonuses | none (all synthetic) |
| Enchants | same caster pattern as Elemental/Mage (Arcanum head, Rugged kits chest/legs/feet, Spellblasting 2H, Greater Resistance cloak), plus necklace Healing Power and bracer Healing Power |
| Trinkets | Elemental's 990114/990115; trinkets.py shows no '!' |
| Ranged | Polished Driftwood Icon 249398 (crafted, Forever), 8% mana regen while casting |

Notes, not findings:
- Hands carry no enchant. Era's Gloves - Healing Power (2617, +30) exists in the db but is a late-Era recipe; Elemental's Nature Damage glove enchant does nothing for heals. Left as is.
- The row's source numbers (two Wowhead lists) were not re-derived; its shares sum to 100.
- Tool fixes: extract.sh now also extracts forever_synthetic_items_shaman_restoration.go. lin2.py weights healing at 1 so
  it prints false 1.7-norm excesses (+20%) for the Restoration set, and has no healing target row.
