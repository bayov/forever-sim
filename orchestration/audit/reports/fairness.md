# Fairness pass 48 (2026-11-22, master 8b5137c6d)

Result: quiet (no new finding; fairness-20261118-1 still open). Range b4b68c52e..8b5137c6d. Full re-sweep (5k, seed 1,
every race) on the 8b5137c6d CLI against the a6677362f baseline; baseline replaced.

## What changed in the watched paths

| Change | Commit | Effect on the comparison |
|---|---|---|
| Two-Handed Weapon Specialization raises all Holy damage (Nov 19 hotfix, standards 41) | 1c80dafff | Ret +4.8 (+0.5%), Shock Ret +9.3, sub-Holy +5.0 |
| Shank procs poisons, enchants and on-hit effects; Stormshroud 3P 18 energy at 3% | 6de6861c5, c7bab3c7d | all six rogues +3.5 to +4.6 (+0.4 to +0.5%) |
| Flurry no longer spent by Heroic Strike / Cleave under Forever (Nov 16 hotfix) | #133 (sim/warrior/talents.go) | Fury and Fury 2H +2.0 (+0.2%); Arms takes no Flurry, unchanged |
| Barov Peasant Caller trinket; Enhancement Phase 0 build | abefd2f86, 5320d5b5b | no report request wears it (Phase 0 is a UI-only preset; report trinkets stay synthetic, 1b) |
| sim/core/buffs.go | (Windfury Totem comment) | comment only |
| Ret sub-Holy vs Undead rotation; standards 41 | 1b4a74f33, e411cbf1e | Undead variant re-simmed (builds.json 1151.0 vs 1151.5 now) |
| DPS Report re-sim of 22 builds | facfbbf3f | builds.json current |

11 of 43 sweep rows moved, each in a class whose code changed. Every caster, hunter, druid, shaman and Arms row is
identical to a6677362f, so no shared change leaked.

## Checks

- Published vs current: builds.json within 1.3 of the loop's 20k run on 68 of 68 (largest Destruction +1.3, MM wolf
  -1.0); no ranking flip. The meta published.dps lags (Affliction +42.5, Lone Wolf -22.4, rogues -5.9 to -15.1,
  Destruction +6.6, Ret Shock +7.2, Fire -4.9) do not reach builds.json; independent confirmation of
  results-20261104-1 / -2, not refiled.
- Race: Balance's report race table still names High Order best on all four Balance builds (+3.8 to +6.1):
  fairness-20261118-1 open, no change. Sweep: report race best within 0.7 elsewhere (Shadow Troll +0.7, Frostfire
  High Order +0.4, Fire Gnome +0.4); MM wolf / tallstrider Night Elf +1.2 / +1.8 over Orc (base MM a tie), unchanged
  since pass 45.
- Gear: `audit1b.py` on 8b5137c6d, 19 of 19 "5a PASS | 1b PASS".
- Client 1.60.1.71037 (ledger 2026-11-20): no ChrRaces or raid-buff change, so no shared re-sweep need beyond this one.

## Ranking

Order unchanged (top Lone Wolf 1291.8, bottom Smite 781.0, 1.65x). Full table: `ref/fairness/table_baseline.md`,
copied into `orchestration/fairness_report.md` §1.

## Baseline

Replaced with the 8b5137c6d sweep (`ref/fairness/sweep_baseline.log`, `table_baseline.md`).

## Notes for the next pass

- `table.py` needs `AUDIT_TREE` exported in the same shell (it raises KeyError otherwise).

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, FTT + WF imbue stacking, boss partial resist log, Warrior Rage k and
demon/pet scaling ranges (see passes/fairness.md).
