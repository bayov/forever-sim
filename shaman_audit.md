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
- [x] 4.2 Swing timer: weapon speed and melee haste (Flurry, Rage of the Farseer, item haste), main hand and off hand sync
- [x] 4.3 Attack table: one roll, in the order miss, dodge, parry, glancing, block, crit, hit
- [x] 4.4 Miss: base chance against the target's level, weapon skill, hit from gear, the hit cap
- [x] 4.5 Enemy dodge and parry: values by level, no parry or block from behind (raid), both from the front (PvP)
- [x] 4.6 Glancing blows: chance against a level +3 target, the damage penalty, weapon skill
- [x] 4.7 Crit: chance, the crit cap that glancing and miss make, crit damage
- [x] 4.8 Weapon skill: base by level, gear skill, its effect on 4.4 to 4.7 (Forever removed the weapon skill racials)
- [x] 4.9 Enemy armor: the mitigation formula, armor debuffs and their Forever values (Sunder, Expose Armor, Faerie Fire, Curse of Recklessness)
- [x] 4.10 Physical damage modifiers: flat bonus damage, percent modifiers
- [x] 4.11 Procs from melee hits: PPM procs (weapon speed), chance procs, which hits can proc what (white, yellow, extra attacks, procs from procs)
- [x] 4.12 Extra attacks: Windfury Totem, Hand of Justice, how they line up with the swing timer
- [x] 4.13 Proc matrix (the user, 2026-10-08): for every ability, what it can proc and what it can't. The abilities are white hits, Windfury Weapon and Totem attacks, Stormstrike, the Flametongue Weapon, Flametongue Totem and Frostbrand procs, the shocks, Lightning Bolt, Chain Lightning, Lava Burst, Fire Nova, totem attacks and Lightning Shield orbs. The procs are Windfury Weapon and Totem, Flametongue Weapon and Totem, Frostbrand, Flurry, Maelstrom Weapon, Elemental Devastation, Clearcasting, Judgement of Wisdom, Water Shield, item and enchant procs (PPM and chance) and racials. The weapon imbues get their own pass here, even though 7.2 to 7.4 have beta damage tests already. Flametongue Totem on white swings only, and both Flametongue procs on landed hits only (blocks included), were settled on the beta (2026-10-10).

## 5. Our spells

- [x] 5.1 Spell hit: base miss against the target's level (17% at +3), hit from gear, the 1% floor
- [?] 5.2 Resistances: target resistance, the level based resistance, average partial resists vs binary spells
- [x] 5.3 Spell crit: base, Intellect, gear, the 1.5 crit multiplier
- [x] 5.4 Spell power: coefficients, school power (fire, nature), spell damage vs spell power under Forever
- [x] 5.5 Spell ranks: the rank each level knows, the penalty for spells learned below level 20
- [x] 5.6 How damage modifiers stack: same kind add (the 1.12 rule), different kinds multiply, target debuffs (Curse of the Elements, the Stormstrike mark)
- [x] 5.7 DoTs: tick timing, what snapshots, refresh, crits on ticks (Flame Shock)
- [x] 5.8 Abilities that use the melee table (Stormstrike, Windfury attacks, the imbue attacks)

## 6. When the enemy hits us

- [x] 6.1 When the enemy attacks us at all: the tank setting (Level 30 solo, PvP), the boss behind the tank at level 60
- [x] 6.2 Enemy damage: weapon damage, attack power, swing speed, parry haste
- [x] 6.3 Our armor: mitigation, Stoneskin Totem, Devotion Aura
- [x] 6.4 Our miss, dodge, parry and block: base values, Agility, Anticipation, parry only with Spirit Weapons, block only with a shield, enemy crits and crushing blows
- [x] 6.5 Spell damage to us: Elemental Warding, resistance auras, the raid damage hits that feed our shield
- [x] 6.6 Health: Stamina, Toughness, Improved Reincarnation, healing (Healing Stream)
- [x] 6.7 PvP mode: the enemy types, time out of melee range

## 7. Shaman abilities

- [x] 7.1 Rank tables: every spell's ranks, learn levels and values at 30 and 60 (vs the ForeverChanges spellbook)

Weapon imbues

The level 30 beta damage tests for 7.2 to 7.4 and 7.24 are in the Findings (section 7, part 2). Each item is still open until its full summary, OK and test, and what each imbue procs and gets procced by is in 4.13.

- [x] 7.2 Windfury Weapon: proc chance, ICD, two attacks, bonus attack power, which hits proc it
- [x] 7.3 Flametongue Weapon: fire damage per hit by weapon speed, coefficient
- [x] 7.4 Rockbiter Weapon: attack power
- [x] 7.5 Frostbrand Weapon: proc rate, damage
- [x] 7.6 Imbue rules: the shaman imbue beside an oil or stone (Forever), one imbue per weapon

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
- [ ] 7.28 Healing Stream Totem. Fixed ahead of the item (2026-10-10, the user's go): Purification no longer raises Healing Stream. Its beta tooltip says "your healing spells", but its spell modifiers list only Healing Wave, Lesser Healing Wave and Chain Heal (class mask 0x1C0) and one Forever spell in the third mask word (0x10). Healing Stream's heal is 0x2000 in both the Forever (70291) and Era (1.15.9) clients. Restorative Totems is right: its talent curves give Mana Spring 5% and Healing Stream 10% a rank, as the beta tooltip says. `periodic_test.go` TestOrcShamanHealingStreamTalents checks 9 a heal at rank 2 with 5 points in each.
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
- [x] 8.24 Elemental Weapons (beta 2026-10-09: Rockbiter +20%, Windfury +40% once, Flametongue +15% on the whole hit, see section 7 part 2)
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
- Changed (2026-10-10, the beta session with the user's OK, bda1b125): Forever changed the Orc race, not the shaman class. A level 1 Orc warlock is off from 1.12 by the same 0 / 0 / -1 / +2 / -3 as our shaman. So every Orc row is shifted by that, and the level 60 Orc offset is now +3 / -3 / +1 / -1 / 0 against the Human base. A naked level 60 Orc shaman is 88 / 52 / 96 / 89 / 100, with 2060 health, 2575 mana and 18.75 mana a second. That's an assumption until release (notes.md At 60). TestEnhancement moved by +0.03% on average, and all shaman tests pass.
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
- Beta, Windfury Weapon's timing (our count from the whole log, 38 procs, 2026-10-10): both attacks landed within 3 ms of the swing or Stormstrike that made them in 30 procs. In 6 procs both landed 101 to 115 ms later. In 2 the first attack landed with the swing and was dodged or parried, and the second landed 104 and 113 ms later. The sim lands both with the swing. So Forever sometimes holds a proc for about 100 ms, more than the 10 ms window in notes.md. It changes nothing we can measure in DPS, so we only record it.

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
- Beta, two-hander with Rage of the Storm (the beta session, 2026-10-10, `beta_results.txt` Shimmering Flats run, commit 70efe9c9): a level 30 Orc at 389 AP with Rage of the Storm (94 to 141, 3.3 speed, +6 Weapon Damage) on level 33 to 35 mobs. The log's raw damage field is before armor, crit and glancing. 75 white hits ran 193 to 238, inside the model's 191.7 to 238.7, so AP held at 389. 30 Stormstrikes ran 222 to 271 (mean 249.0). The sim's normalized 3.6 with Rage of the Storm's +10% gives 220.0 to 271.7, and all 30 are inside. "3.3, then 6.4% more" (224.4 to 279.4) leaves 222 and 223 out, and "3.3, then 5 more flat" (216.4 to 268.1) leaves 271 out. The sim's model stands, assuming Rage of the Storm's +10% works (the user). No talent in the client (70338) touches Stormstrike's damage except Mental Dexterity through AP. 13 more Stormstrikes in a later run fit too (ad8ec9a5). Stormstrike with a dagger stays in Need to Verify (Beta).
- Fixed (2026-10-08, the user): Greater Impact (+7) and Superior Impact (+9) did nothing in the sim, and upstream wowsims has the same gap. They now add to the weapon's damage. hyjal.cc's recipe pages give the same values as 1.12. The rotopt enchant search stops at Enchanting 225, so it never offered them, and no preset changes. TestP1Hunter's gear has Superior Impact and gains about 1%.
- Fixed (2026-10-08, the user): flat "+N damage" effects (Bogling Root, Zandalarian Hero Medallion, the Ragehammer and Sword of Zeal procs, Might of Cenarius) reached only white hits. Stormstrike and Windfury now get them too, as in Classic, through BonusCoefficient 1 like the other classes' weapon specials. No preset uses them.
- `weapon_damage_test.go` TestOrcShamanWeaponDamage fights 30 min at level 60 with Dark Edge of Insanity, Superior Impact and a +20 flat bonus against a target with no armor. Every normal white hit, Windfury hit and Stormstrike lands in its range, and the rolls reach both ends. It runs again with a one-hander, Crul'shorukh, without Impact. It fails without Superior Impact, without the flat bonus on Stormstrike or Windfury, and with Stormstrike's speed 0.1 off on either weapon.
- How we ran the Stormstrike tests: a combat log frame through /run saw no events on the beta client, so the user screenshots the chat combat log instead. The shown amount plus the overkill is the whole hit, and a melee crit is twice a normal hit. This line prints AP (base, plus, minus) and the main hand's damage range at the start:

  ```
  /run print("ap", UnitAttackPower("player")) print("dmg", UnitDamage("player"))
  ```

### 4.2 Swing timer

- The user OK'd it (2026-10-08). The main hand swings every weapon speed divided by our attack speed. The haste rules, the cast pause, the first swing and the 10 ms landing are in 3.2 to 3.5.
- Shamans and paladins can't hold a weapon in the off hand under Forever (the user, 2026-10-08), so main hand and off hand sync doesn't apply. Shields and items held in the off hand stay. The UI already blocked off-hand weapons, but the sim swung one that a gear set had, and rotopt's gear search offered one-handers in the off hand to every class. No preset had one.
- Fixed (2026-10-08): `core.CanDualWield` lets only warriors, rogues and hunters hold an off-hand weapon, like the UI's canDualWield. NewCharacter drops an off-hand weapon from the other classes, so it adds no stats, swings, procs or dual wield miss penalty. rotopt offers the off hand only shields and held items for them. No golden result changed.
- `off_hand_test.go` TestOrcShamanOffHandWeapon runs Deathbringer with and without Crul'shorukh in the off hand on the same seed. It checks the main hand swings every 2.9 sec, the off hand never swings, the stats match and both fights deal the same main hand damage. It fails without the fix.

### 4.3 Attack table

- The user OK'd it (2026-10-09). White hits (auto attacks and Windfury Totem's extra attack) roll once on one table (`spell_outcome.go` outcomeMeleeWhite). From behind: miss, dodge, glancing, crit, hit. In front of the target: miss, dodge, parry, glancing, block, crit, hit. Crits only get what the outcomes before them leave (the crit cap, 4.7).
- Stormstrike and Windfury Weapon's attacks are yellow (outcomeMeleeWeaponSpecialHitAndCrit). The first roll is miss, dodge, then parry and block in front, then hit. A second roll decides the crit. They never glance, and in front a blocked one can't crit.
- The white order fits magey's 2019 Classic tests (white hits only). Creatures dodge from any side, parry and block need the front, and only white swings and extra attacks glance. Yellow crits come from a second roll, like the upstream wowsims Classic sim and the SoD sim. A private server core (vmangos) puts them in the same roll, which would give Stormstrike about 15% more crits against a level 63 boss. We can't tell them apart at level 30.
- Settled (the user, 2026-10-09): Windfury Weapon's attacks are yellow. The client (wago 70291, the same as Era) gives spell 8233 2 extra attacks plus 46 attack power for 1.5 sec, and 1.12 servers swing those white, so they could glance. The beta log names them "Windfury Weapon" hits, where a white swing reads "Melee". The sim already rolled them yellow, like the SoD sim, so only the TODO in `windfury_weapon.go` changed.
- `attack_table_test.go` TestOrcShamanAttackTable fights a level 63 boss for 20 hours at level 60 with Dark Edge of Insanity and 20% extra crit, from behind and in front, then with 60% extra crit from behind to pass the crit cap. It counts every white hit, Stormstrike and Windfury attack: 8% miss, 6.5% dodge, 14% parry and 5% block in front only, 40% glancing on white hits only, crits at the full chance on white hits and from the second roll on yellow ones. It fails when Windfury rolls white, when crit comes before glancing, when yellow crits share the first roll, and when white hits can be parried from behind.

### 4.4 Miss

- The user OK'd it (2026-10-09) and skipped the beta test for now. Miss is 5% plus 0.1% for each point the target's defense (its level times 5) is above our weapon skill, or 0.2% a point when the gap is more than 10 (`target.go` NewAttackTable). Our hit comes off it, and miss never goes below 0. White hits and yellow attacks miss the same, and a shaman has no dual wield penalty (4.2). When the gap is more than 10, (gap - 10) * 0.2% of our hit does nothing. That's 1% for 300 skill against a level 63 boss, so the hit cap is 9%. Level 30 against Vishas (level 32) misses 6%, with no hit ignored.
- Hit comes from gear (one stat for melee and spells) and Tidal Focus. Weapon skill is level times 5 plus skill from gear. Forever's Orc Axe Specialization is +1% crit with axes, not +5 skill (client spell 20574, `racials_forever.go`).
- These are magey's 2019 Classic formulas (Beaza's), and Blizzard's 2019 posts give 8% miss and the ignored first 1% for 300 skill against a level 63. A private server core (vmangos) ignores a flat 1% for any gap over 10 instead. That only differs at 301 to 304 skill or 4 or more levels up, so we keep the per point version. Mobs below level 10 are also missed less (miss times level / 10), which only matters for reading beta logs.
- Tidal Focus gives 1% hit a point. Forever's talents use the trait system: client 70291 has only rank 1 (16179), with the top rank's +5, and TraitDefinitionEffectPoints and CurvePoint give 1 to 5% by rank. ForeverChanges agrees.
- `miss_test.go` TestOrcShamanMissChance checks the miss chance of white hits and Stormstrike in 13 cases: level 60 against a level 63 boss with 0, 1, 5, 9 and 12% hit and with Tidal Focus 5/5, level 30 against levels 25, 30, 32 and 33, and Dwarven Tree Chopper's +2 skill against a level 33. It fails without the ignored hit, with a flat 1% ignored, with the Classic Orc +5 axe skill and without Tidal Focus's hit.

### 4.5 Enemy dodge, parry and block

- The user OK'd it (2026-10-09). Dodge is 5% plus 0.1% for each point of the mob's defense over our weapon skill, skill from gear included: 6.5% against a level 63 boss, 6% against Vishas, and less against lower level mobs. Mobs dodge from any side. Parry is 5% plus 0.1% a point of defense over our level times 5 (gear skill doesn't count), or 0.6% a point when that gap is more than 10: 14% at +3 levels, 6% at +2. Block is a flat 5%. Parry and block only happen in front of the target (4.3).
- Forever's expertise-like stat comes off dodge and parry, 1% for each 1% of it. The item import reads 10 rating as 1%, from wowhead's tooltip, and the client has "Increased Expertise 1/2/3" auras for 1 to 3%. It's on tank sets and a few weapons (Dwarven Tree Chopper 0.6%, Servomechanic Sledgehammer 1%), and no enhancement set has any.
- Blizzard gave 6.5% dodge and 14% parry for a mob 3 levels up in 2019 Classic. magey's 2019 tests fit the dodge formula and 5% block for higher level mobs. Their parry at +1 and +2 levels (5.75 +- 1.0% and 6.55 +- 1.1%) fits our 0.1% a point, and 0.2% too.
- Decision (the user, 2026-10-09): the Level 30 encounter stays behind Vishas, even though we tank him. From the front Level 30 + 5 does 185.9 instead of 199.0, mostly from his 6% parry. PvP mode still lets the target dodge from behind, and we leave that too, because the Level 30 PvP preset attacks from the front.
- Fixed (2026-10-09): glancing went negative against mobs 2 or more levels below us (-40% at level 30 against a level 25). On the one roll white table that ate the block and crit chances after it, so white hits there never blocked and almost never crit. We clamp it at 0 now (`target.go`). The SoD sim has the same gap. No preset fights a lower level mob, and no test suite result changed. It does change sim numbers for the beta's low level mobs.
- `avoidance_test.go` TestOrcShamanEnemyAvoidance rolls the white table and Stormstrike's table 200,000 times each from the front in 8 cases: level 60 against a level 63 boss with Dark Edge of Insanity, with Huge Thorium Battleaxe's 10 skill and with 2% expertise, and level 30 against levels 25, 30, 31 and 32, once with Dwarven Tree Chopper (2 skill, 0.6% expertise). It checks dodge, parry and the 5% block. It fails when gear skill lowers parry, when gear skill doesn't lower dodge, with 0.2% parry a point, when expertise doesn't lower parry and without the glancing fix.

### 4.6 Glancing blows

- The user OK'd it (2026-10-09). Only white hits glance (auto attacks and Windfury Totem's extra attack), only against mobs, and from any side. Yellow attacks never do (4.3), and nothing glances in PvP mode. The chance is 10% plus 2% for each point the mob's defense is above our level times 5, from 0 to 40% (`target.go` NewAttackTable). Skill from gear doesn't lower it. A glancing blow deals a random part of the hit between 1.3 - 0.05 * (defense - skill), at most 0.91, and 1.2 - 0.03 * (defense - skill), from 0.2 to 0.99. Here skill from gear counts. Glancing comes before crit on the one roll, so a glancing blow can't crit.
- Level 60 with 300 skill against a level 63 boss: 40% for 55 to 75% of the hit (65% on average). With Huge Thorium Battleaxe's 310 skill: 91 to 99%. Level 30 against Vishas (level 32): 30% for 80 to 90%. The same level: 10% for 91 to 99%. Mobs below our level: none.
- magey's 2019 Classic tests measured 10.1, 20.3, 29.2 and 40.4% for 0 to 3 levels up, and about 5, 5, 14 and 35% damage lost (27% and 15% at 3 levels up with +2 and +5 skill). Beaza's formulas, which the sim uses, give the same. Their chance takes the lower of our level times 5 and our skill, so a weapon skill below the cap glances more. The sim always uses level times 5 (4.8), and every preset has the cap. The SoD sim has the same formulas without the 40% cap and the 0 floor. Mages, priests and warlocks get a harsher damage penalty in both sims, and the shaman uses the melee one.
- Beta (2026-10-10, our count from the beta session's Shimmering Flats log, WoWCombatLog-101026_150733.txt, up to the end of the Flametongue Weapon run): a level 30 Orc with Rage of the Storm (a two-handed mace) glanced on 13 of 37 white swings against level 33 mobs (35%), 66 of 111 against level 34 (59%) and 6 of 8 against level 35. The mob levels come from the level field on the mobs' own swings. The +3 rate fits our 40%. At +4 the sim's 40% cap is 4.2 standard errors under the 59%, and the formula without the cap gives 50% (2.0 standard errors). The user's two-handed mace skill was 149 of 150 in the 14:50 export. If the chance goes by that skill, it adds 2% at every level (52% at +4, 1.6 standard errors). No preset fights a mob more than 3 levels up, so the cap changes no sim result. The user keeps the cap until we have more data (2026-10-10, and again after the 111 swings at +4), and the beta session counts glancing by mob level in every new log.
- Beta, mobs below our level (the beta session's Thousand Needles run, cc172ba0): none of 61 white swings on level 27 to 29 mobs glanced. The sim gives none below our level.
- Beta, level 34 again (the beta session, 7ae7dd85, Sparkleshell Snappers 18:57 to 19:06): 10 of 21 glanced, so 76 of 132 (58%) in all at level 34.
- Beta, the whole log up to 20:09 (the beta session, 8515a72c): glances out of all white swings were 0 of 175 on level 24 to 28 mobs, 1 of 68 at 29, 0 of 7 at 30, 3 of 11 at 32, 23 of 57 at 33, 102 of 187 at 34 and 6 of 8 at 35. The first glance below our level came on a level 29 Thundering Boulderkin at 19:58. The sim never glances below our level and stops at 40%. One rule fits every row: the chance from our real Two-Handed Maces skill (149, not 150) with no 40% cap, so 2% at -1, 32% at +2, 42% at +3, 52% at +4 and 62% at +5. No sim change yet, because the user waits for more data. The next run is about 150 swings at +2 and about 100 at +3. The level field on our own swings' advanced log lines is our item level (33), not our level, so mob levels have to come from the mobs' own lines.
- Beta, up to 20:18 (the beta session, 905e293a): 1 of 9 white swings glanced at +1 (level 31) and 20 of 50 at +2, so 23 of 60 (38%, 95% range 27 to 51%) at +2 over the whole log. Both the sim's 30% and the real skill rule's 32% fit. No sim change.
- Beta, the whole log up to 20:25 (the beta session, 9435c80b): 3 of 33 at +1 (9%), 25 of 72 at +2 (35%) and 26 of 66 at +3 (39%). The sim's 30% and 40% fit, and so do the real skill rule's 32% and 42%. Telling those apart would take about 2000 swings a level. +1 is low for both rules (3 or fewer of 33 comes 1 time in 12 at 20%). No sim change. The beta session asks the user whether to stop collecting at +2 and +3.
- Parked (the user, through the beta session, da1c40e8): the glancing data is good enough for now, with no sim change.
- Beta, low skill (the beta session, 2ae6ffc0): with Bloody Brass Knuckles (a fist weapon, 1.6 speed) at Unarmed 85 of 150, all 8 landed swings on a level 29 mob glanced. The sim's chance uses our level times 5 in place of our skill, so it gives none there. So the game goes by our real skill when it's under the cap. That only matters with an untrained weapon skill, so no sim change for now. The sheet's melee crit fell by exactly (149 - 85) * 0.04 = 2.56%, so the sheet counts skill under level times 5 too, and fist weapons use Unarmed.
- `glancing_test.go` TestOrcShamanGlancing rolls the white table 20,000 times in 8 cases (level 60 against a level 63 with 300 and 310 skill, level 30 against levels 29 to 34, once with Dwarven Tree Chopper's 152 skill) and checks the glance chance and the lowest and highest glancing damage. It fails when gear skill lowers the chance, when gear skill doesn't help the damage, without the 40% cap and without the 0.91 cap on the low end. All the shaman mechanics tests (TestOrcShaman*) take about 0.3 sec together.

### 4.7 Crit

- The user OK'd it (2026-10-09). Our melee crit is one stat with spell crit under Forever: base crit, Agility, gear, buffs, Thundering Strikes and the Orc axe racial (+1% with an axe, on every attack). Against a mob above our level we lose 0.2% for each point of its defense over our level times 5 (1% a level), and a flat 1.8% more when it's 3 or more levels up. Skill from gear doesn't help. Against a mob below our level we gain 0.04% a point. So we lose 4.8% against a level 63 boss and against a level 33, 2% against Vishas, and gain 1% at level 30 against a level 25 (about 4% against the beta's level 9 boars). White hits only crit in what miss, dodge and glancing leave (45.5% plus our hit from behind against a boss). Yellow attacks roll crit on their own (4.3). A melee crit deals twice a normal hit, as the beta logs show.
- Blizzard's 2019 Classic posts give 1% less crit a level and one more cut on crit from auras against mobs 3 levels up, which magey measured at 1.8%. The SoD sim has the same code as ours. Blizzard's cut only comes off crit from auras (talents, Equip: crit on gear, buffs, consumables), not base crit or Agility, and the sim takes it off all crit (the TODO in `target.go`). That only differs with less than 1.8% crit from auras. Level 60 builds have Thundering Strikes and Leader of the Pack, and Vishas is 2 levels up, so we keep it. Crit lost to weapon skill under level times 5 is 4.8.
- `crit_test.go` TestOrcShamanMeleeCrit checks the crit chance of white hits and Stormstrike against levels 25 to 34 at level 30 and against a level 63 at 60, once with Huge Thorium Battleaxe's 310 skill, and that white and Stormstrike crits deal twice a normal hit. It fails without the 1.8%, when gear skill adds crit, without the bonus against lower mobs and with 1.5x melee crits. The crit cap is in TestOrcShamanAttackTable.

### 4.8 Weapon skill

- The user OK'd it (2026-10-09). Our base skill is our level times 5, a fully trained weapon, and skill from gear adds to it by weapon type, one-handed and two-handed apart (`target.go` GetWeaponSkill). No racial gives skill under Forever (the Orc axe racial is crit, 4.4). Skill from gear lowers miss (4.4) and dodge (4.5) and raises glancing damage (4.6). It doesn't change parry (4.5), the glancing chance (4.6) or crit against mobs (4.7).
- Fixed (2026-10-09, the user): Forever made weapon skill a plain item stat with much smaller amounts, and moved some of it to the expertise-like stat. The client's ItemSparse (70291) has stat type 90 for two-handed axes, 91 for two-handed maces and 96 for daggers, and wowhead's Forever tooltips give the amounts. Our item import kept Classic's skill, because the listing's jsonequip has none of it. `forever_wowhead.go` foreverWeaponSkills now sets it on the 7 Forever items that had Classic skill: Dwarven Tree Chopper none (0.6% expertise only, Classic +2), Servomechanic Sledgehammer +1 two-handed maces (Classic +7), Skilled Fighting Blade +1 daggers (+4), Huge Thorium Battleaxe +2 two-handed axes (+10), Flawless Arcanite Rifle none (+4 guns), Death's Sting +3 and The Hungering Cold +6 (unchanged). MergeItem replaces an item's weapon skill instead of adding to it, and gen_db prints any Forever item that keeps Classic skill without an entry. The other 17 Classic skill items (Edgemaster's Handguards and the rest) aren't in the Forever client or on wowhead Forever yet, so they keep Classic's. No preset uses any of them, and no test suite result changed.
- Left alone (the user, 2026-10-09): skill under level times 5. The beta's sheet loses 0.04% crit per missing point, and 1.12 also adds miss, dodge and glancing, but the sim and the presets assume a trained weapon. Also, 2019 Classic uses 0.04% a point of skill difference against players, and PvP mode uses the mob formulas. At the same level with trained skill both give the same, and no PvP set has skill gear.
- Tests: the 4.4 to 4.7 tests give skill through the player's bonus two-handed axe skill instead of items, so item changes don't move them. TestOrcShamanMissChance checks Huge Thorium Battleaxe's +2 (7.6% miss against a level 63) and TestOrcShamanEnemyAvoidance Dwarven Tree Chopper's 0.6% expertise and no skill. Both fail on the old item data.

### 4.9 Enemy armor

- The user OK'd it (2026-10-09). Our physical damage (white hits, Windfury and Stormstrike) loses armor / (armor + 400 + 85 * our level) (`spell_resistances.go` GetArmorDamageModifier), the 2019 Classic formula, like the wowsims Classic and SoD sims. Fire and nature procs and spells ignore armor (5.2). The armor debuffs take the rank the raid's level allows (`debuffs.go`): Sunder Armor 90 / 180 / 270 / 360 / 450 a stack for 5 stacks, Expose Armor the same per combo point for 5 points (Improved Expose Armor adds nothing under Forever), Faerie Fire and Curse of Recklessness 175 / 285 / 395 / 505. One of Sunder Armor and Expose Armor counts, and one of Faerie Fire and Curse of Recklessness (the user, 2026-10-08). The Level 60 boss's 3731 armor goes to 976 with them (15.1% less damage, 40.4% without), and Vishas's 1063 takes 26.5% with no debuffs.
- Client 70291 has these values and learn levels exactly (Sunder Armor 10 / 22 / 34 / 46 / 58, Expose Armor 14 / 26 / 36 / 46 / 56, Faerie Fire 18 / 30 / 42 / 54, Curse of Recklessness 14 / 28 / 42 / 56). Curse of Recklessness's old attack power is a dummy aura now. Expose Armor's points are 0 in the client, because the server works them out from combo points. All four use a new aura, "(DNT) Mod Armor No Mods", flat armor that armor percentage modifiers don't scale (Devotion Aura and Mark of the Wild use it too). The boss's and Vishas's armor are unverified (the boss in Need to Verify, At 60).
- `armor_test.go` TestOrcShamanArmor checks the target's armor 5 sec in and a normal white hit's damage in 9 cases: Vishas at 30, the boss at 60 with no debuffs, each debuff alone at 60, all four at 60 and at 30 (rank 2 of each), and Expose Armor alone at 30. It fails with 80 a level in the formula, when Curse of Recklessness stacks with Faerie Fire and when Sunder Armor takes its top rank at 30.

### 4.10 Physical damage modifiers

- The user OK'd it (2026-10-10). A white hit, Windfury attack or Stormstrike starts from the weapon roll plus attack power (4.1). Flat "+N damage" from our gear (Bogling Root, Zandalarian Hero Medallion, the Might of Cenarius set) adds to it, and then our percent bonuses multiply the total (`spell_result.go` CalcDamage). Armor comes after (4.9), then flat "+N damage taken" on the target (Gift of Arthas only), then crit or glancing, which also scale the flat part. 2019 Classic and the SoD sim have the same order.
- Rage of the Storm (280604, a level 30 two-handed mace) is the only percent bonus on our physical damage: +10% on Stormstrike, added to Stormstrike's other percent bonuses (it has none today). The other shaman multipliers are on spells and totems. Orcs have no damage racial under Forever, Sanctity Aura is gone, our raid buffs give no physical percent bonus and world buffs are off. Gift of Arthas is off in our presets and needs level 45 under Forever. We add its +8 after armor, and no official source we found settles that. Weapon Mastery's +10% stays in Need to Verify. Rage of the Storm's beta check stays in Need to Verify (Beta).
- `damage_modifiers_test.go` TestOrcShamanDamageModifiers rolls a 100 damage hit with +10 flat damage against no armor, with Whirlwind Axe and with Rage of the Storm. White hits and Windfury deal 110 with both, and Stormstrike deals 110 and 121. It fails when Rage of the Storm gives 20%, when it reaches every attack and when the flat bonus comes after the percent bonuses.

### 4.11 Procs from melee hits

- The user OK'd it (2026-10-10). Each proc effect lists the kinds of hits that can trigger it: white hits, yellow melee attacks (Stormstrike and Windfury attacks), spells, and procs (the Flametongue and Frostbrand damage). Only a landed hit rolls: a hit, crit, glancing blow or block. A PPM proc's chance per landed hit is PPM * weapon speed / 60 (`attack.go` PPMManager). The speed is the weapon's own, so haste only adds swings, and yellow attacks get the same chance (not Stormstrike's normalized speed). A chance proc rolls its flat chance per landed hit. Windfury attacks can trigger what a Stormstrike can, but not Windfury again, because of its 1.5 sec cooldown. Melee procs ignore the Flametongue and Frostbrand damage. What each ability procs is 4.13.
- Client 70291 (wago SpellItemEnchantment and SpellAuraOptions): Windfury Weapon's proc aura (439431, SoD's spell) has 20% and a 1500 ms cooldown, Flametongue Weapon (436519) 100% and Crusader (458112) 100%, all on white hits and melee abilities. The client has no PPM for Crusader, and no spell uses its PPM table, so the client can't tell old weapon speed PPM from the modern haste based kind. 2019 Classic (magey's tests) and the SoD sim use the weapon speed.
- The user remembers one Flametongue Totem proc when Windfury procs, not three (swing and both attacks). Settled in 4.13 part 1: Flametongue Totem procs on white swings only.
- Fixed (2026-10-10, the user's go): a Crusader proc gave 100 - 4 * (level - 60) Strength, so 220 at level 30. It gives 100 up to level 60 now, like the client's Holy Strength (20007) and 2019 Classic twinks (`enchant_effects.go`). Our level 30 enchant search doesn't have Crusader, and no golden changed.
- `proc_test.go` TestOrcShamanMeleeProcs puts Crusader on a level 30 shaman's Whirlwind Axe (6% a landed hit) and rolls 40k white hits, Stormstrikes and Windfury attacks with 50% attack speed. Each procs on 6% of landed hits (4 standard errors), misses and dodges never proc, and a proc gives 100 Strength. It fails with the old Strength formula, when misses proc, when only white hits proc and with PPM / 50.

### 4.12 Extra attacks

- The user OK'd it (2026-10-10). An extra attack is a normal white swing with the main hand that restarts the swing timer (`attack.go` swing and ExtraMHAttack), as in 2019 Classic (magey's tests) and the SoD sim. When a white swing triggers it, the extra swing comes when that hit lands, 10 ms later, and the next swing moves back by those 10 ms. When a yellow attack triggers it mid-swing, the extra swing comes right away and the next one a full swing after it. Windfury Weapon's attacks are yellow attacks, not extra attacks, so they leave the swing timer alone (4.4).
- Windfury Totem procs on 20% of landed melee hits, white or yellow, with a 1.5 sec cooldown that stops it from proccing off its own extra swing (`buffs.go` CreateExtraAttackAuraCommon). In Classic only main hand hits count, under Forever off-hand ones too (below). The extra swing gets 95 / 179 / 246 AP from a buff with 1 charge (2 when a yellow attack triggered it) that white swings use up. Our own Windfury Weapon blocks the totem, so in our presets it only reaches the group, and the totem needs level 32. Hand of Justice procs on landed melee hits, white or yellow, with a 2 sec cooldown. It's in Level 60 Phase 1 and Synthetic. The level 30 presets have neither.
- The Era client (1.15.9) has Hand of Justice at 2% with the 2 sec cooldown, and Windfury Totem's buff at 1.5 sec with 2 charges that only white swings use. The Forever client (70291) differs in two places. Hand of Justice has a 3% proc chance and a second value of 3, and wowhead Forever's tooltip says "1% chance on Melee hit to gain 1 extra attack. Attacks against Dwarves are 3 times as likely to activate this effect." Windfury Totem's buff lasts 1 sec.
- Changed (2026-10-10, the user's go): under Forever Hand of Justice procs on 1% (we take the tooltip as right and the server keeps a third of the 3% procs against anyone but Dwarves, our targets have no race) and Windfury Totem's attack power lasts 1 sec (`item_effects.go`, `buffs.go`). Classic keeps 2% and 1.5 sec. The goldens with Hand of Justice drop (TestEnhancement 1245.5 to 1228.8 on one Phase 1 config, TestP1DPSWarrior 1568.0 to 1532.8), and the rogue ones move a little. We refreshed them. The Level 60 presets weren't rerun.
- Need to Verify (the user, 2026-10-10): whether Windfury Weapon's attacks restart the swing timer (Beta), and whether a Windfury Totem extra attack does (level 32).
- Beta (the beta session, 2026-10-10, `beta_results.txt` commit fa8cd7d1): Windfury Weapon's attacks don't restart the swing timer, as in the sim. After 3 Windfury procs from a Stormstrike with no Flurry change around them (15:09:05.743, 15:30:21.033 and 15:32:16.113), the next white swing came on the old schedule (the swing before plus 3.3 sec) within 36 ms. A restarted timer would have put it 0.9 to 2.8 sec later. The Windfury Totem question stays (level 32).
- Client 70291 and 70338 (found by the beta session, 2026-10-10): Windfury Totem's proc has a 100 ms cooldown, not 1.5 sec. The proc auras of both totems aren't tied to a weapon (no SpellEquippedItems row), so a dual wielder's off-hand hits can trigger them. The SoD and Era client (1.15.9.70003) has the totem as a main hand enchant (1783, 563, 564) that procs its buff on 20% of hits, with no cooldown anywhere in the client. The sim's 1.5 sec came with the SoD sim's "ws wf rework" (2024-09-26) with no source given, and it matches the 1.5 sec the buff lasted in Era. The cooldown is in Need to Verify (level 32), and the sim keeps 1.5 sec until we can check it (the user, 2026-10-10).
- Changed (2026-10-10, the user's go): under Forever a dual wielder's off-hand hits proc Windfury Totem too, as the user expects. The extra swing is still a main hand one. Need to Verify (level 32). Our shaman can't dual wield, but the rogue and warrior tests put Windfury Totem on a dual wielder, so their goldens went up (TestCombatSinisterStrike 868.6 to 890.5 on the first config, TestP1DPSWarrior 1532.8 to 1555.5). We refreshed them. `rogue/dps_rogue/totem_procs_test.go` TestOrcRogueWindfuryTotemOffHand counts the procs from each hand of a dual-wielding level 60 rogue over 5 min (15 main hand, 18 off hand). It fails with the main hand only mask.
- `extra_attacks_test.go` TestOrcShamanExtraAttacks gives a level 30 shaman with Whirlwind Axe an extra attack from its first white swing and one from a Stormstrike at 5 sec, casts Windfury Weapon's two attacks at 10 sec, and checks the swings come at 0, 0.01, 3.61, 5, 8.6, 12.2 and 15.8 sec. It counts Hand of Justice procs over 36000 sec of white swings (1%, 4 standard errors) and checks that Windfury Totem's attack power lasts 1 sec. It fails with Hand of Justice at 2%, with the buff at 1.5 sec and when an extra attack waits a batch window.

### 4.13 Proc matrix, part 1: imbues and weapon totems

- The user OK'd it (2026-10-10). Windfury Weapon (20%, 1.5 sec cooldown), Windfury Totem (20%, 1.5 sec cooldown, Windfury Weapon blocks it), Flametongue Weapon (every hit) and Frostbrand (9 PPM) need a landed main hand melee hit, white or yellow. Flametongue Totem (every hit, Flametongue Weapon blocks it) needs a landed white swing since the fix below. A blocked hit counts as landed. Misses, dodges, parries, spells (shocks, Lightning Bolt, Chain Lightning, Lava Burst, Fire Nova), Searing and Magma Totem attacks, Lightning Shield orbs and the Flametongue and Frostbrand damage never trigger them. A Windfury proc's two attacks can't trigger Windfury again (the cooldown). Flametongue Totem and Windfury Totem share the weapon's one totem slot, so they never meet.
- Client 70291: Windfury Weapon (439431) 20% with 1500 ms and Flametongue Weapon (436519) 100%, both on white hits and melee abilities (SoD's spells, the Era client 1.15.9 has the same). Frostbrand uses the old enchant kind (a server PPM), so the server decides which hits trigger it. We first read the totems as the old kind too, but that was the Era client. The beta session found (2026-10-10) that in Forever (70291 and 70338) the totem buffs are party auras that trigger a proc spell. Flametongue Totem (8230, 8250, 10521, 15036) has 100%, no cooldown and white swings only (ProcTypeMask 0x4). Windfury Totem (8515, 10609, 10612) has 20%, a 100 ms cooldown, and white swings and melee abilities (0x14). The Era client (70003) still has them as main hand enchants. magey's 2019 Classic tests have melee abilities triggering Windfury Totem. The SoD sim uses the same masks for the imbues.
- The October 8 beta notes (Blizzard's WoW Forever Beta Development Notes): "Elemental Focus is no longer activated and immediately consumed when using Fire Nova." The sim already lets Fire Nova's Clearcasting proc stay for the next spell, and TestOrcShamanClearcasting checks it (2.7). The other October 8 shaman notes (Lightning Bolt and Lava Burst ranks, Elemental Fury and Elemental Alacrity swapped, Rage of the Far Seer without cast speed, Flametongue Totem not stacking with Flametongue Weapon or Windfury Totem) are in the sim too.
- Fixed (2026-10-10): on the beta Stormstrike never procced Flametongue Totem (the user), as the client's white swings only mask says. Under Forever it now procs on landed white swings only, extra swings included (`totem_weapon_buffs.go`). Windfury Weapon's attacks weren't tested on their own, but they're melee abilities like Stormstrike, and the user remembers one proc per Windfury proc. Classic keeps every main hand hit, as the Era enchant did. TestEnhancement drops 1 to 2% (1095.9 to 1081.6 on the first config), and we refreshed it. The Level 30 presets weren't rerun.
- Changed (2026-10-10, the user's go): under Forever Flametongue Totem procs on off-hand white swings too, like Windfury Totem (4.12). We assume an off-hand proc goes by the off hand's speed. Only other classes in our group gain from it, and no golden has Flametongue Totem on a dual wielder. TestOrcRogueFlametongueTotemOffHand (`rogue/dps_rogue/totem_procs_test.go`) checks every normal proc of a dual-wielding level 30 rogue over 2 min against the speed of the hand that swung (32 main hand, 46 off hand). It fails with the old mask. Need to Verify.
- Beta (the beta session, 2026-10-10, the same Shimmering Flats run): Flametongue Totem procced on 70 of 70 landed swings (glancing blows, blocks and crits included) and on 0 of 21 misses, dodges and parries, as in the sim. Each of the 12 Windfury Weapon procs from a white swing brought exactly one Flametongue Totem proc, and Stormstrike brought none, as the fix above has it. The proc came 0 to 10 ms after the swing 55 times and 90 to 127 ms after it 15 times.
- Beta (the beta session, 2026-10-10, `beta_results.txt` commit ad8ec9a5): Flametongue Weapon procced on 30 of 30 landed swings (4 blocked) and 12 of 12 landed Stormstrikes, and on none of 10 missed, dodged or parried swings and 8 parried Stormstrikes. That matches the sim, and the beta item is closed.
- `proc_matrix_test.go` TestOrcShamanImbueProcs has four setups (Windfury Weapon with Flametongue Totem, Flametongue Weapon with Flametongue Totem, Frostbrand at 30, Flametongue Weapon with Windfury Totem at 60). It forces each outcome on white hits and Stormstrikes, and hits with the spells, totem attacks, Lightning Shield orbs, the proc damage and Windfury Weapon's attacks, 60 times each 2 sec apart. It checks what each one triggered against the table, and that a Windfury proc always brings exactly 2 attacks. It fails when Flametongue Totem procs on misses, when Stormstrike or Windfury Weapon's attacks proc Flametongue Totem, when Flametongue Weapon doesn't block it, when spells trigger Windfury Weapon and without Windfury's cooldown.

### 4.13 Proc matrix, part 2: talent procs

- The user OK'd it (2026-10-10). Flurry: a melee crit (white hit, Stormstrike or Windfury Weapon attack) gives 3 charges of 5 to 25% attack speed. White swings use them, extra swings included, at most one per 0.5 sec, and a white crit uses one and then gives all 3 back. Windfury Weapon's attacks and Stormstrike don't use them. Maelstrom Weapon: any landed melee hit can give a stack (2 a minute a point, a guess), up to 5 for 30 sec. Lightning Bolt uses them all and Chain Lightning none. Lightning Bolt triggers it at half the chance only with Forever's Totem of the Storm. Elemental Devastation: a spell crit (shocks, Lightning Bolt, Chain Lightning, Lava Burst, Fire Nova) gives 3 / 6 / 9% melee crit for 10 sec. Clearcasting: casting one of those spells gives it 10% of the time, for 15 sec, and the next one is free. A spell never uses up its own proc. None of the four come from the Flametongue and Frostbrand damage, Searing and Magma Totem attacks or Lightning Shield orbs.
- Client 70291: Flurry (16256) triggers on white hits and melee abilities, 5 to 25% by rank. Its buff (16257) has 3 charges that only auto attacks use, 500 ms apart, and lasts 15 sec (the Era client 1.15.9 has the same). Clearcasting's (16246) spell family mask covers exactly Lightning Bolt, Chain Lightning, the three shocks, Lava Burst and Fire Nova. The Maelstrom Weapon buff (408505) lasts 30 sec and changes Lightning Bolt only, the talent (408498) is -4% a point and can trigger from offensive spells too (for Totem of the Storm). Elemental Devastation is 3 / 6 / 9%, its buff (30165) 10 sec, triggered by offensive spells.
- Fixed (2026-10-10, the user's go): Flurry's buff now runs out after 15 sec, for every ruleset, as in both clients. We kept it until the charges were gone. That only showed when white swings stop for long, like out of melee range in PvP. No golden changed.
- Open (Need to Verify): Maelstrom Weapon's proc rate (level 34+), and whether the Flametongue procs trigger Elemental Devastation (Beta).
- Beta (the beta session's Elemental respec, 67c6efc8, 2026-10-10): 4 Fire Nova crits and 1 Flame Shock tick crit gave no Elemental Devastation, and the log would show its buff (30165 has no hide or no-log flag in client 70338). The talent (30160) triggers on any harmful spell through a server script, and Fire Nova's damage (408424) is triggered by the cast's dummy effect. So either Fire Nova's damage doesn't give it (the sim says it does) or the talent doesn't work at all. A shock or Lightning Bolt crit tells them apart. No sim change until then.
  - Up to 21:25 (49b5cce8): Elemental Devastation works on direct spell crits. Both direct Flame Shock crits gave the buff (30165) for 10.0 sec. Nothing else gave it: 3 Flametongue Totem crits and 2 Flame Shock tick crits (as the sim has it), and 4 Fire Nova crits (the sim gives the buff on any spell damage crit, Fire Nova included, `talents.go`).
  - Proposed to the user: Elemental Devastation leaves out Fire Nova.
  - Changed (2026-10-10, the user's go, "fire nova cant trigger elemental devastation"): Fire Nova's crits no longer give the buff (`talents.go` applyElementalDevastation). Fire Nova was 0 of 5 crits by then (the beta session, 2ae6ffc0). The Level 60 preset went from 759.3 to 753.2 DPS (-6.1, the same settings as the 5.6 change). The enhancement goldens fell 0.45% on average (0.05 to 1.4%), because their talents have Elemental Devastation 3/3 and the rotation casts Fire Nova. `talent_procs_test.go` TestOrcShamanTalentProcs now wants no Elemental Devastation from a Fire Nova crit.
  - Up to 22:16 (9294818f): every direct shock crit gave the buff, 9 of 9 since the respec (Earth, Flame and Frost Shock). No Fire Nova crit (0 of 7), Flametongue Totem crit (15 with the buff down) or Flame Shock tick crit gave it. The sim matches since 620a81c2.
  - Up to 21:10 (22f8e477): still no buff. The Flame Shock tick crit doesn't test it, because the talent's trigger (0x10000) is direct spell damage and periodic damage has its own flag (0x40000). The direct shocks had 0 crits in 8 hits, and no Lightning Bolt yet.
- Beta (the same run): Flametongue Totem procced on 2 swings that a player's shield fully absorbed (ABSORB), with the 24 of 24 landed swings in range.
- `talent_procs_test.go` TestOrcShamanTalentProcs, on a level 60 shaman with all four talents: forces hits and crits on white hits, Stormstrikes, Windfury attacks, every damage spell, the Flametongue Totem damage, Searing Totem attacks and Lightning Shield orbs, 60 times each, and checks what triggers Flurry, Maelstrom Weapon and Elemental Devastation. It checks which spells cost nothing under Clearcasting, Flurry's charges (white swings only, 0.5 sec apart, a crit refills) and its 15 sec, and that Chain Lightning keeps Maelstrom Weapon's stacks and Lightning Bolt uses them. It fails when spell crits give Flurry, when proc damage gives Elemental Devastation, when Flurry never runs out, when yellow attacks use its charges, without the 0.5 sec gap, when Chain Lightning uses Maelstrom Weapon and when Clearcasting covers every shaman spell.

### 4.13 Proc matrix, part 3: Judgement of Wisdom, shields, items and racials

- The user OK'd it (2026-10-10). Judgement of Wisdom on the target gives us mana (59 at rank 3) on half of our white hits, even the misses, and on half of our landed Stormstrikes, Windfury Weapon attacks and spells (`debuffs.go`). Proc damage, totem attacks and Lightning Shield orbs never give it. Lightning Shield and Water Shield only take hits on us, at most one orb or globe every 3.5 sec. Our attacks never touch them. The melee item procs (Crusader, Hand of Justice, Sulfuras) trigger on landed white hits, Stormstrikes and Windfury Weapon attacks, never on spells or proc damage, and their own proc damage triggers nothing. Revelation triggers on landed direct spells that don't crit (whether only shocks count is in Need to Verify). The Orc has no proc racials.
- Client 70291 (Era 1.15.9 the same): Judgement of Wisdom triggers when the target takes melee swings, melee abilities, ranged attacks and harmful spells. The 50% and the white misses come from the server (Era, the SoD sim). Every Lightning Shield rank and Water Shield (408510) have 3 charges and a 3500 ms proc cooldown, so we took "Water Shield globe ICD" and "Lightning Shield orb ICD" off the Need to Verify list and updated the code comments.
- Fixed (2026-10-10, the user's go): the client's Lightning Shield also fires on ranged attacks and harmful spells we take (0x222a8, no damage over time ticks), and the sim only fired it on melee hits. The user: shields pop on any damage from an enemy, AoE included, and the raid damage hits per minute stand in for the encounter damage we don't sim. Both shields now go off on any direct hit an enemy lands on us, and the raid damage stream (`shields.go`) fires Lightning Shield orbs at our target as well as spending Water Shield globes. No preset sets the raid damage rate, so no golden changed.
- `shields_test.go` TestOrcShamanShieldTriggers lands an enemy melee hit, a melee miss, a ranged hit, a spell hit, a damage over time tick and our own spell on a fresh Lightning Shield and Water Shield, and checks which ones use a charge. TestOrcShamanRaidDamageLightningShield checks that 12 raid hits a minute fire 3 orbs at the target in 15 sec. They fail when Lightning Shield needs melee hits, when our own damage sets the shields off and when the raid damage only feeds Water Shield.
- `other_procs_test.go` TestOrcShamanOtherProcs puts Crusader and Hand of Justice on a level 60 shaman with Judgement of Wisdom on the target, and forces hits and misses on white hits, Stormstrikes, Windfury attacks, the six damage spells, the Flametongue Totem damage, Searing Totem attacks and Lightning Shield orbs, 2000 times each 2 sec apart (Hand of Justice's 1% with its cooldown). It queues one roll at a time, because 32000 queued actions took the sim 14 sec. It fails when Judgement of Wisdom needs white hits to land, when it comes from proc damage and totems, when Crusader procs from spells and when Hand of Justice only procs from white hits.
- The shaman tests share `weaponSim` (`damage_modifiers_test.go`) for these setups now.

### 5.1 Spell hit

- The user OK'd it (2026-10-10). A spell misses a target 0, 1, 2 and 3 levels above us 4, 5, 6 and 17% of the time, then 11% more a level, at most 99% (`target.go` spellMissChance). Our spell hit comes off it, and it never goes under 1% (`spell_result.go` SpellChanceToMiss). A binary spell lands on (1 - base miss) times its binary hit chance, plus our hit. Under Forever the hit on our gear counts for spells and melee alike (`ruleset.go` unifyEquipHitAndCrit). Tidal Focus gives 1% spell and melee hit a point (Forever text). Nature's Guidance is gone under Forever. Tauren get 1%, Orcs nothing.
- Spells that roll for hit: the shocks, Lightning Bolt, Chain Lightning, Lava Burst, Fire Nova, Searing and Magma Totem attacks, and the Flametongue Weapon, Flametongue Totem and Frostbrand procs. Flame Shock's ticks don't roll, and a missed Flame Shock puts no DoT on. Lightning Shield orbs always hit (until the change below). Our targets: Vishas is level 32 (6%), the Level 60 boss 63 (17%), PvP enemies our own level (4%).
- Fixed (2026-10-10, the user's go): below our level the sim gave 4% too. It's 3, 2 and 1% at 1, 2 and 3 levels below now, as in 1.12. None of our encounters has a lower level target, so no golden changed.
- Kept (the user): Searing and Magma Totem attacks use our spell hit and crit. PvP's 7% a level past 3 levels above us stays out, because our PvP enemies are always our level.
- Beta (the beta session, 2026-10-10, the same Shimmering Flats run): 14 of 71 orbs missed (20%) on level 33 to 35 mobs, and none of the 57 hits crit with 12.3% spell crit (1 time in 1800 by chance). So orbs roll spell hit and never crit. A missed orb still used up its charge. A later run brought the count to 101 orbs, 19 misses and 0 crits.
- Changed (2026-10-10, the user's go, made by the beta session in 9cf69944): under Forever an orb rolls spell hit and never crits (`lightning_shield.go` OutcomeMagicHit). Classic keeps the old always hit. A missed orb uses up its charge, as before. The user calls it settled for now and wants it rechecked with every new log (`beta_handoff.md` "Recheck with every new log"). No golden changed. `shields_test.go` TestOrcShamanLightningShieldOrbOutcome checks it, and TestOrcShamanRaidDamageLightningShield now counts orbs fired (hits and misses), because orbs can miss the level 63 boss.
- Beta, whole log up to 20:18 (the beta session, 905e293a): orbs missed on 3 of 17 at +2, where Classic gives 6%. That comes about 1 time in 12, and the misses at +3 and +4 sit near Classic's rates. No change.
- `miss_test.go` TestOrcShamanSpellMissChance checks Lightning Bolt's miss chance from 5 levels below us to 5 above, against the level 63 boss with Tidal Focus 5/5 (12%), with Tarnished Elven Ring's 1% melee hit (16%), and the 1% floor with Tidal Focus at our level and with 20% hit against the boss. TestOrcShamanSpellMissRolls rolls Lightning Bolt 40000 times against the boss and gets 17% within 4 standard errors (0.75%). The first fails 4 cases with the old 4% below our level.

### 5.2 Resistances

- Shown to the user (2026-10-10), waiting for the OK. A target's resistance to a school lowers our damage of that school, after our Spell Penetration comes off it (`spell_resistances.go`). The resistance over 5 times our level is the fraction that counts. A binary spell is fully resisted 75% times that fraction of the time, in the same roll as its hit (5.1). Earth Shock is our only binary spell (`earth_shock.go`, SpellFlagBinary, as in the SoD sim). Every other spell, proc, totem attack and Lightning Shield orb can be partly resisted, 25, 50 or 75%, on the 2019 Classic table (royalgiraffe), which averages 75% times the fraction up to 2/3 of the cap. Each Flame Shock tick rolls on its own. A DoT with no direct part uses a tenth of the resistance, and we have none. Our physical damage uses armor instead (4.9).
- A target above our level adds 2% average partial resist a level to every spell that isn't binary, and nothing takes it off. That's 6% against the level 63 boss and 4% against Vishas. PvP enemies are our level, so they get none. No preset target has any resistance (the Level 60 boss, Vishas and the PvP enemy), so this level part is all that does anything in our sims.
- Cross-check: the royalgiraffe resist guide (2019 Classic data) has the same rules. Higher level NPCs gain 2% average mitigation a level (24 resistance at level 60 against 63), only against spells that aren't binary, and neither Spell Penetration nor curses remove it. It also has the binary roll, the tenth for pure DoTs and a new roll for each tick (since Classic phase 6). Its binary rule is any spell with an effect besides damage. By that rule Frost Shock and the Frostbrand proc are binary too, because they slow, and the sim has them as partly resistible. That only matters against a target with resistance, or with the level part.
- Beta (our count from the Shimmering Flats log, WoWCombatLog-101026_150733.txt, up to 16:06): not one partial resist in 257 landed hits of spells that aren't binary on level 33 to 35 mobs. The hits were 97 at level 33, 152 at 34 and 8 at 35, from Flametongue Weapon, Lightning Shield, Searing Totem and Flame Shock (direct and ticks). The sim expects about 57 partial resists there (18% of hits at +3, 24% at +4, 30% at +5). The damage shows it too. Flametongue Weapon always hit for its full base amount, and the orbs always for 64 or 65. So the beta has no level based partial resist, at least at level 30.
- Open: the beta mobs were all regular mobs. The Classic rule covers every NPC above our level, but almost all the data behind it is from raid bosses. If Forever kept it for bosses or elites only, we haven't seen it. A dungeon boss test is in notes.md Need to Verify (Beta), for when the user logs in dungeons (2026-10-10).
- Beta, mobs below our level (the beta session's Thousand Needles run, level 27 to 29, cc172ba0): no partial resist on any spell. Classic gives none below our level either.
- Beta, level 34 again (7ae7dd85): no partial resist in 56 more hits.
- Proposed to the user: under Forever, drop the level part (Classic keeps it). Our spell, proc, totem and orb damage would go up 6.4% against the level 63 boss and 4.2% against Vishas. We leave Frost Shock and Frostbrand as they are, because without the level part and with no target resistance, binary or not changes nothing. The user waits for the dungeon logs before the change (2026-10-10).
- Beta, a target with resistance (the beta session's count up to 19:32, 8b8e09fa, checked by us): five level 29 Elder Cloud Serpents resisted part of every Nature hit we landed and none of our Fire hits. All 9 orbs lost 20% or 30% (13 or 19 of 64), and all 4 Earth Shocks lost 30% (36 or 37 of 119 to 123). A fifth Earth Shock was resisted in full. The log shows it as RESIST, where our 11 Earth Shocks that failed the hit roll show as MISS. So the server resists by school resistance in 10% steps, not Classic's 25% steps. It also partly resists Earth Shock, which the sim treats as binary. The log's resisted field is in raw damage, before the Stormstrike mark's 20%. Nothing changes in the sim, because none of our targets has resistance. Earth Shock would only matter if we keep the level part: as a spell that isn't binary it would lose the level part's 6% against the boss too.
  - The respec run (21:10 to 21:25, 49b5cce8): the Elder Cloud Serpents again took 20% and 30% off our Nature hits.
  - Explained (786a09ed): in one Salt Flats Scavenger fight (22:12) every shock was 1 raw higher (about 2%). The user swapped gear then, and our cast lines log 53 spell power from 22:11:30 to 22:13:05, not 49. At 53, Flame Shock is (45.4 + 0.214 * 53) * 1.15 = 65.3 and its ticks (14 + 5.3) * 1.15 = 22.2, as logged. So it isn't a mob or resist effect.

### 5.3 Spell crit

- The user OK'd it (2026-10-10). Our spell crit is 2.3% base plus Intellect (0.0355% a point at 30, item 1.5), crit from gear (one pool with melee crit under Forever, `ruleset.go` unifyEquipHitAndCrit), Thundering Strikes (1% a point), the Orc axe racial (1%) and buffs. Talents that add crit to some spells only (Call of Thunder, Tidal Mastery) are section 8. The shocks, Lightning Bolt, Chain Lightning, Lava Burst, Fire Nova, Searing and Magma Totem attacks and the Flametongue Weapon and Frostbrand procs roll it (`spell_outcome.go` outcomeMagicHitAndCrit), after they land. Flametongue Totem's procs roll a flat 8% since 2026-10-10 (below). Lightning Shield orbs never crit (5.1). Flame Shock's ticks roll their own crit under Forever (5.7). A spell crit deals 1.5 times a hit. Elemental Fury adds 20% of that bonus a point (1.6 times at 1/1, section 8).
- Against a target above our level spell crit isn't cut. The sim works out a 0.3% and 2.1% cut at +2 and +3 (`target.go` SpellCritSuppression), but nothing uses it (`spell_result.go` SpellCritChance, "TODO: Classic verify crit suppression"). Melee crit is cut (4.7).
- Cross-check: the royalgiraffe resist guide (2019 Classic) rolls spell crit apart from hit and resists and mentions no cut against bosses. We found no source for a spell crit cut in Classic.
- Beta (our count from the Shimmering Flats log up to 16:21, with 12.3% spell crit on the sheet and no buffs that add crit): every crit was 1.5 times the hit (Searing Totem, Flametongue Weapon 27 of 18 and 50 of 33, Flame Shock ticks 34 and 35 of 23). Crits by mob level were 15 of 115 hits at level 33 (13.0%), 4 of 95 at level 34 (4.2%, 1 time in 150 by chance at 12.3%) and 1 of 5 at level 35. By spell, Searing Totem crit on 12 of 77 (15.6%) and the Flametongue Weapon and Totem procs on 5 of 103 (4.9%, 1 time in 100 by chance), with 1 of 63 at level 34. The shocks crit on 1 of 18. Most proc hits are at level 34 and most Searing hits at level 33, so we can't tell whether the low numbers come from the level or from the procs.
- Beta: Flame Shock's ticks crit on 2 of 19 (1.5 times). The sim lets them crit too (corrected in 5.7, this line first said it doesn't). Item 5.7.
- Beta, mobs below our level (the beta session's Thousand Needles run, 18:27 to 18:42, level 27 to 29 mobs, recorded in cc172ba0): direct spells crit on 18 of 149 (12.1%, with 12.3% on the sheet). Flametongue Weapon procs crit on 5 of 60, Searing Totem on 8 of 55, Fire Nova on 2 of 17 and the shocks on 3 of 17. Flame Shock ticks crit on 5 of 36. So below our level the procs crit at our full chance, like the sim. The drop at level 34 is still open: most hits there were Flametongue procs (1 crit in 63), and the other spells crit on 3 of 32.
- Beta, whole log at levels 27 to 35 (the beta session, 7ae7dd85, with Windfury Weapon's attacks taken out because they're melee): Searing Totem, the shocks and Fire Nova crit on 29 of 207 (14.0%), Flametongue Weapon on 10 of 133 (7.5%) and Flametongue Totem on 3 of 57 (5.3%). The procs are below the other spells at z = 2.3, and at level 34 the other spells crit on 4 of 46. So the level 34 drop was mostly the procs, at any level. Their rate is close to Healing Stream's 5.1% (5.7), so they may share a cause, like a flat 5% or our crit without Thundering Strikes or without base crit (7.3%). The beta session is asking the user for about 300 more Flametongue Weapon procs on mobs at or below level 30.
- Beta, the log up to 19:26 (the beta session's interim count): Flametongue Weapon now crits on 21 of 187 (11.2%), so its earlier 7.5% was chance and it gets our full chance. Searing Totem, the shocks and Fire Nova crit on 35 of 289 (12.1%), and Flame Shock's ticks on 9 of 89. Flametongue Totem is still 3 of 57 (5.3%, 1 time in 14 at 12.3%), so it stays open. Healing Stream (5.7) no longer has the procs to go with it.
- Beta, the log up to 19:32 (the beta session, 8b8e09fa): Flametongue Weapon crits on 24 of 198 (12.1%), so it's settled at our full chance and the sim is right. Searing Totem, the shocks and Fire Nova crit on 39 of 320 (12.2%), and Flame Shock's ticks on 15 of 122. Flametongue Totem is 5 of 82 (6.1%, 1 time in 19 at 12.3%), still open. The user now plays Windfury Weapon with Flametongue Totem, and about 150 more totem procs settle it. The totem procced on every white swing that landed in range, and on none of 11 Stormstrikes or 12 Windfury Weapon attacks that landed.
- Proposed to the user: no change yet. The beta session counts spell crits by spell and mob level in every new log. Flametongue Weapon on mobs of our own level would show whether the procs crit at our full chance, and Searing Totem on level 34 mobs whether +4 cuts it.
- The user OK'd waiting, and the two runs are in notes.md (Need to Verify) and in `beta_handoff.md` "Recheck with every new log".
- Settled (the beta session, the log up to 19:52, 9c188238): Flametongue Totem rose to 12 of 127 (9.4%). That comes 1 time in 5 at our 12.3% and 1 time in 38 at a flat 5%, so its first 82 procs were low by chance, like Flametongue Weapon's. Both procs crit at our full chance, as in the sim, so nothing changes. Telling 12.3% from 7.3% would take about 500 procs, for well under 1% of DPS. The beta session keeps counting it with every new log.
- Open again (the user, through the beta session, 392ead78): the user wants it confirmed that Flametongue Totem's crit chance moves with our spell crit. 12 of 127 fits our crit, but one crit chance can't show that. The plan is about 500 procs at a second crit chance, from the Elemental respec without Thundering Strikes (with Elemental Devastation and 5.6). It's in notes.md Need to Verify and `beta_handoff.md` item 12. The sim keeps rolling our spell crit for it.
  - Elemental respec (7.41% spell crit): 3 crits in 57 procs up to 21:25 (49b5cce8), 5.3%, with 4.2 expected at our crit. That fits our crit, and fits 5% too. With the fist weapon run (2ae6ffc0, 7.16% after 7 less Intellect) it's 5 of 83 (6.0%). Up to 22:16 (9294818f) it's 18 of 284 (6.3%) at 7.16 to 7.41%.
  - Both crit chances together (2026-10-10): 12 of 127 at 12.3% and 18 of 284 at about 7.3%. Our crit expects 15.6 and 20.7, and a flat 5% expects 6.4 and 14.2. 12 or more of 127 has a chance of 0.026 at 5% and 0.87 at our crit, and the two counts together are 6.7 times as likely with our crit as with 5%. So a flat 5% is unlikely. But a flat 7.3% (the two runs pooled, 30 of 411) fits both counts as well as our crit does, and the two rates (9.4% and 6.3%) are only 1.1 standard errors apart, where following our crit predicts a 5 point gap (the beta session's point, c9d3119c). So the data doesn't yet show the rate moving with our crit. The weak side is the 127 procs at 12.3%. About 400 more after the respec back to Enhancement would make the expected gap about 2.7 standard errors. Still open until the user calls it. The sim keeps rolling our spell crit, which no count so far rules out.
  - Up to 22:18 (the beta session, 57459eea): 13 more crits in 133 procs, so 31 of 417 (7.4%) since the respec at 7.2 to 7.4%. With the 12 of 127 at 12.3%, a flat 5% is now 48 times less likely than our crit. But a flat 7.9% (all 43 of 544) still fits as well as our crit (1.3 to 1). The Enhancement build at 12% is the next step.
  - Changed (the user, 2026-10-10, through the beta session, 79450996): under Forever, Flametongue Totem's procs crit a flat 8% of the time, whatever our crit (`totem_weapon_buffs.go` FlametongueTotemCritChance), until more beta data comes in. `flametongue_totem_test.go` TestOrcShamanFlametongueTotemCrits wants about 77 crits in 1000 procs at both 100% and 0% spell crit. The next run is about 400 procs at the Enhancement build's 12% spell crit, where our spell crit predicts about 49 crits and a flat 8% about 32. It stays in notes.md Need to Verify.
- Beta, whole log up to 20:18 (905e293a): Flametongue Totem crit on 18 of 182 (9.9%), still at one crit chance.
- Beta, whole log up to 20:25 (9435c80b): 22 of 206 (10.7%). The totem procced on 63 of 63 swings in range from 20:09 to 20:25.
- `crit_test.go` TestOrcShamanSpellCrit checks that Lightning Bolt, Earth Shock, Flame Shock and a Searing Totem attack crit at our full chance from 5 levels below us to 4 above, and against the level 63 boss at 60. With 20% more crit at our own level, each crit deals 1.5 times its hit.

### 5.4 Spell power

- The user OK'd it (2026-10-10). A spell hits for its base damage plus its coefficient times our spell power for its school, which is spell power ("damage and healing"), spell damage ("damage" only) and fire or nature power added together (`spell_result.go` GetSchoolDamage). Healing takes spell power and healing power. The percent modifiers then multiply the whole hit (5.6). Stormstrike and Windfury get no spell power. Their coefficient of 1 is for flat "+N damage" effects (4.10).
- Coefficients, the sim and the Forever client (70338, wago.tools SpellEffect) alike: Lightning Bolt 42.9%, 57.1% and then 71.4% from rank 3, Chain Lightning 57.1% (rank 3 51.7%), Earth Shock and Frost Shock 38.6%, Flame Shock 21.4% and 10% a tick, Lightning Shield orbs 26.7%, Searing Totem 1.7%, Magma Totem 3.3%, Fire Nova 21.4% (SoD's damage spells, settled below), Flametongue Weapon and Frostbrand 10% a hit, Flametongue Totem none, Healing Stream 2.2% a tick, Lava Burst 71.4%. Low ranks keep their full coefficient (5.5).
- Beta (our count from the Shimmering Flats log, 91 spell power, the same the beta session used for Searing Totem). The log's raw field drops the fraction, and the shown amount rounds it up at random by its size. Every hit fits the sim:
  - Lightning Shield rank 3: 40 + 26.7% of 91 = 64.30. All 94 hits were raw 64, and 27 of them (29%) were shown as 65.
  - Flame Shock rank 3: 45.4 + 19.47 = 64.87. All 5 were raw 64, and 4 were shown as 65. Its ticks are 14 + 9.1 = 23.1: raw 23 on all 15, shown 24 once.
  - Flame Shock rank 1: ticks 7 + 9.1 = 16.1 (raw 16 on all 4), and one crit of raw 43 (24 + 19.47 = 43.47).
  - Earth Shock rank 4: 83.69 to 89.31 + 35.13 = 118.8 to 124.4. The 15 hits were raw 118 to 123.
  - Searing Totem rank 3 (raw 20 to 26) and Flametongue Totem (raw 18 every time, 548 / 100 * 3.3 = 18.08 with no spell power) fit too.
  - Spell power from 90 to 91.5 fits all the raw numbers, and only 91 fits how often the orbs rounded up (90 would round up 3% of the time).
- The beta also shows no penalty for low ranks. Earth Shock rank 1 is learned at level 4 and Flame Shock rank 1 at 10, and both get their full coefficient. 1.12's penalty for spells learned below level 20 would take Earth Shock rank 1 to 15.4%, about 35 damage instead of 57. The TBC penalty for ranks far below our level would take it to 19.3%. Item 5.5.
- One rank 1 Earth Shock hit raw 57, where the sim's top is 21.64 + 35.13 = 56.77. It's one hit, 0.23 over.
- Flametongue Weapon rank 3 hit raw 33 every time (50 hits) on Rage of the Storm (3.3 speed) with 91 spell power. Without the per-rank cut (7.3) the sim gives 884 / 100 * 3.3 + 9.1 = 38.3, so the cut is 4.3 to 5.3, the same about 5 we saw on three weapons with 6 spell power. With a coefficient that grows with weapon speed (10% * speed / 4) the cut would be 2.7 to 3.7, and with vmangos's 3.85% a second of speed 6.7 to 7.7. So the 10% doesn't grow with speed, as in the sim. This answers the open question in 7.3.
- Settled: Fire Nova's damage spell. Forever's Fire Nova (408341 to 408345) does nothing itself in the client, so the server picks the damage spell. The client has two sets. One is the old totem's (8349, 8502, 8503, 11306, 11307), 10% and 14.3%, which the sim and the SoD sim use. The other is SoD's Fire Nova (408423 to 408428), 21.4%, with a few points less base damage (50, 102, 183, 284 and 403 against 52, 109, 196, 299 and 419). At level 30 rank 2 with 91 spell power the two are almost the same (122.5 to 137.5 against 122.4 to 136.5). At 60 rank 5 they're equal at 225 spell power, and SoD's is 5 more at 300. The combat log names the damage spell, so one cast on the beta answers it.
  - The beta tooltip for rank 2 (408342, 2026-10-10) says 123 to 138. The client's text for it reads $8502s1, so it shows the old totem's damage spell (all five ranks point at 8349, 8502, 8503, 11306 and 11307). At 91 spell power that's 122.5 to 137.5, and SoD's would be 122.4 to 136.5 (122 to 137). So the client agrees with the sim. Only a combat log shows what the server uses, so one cast still confirms it.
- Proposed to the user: no change. Fire Nova's damage spell goes on the beta list (notes.md).
- Changed (2026-10-10, the user's go, the beta session's 0a416e8e): the beta log names SoD's 408424 on all 17 rank 2 hits (16 casts in Thousand Needles, level 27 to 29 mobs, 91 spell power, raw 124 to 136). So the server doesn't use the spell the tooltip reads. The client gives 408424 102 base damage, 12.84% variance, 1.6 more a level from 22 to 27 and a 21.4% coefficient, so 122.41 to 136.53 at 91 spell power. The old 8502 gives 123.01 to 137.01, so the hits fit both and only the spell ID settles it. `fire_totems.go` now has SoD's damage at each rank's fifth level (51.23 to 59.77, 102.94 to 117.06, 182.12 to 205.88, 280.06 to 315.94 and 396.95 to 443.05), rank 1 grows 1.1 a level, and every rank has 21.4%. Only the enhancement golden moved, by 0.02% to 0.07%.
- `spell_power_test.go` TestOrcShamanSpellPowerBeta gives a level 30 Orc 91 spell power and rolls Lightning Shield rank 3, Earth Shock ranks 1 and 4, Flame Shock ranks 1 and 3, Searing Totem rank 3 and Fire Nova rank 2 2000 times each. It checks the sim's lowest and highest normal hit and Flame Shock's ticks, and that every raw number the beta showed can come out of them (the one rank 1 Earth Shock at 57 left out). It fails with a 25% orb coefficient and with the old totem's Fire Nova damage.

### 5.5 Spell ranks

- The user OK'd it (2026-10-10). Every rank of every shaman spell in the sim has the learn level of the Forever client (70338, wago.tools SpellLevels), 114 ranks in 23 spells. Each rank's growth per level and the level it stops growing at match the client too, up to level 60. The client lets some top ranks grow past 60 (Lightning Bolt rank 10 to 61, Lava Burst rank 3 to 68), which a level 60 never reaches.
- A rotation names the top rank, and a lower level character casts the highest rank it knows of the same spell (`spell_ranks.go` rank families). The Level 30 rotations cast Lightning Shield rank 3, Earth Shock rank 4, Flame Shock rank 3, Strength of Earth rank 2, Mana Spring rank 1, Flametongue Totem rank 1 and Fire Nova rank 2, the top rank of each at 30.
- 1.12's penalty for spells learned below level 20 is gone under Forever. The client gives each rank its own coefficient instead: Lightning Bolt rank 1 42.9% and rank 2 57.1%, every other rank the full one. The beta showed Earth Shock rank 1 and Flame Shock rank 1 with full spell power at level 30 (5.4).
- Forever has a penalty of its own. The dev notes of 1 October say spells cast with ranks "vastly below your current level" get less from spell power, and have a lower chance to trigger class abilities and talents (a rank 1 Frostbolt at 60 never procs Frostbite). The character sheet's Spell Damage tooltip explains it. The sim has no such penalty. On the beta, Earth Shock rank 1 and Flame Shock rank 1 got their full spell power at level 30, 21 and 15 levels past the level they stop growing at, so the penalty starts further down than that. It doesn't touch the sim, because our rotations always cast the top rank. It would matter if a rotation ever cast a low rank to save mana.
- Proposed to the user: no change. The sheet tooltip's exact text could tell us where the penalty starts.
- `spell_ranks_test.go` TestOrcShamanSpellRanks names the top rank of each of 17 spells and checks the rank a level 20, 30 and 60 shaman casts, from the client's learn levels. It also checks that a spell we haven't learned yet (Chain Lightning, Windfury Totem at 30) casts nothing.

### 5.6 How damage modifiers stack

- The user OK'd it (2026-10-10). A spell hit is its base damage plus spell power (5.4), times our own percent bonuses, times the target's, and then partial resists (5.2) and crit (5.3) (`spell_result.go` CalcDamage). Every bonus multiplies the whole hit, spell power part included. Each spell keeps two kinds of bonus. Bonuses of the same kind add up (DamageMultiplierAdditive), and the kinds multiply each other (DamageMultiplier and the rest).
- Our talent and item bonuses on a spell's damage. In the Forever client (70338, wago.tools SpellEffect) each one is the same kind, a percent modifier on the spell's damage (aura 108):
  - Concussion, 1% a point on Lightning Bolt, Chain Lightning and Earth Shock. The client's spell mask has only those three, as the tooltip says.
  - Call of Flame, 5% a point on Searing and Magma Totem, Fire Nova, Flame Shock (direct part and ticks) and Lava Burst.
  - Improved Fire Nova, 10% a point on Fire Nova.
  - Improved Lightning Shield, 5% a point on the orbs.
  - Elemental Weapons, 5% a point on the Flametongue and Frostbrand procs.
  - Rage of the Storm, 10% on Stormstrike (4.10).
- Only Fire Nova gets two of them. The sim adds them, so the Level 60 talents (Call of Flame 3, Improved Fire Nova 2) give 1 + 15% + 20% = 1.35. 1.12 adds percent modifiers of the same kind, and the SoD sim adds its talent bonuses too (Concussion, Call of Flame). If Forever multiplied them, it would be 1.15 * 1.2 = 1.38, 2.2% more Fire Nova damage at 60. Our Level 30 talents have none of these talents, so at Level 30 no two bonuses meet on one spell.
- The Stormstrike mark is another kind. The client puts it on the target as 20% more damage taken from us (aura 271) for Lightning Bolt, Chain Lightning and Earth Shock. So it multiplies with Concussion: 1.05 * 1.2 = 1.26 at 5 points. The sim multiplies them.
- Curse of the Elements under Forever: the client's 440892, 1311676, 1311677 and 1311680, learned at 20, 30, 40 and 50, give 4, 6, 8 and 10% more damage taken from every magic school (aura 87, all schools but Physical) and take 30, 45, 60 and 75 off every resistance. The sim has the same (`debuffs.go` foreverCurseOfElementsAura), by the raid's level. It multiplies with all of the above, so a Level 60 Earth Shock on a marked target deals 1.2 * 1.1 = 1.32 times its hit. Our Level 60 presets have it, and the Level 30 ones have no debuffs. Curse of Shadow and Improved Scorch are gone under Forever, and the sim skips them.
- Bonuses on all our spells of a school (Natural Alignment Crystal, Power Infusion, Sayge's Fortune) multiply with the rest, as in 1.12. None are in our presets.
- Lava Burst gets Call of Flame times its own 20% with Flame Shock on the target (1.15 * 1.2 = 1.38). The client's 20% is a script effect, so the server decides how it stacks. Lava Burst is a level 40 talent and isn't in our presets.
- Proposed to the user: no change. A beta respec settles add or multiply: Call of Flame 3 and Improved Fire Nova 2 need 15 points in Elemental, and Stormstrike fits in the other 15. With 91 spell power, Fire Nova rank 2 hits 165.4 to 185.6 if they add and 169.1 to 189.8 if they multiply, so a few casts on a group of mobs tell them apart.
- The user OK'd adding Fire Nova's add or multiply and Lava Burst's stacking to notes.md (Need to Verify, the Beta list and At 60).
- Beta (the beta session's Elemental respec, 67c6efc8, checked by us, 2026-10-10): with Call of Flame 3 and Improved Fire Nova 2 at 49 spell power (the 20:42 export), 25 Fire Nova hits (408424) had raw 160 to 175. Adding the two (x1.35, the sim) allows 153.1 to 172.2, and multiplying them (x1.15 x 1.2 = x1.38) allows 156.5 to 176.0. Raw 173, 173, 174 and 175 are over the add range, so Forever multiplies them. No spell power buff was up from 20:42:34 to 21:10:07 (the beta session's check: only Arcane Intellect, and this build has no Mental Quickness, no elixir, food or flask, and Flametongue and Grounding Totems), so the 49 held. Without the client's base, the average hit before the respec (x1, 91 spell power) and after it (49) gives x1.377 +- 0.013 (the beta session, 22f8e477). x1.35 fits none of the 25. Up to 21:25 (49b5cce8) it's 41 hits with 5 over the add range, every hit fits x1.373 to x1.392, and the average check gives x1.372 +- 0.012. Up to 22:16 (9294818f) it's 104 hits with 12 over, the same x1.373 to x1.392, and x1.377 on average on the client's base. The 23 hits before the respec (91 spell power, no talents) had raw 124 to 136, inside the sim's 122.4 to 136.5. Call of Flame raised Flame Shock's hit and ticks as the sim does (raw 64 and 21), and didn't touch Flametongue Totem (raw 18 on 18 of 18), as in the sim.
- Proposed to the user: Fire Nova's damage takes Call of Flame times Improved Fire Nova (`fire_totems.go`), 2.2% more Fire Nova damage at Level 60. The test then wants 1.15 * 1.2 * 1.1.
- Changed (2026-10-10, the user's go, "For now change nova to multiply them", while the beta session gathers more data): Fire Nova's damage is Call of Flame times (1 + Improved Fire Nova) (`fire_totems.go`). The Level 60 talents have both, and the Level 60 preset went from 757.1 to 759.3 DPS (+2.1, same seed, 20,000 iterations, mksettings with the Level 60 gear, talents and Optimized rotation at 180 sec). The goldens don't take Improved Fire Nova and didn't move. The test below now wants 1.15 * 1.2 * 1.1 for Fire Nova.
- `damage_modifiers_test.go` TestOrcShamanSpellDamageModifiers runs two level 60 sims with the same seed, one with no bonuses and one with Concussion 5, Call of Flame 3, Improved Fire Nova 2, Improved Lightning Shield 3, Curse of the Elements and the Stormstrike mark. Both roll the same hits and crits, so each spell's damage ratio is its multiplier: Earth Shock 1.05 * 1.2 * 1.1, Flame Shock (hit and ticks), Searing Totem and the orbs 1.15 * 1.1, and Fire Nova 1.15 * 1.2 * 1.1. It fails when Fire Nova's two talents add.

### 5.7 DoTs

- The user OK'd it (2026-10-10). Flame Shock is our only DoT (Lava Burst and the totems hit directly). In the sim its DoT goes up only when the direct hit lands, and its ticks never miss (`flame_shock.go`). It ticks 4 times, every 3 sec from the cast, so the last tick comes at 12 sec. Each tick deals the rank's points (14 for rank 3, the client's, with no growth by level) plus 10% of spell power.
- Beta, our count from the Shimmering Flats log (WoWCombatLog-101026_150733.txt). It has our Flame Shocks and those of another shaman in the group (Melba), 42 ticks in all. The first tick came 2.92 to 3.10 sec after the DoT went up (3.02 on average over 11), and the next ones 2.93 to 3.09 sec apart (3.01 over 27). The 7 DoTs that ran out ended on their 4th tick, 11.99 to 12.10 sec after they went up. A miss put no DoT up.
- A Flame Shock on a target that has ours already starts it over. The sim drops the old DoT, and the new one ticks 3 sec later, 4 more times (`dot.go` Apply). So the time since the last tick is lost. Melba's two recasts in the log did the same. One came 0.82 sec after a tick, and the next tick came 3.10 sec after the recast. If the old tick timer had kept going, it would have come after 2.18 sec. Then 4 ticks followed, and the old DoT's last tick never came. Our rotations only cast Flame Shock when its DoT is down, so this doesn't change our numbers.
- The sim takes our spell power and our own percent bonuses when the DoT goes up, and the target's (Curse of the Elements) on each tick (`spell_result.go` Snapshot). The log can't tell us what Forever does, because our spell power didn't change during a DoT. It doesn't matter for our rotations either, because nothing in our presets changes spell power or our percent bonuses during a fight.
- Ticks can crit under Forever. The Forever client gives every player DoT the flag that lets its ticks crit (Attributes_8 0x200, "periodic can crit"), and the Classic Era client (1.15.9) gives it to none. That's Flame Shock, and also Rupture, Garrote, Deadly Poison, Rend, Corruption and the others, and the Healing Stream heal. The beta sheet's crit tooltip says "Most periodic effects can critically strike". Since 2026-09-14 the sim rolls crit on every Forever DoT tick, at our crit chance at the time of the tick (`ruleset.go` canCrit). A spell tick crit deals 1.5 times, and Elemental Fury adds to it like any other crit.
  - Beta: 3 of the 42 ticks crit, 2 of our 19 (12.3% spell crit on the sheet, so 2.3 expected), all at 1.5 times (34 and 35 from 23, and 23 from Melba's 15).
  - Our 5.3 write-up said the sim never lets ticks crit. That was wrong, and it's corrected there. `spell_power_test.go` averaged the ticks with crits counted in, so it now leaves crit ticks out.
- Fixed (2026-10-10, the user's go): Forever's flag also lets the Healing Stream heal crit, and the sim's Healing Stream never crit. Now each heal can crit for 1.5 times under Forever, at our spell crit at the time of the heal (`water_totems.go`). Tidal Mastery and Elemental Fury don't list Healing Stream (class mask 0x2000) in the Forever client, so neither touches its crits, the same as before. The client makes the heal a periodic aura, and Water Shield's proc flags (0x262a8) leave periodic heals out. So the sim now deals it as a periodic heal, and its crits never spend a Water Shield globe. No preset puts down Healing Stream, so no number changed. It only matters for our health in PvP mode.
- Beta (the beta session's count, checked by us, 2026-10-10): Healing Stream does crit, but on only 23 of 449 heals from our own totems (5.1%) in the whole Shimmering Flats and Thousand Needles log, 9 of 194 on us and 14 of 255 on the totems. Each crit was 1.5 times (raw 12 or 13 against 8 or 9). The log shows them as periodic heals (6371). Our 12.3% spell crit is 2.3% base, 5.0% from 141 Intellect and 5% from Thundering Strikes, whose client aura (290) covers all crits. At 12.3% we'd expect 55 crits, give or take 7, so the heal doesn't get our full chance. Searing Totem's bolts in the same log crit at the full chance (20 of 132), so it isn't every spell from a totem. A flat 5% fits best (22.5 expected). Our chance without Thundering Strikes or without base crit, 7.3%, gives 33 give or take 5.5, which 23 makes unlikely but doesn't rule out. The sim gives the heal our full spell crit for now. For a while the Flametongue procs looked as low, but with more log Flametongue Weapon crits at our full chance (5.3). So the heal has its own rule, and a flat 5% fits it best (1 time in 25 at 7.3%, and ruled out at 12.3%). With the log up to 19:32 (8b8e09fa) it's 27 of 549 (4.9%): a count that low comes with a chance of 0.53 at 5% and 0.018 at 7.3%. Up to 19:52 (9c188238) it's 29 of 605 (4.8%), with 0.46 at 5% and 0.008 at 7.3%.
- Fixed (2026-10-10, the user's go): under Forever the heal now crits 5% of the time, whatever our spell crit (`water_totems.go` HealingStreamTotemCritChance). It stays at 5% until we know Forever's rule.
- `periodic_test.go` TestOrcShamanFlameShockTicks puts rank 3's DoT up at 1 sec with 100% spell crit and puts it up again at 8.5 sec with none. It checks ticks at 4 and 7 sec for 21 (crits), none at 10 and 13, and then ticks at 11.5, 14.5, 17.5 and 20.5 for 14. TestOrcShamanHealingStreamCrits heals with rank 2 1000 times at 100% spell crit and 1000 times at none, and expects about 50 crits each time (25 to 75), 9 on a crit and 6 otherwise. It also checks that Water Shield keeps its 3 globes.

### 5.8 Abilities that use the melee table

- The user OK'd it (2026-10-10). Stormstrike and Windfury Weapon's two attacks are our only melee abilities. Flametongue Weapon and Frostbrand procs are spells and roll the spell table (5.1, 5.3). Rockbiter Weapon only adds attack power. A shaman can't dual wield under Forever, so both always hit with the main hand and never take the dual wield miss penalty.
- Both are yellow (4.3). The first roll is miss, dodge, and in front of the target also parry and block. A second roll decides the crit, at twice a normal hit (4.7). They never glance. Our hit and weapon skill work as for white hits (4.4, 4.8).
- Both deal physical damage, so armor counts as for white hits (4.9). Flat "+N damage" effects add to them as to a white hit. Stormstrike's damage is the weapon roll plus attack power at the normalized speed plus 0.3 (4.1). Windfury's attacks use the weapon's own speed with the extra attack power, which Elemental Weapons raises once (4.12).
- Stormstrike costs its mana and starts its 8 sec cooldown whatever the outcome (`stormstrike.go`, the cast pays before the roll). Beta: a parried Stormstrike at 15:08:03 took 125 mana (2412 to 2287), and the next one came 8.02 sec later.
- The mark only goes up when Stormstrike lands, and a blocked one counts as landed (`stormstrike.go` result.Landed). Beta, our 93 Stormstrikes in the Shimmering Flats log: none of the 8 misses, 3 dodges and 19 parries put the mark up. The one blocked Stormstrike did, and so did all 4 crits and the 50 normal hits that left the target alive. The 8 that killed the target put none up, because it was dead. Another shaman's (Backtofront) 2 hits put it up, and their 1 dodge didn't. The client's Stormstrike (17364) is physical (school mask 1), with Normalized Weapon Damage and the mark as its two effects, and nothing for an off hand.
- Proposed to the user: no change.
- Improved Stormstrike's regen buff comes on every Stormstrike cast, landed or not (the user asked, 2026-10-10). The sim rolls it when the cast completes (`talents.go` applyImprovedStormstrike). Beta: all 93 of our casts in the log gave the buff (1238931, 50% regen while casting), the 8 misses, 3 dodges and 19 parries too, 1 to 21 ms before the hit or miss. So the character had both points (100%). The client's buff is 1238931, and the sim's buff carried the talent's ID (1223031). It has the buff's ID now (the user's go), which only changes the name and icon the sim shows.
- `stormstrike_test.go` TestOrcShamanStormstrikeOutcomes casts 400 Stormstrikes at level 60 from in front of a level 63 boss. Every one costs 125 mana and starts the 8 sec cooldown, and the mark is up after a hit, a crit or a block and not after a miss, dodge or parry. It fails when a miss leaves the mark.

### 6.1 When the enemy attacks us at all

- Shown to the user (2026-10-10), waiting for the OK. A target swings at a player only when the encounter names that player as its tank (`environment.go`, the target's tankIndex into the raid's tanks).
- Level 60: no tank, so the boss never swings at us, as when it faces the raid's tank and we stand behind it. Lightning Shield and Water Shield only take the raid damage hits (the class option raidDamageHitsPerMinute, item 6.5), and no preset sets them.
- Level 30: we solo Interrogator Vishas, so we're his tank. He's level 32 and swings every 2.0 sec for 52 to 69 before our armor, with parry haste on (6.2). Each hit can fire Lightning Shield. The preset doesn't put us in front of him, so our own swings never meet his parry or block (the user, 4.5).
- Level 30 PvP: every enemy is our tank too, from in front. The enemy types swing at 1.3 to 3.5 sec, and the caster has no swing at all, so Lightning Shield never fires there.
- Found: the PvP melee downtime moves only us (`pvp.go`). The enemy's own distance never changes, so its swings go on while we're out of melee range (`attack.go` checks the attacker's distance). On the default Level 30 PvP encounter we're out of range 70% of the fight, and the enemy hits us all that time and fires Lightning Shield. On the melee enemies it's 30%. Only the hunter's swing, which stands for its pet, should go on.
- Proposed to the user: stop the enemy's swings while we're out of melee range, except the hunter's. It lowers the PvP numbers, mostly through fewer Lightning Shield orbs.
- Set aside (the user, 2026-10-10). The PvP enemies only swing at us and cast nothing (no encounter AI, `target_ai.go`), so we don't simulate other specs' rotations against us. We leave their swings as they are for now.

### 6.2 Enemy damage

- Shown to the user (2026-10-10), waiting for the OK. An enemy swing deals its minimum damage times 1 to 1.33, plus the minimum times attack power / (14 * 177) (`attack.go` EnemyWeaponDamage). Vishas deals 52 to 69 every 2.0 sec, with no attack power. His numbers come from the 1.12 creature data, and we haven't seen him on Forever. Every PvP enemy uses the same 52 to 69 at its own swing speed. The Level 60 boss (3000, 805 attack power) never swings at us (6.1).
- An enemy crits for 2 times, 5% of the time plus 0.2% a level above us. It crushes for 1.5 times, 15% of the time at 3 or more levels above us and never below (`target.go`). Vishas crits 5.4% and never crushes.
- After a unit parries, its next swing comes sooner (`attack.go` applyParryHaste). Vishas has it, and so do we, because every character that can parry has it. We only parry with Spirit Weapons. Vishas never parries us, because the preset puts us behind him. The PvP enemies don't have it.
- None of this changes our DPS, except how often the enemy swings (each hit can fire Lightning Shield) and our own parry haste. The damage only shows in damage taken.
- Beta (our count from the whole Shimmering Flats and Thousand Needles log, 819 mob swings at us from level 24 to 35 mobs, 2026-10-10):
  - Swing speeds are round numbers: 2.00 sec for most mobs, 1.21 for the Needles Cougar and 3.01 for the Thundering Boulderkin.
  - The normal hits of each mob span 1.22 to 1.37 times their lowest, so the 1.33 spread fits. The log also shows each mob's attack power: 93 to 118 at levels 27 to 35.
  - Mobs almost never crit us: 3 crits in 819 swings (0.4%), where Classic's 5% gives about 41. None of our talents lowers it. So Forever seems to take most crits away from mobs.
  - Crushing blows come from mobs 3 or more levels above us only: 18 of 137 swings at +3 (13.1%) and 89 of 331 at +4 (26.9%). That fits Classic's rule of 15% at +3 and 10% more a level (25% at +4). The sim stops at 15%.
  - Parry haste works the Classic way, for mobs and for us. When more than 60% of the swing is left, the next swing comes 40% of the swing sooner. When 20% to 60% is left, it comes 20% of the swing after the parry. When less than 20% is left, nothing changes. For example, a 2.0 sec mob parried us 1.44 sec after its swing and swung again at 1.85. We parried 2.46 sec into our 3.3 sec swing and swung at 3.13. In 58 parries by mobs, every next swing fits this rule within 0.1 sec. In 25 parries by us, 20 fit within 0.1 sec and 3 within 0.16 (Flurry changes our speed). The other 2 swings came later than an unhasted swing would. Something held them, and with only instant casts around them it may not be a cast. The beta session's recheck with each mob's real swing speed (0193ac04, corrected in 9435c80b): 60 of 61 mob parries within 0.1 sec up to 20:09:26, and 72 of 73 up to 20:25 once a parry logged in the same ms as its swing counts in file order. Ours were 22 of 27 within 0.1 sec, 3 more within 0.16.
- Found: the sim's parry haste always moves the next swing 40% sooner, however little of the swing is left. It only stops the swing from coming earlier than 20% after the last swing, where the Classic rule counts that 20% from the parry. So a parry late in the swing makes the next swing come at once. Our 2.46 sec parry above would give a swing at 2.46 sec in the sim, not 3.13.
- Proposed to the user: fix parry haste to the rule above, leave the mob crit and crushing rules and only record them, and add Vishas' swing to notes.md Need to Verify.
- The user (2026-10-10): fix parry haste. Leave enemy damage to the player as it is and only record what we see, until we have a lot more data. Don't add single mobs to Need to Verify for now.
- Fixed (2026-10-10): parry haste now follows the rule above for every unit (`attack.go` applyParryHaste). On settings rebuilt from the preset files (their totals run a little below the UI's), it costs 0.57 DPS on Level 30 (Vishas, 189.71 to 189.14 at 20,000 iterations) and 0.19 to 0.43 in Level 30 PvP against the melee enemies (default 0.24, hunter 0.23, rogue 0.43, shaman 0.19, warrior 0.24). That's our own parries with Spirit Weapons, which no longer give an instant swing late in the swing. No golden changed, because the Level 60 boss never swings at us.
- Beta, the fist weapon run (21:32 to 22:16, the beta session, 9294818f): 39 of 39 single parries fit the rule within 0.1 sec. The log's only double parry (both Windfury attacks parried in the same ms, 21:53:22.8) didn't. The mob swung 2 ms after it, where the rule gives 0.4 sec or more even applied twice. It's one event, so we only record it.
- `parry_haste_test.go` TestOrcShamanParryHaste parries the enemy's 2.0 sec swing and our Whirlwind Axe swing at 20%, 50% and 85% of the swing. It expects the next swing at 60% and 70% of the swing for the first two and no change for the last. It failed on the old rule for every case but the first.

### 6.3 Our armor

- The user OK'd it (2026-10-10). Armor takes armor / (armor + 400 + 85 * the attacker's level) off physical damage, the Classic formula for attackers below level 60 (`spell_resistances.go` GetArmorDamageModifier). Our armor is the gear's armor (both numbers each piece lists, 1.7) plus 2 a point of Agility. Only physical damage uses it.
- Stoneskin Totem takes a flat 30 off each melee hit after armor, and after a crit doubles the hit. Guardian Totems adds 10% a point (rounded down). It's 30 for every rank, our own totem and another shaman's in the raid buffs alike (`buffs.go` StoneskinTotemAura). No preset puts it down.
- Devotion Aura from the raid buffs gives armor, but only to Alliance characters. We're Horde, so it never reaches us.
- None of this changes our DPS. It only changes damage taken.
- Beta, armor (our count from the whole log, the 14:50 export): the export shows 608 armor, split into 40 base and 568 bonus. The 558 normal mob hits on us after 16:00 (no crits, crushes or blocks, no Stoneskin) fit 568 armor. With 568 the formula predicts 21,484 damage in all against the 21,485 we took, where 608 predicts 21,242. The same armor solves at every mob level from 25 to 35 (553 to 584), so the attacker's level part of the formula is right too.
- The sim gives this gear 562 armor: 482 from the gear and 80 from 40 Agility. The sheet has 3 more Agility (43), which makes 568, the number the hits fit. So the sim's armor matches what the server uses. We don't know what the export's 40 base armor is, but it doesn't reduce damage. The 3 Agility and 5 Intellect the sim is missing come from enchants the sim doesn't have under these IDs (cloak 247, bracers 723 and the Forever enchants 8482 and 8485). That's for 9.6.
- Beta, Stoneskin Totem (our own rank 3, 15:23 to 15:49): the Forever client (70338) gives each rank its own flat reduction to physical damage taken: 4, 7, 11, 16, 22 and 30 (8072, 8156, 8157, 10403, 10404, 10405). Of our 77 normal hits taken while the totem was down, 62 fit 11 taken off before armor, 5 fit 11 after armor, and 10 fit no reduction. Those 10 all came between 15:29:26 and 15:29:56, probably out of the totem's range. With exact rounding, 42 of 67 fit 11 before armor and none 11 after, close to how often armor alone fits exactly (382 of 558).
- Found: the sim's Stoneskin is 30 for every rank, and it comes off after armor. On Forever it's the rank's value (11 at level 30), before armor. It's damage taken only.
- Devotion Aura in the Forever client keeps Classic's armor (55 to 735) under a new aura type (674, Classic's is 22). It doesn't reach us either way.
- Proposed to the user: keep the armor formula and our armor as they are (the beta fits them). Record the Stoneskin difference with the rest of the enemy damage to us, for when the user takes that up.
- The user OK'd it (2026-10-10): keep the armor formula and our armor, and record the Stoneskin difference only.
- `armor_test.go` TestOrcShamanArmorTaken gives a level 30 shaman 482 armor, the export gear's, tanking a level 34 enemy. It checks that our armor is that plus 2 a point of Agility, that 10 Agility adds 20, and that a 100 damage hit loses armor / (armor + 400 + 85 * 34).

### 6.4 Our miss, dodge, parry and block

- The user OK'd it (2026-10-10). An enemy swing at us rolls one table, in the order miss, dodge, parry, block, crit, crush and hit (`spell_outcome.go` outcomeEnemyMeleeWhite). Each of miss, dodge, parry and block is 0.2% lower a level the enemy is above us, and 0.2% higher a level below (`target.go`).
  - Miss: 5% at our level.
  - Dodge: our sheet dodge from base and Agility (1.3), plus Anticipation's 2% a point (`talents.go`, the Forever text says 2%, 4% and 6%).
  - Parry: 5%, only with Spirit Weapons.
  - Block: only with a shield, so never for us.
  - While we cast, we can't dodge, parry or block. Only a miss stops the swing then.
  - Every enemy swing comes from the front.
- What it changes for us: an avoided swing doesn't fire Lightning Shield, and a parry gives us parry haste (6.2). The rest is damage taken.
- Beta (our count from the whole log up to about 20:25, 1,094 mob swings at us, the 14:50 export with 5.43% dodge, 5% parry and Defense 150 of 150):

  | Mob levels | Swings | Miss | Dodge | Parry |
  |---|---|---|---|---|
  | Below us (-6 to -1) | 380 | 4.2% | 6.1% | 4.7% |
  | +1 and +2 | 192 | 7.3% | 3.6% | 3.1% |
  | +3 to +5 (mostly +4) | 506 | 4.2% | 5.7% | 2.8% |

  - The sim expects about 5.4%, 5.8% and 5.4% at -2, and 4.2%, 4.6% and 4.2% at +4. Miss and dodge fit.
  - Parry is low above our level: 20 of 698 (2.9%), where the sim expects about 30 (1 time in 30 by chance). Below our level it fits (18 of 380). Some of those swings may have come from our side or back, where we can't parry. The log doesn't show facing.
  - No swing was blocked, as in the sim.
  - Only 16 swings came while we cast (all Healing Wave and Lesser Healing Wave): 1 miss, no dodge or parry. That's too few to tell whether we can dodge while casting.
- Proposed to the user: no change. Record the low parry with the rest of the enemy damage to us. The user OK'd it (2026-10-10).
- `avoidance_test.go` TestOrcShamanOurAvoidance rolls a level 34 mob's swing at a level 30 shaman with Anticipation 3/3 and Spirit Weapons 20,000 times, and again while we cast. It checks that Anticipation adds 6% dodge, miss and parry are 4.2%, dodge is our sheet dodge less 0.8%, nothing is blocked, and only misses happen while we cast.

### 6.5 Spell damage to us and the raid damage hits

- Shown to the user (2026-10-10). No enemy casts at us, because no preset encounter has an AI (6.1). So Elemental Warding (3, 7 and 10% less Fire, Frost and Nature damage taken, as the Forever text says, `talents.go`) and our resistances never do anything in our sims.
- Raid damage hits (`shields.go`, Enhancement options "Raid hits per minute" and its +/-): the user sets how many direct hits a minute the raid's enemies land on us. Each one fires our Lightning Shield orb at our target or spends a Water Shield globe, the same as a real hit. Both shields still wait 3.5 sec between procs (the client's 3500 ms proc cooldown, 7.15 and 7.16). Nothing else that reacts to a hit taken sees them. With a +/-, each iteration picks its own rate in that range. No preset sets it.
- Found: the hits come at an even pace (one every 60 / rate sec). Real raid damage comes at random times, and with the 3.5 sec wait an even pace gives more procs than random hits at the same rate:

  | Hits a minute | Procs a minute, even | Procs a minute, random |
  |---|---|---|
  | 10 | 10 | 6.3 |
  | 15 | 15 | 8.0 |
  | 20 | 10 | 9.2 |
  | 30 | 15 | 10.9 |
  | 60 | 15 | 13.3 |

  Random hits average one proc every 3.5 sec plus the mean gap between hits. An even pace just above 3.5 sec procs on every hit, and one just below it procs on every second hit.
- Proposed to the user: make the hits come at random times (each gap drawn at random, with the same average rate). It only changes sims that set the option.
- OK from the user (2026-10-10, "make hits land at randomized times around a specific chosen setting for average hits per minute").
- Fixed: each wait for the next hit is drawn from an exponential spread with the set average (60 / rate sec), so the hits land at random times. The +/- still picks each iteration's rate. The "Raid hits per minute" tooltip now says the rate is an average and the hits land at random times.
- Tests:
  - TestOrcShamanRaidDamageHits (replaces TestOrcShamanRaidDamageLightningShield) keeps Lightning Shield at 3 orbs for 9 min at 15 hits a minute. It wants about 72 orbs (8 a minute), within 15. It got 78. The old even pace gives 135 and fails it.
  - TestOrcShamanWaterShield now lands enemy melee hits at set times (every 5 sec, then every 3 sec) instead of using the option, so it still checks each globe and the 3.5 sec wait.

### 6.6 Health

- Shown to the user (2026-10-10). Our health is base health plus 10 a point of Stamina (the first 20 give 1 each), settled and tested in 1.4. It only counts when an enemy attacks us (the Level 30 encounter and PvP, 6.1). Then each hit takes health, healing gives it back, and the sim counts the iterations where we die (Chance of Death in the results). Dying doesn't stop our damage, so health never changes DPS. The gear searches value it through the PvP presets' health weight.
- Toughness (`talents.go`): 2% more Stamina a point, 10% at 5. The Forever client (70291 trait curve 85301 on 16252, aura 137 on Stamina) gives 2, 4, 6, 8 and 10%. The Classic Era client has the old armor from items talent (aura 142). It multiplies our whole Stamina, base and gear, with Blessing of Kings, and drops the fraction (1.8). A naked level 30 Orc goes from 51 to 56 Stamina (56.1) and 665 to 715 health. Not measured on the beta.
- Improved Reincarnation (`talents.go`): 2% more max health a point. The client (curve 85310 on effect 1 of 16184, aura 133) gives 2 and 4%. The sim multiplies our whole health, Stamina's part included: 665 to 691.6 at 2 points. The game keeps whole health, so it would show 691. The other two parts, 10 and 20 min off Reincarnation's cooldown and 10 and 20% more health and mana on a Reincarnation, never come up in a fight. Not measured on the beta.
- Healing on us: Healing Stream (5.7, 5% crit under Forever) and the healing model in the settings (a set heal per second at random times). Only the Level 30 and PvP presets get hit, and neither puts down Healing Stream.
- Proposed to the user: no change. A beta check would be the sheet's health with Toughness or Improved Reincarnation, if a respec takes either.
- OK from the user (2026-10-10, "Ok on 6.6").
- `base_stats_test.go` TestOrcShamanHealthTalents checks a naked level 30 Orc: 51 Stamina and 665 health with no talents, 56 and 715 with Toughness 5/5, 665 * 1.04 with Improved Reincarnation 2/2 and 715 * 1.04 with both. It fails when Toughness keeps the fraction and when Improved Reincarnation leaves out Stamina's health.

### 6.7 PvP mode

- Shown to the user (2026-10-10). PvP mode (the encounter's PvP switch) changes two things in our attack table against the enemy player (`environment.go`). White hits never glance, because the enemy isn't a mob. When the enemy has any of Dodge, Parry or Block set, those replace the level based 5% each. The Level 30 PvP presets attack from in front, so parry and block count. Everything else is the same as against a mob of that level: the miss chances, crit, armor and spell hit (4% miss at the same level).
- The Forever client (70338) has no PvP damage rule for us. Every spell effect's PvpMultiplier is 1 (or 0 on 8 effects), so no shaman spell deals a different amount to a player.
- Time out of melee range (`pvp.go`, Encounter "melee downtime", PvP or not): we go out of range for random stretches of 1 to 10 sec, and the stretches in range are drawn so that the share out of range comes to the setting on average. A fight starts at a random point of that cycle. Out of range we have no white hits, no melee abilities (Stormstrike) and no Fire Nova (the enemy has left the totem). Shocks, Lightning Bolt and our totems' buffs keep going. Searing Totem would keep attacking too, but the PvP rotation doesn't use it.
- The enemy types (`presets.ts`, 2026-10-07): the Level 30 PvP preset fights a "Level 30 Player" (600 armor, 5% dodge, 5% parry), 70% out of range. The six stat weight presets each fight one kind of player built from median level 25 to 30 gear and the class's base dodge: cloth 320 armor and 4.5% dodge, no melee (70% out of range), hunter 770 / 7% / 5% parry (70%), rogue 790 / 14% / 5% (30%), shaman 700 / 6.5% / no parry (30%), warrior or paladin with a two hander 1320 / 4.5% / 5% (30%), and with a shield 2000 / 5.5% / 6% / 12% block (30%). These are our estimates, not beta numbers.
- Found, about the enemy hitting us (recorded only, the user's call on 6.2 to iron out enemy damage later): the enemy never leaves melee range. So it keeps swinging at us, and firing our Lightning Shield, while we are out of range. That's right when it stuns or roots us, but not when it kites us. Its crit is the mob rule (5% at the same level), not a player's crit from Agility. A player's resistances aren't modeled, so none of our spells is partly resisted.
- Proposed to the user: no change. After the OK, a test checks the PvP table (no glancing, the enemy's own dodge, parry and block) and that being out of range stops white hits, Stormstrike and Fire Nova but not the shocks, at the set share of the fight.
- OK from the user (2026-10-10, "ok").
- `pvp_test.go`:
  - TestOrcShamanPvPTable rolls 20,000 white hits from in front of a level 30 enemy player with Whirlwind Axe. With the enemy's own 14% dodge, 5% parry and 12% block it got 4.9% miss, 14.7% dodge, 5.0% parry, 12.4% block and no glancing. With none set it got 5.2%, 5.2% and 5.3%, the level based 5% each.
  - TestOrcShamanMeleeDowntime looks every 0.1 sec for an hour at half the fight out of range. We were out 49.7% of the time, and no white swing happened while out. Out of range, Stormstrike and Fire Nova couldn't be cast and Earth Shock could. Back in range all three could. The swing timer keeps running while we're out, so the held swing goes off as soon as we're back (662 swings in about 1800 sec in range with a 3.6 sec axe, about 2 for each time we come back).

### 7.1 Rank tables

- The user OK'd it (2026-10-10), and asked to rename the Windfury aura IDs as well ("so why not fix it as well?"). A background agent compared every shaman rank in the sim against the newest Forever client on wago.tools (1.60.1.70338, the SpellEffect, SpellLevels, SpellMisc, SpellPower, SpellCooldowns, SpellDuration, SpellCastTimes, SpellAuraOptions and SkillLineAbility tables), 1,025 field checks in all: spell ID, learn level, the level a rank stops growing at, damage or amount (low and high), growth by level, coefficient, mana, cast time, cooldown, duration and ticks. It checked its formula on Fire Nova rank 2 first (408424 gives the sim's 102.94 to 117.06 at 27). We checked the new mismatches in the client tables ourselves.
- Everything matches for Lightning Bolt (ranks 1 to 10), Chain Lightning (1 to 4), Earth Shock (1 to 7), Flame Shock (1 to 6, hit and DoT), Frost Shock (1 to 4), Lightning Shield (1 to 7, orbs included), Water Shield, Searing Totem (1 to 6), Magma Totem (1 to 4), Fire Nova (1 to 5), Flametongue Totem (1 to 4), Grace of Air (1 to 3), Strength of Earth (1 to 5), Mana Spring (1 to 4), Healing Stream (1 to 5), Windfury Weapon (1 to 4), Rockbiter (1 to 7), Stormstrike and Lava Burst (1 to 3). Every totem has a 1 sec GCD. No client rank is missing from the sim.
- Mismatches we already know about, with the user's decisions:
  - Stoneskin Totem: the sim gives every rank rank 6's -30. The client has -4, -7, -11, -16, -22 and -30 (6.3, recorded only).
  - Windfury Totem's proc wait: 100 ms in the client, 1.5 sec in the sim. The sim keeps 1.5 sec until it's checked at level 32 (4.12).
- New mismatches:
  - Mana Tide Totem lasts 13 sec in the client (16190, 17354, 17359, duration index 1073), 12 sec in the sim (`water_totems.go` and `buffs.go` ManaTideTotemDuration). It still ticks 4 times every 3 sec, so only the water slot stays taken 1 sec longer.
  - Frostbrand Weapon rank 1 (8034) is 32 at 20 plus 2.1 a level to 26, so 44.6. The sim has 45 and 13/6 a level (`frostbrand_weapon.go`). Ranks 2 to 5 match. No preset uses Frostbrand.
  - Flametongue Weapon (`flametongue_weapon.go`): the sim rounds each rank's damage (the client's points / 25) to one decimal. Rank 2 is 26.1 for 26.12, rank 3 42 for 42.08, rank 4 69.1 for 69.12 and rank 5 94.9 for 94.88. Ranks 1 and 6 are exact.
  - The Windfury Totem aura IDs the sim uses as names (8514, 10607, 10611) aren't spells in the client. The client's are 8515, 10609 and 10612 (the party auras) and 8516, 10608 and 10610 (the attack power buffs). The rotations look the aura up by 10611, so a rename would have to change them too. It's only a name.
- Small offsets, counted as matches:
  - Below a rank's top level, the sim takes the growth off both ends of the damage range. The client takes it off the base and then spreads it. The average is the same, but the sim's range is a little wider at the learn level (by up to 0.3 a side for Lightning Bolt, 0.9 for Fire Nova and 1.1 for Lava Burst). At the top level they agree.
  - Some ranks keep growing past 60 in the client (Lightning Bolt 10 to 61, Earth Shock 7 to 65, Flame Shock 6 to 67 and others). The sim stops them at 60, which is the same at a level 60 cap.
- Client oddities, for Need to Verify:
  - Flametongue Totem rank 2's party aura (8250) triggers a spell with 0 for its ID, where ranks 1, 3 and 4 name their proc spell. On the beta, rank 2 (levels 38 to 47) may proc nothing.
  - Lightning Overload: the client has its own overload spells, and some don't deal half of the main spell, as the sim does. Lightning Bolt ranks 1 to 5 give 0.52, 0.57, 0.52, 0.47 and 0.45 of the main hit, ranks 6 to 10 about 0.50, and Chain Lightning about 0.53. Chain Lightning rank 3's overload has a 0.2855 coefficient, half of 0.571, where the main spell has 0.517. So the 0.517 may be a client typo. No preset takes Lightning Overload.
- Not in the client tables, so not compared: Searing Totem's fire rate (2.5 sec from the beta logs), the Flametongue Weapon damage formula (a server script), Frostbrand's 9 PPM, which orb each Lightning Shield rank fires, and imbue mana (the sim puts imbues on before the fight).
- Proposed to the user:
  - Mana Tide Totem lasts 13 sec.
  - Frostbrand rank 1: 44.6 at 26 and 2.1 a level.
  - Flametongue Weapon: the exact points / 25 at every rank.
  - Leave the rest: the Windfury aura names, the range offset below the top level, Stoneskin and Windfury Totem's wait (decided earlier).
  - notes.md Need to Verify: Flametongue Totem rank 2 (At 40) and the overload damage (At 60).
- Changed (2026-10-10):
  - Mana Tide Totem lasts 13 sec (`buffs.go` ManaTideTotemDuration, which the shaman's own totem uses now too). The ticks keep their 3 sec period, where the party buff from another shaman used to take a quarter of the duration.
  - Frostbrand rank 1 is 44.6 at 26 and grows 2.1 a level (`frostbrand_weapon.go`).
  - Flametongue Weapon uses the exact N / 25 at every rank (`flametongue_weapon.go`). At 30, rank 3 deals 35.36 per 4 sec of weapon speed (was 35.28).
  - The Windfury Totem aura is 8515, 10609 and 10612, the client's party auras (`air_totems.go` WindfuryBuffAuraId). The five shaman rotations that asked for 10611 ask for 10612. The new and old IDs are one rank family, so a rotation saved before the rename still finds the aura, and a rotation that names rank 3 finds rank 1 or 2 at a lower level (it found nothing below 52 before). The UI database has the three new spell icons in place of the old ones, patched in without regenerating the items.
  - Both rank family items are in notes.md Need to Verify, At 60.
  - The goldens didn't move. We didn't rerun the presets.
- Tests:
  - `spell_ranks_test.go` TestOrcShamanImbueDamage checks a normal Flametongue Weapon hit at each rank's top level (and rank 3 at 30), and Frostbrand rank 1 at 20, 26 and 27, with no spell power. It fails with the old tables (23.49 for 23.51 at rank 2, 45 for 44.6).
  - `spell_ranks_test.go` TestOrcShamanWindfuryTotemAura has a rotation put Windfury Totem down when its aura isn't up, at 40, 50 and 60, asking for 10612 and for 10611. It casts once in 30 sec each time. Without the old IDs in the family, asking for 10611 cast it 11 to 14 times.
  - `mana_returns_test.go` TestOrcShamanManaTide now also wants the totem up until 15 sec after a drop at 2 sec, gone at 15.5 sec, and still 4 ticks.
  - `elemental_weapons_test.go` wants rank 3's 35.36.

### 7.2 Windfury Weapon

- The user OK'd it (2026-10-10). No change to the sim.
- The sim (`windfury_weapon.go`): ranks at 30, 40, 50 and 60, with 46, 119, 249 and 333 bonus attack power at the learn level, growing 7.2, 12.8 and 8.3 a level for 8 levels (rank 4 stays 333 at 60). These match the client (7.1).
- It procs on 20% of landed main hand melee hits: white swings (glancing, crits and blocks included), extra swings like Hand of Justice's, and Stormstrike. Misses, dodges, parries, spells and proc damage don't proc it (4.13). After a proc it can't proc again for 1.5 sec, so its own attacks never proc it.
- A proc makes two yellow main hand attacks right away, each weapon damage plus (attack power + bonus) / 14 times the weapon's own speed, not normalized. They roll the special attack table (no glancing, 4.7), don't restart the swing timer (4.11), don't use Flurry charges but give them on a crit, and proc Flametongue Weapon and the other on-hit effects (4.13). Elemental Weapons multiplies the bonus by 1.13, 1.27 or 1.4, once (8.24). The bonus only goes to the two attacks. On the same weapon it turns Windfury Totem off (7.19).
- Client 70338 (wago.tools): the four enchants (283, 284, 525, 1669) all use one proc aura, 439431 (SoD's spell), with 20%, a 1500 ms cooldown and white swings plus melee abilities (ProcTypeMask 0x14). Its 46 + 7.2 a level is rank 1's, so the server picks the rank. The attacks are 439440, a melee weapon hit (effect 58, melee defense) with no bonus of its own and "requires main hand weapon". Era's buff spells (8233, 8236, 10484, 16361: the bonus attack power, 2 extra attacks and 3 charges for white swings) are still in the client. But the beta logs never show them, so no white swing gets the bonus, as in the sim.
- Beta (our count from both logs up to 22:51, with Windfury Weapon on the weapon in hand, Rage of the Storm or Bloody Brass Knuckles):
  - 145 procs. 133 had two attacks, and in the other 12 the first attack killed the mob.
  - Each proc's first attack came within 150 ms of a landed white swing or Stormstrike. Leaving out killing blows (65), hits with no proc where the mob died within 0.12 sec from something else (9), and hits within 1.5 sec of a proc, white swings procced it on 132 of 689 (19.2%) and Stormstrike on 13 of 83 (15.7%). Together that's 145 of 772, 18.8% plus or minus 1.4%, which fits 20%.
  - The shortest time between two procs was 1.55 sec. A 100 ms cooldown (like Windfury Totem's in the client) would have allowed a proc on 6 hits 0.24 to 1.3 sec after a proc, and on 3 second attacks 100 ms or more after the first. None procced, where 20% expects 1.7, which happens 1 time in 7. So the logs fit 1.5 sec but don't rule out a short cooldown. The item stays in notes.md Need to Verify.
  - Damage: 9 Windfury hits averaged 1.140 times 16 white hits on a 3.6 speed axe at 198 attack power (section 7 part 2). The sim gives 1.144 with the weapon's own speed, and the normalized 3.3 would give 1.080.
  - In 6 of 38 procs both attacks landed about 100 ms after the hit (4.11). We only record it.
- Already tested: `elemental_weapons_test.go` (a normal hit's damage at rank 1 with 0 and 3 points), `attack_table_test.go` (yellow), `proc_matrix_test.go` (which hits proc it), `extra_attacks_test.go` (the swing timer), `damage_modifiers_test.go` (flat damage) and `talent_procs_test.go` (Flurry and the other talents).
- Tests: `windfury_weapon_test.go` TestOrcShamanWindfuryWeapon.
  - In a 2 hour fight at 60 with Stormstrike on cooldown, 457 of 2404 landed hits off cooldown procced (19.0%), within 4 standard errors (3.3%) of 20%. No proc came within 1.5 sec of the last, and every proc made two attacks. With a 100 ms cooldown the test fails 22 times, and with a 25% chance it sees 26.6%.
  - A normal hit with the weapon roll pinned and no armor deals the weapon's minimum plus (AP + 119, 249 or 333) / 14 times its speed, at 40, 50 and 60.

### 7.3 Flametongue Weapon

- The user OK'd the cut (2026-10-10): "put the cut but we need to remember to verify later ranks in the future". Ranks 4 to 6 are in notes.md Need to Verify.
- The sim (`flametongue_weapon.go`): ranks at 10, 18, 26, 36, 46 and 56. Every landed main hand melee hit (white swings, Stormstrike and Windfury Weapon's attacks, blocks included, 4.13) makes a fire hit for the rank's points / 100 times the weapon's own speed, plus 10% of spell power that doesn't grow with speed (5.4). Rank 3 is 884 points at 30 (8.84 per second of speed), and rank 6 is 2810 at 60. The points grow for 6 levels on ranks 1 and 2 and 8 on the others, as in the client (7.1).
- The hit rolls spell hit and spell crit (5.1, 5.3), and a crit deals 1.5 times (Elemental Fury raises it, section 8). Elemental Weapons multiplies the whole hit by 1.05, 1.1 or 1.15 (8.24). The hit doesn't trigger Elemental Devastation (section 8, beta untested in notes.md). On the same weapon it turns Flametongue Totem off (4.13). There's no level based partial resist under Forever (5.2).
- Client 70338 (wago.tools): the six enchants (5, 4, 3, 523, 1665, 1666) have one effect each, the proc aura 436519 (SoD's spell, 100%, white swings and melee abilities). The hit is 10444 "Flametongue Attack", a fire hit with no points of its own and a 0.1 coefficient, so the server fills in each rank's amount. The rank spells "Flametongue Weapon Proc" (8026, 8028, 8029, 10445, 16343, 16344) hold the points the sim uses. The enchants last 60 min.
- Beta, the whole log (WoWCombatLog-101026_150733.txt, 15:41 to 19:27, our count): 230 rank 3 hits on Rage of the Storm (3.3 speed) at level 30. The sheet had 91 spell power (49 from gear and 42 from Mental Quickness 2, 30% of 141 Intellect).
  - Every hit had raw 33, crits included. The amount was 33 on 120 normal hits and 34 on 54, and crits did 50 on 20 and 49 on 4. So the server rounds at random, and the hit is about 33.3 before rounding.
  - The sim gives 884 / 100 * 3.3 + 9.13 = 38.30, so rank 3 hits 5.0 below it. With only the 49 from gear it would give 34.07 and a cut of 0.8. But the 2026-10-08 hits at 6 spell power showed about 5 on the 3.6 speed axe, so the 10% goes on all our spell power and the cut is 5.0 on both weapons.
  - The cut on ranks 1 and 2 (0 and about 2) and the other weapons are in the part 2 pre-checks. The user kept the client's points for every rank on 2026-10-08, and notes.md Need to Verify has rank 6 at 60.
- Already tested: `spell_ranks_test.go` (a normal hit at each rank's top level and rank 3 at 30, no spell power), `elemental_weapons_test.go` (rank 3 with 0 and 3 points) and `proc_matrix_test.go` (which hits proc it, and it turns Flametongue Totem off). Nothing tests the spell power part or the crit.
- Changed (2026-10-10): under Forever, ranks 2 and 3 lose 2 and 5 a hit (`flametongue_weapon.go` FlametongueWeaponCut), before Elemental Weapons. Ranks 1 and 4 to 6 lose nothing until we see them at 36 or above. Classic keeps the full points. The goldens didn't move, and no preset uses Flametongue Weapon.
- Tests:
  - `flametongue_weapon_test.go` TestOrcShamanFlametongueWeapon checks every rank 3 hit at 30 with 91 spell power on Rage of the Storm (3.3) and Bloody Brass Knuckles (1.6): 8.84 a second of speed, minus 5, plus 9.1, and crits at 1.5 times. Rage of the Storm gives 33.27, where the beta averaged 33.3.
  - `spell_ranks_test.go` TestOrcShamanImbueDamage takes 2 and 5 off ranks 2 and 3, and `elemental_weapons_test.go` wants (35.36 / 4 * 3.5 - 5) * 1.15 with 3 points.

### 7.4 Rockbiter Weapon

- The user OK'd it (2026-10-11): "sounds right".
- The sim (`rockbiter_weapon.go`): ranks at 1, 8, 16, 24, 34, 44 and 54. It only adds attack power while the weapon is in hand. Each rank grows a few points a level from its learn level for 6 levels (ranks 1 to 3) or 8 (ranks 4 to 7): 49.5 for rank 1 at 6, 79 for rank 2 at 14, 118 for rank 3 at 22, 193.8 for rank 4 at 32, 355 for rank 5 at 42, 521.8 for rank 6 at 52 and 653 for rank 7 at 60 (it grows on to 686 at 62). Rank 4 is 177.6 at 30.
- Elemental Weapons multiplies the attack power by 1.07, 1.13 or 1.2 (8.24). It also adds flat threat to every main hand hit, which Spirit Weapons turns into 30% more threat instead of 30% less. Threat isn't in our DPS.
- Client 70338 (wago.tools): the enchants (29, 6, 1, 503, 1663, 683, 1664) each hold one passive (10400, 15567, 15568, 15569, 16311, 16312, 16313). Each passive is attack power of 29 + 4.1, 58 + 3.5, 88 + 5, 129 + 8.1, 211 + 18, 393 + 16.1 and 554 + 16.5 a level, growing to level 6, 14, 22, 32, 42, 52 and 62, plus a threat aura with 0 points that the server fills. That's the sim's attack power for every rank. Elemental Weapons (16266, client text) gives Rockbiter 7, 13 and 20%, the sim's numbers. The enchants last 60 min.
- Cross-check: the SoD sim has the same ranks and threat, but cuts the attack power by 90% (a SoD phase 3 change) and uses 14% for Elemental Weapons 2/3. Forever has neither.
- Beta (2026-10-08 and 2026-10-09, part 2 pre-checks): rank 4 at 30 gave 177 attack power on the sheet with no talent points (sim 177.6) and 213 with Elemental Weapons 3/3 (sim 213.1). So no SoD cut.
- Already tested: `elemental_weapons_test.go` (rank 4 at 30 with 0 and 3 points). Nothing tests the other ranks, the growth inside a rank, or Elemental Weapons 1/3 and 2/3.
- No sim change. Tests: `elemental_weapons_test.go` TestOrcShamanElementalWeapons/Rockbiter checks the attack power of each rank at the level it stops growing, rank 3 at 20 (108, still growing), and rank 4 at 30 with 0, 1, 2 and 3 points of Elemental Weapons (177.6 times 1, 1.07, 1.13 and 1.2). It fails with SoD's 14% for 2 points.

### 7.5 Frostbrand Weapon

- The user OK'd it (2026-10-10): "We might test at level 60 later, otherwise sounds good". The proc rate and rank 2's damage are in notes.md Need to Verify.
- The sim (`frostbrand_weapon.go`): ranks at 20, 28, 38, 48 and 58. It procs at 9 PPM on landed main hand melee hits (white swings, Stormstrike and Windfury Weapon's attacks, blocks included, 4.13), so a hit's chance is 9 * weapon speed / 60 (54% on a 3.6 speed weapon). A proc is a frost hit for a flat amount that doesn't depend on weapon speed: 44.6 for rank 1 at 26, 72 for rank 2 at 36, 117 for rank 3 at 46, 159 for rank 4 at 56 and 169.2 for rank 5 at 60. Rank 2 is 54 at 30. It adds 10% of spell power.
- The hit rolls spell hit and spell crit (5.1, 5.3), and a crit deals 1.5 times (Elemental Fury raises it, section 8). Elemental Weapons multiplies it by 1.05, 1.1 or 1.15 (8.24). Like Flametongue Weapon, it doesn't trigger Elemental Devastation or Clearcasting (its proc mask isn't a spell's). It puts an 8 sec debuff on the target that does nothing in the sim. There's no level based partial resist under Forever (5.2).
- Client 70338 (wago.tools): the enchants (2, 12, 524, 1667, 1668) are the old "combat spell" kind with no chance of their own, so the server sets the proc rate. Each casts its rank's Frostbrand Attack (8034, 8037, 10458, 16352, 16353): a 25% movement slow for 8 sec and a frost hit of 32 + 2.1, 48 + 3, 77 + 5, 127 + 4 and 158 + 5.6 a level, growing for 6 levels on rank 1 and 8 on the others, with a 0.1 coefficient. That's the sim's damage (7.1). The enchants last 60 min.
- Cross-check: the SoD sim uses 9 PPM and 0.1 too. Its damage table differs from the client and isn't used.
- Beta: we've never used Frostbrand. Another shaman ("Szm", 20:57 to 21:10 in WoWCombatLog-101026_150733.txt) used rank 2. Its log shows 5 procs (2 hits, 3 misses) on 10 landed melee hits, which fits 9 PPM on a slow weapon but is too few to say. Both hits had raw 63. That fits rank 2 at 30 with Elemental Weapons 3/3 and the 12 spell power in their logged stats, (54 + 1.2) * 1.15 = 63.5, with no Flametongue style cut (2 less would give 61). But we don't know their level or talents.
- The sim's Frostbrand Attack lacks the "passive" and "no cast complete" flags that Flametongue Weapon's hit has, so each proc counts as a cast. No shaman talent reacts to it, because they check for spell damage, so this changes nothing we know of.
- Already tested: `spell_ranks_test.go` (rank 1 at 20, 26 and 27, no spell power) and `proc_matrix_test.go` (which hits proc it). Nothing tests the proc rate, ranks 2 to 5, spell power or Elemental Weapons.
- Changed (2026-10-10): the hit has the same "passive" and "no cast complete" flags as Flametongue Weapon's. The goldens didn't move, and no preset uses Frostbrand.
- Tests: `frostbrand_weapon_test.go` TestOrcShamanFrostbrandWeapon.
  - In a 2 hour fight at 30 against a level 32 target, Rage of the Storm (3.3) procced on 937 of 1916 landed hits (48.9%, want 49.5%) and Bloody Brass Knuckles (1.6) on 989 of 3967 (24.9%, want 24%). Both must be within 4 standard errors. At 10 PPM the test fails on the Knuckles, and a flat chance would fail one of the two.
  - A normal hit with 100 spell power and Elemental Weapons 3/3 deals (flat damage + 10) * 1.15 for rank 2 at 30 (54) and 36 (72), rank 3 at 46 (117), rank 4 at 56 (159) and rank 5 at 60 (169.2).

### 7.6 Imbue rules

- The user OK'd it (2026-10-11): "ok", after asking whether the consumables can ever hold a different shaman imbue. They can't, so we dropped the guard (below). The stacking is in notes.md Need to Verify.
- The sim (`shaman.go` getImbueProcMask and ApplyShamanImbue, `consumes.go`): we pick one shaman imbue in the class settings, and it goes on the main hand. A shaman can't dual wield under Forever, so there's no off hand imbue. A shaman imbue in the consumables' main hand slot (old saved settings, and the TestEnhancement golden) works the same way, and when both name the same imbue it applies once, so Rockbiter's attack power isn't added twice. An oil or stone in the consumables' main hand slot stacks with the shaman imbue. Which totems the imbue turns off is 4.13 and 7.19.
- When the consumables and the class settings name two different shaman imbues, the sim applies both. Nothing we use can do that. The consumables' main hand picker lists only oils, stones, rogue poisons and Forever's imbue scrolls. The Enhancement, Elemental and Warden UIs move a shaman imbue saved in the consumables to the class settings when settings load. Restoration has no class settings imbue, rotopt's `mksettings` puts the imbue only in the consumables, and the Go tests set one or the other. Only a request written by hand with both can do it.
- The sim also stacks the imbue and an oil or stone under the Classic ruleset, where a 1.12 weapon has one temporary enchant. We only sim Forever.
- Client 70338 (wago.tools): the shaman imbues use a new enchant effect (360), as do the rogue poisons, the warlock's Firestone and Spellstone and Forever's "Imbue" scrolls. Oils and sharpening stones keep the old temporary enchant effect (54). So the client keeps them in two places, which fits the stacking.
- Beta (WoWCombatLog-101026_150733.txt): at 19:28:26 Windfury Weapon went on Rage of the Storm and Flametongue 3 came off at the same moment, so a weapon holds one shaman imbue. Windfury put on at 20:54:18 ran out at 21:54:21, so it lasts 60 min (the cast spells say 1800 sec, the enchants 3600). We've never had an oil or stone on beside an imbue.
- No sim change. Tests: `imbue_rules_test.go` TestOrcShamanImbueRules checks at 30 that Rockbiter in the class settings and Brilliant Wizard Oil in the consumables both apply (177.6 attack power, 36 spell power and 1% spell crit), and that Rockbiter in both places or only in the consumables adds 177.6 once. It fails without the guard in ApplyShamanImbue (355.2).

### Pre-checks for sections 3 and 4 (not yet shown to the user)

- Open: spells lose 2.1% crit against a level +3 target (`target.go` SpellCritSuppression). That isn't a 1.12 rule as far as we know. Item 5.3.

### Pre-checks for sections 5 and 6 (not yet shown to the user)

- 6.2 Parry haste on our own swing (the beta session's find, checked in the log, 2026-10-10). In WoWCombatLog-101026_135035.txt we parry a Sunscale Screecher at 13:50:39.0663, 1 ms after our swing at 39.0653. Our next swing comes at 40.8033, 1.738 sec later, where the two after it take 2.826 and 2.940 sec. A 2.88 sec swing (3.6 speed with 25% Flurry) cut by 40% of the swing gives 40.793, 10 ms off. That's the usual parry haste, and the sim gives it to any character that can parry (`attack.go` applyParryHaste). We still need to check whether the user had Spirit Weapons, and to look at the sim's rule for a parry late in the swing.
- 6.1 The Level 60 encounter has no tank, so the boss never swings at us. The Level 30 encounter makes us the tank (`tankIndex: 0`), so Vishas hits us and Lightning Shield fires. That preset doesn't set "in front of target" (only Level 30 PvP does), so our swings never meet Vishas's 6% parry or block, and the user keeps it that way (4.5).

### Pre-checks for section 7, part 1 (not yet shown to the user)

- 7.1 The 2026-10-07 class audit already checked all 458 shaman, rogue and paladin ranks against wowhead Forever and ForeverChanges 70245 (memory class-audit-2026-10-07). Item 7.1 only needs whatever changed since.
- 7.7 Stormstrike under Forever (`stormstrike.go`): 125 mana, 8 sec cooldown, 1.5 sec GCD, a main hand hit for weapon damage on the special attack table, normalized to the weapon type's speed plus 0.3 (4.1). When it lands it marks the target for 12 sec. Our next Lightning Bolt, Chain Lightning or Earth Shock on that target takes the mark and deals 20% more.
- 7.11 The shocks share one cooldown, 6 sec minus 0.2 sec per point of Reverberation (`shocks.go`), with a 1.5 sec GCD. Cost is cut by 2% per point of Convection and 45% with Shamanistic Focus. The two add up, they don't multiply.
- 7.9 Flame Shock: the direct part has a 0.214 spell power coefficient and grows with level up to its rank cap. The DoT is 4 ticks every 3 sec (12 sec) with 0.1 per tick, and it doesn't grow with level (`flame_shock.go`). Burning Totem adds 3 sec, one more tick.

### Pre-checks for section 7, part 2 (not yet shown to the user)

- 7.15 Lightning Shield (`lightning_shield.go`): 3 orbs, 10 min. Any direct hit an enemy lands on us (melee, ranged or spell) and each raid damage hit fire one orb, at most one every 3.5 sec (client 3500 ms, settled in 4.13). 26.7% spell power coefficient on every rank (Forever).
- 7.16 Water Shield (`water_shield.go`): free, 15 sec cooldown, 3 globes, 10 min. Each globe returns 2% of max mana, at most one every 3.5 sec (client 3500 ms, settled in 4.13). Only the raid damage hits per minute option feeds it at level 60.
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
  - Beta (2026-10-10, Shimmering Flats log): rank 3 on Rage of the Storm (3.3 speed) with 91 spell power hit raw 33 every time (50 hits), a cut of 4.3 to 5.3 against the sim's 38.3. That fits a flat 10% coefficient and rules out one that grows with weapon speed (5.4).
  - Flametongue Totem's procs are 489, 697, 947 and 1217 in Classic Era. Forever's (wowhead, and the sim) are 548, 781, 1061 and 1363.
- Beta, level 30 with 3/3 Elemental Weapons (2026-10-09), same Barbaric Battle Axe of Healing at 198 AP:
  - Rockbiter rank 4 gave 213 AP (sim 177.6 * 1.2 = 213.1).
  - The Windfury tooltip says 64 AP (46 * 1.4). On mixed level 8 to 10 boars and scorpids, 28 normal Windfury hits averaged 86.75 and 91 normal white hits 72.54, a ratio of 1.196 +- 0.012. Elemental Weapons applied once gives 1.201, twice 1.281 and not at all 1.144. Seven Windfury hits (79 to 83) were below the lowest a twice hit could do.
  - Fixed (2026-10-09): Elemental Weapons adds to Windfury's extra attack power once. The sim applied it twice, like the SoD sim. `elemental_weapons_test.go` TestOrcShamanElementalWeapons pins the weapon roll and checks every normal Windfury hit with 0 and 3 points, and Rockbiter's attack power. It fails with the double dip. TestEnhancement's results dropped about 1% (the default 1233.1 to 1218.7).
  - Flametongue Weapon with 6 spell power on the same axe: rank 1 hit 19 every time, where (15.84 + 0.6) * 1.15 is 18.9 (16 or 17 with no points). Rank 3 hit 31 or 32, where its 27.5 with no points times 1.15 is 31.6, and the sim without the cut times 1.15 would be 37.3. So the 15% is on the whole hit, like the sim, and the per-rank cut comes off before it. The tooltip still says 5 to 18 for rank 1, so it leaves the talent out.
  - TestOrcShamanElementalWeapons also checks every normal rank 3 Flametongue hit with 0 and 3 points. Frostbrand's 15% uses the same multiplier and isn't tested on the beta.
- 7.22 Searing Totem (`fire_totems.go`): a bolt every 2.5 sec for the totem's 30 to 55 sec lifetime, 1.7% coefficient (Forever). The real attack is every 2.2 sec, but the next bolt waits for the last to land, so we use 2.5 sec without a distance option. The beta shows the totem doesn't wait for its bolt (see the far totem run below), but the 2.5 sec fits anyway.
  - Beta (the beta session, 2026-10-10, `beta_results.txt` commit ad8ec9a5): 23 rank 3 bolts on level 33 and 34 mobs. Each bolt is a 2.22 sec cast (2.13 to 2.29), flies about 0.2 sec, and the totem waits 0 to 0.95 sec (0.26 on average) before the next cast. Bolt to bolt averaged 2.62 sec over 19 gaps, in two groups (2.35 to 2.48 and 2.79 to 2.88). The totem starts casting 3 ms after it goes down, so the first bolt leaves 2.23 to 2.25 sec after placement and lands about 2.45 sec after it. The sim's first bolt comes 2.5 sec after placement and then every 2.5 sec. Damage (raw 20 to 26 against the sim's 20.55 to 26.55 with 91 spell power), crit (4 of 20, 1.5x) and miss (3 of 23) fit the sim. Proposed to the user: 2.6 sec between bolts (withdrawn, see the gap breakdown below).
  - Beta lifetime (the beta session, 2026-10-10, `beta_results.txt` commit e9890230, checked against the log): a rank 3 totem put down at 15:46:49.458 started its last cast 39.58 sec later and never finished it, while we were still fighting. A cast takes 2.13 to 2.29 sec, so the totem was gone between 39.6 and 41.7 sec. That fits the sim's 40 sec. Its first bolt landed 2.26 sec after placement this time, after a 0.11 sec flight. The totem sat idle for 12.6 sec between our two mobs, so this run adds no bolt to bolt gaps.
  - Beta, a whole lifetime (the beta session, 2026-10-10, `beta_results.txt` commit 0c02ee3a): a totem that stayed up its whole life on one level 33 mob put its bolts 2.484 sec apart on average over 15 gaps. 13 gaps were 2.42 to 2.45 sec, and 2 were 2.82 to 2.84 sec, where the totem waited about 0.4 sec before casting. Over both runs it's 2.558 sec across 34 gaps, so the sim's 2.5 sec is inside the spread, and 2.6 may only fit the first mobs. That totem finished a cast at 40.538 sec, and two others started one at 39.58 and 39.67 sec and never finished it. The client gives rank 3 40000 ms, so it lasts 40 sec, give or take a server tick. 4 of 33 hits crit (12%) and 5 of 38 bolts missed. The user waits for more data before changing the time between bolts (2026-10-10).
  - Beta, the gap breakdown (the beta session, 2026-10-10, `beta_results.txt` commit babdb7bc): 42 bolt gaps from all five of our totems, leaving out gaps over 3 sec where the totem switched to a new mob. The cast takes 2.21 sec and the bolt flies 0.1 to 0.27 sec. Next to the mob the next cast starts on the same log line as the bolt's damage, so it looked like the totem waits for its bolt to land, as `fire_totems.go` says (the far totem run below shows it doesn't). A normal gap is 2.43 sec. In 9 of 42 gaps the next cast started about 0.4 sec late, 4 of 8 at 15:44:49 and 2 of 15 at 16:04, and the map positions show that neither the mob nor the user moved during any of them. The mean is 2.511 sec over all 42 gaps and 2.484 sec in the 16:04 fight, where the user stood next to the mob. The 16:04 totem fired 16 bolts in 40 sec, the sim's 40 / 2.5. The 2.6 sec came from the 15:44 fights, where half the gaps had the delay. So the sim's 2.5 sec stands and we withdraw the proposal.
  - Beta, the totem far from the mob (the beta session, 2026-10-10, `beta_results.txt` commit af17dc2f): the user put the totem 13, 16 and 19 yards from a level 33 mob that stayed put. The bolt's flight grew at about 19 yards a second, to 1.0 sec at 19 yards (the client's missile speed for 6351). But the next cast still started about 0.22 sec after the last one finished, the same as next to the mob. So the totem doesn't wait for its bolt to land, and distance only delays each bolt's damage. Over 72 gaps with the same mob standing still the mean is 2.527 sec, the normal gap 2.426 sec, and 18 gaps had the extra 0.4 sec. Under 6 yards it's 2.508 sec over 43 gaps, and at 12 yards or more 2.560 sec over 28. Both totems that lived 40 sec fired 16 bolts, as the sim's 2.5 sec gives (2.53 would give 15). The sim keeps 2.5 sec, and we need no distance option. The beta session proposes to the user fixing the comment in `fire_totems.go` that says the totem waits for the bolt.
- 7.23 Magma Totem: a pulse every 2 sec for 20 sec, 2 less damage per pulse than 1.12 (Forever).
- 7.24 Flametongue Totem: 5 min, fire damage per main hand hit by weapon speed (548 / 25 per 4 sec at rank 1, flat by level). Beta (the beta session, 2ae6ffc0): raw 8 on all 8 procs from Bloody Brass Knuckles (548 / 100 * 1.6 = 8.77), so the proc goes by weapon speed, as in the sim. It takes the fire slot, and placing another fire totem takes it down.
  - Beta, level 30 (2026-10-08): rank 1 on the 3.6 speed axe hit 19 or 20. The sim gives 548 / 100 * 3.6 = 19.73 plus 10% of spell power. Classic Era's 489 would give 17.6, so Forever's 548 holds.
  - With 55 spell power it still hit only 20, where 10% of spell power would make 25. Fixed (2026-10-08): spell power no longer adds to Flametongue Totem (cmangos doesn't add it either). `flametongue_totem_test.go` TestOrcShamanFlametongueTotem checks every normal hit is 548 / 100 * 3.5 with 55 spell power, and fails with the old 10%. TestEnhancement's presets lost 25 to 42 DPS.
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
- Elemental Weapons: Rockbiter +7 / 13 / 20%, Windfury +13 / 27 / 40% applied once (applied twice before 2026-10-09, see section 7 part 2), Flametongue and Frostbrand +5 / 10 / 15%.
- Shamanistic Focus: -45% cost on the shocks and Lightning Shield. Open: is Lightning Shield in Forever's text?
- Anticipation: +2% dodge. Toughness: +2% Stamina.
- Flurry: a melee crit (white, Stormstrike, Windfury attacks) gives 3 charges of 5 to 25% attack speed. White swings use the charges, at most one per 0.5 sec so that both hands don't spend two at once. Windfury Weapon's attacks are yellow, so they don't use charges. Windfury Totem's extra attack is white, so it does. The client's Flurry buff is used by auto attacks only, so extra swings use charges (4.13 part 2). The buff lasts 15 sec (fixed 2026-10-10).
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
