# tank_rotations pass 10

Master 8306f1d52 (2026-11-12), against the pass 8 baseline (707ae92cc; pass 9 spot-checked cb0d928fd and the loop
recorded a0b4d3be9 without benching it). 1,000 iterations, default seed.

## What changed

Since a0b4d3be9: c380bbaef (Rake's direct hit +6.5% AP: cat only, no tank APL casts Rake), 368b2e181 (hunter pet
follows owner melee speed: `sim/core/pet.go`, `unit.go`; only hunter pets set `FollowsOwnerMeleeSpeed`, no tank raid
has one), rotopt phase0 tooling (DPS only). No tank APL, request, rotopt tank template, encounter, tankbench or
tankscore change. Check: tankbench built at a0b4d3be9 gives bit-identical Onyxia rows for druid_bear_balanced and
warrior_protection_balanced, so this range moves nothing.

Unbenched until now: a0b4d3be9 (Windfury Totem rolls on main hand and on-next-swing abilities, from client flags).
Maul's ranks are now in `sim/core/windfury_totem_procs_forever.go`, which fixes tank_rotations-20261110-1 (resolved).

## Baseline (scoring mean of the 6 bosses without Ragnaros; full grid in ref/tank_bench/summary.tsv, rewritten)

| request | weight | score | SafeDPS | TPS | DTPS | healing needed |
|---|---|---|---|---|---|---|
| druid_bear_balanced | 1 | 1717.0 -> 1856.2 | 1519.0 -> 1569.3 | 1428.2 -> 1484.8 | 586.8 -> 586.0 | 471.7 -> 462.0 |
| druid_bear_max_survival | 3 | -578.3 -> -296.5 | 1343.2 -> 1394.2 | 1262.3 -> 1318.7 | 545.2 -> 543.7 | 393.3 -> 375.7 |
| druid_bear_max_threat | 0.333 | 2654.8 -> 2771.5 | 1616.3 -> 1672.8 | 1505.0 -> 1567.7 | 612.2 -> 611.7 | 638.7 -> 633.3 |
| paladin_protection_balanced | 1 | 2753.2 -> 2753.2 | 1688.5 -> 1688.5 | 1501.3 -> 1501.3 | 314.3 -> 314.3 | 245.3 -> 245.3 |
| paladin_protection_max_survival | 3 | 1554.0 -> 1554.0 | 1504.2 -> 1504.2 | 1337.5 -> 1337.5 | 298.3 -> 298.3 | 197.3 -> 197.3 |
| paladin_protection_max_threat | 0.333 | 3171.8 -> 3171.8 | 1693.5 -> 1693.5 | 1506.5 -> 1506.5 | 306.2 -> 306.2 | 256.5 -> 256.5 |
| shaman_warden_balanced | 1 | 1124.7 -> 1124.7 | 1556.7 -> 1556.7 | 1379.0 -> 1379.0 | 675.0 -> 675.0 | 716.8 -> 716.8 |
| shaman_warden_max_survival | 3 | -2736.0 -> -2736.0 | 1502.8 -> 1502.8 | 1348.8 -> 1348.8 | 684.7 -> 684.7 | 691.2 -> 691.2 |
| shaman_warden_max_threat | 0.333 | 2495.7 -> 2495.7 | 1626.2 -> 1626.2 | 1437.7 -> 1437.7 | 718.0 -> 718.0 | 823.3 -> 823.3 |
| warrior_protection | 1* | -46.8 -> -14.2 | 1031.5 -> 1048.2 | 1263.3 -> 1301.5 | 636.5 -> 637.3 | 756.3 -> 757.0 |
| warrior_protection_balanced | 1 | 3226.5 -> 3434.3 | 2100.5 -> 2191.2 | 1952.3 -> 2046.2 | 379.3 -> 377.3 | 342.8 -> 333.3 |
| warrior_protection_max_survival | 3 | 1561.8 -> 1771.3 | 1822.3 -> 1897.5 | 1778.2 -> 1862.5 | 336.7 -> 334.8 | 251.8 -> 245.7 |
| warrior_protection_max_threat | 0.333 | 3957.7 -> 4216.3 | 2196.0 -> 2327.2 | 2028.5 -> 2152.2 | 415.2 -> 415.0 | 443.0 -> 445.2 |

Bear (+3.3% SafeDPS, every boss) and warrior (+4%) rows rise from Windfury on Maul / warrior abilities; paladin and
warden rows are bit-identical to pass 8. Per-boss score_err 15-65; every bear/warrior move is well past it and
attributed to a0b4d3be9, the fix of an open finding, so no new finding.

## Changes tried (paired, 1,000 iterations, 7 bosses, Balanced weight)

| change (druid_bear_balanced) | scoring mean diff | paired err | verdict |
|---|---|---|---|
| Maul queued at >= 10 Rage (from 20) | -2.3 | 3.5 | null |
| Maul queued at >= 30 Rage (from 20) | +2.7 | 5.2 | null |

Maul gaining Windfury doesn't move its best Rage gate. The other lines were tested in pass 8 and nothing they read
changed except the Windfury rule; a fresh rotopt pass on the bear and warrior presets is optional, not a finding.

## Open findings

tank_rotations-20261111-1 (warden re-optimize), -20261104-11, -20261104-3: unchanged, no input they read moved.

## Note for the next pass

Building needs Go protos: `protoc -I=./proto --go_out=./sim/core ./proto/*.proto` (PATH must include /root/go/bin).
