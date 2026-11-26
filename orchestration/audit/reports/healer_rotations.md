# Healer rotations pass 29 (quiet: no new finding; update on -20261124-1, -20261125-2 closed)

- Master: ab3fff700 (`last.master` was bfbee78fc; the ref was last written on 85691655a). Audited 2026-11-26.
- Bench: `tools/healbench` built on master (protos via `protoc -I=./proto --go_out=./sim/core ./proto/*.proto`
  with `/root/go/bin` on PATH), default seed, `-carry`, 1,000 iterations, via
  `tools/audit/healer_rotations/baseline.sh` (7 min). Ref `ref/healer_bench/summary.tsv` refreshed.

## What changed since bfbee78fc

- The one watched file, `sim/core/spell_result.go` (1223e764e), adds a branch inside the
  `sim.Encounter.leveling != nil` block (swings from behind on the leveling player). Healbench never sets
  a leveling session, so it is inert there; confirmed by healbench built at bfbee78fc, 143bf4f72, 8f5b59325,
  1223e764e, 81a2e51fc and 270731faa all giving shaman_restoration_raid broodlord 0.881 / 13.5%, same as master.
- No healer APL, healer request, encounter file, heal, healer talent or healbench main.go change.
- Between the ref's master (85691655a) and bfbee78fc the filler `tools/dps_report/req/mage__frost.json`
  changed (7829e00f1/14de5d62d, Phase 0 gold line: items 22339 -> 272406, 22408 -> 15282 suffix 1958).

## Baseline (1,000 it) vs ref

63 rows: 28 moved (all raid-role rows; tank rows byte-identical), 18 by more than 0.02 carry, mean carry
-0.010, mean mana left +2.

| request | boss | carry ref -> now | other deaths |
|---|---|---|---|
| paladin_holy_raid | golemagg | 2.580 -> 2.489 | |
| paladin_holy_raid | vaelastrasz | 1.042 -> 0.976 | 4.7% -> 5.3% (now over 5%) |
| paladin_holy_raid | onyxia | 0.937 -> 0.991 | 6.4% -> 5.9% |
| paladin_holy_raid | ragnaros | 1.281 -> 1.248 | |
| paladin_holy_raid | chromaggus | 1.113 -> 1.137 | |
| shaman_restoration_raid | broodlord | 0.928 -> 0.881 | 10.1% -> 13.5% |
| shaman_restoration_raid | golemagg | 2.518 -> 2.469 | |
| shaman_restoration_raid | ragnaros | 1.223 -> 1.177 | |
| shaman_restoration_raid | chromaggus | 1.097 -> 1.052 | 1.7% -> 3.9% |
| shaman_restoration_raid | vaelastrasz | 1.102 -> 1.057 | |
| shaman_restoration_raid | magmadar | 1.806 -> 1.775 | |
| shaman_restoration_raid | onyxia | 0.932 -> 0.903 | 7.9% -> 9.2% |
| druid_restoration_raid | golemagg | 2.536 -> 2.494 | |
| druid_restoration_raid | vaelastrasz | 0.995 -> 0.963 | 5.4% -> 5.3% |
| druid_restoration_raid | ragnaros | 1.267 -> 1.238 | |
| druid_restoration_raid | broodlord | 0.916 -> 0.893 | 11.9% -> 14.6% |
| priest_holy_raid | golemagg | 2.508 -> 2.462 | |
| priest_holy_raid | onyxia | 0.988 -> 0.963 | 5.3% -> 6.1% |

Cause confirmed: master healbench with `-filler` = `git show 85691655a:tools/dps_report/req/mage__frost.json`
reproduces the old ref exactly on five rows (shaman raid broodlord 0.928, paladin raid golemagg 2.580 and
onyxia 0.937, druid raid vaelastrasz 0.995, priest raid onyxia 0.988). This is the third time a DPS gear
change re-scaled the raid healers; recorded under -20261124-1 (pin the filler), not a new finding.

## Rotation lines

No rotation changed since pass 28, which tried the Holy priest tank mana gates (-20261125-1) and found none
that wins; not re-run. Open items carried: -20261125-1 (Holy priest Chromaggus OOM, worker fix),
-20261124-1 (filler pin), -20261118-1 (paired confirmation of 7199c8f9c), -20261115-1 (Wild Growth gate x0.75).

## Findings

- None new. -20261124-1: update (third filler-driven move). -20261125-2: closed (recorded, undisputed).
