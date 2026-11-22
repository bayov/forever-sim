# Fairness pass 49 (2026-11-22, master 7a6d1beb9)

Result: findings (1 new: fairness-20261122-1, high). Range 8b5137c6d..7a6d1beb9 (47 commits, 20 watched files).
Economy pass (machine loaded): no full race re-sweep. Movement taken from the loop's shared 20k run (report race) against
the 8b5137c6d sweep baseline; only Frost and Frostfire re-swept at every race (5k, seed 1), since Ice Lance's cut is the
only change here that could move a race pick. Baseline left at 8b5137c6d.

## What changed in the watched paths

| Change | Commit | Effect at the report race (20k shared run vs 5k baseline) |
|---|---|---|
| Ice Lance 0.11 of spell power (user ruling 2026-11-22, retires 0.429) | 4192d6e49 | Frost -110.4 (-11.7%), Frostfire -59.7 (-5.6%) |
| Windfury Totem: Execute rolls twice, no Flurry charge / Bloodthrill off the totem swing (warrior) | b7edf19bb | Fury 2H +18.2 (+1.6%), Fury +11.1 (+1.0%), Arms +1.3 |
| Wrack half-point ticks (sim/core/dot.go SetTickPeriod) | 3960dd1f6 | Affliction +12.4 (+1.2%); other warlocks within 0.5 |
| No Maelstrom stack from the totem; JoW pays only on landed swings | 840029bd8, bd7029410 | Enhancement -7.7 to -12.2 (-0.8 to -1.2%) |
| Hunter pet keeps the AP of its summon | 051a21cd3 | BM / Survival -3.5 to -5.4, MM -1.3 to -2.5, Lone Wolf -0.5 (noise) |
| Linken's Boomerang, Cloak of Fire, Dungeon Set 1 reveal sets, Enh Phase 0 gear | 303023018, 7876cfdd3 | no report request wears them (trinkets synthetic, 1b) |
| Flame Shock, Deep Wounds, Execute fixes | (before 30314ad60) | re-simmed in 30314ad60 |

Every other row (druid, paladin, priest, rogue, Arcane, Fire, Elemental, Demonology, Destruction) is within 0.9 of the
baseline, so no shared change leaked: the sim/core edits (dot.go, attack.go, buffs.go, debuffs.go) show only where the
class commits say.

## Checks

- Published vs current: builds.json (last re-sim 30314ad60) is within 1.5 of the shared run on 40 of 68 builds; the
  28 off are the five classes above. Six specs change place: **fairness-20261122-1** (new, high). The results
  auditor's 2026-11-22 update (results-20261104-2) covers only the mage and hunter rows; independent confirmation of
  those, plus the warrior, Affliction and Enhancement rows it missed.
- Race: Frost re-swept, High Order 834.0 best (Undead 829.2); Frostfire Human 1007.1, High Order 1005.5 (was High
  Order +0.4, now Human +1.6, the report race). Balance unchanged: fairness-20261118-1 open.
- Gear: `audit1b.py` on 7a6d1beb9, 19 of 19 "5a PASS | 1b PASS".
- Shared assumptions: the Windfury Totem rules land for warrior (Flurry, Bloodthrill, Execute) and shaman (Maelstrom);
  the shaman Flurry half stays open as mechanics-20261118-1, not refiled. The hunter pet snapshot leaves out the
  prepull Aspect of the Hawk; the pet takes 10% of 120 RAP, about 1 DPS, below a finding.

## Ranking (shared 20k run, report race; builds.json in brackets)

Top Lone Wolf 1291.3, bottom Smite 781.0, 1.65x. Moves: Fury 2H 5 -> 4, Fury 6 -> 5, Affliction 9 -> 7, Enhancement
7 -> 9, Frostfire 11 -> 15, Frost 22 -> 27. Full table in `orchestration/fairness_report.md` §1.

## Notes for the next pass

- Do a full re-sweep (`sweep.py`, ~20 min) once the five classes are re-simmed and replace the baseline then.
- Helpers used this pass (scratch, worth keeping in mind): shared-run vs baseline at the report race reads
  `tools/dps_report/req/<id>.json` race and `sweep_baseline.log`; sweep.py works with
  `AUDIT_WORK=<dir>` holding a `wowsimcli` symlink to the shared CLI.

## Open items

- Ice Lance: the Low-range item is retired by the 2026-11-22 ruling (0.11). Frost now sits 27th; whether its
  talents and rotation still suit a 0.11 Ice Lance is the rotations auditor's call (results notes the no-Ice-Lance
  APL is stale at 681.8).
- Unchanged: hunter end-of-fight test, FTT + WF imbue stacking, boss partial resist log, Warrior Rage k and demon/pet
  scaling ranges (see passes/fairness.md).
