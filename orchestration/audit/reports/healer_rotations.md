# healer_rotations pass 29

Branch phase0-tankheal 3e9748641 (last: master 85691655a). The finished healer requests moved to their Phase 0 sets
(equipment only: no APL, talent or consume change). druid_restoration is pending (still synthetic).

## Baseline (tools/audit/healer_rotations/baseline.sh, 1,000 iterations, -carry, own role, boss lengths)

Mean carry over 7 bosses, last pass -> now:

| request | last | now |
|---|---|---|
| priest_holy (tank) | 1.267 | 1.194 |
| priest_holy_raid | 1.448 | 1.440 |
| priest_discipline (tank) | 1.266 | 1.228 |
| paladin_holy (tank) | 1.151 | 1.100 |
| paladin_holy_raid | 1.392 | 1.343 |
| shaman_restoration (tank) | 1.276 | 1.201 |
| shaman_restoration_raid | 1.372 | 1.348 |
| druid_restoration (tank) | 1.255 | 1.255 (identical rows) |
| druid_restoration_raid | 1.370 | 1.351 |

41 of 63 rows moved more than 0.02. The cause is the gear change (the spec notes' Phase 0 against synthetic tables
match), not a rotation, so no finding. druid_restoration_raid moved -0.019 with no request change: the Frost mage
filler request changed (tools/dps_report/req/mage__frost.json; ruling healer_fairness-20261124-1). New baseline
written to ref/healer_bench/summary.tsv.

## Mana (180 and 300 s, Golemagg, Magmadar, Chromaggus)

- Holy priest tank: out of mana on Chromaggus at 106 s of 115 (was 114 s); Golemagg at 300 s 233 s. Open -20261125-1.
- Discipline tank: Chromaggus 102 s (was 95 s); Golemagg 300 s 189 s, Magmadar 300 s 245 s. Open -20261104-2.
- Holy paladin and Restoration shaman tank: no OOM inside any fight (the shaman's Chromaggus OOM at 128 to 136 s is
  past the fight's end).
- Raid healers end every fight with 6000 to 7900 mana and 27 to 70% overheal, as before (open gear findings on oils
  and flasks cover the spare mana).

No new findings; new evidence under -20261125-1 and -20261104-2.

---

## Master's pass, carried over from origin/master: Healer rotations pass 29 (quiet: no new finding; update on -20261124-1, -20261125-2 closed)

Merged in from master's audit loop (synthetic-gear tank/healer rows; numbers there are master's baseline, not the
Phase 0 one above). Kept for the verdicts and finding notes that exist only on master.

- Master: ab3fff700 (`last.master` was bfbee78fc; the ref was last written on 85691655a). Audited 2026-11-26.
- Bench: `tools/healbench` built on master (protos via `protoc -I=./proto --go_out=./sim/core ./proto/*.proto`
  with `/root/go/bin` on PATH), default seed, `-carry`, 1,000 iterations, via
  `tools/audit/healer_rotations/baseline.sh` (7 min). Ref `ref/healer_bench/summary.tsv` refreshed.

### What changed since bfbee78fc

- The one watched file, `sim/core/spell_result.go` (1223e764e), adds a branch inside the
  `sim.Encounter.leveling != nil` block (swings from behind on the leveling player). Healbench never sets
  a leveling session, so it is inert there; confirmed by healbench built at bfbee78fc, 143bf4f72, 8f5b59325,
  1223e764e, 81a2e51fc and 270731faa all giving shaman_restoration_raid broodlord 0.881 / 13.5%, same as master.
- No healer APL, healer request, encounter file, heal, healer talent or healbench main.go change.
- Between the ref's master (85691655a) and bfbee78fc the filler `tools/dps_report/req/mage__frost.json`
  changed (7829e00f1/14de5d62d, Phase 0 gold line: items 22339 -> 272406, 22408 -> 15282 suffix 1958).

### Baseline (1,000 it) vs ref

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

### Rotation lines

No rotation changed since pass 28, which tried the Holy priest tank mana gates (-20261125-1) and found none
that wins; not re-run. Open items carried: -20261125-1 (Holy priest Chromaggus OOM, worker fix),
-20261124-1 (filler pin), -20261118-1 (paired confirmation of 7199c8f9c), -20261115-1 (Wild Growth gate x0.75).

### Findings

- None new. -20261124-1: update (third filler-driven move). -20261125-2: closed (recorded, undisputed).
