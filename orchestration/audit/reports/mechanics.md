# Mechanics pass 49 (2026-10-30, master 36a772fd1)

Range audited: d50909a37..36a772fd1. 11 code commits, all the raid sim and compsearch (third audit round). TASKS.md / update ledger: one raid-sim doc commit (98cf9bdbe); no `forever-update/*` merge.

## Get current

- Client: 1.60.1.70707 is new (was 70685). Fetched the mechanics set and diffed by spell (178 spell rows) plus the talent and curve tables.
- Dev notes thread 2360696: still 8 posts; post 7 (edited 10-27 16:47:14) and post 8 (edited 10-27 16:46:55) unchanged.
- Blue posts after 10-29 21:00: Kaivax's podcast post (10-29 16:17) and Linxy's retail PTR note, both before or off topic. No mechanics.
- Server hotfixes: `tools/audit/mechanics/hotfixes.py` is still missing from the tree; not re-read.

## Client 70707 against 70685

| Row | Change | Sim | Verdict |
|---|---|---|---|
| Trueshot Aura 1299346 / 1299348 / 19506 / 20905 (20906 50) | 30/40/50/75 -> 15/22/30/40 | 15/22/30/40/50 | matches (hotfix now in the client) |
| Aspect of the Hawk 13165-14321 (14322 55) | 20/35/50/70/90 -> 15/22/30/38/47 | 15/22/30/38/47/55 | matches |
| Blood Fury 20572 | RecoveryTime 180000 -> 120000 | 2 min | matches |
| Increased Hit Chance 01-05, 10 | new e1 aura 55 (spell hit), same points as the melee hit | gear hit unified (ruleset.go) | matches |
| Endurance 20550 e1 | aura 163 bp 2 misc 0 -> 127 | health only (post 7) | **mechanics-20261030-1** (user ruling) |
| Fengus' Ferocity 22817 | 200 AP / 200 RAP / ... -> four dummy auras | 200 AP / RAP | open item: server scripted or removed? No preset takes it |
| Well Fed 1249521 | +1 stat -> +1 skill 356 (fishing) | not simmed | none |
| Class rows | rogue / druid finishers Attributes_4 0x800; mage Precision 29438 to aura 55 bp 6, Missile Barrage e1 dummy, Ice Lance 8 s category CD, Accelerant mask; hunter Efficiency mask, trap / Counterattack diminishing; two mage talent rank swaps (TraitNodeEntry 130513 5->3, 130525 3->5) and trait spell swaps 135332 / 135339 (400588 / 28574) | | for the class and rank tuning auditors |

SpellItemEnchantment unchanged. TraitDefinitionEffectPoints +1 row (134452 curve 127736, new). CurvePoint: new curves 127736 / 127817, 83108 20/40 -> 25/50 swapped with 126757, 83083 points 3-4 moved to 83095; none of them a shared mechanic.

## Code range

- `buffs.go`: Furious Howl given at the player's level (`ForeverFuriousHowlAt`). Ranks match 70707 (24604 10-24 5+0.6, 24605 24-40 34+0.02, 24603 40-56 60+0.5, 24597 56-60 82+0.4; 83 at 60). Level 60 unchanged.
- `debuffs.go`: Expose Armor and Curse of Recklessness below the top rank get their own labels; `ForeverCurseOfRecklessnessAura` by level (14 / 28 / 42 / 56, armor 175 / 285 / 395 / 505). Level 60 unchanged.
- `warlock/curses.go`: a Forever warlock's own curse now uses the leveled aura (505 at 60, as before).
- `sim_concurrent.go`: empty party slots skip metrics. No mechanic.
- compsearch, web: no mechanic.

## Conclusions

One new finding (Endurance, user ruling). Every other shared row on 70707 matches the sim. Open: Fengus' Ferocity's dummy auras on 70707 (no preset effect).
