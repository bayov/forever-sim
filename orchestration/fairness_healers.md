# Fairness between healers

The healer_fairness auditor keeps this current (`orchestration/audit/passes/healer_fairness.md`). It holds the
healer ranking in each role, its reasons and the open questions. Last pass: 1, 2026-11-04, master c8f5858bd.
Details are in `orchestration/audit/reports/healer_fairness.md`; findings are in
`orchestration/audit/findings/healer_fairness.md`.

## The ranking (healbench carry, mean of the 7 boss files, 1,000 iterations, each file at its own length)

| Role | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Tank healer | Holy priest 1.220 | Discipline priest 1.149 | Shaman 1.103 | Druid 1.040 | Paladin 1.028 |
| Raid healer | Holy priest 1.414 | Paladin 1.413 | Druid 1.413 | Shaman 1.373 | |

## Why

- **Raid role: the four are level.** The first three are within 0.001, and the shaman is 3% behind, last on
  Onyxia, Broodlord, Magmadar and Ragnaros. No one runs out of mana. The file decides the order: the priest leads
  Magmadar, Ragnaros and Vaelastrasz, the paladin Onyxia and Golemagg, the druid Chromaggus.
- **Tank role: the Holy priest's lead (+0.117 over the shaman) is mostly the bench, not the class.**
  - Without others in the raid, Prayer of Mending bounces between the tank and the priest. Holy Binding sets it
    off on the priest, so it is worth 0.085 to the priest. With the raid's other members present it is worth
    about 0 (healer_fairness-20261104-1).
  - The presets' shared "heal only when the target misses the whole heal" gate costs the shaman 0.026, the paladin
    0.018 and the druid 0.011, against the priest's 0.005 (healer_fairness-20261104-2).
  - Inspiration counting all armor while Ancestral Fortitude counts item armor only is worth about 0.01
    (healer_mechanics-20261104-3).
  - What is left is the priest's kit, chiefly Spiritual Healing: about +0.02 to +0.05 over the shaman.
- **Not the cause:**
  - Gear: all four sets are at the same 902 weighted budget. The priest's wand adds 2.4% of it, worth 0.003.
  - Race: each class has its best legal race; the shaman's other races are within 0.002.
  - Consumables and buffs: identical.
  - Mana: every tank healer ends at its carry with 39-89% mana.
- **Each preset suits its role.** On another role's preset, every healer but the paladin carries less.

## Against the logs

The Anniversary kills bring about 5.1 priests, 3.0 shamans, 1.7 paladins and 1.2 druids per 11 healers. The boss
files keep no healing by class, so the sim's order can't be checked against the logs' (healer_fairness-20261104-4).

## Open questions

1. The tank bench's raid (finding 1): fillers kept in tank role, or Prayer of Mending kept off the healers. Then
   re-bench the tank columns of `docs/healing_sim.md` H3.
2. The tank presets' heal-size gate (finding 2), re-tuned per class.
3. Ancestral Fortitude against Inspiration: all armor or item armor, the same rule for both (healer_mechanics-20261104-3).
4. The paladin's tank preset carries as well in the raid role as its raid preset (1.319 against 1.302, 3 files, 500
   iterations). Is the raid preset under-tuned?
5. The classes' "Calls for the user" lists (`docs/healing_{priest,paladin,druid,shaman}.md`) are open rulings.
   This pass found none of them factually wrong. The priest's call 13 (the tank lead "may reflect the shaman
   preset") is answered above: mostly the bench, a little the gate.
