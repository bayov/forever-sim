# Fairness between healers

The healer_fairness auditor keeps this current (`orchestration/audit/passes/healer_fairness.md`). It holds the
healer ranking in each role, its reasons and the open questions. Last pass: 8, 2026-11-12, master cb0d928fd6: quiet (priest Shadow tree size and shaman Trial by Fire rename reach no healer preset; spot cells identical). Pass 7, 2026-11-11, master 707ae92cc5 (Nature's Swiftness on Wild Growth drops the raid druid to last, healer_fairness-20261111-1; other spot cells identical). Pass 6, master ca5e35c3f: shaman changes reach no healer. Pass 5, master 9ba00c2dc: whole-point rolls moved cells by at most 0.002; the raid shaman 1.395.
Details are in `orchestration/audit/reports/healer_fairness.md`; findings are in
`orchestration/audit/findings/healer_fairness.md`.

## The ranking (healbench carry, mean of the 7 boss files, 1,000 iterations, each file at its own length)

| Role | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Tank healer | Holy priest 1.262 | Discipline priest 1.261 | Shaman 1.242 | Druid 1.235 | Paladin 1.176 |
| Raid healer | Holy priest 1.437 | Paladin 1.413 | Shaman 1.395 | Druid 1.391 (1.437 before 9628e85b6) | |

## Why

- **Tank role: four of five within 2%.** With the tank bench's raid kept (healer_fairness-20261104-1, fixed) and the
  gates re-tuned (-2, fixed), the Holy priest's lead over the shaman fell from +0.117 to +0.020. What is left is
  kit, chiefly Spiritual Healing. The tanks now face the boss, which lifted every column by 0.09-0.20 alike.
  - The paladin is 7% behind, last on 4 of 7 files, with mana to spare: its slow Holy Light is cast-time bound.
  - The Discipline priest runs out of mana on Chromaggus at carry (101 s; healer_rotations-20261104-2).
- **Raid role: the druid's preset now spends Nature's Swiftness on Wild Growth** (9628e85b6 made it instant; the
  preset autocasts NS and Wild Growth comes before Healing Touch), costing 0.046 carry, most on Magmadar (2.002 -> 1.818)
  and Onyxia (0.997 -> 0.943); healer_fairness-20261111-1. Before that the role was level but for the shaman (3% behind)
  and the paladin's preset. The paladin's raid preset loses
  0.025 to its own Holy Light lines; a Flash of Light raid paladin carries 1.438, level with the first two
  (healer_fairness-20261104-5). The file decided the order (pass 6): the priest led Magmadar and Ragnaros, the
  druid Onyxia, Broodlord and Chromaggus, the paladin Golemagg.
- **Not the cause:**
  - Gear: all four sets at the same 902 weighted budget; slot 16 is stat-less for all.
  - Race: each class has its best legal race; the shaman's other races are within 0.002.
  - Consumables, buffs, debuffs, encounter: identical in all 9 requests.
  - Mana: only the Disc tank preset on Chromaggus runs out at carry.

## Against the logs

The Anniversary kills bring about 5.1 priests, 3.0 shamans, 1.7 paladins and 1.2 druids per 11 healers. The boss
files keep no healing by class, so the sim's order can't be checked against the logs' (healer_fairness-20261104-4).

## Open questions

1. The druid raid preset's Nature's Swiftness (finding 20261111-1): cast it before Healing Touch, not on Wild Growth.
2. The paladin raid preset's re-tune (finding -5): drop or gate Holy Light, Flash of Light 6 at 0.7x.
3. HPS by healer class in the boss files (finding -4), to check the order against the logs.
4. The classes' "Calls for the user" lists (`docs/healing_{priest,paladin,druid,shaman}.md`) are open rulings.
