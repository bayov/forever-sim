# Fairness pass 12 (2026-10-28, master 224b8a4)

Result: quiet (no new finding, no open finding).

## Why it ran

Since 5d27b00: 9 race-tuned rotation files (standards 29) in tools/dps_report/req/races/: High Order Arcane and
Frostfire, Troll Elemental (both builds), Gnome warlocks (4 specs), Night Elf Fury (commits 4c9f2d32c..1bffad948).
No sim code, preset, gear, consumable or standards change. The commit says no preset changes race.

## What I checked

- Each new file is a rotation-only overlay (rotation, gain, note, how), with no encounter, gear, buff or
  consumable fields, so only that race's sim in that build's race table changes (race_override.py).
- Stated gains are all over 2 SE (standards 29 keep rule): Frostfire HO +18.8 (se 0.82), Fury NE +9.3 (0.38),
  Arcane HO +4.9 (0.11), Elemental Troll +4.9 (0.23), Demonology Gnome +4.1 (0.21).
- Best race against report race, from ui/dps_report/data/races.json (after the re-sim): every preset race is the
  best race or within 1.0. No tuned race overtakes its preset: Demonology Gnome 1050.0 vs Orc 1051.6, Fury NE
  977.4 vs Orc 982.6, Elemental Troll 767.1 vs Windshaper 771.6, Arcane HO 896.0 vs Gnome 901.5.
- Frostfire Gnome 1089.2 vs Human 1088.2 (exactly 1.0 in the 5k table): paired 20k, seeds 1 / 1000003, Gnome
  1088.41 / 1088.18 vs Human 1087.50 / 1087.98, so +0.56 on average. That's a tie, and no note is needed.
  Fire Gnome +0.5 and Affliction Orc +0.4 are ties too (pass 11 measured them).
- Published against current: vs_published.txt gaps (Frost -35.5, Affliction -17.2) are the stale metas'
  published.dps, results-20261027-1 (open, not re-filed; my independent confirmation: builds.json is current).
- Gear (audit1b) and the full sweep weren't rerun: no gear, database or shared-code change. The baseline stays 5d27b00.

## Ranking

The preset numbers didn't change (the race files don't touch preset races). The table in fairness_report.md §1
and ref/fairness/table_baseline.md is still current.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, FTT + Windfury stacking, boss partial resists, Rage k and
pet scaling (see passes/fairness.md).
