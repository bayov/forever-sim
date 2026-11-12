# Fairness pass 36 (2026-11-12, master a0b4d3b)

Result: findings (1 new, 1 resolved). Watched files changed since 33db32b (23): Windfury Totem on main hand abilities
(a0b4d3be9, standards 17 text), Venom on Instant Poison (e2322baa1), defaults switched to each class's best 180 s build
(7462c17c8), priest tree size, shaman Trial by Fire rename, druid comment, compsearch drops. Full pass: `audit1b.py` and a
full 5k race sweep on the loop's a0b4d3b CLI and tree. New baseline: a0b4d3b.

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Windfury Totem rolls on white MH swings and on MH abilities whose client spell needs a MH weapon or comes on the next swing (generated list, logs of Nov 6-11) | a0b4d3be9 | Arms +171.0 (24th -> 11th), Fury 2H +105.8, Fury +50.5, Survival +70.8, rogues part of +21 to +58. Enhancement, Feral, Ret unchanged (Stormstrike, Claw/Rake, Holy Strike never roll it in logs). Raptor Strike now rolls like Heroic Strike / Cleave: fairness-20261109-1 resolved |
| Venom's +20% reaches Instant Poison | e2322baa1 | Rogues, with the totem: Combat Swords +46.5, SF Swords +57.6, Mutilate +38.1, SF Daggers +34.8, Subtlety +24.3, Combat Daggers +21.2 |
| Defaults: best 180 s build per class | 7462c17c8 | Combat Daggers chosen over Swords on 707ae92 numbers; the two commits above flip the order (fairness-20261112-1). Fury DW, Affliction, Lone Wolf still hold |
| Priest TalentTreeSizes, shaman Trial by Fire rename, druid comment | 33db32b..a0b4d3b | No move (Shadow, Smite, Elemental, Enhancement, Feral identical) |

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Movement against the 707ae92 baseline: only warriors, Survival and rogues moved, all from a0b4d3be9 / e2322baa1. No leak.
- The new rule is spell-ID based; Raptor Strike's hit spell (tag 1, SpellID 2973 / 14260-14266) is in the list, Mongoose
  Bite and Wing Clip too. Even-handed across warrior, hunter and rogue on-swing / MH abilities.
- Published against current: the 19 builds of results-20261112-1 (Fury -85.8, Fury 2H -119.0, Arms -46.3, Survival -53.4,
  rogues -0.4 to -18.3) plus older stale metas (hunters about -21, Feral -14.6, Affliction +37.0, Ret -83.2, Fire -4.1).
  Independent confirmation of the results auditor's findings, not refiled.
- Best race: report race is best within noise everywhere (Combat Daggers Gnome +0.6, Fire Gnome +0.4, MM Night Elf +0.3).

## Ranking now (report race, 5k)

Lone Wolf 1291.8, MM 1232.7, BM 1174.7, Feral 1124.4, Affliction 1071.1, Frostfire 1065.7, Destruction 1063.2,
Demonology 1055.3, Fire 1053.8, Fury DW 1050.2, Arms 1041.7, Fury 2H 1032.0, Enhancement 1030.7, Ret 1008.9,
Destruction Fire 997.9, Survival 987.8, Elemental 977.4, Arcane 963.6, Balance 952.1, Shadow 948.6, Combat Swords 946.1,
Frost 943.6, SF Swords 938.7, Combat Daggers 936.0, SF Daggers 904.4, Mutilate 865.7, Subtlety 863.6, Smite 781.0.
Top over bottom 1.65x. Full table: `orchestration/audit/ref/fairness/table_baseline.md`.

## Open for the next pass

- After the re-sim of results-20261112-1, check published against current for the 19 moved builds and the older stale metas.
- fairness-20261112-1: confirm the Combat Swords / Daggers order at 20k after the re-sim, and that the default follows it.
