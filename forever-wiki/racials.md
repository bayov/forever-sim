# Racials

Every race gets four racials: two active, two passive. Blizzard's stated goal is similar overall value split differently across offence, defence and utility. Resistance racials are gone. Since client 1.60.1.70425 no racial depends on the weapon held: Human Ruthless and Orc Savage give crit with any weapon, and the Dwarf's mace racial became Earth Fury (crit damage).

The text below is the client's, checked line by line against build 1.60.1.70425 (wago.tools Spell, SpellEffect, SpellEquippedItems, SpellCooldowns, SpellAuraOptions, 2026-10-16; first read from build 1.60.1.69876 through hyjal.cc's racials page), with the cost line the client shows. Each change 70425 made from 70338 says so, with 70338's text. A racial that reads differently per class lists each version. Berserking's duration is missing from the client data ("?"), the sim keeps Classic's 10 sec. The earlier BlizzCon transcriptions (wowforevertalents.com, Method, Mobalytics) agreed with this text except for the numbers the client added: Eureka!'s cost cut, Shatter Curse's 15%, Stoneform's 10%, Cannibalize's 7% ticks, Plainsrunning's 1% per 5 sec to 30%, Touch of the Grave's 10% for casters.

Legend: NEW = new in Forever, CHANGED = changed from Classic, same = same as Classic, REMOVED = Classic racial not in Forever.

## New race and class combinations

Six at launch: Gnome Priest, Human Hunter, Dwarf Shaman, Orc Mage, Troll Warlock, Undead Paladin. More are planned. Undead Paladins get their own class quest line starting around level 10 at Banderian Keep in Tirisfal and a unique steed at 60.

## Alliance

### Human

Classes: Hunter (new), Mage, Paladin, Priest, Rogue, Warlock, Warrior.

- Will to Survive (active) NEW: [Instant; 3 min cooldown] Instantly removes all Stun effects.
- Perception (active) same: [Instant; 3 min cooldown] Dramatically increases stealth detection for 20 sec.
- Ruthless (passive) CHANGED: Increases your critical strike chance with all spells and attacks by 2%. (70425 renamed it from Sword Specialization and dropped the weapon rule; 70338: "... by 2% while you have a sword or two-handed sword equipped".)
- The Human Spirit (passive) same: Spirit increased by 5%.
- Diplomacy REMOVED (was: Reputation gains increased by 10%).
- Mace Specialization REMOVED (was: Skill with Maces and Two-Handed Maces increased by 5).

### Dwarf

Classes: Hunter, Paladin, Priest, Rogue, Shaman (new), Warrior.

- Stoneform (active) CHANGED: [Instant; 3 min cooldown] Instantly removes and grants immunity to all Bleed, Poison, and Disease effects, and reduces all Physical damage taken by 10% for 8 sec.
- Find Treasure (active) CHANGED: [Instant] Allows the dwarf to sense nearby treasure, making it appear on the minimap. Lasts until cancelled.
- Earth Fury (passive) NEW: Increases your critical strike bonus to damage and healing by 2%. (70425 replaced 70338's Mace Specialization, "Increases your critical strike chance with all spells and attacks by 1% while you have a mace or two-handed mace equipped".)
- Big Game Hunter (passive) NEW: Damage dealt versus Beasts increased by 5%.
- Frost Resistance REMOVED. Gun Specialization REMOVED.

### Night Elf

Classes: Druid, Hunter, Priest, Rogue, Warrior.

- Elune's Light (active) NEW: [Instant; 3 min cooldown] Increases your critical strike chance with all spells and attacks by 10% for 15 sec.
- Shadowmeld (active) CHANGED: [Instant; 10 sec cooldown] Activate to slip into the shadows, reducing the chance for enemies to detect your presence. Lasts until cancelled or upon moving. Using this ability in combat discourages enemies from attacking you, but increases the cooldown to 2 min.
  - The client row (20580, 1.60.1.70532): no global cooldown, a 10 sec category cooldown (2 min in combat by the tooltip), stealth (aura 16), a threat drop (aura 103, -999999), the cancel auto attack bit and no out-of-combat-only bit, so it works in combat. The sim has it since 2026-10-23, as wowsims/forever's shadowmeld.go does: 2 min cooldown, only in the fight (not before the pull: orchestration/standards.md item 13), off the global cooldown, it stops auto attack, and the next cast ends it. While it is up the Night Elf counts as stealthed, so Ravage (druid) and Ambush and Garrote (rogue) can follow it. The threat drop isn't modelled; it changes no damage. It is not a major cooldown, so only an APL line casts it. Shadowmeld into Ravage puts the Night Elf ahead of the Tauren for Feral (785.1 against 781.8 at 180 sec), and Shadowmeld into Ambush puts it ahead of the Orc for Subtlety (674.3 against 671.5). For the other dagger rogues it is worth +0.4 to +2.4, and the Skyborne Windshaper stays ahead (sim-status.md).
- Quickness (passive) CHANGED: Dodge chance increased by 1% and movement speed increased by 2%. Night Elf Rogues and Druids are harder to detect in Stealth as if they were 1 level higher.
- Wisp Spirit (passive) CHANGED: Transform into a wisp upon death, increasing movement speed by 75%.
- Nature Resistance REMOVED.

### Gnome

Classes: Mage, Priest (new), Rogue, Warlock, Warrior.

- Escape Artist (active) same: [Instant; 2 min cooldown] Instantly escape the effects of any movement impairing effect and gain immunity to those effects for 3 sec.
- Eureka! (active) NEW: [Instant; 2 min cooldown] lasts 8 sec. The 2026-09-24 build set the discount to 10% for every class (it was 20% Energy, 40% Rage, 50% Mana and 15% for Priests). Client 70425 dropped the 3 charges ("Your next 3 ...") for the 8 sec duration:
  - Rogue: Your non-periodic damaging abilities have their Energy cost reduced by 10% and deal 10% more damage.
  - Warrior: Your non-periodic damaging abilities have their Rage cost reduced by 10% and deal 10% more damage.
  - Warlock: Your damaging abilities have their Mana cost reduced by 10% and deal 10% more damage.
  - Mage: Your non-periodic damaging abilities have their Mana cost reduced by 10% and deal 10% more damage.
  - Priest: Your non-periodic damaging or healing abilities have their Mana cost reduced by 10% and deal 10% more damage or healing.
  - The sim multiplies the 10% off with the other percent cost cuts since 2026-10-23 (it added to them before): each percent cost row is its own factor in game (wowsims/forever fef2bcaebd). With Cataclysm 3/3 a Destruction spell costs 81% under Eureka!, not 80%.
  - Which spells count is set by each class's own Eureka! spell (Rogue 1259812, Warrior 1259813, Mage 1259817, Warlock 1259821, Priest 1259823): a class mask for the cost, one for direct damage and one for periodic damage. Only the Warlock's Drain Life, Drain Soul and Wrack and the Priest's Mind Flay, Penance and Starshards get it on their ticks ("channeled spells do not count as periodics", development notes 2026-10-01). Corruption, the banes, Siphon Life, Incinerate, Shadow Word: Pain, Devouring Plague, Rupture, Garrote and Rend get nothing; Immolate and Holy Fire get it on the hit only. The sim follows the masks (sim/core/racials_forever_eureka.go).
- Expansive Mind (passive) CHANGED:
  - Mage, Priest, Warlock: Maximum Mana increased by 5%.
  - Rogue: Maximum Energy increased by 5%.
  - Warrior: Maximum Rage increased by 5%.
- Engineering Specialization (passive) CHANGED: Your gnomish ingenuity reduces the rate of engineering devices failing or backfiring when you use them by 20%.
- Arcane Resistance REMOVED.

### High Order Skyborne (new, Alliance)

Classes: Druid, Hunter, Mage, Rogue, Warrior.

- Walk on Air (active) NEW: [Instant; 1.5 min cooldown, 2 min in 70338] Glide downward through the air for 10 sec while controlling your direction of travel.
- Read Leyline (active) NEW: [2 sec cast; 2 min cooldown] Attempt to tap into the power of a nearby leyline, increasing your Health and Mana regeneration by 100%. Lasts 30 sec if no leyline is nearby, and 15 min if one is found. (15 sec in 70338: client 70425 moved Energized 1270842 from DurationIndex 8 to 9, wowsims/forever e307ef4b0a.)
  - The sim (since 2026-10-23) casts it only from an APL and gives the 30 sec version. Energized is aura 110 (MOD_POWER_REGEN_PERCENT), Innervate's and Evocation's, so it doubles the Spirit part of the regen and not mp5; wowsims/forever doubles the whole regen. No preset casts it: it lost on the Level 60 Frost, Arcane, Fire and Balance builds at 180 sec, and Frost never gets low enough on mana to want it, even at 480 sec.
- Wind Blessed (passive) NEW: Increases your spellcasting, melee, and ranged Haste by 3%. (1% in 70338.)
- Elemental Insight (passive) NEW: Damage dealt versus Elementals increased by 5%.

## Horde

### Orc

Classes: Hunter, Mage (new), Rogue, Shaman, Warlock, Warrior.

- Blood Fury (active) CHANGED: [Instant; 3 min cooldown] Increases Attack Power and Spell Power by 10% for 15 sec. (2 min cooldown in 70338.)
- Shatter Curse (active) NEW: [Instant; 3 min cooldown] Instantly removes all Curses and Banes, and reduces all Magical damage taken by 5% for 8 sec. (70338: "Instantly removes and grants immunity to all Curses and Banes, and reduces all Magical damage taken by 15%".)
- Savage (passive) CHANGED: Increases your critical strike chance with all spells and abilities by 1%. Client 70425 renamed it from Axe Specialization and dropped the axe requirement (70338: "while you have an axe or a two-handed axe equipped").
- Hardiness (passive) CHANGED: Duration of Stun effects on you reduced by 20%.
- Command REMOVED (was: Hunter and Warlock pet damage +5%).

### Undead

Classes: Mage, Paladin (new), Priest, Rogue, Warlock, Warrior.

- Will of the Forsaken (active) CHANGED: [Instant; 2 min cooldown] Instantly removes all Charm, Fear and Sleep effects.
- Cannibalize (active) CHANGED: [5 yd range; Instant; 2 min cooldown] When activated, regenerates 7% of total Health and 7% of total Mana every 2 sec for 10 sec. Only works on Humanoid or Undead corpses within 5 yds. Any movement, action, or damage taken while Cannibalizing will cancel the effect.
- Underwater Breathing (passive) same: Underwater breath lasts 300% longer than normal.
- Touch of the Grave (passive) NEW: Your spells and attacks have a chance to drain Health from the target, up to 5% of your maximum Health. Client 70425: 20% (Warrior, Rogue, Paladin, spell 1260189) or 30% (Mage, Priest, Warlock, 1260201) a hit, at most once every 25 sec (SpellAuraOptions). It was 5% and 10% with no lockout before. Since client 70425 the drain no longer scales with Shadow damage bonuses on the target or the caster (development notes 2026-10-15: "Touch of the Dead no longer scales with Shadow increases on the target or the caster"); the sim gives it no damage modifiers at all.
  - Since the 2026-09-24 build it does not break crowd control, and abilities with no damage component (Distract, Pick Pocket, Polymorph, Expose Armor) no longer activate it. A damage over time spell activates it on the cast, not on its ticks.
- Shadow Resistance REMOVED.

### Tauren

Classes: Druid, Hunter, Shaman, Warrior.

- War Stomp (active) same: [0.5 sec cast; 1.5 min cooldown, 2 min in 70338] Stuns up to 5 enemies within 8 yds for 2 sec.
- Cultivation (active) CHANGED: [Instant] Cultivate a nearby herb, growing a duplicate you can harvest without requiring Herbalism skill. Each herb may only be cultivated once.
- Plainsrunning (passive) NEW: Gain 1% increased movement speed every 5 sec spent moving, up to a maximum of 30% increase. Taking damage or standing still will reduce this effect. In addition, you move 3% faster while mounted. (The mounted 3% is new in 70425.)
- Endurance (passive) CHANGED: Total Health increased by 5% and critical strike bonus to damage and healing increased by 2%. (70338: "Total Health increased by 5% and chance to hit increased by 1%".)
- Nature Resistance REMOVED.

### Troll

Classes: Hunter, Mage, Priest, Rogue, Shaman, Warlock (new), Warrior.

- Berserking (active) CHANGED: [Instant; 3 min cooldown] Increases your spellcasting and attack speed by 10% for 10 sec.
- Rapid Regeneration (active) NEW: [Channeled; 3 min cooldown] Regenerate 50% of your maximum Health over 6 sec. Any movement, action, or damage taken will cancel the effect.
- Beast Slaying (passive) same: Damage dealt versus Beasts increased by 5%.
- Regeneration (passive) CHANGED: Health regeneration rate increased by 20%. In addition, 20% of total Health regeneration will continue during combat. (10% and 10% in 70338.)
- Bow Specialization REMOVED. Throwing Specialization REMOVED.

### Windshaper Skyborne (new, Horde)

Classes: Druid, Hunter, Rogue, Shaman, Warrior.

- Walk on Air (active) NEW: [Instant; 1.5 min cooldown, 2 min in 70338] Glide downward through the air for 10 sec while controlling your direction of travel.
- Skysight (active) NEW: [0.5 sec cast; 2 min cooldown] Attempt to draw power from a convergence of elements and receive its blessing, increasing your movement and mounted movement speeds by 10%. Lasts 30 sec if no elemental convergence is nearby, and 15 min if one is found.
- Wind Blessed (passive) NEW: Increases your spellcasting, melee, and ranged Haste by 3%. (1% in 70338.)
- Elemental Insight (passive) NEW: Damage dealt versus Elementals increased by 5%.

## DPS-relevant summary (client 70425)

For a DPS sim comparing races within a faction:

- Horde: Orc gives 1% crit with any weapon (Savage) plus Blood Fury, 10% AP and SP for 15 sec every 3 min. Troll gives Berserking, 10% haste for 10 sec every 3 min, and 5% vs Beasts. Tauren gives 2% more crit damage (Endurance; no hit any more). Undead gives Touch of the Grave, a 20% (30% for casters) chance a hit to drain up to 5% of maximum health, once every 25 sec. Windshaper Skyborne gives 3% haste and 5% vs Elementals.
- Alliance: Human gives 2% crit with any weapon (Ruthless) and 5% Spirit. Dwarf gives 2% more crit damage (Earth Fury) and 5% vs Beasts. Night Elf gives Elune's Light, 10% crit for 15 sec every 3 min. Gnome gives Eureka!, 10% less cost and 10% more damage on its abilities for 8 sec every 2 min, and a 5% bigger Energy, Rage or Mana pool. High Order Skyborne gives 3% haste and 5% vs Elementals.

No racial depends on the weapon or on weapon skill any more. The sim's numbers per race are in [sim-status.md](sim-status.md).
