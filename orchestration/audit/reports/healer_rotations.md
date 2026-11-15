# Healer rotations pass 14 (findings: healer_rotations-20261115-1; -20261104-2 still open)

- Master: 25ce81de7 (`last.master` was 70ea588bb). Audited 2026-11-15.
- Bench: `tools/healbench` built from this checkout (Go protos generated with `protoc -I=./proto --go_out=./sim/core`
  first), default seed, `-carry`, 1,000 iterations, every request on each of the 7 boss files at its own length in
  its role, via `tools/audit/healer_rotations/baseline.sh` (now resumable: it keeps the rows already in its output,
  since a full run took about an hour on a loaded 4-core box; `REQS=` limits it to some requests).

## What changed since 70ea588bb

- `sim/druid/healing.go`: Wild Growth's tick coefficient 0.033 -> 0.066 (client 1.60.1.70938). Used only by the
  Restoration druid raid healer.
- `sim/paladin/*` (Two-Handed Weapon Specialization's Holy part by defense type, Judgement of Command on melee
  defense, seal changes) and `sim/shaman/windfury_weapon.go`: damage side only. The Holy paladin and Restoration
  shaman rows are bit-identical.
- No healer APL, healbench request, encounter file or `tools/healbench/main.go` changed.

## Baseline

56 of 63 rows are identical to the old reference. The 7 `druid_restoration_raid` rows moved (carry, ref -> now):

| boss | carry | other died | mana left |
|---|---|---|---|
| onyxia | 0.991 -> 0.944 | 5.6% -> 7.1% | 6150 -> 6314 |
| broodlord | 0.924 -> 0.911 | 11.3% -> 10.4% | 6688 -> 6733 |
| chromaggus | 1.094 -> 1.107 | 1.7% -> 1.6% | 7131 -> 7202 |
| golemagg | 2.630 -> 2.634 | 0.0% | 7260 -> 7271 |
| magmadar | 1.829 -> 1.839 | 0.0% | 6702 -> 6816 |
| ragnaros | 1.278 -> 1.264 | 0.3% -> 0.7% | 5418 -> 5842 |
| vaelastrasz | 1.152 -> 1.048 | 2.8% -> 3.3% | 6070 -> 6066 |

Mean -0.022 carry: a heal twice as strong carries less. `ref/healer_bench/summary.tsv` is rewritten with the new rows.

## Changes tried (druid_restoration_raid, Wild Growth line's gate, paired, 1,000 iterations)

The gate is `healthMissing(MostHurt) >= spellExpectedHealing(Wild Growth)`, which doubled with the coefficient.

| boss | now | x0.5 | x0.75 | no gate |
|---|---|---|---|---|
| onyxia | 0.944 | 0.972 | 0.964 | 0.917 |
| broodlord | 0.911 | 0.921 | 0.924 | 0.899 |
| chromaggus | 1.107 | 1.132 | 1.115 | 1.096 |
| golemagg | 2.634 | 2.458 | 2.603 | 2.438 |
| magmadar | 1.839 | 1.793 | 1.820 | 1.734 |
| ragnaros | 1.264 | 1.271 | 1.266 | 1.247 |
| vaelastrasz | 1.048 | 1.210 | 1.181 | 1.177 |
| mean vs now | | +0.001 | +0.018 | -0.034 |

x0.75 wins on 5 of 7 bosses and brings Onyxia's other deaths back to 5.6%. Mana is not binding (over 5,000 left
everywhere), so the 180/300 s runs were skipped. Filed as healer_rotations-20261115-1. Other rotation lines were not
re-screened: nothing else they depend on moved.

## Open findings

- healer_rotations-20261115-1 (new): see above. Needs a worker fix.
- healer_rotations-20261104-2 (Discipline tank healer OOM on Chromaggus): row unchanged. Needs a worker fix.
