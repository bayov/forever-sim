# Fairness pass 34 (2026-11-10, master ca5e35c)

Result: quiet (no new finding). Watched files changed since 91ad195: `sim/shaman/talents.go` and
`sim/shaman/flametongue_weapon.go` (bcc5d6cd4), `tools/database/forever_enchants.go` and the item database
(04bdcdb9e, a778e524b), test results and report hashes. Full pass: `audit1b.py` and a full 5k race sweep on the loop's
ca5e35c CLI and tree. New baseline: ca5e35c.

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Maelstrom Weapon stacks on every landed melee hit under Forever (was 4 PPM a point); Totem of the Storm 50% a bolt | bcc5d6cd4 | Enhancement only: level60 -0.3, imp-gw -2.4, water-shield -4.3, mana-tide +0.3 (5k, seed 1). At 5/5 with a 3.4 sec two-hander 20 PPM was already a 100% chance a hit, so the rule barely moves DPS; the rest is the random stream |
| Flametongue Weapon rank 5 cut by 7 | bcc5d6cd4 | none at 60 (rank 6 keeps the full points) |
| Enchants 41 / 66 give +2 / +3 Stamina | 04bdcdb9e | none (Stamina only; only Enhancement's hashes moved) |
| Item database fold-in (67 added, 3 green variants to base stats) | a778e524b | none: every non-shaman build identical to 0.1 |

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Movement against the 66a243c baseline: only the four Enhancement builds moved, and only that class's code changed.
  No shared change leaked. Elemental identical.
- Published against current: no new gap. The 15 Windfury Totem gaps (results-20261109-1) and the older Enhancement
  -11.9, Feral +3.6, Elemental +7.3, Arcane +7.9 (results-20261104-1/-2) are unchanged; the shared run's
  `vs_published.txt` agrees to ±1. Independent confirmation only, not refiled.
- Best race: the report race is best everywhere except Arms (Undead 888.2 vs Orc 887.6, noise). Enhancement Orc
  still leads Windshaper by 4.7 to 5.5 except water-shield (+1.4).
- fairness-20261109-1 (Heroic Strike / Cleave proc Windfury Totem, Raptor Strike does not) is still open: no change to
  `sim/core/buffs.go`, `sim/warrior` or `sim/hunter`.

## Ranking now (report race, 5k)

Lone Wolf 1314.4, MM 1254.6, BM 1195.7, Enhancement 1147.5, Feral 1142.6, Frostfire 1065.7, Destruction 1063.2,
Fire 1058.2, Demonology 1055.3, Fury DW 1038.1, Affliction 1034.2, Ret 1008.9, Destruction Fire 997.9, Elemental 977.4,
Arcane 970.2, Fury 2H 961.8, Balance 952.1, Shadow 948.6, Survival 945.7, Frost 943.6, Combat Daggers 933.3,
Combat Swords 918.0, SF Swords 899.7, Arms 887.6, SF Daggers 887.6, Subtlety 858.1, Mutilate 843.8, Smite 781.0.
Top over bottom 1.68x. Full table: `orchestration/audit/ref/fairness/table_baseline.md`.

## Open for the next pass

- After the re-sim of results-20261109-1, check that published matches current for all 15 moved specs.
- The ruling on fairness-20261109-1 (Option A: Survival +25; Option B: Fury -16.6, Fury 2H -3.5, Arms -4.8).
