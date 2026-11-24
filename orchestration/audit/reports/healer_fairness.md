# Healer fairness pass 28 (2026-11-24, master eceb09019)

Result: findings. New: -20261124-1 (high, the raid role's filler mage moved onto Phase 0 gear and reordered the raid
row). Still open: -20261115-1 (high, raid druid not re-tuned for the 0.066 Wild Growth), -20261104-4 (low).

Due because `sim/core/mana.go` changed (ed1df1d51, EV-0114: Innervate and Read Leyline's Energized now scale the MP5
share that isn't a periodic energize, i.e. gear MP5 but not Blessing of Wisdom or Mana Spring). Among the healer
requests only the two druid presets cast Innervate (self) and no preset is High Order (Read Leyline). Paired
(seed 101, 1,000 it) against eceb09019 with ed1df1d51 reverted (scratch worktree): the 14 druid cells are identical
but the tank druid's Chromaggus, 1.027 -> 1.037 (mean 1.267 -> 1.268). Negligible; no finding.

Pass 27 kept pass 21's cells without a re-bench; the reverted build's raid druid differed from them (1.385 -> 1.372),
so this pass re-benched all 63 cells at eceb09019 (`ref/healer_fairness/bench_eceb090.tsv`; healbench built with
`go build -tags with_db`, `tools/audit/healer_fairness/screen.sh`, seed 101, 1,000 it).

| Preset (role) | Onyxia | Brood | Chrom | Golemagg | Magmadar | Ragnaros | Vael | Mean (pass 21) |
|---|---|---|---|---|---|---|---|---|
| shaman_restoration (tank) | 1.391 | 0.944 | 1.010 | 1.984 | 2.111 | 0.598 | 0.922 | 1.280 (1.280) |
| priest_holy (tank) | 1.378 | 0.931 | 0.970 | 2.000 | 2.154 | 0.600 | 0.910 | 1.278 (1.278) |
| priest_discipline (tank) | 1.363 | 0.958 | 0.871 | 2.027 | 2.143 | 0.602 | 0.938 | 1.272 (1.272) |
| druid_restoration (tank) | 1.401 | 0.898 | 1.037 | 1.952 | 2.025 | 0.638 | 0.928 | 1.268 (1.267) |
| paladin_holy (tank) | 1.241 | 0.883 | 0.908 | 1.806 | 1.867 | 0.568 | 0.789 | 1.152 (1.152) |
| priest_holy_raid (raid) | 0.980 | 0.924 | 1.057 | 2.512 | 2.014 | 1.339 | 1.197 | 1.432 (1.472) |
| paladin_holy_raid (raid) | 0.977 | 0.904 | 1.149 | 2.527 | 1.877 | 1.248 | 1.027 | 1.387 (1.418) |
| druid_restoration_raid (raid) | 0.988 | 0.906 | 1.118 | 2.527 | 1.828 | 1.263 | 0.971 | 1.372 (1.385) |
| shaman_restoration_raid (raid) | 0.889 | 0.884 | 1.062 | 2.495 | 1.778 | 1.202 | 1.096 | 1.344 (1.390) |

Tank row: pass 21's but for the druid's Chromaggus. Raid row: every preset 0.013 to 0.046 lower, and the shaman drops
from second-to-last to last. Cause: 058a07457 put `tools/dps_report/req/mage__frost.json`, healbench's default filler
(7 copies in the raid role), on Phase 0 items instead of the synthetic set. With `-filler` set to a6677362f's version of
that file, the same binary gives Holy 1.473, paladin 1.418, shaman 1.395, druid 1.390 (`ref/healer_fairness/
raid_oldfiller_eceb090.tsv`), pass 21's row within 0.005. The shift by file runs +0.042 (druid Onyxia) to -0.152
(paladin Golemagg). So the raid ranking depends on the DPS report's mage gear, which nobody tunes for the healing
bench: finding -20261124-1 asks for healbench's own filler or a ruling that it tracks the report.

Same terms (step 1): the 9 healer requests are unchanged since pass 11, so everything but the healer row is still
identical. Logs (step 2): still blocked by -20261104-4. Roles (step 3): unchanged.

Open item: why the filler's gear moves carry this much (health, armor, fire resistance on Golemagg/Vaelastrasz?) was
not measured.

The rest is pass 21's (master a6677362f), kept for background:

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
