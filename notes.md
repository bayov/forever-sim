* Which is better, Flame Shock or Earth Shock when Stormstrike debuff is up?
* Consider rotation with less Fire Nova to be more mana efficient
* Level 30 + 5 and Level 30 PvP: cast Lightning Shield before the pull (at -4.5 s, ahead of Strength of Earth at -3 s) instead of on a free GCD in the fight (the user, 2026-10-08)

# Need to Verify

Mechanics the sim assumes but no source has settled yet. "Beta" marks the ones we can test at level 30 or below.

## Shaman

* Beta: Elemental Weapons on Windfury Weapon applies once or twice (sim: twice, like SoD). With a 3.6 speed weapon, the Windfury hit minus a white hit should be about 11.8 AP worth with 0 points, 16.6 if it applies once and 23.2 if twice (`sim/shaman/windfury_weapon.go`).
* Beta: do Flametongue Weapon and Flametongue Totem procs trigger Elemental Devastation?
* Beta: Water Shield globe ICD (sim: 3.5 s, the tooltip says "every few seconds").
* Beta: Lightning Shield orb ICD (sim: 3.5 s, the vanilla value is unknown).
* Maelstrom Weapon proc rate (sim: 2 PPM per point). The talent needs level 34+.
* Rage of the Farseer cooldown (sim: 3 min, from client data, the tooltip shows none).
* wowhead spell details (rate limited last time): Call of Flame 16038 (does it touch Flametongue Totem, and does it add with Improved Fire Nova), Improved Fire Nova 16086, Elemental Weapons 16266, Concussion 16035.
* Revelation (weapon enchant): only shocks trigger it, no ICD, flat proc chance (sim: 7.2%).
* Stacking of the Forever-only elixirs with each other.
* Beta: Spirit regen of the classes we haven't measured (druid, hunter, mage, priest, warlock). The sim gives every class the shaman and paladin formula under Forever: 0.25 mana a second for each of the first 50 Spirit, 0.125 past that. One GetManaRegen reading below 50 Spirit and one above settles a class.
* Rage of the Storm (280604) source. We assume a level 30 shaman quest chain.

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
* Beta: Undead paladin base stats. 1.12 has no Undead paladin, so the sim builds the row from the Human paladin and the Undead warrior. The beta shows 34 base Spirit at level 20, the sim has 39. The other four stats are still unchecked.

## At 60

Things we can only check on release, because the beta stops at level 30.

* Naked level 60 Orc shaman sheet: attributes, health, mana, attack power, melee and spell crit, dodge, and how much crit one point of Agility and Intellect gives at 60. The level 30 sheet only checks the lower level tables.
* Attack table against a level 63 boss with 300 weapon skill (sim: 8% miss and 9% hit to cap, 6.5% dodge, 40% glancing at 65% damage on average, crit cut by 4.8%). A level 33 mob at level 30 tests the same +3 rules, but with different numbers.
* Spells against a level 63 boss (sim: 17% miss, 6% average partial resist, crit cut by 2.1%).
* Every shaman rank above what a level 30 knows, and the spells learned after 30 (in 1.12, Windfury Totem and Chain Lightning at 32, Grace of Air Totem at 42). wowhead shows the imbues at levels 62 to 68, so their level 60 values are our own scaling.
* Shaman talents in rows 6 and 7 (Elemental Fury, Maelstrom Weapon, Lava Burst, Rage of the Farseer).
* Blessing of Wisdom at its top rank, and Judgement of Wisdom (1.12 learns it at 38). Both need a paladin past 30.
* Blood Fury at 60. wowhead shows a flat 169 attack power. The beta tooltip at 30 tells us whether that number grows with level.
* The level 60 trinkets in the presets (Hand of Justice, Blackhand's Breadth).
* Raid buffs and totems at their top ranks, at their Forever values.
* The level 60 raid boss: is it level 63, and its armor (sim: 3731).
* The 100% cap on mana regen while casting. Mindfulness 3/3, Improved Stormstrike and Polished Driftwood Icon add up to 108%, and the sim caps it at 100% like cmangos.
