# Healer fairness pass 24 (2026-11-22, master 8b5137c6d)

Result: quiet. No new findings. Still open: -20261115-1 (high, raid druid not re-tuned for the 0.066 Wild Growth;
last at 1.385, 0.005 behind the shaman), -20261104-4 (low).

Due because 10 watched files changed since b4b68c52e (net diff to 8b5137c6d):
- `sim/paladin/talents.go`, `judgement.go`, `consecration.go`, `soc.go`, `sor.go`, `seal_of_fury.go`: Two-Handed Weapon
  Specialization's Holy part moves to the Holy school under Forever (Nov 19 hotfix) and follows item swaps. It applies only
  with a two-handed main hand; neither Holy paladin preset carries one, and it is a damage multiplier, not healing.
- `sim/shaman/talents.go`, `stormstrike.go`, `item_sets_pve.go`, `sim/core/buffs.go`: comments only (Maelstrom Weapon PPM
  source, Dungeon Set 1 in set 1679, Windfury Totem client rows).
- `orchestration/standards.md` 41 (Blizzard's word overrides an earlier ruling): no open healer finding rests on a ruling it retires.
- No `tools/healbench/req/` or `tools/encounters/` change.
- Check: all 63 cells, both roles, seed 101, 1,000 it (`screen.sh`, healbench built on 8b5137c6d with `-tags with_db`
  in the repo, no worktree needed): identical to `ref/healer_fairness/bench_a667736.tsv`, so sections 1-3 below stand.

The rest is pass 21's (master a6677362f), still current:

## 1. Same terms

Requests unchanged, so pass 11's check holds: everything but the healer row is identical in all 9 requests.

## 2. The numbers (healbench -carry, 1,000 iterations, seed 101)

Healbench built from a6677362f (`go build -tags with_db`). All 63 cells re-run with
`tools/audit/healer_fairness/screen.sh` (`ref/healer_fairness/bench_a667736.tsv`; columns request, role, boss, carry,
EHPS, overheal). Against `bench_25ce81d.tsv`: tank paladin, Holy and Discipline priest identical (21 cells); the
other 42 moved, mostly within ±0.04, largest the raid shaman on Magmadar (1.887 -> 1.778) and the tank shaman on
Onyxia (1.340 -> 1.391). Those are the presets that idle between casts, which is where 7199c8f9c acts.

| Preset | Onyxia | Brood | Chrom | Golemagg | Magmadar | Ragnaros | Vael | Mean (was) |
|---|---|---|---|---|---|---|---|---|
| Shaman tank | 1.391 | 0.948 | 1.009 | 1.986 | 2.102 | 0.599 | 0.922 | 1.280 (1.257) |
| Holy tank | 1.378 | 0.931 | 0.970 | 2.000 | 2.154 | 0.600 | 0.910 | 1.278 (1.278) |
| Disc tank | 1.363 | 0.958 | 0.871 | 2.027 | 2.143 | 0.602 | 0.938 | 1.272 (1.272) |
| Druid tank | 1.401 | 0.898 | 1.032 | 1.951 | 2.025 | 0.637 | 0.928 | 1.267 (1.269) |
| Paladin tank | 1.241 | 0.883 | 0.908 | 1.806 | 1.867 | 0.568 | 0.789 | 1.152 (1.152) |
| Holy raid | 0.990 | 0.911 | 1.073 | 2.589 | 2.100 | 1.378 | 1.264 | 1.472 (1.461) |
| Paladin raid | 1.008 | 0.922 | 1.141 | 2.679 | 1.887 | 1.270 | 1.022 | 1.418 (1.430) |
| Shaman raid | 0.935 | 0.918 | 1.092 | 2.636 | 1.778 | 1.235 | 1.139 | 1.390 (1.412) |
| Druid raid | 0.946 | 0.873 | 1.102 | 2.625 | 1.857 | 1.250 | 1.042 | 1.385 (1.392) |

| Role | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Tank | Shaman 1.280 | Holy priest 1.278 | Disc priest 1.272 | Druid 1.267 | Paladin 1.152 |
| Raid | Holy priest 1.472 | Paladin 1.418 | Shaman 1.390 | Druid 1.385 | |

Spread: tank 1.152-1.280 (11%, the paladin; the top four within 1%), raid 1.385-1.472 (6.3%). Nothing past the 15%
line. The tank shaman's move to first is 0.002 over the Holy priest, within seed noise; the order of the top four
tank healers is not meaningful at one seed. The raid shaman lost 0.022 (Magmadar -0.109, Golemagg -0.038) and the
Holy priest gained 0.011 (Magmadar and Ragnaros +0.038): reacting on time changes which heal the idle presets pick.
Per file: tank Onyxia, Chromaggus and Ragnaros go to the druid, Broodlord, Golemagg and Vaelastrasz to the
Discipline priest, Magmadar to the Holy priest; raid Onyxia, Broodlord, Chromaggus, Golemagg to the paladin,
Magmadar, Ragnaros, Vaelastrasz to the Holy priest.

Not re-confirmed on a second seed this pass; the healer_rotations pass ran its own baseline on this master at the
same time and may want to re-check the raid shaman's Nature's Swiftness gate after the timing fix.

## 3. What a role rewards

Each preset benched only in its own role, each file at its own length. Unchanged.

## Against the logs

Still can't be checked: the boss files keep no healing by class (-20261104-4).

## Open items

- -20261115-1 (high): re-tune the raid druid for the 0.066 Wild Growth; try Wild Growth in the tank preset.
- -20261104-4 (low): HPS by healer class in the boss files.
- Carried: the raid shaman Totem of Tides relic screen (tank variant was -0.004; low priority).
- Housekeeping: this pass's scratch landed in `/healer_fairness/` (unset `$AUDIT_WORK` in one shell); a copy of the
  bench is in the ref file above, the directory can be deleted.
