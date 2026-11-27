# Fairness pass 65 (2026-11-27, master 34994816e)

Findings pass. Range f9fa79917..34994816e (105 watched files): the Armaments of Battle rewards out of Phase 0
(287da91da, 6bd2d60a2: lists, nine DPS gear sets, warlock Bane of Doom hold), Chained Essence of Eranikus cloud
ticks at a fixed 2.7% crit (5591bd019, sim/common/item_effects.go), the Enhancement trinket re-rank (3698a73d7),
and the Shockadin presets added and dropped again (27a8031dc..d70972fa3). No full race sweep: no racial, raid
or sim/core change, so the 2af970d74 sweep stays the baseline.

## Published against current

The loop's 20k run (/tmp/audit-work/runs/vs_published.txt): every build within 1.5 of published (largest Feral and
Lone Wolf -1.5, the Chained Essence crit fix on its wearers). The requests did not change, so this only says the
published numbers match the requests, not the gear sets.

## Gear sets against requests (new: tools/audit/fairness/gearset.py)

phase0_check.py flags 37 requests whose equipment is no longer their gear set. Paired 10k, seeds 1 / 1000003,
request with only its equipment set to the current gear set:

| Build | Request | Gear set | Delta | Swap |
|---|---|---|---|---|
| Affliction | 962.8 | 945.6 | -17.2 | Talisman of Ascendance -> Chained Essence |
| Affliction duties / Shadow Bolt | | | -17.7 / -18.4 / -10.2 | same |
| Demonology (duties) | 876.8 | 867.2 | -9.6 (-10.5 / -9.3) | same |
| Destruction (duties) | 868.4 | 866.8 | -1.5 (-2.5 / -2.6) | Talisman -> 272438 |
| Destruction Fire (duties) | 825.6 | 824.9 | -0.7 (-0.7 / -1.2) | same |
| Balance / FF / Wrath / Wrath FF | 801.8 | 796.2 | -5.7 / -5.1 / -2.7 / -1.9 | Talisman -> Chained Essence |
| Feral (FF duty) | 1057.5 | 1054.1 | -3.3 (-3.4) | Bracers of Subterfuge -> 10256 |
| Fury 2H (duties) | 994.5 | 991.7 | -2.8 (-2.6 / -2.6) | same |
| Elemental (mana-tide) | 856.2 | 857.9 | +1.6 (+0.8) | Talisman, Chained Essence -> 13968, 18471 |
| Smite | 678.7 | 678.7 | -0.0 | Verimonde's -> 272083 |
| Enhancement (imp-gw, mana-tide, water-shield) | 966.4 | 972.5 | +6.1 (+6.5 / +6.0 / +5.4) | Chained Essence -> Maelstrom |

Rank effect: Affliction falls from 15th to below Survival and Enhancement imp-gw; Demonology falls below
Destruction and Enhancement mana-tide. Filed as fairness-20261127-1 (high, orchestrator fix).

## Chained Essence crit fix on other wearers

Rogues, Feral and Survival already wear Maelstrom beside Chained Essence. Ret and Shock Ret wear Chained Essence
without Maelstrom: Shock Ret +2.9 with Maelstrom (paired 10k), added to fairness-20261126-4.

## Open findings

fairness-20261126-4 (Ret Maelstrom): new evidence above. The other six (Shahram stun, Arcane / Fury race,
explosives, Demonology race, MM race, Fire race, Balance race) stand: no request changed in range.

## Baseline

Kept at 2af970d74 (sweep_baseline.log, table_baseline.md). Take a new one after fairness-20261127-1's re-sim.
