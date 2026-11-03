# Fairness pass 26 (2026-11-03, master f23ea25)

Result: quiet (no new finding). Targeted pass: 21 watched files since f5c99ad, all class-local (warlock Affliction
rework, rogue Subtlety meta race, Enhancement Earth Shock hold, compsearch pool rename). No `sim/core`, standards or
database change, so no full re-sweep; the f5c99ad baseline stays.

## What I checked

- Shared 20k run (`runs/vs_published.txt`): 62 of 68 within 0.5 of published; the six shamans are the known stale metas
  (results-20261102-1 update, clears with the regeneration). Against the previous shared run every build is within 0.2.
- Race sweep of the changed builds, 20k, seed 1, the loop's f23ea25 CLI and tree:

| Build | Report race | Best | Runner-up |
|---|---|---|---|
| warlock__affliction (Wrack, Imp sacrificed) | Undead 1034.1 | Undead | Orc 1032.8, Human 1030.7 |
| warlock__affliction_shadow_bolt | Orc 1015.6 | Orc | Human 1013.5, Undead 1012.4 |
| rogue__subtlety | NightElf 881.7 | NightElf | Undead 880.2 |
| shaman__enhancement-level60 | Orc 1159.4 | Orc | Windshaper 1154.8 |
| shaman__enhancement-imp-gw | Orc 1149.8 | Orc | Windshaper 1145.6 |

- Affliction duty variants carry the base's race (Undead) and talents; metas and requests agree.
- Movement traced: Affliction 1015.6 -> 1034.1 is the new preset (dc0a0fcfa, 734fe0e93), published updated; the
  Shadow Bolt build kept its number as the variant. Enhancement +0.2 (203b686f6). Nothing else moved: no leak.
- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".

## Results auditor overlap

- results-20261103-1 (Subtlety meta race "Night Elf"): fixed by 5b766ca2b; NightElf is the best race (+1.5 over Undead).

## Ranking

Unchanged order from pass 25 (see `orchestration/fairness_report.md` §1); Affliction 1034.1 stays 15th, between
Survival 1041.1 and Destruction Fire 997.1. Top over bottom 1.73x.
