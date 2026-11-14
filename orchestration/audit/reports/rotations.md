# Rotations pass 65 (2026-11-14, master 1bf5feb4b)

Last audited master: 3a86d4a6b. Client 1.60.1.70888 data (unchanged). Result: quiet.

## What changed (3a86d4a6b..1bf5feb4b)

| path | change | reaches presets? |
|---|---|---|
| `sim/core/statweight.go` (2745da6ce) | stat-weight attribute step 1 -> 10 (Kings rounding) | no: stat weights only; the `.results` diffs are weight lines |
| `ui/enhancement_shaman/apls/magma.apl.json` (7858b83e5) | Flame Shock moved above Stormstrike | yes: shaman__enhancement-water-shield req rotation |
| `tools/dps_report/req/shaman__enhancement-imp-gw.json` | gear piece 990007 -> 990040 (hit-1 step) | gear only, no rotation change |
| `tools/rotopt/*`, BiS list, report data | optimizer and report output | no rotation logic |

`apl_logic_diff.py 3a86d4a6bf60 1bf5feb4b789`: 1 APL file touched (magma, the reorder above). No baseline refresh: no
sim change that the presets reach.

## Verified merged change: Water Shield Flame Shock above Stormstrike

The worker's re-search (seeds 1 / 1000003) reported +0.4 to +3.0. Spot-check with `ab.py` (20k it, 8 blocks,
shared CLI 1bf5feb4b789), A = master req, B = the old rotation (Stormstrike first), B-A at 120 / 180 / 300 s:

| set | 120 | 180 | 300 |
|---|---|---|---|
| S0 1e6 | -2.63 ± 1.08 | -1.95 ± 0.76 | -1.27 ± 0.45 |
| S0 2e8 | -2.15 ± 0.69 | -2.30 ± 0.51 | -0.88 ± 0.35 |
| pooled | -2.39 | -2.13 | -1.08 |

A at 180 s: 1000.9. The new order wins at every length; the merge holds.

## Open findings (all still unapplied on 1bf5feb4b)

- rotations-20261112-1: rogue poisons swap (Mutilate, Seal Fate, Subtlety still `mainHandPoison: DeadlyPoison`).
- rotations-20261109-1: Fury 2H 90-rage Heroic Strike line still present.
- rotations-20261105-1: Feral Ferocious Bite gate still 45.

## Optional

None new.
