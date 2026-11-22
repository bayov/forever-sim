# Healer rotations pass 22 (quiet)

- Master: 8b5137c6d (`last.master` was b4b68c52e). Audited 2026-11-22.
- Bench: `tools/healbench` built in the audit checkout (protos generated with
  `protoc -I=./proto --go_out=./sim/core ./proto/*.proto`; setup.sh still doesn't do it), default seed, `-carry`,
  1,000 iterations, via `tools/audit/healer_rotations/baseline.sh`.

## What changed since b4b68c52e

- `sim/paladin/talents.go` (1c80dafff, Nov 19 hotfix): Two-Handed Weapon Specialization now multiplies all Holy
  school damage (`SchoolDamageDealtMultiplier[Holy]`). Damage only, and a Retribution talent; heals don't take it.
- `sim/paladin/{consecration,judgement,seal_of_fury,soc,sor}.go`: the matching removal of the per-spell Holy
  multiplier. Damage spells, none in a Holy paladin healer APL's healing.
- `sim/shaman/talents.go`: Maelstrom Weapon PPM from client 1.60.1.70959 (Enhancement). `sim/shaman/stormstrike.go`
  (Enhancement). `sim/shaman/item_sets_pve.go`: comment only (Dungeon Set 1 in set 1679).
- No healer APL, request, encounter file, `sim/druid`/`sim/priest` or healbench main.go change.

## Baseline (1,000 it) vs ref

All 63 rows (9 requests x 7 boss files, each in its role) byte-identical to `ref/healer_bench/summary.tsv`, which
confirms none of the changes reaches the healers. Ref kept as is.

Rows behind the open findings:

| request | boss | carry | mana left | OOM | other died |
|---|---|---|---|---|---|
| druid_restoration_raid | onyxia | 0.946 | 6,303 | (1,403 s) | 5.8% |
| priest_discipline (tank) | chromaggus | 0.871 | 341 | 101 s | - |
| shaman_restoration_raid | magmadar | 1.778 | 7,638 | (2,322 s) | 0.0% |

## Rotation lines

No line changed and the baseline is identical, so the line-by-line findings of earlier passes stand; no new
one-line variants were tried this pass.

## Findings

No new findings. Open, unchanged: -20261118-1 (baseline moved by 7199c8f9c, paired confirmation still pending),
-20261115-1 (Wild Growth gate x0.75), -20261104-2 (Discipline tank OOM on Chromaggus).
