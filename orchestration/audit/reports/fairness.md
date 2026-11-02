# Fairness pass 23 (2026-11-02, master 796495d)

Result: quiet (no new finding). Full pass: sim/core, class presets, requests, standards 33 / 34 and the gear sets changed.

## Why it ran

Since b647a21 (pass 22 audited fa865a3's content), 192 watched files: standards 33 (casters at the longest range their
rotation reaches, fa2d49997; weaving hunters searched, BM at 12 yd, 6c357819c), standards 34 (Demonology's own set,
be5126ccd), Balance Wrath as Night Elf (32c825cff), the warrior (f62b4a582), rogue (da5be5321) and MM / Lone Wolf
(bb8edc2fe) re-searches, and the re-sims (63c461bfc, 722ccf78b). `sim/core` changes are the tank / healing / practice-room
hooks; none moved a build whose request did not change.

## What I checked

- Race sweep (`sweep.py`, 5k, seed 1) on the 796495d tree with the loop's shared CLI; `table.py` against the fa865a3
  baseline. Table: `orchestration/fairness_report.md` §1 and `ref/fairness/table_baseline.md` (baseline replaced).
- Gear: `audit1b.py` 20 of 20 "5a PASS | 1b PASS" (Demonology now on its own set; fairness-20261031-1 confirmed fixed).
- Published against current: loop's 20k run, 0 of 73 moved more than 0.5 (`runs/vs_published.txt`). In the 5k sweep the
  largest gaps are Arcane +1.7 and Fury 2H +1.3, which the 20k run shows are noise. Confirms results' staleness findings
  stay fixed; results-20261102-1 (mana-tide 0.4) is below my tolerance.
- Movement: every build that moved over 0.5 changed its own request in a named commit (Demonology +7.4, Fury 2H +6.1,
  Combat Swords +5.6, Seal Fate Swords +4.9, Combat Daggers +3.2, Lone Wolf +2.7, MM +2.2, Balance +1.3, Wrath -1.8,
  Destruction Fire -1.1, BM +1.0). Classes with no request change did not move (Fury DW, Arms, Ret, Survival, Shadow,
  Smite, Enhancement, Elemental, Feral all +0.0).
- Distances (standards 33) are applied by one rule to every ranged build: Frostfire / Arcane / Elemental 20, Shadow 22,
  Fire / Frost / Smite / all warlocks 30, Balance 36, MM / Lone Wolf 25, BM 12 (searched). Evenly applied.
- Best race: Fire 5k Gnome +1.5 over Human, 20k only +0.7 (Gnome 1057.8 / Human 1057.1): noise-level, no note.
  Affliction Orc +0.9 at 20k (was +1.4): evidence added to fairness-20261101-2.

## Ranking (5k sweep, report race)

Lone Wolf 1090.5, Frostfire 1076.2, MM 1065.9, Fire 1056.4, Demonology 1053.9, BM 1046.2, Fury 2H 1045.7, Fury 1041.6,
Ret 1039.2, Survival 1024.8, Affliction 1013.1, Arms 1008.7, Destruction 994.9, Destruction Fire 976.7, Shadow 948.7,
Frost 943.6, Feral 942.4, Arcane 939.2, Balance 891.9, Enhancement 883.0, Combat Daggers 791.6, Combat Swords 782.1,
Elemental 769.0, Smite 761.2, Seal Fate Daggers 752.9, Seal Fate Swords 747.3, Mutilate 743.6, Subtlety 733.3.
Top over bottom 1.49x.

## Open

fairness-20261101-2 (Affliction race, now +0.9). Master moved past 796495d to 8153736 during the pass (mage Ice Lance
shares Fire Blast's cooldown, d7541708d, and a four-mage re-sim): not audited here; the next pass covers it.
