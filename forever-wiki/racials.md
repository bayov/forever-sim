# Racials

Every race gets four racials: two active, two passive. Blizzard's stated goal is similar overall value split differently across offence, defence and utility. Resistance racials are gone. Since client 1.60.1.70425 no racial depends on the weapon held: Human Ruthless gives crit with any weapon, and the Dwarf's mace racial became Earth Fury (crit damage), which the 2026-10-26 pass replaced with Artisan's Precision (hit).

**The 2026-10-26 racial pass.** Blizzard's development notes of 2026-10-26 (post 7, "Racial Abilities and Traits") retune several racials server side; client 1.60.1.70630 still has the old rows, so the text below gives the notes' version, marked "2026-10-26", with the client's in brackets. The sim follows the notes (sim/core/racials_forever.go), and the UI's racials.json carries them through tools/sources/racials_hotfix.json: Human Ruthless 1% crit (was 2%) and The Human Spirit 10% Spirit (was 5%); Dwarf Earth Fury removed for Artisan's Precision, +1% hit with spells and attacks; Tauren Endurance only 5% health, Plainsrunning 2% a stack to 60% and no mounted bonus; Orc Savage removed and Command back, +5% Hunter and Warlock pet damage; Skyborne Wind Blessed 2% haste (was 3%), Read Leyline 50% regen for 15 sec (100% while Attuned), Skysight 15% for 15 sec (30% while Charged).

Otherwise the text below is the client's, checked line by line against build 1.60.1.70425 (unchanged in 70630 but for Eureka!'s "for 8 sec") (wago.tools Spell, SpellEffect, SpellEquippedItems, SpellCooldowns, SpellAuraOptions, 2026-10-16; first read from build 1.60.1.69876 through hyjal.cc's racials page), with the cost line the client shows. Each change 70425 made from 70338 says so, with 70338's text. A racial that reads differently per class lists each version. Berserking's duration is missing from the client data ("?"); the sim uses the 15 sec of the 2026-10-27 notes. The earlier BlizzCon transcriptions (wowforevertalents.com, Method, Mobalytics) agreed with this text except for the numbers the client added: Eureka!'s cost cut, Shatter Curse's 15%, Stoneform's 10%, Cannibalize's 7% ticks, Plainsrunning's 1% per 5 sec to 30%, Touch of the Grave's 10% for casters.

Legend: NEW = new in Forever, CHANGED = changed from Classic, same = same as Classic, REMOVED = Classic racial not in Forever.

## New race and class combinations

Six at launch: Gnome Priest, Human Hunter, Dwarf Shaman, Orc Mage, Troll Warlock, Undead Paladin. More are planned. Undead Paladins get their own class quest line starting around level 10 at Banderian Keep in Tirisfal and a unique steed at 60.

## Base attributes

Forever redesigned the races' attributes (user ruling 2026-10-25). At level 1 every race is the class baseline plus a race offset that adds up to 0 over Strength, Agility, Stamina and Intellect and leaves Spirit alone. The sim (sim/core/base_stats.go RaceOffsets at 60, tools/gen_level_base_stats.py below 60) takes, relative to the Human:

| Race | Str | Agi | Sta | Int | Spi | Classic 1.12 (Str, Agi, Sta, Int, Spi) |
| --- | --- | --- | --- | --- | --- | --- |
| Human | 0 | 0 | 0 | 0 | 0 | 0, 0, 0, 0, 0 |
| Dwarf | +2 | -2 | +1 | -1 | 0 | +2, -4, +3, -1, -1 |
| Night Elf | -2 | +2 | 0 | 0 | 0 | -3, +5, -1, 0, 0 |
| Gnome | -3 | +1 | -1 | +3 | 0 | -5, +3, -1, +3, 0 |
| High Order Skyborne | -1 | +1 | -1 | +1 | 0 | (new race) |
| Orc | +3 | -3 | +1 | -1 | 0 | +3, -3, +2, -3, +3 |
| Undead | +2 | -1 | +1 | -2 | 0 | -1, -2, +1, -2, +5 |
| Tauren | +2 | -2 | +2 | -2 | 0 | +5, -5, +2, -5, +2 |
| Troll | +1 | +2 | 0 | -3 | 0 | +1, +2, +1, -4, +1 |
| Windshaper Skyborne | -1 | +1 | -1 | +1 | 0 | (new race) |

Sources: our naked level 1 beta sheets (2026-10-10, Horde, several classes each, the same shift on every class), wowsims/forever 6296e8546d (#597, 2026-10-20: naked level 1 warrior sheets of every race against a Human, this exact table with Spirit 0 everywhere), and their Skyborne sheets (2026-09-22: four Human/High Order class pairs and a Windshaper warrior, the same delta). The Human keeps its Spirit racial (10% since 2026-10-26) on top. We assume the offsets hold at every level, 60 included. Open: wowsims/forever 6fc22306b1 (#619, 2026-10-21) put the seven original races back on Classic's offsets, saying Forever restored them, with no sheet or source; a naked level 1 sheet taken after 2026-10-21 settles it (notes.md).

## Alliance

### Human

Classes: Hunter (new), Mage, Paladin, Priest, Rogue, Warlock, Warrior.

- Will to Survive (active) NEW: [Instant; 3 min cooldown] Instantly removes all Stun effects.
- Perception (active) same: [Instant; 3 min cooldown] Dramatically increases stealth detection for 20 sec.
- Ruthless (passive) CHANGED: Increases your critical strike chance with all spells and attacks by 1%. (2026-10-26: "Ruthless now increases critical chance by 1% (was 2%)"; client 70630 still reads 2%. 70425 renamed it from Sword Specialization and dropped the weapon rule; 70338: "... by 2% while you have a sword or two-handed sword equipped".)
- The Human Spirit (passive) CHANGED: Spirit increased by 10%. (2026-10-26: "Human Spirit now increases Spirit by 10% (was 5%)"; client 70630 reads 5%, Classic's.)
- Diplomacy REMOVED (was: Reputation gains increased by 10%).
- Mace Specialization REMOVED (was: Skill with Maces and Two-Handed Maces increased by 5).

### Dwarf

Classes: Hunter, Paladin, Priest, Rogue, Shaman (new), Warrior.

- Stoneform (active) CHANGED: [Instant; 3 min cooldown] Instantly removes and grants immunity to all Bleed, Poison, and Disease effects, and reduces all Physical damage taken by 10% for 8 sec.
- Find Treasure (active) CHANGED: [Instant] Allows the dwarf to sense nearby treasure, making it appear on the minimap. Lasts until cancelled.
- Artisan's Precision (passive) NEW: Increases your hit chance with spells and attacks by 1%. (2026-10-26: "Earth Fury (2% increased critical strike damage and healing) removed. Added: Artisan's Precision". No client spell yet; client 70630 still has Earth Fury 1259719, "Increases your critical strike bonus to damage and healing by 2%", which 70425 put in place of 70338's Mace Specialization, 1% crit with a mace or two-handed mace.) The sim adds 1% melee (and ranged) hit and 1% spell hit.
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
  - Rogue: Your non-periodic damaging abilities have their Energy cost reduced by 10% and deal 10% more damage for 8 sec.
  - Warrior: Your non-periodic damaging abilities have their Rage cost reduced by 10% and deal 10% more damage for 8 sec.
  - Warlock: Your non-periodic damaging abilities have their Mana cost reduced by 10% and deal 10% more damage for 8 sec.
  - Mage: Your damaging abilities have their Mana cost reduced by 10% and deal 10% more damage for 8 sec.
  - Priest: Your non-periodic damaging or healing abilities have their Mana cost reduced by 10% and deal 10% more damage or healing for 8 sec.
  - (Client 70630 added "for 8 sec" to each.) The development notes' "Changed Eureka on every class to a 10% discount on Mana, Rage, or Energy abilities" is the 2026-09-24 change above; it reappears in the 2026-10-26 edit of that post, but the later posts and client 70630 keep the 10% damage and the class masks, so the sim keeps both halves.
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
- Read Leyline (active) NEW: [2 sec cast; 2 min cooldown] Increases your Health and Mana regeneration by 50% for 15 sec. Used at a Ley Line it also grants Attuned for 1 hour (it persists through death), which raises the bonus to 100%. It cannot be used while Innervate is active. (2026-10-26: "now increases Mana and Health regeneration by 50% (was 100%), with a 15-second duration". Client 70630 still reads "increasing your Health and Mana regeneration by 100%. Lasts 30 sec if no leyline is nearby, and 15 min if one is found"; 15 sec in 70338, client 70425 moved Energized 1270842 from DurationIndex 8 to 9, wowsims/forever e307ef4b0a.)
  - The sim (since 2026-10-23) casts it only from an APL, and since 2026-10-26 gives 50% for 15 sec; its fights have no Ley Line, so never the Attuned 100%. Energized is aura 110 (MOD_POWER_REGEN_PERCENT), Innervate's and Evocation's, so it raises the Spirit part of the regen and not mp5; wowsims/forever raises the whole regen. No preset casts it: it lost on the Level 60 Frost, Arcane, Fire and Balance builds at 180 sec, and Frost never gets low enough on mana to want it, even at 480 sec.
- Wind Blessed (passive) NEW: Increases your spellcasting, melee, and ranged Haste by 2%. (2026-10-26: "Wind-Blessed Haste bonus reduced to 2% (was 3%)"; clients 70425 to 70630 read 3%, 70338 1%.)
- Elemental Insight (passive) NEW: Damage dealt versus Elementals increased by 5%.

## Horde

### Orc

Classes: Hunter, Mage (new), Rogue, Shaman, Warlock, Warrior.

- Blood Fury (active) CHANGED: [Instant; 2 min cooldown] Increases Attack Power and Spell Power by 10% for 15 sec. The client (70425 to 70634) reads a 3 min cooldown; the development notes of 2026-10-27 (post 8) put it back to 2 min, as in 70338, and the sim follows them.
- Shatter Curse (active) NEW: [Instant; 3 min cooldown] Instantly removes all Curses and Banes, and reduces all Magical damage taken by 5% for 8 sec. (70338: "Instantly removes and grants immunity to all Curses and Banes, and reduces all Magical damage taken by 15%".)
- Command (passive) same: Damage dealt by Hunter and Warlock pets increased by 5%. (2026-10-26: "Savage removed (1% increased critical strike chance). Command re-added, increasing Warlock and Hunter pet damage by 5%". Client 70630 still has Savage 20574, "Increases your critical strike chance with all spells and abilities by 1%", which 70425 made of Axe Specialization by dropping the axe requirement, and no Command row; the UI uses Classic's 21563.) The sim multiplies the damage of a hunter's or warlock's pet by 1.05, guardians (hawks, Infernal) not included, as Classic's Command.
- Hardiness (passive) CHANGED: Duration of Stun effects on you reduced by 20%.
- Savage REMOVED (2026-10-26).

### Undead

Classes: Mage, Paladin (new), Priest, Rogue, Warlock, Warrior.

- Will of the Forsaken (active) CHANGED: [Instant; 2 min cooldown] Instantly removes all Charm, Fear and Sleep effects.
- Cannibalize (active) CHANGED: [5 yd range; Instant; 2 min cooldown] When activated, regenerates 7% of total Health and 7% of total Mana every 2 sec for 10 sec. Only works on Humanoid or Undead corpses within 5 yds. Any movement, action, or damage taken while Cannibalizing will cancel the effect.
- Underwater Breathing (passive) same: Underwater breath lasts 300% longer than normal.
- Touch of the Grave (passive) NEW: Your spells and attacks have a chance to drain Health from the target, up to 5% of your maximum Health. Client 70425: 20% (Warrior, Rogue, Paladin, spell 1260189) or 30% (Mage, Priest, Warlock, 1260201) a hit, at most once every 25 sec in the client (SpellAuraOptions, still in 70634), 15 sec since the development notes of 2026-10-27 (post 8), which the sim follows. It was 5% and 10% with no lockout before. Since client 70425 the drain no longer scales with Shadow damage bonuses on the target or the caster (development notes 2026-10-15: "Touch of the Dead no longer scales with Shadow increases on the target or the caster"); the sim gives it no damage modifiers at all.
  - Since the 2026-09-24 build it does not break crowd control, and abilities with no damage component (Distract, Pick Pocket, Polymorph, Expose Armor) no longer activate it. A damage over time spell activates it on the cast, not on its ticks (Shadow Word: Pain, the development notes' example). The sim rolls on the cast of a spell with a damage multiplier and on each landed attack that comes without a cast; periodic ticks never roll.
  - Client 70630 takes four attributes off the drain (1260198): Ignore Caster Modifiers (Attributes_3 0x20000000) and three healing ones (Attributes_9 0x1000000 Ignore Caster Healing Modifiers, Attributes_10 0x2, Attributes_13 0x4000000). It keeps Ignore Damage Taken Modifiers (Attributes_4 0x100) and Ignore Caster Damage Modifiers (Attributes_6 0x20000000), and the drain (effect 9, 5 base points) has no spell power coefficient, so its damage still takes no modifier on either side, as the sim has it.
- Shadow Resistance REMOVED.

### Tauren

Classes: Druid, Hunter, Shaman, Warrior.

- War Stomp (active) same: [0.5 sec cast; 1.5 min cooldown, 2 min in 70338] Stuns up to 5 enemies within 8 yds for 2 sec.
- Cultivation (active) CHANGED: [Instant] Cultivate a nearby herb, growing a duplicate you can harvest without requiring Herbalism skill. Each herb may only be cultivated once.
- Plainsrunning (passive) NEW: Gain 2% increased movement speed every 5 sec spent moving, up to a maximum of 60% increase. Taking damage or standing still will reduce this effect. (2026-10-26: 2% a stack to 60%, was 1% to 30%, and no longer faster while mounted. Client 70630 reads 1% to 30% and "In addition, you move 3% faster while mounted", new in 70425.)
- Endurance (passive) CHANGED: Total Health increased by 5%. (2026-10-26: "Endurance no longer provides any modifiers beyond 5% increased Health (no hit chance or critical damage modifiers)". Client 70630 still reads "and critical strike bonus to damage and healing increased by 2%", 70425's; 70338: "and chance to hit increased by 1%".)
- Nature Resistance REMOVED.

### Troll

Classes: Hunter, Mage, Priest, Rogue, Shaman, Warlock (new), Warrior.

- Berserking (active) CHANGED: [Instant; 3 min cooldown] Increases your spellcasting and attack speed by 10% for 15 sec. 15 sec since the development notes of 2026-10-27 (post 8); the client's duration is missing ("?") and the sim had Classic's 10 sec before.
- Rapid Regeneration (active) NEW: [Channeled; 3 min cooldown] Regenerate 50% of your maximum Health over 6 sec. Any movement, action, or damage taken will cancel the effect.
- Beast Slaying (passive) same: Damage dealt versus Beasts increased by 5%.
- Regeneration (passive) CHANGED: Health regeneration rate increased by 20%. In addition, 20% of total Health regeneration will continue during combat. (10% and 10% in 70338.)
- Bow Specialization REMOVED. Throwing Specialization REMOVED.

### Windshaper Skyborne (new, Horde)

Classes: Druid, Hunter, Rogue, Shaman, Warrior.

- Walk on Air (active) NEW: [Instant; 1.5 min cooldown, 2 min in 70338] Glide downward through the air for 10 sec while controlling your direction of travel.
- Skysight (active) NEW: [0.5 sec cast; 2 min cooldown] Increases your movement speed by 15% for 15 sec. Used at an elemental convergence it also grants Charged for 1 hour (it persists through death), which raises the bonus to 30%. (2026-10-26: "Skysight is now a 15% increased movement speed (was 10%), with a 15-second duration". Client 70630 reads "increasing your movement and mounted movement speeds by 10%. Lasts 30 sec if no elemental convergence is nearby, and 15 min if one is found".) Movement only; the sim leaves it out.
- Wind Blessed (passive) NEW: Increases your spellcasting, melee, and ranged Haste by 2%. (2026-10-26; clients 70425 to 70630 read 3%, 70338 1%.)
- Elemental Insight (passive) NEW: Damage dealt versus Elementals increased by 5%.

## DPS-relevant summary (development notes 2026-10-26 over client 70630)

For a DPS sim comparing races within a faction:

- Horde: Orc gives Blood Fury, 10% AP and SP for 15 sec every 2 min, and 5% more pet damage for Hunters and Warlocks (Command); no crit racial any more. Troll gives Berserking, 10% haste for 15 sec every 3 min, and 5% vs Beasts. Tauren gives only 5% health (Endurance; no hit or crit damage any more). Undead gives Touch of the Grave, a 20% (30% for casters) chance a hit to drain up to 5% of maximum health, once every 15 sec. Windshaper Skyborne gives 2% haste and 5% vs Elementals.
- Alliance: Human gives 1% crit with any weapon (Ruthless) and 10% Spirit. Dwarf gives 1% hit (Artisan's Precision) and 5% vs Beasts. Night Elf gives Elune's Light, 10% crit for 15 sec every 3 min. Gnome gives Eureka!, 10% less cost and 10% more damage on its abilities for 8 sec every 2 min, and a 5% bigger Energy, Rage or Mana pool. High Order Skyborne gives 2% haste and 5% vs Elementals.

No racial depends on the weapon or on weapon skill any more. The sim's numbers per race are in [sim-status.md](sim-status.md).
