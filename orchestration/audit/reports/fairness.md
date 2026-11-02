# Fairness pass 22 (2026-11-02, master fa865a3)

Result: quiet (no new finding). Light pass: no sim, database, gear or standards change.

## Why it ran

Since 5a6e9c6, 81 watched files: 231d5492a drops the deep Holy Shock Ret build (paladin__level_60_shock, its
-duty-crusader / -duty-wisdom runs, _undead and its Dwarf race file; user ruling) and makes Frostfire the mage default;
3fd9b46d2 refreshes every meta's published / verify number from the deployed 2026-11-02 report. `sim/` changed only in
`sim/compsearch` (drop table), `sim/raid_forever_test.go` and `report_hashes.txt`: no class or core code.

## What I checked

- Sim unchanged, so the 5a6e9c6 race sweep stands. The loop's master run (20k, seed 1) against the metas: 0 of 73 moved
  more than 0.5 (`runs/vs_published.txt`).
- `table.py` on the 5a6e9c6 sweep (Shock builds removed) against the new metas: every base build and variant within 1.0
  of its published number (largest: Frostfire -1.0, Demonology +1.0, Frost -0.9, Affliction +0.8). The 62-of-77 staleness
  (results-20261101-1) is fixed on master; I confirm it independently.
- Best race: unchanged; only Affliction (Orc +1.3 over Human) flags, the open fairness-20261101-2.
- Gear: no gear_sets / database change; `audit1b.py` not rerun (19 of 20, Demonology 5a, fairness-20261031-1 open).
- Dropping deep Holy Shock and switching the mage default change no shared assumption.

## Ranking (5k sweep, report race)

Lone Wolf 1087.8, Frostfire 1075.6, MM 1063.7, Fire 1056.7, Demonology 1046.5, BM 1045.2, Fury 1041.6, Fury 2H 1039.6,
Ret 1039.2, Survival 1024.8, Affliction 1012.9, Arms 1008.7, Destruction 995.0, Destruction Fire 977.8, Shadow 948.7,
Frost 944.0, Feral 942.4, Arcane 939.4, Balance 890.6, Enhancement 883.0, Combat Daggers 788.4, Combat Swords 776.5,
Elemental 769.0, Smite 761.2, Seal Fate Daggers 751.9, Mutilate 742.9, Seal Fate Swords 742.4, Subtlety 732.4.
Top over bottom 1.49x. Full table: `orchestration/fairness_report.md` §1 and `ref/fairness/table_baseline.md`.

## Open (unchanged)

fairness-20261101-2 (Affliction race), fairness-20261101-1 (caster range, user ruling), fairness-20261031-1 (Demonology set, user ruling).
