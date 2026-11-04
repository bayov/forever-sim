# Results audit, pass 77 (master 6b480b060, 2026-11-04)

Trigger: 148 watched files (rogue hotfix 113483, warrior hotfix 113539, healing priest/paladin, UI).

- Shared run (68 builds, 20k, seed 1) vs builds.json: all 68 within 0.15 except none; the earlier warrior and shaman gaps closed (3966c9d26, 07:51 shaman re-sim).
- Run vs metas: 62 exact. Five rogue metas lag the hotfix re-sim: seal-fate-daggers -17.3, seal-fate-swords -16.6, mutilate -7.8, combat-swords -4.9, combat-daggers +3.2 (finding results-20261104-1). Cause of the moves: 316a600ac (Mutilate 44/63/88, Shank, Venom, tree); builds.json (e9185b1aa) already carries them.
- Paladin duty-crusader/-wisdom and subtlety differ 0.2 to 0.4 from meta (within noise of rounding; not flagged).
- Moves vs old reference (now replaced by this run): the five rogues above, warlock affliction x3 (+18 to +21, earlier re-search), warrior Fury 2H rewrites (+2 to +3.4).
- Drift (regen.py): same as the reference except earlier fixes (feral duty-ff explosive, rogue subtlety race, affliction_shadow_bolt rename). No new drift.
- Unread fields: no output besides expected.
- Result: findings (1 low). Reference req_summary.txt replaced.
