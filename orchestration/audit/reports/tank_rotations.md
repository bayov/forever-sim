# tank_rotations pass 35

master 34994816e7, 2026-11-27. Previous pass on f9fa79917b; baseline compared against `ref/tank_bench/summary.tsv`
(pass 34). Result: quiet (no new findings).

## What changed

No tank APL, tankbench request, rotopt tank template, encounter file, tankbench or tankscore change. The watched
changes are the tank `bis/*.bis.json` lists (doubtful Dawn quest items dropped as alternatives; no worn item changes)
and one engine change in `sim/common/item_effects.go`: Chained Essence of Eranikus (10455) cloud ticks now crit at a
fixed 2.7% (57 of 2151 log ticks, 2026-11-26 log audit) instead of at the wearer's spell crit.

## Baseline (1,000 iterations, default seed, scoring mean at the preset's weight)

| request | score | was | SafeDPS | was | healing needed | was |
|---|---|---|---|---|---|---|
| druid_bear_balanced | 3728.7 | 3771.5 | 2933.7 | 2955.2 | 758.5 | 758.5 |
| druid_bear_max_survival | 1025.3 | 1048.0 | 1903.2 | 1914.3 | 344.8 | 344.8 |
| druid_bear_max_threat | 5568.5 | 5606.7 | 3210.7 | 3229.7 | 918.0 | 918.0 |
| paladin_protection_balanced | 2954.7 | 2965.3 | 2092.5 | 2097.7 | 441.0 | 441.0 |
| paladin_protection_max_survival | 895.2 | 906.5 | 1671.5 | 1676.8 | 292.5 | 292.5 |
| paladin_protection_max_threat | 3925.8 | 3936.3 | 2200.2 | 2205.3 | 514.7 | 514.7 |
| shaman_warden_balanced | 2013.0 | 2022.7 | 1973.2 | 1977.8 | 698.3 | 698.3 |
| shaman_warden_max_survival | -1212.8 | -1212.8 | 1623.5 | 1623.5 | 548.7 | 548.7 |
| shaman_warden_max_threat | 3417.8 | 3428.0 | 2119.5 | 2124.7 | 893.2 | 893.2 |
| warrior_protection | -399.2 | -399.2 | 854.5 | 854.5 | 755.5 | 755.5 |
| warrior_protection_balanced | 3353.8 | 3360.5 | 2414.7 | 2418.3 | 500.0 | 500.0 |
| warrior_protection_max_survival | 1218.8 | 1218.8 | 1763.0 | 1763.0 | 265.7 | 265.7 |
| warrior_protection_max_threat | 4718.8 | 4725.8 | 2732.8 | 2735.8 | 743.5 | 743.5 |

Attribution: exactly the 10 requests that wear 10455 moved (all 7 bosses each); the 3 that don't (Warden Max
survival, warrior Max survival, legacy warrior_protection) are bit-identical. Healing needed is unchanged everywhere,
so the move is threat only. Bears move most (-1.1% at Balanced, SafeDPS -21) since their high crit fed the old roll;
plate tanks -5 to -11. The change is log-backed (mechanics/log-audit), so it is not a finding here. Whether 10455 is
still the bears' best trinket after the cut is tank_gear's question. New `ref/tank_bench/summary.tsv` written.

## Rotation lines

No rotation line changed since pass 34, and none reads the cloud's crit; no new A/B tried this pass.

## Findings

- Still open, user calls: 20261104-11 (warrior Max survival trinket hold), 20261104-3 (paladin Bulwark 13 s). Not re-tested.

## Open item not done

The Windfury main-hand timing item (pass doc, user request 2026-11-15) is still not done: it needs new APL
conditions for swing timing across four tank specs on two seed sets, more than one pass budget. Left in the pass doc.
