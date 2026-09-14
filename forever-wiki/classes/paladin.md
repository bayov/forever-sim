# Paladin

Source: wowforevertalents.com dataset (BlizzCon 2026 footage overlaid on Classic Era 1.15.9 client tables), data version 2026-09-13/14. See [../README.md](../README.md) for caveats. Numbers on new talents were read off level 38 demo characters, so absolute damage and mana values are not level 60 values.

Talents: 52 total, 20 new, 21 changed, 11 unchanged (verified).

Row N needs 5*(N-1) points spent in that tree. Row 1 is the top row. Rows 3, 4, 5 and 7 hold the 11, 16, 21 and 31 point talents.

Warning on multi-rank NEW and CHANGED talents: the footage showed rank 1 only. The dataset scaled the other ranks linearly from rank 1, which sometimes produces nonsense (for example Maelstrom Weapon rank 5 reading 'stacks up to 25 times'). Treat rank 1 as observed and the rest as a guess until the beta client is out. The [ranks 2+ unverified] tag marks every multi-rank new or changed talent. Read the Note line, it says when a higher rank was actually seen on a BlizzCon slide.

## Spellbook changes (trainer abilities)

33 of 51 trainer abilities confirmed from footage. Abilities not listed below are either verified unchanged or still Classic placeholders (unverified).

### New abilities

- **Holy Strike** (retribution, 20 Mana, Melee Range, Instant, 12 sec cooldown)
  - An instant strike that causes 40% weapon damage plus an additional 36 to 46 as Holy damage.
  - Note: Baseline Retribution strike with ranks: the level 38 demo paladin knew Rank 5 (tooltip not hovered) and the BlizzCon slide showed Rank 8. The Improved Holy Strike and Iron Creed talents build on it.
- **Blessing of Kings** (protection, 75 Mana, 30 yd range, Instant)
  - Places a Blessing on the friendly target, increasing total stats by 10% for 1 hour. Players may only have one Blessing on them per Paladin at any one time.
  - Note: A talent in Classic, in Forever it is a trainer blessing in the Protection tab and lasts 1 hour.
- **Seal of Fury** (protection, 200 Mana, Instant)
  - Fills the Paladin with divine fury for 30 sec, causing melee attacks to deal an additional 37 Holy damage. While a shield is equipped, each attack also grants an absorb shield equal to 50% of the Holy damage dealt. Only one Seal can be active on the Paladin at any one time. Unleashing this Seal's energy causes 163 to 178 Holy damage to an enemy and taunts the target to attack you for 4 sec.
  - Note: The new tanking seal: Holy damage on hit, an absorb shield while a shield is equipped, and a taunting Judgement. Rank 4 on the level 38 demo paladin, Rank 7 on the BlizzCon slide with the same wording and roughly doubled numbers.

### Changed abilities

- **Blessing of Might** (retribution, Level 32, 60 Mana, 30 yd range, Instant)
  - Places a Blessing on the friendly target, increasing melee attack power by 61 for 1 hour. Players may only have one Blessing on them per Paladin at any one time.
  - Classic: Places a Blessing on the friendly target, increasing melee attack power by 185 for 5 min. Players may only have one Blessing on them per Paladin at any one time.
  - Note: Blessings now last 1 hour. Rank 4 read 61 attack power on the level 38 paladin where Classic rank 4 gives 85.
- **Judgement** (retribution, Level 4, 6% of base mana, 10 yd range, Instant, 10 sec cooldown)
  - Unleash the energy of a Seal spell upon an enemy. Does not consume the Seal. Refer to individual Seals for Judgement effect.
  - Classic: Unleashes the energy of a Seal spell upon an enemy. Refer to individual Seals for Judgement effect.
  - Note: Judgement no longer consumes the Seal. Still costs 6% of base mana as in Classic: 56 Mana on the level 38 demo paladin, 91 Mana on the level 60 BlizzCon slide.
- **Seal of the Crusader** (retribution, Level 32, 90 Mana, Instant)
  - Fills the Paladin with the spirit of a crusader for 30 sec, granting 157 melee attack power. The Paladin also attacks 40% faster, but deals less damage with each attack. Only one Seal can be active on the Paladin at any one time.

Unleashing this Seal's energy will judge an enemy for 40 sec, increasing Holy damage taken by up to 92. Your melee strikes will refresh the spell's duration. Only one Judgement per Paladin can be active at any one time.
  - Classic: Fills the Paladin with the spirit of a crusader for 30 sec, granting 325 melee attack power. The Paladin also attacks 40% faster, but deals less damage with each attack. Only one Seal can be active on the Paladin at any one time.

Unleashing this Seal's energy will judge an enemy for 10 sec, increasing Holy damage taken by up to 140. Your melee strikes will refresh the spell's duration. Only one Judgement per Paladin can be active at any one time.
  - Note: Judgement effects now last 40 sec instead of 10. Rank 4 judgement increases Holy damage taken by up to 92 (Classic 80), the 157 attack power matches Classic at level 38.
- **Retribution Aura** (retribution, Level 36, Instant)
  - Causes 18 Holy damage to any creature that strikes a party member within 30 yards. Players may only have one Aura on them per Paladin at any one time.
  - Classic: Causes 20 Holy damage to any creature that strikes a party member within 30 yards. Players may only have one Aura on them per Paladin at any one time.
  - Note: Rank 3 read 18 Holy damage per hit on the level 38 paladin where Classic rank 3 does 12.
- **Divine Protection** (protection, Level 18, 35 Mana, Instant, 5 min cooldown)
  - You are protected from all physical attacks and spells for 8 sec, but during that time you cannot attack or use physical abilities yourself. Applies Forbearance for 1 min. Cannot be cast while Forbearance is active.
  - Classic: You are protected from all physical attacks and spells for 8 sec, but during that time you cannot attack or use physical abilities yourself. Once protected, the target cannot be made invulnerable by Divine Shield, Divine Protection or Blessing of Protection again for 1 min.
  - Note: Uses the Forbearance debuff instead of the old shared immunity lockout text.
- **Blessing of Protection** (protection, Level 38, 65 Mana, 30 yd range, Instant, 5 min cooldown)
  - A targeted party member is protected from all physical attacks for 10 sec, but during that time they cannot attack or use physical abilities. Players may only have one Blessing on them per Paladin at any one time. Applies Forbearance for 1 min. Cannot be cast while Forbearance is active.
  - Classic: A targeted party member is protected from all physical attacks for 10 sec, but during that time they cannot attack or use physical abilities. Players may only have one Blessing on them per Paladin at any one time. Once protected, the target cannot be made invulnerable by Divine Shield, Divine Protection or Blessing of Protection again for 1 min.
  - Note: Uses the Forbearance debuff instead of the old shared immunity lockout text.
- **Righteous Fury** (protection, Level 16, 280 Mana, Instant)
  - Increases the threat generated by your Holy attacks by 90%. Lasts 30 min.
  - Classic: Increases the threat generated by your Holy attacks by 60%. Lasts 30 min.
  - Note: Threat bonus raised from 60% to 90%.
- **Blessing of Salvation** (protection, Level 26, 75 Mana, 30 yd range, Instant)
  - Places a Blessing on the party member, reducing the amount of threat generated by 30% for 1 hour. Players may only have one Blessing on them per Paladin at any one time.
  - Classic: Places a Blessing on the party member, reducing the amount of all threat generated by 30% for 5 min. Players may only have one Blessing on them per Paladin at any one time.
  - Note: Lasts 1 hour instead of 5 min.
- **Shadow Resistance Aura** (protection, Level 28, Instant)
  - Gives 30 additional Shadow resistance to all party and raid members within 30 yards. Players may only have one Aura on them per Paladin at any one time.
  - Classic: Gives 60 additional Shadow resistance to all party members within 30 yards. Players may only have one Aura on them per Paladin at any one time.
  - Note: Now covers raid members as well as the party.
- **Divine Shield** (protection, Level 34, 75 Mana, Instant, 5 min cooldown)
  - Protects the paladin from all damage and spells for 10 sec, but reduces all damage you deal by 50%. Applies Forbearance for 1 min. Cannot be cast while Forbearance is active.
  - Classic: Protects the paladin from all damage and spells for 12 sec, but increases the time between your attacks by 100%. Once protected, the target cannot be made invulnerable by Divine Shield, Divine Protection or Blessing of Protection again for 1 min.
  - Note: Now reduces all damage you deal by 50% instead of slowing your attacks, and applies Forbearance.
- **Fire Resistance Aura** (protection, Level 36, Instant)
  - Gives 30 additional Fire resistance to all party and raid members within 30 yards. Players may only have one Aura on them per Paladin at any one time.
  - Classic: Gives 60 additional Fire resistance to all party members within 30 yards. Players may only have one Aura on them per Paladin at any one time.
  - Note: Now covers raid members as well as the party.
- **Holy Light** (holy, Level 38, 365 Mana, 40 yd range, 2.5 sec cast)
  - Heals a friendly target for 610 to 682.
  - Classic: Heals a friendly target for 1590 to 1770.
  - Note: Rank 6 healed 610 to 682 on the level 38 paladin where the Classic client gives 698 to 780 at that level.
- **Lay on Hands** (holy, Level 30, 40 yd range, Instant, 20 min cooldown)
  - Heals a friendly target for an amount equal to the Paladin's maximum health and restores 250 of their mana. Drains all of the Paladin's remaining Mana when used, but does not interrupt Mana regeneration.
  - Classic: Heals a friendly target for an amount equal to the Paladin's maximum health and restores 550 of their mana. Drains all of the Paladin's remaining mana when used.
  - Note: Cooldown cut from 1 hour to 20 min and it no longer interrupts mana regeneration.
- **Blessing of Wisdom** (holy, Level 34, 65 Mana, 30 yd range, Instant)
  - Places a Blessing on the friendly target, restoring 24 mana every 5 seconds for 1 hour. Players may only have one Blessing on them per Paladin at any one time.
  - Classic: Places a Blessing on the friendly target, restoring 33 mana every 5 seconds for 5 min. Players may only have one Blessing on them per Paladin at any one time.
  - Note: Lasts 1 hour, rank 3 restores 24 mana per 5 sec where Classic rank 3 restores 20.
- **Exorcism** (holy, Level 36, 180 Mana, 30 yd range, Instant, 15 sec cooldown)
  - Causes 194 to 218 Holy damage to an Undead or Demon target.
  - Classic: Causes 505 to 563 Holy damage to an Undead or Demon target.
  - Note: Rank 3 read 194 to 218 damage on the level 38 paladin where the Classic client gives 221 to 249 at that level.
- **Flash of Light** (holy, Level 34, 70 Mana, 40 yd range, 1.5 sec cast)
  - Heals a friendly target for 101 to 114.
  - Classic: Heals a friendly target for 348 to 388.
  - Note: Rank 3 healed 101 to 114 on the level 38 paladin where the Classic client gives 151 to 169 at that level.
- **Consecration** (holy, 565 Mana, Instant, 8 sec cooldown)
  - Consecrates the land beneath the Paladin, doing 96 Holy damage over 8 sec to enemies who enter the area. The first 4 enemies who enter the area will take an additional 233 damage over 8 sec.
  - Classic: Consecrates the land beneath Paladin, doing 384 Holy damage over 8 sec to enemies who enter the area.
  - Note: Redesigned and baseline (Rank 2 known at level 38 without the Classic talent): a smaller ground effect plus extra damage to the first four enemies who enter it. Rank 5 on the BlizzCon slide keeps Classic's 565 Mana but deals 96 to everyone plus 233 to the first four, where Classic Rank 5 dealt 384 to everyone. The level 38 tooltip was partly obscured, the slide confirms its wording.
- **Seal of Light** (holy, Level 30, 110 Mana, Instant)
  - Fills the Paladin with divine light for 30 sec, giving each melee attack a chance to heal the Paladin for 39. Only one Seal can be active on the Paladin at any one time.

Unleashing this Seal's energy will judge an enemy for 40 sec, granting melee attacks made against the judged enemy a chance of healing the attacker for 25. Your melee strikes will refresh the spell's duration. Only one Judgement per Paladin can be active at any one time.
  - Classic: Fills the Paladin with divine light for 30 sec, giving each melee attack a chance to heal the Paladin for 94. Only one Seal can be active on the Paladin at any one time.

Unleashing this Seal's energy will judge an enemy for 10 sec, granting melee attacks made against the judged enemy a chance of healing the attacker for 61. Your melee strikes will refresh the spell's duration. Only one Judgement per Paladin can be active at any one time.
  - Note: Judgement effects now last 40 sec instead of 10.
- **Seal of Wisdom** (holy, Level 38, 135 Mana, Instant)
  - Fills the Paladin with divine wisdom for 30 sec, giving each melee attack a chance to restore 50 of the Paladin's mana. Only one Seal can be active on the Paladin at any one time.

Unleashing this Seal's energy will judge an enemy for 40 sec, granting attacks and spells used against the judged enemy a chance to restore 33 mana to the attacker. Your melee strikes will refresh the spell's duration. Only one Judgement per Paladin can be active at any one time.
  - Classic: Fills the Paladin with divine wisdom for 30 sec, giving each melee attack a chance to restore 90 of the Paladin's mana. Only one Seal can be active on the Paladin at any one time.

Unleashing this Seal's energy will judge an enemy for 10 sec, granting attacks and spells used against the judged enemy a chance to restore 59 mana to the attacker. Your melee strikes will refresh the spell's duration. Only one Judgement per Paladin can be active at any one time.
  - Note: Judgement effects now last 40 sec instead of 10.

### Verified unchanged

Devotion Aura, Hammer of Justice, Blessing of Freedom, Concentration Aura, Seal of Justice, Divine Intervention, Seal of Righteousness, Purify, Redemption, Sense Undead, Turn Undead

### Not yet verified (Classic text assumed)

Seal of Command, Greater Blessing of Might, Frost Resistance Aura, Blessing of Sanctuary, Blessing of Sacrifice, Holy Shield, Greater Blessing of Kings, Greater Blessing of Salvation, Greater Blessing of Sanctuary, Blessing of Light, Summon Warhorse, Cleanse, Hammer of Wrath, Holy Shock, Holy Wrath, Greater Blessing of Wisdom, Greater Blessing of Light, Summon Charger

## Talents

### Holy

18 talents, 8 new. 31-point talent: Light's Vigil.

#### Row 1 (0 points)

- **Improved Holy Strike** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Reduces the cooldown of your Holy Strike ability by 1 sec.
  - Rank 2: Reduces the cooldown of your Holy Strike ability by 2 sec.
- **Divine Strength** (5 ranks, column 2) [unchanged]
  - Rank 1: Increases your Strength by 2%.
  - Rank 2: Increases your Strength by 4%.
  - Rank 3: Increases your Strength by 6%.
  - Rank 4: Increases your Strength by 8%.
  - Rank 5: Increases your Strength by 10%.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Divine Intellect** (5 ranks, column 3) [unchanged]
  - Rank 1: Increases your total Intellect by 2%.
  - Rank 2: Increases your total Intellect by 4%.
  - Rank 3: Increases your total Intellect by 6%.
  - Rank 4: Increases your total Intellect by 8%.
  - Rank 5: Increases your total Intellect by 10%.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.

#### Row 2 (5 points)

- **Healing Light** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the amount healed by your Holy Light, Flash of Light, and Holy Shock spells by 4%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Classic (3 ranks):
    - Rank 1: Increases the amount healed by your Holy Light and Flash of Light spells by 4%.
    - Rank 2: Increases the amount healed by your Holy Light and Flash of Light spells by 8%.
    - Rank 3: Increases the amount healed by your Holy Light and Flash of Light spells by 12%.
  - Note: Rank 1 reads 4% in the footage (our earlier 12% was a maxed-rank screenshot). Ranks 2-3 not yet seen.
- **Spiritual Focus** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives your Flash of Light, Holy Light, and Light's Vigil spells a 70% chance to not lose casting time when you take damage.
  - Rank 2: Gives your Flash of Light, Holy Light, and Light's Vigil spells a 100% chance to not lose casting time when you take damage.
  - Classic (5 ranks):
    - Rank 1: Gives your Flash of Light and Holy Light spells a 14% chance to not lose casting time when you take damage.
    - Rank 2: Gives your Flash of Light and Holy Light spells a 28% chance to not lose casting time when you take damage.
    - Rank 3: Gives your Flash of Light and Holy Light spells a 42% chance to not lose casting time when you take damage.
    - Rank 4: Gives your Flash of Light and Holy Light spells a 56% chance to not lose casting time when you take damage.
    - Rank 5: Gives your Flash of Light and Holy Light spells a 70% chance to not lose casting time when you take damage.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Improved Seals** (3 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage done by your Seals and Judgements by 5%.
  - Rank 2: Increases the damage done by your Seals and Judgements by 10%.
  - Rank 3: Increases the damage done by your Seals and Judgements by 15%.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Unyielding Faith** (2 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your chance to resist Fear and Disorient effects by an additional 5%.
  - Rank 2: Increases your chance to resist Fear and Disorient effects by an additional 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Voice of Truth** (1 rank, column 1) [NEW]
  - Cost: Instant, 3 min cooldown
  - Grants you immunity to Silence and Interrupt effects. Lasts 6 sec.
  - Note: Slot from the third-party dataset, not yet verified in footage by us.
- **Reverence** (3 ranks, column 2) [NEW] [ranks 2+ unverified]
  - Rank 1: Allows 10% of your Mana regeneration to continue while casting.
  - Rank 2: Allows 20% of your Mana regeneration to continue while casting.
  - Rank 3: Allows 30% of your Mana regeneration to continue while casting.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Purifying Power** (2 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Reduces the mana cost of your Cleanse and Purify spells by 10% and reduces the cooldown of your Exorcism and Holy Wrath spells by 17%.
  - Rank 2: Reduces the mana cost of your Cleanse and Purify spells by 20% and reduces the cooldown of your Exorcism and Holy Wrath spells by 34%.
  - Note: Higher-rank values taken from the nikftw dataset.

#### Row 4 (15 points)

- **Infusion of Light** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Your Holy Shock and Flash of Light critical hits reduce the cast time of your next Holy Light cast within 15 sec by 0.5 sec.
  - Rank 2: Your Holy Shock and Flash of Light critical hits reduce the cast time of your next Holy Light cast within 15 sec by 1.0 sec.
  - Note: Rank 1 read in footage, rank 2 verified on the BlizzCon slide.
- **Illumination** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Requires: Reverence (3/3)
  - Rank 1: After getting a critical effect from your Flash of Light, Holy Light, Light's Vigil, or Holy Shock heal spell you have a 20% chance to gain Mana equal to 50% of the base cost of the spell.
  - Rank 2: After getting a critical effect from your Flash of Light, Holy Light, Light's Vigil, or Holy Shock heal spell you have a 40% chance to gain Mana equal to 50% of the base cost of the spell.
  - Rank 3: After getting a critical effect from your Flash of Light, Holy Light, Light's Vigil, or Holy Shock heal spell you have a 60% chance to gain Mana equal to 50% of the base cost of the spell.
  - Rank 4: After getting a critical effect from your Flash of Light, Holy Light, Light's Vigil, or Holy Shock heal spell you have a 80% chance to gain Mana equal to 50% of the base cost of the spell.
  - Rank 5: After getting a critical effect from your Flash of Light, Holy Light, Light's Vigil, or Holy Shock heal spell you have a 100% chance to gain Mana equal to 50% of the base cost of the spell.
  - Classic (5 ranks):
    - Rank 1: After getting a critical effect from your Flash of Light, Holy Light, or Holy Shock heal spell, gives you a 20% chance to gain Mana equal to the base cost of the spell.
    - Rank 2: After getting a critical effect from your Flash of Light, Holy Light, or Holy Shock heal spell, gives you a 40% chance to gain Mana equal to the base cost of the spell.
    - Rank 3: After getting a critical effect from your Flash of Light, Holy Light, or Holy Shock heal spell, gives you a 60% chance to gain Mana equal to the base cost of the spell.
    - Rank 4: After getting a critical effect from your Flash of Light, Holy Light, or Holy Shock heal spell, gives you a 80% chance to gain Mana equal to the base cost of the spell.
    - Rank 5: After getting a critical effect from your Flash of Light, Holy Light, or Holy Shock heal spell, gives you a 100% chance to gain Mana equal to the base cost of the spell.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Divine Favor** (1 rank, column 3) [unchanged]
  - Requires: Illumination (5/5)
  - Cost: 37 Mana, Instant, 2 min cooldown
  - When activated, gives your next Flash of Light, Holy Light, or Holy Shock spell a 100% critical effect chance.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.

#### Row 5 (20 points)

- **Divine Precision** (3 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Requires: Holy Shock (1/1)
  - Rank 1: Increases your chance to hit with Holy spells by 6%.
  - Rank 2: Increases your chance to hit with Holy spells by 12%.
  - Rank 3: Increases your chance to hit with Holy spells by 18%.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Holy Shock** (1 rank, column 2) [CHANGED]
  - Cost: 160 Mana, Instant, Enemy: 20 yd range, Friendly: 40 yd range, 10 sec cooldown
  - Blasts the target with Holy energy, causing 138 to 149 Holy damage to an enemy, or 119 to 128 healing to an ally.
  - Classic: Blasts the target with Holy energy, causing 204 to 220 Holy damage to an enemy, or 204 to 220 healing to an ally.
  - Note: Redesigned: 160 Mana, 10 sec cooldown and a 40 yd friendly range where Classic Rank 1 cost 225 Mana on a 30 sec cooldown at 20 yd. The numbers differ between the level 38 demo paladin (129-139 damage, 110-118 healing) and the level 60 BlizzCon slide (shown here).
- **Consecrated Ground** (2 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Gives your Holy spells 5% increased damage against the first 4 enemies that enter your Consecration.
  - Rank 2: Gives your Holy spells 10% increased damage against the first 4 enemies that enter your Consecration.
  - Note: Rank 2 verified on the BlizzCon slide, rank 1 from the nikftw screenshots.

#### Row 6 (25 points)

- **Holy Power** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike chance of your Holy Shock spell by 3%, and all other spells by 1%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Rank 4: Rank 4 values not yet captured from footage.
  - Rank 5: Rank 5 values not yet captured from footage.
  - Classic (5 ranks):
    - Rank 1: Increases the critical effect chance of your Holy spells by 1%.
    - Rank 2: Increases the critical effect chance of your Holy spells by 2%.
    - Rank 3: Increases the critical effect chance of your Holy spells by 3%.
    - Rank 4: Increases the critical effect chance of your Holy spells by 4%.
    - Rank 5: Increases the critical effect chance of your Holy spells by 5%.
  - Note: Redesigned around Holy Shock: rank 1 gives Holy Shock 3% crit and other spells 1%. Ranks 2-5 not yet seen.

#### Row 7 (30 points)

- **Light's Vigil** (1 rank, column 2) [NEW]
  - Requires: Holy Shock (1/1)
  - Cost: 730 Mana, 1.5 sec cast, Enemy: 20 yd range, Friendly: 40 yd range, 6 sec cooldown
  - Applies Light's Vigil to the target for 30 sec. Your next Holy Shock cast on them triggers no cooldown and causes friendly targets to heal their party for 329 to 348, or enemy targets to suffer 184 to 199 Holy damage and refund 75% of Light's Vigil's Mana cost. You may only have 1 Light's Vigil active per Paladin, per party.
  - Note: Holy capstone. The numbers differ between the level 38 demo paladin (315-333 party heal, 175-189 damage) and the level 60 BlizzCon slide (shown here).

### Protection

16 talents, 5 new. 31-point talent: Holy Shield.

#### Row 1 (0 points)

- **Toughness** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your armor value from items by 2%.
  - Rank 2: Increases your armor value from items by 4%.
  - Rank 3: Increases your armor value from items by 6%.
  - Rank 4: Increases your armor value from items by 8%.
  - Rank 5: Increases your armor value from items by 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Redoubt** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Damaging melee attacks against you have a 10% chance to increase your chance to block by 6%. Lasts 10 sec or 5 blocks.
  - Rank 2: Damaging melee attacks against you have a 20% chance to increase your chance to block by 6%. Lasts 10 sec or 5 blocks.
  - Rank 3: Damaging melee attacks against you have a 30% chance to increase your chance to block by 6%. Lasts 10 sec or 5 blocks.
  - Rank 4: Damaging melee attacks against you have a 40% chance to increase your chance to block by 6%. Lasts 10 sec or 5 blocks.
  - Rank 5: Damaging melee attacks against you have a 50% chance to increase your chance to block by 6%. Lasts 10 sec or 5 blocks.
  - Classic (5 ranks):
    - Rank 1: Increases your chance to block attacks with your shield by 6% after being the victim of a critical strike. Lasts 10 sec or 5 blocks.
    - Rank 2: Increases your chance to block attacks with your shield by 12% after being the victim of a critical strike. Lasts 10 sec or 5 blocks.
    - Rank 3: Increases your chance to block attacks with your shield by 18% after being the victim of a critical strike. Lasts 10 sec or 5 blocks.
    - Rank 4: Increases your chance to block attacks with your shield by 24% after being the victim of a critical strike. Lasts 10 sec or 5 blocks.
    - Rank 5: Increases your chance to block attacks with your shield by 30% after being the victim of a critical strike. Lasts 10 sec or 5 blocks.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 2 (5 points)

- **Precision** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your chance to hit with all spells and attacks by 1%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Classic (3 ranks):
    - Rank 1: Increases your chance to hit with melee weapons by 1%.
    - Rank 2: Increases your chance to hit with melee weapons by 2%.
    - Rank 3: Increases your chance to hit with melee weapons by 3%.
  - Note: Now applies to spells as well as attacks. Ranks 2-3 not yet seen.
- **Guardian's Favor** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the cooldown of your Blessing of Protection by 1 min and increases the duration of your Blessing of Freedom by 3 sec.
  - Rank 2: Reduces the cooldown of your Blessing of Protection by 120 sec and increases the duration of your Blessing of Freedom by 6 sec.
  - Classic (2 ranks):
    - Rank 1: Reduces the cooldown of your Blessing of Protection by 60 sec and increases the duration of your Blessing of Freedom by 3 sec.
    - Rank 2: Reduces the cooldown of your Blessing of Protection by 120 sec and increases the duration of your Blessing of Freedom by 6 sec.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Anticipation** (5 ranks, column 4) [unchanged]
  - Rank 1: Increases your Defense Skill by 4.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Increases your Defense skill by 6.
  - Rank 4: Increases your Defense skill by 8.
  - Rank 5: Increases your Defense skill by 10.

#### Row 3 (10 points)

- **Improved Seal of Fury** (1 rank, column 1) [NEW]
  - When Seal of Fury's shield is fully absorbed, restore 60 Mana, increased by 15% per level the attacker is above you, up to 45%.
  - Note: The Mana restored is one per level: 38 on the level 38 demo paladin, 60 on the level 60 BlizzCon slide (shown here).
- **Improved Righteous Fury** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: While Righteous Fury is active, all damage taken is reduced by 2%.
  - Rank 2: While Righteous Fury is active, all damage taken is reduced by 4%.
  - Rank 3: While Righteous Fury is active, all damage taken is reduced by 6%.
  - Classic (3 ranks):
    - Rank 1: Increases the amount of threat generated by your Righteous Fury spell by 16%.
    - Rank 2: Increases the amount of threat generated by your Righteous Fury spell by 33%.
    - Rank 3: Increases the amount of threat generated by your Righteous Fury spell by 50%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Shield Specialization** (3 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Requires: Redoubt (5/5)
  - Rank 1: Increases the amount of damage absorbed by your shield by 10%, and gives your blocks a 33% chance to restore 6% of your maximum Mana. May only occur once every 3 sec.
  - Rank 2: Increases the amount of damage absorbed by your shield by 20%, and gives your blocks a chance to restore 6% of your maximum Mana (rank 2 chance not yet captured, 33% at rank 1, 100% at rank 3). May only occur once every 3 sec.
  - Rank 3: Increases the amount of damage absorbed by your shield by 30%, and gives your blocks a 100% chance to restore 6% of your maximum Mana. May only occur once every 3 sec.
  - Classic (3 ranks):
    - Rank 1: Increases the amount of damage absorbed by your shield by 10%.
    - Rank 2: Increases the amount of damage absorbed by your shield by 20%.
    - Rank 3: Increases the amount of damage absorbed by your shield by 30%.
  - Note: Every rank carries the block-to-Mana clause Classic lacked: a 33% chance at rank 1 (footage) and 100% at rank 3 (BlizzCon slide), the rank 2 chance has not been seen.
- **Sacred Duty** (2 ranks, column 4) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases your total Stamina by 2% and reduces the cooldown of your Divine Shield, Divine Protection, and Templar's Bulwark spells by 30 sec.
  - Rank 2: Increases your total Stamina by 4% and reduces the cooldown of your Divine Shield, Divine Protection, and Templar's Bulwark spells by 60 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Swift Judgement** (1 rank, column 1) [NEW]
  - Requires: Improved Seal of Fury (1/1)
  - Cost: Instant, 1 min cooldown
  - Finishes the remaining cooldown on your Judgement ability and reduces the Mana cost of your next Judgement by 100%.
  - Note: Text verified on the BlizzCon slide. The cooldown read 1 min on the show-floor demo (footage and screenshot) but 50 sec on the panel slide, the demo value is shown.
- **One-Handed Weapon Specialization** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage you deal with one-handed melee weapons by 3%.
  - Rank 2: Increases the damage you deal with one-handed melee weapons by 4%.
  - Rank 3: Increases the damage you deal with one-handed melee weapons by 6%.
  - Classic (5 ranks):
    - Rank 1: Increases the damage you deal with one-handed melee weapons by 2%.
    - Rank 2: Increases the damage you deal with one-handed melee weapons by 4%.
    - Rank 3: Increases the damage you deal with one-handed melee weapons by 6%.
    - Rank 4: Increases the damage you deal with one-handed melee weapons by 8%.
    - Rank 5: Increases the damage you deal with one-handed melee weapons by 10%.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Improved Hammer of Justice** (3 ranks, column 3) [unchanged]
  - Rank 1: Decreases the cooldown of your Hammer of Justice spell by 5 sec.
  - Rank 2: Decreases the cooldown of your Hammer of Justice spell by 10 sec.
  - Rank 3: Decreases the cooldown of your Hammer of Justice spell by 15 sec.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.

#### Row 5 (20 points)

- **Templar's Bulwark** (1 rank, column 2) [NEW]
  - Cost: 110 Mana, Instant, 5 min cooldown
  - When activated, this ability grants you an absorb shield equal to 100% of your maximum health for 8 sec. Applies Forbearance for 1 min. Cannot be cast while Forbearance is active.
  - Note: Text, cost and cooldown verified on the BlizzCon slide.
- **Reckoning** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives you a 8% chance to gain an extra attack after Blocking a melee attack and a 20% chance to gain an extra attack after being the victim of a non-periodic critical strike.
  - Rank 2: Gives you a 16% chance to gain an extra attack after Blocking a melee attack and a 40% chance to gain an extra attack after being the victim of a non-periodic critical strike.
  - Rank 3: Gives you a 24% chance to gain an extra attack after Blocking a melee attack and a 60% chance to gain an extra attack after being the victim of a non-periodic critical strike.
  - Rank 4: Gives you a 32% chance to gain an extra attack after Blocking a melee attack and a 80% chance to gain an extra attack after being the victim of a non-periodic critical strike.
  - Rank 5: Gives you a 40% chance to gain an extra attack after Blocking a melee attack and a 100% chance to gain an extra attack after being the victim of a non-periodic critical strike.
  - Note: Rank 5 verified on the BlizzCon slide, ranks 2-4 from the nikftw dataset.

#### Row 6 (25 points)

- **Iron Creed** (5 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the threat generated by your Holy Strike ability by 5%. While Righteous Fury is active, Holy Strike also reduces your damage taken by 2% for 6 sec.
  - Rank 2: Increases the threat generated by your Holy Strike ability 10%. While Righteous Fury is active, Holy Strike also reduces your damage taken by 4% for 6 sec.
  - Rank 3: Increases the threat generated by your Holy Strike ability 15%. While Righteous Fury is active, Holy Strike also reduces your damage taken by 6% for 6 sec.
  - Rank 4: Increases the threat generated by your Holy Strike ability 20%. While Righteous Fury is active, Holy Strike also reduces your damage taken by 8% for 6 sec.
  - Rank 5: Increases the threat generated by your Holy Strike ability 25%. While Righteous Fury is active, Holy Strike also reduces your damage taken by 10% for 6 sec.
  - Note: Rank 1 verified in footage and rank 5 on the BlizzCon slide (the slide's 'ability 25%' typo is Blizzard's), ranks 2-4 from the nikftw dataset.

#### Row 7 (30 points)

- **Holy Shield** (1 rank, column 2) [CHANGED]
  - Requires: Templar's Bulwark (1/1)
  - Cost: 150 Mana, Instant, 10 sec cooldown
  - Increases chance to block by 20% for 10 sec, and deals 110 Holy damage for each attack blocked while active. Damage caused by Holy Shield causes 20% additional threat. Each block expends a charge. 4 charges.
  - Classic: Increases chance to block by 30% for 10 sec, and deals 65 Holy damage for each attack blocked while active. Damage caused by Holy Shield causes 20% additional threat. Each block expends a charge. 4 charges.
  - Note: 150 Mana, instant, 10 sec cooldown, requires a shield. Capstone slot per the third-party dataset, text from your screenshot.

### Retribution

18 talents, 7 new. 31-point talent: Twist of Light.

#### Row 1 (0 points)

- **Deflection** (5 ranks, column 2) [unchanged]
  - Rank 1: Increases your Parry chance by 1%.
  - Rank 2: Increases your Parry chance by 2%.
  - Rank 3: Increases your Parry chance by 3%.
  - Rank 4: Increases your Parry chance by 4%.
  - Rank 5: Increases your Parry chance by 5%.
- **Benediction** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the Mana cost of all instant cast spells and abilities by 2%.
  - Rank 2: Reduces the Mana cost of all instant cast spells and abilities by 4%.
  - Rank 3: Reduces the Mana cost of all instant cast spells and abilities by 6%.
  - Rank 4: Reduces the Mana cost of all instant cast spells and abilities by 8%.
  - Rank 5: Reduces the Mana cost of all instant cast spells and abilities by 10%.
  - Classic (5 ranks):
    - Rank 1: Reduces the Mana cost of your Judgement and Seal spells by 3%.
    - Rank 2: Reduces the Mana cost of your Judgement and Seal spells by 6%.
    - Rank 3: Reduces the Mana cost of your Judgement and Seal spells by 9%.
    - Rank 4: Reduces the Mana cost of your Judgement and Seal spells by 12%.
    - Rank 5: Reduces the Mana cost of your Judgement and Seal spells by 15%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 2 (5 points)

- **Improved Judgement** (2 ranks, column 1) [unchanged]
  - Rank 1: Decreases the cooldown of your Judgement ability by 1 sec.
  - Rank 2: Decreases the cooldown of your Judgement spell by 2 sec.
- **Holy Conduit** (2 ranks, column 2) [NEW] [ranks 2+ unverified]
  - Rank 1: Reduces the mana cost of your Consecration, Holy Wrath, Exorcism, and Hammer of Wrath spells by 20%.
  - Rank 2: Reduces the mana cost of your Consecration, Holy Wrath, Exorcism, and Hammer of Wrath spells by 40%.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Conviction** (5 ranks, column 3) [unchanged]
  - Rank 1: Increases your chance to get a critical strike with melee attacks by 1%.
  - Rank 2: Increases your chance to get a critical strike with melee weapons by 2%.
  - Rank 3: Increases your chance to get a critical strike with melee weapons by 3%.
  - Rank 4: Increases your chance to get a critical strike with melee weapons by 4%.
  - Rank 5: Increases your chance to get a critical strike with melee weapons by 5%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Vindication** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives your damaging melee attacks a chance to reduce the target's Attack Power by 42, and increase your Attack Power by 1% for 30 sec.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Gives your damaging melee attacks a chance to reduce the target's Attack Power by 204, and increase your Attack Power by 3% for 30 sec.
  - Classic (3 ranks):
    - Rank 1: Gives the Paladin's damaging melee attacks a chance to reduce the target's Strength and Agility by 5% for 10 sec.
    - Rank 2: Gives the Paladin's damaging melee attacks a chance to reduce the target's Strength and Agility by 10% for 10 sec.
    - Rank 3: Gives the Paladin's damaging melee attacks a chance to reduce the target's Strength and Agility by 15% for 10 sec.
  - Note: Redesigned: an Attack Power debuff plus a self buff instead of Classic's Strength and Agility reduction. Rank 3 from the level 60 BlizzCon slide, rank 1 was read at level 38 from the nikftw screenshots, so the two are not on one scale, rank 2 not yet seen.
- **Sanctified Judgement** (3 ranks, column 2) [NEW] [ranks 2+ unverified]
  - Rank 1: Gives your Judgement ability a 33% chance to return 20% of the Mana cost of the judged seal.
  - Rank 2: Gives your Judgement ability a 66% chance to return 40% of the Mana cost of the judged seal.
  - Rank 3: Gives your Judgement ability a 100% chance to return 60% of the Mana cost of the judged seal.
- **Seal of Command** (1 rank, column 3) [unchanged]
  - Cost: 65 Mana, Instant
  - Gives the Paladin a chance to deal additional Holy damage equal to 70% of normal weapon damage. Only one Seal can be active on the Paladin at any one time. Lasts 30 sec. Unleashing this Seal's energy will judge an enemy, instantly causing 68 to 73 Holy damage, 137 to 146 if the target is stunned or incapacitated.
- **Pursuit of Justice** (2 ranks, column 4) [unchanged]
  - Rank 1: Increases movement speed and mounted movement speed by 8%. This does not stack with other movement speed increasing effects.
  - Rank 2: Increases movement and mounted movement speed by 8%. This does not stack with other movement speed increasing effects.

#### Row 4 (15 points)

- **Eye for an Eye** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: All critical strikes against you cause 5% of the damage taken to the attacker as well. The damage caused by Eye for an Eye will not exceed 50% of the Paladin's total health.
  - Rank 2: All critical strikes against you cause 10% of the damage taken to the attacker as well. The damage caused by Eye for an Eye will not exceed 50% of the Paladin's total health.
  - Classic (2 ranks):
    - Rank 1: All spell criticals against you cause 15% of the damage taken to the caster as well. The damage caused by Eye for an Eye will not exceed 50% of the Paladin's total health.
    - Rank 2: All spell criticals against you cause 30% of the damage taken to the caster as well. The damage caused by Eye for an Eye will not exceed 50% of the Paladin's total health.
- **Sacred Arbiter** (1 rank, column 3) [NEW]
  - Increases the damage of your Holy Strike ability by 10% and causes it to refresh all Judgement effects on the target.
  - Note: Text verified on the BlizzCon slide.
- **Crusade** (2 ranks, column 4) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases all damage dealt by 1%. Increased by an additional 1% against Demon and Undead targets.
  - Rank 2: Increases all damage dealt by 2%. Increased by an additional 2% against Demon and Undead targets.

#### Row 5 (20 points)

- **Two-Handed Weapon Specialization** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage you deal with two-handed melee weapons by 3%.
  - Rank 2: Increases the damage you deal with two-handed melee weapons by 6%.
  - Rank 3: Increases the damage you deal with two-handed melee weapons by 9%.
  - Classic (3 ranks):
    - Rank 1: Increases the damage you deal with two-handed melee weapons by 2%.
    - Rank 2: Increases the damage you deal with two-handed melee weapons by 4%.
    - Rank 3: Increases the damage you deal with two-handed melee weapons by 6%.
- **Vengeance** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Requires: Sanctified Judgement (3/3)
  - Rank 1: Increases your Physical and Holy damage dealt by 1% for 30 sec after landing a critical strike. Stacks up to 5 times.
  - Rank 2: Increases your Physical and Holy damage dealt by 2% for 30 sec after landing a critical strike. Stacks up to 5 times.
  - Rank 3: Increases your Physical and Holy damage dealt by 3% for 30 sec after landing a critical strike. Stacks up to 5 times.
  - Classic (5 ranks):
    - Rank 1: Gives you a 3% bonus to Physical and Holy damage you deal for 8 sec after dealing a critical strike from a weapon swing, spell, or ability.
    - Rank 2: Gives you a 6% bonus to Physical and Holy damage you deal for 8 sec after dealing a critical strike from a weapon swing, spell, or ability.
    - Rank 3: Gives you a 9% bonus to Physical and Holy damage you deal for 8 sec after dealing a critical strike from a weapon swing, spell, or ability.
    - Rank 4: Gives you a 12% bonus to Physical and Holy damage you deal for 8 sec after dealing a critical strike from a weapon swing, spell, or ability.
    - Rank 5: Gives you a 15% bonus to Physical and Holy damage you deal for 8 sec after dealing a critical strike from a weapon swing, spell, or ability.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Repentance** (1 rank, column 3) [unchanged]
  - Requires: Seal of Command (1/1)
  - Cost: 60 Mana, Instant, 20 yd range, 1 min cooldown
  - Puts the enemy target in a state of meditation, incapacitating them for up to 6 sec. Any damage caused will awaken the target. Only works against Humanoids.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 6 (25 points)

- **Champion of the Light** (3 ranks, column 2) [NEW] [ranks 2+ unverified]
  - Requires: Vengeance (3/3)
  - Rank 1: Increases your spell damage and healing by up to 33% of your Intellect.
  - Rank 2: Increases your spell damage and healing by up to 66% of your Intellect.
  - Rank 3: Increases your spell damage and healing by up to 100% of your Intellect.
  - Note: Rank 3 verified on the BlizzCon slide (100%, not the 99% read from the nikftw screenshot), rank 1 read in footage.
- **Instrument of Law** (2 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Requires: Repentance (1/1)
  - Rank 1: Reduces the cast time of your Hammer of Wrath by 0.5 sec, and reduces the threat you generate by 10% while Righteous Fury is not active.
  - Rank 2: Reduces the cast time of your Hammer of Wrath by 1.0 sec, and reduces all threat you generate by 20% while Righteous Fury is not active.
  - Note: Rank 1 verified in footage, rank 2 on the BlizzCon slide.

#### Row 7 (30 points)

- **Twist of Light** (1 rank, column 2) [NEW]
  - Requires: Champion of the Light (3/3)
  - When you replace your Seal of Command, Seal of Righteousness, Seal of Fury, or Seal of Justice with a different Seal, gain an Echo. Your next melee attack applies the replaced Seal's effects, consuming the Echo.
  - Note: Retribution capstone, text verified on the BlizzCon slide. The demo tooltip showed Requires Level 40.
