# Fairness pass 5 (2026-10-25, master 016daa6)

Result: findings (no new ones). fairness-20261025-1 (Marksmanship's report race) is still open, with new evidence. Full pass.

## Why it ran

40 watched files changed since 54878fa: every hunter keeps up its own Hunter's Mark and the Mark and Scorpid Sting duty presets are gone (e71b90f5, 214bc117, standards 23 ruling 2026-10-25), sim/hunter/hunters_mark.go, the hunter requests and APLs, mksettings spec_hunter.go, and item database fixes (Wall of the Dead, Changuk Smasher, Axe of Inevitable Ruin).

## What I checked

- Gear: `audit1b.py`, all 19 sets "5a PASS | 1b PASS".
- Race sweep: 36 base requests and variants, every race, 5000 iterations, seed 1, on the 016daa6 tree. Baseline refreshed (`ref/fairness/`, sha 016daa6).
- Movement: only hunters moved (BM -0.2, MM -1.0, SV -1.4, LW -3.1), all from the hunters now casting their own mark (the ruling's 0.5 to 2.5 DPS cost). Every non-hunter build and variant is identical to 54878fa, so the database changes leaked into no preset.
- Shared assumptions: Hunter's Mark adds ranged attack power only (sim/core/debuffs.go huntersMarkAura); no request (0 of 57) carries the raid's mark, so its removal is hunter-only. Raid Sunder is on in all 57 requests. Hunter own-mark vs warrior raid Sunder is the user's ruling; not re-flagged.
- Published against current: all within 1.9 at 5000 iterations; the shared 20k run says 0 of 57 off by more than 0.5.
- Best race: Marksmanship Human +2.6 / +2.3 over its Windshaper (20k, seeds 1 / 1000003) on the new request. Evidence added to fairness-20261025-1. Lone Wolf stays Windshaper.
- results-20261025-1 (mksettings vs requests consumes) is the results auditor's; not duplicated.
- Top over bottom: 1.64x.

## Ranking (5000 iterations, seed 1)

| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Gap | Before (54878fa) |
|---|---|---|---|---|---|---|---|---|
| 1 | Beast Mastery | Human | 1171.9 | Orc 1167.8 / Human | Human (1171.9) | 1171.5 | +0.4 | 1172.1 (-0.0%) |
| 2 | Marksmanship | Human | 1168.4 | Windshaper 1166.7 / Human | Windshaper (1166.7) | 1166.1 | +0.6 | 1169.4 (-0.1%) |
| 3 | Lone Wolf (hunter) | Windshaper | 1156.6 | Windshaper / Human 1156.1 | Windshaper (1156.6) | 1156.8 | -0.2 | 1159.7 (-0.3%) |
| 4 | Fire | Human | 1100.8 | Orc 1096.2 / Human | Human (1100.8) | 1101.0 | -0.2 | 1100.8 (+0.0%) |
| 5 | Frostfire | Human | 1098.0 | Orc 1085.8 / Human | Human (1098.0) | 1096.4 | +1.6 | 1098.0 (+0.0%) |
| 6 | Fury (dual wield) | Orc | 1069.1 | Orc / Human 1064.9 | Orc (1069.1) | 1068.7 | +0.4 | 1069.1 (+0.0%) |
| 7 | Demonology | Human | 1064.9 | Orc 1062.1 / Human | Human (1064.9) | 1065.1 | -0.2 | 1064.9 (+0.0%) |
| 8 | Survival | Human | 1055.7 | Orc 1054.7 / Human | Human (1055.7) | 1055.0 | +0.7 | 1057.1 (-0.1%) |
| 9 | Affliction | Human | 1036.4 | Orc 1033.4 / Human | Human (1036.4) | 1035.3 | +1.1 | 1036.4 (+0.0%) |
| 10 | Fury (two-hand) | Orc | 1034.5 | Orc / Human 1030.9 | Orc (1034.5) | 1033.7 | +0.8 | 1034.5 (+0.0%) |
| 11 | Destruction | Human | 1012.2 | Orc 1008.9 / Human | Human (1012.2) | 1011.8 | +0.4 | 1012.2 (+0.0%) |
| 12 | Destruction Fire | Human | 987.5 | Orc 987.1 / Human | Human (987.5) | 987.4 | +0.1 | 987.5 (+0.0%) |
| 13 | Arcane | Human | 983.2 | Orc 983.0 / Human | Human (983.2) | 982.5 | +0.7 | 983.2 (+0.0%) |
| 14 | Shadow | Human | 961.7 | Troll 954.5 / Human | Human (961.7) | 961.8 | -0.1 | 961.7 (+0.0%) |
| 15 | Arms | Orc | 956.5 | Orc / Human 955.4 | Orc (956.5) | 957.4 | -0.9 | 956.5 (+0.0%) |
| 16 | Retribution | Human | 947.0 | Undead 940.1 / Human | Human (947.0) | 945.1 | +1.9 | 947.0 (+0.0%) |
| 17 | Frost | High Order | 946.8 | Orc 944.8 / High Order | High Order (946.8) | 946.6 | +0.2 | 946.8 (+0.0%) |
| 18 | Enhancement | Orc | 887.7 | Orc / Dwarf 878.9 | Orc (887.7) | 888.1 | -0.4 | 887.7 (+0.0%) |
| 19 | Feral | Night Elf | 868.5 | Tauren 862.7 / Night Elf | Night Elf (868.5) | 868.1 | +0.4 | 868.5 (+0.0%) |
| 20 | Balance | Tauren | 856.7 | Tauren / Night Elf 854.1 | Tauren (856.7) | 856.3 | +0.4 | 856.7 (+0.0%) |
| 21 | Elemental | Windshaper | 777.1 | Windshaper / Dwarf 773.0 | Windshaper (777.1) | 777.4 | -0.3 | 777.1 (+0.0%) |
| 22 | Rogue Combat Daggers | Windshaper | 770.5 | Windshaper / High Order 769.9 | Windshaper (770.5) | 770.3 | +0.2 | 770.5 (+0.0%) |
| 23 | Smite | Undead | 765.5 | Undead / Gnome 745.6 | Undead (765.5) | 766.1 | -0.6 | 765.5 (+0.0%) |
| 24 | Rogue Combat Swords | Orc | 764.4 | Orc / High Order 763.7 | Windshaper (763.9) | 764.0 | -0.1 | 764.4 (+0.0%) |
| 25 | Rogue Seal Fate Daggers | Orc | 732.1 | Orc / Human 731.7 | Orc (732.1) | 731.5 | +0.6 | 732.1 (+0.0%) |
| 26 | Rogue Seal Fate Swords | Human | 729.4 | Orc 729.4 / Human | Orc (729.4) | 728.9 | +0.5 | 729.4 (+0.0%) |
| 27 | Rogue Mutilate | Orc | 719.4 | Orc / Human 718.7 | Windshaper (719.1) | 718.9 | +0.2 | 719.4 (+0.0%) |
| 28 | Rogue Subtlety | Orc | 715.8 | Orc / Night Elf 715.7 | Night Elf (715.7) | 715.9 | -0.2 | 715.8 (+0.0%) |

| Variant | Best race | DPS | Report's published (at its race) | Before |
|---|---|---|---|---|
| druid__balance-wrath | Windshaper | 814.0 | 813.3 (at Tauren 813.7) | 814.0 |
| paladin__level_60_shock | Human | 869.4 | 869.0 (at Human 869.4) | 869.4 |
| paladin__level_60_sub_holy | Human | 949.8 | 949.1 (at Human 949.8) | 949.8 |
| priest__shadow-holy | Human | 918.9 | 918.9 (at Human 918.9) | 918.9 |
| shaman__elemental-level60-mana-tide | Windshaper | 762.2 | 762.1 (at Windshaper 762.2) | 762.2 |
| shaman__enhancement-imp-gw | Orc | 883.8 | 884.3 (at Orc 883.8) | 883.8 |
| shaman__enhancement-mana-tide | Orc | 798.9 | 800.0 (at Orc 798.9) | 798.9 |
| shaman__enhancement-water-shield | Windshaper | 839.7 | 839.9 (at Orc 839.7) | 839.7 |

Top over bottom: 1.64x

## Notes for the next pass

- `druid__feral-cat.meta.json` and one other meta still say "Night Elf" with a space; `table.py` crashes on it. I ran it against a scratch copy of the metas with the spaces stripped (`jq '.race|=gsub(" ";"")'` into $AUDIT_WORK/fairness/tree, then AUDIT_TREE=that). The tool fix is `m["race"].replace(" ", "")`; I was told not to edit tools outside orchestration/audit this pass.
- Open items carried: hunter end-of-fight test, Ice Lance Low range, FTT + Windfury stacking, boss partial resist log, Rage k and pet scaling.
