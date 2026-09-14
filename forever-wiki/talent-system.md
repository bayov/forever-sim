# Talent system

## Structure

- Three trees per class, seven rows, 51 points at 60. Same shape as Classic.
- One-point milestone talents at 11, 21 and 31 points, now joined by a new 16-point milestone (row 4). New talents sit on that row in eight of nine classes: Mangle (Feral), Hot Streak (Fire), Raging Blows (Fury), and so on.
- Row N needs 5*(N-1) points in the tree. Row 7 is the 31-point talent.
- Legacy Tree perk "Talented" (5 ranks): gain talent points every level starting at level 9 instead of 10, still capped at 51 total. See [legacy-tree.md](legacy-tree.md).
- Blizzard's framing: some talents are exactly as in 1.12, some adjusted, and talents whose only job was a group buff became baseline (Divine Spirit, Blessing of Kings, Improved Mark of the Wild, Improved Battle Shout). "Improved X" filler was cut heavily (Warlock lost eleven of them). Weak 31-point talents were replaced.
- Every tree gained at least one signature spell from a later expansion at a Classic point cost, in rows 3 to 7: Mutilate, Penance, Prayer of Mending, Lava Burst, Riptide, Arcane Blast, Hot Streak, Ice Lance, Mangle, Berserk, Wild Growth, Incinerate.

## Counts

From the wowforevertalents.com dataset (470 talents across 27 trees, all 470 confirmed from footage). "Changed" includes relocated talents with the same effect.

| Class | Total | New | Changed | Unchanged | 31-point talents (Forever) |
|---|---|---|---|---|---|
| Warrior | 54 | 11 | 31 | 12 | Mortal Strike, Bloodthirst (reworked), Shield Slam (reworked) |
| Paladin | 52 | 20 | 21 | 11 | Light's Vigil (new), Holy Shield (reworked), Twist of Light (new) |
| Hunter | 50 | 15 | 20 | 15 | Bestial Wrath, Sniper Shot (reworked), Lacerating Strikes (new) |
| Rogue | 53 | 10 | 26 | 17 | Venom (new), Adrenaline Rush, Thousand Cuts (new) |
| Priest | 53 | 14 | 27 | 12 | Power Infusion (reworked), Prayer of Mending (new), Shadowform (reworked) |
| Shaman | 50 | 15 | 25 | 10 | Lava Burst (new), Rage of the Farseer (new), Riptide (new) |
| Mage | 54 | 6 | 32 | 16 | Arcane Power (reworked), Combustion (reworked), Ice Barrier (reworked) |
| Warlock | 52 | 23 | 26 | 3 | Drain Hope (new), Demonic Pact (new), Incinerate (new) |
| Druid | 52 | 15 | 23 | 14 | Moonkin Form (reworked), Berserk (new), Wild Growth (new) |

wowforever.quest counts it as 119 new, 263 reworked, 19 relocated, 69 unchanged, 81 Classic talents removed (5 became trainer spells), 14 new 31-point talents. The two sites classify borderline cases differently, the per-talent text in [classes/](classes/) is what matters.

## Baseline moves (talent to trainer spell)

- Warrior: Tactical Mastery (baseline, keeps 10 Rage on stance change, talent Improved Tactical Mastery adds 3 per rank), Victory Rush (new baseline, level 1 Arms). Intercept was not in the demo warrior's spellbook (marked removed by the dataset, unclear if it moved).
- Paladin: Consecration (level 20), Blessing of Kings, Holy Strike (new, level 6), Seal of Fury (new).
- Hunter: Aimed Shot baseline. Aimed Shot and Multi-Shot share a cooldown. Traps usable in combat. New pet bar commands.
- Priest: Shadow Word: Death, Devouring Plague, Fear Ward baseline. Divine Spirit is a trainer spell.
- Shaman: Fire Nova is a spell (Fire Nova Totem removed). Totemic Projection, Totemic Recall, Call of the Elements, Call of the Ancestors are new spells (exact effects of the Call spells not captured).
- Mage: Frostfire Bolt is a trainer spell. Ice Block moved into the Frost tree.
- Warlock: Curse of Agony and Curse of Doom became Banes. Incubus is a separate pet.
- Druid: Nature's Grasp and Omen of Clarity are trainer spells. Revive is a new out-of-combat resurrection. Improved Mark of the Wild baseline.

## Removed and added talents by name

Computed by diffing talent names in this repo's Classic trees (`ui/core/talents/trees/*.json`) against the Forever dataset. Some "removed plus added" pairs are renames with a changed effect (for example Improved Backstab became Puncturing Wounds, Storm Reach became Elemental Reach, Improved Fire Totems became Improved Fire Nova, Lethal Shots became Lethal Attacks, Improved Curse of Agony became Improved Bane of Agony, Fel Stamina became Fel Vitality, Improved Mana Shield became Arcane Shielding). Check the class file before treating one as gone.

### Warrior

- Removed: Axe Specialization, Mace Specialization, Sword Specialization, Polearm Specialization, One-Handed Weapon Specialization, Improved Battle Shout, Improved Demoralizing Shout, Improved Shield Block, Improved Taunt, Tactical Mastery (now baseline).
- Added: Spearing Strike, Bloodthrill, Weaponmaster (Arms), Boundless Rage, Raging Blows, Precision (Fury), Master of Defense, Vanguard, Vitality, Focused Rage, Bastion (Protection), Improved Tactical Mastery.
- Weaponmaster replaces the four weapon specializations. Victory Rush and Tactical Mastery are baseline.

### Paladin

- Removed: Blessing of Kings, Blessing of Sanctuary, Consecration, Improved Blessing of Might, Improved Blessing of Wisdom, Improved Concentration Aura, Improved Devotion Aura, Improved Lay on Hands, Improved Retribution Aura, Improved Seal of Righteousness, Improved Seal of the Crusader, Lasting Judgement, Sanctity Aura.
- Added: Improved Holy Strike, Voice of Truth, Reverence, Purifying Power, Infusion of Light, Divine Precision, Consecrated Ground, Light's Vigil (Holy), Improved Seal of Fury, Sacred Duty, Swift Judgement, Templar's Bulwark, Iron Creed (Protection), Holy Conduit, Sanctified Judgement, Sacred Arbiter, Crusade, Champion of the Light, Instrument of Law, Twist of Light (Retribution), Improved Seals.
- Judgement no longer consumes the Seal. Holy Shock moved from 31 to 21 points with a 10 sec cooldown.

### Hunter

- Removed: Aimed Shot (baseline), Humanoid Slaying, Monster Slaying, Improved Aspect of the Hawk, Improved Eyes of the Beast, Improved Feign Death, Improved Hunter's Mark, Improved Scorpid Sting, Improved Serpent Sting, Killer Instinct, Lethal Shots, Thick Hide, Trap Mastery, Wyvern Sting.
- Added: Focused Fire, Summon Hawk (Beast Mastery), Careful Aim, Rapid Killing, Lone Wolf, Rapid Recuperation, Sniper Shot (Marksmanship), Improved Tracking, Survival Tactics, Predator's Edge, Resourcefulness, Expose Prey, Survivalist's Discipline, Strider Kick, Lacerating Strikes (Survival), Deadly Aspects, Improved Stings, Lethal Attacks.
- Survival rebuilt as a melee and traps tree around Mongoose Bite.

### Rogue

- Removed: Dagger Specialization, Fist Weapon Specialization, Mace Specialization, Sword Specialization, Deadliness, Improved Backstab, Improved Sap, Sleight of Hand.
- Added: Mutilate, Venom, Puncturing Wounds (Assassination), Restless Blades, Hack and Slash (Combat), Dirty Tricks, Improved Distract, Quietus, Cutthroat, Thousand Cuts (Subtlety).
- The four weapon specializations collapsed into Hack and Slash, which also enables axes.

### Priest

- Removed: Divine Spirit (trainer), Force of Will, Healing Focus, Improved Power Word: Fortitude, Improved Prayer of Healing, Improved Vampiric Embrace, Lightwell, Unbreakable Will.
- Added: Power in Light, Twin Disciplines, Holy Precision, Soul Warding, Penance, Renewed Hope, Divine Aegis (Discipline), Binding Heal, Litany of Light, Prayer of Mending (Holy), Improved Mind Flay, Devouring Contagion, Early Demise, Twilight Focus (Shadow).

### Shaman

- Removed: Elemental Mastery, Enhancing Totems, Healing Grace, Improved Fire Totems, Improved Weapon Totems, Lightning Mastery, Nature's Guidance, Parry, Shield Specialization, Storm Reach, Totemic Mastery, Two-Handed Axes and Maces, Weapon Mastery.
- Added: Lightning Overload, Earthbound, Lava Burst, Elemental Alacrity, Elemental Reach, Improved Fire Nova (Elemental), Mental Dexterity, Shamanistic Focus, Mental Quickness, Improved Stormstrike, Maelstrom Weapon, Rage of the Farseer, Spirit Weapons (Enhancement), Mindfulness, Water Shield, Riptide, Natural Grace (Restoration).
- Enhancement now scales with Intellect (Mental Dexterity: AP equal to 33% of Int at rank 1, 3 ranks, higher ranks extrapolated by the dataset. Mental Quickness: spell damage from 15% of Int at rank 1, 2 ranks). Two-handed weapons no longer have a talent. Windfury Weapon on main hand no longer stacks with own Windfury Totem. Rage of the Farseer is a personal 30% haste for 25 sec on 3 min cooldown.

### Mage

- Removed: Improved Arcane Explosion, Improved Arcane Missiles, Improved Fire Blast, Improved Mana Shield, Incinerate (renamed Incineration), Magic Attunement.
- Added: Arcane Geometry, Arcane Blast, Missile Barrage, Arcane Impact, Arcane Shielding, Improved Channeling (Arcane), Hot Streak, Wake of Fire, Incineration (Fire), Ice Lance, Fingers of Frost (Frost).

### Warlock

- Removed: Dark Pact, Devastation, Emberstorm, Fel Intellect, Fel Stamina, Grim Reach, Improved Curse of Agony, Improved Curse of Exhaustion, Improved Curse of Weakness, Improved Drain Life, Improved Drain Mana, Improved Drain Soul, Improved Firebolt, Improved Firestone, Improved Healthstone, Improved Immolate, Improved Lash of Pain, Improved Searing Pain, Improved Spellstone, Improved Subjugate Demon.
- Added: Malediction, Soul Harvesting, Improved Drains, Pandemic, Malevolence, Soul Siphon, Drain Hope, Improved Bane of Agony (Affliction), Demonic Aegis, Demonic Energies, Decimation, Demonic Brand, Improved Felhunter, Demonic Knowledge, Demonic Pact, Fel Vitality (Demonology), Molten Skin, Bane of Havoc, Fire and Brimstone, Shadow and Flame, Incinerate, Agonizing Flames (Destruction).

### Druid

- Removed: Blood Frenzy, Faerie Fire (Feral), Feline Swiftness (renamed Feral Swiftness), Feral Aggression, Improved Enrage, Improved Healing Touch, Improved Mark of the Wild (baseline), Improved Nature's Grasp, Improved Shred (renamed Shredding Attacks), Improved Thorns, Natural Weapons, Nature's Grasp (trainer), Omen of Clarity (trainer).
- Added: Genesis, Nature's Majesty, Nature's Splendor, Balance of Nature, Overgrowth, Eclipse (Balance), Mangle, Predatory Instincts, King of the Jungle, Natural Reaction, Rend and Tear, Berserk, Feral Swiftness, Shredding Attacks (Feral), Gift of the Earthmother, Living Spirit, Wild Growth, Naturalist (Restoration).

## Data quality per talent

Each talent in [classes/](classes/) carries a status:

- NEW: no Classic equivalent. Rank 1 text from footage. Higher ranks often extrapolated (the note says so).
- CHANGED: text differs from Classic. The Classic text is shown next to it.
- unchanged: verified same as Classic in footage.

The footage was of level 38 characters, so numbers inside ability tooltips (Stormstrike mana cost, Lava Burst damage) are level 38 numbers.
