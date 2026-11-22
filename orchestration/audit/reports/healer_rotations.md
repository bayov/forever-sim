# Healer rotations pass 23 (quiet)

- Master: 7a6d1beb9 (`last.master` was 8b5137c6d). Audited 2026-11-22. Client build 1.60.1.71037 (gate).
- Bench: `tools/healbench` built in the audit checkout, default seed, `-carry`, 1,000 iterations, via
  `tools/audit/healer_rotations/baseline.sh`. setup.sh's `protoc-gen-go` install had not landed in this session;
  `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` then `make sim/core/proto/api.pb.go` fixed it.

## What changed since 8b5137c6d (the 4 watched files)

- `sim/core/dot.go`: `SetTickPeriod` now retimes a dot on the grid of its own expiry (Wrack, warlock). Only
  Wrack calls it; no heal-over-time changes its period, so healer HoTs are untouched.
- `sim/paladin/item_sets_pve.go`: Soulforge Armor 4-piece reuses an existing Crusader's Wrath aura (Forever set
  1673 stacking with Lightforge). No healer preset wears either set.
- `sim/shaman/flame_shock.go`: Flame Shock's GCD takes haste under Forever. Not in any Restoration shaman APL.
- `sim/shaman/talents.go`: comment only (Maelstrom Weapon and Windfury Totem, Enhancement).
- No healer APL, request, encounter file, `sim/druid`/`sim/priest` heal or healbench main.go change.

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
