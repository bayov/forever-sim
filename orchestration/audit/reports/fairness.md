# Fairness pass 8 (2026-10-26, master f62cb79)

Result: quiet (no new findings). Full pass, baseline refreshed to f62cb79.

## Why it ran

Watched files changed since f796d99: 32a6848f9 (the seven wowsims/forever mechanics as one-value assumptions: Multi-Shot 0.5 sec cast, live spell DoT / channel ticks, hunter base RAP 110, failed Bite keeps energy and combo points, casting regen held to a paid channel's end) in sim/core (base_stats, cast, dot, mana, spell_result, unit), sim/druid/ferocious_bite.go, sim/hunter/multi_shot.go; standards 27 and 28 (rulings 2026-10-26). The full regeneration 758605874 re-simmed all 57 builds on that sim.

## What I checked

- Gear: `audit1b.py`, all 19 sets "5a PASS | 1b PASS".
- Race sweep: 36 base requests and variants, every race, 5000 iterations, seed 1, on the f62cb79 tree.
- Shared 20k run (`runs/master`) against the deployed `ui/dps_report/data/builds.json`: 0 of 57 builds differ by more than 0.5 DPS. The report is current.
- Read the shared mechanics in sim/core: live ticks (`dot.liveTick`) apply to every non-physical, non-helpful spell dot under Forever, and the channel regen hold (`mana.go holdFiveSecondRule`) to every mana-paid channel. Neither is gated by class, so no spec gets a mechanic another is denied.
- `tools/audit/fairness/table.py`: the Night Elf race-name fix pass 6 described was not in the file (the Feral row crashed again); restored. The gap column now reads the deployed builds.json (falling back to meta.published), because the req metas' `published.dps` is not refreshed by the regeneration.

## Ranking (report race, 5000 iterations)

Moves over 0.3 DPS since 74b3a5f; every other spec is within 0.1.

| # | Spec | DPS | Before (74b3a5f) |
|---|---|---|---|
| 1 | Beast Mastery | 1143.6 | 1141.4 |
| 2 | Marksmanship | 1136.3 | 1139.6 |
| 3 | Lone Wolf | 1123.9 | 1127.9 |
| 4 | Fire | 1100.4 | 1100.8 |
| 5 | Frostfire | 1097.6 | 1098.0 |
| 6 | Demonology | 1060.1 | 1064.9 |
| 7 | Affliction | 1030.4 | 1036.4 |
| 10 | Destruction Fire | 988.1 | 987.5 |
| 14 | Shadow | 956.8 | 961.7 |
| 20 | Balance | 852.5 | 856.7 |
| 21 | Elemental | 776.6 | 777.1 |
| 24 | Smite | 762.9 | 765.5 (was #23, swaps with Combat Swords 764.4) |

Variants: Balance Wrath -2.1, Shadow + Holy -4.7. Full table: `ref/fairness/table_baseline.md` and `orchestration/fairness_report.md` §1. Top over bottom 1.60x (was 1.59x).

## Conclusions

- Every move is attributed to 32a6848f9: the DoT and channel specs (Affliction, Demonology, Shadow, Balance, Smite's Holy Fire) lose to live ticks reading power and buffs at the tick, not the cast; the hunters move with the Multi-Shot cast and base RAP 110 (BM +2.2, MM -3.3, Lone Wolf -4.0, Survival +0.2). Feral moved +0.1 with the Bite change. Classes with no DoT or channel and no code change (warriors, rogues, Arcane, Ret, Enhancement, Frost) are unchanged to the decimal, so nothing leaked.
- Published against current (builds.json): every build within 1.9 at 5000 iterations (Ret +1.9, Frostfire +1.5 are the usual seed noise at this seed).
- The req metas' `meta.published.dps` is now 3 to 6 DPS stale for the 23 moved builds (e.g. Affliction 1035.3 against builds.json 1029.4), which is why `runs/vs_published.txt` shows those gaps. Same situation as pass 6 (the dashboard is current), so not a finding, but any tool comparing against meta.published will report false gaps until the metas are refreshed.
- Best race: Lone Wolf High Order ties the report's Windshaper (1123.9 each); Subtlety Orc vs Night Elf is the known 20k tie from pass 6. Keep all.
- results-20261025-1 is the results auditor's; nothing to add.
