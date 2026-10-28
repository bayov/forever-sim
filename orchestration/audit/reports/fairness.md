# Fairness pass 14 (2026-10-28, master dd663e6)

Result: quiet (no new finding, no open finding). Full pass: gear check, full race sweep, new baseline.

## Why it ran

Since c9d93dd, 78 commits. Watched paths: standards.md got rule 31 (rotation lines state their intent, checked at
120 / 180 / 300 s and 180 +- 20 s) and the rule 30 source wording. Rotation-intent commits: warrior (dd4ff5d91),
mage (8624ba126), warlock Troll Berserking (d387da420), Elemental lower-rank Lightning Bolt (3ad93a86b), Holy Shock
Light's Vigil (d8185e98f); Smite without Chastise (e419e7f50); the Dwarf gear-trade split of Artisan's Precision in
the race breakdown (5f96ef4c0); priest Shadowguard split (report data). Most of the 103 req files changed only meta
text (phrasing passes). No sim/ or proto/ change; no metas' `published.dps` changed.

## What I checked

- Gear: `audit1b.py` passes 5a and 1b on all 19 lines.
- Report races: every preset race is +0.0 against the 1274e15 baseline (5k, seed 1). Ranking unchanged.
- Published against current: shared 20k run (`vs_published.txt`) has 57 of 58 within 0.5. Frostfire +0.9
  (1087.5 -> 1088.4) matches 8624ba126's own measurement (+0.25 at 180 s, +0.8 at 180 +- 20 s); within tolerance.
  The 5k gaps (Frostfire +2.4, Affliction +1.6) are unchanged since pass 11.
- Non-report races that moved, each attributed: Fire Orc +2.3 / Troll +0.4 and Frostfire Orc +1.4, Undead -1.6,
  Gnome -1.1, High Order +0.6 (8624ba126; Frostfire Orc and Undead race files removed as no longer gaining);
  Elemental every race +0.1 to +0.5 (3ad93a86b); warlock Troll -2.0 / +1.4 / -0.2 / +0.9 (d387da420); Fury 2H
  Night Elf +7.1, Fury / Arms Night Elf +0.1 / +0.2 (dd4ff5d91). None overtakes its preset race.
- Smite (no Chastise) and the Dwarf gear-trade change move no race total.
- Best race against report race: Fire Gnome +1.1, Affliction Orc +0.1, both 20k-measured ties (passes 11, 12).
- Rule 31 is a robustness rule; at 180 s it moved no preset. Whether the classes it did not touch (rogue, hunter,
  druid, Enhancement, Shadow) hold their ranking at 120 / 300 s is not checked here (rotations auditor's scope).

## Tool fix (again)

Pass 13's fix to `tools/audit/fairness/sweep.py` (apply each race's req/races/ override through the tree's
`race_override.apply`) never reached the branch. The first sweep this pass ran without it and showed Dwarf -3.5 to
-8.2, Gnome rogues, Smite races and High Order mages 2 to 15 down. Re-applied (uncommitted helper change in the
working tree, for the run's commit); the rerun is the baseline.

## Ranking (top 8; full table in fairness_report.md §1 and ref/fairness/table_baseline.md)

| 1 | Beast Mastery | Orc | 1131.8 | Orc / Night Elf 1119.5 | Orc (1131.8) | 1131.8 | +0.0 | 1131.8 (+0.0%) |
| 2 | Marksmanship | Orc | 1116.1 | Orc / Night Elf 1109.6 | Orc (1116.1) | 1115.7 | +0.4 | 1116.1 (+0.0%) |
| 3 | Lone Wolf (hunter) | Night Elf | 1093.8 | Windshaper 1093.5 / Night Elf | Night Elf (1093.8) | 1094.3 | -0.5 | 1093.8 (+0.0%) |
| 4 | Fire | Gnome | 1090.6 | Undead 1086.9 / Gnome | Human (1089.5) | 1089.0 | +0.5 | 1090.6 (+0.0%) |
| 5 | Frostfire | Human | 1089.9 | Orc 1080.5 / Human | Human (1089.9) | 1087.5 | +2.4 | 1089.9 (+0.0%) |
| 6 | Demonology | Orc | 1052.0 | Orc / Gnome 1050.7 | Orc (1052.0) | 1051.8 | +0.2 | 1052.0 (+0.0%) |
| 7 | Survival | Orc | 1028.8 | Orc / Night Elf 1019.1 | Orc (1028.8) | 1028.9 | -0.1 | 1028.8 (+0.0%) |
| 8 | Affliction | Orc | 1019.8 | Orc / Human 1019.7 | Human (1019.7) | 1018.1 | +1.6 | 1019.8 (+0.0%) |
| 9 | Destruction | Orc | 1002.1 | Orc / Human 1000.1 | Orc (1002.1) | 1001.6 | +0.5 | 1002.1 (+0.0%) |
| 10 | Destruction Fire | Undead | 987.4 | Undead / Human 982.0 | Undead (987.4) | 986.4 | +1.0 | 987.4 (+0.0%) |

Top over bottom: 1.58x (unchanged). The baseline is now dd663e6.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, Flametongue Totem with the Windfury imbue, boss partial resists,
Rage k and pet scaling (see passes/fairness.md).
