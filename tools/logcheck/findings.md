# Log audit findings

What the log audit found that we haven't settled yet. The `log-audit` skill fills it in every
hour while its session is open, and we go through it together on `/log-audit` (the skill's
sections 2 and 6 have the format).

## Fix in the sim

### Stormstrike hits about 1.6x the sim from Nov 11 10:37 to Nov 13 04:59

- **Flag**: `Stormstrike (17364): raw`, mob levels 46 to 59 (114 flags, 49 windows of `WoWCombatLog-111126_095013.txt`, Nov 11 10:37 to Nov 12 07:41). `Melee: raw` stays close on most of them.
- **Windows**: `111126_095013 10:37-10:54` log 544 to 552 vs the sim's 330 to 332 (36 and 21 hits, 1.65x, Staves). `11:02-11:22` 523 vs 337 (33 hits). `13:56-13:58` with Rockbiter 786 vs 503 (7 hits). `23:15-23:19` 814 vs 333 (2.44x), `03:08-03:16` 827 vs 335 (2.47x), both with Shadow Oil and sheet AP about 790 over the sim's. `05:25-05:42` 625 to 639 vs 376 (38 hits). Almost every window sits at 1.55 to 1.70x (Alcaz Willcrusher and The Chief's Enforcer, both two-handers). Only `07:38-07:41` (Rockbiter, level 59 boss) fits: 557.6 vs 523.9 (1.06x, 9 hits). The last window before Nov 11 10:37 with Stormstrike (`075640 00:10:31`, Rockbiter, client tables 70862) fit: 731.8 vs 725.0.
- **What we think**: the sim's Stormstrike takes the 2026-11-10 class tuning (100% weapon damage, 8 sec cooldown, `StormstrikeWeaponDamagePercent` in `sim/shaman/stormstrike.go`), but its own comment says no server row carries it (`SpellEffect 1368308` still 150). 1.5x fits the windows at 1.55 to 1.70x, and Rockbiter windows with a matching sheet AP (13:56, 14:32 to 20:56) are 1.55 to 1.6x too, so the AP gap below isn't the cause. The one fitting window (07:38) could be the tuning going live, or a one-off (the sim's Rockbiter AP was 610 over the sheet's there). Touches the Stormstrike notes in `notes.md`.
- **Nov 13 update**: `111226_132121` Nov 12 13:52 to 18:10 (18 windows) and `111226_230324` Nov 13 04:36 to 04:59 (client tables 70888) are still 1.56 to 1.77x (log 523 to 639 vs the sim's 322 to 369, 5 to 143 hits a window). But every window of `111326_101810` (client tables 1.60.1.70938, Nov 13 10:53 to 12:48) fits: `11:29-11:43` 365 vs 364 (10 hits), `11:47-11:51` 372 vs 361, `12:09-12:19` 389 vs 379 (30), `12:03-12:05` 393 vs 360 and `12:24-12:31` 418 vs 387 (a few hits, Melee raw is +8% to +10% there too). The other windows of that log (`10:53`, `11:44`, `11:51`, `11:56`, `11:59`, `12:22`, `12:35`, `12:39`) had no Stormstrike raw flag.
- **What we think now**: the 100% weapon damage in the sim is the live value, and it went live with the client tables bump 70888 to 70938 between Nov 13 04:59 and 10:53. The 1.5x of Nov 11 to 13 04:59 was the old row. So no sim change is needed.
- **Next step**: nothing to change in the sim. Move it to Done at the next `/log-audit` if the user agrees, and recheck the cooldown (6 or 8 sec) from cast gaps in `111326_101810`.
- **Nov 16 update**: the Nov 15 and 16 logs (`111526_162618`, `111626_123605`, client tables 70938) keep fitting: Stormstrike raw is 1.00x the sim in 15 windows (90 hits) and 1.03x in 4 (25 hits) at mob levels up to 60. The 100% weapon damage is the live value.
- **Nov 17 update**: `111626_123605` Nov 16 evening (49 windows run, 18:32 to 23:04, levels 54 to 60): Stormstrike raw 383.0 vs the sim's 361.5 (4 hits, `22:44:55`) and 378.5 vs 362.1 (4 hits, `23:00:21`), +5%, the same size as `Melee: raw` there. Still no 1.5x.
- **Nov 18 update**: `111726_171435` Nov 17 18:20 to 22:34 (61 windows run, 70 skipped under 50 attacks): Stormstrike raw is 0.95 to 0.99x the sim at levels 55 to 59 (16 hits), 1.07x at 60 (15 hits), 1.06x at 61 (10) and 1.11x at 62 (33 hits, 433.6 vs 391.5). Still no 1.5x. The rise with mob level matches `Melee: raw` and the level finding under "Collect more data".
- **Nov 20 update**: `111926_175614` and `112026_111427` (Nov 19 22:13 to Nov 20 16:23): Stormstrike raw is 1.03x the sim at level 60 (15 hits, 376.6 vs 367.1) and 1.08x at 62 (28 hits, 447.8 vs 414.5), the same size as `Melee: raw` (1.07 to 1.09x at 60 to 62). Still no 1.5x.
- **Nov 22 update**: `112126_201216` and `112226_113558` (45 windows run, Nov 21 20:14 to Nov 22 13:43): Stormstrike raw is 1.00x the sim at levels 58 to 60 in the Nov 21 and Nov 22 morning windows (48 hits, 419.1 vs 418.3 and 380.5 vs 383.2) and 1.05x in the afternoon at 58 to 60 (31 hits). At 62 it is 1.28x (6 hits, 507.0 vs 396.6 in `13:42:21`), the level shape of `Melee: raw` (1.24x there, 26 hits). Still no 1.5x.
- **Nov 25 update**: `112326_183423`, `112426_110702` and `112426_172127` (Nov 23 18:56 to Nov 24 18:06, 34 windows run, 134 skipped under 50 attacks): Stormstrike raw is 1.04x the sim at 58 and 59 (33 hits, 363.3 vs 350.9, 404.9 vs 389.5), 1.06x at 61 (8 hits), 1.09x at 56 (11 hits). Still no 1.5x, the same size as `Melee: raw`.
- **Status**: open, first seen 2026-11-12, updated 2026-11-25 (log fits, ready for Done).

## Collect more data

### Fixed-damage hits move with the mob's level around 60

- **Flag**: `Fiery Weapon (13897): raw` (NEW, 25 flags), and the same drift in `Cloud of Chained Essence (1300240) ticks: raw` (NEW), `Magma Totem (10581): raw`, `Flame Shock (10448) ticks: raw`, `Windfury Weapon (439440): raw`, `Earth Shock (10414): raw`, `Frost Shock (10473): raw` (NEW). Mob levels 58 to 62, our level 60.
- **Windows**: `111526_162618` 19:20 to 22:32 and `111626_123605` 14:35 to 16:21. Fiery Weapon (flat 40 in the sim) is 38 on every hit at level 58 (8 windows, 74 hits, ratio 0.96), 39 to 40 at 59 to 61 (ratio 1.00), and 42 on every hit at level 62 (5 windows, 59 hits, ratio 1.05), all at the same sheet. Cloud of Chained Essence (140 in the sim) is 138 at 58, 143 at 61 (14 hits) and 148 at 62 (9 hits). Magma Totem runs 0.98 to 1.01 at 56 to 60, 1.02 at 61 (81 hits) and 1.045 at 62 (98 hits, 95.0 vs 90.0 in `16:19:47`). Flame Shock ticks 1.02 at 56 to 61, 1.07 at 62 (59 hits). Frost Shock 600.5 vs 566.3 and Earth Shock 683.6 vs 593.2 at 62 (6 and 5 hits). Melee at level 62 does not move (0.97, 80 hits).
- **What we think**: the damage a hit does depends on the target's level, about 1 point or 2.5% a level on a flat 40, and the sim's hit doesn't. It would also explain the 1 to 3 less on dungeon mobs in "Dungeon mobs take about 2 less" (level 58 and below take less). Weak signal: level 62 mobs are few, and the sheet's Nature power is 16 over the sim's in 68 of these 84 windows (see the sheet finding), which adds about 1 to 2% on its own.
- **Next step**: collect level 58 to 63 hits of Fiery Weapon (flat 40) and the Chained Essence cloud (flat 140) against the same fight, then fit a per level factor and ask the user before changing `sim/core` damage. Check whether the ledger's older windows at levels above 60 show it too.
- **Nov 17 update**: `111626_123605` Nov 16 18:32 to 23:04 (49 windows) has the drift on the low side too, at levels 53 to 60. Level 54 in `22:46:40`: Magma Totem 88.1 vs 89.2 (51 hits, 0.99), Fiery Weapon 38 vs 40 (8 hits, 0.95), Cloud of Chained Essence 138 vs 140 (9 hits), Flame Shock ticks 86.3 vs 88.3 at 54 in `22:21:59`. At 55 to 60 the log is 1 to 2 above (Magma Totem 90 vs 89 in about 30 windows, Flame Shock ticks 93 vs 91 in about 25, Fire Nova 311 to 318 vs 308 to 310 in 11), the same size as the sheet's 16 Nature power gap would give on spells. It supports a per level factor, but doesn't settle it, and no level 62 mobs in these windows.
- **Nov 18 update**: `111726_171435` Nov 17 17:42 to 23:40 (61 windows run, mob levels 50 to 62) settles the shape, summed hits log vs sim by mob level. Magma Totem: 1.00 at 51 to 55, 1.01 at 59 and 60 (691 hits at 60), 1.022 at 61 (110 hits), 1.066 at 62 (36 hits, 96.7 vs 90.7). Flame Shock ticks: 1.02 at 52 to 60 (170 hits at 60), 1.03 at 61 (70), 1.06 at 62 (72 hits, 98.3 vs 92.6). Fire Nova: 1.01 at 55 and 60, 1.04 at 61, 1.06 at 62 (11 hits). Earth Shock: 1.014 at 59 and 60, 1.04 at 61, 1.06 at 62 (20 hits, 630.8 vs 594.4). Fiery Weapon 42.0 vs 40 on all 65 hits at 62, 38.0 at 56 (5 hits). Shadow Bolt 1.04 to 1.06 at 53 to 62 (no clear step, the oil's known gap). Windfury Weapon and Melee/Stormstrike climb the same way (1.01 at 60, 1.09 at 61 and 62), see the sheet finding. So fixed-damage hits step up about 2% a level from 60 to 62 on top of a flat 1 to 2% the spell power gap gives, and fall below the sim at 56 and under. Cloud of Chained Essence is 138 at 56, 139 at 57 and 143 at 61 (13 hits). `18:40:08` has 183.1 vs 140 at level 55 (13 hits, 140 to 280), two clouds at once like the 280 of Nov 16.
- **Nov 20 update**: `111926_175614` and `112026_111427` (36 windows run, 17 skipped under 50 attacks, levels 57 to 62), summed hits log vs sim by mob level. Magma Totem: 1.011 at 58 to 60 (1522 hits), 1.033 at 61 (232), 1.055 at 62 (192, 94.9 vs 90.0). Flame Shock ticks: 1.02 at 58 to 60, 1.041 at 61 (69), 1.061 at 62 (67, 99.3 vs 93.6). Fire Nova 1.015 at 60, 1.031 at 61, 1.075 at 62 (22 hits). Earth Shock 1.013 at 59 and 60, 1.077 at 62 (19 hits, 642.4 vs 596.6). Fiery Weapon 41 vs 40 at 61 and 62 (113 hits, 1.025). Windfury Weapon 1.024 at 59, 1.047 at 60, 1.094 at 61. Same shape as Nov 18: flat to 60, rising about 2% a level above it. `Grounded Electroshock (1315996)` (NEW, our proc) is 62 vs 60 at level 62 (5 hits), the same drift.
- **Nov 22 update**: `112126_201216` and `112226_113558` (45 windows run, 9 skipped under 50 attacks, levels 50 to 62), summed hits log vs sim by mob level. Magma Totem: 0.989 at 54 (118 hits), 1.007 to 1.010 at 58 to 60 (856 hits), 1.023 at 61 (27), 1.058 at 62 (24 hits, 95.2 vs 90.0 in `13:42:21`). Flame Shock ticks: 1.024 to 1.026 at 58 to 60 (240 hits), 1.032 at 61 (14), 1.074 at 62 (14 hits, 101 vs 94). Fire Nova 1.012 at 58, 1.039 at 61, 1.071 at 62 (6 hits). Earth Shock 1.011 to 1.015 at 58 and 59, 1.047 at 61, 1.068 at 62 (7 hits). Fiery Weapon 38.0 vs 40 at 54 (33 hits, 0.95), 42.0 vs 40 at 62 (22 hits, 1.05). Windfury Weapon 1.011 at 60, 1.10 at 61, 1.174 at 62 (15 hits). Cloud of Chained Essence 138 at 53 and 56, 143 at 61 (5 hits). Same shape as Nov 20, with the 62 step a little larger on Windfury Weapon and Melee.
- **Nov 25 update**: `112326_183423`, `112426_110702` and `112426_172127` (34 windows run, levels 55 to 63), flagged hits only, so few at 61 and above. Magma Totem 1.010 at 56 to 58 (53 hits), 0.989 at 60 (10 hits), 0.996 at 61 (20 hits, 90.3 vs 90.7). Flame Shock ticks 1.00 at 56, 1.02 at 55 and 58, 0.989 at 61 (11 hits). Windfury Weapon 1.033 at 58 (18 hits) and 1.056 at 63 (10 hits, 436.1 vs 412.8). No level 62 spell window, so the 62 step is not rechecked.
- **Nov 26 update**: `112526_195841` `22:00:52` (Alterac Valley NPCs): Flame Shock ticks 96.7 vs 93.5 at level 58 (6 hits, 1.034) and 98.3 vs 93.1 at 61 (6 hits, 1.056), Magma Totem 91 vs 90.5 at 58 and 60 (22 hits, 1.010). Same curve as before. No level 62 spell hits.
- **Status**: open, first seen 2026-11-16, last updated 2026-11-26.

### Fiery Weapon partial resists and rounding

- **Flag**: `Fiery Weapon (13897): partial` (10 windows), `amount` (11 windows plus 7 all-levels), `immune` 1 of 31, `absorb` 1 of 5, `other` 3 of 7 (NEW).
- **Windows**: partial on fire-resistant mobs, like 8 of 8 at level 57 and 5 of 5 at 58 in `111526_162618 20:34:12-20:41:06`, 10 of 10 at level 60 in `20:47:39-20:52:21`, and 2 to 4 of 6 to 14 in `111626_123605 14:14` to `16:09`. `amount` rounds up on 1 to 4 of 4 to 13 hits, mostly at level 61 (the log's hit is 40.0 to 40.7 vs the sim's flat 40.0).
- **What we think**: the partials are mobs with real fire resistance, the same cause as the Searing Totem and Flame Shock partials in `known.json`. The rounding is the 16 spell power the sim lacks (see the sheet finding) or the level effect above, not a separate mechanic.
- **Next step**: add `Fiery Weapon (13897): partial` to `known.json` at the next `/log-audit` if the user agrees, then recheck `amount` after the sheet gap is found.
- **Nov 18 update**: `111726_171435` has `Fiery Weapon (13897): immune` in 17 flag lines (4 of 9 at 52, 6 of 8 at 54, 11 of 11 at 54 in `17:56:40`, 5 of 8 at level 50), on the same fire-immune mobs as the known `Magma Totem (10581): immune` and `Fire Nova (408428): immune` rows. Same cause, so add it to `known.json` with them. `amount` (Magma Totem rounds up on 46 of 59 and 34 of 41 hits, the log's hit 90.7 to 90.9 vs 90.2) is the spell power gap.
- **Nov 22 update**: `112226_113558` `11:44` to `11:52` (levels 54 and 55): `Fiery Weapon (13897): partial` 4 of 14, 15 of 16 and 4 of 6, with `Magma Totem (10581): partial` 23 of 50, 47 of 57, `Fire Nova (408428): partial` 5 of 13, 13 of 14 in the same windows (fire-resistant mobs). `Fiery Weapon (13897): immune` 3 of 6, 6 of 6 and 12 of 12 in `112126_201216` `20:17` to `20:29` (levels 52 to 56, fire-immune mobs), and `Magma Totem (10581): amount` rounds up on 13 to 20 of 14 to 23 hits at 52 to 58 (the log's hit 90.6 to 90.99 vs the sim's 90.19). Same causes as above.
- **Status**: open, first seen 2026-11-16, last updated 2026-11-22.

### Shadow Oil's Shadow Bolt hits about 11 harder than the sim

- **Flag**: `Shadow Bolt (350026): raw`, levels 57 to 60.
- **Windows**: `111326_101810` (after the 2026-11-13 fix): 129.7 vs the sim's 118.6 (6 hits), 129.8 vs 118.6 (8), 128.9 vs 118.2 (10).
- **What we think**: the sim gives 63 to 71 plus 0.56 of spell power (`shadowOilDamageRange` in `sim/core/consumes.go`), from client 70338 with scaling capped at level 17. The log's 122 to 140 needs about 11 more: a higher base, a bigger coefficient, or scaling past 17.
- **Next step**: more hits, and read 350026's SpellEffect and SpellLevels in the latest client tables.
- **Status**: open, first seen 2026-11-13, last updated 2026-11-13.

### Dungeon mobs take about 2 less spell damage per hit

- **Flag**: `Flame Shock (8050/8053/10447)` ticks and direct `raw`, `Lightning Shield (26365..26369): raw`, `Earth Shock (8045): raw`, only in dungeon windows.
- **Windows**: Nov 5-8 dungeon runs (Wailing Caverns, Shadowfang Keep, Razorfen Kraul, Gnomeregan): Flame Shock rank 2 ticks 12 to 10, rank 4 ticks 26 to 24, direct 54 to 52 and 101 to 99, orbs 24 to 22-23. Open world mobs of the same level, same hour: full damage.
- **What we think**: a per-mob trait, not per dungeon or time. 54 of 135 even-level dungeon mobs, 0 of 144 odd-level dungeon mobs and 5 of 739 world mobs. Same sheet spell power and auras. Guess: a hidden per-mob resistance (but see Maraudon below: a flat 2 off Physical too). See the Need to Verify item under "All classes".
- **Window** `WoWCombatLog-110926_151906.txt 23:11:05-23:26:03` (Uldaman): every level 43 mob (Stonevault Brawler and Geomancer, Earthen Rocksmasher and Sculptor, Deadly Cleft Scorpid, Shadowforge Archaeologist and Darkcaster) takes 2 less: Magma Totem 51 to 52 (157 hits) vs 53 to 54, Flametongue Attack 72 vs 74, Fire Nova 189.8 vs 191 to 192. Level 42 and 44 mobs hit by the same Magma tick take full damage. So odd-level dungeon mobs have it too.
- **Window** `WoWCombatLog-110926_151906.txt 23:51:48-23:55:38`: level 47 mobs take about 1 less than level 44 and 45 in the same window: Magma Totem 52 (15 hits) vs the sim's 53, Flame Shock ticks 46.3 vs 47, Flametongue Attack 73.2 vs 74.0. Smaller than 2, so maybe only some of them have it.
- **Window** `WoWCombatLog-111026_075640.txt 20:06:50` and `21:02:11` (Maraudon, 2908 and 2748 attacks): odd-level mobs take 2 less again. Magma Totem 51 vs the sim's 53 (323 hits), Flametongue Attack 78 to 79 vs 80 at levels 45 and 47, Flame Shock ticks 44.7 to 45 vs 46, Farraki Enforcement 292 vs 294 (13 hits). On those mobs Lightning Shield equals the sim, because the +2 of "Spell hits run about 6 spell power above the sim" cancels it. Farraki Enforcement is Physical, and 2 off 294 is the same 2 as off Magma Totem's 53, so it looks like a flat 2 less damage taken on every school, not a resistance (a resistance would take a share, and none off Physical).
- **Window** `WoWCombatLog-111126_095013.txt` Nov 11 10:37 to Nov 12 07:17 (Nov 12 audit): the same 2 to 3 less turns up at levels 47 to 55 on spells, mostly in what looks like open world fights (10:37 to 11:48 has Overlord Mok'Morokk, a Dustwallow Marsh mob, zones not checked). Magma Totem 52.0 vs the sim's 54.0 in 13 windows (30 to 37 hits each), Lightning Shield (26370) 172 to 173 vs 174 to 176 in 24 windows, Flame Shock (10448) ticks 68 to 69 vs 70 to 71 in 6, Searing Totem (10435) 37.1 vs 40.0 and 38.2 vs 39.9 in 13, Flametongue Attack 84.8 to 85.1 vs 87.6 to 88.0, Lightning Bolt (15207) 342 to 345 vs 346 to 350 in 12, Earth Shock (10413) 411 to 415 vs 415 to 417 in 6. The sheet's Nature power was 0.6 to 1.5 over the sim's in those windows, which is too small for all of it. If a flat 2 off every hit holds at these levels, it isn't a dungeon trait.
- **Window** `WoWCombatLog-111226_132121.txt` Nov 12 13:52 to 18:10 (about 20 windows, levels 55 to 59): Lightning Shield (26363) 218.2 vs the sim's 220.1 (33 hits), 224 vs 225.1, 227.4 vs 228.2, Magma Totem (10581) 72.0 vs 74.0 (29 hits), 74.0 vs 74.5, 75.0 vs 76.0 (36 hits), Lightning Bolt (15208) 388.2 vs 391.4 (26 hits), 409.5 vs 413.3, 420.1 vs 429.9. Still 1 to 3 less on every school.
- **Window** `WoWCombatLog-111326_101810.txt` Nov 13 11:29 to 12:39 (tables 70938): Magma Totem 75 vs 76.0 in 8 windows, so still 1 less (it rounds up on 9 to 23 of 13 to 30 hits, the log's hit 76.4 to 76.9 vs the sim's 76.05, so it may be spell power). The other schools run above the sim there (Lightning Shield 238 vs 233, Lightning Bolt 440 vs 426, Earth Shock 435 vs 426), which fits the sheet's Nature power being 16.5 over the sim's (see the AP finding below), so the -2 is hidden by the spell power gap.
- **Next step**: check the unfiltered logs for those mob GUIDs. Then Flame Shock a dungeon trash mob and an open world mob of the same level, 30 ticks each.
- **Status**: open, first seen 2026-11-08, last updated 2026-11-13.

### Maelstrom Weapon cut at 4 stacks may not be linear

- **Flag**: none (from the cast times, not logcheck).
- **Windows**: Nov 9 21:49 to 23:30 at 5/5: Lightning Bolt at 4 stacks took 0.52 sec avg in 13 casts (0.43 to 0.71, most 0.50 to 0.51). The sim gives 0.30 (80% off). At 4/5 with 4 stacks 0.48 (4 casts) vs the sim's 0.54, at 3/5 0.82 (2) vs 0.78.
- **Per cast** (2026-11-10): there's no 0.5 sec floor. At 4/5, 26 bolts at 5 stacks took 0.21 to 0.47 sec (avg 0.31). At 5/5, 4 stacks took 0.50 to 0.51 on 10 of 13 casts, one bolt at 3 stacks took 1.01 and one at 2 stacks 1.50. Those fit 0.5 sec per stack short of 5 at 5/5. Ranks 2 to 4 fit the sim's straight line.
- **What we think**: 5 stacks fit the sim exactly, but 4 stacks at 5/5 cut 67%, not 80%. Maybe the stack count at cast start is off by one in our reading, or the cut isn't linear in stacks. It matters little since we cast at 5 stacks.
- **Next step**: count more 2 to 4 stack casts in new logs before changing anything (`mw/cast.py` in the scratch dir did it, worth moving into `tools/combat_log/`).
- **Status**: open, first seen 2026-11-09, last updated 2026-11-10.

### Earth Shock is never partly resisted once it lands

- **Flag**: `Earth Shock (8045/8046/10412)`: `resisted` on landed hits.
- **Windows**: Nov 5-8 on Nature-resistant mobs. Earth Shock had 77 whole hits, 0 partial and 31 RESIST, while Lightning Shield orbs on the same mobs were partly resisted 60 of 63. Druid of the Fang alone: 57 whole, 0 partial, 27 RESIST, orbs 31 of 31 partial.
- **What we think**: the sim gives a landed binary spell a partial roll (`sim/core/spell_resistances.go`, `sim/shaman/earth_shock.go`). The October beta serpents (6 landed, all partial) disagree, so live may have changed.
- **Next step**: the user's ruling (2026-11-08) is to keep monitoring. Count Earth Shock partials on resistant mobs in each new log.
- **Nov 20 update**: Blackwing Lair Chromatic mobs show it again, now with a high Nature resist. `111926_175614 23:23:28-23:25:09`: Earth Shock `miss` (RESIST) 5 of 7 on Chromatic Elite Guards at level 59 (the sim gives 0.7%), and `Grounded Electroshock (1315996)` partial 7 of 7 landed on the same mobs. `112026_111427 15:53:34-16:01:48`: Earth Shock RESIST 3 of 7 at level 62 (sim 1.5%): Chromatic Whelp, Dragonspawn and Gyth. These are Nature-resistant mobs, logcheck's mob has no resist, so the rates don't tell us whether landed hits can be partial. Nothing to change.
- **Status**: open, first seen 2026-11-08, last updated 2026-11-20.

## logcheck issues

### Sheet Attack power far from the sim's on some exports

- **Flag**: `sheet` (Attack power), and `Melee: raw` in the same windows.
- **Windows**: `111126_095013`, sheet vs sim: `10:37-10:54` 1300 vs 610, `23:15-23:47` 1465 to 1508 vs 709 to 714, `00:26-01:04` 1476 vs 733, `03:08-03:43` 1412 to 1505 vs 647 to 720 (Shadow Oil and Flametongue, Ghost Wolf in the export's buffs on most). The other way: `19:21-19:31` sheet 732 vs sim 1337, `07:33-07:41` 869 to 988 vs 1479 to 1570 (Rockbiter, the sim adds its AP and the export doesn't have it). Logged AP in the fights is 1318 to 2311, over the sheet in most. `Melee: raw` 501 vs 317 in `23:15-23:19`, and 506 vs 318 in `03:08-03:16`, but 347 vs 334 in `10:37-10:54`.
- **What we think**: an AP source in the sheet that `buffAuras` in `tools/logcheck/buffs.go` doesn't read (about 700 to 800 AP, up in some exports only), or the export reading the sheet before an AP aura applies. Not found. It skews the Melee rows in the windows listed, not the spells.
- **Nov 13 update**: `111226_132121` Nov 12 13:52 to 18:10 (19 windows) sheet AP is 80 to 109 over the sim's, `111226_230324` `04:36-04:57` (5 windows) 1532 vs 786 (-746, Health -300 too, buffs up: Authority, Well Fed, Strength of Earth), and `111326_101810` `11:44`, `11:45`, `11:47` (+78), `12:22` to `12:35` (+60 to +76) the other way. `11:29-11:43` fits (815 vs 815), but its Nature power is 109 vs 92.5 (-16.5), which fits Lightning Shield 238 vs 233, Lightning Bolt 440 vs 426 and Earth Shock 435 vs 426. Melee raw in `111326_101810` runs +4% to +15% over the sim in 20 windows (like 372 vs 324 at `11:45`, 374 vs 337 at `12:03`, 391 vs 359 at `12:07`, 424 vs 401 at `12:32`), though Stormstrike fits there.
- **Next step**: for the 23:15, 03:08 and Nov 13 04:36 exports, list the auras up on us from the log and the sheet's AP rows (`Export.lua`), then find the one that adds the 700 to 800. For `11:29`, find the +16.5 Nature power (an elixir or buff `buffAuras` doesn't read).
- **Nov 16 update**: `111526_162618` and `111626_123605` (84 windows run, 63 skipped under 50 attacks): the sheet's Nature power is 16 over the sim's (+15.7 to +16.5) in 68 windows and fits in 16 (like `111626 15:15` and `15:52`, +1.5). AP is within 60 in all but 5 windows (+124 at `19:02:49`, +126 at `19:30:26`, +699 at `19:58:20`, -92 at `22:13:51`, +254 at `16:19:28`). `Melee: raw` is 1.04x the sim at level 60 and below on Nov 15 (42 windows, 491 hits) and 1.06x on Nov 16 (21 windows, 218 hits), so the Nov 13 gap stays. The +16 Nature power is the same one as in `111326_101810 11:29` and probably an item or buff `buffAuras` doesn't read.
- **Nov 17 update**: `111626_123605` Nov 16 18:32 to 23:04 (49 windows run, 28 skipped under 50 attacks): the sheet's Nature power is 16 to 17 over the sim's in 39 windows (-16.70 in 24, -15.70 in 15) and fits in 6 (+1.2 to +2.3). AP is within 15 in all but `18:57:04` (-110) and `22:25:09` (-80). `Melee: raw` is 1.04x to 1.12x the sim in 30 windows (like 414.9 vs 374.8 at `19:11:07`, 467.3 vs 396.9 at `18:58:57`), so the Nov 13 gap stays.
- **Nov 18 update**: `111726_171435` Nov 17 (61 windows run): sheet AP is within 4 of the sim's in every window the digest lists (-1 to -4) and Nature power is 17 over in the windows with a Mageblood export, yet `Melee: raw` is still high: 1.03 to 1.05x at levels 55 to 60 (8 to 78 hits a level), 1.06 at 54 (54 hits), 1.07 to 1.11 at 50 to 53 (few hits), 1.09 at 61 (73 hits) and 1.13 at 62 (66 hits, 383.4 vs 340.4). So the gap is not AP. It looks level shaped like the fixed-damage drift above, plus a flat 3 to 5%. Windfury Weapon raw 1.01 at 60, 1.09 at 61 and 62 (52 hits at 62). Next step stays: pin the melee gap with `tools/combat_log/` once the level shape is known.
- **Nov 22 update**: `112226_113558` (37 windows run, 9 skipped) and `112126_201216` (8 windows): every window has the sheet flag, the Nature power gap is -16.5 as before (`11:53:19`), and sheet AP is within 7 of the sim's in the morning, -53 to -59 in the afternoon (`13:08` to `13:42`, -165 at `13:28`), +113 at `112126 20:42`. `Melee: raw` is 1.05x at levels 60 and below on Nov 21 (40 hits), 1.25x in the Nov 22 morning (164 hits, 355.3 vs 283.8) and 1.11x in the afternoon (180 hits, 398.0 vs 360.0), 1.17x at 61 and 62 (53 hits). The morning sim average is low because its Melee range starts at 41 (`11:53:19`, `12:04:35`, `12:08`: sim avg 125 to 217 vs the log's 336 to 371), a weak weapon the sim swings far more than the log did, in windows the trinket and weapon swap `change` flags cover. Stormstrike fits there (0.99x, 30 hits), so it isn't the AP. Not pinned, only the weapon mix is a guess.
- **Nov 25 update**: all 34 windows run have the sheet flag. Nature power is within +1.2 to +2.3 of the sim's in 26 windows and -13.6 to -16.6 in 5, Attack power is within 3 in most (-80 in 2, -16 in 1). `Melee: raw` is 1.04 to 1.06x at 57 and 58 (124 hits), 0.98x at 59 (20 hits), 1.08x at 61 (30 hits), 1.07x at 63 (12 hits). The Nov 22 morning 1.25x does not come back.
- **Status**: open, first seen 2026-11-12, updated 2026-11-25.

### Trinket swaps and skill ups flag as a character change

- **Flag**: `change` (30 windows of `111226_132121` and `111326_101810`).
- **Windows**: 17 of them say "the character changed ... Vigilance Charm / Chained Essence of Eranikus / Jewel or Rune of the Guard Captain" (on-use trinkets swapped), 12 say "weapon skill went up during the window, Two-Handed Maces 281 to 295 of 295", and one is a swap of Alcaz Willcrusher for Fist of the Damned and nothing.
- **What we think**: the audit treats a trinket swap as a gear change, but the sim's trinkets don't change the rows we compare. With history present, `change` should mean a real change to the damage rows (weapon, stat gear). The skill-up flag is the known "skill goes up as we swing" caveat.
- **Next step**: have the audit leave trinket swaps out of `change` unless the trinket has an effect the sim runs (and say skill ups only once per window).
- **Nov 16 update**: 71 of the 84 windows of `111526_162618` and `111626_123605` say "the character changed" (trinkets Vigilance Charm, Chained Essence of Eranikus, Rune and Jewel of the Guard Captain, Smokey's Lighter; Argent Dawn Commission), 17 say "weapon skill went up" (Two-Handed Axes 274 to 299 of 300). Two real swaps: Malicious Axe to Bonechill Hammer with a shield at `19:54:33`, and Totem of Sustaining to Egan's Blaster at `20:28:02`. The ledger line now leaves out 369 windows because of this, so the flag counts there are thin.
- **Nov 17 update**: the Nov 16 evening windows of `111626_123605` say the same: 48 "the character changed" (trinket swaps) and 3 "weapon skill went up" in 49 windows.
- **Nov 18 update**: `111726_171435` Nov 17: 49 of the 61 windows run say "the character changed" (the same trinket swaps, Smokey's Lighter, Vigilance Charm, Chained Essence of Eranikus, Rune and Jewel of the Guard Captain).
- **Nov 20 update**: `111926_175614` and `112026_111427`: 32 of 36 windows run say "the character changed" (the same trinket swaps: Smokey's Lighter, Chained Essence of Eranikus, Rune and Jewel of the Guard Captain, Vigilance Charm, Bolt's Charged Eyeball, Blackhand's Breadth).
- **Nov 22 update**: `112126_201216` and `112226_113558`: 22 of 36 windows run say "the character changed" (Chained Essence of Eranikus, Jewel of the Guard Captain, Blackhand's Breadth, Barov Peasant Caller trinket swaps). The morning windows of `112226_113558` also swing a low damage weapon (the log's Melee min is 44 to 45) that the sim doesn't, which looks like a real swap and skews `Melee: raw` (see the sheet finding).
- **Nov 25 update**: `112326_183423`, `112426_110702` and `112426_172127`: 16 of 34 windows run say "the character changed" (Jewel of the Guard Captain, Blackhand's Breadth, Chained Essence of Eranikus, Seal of the Dawn, Barov Peasant Caller trinket swaps).
- **Status**: open, first seen 2026-11-13, updated 2026-11-25.

### Mindless Undead take a third less Fire damage

- **Flag**: `Magma Totem (10581): raw` and `Fire Nova (408428): raw`, mob levels 56 and 57 (not NEW, the rows existed).
- **Windows**: `111526_162618 19:48:25-19:50:58` (Stratholme): Magma Totem 60.0 on every hit (30 hits) vs the sim's 89.5, Fire Nova 208.4 and 207.2 vs 306.4 (30 hits). `22:25:53-22:28:45`: Magma Totem 63.6 and 65.7 (32 hits, 62 or 63 on the Mindless Undead and 90 on level 59 to 61 mobs next to them), Fire Nova 195.1 and 194.4 vs 309 to 311. Ramstein the Gorger, Venom Belcher and Bile Spewer in the same fights take the sim's 90 and 313.
- **What we think**: a mob trait, not our character: Mindless Undead (level 56 and 57, Stratholme) take about 0.67x Fire damage, like the odd per mob rows in "Dungeon mobs take about 2 less". It is either a fire resistance that the log counts in raw or a damage taken modifier on the mob. Not a sim error.
- **Next step**: at the next `/log-audit` add a `known.json` entry for these two abilities with this reason, or have logcheck drop Mindless Undead from the fight. Check Flame Shock and Searing Totem on them too.
- **Status**: open, first seen 2026-11-16, last updated 2026-11-16.

### Melee and Stormstrike a few percent high in the Nov 10 morning

- **Flag**: `Stormstrike (17364): raw` and `Melee: raw`, mob levels 43 to 49.
- **Windows**: `WoWCombatLog-111026_075640.txt` 08:26 to 12:32. After the Enchanted Lute fix (see Done) the sheet matches, but Stormstrike still runs +2% to +4% in the log in some windows (09:27: 660.8 vs 638.1 in 19 hits, 652.2 vs 633.2 in 17) and Melee up to +4% (12:20: 441.4 vs 421.2 in 9).
- **What we think**: the Lute (and its 9 Strength and Agility) may have been up during fights where logcheck's Camp Benefits stretches don't reach, or another camp item came and went. Smaller than before, and few hits.
- **Next step**: watch new logs, now with the elixir and Lute fixes. Dig only if it recurs.
- **Status**: open, first seen 2026-11-10, last updated 2026-11-11.

### BayovCore cuts the gear list short

- **Flag**: none now (logcheck works around it).
- **Windows**: history records 194, 198 and 234 (Nov 5) lost the gloves, cloak and main hand; fourteen Nov 7 logins lost the relic (slot 18).
- **What we think**: `isArray` in BayovCore's `Modules/Export.lua` writes the gear table as a list cut at the first gap. `historyGear` in `tools/logcheck/history.go` now keeps the slots a login's list dropped.
- **Next step**: fix `isArray` in the addon so it always writes an object.
- **Status**: open, first seen 2026-11-08, last updated 2026-11-08.

### NPC debuffs on us are spread over the whole window

- **Flag**: `Melee: raw`, level 62 (Vanndar Stormpike, Alterac Valley).
- **Windows**: `112526_195841 20:30:35-20:32:59`: log 288.3 in 11 hits (229 to 401) vs the sim's 366.4 (283 to 408).
- **What we think**: Stonehearth Marshal's Demoralizing Shout (23511) landed on us at 20:32:19.5 and again at :37.7. Our AP dropped from 1663 to 1059 each time. Hits at full AP were 401 to 415, hits under the shout 229 to 267. logcheck applies the shout as a share of the whole window (21%), not at the time of each hit, so the sim's average stays high. Not a sim error.
- **Next step**: have logcheck apply debuffs on us by their real start and end times in the log. Until then, battleground windows with NPC debuffs on us read low on physical damage.
- **Status**: open, first seen 2026-11-26.

### Sheet read with a fishing pole equipped

- **Flag**: `sheet` (Strength +26 to +29, AP +60 to +66, melee crit +5.3 to +6.7%, spell crit +1%).
- **Windows**: `112526_195841 20:14:12-20:16:11` and `22:00:52-22:02:57`.
- **What we think**: we fished between fights. History line 36499 swaps the main hand to the fishing pole (6256), the sheet is read at 36544, and Bonecrusher (18420) comes back at 36546. Line 36798 to 36827 repeats this. Bonecrusher's 26 Strength and 14 crit rating (29 Strength with Kings) are the whole gap. `buildExports` in `tools/logcheck/history.go` carries the last sheet onto later gear.
- **Next step**: skip the sheet check (or say "sheet read with another weapon") when the weapon slots at the sheet's line differ from the window's gear.
- **Status**: open, first seen 2026-11-26.

## Watching

### Searing Totem crits high at level 37 once

- **Flag**: `Attack (6352): crit` (NEW), mob level 37.
- **Windows**: `WoWCombatLog-110826_085019.txt 18:22:24-18:57:33`: 26 of 133 (19.5%) vs the sim's 9.7%.
- **What we think**: noise. Across all of Nov 8 and 9 our Searing Totem bolts crit 290 of 2623 (11.1%) and rank 3 (6351) 76 of 663 (11.5%). The sim gives 9.7% to 10.9%, which is about 1.8 standard deviations from 11.1%.
- **Next step**: none unless it shows again.
- **Status**: open, first seen 2026-11-10, last updated 2026-11-10.

### Flametongue Attack rounds up at level 29

- **Flag**: `Flametongue Attack (10444): amount` (NEW), mob level 29.
- **Windows**: `WoWCombatLog-110626_104707.txt 01:30:50-01:36:08` (Nov 7): rounded up on 3 of 4, the sim's hit of 34.040 expects 0.2. The log puts the hit at 34.77 to 35.00.
- **What we think**: 4 hits only, from a window the imbue cuts split again. It fits a bit of spell power the sim lacks (a buff), like "Spell hits a point or so off the sim".
- **Next step**: none unless it shows again.
- **Status**: open, first seen 2026-11-10, last updated 2026-11-10.

### Sheet Attack power and Dodge a bit high at level 6

- **Flag**: `sheet` (Attack power +0.50, Dodge % +0.12).
- **Windows**:
  - `WoWCombatLog-101626_165126.txt 23:45:55-23:59:16` (`Player-4620-01872808`, level 6). The sheet has AP 97 and Dodge 4.43%, but the sim has 97.5 and 4.55%. Every other sheet row matches, Armor 170 included.
  - `WoWCombatLog-110826_085019.txt 22:42:40-23:22:34` (`Player-6940-0038ED1A`, level 43, the first window on the new Orc shaman base stats). The sim is again a bit high: Attack power +1.50, Dodge % +0.36, Nature power +0.30. The export was taken in Ghost Wolf with Well-Rested, which shouldn't move these. Only 29 swings and no damage row flagged.
  - The three Nov 8-9 windows after it (`23:58:01`, `00:06:11`, `00:14:26`, level 43 to 44) show the same Dodge % +0.28 to +0.32 and Nature power +0.30 in the sim, so the dodge gap is steady, not noise.
- **What we think**: the sim's own base stats at low level, not logcheck. The half point of AP may be the game rounding AP down. The dodge may be the Agility to dodge ratio at level 6 (`sim/core/base_stats.go`). The AP moves Melee raw damage by about 0.1, and our dodge doesn't touch our damage rows.
- **Next step**: see if it comes back at other levels. If it does, check the AP rounding and the dodge per Agility by level.
- **Status**: open, first seen 2026-10-18, last updated 2026-11-09 (steady at level 43 to 44, so the dodge per Agility by level is worth a check).

### Immunes, partial resists and odd mobs at levels 48 to 59

- **Flag**: `immune` and `partial` on Lightning Bolt (15207), Fire Nova (408427, 408428), Magma Totem (10580), Searing Totem (10435), Lightning Shield (26370), Flame Shock (10448) ticks, Stormstrike, Melee. `never` on Cloud of Chained Essence (1300240) ticks and M73 Frag Grenade (13808). `procs` on Searing Totem (10435), `amount` on Lightning Shield (26370) and Magma Totem, `miss` on Earth Shock (10413) (3 of 8), `absorb` on Magma Totem.
- **Windows**: `111126_095013` Nov 11 10:37 to Nov 12 07:41. Most are the same causes as the known entries: Nature-immune mobs (Lightning Bolt 5 of 5 at level 55 in `01:26-02:02`, Magma Totem 8 of 39 at level 54 in `07:20-07:22`), resistant mobs (Lightning Bolt partial 12 of 20 at 50 in `00:26-00:55`, Flametongue Attack 24 of 65 at level 56 in `07:04-07:17`), and the Searing Totem rank 10435 rotation gap, the same as ranks 6350 to 6352. Cloud of Chained Essence and the grenade are from mobs or others, not us.
- **What we think**: noise or the same causes as the `known.json` entries, which don't have these ranks yet. One window each for `miss` and `amount`.
- **Nov 13 update**: the same causes in the new windows. `Fire Nova (408428)` partial in 14 windows, immune in 6 and raw in 8 (`251 to 266` vs the sim's 258 to 259, both ways), `Magma Totem (10581)` partial 5 and immune 2, `Lightning Bolt (15208)` immune 2, partial 2, other 1 (all `111226_132121 16:22-16:40`), Lightning Shield (26363) immune, partial and never/procs 1 each, Earth Shock (10413) immune and absorb 1, `Attack (10435)` partial 10 and absorb 3, `Flame Shock (10448) ticks` absorb and partial 1 each. `Cloud of Chained Essence (1300240)` never (14 windows) is a mob's. The `Attack (10435): procs` flag (84 windows) is the known rotation gap, and `Earthbind (3600)` at `11:51` is known.
- **Next step**: at the next `/log-audit` add `Attack (10435): procs` and the immune and partial rows above to `known.json` with the same reasons, if the user agrees.
- **Status**: open, first seen 2026-11-12, updated 2026-11-13.

### Nov 15 and 16 one-off flags

- **Flag**: `Stormstrike (17364): dodge` 7 of 10 (NEW, level 60), `Attack (10436): immune` 1 of 19 (NEW, level 58), `Flame Shock (10448): partial` 2 of 6, 2 of 7 and 5 of 5 (NEW, levels 57 and 60), `Cloud of Chained Essence (1300240) ticks: absorb` 1 of 6, `Windfury Totem (10610): never` and `procs` (another shaman's totem in 12 and 6 windows).
- **Windows**: `111526_162618 20:47:39-20:52:21` (dodge, Flame Shock partial), `20:58:37-21:06:15` (immune), `22:13:51-22:16:13` (absorb).
- **What we think**: noise. 10 Stormstrikes in one window, one immune hit, and partials on fire-resistant mobs of the same window as the Fiery Weapon partials. The Windfury Totem rows are a party member's totem, which the sim doesn't run.
- **Nov 17 update**: `111626_123605` Nov 16 18:32 to 23:04, same kinds: `Windfury Totem (10610): never` in 8 windows and `procs` in 5 (24 to 56 per 100 white swings, a party member's totem). `Smokey's Lighter (17283): never` in 4 windows (`18:39`, `22:39`, `22:50`, `22:51`, our on-use trinket, which the sim doesn't use). `Attack (10436): procs` in 4 windows (`18:39:24` 66.7 vs 195.4, `22:21:59` 900 vs 104.8, `22:51:58` 150 vs 826.7 and 80 vs 392.4), the Searing Totem rotation gap of `known.json`'s other ranks. `Magma Totem (10581): partial` 6 of 23 and 2 of 9 and `Fiery Weapon (13897): partial` 4 of 7 in `22:25:09` (resistant mobs). `Cloud of Chained Essence (1300240) ticks: raw` 280.0 vs 140 in 16 hits at `22:23:55`, exactly double, so two clouds at once. Nothing new.
- **Nov 18 update**: `111726_171435` Nov 17 (61 windows): `Windfury Totem (10610): never` in 19 windows and `procs` in 13 (22.7 to 46.2 per 100 white swings, avg 32), `Smokey's Lighter (17283): never` in 7, `Attack (10436): procs` in 5 (20.8 to 57.4 vs the sim's 105 to 151, one 350.0 vs 49.7), all the same causes (a party member's totem, our on-use trinket, the totem rotation gap). `Melee: miss` 3 of 31 and 3 of 25 at level 60 (the sim gives 0) is noise. All kinds were seen before (no NEW flags).
- **Next step**: none unless they come back. At the next `/log-audit` add `Attack (10436): procs` and `Smokey's Lighter (17283): never` to `known.json`.
- **Nov 20 update**: `111926_175614` and `112026_111427` (36 windows): `Windfury Totem (10610): never` in many windows (a party member's totem, 24 windows in the ledger), `Smokey's Lighter (17283): never`, `Attack (10436): procs` (17.0 to 17.1 per 100 white swings vs the sim's 81 to 95 at levels 59 and 60 in `22:13:58`), and partial resists on fire-resistant mobs (`Magma Totem`, `Fire Nova`, `Fiery Weapon`, `Flame Shock` ticks, `Attack (10436)`), plus `Fiery Weapon (13897): immune` 3 of 27 at level 60. Same causes as above, no NEW flags beyond Grounded Electroshock and Earth Shock (see their findings).
- **Nov 22 update**: `112126_201216` and `112226_113558` (45 windows run): `Windfury Totem (10610): never` in 18 windows and `procs` in 11 (16.7 to 41.7 per 100 white swings, a party member's totem), `Attack (10436): procs` in 5 (17.9 to 66.7 vs the sim's 65 to 352, the totem rotation gap), `Fiery Weapon (13897): procs` 92 to 125 per 100 vs the sim's 14 to 50 in 3 windows (`11:56`, `12:04`, `12:08`, level 55 and 56, the rotation melees more than the sim), `Windfury Weapon (439440): procs` 85.7 and 150 per 100 in 2 (`11:53`, `12:04`, same cause), `Melee: miss` 2 of 11 and 8 of 42 (level 50 and 59, noise), `Melee: glance` 1 of 11, `Stormstrike (17364): miss` 1 of 13, `Windfury Weapon (439440): miss` 1 of 20, `Cloud of Chained Essence (1300240) ticks: absorb` 1 of 6. All kinds were seen before (no NEW flags). `Smokey's Lighter (17283): never` didn't show.
- **Nov 25 update**: `112326_183423`, `112426_110702` and `112426_172127` (34 windows run): `Windfury Totem (10610): never` in 5 windows and `procs` in 2 (30.0 and 33.3 per 100 white swings, a party member's totem), `Earthbind (3600): never` and `procs` in 2 (known), `Attack (10436): procs` in 1, `Flame Shock (10448) ticks: never` in 1.
- **Status**: open, first seen 2026-11-16, last updated 2026-11-25.

### Frost Shock raw and one Melee miss at level 58 and 59

- **Flag**: `Frost Shock (10473): raw` (level 58), `Melee: miss` (level 59).
- **Windows**: `112326_183423 22:26:06` Frost Shock 589.0 vs the sim's 582.6 (7 hits, inside the sim's 568 to 597 range). `23:15:37` Melee miss 1 of 9 (11.1%) vs the sim's 0.0%.
- **What we think**: noise. Both are small samples (7 and 9). The miss may be a below-skill swing, and Frost Shock fits the fixed-damage drift above.
- **Next step**: none unless it shows again.
- **Status**: open, first seen 2026-11-25, last updated 2026-11-25.

## Done

### Warchief's Blessing left out of the sheet check

All three Nov 25 battleground sheets (`112526_195841`) were 300 health over the sim. Warchief's Blessing (16609) has dummy auras in the client (300, 15, 10), so logcheck read no stats from it. We added it to `buffOverrides` in `tools/logcheck/buffs.go` as 300 health and 10 mana per 5 sec (2026-11-26).

### Shadow Oil replaced Windfury Weapon in logcheck

- **Done**: 2026-11-13. `setImbue` (`tools/logcheck/main.go`) and the audit's window cuts took any imbue cast, so a Shadow Oil cast over Windfury Weapon left the sim with no Windfury Weapon ("the sim never did it", 37.8 to 60.9 procs per 100 swings in the log) and cut the window at the oil. Both now take only shaman imbues (`shamanImbueFromName`). After `-redo` the Nov 13 windows have no Windfury Weapon `procs` or `never` flag. Its raw damage reads 4 to 10% high, like Melee.

### Windfury Weapon and Windfury Totem keep separate cooldowns

- **Done**: 2026-11-13, sim right. Nov 13 log with both: 109 procs of the two within 100 ms of each other, one white swing proced both 53 times, and the imbue's 1.5 sec held (shortest gap 1.619 sec). Rates fit the sim's 20%. In `notes.md` and TASKS.md (`wf_weapon_totem_icd`).

### Maelstrom Weapon now 10 procs a minute at 5/5

- **Done**: 2026-11-13, sim right. After the Nov 10 class tuning a stack comes on 56% of landed hits with a 3.2 sec weapon (294/527) and 59% with a 3.4 sec one (643/1084), not every hit. The sim's 10 PPM since 2026-11-10 gives 53% and 57%. In `notes.md`.

### Searing Totem rank 5 and Scholomance mobs

- **Done**: 2026-11-13. `known.json` has `Attack (10435)` procs, partial and absorb like the other ranks, absorbs on Scholomance mobs (Bone Armor, Shadow Shield) and immunes on Risen Aberration.

### Weakening Disease ticks for 15 under Forever

- **Done**: 2026-11-13. Client values (15 every 2 sec for 14 sec) in the sim. Then the rate (the user's go): 3.9 PPM under Forever (`sim/common/item_effects.go`, Frightskull Shaft), from 21.8% of landed white swings (190/873) and 22.8% of Stormstrikes (61/268) with the 3.4 sec shaft. It was a guessed 0.5. Ticks per 100 swings still flag both ways, because they depend on how long mobs live (`known.json`).

### Windfury Weapon attacks proc Windfury Totem

- **Done**: 2026-11-13 (the user's go). In the Nov 13 log the imbue's attacks proced the totem on 54 of 174 imbue procs (31%) when no totem proc came in the 150 ms before, and on 0 of 55 when one did (the totem's 100 ms cooldown). The sim keyed its imbue attacks under 8232 to 16362, not the log's 439440, so they never rolled the totem. `windfuryWeaponAttackIds` in `sim/core/buffs.go` lets them. The 100 ms cooldown and the extra swing rule already give the 0 of 55.

### Flametongue Attack rank 6 hits about 5.5 less

- **Done**: 2026-11-13 (the user's go). `FlametongueWeaponCut[6]` is 5.5 under Forever. With 7 the Nov 12 and 13 windows read 1.4 low on average, with 5.5 they sit within about 1 (dungeon mobs 2 to 3 lower, see "Dungeon mobs take about 2 less").

### Shadow Oil's Shadow Bolt in the sim and logcheck

- **Done**: 2026-11-13 (the user's go). The sim had the oil but keyed its proc under Classic's 1382, and logcheck never gave it to the sim because the export shows only the imbue as the main hand's temporary enchant. Now the sim keys it 350026 under Forever (Frost Oil 350025), and logcheck takes the last oil or stone cast when the export has none. Proc rate fits (13.8 vs 11.5 per 100 swings, p 0.68). The damage doesn't yet, see "Shadow Oil's Shadow Bolt hits about 11 harder".

### Windfury Totem procs from main hand abilities, not only white swings

- **Flag**: no logcheck flag (asked by the user, 2026-11-12). Every player with the totem's buff, all mob levels.
- **Windows**: the 14 live logs of Nov 6-11 (`110626_104707` to `111126_095013`), 4,737 procs on 93 players (16 warriors, 13 rogues, 17 druids, 19 paladins, 13 hunters, 18 shamans). The beta logs have none. Procs where the ability was the only hit in the 150 ms before it, out of landed casts with the totem up: Cleave 64/308, Heroic Strike 21/109, Maul 32/174, Raptor Strike 16/127, Sunder Armor 57/428, Revenge 33/257, Execute 20/186, Rend 4/39, Sinister Strike 178/1610, Eviscerate 18/206, Wing Clip 6/37, Ghostly Strike 4/15, Kidney Shot 2/19, Mongoose Bite 3/32. Never: Stormstrike 0/4467, Holy Strike 0/1175, Bloodthirst 0/258, Whirlwind 0/292, Shield Slam 0/125, Claw 0/165, Swipe 0/1634, Rake 0/91, Cheap Shot 0/45, Judgements, seal and poison and item procs, off-hand swings (2% of 2,361, each with a main hand swing 80 ms later), Auto Shot 0/697, extra swings 0/3381. White main hand swings 979/6592 (14.9%) on one-weapon players.
- **What we think**: the client decides it. The totem's aura (8515, 10609, 10612) has ProcChance 20 and ProcTypeMask 0x14 (white swings plus melee abilities) with no class mask. The abilities that proc it are exactly those whose spell has Attributes_3 0x400 (needs a main hand weapon) or Attributes_0 0x4 (on next swing). Execute's cast has the flag and its damage spell 20647 doesn't, so the cast counts. A landed hit with no damage counts (Sunder, Rend, Kidney Shot), as the sim already has it (`wf_totem_nodamage`). The sim's rule since 2026-11-08 (`WindfuryTotemBuffAura` in `sim/core/buffs.go`, `ProcMaskMeleeMHAuto` under Forever) leaves out every ability, so warriors, rogues, ferals and hunters get too few procs. The 14.9% white rate is under 20% because "totem up" is a guess (the buff isn't logged on others, so out-of-range time counts). The proc gap of 98 ms fits the 100 ms cooldown. Touches notes.md Shaman "Windfury Totem procs from white main hand swings only".
- **Next step**: needs the user's go. Under Forever, let the totem roll on white main hand swings, on-next-swing attacks and main hand abilities whose client spell has Attributes_3 0x400, and not on Stormstrike, Holy Strike, Bloodthirst, Whirlwind, Shield Slam, Claw, Swipe, Rake, seal procs or Judgements. Then rerun the warrior, rogue and feral report builds. Scripts: `parse.py`, `core.py`, `report.py` in the 2026-11-12 session scratchpad `wf/` (not kept).
- **Done**: 2026-11-12 (the user's go). `WindfuryTotemBuffAura` now rolls under Forever on white main hand swings and the abilities in `sim/core/windfury_totem_procs_forever.go` (313 client spells, made by `tools/sources/wf_totem_procs.py`). The sim's abilities match the logs one for one. 19 report builds moved (warriors, rogues, Survival).

### Camp items from other players and the Enchanted Lute

- **Done**: 2026-11-11. Every item at the campfire buffs everyone sitting there, other players' items too, and Camp Benefits (1229741) doesn't say which items. The Nov 10 morning sheets at level 47 had +9 to every stat and +229 armor, which is the Enchanted Lute's row for levels 40 to 49 (211 armor, 9 stats, 12 resistances; bindonforever.com/camping). `buffs.go` now adds the Lute (`luteBuff`, by level, not with Mark of the Wild) when the export's sheet shows it, the way it adds the Camp Chair's crit. 5 windows found it, and their sheet gaps went back to the usual few points. The other items (First Aid Kit, Fish Bowl, Sharpening Wheel, Incense Candle, Faction Banner, Lodestone, Mana Well) aren't read yet.

### Our own elixirs and scrolls were left out of the replay

- **Done**: 2026-11-11. `windowBuffs` in `tools/logcheck/buffs.go` skipped every buff we put on ourselves, on the idea that the replay casts it. It can't cast elixirs or scrolls, so whenever Greater Intellect (11396) was up the replay lacked 25 Intellect: about 8 spell power through Mental Quickness and 42 attack power through Mental Dexterity. That was the whole "spell hits run about 6 spell power above the sim" gap (Lightning Shield +5.8 spell power over 399 hits with the elixir, -0.1 over 614 without) and most of the Nov 10 Melee and Stormstrike gap (Maraudon +2.5% and +4.3% to -0.5% and +1.0%). A Scroll of Intellect (8098) gave the rest. Now a buff of ours is skipped only when the sim has its spell (`simSpellIDs` in `replay.go`). On the Nov 10 log, raw flags fell from 50 to 28 for Lightning Shield, 39 to 19 for Stormstrike and 27 to 15 for Melee. Blood Fury, level scaling and coefficients were ruled out, and the coefficients match the client.

### Farraki Enforcement hit for 588 until Nov 10 17:30

- **Done**: 2026-11-11. Trolls take double, as the tooltip says (and the comment in `sim/common/item_effects.go`). The 588s were Zul'Farrak and other troll mobs, the 294s after 17:43 were not. Into `known.json` as `raw`, since the sim's mob is never a troll.

### Stolen Power's spell power left out of the sim's character

- **Done**: 2026-11-11. Stolen Power (1318002, +16 spell damage to all schools) was up at export time in the Nov 10 22:28 to 00:50 windows, and the sheet's Nature power was 14 to 15 above the sim. `buffAuras` in `tools/logcheck/buffs.go` now reads aura 13 (spell damage by school mask). After `-redo`: Nature power +1 to +2 (the usual), Lightning Shield 134.0 vs 134.5 at 22:28.

### Maraudon Nature immunes and partial resists

- **Done**: 2026-11-11. Into `known.json`: `immune` for Lightning Shield (26369), Earth Shock (10412) and Lightning Bolt (10392) (earth elementals and oozes, 5% to 62% of hits), and `partial` for Lightning Bolt (10392) (Noxxion, Razorlash, Vile Larva) and Magma Totem (10580) (Sprite Darters).

### Glances far past the 40% cap with a new two-hander

- **Done**: 2026-11-11. Staves at 202, 214 and 225 of 245 glanced 80% to 87% (61 of 72 swings, levels 46 to 50) where the sim gives 40% to 41%. Added to the Need to Verify item on big skill gaps (`notes.md`, glancing keeps rising past the 40% cap). Leveling only.

### Fire Nova rank 1 keeps its full coefficient

- **Done**: 2026-11-10. The sim has given the downranking penalty to heals only since 2026-11-08 (`sim/core/downrank.go`), and the "Forever's downranking penalty" Need to Verify item tracks the heal side. The redo has the log at 40.2, 40.2 and 40.5 against the sim's 40.0 at levels 36, 37 and 39. Level 38 (dungeon, 37.6) and levels 33 to 35 (38.8 to 39.4 vs 40.0 to 40.1) still run a little under the sim.

### Two Forever enchants changed under the same ID

- **Done**: 2026-11-10 (04bdcdb9). `tools/database/forever_enchants.go` takes the client's stats for enchants 41 (+2 Stamina, was +5 Health) and 66 (+3 Stamina, was +1), and keeps their IDs.

### Buffs at export time make the sheet higher than the sim

- **Done**: 2026-11-10 (490e230b). `tools/logcheck/buffs.go` adds the buffs up at export time from the client's spell rows (Well Fed and the Loyal Companion by hand, the Camp Chair's 2% crit only when the sheet shows it). Others' buffs follow the log's applies and removes during the fight. A sheet read in a form isn't flagged. Ledger: `sheet` 227 of 399 windows to 176 of 422, `Melee: raw` 557 to 274, `Stormstrike: raw` 335 to 191. What's left is in "Melee and Stormstrike now run higher in the sim than in the log".

### Evades counted as outcomes the sim never rolls

- **Done**: 2026-11-10 (490e230b). `tallyLog` in `tools/logcheck/main.go` drops EVADE outcomes. `evade` flags went from 20 to 0.

### The first imbue cast in a log doesn't end the window

- **Done**: 2026-11-10 (490e230b). `imbueCuts` in `tools/logcheck/audit.go` reads the imbue before the first cast from its procs. Nov 10 is now 08:26-08:37:49 (Flametongue) and 08:37:49-08:48 (Rockbiter), and the Flametongue never/procs flags are gone.

### Maelstrom Weapon gives a stack on every landed melee hit

- **Done**: 2026-11-10 (the user's go). Our Nov 8-9 logs at every rank gave a stack on 2756 of 2760 landed white swings, 1553 of 1555 Stormstrikes and 428 of 430 Windfury swings, none on 421 misses. `applyMaelstromWeapon` in `sim/shaman/talents.go` now does that under Forever (it was 4 PPM a point), Totem of the Storm gives 50% a bolt, and the `sham_mw_ppm` report assumption is one value. The swing timer already matched: an instant bolt leaves it alone, a bolt with a cast time starts it over.

### Flametongue Weapon rank 5 hits about 7 less than the sim

- **Done**: 2026-11-10 (the user's go). `FlametongueWeaponCut[5]` is 7 (`sim/shaman/flametongue_weapon.go`). The Nov 9 23:46 window now has the sim at 73.8 against the log's 74.0 to 74.4, and the party buffs explain the rest. Still to check on another weapon speed (Stinging Viper 2.8) and at ranks 4 and 6.

### An imbue cut left the next imbue's cast inside the window

- **Done**: 2026-11-10. `imbueCuts` in `tools/logcheck/audit.go` ends the window a second before the cast. The Nov 9 23:46-23:50 window now runs with Flametongue.

### Audit stops on a login record read before the character loaded

- **Done**: 2026-11-09. BayovCore wrote a login at Nov 9 17:49 with no gear, an empty skill table as `{}` and a talents error. `replayHistory` in `tools/logcheck/history.go` now skips such a login and reads `{}` as no skills.

### Windows mixed two imbues

- **Flag**: `Flametongue Attack (10444): procs` (log 35 to 60 per 100 landed white swings, the sim 143 to 155).
- **Done**: 2026-11-09. Window 22:09-23:10 held 45 min of Rockbiter Weapon, then Flametongue Weapon, and the sim used Flametongue for all of it. `imbueCuts` in `tools/logcheck/audit.go` now ends a window when we cast another imbue or rank. On Flametongue alone the log gives 152 of 167 landed swings and 66 of 69 Stormstrikes, which fits the sim. The per 100 white swings row counts Stormstrike procs too, on both sides.

### Partial resists on resistant mobs

- **Done**: 2026-11-09. Searing Totem bolts (6352) and Flame Shock ticks on Fire-resistant whelps and dragonkin, Lightning Shield orbs on Nature-resistant Woodpaw gnolls, Grimtotem Naturalist and Wandering Forest Walker. Into `known.json` as `partial`, since the sim's mob has no resistances.

### Orc shaman base stats a point or two off

- **Flag**: `sheet` (Strength +1 to 2, Agility -1 to 2, Intellect -1 to 2 at most levels).
- **Windows**: Nov 5-8, most windows, small residue left after the buffs.
- **What we think**: Forever's Orc shaman base rows differ from the sim's 1.12 rows (`sim/core/base_stats.go`). It may also explain the level 6 "Watching" finding.
- **Next step**: read the base rows from a few unbuffed sheets at different levels and compare row by row.
- **Done**: 2026-11-08. The sheet rows for levels 2 to 43 are now in `FOREVER_ATTRIBUTES` (`tools/gen_level_base_stats.py`), and `sim/core/base_stats_levels_auto_gen.go` is regenerated from it. `TestForeverRaceBaseAttributes` pins levels 3, 15, 30 and 43. Levels 44 to 60 stay on 1.12 until we have sheets there. The gap grows with level, so level 60 may be about Str +3, Agi -2 to -3 and Int -2 off.

### Spell damage rolls round to nearest, the sim rolled evenly

- **Flag**: `Attack (3606/6350/6351): raw` (Searing Totem), about 0.5 low in the sim.
- **Done**: 2026-11-08 (in the working copy, goldens still to update). The engine never floored: the floor is logcheck's `toWhole`, which matches the log's raw field for steady hits (8.8 shows 8). The game rolls whole points: 3606 (9 to 11) hit 9 / 10 / 11 on 193 / 370 / 203 bolts, and amount equals raw on those, so nothing fractional rides on the roll. 6350 and 6352 show the same half weight ends. `sim.Roll` now rounds its roll to the nearest point, and the spell power bonus comes on top, so logcheck's floor gives the game's raw. Averages stay the same for whole ranges (goldens move by under 0.4%). After `-redo` the 3606 raw flags went from 36 to 5 and 6350 from 39 to 18. What's left is the dungeon mobs (Wailing Caverns, see "Dungeon mobs take about 2 less") and 6351 about 3 high (separate). Melee is unchanged (its own random source): our white swings at one attack power don't form a flat range, so their ends tell nothing.

### Flametongue Attack never fired in the sim

- **Flag**: `Flametongue Attack (10444/29469/29470): never` and `procs`.
- **Done**: 2026-11-08. The sim keys the imbue proc under Flametongue Weapon's rank, the log under Flametongue Attack. `nameAliases` in `tools/logcheck/report.go` now pairs them. After `-redo` the never and procs rows are gone, a raw row is left in 3 windows (the AP and spell power gap above).

### Earth Shock raw about 20% high in the sim

- **Flag**: `Earth Shock (8045/8046/10412): raw`.
- **Done**: 2026-11-08. The game leaves the Stormstrike mark out of the log's raw field (amount over raw is 1.004 with no Stormstrike before it, 1.202 right after one), but the sim put it in the spell's multiplier. `applyStormstrike` in `sim/shaman/stormstrike.go` now adds the mark after the damage calc for Earth Shock, Lightning Bolt and Chain Lightning, with the same damage and threat (shaman tests unchanged). After `-redo` Earth Shock 10412 went from 6 windows to 1 and 8046 cleared.

### Wrong character from history and items

- **Flag**: `sheet`, `Melee: raw` (31 vs 3.0 in `110526_011432 02:53:44`, the sim swung unarmed).
- **Done**: 2026-11-08. logcheck keeps gear slots a history list dropped (`historyGear`), uses the client's enchant stats when the sim's differ, adds Forever random suffixes from the item link's bonus IDs (`ItemBonus`, `RandPropPoints`), and falls back to the export's main-hand imbue when the log has no imbue proc (Rockbiter: AP gap -103.5 to -10.5 in 10:49 Nov 5).

### Rotation and fight differences

- **Done**: 2026-11-08. Into `known.json`: Searing Totem `procs` (the sim's player barely melees) and Earthbind `never` / `procs` (we drop it by hand), plus the Oct 21 Earth Shock 8042 and Lightning Bolt 403 `raw` rows (the log predates hotfix 113380, and the old rank 1 values fit it exactly). Single absorbs, partial resists, immunes (Nature-immune mobs) and evades are noise.

### Audit stops on a history record whose gear is a JSON array

- **Flag**: none. The whole `-audit` run exits before any window runs, so nothing new gets checked.
- **Windows**: all. Error: `~/forever-logs/history/Player-6940-0038ED1A.jsonl:194: json: cannot unmarshal array into Go struct field historyRecord.gear of type map[string]jsontext.Value`. The record is dated 2026-11-05 02:49:32. Its `gear` is `[null,null,null,{...Primitive Mantle...},...]`.
- **What we think**: the addon saves the gear as a Lua table. When the low slots are empty (a new character), the table has no string keys and serializes as an array with nulls, not an object. `historyRecord.Gear` in `tools/logcheck/history.go` is a `map[string]json.RawMessage`, so it rejects the array.
- **Next step**: let `Gear` take both shapes (decode an array into slot keys "1".."n", skipping nulls), or make the addon always write an object. Then rerun `-audit`.
- **Done**: 2026-11-08. `historyGear` in `tools/logcheck/history.go` now takes both shapes (an array becomes slots "1".."n", nulls skipped). The audit then ran, but 4 windows failed with `SpellItemEnchantment.csv: EOF`, because parallel windows downloaded the same wago table and one read it half written. `download` in `tools/logcheck/wago.go` now writes a temporary file and renames it, and `-audit` reruns a window whose last run failed (`tools/logcheck/audit.go`).

### Low level gear missing from the sim's database

- **Flag**: `sheet`, and `Melee: raw` at mob levels 6 and 7.
- **Windows**:
  - `WoWCombatLog-101626_165126.txt 23:45:55-23:59:16` (Bayov-Oom, `Player-4620-01872808`, level 6). The sheet has Armor 170 and AP 97, but the sim has 38 and 48. Melee raw damage is 28.9 avg in 32 hits at level 6 and 30.4 in 18 hits at level 7. The sim's is 3.0 at both levels.
- **What we think**: logcheck built the wrong character. All 9 worn items are left off as "not in the sim's database": Primitive Mantle (154), Ragged Leather Vest (1364), Belt (1369), Boots (1367) and Bracers (1370), Dust-covered Leggings (4921), Battleworn Leather Gloves (4914), Flimsy Chain Cloak (2652) and Primitive Walking Stick (5778). So the sim swings unarmed (3 damage), with too little armor and AP. The sim's item DB leaves out the grey and white starter gear a new character wears. `character.equip` in `tools/logcheck/export.go` skips any item missing from `core.ItemsByID`. Every new character's first windows will flag like this, so their Melee rows don't tell us anything about the sim. The outcome rows and Earth Shock fit.
- **Next step**: When an item is missing from the DB, have `equip` build it from the client's item tables (logcheck already loads them for the wago build). It needs the weapon damage, speed, armor and stats. Another option is to add the low level items to the sim's DB. Until then, the audit could skip a window whose weapon is missing instead of flagging it.
- **Done**: 2026-10-18. gen_db already builds these items, but it writes the grey and white gear and anything below item level 10 to `assets/database/leftover_db.bin`, which only the UI loads. logcheck now adds the leftover items to the sim's item table at startup (`addLeftoverItems` in `tools/logcheck/export.go`). They match the sheet: Kolkar Hammer 5-11 at 2.4 speed, and Armor 170 = 170 at 23:45. With `-redo` the window's `Melee: raw` flag is gone. Only Primitive Mantle (154) is still left off. It's a shirt with no stats.
