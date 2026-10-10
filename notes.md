* Which is better, Flame Shock or Earth Shock when Stormstrike debuff is up?
* Consider rotation with less Fire Nova to be more mana efficient
* Level 30 + 5 and Level 30 PvP: cast Lightning Shield before the pull (at -4.5 s, ahead of Strength of Earth at -3 s) instead of on a free GCD in the fight (the user, 2026-10-08)

# Need to Verify

Mechanics the sim assumes but no source has settled yet. "Beta" marks the ones we can test at level 30 or below.

## Shaman

* Flametongue Weapon's damage per hit at level 60. At level 30 each rank hits a flat amount below the sim on every weapon speed: 0 for rank 1, about 2 for rank 2 and about 5 for rank 3 (see shaman_audit.md 7.3). No database or server core has this cut, so the sim keeps the client's points for every rank for now. At 60, rank 6 should hit 2810 / 100 * speed + 10% of spell power (101.2 on a 3.6 speed weapon). If the per-rank cut keeps growing, it would be about 28 less (`sim/shaman/flametongue_weapon.go`). The cut comes off before Elemental Weapons' 15% (beta, 2026-10-09).
* Beta: do Flametongue Weapon and Flametongue Totem procs trigger Elemental Devastation?
* Does a Windfury Totem extra attack restart the swing timer (the user, 2026-10-10)? The sim restarts it, as in Era: when a Stormstrike procs it mid-swing, the extra swing comes right away and the next swing a full swing after it. The totem needs level 32, and our own Windfury Weapon blocks it, so it takes a shaman with Flametongue or Rockbiter Weapon. Item 4.12.
* The proc cooldowns of Windfury Totem and Windfury Weapon (the user, 2026-10-10). The sim gives both 1.5 sec. For the totem it comes from the SoD sim, the same as the 1.5 sec the totem's attack power buff lasted in Era. The Forever client (70291 and 70338) gives the totem's proc 100 ms, and the SoD and Era client (1.15.9) has the totem as a weapon enchant with no cooldown at all. The imbue's 1.5 sec is the client's (439431, 1500 ms in both clients). Two procs closer than the cooldown rule it out, like a Stormstrike proc and then one from the next white swing. The shortest gap between two procs takes a lot of data, so we wait for many players' logs after release (warcraftlogs, for example). Items 4.11 and 4.12.
* Do off-hand hits proc Windfury Totem and Flametongue Totem (the user, 2026-10-10)? The sim says yes under Forever, because the user expects it and the Forever client ties neither totem's proc to a weapon. We also assume an off-hand Flametongue Totem proc goes by the off hand's speed. It takes a dual wielder in a shaman's group, and Windfury Totem needs level 32. Items 4.12 and 4.13.
* Maelstrom Weapon proc rate (sim: 2 PPM per point). The talent needs level 34+.
* Weapon Mastery (+10% damage with all weapons, the sim doesn't give it). Blizzard's Mage and Shaman deep dive (2026-10-07) says it moved to baseline, with Enhancing Totems, Two-Handed Axes and Maces and Improved Weapon Totems. The other two totem talents are already part of the client's values: Grace of Air is 77 * 1.15 = 89 and Flametongue Totem 489 * 1.12 = 548, and the beta's Flametongue Totem hits showed nothing on top. But the client (1.60.1.70291, wago.tools) has no baseline spell for Weapon Mastery, only the old talent ranks, and ForeverChanges lists it as removed. At level 30 the beta sheet showed no bonus: UnitDamage gave 97.83 to 109.83 with a percent of 1, exactly the 23-35 axe plus 2.7 * 388 / 14. It could still come at a higher level, because the talent was in row 6. Beta: white hits on critters (almost no armor, and the overkill gives the whole hit) above the UnitDamage range would show it.
* wowhead spell details (rate limited last time): Call of Flame 16038 (does it touch Flametongue Totem, and does it add with Improved Fire Nova), Improved Fire Nova 16086, Elemental Weapons 16266, Concussion 16035.
* Revelation (weapon enchant): only shocks trigger it, no ICD, flat proc chance (sim: 7.2%).
* Stacking of the Forever-only elixirs with each other.
* Beta: Spirit regen of the classes we haven't measured (druid, hunter, mage, priest, warlock). The sim gives every class the shaman and paladin formula under Forever: 0.25 mana a second for each of the first 50 Spirit, 0.125 past that. One GetManaRegen reading below 50 Spirit and one above settles a class.
* Rage of the Storm (280604) source. We assume a level 30 shaman quest chain.
* Beta: Stormstrike with a dagger. We give daggers the normalized 1.7 plus 0.3, so 2.0, but only two-handers and one-handers are tested. The test is the same as the one-hander's: a dagger with no Stormstrike bonus, boars of one level, and about 30 normal Stormstrikes and 30 normal white hits. The /run line for AP and weapon damage is in shaman_audit.md 4.1.
* Server batch window (sim: white hits land with their procs 10 ms after the swing, like Classic Era servers). The user thinks Forever uses 10 ms too. A 400 ms window like the 1.12 servers would change how procs line up.
* Beta: the glancing chance against mobs 4 or more levels above us (sim: at most 40%). At level 30 the beta glanced on 66 of 111 white swings against level 34 mobs (59%) and 6 of 8 against level 35. Without the cap the sim would give 50% and 60%. The user waits for more data (2026-10-10). Item 4.6.
* Beta: partial resists on dungeon bosses and elites above our level. Classic gives every NPC above our level 2% average partial resist a level, on our spells that can be partly resisted. The beta showed none in 294 hits on regular mobs 3 to 5 levels up, so the sim is set to drop it under Forever (5.2). But if Forever kept it for bosses or elites only, we haven't seen it yet. Classic rules give about 24% of hits partly resisted on Bloodmage Thalnos (SM Graveyard, level 34) and about 12% on Interrogator Vishas (level 32, our Level 30 target). About 100 hits on either one settles it: Lightning Shield, Searing Totem, Flametongue procs or Flame Shock, not Earth Shock. The user may log in dungeons (2026-10-10). Item 5.2.

## Rogue

* Opportunity x Aggression on Backstab: add or multiply (sim: multiply, 1.12 rules say add).
* Venom x Vile Poisons: add or multiply.
* Murder crit bonus double dip (deliberate in `target.go`).
* Blade Flurry: does the copied hit take the damage modifiers a second time?
* Cutthroat, Puncturing Wounds and Quietus: only rank 1 was seen, the sim scales them linearly.
* Mutilate dagger requirement (the tooltip doesn't repeat it, the sim assumes it still applies).
* Preparation resets only Cold Blood, Shadowstep and Vanish.
* Precision scope.

## Paladin

* Improved Seals: Seal of Righteousness base only vs Seal of Command whole hit.
* Consecrated Ground: all Holy damage and every target?
* Holy Power scope.
* Swift Judgement and Pursuit of Justice are not implemented.
* Beta: Undead paladin base stats. 1.12 has no Undead paladin, so the sim builds the row from the Human paladin and the Undead warrior. A naked level 20 Undead paladin on the beta (2026-10-10) has Str 44, Agi 30, Sta 42, Int 30, Spi 34, and the sim has 41 / 29 / 42 / 30 / 39. Everything built from them (health, mana, AP, crit, dodge) follows the sim's formulas. A level 1 Undead paladin and an Undead level 1 caster would show whether Forever changed the Undead race (the Undead rogue too) or the paladin's growth.

## At 60

Things we can only check on release, because the beta stops at level 30.

* Naked level 60 Orc shaman sheet: attributes, health, mana, attack power, melee and spell crit, dodge, and how much crit one point of Agility and Intellect gives at 60. The level 30 sheet only checks the lower level tables.
* Attack table against a level 63 boss with 300 weapon skill (sim: 8% miss and 9% hit to cap, 6.5% dodge, 40% glancing at 65% damage on average, crit cut by 4.8%). A level 33 mob at level 30 tests the same +3 rules, but with different numbers.
* Spells against a level 63 boss (sim: 17% miss, 6% average partial resist, no crit cut).
* Every shaman rank above what a level 30 knows, and the spells learned after 30 (in 1.12, Windfury Totem and Chain Lightning at 32, Grace of Air Totem at 42). wowhead shows the imbues at levels 62 to 68, so their level 60 values are our own scaling.
* Shaman talents in rows 6 and 7 (Elemental Fury, Maelstrom Weapon, Lava Burst, Rage of the Farseer).
* Blessing of Wisdom at its top rank, and Judgement of Wisdom (1.12 learns it at 38). Both need a paladin past 30.
* Blood Fury at 60. wowhead shows a flat 169 attack power. The beta tooltip at 30 tells us whether that number grows with level.
* The level 60 trinkets in the presets (Hand of Justice, Blackhand's Breadth).
* Raid buffs and totems at their top ranks, at their Forever values.
* The level 60 raid boss: is it level 63, and its armor (sim: 3731).
* The 100% cap on mana regen while casting. Mindfulness 3/3, Improved Stormstrike and Polished Driftwood Icon add up to 108%, and the sim caps it at 100% like cmangos.
