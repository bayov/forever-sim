# Fairness pass 38 (2026-11-13, master 71562a8)

Result: quiet (no new finding; fairness-20261112-1 resolved). Watched files changed since 8306f1d (52), most of them
tank and healer presets and rotations (tankheal merge a6aabc896). Partial pass: the machine was shared with healer passes,
so instead of the full 20 min race sweep I compared the loop's 20k master run against the 8306f1d sweep and re-swept by
race only the builds that moved.

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Judgement is a ranged attack (Nov 11 hotfix) | f1c07e13a | Ret -4.3, Shock Ret -7.5, Sub-Holy Ret -5.6 (5k, report race). Ret stays 15th; Undead still best (Human -4.2, Dwarf -9.6) |
| Windfury re-tunes: Survival Wing Clip without mana floor, Fury Battle Shout duty HS 60 in Reckless Swings / fillerGap 1 | 8f5de1c87 | Survival +1.2 (noise at 180 s), Fury BS duty +5. Base builds held, per the commit's two-seed check |
| Combat Swords back as rogue default | 8f5de1c87 | Resolves fairness-20261112-1 |
| Windfury rule by client flag (standards 17) merged into tankheal; master's rule kept | 2a0208136, c1502631c | Already in the 8306f1d baseline: no DPS build moved apart from Ret |
| Healer health values, tank/healer presets, enchants, consumables | 65b5d49cf and others | No DPS build moved |

## What I checked

- Movement: loop's 20k run (`$AUDIT_WORK/runs/master/summary.tsv`) against `ref/fairness/sweep_baseline.log` (5k): only
  the three Ret builds moved more than 1.5. All attributed to f1c07e13a (paladin only). No shared leak.
- Report data: `ui/dps_report/data/builds.json` (e9a8857e7) matches the 20k run within 1.5 on all 68 builds.
- meta.published.dps still lags on 45 metas (Fury 2H -119, Sub-Holy -97, Ret -88, Fury -85, Arms -45, Affliction +37,
  Survival -26, hunters -9 to -22, rogues -1 to -19, Feral +14). Same bookkeeping as before; results-20261104-2 covers it.
- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Best race: paladin rows re-swept (Undead best everywhere). Others unchanged since pass 37.

## Ranking now (report race, 5k; Ret rows from this pass, the rest from pass 37)

Lone Wolf 1291.8, MM 1234.7, BM 1182.0, Feral 1153.8, Affliction 1071.1, Frostfire 1065.7, Destruction 1063.2,
Demonology 1055.3, Fire 1053.8, Fury DW 1050.2, Arms 1041.7, Fury 2H 1032.0, Enhancement 1030.7, Survival 1013.9,
Ret 1004.6, Destruction Fire 997.9, Elemental 977.4, Arcane 963.6, Balance 952.1, Shadow 948.6, Combat Swords 946.1,
Frost 943.6, SF Swords 938.7, Combat Daggers 936.0, SF Daggers 904.4, Mutilate 865.7, Subtlety 863.6, Smite 781.0.
Top over bottom 1.65x.

## Baseline

`ref/fairness/sweep_baseline.log` and `table_baseline.md`: the 8306f1d full sweep with the three paladin lines replaced
by this pass's 71562a8 re-sweep. Baseline sha 71562a8. Next full sweep: when sim/core or a shared rule moves many builds.

## Open for the next pass

- Metas' published.dps: re-check when results-20261104-2 is fixed.
