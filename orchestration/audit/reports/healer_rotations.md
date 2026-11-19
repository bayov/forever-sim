# Healer rotations pass 20 (quiet)

- Master: 8f388c7f3 (`last.master` was eebd00b2b). Audited 2026-11-19.
- Bench: `tools/healbench` built in the audit checkout (protos generated with `protoc -I=./proto --go_out=./sim/core
  ./proto/*.proto`; setup.sh still doesn't do it), default seed 101, `-carry`, 1,000 iterations, via
  `tools/audit/healer_rotations/baseline.sh`.

## What changed since eebd00b2b

- 65f0ea9d6: `swiftmendHot` (sim/druid/healing.go) loops the Rejuvenation and Regrowth rank lists directly,
  skipping nil ranks, instead of building a filtered copy with `balanceSpells` on each call. Same ranks, same
  order, same tie rule (strict `<` on expiry), so the same HoT is consumed. A speedup, no behaviour change.
- No healer APL, request, encounter file or healbench main.go change.

## Baseline (1,000 it) vs ref

All 63 rows (9 requests x 7 boss files, each in its role) byte-identical to `ref/healer_bench/summary.tsv`,
which confirms the refactor is a no-op. Ref kept as is.

Rows behind the open findings:

| request | boss | carry | mana left | OOM | other died |
|---|---|---|---|---|---|
| druid_restoration_raid | onyxia | 0.946 | 6,303 | (1,403 s) | 5.8% |
| priest_discipline (tank) | chromaggus | 0.871 | 341 | 101 s | - |
| shaman_restoration_raid | magmadar | 1.778 | 7,638 | (2,322 s) | 0.0% |

## Rotation lines

No line changed, and the baseline is identical, so the line-by-line findings of earlier passes stand; no new
one-line variants were tried this pass.

## Findings

No new findings. Open, unchanged: -20261118-1 (baseline moved by 7199c8f9c, paired confirmation still pending),
-20261115-1 (Wild Growth gate x0.75), -20261104-2 (Discipline tank OOM on Chromaggus).
