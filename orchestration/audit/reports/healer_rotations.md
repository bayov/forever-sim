# Healer rotations pass 17 (quiet; healer_rotations-20261115-1 and -20261104-2 still open)

- Master: c198d81e3 (`last.master` was 2f7e255af). Audited 2026-11-18.
- Bench: `tools/healbench` built from this checkout (Go protos generated with `make sim/core/proto/api.pb.go`
  first), default seed, `-carry`, 1,000 iterations, via `tools/audit/healer_rotations/baseline.sh` (about 14 min
  on 4 cores, while other auditors were running).

## What changed since 2f7e255af

- `sim/priest/priest.go`, `sim/priest/talents.go` (ac0217aed): `AddPartyBuffs` clears `PartyBuffs.dark_inspiration`
  for the party of a priest who has Dark Inspiration, because that priest already gives the party the real aura.
  None of the three healer priest requests has Shadow points (talents `005005031330011531-0350510303`,
  `00500503113-03505103030121531`, `00500503113-33505100030121531`), and no healbench request sets
  `dark_inspiration`. The new code can't reach the healers.
- No healer APL, healbench request, encounter file or `tools/healbench/main.go` changed.

## Baseline

All 63 rows (9 requests x 7 boss files, each in its own role) were re-simmed in full this time, since pass 16
stopped after 23 rows. Every row is byte-identical to `ref/healer_bench/summary.tsv`, so the reference stays
as it is. Rows that bear on the open findings:

| request | boss | carry | mana left | OOM at | other deaths |
|---|---|---|---|---|---|
| druid_restoration_raid | onyxia | 0.944 | 6,314 | - | 7.1% |
| priest_discipline | chromaggus | 0.871 | 341 | 101 s | 23.6% |

## Changes tried

None. No rotation or kit input moved, and the two open findings already carry their A/B numbers.

## Findings

No new findings. healer_rotations-20261115-1 (Wild Growth gate) and -20261104-2 (Discipline OOM on Chromaggus)
remain open and unchanged.
