# Healer rotations pass 19 (findings: healer_rotations-20261118-1)

- Master: a6677362f (`last.master` was 8219f1083). Audited 2026-11-18.
- Bench: `tools/healbench` built in the audit checkout (protos generated with `protoc -I=./proto --go_out=./sim/core
  ./proto/*.proto`; setup.sh doesn't do it), default seed 101, `-carry`, 1,000 iterations, via
  `tools/audit/healer_rotations/baseline.sh`.

## What changed since 8219f1083

- 7199c8f9c: `readyNotBefore` (idle skip) now stops at the first hit or heal on a heal candidate the healer
  doesn't see yet, a reaction time on. Before, a skipping healer saw a hit up to several 50 ms wake-ups late,
  and the skip's adaptivity (aplIdleTries, kept across iterations) made results depend on earlier fights and core
  count. The only behaviour change for healers in the window.
- bbfd014e4 (heal_sight.go cache, heal target and health-history speedups, spell metrics), c8b5275b1 (DPS report
  switches folded into sim code, inert without env vars), c788f8717 (PGO profile): claimed bit-identical.
- No healer APL, request, encounter file or healbench main.go change.

## Baseline (1,000 it) vs ref

42 of 63 rows moved; 21 held byte-identical (all Holy paladin tank, Discipline and Holy priest tank rows,
most of whose fights rarely idle). Moves over 0.02 carry:

| request | boss | ref | now | delta |
|---|---|---|---|---|
| shaman_restoration_raid | magmadar | 1.887 | 1.778 | -0.109 |
| shaman_restoration | onyxia | 1.340 | 1.391 | +0.051 |
| shaman_restoration | golemagg | 1.938 | 1.986 | +0.048 |
| shaman_restoration | magmadar | 2.062 | 2.102 | +0.040 |
| druid_restoration_raid | broodlord | 0.911 | 0.873 | -0.038 |
| shaman_restoration_raid | golemagg | 2.674 | 2.636 | -0.038 |
| priest_holy_raid | magmadar | 2.062 | 2.100 | +0.038 |
| priest_holy_raid | ragnaros | 1.340 | 1.378 | +0.038 |
| shaman_restoration_raid | broodlord | 0.885 | 0.918 | +0.033 |
| shaman_restoration | broodlord | 0.978 | 0.948 | -0.030 |
| paladin_holy_raid | magmadar | 1.914 | 1.887 | -0.027 |
| shaman_restoration | chromaggus | 0.984 | 1.009 | +0.025 |
| paladin_holy_raid | golemagg | 2.702 | 2.679 | -0.023 |
| shaman_restoration_raid | chromaggus | 1.115 | 1.092 | -0.023 |

Signs are mixed and mana left barely moves (within ~100), consistent with a reshuffled stream plus earlier
reactions. Death-floor notes: shaman raid Onyxia other deaths 7.4% -> 10.6%; shaman tank Broodlord 3.6% -> 5.7%;
druid raid Onyxia 7.1% -> 5.8%. Filed as healer_rotations-20261118-1. `ref/healer_bench/summary.tsv` replaced.

## Not done (cut short by the run's time limit)

- Paired old (8219f1083) vs new at 4,000 it on seeds 101 and 7001 for the big movers, to split noise from the
  sight change. An old healbench builds from `/tmp/audit-cache/cli/8219f1083fee/tree` (`go build -tags=with_db
  ./tools/healbench`).
- Step 3 (per-line rotation checks, 180/300 s mana side) not re-run this pass: no rotation changed.

## Open findings

- -20261118-1 (new): baseline moved by 7199c8f9c, unconfirmed.
- -20261115-1: druid raid Wild Growth gate; Onyxia now 0.946, 5.8% other deaths.
- -20261104-2: Disc tank OOM on Chromaggus; unchanged (0.871, 341 mana, OOM 101 s).
