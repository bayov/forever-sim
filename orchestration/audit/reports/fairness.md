# Fairness pass 37 (2026-11-12, master 8306f1d)

Result: quiet (no new finding, none resolved). Watched files changed since a0b4d3b (6): hunter pet melee speed and
autocast (368b2e181, with sim/core/pet.go, unit.go), Rake's hit takes 6.5% AP (c380bbaef), report hashes. Full pass:
`audit1b.py` and a full 5k race sweep on the loop's 8306f1d CLI and tree. New baseline: 8306f1d.

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Hunter pet swings at its owner's melee/attack speed (Pet.FollowsOwnerMeleeSpeed, hunter only); autocast passes an ability short of focus | 368b2e181 | Survival +26.1 (16th -> 14th, passes Ret and Destruction Fire), BM +7.3, MM +2.0 (variants +4 to +27). Lone Wolf (no pet) identical |
| Rake's direct hit adds 0.065 AP (user ruling 2026-11-12, 70862 log) | c380bbaef | Feral +29.4 (+2.6%), stays 4th |
| Core pet.go / unit.go | 368b2e181 | The old DK hook became the hunter-only flag. Warlock demons unchanged (all warlock rows identical), standards 14 intact |

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Movement against the a0b4d3b baseline: only hunters (pet) and Feral moved, both attributed. No leak.
- Report data: ui/dps_report/data/builds.json matches the loop's 20k master run within 1.5 DPS on all 68 builds
  (re-sim bff0169e4 and 2b62e304a), so the dashboard ranking is current.
- meta.published.dps still lags on 45 metas (Fury 2H -119, Fury -85, Ret -82, Sub-Holy -92, Arms -45, Affliction +37,
  Survival -27, hunters -9 to -22, rogues -1 to -19, Feral +14.5). Bookkeeping in the req metas, not the published report;
  the results auditor's open results-20261104-2 covers it. Not refiled.
- Best race: report race is best within noise (MM Night Elf +0.3, Combat Daggers Gnome +0.6, Fire Gnome +0.4). The MM pet
  variants run Orc, Night Elf +1.0 (wolf) and +1.5 (tallstrider) at 5k; builds.json at 20k agrees (+1.0, +1.2). Note for the
  hunter owner, below finding level (variants keep the base build's race).
- fairness-20261112-1: confirmed at 20k after the re-sim, Combat Swords 946.4 over default Combat Daggers 936.2.

## Ranking now (report race, 5k)

Lone Wolf 1291.8, MM 1234.7, BM 1182.0, Feral 1153.8, Affliction 1071.1, Frostfire 1065.7, Destruction 1063.2,
Demonology 1055.3, Fire 1053.8, Fury DW 1050.2, Arms 1041.7, Fury 2H 1032.0, Enhancement 1030.7, Survival 1013.9,
Ret 1008.9, Destruction Fire 997.9, Elemental 977.4, Arcane 963.6, Balance 952.1, Shadow 948.6, Combat Swords 946.1,
Frost 943.6, SF Swords 938.7, Combat Daggers 936.0, SF Daggers 904.4, Mutilate 865.7, Subtlety 863.6, Smite 781.0.
Top over bottom 1.65x. Full table: `orchestration/audit/ref/fairness/table_baseline.md`.

## Open for the next pass

- fairness-20261112-1: check that the Combat default follows Swords once the orchestrator applies it.
- Metas' published.dps: re-check when results-20261104-2 is fixed.
