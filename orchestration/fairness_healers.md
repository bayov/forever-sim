# Fairness between healers

The healer_fairness auditor keeps this current (`orchestration/audit/passes/healer_fairness.md`). It holds the
healer ranking in each role, its reasons and the open questions. Last pass: 22, 2026-11-19, master 8f388c7f3: quiet; 65f0ea9d6's allocation-free Swiftmend HoT pick is behaviour-neutral, the druid's 14 cells identical to pass 21. Pass 21, 2026-11-18, master a6677362f: quiet; all 63 cells re-benched (`orchestration/audit/ref/healer_fairness/bench_a667736.tsv`) after 7199c8f9c made an idling healer see a hit a reaction time late instead of up to several wake-ups late: the presets that idle moved (tank shaman 1.257 -> 1.280, now first by 0.002; raid shaman 1.412 -> 1.390, Holy raid 1.461 -> 1.472, paladin raid 1.430 -> 1.418, druid raid 1.392 -> 1.385), the tank paladin and tank priests identical; the hasted spell GCD reaches no healer preset (no haste in their requests). Pass 20, 2026-11-18, master 8219f1083: quiet; Forever's hasted dot ticks (Hurricane, Volley) reach no HoT or healer channel, 28 cells (druid both roles, raid shaman, tank Holy priest) identical. Pass 19, 2026-11-18, master c198d81e3: quiet; Dark Inspiration's given form now drops for a party whose priest has the talent, and no healer preset has Shadow points or the party buff, so inert; 35 cells re-run (priest tank, every raid healer but the druid) identical. Pass 18, 2026-11-18, master 2f7e255af: quiet; the update_1117 items (Idol of Rejuvenation, Totem of Tides, Darkmoon Blue Dragon and Heroism) and the Holy Strike and Vampiric Embrace hotfixes reach no healer preset, 39 of 63 cells re-run identical; Idol of Rejuvenation and Totem of Tides screened as relics, neither beats the shipped one past noise. Pass 17, 2026-11-17, master 5f20dfdf4: quiet; Forever's Five Thunders set (1679) is worn by no healer preset, all 63 cells identical to pass 15. Pass 16, 2026-11-16, master f30078da1: quiet; standards 17 (Windfury Weapon and Totem cooldowns) and the item DB refresh reach no healer, all 63 cells identical to pass 15. Pass 15, 2026-11-15, master 25ce81de7: findings; Wild Growth's tick doubled (0.033 -> 0.066, client 70938) and the raid druid fell 1.414 -> 1.392, third to last, because its preset isn't re-tuned for it (healer_fairness-20261115-1); the other 56 cells identical (`orchestration/audit/ref/healer_fairness/bench_25ce81d.tsv`). Pass 12, 2026-11-13, master 97da1cca5: quiet; the Unsealed Wrath and Windfury extra-attack rules reach no healer, all 63 cells identical to pass 11. Pass 11, 2026-11-13, master 71562a8b0: quiet; the whole bench re-run after the tankheal merge (`orchestration/audit/ref/healer_fairness/bench_71562a8.tsv`), the 7 unchanged presets identical, the raid priest and paladin at their re-tuned 1.461 and 1.430, the table below now measured on master. Pass 9, 2026-11-12, master 4fb90c298: 1 new finding (the Holy priest raid preset keeps Inspiration, dead without a tank, instead of Twilight Focus: +0.021 raid carry, healer_fairness-20261112-1); the table below re-benched and confirmed, a talent screen of each class's likeliest swaps found no other gain. Pass 8, 2026-11-12, master cb0d928fd6: quiet (priest Shadow tree size and shaman Trial by Fire rename reach no healer preset; spot cells identical). Pass 7, 2026-11-11, master 707ae92cc5 (Nature's Swiftness on Wild Growth drops the raid druid to last, healer_fairness-20261111-1; other spot cells identical). Pass 6, master ca5e35c3f: shaman changes reach no healer. Pass 5, master 9ba00c2dc: whole-point rolls moved cells by at most 0.002; the raid shaman 1.395.
Details are in `orchestration/audit/reports/healer_fairness.md`; findings are in
`orchestration/audit/findings/healer_fairness.md`.

## The ranking (healbench carry, mean of the 7 boss files, 1,000 iterations, each file at its own length)

| Role | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Tank healer | Shaman 1.280 | Holy priest 1.278 | Discipline priest 1.272 | Druid 1.267 | Paladin 1.152 |
| Raid healer | Holy priest 1.472 | Paladin 1.418 | Shaman 1.390 | Druid 1.385 (awaiting re-tune for the 0.066 Wild Growth) | |

Measured on master a6677362f (pass 21, all 63 cells, seed 101). The top four tank healers are within 1%, so their order is
within seed noise. Before (pass 15-20 table): tank Holy 1.278, Disc 1.272, druid 1.269, shaman 1.257, paladin 1.152;
raid Holy 1.461, paladin 1.430, shaman 1.412, druid 1.392.

Confirmed on master 8219f1083 (pass 20, 28 cells: druid tank and raid, raid shaman, tank Holy priest, all identical), c198d81e3 (pass 19, 35 cells: priest tank, raid priest, paladin and shaman) and 2f7e255af (pass 18, 39 of 63 cells: all tank, raid druid), together all 63 identical; 5f20dfdf4 (pass 17) and f30078da1 (pass 16), all 63 cells identical, and 25ce81de7 (pass 15; only the raid druid moved). Before: master 97da1cca5 (pass 12) and 71562a8b0 (pass 11), all 63 cells. Re-benched on 4fb90c298 (pass 9, `orchestration/audit/ref/healer_bench/summary.tsv`); the two raid fixes of
2026-11-12 on top (seed 101, `orchestration/audit/ref/healer_fairness/retune_25f65da_s101.tsv`): the Holy raid
priest takes Twilight Focus for Inspiration, 1.440 -> 1.461 (finding -20261112-1), and the raid paladin is a Flash
of Light healer (6 and 4 at 0.7x, no Holy Light or Divine Favor), 1.400 -> 1.430 (finding -5); both on two seeds.
An enchant screen (head Lesser Arcanum of Voracity Spirit or Rumination for Intellect, legs Voracity Intellect for
the Mystic Rugged Armor Kit, a kit on the bare hands, chest Greater Stats; Holy tank, shaman raid and paladin tank,
seed 101) moved no mean by more than 0.006 either way, so the presets' enchants stay.

2026-11-12, on the new consumables (Consumables in `docs/healing_sim.md`), seed 101, after the healer_rotations
fixes of that day (druid tank NS at 45%, shaman raid NS held for an instant Healing Wave under 65%, the shaman
tank's dead Lesser Healing Wave 5 line dropped). The rows before them (master e05135ad3, old consumables): tank
druid 1.273, Holy 1.262, Discipline 1.261, shaman 1.242, paladin 1.176; raid druid 1.438, Holy 1.437, paladin
1.413, shaman 1.395.

## Why

- **Tank role: four of five within 1%** (pass 21; 3% before the healers' reaction-time fix 7199c8f9c) (2% before the druid's re-tune of e05135ad3). With the tank bench's raid kept (healer_fairness-20261104-1, fixed) and the
  gates re-tuned (-2, fixed), the Holy priest's lead over the shaman fell from +0.117 to +0.020. What is left is
  kit, chiefly Spiritual Healing. The tanks now face the boss, which lifted every column by 0.09-0.20 alike.
  - The paladin is 7% behind, last on 4 of 7 files, with mana to spare: its slow Holy Light is cast-time bound.
  - The Discipline priest runs out of mana on Chromaggus at carry (101 s; healer_rotations-20261104-2).
- **Druid presets re-tuned for Nature's Swiftness on Wild Growth** (healer_fairness-20261111-1, fixed in e05135ad3):
  9628e85b6 made Wild Growth instant under NS and the raid preset's autocast NS went on it (mean 1.437 -> 1.391).
  Both druid presets now hold NS on their own line (raid: instant Regrowth under 45%; tank: instant Healing Touch 11
  under 60%) and take Heart of the Wild 2 for Subtlety 2: raid 1.438, level first with the Holy priest; tank 1.273,
  first (+0.011). Both gains come from holding an emergency cooldown; the shaman's raid preset now holds its
  Nature's Swiftness too (2026-11-12, +0.020). Before 9628e85b6 the raid role was level but for the shaman (3% behind)
  and the paladin's preset. The paladin's raid preset lost
  0.03 to its own Holy Light lines; as a Flash of Light healer it is second, 1.430 (healer_fairness-20261104-5,
  fixed 2026-11-12). The Holy priest's raid lead (+0.031 over the paladin) is Twilight Focus on Vaelastrasz's
  raid-wide fire plus Prayer of Healing. The file decided the order (pass 6): the priest led Magmadar and Ragnaros, the
  druid Onyxia, Broodlord and Chromaggus, the paladin Golemagg.
- **Not the cause:**
  - Gear: all four sets at the same 902 weighted budget; slot 16 is stat-less for all.
  - Race: each class has its best legal race; the shaman's other races are within 0.002.
  - Talents (pass 9 screen, own role, 1,000 it): the shaman's Healing Focus (raid) and Ancestral Healing (tank)
    picks are right by 0.039 and 0.085, the druid's Nature's Splendor over Heart of the Wild 5 by 0.011 / 0.016. The
    Holy raid priest is the one raid preset without its pushback talent (Twilight Focus; the shaman, paladin and
    druid take Healing Focus, Spiritual Focus, Nature's Focus): -0.021, all of it Vaelastrasz (+0.137 with it).
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

1. (Closed 2026-11-12: the shaman raid preset now holds Nature's Swiftness, +0.020 carry, healer_rotations-20261112-2.)
2. (Closed 2026-11-12: the paladin raid preset is a Flash of Light healer, 1.400 -> 1.430, finding -5.)
3. (Closed 2026-11-12: the Holy raid priest takes Twilight Focus 3 for Inspiration 3, 1.440 -> 1.461, finding -20261112-1.)
4. HPS by healer class in the boss files (finding -4), to check the order against the logs.
6. The raid druid preset for the doubled Wild Growth (finding -20261115-1): a gate at the old threshold wins Vaelastrasz (+0.16) and loses Golemagg (-0.18), mean 1.394; it needs a rotation re-tune, and the tank preset may want a Wild Growth line.
5. The classes' "Calls for the user" lists (`docs/healing_{priest,paladin,druid,shaman}.md`) are open rulings.
