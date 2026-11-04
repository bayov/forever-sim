# Fairness pass 29 (2026-11-04, master 6b480b0)

Result: quiet (no new finding). 141 watched files changed since fb584ea (97 commits), including `sim/core`, the item
database and the warrior / rogue / shaman code, so this was a full pass: `audit1b.py` and a full 5k race sweep on a
CLI and tree built at 6b480b0 (origin/master had moved to 41aa107 meanwhile; not audited). New baseline: 6b480b0.

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Warrior hotfix push 113539 (tree, Spearing Strike, Raging Blows out, Reckless Swings on the GCD) | 7cd399e88, e07adf11b | Fury 2H -20.0, Fury DW -17.8, Arms -9.9; published within 0.3 |
| Rogue hotfix push 113483 (tree, Shank, Mutilate, Venom) | 316a600ac, e9185b1aa | SF Daggers -18.2, SF Swords -16.6, Mutilate -8.5, Combat Swords -4.8, Combat Daggers +2.7, Sub -0.4 |
| Forever downranking moved into sim/core, applied to every spell | 8f304b240, b57f5cc0a | no report build moves (every build casts top ranks; probe abaec7e0b agrees) |
| Shaman Chain Lightning 0.714, metas refreshed | a4fc1d715 | no DPS change (CL not in the rotations); shaman metas now current |
| Power Infusion heals, absorb pipeline, healer classes | 72b236b48 and others | no DPS report build moves |
| Item db for 70755, random suffixes to 60 | 993314e0e, bd00cd090 | no DPS change in any class (synthetic sets unchanged) |
| Labeled random streams in the report sims | ca0ca1285 | report-tool only; shared 20k run matches builds |
| Standards 40 (Enhancement sims both imbue pairings on endgame gear) | 0998dd5b7 | ruling for later gear; Phase 1 synthetic keeps WF Weapon + FT Totem |

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Sweep (5k, seed 1, every race) against the f331f0c baseline: every hunter, mage, warlock, druid, paladin, priest and
  shaman build identical to 0.1. Only warrior and rogue moved, both from their own hotfix commits. Nothing shared leaked.
- Published against current: within 1.4 everywhere but five rogue metas (SF Daggers -17.6, SF Swords -15.9, Mutilate
  -8.2, Combat Swords -5.2, Combat Daggers +3.2). Already open as results-20261104-1 (metas lag; builds.json is right);
  independent confirmation here, not filed again. results-20261102-1's shaman gaps are gone (+0.3 to +2.1 now).
- Best race: report race is best within noise. Combat Swords Undead 0.7 under Orc, Combat Daggers Undead 0.2 under
  Gnome (both under the 1.0 note threshold).

## Ranking

Lone Wolf 1314.4, MM 1254.6, BM 1195.7, Enhancement 1159.9, Fury 2H 1150.7, Feral 1139.7, Fury DW 1135.8, Ret 1091.3,
Arms 1088.0, ... Smite 781.0 last. Top over bottom 1.68x. Full table: `orchestration/fairness_report.md` §1.

## Open items (carried)

- None new. Next full sweep when `sim/` or the database changes again.
