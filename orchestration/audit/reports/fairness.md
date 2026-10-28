# Fairness pass 11 (2026-10-28, master 5d27b00)

Result: quiet (no new finding; fairness-20261027-1 resolved).

## Why it ran

Since 37a8a8c: post 8 racials (ae8522c14: Blood Fury 2 min, Berserking 15 sec, Touch of the Grave 15 sec ICD),
standards 29 race-tuned rotations and gear (f6116e453, 1526b8b48, req/races/*.json, 32 files), and the presets'
race re-picks for hunters, Balance, Arcane, rogues, paladins, Enhancement, Arms, Shadow and warlocks.

## What I checked

- Gear: `audit1b.py`, all 19 sets "5a PASS | 1b PASS".
- Shared code: the racial changes are race-gated (Orc / Troll / Undead only); builds on other races moved 0.0.
- Race sweep (5000 it, seed 1) with each race's req/races override applied. Tooling: `sweep.py` now applies the
  tree's `race_override.py` (before it simmed every race with the preset's rotation, unfair to tuned races), and
  `table.py` now really reads builds.json (pass 10 said it did; it still read the metas).
- 20k paired seeds (1 / 1000003) on the two non-trivial best-race gaps: Fire Gnome 1089.6 vs Human 1088.8 (+0.7),
  Affliction Orc 1019.0 vs Human 1018.3 (+0.6). Both ties; no note needed.
- Published against current: 5k sweep within 2.4 of builds.json (Frostfire +2.4, Affliction +1.6, noise); the
  metas' published.dps still lag (results-20261027-1, open, not re-filed).

## Movement (vs baseline 06a8987, best race)

| Build | Before | Now | Cause |
|---|---|---|---|
| Fury DW / 2H (Orc) | 977.1 / 956.0 | 983.1 / 963.3 | Blood Fury 2 min |
| BM / MM / Survival (Orc) | 1146.3 / 1132.7 / 1024.8 | 1150.6 / 1137.7 / 1029.7 | Blood Fury 2 min |
| Rogues (Undead) | | +0.3% to +0.9% | Touch of the Grave 15 sec ICD |
| Arms, Ret, Shadow, Smite, Destr. Fire (Undead) | | +0.3% to +0.5% | same |
| Demonology, Destruction (now Orc) | 1049.5 / 1000.1 | 1052.0 / 1002.1 | race switch + Blood Fury |
| Ret Shock (Undead) | 866.5 | 887.7 | Light's Vigil rotation 35928cc48 |
| everything else | | 0.0 to +0.1% | none |

No unexplained move. Every report race is the best race. Ranking order unchanged at the top; Survival (1029.7)
now sits above Affliction (1019.8). Top over bottom 1.60x. Full table: `ref/fairness/table_baseline.md`,
fairness_report.md §1.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, FTT + Windfury, boss partial resist, Rage k and pet scaling.
