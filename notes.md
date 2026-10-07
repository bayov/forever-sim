* Which is better, Flame Shock or Earth Shock when Stormstrike debuff is up?
* Consider rotation with less Fire Nova to be more mana efficient

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
