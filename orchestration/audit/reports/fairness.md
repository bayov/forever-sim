# Fairness pass 25 (2026-11-03, master f5c99ad)

Result: quiet (no new finding). Full pass, forced by round 70753 phase 3: all nine classes, `sim/core` (race base
attributes, racials, Throw, potions), the requests and standards 37 changed (245 watched files since 9da90fc).

## What I checked

- Race sweep (`sweep.py`, 5k, seed 1) with the loop's shared CLI and tree (`/tmp/audit-cache/cli/f5c99ad018aa`);
  `table.py` against the a0b3069 baseline. Table: `orchestration/fairness_report.md` §1 and
  `ref/fairness/table_baseline.md` (baseline replaced, sha in `state/fairness.json`).
- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Published against current: loop's 20k run (`runs/vs_published.txt`), 62 of 68 within 0.5; the six that moved are the
  shaman metas, stale by design until the regeneration. My 5k gaps on the base builds: at most Frostfire +1.4,
  Demonology +1.4, Frost -1.1 (noise).
- Movement since a0b3069, each traced to its class's round file (`orchestration/round_70753/<class>.md`; their final
  numbers match the sweep): hunters +16 to +224 (hotfix and notes shares), Feral +189 and Balance +59 (notes), rogues
  +123 to +210 (rework), warriors +89 to +125 (rework), Ret +52 (Vengeance 5, Holy Two-Handed Specialization),
  Destruction +66 and Destruction Fire +20 (Shadow and Flame), Enhancement +277 and Elemental +201 (notes), Arcane +23
  (talent re-search). Fire, Frost, Frostfire, Shadow, Smite, Affliction, Demonology within 0.7: the shared `sim/core`
  changes did not leak into builds whose class had no damage change.
- Best race after Classic's race base attributes (7598b8015): the report race is the best everywhere within noise
  (Combat Swords Undead 0.5 under Orc; MM tallstrider / wolf variants Orc 0.8 / 0.5 under Night Elf).
- Shared raid assumptions (`/tmp` script grouping the 45 base requests' raid buffs, debuffs, blessings): differences
  are by role only (casters without melee totems, own-class buffs dropped, warriors' own pull Sunder per standard 32).
  Feral is the one mana user without Mana Spring: measured 0.0 DPS (1131.1 both ways, 20k, seed 1). Salvation still
  present everywhere (phase 3 to-do, not a finding).
- Notes-only values (standards 37): checked the round files' lists against the notes' text for druid (Shred 200, Claw
  140, Ravage 435, bleeds x1.25, Starfire x1.40, Rend and Tear 3%), hunter (0.15, 5%), paladin (Vengeance 5, Two-Handed
  2/4/6% Holy and Physical), rogue (poisons 0.009 / 0.01, Mutilate 90%, Flawless Execution 15, Aggression 5/10/15),
  shaman (Lightning Overload, Maelstrom x2, Stormstrike 150% / 6 s, Mental Dexterity, Flurry), warlock (Wrack 0.286 a
  1 s tick, 86 at 60; Shadow and Flame 4%) and warrior (Slam 15 s, Improved Slam, Tactical Mastery, Bloodthrill,
  Flurry, Reckless Swings, Improved Whirlwind). All match; druid's are gated on `Env.IsForever()`. The hunter shares'
  missing gate is already results-20261103-2.

## Results auditor overlap

- results-20261103-1 (Subtlety meta race "Night Elf"): confirmed; it crashed `table.py`, which now strips the space.

## Ranking (5k sweep, report race)

Lone Wolf 1314.4, MM 1254.6, BM 1195.7, Fury 2H 1170.7, Enhancement 1159.5, Fury 1153.6, Feral 1131.0, Arms 1097.9,
Ret 1090.7, Frostfire 1065.7, Destruction 1063.2, Fire 1058.2, Demonology 1055.3, Survival 1041.1, Affliction 1015.8,
Destruction Fire 997.1, Elemental 970.4, Combat Swords 963.3, Arcane 962.3, Seal Fate Swords 957.7, Balance 950.9,
Shadow 948.6, Combat Daggers 947.3, Frost 943.6, Seal Fate Daggers 921.3, Subtlety 881.2, Mutilate 866.4, Smite 761.9.
Top over bottom 1.73x (was 1.49x): the notes lifted every melee and hunter class while mage, priest and Affliction /
Demonology stayed put.

## Open

None of mine. The pass doc's long-standing open items (Ice Lance Low range, FTT stacking, boss partial resist, Rage k)
are unchanged; the Skyborne base stats stay a round README open question.
