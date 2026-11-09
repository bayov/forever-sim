# Log audit findings

What the log audit found that we haven't settled yet. The `log-audit` skill fills it in every
hour while its session is open, and we go through it together on `/log-audit` (the skill's
sections 2 and 6 have the format).

## Fix in the sim

### Maelstrom Weapon gives a stack on every landed melee hit

- **Flag**: none (logcheck doesn't count the buff). Measured from the logs, ranks 1 to 5.
- **Windows**: Nov 8 20:35 to Nov 9 23:30 (1/5 from Nov 8 20:35, 2/5 from Nov 9 12:20, 3/5 from 15:19, 4/5 from 18:51, 5/5 from 21:49). A stack on 2756 of 2760 landed white swings, 1553 of 1555 Stormstrikes and 428 of 430 Windfury extra swings, the same at every rank. Misses, dodges and parries 0 of 421.
- **What we think**: the server gives a stack on every landed melee hit (client 408498 has ProcChance 100). The rank only sets the cast time cut. The sim's 4 PPM a point (`sim/shaman/talents.go` 556-603) gives 22.6%, 45%, 68% and 90% a hit at ranks 1 to 4 with a 3.39 sec weapon, and under 100% at 5/5 with a weapon faster than 3 sec. Lightning Bolt at 5 stacks matches the sim's 4% per rank per stack (0.91, 0.61, 0.31 sec and instant at 2/5 to 5/5). The rest matches too: 5 stacks, 30 sec from the last stack, all used by Lightning Bolt at cast end, none from spells, no cooldown. Need to Verify item "Maelstrom Weapon proc rate" in notes.md.
- **Next step**: the user's go to make it a 100% proc on landed melee hits (white, Stormstrike, Windfury), then rerun the shaman presets and the leveling plan, and drop the `sham_mw_ppm` report assumption. Totem of the Storm's Lightning Bolt trigger (half the melee chance) would become 50% a bolt.
- **Status**: open, first seen 2026-11-08, last updated 2026-11-09.

### Flametongue Weapon rank 5 hits about 7 less than the sim

- **Flag**: `Flametongue Attack (10444): raw`, every Nov 9 window with rank 5 (mob levels 40 to 46).
- **Windows**: `WoWCombatLog-110926_151906.txt 22:54:03-23:10:55` and `23:11:05-23:26:03` (character level 47, Mograine's Might, 3.4 sec): log 74.0 to 74.6 raw against the sim's 81.1 to 81.5 (81.8 before floor). Level 43 Uldaman mobs show 72, the dungeon mob trait below.
- **What we think**: rank 5 takes a flat cut like ranks 2 and 3 (2 and 5). The sim keeps the client's points for ranks 4 to 6 (`FlametongueWeaponCut` in `sim/shaman/flametongue_weapon.go` 26-32). The gap is 7.1 to 8 raw, so 6.9 to 7.4 before Elemental Weapons' 1.15. Only one weapon speed so far, so flat versus speed scaled isn't settled. shaman_audit.md 7.3 has ranks 2 and 3.
- **Next step**: the user's go to set `FlametongueWeaponCut[5]` to 7. Then check on another weapon speed (Stinging Viper 2.8) and at rank 4 (level 36) and 6 (level 56).
- **Status**: open, first seen 2026-11-09, last updated 2026-11-09.

### Fire Nova rank 1 keeps its full coefficient at 36 to 39

- **Flag**: `Fire Nova (408423): raw`, mob levels 26 to 42, 4 windows.
- **Windows**: `WoWCombatLog-110726_203718.txt 21:55:17` (level 36 to 39, 38 spell power): log 40.2 avg in 56 hits, sim 37.3. Every window from level 26 to 42 has the log at about 40.0.
- **What we think**: full coefficient gives 35.0 + 0.143 x 38 = 40.4, the log. The sim's downranking penalty (`sim/core/downrank.go`, rank learned at 12, so 50% at 39) gives 37.7, the sim. This is the newer low rank log the "Forever's downranking penalty" Need to Verify item asked for, and it says no penalty at a 27 level gap. Windows at levels 18 to 25 have the log about 1.5 below the sim (33.0 vs 34.5), cause unknown.
- **Next step**: the user's ruling on the penalty (drop it, or shrink it). The reduced proc chance part isn't modelled either way.
- **Status**: open, first seen 2026-11-08, last updated 2026-11-08.

### Two Forever enchants changed under the same ID

- **Flag**: `sheet` (Stamina -2, Health -15 from level 12 on).
- **Windows**: every Nov 5-8 window from level 12.
- **What we think**: enchant 41 went from +5 Health to +2 Stamina and 66 from +1 to +3 Stamina in the Forever client. The sim's `db.json` still has the old values. logcheck now takes the client's stats when they differ, so the audit no longer shows it, but the sim's own gear does.
- **Next step**: refresh the enchants from the client (`/update-sources`), or fix the two by hand.
- **Status**: open, first seen 2026-11-08, last updated 2026-11-08.

## Collect more data

### Dungeon mobs take about 2 less spell damage per hit

- **Flag**: `Flame Shock (8050/8053/10447)` ticks and direct `raw`, `Lightning Shield (26365..26369): raw`, `Earth Shock (8045): raw`, only in dungeon windows.
- **Windows**: Nov 5-8 dungeon runs (Wailing Caverns, Shadowfang Keep, Razorfen Kraul, Gnomeregan): Flame Shock rank 2 ticks 12 to 10, rank 4 ticks 26 to 24, direct 54 to 52 and 101 to 99, orbs 24 to 22-23. Open world mobs of the same level, same hour: full damage.
- **What we think**: a per-mob trait, not per dungeon or time. 54 of 135 even-level dungeon mobs, 0 of 144 odd-level dungeon mobs and 5 of 739 world mobs. Same sheet spell power and auras. Guess: a hidden per-mob resistance. See the Need to Verify item under "All classes".
- **Window** `WoWCombatLog-110926_151906.txt 23:11:05-23:26:03` (Uldaman): every level 43 mob (Stonevault Brawler and Geomancer, Earthen Rocksmasher and Sculptor, Deadly Cleft Scorpid, Shadowforge Archaeologist and Darkcaster) takes 2 less: Magma Totem 51 to 52 (157 hits) vs 53 to 54, Flametongue Attack 72 vs 74, Fire Nova 189.8 vs 191 to 192. Level 42 and 44 mobs hit by the same Magma tick take full damage. So odd-level dungeon mobs have it too.
- **Next step**: check the unfiltered logs for those mob GUIDs. Then Flame Shock a dungeon trash mob and an open world mob of the same level, 30 ticks each.
- **Status**: open, first seen 2026-11-08, last updated 2026-11-08.

### Maelstrom Weapon cut at 4 stacks may not be linear

- **Flag**: none (from the cast times, not logcheck).
- **Windows**: Nov 9 21:49 to 23:30 at 5/5: Lightning Bolt at 4 stacks took 0.52 sec avg in 13 casts (0.43 to 0.71, most 0.50 to 0.51). The sim gives 0.30 (80% off). At 4/5 with 4 stacks 0.48 (4 casts) vs the sim's 0.54, at 3/5 0.82 (2) vs 0.78.
- **What we think**: 5 stacks fit the sim exactly, but 4 stacks at 5/5 cut 67%, not 80%. Maybe the stack count at cast start is off by one in our reading, or the cut isn't linear in stacks. It matters little since we cast at 5 stacks.
- **Next step**: count more 2 to 4 stack casts in new logs before changing anything (`mw/cast.py` in the scratch dir did it, worth moving into `tools/combat_log/`).
- **Status**: open, first seen 2026-11-09, last updated 2026-11-09.

### Earth Shock is never partly resisted once it lands

- **Flag**: `Earth Shock (8045/8046/10412)`: `resisted` on landed hits.
- **Windows**: Nov 5-8 on Nature-resistant mobs. Earth Shock had 77 whole hits, 0 partial and 31 RESIST, while Lightning Shield orbs on the same mobs were partly resisted 60 of 63. Druid of the Fang alone: 57 whole, 0 partial, 27 RESIST, orbs 31 of 31 partial.
- **What we think**: the sim gives a landed binary spell a partial roll (`sim/core/spell_resistances.go`, `sim/shaman/earth_shock.go`). The October beta serpents (6 landed, all partial) disagree, so live may have changed.
- **Next step**: the user's ruling (2026-11-08) is to keep monitoring. Count Earth Shock partials on resistant mobs in each new log.
- **Status**: open, first seen 2026-11-08, last updated 2026-11-08.

## logcheck issues

### Buffs at export time make the sheet higher than the sim

- **Flag**: `sheet` (31 of 31 windows), `Melee: raw` (24), `Stormstrike (17364): raw` (13), small spell raw rows like Flame Shock rank 4 +1.
- **Windows**: e.g. 07:42:29 on Nov 5: Stamina -6 (Stamina scroll +4, Rock Skin +2), AP -28.5 (Blessing of Might +25, Well Fed), crit -1.9 (Camp Benefits, whose chair part is +2% crit and isn't in the buff list). Devotion Aura or the Armor scroll explain Armor -150 to -194, Blessing of Kings the big Nov 7 stat gaps. The AP gap explains the white damage gap (AP/14 x 3.3 = +6.7 expected, +6.5 seen), and Stormstrike's formula holds once the AP from white hits is used.
- **Window** `WoWCombatLog-110826_085019.txt 21:00:57-21:17:14` (levels 36 to 43, Mograine's Might): the sheet runs the other way, AP +424 in the sim, because the export was taken in Ghost Wolf (buffs: Well-Rested, Ghost Wolf). The fight still shows the usual gap: Melee raw 386.4 in 31 hits vs the sim's 365.6 at level 41, Stormstrike 581.1 in 15 hits vs 559.9, Flame Shock rank 4 ticks 47 vs 46.3 and direct 179 vs 178. So a sheet taken in a form says nothing about the fight, and the AP for damage rows should come from the window's white hits.
- **Window** `WoWCombatLog-110826_085019.txt 21:17:30-21:34:39` (levels 39 to 41, level 43, Mograine's Might at 209-210 of 215): the export at 21:22:50 had Blessing of Kings, Well Fed, Stoneskin and Windfury Totem up. The sheet is Strength +16, Stamina +12, Intellect +10, Spirit +12 and AP +46 over the sim, which Kings (10%) and Well Fed explain. Melee raw 378.8 in 32 hits vs the sim's 363.1 at level 40, Stormstrike 594.2 in 19 vs 564.0, and the spell rows 1 over (Flame Shock 179 vs 178, Lightning Shield 126 vs 125, about the sheet's 2.4 extra Nature power). Blood Fury shows in the logged AP (1478 to 1626). One glance in 28 swings at level 41 (sim 0%) fits the skill going 209 to 210 in the window, at the edge of the glance threshold.
- **Windows** `WoWCombatLog-110826_085019.txt 23:58:01-00:02:56`, `00:06:11-00:09:28` and `00:14:26-00:16:53` (Nov 8-9, level 43 to 44, mobs 40 to 44): the sheet is AP +65 to +71 over the sim with Camp Benefits, Well Fed and Veteran Adventurer's Loyal Companion up. Melee raw runs 374.6 to 389.9 against the sim's 357 to 361, and Stormstrike 585.3 to 599.0 against 553.6 to 559.7, the usual AP gap.
- **Windows** `WoWCombatLog-110826_085019.txt 00:51:26-01:10:15` and `01:15:02-01:23:24` (Nov 9, level 44, mobs 40 to 45): the first had Camp Benefits and Well Fed (AP +71.5 on the sheet), and the second had another shaman's Strength of Earth and Windfury Totem plus Blessing of Kings (Strength +18, AP +51.5). Melee raw runs 373.9 to 390.7 against the sim's 359 to 370, and Stormstrike 569.1 to 610.8 against 553.0 to 583.3.
- **Window** `WoWCombatLog-110926_151906.txt 23:11:05-23:26:03`: Lightning Bolt 286.5 to 289.3 raw against the sim's 282. Mark of the Wild, Blessing of Kings, Fortitude and Divine Spirit give 23 Intellect, which Mental Quickness 2/2 turns into 6.9 spell power (sheet Nature power 59.0 vs 52.7), so +4.9 a bolt. Before the buffs at 23:11 the bolts were 284.3. Lightning Shield, Earth Shock, Magma Totem and Fire Nova move the same way.
- **What we think**: history records carry `buffs` and the sheet check now lists them, but logcheck doesn't add their stats to the sim. Other players' Blessing of Might mid-window also moves AP (244 to 271 in 06:37).
- **Next step**: add the stats of the buffs up at export time to the sim's character (a table of the common ones), or compare the sheet against the sim plus those buffs, and take the AP for damage rows from the window's white hits.
- **Status**: open, first seen 2026-11-08, last updated 2026-11-09 (hourly run, Nov 9 windows).

### BayovCore cuts the gear list short

- **Flag**: none now (logcheck works around it).
- **Windows**: history records 194, 198 and 234 (Nov 5) lost the gloves, cloak and main hand; fourteen Nov 7 logins lost the relic (slot 18).
- **What we think**: `isArray` in BayovCore's `Modules/Export.lua` writes the gear table as a list cut at the first gap. `historyGear` in `tools/logcheck/history.go` now keeps the slots a login's list dropped.
- **Next step**: fix `isArray` in the addon so it always writes an object.
- **Status**: open, first seen 2026-11-08, last updated 2026-11-08.

### Evades counted as outcomes the sim never rolls

- **Flag**: `Melee: evade`, `Stormstrike (17364): evade`, `Earth Shock (10412): evade` (NEW), mob level 42.
- **Windows**: `WoWCombatLog-110826_085019.txt 23:58:01-00:02:56`: Melee 2 of 23, Stormstrike 1 of 11 and Earth Shock 1 of 9 evaded. The sim gives 0%.
- **What we think**: a mob that leashes back or resets evades everything for a moment. That's the fight, not a combat table outcome, and the sim's target never evades. logcheck counts the evades as attempts against the mob's level.
- **Next step**: have logcheck drop EVADE lines (or the attacks on a mob while it evades) before it builds the outcome rows, in `tools/logcheck`. Until then, read an evade flag as a leash.
- **Status**: open, first seen 2026-11-09, last updated 2026-11-09.

## Watching

### Sheet Attack power and Dodge a bit high at level 6

- **Flag**: `sheet` (Attack power +0.50, Dodge % +0.12).
- **Windows**:
  - `WoWCombatLog-101626_165126.txt 23:45:55-23:59:16` (`Player-4620-01872808`, level 6). The sheet has AP 97 and Dodge 4.43%, but the sim has 97.5 and 4.55%. Every other sheet row matches, Armor 170 included.
  - `WoWCombatLog-110826_085019.txt 22:42:40-23:22:34` (`Player-6940-0038ED1A`, level 43, the first window on the new Orc shaman base stats). The sim is again a bit high: Attack power +1.50, Dodge % +0.36, Nature power +0.30. The export was taken in Ghost Wolf with Well-Rested, which shouldn't move these. Only 29 swings and no damage row flagged.
  - The three Nov 8-9 windows after it (`23:58:01`, `00:06:11`, `00:14:26`, level 43 to 44) show the same Dodge % +0.28 to +0.32 and Nature power +0.30 in the sim, so the dodge gap is steady, not noise.
- **What we think**: the sim's own base stats at low level, not logcheck. The half point of AP may be the game rounding AP down. The dodge may be the Agility to dodge ratio at level 6 (`sim/core/base_stats.go`). The AP moves Melee raw damage by about 0.1, and our dodge doesn't touch our damage rows.
- **Next step**: see if it comes back at other levels. If it does, check the AP rounding and the dodge per Agility by level.
- **Status**: open, first seen 2026-10-18, last updated 2026-11-09 (steady at level 43 to 44, so the dodge per Agility by level is worth a check).

### Spell hits a point or so off the sim at level 44

- **Flag**: `Flame Shock (10447): amount` (NEW), and `Lightning Shield (26369): raw` at mob levels 42 to 45.
- **Windows**:
  - `WoWCombatLog-110826_085019.txt 00:51:26-01:10:15` (level 44): Flame Shock's direct hit is 179 raw on all 8, and its amount rounded up on 4 of 8. The sim's hit is 179.050, so it expects 0.9 round ups, and the log puts the hit at 179.40 to 179.90. Lightning Shield is 126 raw on all 6 hits, but the sim's 125.967 floors to 125.
  - `01:15:02-01:23:24`: Lightning Shield 126 to 128 raw in 38 hits against the sim's 125 to 127. Blessing of Kings was up there, so that window only counts weakly.
  - `01:26:45-01:36:47` and `01:37:51-01:42:21` (level 44, the character changed at 01:26): now the sim is the higher one. Lightning Shield logs 126.0 to 126.8 avg in 63 hits against the sim's 127.1 to 127.2 (127.2 to 128.9 before floor), and Earth Shock 268.8 in 17 hits against 270.0. The sheet still shows Nature power within 0.7 of the sim.
- **What we think**: the sim's spell hits sit 0.03 to 0.85 below the game's, about 0.1 to 4 spell power. The export's sheet shows Nature power within 0.7 of the sim. So it's either a fraction of spell power the sim leaves out, or the game rounding the hit (or the spell power bonus) up where the sim's raw is floored (`toWhole` in `tools/logcheck`). It moves damage by under 0.5%. The sign flipped when the character changed at 01:26, so a gear or talent spell power difference fits better than rounding.
- **Next step**: diff the sim's character against the 01:26 export (what changed, and which item or talent the sim counts for 2 to 4 spell power the game doesn't). Then see if it repeats at other levels and with no buffs. If it does, compare the sim's spell power against the sheet's to the decimal, and check whether round to nearest instead of floor clears the Lightning Shield row on old windows.
- **Status**: open, first seen 2026-11-09, last updated 2026-11-09.

## Done

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
