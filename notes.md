* Which is better, Flame Shock or Earth Shock when Stormstrike debuff is up?
* Consider rotation with less Fire Nova to be more mana efficient
* Level 30 + 5 and Level 30 PvP: cast Lightning Shield before the pull (at -4.5 s, ahead of Strength of Earth at -3 s) instead of on a free GCD in the fight (the user, 2026-10-08)

# Need to Verify

Mechanics the sim assumes but no source has settled yet. "Beta" marks the ones we can test at level 30 or below.

## Shaman

* Flametongue Weapon's damage per hit on ranks 4 to 6 (levels 36 and up). At level 30 each rank hits a flat amount below the client's points on every weapon speed: 0 for rank 1, about 2 for rank 2 and 5.0 for rank 3 (see shaman_audit.md 7.3). The sim takes 2 and 5 off ranks 2 and 3 under Forever since 2026-10-10 (the user's go), before Elemental Weapons' 15%. Ranks 4 to 6 keep the client's points until we see them (`FlametongueWeaponCut` in `sim/shaman/flametongue_weapon.go`). At 60, rank 6 would hit 2810 / 100 * speed + 10% of spell power (101.2 on a 3.6 speed weapon) with no cut. If the cut keeps growing by rank ((points at the learn level - 326) / 77 fits ranks 2 and 3), it would be about 28 less. Check each rank at 36, 46 and 56 when the beta opens those levels.
* Does a Windfury Totem extra attack restart the swing timer (the user, 2026-10-10)? The sim restarts it, as in Era: when a Stormstrike procs it mid-swing, the extra swing comes right away and the next swing a full swing after it. The totem needs level 32, and our own Windfury Weapon blocks it, so it takes a shaman with Flametongue or Rockbiter Weapon. Item 4.12.
* The proc cooldowns of Windfury Totem and Windfury Weapon (the user, 2026-10-10). The sim gives both 1.5 sec. For the totem it comes from the SoD sim, the same as the 1.5 sec the totem's attack power buff lasted in Era. The Forever client (70291 and 70338) gives the totem's proc 100 ms, and the SoD and Era client (1.15.9) has the totem as a weapon enchant with no cooldown at all. The imbue's 1.5 sec is the client's (439431, 1500 ms in both clients). Two procs closer than the cooldown rule it out, like a Stormstrike proc and then one from the next white swing. The shortest gap between two procs takes a lot of data, so we wait for many players' logs after release (warcraftlogs, for example). Items 4.11 and 4.12.
* Do off-hand hits proc Windfury Totem and Flametongue Totem (the user, 2026-10-10)? The sim says yes under Forever, because the user expects it and the Forever client ties neither totem's proc to a weapon. We also assume an off-hand Flametongue Totem proc goes by the off hand's speed. It takes a dual wielder in a shaman's group, and Windfury Totem needs level 32. Items 4.12 and 4.13.
* Beta: Frostbrand Weapon's proc rate and damage (sim: 9 procs a minute, so 9 * speed / 60 a landed main hand hit, and rank 2 hits a flat 54 at level 30 plus 10% of spell power, times 1.15 with Elemental Weapons 3/3). The client leaves the proc rate to the server, and we've never used Frostbrand on the beta. About 100 landed hits with Rage of the Storm (49.5% at 9 a minute) show the rate, and the hits show whether rank 2 loses a flat amount like Flametongue Weapon (see shaman_audit.md 7.3). Another shaman's two rank 2 hits fit no cut (raw 63 for (54 + 1.2) * 1.15 = 63.5), but we don't know their level or talents. We might test the later ranks at 60 too (the user, 2026-10-10). Item 7.5.
* Maelstrom Weapon proc rate (sim: 2 PPM per point). The talent needs level 34+.
* wowhead spell details (rate limited last time): Call of Flame 16038 (the beta answered both: it doesn't touch Flametongue Totem, and it multiplies with Improved Fire Nova), Improved Fire Nova 16086, Elemental Weapons 16266. Concussion 16035 is settled by the client data (shaman_audit.md 8.2).
* Revelation (weapon enchant): only shocks trigger it, no ICD, flat proc chance (sim: 7.2%).
* Stacking of the Forever-only elixirs with each other.
* Beta: Spirit regen past 50 Spirit for the classes other than the shaman and paladin. Under 50, every class with mana gave 0.25 a sec per Spirit at level 1, which fits the formula below (2026-10-10). The sim gives every class the shaman and paladin formula under Forever: 0.25 mana a second for each of the first 50 Spirit, 0.125 past that. One GetManaRegen reading below 50 Spirit and one above settles a class.
* Rage of the Storm (280604) source. We assume a level 30 shaman quest chain.
* Server batch window (sim: white hits land with their procs 10 ms after the swing, like Classic Era servers). The user thinks Forever uses 10 ms too. A 400 ms window like the 1.12 servers would change how procs line up.
* Beta: the glancing chance by mob level (sim: 10% plus 2% for each point of the mob's defense over our level times 5, at most 40%, so 0 below our level, 30% at +2 and 40% from +3). At level 30 with Two-Handed Maces at 149 of 150, our white swings glanced on 0 of 175 against mobs 2 to 6 levels below, 1 of 68 at -1 (level 29, 19:58), 0 of 7 at +0, 3 of 33 at +1, 25 of 72 at +2, 26 of 66 at +3 (40%), 102 of 187 at +4 (55%) and 6 of 8 at +5 (2026-10-10). The glance at -1 can't happen in the sim, and +4 is well over its 40% cap. Both fit the same rule on our real weapon skill with no 40% cap: 10% plus 2% a point over 149 gives 2% at -1, 32% at +2, 42% at +3, 52% at +4 and 62% at +5. The sims we use only meet +2 (Vishas) and +3 (the level 60 boss, 40% either way with full skill), so those two matter most. Up to 20:25, +2 (35%) and +3 (39%) fit both rules, which differ by only 2 points there (about 2000 swings a level to tell apart). +1 is low for both (3 of 33, about 1 time in 12 at 20%). The user called it good enough for now (2026-10-10), so the sim stays as it is. Later a fist weapon at Unarmed 85 glanced on all 8 landed swings on a level 29 mob, so glancing goes by our real skill when it's under the most for our level. The sims assume full skill, so that changes nothing. Item 4.6.
* Beta: do Call of Flame and Improved Fire Nova add or multiply on Fire Nova (sim: multiply since 2026-10-10, 1.15 * 1.2 = 1.38 at 60, where adding gives 1.35)? In the client both are the same kind of percent bonus on the spell's damage, and 1.12 adds those. A respec with Call of Flame 3 and Improved Fire Nova 2 settles it. That takes 15 points in Elemental, and Stormstrike fits in the other 15. With 91 spell power Fire Nova rank 2 (SoD's damage spell, 21.4%) hits 165.3 to 184.3 if they add and 168.9 to 188.4 if they multiply, so a few casts on a group of mobs are enough. The respec run (2026-10-10, 49 spell power) had 5 of 41 hits over the add range (raw 173 to 175, add allows up to 172.2), so they multiply. The sim multiplies them since 2026-10-10 (the user's go, while more data comes in). Item 5.6.
* Beta: partial resists on dungeon bosses and elites above our level. Classic gives every NPC above our level 2% average partial resist a level, on our spells that can be partly resisted. The beta showed none in 294 hits on regular mobs 3 to 5 levels up, so the sim is set to drop it under Forever (5.2). But if Forever kept it for bosses or elites only, we haven't seen it yet. Classic rules give about 24% of hits partly resisted on Bloodmage Thalnos (SM Graveyard, level 34) and about 12% on Interrogator Vishas (level 32, our Level 30 target). About 100 hits on either one settles it: Lightning Shield, Searing Totem, Flametongue procs, Flame Shock or Earth Shock (partly resisted under Forever since 2026-10-11, 7.8). The user may log in dungeons (2026-10-10). Item 5.2.
* Beta: how a mob's school resistance works under Forever. The level 29 Elder Cloud Serpents took exactly 20% or 30% of the raw damage off every Nature hit, in 10% steps where Classic uses 25% steps, and fully resisted 3 Earth Shocks (logged as RESIST, not MISS). A Crocolisk took 30% off one Lightning Shield orb (20:55:05). The sim uses Classic's table (25, 50 or 75% off, 75% of resistance over 5 times our level on average), and since 2026-10-11 it partly resists Earth Shock under Forever like our other spells (shaman_audit.md 5.2 and 7.8). Open: the steps and how often each comes, how often a full resist comes and for which spells, whether Frost Shock and Frostbrand (they slow) are binary, and what Spell Penetration and Curse of the Elements take off. A few hundred hits per school on mobs with resistance (Elder Cloud Serpents for Nature, fire elementals for Fire) would tell.
* Beta: Elemental Devastation's size (sim: 3% melee crit a point, so 9% at 3/3, from the client's talent curve). 56 white swings with the buff up crit 14.3% +- 4.7, against 10.4% without, which can't tell 9% from 3%. A sheet reading (UnitStat or /bayov export) within 10 sec of a shock crit settles it (shaman_audit.md 8.6).

## All classes

* Beta: crit and dodge per Agility of the warrior, hunter and warlock. At level 1 the warrior gets 0.2500 a point (sim 0.2572), the hunter 0.2174 crit and 0.4348 dodge (sim 0.1783 and 0.3567), and the warlock 0.15 (sim 0.1249). The paladin, rogue, priest, mage and druid match the sim at level 1 (2026-10-10).

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

## Warrior

* (nothing yet)

## Hunter

* (nothing yet)

## Priest

* (nothing yet)

## Mage

* (nothing yet)

## Warlock

The warlock runs the Forever client tree and spell ranks since 2026-10-12. What the client leaves to the server, and what we assumed for it:

* Soul Siphon's count. The client gives only "4% per each of your other Affliction effects active on the target, up to a maximum increase of 12%" a rank. The sim counts Corruption, the bane, Siphon Life and the curse, snapshots the count when the drain starts, and multiplies it with Improved Drains and the other bonuses. A Drain Life with and without Corruption up (one rank of the talent, combat log) settles the count and whether it adds.
* Decimation's damage part. Only the tooltip says "they deal 3% increased damage" (Shadow Bolt and Searing Pain on a target below 35%), the client's effect has 0 points. The sim gives the two spells 3% a rank against a target in the 35% execute phase. Needs a level 40 Demonology warlock and a mob under 35%.
* Demonic Brand's damage. The 69876 client tooltip says 65 to 68 a hit, 70338 shows a placeholder. The sim has 65 to 68, no spell power, and lets it crit. Needs a level 40 Demonology warlock.
* The demons don't scale with the warlock. Firebolt and Lash of Pain have spell power coefficients in the client (57.1% and 42.9%), and Demonic Knowledge gives the demon spell damage, so Forever may give the demons some of the warlock's spell power. The sim gives them none, as in 1.12, and keeps the 1.12 demon stats and melee damage (server side). A Firebolt with and without a spell power item on the warlock settles it (beta, any level). The melee is the least sourced input of the Level 60 numbers: the Succubus is about 7% of Demonology's damage (66 DPS, 6.2% melee), and nothing we could find (client 70338 DB2, wow.gg's Forever Demonology guide, ForeverChanges) gives Forever's demon melee. If Forever halved it like the spells, Demonology loses 29 at 180 sec and Destruction, Affliction and Affliction damage over time 26 each (measured with the Succubus's physical damage halved); the Felhunter's melee would drop with it, so the Succubus would likely stay the pick. A Succubus's melee hits in a beta log at level 60 settle it.
* The demons take the players' raid buffs, as in 1.12. Forever's hunter pets take none (Blizzard's hunter deep dive, part of the rework that makes them scale with the hunter's gear, `Pet.IgnoresPlayerBuffs`), but no source says the same of demons, which keep their 1.12 stats here. Taking the raid buffs off the demon costs 12 DPS on Destruction and 14 on Demonology at 180 sec (the Succubus goes from 62.5 to 50.3 and from 75.8 to 61.9 DPS; Blessing of Might is already off). If demons turn out to skip player buffs too, set the flag on the warlock pets. A demon's sheet with and without a raid buff settles it (beta, any level).
* The demons' melee. Forever halved the demons' spells (Firebolt 43 to 48 at rank 7, 1.12 85 to 96, Lash of Pain 50, 1.12 99), but their melee damage and stats are server side, so the sim keeps 1.12's (the Succubus hits for 95 to 131 every 2 sec at 60). That makes the Succubus 55 to 60 DPS better than the Imp for the Affliction and Destruction builds, where 1.12 had them closer. A level 30 Succubus's and Imp's damage on the beta (combat log against a training dummy) would show whether the melee was cut too.
* Demonic Embrace's Spirit. The tooltip lost 1.12's Spirit cost, but the client still has a Spirit effect on the talent with the server filling in its value. The sim takes no Spirit. One look at the character sheet with the talent settles it (beta, level 10).
* Two Demonic Sacrifice buffs at once. With Demonic Pact a warlock could sacrifice the Imp, summon the Succubus and sacrifice it too. The sim keeps one buff (a new sacrifice replaces the old one), like 1.12's mutually exclusive buffs. Needs a level 60 Demonology warlock.
* Incinerate's 25% with Immolate up. The client's second effect is a script with 25 in it, the sim takes 25% more of the whole hit, spell power part included. The beta's level 40 Incinerate on a mob with and without Immolate settles it.
* Soul Shards are not counted. Shadowburn and Soul Fire cost one, Decimation and Shadow and Flame give some back. The presets assume the warlock brings enough to a raid fight.
* Improved Shadow Bolt's Shadow Vulnerability counts for every Shadow hit and tick of the warlock (aura 270, damage taken from the caster), not for the demon.
* Wrack's 10% (aura 271 on Corruption and Bane of Agony) multiplies with the other bonuses, and only while the channel runs.
* Rain of Fire and Hellfire keep their 1.12 numbers. Forever's Rain of Fire is an area trigger the client does not fill in. No preset uses either.
* Life Tap costs health the sim does not track unless the warlock is the tank, so the presets tap as much as they like.
* One +2% crit elixir per character. The client presents Mongoose, Grizzly, Owl, Sages and Whale as role variants of one elixir (all item level 55, category 2593), so the presets take one (the orchestrator's ruling for every class): Elixir of Sages with the non-crit Elixir of Greater Intellect, 0.1 to 0.6 ahead of the Owl with Elixir of Greater Spirit. If two can be up at once after all, the Owl on top was worth about +20. Two elixirs on a beta character settle it.


## Druid

* (nothing yet)

## At 60

Things we can only check on release, because the beta stops at level 30.

* Naked level 60 Orc shaman sheet: attributes, health, mana, attack power, melee and spell crit, dodge, and how much crit one point of Agility and Intellect gives at 60. The level 30 sheet only checks the lower level tables. The race offsets too: the sim assumes the Forever Horde race offsets seen at level 1 (2026-10-10) hold at 60, so a naked level 60 Orc shaman, Undead paladin or Undead rogue checks them. The Alliance races and the High Order still have 1.12's or none.
* Attack table against a level 63 boss with 300 weapon skill (sim: 8% miss and 9% hit to cap, 6.5% dodge, 40% glancing at 65% damage on average, crit cut by 4.8%). A level 33 mob at level 30 tests the same +3 rules, but with different numbers.
* Spells against a level 63 boss (sim: 17% miss, 6% average partial resist, no crit cut).
* Every shaman rank above what a level 30 knows, and the spells learned after 30 (in 1.12, Windfury Totem and Chain Lightning at 32, Grace of Air Totem at 42). wowhead shows the imbues at levels 62 to 68, so their level 60 values are our own scaling.
* Does Flametongue Totem rank 2 (levels 38 to 47) proc at all? In the Forever client (70338) its party aura (8250) triggers spell 0, where ranks 1, 3 and 4 name their proc spell. The sim gives rank 2 the usual proc (8248). shaman_audit.md 7.1.
* Lightning Overload's damage (sim: half of the main hit). The Forever client (70338) has its own overload spells, and they deal 0.52, 0.57, 0.52, 0.47 and 0.45 of the main hit for Lightning Bolt ranks 1 to 5, about 0.50 for ranks 6 to 10 and about 0.53 for Chain Lightning. Chain Lightning rank 3's overload has a 0.2855 coefficient, half of 0.571, where the main spell has 0.517, so the 0.517 may be a client typo. No preset takes the talent. shaman_audit.md 7.1.
* Shaman talents in rows 6 and 7 (Elemental Fury, Maelstrom Weapon, Lava Burst, Rage of the Farseer). Lava Burst's 20% with Flame Shock on the target is up to the server, and the sim multiplies it with Call of Flame (1.15 * 1.2 = 1.38 at 3 points, 1.35 if they add, item 5.6).
* Blessing of Wisdom at its top rank, and Judgement of Wisdom (1.12 learns it at 38). Both need a paladin past 30.
* The level 60 trinkets in the presets (Hand of Justice, Blackhand's Breadth).
* Raid buffs and totems at their top ranks, at their Forever values.
* The level 60 raid boss: is it level 63, and its armor (sim: 3731).
* The 100% cap on mana regen while casting. Mindfulness 3/3, Improved Stormstrike and Polished Driftwood Icon add up to 108%, and the sim caps it at 100% like cmangos.
