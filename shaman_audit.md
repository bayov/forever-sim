# Shaman Mechanics Audit

We go over every shaman mechanic the sim models, one at a time. For each one:

1. Summarize what the sim does today, with file references.
2. Cross-check it with ForeverChanges (70245), wowhead Forever, the beta client data and the SoD sim.
3. The user OKs the logic, or we fix it first.
4. Test the logic and its edge cases in our implementation (a Go test or a focused sim run).

Anything no source settles goes to `notes.md` under `# Need to Verify`, marked "Beta" when we can test it at level 30 or below.

Scope: enhancement shaman under the Forever ruleset, at level 60 (raid), Level 30 + 5 (solo Vishas) and Level 30 PvP. Healing and elemental-only mechanics are out of scope, unless an enhancement build takes the talent.

Status: `[ ]` not started, `[?]` summarized and waiting for the user's OK, `[t]` OK'd and waiting for its test, `[x]` OK'd and tested.

## 1. Character stats

- [x] 1.1 Base stats by race and level (Orc shaman at 30 and 60): attributes, base health and mana, base attack power, base melee and spell crit, base dodge
- [x] 1.2 Strength: attack power, block value
- [x] 1.3 Agility: melee crit by level, dodge, armor, attack power (none for a shaman)
- [x] 1.4 Stamina: health
- [x] 1.5 Intellect: mana, spell crit by level
- [x] 1.6 Spirit (the regen formula is 2.2)
- [x] 1.7 Forever gear rules: armor listed twice, hit and crit as flat percent, one hit and one crit stat for melee and spells, item haste, school spell power (fire, nature), bonus weapon damage
- [x] 1.8 Stat multipliers and their order (Blessing of Kings, Ancestral Knowledge, Toughness)
- [x] 1.9 The stats panel: it shows our own Strength of Earth and Grace of Air totems on top of the real stats

## 2. Mana and regen

- [x] 2.1 Mana pool: base mana by level, Intellect to mana
- [x] 2.2 Spirit regen formula, and continuous regen under Forever (no 2 sec ticks)
- [x] 2.3 Five second rule: what starts it, regen while casting (Mindfulness, Polished Driftwood Icon, Improved Stormstrike)
- [x] 2.4 MP5: gear, Blessing of Wisdom, Mageblood, Mana Spring
- [x] 2.5 Mana returns: Judgement of Wisdom on melee hits, Water Shield globes, Mana Tide
- [x] 2.6 Potions and Demonic Rune: amounts, shared cooldown, when the rotation uses them
- [x] 2.7 Mana costs: base cost by rank, cost modifiers (Totemic Focus, Convection, Shamanistic Focus, Clearcasting, Maelstrom Weapon), what happens when we can't pay

## 3. Timing

- [x] 3.1 GCD: 1.5 sec for spells, 1 sec for totems, the 1 sec floor, whether spell haste shortens it
- [x] 3.2 Cast times (Lightning Bolt, Chain Lightning) and what a cast does to the swing timer
- [x] 3.3 How melee haste and cast speed modifiers stack (multiplicative)
- [x] 3.4 Prepull: totems, shield and imbue before the pull, the first swing at pull, totem time across the pull
- [x] 3.5 Reaction and batching delays (if the sim has any)

## 4. Our white attacks

- [x] 4.1 Weapon damage: the damage roll, attack power (AP / 14 * weapon speed), normalized damage for specials
- [ ] 4.2 Swing timer: weapon speed and melee haste (Flurry, Rage of the Farseer, item haste), main hand and off hand sync
- [ ] 4.3 Attack table: one roll, in the order miss, dodge, parry, glancing, block, crit, hit
- [ ] 4.4 Miss: base chance against the target's level, weapon skill, hit from gear, the hit cap
- [ ] 4.5 Enemy dodge and parry: values by level, no parry or block from behind (raid), both from the front (PvP)
- [ ] 4.6 Glancing blows: chance against a level +3 target, the damage penalty, weapon skill
- [ ] 4.7 Crit: chance, the crit cap that glancing and miss make, crit damage
- [ ] 4.8 Weapon skill: base by level, gear skill, its effect on 4.4 to 4.7 (Forever removed the weapon skill racials)
- [ ] 4.9 Enemy armor: the mitigation formula, armor debuffs and their Forever values (Sunder, Expose Armor, Faerie Fire, Curse of Recklessness)
- [ ] 4.10 Physical damage modifiers: flat bonus damage, percent modifiers
- [ ] 4.11 Procs from melee hits: PPM procs (weapon speed), chance procs, which hits can proc what (white, yellow, extra attacks, procs from procs)
- [ ] 4.12 Extra attacks: Windfury Totem, Hand of Justice, how they line up with the swing timer

## 5. Our spells

- [ ] 5.1 Spell hit: base miss against the target's level (17% at +3), hit from gear, the 1% floor
- [ ] 5.2 Resistances: target resistance, the level based resistance, average partial resists vs binary spells
- [ ] 5.3 Spell crit: base, Intellect, gear, the 1.5 crit multiplier
- [ ] 5.4 Spell power: coefficients, school power (fire, nature), spell damage vs spell power under Forever
- [ ] 5.5 Spell ranks: the rank each level knows, the penalty for spells learned below level 20
- [ ] 5.6 How damage modifiers stack: same kind add (the 1.12 rule), different kinds multiply, target debuffs (Curse of the Elements, the Stormstrike mark)
- [ ] 5.7 DoTs: tick timing, what snapshots, refresh, crits on ticks (Flame Shock)
- [ ] 5.8 Abilities that use the melee table (Stormstrike, Windfury attacks, the imbue attacks)

## 6. When the enemy hits us

- [ ] 6.1 When the enemy attacks us at all: the tank setting (Level 30 solo, PvP), the boss behind the tank at level 60
- [ ] 6.2 Enemy damage: weapon damage, attack power, swing speed, parry haste
- [ ] 6.3 Our armor: mitigation, Stoneskin Totem, Devotion Aura
- [ ] 6.4 Our miss, dodge, parry and block: base values, Agility, Anticipation, parry only with Spirit Weapons, block only with a shield, enemy crits and crushing blows
- [ ] 6.5 Spell damage to us: Elemental Warding, resistance auras, the raid damage hits that feed Water Shield
- [ ] 6.6 Health: Stamina, Toughness, Improved Reincarnation, healing (Healing Stream)
- [ ] 6.7 PvP mode: the enemy types, time out of melee range

## 7. Shaman abilities

- [ ] 7.1 Rank tables: every spell's ranks, learn levels and values at 30 and 60 (vs the ForeverChanges spellbook)

Weapon imbues

- [ ] 7.2 Windfury Weapon: proc chance, ICD, two attacks, bonus attack power, which hits proc it
- [ ] 7.3 Flametongue Weapon: fire damage per hit by weapon speed, coefficient
- [ ] 7.4 Rockbiter Weapon: attack power
- [ ] 7.5 Frostbrand Weapon: proc rate, damage
- [ ] 7.6 Imbue rules: the shaman imbue beside an oil or stone (Forever), one imbue per weapon

Strikes and shocks

- [ ] 7.7 Stormstrike: weapon damage, the nature mark (Forever: personal), its charges and duration, cooldown
- [ ] 7.8 Earth Shock
- [ ] 7.9 Flame Shock: direct and DoT parts
- [ ] 7.10 Frost Shock
- [ ] 7.11 Shared shock cooldown

Nature spells

- [ ] 7.12 Lightning Bolt
- [ ] 7.13 Chain Lightning: targets, damage falloff per jump
- [ ] 7.14 Lava Burst

Shields

- [ ] 7.15 Lightning Shield: orbs, damage, ICD, what triggers it
- [ ] 7.16 Water Shield (Forever): free, 15 sec cooldown, globes, mana per globe

Totems

- [ ] 7.17 Totem rules: one per element, durations, the 1 sec GCD, totem health ignored, placing them before the pull
- [ ] 7.18 The weapon totem slot: one totem buff per weapon, our imbue turns off the same kind of totem buff, another shaman's totem in the raid buffs
- [ ] 7.19 Windfury Totem: extra attack, bonus attack power, proc chance, ICD, which hits proc it
- [ ] 7.20 Grace of Air Totem
- [ ] 7.21 Strength of Earth Totem
- [ ] 7.22 Searing Totem: attack rate, damage, coefficient, lifetime
- [ ] 7.23 Magma Totem
- [ ] 7.24 Flametongue Totem: the weapon buff and its own attack (16389)
- [ ] 7.25 Fire Nova (Forever spell): damage, cooldown, targets, needs a fire totem or not
- [ ] 7.26 Mana Spring Totem
- [ ] 7.27 Mana Tide Totem
- [ ] 7.28 Healing Stream Totem
- [ ] 7.29 Utility totems the rotation may place: Stoneskin, Tremor, Windwall

## 8. Talents

Elemental

- [ ] 8.1 Convection
- [ ] 8.2 Concussion
- [ ] 8.3 Elemental Warding
- [ ] 8.4 Reverberation
- [ ] 8.5 Call of Flame
- [ ] 8.6 Elemental Devastation
- [ ] 8.7 Elemental Focus (Clearcasting)
- [ ] 8.8 Elemental Alacrity
- [ ] 8.9 Improved Fire Nova
- [ ] 8.10 Eye of the Storm (not implemented)
- [ ] 8.11 Call of Thunder
- [ ] 8.12 Elemental Reach (not implemented)
- [ ] 8.13 Lightning Overload
- [ ] 8.14 Earthbound (not implemented)
- [ ] 8.15 Elemental Fury
- [ ] 8.16 Lava Burst

Enhancement

- [ ] 8.17 Earth's Grasp (not implemented)
- [ ] 8.18 Thundering Strikes
- [ ] 8.19 Ancestral Knowledge
- [ ] 8.20 Guardian Totems
- [ ] 8.21 Mental Dexterity
- [ ] 8.22 Improved Ghost Wolf (not implemented)
- [ ] 8.23 Improved Lightning Shield
- [ ] 8.24 Elemental Weapons
- [ ] 8.25 Shamanistic Focus
- [ ] 8.26 Anticipation
- [ ] 8.27 Toughness
- [ ] 8.28 Flurry
- [ ] 8.29 Stormstrike
- [ ] 8.30 Spirit Weapons
- [ ] 8.31 Mental Quickness
- [ ] 8.32 Improved Stormstrike
- [ ] 8.33 Maelstrom Weapon
- [ ] 8.34 Rage of the Farseer

Restoration

- [ ] 8.35 Improved Healing Wave (not implemented)
- [ ] 8.36 Totemic Focus
- [ ] 8.37 Mindfulness
- [ ] 8.38 Natural Grace
- [ ] 8.39 Tidal Focus
- [ ] 8.40 Improved Reincarnation
- [ ] 8.41 Ancestral Healing (not implemented)
- [ ] 8.42 Healing Focus (not implemented)
- [ ] 8.43 Water Shield
- [ ] 8.44 Tidal Mastery
- [ ] 8.45 Restorative Totems
- [ ] 8.46 Mana Tide Totem
- [ ] 8.47 Healing Way (not implemented)
- [ ] 8.48 Nature's Swiftness
- [ ] 8.49 Purification
- [ ] 8.50 Riptide (not implemented)

Tree rules

- [ ] 8.51 Points per level, the bonus points (Level 30 + 5), tier and prerequisite rules in the talent search

## 9. Race, gear and consumables

- [ ] 9.1 Orc racials under Forever: Blood Fury, Axe Specialization (crit with an axe)
- [ ] 9.2 Relics: Burning Totem, Totem of the Storm (Forever), Polished Driftwood Icon, Totem of Rage, Totem of the Storm (Classic)
- [ ] 9.3 Rage of the Storm (Stormstrike +10%)
- [ ] 9.4 Set bonuses: Enhancement Synthetic (Phase 1), the real sets in the phase gear
- [ ] 9.5 Trinkets in the presets: Hand of Justice, Blackhand's Breadth
- [ ] 9.6 Enchants: Revelation, Crusader, Impact, stat enchants, Forever enchants
- [ ] 9.7 Engineering: Goblin Sapper Charge, Goblin Land Mine, the shared explosive cooldown
- [ ] 9.8 Elixirs, flasks, food, scrolls and juju: what stacks with what, the Forever-only elixirs
- [ ] 9.9 Weapon oils and stones beside the shaman imbue

## 10. Raid buffs and debuffs (Forever values)

- [ ] 10.1 Blessings: Might, Kings, Wisdom
- [ ] 10.2 Other raid buffs: Battle Shout, Trueshot Aura, Leader of the Pack, Moonkin Aura, Retribution Aura, Sanctity Aura, Arcane Intellect, Divine Spirit, Fortitude, Gift of the Wild, Blood Pact, Devotion Aura, resistance auras, scrolls
- [ ] 10.3 Another shaman's totems in the raid buffs (Enhancing Totems is gone)
- [ ] 10.4 Debuffs: Curse of the Elements, Curse of Recklessness, Curse of Shadow, Sunder Armor, Expose Armor, Faerie Fire, Judgement of Wisdom, Light and the Crusader, Insect Swarm, Thunder Clap, Demoralizing Shout and Roar, Scorpid Sting

## 11. Encounter and rotation

- [ ] 11.1 Encounter: target level, armor, health, duration variation, execute phases, more targets (Fire Nova, Magma Totem, Chain Lightning)
- [ ] 11.2 Shaman APL values (totem timers, shield charges and the like)
- [ ] 11.3 APL edge cases: a swing due during a cast, replacing a totem early, Fire Nova on Clearcasting, mana thresholds
- [ ] 11.4 Movement and range in PvP (time out of melee range)

## Findings

Fixes and open questions found along the way, by item.

### 1.1 Base stats

- Fixed (2026-10-07): the generator takes base mana from the client table for every class, and the Orc shaman's level 1 and 30 attributes from the beta (`tools/gen_level_base_stats.py`). The client table also agrees with the hand-checked level 60 mana of the mage (1213) and the priest (1376), where cmangos was 60 off at most levels. `sim/shaman/enhancement/base_stats_test.go` pins a naked Orc shaman at levels 1, 30 and 60. Troll and Tauren shamans at 30 still use the 1.12 rows.
- The stats panel adds our own Strength of Earth (+53) and Grace of Air (+89) as build-phase auras. Display only.
- Every character shows 5% parry and 5% block in the panel (`character.go` addUniversalStatDependencies). Item 6.4 checks whether the attack table uses them for a shaman without Spirit Weapons or a shield.
- Forever calls crit "one stat across melee, ranged and spell". The beta sheet still shows two numbers, melee 0.38% and spell 3.93%, so the sim is right to keep base melee crit (1.7% + Agility) apart from base spell crit (2.3% + Intellect).

Beta sheet, naked level 30 Orc shaman with no talents (2026-10-07). These come from the panel and its tooltips. The user warned that beta tooltips can be wrong, so we change nothing based on them without asking.

| | Beta | Sim |
|---|---|---|
| Str / Agi / Sta / Int / Spi | 51 / 30 / 51 / 46 / 53 | 49 / 31 / 52 / 45 / 56 |
| Health (base) | 665 (335) | 675 (335) |
| Mana (base) | 1075 (665) | 1113 (718) |
| Attack power | 142 | 138 |
| Melee crit | 0.38% shown, 4.30% before the unarmed skill penalty | 4.39% |
| Spell crit | 3.93% | 3.90% |
| Dodge | 4.3% | 4.39% |
| Armor | 60 | 62 |

- The Lua API (not the tooltips) gives the same numbers: UnitStat 51 / 30 / 51 / 46 / 53, 665 health, 1075 mana, 142 AP, 60 armor, melee crit 0.384, fire and nature spell crit 3.933, dodge 4.304, 0 melee and spell hit. GetParryChance is 0 and GetBlockChance is 5 with no shield. The sim shows 5% parry for everyone, so item 6.4 checks that our attack table doesn't use it. UnitDefense doesn't exist in this client.
- Fresh level 1 Orc shaman on the beta: 24 / 17 / 22 / 20 / 22 (1.12 table: 24 / 17 / 23 / 18 / 25), 67 health, 75 mana, 30 AP, 4.9% spell crit, 4.3% dodge. Health and AP fit the formulas. Base mana is 55 (the 1.12 table has 53). Spell crit fits the sim's level 1 curve (4.875%). Dodge fits too once the untrained defense (1 of 5) takes 4 × 0.04% off the sim's 4.50%.
- Stamina is 1 lower and Spirit 3 lower at both levels, so Forever changed the starting stats. The other stats drift differently between level 1 and 30, so the per-level growth changed too. The level 60 row can't be guessed from this.
- Base mana: the client's base mana table, which we already have in `assets/db_inputs/basestats/octbasempbyclass.txt`, has 55 at level 1 and 665 at level 30, the same as the beta. The original WoW table on warcraft.wiki.gg (Base_mana) agrees. The generator reads base mana from the cmangos table instead, which has 53 and 718. For the shaman, cmangos differs from the client table at levels 1, 22, 26, 30 and 52. For the paladin, hunter, warlock and druid only level 1 differs. For the mage and priest almost every level differs.
- Early note for 2.2: GetManaRegen gives 12.876 a second at level 30 (53 Spirit, 46 Intellect) and 5.501 at level 1 (22 Spirit, 20 Intellect) while not casting, and 0.001 while casting at both. The 1.12 formula (15 + Spirit / 5 per 2 sec) gives 12.8 and 9.7. So Forever's Spirit regen isn't the 1.12 formula, at least at low level. The 0.001 looks like the modern engine's floor.
- Every formula matches: 2 AP per Str, 1 block value per 20 Str (rounded down), 0.0868% crit and dodge per Agi at 30, 2 armor per Agi, 1 health for each of the first 20 Sta then 10, 1 mana for each of the first 20 Int then 15, 0.0355% spell crit per Int at 30, base crit 1.7% melee and 2.3% spell, AP base 2 × level − 20, Spirit regen (15 + Spi / 5) per 2 sec = 64 per 5 sec at 53 Spi.
- Mismatch: the attributes. The sim's level 30 row comes from the cmangos 1.12 table (`tools/gen_level_base_stats.py`). The beta differs by 1 to 3 points in every stat.
- Mismatch: base mana at 30. The 1.12 table has 718 at level 30, above level 31's 699, and the beta has 665. That is right between the table's 631 at level 29 and 699 at level 31, so the 718 looks like a bad row in the table.
- The beta's melee crit loses 0.04% per point of weapon skill under 5 × level. With unarmed at 52 of 150 that is 3.92%, which takes 4.30% down to 0.38%. Item 4.8.
- The crit tooltip says "Most periodic effects can critically strike". Item 5.7 (Flame Shock ticks).
- Melee crits deal double damage and spell crits 1.5 times. Items 4.7 and 5.3.
- The armor tooltip matches the 1.12 formula: 60 / (60 + 400 + 85 × 30) = 1.99%. Item 6.3.
- Creature crits deal double damage. Creatures 3 or more levels above us can crush for 150%. Item 6.4.
- Players dodge only from the front. Creatures dodge from any direction. Item 4.5.

### 1.2 to 1.6 Attributes

Geared level 30 Orc shaman on the beta (2026-10-08): 106 Str, 43 Agi, 106 Sta, 137 Int, 83 Spi. Every rate matches the sim. 212 AP and 5 block value from Strength, 3.7% crit and dodge and 86 armor from Agility, 880 health from Stamina, 1775 mana and 4.9% spell crit from Intellect, 83 mana per 5 sec from Spirit. Melee crit 5.39% (5.43% less 0.04% for a 149 of 150 mace skill), spell crit 7.16%, dodge 5.43%, armor 613 for 17.20% (the 1.12 formula against level 30). `base_stats_test.go` TestOrcShamanAttributeRates gives a naked Orc the same bonuses and checks all of it.

- The beta's spell crit per Intellect at 30 is exactly 0.0355%. The sim scales the level 60 rate along the client curve and gets 0.03554%, 0.006% crit too much at 137 Intellect. Too small to matter, so we left it.
- For 4.1: the two handed mace shows 172 to 225 at 3.6 speed with 252 AP. 252 / 14 × 3.6 = 64.8 on top of the weapon's own damage. We need the weapon and its enchant to check the rest.
- For 1.7 and 5.4: the sheet shows 49 spell damage and 59 spell healing. 1.7 below explains the 10 point gap.

### 1.7 Forever gear rules

The beta (2026-10-08) gave 49 for GetSpellBonusDamage on nature and fire, and 59 for GetSpellBonusHealing, from six items and three enchants. Berylline Pads 7, Green Silk Armor 9, Skirmisher's Leather Belt 7, Totemic Leather Leggings 8 and Spidersilk Boots 7 say "damage and healing". Naga Battle Gloves say "healing by up to 15 and damage by up to 5". The Mystic Heavy Armor Kit (legs) adds 4 and the Mystic Medium Armor Kit (feet) adds 2.

- The sim has the same three stats. SpellPower counts for damage and healing, SpellDamage for damage only and HealingPower for healing only (`spell_result.go` GetSchoolDamage and HealingPower). The gloves are stored as 5 SpellPower and 10 HealingPower, so the sim gets 49 and 59 too. `spell_power_test.go` TestOrcShamanSpellPower equips the same gear and checks Lightning Bolt, Flame Shock and Healing Stream Totem, plus 431 armor.
- Hit and crit: the user says nearly every Forever item with hit or crit means both melee and spell (and healing). The sim already makes hit and crit from gear count for both. We had no hit or crit item on the beta to test it.
- Armor adds up normally (user). The test above checks the items' armor plus the kits.
- Crit from consumables and buffs is still in its own pool (Elixir of the Mongoose, Leader of the Pack, Moonkin Aura). That moves to items 9.8 and 10.2.
- Item haste and bonus weapon damage move to 4.1, 4.2 and 4.10.
- The sim has no Healing Wave, Lesser Healing Wave or Chain Heal. Healing Stream Totem is the only shaman heal.

### 1.8 Stat multipliers

Beta, geared level 30 Orc shaman with Ancestral Knowledge 5/5 (2026-10-08): UnitStat gives 150 Intellect and UnitPowerMax gives 2635 mana. Without the talent it was 137. 137 × 1.10 = 150.7, so the talent multiplies base and gear Intellect together and the server drops the fraction. The mana agrees, because 665 + 20 + 130 × 15 = 2635. The tooltip shows 5.3% spell crit, and 150 × 0.0355 = 5.33%.

With Blessing of Kings on top the beta showed 165 Intellect and 2860 mana. 137 × 1.10 × 1.10 = 165.77, so Kings and the talent multiply (adding them would give 164), and mana again counts the whole 165.

- Fixed (2026-10-08): the sim kept the fraction (150.7 Intellect and 2645.5 mana). Under Forever we now drop the fraction from all five attributes after their multipliers and before anything converts them (`stats/deps.go` FloorAttributes, set in `environment.go`). An attribute change mid-fight recomputes the whole stat sheet, because the change on its own can't be rounded (`unit.go` AddStatsDynamic). Classic keeps fractions, as upstream does.
- `stat_multipliers_test.go` TestOrcShamanStatMultipliers checks both beta readings and a +7 Intellect change under both multipliers (165 to 174, not 173.47).
- Preset DPS drops by about 0.4: Level 60 850.4 / 840.4 / 828.4 at 120 / 180 / 300 sec (was 850.8 / 840.8 / 828.7), Level 30 + 5 217.3 / 208.7 / 182.9 at 60 / 120 / 300 sec (was 217.7 / 209.1 / 183.4). Goldens refreshed for every Forever test that has a multiplier.
- Toughness uses the same code path but we didn't measure it on the beta. Item 8.27.
- The other four attributes drop the fraction too (beta, Kings on, same gear as the 1.2 to 1.6 sheet): UnitStat 116 / 47 / 116 / 91 for 116.6 / 47.3 / 116.6 / 91.3. What they convert to agrees: AP 272 (+20 over 252, keeping the fraction gives 273.2), health 1315 (+100, keeping it gives 1321), crit 5.7396 (+4 × 0.0868, keeping it gives 5.7656), GetManaRegen 17.626 (12.5 + 41 × 0.125 + the 0.001 floor). The test checks all of them.
- The same character without Kings read AP 252, health 1215, crit 5.3924 and GetManaRegen 16.626, so the differences above are measured on both sides.
- GetManaRegen's casting value read 1.411 with Kings and 1.331 without, where every earlier reading had 0.001. Both are 0.001 plus exactly 8% of the Spirit regen. It went back to 0.001 with no gear, and the source is Polished Driftwood Icon ("Allows 8% of your Mana regeneration to continue while casting"). The sim already has it (`sim/shaman/items.go`), and `mana_regen_test.go` TestOrcShamanDriftwoodIcon checks 16.625 and 1.33. This also shows that casting regen is a share of the full Forever Spirit regen. Item 2.3 still has the five second rule timing.

### 1.9 The stats panel

- Fixed (2026-10-08, user's call): the panel turns on every aura with a build phase after the shaman has registered its spells, so it showed our own Strength of Earth and Grace of Air whenever the shaman knew them. At 60 that was +53 Strength and +89 Agility, though the Level 60 preset puts down Windfury Totem and never Grace of Air. The fight was never affected. Our own totem buffs now have no build phase (`totems.go` ownTotemAura). A totem picked in the raid buffs (another shaman) is already permanent by then and stays on the panel. `stats_panel_test.go` TestOrcShamanStatsPanel checks both. No golden changed.
- The panel's 5% parry and 5% block for everyone is item 6.4.

### Pre-checks for 1.2 to 1.9 (not yet shown to the user)

- 1.2 Strength: 2 attack power a point (`base_stats.go` APPerStrength), 1 block value per 20. Same as Classic.
- 1.3 Agility: no attack power for a shaman (Classic, TBC added it). Crit and dodge per point follow the 1.12 curve, 2 armor a point.
- 1.4 Stamina: 10 health a point, the first 20 give 1 each.
- 1.5 Intellect: 15 mana a point, the first 20 give 1 each. Spell crit per point follows the 1.12 curve.
- 1.7 Only equipment hit and crit are made universal (`ruleset.go` unifyEquipHitAndCrit). Crit from consumes and buffs stays in its own pool: Elixir of the Mongoose and Leader of the Pack are melee only, Moonkin Aura is spell only. Open: under Forever's single crit stat, do they count for both? Items 9.8 and 10.2.
- 1.7 unifyEquipHitAndCrit sums melee and spell hit from gear. An item that lists both would count twice. Forever items list one hit stat, so this only matters for Classic items.
- 1.8 Percent stat modifiers multiply each other: Blessing of Kings 10% times Ancestral Knowledge 2% a rank on Intellect, Toughness 2% a rank on Stamina. 1.12 multiplies them too.

### 2.2 Spirit regen (beta, 2026-10-07)

- Fixed (2026-10-07): under Forever every player class uses the measured formula (`mana.go` ForeverSpiritManaRegenPerSecond). Pets keep their own, and the Classic ruleset keeps 1.12. `mana_regen_forever_test.go` pins the 7 Spirit readings and `enhancement/mana_regen_test.go` a naked Orc shaman at 1, 30 and 60. Level 60 preset 850.8 / 840.8 / 828.7 at 120 / 180 / 300 sec (was 850.5 / 839.2 / 825.3), the mana lasts about 30 sec longer. Level 30 + 5 217.7 / 209.1 / 183.4 at 60 / 120 / 300 sec (was 217.6 / 208.6 / 182.0, already with the 1.1 stats).

GetManaRegen on the level 30 Orc shaman, while not casting, in mana a second (the 0.001 floor taken off):

| Spirit | Intellect | Beta | Sim today (7.5 + Spirit / 10) | 6.25 + Spirit / 8 |
|---|---|---|---|---|
| 53 | 46 | 12.875 | 12.8 | 12.875 |
| 60 | 46 | 13.75 | 13.5 | 13.75 |
| 65 | 46 | 14.375 | 14.0 | 14.375 |
| 69 | 46 | 14.875 | 14.4 | 14.875 |
| 53 | 72 | 12.875 | 12.8 | 12.875 |

- Every level 30 reading fits 6.25 + Spirit / 8 a second exactly. That is 12.5 + Spirit / 4 per 2 sec, the 1.12 priest and mage formula (`priest.go`, `mage.go`). Intellect plays no part.
- The level 1 reading (22 Spirit, 5.5) doesn't fit it (9.0). Spirit / 4 a second fits. One shape fits all six readings: each of the first 50 Spirit gives 0.25 a second and each one past 50 gives 0.125. Past 50 that is the same as 6.25 + Spirit / 8. A level 30 or 60 shaman always has more than 50 Spirit, so only low levels would tell the two apart.
- Level 20 Undead paladin, geared: 34 Spirit gives 8.5 and 42 Spirit gives 10.5, both exactly Spirit / 4. So the paladin has the same formula, the 0.25 a Spirit below 50 holds at level 20 too, and nothing in it depends on level. In 1.12 the paladin and shaman shared 7.5 + Spirit / 10, so Forever replaced it for both.
- The formula: 0.25 mana a second for each of the first 50 Spirit, and 0.125 for each point past 50. The two parts meet at 50 (12.5 a second), which is why 6.25 + Spirit / 8 fits every reading above 50.
- The Spirit tooltip agrees with the server: 64 per 5 sec at 53 Spirit and 74 at 69 (14.875 × 5 = 74.4). Health regen in the tooltip still fits the 1.12 shaman formula (0.11 × Spirit + 7 per 2 sec).
- Real regen, sampled every 0.5 sec after a cast: the mana rose about 6 to 7 every half second (continuous, no 2 sec ticks), 51 mana from 6.0 to 10.0 sec, which is 12.75 a second against 12.875 with whole numbers. One step of +14 at 5.5 to 6.0 sec looks like the client catching up with the server once regen started.

### 2.1 Mana pool

- The beta readings from 1.1, 1.5 and 1.8 cover it: 1075 mana naked at 30, 2440 geared, 2635 with Ancestral Knowledge and 2860 with Kings on top. Base mana comes from the client table (1.1). Intellect gives 1 mana for each of the first 20 points and 15 for each one after. `base_stats_test.go` and `stat_multipliers_test.go` check these numbers. Nothing changed.

### 2.3 Five second rule

- The user OK'd it on 2026-10-08: it works like the other Classic versions. cmangos agrees (`Spell::TakePower`). Spending mana on a spell starts it. A spell with a cast time pays at the end of the cast, so we regen in full during the cast. A spell that costs nothing starts nothing.
- The sim: `mana.go` SpendCost runs the rule until 5 sec after the mana is paid, and a hardcast pays in its OnComplete (`cast.go`). Water Shield has no cost, and Clearcasting brings the cost to 0. Under Forever the bar settles up to the rule's end, so the casting rate and the full rate split exactly there.
- What we keep while casting: MP5 in full, plus a share of Spirit regen. Mindfulness gives 17 / 33 / 50% (the beta client text, the forever-wiki has 17 / 34 / 51), Polished Driftwood Icon 8% (beta, 1.8) and Improved Stormstrike 50% at both ranks for 15 sec (the client text).
- Fixed (2026-10-08): cmangos caps the share at 100% (`Player::UpdateManaRegen`), and the sim had no cap. Only Mindfulness 3/3, Improved Stormstrike and the icon together reach it (108%). No level 60 gear set has the icon, so no preset or golden changed.
- `five_second_rule_test.go`: TestOrcShamanFiveSecondRule follows the mana of a level 30 shaman with the icon through a Flame Shock, a Lightning Bolt and a free Flame Shock (Clearcasting). TestOrcShamanCastingRegenCap checks that Improved Stormstrike changes the rate the moment it starts and ends, and the cap. Moving the rule's start to the start of the cast, letting free casts start it, or taking out the cap each make the tests fail.

### 2.4 MP5

- MP5 gives MP5 / 5 mana a second, with no pause, and in full during the five second rule. Gear MP5 counts once. Blessing of Wisdom has Forever's ranks by level (12 / 18 / 24 / 30 / 36 / 40), and Improved Blessing of Wisdom does nothing under Forever. Mageblood Potion is 12 MP5 (Classic).
- Fixed (2026-10-08, user's call): Blessing of Wisdom and another shaman's Mana Spring Totem stack. The sim gave only Blessing of Wisdom when both were picked, because in Classic they came from opposite factions (`buffs.go`).
- Fixed (2026-10-08, user's call): the Mana Spring raid buff picks its rank by level (10 / 15 / 20 / 25 MP5 at 26 / 36 / 46 / 56, ×1.25 improved). It was always 25.
- `mp5_test.go` TestOrcShamanMP5Buffs checks levels 20, 30 and 60 under Forever and level 60 under Classic, and that MP5 continues while casting. No preset picks the Mana Spring raid buff. The goldens of the Forever tests with every raid buff changed (the mana users gain 31.25 MP5, up to +130 DPS on multi target, and the others show it only in the stats line).
- Not checked on the beta: whether GetManaRegen counts MP5. It needs an MP5 item or a Blessing of Wisdom from someone else.

### 2.5 Mana returns

- Judgement of Wisdom: 59 mana (rank 3) on a 50% roll from every direct hit, melee or spell. White swings roll even when they miss or are dodged (Classic). Procs and phantom spells don't. The user keeps rank 3 at every level (2026-10-08), because we only sim a raid at level 60. Its Forever value is still unknown.
- Water Shield: 3 globes of 2% max mana, free, 15 sec cooldown, 10 min (the beta client text). A landed spell, melee or ranged hit on the shaman or a heal crit spends a globe, at most one every 3.5 sec. The client only says "every few seconds". The user keeps 3.5 sec (2026-10-08): raid damage comes once every 5 to 15 sec and is the user's setting, so the wait never shows at those rates. No preset sets the raid damage rate, so it is 0 by default.
- Mana Tide Totem: 88 / 197 / 290 mana every 3 sec, 4 times, to the party members with mana, 5 min cooldown. It takes the water slot and stops Mana Spring Totem. Rank 1 matches the beta client text.
- `mana_returns_test.go`: TestOrcShamanJudgementOfWisdom (457 procs on 901 unarmed swings at level 60, 50.7%, 135 of them missed or dodged, 59 each, 42.8% if misses didn't roll), TestOrcShamanWaterShield (a hit every 5 sec spends a globe each time, a hit every 3 sec only every other hit), TestOrcShamanManaTide (ticks at 3 / 6 / 9 / 12 sec after the drop, Mana Spring gone). Taking the miss rolls out of Judgement of Wisdom, a 1 sec globe wait, or leaving Mana Spring up under Mana Tide each make the tests fail. Nothing changed in the sim.

### 2.6 Potions and Demonic Rune

- wowhead Forever (2026-10-08, the user's pick of source) has the Classic values for every mana potion (Lesser 280-360, Mana 455-585, Greater 700-900, Superior 900-1500, Major 1350-2250) and for Demonic Rune (900-1500 mana for 600-1000 health). Potions share a 2 min cooldown. Demonic Rune, healthstones and Minor Recombobulator share another one.
- Fixed (2026-10-08): Lesser Mana Potion gave 270-330. Minor Recombobulator had a 2 min cooldown where both versions say 5 min, and under Forever it gives 96-160 mana (Classic 150-250). No preset uses either, and no golden changed.
- Demonic Rune doesn't take its health in the sim. Only the Level 60 preset uses it, and nothing hits the shaman there.
- The rotation uses one once the missing mana is at least its top roll plus 2 sec of regen while casting.
- `potions_test.go` TestOrcShamanManaConsumables uses each 300 times and checks the roll range, the cooldown and when the rotation uses it. The old Lesser Mana Potion values make it fail.

### 2.7 Mana costs

- Every rank's mana cost (95 ranks, totems and Lava Burst included) matches wowhead Forever (2026-10-08). Water Shield shows no cost there and is free in the sim.
- The user OK'd the modifiers (2026-10-08). Convection takes 2% a point off shocks, Lightning Bolt, Chain Lightning and Lava Burst. Shamanistic Focus takes 45% off shocks and Lightning Shield. Totemic Focus takes 5% a point off totems. Maelstrom Weapon takes 4% a point a stack off the next Lightning Bolt's cost and cast time. The percents add (as in cmangos) and the cost never goes below 0.
- Clearcasting comes on 10% of the damage spells we cast, Fire Nova included, and makes the next one free. Totem attacks, Lightning Shield orbs and imbue procs don't give it or use it up. A Lightning Bolt uses it up at the end of its cast.
- Fire Nova isn't a totem under Forever, so Totemic Focus doesn't reach it (the user, 2026-10-08).
- A spell we can't pay for doesn't cast, and the sim marks us out of mana until we can. The rotation moves on to the next spell.
- `mana_costs_test.go` has TestOrcShamanManaCostTalents, TestOrcShamanMaelstromWeaponCost, TestOrcShamanClearcasting and TestOrcShamanNotEnoughMana. They fail when Fire Nova stops counting as a damage spell, when Searing Totem's attacks count, when Lightning Bolt doesn't use up Maelstrom Weapon, and when Clearcasting procs 20%.
- Fixed (2026-10-08, the user's go): Wrath and TBC spell IDs that wowhead Forever doesn't know. Lava Burst used 51505 on every rank and now has 408490, 1238299 and 1238300 as one rank family (`lava_burst_test.go`). Rage of the Farseer 2825 is now 425336, the Maelstrom Weapon buff 51530 is now 408505, and the Improved Stormstrike buff 51521 is now 1223031. The rotations and rotopt follow. The talent tree also had Wrath IDs for Lightning Overload, Mental Dexterity, Shamanistic Focus, Improved Stormstrike, Maelstrom Weapon, Water Shield and Riptide, and now has the client's Forever IDs. No golden changed.

### 3.1 GCD

- The user OK'd it (2026-10-08). wowhead Forever's spell pages give 1.5 sec for Earth Shock, Stormstrike, Water Shield and Fire Nova, and 0 for Rage of the Farseer and Nature's Swiftness. Totems take 1 sec. Haste never shortens the GCD, as in 1.12, so the 1 sec floor never comes into play. A cast longer than 1.5 sec holds the GCD until it ends, and a shorter one (Maelstrom Weapon, haste) still holds it for 1.5 sec.
- `gcd_test.go` TestOrcShamanGCD casts every shaman spell and totem at level 60, Lightning Bolt at 0, 25 and 100% haste and with 3 Maelstrom Weapon stacks, and the two off-GCD spells during the GCD. It fails when haste shortens the GCD, when totems take 1.5 sec, and when Rage of the Farseer goes on the GCD.
- Fixed (2026-10-08): casting Stoneskin Totem or Windwall Totem from a rotation crashed the sim. Stoneskin built its buff on the first cast, and Windwall has no buff and the air slot code expected one. Stoneskin's buff also added its -30 a second time when it ended. `stoneskin_test.go` checks the buff. No golden changed.
- For 7.17: the earth totems don't take the old totem's buff away, so Stoneskin and Strength of Earth can both be up. Only the air totems swap the buff. Stoneskin also always gives rank 6's -30, whatever the level.

### 3.2 Cast times and the swing timer

- Cast times match wowhead Forever (2026-10-08): Lightning Bolt 1.5 / 2 sec at ranks 1 and 2 and 2.5 sec from rank 3, Chain Lightning 2 sec and Lava Burst 2.5 sec at every rank.
- The user OK'd the swing timer (2026-10-08). We don't swing during a cast, and the swing timer starts over when it ends, so the next swing comes one full swing after the cast. Instant spells don't touch it.
- Fixed (2026-10-08, the user): an instant Lightning Bolt (5 Maelstrom Weapon stacks or Nature's Swiftness) leaves the swing timer alone, like a shock. It used to start the swing timer over too (`electric_spell.go` stopMeleeForCast). TestEnhancement goes up about 3%. The Level 60 preset is 872.9 / 864.3 / 853.4 at 120 / 180 / 300 sec (was 850.4 / 840.4 / 828.4). Level 30 + 5 has no Maelstrom Weapon and doesn't change. The rotation knobs were tuned on the old behavior.
- `swing_timer_test.go` TestOrcShamanCastSwingTimer casts 0.4 sec before a swing and checks the next swing after Lightning Bolt, Chain Lightning, a 3 stack bolt, a 5 stack bolt, a Nature's Swiftness bolt and Earth Shock. The old behavior fails it.

### 3.3 Haste

- The user OK'd it (2026-10-08). It matches Classic Era and wowsims/sod. Attack speed effects multiply (Slice and Dice 40% with Blade Flurry 20% and Juju Flurry 3% is 73%). Our sources are Flurry, Rage of the Farseer (attack speed only since the 2026-09-24 beta build), Berserking, Battle Squawk, Juju Flurry and haste on gear, which multiplies per item. Cast speed is kept apart and multiplies too (Berserking, spell haste on gear, the 1% set bonus).
- When attack speed changes mid-swing, the rest of the swing scales with the new speed, both when an effect comes up and when it drops. A cast in progress keeps the cast time it started with. Ranged attacks don't rescale mid-shot.
- Berserking for mana users cuts attack and cast time by 10 to 30% (30% is +42.9% speed), and for warriors and rogues it adds 10 to 30% speed. That's from wowsims/sod and we found no Classic source. It doesn't matter for an Orc.
- `haste_test.go` TestOrcShamanHaste stacks Flurry 5/5 and Rage of the Farseer halfway through a swing and drops them one at a time. It checks the swing speed, when the swing in progress lands and that cast speed stays at 1. It fails when the swing in progress doesn't rescale and when Rage of the Farseer gets cast speed.
- For section 8: Flurry gives 5 / 10 / 15 / 20 / 25% here, where Classic gave 10 to 30%.

### 3.4 Prepull

- The user OK'd it (2026-10-08). The sim starts at the first prepull cast with a full mana bar, and we regen from there. The weapon imbue is on from the start with no cast, mana or GCD. Each prepull cast pays its mana and starts the five second rule, so a totem at -1 sec keeps it going until +4 sec. Shield and totem time counts from the cast: 10 min for Lightning Shield, 5 min for the earth, air and water totems and Flametongue Totem, 30 to 55 sec for Searing Totem and 20 sec for Magma Totem. The totem's buff runs out with it.
- At 0 the auto attacks start and the first swing lands right away. The GCD of a totem cast at -1 sec ends at 0. Cooldowns start at 0. Nothing hits the target before the pull: Searing Totem cast at -1 sec first attacks at +1.5 sec, one 2.5 sec interval after the cast. When the first attack should come is part of 7.22.
- `prepull_test.go` TestOrcShamanPrepull casts Lightning Shield at -6 sec and Strength of Earth, Mana Spring and Searing Totem after it. It checks the mana after the first cast, the imbue, the GCD and the five second rule at the pull, every shield, totem and buff expiration, the first swing at 0 and the first Searing Totem attack. It fails when the pull comes 0.5 sec late, when the imbue isn't on, when the five second rule is 4 sec, and when Strength of Earth's buff or Mana Spring Totem gets a Classic duration.
- For the rotation session: the Level 30 presets should cast Lightning Shield before the pull (the user, 2026-10-08). It's in notes.md.

### 3.5 Reaction and batching delays

- The user OK'd it (2026-10-08). White hits roll when they swing and land 10 ms later with their procs, like the 10 ms batch window on Classic Era servers. The user thinks Forever uses 10 ms too, and it's in notes.md Need to Verify. Spells land at once, except that Lightning Bolt and Lava Burst travel at 20 yd/s from the player's distance (0 for enhancement in the UI, 5 yards in rotopt's settings). When the GCD is free and nothing can go, the rotation looks again every 50 ms and after every swing.
- Changed (2026-10-08, the user): the rotation sees a proc only after the player's reaction time, 200 ms by default. That's an aura or stacks a proc gives us, like Clearcasting, Maelstrom Weapon, Flurry and trinket procs. What a rotation action does shows at once, because we queue spells, and so does losing an aura or stacks. The GCD and cooldowns need no reaction time. It reaches every rotation condition that reads an aura (active, stacks, remaining time). Reaction Time is at the end of the Player section for every spec, a saved 0 reads as the default, and rotopt's mksettings writes 200 ms (was 150). The test suites keep their 150 ms and the shaman mechanics tests 0.
- `reaction_time_test.go` TestOrcShamanReactionTime casts Lightning Bolt at 5 Maelstrom Weapon stacks and Blood Fury when Rage of the Farseer is up. Each bolt goes 200 to 250 ms after the 5th stack, and Blood Fury goes with Rage of the Farseer. It fails when stacks show at once and when our own casts count as procs.

### 4.1 Weapon damage

- A white hit's base damage is a random number between the weapon's min and max damage, plus AP / 14 times the weapon's speed (`attack.go` CalculateWeaponDamage). The AP is ours plus any attack power the target gives us. Striking and Impact add to the weapon's min and max damage, so white hits, Windfury and Stormstrike all get them. Windfury's extra attack uses the weapon's speed and adds its bonus AP. Normalized speeds are 1.7 for daggers, 2.4 for one-handers, 3.3 for two-handers and 2.8 for ranged. Under Forever, Stormstrike uses them plus 0.3 (next item).
- Changed (2026-10-08, the user): under Forever, Stormstrike's attack power bonus uses the normalized speed plus 0.3 (`StormstrikeExtraSpeed`): 3.6 for two-handers, 2.7 for one-handers, and 2.0 for daggers, which we haven't tested. It used the weapon's own speed, like the SoD sim. Forever's spell data says "Normalized Weapon Damage" and Classic's says "Deal Weapon Damage" (wowhead, and wago.tools for client 70245, which has no flat bonus, per-level growth or attack power coefficient on it). The same Forever data gives the mark 21%, but the tooltip and ForeverChanges (client 70245: talents, changes and spellbook) all say 20%, like the sim. The mark covers Lightning Bolt, Chain Lightning and Earth Shock (its spell class mask), like the sim.
- The beta tests (level 30 Orc, no buffs, Elder Mottled Boars). The ratio is average Stormstrike over average white hit, which armor doesn't change because both hit the same boars. Crits count as half.

  | Weapon | AP | Boars | White hits | Stormstrikes | Ratio |
  |---|---|---|---|---|---|
  | Twin-bladed Axe of the Owl (2H, 23-35, 2.7) | 388 | 8 to 10 | 10 | 12 | 1.248 |
  | Barbaric Battle Axe of Healing (2H, 25-38, 3.6) | 390 | 8 to 10 | 16 | 19 | 0.977 |
  | Twin-bladed Axe of the Owl | 195 | 8 to 10 | 22 | 12 | 1.200 |
  | Twin-bladed Axe of the Owl | 195 | 9 only | 34 | 18 | 1.171 |
  | Bloody Brass Knuckles (1H, 24-46, 1.6) | 204 | 8 only | 27 | 14 | 1.277 |

  The weapon's own speed is far off (ratio 1 every time). So is the usual 3.3 for two-handers. On boars of one level the armor is the same, so the lowest and highest hits count too. In the level 9 run, a 53 white hit and two 75 Stormstrikes rule out every two-hand speed below 3.55. In the level 8 run the one-hand speed fits from 2.28 to 2.87, best 2.72, and plain 2.4 is 2 standard errors low. Speed 3.6 misses only the slow axe session, where Stormstrike averaged 2.3% below white (2.4 standard errors). Two other models fit every session as well: the usual speed then 6.4% more, or the usual speed then 5 more flat. The user picked 3.6, and the one-hander's 2.7 makes it "plus 0.3". At level 60 with the 3.4 speed Synthetic two-hander, 3.6 adds about 4% to Stormstrike damage, about the same as 6.4% more, while 5 more flat would take away 1%. notes.md Need to Verify has the test that tells them apart.
- Fixed (2026-10-08, the user): Greater Impact (+7) and Superior Impact (+9) did nothing in the sim, and upstream wowsims has the same gap. They now add to the weapon's damage. hyjal.cc's recipe pages give the same values as 1.12. The rotopt enchant search stops at Enchanting 225, so it never offered them, and no preset changes. TestP1Hunter's gear has Superior Impact and gains about 1%.
- Fixed (2026-10-08, the user): flat "+N damage" effects (Bogling Root, Zandalarian Hero Medallion, the Ragehammer and Sword of Zeal procs, Might of Cenarius) reached only white hits. Stormstrike and Windfury now get them too, as in Classic, through BonusCoefficient 1 like the other classes' weapon specials. No preset uses them.
- `weapon_damage_test.go` TestOrcShamanWeaponDamage fights 30 min at level 60 with Dark Edge of Insanity, Superior Impact and a +20 flat bonus against a target with no armor. Every normal white hit, Windfury hit and Stormstrike lands in its range, and the rolls reach both ends. It runs again with a one-hander, Crul'shorukh, without Impact. It fails without Superior Impact, without the flat bonus on Stormstrike or Windfury, and with Stormstrike's speed 0.1 off on either weapon.
- How we ran the Stormstrike tests: a combat log frame through /run saw no events on the beta client, so the user screenshots the chat combat log instead. The shown amount plus the overkill is the whole hit, and a melee crit is twice a normal hit. This line prints AP (base, plus, minus) and the main hand's damage range at the start:

  ```
  /run print("ap", UnitAttackPower("player")) print("dmg", UnitDamage("player"))
  ```

### Pre-checks for sections 3 and 4 (not yet shown to the user)

- 4.3 White hits roll once. From behind: miss, dodge, glancing, crit, hit. In front (`InFrontOfTarget`): miss, dodge, parry, glancing, block, crit, hit (`spell_outcome.go` outcomeMeleeWhite).
- 4.4 to 4.7 (`target.go` NewAttackTable), level 60 with 300 skill vs a level 63 boss: 8% miss, and the first 1% of hit from gear does nothing (9% to cap). 6.5% dodge, 14% parry from the front, 40% glancing for 55 to 75% damage (65% on average). Crit is cut by 4.8% (3% from the level gap, 1.8% aura suppression, taken off all crit rather than only crit from auras, see the TODO).
- Level 30 vs Vishas (level 32) with 150 skill: 6% miss, 6% dodge, 6% parry from the front, 30% glancing for 80 to 90% damage, crit cut by 2%.
- Open: spells lose 2.1% crit against a level +3 target (`target.go` SpellCritSuppression). That isn't a 1.12 rule as far as we know. Item 5.3.
- Open: Forever adds an expertise-like stat that lowers dodge and parry, and weapon skill on gear. Neither is in the attack table yet. Items 4.5 and 4.8.

### Pre-checks for sections 5 and 6 (not yet shown to the user)

- 5.1 Spell miss against a target 0 / 1 / 2 / 3 levels above us is 4 / 5 / 6 / 17%, then 11% more a level, at most 99% (`target.go` spellMissChance). Same as 1.12.
- 5.2 Partial resists follow the royalgiraffe resist guide (`spell_resistances.go`). A target above our level adds 2% average mitigation a level (6% for a level 63 boss, 4% for Vishas), for spells that aren't binary. A DoT with no direct part takes a tenth of the resistance.
- 5.2 Earth Shock is binary (`earth_shock.go`, SpellFlagBinary), so it takes no level-based partial resist. Frost Shock and Flame Shock aren't. Open: which shaman spells 1.12 treats as binary. Items 7.8 to 7.10.
- 6.1 The Level 60 encounter has no tank, so the boss never swings at us. The Level 30 encounter makes us the tank (`tankIndex: 0`), so Vishas hits us and Lightning Shield fires. Possible bug: that preset doesn't set "in front of target" (only Level 30 PvP does), so our swings never meet Vishas's 6% parry or block, even though we tank him from the front. Items 4.5 and 6.1.

### Pre-checks for section 7, part 1 (not yet shown to the user)

- 7.1 The 2026-10-07 class audit already checked all 458 shaman, rogue and paladin ranks against wowhead Forever and ForeverChanges 70245 (memory class-audit-2026-10-07). Item 7.1 only needs whatever changed since.
- 7.7 Stormstrike under Forever (`stormstrike.go`): 125 mana, 8 sec cooldown, 1.5 sec GCD, a main hand hit for weapon damage on the special attack table, normalized to the weapon type's speed plus 0.3 (4.1). When it lands it marks the target for 12 sec. Our next Lightning Bolt, Chain Lightning or Earth Shock on that target takes the mark and deals 20% more.
- 7.11 The shocks share one cooldown, 6 sec minus 0.2 sec per point of Reverberation (`shocks.go`), with a 1.5 sec GCD. Cost is cut by 2% per point of Convection and 45% with Shamanistic Focus. The two add up, they don't multiply.
- 7.9 Flame Shock: the direct part has a 0.214 spell power coefficient and grows with level up to its rank cap. The DoT is 4 ticks every 3 sec (12 sec) with 0.1 per tick, and it doesn't grow with level (`flame_shock.go`). Burning Totem adds 3 sec, one more tick.

### Pre-checks for section 7, part 2 (not yet shown to the user)

- 7.15 Lightning Shield (`lightning_shield.go`): 3 orbs, 10 min. A melee hit on us that lands fires one orb, at most one every 3.5 sec (the ICD is a guess, "TODO: Does vanilla have an ICD?"). 26.7% spell power coefficient on every rank (Forever).
- 7.16 Water Shield (`water_shield.go`): free, 15 sec cooldown, 3 globes, 10 min. Each globe returns 2% of max mana, at most one every 3.5 sec (a guess borrowed from Lightning Shield). Only the raid damage hits per minute option feeds it at level 60.
- 7.3 Flametongue Weapon (`flametongue_weapon.go`): a fire hit on each main hand hit for (rank damage / 4) times weapon speed, 112.4 per 4 sec at level 60. Its 0.1 spell power coefficient does not scale with weapon speed. Open: does Forever scale the coefficient with speed? Elemental Weapons adds 5 / 10 / 15%.
- Beta, level 30 with 0/3 Elemental Weapons (2026-10-08):
  - Rockbiter rank 4 gave 177 AP (sim 177.6).
  - The Windfury tooltip says 46 AP (sim 46). The log shows two "Windfury Weapon" hits per proc. With the Barbaric Battle Axe of Healing (25 to 38, 3.6 speed) at 198 AP, 9 Windfury hits averaged 75.1 and 16 white hits 65.9, a ratio of 1.140 (sim 1.144).
  - Flametongue with the same axe and 6 spell power hit 27 or 28 Fire (9 and 8 hits), so about 27.5. The sim gives 35.36 * 3.6 / 4 + 0.6 = 32.4, 15% more. The tooltip's "11 to 36" matches the client's 884 / 77 - 1 and 884 / 25, plus 0.6 for the spell power. With 0 spell power the tooltip says 10 to 35, so the tooltip rounds to the nearest number and its "+ 0" terms are the spell power part.
  - Rank 3 with the Twin-bladed Axe of the Owl (2.7) hit 18.6 (sim 23.9) and with Bloody Brass Knuckles (1.6) 9.1 (sim 14.1). So rank 3 grows 8.84 per second of speed like the sim, but every hit is about 5 lower.
  - Rank 1 with the slow axe hit 16 or 17 (3 and 3, crits 25 and 24 at 1.5x), so 16.5 against the sim's 15.84 + 0.6 = 16.4. Rank 1 has no cut.
  - Searing Totem rank 3 hit 19 to 24 on the same mobs (client and sim 19 to 25), so the mobs don't take fire damage off.
  - hyjal.cc's proc spells (client) give the same points as the sim: 326 + 19 a level (rank 1, learned at 10), 479 + 29 (18), 716 + 42 (26), 1144 + 73 (36), 1876 + 62 (46), 2498 + 78 (56), each growing for 6 levels (8 for rank 3 on). vmangos uses points / 100 * speed with 3.85% of spell power per second of speed, and rounds at random.
  - Rank 1 on the knuckles hit 7 (sim 7.04). Rank 2 hit 22 on the slow axe (sim 23.5 + 0.6) and 8 or 9 on the knuckles (sim 10.45), so about 2 lower at both speeds. Each rank loses a flat amount per hit, the same on every weapon: 0 for rank 1, about 2 for rank 2 and about 5 for rank 3. (L - 326) / 77 fits (2.0 and 5.1), where L is the rank's points at its learn level, but we know no reason for it.
  - Other sources (2026-10-08) all give the client's points: Classic Era 1.15.9 (wago.tools SpellEffect), vmangos's 1.10 to 1.12 rows, wowhead's Classic, TBC and Forever tooltips, and ForeverChanges. vmangos's pre-1.10 rows are about 10% lower (296 + 17, 435 + 26, 651 + 38) but don't fit either, because rank 1 would hit 14.3 on the slow axe. vmangos, cmangos and mangoszero all use points / 100 * speed with no cut. Forever's dev notes only change Flametongue Totem, and its downranking change only takes gear bonus off low ranks. We found no report of Flametongue hitting below its tooltip, so the cut is likely a Forever server change or bug.
  - Decision (user, 2026-10-08): the sim keeps the client's points for every rank for now, and we check rank 6 at level 60 (notes.md Need to Verify).
  - Flametongue Totem's procs are 489, 697, 947 and 1217 in Classic Era. Forever's (wowhead, and the sim) are 548, 781, 1061 and 1363.
- 7.22 Searing Totem (`fire_totems.go`): a bolt every 2.5 sec for the totem's 30 to 55 sec lifetime, 1.7% coefficient (Forever). The real attack is every 2.2 sec, but the next bolt waits for the last to land, so we use 2.5 sec without a distance option.
- 7.23 Magma Totem: a pulse every 2 sec for 20 sec, 2 less damage per pulse than 1.12 (Forever).
- 7.24 Flametongue Totem: 5 min, fire damage per main hand hit by weapon speed (548 / 25 per 4 sec at rank 1, flat by level). It takes the fire slot, and placing another fire totem takes it down.
  - Beta, level 30 (2026-10-08): rank 1 on the 3.6 speed axe hit 19 or 20. The sim gives 548 / 100 * 3.6 = 19.73 plus 10% of spell power. Classic Era's 489 would give 17.6, so Forever's 548 holds. Open: whether spell power adds to it (cmangos says it doesn't scale with gear).
- 7.25 Fire Nova (Forever spell): needs a fire totem down, 10 sec cooldown, 1.5 sec GCD, Totemic Focus doesn't discount it. Damage is Classic's (rank 1 grows over 5 levels). Call of Flame and Improved Fire Nova add up.
- 7.17 Totems last 5 min under Forever (Searing, Magma and Fire Nova keep their own times). The weapon totem slot rules are in `core/totem_weapon_buffs.go` and `shaman.go` setTotemWeaponBuff.

### Pre-checks for section 8 (not yet shown to the user)

What each talent does in the sim today (`talents.go` unless noted). Values are per point.

- Convection: -2% mana cost on Lightning Bolt, Chain Lightning, the shocks and Lava Burst.
- Concussion: +1% damage on Lightning Bolt, Chain Lightning and Earth Shock, added to the other percent bonuses.
- Elemental Warding: 3 / 7 / 10% less fire, frost and nature damage taken.
- Reverberation: -0.2 sec on the shared shock cooldown (`shocks.go`).
- Call of Flame: +5% on Flame Shock and Lava Burst (multiplied), on the fire totems' base damage, and on Fire Nova added to Improved Fire Nova.
- Elemental Devastation: a spell crit gives +3% melee crit for 10 sec. Flametongue Weapon and Totem procs don't count (they are procs, not spells).
- Elemental Focus: 10% chance on each damaging shaman spell cast that the next one costs nothing.
- Elemental Alacrity: -0.17 / 0.33 / 0.5 sec cast time on Lightning Bolt, Chain Lightning and Lava Burst.
- Improved Fire Nova: +10% Fire Nova damage and -2 sec on its cooldown.
- Call of Thunder: +3% crit on Lightning Bolt and Chain Lightning (`electric_spell.go`).
- Lightning Overload: 3 / 7 / 10% chance to cast a second bolt (`lightning_overload.go`).
- Elemental Fury: +20% crit damage bonus on shaman fire, frost and nature spells plus the Searing and Magma Totem attacks. Flametongue Totem's attack isn't included.
- Thundering Strikes: +1% melee and +1% spell crit.
- Ancestral Knowledge: +2% Intellect.
- Guardian Totems: Stoneskin Totem (`earth_totems.go`).
- Mental Dexterity: attack power equal to 33 / 67 / 100% of Intellect. The older wiki tree reads 33 / 66 / 99%, the beta client data 33 / 67 / 100%.
- Improved Lightning Shield: +5 / 10 / 15% on the whole orb hit (`lightning_shield.go`).
- Elemental Weapons: Rockbiter +7 / 14 / 20%, Windfury +13 / 27 / 40% applied twice, Flametongue and Frostbrand +5 / 10 / 15%.
- Shamanistic Focus: -45% cost on the shocks and Lightning Shield. Open: is Lightning Shield in Forever's text?
- Anticipation: +2% dodge. Toughness: +2% Stamina.
- Flurry: a melee crit (white, Stormstrike, Windfury attacks) gives 3 charges of 5 to 25% attack speed. White swings use the charges, at most one per 0.5 sec so that both hands don't spend two at once. Windfury Weapon's attacks are yellow, so they don't use charges. Windfury Totem's extra attack is white, so it does. Open: which extra attacks use Flurry charges under Forever.
- Stormstrike: see 7.7.
- Spirit Weapons: lets us parry, and threat 0.7 (1.3 with Rockbiter). Small bug: it reads Rockbiter from the consumes imbue, not the Forever shaman imbue. Threat isn't in our DPS numbers.
- Mental Quickness: spell power equal to 15% of Intellect.
- Improved Stormstrike: 50% chance a point that Stormstrike gives 50% regen while casting for 15 sec, and 50% a point that a dodge or parry of an enemy attack on us resets Stormstrike.
- Maelstrom Weapon: 2 procs a minute per point from landed melee hits, at most 5 stacks for 30 sec. Each stack cuts Lightning Bolt's cast time 4% and its cost 4% per point. Lightning Bolt uses all stacks. Chain Lightning doesn't use them. The proc rate is a guess.
- Rage of the Farseer: +30% attack speed for 25 sec, 3 min cooldown (a guess).
- Totemic Focus: -5% totem mana cost, not Fire Nova.
- Mindfulness: 17 / 33 / 50% Spirit regen while casting.
- Natural Grace: -5% threat on shaman spells.
- Tidal Focus: -1% healing cost, +1% melee and spell hit.
- Improved Reincarnation: +2% health.
- Tidal Mastery: +1% heal crit.
- Restorative Totems and Purification: Healing Stream and Mana Spring.
- Water Shield and Mana Tide Totem: see 7.16 and 7.27.
- Nature's Swiftness: the next nature spell with a cast time is instant, 3 min.
- Not implemented: Eye of the Storm, Elemental Reach, Earthbound, Earth's Grasp, Improved Ghost Wolf, Improved Healing Wave, Ancestral Healing, Healing Focus, Healing Way, Riptide.

### Pre-checks for sections 9 to 11 (not yet shown to the user)

- 9.1 Orc under Forever (`racials_forever.go`): Axe Specialization gives +1% melee and spell crit while the main or off hand is an axe. It is checked once at the start, so an item swap doesn't update it. Blood Fury: +10% attack power and spell power for 15 sec, 2 min cooldown. Open: the sim takes 10% of our attack power when Blood Fury starts and keeps that amount. A real percent buff would follow our attack power as it changes during the 15 sec, for example when a new Strength of Earth goes down. It also triggers the 1.5 sec GCD. Does Forever's Blood Fury use the GCD?
- 9.2 Relics (`shaman/items.go`): Burning Totem adds a Flame Shock tick, Polished Driftwood Icon 8% regen while casting, Totem of Rage +30 to shocks, Classic Totem of the Storm +33 to Lightning Bolt and Chain Lightning, Forever Totem of the Storm lets Lightning Bolt proc Maelstrom at half the chance.
- 9.3 Rage of the Storm: Stormstrike +10%, added to the other percent bonuses.
- 9.5 Hand of Justice: 2% chance for an extra attack on any melee hit that lands, white or yellow (`common/item_effects.go`).
- 10 Forever buff values live in `core/buffs.go` behind IsForever (Blessing of Wisdom, Strength of Earth 53, Grace of Air 89, and others). Under Forever both factions get blessings and totems. Judgement of Wisdom is still Classic (see 2.5).
- 11.1 Level 60: level 63 boss, 3731 armor, 180 sec plus or minus 20, no tank so it never hits us, every debuff on. Level 30: Vishas level 32, 1063 armor, 120 sec plus or minus 5, we are the tank (and the in front bug from 6.1).

### Source checks against the beta client talent texts (`tools/forever_talents/client/shaman.json`)

- Rage of the Farseer: build 70009 says "Instant; 3 min cooldown". That settles the 3 min the sim guessed. The code comment and the notes.md Need to Verify line can be updated at 8.34.
- Shamanistic Focus: "Reduces the mana cost of your Shock and Lightning Shield spells by 45%." Lightning Shield is in, the sim is right.
- Mental Dexterity: 33 / 67 / 100% of Intellect, the sim is right (the wiki's 99% was an older tree).
- Stormstrike: "Instantly strike for normal weapon damage", so not normalized, the sim is right.
- Elemental Devastation: "Your offensive spell critical strikes". It doesn't say whether Flametongue procs count. Still open.
- Elemental Focus: "after casting any Fire, Frost, or Nature damage spell", for "your next damage spell". Matches the sim.
- Improved Reincarnation: +2% max health a point. Matches.
- Flurry: "your next 3 swings after dealing a melee critical strike". Whether extra attacks count as swings is still open.
- Blood Fury (9.1): wowhead's Forever tooltip (nether endpoint, dataEnv 17, spell 20572) says "Instant, 2 min cooldown. Increases your attack power by 169 for 15 sec." That is a flat value with no spell power. The sim follows the Forever deep dive's "+10% attack power and spell power". The two disagree, so this goes to the beta check list.
- Judgement of Wisdom (2.5): wowhead's Forever tooltips for 20186, 20354 and 20355 come back empty, and its search is behind a 403 right now. Still open.
