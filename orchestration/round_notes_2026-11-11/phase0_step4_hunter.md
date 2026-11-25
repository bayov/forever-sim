# Phase 0 step 4: hunters

Step 4 of the Phase 0 brief (`briefs/phase0_step4_1124.md`) for hunter_bm, hunter_mm (with Lone Wolf) and hunter_sv.
Every DPS below is wowsimcli on the report request, 180 ± 20s, 20,000 iterations, seed 1. A list's own `dps` field is
its ranking score, about 13 higher than this 180s measure.

## Per build

| Build | Phase 0 | Synthetic (old published) | Change |
|---|---|---|---|
| Beast Mastery | 1009.46 | 1115.4 | -106.0 |
| Beast Mastery, wolf | 964.39 | 1065.6 | -101.2 |
| Beast Mastery, tallstrider | 955.14 | 1057.1 | -102.0 |
| Marksmanship | 1090.51 | 1201.3 | -110.8 |
| Marksmanship, wolf | 1070.33 | 1179.4 | -109.1 |
| Marksmanship, tallstrider | 1064.60 | 1173.8 | -109.2 |
| Lone Wolf | 1173.97 | 1313.7 | -139.7 |
| Survival | 926.70 | 1012.1 | -85.4 |
| Survival, wolf | 904.82 | 989.1 | -84.3 |
| Survival, tallstrider | 897.66 | 982.5 | -84.8 |

Each request sims to its list's 180s measure within noise (Beast Mastery 1009.0, Marksmanship 1090.4, Lone Wolf 1174.1,
Survival 926.9).

## Talents and rotations at the Phase 0 hit

The ranged sets sit at 7.6% hit after the re-rank, Survival at 9% with Surefooted 3/3. Talent searches on the Phase 0
settings (2,000 iterations, 120, 180 and 300s together):

- Marksmanship 502-0053551501503051-5, Lone Wolf 5-0053552511503051-5 and Survival 502-0050051-500230230250022151:
  unchanged. Survival keeps Surefooted 3/3 at 9% hit.
- Beast Mastery changed from 5021201505021241-00510504005 to 5021201505011251-00520503005: Spirit Bond 2 to 1,
  Frenzy 4 to 5, Improved Stings 1 to 2, Improved Arcane Shot 4 to 3. At 20,000 iterations it gains 2.7 to 3.9 at 120
  and 180s and 7.8 to 8.1 at 300s on all three pets. The cause is the thinner Phase 0 mana, not hit.
- Rotations: no change.

## List changes

- Black Dragonscale Helm and Windreaver Greaves went over the gold line. Re-ranked head and feet (`RANK_ONLY=1
  SLOTS=head,feet`) on the Marksmanship and Survival lists, since a full RANK_ONLY on all three ran past the 2h limit:
  - Marksmanship: Beastmaster's Cap and Beastmaster's Treads.
  - Survival: Outlaw's Collar and Beastmaster's Treads.
- Beast Mastery had a full RANK_ONLY on its new talents. It picks Beastmaster's Cap, Beastmaster's Treads, Beastmaster's
  Bindings (wrist) and Ankle Slicers (ranged, was Assassin's Throwing Axe). Score 1021.67.
- Lone Wolf has its own list now (`hunter_lw` in phase0_specs.json, `level60_phase0_hunter_lw`), ranked from the
  Marksmanship list on Lone Wolf talents at Night Elf with shoulder, head and feet open: Theramore Epaulets, Beastmaster's
  Cap, Beastmaster's Treads. Score 1190.05. It used to share the Marksmanship set.

## Requests and race rows

- All ten requests were regenerated with mksettings on the Phase 0 sets, with no `_hit±N` steps. The metas carry the
  gear set, the verified DPS and Beast Mastery's new talents.
- The caveat about a 2.5s bow with Rhok'delar's weapon DPS described the synthetic gear. It now names the Phase 0 ranged
  pick.
- Faction sets for other-faction rows: shoulders re-ranked at Human or Orc (lists in `phase0_step4_hunter/`).
  - `level60_phase0_hunter_bm_alliance` and `_mm_alliance` and `_sv_alliance`: Theramore Epaulets for the Alliance races.
  - `level60_phase0_hunter_lw_horde`: Darkspear Epaulets for the Horde races.
- race_gear.py --write on all four specs. Lone Wolf's Dwarf trades its 0.6% excess hit for +5.0 DPS.

## Problem

race_gear.py puts every other-faction race on its faction set and skips the Dwarf excess-hit trade for them. So the
Beast Mastery, Marksmanship and Survival Dwarf rows carry 0.6% (ranged) or 1% (Survival) hit over the cap without
re-spending it, likely a few DPS low (Lone Wolf's trade was worth 5.0). The shared script was left as is.

## Full RANK_ONLY check (publish brief, 2026-11-25)

A full `RANK_ONLY` on Marksmanship (2 cores, the cap while other sims ran) took about 1h to rank every enchant and
slot once. The only confirmed change was wrist: Beastmaster's Bindings +0.6 ± 0.2 over Bracers of the Eclipse at
20,000 iterations (the current list has it 0.29 below). Every other slot kept its pick. The script then started its
second pass on the new set and was stopped at the 2h limit before writing, so nothing was applied: a 0.3 to 0.6
swing on the wrist is a tie, and the lists, sets and requests stay. Survival was not run (another 1 to 2h); its
wrist list has Beastmaster's Bindings outside the top 3. Both stay on the head and feet re-rank.
