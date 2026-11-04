# Hunter: round 70753 phase 5, talent re-search with the hit-trade gear steps

Brief: `briefs/phase5_hit_trade_class.md` (class hunter). Branch `phase5-hit-hunter`, from `phase5-hit-trade` aed188a4b.

## Summary

No Hunter build wears a hit step: every build's total hit on its Level 60 set is the 8.00% cap already, so the cap
step is +0 for all ten (the tool: `hit step +0 (...; cap step +0, total hit on the set 8.00% against 8%)`). The ranged
builds have no hit talent and Orc / Night Elf no hit racial. Survival's Surefooted is counted in its 8.00%.
`-hit-span` wasn't used (no dual wield, no hybrid), so the search never asked for a step that doesn't exist.

Talents: Marksmanship and Lone Wolf hold. Beast Mastery moves Spirit Bond 2 -> 1 for Frenzy 4 -> 5, which is flat (inside the error).
Survival moves Resourcefulness -1 for Rapid Killing +1. That is flat at 120 and 180 s and +7.4 at 300 s.

| Build | Old talents | New talents | Change | Gear (step) | Total hit old / new (cap 8) |
|---|---|---|---|---|---|
| beast-mastery (and -wolf, -tallstrider) | 5021201505021241-00510504005 | 5021201505011251-00510504005 | Spirit Bond 2 -> 1, Frenzy 4 -> 5 | `ui/hunter/gear_sets/level60_phase1_synthetic_ranged.gear.json` (+0) | 8.00 / 8.00 |
| marksmanship (and -wolf, -tallstrider) | 502-0053551501503051-5 | same | none | same set (+0) | 8.00 / 8.00 |
| lone-wolf | 5-0053552511503051-5 | same | none | same set (+0) | 8.00 / 8.00 |
| survival (and -wolf, -tallstrider) | 502-005005-500230230250122151 | 502-0050051-500230230250022151 | Resourcefulness -1 (Survival tree), Rapid Killing +1 (Marksmanship tree) | `ui/hunter/gear_sets/level60_phase1_synthetic_survival.gear.json` (+0) | 8.00 / 8.00 |

The variants keep their base build's talents (meta `baseBuild`); they were confirmed with the base build's new talents.

## DPS (20,000 iterations, seed 1, each length on its own, ± error)

Old: current talents, current gear, no `-hit-steps`. New: picked talents, `-hit-steps ... -hit-step +0`.

| Build | Old 120 / 180 / 300 | New 120 / 180 / 300 |
|---|---|---|
| beast-mastery | 1197.3±0.3 / 1195.6±0.2 / 1176.3±0.2 | 1197.0±0.3 / 1195.4±0.2 / 1176.0±0.2 |
| beast-mastery-wolf | 1151.2±0.3 / 1150.5±0.2 / 1132.7±0.2 | 1151.8±0.3 / 1150.7±0.2 / 1132.9±0.2 |
| beast-mastery-tallstrider | 1138.3±0.3 / 1136.9±0.2 / 1119.2±0.2 | 1138.8±0.3 / 1137.4±0.2 / 1119.6±0.2 |
| marksmanship | 1268.9±0.4 / 1254.3±0.3 / 1217.6±0.2 | same (same talents, step +0) |
| marksmanship-wolf | 1243.1±0.3 / 1228.3±0.3 / 1192.4±0.2 | same |
| marksmanship-tallstrider | 1236.0±0.3 / 1220.9±0.3 / 1185.8±0.2 | same |
| lone-wolf | 1332.4±0.4 / 1313.7±0.3 / 1264.7±0.3 | same |
| survival | 1048.6±0.4 / 1041.2±0.3 / 1032.7±0.2 | 1048.4±0.4 / 1040.6±0.3 / 1040.1±0.2 |
| survival-wolf | 1020.7±0.4 / 1014.5±0.3 / 1006.9±0.2 | 1021.2±0.4 / 1014.3±0.3 / 1014.4±0.2 |
| survival-tallstrider | 1013.8±0.4 / 1007.4±0.3 / 999.5±0.2 | 1014.2±0.4 / 1007.4±0.3 / 1007.2±0.2 |

The old numbers reproduce the round 70753 table (`orchestration/round_70753/hunter.md`).

Search scores (the mean of 120 / 180 / 300 s, 2,000 iterations a candidate): BM 1189.5 (from about 1188.6), SV 1042.9 (from 1041.4).
MM 1246.9 and LW 1303.6 found no move.

## Notes and odd things

- None of the talent moves has anything to do with hit: with every step at +0, this was a plain re-search of the round 70753 talents.
  - BM's move is noise: -0.3 / -0.2 / -0.3 on the Orc base and +0.2 to +0.6 on the variants. I'd keep the current BM talents.
  - Survival's Rapid Killing is worth one more Rapid Fire in a 300 s fight (+7.4 on all three builds) and nothing at 120 or 180 s.
    It was measured on seed 1 only. The phase 5 session can take or leave it.
- `-require frenzy` (BM keyTalents) fails, because the current build has Frenzy at 4/5 and the pin wants max rank ("starting talents break the tree rules").
  So Frenzy isn't pinned. The search put it to 5/5 anyway.
- The ranged set has only minus steps (hit-1 ... hit-4) and Survival's has hit-1, hit+1 ... hit+3. No build reached for any of them.
- To build the tool, the checkout needed the gitignored generated protos (`sim/core/proto/*.pb.go`). I generated them locally and committed nothing.

## Commands

Run from the repo root. The binary is `go build -tags=with_db -o rotopt ./tools/rotopt`. The knobs are those of `tools/rotopt/hunter_presets.sh` expanded:

```
BM_K="aimed=1,aimedMana=0,hawk=2,intimidation=0,serpent=1,sniper=1,sniperFirst=0,rapidPrepull=1,arcaneFirst=1,arcaneMana=0,serpentLate=2,serpentMana=0,holdElune=1,weave=1,weaveRaptor=1,weaveDelay=1.2,weaveShotSlack=0.75,weaveShotSlack=0"
MM_K="aimed=1,aimedMana=0,hawk=2,intimidation=0,serpent=1,sniper=1,sniperFirst=0,rapidPrepull=1,arcaneFirst=1,arcaneMana=0,serpentLate=0,serpentMana=0,aimedMana=0,holdRapid=1,holdElune=1,sniperManaPerSec=10,weave=1,weaveRaptor=1,weaveDelay=1.2,weaveShotSlack=0.75,weaveManaPerSec=25"
LW_K= MM_K with serpentLate=1
SV_K="lacerate=1,lacerateMana=20,strikerMana=0,trap=1,trapMana=30,wingClip=1,wingClipMana=50,wingClipHold=2,trapFirst=1"
RANGED_HIT="-hit-steps ui/hunter/gear_sets/level60_phase1_synthetic_ranged -hit-cap ranged:8"
SV_HIT="-hit-steps ui/hunter/gear_sets/level60_phase1_synthetic_survival -hit-cap melee:8"
RANGED_FREEZE=weaveReact,weaveMeleeWait,markPrepull,leylineMana,hawk,intimidation
SV_FREEZE=markPrepull,leylineMana
```

Talent searches (four, side by side):

```
rotopt -template hunter_ranged -settings tools/dps_report/req/hunter__beast-mastery.json -knobs "$BM_K" -talent-search -talents 5021201505021241-00510504005 -combined -durs 120,180,300 $RANGED_HIT -require bestialWrath,unleashedFury,ferocity,deadlyAspects,carefulAim,lethalAttacks,mortalShots -freeze $RANGED_FREEZE -iters 2000
rotopt -template hunter_ranged -settings tools/dps_report/req/hunter__marksmanship.json -knobs "$MM_K" -talent-search -talents 502-0053551501503051-5 -combined -durs 120,180,300 $RANGED_HIT -require deadlyAspects,carefulAim,lethalAttacks,mortalShots,improvedArcaneShot,sniperShot,barrage,focusedFire -freeze $RANGED_FREEZE -iters 2000
rotopt -template hunter_ranged -settings tools/dps_report/req/hunter__lone-wolf.json -knobs "$LW_K" -talent-search -talents 5-0053552511503051-5 -combined -durs 120,180,300 $RANGED_HIT -require loneWolf,deadlyAspects,carefulAim,lethalAttacks,mortalShots,improvedArcaneShot,sniperShot,rangedWeaponSpecialization -freeze $RANGED_FREEZE -iters 2000
rotopt -template hunter_sv -settings tools/dps_report/req/hunter__survival.json -knobs "$SV_K" -talent-search -talents 502-005005-500230230250122151 -combined -durs 120,180,300 $SV_HIT -require predatorsEdge,exposePrey,striderKick,laceratingStrikes,deadlyAspects,lightningReflexes,surefooted,cleverTraps -freeze $SV_FREEZE -iters 2000
```

Confirmation. Each `<d>` of 120, 180 and 300 is its own run. A variant takes its own request with its base build's template, knobs and talents:

```
old: rotopt -template <t> -settings tools/dps_report/req/hunter__<id>.json -knobs "<knobs>" -talents <old> -eval -confirm-iters 20000 -durs <d>
new: rotopt -template <t> -settings tools/dps_report/req/hunter__<id>.json -knobs "<knobs>" -talents <new> <hit flags> -hit-step +0 -eval -confirm-iters 20000 -durs <d>
hit: the same lines with -print-stats in place of -eval ... -durs
```
