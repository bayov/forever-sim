# tank_rotations pass 28

Master eceb090196 (2026-11-24), against pass 27's master c7e3b3e1ef and the baseline of 566067a394.
Quiet: no new finding.

## What changed

Gate (c7e3b3e .. eceb090): `tools/rotopt/warrior_tank.go` (`stoneformAt` knob), Stoneform from 20 s in the warrior
Balanced and Max survival requests and UI APLs (dfd0159, already checked at the end of pass 27). Engine changes since the
baseline: `sim/core/racials_forever.go` (Troll Regeneration, Forever Stoneform: 10% less Physical damage for 8 s, 3 min,
GCD, from client 1.60.1.71037), `base_stats.go` (Skyborne offsets EV-0093, the shaman's level 60 class row from a Dwarf
shaman's sheet EV-0181), `mana.go` (Innervate/Energized scale MP5), `buffs.go`. No encounter, tankbench or tankscore change.

## Baseline (1,000 iterations, default seed; `ref/tank_bench/summary.tsv` rewritten)

Every row is identical to the old baseline except the three Warden requests. The warrior presets are Undead and the
others aren't Dwarves or Trolls, so the new racials don't fire in any tank request.

| request (scoring mean) | score old -> new | SafeDPS | DTPS | healing needed |
|---|---|---|---|---|
| shaman_warden_balanced | 1284.5 -> 1264.7 | 1551.0 -> 1548.3 | 660.0 -> 662.0 | 658.7 -> 663.2 |
| shaman_warden_max_survival | -2245.5 -> -2263.5 | 1489.5 (same) | 671.2 -> 672.0 | 634.5 -> 636.5 |
| shaman_warden_max_threat | 2624.8 -> 2621.7 | 1708.7 -> 1706.8 | 683.2 -> 685.2 | 860.7 -> 857.7 |

Each Warden row moves within one or two of its own errors, but every boss moves the same way (DTPS +1 to +3, TPS -1 to
-5). That fits the sourced shaman base stat row (EV-0181, 0352d95dd), which is a data fix, not a rotation problem, so
there's no finding. Other scoring means: warrior Balanced 3454.2, Max threat 4233.3, Max survival 1832.8; paladin
2809.0 / 3243.7 / 1787.2; bear 2130.0 / 3131.8 / 320.7.

## Rotation lines

- Stoneform (warrior Balanced and Max survival, `currentTime >= 20s`): no tank request is a Dwarf, so the line does
  nothing here. Its timing was searched by the worker (`racials_1124/stoneform_i10000_s1.tsv`, docs/tank_warrior.md
  line 553) and pass 27 checked that the APLs match rotopt. The Stoneform aura uses the timer and doesn't read health,
  so it's fine in a scoring preset.

## Windfury main-hand timing (pass doc Open item, partial)

The test: keep main-hand abilities out of the totem's 100 ms proc cooldown (the ICD starts on a proc, `sim/core/buffs.go`
2131 and 2253). A `wait` for `auraInternalCooldown(10610)` goes in front of the first main-hand ability
(`tools/audit/tank_rotations/mk_wf_wait.py`). The runs are paired `abtest -iters 3000`, default seed, on all seven bosses:

| request | insert before | scoring-mean diff | se | verdict |
|---|---|---|---|---|
| warrior_protection_balanced | Revenge (12) | -31.3 | 5.8 | loses |
| warrior_protection_max_threat | Shield Slam (11) | -52.1 | 3.1 | loses |
| warrior_protection_max_survival | Revenge (12) | -6.6 | 9.4 | null |
| paladin_protection_balanced | top (1) | +0.2 | 1.6 | null |
| shaman_warden_balanced | Stormstrike (6) | -0.8 | 3.0 | null |
| druid_bear_balanced | top (1) | -1.9 | 4.6 | null |

For the warrior, the delay loses more threat than the extra rolls bring back (SafeDPS -9 to -37 on every boss). No gain
for the others. Not tried yet, left for a later pass: spreading abilities between swings, and holding Stormstrike for
Windfury Weapon's cooldown (no tank request uses Windfury Weapon). The paladin, Warden and bear gaps are within
noise, and pass 28 didn't confirm that the ICD aura resolves for those classes: a `null` there may mean the line never
fired.

## Open findings

-20261122-1, -20261104-11 and -20261104-3 are unchanged: no paladin APL or warrior trinket line changed.

## Repro

```sh
W=$AUDIT_WORK/tank_rotations; mkdir -p $W/base
go build -tags=with_db -o $AUDIT_WORK/tankbench ./tools/tankbench
for r in tools/tankbench/req/*.json; do for b in tools/encounters/*.json; do
  $AUDIT_WORK/tankbench -iters 1000 -encounter $b $r > $W/base/$(basename $r .json)__$(basename $b .json).txt; done; done
python3 tools/audit/tank_rotations/bench_summary.py   # then fix line 1's master/date
go build -tags=with_db -o $AUDIT_WORK/abtest ./tools/audit/tank_rotations/abtest
python3 tools/audit/tank_rotations/mk_wf_wait.py tools/tankbench/req/warrior_protection_balanced.json $W/B.json 12 10610
$AUDIT_WORK/abtest -iters 3000 tools/tankbench/req/warrior_protection_balanced.json $W/B.json tools/encounters/*.json
```
