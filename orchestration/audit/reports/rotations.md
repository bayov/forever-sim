# Rotations pass 84 (2026-11-26, master f9fa79917)

Result: quiet. Range 2af970d74..f9fa79917: Bear tank Phase 0 work (Lacerate EV-0040 fix, re-searched tank APLs),
Totem of Sustaining 4th Water Shield globe, rotopt Phase 0 tank/healer prices.

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py 2af970d74 f9fa79917b` | only `ui/feral_tank_druid/apls/*` (Balanced/Max threat drop Demoralizing Roar, Enrage under 80 Rage, Lacerate refresh under 10 s). Tank APLs, no DPS preset uses them |
| Req rotations changed | none (`tools/dps_report/req` untouched) |
| Non-test sim `.go` changed | `sim/druid/bear_attacks.go` (Lacerate hit counts stacks after the cast): bear-only spell, no req casts 1235827. `sim/shaman/water_shield.go` + `items.go` (Totem of Sustaining 23200 adds a globe): no req wears 23200. No preset reached, no regress run |
| Shared run vs published (`runs/vs_published.txt`) | 14 of 68 moved > 0.5 DPS, all within ±0.15% (noise level, no class pattern) |
| Open findings landed? | 20261126-3, 20261122-1, 20261112-1 still open (no change in range) |
| Baseline | unchanged: `baseline_ab3fff7_3000it_seed424242.tsv` |

## Optional

None new.
