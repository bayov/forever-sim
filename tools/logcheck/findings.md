# Log audit findings

What the log audit found that we haven't settled yet. The `log-audit` skill fills it in every
hour while its session is open, and we go through it together on `/log-audit` (the skill's
sections 2 and 6 have the format).

## Fix in the sim

### Windfury Weapon attacks proc Windfury Totem, the sim's don't

- **Flag**: none (asked by the user, 2026-11-13). Levels 57 to 61, Frightskull Shaft.
- **Windows**: `111326_101810` 12:34 to 14:57, Windfury Weapon and Windfury Totem up together. A totem proc landed 10 to 20 ms after an imbue attack on 47 of 162 imbue procs (29%), 13 of them on imbue procs from Stormstrike, which never procs the totem by itself.
- **What we think**: the imbue's attack is client spell 439440, and it is in the totem's ability table (`sim/core/windfury_totem_procs_forever.go:45`). But the sim's imbue attack carries spell ID 10486 (`WindfuryWeaponSpellId`, `sim/shaman/windfury_weapon.go:35`), so `windfuryTotemRollsOnForever` (`sim/core/buffs.go:2213`) doesn't let it roll the totem. Everything else matches: separate cooldowns, the imbue's 20% / 1.5 sec, the totem's 20% / 100 ms, the totem's extra swing rolling the imbue (26 of 118), +246 and about +445 AP.
- **Next step**: needs the user's go. Let the sim's imbue attack roll the totem under Forever (check it by 439440, or add the imbue's rank IDs to the table), then rerun the Enhancement report builds.
- **Status**: open, first seen 2026-11-13, last updated 2026-11-13.

### Weakening Disease procs about 8 times as often as the sim's

- **Flag**: `Weakening Disease (18633) ticks: procs` and `raw`, levels 57 to 61 (Frightskull Shaft).
- **Windows**: the unattended run first saw it (`111326_101810` 12:07 to 12:39, ticks 14 or 15 vs the sim's 8, 95 to 183 per 100 white swings vs 35 to 45). `111326_101810` 12:07 to 14:57: 259 procs (239 applied or refreshed, 20 resisted) in 37 min of melee. 21.8% of landed white swings (190/873) and 22.8% of Stormstrikes (61/268). Log ticks per 100 swings 112, the sim's 49.
- **What we think**: the client has no rate (ProcChance 101, no PPM row), and the sim's 0.5 PPM is a guess (`sim/common/item_effects.go`, Frightskull Shaft). 22% with a 3.4 sec weapon is about 3.9 PPM. The tick of 15 every 2 sec for 14 sec is the client's and went into the sim on 2026-11-13 (it had 8 for 30 sec).
- **Next step**: needs the user's go. Under Forever about 3.9 PPM (or a flat 22% on main hand hits; one weapon speed can't tell them apart).
- **Status**: open, first seen 2026-11-13, last updated 2026-11-13.

### Stormstrike hits about 1.6x the sim from Nov 11 10:37 to Nov 13 04:59

- **Flag**: `Stormstrike (17364): raw`, mob levels 46 to 59 (114 flags, 49 windows of `WoWCombatLog-111126_095013.txt`, Nov 11 10:37 to Nov 12 07:41). `Melee: raw` stays close on most of them.
- **Windows**: `111126_095013 10:37-10:54` log 544 to 552 vs the sim's 330 to 332 (36 and 21 hits, 1.65x, Staves). `11:02-11:22` 523 vs 337 (33 hits). `13:56-13:58` with Rockbiter 786 vs 503 (7 hits). `23:15-23:19` 814 vs 333 (2.44x), `03:08-03:16` 827 vs 335 (2.47x), both with Shadow Oil and sheet AP about 790 over the sim's. `05:25-05:42` 625 to 639 vs 376 (38 hits). Almost every window sits at 1.55 to 1.70x (Alcaz Willcrusher and The Chief's Enforcer, both two-handers). Only `07:38-07:41` (Rockbiter, level 59 boss) fits: 557.6 vs 523.9 (1.06x, 9 hits). The last window before Nov 11 10:37 with Stormstrike (`075640 00:10:31`, Rockbiter, client tables 70862) fit: 731.8 vs 725.0.
- **What we think**: the sim's Stormstrike takes the 2026-11-10 class tuning (100% weapon damage, 8 sec cooldown, `StormstrikeWeaponDamagePercent` in `sim/shaman/stormstrike.go`), but its own comment says no server row carries it (`SpellEffect 1368308` still 150). 1.5x fits the windows at 1.55 to 1.70x, and Rockbiter windows with a matching sheet AP (13:56, 14:32 to 20:56) are 1.55 to 1.6x too, so the AP gap below isn't the cause. The one fitting window (07:38) could be the tuning going live, or a one-off (the sim's Rockbiter AP was 610 over the sheet's there). Touches the Stormstrike notes in `notes.md`.
- **Nov 13 update**: `111226_132121` Nov 12 13:52 to 18:10 (18 windows) and `111226_230324` Nov 13 04:36 to 04:59 (client tables 70888) are still 1.56 to 1.77x (log 523 to 639 vs the sim's 322 to 369, 5 to 143 hits a window). But every window of `111326_101810` (client tables 1.60.1.70938, Nov 13 10:53 to 12:48) fits: `11:29-11:43` 365 vs 364 (10 hits), `11:47-11:51` 372 vs 361, `12:09-12:19` 389 vs 379 (30), `12:03-12:05` 393 vs 360 and `12:24-12:31` 418 vs 387 (a few hits, Melee raw is +8% to +10% there too). The other windows of that log (`10:53`, `11:44`, `11:51`, `11:56`, `11:59`, `12:22`, `12:35`, `12:39`) had no Stormstrike raw flag.
- **What we think now**: the 100% weapon damage in the sim is the live value, and it went live with the client tables bump 70888 to 70938 between Nov 13 04:59 and 10:53. The 1.5x of Nov 11 to 13 04:59 was the old row. So no sim change is needed.
- **Next step**: nothing to change in the sim. Move it to Done at the next `/log-audit` if the user agrees, and recheck the cooldown (6 or 8 sec) from cast gaps in `111326_101810`.
- **Status**: open, first seen 2026-11-12, updated 2026-11-13 (log now fits).

### Flametongue Attack rank 6 (10444) hits 6 to 8 less than the sim

- **Flag**: `Flametongue Attack (10444): raw`, mob levels 55 to 59 (about 21 windows).
- **Windows**: `111226_132121` Nov 12 14:49 to 17:41: 95 vs 103.3 (5 hits), 94 vs 102.3 (19), 98 vs 106 (21), 101 vs 109 (29), 96.7 vs 105.5, 101 vs 109 (11). `111326_101810` Nov 13 11:45 to 12:32: 104 to 105 vs 111 (16, 13, 16, 7, 7 hits and more), 110 to 111 vs 117 at 12:07 to 12:32 (24, 24, 17 hits), 108.7 vs 117.2 at 12:32. Earlier `111126_095013` Nov 12 00:26 to 07:04 was 2 low at 85 vs 87, and 13 low at 04:01 to 07:04 (85 to 87 vs 99 to 100, a different weapon, not checked).
- **What we think**: the same as rank 5's cut of 7 (see Done, "Flametongue Weapon rank 5 hits about 7 less"). `FlametongueWeaponCut` in `sim/shaman/flametongue_weapon.go` has 0 for rank 6 because we hadn't seen it, and the log's whole numbers (94 to 111, flat in a window) say the cut is about 6 to 8 here too. Which weapon and swing speed the windows used isn't checked (the hit comes from the weapon's swing speed).
- **Next step**: needs the user's go. Set `FlametongueWeaponCut[6]` to about 7 under Forever, then `-redo` these windows and check the weapon speeds fit. Touches the "Flametongue Weapon" line in `notes.md` if it has one.
- **Status**: open, first seen 2026-11-13, updated 2026-11-13.

## Collect more data

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
- **Status**: open, first seen 2026-11-08, last updated 2026-11-08.

## logcheck issues

### Shadow Oil's Shadow Bolt isn't in the sim

- **Flag**: `Shadow Bolt (350026): never` and `procs` (levels 48 to 59, 12 windows from `111126_095013 00:26-00:55` to `07:23-07:33`, 11 to 19 per 100 landed white swings).
- **Windows**: every window with the note "shaman imbue from the log: Shadow Oil" (21 reports).
- **What we think**: Shadow Oil is a weapon oil, and logcheck replays it as an imbue the sim doesn't have, so its proc never fires. Spell damage from it is small, so the other rows hold.
- **Nov 13 update**: 20 more windows (`111226_132121` Nov 12 13:52 to 18:10 and `111326_101810` Nov 13 11:29 to 12:39) flag Shadow Bolt the same way, 9 to 20 per 100 landed swings. Two windows show the imbue pick itself is off. `120333`, `120612`, `120714`, `120939` (Nov 13 12:03 to 12:19) read "Shadow Oil", but the log has Flametongue Attack procs (114 to 150 per 100 swings) and the sim gives 0. `123930` (12:39 to 12:48) reads Shadow Oil while the log has `Windfury Weapon (439440)` procs (40 to 68 per 100 swings) and the sim gives 0. So the imbue can change between oils, Flametongue and Windfury inside a window without a cast logcheck sees, and the audit should not take Shadow Oil as the imbue when the log's procs say otherwise.
- **Next step**: the imbue pick is fixed (2026-11-13, see Done, "Shadow Oil replaced Windfury Weapon in logcheck"), and `known.json` has `Shadow Bolt (350026)` `never` and `procs`. Left: add the oil's proc to the sim if we want the numbers.
- **Status**: open, first seen 2026-11-12, updated 2026-11-13.

### Sheet Attack power far from the sim's on some exports

- **Flag**: `sheet` (Attack power), and `Melee: raw` in the same windows.
- **Windows**: `111126_095013`, sheet vs sim: `10:37-10:54` 1300 vs 610, `23:15-23:47` 1465 to 1508 vs 709 to 714, `00:26-01:04` 1476 vs 733, `03:08-03:43` 1412 to 1505 vs 647 to 720 (Shadow Oil and Flametongue, Ghost Wolf in the export's buffs on most). The other way: `19:21-19:31` sheet 732 vs sim 1337, `07:33-07:41` 869 to 988 vs 1479 to 1570 (Rockbiter, the sim adds its AP and the export doesn't have it). Logged AP in the fights is 1318 to 2311, over the sheet in most. `Melee: raw` 501 vs 317 in `23:15-23:19`, and 506 vs 318 in `03:08-03:16`, but 347 vs 334 in `10:37-10:54`.
- **What we think**: an AP source in the sheet that `buffAuras` in `tools/logcheck/buffs.go` doesn't read (about 700 to 800 AP, up in some exports only), or the export reading the sheet before an AP aura applies. Not found. It skews the Melee rows in the windows listed, not the spells.
- **Nov 13 update**: `111226_132121` Nov 12 13:52 to 18:10 (19 windows) sheet AP is 80 to 109 over the sim's, `111226_230324` `04:36-04:57` (5 windows) 1532 vs 786 (-746, Health -300 too, buffs up: Authority, Well Fed, Strength of Earth), and `111326_101810` `11:44`, `11:45`, `11:47` (+78), `12:22` to `12:35` (+60 to +76) the other way. `11:29-11:43` fits (815 vs 815), but its Nature power is 109 vs 92.5 (-16.5), which fits Lightning Shield 238 vs 233, Lightning Bolt 440 vs 426 and Earth Shock 435 vs 426. Melee raw in `111326_101810` runs +4% to +15% over the sim in 20 windows (like 372 vs 324 at `11:45`, 374 vs 337 at `12:03`, 391 vs 359 at `12:07`, 424 vs 401 at `12:32`), though Stormstrike fits there.
- **Next step**: for the 23:15, 03:08 and Nov 13 04:36 exports, list the auras up on us from the log and the sheet's AP rows (`Export.lua`), then find the one that adds the 700 to 800. For `11:29`, find the +16.5 Nature power (an elixir or buff `buffAuras` doesn't read).
- **Status**: open, first seen 2026-11-12, updated 2026-11-13.

### Trinket swaps and skill ups flag as a character change

- **Flag**: `change` (30 windows of `111226_132121` and `111326_101810`).
- **Windows**: 17 of them say "the character changed ... Vigilance Charm / Chained Essence of Eranikus / Jewel or Rune of the Guard Captain" (on-use trinkets swapped), 12 say "weapon skill went up during the window, Two-Handed Maces 281 to 295 of 295", and one is a swap of Alcaz Willcrusher for Fist of the Damned and nothing.
- **What we think**: the audit treats a trinket swap as a gear change, but the sim's trinkets don't change the rows we compare. With history present, `change` should mean a real change to the damage rows (weapon, stat gear). The skill-up flag is the known "skill goes up as we swing" caveat.
- **Next step**: have the audit leave trinket swaps out of `change` unless the trinket has an effect the sim runs (and say skill ups only once per window).
- **Status**: open, first seen 2026-11-13, updated 2026-11-13.

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

## Done

### Shadow Oil replaced Windfury Weapon in logcheck

- **Done**: 2026-11-13. `setImbue` (`tools/logcheck/main.go`) and the audit's window cuts took any imbue cast, so a Shadow Oil cast over Windfury Weapon left the sim with no Windfury Weapon ("the sim never did it", 37.8 to 60.9 procs per 100 swings in the log) and cut the window at the oil. Both now take only shaman imbues (`shamanImbueFromName`). After `-redo` the Nov 13 windows have no Windfury Weapon `procs` or `never` flag. Its raw damage reads 4 to 10% high, like Melee.

### Windfury Weapon and Windfury Totem keep separate cooldowns

- **Done**: 2026-11-13, sim right. Nov 13 log with both: 109 procs of the two within 100 ms of each other, one white swing proced both 53 times, and the imbue's 1.5 sec held (shortest gap 1.619 sec). Rates fit the sim's 20%. In `notes.md` and TASKS.md (`wf_weapon_totem_icd`).

### Maelstrom Weapon now 10 procs a minute at 5/5

- **Done**: 2026-11-13, sim right. After the Nov 10 class tuning a stack comes on 56% of landed hits with a 3.2 sec weapon (294/527) and 59% with a 3.4 sec one (643/1084), not every hit. The sim's 10 PPM since 2026-11-10 gives 53% and 57%. In `notes.md`.

### Searing Totem rank 5 and Scholomance mobs

- **Done**: 2026-11-13. `known.json` has `Attack (10435)` procs, partial and absorb like the other ranks, absorbs on Scholomance mobs (Bone Armor, Shadow Shield) and immunes on Risen Aberration.

### Weakening Disease ticks for 15 under Forever

- **Done**: 2026-11-13. Client values (15 every 2 sec for 14 sec) in the sim. The proc rate is open under "Fix in the sim".

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
