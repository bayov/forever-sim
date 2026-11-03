# Fairness pass 24 (2026-11-03, master a0b3069)

Result: quiet (no new finding). Full pass: sim/core, sim/mage, sim/warrior, warlock and hunter requests and standards.md changed.

## Why it ran

Since 796495d (pass 23), 81 watched files: Ice Lance shares Fire Blast's category 19 cooldown and Fingers of Frost's
buff id (d7541708d, user ruling 2026-11-02) with the four-mage re-sim (5445fc92a); the warlocks hold the second Bane of
Doom so its tick lands in the 2 min cooldowns (58a1c3593, re-sim 3fce3491c); the ranged hunters drop the Goblin Sapper
and Affliction is an Orc (b33ad307d); standards 26 records the Balance Subtlety ruling (wording only); Blood Craze for
warriors (407f959a5); `sim/core` heal-rank cap, Expected Healing APL value, raid damage shares and tank score (healing
and tank sims); the Restoration shaman spec.

## What I checked

- Race sweep (`sweep.py`, 5k, seed 1) on the a0b3069 tree; `table.py` against the 796495d baseline. Table:
  `orchestration/fairness_report.md` §1 and `ref/fairness/table_baseline.md` (baseline replaced).
- Gear: `audit1b.py` 20 of 20 "5a PASS | 1b PASS".
- Published against current: loop's 20k run, 0 of 73 moved more than 0.5 (`runs/vs_published.txt`). 5k gaps at most
  Arcane +1.7, Frostfire +1.4, Fury 2H +1.3: noise. results-20261102-1 (mana-tide meta 0.4) is below my tolerance.
- Movement, each traced to its own class's commit:

| Build | Before | Now | Move | Cause |
|---|---|---|---|---|
| Frostfire | 1076.2 | 1065.7 | -10.5 | Ice Lance on Fire Blast's cooldown (d7541708d) |
| Destruction | 994.9 | 997.3 | +2.4 | Bane of Doom hold (58a1c3593) |
| Affliction | 1013.7 | 1015.8 | +2.1 | Orc (b33ad307d) and the Doom hold |
| Demonology | 1053.9 | 1055.6 | +1.7 | Doom hold |
| Destruction Fire | 976.7 | 977.4 | +0.7 | Doom hold |
| BM tallstrider / MM / LW / pet variants | | | -0.1 to -0.8 | no Sapper (b33ad307d) |

  Fire, Frost, Arcane and every other class: +0.0. The shared `sim/core` changes only touch healing rotations
  (rank cap applies to ProcMaskSpellHealing casts), raid damage events with `Share` set or tank-only raids, and tank
  score merging; Blood Craze heals only. None moved a DPS build.
- Explosives after b33ad307d: the ranged hunters keep the EZ-Thro Dark Bomb at 12 / 25 yd; melee keep the Sapper; the
  casters carry none (standards 11, "skip when it loses"). Even.
- Ice Lance / Fire Blast shared cooldown is a mage-only client rule with a user ruling; not a fairness question.
- Best race: report race is best everywhere except Fire (Gnome +1.5 at 5k, +0.7 at 20k last pass: noise).
  fairness-20261101-2 (Affliction race) was already resolved by b33ad307d; confirmed Orc 1015.8 / Human 1013.9.

## Ranking (5k sweep, report race)

Lone Wolf 1090.4, MM 1065.8, Frostfire 1065.7, Fire 1056.4, Demonology 1055.6, BM 1045.9, Fury 2H 1045.7, Fury 1041.6,
Ret 1039.2, Survival 1024.8, Affliction 1015.8, Arms 1008.7, Destruction 997.3, Destruction Fire 977.4, Shadow 948.7,
Frost 943.6, Feral 942.4, Arcane 939.2, Balance 891.9, Enhancement 883.0, Combat Daggers 791.6, Combat Swords 782.1,
Elemental 769.0, Smite 761.2, Seal Fate Daggers 752.9, Seal Fate Swords 747.3, Mutilate 743.6, Subtlety 733.3.
Top over bottom 1.49x. Frostfire drops from 2nd to 3rd (0.1 under MM).

## Open

None of mine. The pass doc's long-standing open items (Ice Lance Low range, FTT stacking, boss partial resist, Rage k)
are unchanged.
