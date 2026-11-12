# Healer fairness pass 9 (2026-11-12, master 4fb90c298)

Result: 1 new finding. healer_fairness-20261112-1 (high): the Holy priest raid preset keeps Inspiration (dead without a
tank) instead of Twilight Focus; +0.021 mean raid carry with the swap. Still open: -20261104-5 (high, paladin raid
APL, now +0.024), -4 (low). Forced re-audit after the day's healer changes (druid NS holds and Heart of the Wild 2,
shaman raid NS for Healing Wave under 65%, the seen-health engine fix, the new consumables). HEAD moved to a2d1dceaa
during the pass with audit files only, so the numbers hold for it too.

## 1. Same terms (the 9 requests)

Everything but the healer row is byte-identical in all 9 (md5 of the request without the player): encounter, raid,
buffs, debuffs, sim options. Player rows: level 60, Engineering + Alchemy, Blessing of Kings and Improved Wisdom,
reaction 200 ms (healbench overrides 0.3 s ±25%), 20 yd. Consumables: Supreme Power, Greater Mageblood, Greater
Cleric's Elixir, Greater Intellect, Sages, Cerebral Cortex, Major Mana Potion, Demonic Rune in all 9; food and oil
each preset's own pick by objective (`docs/healing_sim.md`, Consumables, healer_gear-20261112-1 to -3): Wicked
Smoothie for the druid, Runn Tum Tuber Surprise for the rest; Lesser Mana Oil for paladin raid and Holy tank, Wizard
Oil elsewhere. Race: Human priest/paladin, Skyborne Windshaper shaman/druid (race screens in the class docs, all
within 0.02). Slot 16 stat-less for all; the paladin's 990314 off-hand is inside its budget.

Food the gear pass did not screen for the shaman and paladin (Wicked Smoothie for Runn Tum Tuber, raid role):
shaman 1.412 -> 1.415 at seed 101, 1.411 -> 1.411 at seed 202; paladin 1.400 -> 1.399. Level; no finding.

## 2. The numbers (healbench -carry, 1,000 iterations, seed 101, each file at its own length)

Regenerated `ref/healer_bench/summary.tsv` (tools/audit/healer_rotations/baseline.sh; healbench needs `-tags with_db`).

| Request | Onyxia | Brood | Chrom | Golem | Magma | Rag | Vael | Mean | Rank |
|---|---|---|---|---|---|---|---|---|---|
| priest_holy (tank) | 1.378 | .931 | .970 | 2.000 | 2.154 | .600 | .910 | 1.278 | 1 |
| priest_discipline (tank) | 1.363 | .958 | .871 | 2.027 | 2.143 | .602 | .938 | 1.272 | 2 |
| druid_restoration (tank) | 1.403 | .893 | 1.023 | 1.951 | 2.045 | .634 | .932 | 1.269 | 3 |
| shaman_restoration (tank) | 1.340 | .978 | .984 | 1.938 | 2.062 | .587 | .911 | 1.257 | 4 |
| paladin_holy (tank) | 1.241 | .883 | .908 | 1.806 | 1.867 | .568 | .789 | 1.152 | 5 |
| priest_holy_raid | .978 | .923 | 1.066 | 2.608 | 2.054 | 1.328 | 1.125 | 1.440 | 1 |
| druid_restoration_raid | .991 | .924 | 1.094 | 2.630 | 1.829 | 1.278 | 1.152 | 1.414 | 2 |
| shaman_restoration_raid | .946 | .885 | 1.115 | 2.674 | 1.887 | 1.236 | 1.138 | 1.412 | 3 |
| paladin_holy_raid | .977 | .900 | 1.087 | 2.684 | 1.903 | 1.290 | .957 | 1.400 | 4 |

Against the last baseline (8f3f74f3f): only druid tank 1.242 -> 1.269 (NS at 45%) and shaman raid 1.392 -> 1.412
(held NS); the other 7 means are unchanged. Spread: tank 1.152-1.278 (paladin 10% behind the leader, cast-time bound
with mana to spare, as before), raid 1.400-1.440 (3%). Only the Disc tank preset runs out of mana (Chromaggus, 101 s).

## 3. Talent screen (one swap per likely choice, own role, 1,000 it, paired seed 101)

No healer talent-search tool exists; covered the choices most likely to move carry. Rows in
`ref/healer_fairness/screen_4fb90c2.tsv`, helper `tools/audit/healer_fairness/screen.sh`.

| Preset | Swap | Mean | Change | Verdict |
|---|---|---|---|---|
| Holy priest raid | Inspiration 3 -> Twilight Focus 3 | 1.461 (s202 1.469) | +0.021 (+0.020) | **finding -20261112-1** (Vael +0.137) |
| Holy priest tank | same | 1.205 | -0.073 | preset right |
| Shaman raid | Healing Focus 3 -> Ancestral Healing 3 | 1.373 | -0.039 | preset right (Vael -0.156) |
| Shaman tank | Ancestral Healing 3 -> Healing Focus 3 | 1.172 | -0.085 | preset right |
| Druid raid | Nature's Splendor + Moonglow -> Heart of the Wild 5, Subtlety 1 (`05002--5553015153113051`) | 1.403 | -0.011 | preset right |
| Druid tank | same | 1.253 | -0.016 | preset right |

Not simmed, by reading the trees: the paladin's 10 dead points (Divine Strength 3, Improved / Sanctified Judgement,
Pursuit of Justice) have no healing talent to go to within reach; the priest's Improved Power Word: Shield 1 and
Holy Nova 1 are gate fillers with no healing alternative in their rows; the druid's Moonglow 3 is the gate to Nature's
Splendor (the screen above prices it). Heart of the Wild 2 for Subtlety and Nature's Swiftness were measured by the
druid re-tune (e05135ad3). Disc was re-screened by healer_rotations (8fd5ddcb9).

## 4. Rotation / role picks

- Paladin raid (open -5): the tank preset's APL in the raid role, with the raid oil, 1.424 against 1.400 (+0.024; 6 of 7
  files up, Ragnaros -0.031); dropping only the Holy Light 9/8 and Divine Favor lines 1.403 (+0.003). The gain is now
  mostly the Flash of Light gate, not the Holy Light lines.
- Each role preset is benched only in its own role (baseline.sh), at its file's length.

## 5. Against the logs

Unchanged: the boss files keep no HPS by class (-4). Composition about 5.1 priests, 3.0 shamans, 1.7 paladins,
1.2 druids per 11 healers; the sim's raid order Holy > druid ≈ shaman > paladin agrees with priests and shamans being
taken most, but can't be checked further.
