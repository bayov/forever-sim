# Healer fairness pass 17 (2026-11-17, master 5f20dfdf4)

Result: quiet. No new findings. Still open: -20261115-1 (high, raid druid not re-tuned for the 0.066 Wild Growth),
-20261104-4 (low).

Due because master changed 1 watched file since 1a85df1d7: `sim/shaman/item_sets_pve.go`. What changed in
1a85df1d7..5f20dfdf4 that could reach a healer:
- 6875e3b1e: Forever's Five Thunders set (1679, items 227001-227024) with its 2/3/4/6 piece bonuses. No healer
  request or healer UI preset wears any of those items (healer rows use the synthetic set, standards 35), so it
  can't reach a healer.
- Warrior, hunter and rogue mechanics, the Manual Crowd Pummeler cooldown, rotopt gear search, Enhancement BiS: no healer.
- `tools/healbench/`, `tools/encounters/`, `sim/core`, `proto/`, `ui/healing_sim`, druid/priest/paladin sim code: unchanged.

## 1. Same terms

Requests unchanged, so pass 11's check holds: everything but the healer row is identical in all 9 requests.

## 2. The numbers (healbench -carry, 1,000 iterations, seed 101, each file at its own length)

Healbench built from 5f20dfdf4 (`go build -tags with_db ./tools/healbench`; without the tag the item DB is empty
and the synthetic items panic), each of the 63 (request, role, boss) cells of `ref/healer_fairness/bench_25ce81d.tsv`
re-run: carry and EHPS identical in all 63, so no new ref file. Table (unchanged from pass 15):

| Request | Onyxia | Brood | Chrom | Golem | Magma | Rag | Vael | Mean | Rank |
|---|---|---|---|---|---|---|---|---|---|
| priest_holy (tank) | 1.378 | .931 | .970 | 2.000 | 2.154 | .600 | .910 | 1.278 | 1 |
| priest_discipline (tank) | 1.363 | .958 | .871 | 2.027 | 2.143 | .602 | .938 | 1.272 | 2 |
| druid_restoration (tank) | 1.403 | .893 | 1.023 | 1.951 | 2.045 | .634 | .932 | 1.269 | 3 |
| shaman_restoration (tank) | 1.340 | .978 | .984 | 1.938 | 2.062 | .587 | .911 | 1.257 | 4 |
| paladin_holy (tank) | 1.241 | .883 | .908 | 1.806 | 1.867 | .568 | .789 | 1.152 | 5 |
| priest_holy_raid | .988 | .906 | 1.065 | 2.607 | 2.062 | 1.340 | 1.262 | 1.461 | 1 |
| paladin_holy_raid | 1.018 | .912 | 1.146 | 2.702 | 1.914 | 1.283 | 1.035 | 1.430 | 2 |
| shaman_restoration_raid | .946 | .885 | 1.115 | 2.674 | 1.887 | 1.236 | 1.138 | 1.412 | 3 |
| druid_restoration_raid | .944 | .911 | 1.107 | 2.634 | 1.839 | 1.264 | 1.048 | 1.392 | 4 (was 3, 1.414) |

Spread: tank 1.152-1.278, raid 1.392-1.461 (4.7%). Nothing past the 15% line. The raid druid's fall to last
is -20261115-1, still open (no change to `sim/druid` or its preset on master).

## 3. What a role rewards

Each preset benched only in its own role, each file at its own length. Unchanged.

## Against the logs

Still can't be checked: the boss files keep no healing by class (-20261104-4).

## Open items

- -20261115-1 (high): re-tune the raid druid for the 0.066 Wild Growth; try Wild Growth in the tank preset.
- -20261104-4 (low): HPS by healer class in the boss files.
