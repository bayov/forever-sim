# Fairness pass 43 (2026-11-15, master 25ce81d)

Result: quiet (no new finding). Since 70ea588: client 70938 update (Imp Firebolt ranks, Dawn items, Wild Growth,
78c395984), the Nov 14 update (9ec69044f: Windfury Totem on every landed main hand melee hit, Judgement of Command
melee, Wildstalker 5pc Multi-Shot, item stats), standards 17 amended (f1ab9ccd1), Two-Handed Weapon Specialization's
Holy part (baa363615), JoC on melee defense (6ce77e5ef), re-tunes for Arms, Enhancement, Feral and Ret, and the DPS
Report re-sim of 30 builds (2694dc2ed). sim/core changed, so a full pass: every base request at every race (5k, seed 1,
the loop's 25ce81d CLI and tree, `AUDIT_WORK=/tmp/audit-work/fairness`).

## What moved against the 70ea588 baseline

| Build | Before | Now | Cause |
|---|---|---|---|
| Fury 2H | 1032.0 | 1103.6 (+6.9%) | WF Totem on Bloodthirst / Whirlwind MH strike (9ec69044f) |
| Fury DW | 1050.2 | 1092.8 (+4.1%) | same |
| Ret (Shock Ret, sub-Holy) | 983.8 (970.4, 974.3) | 1024.2 (1010.7, 1011.9) | Holy Strike and melee JoC roll WF Totem; JoC melee; twist re-tune (c26878d48) |
| Feral | 1156.5 | 1186.0 (+2.6%) | Claw/Shred/Rake roll the totem; Bite at 65 and totem gate (f9990ec4e) |
| Arms | 1041.7 | 1069.5 (+2.7%) | WF Totem rule; strikes wait out the 100 ms (a8ea43990) |
| Enhancement (3 variants) | 1054.1 | 1081.1 (+26 to +28 each) | Stormstrike rolls the totem; Stormstrike hold for WF Weapon (4d0c2ceac) |
| Survival (2 variants) | 1014.8 | 1035.2 (+20 each) | Mongoose Bite / Wing Clip now roll it |
| Subtlety | 863.4 | 866.6 | Hemorrhage now rolls it |

Every other row (casters, hunter ranged builds, the five other rogue builds, warlocks incl. Affliction, Fire) is
identical to 70ea588 to the decimal. Rogue SS, Backstab, Eviscerate etc. were already in the old client-flag list, so
the rogues' standing is consistent. report_hashes.txt moves exactly the 30 melee requests the report re-simmed: no
shared leak. Imp Firebolt (78c395984) touches only warlock pet ranks 2-4; the report's Imp row did not move.

## What I checked

- Rule symmetry: `windfuryTotemRollsOnForever` (sim/core/buffs.go:2216) = any `ProcMaskMeleeMH` hit without
  `SpellFlagNoWindfuryTotem` (paladin seal procs only), extra attacks excluded at buffs.go:2192. Other MH-mask procs
  (Ravager, Frostguard, warrior T2 prot 6pc Retaliate) are not in any DPS request.
- Rotation follow-ups were searched per melee class: Arms and Feral took the 100 ms gate, Fury lost with it,
  Enhancement holds for WF Weapon. Survival and Ret were not reported as gate-searched; the measured gains elsewhere
  were +0.2 to +2.0, below a finding.
- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Published vs current: builds.json matches the shared 5k run of master within 1.5 on all 68 builds. Metas lag for
  Affliction (+37.0), Lone Wolf / MM / BM (-13.6 to -21.9), four rogues (-10.6 to -19.6), Fire (-4.1). Same pattern as
  results-20261104-1 / -2; confirmed independently, not refiled.
- Best race: report race is best or within noise everywhere (MM Orc vs NE +0.3, MM variants +1.0 / +1.5, Fire Human
  vs Gnome +0.4, Combat Daggers Undead vs Gnome +0.6), unchanged from 70ea588.

## Ranking now (report race, 5k)

Feral 4th -> 3rd, Fury 2H 9th -> 5th, Ret 17th -> 15th, Survival 15th -> 14th. Top over bottom 1.65x. Full table:
`ref/fairness/table_baseline.md` and `orchestration/fairness_report.md` §1.

## Baseline

`ref/fairness/sweep_baseline.log` and `table_baseline.md` replaced with the 25ce81d full sweep.

## Open for the next pass

- Gate search for Survival and Ret under the totem rule (small, about +1 DPS expected; rotation auditor's area).
- Next full re-sweep on any sim/core, request or database change.
