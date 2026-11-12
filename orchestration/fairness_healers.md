# Fairness between healers

The healer_fairness auditor keeps this current (`orchestration/audit/passes/healer_fairness.md`). It holds the
healer ranking in each role, its reasons and the open questions. Last pass: 8, 2026-11-12, master cb0d928fd6: quiet (priest Shadow tree size and shaman Trial by Fire rename reach no healer preset; spot cells identical). Pass 7, 2026-11-11, master 707ae92cc5 (Nature's Swiftness on Wild Growth drops the raid druid to last, healer_fairness-20261111-1; other spot cells identical). Pass 6, master ca5e35c3f: shaman changes reach no healer. Pass 5, master 9ba00c2dc: whole-point rolls moved cells by at most 0.002; the raid shaman 1.395.
Details are in `orchestration/audit/reports/healer_fairness.md`; findings are in
`orchestration/audit/findings/healer_fairness.md`.

## The ranking (healbench carry, mean of the 7 boss files, 1,000 iterations, each file at its own length)

| Role | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Tank healer | Druid 1.273 (1.235 before e05135ad3) | Holy priest 1.262 | Discipline priest 1.261 | Shaman 1.242 | Paladin 1.176 |
| Raid healer | Druid 1.438 (1.391 after 9628e85b6, before e05135ad3) | Holy priest 1.437 | Paladin 1.413 | Shaman 1.395 | |

## Why

- **Tank role: four of five within 3%** (2% before the druid's re-tune of e05135ad3). With the tank bench's raid kept (healer_fairness-20261104-1, fixed) and the
  gates re-tuned (-2, fixed), the Holy priest's lead over the shaman fell from +0.117 to +0.020. What is left is
  kit, chiefly Spiritual Healing. The tanks now face the boss, which lifted every column by 0.09-0.20 alike.
  - The paladin is 7% behind, last on 4 of 7 files, with mana to spare: its slow Holy Light is cast-time bound.
  - The Discipline priest runs out of mana on Chromaggus at carry (101 s; healer_rotations-20261104-2).
- **Druid presets re-tuned for Nature's Swiftness on Wild Growth** (healer_fairness-20261111-1, fixed in e05135ad3):
  9628e85b6 made Wild Growth instant under NS and the raid preset's autocast NS went on it (mean 1.437 -> 1.391).
  Both druid presets now hold NS on their own line (raid: instant Regrowth under 45%; tank: instant Healing Touch 11
  under 60%) and take Heart of the Wild 2 for Subtlety 2: raid 1.438, level first with the Holy priest; tank 1.273,
  first (+0.011). Both gains come from holding an emergency cooldown; the shaman's raid preset still autocasts
  its Nature's Swiftness. Before 9628e85b6 the raid role was level but for the shaman (3% behind)
  and the paladin's preset. The paladin's raid preset loses
  0.025 to its own Holy Light lines; a Flash of Light raid paladin carries 1.438, level with the first two
  (healer_fairness-20261104-5). The file decided the order (pass 6): the priest led Magmadar and Ragnaros, the
  druid Onyxia, Broodlord and Chromaggus, the paladin Golemagg.
- **Not the cause:**
  - Gear: all four sets at the same 902 weighted budget; slot 16 is stat-less for all.
  - Race: each class has its best legal race; the shaman's other races are within 0.002.
  - Buffs, debuffs, encounter: identical in all 9 requests. Consumables: the same tier and slots in all 9, each
    preset's food and oil its own best by its objective (2026-11-12, healer_gear-20261112-1 to -3; Greater
    Cleric's Elixir, Supreme Power everywhere; Wicked Smoothie for the druid, Runn Tum Tuber Surprise for the
    rest; Wizard Oil, Lesser Mana Oil for the paladin raid and Holy tank presets; `docs/healing_sim.md`,
    Consumables). Before and after, 1,000 iterations seed 101: tank druid 1.237 -> 1.242, shaman 1.248 -> 1.257,
    paladin 1.149 -> 1.152, Holy 1.263 -> 1.278, Discipline 1.261 -> 1.272; raid druid 1.419 -> 1.414, shaman
    1.389 -> 1.392, paladin 1.394 -> 1.400, Holy 1.440 -> 1.440; the order in each role unchanged.
  - Mana: only the Disc tank preset on Chromaggus runs out at carry.

## Against the logs

The Anniversary kills bring about 5.1 priests, 3.0 shamans, 1.7 paladins and 1.2 druids per 11 healers. The boss
files keep no healing by class, so the sim's order can't be checked against the logs' (healer_fairness-20261104-4).

## Open questions

1. The shaman raid preset autocasts Nature's Swiftness; the druids' held NS gained 0.04 carry (tank and raid).
2. The paladin raid preset's re-tune (finding -5): drop or gate Holy Light, Flash of Light 6 at 0.7x.
3. HPS by healer class in the boss files (finding -4), to check the order against the logs.
4. The classes' "Calls for the user" lists (`docs/healing_{priest,paladin,druid,shaman}.md`) are open rulings.
