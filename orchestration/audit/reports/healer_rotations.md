# Healer rotations pass 7 (findings: 1 new, high; -2 and -3 still open)

- Master: 707ae92cc (`last.master` was 27d0d59d6). Audited 2026-11-11.
- Bench: `tools/healbench` built from this checkout (Go protos generated with `protoc -I=./proto --go_out=./sim/core
  ./proto/*.proto` first), healbench's default seed, `-carry`, 1,000 iterations, each request on each of the 7 boss
  files at the file's own length in its role. Script: `tools/audit/healer_rotations/baseline.sh [out.tsv] [iters]`
  (reads `$AUDIT_WORK/healbench`, 2 benches at once, about 25 minutes on 4 cores). It was missing again and is
  rewritten under `tools/audit/healer_rotations/`; the run's commit must include it.

## What changed

The gate fired on 9 files:

- `sim/druid/healing.go` (9628e85b6): Nature's Swiftness also makes Wild Growth instant (hotfix push 113885).
  This moves the Restoration druid raid healer (below).
- `sim/core/dot.go` (f20d5938a): no tick after a dot's aura ran out when the expiry is processed late, plus
  `SetTickPeriod` for Wrack. Measured no effect on any healer row (bench with only this file reverted gives
  the same druid raid numbers as master).
- `sim/druid/items.go`: Idol of Ferocity (feral only). `sim/shaman/*`: the Nov 10 Enhancement tuning (Mental
  Dexterity, Maelstrom Weapon PPM, Stormstrike, totems); neither shaman healer request takes those talents.
- No healer APL, healbench request, `tools/encounters/*.json` or `tools/healbench/main.go` changed.

## The baseline

56 of 63 rows byte-identical to the old `ref/healer_bench/summary.tsv`. The 7 Restoration druid raid healer rows
all fell (ref -> master, carry):

| boss | ref | master | change |
|---|---|---|---|
| onyxia 137 s | 0.997 | 0.943 | -0.054 |
| broodlord 44 s | 0.920 | 0.906 | -0.014 |
| chromaggus 115 s | 1.109 | 1.092 | -0.017 |
| golemagg 56 s | 2.660 | 2.658 | -0.002 |
| magmadar 54 s | 2.002 | 1.818 | -0.184 |
| ragnaros 97 s | 1.270 | 1.240 | -0.030 |
| vaelastrasz 31 s | 1.102 | 1.079 | -0.023 |

Bisected to 9628e85b6 (healing.go reverted alone restores the ref rows exactly). New ref written.

## The druid raid healer and Nature's Swiftness

NS is left to `autocastOtherCooldowns`, which raises the buff at 0 s. Before the hotfix the next affected cast was
line 4's emergency Regrowth (lowest under 60%), an instant big heal when it mattered. Now line 5's Wild Growth,
an ordinary cast, uses it first. Variants (raid role, own lengths, 1,000 it, paired seed):

| variant | ony | brood | chrom | golem | magma | rag | vael | mean |
|---|---|---|---|---|---|---|---|---|
| master | 0.943 | 0.906 | 1.092 | 2.658 | 1.818 | 1.240 | 1.079 | 1.391 |
| A: NS line before line 4, same condition | 0.998 | 0.919 | 1.108 | 2.655 | 2.009 | 1.274 | 1.108 | 1.439 |
| B: Wild Growth only when NS not up | 0.977 | 0.945 | 1.114 | 2.667 | 1.889 | 1.264 | 0.984 | 1.406 |
| C: NS line before Healing Touch 11 | 0.922 | 0.906 | 1.094 | 2.635 | 1.830 | 1.232 | 1.057 | 1.382 |

Mana side, A against master (carry, mana left): 180 s Magmadar 1.585 -> 1.660, Ragnaros 1.172 -> 1.191, Onyxia
0.823 -> 0.865; 300 s 1.467 -> 1.578 (6422 -> 6475), 1.059 -> 1.137 (4224 -> 4455), 0.762 -> 0.785 (3964 ->
4229). A wins at every length and on the old rule gives the old numbers, so it is safe either way. Filed as
healer_rotations-20261111-1 (high).

## Rotation lines

No other line changed since pass 2. Open: -2 (Discipline tank healer OOM on Chromaggus, row byte-identical) and
-3 (shaman tank healer's Lesser Healing Wave 5 dead line), both unchanged.

## Open items for the next pass

- Re-run the druid raid rows once -1 is fixed. The tank druid healer could also hold NS on an explicit line,
  but it has no Wild Growth, so autocast still reaches its emergency Regrowth (rows unchanged; not measured).
- Emergency Flash Heal at 35% for Disc (still not run). `-labeled` for tighter paired comparisons.
