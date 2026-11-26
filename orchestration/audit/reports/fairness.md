# Fairness pass 64 (2026-11-26, master f9fa79917)

Quiet pass. Range 2af970d74..f9fa79917 (53 watched files) is the tank and healer Phase 0 work: their gear sets,
lists, requests and prices, standards 1b / 8a / 35 / 36 / 43 updated for tanks and healers, the Bear Lacerate
stack fix and Water Shield's 4th globe with Totem of Sustaining. No DPS request, DPS preset, APL or list changed.
No full race sweep: nothing in range touches a DPS build, racials or the raid setup, so the 2af970d74 sweep stays
the baseline.

## What could reach a DPS build, checked

| Change | DPS effect |
|---|---|
| sim/druid/bear_attacks.go (Lacerate hits scale with stacks after the cast, EV-0040) | Bear form only. Feral cat and Balance within 0.5 of published in the loop's run; 6763bbb8d re-simmed the druid DPS builds, unchanged |
| sim/shaman water_shield.go, items.go (Totem of Sustaining +1 globe) | No DPS request or gear set wears 23200 (grep of tools/dps_report/req, ui/*/gear_sets) |
| standards.md 1b, 8a, 29, 35, 36, 43 | Text for tanks and healers only (UPDATED/SUPERSEDED notes, tank and healer gold line) |
| phase0_content.json (254 new tank/healer items, tank/healer goldPerDps) | Two DPS-visible tag changes: Frozen Steel Vambraces doubtful -> out (Ret wrist rank 14, loss 17.4, not a pick); Lesser Eternal Wand untagged -> doubtful (warlock ranged rank 17, loss 3.4, not a pick; doubtful stays in by standards 42). No pick moves |
| phase0_enchants.json | Tank and healer enchants only |

## Published against current

The loop's 20k run (/tmp/audit-work/runs/vs_published.txt): 14 of 68 builds move more than 0.5, the largest Fury 2H
duty-sunder +1.4 (0.15%), all inside noise. The published ranking still matches what the requests sim.

## Open findings

All seven open findings stand unchanged: no Ret, Arcane, Fury DW, Demonology, Destruction, MM, Balance, Fire, Shadow
request or list changed in range, no stun or explosive code changed.

## Baseline

Unchanged: 2af970d74 (sweep_baseline.log, table_baseline.md).
