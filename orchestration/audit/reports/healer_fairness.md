# Healer fairness pass 15 (2026-11-15, master 25ce81de7)

Result: findings. New: -20261115-1 (high; Wild Growth's doubled tick lowers the raid druid's carry 1.414 -> 1.392,
from third to last in the raid role). Still open: -20261104-4 (low).

Due because master changed 9 watched files since 70ea588bb. What reaches a healer:
- `sim/druid/healing.go`: Wild Growth tick coefficient 0.033 -> 0.066 (client 1.60.1.70938). The raid druid preset
  casts Wild Growth; the tank preset doesn't.
- `sim/paladin/talents.go`, `judgement.go`, `soc.go`, `sor.go`, `seal_of_fury.go`, `consecration.go`: Two-Handed Weapon
  Specialization's Holy part and Judgement of Command's defense type. Damage only; the Holy paladin presets heal.
- `sim/shaman/windfury_weapon.go`, standards 17: Windfury Totem and Weapon melee rules. No healer melees.

## 1. Same terms

`tools/healbench/req/`, `tools/encounters/`, `ui/healing_*` and `proto/` unchanged in 70ea588bb..25ce81de7, so pass
11's check holds: everything but the healer row is identical in all 9 requests; level 60, the same consumable
tier, Blessing of Kings and Improved Wisdom, 200 ms reaction, each class's best race. No unexplained difference.

## 2. The numbers (healbench -carry, 1,000 iterations, seed 101, each file at its own length)

Healbench built from 25ce81de7 (`--tags=with_db`); rows in `ref/healer_fairness/bench_25ce81d.tsv` (summary.tsv
format, `tools/audit/healer_rotations/baseline.sh`). 56 of 63 cells identical to `bench_71562a8.tsv`; only the raid
druid's 7 moved.

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

Spread: tank 1.152-1.278 (unchanged), raid 1.392-1.461 (4.7%, was 3.5%). Nothing past the 15% line.

### The raid druid and Wild Growth (-20261115-1)

Paired against a 0.033 build of the same master (scratch worktree, only the constant changed), carry:

| File | seed | 0.033 | 0.066 |
|---|---|---|---|
| Vaelastrasz | 101 / 202 / 303 | 1.152 / 1.158 / 1.145 | 1.048 / 1.039 / 1.031 |
| Onyxia | 101 / 202 / 303 | .991 / .985 / .984 | .944 / .947 / .987 |

A heal worth twice as much costing carry means the preset doesn't use it well. Its Wild Growth line waits until the
most hurt misses `spellExpectedHealing(Wild Growth)`, which doubled. Gate screen (seed 101, 1,000 it):

| Wild Growth gate | Onyx | Brood | Chrom | Golem | Magma | Rag | Vael | Mean |
|---|---|---|---|---|---|---|---|---|
| 1.0x (preset) | .944 | .911 | 1.107 | 2.634 | 1.839 | 1.264 | 1.048 | 1.392 |
| 0.5x (old threshold) | .972 | .921 | 1.132 | 2.458 | 1.793 | 1.271 | 1.210 | 1.394 |
| 0.25x | .930 | .910 | 1.103 | 2.344 | 1.671 | 1.256 | 1.174 | 1.341 |

More Wild Growth wins Vaelastrasz (raid-wide fire) and loses Golemagg and Magmadar, so it needs a proper rotation
re-tune (healer_rotations / rotopt), not one constant. Sent as a worker fix.

## 3. What a role rewards

Each preset benched only in its own role, each file at its own length. Unchanged.

## Against the logs

Still can't be checked: the boss files keep no healing by class (-20261104-4).

## Open items

- -20261115-1 (high): re-tune the raid druid for the 0.066 Wild Growth; try Wild Growth in the tank preset.
- -20261104-4 (low): HPS by healer class in the boss files.
