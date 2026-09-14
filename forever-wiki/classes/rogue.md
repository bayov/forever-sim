# Rogue

Source: wowforevertalents.com dataset (BlizzCon 2026 footage overlaid on Classic Era 1.15.9 client tables), data version 2026-09-13/14. See [../README.md](../README.md) for caveats. Numbers on new talents were read off level 38 demo characters, so absolute damage and mana values are not level 60 values.

Talents: 53 total, 10 new, 26 changed, 17 unchanged (verified).

Row N needs 5*(N-1) points spent in that tree. Row 1 is the top row. Rows 3, 4, 5 and 7 hold the 11, 16, 21 and 31 point talents.

Warning on multi-rank NEW and CHANGED talents: the footage showed rank 1 only. The dataset scaled the other ranks linearly from rank 1, which sometimes produces nonsense (for example Maelstrom Weapon rank 5 reading 'stacks up to 25 times'). Treat rank 1 as observed and the rest as a guess until the beta client is out. The [ranks 2+ unverified] tag marks every multi-rank new or changed talent. Read the Note line, it says when a higher rank was actually seen on a BlizzCon slide.

## Spellbook changes (trainer abilities)

0 of 33 trainer abilities confirmed from footage. Abilities not listed below are either verified unchanged or still Classic placeholders (unverified).

### Not yet verified (Classic text assumed)

Sinister Strike, Backstab, Gouge, Evasion, Sprint, Kick, Feint, Pick Lock, Stealth, Pick Pocket, Sap, Distract, Vanish, Detect Traps, Disarm Trap, Blind, Safe Fall, Hemorrhage, Crippling Poison, Instant Poison, Poisons, Mind-numbing Poison, Deadly Poison, Wound Poison, Blinding Powder, Eviscerate, Slice and Dice, Expose Armor, Garrote, Ambush, Rupture, Cheap Shot, Kidney Shot

## Talents

### Assassination

17 talents, 2 new. 31-point talent: Venom.

#### Row 1 (0 points)

- **Improved Gouge** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the duration of your Gouge ability by 0.5 sec.
  - Rank 2: Increases the effect duration of your Gouge ability by 1 sec.
  - Rank 3: Increases the effect duration of your Gouge ability by 1.5 sec.
  - Classic (3 ranks):
    - Rank 1: Increases the effect duration of your Gouge ability by 0.5 sec.
    - Rank 2: Increases the effect duration of your Gouge ability by 1 sec.
    - Rank 3: Increases the effect duration of your Gouge ability by 1.5 sec.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Remorseless Attacks** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: After killing a non-trivial enemy, gives you a 20% increased critical strike chance on your next Sinister Strike, Backstab, Ambush, Mutilate, or Ghostly Strike. Lasts 20 sec.
  - Rank 2: After killing a non-trivial enemy, gives you a 40% increased critical strike chance on your next Sinister Strike, Backstab, Ambush, Mutilate, or Ghostly Strike. Lasts 20 sec.
  - Classic (2 ranks):
    - Rank 1: After killing an opponent that yields experience or honor, gives you a 20% increased critical strike chance on your next Sinister Strike, Backstab, Ambush, or Ghostly Strike. Lasts 20 sec.
    - Rank 2: After killing an opponent that yields experience or honor, gives you a 40% increased critical strike chance on your next Sinister Strike, Backstab, Ambush, or Ghostly Strike. Lasts 20 sec.
  - Note: Now also triggers on Mutilate. Rank 2 extrapolated.
- **Malice** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your critical strike chance with all attacks and Poisons by 1%.
  - Rank 2: Increases your critical strike chance with all attacks and Poisons by 2%.
  - Rank 3: Increases your critical strike chance with all attacks and Poisons by 3%.
  - Rank 4: Increases your critical strike chance with all attacks and Poisons by 4%.
  - Rank 5: Increases your critical strike chance with all attacks and Poisons by 5%.
  - Classic (5 ranks):
    - Rank 1: Increases your critical strike chance by 1%.
    - Rank 2: Increases your critical strike chance by 2%.
    - Rank 3: Increases your critical strike chance by 3%.
    - Rank 4: Increases your critical strike chance by 4%.
    - Rank 5: Increases your critical strike chance by 5%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 2 (5 points)

- **Ruthlessness** (3 ranks, column 1) [unchanged]
  - Rank 1: Gives your finishing moves a 20% chance to add a Combo Point to your target.
  - Rank 2: Gives your finishing moves a 40% chance to add a combo point to your target.
  - Rank 3: Gives your finishing moves a 60% chance to add a combo point to your target.
- **Murder** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases all damage dealt by 2% against Humanoid and Giant targets.
  - Rank 2: Increases all damage dealt by 4% against Humanoid and Giant targets.
  - Classic (2 ranks):
    - Rank 1: Increases all damage caused against Humanoid, Giant, Beast and Dragonkin targets by 1%.
    - Rank 2: Increases all damage caused against Humanoid, Giant, Beast and Dragonkin targets by 2%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Improved Slice and Dice** (3 ranks, column 4) [unchanged]
  - Rank 1: Increases the duration of your Slice and Dice ability by 15%.
  - Rank 2: Increases the duration of your Slice and Dice ability by 30%.
  - Rank 3: Increases the duration of your Slice and Dice ability by 45%.

#### Row 3 (10 points)

- **Relentless Strikes** (1 rank, column 1) [unchanged]
  - Your finishing moves have a 20% chance per Combo Point to restore 25 Energy.
- **Improved Expose Armor** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the Energy cost of your Expose Armor ability by 5, and refunds 1 Combo Point when cast with 5 Combo Points.
  - Rank 2: Reduces the Energy cost of your Expose Armor ability by 10, and refunds 2 Combo Points when cast with 5 Combo Points.
  - Classic (2 ranks):
    - Rank 1: Increases the armor reduced by your Expose Armor ability by 25%.
    - Rank 2: Increases the armor reduced by your Expose Armor ability by 50%.
  - Note: Reworked: Energy cost and Combo Point refund instead of extra armor reduction. Rank 2 extrapolated.
- **Lethality** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Requires: Malice (5/5)
  - Rank 1: Increases the critical strike damage bonus of your Sinister Strike, Gouge, Backstab, Mutilate, Ghostly Strike, and Hemorrhage abilities by 6%.
  - Rank 2: Increases the critical strike damage bonus of your Sinister Strike, Gouge, Backstab, Mutilate, Ghostly Strike, and Hemorrhage abilities by 12%.
  - Rank 3: Increases the critical strike damage bonus of your Sinister Strike, Gouge, Backstab, Mutilate, Ghostly Strike, and Hemorrhage abilities by 18%.
  - Rank 4: Increases the critical strike damage bonus of your Sinister Strike, Gouge, Backstab, Mutilate, Ghostly Strike, and Hemorrhage abilities by 24%.
  - Rank 5: Increases the critical strike damage bonus of your Sinister Strike, Gouge, Backstab, Mutilate, Ghostly Strike, and Hemorrhage abilities by 30%.
  - Classic (5 ranks):
    - Rank 1: Increases the critical strike damage bonus of your Sinister Strike, Gouge, Backstab, Ghostly Strike, and Hemorrhage abilities by 6%.
    - Rank 2: Increases the critical strike damage bonus of your Sinister Strike, Gouge, Backstab, Ghostly Strike, and Hemorrhage abilities by 12%.
    - Rank 3: Increases the critical strike damage bonus of your Sinister Strike, Gouge, Backstab, Ghostly Strike, and Hemorrhage abilities by 18%.
    - Rank 4: Increases the critical strike damage bonus of your Sinister Strike, Gouge, Backstab, Ghostly Strike, and Hemorrhage abilities by 24%.
    - Rank 5: Increases the critical strike damage bonus of your Sinister Strike, Gouge, Backstab, Ghostly Strike, and Hemorrhage abilities by 30%.
  - Note: Now also covers Mutilate and Hemorrhage. Requires Malice.

#### Row 4 (15 points)

- **Vile Poisons** (5 ranks, column 1) [unchanged]
  - Rank 1: Increases the damage dealt by your poisons by 4% and gives your poisons an additional 8% chance to resist dispel effects.
  - Rank 2: Increases the damage dealt by your poisons by 8% and gives your poisons an additional 16% chance to resist dispel effects.
  - Rank 3: Increases the damage dealt by your poisons by 12% and gives your poisons an additional 24% chance to resist dispel effects.
  - Rank 4: Increases the damage dealt by your poisons by 16% and gives your poisons an additional 32% chance to resist dispel effects.
  - Rank 5: Increases the damage dealt by your poisons by 20% and gives your poisons an additional 40% chance to resist dispel effects.
- **Cold Blood** (1 rank, column 2) [unchanged]
  - Cost: Instant, 3 min cooldown
  - When activated, increases the critical strike chance of your next Sinister Strike, Backstab, Ambush, Eviscerate, or Mutilate by 100%.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Improved Poisons** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the chance to apply Poisons to your target by 2%, and gives Poison applications a 10% chance to not consume a charge.
  - Rank 2: Increases the chance to apply Poisons to your target by 4%, and gives Poison applications a 20% chance to not consume a charge.
  - Rank 3: Increases the chance to apply Poisons to your target by 6%, and gives Poison applications a 30% chance to not consume a charge.
  - Rank 4: Increases the chance to apply Poisons to your target by 8%, and gives Poison applications a 40% chance to not consume a charge.
  - Rank 5: Increases the chance to apply Poisons to your target by 10%, and gives Poison applications a 50% chance to not consume a charge.
  - Classic (5 ranks):
    - Rank 1: Increases the chance to apply poisons to your target by 2%.
    - Rank 2: Increases the chance to apply poisons to your target by 4%.
    - Rank 3: Increases the chance to apply poisons to your target by 6%.
    - Rank 4: Increases the chance to apply poisons to your target by 8%.
    - Rank 5: Increases the chance to apply poisons to your target by 10%.
  - Note: Now also lets Poison applications skip consuming a charge. Ranks 2 to 5 extrapolated.

#### Row 5 (20 points)

- **Vigor** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your maximum Energy by 5.
  - Rank 2: Increases your maximum Energy by 10.
  - Classic: Increases your maximum Energy by 10.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Mutilate** (1 rank, column 2) [NEW]
  - Cost: 60 Energy, Melee Range, Instant
  - Instantly attacks with both weapons for 75% weapon damage plus an additional 13 with each weapon. Damage increased by 20% against Poisoned targets. Awards 2 Combo Points.
  - Note: 60 Energy, melee range, instant.
- **Improved Kidney Shot** (2 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Enemies Stunned by your Kidney Shot ability take 5% increased damage from your poisons and attacks.
  - Rank 2: Enemies Stunned by your Kidney Shot ability take 10% increased damage from your poisons and attacks.
  - Classic (3 ranks):
    - Rank 1: While affected by your Kidney Shot ability, the target receives an additional 3% damage from all sources.
    - Rank 2: While affected by your Kidney Shot ability, the target receives an additional 6% damage from all sources.
    - Rank 3: While affected by your Kidney Shot ability, the target receives an additional 9% damage from all sources.
  - Note: Two ranks now, rank 2 extrapolated.

#### Row 6 (25 points)

- **Seal Fate** (5 ranks, column 3) [unchanged]
  - Rank 1: Your critical strikes from abilities that add combo points  have a 20% chance to add an additional combo point.
  - Rank 2: Your critical strikes from abilities that add combo points  have a 40% chance to add an additional combo point.
  - Rank 3: Your critical strikes from abilities that add combo points  have a 60% chance to add an additional combo point.
  - Rank 4: Your critical strikes from abilities that add combo points  have a 80% chance to add an additional combo point.
  - Rank 5: Your critical strikes from abilities that add combo points  have a 100% chance to add an additional combo point.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.

#### Row 7 (30 points)

- **Venom** (1 rank, column 2) [NEW]
  - Requires: Mutilate (1/1)
  - Cost: 25 Energy, 1 to 5 Combo Points, Instant
  - Finishing move that increases the damage of your Poisons by 30% and your chance to apply Poisons by 10%. Lasts longer per combo point:
   1 point: 9 seconds
   2 points: 12 seconds
   3 points: 15 seconds
   4 points: 18 seconds
   5 points: 21 seconds
  - Note: 25 Energy, 1 to 5 Combo Points, instant. Requires Mutilate.

### Combat

17 talents, 3 new. 31-point talent: Adrenaline Rush.

#### Row 1 (0 points)

- **Improved Eviscerate** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage done by your Eviscerate ability by 7%.
  - Rank 2: Increases the damage done by your Eviscerate ability by 10%.
  - Rank 3: Increases the damage done by your Eviscerate ability by 15%.
  - Classic (3 ranks):
    - Rank 1: Increases the damage done by your Eviscerate ability by 5%.
    - Rank 2: Increases the damage done by your Eviscerate ability by 10%.
    - Rank 3: Increases the damage done by your Eviscerate ability by 15%.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Improved Sinister Strike** (2 ranks, column 2) [unchanged]
  - Rank 1: Reduces the Energy cost of your Sinister Strike ability by 3.
  - Rank 2: Reduces the Energy cost of your Sinister Strike ability by 5.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Lightning Reflexes** (5 ranks, column 3) [unchanged]
  - Rank 1: Increases your Dodge chance by 1%.
  - Rank 2: Increases your Dodge chance by 2%.
  - Rank 3: Increases your Dodge chance by 3%.
  - Rank 4: Increases your Dodge chance by 4%.
  - Rank 5: Increases your Dodge chance by 5%.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.

#### Row 2 (5 points)

- **Puncturing Wounds** (3 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike chance of your Backstab by 10% and your Mutilate by 5%, and gives Backstab a 15% chance to add an additional Combo Point.
  - Rank 2: Increases the critical strike chance of your Backstab by 20% and your Mutilate by 10%, and gives Backstab a 30% chance to add an additional Combo Point.
  - Rank 3: Increases the critical strike chance of your Backstab by 30% and your Mutilate by 15%, and gives Backstab a 45% chance to add an additional Combo Point.
  - Note: Replaces Improved Backstab. Ranks 2 and 3 extrapolated.
- **Deflection** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Requires: Improved Sinister Strike (2/2)
  - Rank 1: Increases your Parry chance by 2%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Increases your Parry chance by 3%.
  - Classic (5 ranks):
    - Rank 1: Increases your Parry chance by 1%.
    - Rank 2: Increases your Parry chance by 2%.
    - Rank 3: Increases your Parry chance by 3%.
    - Rank 4: Increases your Parry chance by 4%.
    - Rank 5: Increases your Parry chance by 5%.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Precision** (3 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your chance to hit with all attacks and Poisons by 1%.
  - Rank 2: Increases your chance to hit with all attacks and Poisons by 2%.
  - Rank 3: Increases your chance to hit with all attacks and Poisons by 3%.
  - Classic (5 ranks):
    - Rank 1: Increases your chance to hit with melee weapons by 1%.
    - Rank 2: Increases your chance to hit with melee weapons by 2%.
    - Rank 3: Increases your chance to hit with melee weapons by 3%.
    - Rank 4: Increases your chance to hit with melee weapons by 4%.
    - Rank 5: Increases your chance to hit with melee weapons by 5%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 3 (10 points)

- **Endurance** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the cooldown of your Sprint and Evasion abilities by 30%.
  - Rank 2: Reduces the cooldown of your Sprint and Evasion abilities by 60%.
  - Classic (2 ranks):
    - Rank 1: Reduces the cooldown of your Sprint and Evasion abilities by 45 sec.
    - Rank 2: Reduces the cooldown of your Sprint and Evasion abilities by 1.5 min.
  - Note: Percentage cooldown reduction now. Rank 2 extrapolated.
- **Riposte** (1 rank, column 2) [unchanged]
  - Requires: Deflection (3/3)
  - Cost: 10 Energy, Melee Range, Instant, 6 sec cooldown
  - A strike that becomes active after parrying an opponent's attack.  This attack deals 150% weapon damage and disarms the target for 6 sec.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Improved Sprint** (2 ranks, column 4) [unchanged]
  - Rank 1: Gives a 50% chance to remove all movement impairing effects when you activate your Sprint ability.
  - Rank 2: Gives a 100% chance to remove all movement impairing effects when you activate your Sprint ability.

#### Row 4 (15 points)

- **Improved Kick** (2 ranks, column 1) [unchanged]
  - Rank 1: Gives your Kick ability a 50% chance to silence the target for 2 sec.
  - Rank 2: Gives your Kick ability a 100% chance to silence the target for 2 sec.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Restless Blades** (1 rank, column 2) [NEW]
  - Your damaging finishing moves reduce the remaining cooldown of your Adrenaline Rush, Blade Flurry, Evasion, Sprint, and Vanish abilities by 2 sec per combo point.
  - Note: Placed from the third-party dataset, not yet verified in footage by us.
- **Dual Wield Specialization** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Requires: Precision (5/3)
  - Rank 1: Increases the damage done by your off-hand weapon by 5%.
  - Rank 2: Increases the damage done by your off-hand weapon by 10%.
  - Rank 3: Increases the damage done by your off-hand weapon by 15%.
  - Rank 4: Increases the damage done by your off-hand weapon by 20%.
  - Rank 5: Increases the damage done by your off-hand weapon by 25%.
  - Classic (5 ranks):
    - Rank 1: Increases the damage done by your offhand weapon by 10%.
    - Rank 2: Increases the damage done by your offhand weapon by 20%.
    - Rank 3: Increases the damage done by your offhand weapon by 30%.
    - Rank 4: Increases the damage done by your offhand weapon by 40%.
    - Rank 5: Increases the damage done by your offhand weapon by 50%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 5 (20 points)

- **Blade Flurry** (1 rank, column 2) [unchanged]
  - Requires: Restless Blades (1/1)
  - Cost: 25 Energy, Instant, 2 min cooldown
  - Increases your melee attack speed by 20% and your melee attacks strike an additional nearby opponent. Lasts 15 sec.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Hack and Slash** (5 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Gives your melee weapon attacks a benefit depending on the weapon. Axe/Sword: Your successful melee attacks have a 1% chance to trigger an extra attack on the target. Dagger/Fist: Increases your critical strike chance by 1%. Mace: Your attacks ignore 3% of your target's armor.
  - Rank 2: Gives your melee weapon attacks a benefit depending on the weapon. Axe/Sword: Your successful melee attacks have a 2% chance to trigger an extra attack on the target. Dagger/Fist: Increases your critical strike chance by 2%. Mace: Your attacks ignore 6% of your target's armor.
  - Rank 3: Gives your melee weapon attacks a benefit depending on the weapon. Axe/Sword: Your successful melee attacks have a 3% chance to trigger an extra attack on the target. Dagger/Fist: Increases your critical strike chance by 3%. Mace: Your attacks ignore 9% of your target's armor.
  - Rank 4: Gives your melee weapon attacks a benefit depending on the weapon. Axe/Sword: Your successful melee attacks have a 4% chance to trigger an extra attack on the target. Dagger/Fist: Increases your critical strike chance by 4%. Mace: Your attacks ignore 12% of your target's armor.
  - Rank 5: Gives your melee weapon attacks a benefit depending on the weapon. Axe/Sword: Your successful melee attacks have a 5% chance to trigger an extra attack on the target. Dagger/Fist: Increases your critical strike chance by 5%. Mace: Your attacks ignore 15% of your target's armor.
  - Note: Replaces the weapon specialization talents. Ranks 2 to 5 extrapolated.

#### Row 6 (25 points)

- **Weapon Expertise** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Requires: Blade Flurry (1/1)
  - Rank 1: Reduces the chance for your attacks to be Dodged or Parried by 1%.
  - Rank 2: Reduces the chance for your attacks to be Dodged or Parried by 2%.
  - Classic (2 ranks):
    - Rank 1: Increases your skill with Sword, Fist and Dagger weapons by 3.
    - Rank 2: Increases your skill with Sword, Fist and Dagger weapons by 5.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Aggression** (3 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Requires: Hack and Slash (5/5)
  - Rank 1: Increases the damage of your Sinister Strike, Backstab, and Eviscerate abilities by 2%.
  - Rank 2: Increases the damage of your Sinister Strike, Backstab, and Eviscerate abilities by 4%.
  - Rank 3: Increases the damage of your Sinister Strike, Backstab, and Eviscerate abilities by 6%.
  - Classic (3 ranks):
    - Rank 1: Increases the damage of your Sinister Strike and Eviscerate abilities by 2%.
    - Rank 2: Increases the damage of your Sinister Strike and Eviscerate abilities by 4%.
    - Rank 3: Increases the damage of your Sinister Strike and Eviscerate abilities by 6%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 7 (30 points)

- **Adrenaline Rush** (1 rank, column 2) [unchanged]
  - Cost: Instant, 5 min cooldown
  - Increases your Energy regeneration rate by 100% for 15 sec.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.

### Subtlety

19 talents, 5 new. 31-point talent: Thousand Cuts.

#### Row 1 (0 points)

- **Camouflage** (5 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces your speed penalty from your Stealth ability by 3% and reduces its cooldown by 2 sec.
  - Rank 2: Reduces your speed penalty from your Stealth ability by 6% and reduces its cooldown by 4 sec.
  - Rank 3: Reduces your speed penalty from your Stealth ability by 9% and reduces its cooldown by 6 sec.
  - Rank 4: Reduces your speed penalty from your Stealth ability by 12% and reduces its cooldown by 8 sec.
  - Rank 5: Reduces your speed penalty from your Stealth ability by 15% and reduces its cooldown by 10 sec.
  - Classic (5 ranks):
    - Rank 1: Increases your speed while stealthed by 3% and reduces the cooldown of your Stealth ability by 1 sec.
    - Rank 2: Increases your speed while stealthed by 6% and reduces the cooldown of your Stealth ability by 2 sec.
    - Rank 3: Increases your speed while stealthed by 9% and reduces the cooldown of your Stealth ability by 3 sec.
    - Rank 4: Increases your speed while stealthed by 12% and reduces the cooldown of your Stealth ability by 4 sec.
    - Rank 5: Increases your speed while stealthed by 15% and reduces the cooldown of your Stealth ability by 5 sec.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Master of Deception** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the chance enemies have to detect you while in Stealth mode as if you were 1 level higher.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Classic (5 ranks):
    - Rank 1: Reduces the chance enemies have to detect you while in Stealth mode.
    - Rank 2: Reduces the chance enemies have to detect you while in Stealth mode. More effective than Master of Deception (Rank 1).
    - Rank 3: Reduces the chance enemies have to detect you while in Stealth mode. More effective than Master of Deception (Rank 2).
    - Rank 4: Reduces the chance enemies have to detect you while in Stealth mode. More effective than Master of Deception (Rank 3).
    - Rank 5: Reduces the chance enemies have to detect you while in Stealth mode. More effective than Master of Deception (Rank 4).
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Opportunity** (2 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage dealt by your Backstab, Garrote, Ambush, and Mutilate abilities by 5%.
  - Rank 2: Increases the damage dealt by your Backstab, Garrote, Ambush, and Mutilate abilities by 10%.
  - Classic (5 ranks):
    - Rank 1: Increases the damage dealt when striking from behind with your Backstab, Garrote, or Ambush abilities by 4%.
    - Rank 2: Increases the damage dealt when striking from behind with your Backstab, Garrote, or Ambush abilities by 8%.
    - Rank 3: Increases the damage dealt when striking from behind with your Backstab, Garrote, or Ambush abilities by 12%.
    - Rank 4: Increases the damage dealt when striking from behind with your Backstab, Garrote, or Ambush abilities by 16%.
    - Rank 5: Increases the damage dealt when striking from behind with your Backstab, Garrote, or Ambush abilities by 20%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 2 (5 points)

- **Setup** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives you a 33% chance to add a Combo Point to your target after Dodging an attack or fully resisting a spell.
  - Rank 2: Gives you a 66% chance to add a Combo Point to your target after Dodging an attack or fully resisting a spell.
  - Rank 3: Gives you a 100% chance to add a Combo Point to your target after Dodging an attack or fully resisting a spell.
  - Classic (3 ranks):
    - Rank 1: Gives you a 15% chance to add a combo point to your target after dodging their attack or fully resisting one of their spells.
    - Rank 2: Gives you a 30% chance to add a combo point to your target after dodging their attack or fully resisting one of their spells.
    - Rank 3: Gives you a 45% chance to add a combo point to your target after dodging their attack or fully resisting one of their spells.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Elusiveness** (2 ranks, column 2) [unchanged]
  - Rank 1: Reduces the cooldown of your Vanish and Blind abilities by 45 sec.
  - Rank 2: Reduces the cooldown of your Vanish and Blind abilities by 1.5 min.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Dirty Tricks** (2 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Reduces the Energy cost of your Sap and Blind abilities by 25%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Note: Placed from the third-party dataset (rank 1 text only), not yet verified in footage by us.
- **Improved Ambush** (3 ranks, column 4) [unchanged]
  - Rank 1: Increases the critical strike chance of your Ambush ability by 15%.
  - Rank 2: Increases the critical strike chance of your Ambush ability by 30%.
  - Rank 3: Increases the critical strike chance of your Ambush ability by 45%.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Initiative** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives you a 33% chance to add an additional combo point to your target when using your Ambush, Garrote, or Cheap Shot ability.
  - Rank 2: Gives you a 66% chance to add an additional combo point to your target when using your Ambush, Garrote, or Cheap Shot ability.
  - Rank 3: Gives you a 100% chance to add an additional combo point to your target when using your Ambush, Garrote, or Cheap Shot ability.
  - Classic (3 ranks):
    - Rank 1: Gives you a 25% chance to add an additional combo point to your target when using your Ambush, Garrote, or Cheap Shot ability.
    - Rank 2: Gives you a 50% chance to add an additional combo point to your target when using your Ambush, Garrote, or Cheap Shot ability.
    - Rank 3: Gives you a 75% chance to add an additional combo point to your target when using your Ambush, Garrote, or Cheap Shot ability.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Ghostly Strike** (1 rank, column 2) [CHANGED]
  - Cost: 40 Energy, Melee Range, Instant, 20 sec cooldown
  - A strike that deals 125% (180% if a Dagger is equipped in your Main Hand) weapon damage and increases your chance to dodge by 15% for 7 sec. Awards 1 combo point.
  - Classic: A strike that deals 125% weapon damage and increases your chance to dodge by 15% for 7 sec. Awards 1 combo point.
  - Note: 40 Energy, melee range, instant, 20 sec cooldown. Dagger bonus is new.
- **Improved Distract** (2 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the radius of your Distract ability by 3 yds, and further reduces the Stealth detection of distracted enemies as though they were an additional 1 level lower.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Note: Placed from the third-party dataset (rank 1 text only), not yet verified in footage by us.

#### Row 4 (15 points)

- **Heightened Senses** (2 ranks, column 1) [unchanged]
  - Rank 1: Increases your Stealth detection as if you were 1 level higher and reduces your chance to be hit by spells and ranged attacks by 2%.
  - Rank 2: Increases your Stealth detection as if you were 2 levels higher and reduces your chance to be hit by spells and ranged attacks by 4%.
  - Note: Moved up one row.
- **Premeditation** (1 rank, column 2) [CHANGED]
  - Cost: 20 yd range, Instant, 2 min cooldown
  - Adds 2 Combo Points to your target. You must add to or use those combo points within 20 sec or the combo points are lost.
  - Classic: When used, adds 2 combo points to your target. You must add to or use those combo points within 10 sec or the combo points are lost.
  - Note: 20 yd range, instant, 2 min cooldown. No longer the capstone, combo points now last 20 sec.
- **Serrated Blades** (3 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Requires: Improved Distract (2/2)
  - Rank 1: Causes your attacks to ignore 3% of your target's Armor and increases the damage dealt by your Rupture ability by 10%.
  - Rank 2: Causes your attacks to ignore 6% of your target's Armor and increases the damage dealt by your Rupture ability by 20%.
  - Rank 3: Causes your attacks to ignore 9% of your target's Armor and increases the damage dealt by your Rupture ability by 30%.
  - Classic (3 ranks):
    - Rank 1: Causes your attacks to ignore 100 of your target's Armor and increases the damage dealt by your Rupture ability by 10%. The amount of Armor reduced increases with your level.
    - Rank 2: Causes your attacks to ignore 200 of your target's Armor and increases the damage dealt by your Rupture ability by 20%. The amount of Armor reduced increases with your level.
    - Rank 3: Causes your attacks to ignore 300 of your target's Armor and increases the damage dealt by your Rupture ability by 30%. The amount of Armor reduced increases with your level.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 5 (20 points)

- **Dirty Deeds** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Requires: Heightened Senses (2/2)
  - Rank 1: Reduces the Energy cost of your Cheap Shot and Garrote abilities by 10, and your Garrote ability no longer requires you to be behind your target.
  - Rank 2: Reduces the Energy cost of your Cheap Shot and Garrote abilities by 20, and your Garrote ability no longer requires you to be behind your target.
  - Classic (2 ranks):
    - Rank 1: Reduces the Energy cost of your Cheap Shot and Garrote abilities by 10.
    - Rank 2: Reduces the Energy cost of your Cheap Shot and Garrote abilities by 20.
  - Note: Garrote no longer needs you behind the target. Rank 2 extrapolated.
- **Preparation** (1 rank, column 2) [unchanged]
  - Requires: Premeditation (1/1)
  - Cost: Instant, 10 min cooldown
  - When activated, this ability immediately finishes the cooldown on your other Rogue abilities.
  - Note: Instant, 10 min cooldown.
- **Hemorrhage** (1 rank, column 3) [CHANGED]
  - Requires: Serrated Blades (3/3)
  - Cost: 35 Energy, Melee Range, Instant
  - An instant strike that deals 100% weapon damage (145% if a Dagger is equipped) and causes the target to take 15% increased Rupture damage from the Rogue. Lasts 15 sec. Awards 1 Combo Point.
  - Classic: An instant strike that damages the opponent and causes the target to hemorrhage, increasing any Physical damage dealt to the target by up to 3. Lasts 30 charges or 15 sec. Awards 1 combo point.
  - Note: 35 Energy, melee range, instant. Reworked around Rupture.

#### Row 6 (25 points)

- **Quietus** (5 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Requires: Dirty Deeds (2/2)
  - Rank 1: Your Sinister Strike, Ghostly Strike, and Hemorrhage abilities cause 2% more damage against targets below 35% health.
  - Rank 2: Your Sinister Strike, Ghostly Strike, and Hemorrhage abilities cause 4% more damage against targets below 35% health.
  - Rank 3: Your Sinister Strike, Ghostly Strike, and Hemorrhage abilities cause 6% more damage against targets below 35% health.
  - Rank 4: Your Sinister Strike, Ghostly Strike, and Hemorrhage abilities cause 8% more damage against targets below 35% health.
  - Rank 5: Your Sinister Strike, Ghostly Strike, and Hemorrhage abilities cause 10% more damage against targets below 35% health.
  - Note: Footage shows an arrow from Dirty Deeds into Quietus.
- **Cutthroat** (5 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Requires: Hemorrhage (1/1)
  - Rank 1: Your Backstab has a 3% chance to cause your next Ambush within 10 sec to not require Stealth.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Rank 4: Rank 4 values not yet captured from footage.
  - Rank 5: Rank 5 values not yet captured from footage.
  - Note: Placed from the third-party dataset (rank 1 text only), not yet verified in footage by us.

#### Row 7 (30 points)

- **Thousand Cuts** (1 rank, column 2) [NEW]
  - Requires: Preparation (1/1)
  - When your Rupture ability deals periodic damage, the Energy cost of your next Hemorrhage or Backstab ability within 10 sec is reduced by 3, stacking up to 5 times.
