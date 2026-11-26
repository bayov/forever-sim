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
