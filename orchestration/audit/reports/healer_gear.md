# Healer gear pass 1 (master a1b1163d2)

Result: quiet, no findings. Forced first pass after the healers merged; `last.master` was a seed (36cd929a2), so
everything was checked, not just a diff. Recomputed with a parser over the `.go` files (weights Sta/Int/Spi/crit 1,
SP 0.86, Healing 0.45, MP5 2.4; L = 1.125 ilvl + 2; pure caster f 1.053 / 1.026 / 0.995; shape 0.661 ilvl + 2.92).

## Pieces (linear vs slot mod x L x f; 1.7-norm vs shape target; limits +-1% / +-6%)

| Set | Armor type | Linear range | Norm range | Weapon(s) vs rate | Verdict |
|---|---|---|---|---|---|
| Shaman 990201-13 | mail on all armor pieces | -0.4% .. +0.9% | -2.8% .. +0.7% | staff 193.2 vs 193.4 (-0.1%) | ok |
| Paladin 990301-14 | plate | -0.4% .. +0.3% | -2.8% .. +1.7% | MH 148.9 vs 149.0, OH 43.2 vs 43.4 | ok |
| Druid 990401-13 | leather | -0.5% .. +0.3% | -2.3% .. +3.9% (neck, both rings) | staff 193.9 vs 193.4 (+0.3%) | ok |
| Priest 990501-14 | cloth | -0.4% .. +0.5% | -3.2% .. -0.7% | staff 194.0 vs 193.4; wand 21.4 vs 21.3 | ok |

Item levels 66 (head/shoulder/chest/hands/legs/feet), 71 (neck/back/wrist/waist/rings), 76 staff, 71 hand/wand:
as standards 1b. No Hit on any piece. Druid neck/rings sit at +3.5..+3.9% on the norm: inside the 6% band.

## Set shares (no trinkets; target = healer row of gear_targets.md; tolerance Sta 1.5, rest 3)

| Set | Sta | Int | Spi | SP | Heal | Crit | MP5 | top stat / non-Sta |
|---|---|---|---|---|---|---|---|---|
| Shaman set (row) | 10.6 (10.8) | 17.5 (17.5) | 10.1 (10.0) | 22.4 (22.2) | 26.9 (27.1) | 7.0 (7.0) | 5.6 (5.4) | Heal 30.1% |
| Paladin set (row) | 14.9 (14.7) | 18.8 (18.7) | 10.0 (10.0) | 21.7 (21.9) | 26.9 (26.8) | 2.8 (2.8) | 5.1 (5.1) | Heal 31.6% |
| Druid set (row) | 14.9 (14.8) | 18.5 (18.7) | 9.9 (10.0) | 17.5 (17.5) | 35.4 (35.2) | 1.0 (1.0) | 2.9 (2.8) | Heal 41.6% |
| Priest set (row) | 12.4 (12.1) | 15.0 (15.3) | 10.3 (10.0) | 26.8 (27.5) | 21.1 (21.1) | 6.0 (5.9) | 8.5 (8.1) | SP 30.6% |

All within tolerance; Spirit about 10%; guardrail (55%) holds with room.

## Trinkets, relics, enchants

- Trinkets are the DPS casters' (990114/5 Elemental, 992014/5 Retribution, 995033/4 Priest, 998014/5 Balance). Stat
  trinkets recompute to 45.8 / 45.5 / 46.0 / 46.0 linear (target 45.9 +-1%) and norm 36.2 / 34.7 / 36.0 / 36.0
  (target 36.8 +-6%; Retribution's -5.7% is the closest). Each on-use has a sim effect (`sim/*/forever_synthetic_trinkets.go`).
  They carry crit, so a healer with a 1% crit row (druid) gets crit from trinkets; shares exclude trinkets (5a), so fine.
- Relics modelled: Polished Driftwood Icon 249398 (`sim/shaman/items.go:169`, 8% casting regen), Libram of Economy
  272436 (`sim/paladin/healing.go:233`, Holy Light -5%), Idol of Swiftness 279250 (`sim/druid/items_restoration.go`,
  Swiftmend -3 s). All three are crafted or Stranglethorn drops, Phase 1. Priest: synthetic wand 990514.
- Enchants (same as the DPS casters' where the slot matches): Lesser Arcanum of Voracity (head), Necklace - Healing
  Power (heals, unlike the casters' Spell Power), Cloak - Greater Resistance (the DPS casters' too), Bracer - Healing
  Power (+24 heal), Mystic Rugged Armor Kit on chest/legs/feet, 2H Spellblasting (+65 SP; beats Weapon - Healing
  Power 55 x 0.45 = 24.8 against 55.9). Gloves unenchanted: Gloves - Healing Power is Phase 5, the Forever school
  "Power" gloves enchants are damage-only. Paladin: MH Mighty Intellect (22) vs Weapon - Healing Power (55 x 0.45 =
  24.8): a 2.8-point (0.3%) call not worth a finding; OH Wisdom. Rings unenchanted (need an Enchanter).

## Rows (gear_targets.md, Healers)

- Every row sums to 100.0 and its top stat / non-Sta column recomputes (30.4 / 31.4 / 41.3 / 31.3%).
- Priest and paladin rows re-derived from the sources' numbers quoted in the doc (priest = average of the Classic
  healing priest list and the Forever shaman list, Spirit reset to 10%; paladin = Classic list, healing split 45/55
  after the Spirit renormalisation): match within 0.7 points.
- NOT done: recomputing from the cached Wowhead pages. `~/.cache/wow-data/wowhead/pages` is empty in this
  container and the pass may not fetch. The druid row (no quoted source numbers in the doc) and the shaman
  row's list picks are unchecked beyond arithmetic. A later pass in a session with the cache should do it.
- The docs' "Calls for the user" lists were not read for factual errors this pass.

## Next pass
Watch for the sets being re-tuned (the synth scripts in `orchestration/tools/healers/`); re-run the same
parser (not committed; ~60 lines, regexes over `Id: ... Name` blocks) and the row recompute with a warm cache.
