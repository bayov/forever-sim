# Fairness pass 41 (2026-11-14, master 1bf5feb)

Result: quiet (no new finding). Watched files changed since 3a86d4a: sim/core/statweight.go (attributes stepped by 10 so
Kings' 10% isn't rounded away, 2745da6ce: stat weights only), assets/database (in-game reveals, 668c7bc40), the
Enhancement re-sim after the Nov 13 log audit (760d4c1e5: Imp GW on the new hit-1 synthetic set, Water Shield moves
Stormstrike after Flame Shock, four metas, presets.ts), report_hashes and test .results (stat-weight step). Full race
sweep (5k, seed 1, every base request at every race, the loop's 1bf5feb CLI and tree). Client 1.60.1.70938.

## What moved against the 3a86d4a baseline

| Build | Before | Now | Cause |
|---|---|---|---|
| Enhancement imp-gw (Orc) | 1038.3 | 1040.9 (+2.6) | hit-1 regear (760d4c1e5); presets.ts note claims +1.2 to +2.6, consistent |
| Enhancement water-shield | Orc 1000.1 | Windshaper 1000.1, Orc 999.7 | Stormstrike reorder; net zero, race order is a 0.4 tie (noise) |

No other sweep line changed (diff of sweep logs touches only these two). statweight.go only changes the stat-weight
step, and the database edits reach no report request: no shared leak. The hit-1 set is a hit-for-other-stats trade under
the ruled phase 5 method, and audit1b still passes.

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Best race: unchanged; report race is best or within 0.5 everywhere (Fire Gnome +0.4, MM Night Elf +0.3,
  Combat Daggers Gnome +0.6 as before, Enhancement water-shield/mana-tide Windshaper ties).
- Published: the four Enhancement metas now match the sweep within 0.4 (pass 40's open item closed). builds.json
  matches the shared 20k run on 68 of 68 (results update 2026-11-14). Other metas' gaps unchanged: Fury 2H -119, Fury
  -86, Arms -46, Affliction +37, Survival -26, Lone Wolf -22, MM -20, rogues -10 to -20, BM -14, feral +15. All already
  under results-20261104-1/-2 and -20261112-1; confirmed independently, not refiled.

## Ranking now (report race, 5k)

Unchanged from pass 40: Lone Wolf 1291.8, MM 1234.7, BM 1182.0, Feral 1153.8, Affliction 1071.1, Frostfire 1065.7,
Destruction 1063.2, Demonology 1055.3, Fire 1053.8, Enhancement 1054.1, Fury DW 1050.2, Arms 1041.7, Fury 2H 1032.0,
Survival 1014.8, Destruction Fire 997.9, Ret 983.8, Elemental 977.4, Arcane 963.6, Balance 952.1, Shadow 948.6,
Combat Swords 943.9, Frost 943.6, SF Swords 938.7, Combat Daggers 936.0, SF Daggers 904.4, Mutilate 865.7,
Subtlety 863.4, Smite 781.0. Top over bottom 1.65x. Full table: `ref/fairness/table_baseline.md`, and
`orchestration/fairness_report.md` §1.

## Baseline

`ref/fairness/sweep_baseline.log` and `table_baseline.md` replaced with this sweep. Baseline sha 1bf5feb.

## Open for the next pass

- Other metas' published.dps: re-check when results-20261104-1/-2 and -20261112-1 are fixed.
