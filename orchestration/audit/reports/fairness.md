# Fairness pass 45 (2026-11-17, master 5f20dfdf4)

Result: quiet (no new finding; the published lag is results-20261104-2, confirmed below). Range 1a85df1d7..5f20dfdf4.

## What changed in the watched paths

| Change | Commit | Effect on the comparison |
|---|---|---|
| Hunter pet white swings take none of the hunter's crit (wowsims/forever 7e0e0c2e46) | b075ae64d | BM -5.6%, MM -2.6%, Survival -2.2%; Lone Wolf (no pet) unchanged |
| Raptor Strike: queued keeps the DW miss penalty; Predator's Edge counts twice | eada6763e, 4a8541e46 | Survival / melee-weaving hunters (folded into the moves above) |
| Immolation and Explosive Trap roll spell hit | 29481dd0e | Survival (folded in) |
| Warrior hotfix 114097 (Unbridled Wrath, Flurry, Reckless Swings on abilities); avoided Execute keeps Rage | ad1d7f469, 2ebdc14d6, 3e1214a69 | Arms +9.5, Fury +9.8, Fury 2H +7.8 (+0.7 to +0.9%) |
| Rogue Death Mark always hits, cap uses sheet crit | cf95c3e32 | Subtlety +5.4 (+0.6%); other rogues unchanged |
| Five Thunders set (1679), Manual Crowd Pummeler 3 min CD, core ItemSetOf / Pet.InheritedStats | 6875e3b1e, 6ea6e47b4, 39b94e364 | no report request uses them; no build moved |
| report_hashes.txt, goldens | d5e59ef48 | test data only |

## Checks

- Shared 20k run of 5f20dfdf4 vs `ref/fairness/sweep_baseline.log` (25ce81d) at each report race: only the 13 hunter /
  warrior / Subtlety base builds moved beyond 2.1; every other class within noise. No leak from sim/core.
- Race re-sweep (5k, seed 1, every race) of the 14 builds in changed classes: report race is still best within noise
  (MM Night Elf 1202.3 vs Orc 1201.7; Subtlety Night Elf 872.0 vs Undead 871.5; Fury 2H Orc 1111.4 vs NE 1106.3).
- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Symmetry: the pet-crit rule is hunter-only by its source; warlock demons keep their (separate) inheritance, already
  covered by the open "demon / pet scaling" item. Not flagged.
- Published vs current: builds.json, current on 68 of 68 at the 2026-11-15 re-sim, now lags 20 of 68 builds by more
  than 1.5 (BM -67 to -76, MM -30 to -33, Survival -19 to -22, warriors +8 to +10, Subtlety +5). Ranking flips against
  the published one: Survival under Ret, Arms over Affliction. The results auditor recorded this lag in the same run
  (update 2026-11-17 under results-20261104-2); independently confirmed here, with the ranking effect, not refiled.
  Older meta-only lags unchanged: Affliction +37, four rogues -10.6 to -19.6, Fire -4.1.

## Ranking (5k, best race; top 10 and movers)

| # | Spec | DPS | Before (25ce81d) |
|---|---|---|---|
| 1 | Lone Wolf | 1291.8 | 1291.8 |
| 2 | Marksmanship | 1202.3 | 1235.0 |
| 3 | Feral | 1186.0 | 1186.0 |
| 4 | Beast Mastery | 1115.5 | 1182.0 |
| 5 | Fury 2H | 1111.4 | 1103.6 |
| 6 | Fury | 1102.6 | 1092.8 |
| 7 | Enhancement | 1081.1 | 1081.1 |
| 8 | Arms | 1079.0 | 1069.5 |
| 9 | Affliction | 1071.1 | 1071.1 |
| 14 | Retribution | 1024.2 | 1024.2 |
| 15 | Survival | 1012.9 | 1035.2 |
| 26 | Subtlety | 872.0 | 866.6 |

Top over bottom 1.65x. Full table: `ref/fairness/table_baseline.md`.

## Baseline

Replaced: the 14 re-swept lines in `sweep_baseline.log` are from 5f20dfdf4; the rest stay from 25ce81d (their classes'
code did not change and the 20k run confirms them). `table_baseline.md` regenerated. `orchestration/fairness_report.md`
§1 updated, old table in `<details>`.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, FTT + WF imbue stacking, boss partial resist log, Warrior Rage k and
demon/pet scaling ranges (see passes/fairness.md).
