# Healer rotations pass 24 (findings)

- Master: 566067a39 (`last.master` was c78d983ac). Audited 2026-11-24.
- Bench: `tools/healbench` built on master, default seed, `-carry`, 1,000 iterations, via
  `tools/audit/healer_rotations/baseline.sh`.

## What changed since c78d983ac (the 3 watched files)

- `sim/druid/rip.go`, `sim/druid/rake.go`: Rip and Rake tick attack power shares from Forever logs (be773f5eb).
  Feral bleeds only; no Restoration druid APL casts them.
- `sim/druid/tuning_70753.go`: comment only.
- No healer APL, healer request, encounter file, heal, talent or healbench main.go change.

## Baseline (1,000 it) vs ref

- Tank role (35 rows): byte-identical.
- Raid role (28 rows): all moved, 16 by more than 0.02 carry, signs mixed (largest: paladin_holy_raid golemagg
  2.679 -> 2.775, vaelastrasz rows -0.03 to -0.05 for paladin/priest/shaman, +0.052 druid).
- Cause, confirmed: the ref predates b8b89d6a8 (Frost mage Ice Lance re-tune, 2026-11-23), which changed
  `tools/dps_report/req/mage__frost.json`, healbench's default `-filler`. Master healbench with
  `-filler` = the pre-b8b89d6a8 file reproduces the old ref row exactly (paladin_holy_raid golemagg 2.679).
  The raid healers' score depends on a DPS preset's talents: finding healer_rotations-20261124-1.
- Ref `ref/healer_bench/summary.tsv` refreshed on master.

| request (raid) | boss | carry ref | carry now |
|---|---|---|---|
| druid_restoration_raid | magmadar | 1.857 | 1.907 |
| druid_restoration_raid | vaelastrasz | 1.042 | 1.094 |
| paladin_holy_raid | golemagg | 2.679 | 2.775 |
| paladin_holy_raid | ragnaros | 1.270 | 1.315 |
| priest_holy_raid | magmadar | 2.100 | 2.145 |
| priest_holy_raid | vaelastrasz | 1.264 | 1.217 |
| shaman_restoration_raid | chromaggus | 1.092 | 1.127 |
| shaman_restoration_raid | vaelastrasz | 1.139 | 1.096 |

## Rotation lines

Not re-tried this pass: no healer rotation or heal changed. Open findings unchanged: -20261118-1, -20261115-1,
-20261104-2.

## Left over

- A stray binary `/healbench_old` (old-build check) was left at the filesystem root; the safety check refused
  its removal. Delete it by hand.
