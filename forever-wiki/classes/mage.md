# Mage

Source: wowforevertalents.com dataset (BlizzCon 2026 footage overlaid on Classic Era 1.15.9 client tables), data version 2026-09-13/14. See [../README.md](../README.md) for caveats. Numbers on new talents were read off level 38 demo characters, so absolute damage and mana values are not level 60 values.

Talents: 54 total, 6 new, 32 changed, 16 unchanged (verified).

Row N needs 5*(N-1) points spent in that tree. Row 1 is the top row. Rows 3, 4, 5 and 7 hold the 11, 16, 21 and 31 point talents.

Warning on multi-rank NEW and CHANGED talents: the footage showed rank 1 only. The dataset scaled the other ranks linearly from rank 1, which sometimes produces nonsense (for example Maelstrom Weapon rank 5 reading 'stacks up to 25 times'). Treat rank 1 as observed and the rest as a guess until the beta client is out. The [ranks 2+ unverified] tag marks every multi-rank new or changed talent. Read the Note line, it says when a higher rank was actually seen on a BlizzCon slide.

## Spellbook changes (trainer abilities)

0 of 48 trainer abilities confirmed from footage. Abilities not listed below are either verified unchanged or still Classic placeholders (unverified).

### Not yet verified (Classic text assumed)

Frost Armor, Frostbolt, Frost Nova, Blizzard, Frost Ward, Cone of Cold, Ice Armor, Ice Barrier, Fireball, Fire Blast, Flamestrike, Fire Ward, Scorch, Pyroblast, Blast Wave, Arcane Intellect, Conjure Water, Conjure Food, Arcane Missiles, Polymorph, Dampen Magic, Slow Fall, Arcane Explosion, Detect Magic, Amplify Magic, Remove Lesser Curse, Blink, Evocation, Mana Shield, Teleport: Ironforge, Teleport: Orgrimmar, Teleport: Stormwind, Teleport: Undercity, Counterspell, Conjure Mana Agate, Teleport: Darnassus, Teleport: Thunder Bluff, Mage Armor, Conjure Mana Jade, Portal: Ironforge, Portal: Orgrimmar, Portal: Stormwind, Portal: Undercity, Conjure Mana Citrine, Portal: Darnassus, Portal: Thunder Bluff, Arcane Brilliance, Conjure Mana Ruby

## Talents

### Arcane

18 talents, 3 new. 31-point talent: Arcane Power.

#### Row 1 (0 points)

- **Wand Specialization** (2 ranks, column 1) [unchanged]
  - Rank 1: Increases your damage with Wands by 13%.
  - Rank 2: Increases your damage with Wands by 25%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Arcane Focus** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your chance to hit with your Arcane spells by 1%.
  - Rank 2: Increases your chance to hit with your Arcane spells by 2%.
  - Rank 3: Increases your chance to hit with your Arcane spells by 3%.
  - Rank 4: Increases your chance to hit with your Arcane spells by 4%.
  - Rank 5: Increases your chance to hit with your Arcane spells by 5%.
  - Classic (5 ranks):
    - Rank 1: Reduces the chance that the opponent can resist your Arcane spells by 2%.
    - Rank 2: Reduces the chance that the opponent can resist your Arcane spells by 4%.
    - Rank 3: Reduces the chance that the opponent can resist your Arcane spells by 6%.
    - Rank 4: Reduces the chance that the opponent can resist your Arcane spells by 8%.
    - Rank 5: Reduces the chance that the opponent can resist your Arcane spells by 10%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Improved Channeling** (5 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Gives you a 20% chance to avoid interruption caused by damage while channeling Arcane Missiles and a 14% chance while casting Arcane Blast.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Rank 4: Rank 4 values not yet captured from footage.
  - Rank 5: Rank 5 values not yet captured from footage.

#### Row 2 (5 points)

- **Arcane Subtlety** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces your target's resistance to all your spells by 8 and reduces the threat caused by your Arcane spells by 15%.
  - Rank 2: Reduces your target's resistance to all your spells by 15 and reduces the threat caused by your Arcane spells by 30%.
  - Classic (2 ranks):
    - Rank 1: Reduces your target's resistance to all your spells by 5 and reduces the threat caused by your Arcane spells by 20%.
    - Rank 2: Reduces your target's resistance to all your spells by 10 and reduces the threat caused by your Arcane spells by 40%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Magic Absorption** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases all your resistances by 5 and causes all spells you fully resist to restore 1% of your total mana. Cannot trigger more often than 1 time per sec.
  - Rank 2: Increases all your resistances by 10 and causes all spells you fully resist to restore 2% of your total mana. Cannot trigger more often than 1 time per sec.
  - Classic (5 ranks):
    - Rank 1: Increases all resistances by 2 and causes all spells you fully resist to restore 1% of your total mana. 1 sec. cooldown.
    - Rank 2: Increases all resistances by 4 and causes all spells you fully resist to restore 2% of your total mana. 1 sec. cooldown.
    - Rank 3: Increases all resistances by 6 and causes all spells you fully resist to restore 3% of your total mana. 1 sec. cooldown.
    - Rank 4: Increases all resistances by 8 and causes all spells you fully resist to restore 4% of your total mana. 1 sec. cooldown.
    - Rank 5: Increases all resistances by 10 and causes all spells you fully resist to restore 5% of your total mana. 1 sec. cooldown.
- **Arcane Concentration** (5 ranks, column 3) [unchanged]
  - Rank 1: Gives you a 2% chance of entering a Clearcasting state after any damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
  - Rank 2: Gives you a 4% chance of entering a Clearcasting state after any damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
  - Rank 3: Gives you a 6% chance of entering a Clearcasting state after any damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
  - Rank 4: Gives you a 8% chance of entering a Clearcasting state after any damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
  - Rank 5: Gives you a 10% chance of entering a Clearcasting state after any damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
- **Arcane Resilience** (2 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your Armor by an amount equal to 25% of your Intellect.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Classic: Increases your armor by an amount equal to 50% of your Intellect.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Arcane Geometry** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the range of your Arcane spells by 3 yards.
  - Rank 2: Rank 2 values not yet captured from footage.
- **Arcane Impact** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike chance of your Arcane spells by 2%.
  - Rank 2: Increases the critical strike chance of your Arcane spells by 4%.
  - Rank 3: Increases the critical strike chance of your Arcane spells by 6%.
  - Classic (3 ranks):
    - Rank 1: Increases the critical strike chance of your Arcane Explosion spell by an additional 2%.
    - Rank 2: Increases the critical strike chance of your Arcane Explosion spell by an additional 4%.
    - Rank 3: Increases the critical strike chance of your Arcane Explosion spell by an additional 6%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Arcane Blast** (1 rank, column 4) [NEW]
  - Cost: 122 Mana, 30 yd range, 2.5 sec cast
  - Blasts the target with energy, dealing 95 to 104 Arcane damage. Each time you cast Arcane Blast, the damage of all your other spells is increased by 10% and the mana cost of Arcane Blast is increased by 175%. Effect stacks up to 4 times and lasts 8 sec or until any other damage spell is cast.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Arcane Shielding** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Decreases the Mana lost per point of damage taken when your Mana Shield spell is active by 17% and increases the resistances granted by your Mage Armor spell by 25%.
  - Rank 2: Decreases the Mana lost per point of damage taken when your Mana Shield spell is active by 34% and increases the resistances granted by your Mage Armor spell by 50%.
  - Classic (2 ranks):
    - Rank 1: Decreases the mana lost per point of damage taken when Mana Shield is active by 10%.
    - Rank 2: Decreases the mana lost per point of damage taken when Mana Shield is active by 20%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Improved Counterspell** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Your Counterspell also Silences the target for 2 sec.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Classic (2 ranks):
    - Rank 1: Gives your Counterspell a 50% chance to silence the target for 4 sec.
    - Rank 2: Gives your Counterspell a 100% chance to silence the target for 4 sec.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Arcane Meditation** (3 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Requires: Arcane Concentration (5/5)
  - Rank 1: Allows 17% of your Mana regeneration to continue while casting.
  - Rank 2: Allows 34% of your Mana regeneration to continue while casting.
  - Rank 3: Allows 51% of your Mana regeneration to continue while casting.
  - Classic (3 ranks):
    - Rank 1: Allows 5% of your Mana regeneration to continue while casting.
    - Rank 2: Allows 10% of your Mana regeneration to continue while casting.
    - Rank 3: Allows 15% of your Mana regeneration to continue while casting.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Missile Barrage** (1 rank, column 4) [CHANGED]
  - Gives your Arcane Blast spell a 40% chance, and your Fireball, Frostbolt, and Frostfire Bolt spells a 20% chance to reduce the channeled duration of your next Arcane Missiles spell by 50%, reduce the Mana cost by 100%, and missiles will fire every 0.5 sec.
  - Note: Replaced ability_mage_missiles (404 on zamimg) with the WotLK Missile Barrage icon.

#### Row 5 (20 points)

- **Presence of Mind** (1 rank, column 2) [unchanged]
  - Cost: Instant, 3 min cooldown
  - When activated, your next Mage spell with a casting time less than 10 sec becomes an instant cast spell.
- **Arcane Mind** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your Intellect by 2% and increases the critical strike damage bonus of your Arcane spells by 20%.
  - Rank 2: Increases your Intellect by 4% and increases the critical strike damage bonus of your Arcane spells by 40%.
  - Rank 3: Increases your Intellect by 6% and increases the critical strike damage bonus of your Arcane spells by 60%.
  - Rank 4: Increases your Intellect by 8% and increases the critical strike damage bonus of your Arcane spells by 80%.
  - Rank 5: Increases your Intellect by 10% and increases the critical strike damage bonus of your Arcane spells by 100%.
  - Classic (5 ranks):
    - Rank 1: Increases your maximum Mana by 2%.
    - Rank 2: Increases your maximum Mana by 4%.
    - Rank 3: Increases your maximum Mana by 6%.
    - Rank 4: Increases your maximum Mana by 8%.
    - Rank 5: Increases your maximum Mana by 10%.

#### Row 6 (25 points)

- **Arcane Instability** (3 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Requires: Arcane Mind (5/5)
  - Rank 1: Increases the damage done by all your spells by 1% and your critical strike chance with all attacks by 1%.
  - Rank 2: Increases the damage done by all your spells by 2% and your critical strike chance with all attacks by 2%.
  - Rank 3: Increases the damage done by all your spells by 3% and your critical strike chance with all attacks by 3%.
  - Classic (3 ranks):
    - Rank 1: Increases your spell damage and critical strike chance by 1%.
    - Rank 2: Increases your spell damage and critical strike chance by 2%.
    - Rank 3: Increases your spell damage and critical strike chance by 3%.
  - Note: Rank 1 from footage. Rank 2 was still Classic text, rewritten from rank 1. Rank 3 already matched the Forever wording.

#### Row 7 (30 points)

- **Arcane Power** (1 rank, column 2) [CHANGED]
  - Requires: Presence of Mind (1/1)
  - Cost: Instant, 3 min cooldown
  - For the next 15 sec, your spells deal 30% more damage while costing 30% more mana to cast.
  - Classic: When activated, your spells deal 30% more damage while costing 30% more mana to cast. This effect lasts 15 sec.

### Fire

17 talents, 2 new. 31-point talent: Combustion.

#### Row 1 (0 points)

- **Wake of Fire** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Reduces the cooldown of your Fire Blast spell by 1 sec. Killing a non-trivial target increases the critical strike chance of your next Fire Blast cast within 20 sec by 25%.
  - Rank 2: Reduces the cooldown of your Fire Blast spell by 2 sec. Killing a non-trivial target increases the critical strike chance of your next Fire Blast cast within 20 sec by 50%.
  - Note: Rank 2 confirmed on screen: 2 sec and 50%.
- **Incineration** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike chance of your Fire Blast and Scorch spells by 2%.
  - Rank 2: Increases the critical strike chance of your Fire Blast and Scorch spells by 4%.
  - Rank 3: Increases the critical strike chance of your Fire Blast and Scorch spells by 4%.
  - Classic (2 ranks):
    - Rank 1: Increases the critical strike chance of your Fire Blast and Scorch spells by 2%.
    - Rank 2: Increases the critical strike chance of your Fire Blast and Scorch spells by 4%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Fireball** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the casting time of your Fireball spell by 0.1 sec.
  - Rank 2: Reduces the casting time of your Fireball spell by 0.2 sec.
  - Rank 3: Reduces the casting time of your Fireball spell by 0.3 sec.
  - Rank 4: Reduces the casting time of your Fireball spell by 0.4 sec.
  - Rank 5: Reduces the casting time of your Fireball spell by 0.5 sec.

#### Row 2 (5 points)

- **Ignite** (5 ranks, column 1) [unchanged]
  - Rank 1: Your critical strikes from Fire damage spells cause the target to burn for an additional 8% of your spell's damage over 4 sec.
  - Rank 2: Your critical strikes from Fire damage spells cause the target to burn for an additional 16% of your spell's damage over 4 sec.
  - Rank 3: Your critical strikes from Fire damage spells cause the target to burn for an additional 24% of your spell's damage over 4 sec.
  - Rank 4: Your critical strikes from Fire damage spells cause the target to burn for an additional 32% of your spell's damage over 4 sec.
  - Rank 5: Your critical strikes from Fire damage spells cause the target to burn for an additional 40% of your spell's damage over 4 sec.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Flame Throwing** (2 ranks, column 2) [unchanged]
  - Rank 1: Increases the range of your Fire spells by 3 yards.
  - Rank 2: Increases the range of your Fire spells by 6 yards.
  - Note: Placed from the third-party dataset, not yet verified in footage by us.
- **Impact** (3 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives your Fire spells a 2% chance to stun the target for 2 sec.
  - Rank 2: Gives your Fire spells a 4% chance to stun the target for 2 sec.
  - Rank 3: Gives your Fire spells a 6% chance to stun the target for 2 sec.
  - Classic (5 ranks):
    - Rank 1: Gives your Fire spells a 2% chance to stun the target for 2 sec.
    - Rank 2: Gives your Fire spells a 4% chance to stun the target for 2 sec.
    - Rank 3: Gives your Fire spells a 6% chance to stun the target for 2 sec.
    - Rank 4: Gives your Fire spells a 8% chance to stun the target for 2 sec.
    - Rank 5: Gives your Fire spells a 10% chance to stun the target for 2 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Burning Soul** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives your Fire spells a 23% chance to not lose casting time when you take damage and reduces the threat caused by your Fire spells by 10%.
  - Rank 2: Gives your Fire spells a 70% chance to not lose casting time when you take damage and reduces the threat caused by your Fire spells by 30%.
  - Rank 3: Gives your Fire spells a 70% chance to not lose casting time when you take damage and reduces the threat caused by your Fire spells by 30%.
  - Classic (2 ranks):
    - Rank 1: Gives your Fire spells a 35% chance to not lose casting time when you take damage and reduces the threat caused by your Fire spells by 15%.
    - Rank 2: Gives your Fire spells a 70% chance to not lose casting time when you take damage and reduces the threat caused by your Fire spells by 30%.
- **Improved Flamestrike** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike chance of your Flamestrike spell by 5%.
  - Rank 2: Increases the critical strike chance of your Flamestrike spell by 10%.
  - Rank 3: Increases the critical strike chance of your Flamestrike spell by 15%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Pyroblast** (1 rank, column 3) [CHANGED]
  - Cost: 125 Mana, 35 yd range, 6 sec cast
  - Hurls an immense fiery boulder that causes 155 to 185 Fire damage and an additional 76 Fire damage over 12 sec.
  - Classic: Hurls an immense fiery boulder that causes 149 to 195 Fire damage and an additional 56 Fire damage over 12 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Improved Scorch** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Your Scorch spell has a 33% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases all Fire damage you deal to your target by 3% and lasts 30 sec, stacking up to 5 times.
  - Rank 2: Your Scorch spells have a 66% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases the Fire damage dealt to your target by 3% and lasts 30 sec. Stacks up to 5 times.
  - Rank 3: Your Scorch spells have a 100% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases the Fire damage dealt to your target by 3% and lasts 30 sec. Stacks up to 5 times.
  - Classic (3 ranks):
    - Rank 1: Your Scorch spells have a 33% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases the Fire damage dealt to your target by 3% and lasts 30 sec. Stacks up to 5 times.
    - Rank 2: Your Scorch spells have a 66% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases the Fire damage dealt to your target by 3% and lasts 30 sec. Stacks up to 5 times.
    - Rank 3: Your Scorch spells have a 100% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases the Fire damage dealt to your target by 3% and lasts 30 sec. Stacks up to 5 times.
- **Improved Fire Ward** (2 ranks, column 2) [unchanged]
  - Rank 1: Causes your Fire Ward to have a 10% chance to reflect Fire spells while active.
  - Rank 2: Causes your Fire Ward to have a 20% chance to reflect Fire spells while active.
- **Hot Streak** (1 rank, column 3) [NEW]
  - Requires: Pyroblast (1/1)
  - Your non-periodic critical strikes with Fireball, Frostfire Bolt, Fire Blast, and Scorch grant Hot Streak for 15 sec. Hot Streak reduces the cast time of Pyroblast by 25%, stacking up to 3 times.
- **Master of Elements** (3 ranks, column 4) [unchanged]
  - Rank 1: Your Fire and Frost critical strikes will refund 10% of their base mana cost.
  - Rank 2: Your Fire and Frost spell criticals will refund 20% of their base mana cost.
  - Rank 3: Your Fire and Frost spell criticals will refund 30% of their base mana cost.

#### Row 5 (20 points)

- **Critical Mass** (3 ranks, column 2) [unchanged]
  - Rank 1: Increases the critical strike chance of your Fire spells by 2%.
  - Rank 2: Increases the critical strike chance of your Fire spells by 4%.
  - Rank 3: Increases the critical strike chance of your Fire spells by 6%.
- **Blast Wave** (1 rank, column 3) [unchanged]
  - Cost: 215 Mana, Instant, 45 sec cooldown
  - A wave of flame radiates outward from the caster, damaging all enemies caught within the blast for 160 to 191 Fire damage, and Dazing them for 50% reduced movement speed for 6 sec.
  - Note: Placed from the third-party dataset, not yet verified in footage by us.

#### Row 6 (25 points)

- **Fire Power** (5 ranks, column 3) [unchanged]
  - Rank 1: Increases the damage done by your Fire spells by 2%.
  - Rank 2: Increases the damage done by your Fire spells by 4%.
  - Rank 3: Increases the damage done by your Fire spells by 6%.
  - Rank 4: Increases the damage done by your Fire spells by 8%.
  - Rank 5: Increases the damage done by your Fire spells by 10%.

#### Row 7 (30 points)

- **Combustion** (1 rank, column 2) [CHANGED]
  - Requires: Critical Mass (3/3)
  - Cost: Instant, 3 min cooldown
  - When activated, this spell causes each of your Fire damage spell hits to increase your critical strike chance with Fire damage spells by 10%. This effect lasts until you have caused 4 non-periodic critical strikes with Fire spells.
  - Classic: When activated, this spell causes each of your Fire damage spell hits to increase your critical strike chance with Fire damage spells by 10%. This effect lasts until you have caused 3 critical strikes with Fire spells.

### Frost

19 talents, 1 new. 31-point talent: Ice Barrier.

#### Row 1 (0 points)

- **Frost Warding** (2 ranks, column 1) [unchanged]
  - Rank 1: Increases the Armor and resistance given by your Frost Armor and Ice Armor spells by 15%. In addition, gives your Frost Ward a 10% chance to reflect Frost spells and effects while active.
  - Rank 2: Increases the armor and resistances given by your Frost Armor and Ice Armor spells by 30%. In addition, gives your Frost Ward a 20% chance to reflect Frost spells and effects while active.
- **Improved Frostbolt** (5 ranks, column 2) [unchanged]
  - Rank 1: Reduces the casting time of your Frostbolt spell by 0.1 sec.
  - Rank 2: Reduces the casting time of your Frostbolt spell by 0.2 sec.
  - Rank 3: Reduces the casting time of your Frostbolt spell by 0.3 sec.
  - Rank 4: Reduces the casting time of your Frostbolt spell by 0.4 sec.
  - Rank 5: Reduces the casting time of your Frostbolt spell by 0.5 sec.
- **Elemental Precision** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your chance to hit with Frost and Fire spells by 1%.
  - Rank 2: Increases your chance to hit with Frost and Fire spells by 2%.
  - Rank 3: Increases your chance to hit with Frost and Fire spells by 3%.
  - Rank 4: Increases your chance to hit with Frost and Fire spells by 4%.
  - Rank 5: Increases your chance to hit with Frost and Fire spells by 5%.
  - Classic (3 ranks):
    - Rank 1: Reduces the chance that the opponent can resist your Frost and Fire spells by 2%.
    - Rank 2: Reduces the chance that the opponent can resist your Frost and Fire spells by 4%.
    - Rank 3: Reduces the chance that the opponent can resist your Frost and Fire spells by 6%.
  - Note: Rank 1 from footage (hit, not resist). Ranks 2+ were still Classic resist text, rewritten from rank 1 with linear scaling.

#### Row 2 (5 points)

- **Ice Shards** (5 ranks, column 1) [unchanged]
  - Rank 1: Increases the critical strike damage bonus of your Frost spells by 20%.
  - Rank 2: Increases the critical strike damage bonus of your Frost spells by 40%.
  - Rank 3: Increases the critical strike damage bonus of your Frost spells by 60%.
  - Rank 4: Increases the critical strike damage bonus of your Frost spells by 80%.
  - Rank 5: Increases the critical strike damage bonus of your Frost spells by 100%.
- **Permafrost** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the duration of your Chill effects by 11% and reduces the target's speed by an additional 3%.
  - Rank 2: Increases the duration of your Chill effects by 22% and reduces the target's speed by an additional 6%.
  - Rank 3: Increases the duration of your Chill effects by 33% and reduces the target's speed by an additional 9%.
  - Classic (3 ranks):
    - Rank 1: Increases the duration of your Chill effects by 1 sec and reduces the target's speed by an additional 4%.
    - Rank 2: Increases the duration of your Chill effects by 2 secs and reduces the target's speed by an additional 7%.
    - Rank 3: Increases the duration of your Chill effects by 3 secs and reduces the target's speed by an additional 10%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Improved Frost Nova** (2 ranks, column 3) [unchanged]
  - Rank 1: Reduces the cooldown of your Frost Nova spell by 2 sec.
  - Rank 2: Reduces the cooldown of your Frost Nova spell by 4 sec.
- **Frostbite** (3 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives your Chill effects a 5% chance to freeze the target for 5 sec.
  - Rank 2: Gives your Chill effects a 10% chance to freeze the target for 5 sec.
  - Rank 3: Gives your Chill effects a 15% chance to freeze the target for 5 sec.

#### Row 3 (10 points)

- **Piercing Ice** (3 ranks, column 1) [unchanged]
  - Rank 1: Increases the damage done by your Frost spells by 2%.
  - Rank 2: Increases the damage done by your Frost spells by 4%.
  - Rank 3: Increases the damage done by your Frost spells by 6%.
- **Frost Channeling** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the mana cost of your Frost spells by 5% and reduces the threat caused by your Frost spells by 10%.
  - Rank 2: Reduces the mana cost of your Frost spells by 10% and reduces the threat caused by your Frost spells by 20%.
  - Rank 3: Reduces the mana cost of your Frost spells by 15% and reduces the threat caused by your Frost spells by 30%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Ice Lance** (1 rank, column 3) [CHANGED]
  - Requires: Improved Frost Nova (2/2)
  - Cost: 45 Mana, 30 yd range, Instant
  - Deals 28 to 33 Frost damage to an enemy target. Deals 300% increased damage to Frozen targets.
  - Note: Filled the question-mark placeholder with the WotLK Ice Lance icon.
- **Improved Blizzard** (3 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Adds a Chill effect to your Blizzard spell. This effect lowers the target's movement speed by 15% for 1.5 sec.
  - Rank 2: Adds a Chill effect to your Blizzard spell. This effect lowers the target's movement speed by 30% for 1.5 sec.
  - Rank 3: Adds a Chill effect to your Blizzard spell. This effect lowers the target's movement speed by 45% for 1.5 sec.
  - Classic (3 ranks):
    - Rank 1: Adds a chill effect to your Blizzard spell. This effect lowers the target's movement speed by 30%. Lasts 1.5 sec.
    - Rank 2: Adds a chill effect to your Blizzard spell. This effect lowers the target's movement speed by 50%. Lasts 1.5 sec.
    - Rank 3: Adds a chill effect to your Blizzard spell. This effect lowers the target's movement speed by 65%. Lasts 1.5 sec.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 4 (15 points)

- **Arctic Reach** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the range of your Frostbolt and Blizzard spells and the radius of your Frost Nova and Cone of Cold spells by 10%.
  - Rank 2: Increases the range of your Frostbolt and Blizzard spells and the radius of your Frost Nova and Cone of Cold spells by 20%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Ice Block** (1 rank, column 2) [unchanged]
  - Cost: 15 Mana, Instant, 5 min cooldown
  - You become encased in a block of ice, protecting you from all physical attacks and spells for 10 sec, but during that time you cannot attack, move or cast spells.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Shatter** (3 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike chance of all your spells against frozen targets by 10%.
  - Rank 2: Increases the critical strike chance of all your spells against frozen targets by 20%.
  - Rank 3: Increases the critical strike chance of all your spells against frozen targets by 30%.
  - Classic (5 ranks):
    - Rank 1: Increases the critical strike chance of all your spells against frozen targets by 10%.
    - Rank 2: Increases the critical strike chance of all your spells against frozen targets by 20%.
    - Rank 3: Increases the critical strike chance of all your spells against frozen targets by 30%.
    - Rank 4: Increases the critical strike chance of all your spells against frozen targets by 40%.
    - Rank 5: Increases the critical strike chance of all your spells against frozen targets by 50%.

#### Row 5 (20 points)

- **Improved Cone of Cold** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage dealt by your Cone of Cold spell by 12%.
  - Rank 2: Increases the damage dealt by your Cone of Cold spell by 25%.
  - Rank 3: Increases the damage dealt by your Cone of Cold spell by 35%.
  - Classic (3 ranks):
    - Rank 1: Increases the damage dealt by your Cone of Cold spell by 15%.
    - Rank 2: Increases the damage dealt by your Cone of Cold spell by 25%.
    - Rank 3: Increases the damage dealt by your Cone of Cold spell by 35%.
- **Cold Snap** (1 rank, column 2) [CHANGED]
  - Cost: Instant, 10 min cooldown
  - Finishes the remaining cooldown on all your other Frost spells.
  - Classic: When activated, this spell finishes the cooldown on all of your Frost spells.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Fingers of Frost** (2 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Requires: Ice Lance (1/1)
  - Rank 1: Gives your Chill effects a 15% chance to grant you the Fingers of Frost effect, which treats your next 1 spell cast as if the target were Frozen. Lasts 15 sec.
  - Rank 2: Gives your Chill effects a 30% chance to grant you the Fingers of Frost effect, which treats your next 1 spell cast as if the target were Frozen. Lasts 15 sec.

#### Row 6 (25 points)

- **Winter's Chill** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives your Frost damage spells a 20% chance to apply the Winter's Chill effect, which increases the chance your Ice Lance and Frostbolt spells will critically hit the target by 2% for 15 sec. Stacks up to 1 times.
  - Rank 2: Gives your Frost damage spells a 40% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
  - Rank 3: Gives your Frost damage spells a 60% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
  - Rank 4: Gives your Frost damage spells a 80% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
  - Rank 5: Gives your Frost damage spells a 100% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
  - Classic (5 ranks):
    - Rank 1: Gives your Frost damage spells a 20% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
    - Rank 2: Gives your Frost damage spells a 40% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
    - Rank 3: Gives your Frost damage spells a 60% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
    - Rank 4: Gives your Frost damage spells a 80% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
    - Rank 5: Gives your Frost damage spells a 100% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.

#### Row 7 (30 points)

- **Ice Barrier** (1 rank, column 2) [CHANGED]
  - Requires: Cold Snap (1/1)
  - Cost: 305 Mana, Instant, 30 sec cooldown
  - Instantly shields you, absorbing 431 damage. Lasts 1 min. While the shield holds, your spellcasts will not be interrupted or delayed from taking damage.
  - Classic: Instantly shields you, absorbing 455 damage. Lasts 1 min. While the shield holds, spells will not be interrupted.
