# Healer rotations pass 25 (findings: update on -20261124-1, no new finding)

- Master: eceb09019 (`last.master` was e29cfe00f). Audited 2026-11-24.
- Bench: `tools/healbench` built on master (after `protoc -I=./proto --go_out=./sim/core ./proto/*.proto`;
  `setup.sh` doesn't generate the protos), default seed, `-carry`, 1,000 iterations, via
  `tools/audit/healer_rotations/baseline.sh`.

## What changed since e29cfe00f

- Watched: `sim/core/mana.go` (ed1df1d51, EV-0114): Innervate and Read Leyline's Energized now scale the MP5 too,
  except the MP5 the sim adds for Blessing of Wisdom and Mana Spring (`EnergizeMP5`). Reaches the Restoration
  druid presets (self-Innervate at 35% mana).
- Not watched but healer-relevant: e19d2df85 (Skyborne Windshaper base offsets: Int -1, Spi +3 against the Human,
  was +1/0; druid and shaman presets are Windshaper), 0352d95dd (shaman class row Int 90 -> 87), and 058a07457
  (`tools/dps_report/req/mage__frost.json`, healbench's default `-filler`, moved to the Phase 0 set).
- No healer APL, healer request, encounter file, heal, talent or healbench main.go change.

## Baseline (1,000 it) vs ref

| request | rows same | moved | over 0.02 |
|---|---|---|---|
| druid_restoration | 0 | 7 | 0 |
| shaman_restoration | 0 | 7 | 0 |
| paladin_holy, priest_discipline, priest_holy | 21 | 0 | 0 |
| druid_restoration_raid | 0 | 7 | 4 |
| paladin_holy_raid | 0 | 7 | 4 |
| priest_holy_raid | 0 | 7 | 6 |
| shaman_restoration_raid | 0 | 7 | 6 |

- Tank druid: carry within +0.005; Chromaggus mana left 2,276 -> 2,711 (projected OOM 191 -> 228 s), from the
  Innervate MP5 scaling. Tank shaman: carry within 0.009, about 90 mana less left everywhere (Int -4 net).
- Raid rows: all moved by the filler. Master healbench with `-filler` = e29cfe00f's frost request gives the old ref
  on all 28 rows within 0.013 (paladin and shaman rows exact), so the raid moves are 058a07457's, filed as an
  update on healer_rotations-20261124-1. Only residual above 0.02: druid_restoration_raid magmadar -0.032
  (Innervate/Windshaper; same-seed but a different random stream, carry far above 1.0).

| request (raid) | boss | carry ref | carry now | other deaths |
|---|---|---|---|---|
| paladin_holy_raid | golemagg | 2.775 | 2.527 | 0.0% -> 0.1% |
| paladin_holy_raid | ragnaros | 1.315 | 1.248 | |
| shaman_restoration_raid | golemagg | 2.667 | 2.495 | |
| shaman_restoration_raid | onyxia | 0.940 | 0.889 | 7.8% -> 9.3% |
| priest_holy_raid | magmadar | 2.145 | 2.014 | |
| priest_holy_raid | onyxia | 1.007 | 0.980 | 4.6% -> 7.0% |
| druid_restoration_raid | vaelastrasz | 1.094 | 0.971 | 3.2% -> 5.7% |
| druid_restoration_raid | golemagg | 2.622 | 2.527 | |
| druid_restoration_raid | onyxia | 0.951 | 0.988 | 6.9% -> 6.5% |

- Ref `ref/healer_bench/summary.tsv` refreshed on master.

## Rotation lines tried (paired, default seed)

Innervate is stronger now, so its 35% trigger was re-tested on both druid presets.

| variant | tank, own lengths | raid, own lengths | tank 180/300 s (Chromaggus, Golemagg, Ragnaros) |
|---|---|---|---|
| Innervate at 20% | Magmadar +0.051 at 1k, +0.017 at 4k (2.074 -> 2.091); others within 0.004 | identical carry | within 0.005 either way |
| Innervate at 50% | Golemagg -0.029, Magmadar +0.002, Ragnaros -0.006, Chromaggus +0.008 | identical carry | not run |

No win worth a change: the 35% line stands. The raid druid never gets low enough for the trigger to matter.

Mana side: the druid tank healer runs dry on Chromaggus at 180 s (153 s) and 300 s (147 s) on every variant;
at Chromaggus's own length it doesn't. Recorded here only, like the Discipline OOM (-20261104-2), since that
fight held at 180 s is a mana-starved case for every tank healer.

## Open findings

-20261124-1 (updated), -20261118-1, -20261115-1, -20261104-2: rows otherwise as the table shows; the
Wild Growth gate and Discipline variants were not re-measured.
