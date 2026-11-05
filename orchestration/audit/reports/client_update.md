# Client updates: pass 10 (2026-11-05, master a079415316)

## Sources

| Source | Result |
|---|---|
| wago.tools (`wago.py latest --probe 300`) | 1.60.1.70790 (the launch client) |
| `table_summary.py 70755 70790` | 219 spells changed; trait trees 1111 and 1117; 5 poison enchants' charges; ItemSparse only on 6 keys and pages; ChrRaces, ItemEffect, SpellCastTimes, SpellDuration identical |
| Dev notes thread 2360696 | JSON rate-limited; the slug still reads "updated-november-2", so no new update post |
| foreverchanges.pro | 70790: "writes the November 3 hotfixes into the client files"; a November 4 hotfix: Defiance, Defensive Stance (threat) and Chain Lightning "2 sec instead of 2.5" |
| wago.tools/hotfixes | Chain Lightning SpellMisc 314274 / 322672: push 113341 sends CastingTimeIndex 19 (2.5 sec) to 70755 and 70790 clients up to 2026-11-04 23:43Z |

## DPS-relevant changes and sim status

| Change | Sim status |
|---|---|
| Hunter Aimed/Arcane/Serpent/Sniper, Aspect of the Hawk, Trueshot Aura (HUN-HF-03..10) | landed, values match |
| Rogue rework (push 113483, ROG-HF-01..06), tree 1111, poison charges | landed |
| Paladin Vengeance 5, 2H Spec Holy (PAL-HF-01/02) | landed |
| Shaman push 113380 (LB, CL, shocks, Stormstrike 150%/6 s, Call of Thunder) | landed |
| Chain Lightning cast 2.5 s (client index 19; FC says the 11-04 hotfix makes it 2 s) | landed at 2.5 s; FC contradicts the raw hotfix rows and the notes (SHA-HF-01) |
| Warrior Slam 15 s, Improved Slam, Improved Whirlwind, Reckless Swings, Spearing Strike 75% | landed |
| Defiance / Defensive Stance threat (11-04 hotfix) | landed (n/a to DPS) |
| Warlock Wrack 0.286, 2.5 a level | landed |
| Druid Shred/Claw/Ravage, Starfire x1.4 | landed, match within rounding |
| Druid Rip/Rake/Ferocious Bite exact rank values vs the sim's x1.25 | partial: finding client_update-20261105-1 (low, est. 0.1-0.2% feral) |

## Conclusions

- 70790 moves the server hotfixes into the client; the sim already had all of them. One low finding: feral bleeds and Bite should take the exact client numbers.
- Ledger: section "Build 1.60.1.70790" with rows DRU-U-16 and SHA-HF-01; Summary 767 rows.
- For the orchestrator: `foreverClientBuild` should become 1.60.1.70790.
- Open findings 20261030-2 (Missile Barrage dummy) and 20261030-4 (Fengus' Ferocity dummies): 70790 doesn't touch 400589 or 22817; unchanged.
