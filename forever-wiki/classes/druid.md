# Druid

Source: wowforevertalents.com dataset (BlizzCon 2026 footage overlaid on Classic Era 1.15.9 client tables), data version 2026-09-13/14. See [../README.md](../README.md) for caveats. Numbers on new talents were read off level 38 demo characters, so absolute damage and mana values are not level 60 values.

Talents: 52 total, 15 new, 23 changed, 14 unchanged (verified).

Row N needs 5*(N-1) points spent in that tree. Row 1 is the top row. Rows 3, 4, 5 and 7 hold the 11, 16, 21 and 31 point talents.

Warning on multi-rank NEW and CHANGED talents: the footage showed rank 1 only. The dataset scaled the other ranks linearly from rank 1, which sometimes produces nonsense (for example Maelstrom Weapon rank 5 reading 'stacks up to 25 times'). Treat rank 1 as observed and the rest as a guess until the beta client is out. The [ranks 2+ unverified] tag marks every multi-rank new or changed talent. Read the Note line, it says when a higher rank was actually seen on a BlizzCon slide.

## Spellbook changes (trainer abilities)

6 of 53 trainer abilities confirmed from footage. Abilities not listed below are either verified unchanged or still Classic placeholders (unverified).

### New abilities

- **Revive** (restoration)
  - Returns the spirit to the body, restoring a dead target to life. Cannot be cast when in combat. Exact values not yet captured from footage.
  - Note: Out-of-combat resurrection, Rank 3 known by level 38. Tooltip not yet captured.
- **Omen of Clarity** (balance, Passive)
  - Your spells and attacks have a chance to grant you Clearcasting, reducing the Mana, Rage, or Energy cost of your next damage or healing spell or offensive ability by 100%. Clearcasting is not consumed by Wrath or by spells or abilities that cost no resources.
  - Note: A Restoration talent in Classic, in Forever it is a baseline passive listed in the Balance tab.

### Changed abilities

- **Enrage** (feral-combat, Level 12, Instant, 1 min cooldown)
  - Instantly generates 10 Rage and another 20 Rage over 10 sec, but reduces base armor by 27% in Bear Form and 16% in Dire Bear Form. The druid is considered in combat for the duration.
  - Classic: Generates 20 rage over 10 sec, but reduces base armor by 27% in Bear Form and 16% in Dire Bear Form. The druid is considered in combat for the duration.
  - Note: Now grants 10 Rage instantly on top of the 20 Rage over 10 sec.
- **Tiger's Fury** (feral-combat, Level 24, 30 Energy, Instant, 1 sec cooldown)
  - Increases damage done by 40 for 6 sec.
  - Note: Listed without a rank at level 38, where Classic shows Rank 2, appears to be a single-rank ability now. Values not yet captured.
- **Wrath** (balance, Level 38, 80 Mana, 30 yd range, 2 sec cast)
  - Causes 38 to 43 Nature damage to the target.
  - Classic: Causes 249 to 277 Nature damage to the target.
  - Note: The Rank 6 tooltip on a level 38 druid read 80 Mana, 2 sec cast, 38 to 43 damage, where Classic Rank 6 costs 125 Mana for 149 to 167 damage. Other ranks not seen.

### Removed abilities

- **Cure Poison** (restoration, Level 14, 16% of base mana, 30 yd range, Instant)
  - Cures 1 poison effect on the target.
  - Note: Not in the level 38 demo druid's Restoration tab, which shows every other Classic spell of that level, Abolish Poison is present.

### Not yet verified (Classic text assumed)

Bear Form, Demoralizing Roar, Growl, Maul, Bash, Aquatic Form, Swipe, Cat Form, Claw, Prowl, Rip, Shred, Rake, Dash, Challenging Roar, Cower, Faerie Fire (Feral), Travel Form, Ferocious Bite, Ravage, Track Humanoids, Frenzied Regeneration, Pounce, Dire Bear Form, Feline Grace, Healing Touch, Mark of the Wild, Rejuvenation, Regrowth, Rebirth, Remove Curse, Abolish Poison, Insect Swarm, Tranquility, Innervate, Gift of the Wild, Moonfire, Thorns, Entangling Roots, Teleport: Moonglade, Faerie Fire, Hibernate, Nature's Grasp, Starfire, Soothe Animal, Hurricane, Barkskin

## Talents

### Balance

17 talents, 6 new. 31-point talent: Moonkin Form.

#### Row 1 (0 points)

- **Improved Wrath** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the cast time of your Wrath spell by 0.1 sec and its Mana cost by 10%.
  - Rank 2: Reduces the cast time of your Wrath spell by 0.2 sec and its Mana cost by 20%.
  - Rank 3: Reduces the cast time of your Wrath spell by 0.3 sec and its Mana cost by 30%.
  - Rank 4: Reduces the cast time of your Wrath spell by 0.4 sec and its Mana cost by 40%.
  - Rank 5: Reduces the cast time of your Wrath spell by 0.5 sec and its Mana cost by 50%.
  - Classic (5 ranks):
    - Rank 1: Reduces the cast time of your Wrath spell by 0.1 sec.
    - Rank 2: Reduces the cast time of your Wrath spell by 0.2 sec.
    - Rank 3: Reduces the cast time of your Wrath spell by 0.3 sec.
    - Rank 4: Reduces the cast time of your Wrath spell by 0.4 sec.
    - Rank 5: Reduces the cast time of your Wrath spell by 0.5 sec.
- **Genesis** (5 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the periodic damage and healing done by your spells and abilities by 1%.
  - Rank 2: Increases the periodic damage and healing done by your spells and abilities by 2%.
  - Rank 3: Increases the periodic damage and healing done by your spells and abilities by 3%.
  - Rank 4: Increases the periodic damage and healing done by your spells and abilities by 4%.
  - Rank 5: Increases the periodic damage and healing done by your spells and abilities by 5%.
  - Note: Replaced a missing zamimg file (spell_arcane_natureguardian) with the WotLK Genesis icon.

#### Row 2 (5 points)

- **Moonglow** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the Mana cost of your spells by 3%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Classic (3 ranks):
    - Rank 1: Reduces the Mana cost of your Moonfire, Starfire, Wrath, Healing Touch, Regrowth and Rejuvenation spells by 3%.
    - Rank 2: Reduces the Mana cost of your Moonfire, Starfire, Wrath, Healing Touch, Regrowth and Rejuvenation spells by 6%.
    - Rank 3: Reduces the Mana cost of your Moonfire, Starfire, Wrath, Healing Touch, Regrowth and Rejuvenation spells by 9%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Moonfire** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage and critical strike chance of your Moonfire spell by 5%.
  - Rank 2: Increases the damage and critical strike chance of your Moonfire spell by 10%.
  - Classic (5 ranks):
    - Rank 1: Increases the damage and critical strike chance of your Moonfire spell by 2%.
    - Rank 2: Increases the damage and critical strike chance of your Moonfire spell by 4%.
    - Rank 3: Increases the damage and critical strike chance of your Moonfire spell by 6%.
    - Rank 4: Increases the damage and critical strike chance of your Moonfire spell by 8%.
    - Rank 5: Increases the damage and critical strike chance of your Moonfire spell by 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Nature's Majesty** (2 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases your critical strike chance with spells and melee attacks by 2%.
  - Rank 2: Increases your critical strike chance with spells and melee attacks by 4%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Nature's Reach** (2 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the range of your offensive Balance spells by 10% and increases the chance for all your spells and attacks to hit by 2%.
  - Rank 2: Increases the range of your offensive Balance spells by 20% and increases the chance for all your spells and attacks to hit by 4%.
  - Classic (2 ranks):
    - Rank 1: Increases the range of your Wrath, Entangling Roots, Faerie Fire, Moonfire, Starfire, and Hurricane spells by 10%.
    - Rank 2: Increases the range of your Wrath, Entangling Roots, Faerie Fire, Moonfire, Starfire, and Hurricane spells by 20%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Improved Entangling Roots** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage done by your Entangling Roots spell by 25%, and its victims can take up to 25% more damage without interrupting the effect.
  - Rank 2: Increases the damage done by your Entangling Roots spell by 50%, and its victims can take up to 50% more damage without interrupting the effect.
  - Rank 3: Increases the damage done by your Entangling Roots spell by 75%, and its victims can take up to 75% more damage without interrupting the effect.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Nature's Splendor** (1 rank, column 3) [NEW]
  - Requires: Nature's Majesty (2/2)
  - Increases the duration of your Moonfire and Rejuvenation spells by 3 sec, your Regrowth spell by 6 sec, and your Insect Swarm spell by 2 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Balance of Nature** (5 ranks, column 4) [NEW] [ranks 2+ unverified]
  - Rank 1: Each time you cast a Nature spell, your next Arcane damage spell within 10 sec deals 1% increased damage. Each time you cast an Arcane spell, your next Nature damage spell within 10 sec deals 1% increased damage.
  - Rank 2: Each time you cast a Nature spell, your next Arcane damage spell within 10 sec deals 2% increased damage. Each time you cast an Arcane spell, your next Nature damage spell within 10 sec deals 2% increased damage.
  - Rank 3: Each time you cast a Nature spell, your next Arcane damage spell within 10 sec deals 3% increased damage. Each time you cast an Arcane spell, your next Nature damage spell within 10 sec deals 3% increased damage.
  - Rank 4: Each time you cast a Nature spell, your next Arcane damage spell within 10 sec deals 4% increased damage. Each time you cast an Arcane spell, your next Nature damage spell within 10 sec deals 4% increased damage.
  - Rank 5: Each time you cast a Nature spell, your next Arcane damage spell within 10 sec deals 5% increased damage. Each time you cast an Arcane spell, your next Nature damage spell within 10 sec deals 5% increased damage.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Insect Swarm** (1 rank, column 1) [CHANGED]
  - Cost: 45 Mana, 30 yd range, Instant
  - The enemy target is swarmed by insects, decreasing their chance to hit by 2% and causing 55 Nature damage over 12 sec.
  - Classic: The enemy target is swarmed by insects, decreasing their chance to hit by 2% and causing 66 Nature damage over 12 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Vengeance** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Requires: Improved Moonfire (2/2)
  - Rank 1: Increases the critical strike damage bonus of your Arcane and Nature spells by 20%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Rank 4: Rank 4 values not yet captured from footage.
  - Rank 5: Rank 5 values not yet captured from footage.
  - Classic (5 ranks):
    - Rank 1: Increases the critical strike damage bonus of your Starfire, Moonfire, and Wrath spells by 20%.
    - Rank 2: Increases the critical strike damage bonus of your Starfire, Moonfire, and Wrath spells by 40%.
    - Rank 3: Increases the critical strike damage bonus of your Starfire, Moonfire, and Wrath spells by 60%.
    - Rank 4: Increases the critical strike damage bonus of your Starfire, Moonfire, and Wrath spells by 80%.
    - Rank 5: Increases the critical strike damage bonus of your Starfire, Moonfire, and Wrath spells by 100%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Starfire** (5 ranks, column 3) [unchanged]
  - Rank 1: Reduces the cast time of Starfire by 0.1 sec and Starfire has a 3% chance to stun its target for 3 sec.
  - Rank 2: Reduces the cast time of Starfire by 0.2 sec and Starfire has a 6% chance to stun its target for 3 sec.
  - Rank 3: Reduces the cast time of Starfire by 0.3 sec and Starfire has a 9% chance to stun its target for 3 sec.
  - Rank 4: Reduces the cast time of Starfire by 0.4 sec and Starfire has a 12% chance to stun its target for 3 sec.
  - Rank 5: Reduces the cast time of Starfire by 0.5 sec and Starfire has a 15% chance to stun its target for 3 sec.
  - Note: Higher-rank values taken from the nikftw dataset.

#### Row 5 (20 points)

- **Overgrowth** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the maximum number of targets you may have affected by Entangling Roots by 1.
  - Rank 2: Increases the maximum number of targets you may have affected by Entangling Roots by 2.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Nature's Grace** (1 rank, column 2) [unchanged]
  - All non-periodic spell criticals grace you with a blessing of nature, increasing your spellcasting speed and reducing your global cooldown by 10% for 3 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Eclipse** (3 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Your Wrath spell reduces the cast time of your next 2 Starfire spells by 0.17 sec. Stores up to 4 charges. Lasts 15 sec.
  - Rank 2: Your Wrath spell reduces the cast time of your next 2 Starfire spells by 0.34 sec. Stores up to 4 charges. Lasts 15 sec.
  - Rank 3: Your Wrath spell reduces the cast time of your next 2 Starfire spells by 0.50 sec. Stores up to 4 charges. Lasts 15 sec.
  - Note: Higher-rank values taken from the nikftw dataset.

#### Row 6 (25 points)

- **Moonfury** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage done by your Arcane and Nature spells by 1%.
  - Rank 2: Increases the damage done by your Arcane and Nature spells by 2%.
  - Rank 3: Increases the damage done by your Arcane and Nature spells by 3%.
  - Rank 4: Increases the damage done by your Arcane and Nature spells by 4%.
  - Rank 5: Increases the damage done by your Arcane and Nature spells by 5%.
  - Classic (5 ranks):
    - Rank 1: Increases the damage done by your Starfire, Moonfire and Wrath spells by 2%.
    - Rank 2: Increases the damage done by your Starfire, Moonfire and Wrath spells by 4%.
    - Rank 3: Increases the damage done by your Starfire, Moonfire and Wrath spells by 6%.
    - Rank 4: Increases the damage done by your Starfire, Moonfire and Wrath spells by 8%.
    - Rank 5: Increases the damage done by your Starfire, Moonfire and Wrath spells by 10%.

#### Row 7 (30 points)

- **Moonkin Form** (1 rank, column 2) [CHANGED]
  - Cost: 283 Mana, Instant
  - Transforms the Druid into Moonkin Form. While in this form, the armor contribution from items is increased by 360% and all party members within 45 yards have their critical chance increased by 3%, exclusive with Leader of the Pack. The Moonkin cannot cast healing spells while shapeshifted. The act of shapeshifting frees the caster of Polymorph and Movement Impairing effects.
  - Classic: Transforms the Druid into Moonkin Form. While in this form the armor contribution from items is increased by 360% and all party members within 30 yards have their spell critical chance increased by 3%. The Moonkin can only cast Balance spells while shapeshifted.

The act of shapeshifting frees the caster of Polymorph and Movement Impairing effects.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

### Feral Combat

19 talents, 6 new. 31-point talent: Berserk.

#### Row 1 (0 points)

- **Ferocity** (5 ranks, column 2) [unchanged]
  - Rank 1: Reduces the cost of your Maul, Swipe, Claw, and Rake abilities by 1 Rage or Energy.
  - Rank 2: Reduces the cost of your Maul, Swipe, Claw, and Rake abilities by 2 Rage or Energy.
  - Rank 3: Reduces the cost of your Maul, Swipe, Claw, and Rake abilities by 3 Rage or Energy.
  - Rank 4: Reduces the cost of your Maul, Swipe, Claw, and Rake abilities by 4 Rage or Energy.
  - Rank 5: Reduces the cost of your Maul, Swipe, Claw, and Rake abilities by 5 Rage or Energy.
- **Heart of the Wild** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your Intellect by 2%. In addition, while in Bear Form or Dire Bear Form your Stamina is increased by 4% and while in Cat Form your Strength is increased by 2%.
  - Rank 2: Increases your Intellect by 4%. In addition, while in Bear Form or Dire Bear Form your Stamina is increased by 8% and while in Cat Form your Strength is increased by 4%.
  - Rank 3: Increases your Intellect by 6%. In addition, while in Bear Form or Dire Bear Form your Stamina is increased by 12% and while in Cat Form your Strength is increased by 6%.
  - Rank 4: Increases your Intellect by 8%. In addition, while in Bear Form or Dire Bear Form your Stamina is increased by 16% and while in Cat Form your Strength is increased by 8%.
  - Rank 5: Increases your Intellect by 10%. In addition, while in Bear Form or Dire Bear Form your Stamina is increased by 20% and while in Cat Form your Strength is increased by 10%.
  - Classic (5 ranks):
    - Rank 1: Increases your Intellect by 4%. In addition, while in Bear or Dire Bear Form your Stamina is increased by 4% and while in Cat Form your Strength is increased by 4%.
    - Rank 2: Increases your Intellect by 8%. In addition, while in Bear or Dire Bear Form your Stamina is increased by 8% and while in Cat Form your Strength is increased by 8%.
    - Rank 3: Increases your Intellect by 12%. In addition, while in Bear or Dire Bear Form your Stamina is increased by 12% and while in Cat Form your Strength is increased by 12%.
    - Rank 4: Increases your Intellect by 16%. In addition, while in Bear or Dire Bear Form your Stamina is increased by 16% and while in Cat Form your Strength is increased by 16%.
    - Rank 5: Increases your Intellect by 20%. In addition, while in Bear or Dire Bear Form your Stamina is increased by 20% and while in Cat Form your Strength is increased by 20%.
  - Note: Rank 1 from footage (2% Int, 4% Bear/Dire Bear Stam, 2% Cat Str). Ranks 2-5 use the same per-rank step as Classic (Int +2, Stam +4, Str +2 each rank).

#### Row 2 (5 points)

- **Feral Swiftness** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your movement speed while in Cat Form by 15%, and increases your chance to Dodge by 2%.
  - Rank 2: Increases your movement speed while in Cat Form by 30%, and increases your chance to Dodge by 4%.
  - Note: Replaced a missing zamimg file (spell_druid_feralswiftness) with the WotLK/vanilla icon.
- **Feral Instinct** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases threat caused in Bear Form and Dire Bear Form by 5% and reduces the chance enemies have to detect you while Prowling.
  - Rank 2: Increases threat caused in Bear Form and Dire Bear Form by 10% and reduces the chance enemies have to detect you while Prowling.
  - Rank 3: Increases threat caused in Bear Form and Dire Bear Form by 15% and reduces the chance enemies have to detect you while Prowling.
  - Classic (5 ranks):
    - Rank 1: Increases threat caused in Bear and Dire Bear Form by 3% and reduces the chance enemies have to detect you while Prowling.
    - Rank 2: Increases threat caused in Bear and Dire Bear Form by 6% and reduces the chance enemies have to detect you while Prowling.
    - Rank 3: Increases threat caused in Bear and Dire Bear Form by 9% and reduces the chance enemies have to detect you while Prowling.
    - Rank 4: Increases threat caused in Bear and Dire Bear Form by 12% and reduces the chance enemies have to detect you while Prowling.
    - Rank 5: Increases threat caused in Bear and Dire Bear Form by 15% and reduces the chance enemies have to detect you while Prowling.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Brutal Impact** (2 ranks, column 3) [unchanged]
  - Rank 1: Increases the stun duration of your Bash and Pounce abilities by 0.5 sec and reduces the cooldown of Bash by 15 sec.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Thick Hide** (3 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Rank 1: While in Bear Form, Cat Form, Dire Bear Form, or Moonkin Form, you gain 1 additional base Armor per level and another 0.67 base Armor for each point of defense skill beyond five times your level. This amount can be further increased by multipliers from those forms.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Classic (5 ranks):
    - Rank 1: Increases your Armor contribution from items by 2%.
    - Rank 2: Increases your Armor contribution from items by 4%.
    - Rank 3: Increases your Armor contribution from items by 6%.
    - Rank 4: Increases your Armor contribution from items by 8%.
    - Rank 5: Increases your Armor contribution from items by 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Savage Fury** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage caused by your Claw, Rake, Shred, Maul, and Swipe abilities by 5%.
  - Rank 2: Increases the damage caused by your Claw, Rake, Shred, Maul, and Swipe abilities by 10%.
  - Classic (2 ranks):
    - Rank 1: Increases the damage caused by your Claw, Rake, Maul and Swipe abilities by 10%.
    - Rank 2: Increases the damage caused by your Claw, Rake, Maul and Swipe abilities by 20%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Feral Charge** (1 rank, column 3) [CHANGED]
  - Cost: 5 Rage, 8-25 yd range, Instant, 15 sec cooldown, Feral Charge (Cat): 8-25 yd range, Instant, 30 sec cooldown
  - Charge an enemy, immobilizing them and interrupting any spell they are casting for 4 sec.

Feral Charge (Cat)
Leap behind an enemy, Dazing them for 3 sec.
  - Classic: Causes you to charge an enemy, immobilizing and interrupting any spell being cast for 4 sec.
- **Sharpened Claws** (2 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your critical strike chance while in Bear Form, Dire Bear Form, or Cat Form by 3%.
  - Rank 2: Increases your critical strike chance while in Bear Form, Dire Bear Form, or Cat Form by 6%.
  - Classic (3 ranks):
    - Rank 1: Increases your critical strike chance while in Bear, Dire Bear or Cat Form by 2%.
    - Rank 2: Increases your critical strike chance while in Bear, Dire Bear or Cat Form by 4%.
    - Rank 3: Increases your critical strike chance while in Bear, Dire Bear or Cat Form by 6%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Shredding Attacks** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the Energy cost of your Shred ability by 6 and reduces the Rage cost of your Lacerate ability by 1.
  - Rank 2: Reduces the Energy cost of your Shred ability by 12 and reduces the Rage cost of your Lacerate ability by 2.
  - Rank 3: Reduces the Energy cost of your Shred ability by 18 and reduces the Rage cost of your Lacerate ability by 3.
  - Classic (2 ranks):
    - Rank 1: Reduces the Energy cost of your Shred ability by 6.
    - Rank 2: Reduces the Energy cost of your Shred ability by 12.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Mangle** (1 rank, column 2) [NEW]
  - Requires: Savage Fury (2/2)
  - Cost: 20 Rage, Melee Range, Instant, 6 sec cooldown
  - Mangle the target for 100% normal damage plus 26.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Predatory Strikes** (3 ranks, column 3) [unchanged]
  - Rank 1: Increases your melee attack power in Cat, Bear, and Dire Bear Forms by 50% of your level.
  - Rank 2: Increases your melee attack power in Cat, Bear, and Dire Bear Forms by 100% of your level.
  - Rank 3: Increases your melee attack power in Cat, Bear, and Dire Bear Forms by 150% of your level.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Primal Fury** (2 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Requires: Sharpened Claws (2/2)
  - Rank 1: Gives you a 50% chance to gain an additional 5 Rage any time you get a critical strike while in Bear Form or Dire Bear Form. In addition, your non-periodic critical strikes from Cat Form abilities that generate Combo Points have a 50% chance to add an additional Combo Point.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Classic (2 ranks):
    - Rank 1: Gives you a 50% chance to gain an additional 5 Rage anytime you get a critical strike while in Bear and Dire Bear Form.
    - Rank 2: Gives you a 100% chance to gain an additional 5 Rage anytime you get a critical strike while in Bear and Dire Bear Form.

#### Row 5 (20 points)

- **Predatory Instincts** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike damage bonus of your melee abilities by 10%.
  - Rank 2: Increases the critical strike damage bonus of your melee abilities by 20%.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Leader of the Pack** (1 rank, column 2) [CHANGED]
  - While in Cat Form, Bear Form, or Dire Bear Form, the Leader of the Pack increases the critical strike chance of all party members within 45 yards by 3%, exclusive with Moonkin Aura.
  - Classic: While in Cat, Bear or Dire Bear Form, the Leader of the Pack increases ranged and melee critical chance of all party members within 45 yards by 3%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **King of the Jungle** (3 ranks, column 4) [NEW] [ranks 2+ unverified]
  - Rank 1: Tiger's Fury now instantly grants you 20 Energy.
  - Rank 2: Tiger's Fury now instantly grants you 40 Energy.
  - Rank 3: Tiger's Fury now instantly grants you 60 Energy.
  - Note: Filled the question-mark placeholder with the WotLK talent icon.

#### Row 6 (25 points)

- **Natural Reaction** (5 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases your dodge chance by 1%, and gives you a 20% chance to gain 5 Rage each time you dodge.
  - Rank 2: Increases your dodge chance by 2%, and gives you a 40% chance to gain 5 Rage each time you dodge.
  - Rank 3: Increases your dodge chance by 3%, and gives you a 60% chance to gain 5 Rage each time you dodge.
  - Rank 4: Increases your dodge chance by 4%, and gives you a 80% chance to gain 5 Rage each time you dodge.
  - Rank 5: Increases your dodge chance by 5%, and gives you a 100% chance to gain 5 Rage each time you dodge.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Rend and Tear** (5 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Requires: Predatory Strikes (3/3)
  - Rank 1: Increases damage done by your melee abilities on Bleeding targets by 2%.
  - Rank 2: Increases damage done by your melee abilities on Bleeding targets by 4%.
  - Rank 3: Increases damage done by your melee abilities on Bleeding targets by 6%.
  - Rank 4: Increases damage done by your melee abilities on Bleeding targets by 8%.
  - Rank 5: Increases damage done by your melee abilities on Bleeding targets by 10%.
  - Note: Filled the question-mark placeholder with the WotLK talent icon.

#### Row 7 (30 points)

- **Berserk** (1 rank, column 2) [NEW]
  - Requires: Leader of the Pack (1/1)
  - Cost: Instant, 3 min cooldown
  - Causes your Mangle ability to strike up to 3 targets, removes its cooldown, and increases the critical strike chance of your Combo Point-generating abilities by 100%. Clears and grants immunity to Fear effects for the duration. Lasts 15 sec.

### Restoration

16 talents, 3 new. 31-point talent: Wild Growth.

#### Row 1 (0 points)

- **Nature's Focus** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives you a 14% chance to avoid interruption caused by damage while casting Arcane and Nature spells.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Rank 4: Rank 4 values not yet captured from footage.
  - Rank 5: Rank 5 values not yet captured from footage.
  - Classic (5 ranks):
    - Rank 1: Gives you a 14% chance to avoid interruption caused by damage while casting the Healing Touch, Regrowth and Tranquility spells.
    - Rank 2: Gives you a 28% chance to avoid interruption caused by damage while casting the Healing Touch, Regrowth and Tranquility spells.
    - Rank 3: Gives you a 42% chance to avoid interruption caused by damage while casting the Healing Touch, Regrowth and Tranquility spells.
    - Rank 4: Gives you a 56% chance to avoid interruption caused by damage while casting the Healing Touch, Regrowth and Tranquility spells.
    - Rank 5: Gives you a 70% chance to avoid interruption caused by damage while casting the Healing Touch, Regrowth and Tranquility spells.
  - Note: Now protects all Arcane and Nature spells instead of Healing Touch, Regrowth and Tranquility. Only rank 1 seen.
- **Furor** (5 ranks, column 3) [unchanged]
  - Rank 1: Gives you a 20% chance to gain 10 Rage when you shapeshift into Bear Form or Dire Bear Form. When you shift into Cat Form, you will regain 20% of the Energy you had when you were last in Cat Form, plus 2 Energy for each second you spent not in Bear Form, Cat Form, or Dire Bear Form, up to a maximum of 20 Energy.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Rank 4: Rank 4 values not yet captured from footage.
  - Rank 5: Gives you a 100% chance to gain 10 Rage when you shapeshift into Bear Form or Dire Bear Form. When you shift into Cat Form, you will regain 100% of the Energy you had when you were last in Cat Form, plus 10 Energy for each second you spent not in Bear Form, Cat Form, or Dire Bear Form, up to a maximum of 100 Energy.
  - Note: Ranks 1 and 5 read from footage, ranks 2-4 not yet seen.

#### Row 2 (5 points)

- **Naturalist** (5 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the cast time of your Healing Touch spell by 0.1 sec and increases all damage you deal by 1%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Rank 4: Rank 4 values not yet captured from footage.
  - Rank 5: Rank 5 values not yet captured from footage.
  - Note: Forever adds a damage bonus per rank. Only rank 1 seen.
- **Subtlety** (3 ranks, column 2) [unchanged]
  - Rank 1: Reduces the threat generated by your Healing spells by 4%.
  - Rank 2: Reduces the threat generated by your Healing spells by 8%.
  - Rank 3: Reduces the threat generated by your Healing spells by 12%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Natural Shapeshifter** (3 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the mana cost of all shapeshifting by 10%.
  - Rank 2: Reduces the mana cost of all shapeshifting by 20%.
  - Rank 3: Reduces the mana cost of all shapeshifting by 30%.

#### Row 3 (10 points)

- **Reflection** (3 ranks, column 2) [unchanged]
  - Rank 1: Allows 5% of your Mana regeneration to continue while casting.
  - Rank 2: Allows 10% of your Mana regeneration to continue while casting.
  - Rank 3: Allows 15% of your Mana regeneration to continue while casting.
- **Gift of Nature** (5 ranks, column 3) [unchanged]
  - Rank 1: Increases the effect of all healing spells by 2%.
  - Rank 2: Increases the effect of all healing spells by 4%.
  - Rank 3: Increases the effect of all healing spells by 6%.
  - Rank 4: Increases the effect of all healing spells by 8%.
  - Rank 5: Increases the effect of all healing spells by 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Gift of the Earthmother** (1 rank, column 4) [NEW]
  - Reduces the global cooldown by 0.5 seconds on your Rejuvenation, Swiftmend, and Wild Growth spells.
  - Note: Placed from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Tranquil Spirit** (5 ranks, column 2) [unchanged]
  - Rank 1: Reduces the mana cost of your Healing Touch and Tranquility spells by 2%.
  - Rank 2: Reduces the mana cost of your Healing Touch and Tranquility spells by 4%.
  - Rank 3: Reduces the mana cost of your Healing Touch and Tranquility spells by 6%.
  - Rank 4: Reduces the mana cost of your Healing Touch and Tranquility spells by 8%.
  - Rank 5: Reduces the mana cost of your Healing Touch and Tranquility spells by 10%.
- **Improved Rejuvenation** (3 ranks, column 3) [unchanged]
  - Rank 1: Increases the effect of your Rejuvenation spell by 5%.
  - Rank 2: Increases the effect of your Rejuvenation spell by 10%.
  - Rank 3: Increases the effect of your Rejuvenation spell by 15%.
- **Swiftmend** (1 rank, column 4) [CHANGED]
  - Cost: 162 Mana, 40 yd range, Instant, 15 sec cooldown
  - Instantly heals a target with an active Rejuvenation or Regrowth effect for an amount equal to the full duration of the periodic effect of one of those spells.
  - Classic: Consumes a Rejuvenation or Regrowth effect on a friendly target to instantly heal them an amount equal to 12 sec. of Rejuvenation or 18 sec. of Regrowth.

#### Row 5 (20 points)

- **Nature's Swiftness** (1 rank, column 1) [unchanged]
  - Requires: Naturalist (5/5)
  - Cost: Instant, 3 min cooldown
  - When activated, your next Nature spell becomes an instant cast spell.
- **Living Spirit** (3 ranks, column 2) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases your Spirit by 5%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Note: Placed from the third-party dataset (rank 1 text only), not yet verified in footage by us.
- **Improved Tranquility** (2 ranks, column 4) [unchanged]
  - Rank 1: Reduces threat caused by Tranquility by 50%.
  - Rank 2: Reduces threat caused by Tranquility by 100%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 6 (25 points)

- **Improved Regrowth** (5 ranks, column 3) [unchanged]
  - Requires: Improved Rejuvenation (3/3)
  - Rank 1: Increases the critical effect chance of your Regrowth spell by 10%.
  - Rank 2: Increases the critical effect chance of your Regrowth spell by 20%.
  - Rank 3: Increases the critical effect chance of your Regrowth spell by 30%.
  - Rank 4: Increases the critical effect chance of your Regrowth spell by 40%.
  - Rank 5: Increases the critical effect chance of your Regrowth spell by 50%.
  - Note: Placed from the third-party dataset, not yet verified in footage by us.

#### Row 7 (30 points)

- **Wild Growth** (1 rank, column 2) [NEW]
  - Requires: Living Spirit (3/3)
  - Cost: 550 Mana, 40 yd range, Instant, 6 sec cooldown
  - Heals the target and their party for 285 over 7 sec. Party members must be within 43 yards of target. The amount healed is applied quickly at first, and slows down as Wild Growth reaches its full duration.
  - Note: Replaced a missing zamimg file (ability_druid_wildgrowth) with the WotLK talent icon.
