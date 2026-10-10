* Which is better, Flame Shock or Earth Shock when Stormstrike debuff is up?
* Consider rotation with less Fire Nova to be more mana efficient
* Level 30 + 5 and Level 30 PvP: cast Lightning Shield before the pull (at -4.5 s, ahead of Strength of Earth at -3 s) instead of on a free GCD in the fight (the user, 2026-10-08)

# Need to Verify

Mechanics the sim assumes but no source has settled yet. "Beta" marks the ones we can test at level 30 or below.

## Shaman

* Flametongue Weapon's damage per hit at level 60. At level 30 each rank hits a flat amount below the sim on every weapon speed: 0 for rank 1, about 2 for rank 2 and about 5 for rank 3 (see shaman_audit.md 7.3). No database or server core has this cut, so the sim keeps the client's points for every rank for now. At 60, rank 6 should hit 2810 / 100 * speed + 10% of spell power (101.2 on a 3.6 speed weapon). If the per-rank cut keeps growing, it would be about 28 less (`sim/shaman/flametongue_weapon.go`). The cut comes off before Elemental Weapons' 15% (beta, 2026-10-09).
* Beta: do Flametongue Weapon procs trigger Elemental Devastation? On the beta (2026-10-10) the talent worked on both direct Flame Shock crits. It didn't work on 3 Flametongue Totem crits or on Flame Shock ticks, as in the sim. It didn't work on 5 Fire Nova crits either, and the sim leaves Fire Nova out since 2026-10-10 (the user's go). Flametongue Weapon is still untested.
* Does a Windfury Totem extra attack restart the swing timer (the user, 2026-10-10)? The sim restarts it, as in Era: when a Stormstrike procs it mid-swing, the extra swing comes right away and the next swing a full swing after it. The totem needs level 32, and our own Windfury Weapon blocks it, so it takes a shaman with Flametongue or Rockbiter Weapon. Item 4.12.
* The proc cooldowns of Windfury Totem and Windfury Weapon (the user, 2026-10-10). The sim gives both 1.5 sec. For the totem it comes from the SoD sim, the same as the 1.5 sec the totem's attack power buff lasted in Era. The Forever client (70291 and 70338) gives the totem's proc 100 ms, and the SoD and Era client (1.15.9) has the totem as a weapon enchant with no cooldown at all. The imbue's 1.5 sec is the client's (439431, 1500 ms in both clients). Two procs closer than the cooldown rule it out, like a Stormstrike proc and then one from the next white swing. The shortest gap between two procs takes a lot of data, so we wait for many players' logs after release (warcraftlogs, for example). Items 4.11 and 4.12.
* Do off-hand hits proc Windfury Totem and Flametongue Totem (the user, 2026-10-10)? The sim says yes under Forever, because the user expects it and the Forever client ties neither totem's proc to a weapon. We also assume an off-hand Flametongue Totem proc goes by the off hand's speed. It takes a dual wielder in a shaman's group, and Windfury Totem needs level 32. Items 4.12 and 4.13.
* Maelstrom Weapon proc rate (sim: 2 PPM per point). The talent needs level 34+.
* wowhead spell details (rate limited last time): Call of Flame 16038 (the beta answered both: it doesn't touch Flametongue Totem, and it multiplies with Improved Fire Nova), Improved Fire Nova 16086, Elemental Weapons 16266, Concussion 16035.
* Revelation (weapon enchant): only shocks trigger it, no ICD, flat proc chance (sim: 7.2%).
* Stacking of the Forever-only elixirs with each other.
* Beta: Spirit regen past 50 Spirit for the classes other than the shaman and paladin. Under 50, every class with mana gave 0.25 a sec per Spirit at level 1, which fits the formula below (2026-10-10). The sim gives every class the shaman and paladin formula under Forever: 0.25 mana a second for each of the first 50 Spirit, 0.125 past that. One GetManaRegen reading below 50 Spirit and one above settles a class.
* Rage of the Storm (280604) source. We assume a level 30 shaman quest chain.
* Beta: Stormstrike with a dagger. We give daggers the normalized 1.7 plus 0.3, so 2.0, but only two-handers and one-handers are tested. The test is the same as the one-hander's: a dagger with no Stormstrike bonus, boars of one level, and about 30 normal Stormstrikes and 30 normal white hits. The /run line for AP and weapon damage is in shaman_audit.md 4.1.
* Server batch window (sim: white hits land with their procs 10 ms after the swing, like Classic Era servers). The user thinks Forever uses 10 ms too. A 400 ms window like the 1.12 servers would change how procs line up.
* Beta: the glancing chance by mob level (sim: 10% plus 2% for each point of the mob's defense over our level times 5, at most 40%, so 0 below our level, 30% at +2 and 40% from +3). At level 30 with Two-Handed Maces at 149 of 150, our white swings glanced on 0 of 175 against mobs 2 to 6 levels below, 1 of 68 at -1 (level 29, 19:58), 0 of 7 at +0, 3 of 33 at +1, 25 of 72 at +2, 26 of 66 at +3 (40%), 102 of 187 at +4 (55%) and 6 of 8 at +5 (2026-10-10). The glance at -1 can't happen in the sim, and +4 is well over its 40% cap. Both fit the same rule on our real weapon skill with no 40% cap: 10% plus 2% a point over 149 gives 2% at -1, 32% at +2, 42% at +3, 52% at +4 and 62% at +5. The sims we use only meet +2 (Vishas) and +3 (the level 60 boss, 40% either way with full skill), so those two matter most. Up to 20:25, +2 (35%) and +3 (39%) fit both rules, which differ by only 2 points there (about 2000 swings a level to tell apart). +1 is low for both (3 of 33, about 1 time in 12 at 20%). The user called it good enough for now (2026-10-10), so the sim stays as it is. Later a fist weapon at Unarmed 85 glanced on all 8 landed swings on a level 29 mob, so glancing goes by our real skill when it's under the most for our level. The sims assume full skill, so that changes nothing. Item 4.6.
* Beta: does Flametongue Totem's proc crit chance follow our crit (sim: it rolls our spell crit)? The user wants this settled eventually (2026-10-10). With 12.3% spell crit and 10.4% melee crit on the sheet, it crit on 12 of 127 procs (9.4%), at 1.5 times a normal hit (27 on 18), so it's a spell crit. A flat 5% that ignores our stats is unlikely (1 time in 38), but one crit chance can't show that the rate moves with it. The log names us as its source, like Flametongue Weapon (24 of 198 at our full chance), while Healing Stream's heals come from the totem and crit at about 5% (29 of 605). The test is the same count at a second crit chance, about 500 procs each, for example the Elemental respec without Thundering Strikes (5% less crit). Item 5.3.
* Beta: do Call of Flame and Improved Fire Nova add or multiply on Fire Nova (sim: add, 1 + 15% + 20% = 1.35 at 60, where multiplying gives 1.38)? In the client both are the same kind of percent bonus on the spell's damage, and 1.12 adds those. A respec with Call of Flame 3 and Improved Fire Nova 2 settles it. That takes 15 points in Elemental, and Stormstrike fits in the other 15. With 91 spell power Fire Nova rank 2 (SoD's damage spell, 21.4%) hits 165.3 to 184.3 if they add and 168.9 to 188.4 if they multiply, so a few casts on a group of mobs are enough. The respec run (2026-10-10, 49 spell power) had 5 of 41 hits over the add range (raw 173 to 175, add allows up to 172.2), so they multiply. The sim multiplies them since 2026-10-10 (the user's go, while more data comes in). Item 5.6.
* Beta: partial resists on dungeon bosses and elites above our level. Classic gives every NPC above our level 2% average partial resist a level, on our spells that can be partly resisted. The beta showed none in 294 hits on regular mobs 3 to 5 levels up, so the sim is set to drop it under Forever (5.2). But if Forever kept it for bosses or elites only, we haven't seen it yet. Classic rules give about 24% of hits partly resisted on Bloodmage Thalnos (SM Graveyard, level 34) and about 12% on Interrogator Vishas (level 32, our Level 30 target). About 100 hits on either one settles it: Lightning Shield, Searing Totem, Flametongue procs or Flame Shock, not Earth Shock. The user may log in dungeons (2026-10-10). Item 5.2.

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

## At 60

Things we can only check on release, because the beta stops at level 30.

* Naked level 60 Orc shaman sheet: attributes, health, mana, attack power, melee and spell crit, dodge, and how much crit one point of Agility and Intellect gives at 60. The level 30 sheet only checks the lower level tables. The race offsets too: the sim assumes the Forever Horde race offsets seen at level 1 (2026-10-10) hold at 60, so a naked level 60 Orc shaman, Undead paladin or Undead rogue checks them. The Alliance races and the High Order still have 1.12's or none.
* Attack table against a level 63 boss with 300 weapon skill (sim: 8% miss and 9% hit to cap, 6.5% dodge, 40% glancing at 65% damage on average, crit cut by 4.8%). A level 33 mob at level 30 tests the same +3 rules, but with different numbers.
* Spells against a level 63 boss (sim: 17% miss, 6% average partial resist, no crit cut).
* Every shaman rank above what a level 30 knows, and the spells learned after 30 (in 1.12, Windfury Totem and Chain Lightning at 32, Grace of Air Totem at 42). wowhead shows the imbues at levels 62 to 68, so their level 60 values are our own scaling.
* Shaman talents in rows 6 and 7 (Elemental Fury, Maelstrom Weapon, Lava Burst, Rage of the Farseer). Lava Burst's 20% with Flame Shock on the target is up to the server, and the sim multiplies it with Call of Flame (1.15 * 1.2 = 1.38 at 3 points, 1.35 if they add, item 5.6).
* Blessing of Wisdom at its top rank, and Judgement of Wisdom (1.12 learns it at 38). Both need a paladin past 30.
* Blood Fury at 60. wowhead shows a flat 169 attack power. The beta tooltip at 30 tells us whether that number grows with level.
* The level 60 trinkets in the presets (Hand of Justice, Blackhand's Breadth).
* Raid buffs and totems at their top ranks, at their Forever values.
* The level 60 raid boss: is it level 63, and its armor (sim: 3731).
* The 100% cap on mana regen while casting. Mindfulness 3/3, Improved Stormstrike and Polished Driftwood Icon add up to 108%, and the sim caps it at 100% like cmangos.
