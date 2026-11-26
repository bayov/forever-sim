# healer_rotations pass 30

master f9fa79917 (last: 3e9748641). Quiet: no new findings.

## What changed

The 8 watched files: sim/core/spell_result.go (leveling sessions only: hits from behind), sim/shaman/items.go and
water_shield.go (Totem of Sustaining 23200 gives Water Shield a 4th globe), and five healer `bis/*.bis.json` lists
(price and source notes only, commit 266e5357f; no pick changed). No healer APL, healbench request, encounter file,
healbench code or filler request (tools/dps_report/req/mage__frost.json) changed. Other core changes in the range
(spell_outcome.go stun/behind checks, environment.go, unit.go, Cannibalize registration) are leveling-gated or
inert on the bench.

## Baseline (tools/audit/healer_rotations/baseline.sh, 1,000 iterations, -carry, own role, boss lengths)

All 63 rows are byte-identical to ref/healer_bench/summary.tsv (cmp), so the means of pass 29 stand:

| request | mean carry |
|---|---|
| priest_holy (tank) | 1.194 |
| priest_holy_raid | 1.440 |
| priest_discipline (tank) | 1.228 |
| paladin_holy (tank) | 1.100 |
| paladin_holy_raid | 1.343 |
| shaman_restoration (tank) | 1.201 |
| shaman_restoration_raid | 1.348 |
| druid_restoration (tank) | 1.255 |
| druid_restoration_raid | 1.351 |

Mana: unchanged from pass 29 (Holy and Discipline tank still OOM on Chromaggus, open -20261125-1, -20261104-2).

## Changes tried

Totem of Sustaining (23200, now 4 Water Shield globes) in place of Totem of Urgency (279249) in the shaman requests,
paired seeds, 1,000 iterations, carry Urgency -> Sustaining:

| row | Urgency | Sustaining | delta |
|---|---|---|---|
| tank Onyxia | 1.328 | 1.371 | +0.043 |
| tank Broodlord | 0.927 | 0.950 | +0.023 |
| tank Chromaggus | 0.924 | 0.918 | -0.006 |
| tank Golemagg | 1.831 | 1.801 | -0.030 |
| tank Vaelastrasz | 0.859 | 0.849 | -0.010 |
| tank Ragnaros | 0.558 | 0.553 | -0.005 |
| tank Magmadar | 1.980 | 1.969 | -0.011 |
| raid Magmadar | 1.750 | 1.743 | -0.007 |
| raid Ragnaros | 1.210 | 1.205 | -0.005 |
| tank Onyxia 300 s | 0.785 | 0.782 | -0.003 |
| tank Golemagg 300 s | 0.958 | 0.948 | -0.010 |

Tank mean +0.001 over 7 bosses, raid and 300 s rows lose: Urgency stays, no finding. Sustaining ends Chromaggus
with less mana (731 against 866), so the extra globe doesn't pay for Urgency's Lesser Healing Wave crits.

No new findings. No open finding was fixed on master.
