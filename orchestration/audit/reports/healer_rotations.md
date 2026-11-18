# Healer rotations pass 18 (quiet; healer_rotations-20261115-1 and -20261104-2 still open)

- Master: 8219f1083 (`last.master` was c198d81e3). Audited 2026-11-18.
- Bench: `tools/healbench` built from the shared worktree `/tmp/audit-cache/cli/8219f1083fee/tree` (protos
  already generated there; the audit checkout has none), default seed, `-carry`, 1,000 iterations, via
  `tools/audit/healer_rotations/baseline.sh` run with that tree as cwd.

## What changed since c198d81e3

- `sim/core/dot.go` (8219f1083): under Forever, a dot with `AffectedByCastSpeed` now ticks at a ms-rounded hasted
  period; with `HasteReducesDuration` it keeps its tick count (Volley), otherwise it keeps its base duration and
  gains the ticks that fit (Hurricane). Era and dots without the flag are unchanged (`hastedTicks` returns the base
  tick length and 0 extra ticks).
- `sim/druid/hurricane.go`, `sim/hunter/volley.go`: the only two spells that set `AffectedByCastSpeed`
  (grep of `sim/`, non-test). No healer HoT (Renew, Rejuvenation, Regrowth, Wild Growth, Tranquility) or channel
  sets it, so the change can't reach the healers.
- No healer APL, healbench request, encounter file or `tools/healbench/main.go` changed.

## Baseline

All 63 rows (9 requests x 7 boss files, each in its own role) re-simmed in full: byte-identical to
`ref/healer_bench/summary.tsv`, so the reference stays. Rows that bear on the open findings:

| request | boss | carry | mana left | OOM at | other deaths |
|---|---|---|---|---|---|
| druid_restoration_raid | onyxia | 0.944 | 6,314 | - | 7.1% |
| priest_discipline | chromaggus | 0.871 | 341 | 101 s | 23.6% |

## Changes tried

None. No rotation or kit input moved, and the open findings already carry their A/B numbers.

## Findings

No new findings. healer_rotations-20261115-1 (Wild Growth gate) and -20261104-2 (Discipline OOM on Chromaggus)
remain open and unchanged.
