# Fairness pass 33 (2026-11-09, master 66a243c)

Result: findings (1 new: fairness-20261109-1). Watched files changed since 9ba00c2: `sim/core/buffs.go`
(Windfury Totem white-swing-only, 0ce9dbf40), `sim/core/base_stats_levels_auto_gen.go` (Orc shaman levels 2-43,
3b0983bc6, leveling only, no level 60 row changed), the item database (8051c4322), test results and report hashes.
Full pass: `audit1b.py`, a full 5k race sweep on the loop's 66a243c CLI and tree, a bisect of the Feral move, and two
measured options for the on-next-swing asymmetry. New baseline: 66a243c.

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Windfury Totem procs from white MH swings only under Forever (user ruling 2026-11-08, "for now") | 0ce9dbf40 | every raid-totem melee build drops: Arms -200.3, Fury 2H -189.0, Fury DW -97.6, Survival -94.6, Ret -94.2, Feral -27.0, rogues -19.6 to -46.5. Enhancement (own imbue) and all casters / ranged unchanged |
| Orc shaman base attributes, levels 2-43 | 3b0983bc6 | none at level 60 |
| Item database fold-in (331 added, 2 changed) | 8051c4322 | none: every non-melee build identical to 0.1; audit1b 19/19 PASS |

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Movement against the 9ba00c2 baseline: only melee builds that take the raid's Windfury Totem moved. Feral bisected
  (5k, seed 1): c1b90a5 1169.6, 0ce9dbf 1142.6, 66a243c 1142.6. So Feral's -27.0 is also 0ce9dbf40. Correction to
  results-20261109-1, which lists feral as "not from this commit": its published 1139.0 only looked close because it
  was already 30.6 behind. Nothing shared leaked beyond that commit.
- Published against current: independent confirmation of results-20261109-1 (same numbers to ±1 at 5k). The report now
  has the order wrong for 15 specs: Fury 2H published 5th, sims 16th; Arms 9th -> 24th; Fury DW 7th -> 10th;
  Ret 8th -> 12th; Survival 14th -> 19th. Feral's old +30.6 gap closes to +3.6. Elemental +7.3, Arcane +7.9 and
  Enhancement -11.6 are unchanged and still open under results-20261104-1/-2.
- Best race: the report race is best everywhere except Arms (Undead 888.2 vs Orc 887.6, +0.6, noise).
- On-next-swing attacks under the new rule (fairness-20261109-1): Heroic Strike / Cleave carry
  `MHSpecial | MHAuto` and still proc the totem. The Raptor Strike hit is `MHSpecial` only, so it no longer does.
  The commit comment says Heroic Strike is "left out", which the code does not do. Measured with an env-gated scratch
  hack (5k, seed 1): Raptor Strike procs -> Survival +24.9 / wolf +24.8 / tallstrider +25.5; Heroic Strike and Cleave
  don't proc -> Fury -16.6, Fury 2H -3.5, Arms -4.8.

## Ranking now (report race, 5k)

Lone Wolf 1314.4, MM 1254.6, BM 1195.7, Enhancement 1147.8, Feral 1142.6, Frostfire 1065.7, Destruction 1063.2,
Fire 1058.2, Demonology 1055.3, Fury DW 1038.1, Affliction 1034.2, Ret 1008.9, Destruction Fire 997.9, Elemental 977.4,
Arcane 970.2, Fury 2H 961.8, Balance 952.1, Shadow 948.6, Survival 945.7, Frost 943.6, Combat Daggers 933.3,
Combat Swords 918.0, SF Swords 899.7, Arms 887.6, SF Daggers 887.6, Subtlety 858.1, Mutilate 843.8, Smite 781.0.
Top over bottom 1.68x. Full table: `orchestration/audit/ref/fairness/table_baseline.md`.

## Open for the next pass

- After the re-sim of results-20261109-1, check that published matches current for all 15 moved specs.
- The rule rests on Stormstrike alone. Whether Mortal Strike, Sinister Strike, Crusader Strike or Shred procs the
  totem in game is the open question under results-20261109-1. Heroic Strike vs Raptor Strike is fairness-20261109-1.
