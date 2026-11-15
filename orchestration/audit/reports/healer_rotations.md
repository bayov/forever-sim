# Healer rotations pass 13 (quiet: no new findings; -20261104-2 still open)

- Master: 70ea588bb (`last.master` was d630df2cb). Audited 2026-11-15.
- Bench: `tools/healbench` built from this checkout (Go protos generated with `protoc -I=./proto --go_out=./sim/core`
  first; a fresh session lacks them), default seed, `-carry`, 1,000 iterations, every request on each of the 7 boss
  files at its own length in its role, via `tools/audit/healer_rotations/baseline.sh`.

## What changed since d630df2cb

- Gate: `sim/druid/ferocious_bite.go` (an avoided Ferocious Bite refunds 80% of its energy) and `sim/druid/forms.go`
  (Sharpened Claws now adds spell crit as well as melee crit in the feral forms, aura 290 A_MOD_CRIT_PCT). Both are
  feral-form only: the Restoration druid requests never shift and take no Sharpened Claws, so no heal moves.
- Otherwise: feral test results, report hashes, DPS report data, wsf_compare tools, audit docs.
- No healer APL, healbench request, encounter file, heal/mana engine code or `tools/healbench/main.go` changed.

## Baseline

All 63 rows (9 requests x 7 boss files) are identical to `ref/healer_bench/summary.tsv` in every column, so the
reference is unchanged. Rotation lines were not re-screened: nothing they depend on moved since pass 12.

## Open findings

- healer_rotations-20261104-2 (Discipline tank healer OOM on Chromaggus): unchanged, carry 0.871, 341 mana left,
  OOM at 101 s (byte-identical row). Needs a worker fix.
