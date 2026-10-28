# Fairness pass 13 (2026-10-28, master 1274e15)

Result: quiet (no new finding, no open finding). Full pass: gear check, full race sweep, new baseline.

## Why it ran

Since 224b8a4: 73cedc9a9 takes the launch realms' 2026-10-28 server hotfix (user ruling "Adopt both"). Under Forever,
Trueshot Aura (sim/core/buffs.go) is now 15 / 22 / 30 / 40 / 50 ranged attack power, so rank 5 is 50 instead of the
2026-10-15 ruling's 100. Aspect of the Hawk (sim/hunter/aspects.go) ranks 1-5 are lowered, a level 60 hunter casts
rank 6 (55, spell 14322), and the BM / MM / LW requests name 14322. 5b019f33c then re-simmed the whole report and
refreshed every meta's published DPS. standards.md got rule 30, about the Changes tab only. The other commits only
touch the report UI and deploy.

## What I checked

- Gear: `audit1b.py` passes 5a and 1b on all 19 lines.
- Shared assumption: Trueshot Aura gives ranged attack power only. Other classes' goldens move only their stat
  sheet's ranged attack power (-50). In the sweep, every non-hunter report race is +0.0 against the 5d27b00 baseline.
- Hunter moves (5000 iterations, seed 1, all 8 races): LW -24.1 to -24.7 (-2.2%), MM -21.0 to -22.0 (-1.9%),
  BM -17.8 to -18.8 (-1.6%), Survival -0.9 to -1.0 (-0.1%). The moves are uniform across races and match the
  commit's golden moves (-1.5 to -1.9%; LW a little more). They are all attributed to 73cedc9a9.
- The other moved rows are the 9 race-tuned rotations from pass 12 (Frostfire High Order +14.7, Fury Night Elf
  +8.4, Arcane High Order +4.7, Elemental Troll +4.4 / +4.6, Gnome warlocks +1.2 to +4.8). None overtakes its
  preset race.
- Published against current: builds.json matches the sweep within 0.5, except Frostfire +2.4 and Affliction +1.6.
  Both gaps are unchanged since pass 11 and are 5k noise. The shared 20k run (`vs_published.txt`) has 0 of 58
  requests more than 0.5 off, so results-20261027-1's stale metas are fixed by 5b019f33c.
- Best race against report race: Fire Gnome +1.1 and Affliction Orc +0.1. Both were measured as ties at 20k
  (passes 11 and 12).

## Tool fix

The committed `tools/audit/fairness/sweep.py` set only the race. It didn't apply each race's own req/races/
override (standards 29), which the 5d27b00 baseline did. On a first run, every overridden race looked 1.6 to 13 DPS
down (Dwarf in 10 builds, Gnome rogues, most Smite races, High Order Fire / Balance, Frostfire Orc / Troll / Undead).
The sweep now calls the tree's `race_override.apply(req, id, race)`, and the rerun matches the baseline exactly
outside the hunters and the new race files. This is an uncommitted helper change in the working tree; the run's
commit picks it up.

## Ranking (top 8; the full table is in fairness_report.md §1 and ref/fairness/table_baseline.md)

| # | Spec | Race | DPS | Before |
|---|---|---|---|---|
| 1 | Beast Mastery | Orc | 1131.8 | 1150.6 |
| 2 | Marksmanship | Orc | 1116.1 | 1137.7 |
| 3 | Lone Wolf | Night Elf | 1093.8 | 1118.4 |
| 4 | Fire | Human | 1089.5 | 1089.5 |
| 5 | Frostfire | Human | 1089.9 | 1089.9 |
| 6 | Demonology | Orc | 1052.0 | 1052.0 |
| 7 | Survival | Orc | 1028.8 | 1029.7 |
| 8 | Affliction | Human | 1019.7 | 1019.7 |

Top over bottom: 1.58x (was 1.60x). The baseline is now 1274e15.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, Flametongue Totem with the Windfury imbue, boss partial resists,
Rage k and pet scaling (see passes/fairness.md).
