# Healer rotations pass 28 (findings: new -20261125-2 (low), updates on -20261125-1, -20261104-2)

- Master: 85691655a (`last.master` was c3bbe33ea). Audited 2026-11-25.
- Bench: `tools/healbench` built on master (needs `protoc -I=./proto --go_out=./sim/core ./proto/*.proto`
  with `/root/go/bin` on PATH; `setup.sh` doesn't generate the protos), default seed, `-carry`, 1,000
  iterations, via `tools/audit/healer_rotations/baseline.sh`. Ref `ref/healer_bench/summary.tsv` refreshed.
- New helper: `tools/audit/healer_rotations/variant.py` adds a mana-percent gate to one rotation line of a
  request (for the one-line A/B runs).

## What changed since c3bbe33ea

- The wowsims/forever evidence batches: caster rules (efeb06b95: resist constant 8L+19, 75% partial cap,
  binary spells, overloads, Flurry, Shatter), tank threat and damage-taken rules (2f42f4a44), hunter, items and
  procs batches, and 22d2cf952 (six earlier rulings replaced by the log).
- No healer APL, healer request, encounter file, heal, healer talent, healbench main.go or filler
  (`tools/dps_report/req/mage__frost.json`) change.

## Baseline (1,000 it) vs ref

63 rows: 56 moved, 14 by more than 0.02 carry, mean carry +0.007, mean mana left -4. Tank rows moved at most
0.011 except paladin_holy golemagg.

| request | boss | carry ref -> now | other deaths |
|---|---|---|---|
| shaman_restoration_raid | broodlord | 0.852 -> 0.928 | 13.1% -> 10.1% |
| priest_holy_raid | magmadar | 2.003 -> 2.058 | |
| paladin_holy_raid | onyxia | 0.988 -> 0.937 | 5.8% -> 6.4% |
| paladin_holy_raid | golemagg | 2.542 -> 2.580 | |
| paladin_holy_raid | broodlord | 0.882 -> 0.918 | 10.7% |
| paladin_holy_raid | chromaggus | 1.149 -> 1.113 | |
| druid_restoration_raid | vaelastrasz | 0.964 -> 0.995 | 6.7% -> 5.4% |
| druid_restoration_raid | magmadar | 1.790 -> 1.815 | |
| shaman_restoration_raid | magmadar | 1.781 -> 1.806 | |
| priest_holy_raid | golemagg | 2.484 -> 2.508 | |
| paladin_holy | golemagg | 1.778 -> 1.800 | |
| druid_restoration_raid | ragnaros | 1.246 -> 1.267 | |
| paladin_holy_raid | vaelastrasz | 1.021 -> 1.042 | |
| druid_restoration_raid | chromaggus | 1.133 -> 1.113 | |

Cause: healbench built at efeb06b95 already gives the new values on four sample rows, unchanged through
4f39862f4, becbcd30d and 22d2cf952. The resist rule changes the raid's magic damage taken and the random
stream. It is an evidence-backed rule, so -20261125-2 is low and needs no ruling.

Mana failures at judged lengths, unchanged: priest_discipline Chromaggus 0.828, OOM 95 s (-20261104-2);
priest_holy Chromaggus 0.914, OOM 114 s (-20261125-1); priest_holy Golemagg 300 s OOM 271 s.

## Changes tried (priest_holy tank, paired, 1,000 it)

| variant | chromaggus | golemagg 300 s | ragnaros 300 s | other files |
|---|---|---|---|---|
| master | 0.914 / OOM 114 s | 0.932 / OOM 271 s | 0.323 | |
| Greater Heal 5 only at >= 30% mana | 0.912 / 114 s | 0.932 | 0.323 | same |
| Greater Heal 5 only at >= 50% mana | 0.920 / 114 s | 0.932 | 0.323 | Magmadar -0.003 |
| Renew only at >= 30% mana | 0.905 / no OOM | 0.911 | 0.320 | same |
| Renew only at >= 50% mana | 0.902 / no OOM | 0.895 | 0.309 | Magmadar +0.009 |

None wins: gating Renew stops the OOM but costs carry. Greater Heal 5 at 50% is +0.006 on Chromaggus, within
noise at 1,000 it. The Holy tank mana fix needs a worker (a cheaper pattern), as -20261125-1 says.

## Open findings

-20261125-2 (new, low, baseline history), -20261125-1 (Holy OOM, tried gates), -20261124-1 (filler pin),
-20261118-1, -20261115-1, -20261104-2 (Disc OOM), and the rest unchanged.
