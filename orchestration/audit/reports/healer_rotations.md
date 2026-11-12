# Healer rotations pass 8 (quiet: no new findings; -1, -2, -3 still open)

- Master: cb0d928fd (`last.master` was 33db32b36). Audited 2026-11-12.
- Bench: `tools/healbench` built from this checkout (Go protos generated with `protoc -I=./proto --go_out=./sim/core
  ./proto/*.proto` first), healbench's default seed, `-carry`, 1,000 iterations, each request on each of the 7 boss
  files at the file's own length in its role. Script: `tools/audit/healer_rotations/baseline.sh [out.tsv] [iters]
  [extra flags]` (reads `$AUDIT_WORK/healbench`, 2 benches at once, about 12 minutes on 4 cores). It was missing
  again and is rewritten; it now writes the summary.tsv format directly. The run's commit must include it.

## What changed

The gate fired on 3 files, all from cb0d928fd (talent tooltip text):

- `sim/druid/talents.go`: comment only (Rend and Tear).
- `sim/priest/priest.go`: `TalentTreeSizes` Shadow 18 -> 17. The commit notes the last size is never read; the
  three priest requests have no Shadow points anyway.
- `sim/shaman/talents.go` + `proto/shaman.proto`: `ImprovedFireNova` renamed `TrialByFire` (same field). Neither
  shaman healer request takes it.
- No healer APL, healbench request, `tools/encounters/*.json` or `tools/healbench/main.go` changed.

## The baseline

All 63 rows (9 requests x 7 boss files) are byte-identical to `ref/healer_bench/summary.tsv` (`diff` empty), so
the ref is left as it is. No row moved.

## Rotation lines

No rotation changed since pass 7, so no new one-line A/B runs. Open findings re-checked by the baseline:

- -20261111-1 (Resto druid raid: Nature's Swiftness eaten by Wild Growth): still open, rows unchanged
  (Magmadar 1.818). Proposed fix (explicit NS line, +0.048 mean carry) not applied yet.
- -20261104-2 (Disc tank healer OOM on Chromaggus at 101 s): unchanged.
- -20261104-3 (Resto shaman tank LHW 5 dead line): unchanged, APL untouched.

## Open items for the next pass

- Emergency Flash Heal at 35% for Disc (still not run). `-labeled` for tighter paired comparisons.
