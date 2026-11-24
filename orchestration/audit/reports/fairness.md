# Fairness pass 53 (2026-11-24, master 566067a39)

Result: quiet (no new finding). Range c78d983ac..566067a39, 9 watched files (sim/druid, sim/hunter, sim/warrior, report_hashes).
Baseline left at 8b5137c6d (no full sweep).

## What changed in the watched paths

| Change | Commit | Effect at the report race (paired, 20k, seed 1, c78d983ac CLI vs 566067a39 CLI) |
|---|---|---|
| Rip 0.0114 AP / CP and Rake 0.065 AP a tick, from Forever logs | be773f5eb | Feral 1186.2 -> 1171.2 (-15.0), FF duty -14.9 |
| Heroic Strike / Cleave spend a Flurry and a Reckless Swings charge a cast; UW not from Shield Slam / Revenge / Bash | cd71e78c2 | Fury -6.2, Fury duties -4.5 / -5.1, Fury 2H -1.0 / -1.2 / -1.5, Arms 0.0, Arms dragonkin +0.2 |
| Hunter pet summoned under a prepull Aspect of the Hawk keeps 10% of its RAP | b82ad9a1e | BM +1.2 to +1.4, MM +0.7 to +0.8, Survival / Lone Wolf 0.0 |

The other 53 builds: 0.0 change. No shared leak (no sim/core change).

## Checks

- Published vs current (builds.json): 7 of 68 lag by more than 1.5: Feral -14.8 / -14.6, Fury -5.5 / -4.6 / -4.6,
  BM wolf +1.6, Fury 2H battle-shout -1.5. Places 5 and 6 swap in the sim (BM 1112.8 over Fury DW 1109.5) but not
  in the published report. Already filed by the results auditor (results-20261104-2, update 2026-11-24, same numbers);
  independent confirmation here, not refiled. Fix is the Feral and Fury re-sim.
- Symmetry: the Rip / Rake and warrior Flurry changes are class rules from Forever logs (user ruling 2026-11-22, a
  measured log wins). The Hawk summon default touches hunter pets only; warlock demons take no player buffs
  (standards 14), so no cross-class asymmetry. Note for the mechanics auditor, not a fairness point: cd71e78c2's
  Heroic Strike / Cleave Flurry charge reverses 30a645072 (Blizzard Nov 16 hotfix, "no longer spend a charge");
  the commit rests on wowsims/forever EV-0138 "after push 114115". Its share of the Fury move was not split out.
- Race (sweep.py, 5k, seed 1, the six touched base builds): every report race is still best. Feral NE +18.8;
  BM Orc +3.6; Arms / Fury / Fury 2H Orc +3.3 / +5.5 / +5.9; MM NE +1.0 over Orc (baseline +0.6; noise).
  Balance High Order (fairness-20261118-1) still open, no new evidence.
- Gear: no synthetic set, request or database change; audit1b last 19/19 PASS at 5518f4873.

## Ranking

Top Lone Wolf 1291.3, bottom Smite 781.0, 1.65x. Feral stays 3rd (1171.2); BM 5th, Fury DW 6th in the sim.
Table in `orchestration/fairness_report.md` §1.
