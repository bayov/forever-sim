# Fairness pass 47 (2026-11-18, master a6677362f)

Result: findings (1 new). Range 8219f1083..a6677362f (46 commits). Full re-sweep (5k, seed 1, every race) on the
loop's a6677362f CLI; baseline replaced.

## What changed in the watched paths

| Change | Commit | Effect on the comparison |
|---|---|---|
| Haste shortens a spell's GCD, floor 1s (standards 27) | ee35da473, c2efe4ae3 | casters only: Destruction +5.9, Affliction +5.7, Balance +4.2, Elemental +3.8, Frost +1.4, Demonology +1.3, Shadow +0.7; Ret -2.0 (rotations timed for 1.5s, re-time deferred by user ruling, rotations.md) |
| Report switches folded into code, faster start-up, healbench speedups | c8b5275b1, cc5a55f47, bbfd014e4 | no DPS build moved beyond 5k noise |
| Healers' idle skip wakes for a hit | 7199c8f9c | healing only |
| DPS Report full regeneration | 392503f55 | builds.json current |

Every melee, hunter and rogue row is identical to the 5f20dfdf4 baseline: the GCD change did not leak to physical abilities.

## Checks

- Published vs current: builds.json within 1.3 of the shared 20k run on 68 of 68 (no ranking flip). The meta
  published.dps lags (Affliction +42.5, Lone Wolf -22.4, rogues -10.5 to -20.1, Destruction +6.6, Balance +3.9,
  Elemental +4.0, Fire -4.9) do not reach builds.json; independent confirmation of results-20261104-1 / -2, not refiled.
- Race: **Balance** (fairness-20261118-1). The report's race table now names High Order best on all four Balance builds
  (+3.8 to +6.1). Paired 20k seeds 1 / 1000003: Balance NE 956.1 / 956.2 vs High Order with its race rotation
  959.8 / 959.8; Wrath NE 866.0 / 866.1 vs High Order 871.9 / 871.9. Cause: the hasted GCD raises the Skyborne
  haste racial. Elsewhere the report race is best within 0.7 (Shadow Troll +0.7, Combat Daggers Gnome +0.6, Frostfire
  High Order +0.4, Fire Gnome +0.4); MM wolf / tallstrider Night Elf +1.3 / +1.5 over Orc in the report's 10k table,
  unchanged since pass 45 and a tie on the base MM build.
- Gear: `audit1b.py` on a6677362f, 19 of 19 "5a PASS | 1b PASS".
- Note: `sweep.py` swaps only the race, so for Balance it misses the race-tuned rotation (its table shows Night Elf best);
  the report's race table and the paired check above use it.

## Ranking

Order unchanged (top Lone Wolf 1291.8, bottom Smite 781.0, 1.65x). Full table: `ref/fairness/table_baseline.md`,
copied into `orchestration/fairness_report.md` §1.

## Baseline

Replaced with the a6677362f sweep (`ref/fairness/sweep_baseline.log`, `table_baseline.md`).

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, FTT + WF imbue stacking, boss partial resist log, Warrior Rage k and
demon/pet scaling ranges (see passes/fairness.md).
