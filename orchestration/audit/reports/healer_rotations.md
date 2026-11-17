# Healer rotations pass 15 (quiet; healer_rotations-20261115-1 and -20261104-2 still open)

- Master: 5f20dfdf4 (`last.master` was 1a85df1d7). Audited 2026-11-17.
- Bench: `tools/healbench` built from this checkout (Go protos generated with `protoc -I=./proto --go_out=./sim/core`
  first, after `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`), default seed, `-carry`, 1,000
  iterations, every request on each of the 7 boss files at its own length in its role, via
  `tools/audit/healer_rotations/baseline.sh`. About 40 minutes on a 4-core box shared with other auditors.

## What changed since 1a85df1d7

- `sim/shaman/item_sets_pve.go`: Forever's reworked Five Thunders set (1679), with its 2/3/4/6 piece bonuses.
  Neither `shaman_restoration` nor `shaman_restoration_raid` wears any of its 32 pieces (checked against
  `assets/database/db.json`), so the healer rows can't move.
- `sim/core/item_sets.go` (`ItemSetOf`) and `sim/core/pet.go` (`InheritedStats`): new read-only accessors for
  rotopt. No behavior change.
- No healer APL, healbench request, encounter file or `tools/healbench/main.go` changed.

## Baseline

All 63 rows are byte-identical to `ref/healer_bench/summary.tsv` (`cmp` clean), so the reference stays as it is.
Rows that open findings track:

| request | boss | carry | mana left | OOM at |
|---|---|---|---|---|
| druid_restoration_raid | onyxia | 0.944 | 6,314 | (1,419 s) |
| priest_discipline | chromaggus | 0.871 | 341 | 101 s |

## Rotation lines

Nothing changed in any rotation, request or the heal engine, so the line-by-line review and the 180/300 s mana
runs from pass 14 still stand. No new variants were tried.

## Findings

- New: none.
- Still open: healer_rotations-20261115-1 (Wild Growth gate, druid raid healer), healer_rotations-20261104-2
  (Discipline tank healer OOM on Chromaggus). Both re-confirmed unchanged on this master.
