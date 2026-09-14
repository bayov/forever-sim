# Hunter

Source: wowforevertalents.com dataset (BlizzCon 2026 footage overlaid on Classic Era 1.15.9 client tables), data version 2026-09-13/14. See [../README.md](../README.md) for caveats. Numbers on new talents were read off level 38 demo characters, so absolute damage and mana values are not level 60 values.

Talents: 50 total, 15 new, 20 changed, 15 unchanged (verified).

Row N needs 5*(N-1) points spent in that tree. Row 1 is the top row. Rows 3, 4, 5 and 7 hold the 11, 16, 21 and 31 point talents.

Warning on multi-rank NEW and CHANGED talents: the footage showed rank 1 only. The dataset scaled the other ranks linearly from rank 1, which sometimes produces nonsense (for example Maelstrom Weapon rank 5 reading 'stacks up to 25 times'). Treat rank 1 as observed and the rest as a guess until the beta client is out. The [ranks 2+ unverified] tag marks every multi-rank new or changed talent. Read the Note line, it says when a higher rank was actually seen on a BlizzCon slide.

## Spellbook changes (trainer abilities)

45 of 63 trainer abilities confirmed from footage. Abilities not listed below are either verified unchanged or still Classic placeholders (unverified).

### New abilities

- **Assist** (pet)
  - Your pet will assist you.
  - Note: Pet stance that replaces Classic's Aggressive stance.
- **Mine!** (pet, 20 Focus, Melee Range, Instant, 1 min cooldown)
  - Grabs an enemy's weapon with its talons, causing 20 to 24 damage and disarming them for 4 sec.
  - Note: Bird of Prey ability, Rank 3 at level 38. A disarm, like the later Snatch.
- **Move To** (pet)
  - Orders your pet to move to a target location.
  - Note: New pet command.

### Changed abilities

- **Eyes of the Beast** (beast-mastery, Level 14, 20 Mana, Unlimited range, 2 sec cast)
  - Take direct control of your pet and see through its eyes for 2 min.
  - Classic: Take direct control of your pet and see through its eyes for 1 min.
  - Note: Lasts 2 min instead of 1 min.
- **Aspect of the Beast** (beast-mastery, Level 30, 50 Mana, Instant)
  - The hunter takes on the aspects of a beast, becoming untrackable and increasing Melee Attack Power by 50. Only one Aspect can be active at a time.
  - Classic: The hunter takes on the aspects of a beast, becoming untrackable. Only one Aspect can be active at a time.
  - Note: Now ranked (Rank 1 at level 38) and adds melee attack power on top of being untrackable.
- **Raptor Strike** (survival, Level 32, 55 Mana, Melee Range, Next melee, 6 sec cooldown)
  - A strong attack that deals melee weapon damage plus 35.
  - Classic: A strong attack that increases melee damage by 140.
  - Note: Now worded as weapon damage plus a bonus (rank 5: plus 35, where Classic rank 5 added 50).
- **Immolation Trap** (survival, Level 36, 135 Mana, Instant, 30 sec cooldown)
  - Place a Fire trap that will burn the first enemy to approach for 340 Fire damage over 15 sec. Trap will exist for 1 min. Only one Fire trap can be active at a time.
  - Classic: Place a fire trap that will burn the first enemy to approach for 690 Fire damage over 15 sec. Trap will exist for 1 min. Traps can only be placed when out of combat. Only one trap can be active at a time.
  - Note: Traps can now be placed in combat (the Classic out-of-combat clause is gone), one Fire trap and one Frost trap can be active at the same time, and the cooldown is 30 sec instead of 15.
- **Mongoose Bite** (survival, Level 30, 40 Mana, Melee Range, Instant, 5 sec cooldown)
  - Counterattack the enemy for melee weapon damage plus 22. Can only be performed after you dodge.
  - Classic: Counterattack the enemy for 115 damage. Can only be performed after you dodge.
  - Note: Now scales with weapon damage (Classic rank 2 was a flat 45 damage).
- **Freezing Trap** (survival, Level 20, 50 Mana, Instant, 30 sec cooldown)
  - Place a Frost trap that freezes the first enemy that approaches, preventing all action for up to 10 sec. Any damage caused will break the ice. Trap will exist for 1 min. Only one Frost trap can be active at a time.
  - Classic: Place a frost trap that freezes the first enemy that approaches, preventing all action for up to 20 sec. Any damage caused will break the ice. Trap will exist for 1 min. Traps can only be placed when out of combat. Only one trap can be active at a time.
  - Note: Traps can now be placed in combat (the Classic out-of-combat clause is gone), one Fire trap and one Frost trap can be active at the same time, and the cooldown is 30 sec instead of 15.
- **Frost Trap** (survival, Level 28, 60 Mana, Instant, 30 sec cooldown)
  - Place a Frost trap that creates an ice slick around itself for 30 sec when the first enemy approaches it. All enemies within 10 yards will be slowed by 60% while in the area of effect. Trap will exist for 1 min. Only one Frost trap can be active at a time.
  - Classic: Place a frost trap that creates an ice slick around itself for 30 sec when the first enemy approaches it. All enemies within 10 yards will be slowed by 60% while in the area of effect. Trap will exist for 1 min. Traps can only be placed when out of combat. Only one trap can be active at a time.
  - Note: Traps can now be placed in combat (the Classic out-of-combat clause is gone), one Fire trap and one Frost trap can be active at the same time, and the cooldown is 30 sec instead of 15.
- **Feign Death** (survival, Level 30, 80 Mana, Instant, 30 sec cooldown)
  - Feign death which may trick enemies into ignoring you. Lasts 6 min. If not cancelled, after 6 min you will actually die.
  - Classic: Feign death which may trick enemies into ignoring you. Lasts up to 6 min.
  - Note: The tooltip now spells out that you die if it is not cancelled within 6 min.
- **Explosive Trap** (survival, Level 34, 275 Mana, Instant, 30 sec cooldown)
  - Place a Fire trap that explodes when an enemy approaches, causing 103 to 134 Fire damage and 150 additional Fire damage over 20 sec to all within 10 yards. Trap will exist for 1 min. Only one Fire trap can be active at a time.
  - Classic: Place a fire trap that explodes when an enemy approaches, causing 208 to 264 Fire damage and 330 additional Fire damage over 20 sec to all within 10 yards. Trap will exist for 1 min. Traps can only be placed when out of combat. Only one trap can be active at a time.
  - Note: Traps can now be placed in combat (the Classic out-of-combat clause is gone), one Fire trap and one Frost trap can be active at the same time, and the cooldown is 30 sec instead of 15.
- **Serpent Sting** (marksmanship, Level 34, 115 Mana, 8-35 yd range, Instant)
  - Stings the target, causing 170 Nature damage over 15 sec. Only one Sting per Hunter can be active on any one target.
  - Classic: Stings the target, causing 555 Nature damage over 15 sec. Only one Sting per Hunter can be active on any one target.
  - Note: Rank 5 read 170 Nature damage over 15 sec on the level 38 hunter, where Classic rank 5 does 210, other ranks not seen.
- **Arcane Shot** (marksmanship, Level 36, 105 Mana, 8-35 yd range, Instant, 6 sec cooldown)
  - An instant shot that causes 94 Arcane damage.
  - Classic: An instant shot that causes 183 Arcane damage.
  - Note: Rank 5 read 94 Arcane damage on the level 38 hunter, where the Classic client gives 83 at that level, other ranks not seen.
- **Hunter's Mark** (marksmanship, Level 22, 30 Mana, 100 yd range, Instant)
  - Places the Hunter's Mark on the target, increasing the Ranged Attack Power of all attackers against that target by 59. In addition, the target of this ability can always be seen by the hunter whether it stealths or turns invisible. The target also appears on the mini-map. Lasts for 2 min.
  - Classic: Places the Hunter's Mark on the target, increasing the Ranged Attack Power of all attackers against that target by 110. In addition, the target of this ability can always be seen by the hunter whether it stealths or turns invisible. The target also appears on the mini-map. Lasts for 2 min.
  - Note: Rank 2 read 59 ranged attack power on the level 38 hunter, where Classic rank 2 gives 45, other ranks not seen.
- **Multi-Shot** (marksmanship, Level 18, 145 Mana, 8-35 yd range, 0.5 sec cast, 6 sec cooldown)
  - Fires several missiles, hitting 3 targets. Multi-Shot shares its cooldown with Aimed Shot.
  - Classic: Fires several missiles, hitting 3 targets for an additional 150 damage.
  - Note: Rank 1 is still the known rank at level 38 (Classic taught Rank 2 at 30). Cooldown is 6 sec instead of 10 and it now shares that cooldown with Aimed Shot.
- **Scorpid Sting** (marksmanship, Level 22, 94 Mana, 8-35 yd range, Instant)
  - Stings the target, reducing the target's chance to hit with all attacks by 2% for 20 sec. Only one Sting per Hunter can be active on any one target.
  - Classic: Stings the target, reducing Strength and Agility by 68 for 20 sec. Only one Sting per Hunter can be active on any one target.
  - Note: Redesigned: reduces the target's chance to hit by 2% instead of lowering Strength and Agility, and only Rank 1 is known at level 38 (Classic taught Rank 2 at 32).
- **Rapid Fire** (marksmanship, Level 26, 100 Mana, Instant, 5 min cooldown)
  - Increases ranged and melee attack speed by 40% for 15 sec.
  - Classic: Increases ranged attack speed by 40% for 15 sec.
  - Note: Now also speeds up melee attacks.
- **Aimed Shot** (marksmanship, Level 36, 160 Mana, 8-35 yd range, 2 sec cast, 6 sec cooldown)
  - An aimed shot that increases ranged damage by 55. Aimed Shot shares its cooldown with Multi-Shot.
  - Classic: An aimed shot that increases ranged damage by 600.
  - Note: Rank 3 on the level 38 hunter: 2 sec cast, plus 55 damage, and it now shares its cooldown with Multi-Shot (Classic rank 3: 3 sec cast, plus 200).
- **Viper Sting** (marksmanship, Level 36, 135 Mana, 8-35 yd range, Instant, 15 sec cooldown)
  - Stings the target, draining 616 mana over 8 sec. Only one Sting per Hunter can be active on any one target.
  - Classic: Stings the target, draining 1108 mana over 8 sec. Only one Sting per Hunter can be active on any one target.
  - Note: Gained a 15 sec cooldown, the drain itself is unchanged.

### Verified unchanged

Aspect of the Monkey, Aspect of the Hawk, Call Pet, Dismiss Pet, Feed Pet, Revive Pet, Tame Beast, Mend Pet, Eagle Eye, Scare Beast, Aspect of the Cheetah, Beast Lore, Track Beasts, Track Humanoids, Wing Clip, Track Undead, Disengage, Track Hidden, Track Elementals, Track Demons, Auto Shot, Concussive Shot, Distracting Shot, Flare, Passive

### Not yet verified (Classic text assumed)

Aspect of the Pack, Aspect of the Wild, Tranquilizing Shot, Track Giants, Counterattack, Track Dragonkin, Wyvern Sting, Volley, Trueshot Aura, Attack, Claw, Defensive, Dive, Follow, Great Stamina, Growl, Natural Armor, Stay

## Talents

### Beast Mastery

16 talents, 2 new. 31-point talent: Bestial Wrath.

#### Row 1 (0 points)

- **Deadly Aspects** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: While Aspect of the Hawk is active, all normal ranged attacks have a 2% chance of increasing ranged attack speed by 30% for 12 sec. While Aspect of the Beast is active, all melee auto attacks have a 2% chance of increasing melee attack speed by 30% for 12 sec.
  - Rank 2: While Aspect of the Hawk is active, all normal ranged attacks have a 4% chance of increasing ranged attack speed by 60% for 24 sec. While Aspect of the Beast is active, all melee auto attacks have a 4% chance of increasing melee attack speed by 60% for 24 sec.
  - Rank 3: While Aspect of the Hawk is active, all normal ranged attacks have a 6% chance of increasing ranged attack speed by 90% for 36 sec. While Aspect of the Beast is active, all melee auto attacks have a 6% chance of increasing melee attack speed by 90% for 36 sec.
  - Rank 4: While Aspect of the Hawk is active, all normal ranged attacks have a 8% chance of increasing ranged attack speed by 120% for 48 sec. While Aspect of the Beast is active, all melee auto attacks have a 8% chance of increasing melee attack speed by 120% for 48 sec.
  - Rank 5: While Aspect of the Hawk is active, all normal ranged attacks have a 10% chance of increasing ranged attack speed by 150% for 60 sec. While Aspect of the Beast is active, all melee auto attacks have a 10% chance of increasing melee attack speed by 150% for 60 sec.
  - Classic (5 ranks):
    - Rank 1: While Aspect of the Hawk is active, all normal ranged attacks have a 1% chance of increasing ranged attack speed by 30% for 12 sec.
    - Rank 2: While Aspect of the Hawk is active, all normal ranged attacks have a 2% chance of increasing ranged attack speed by 30% for 12 sec.
    - Rank 3: While Aspect of the Hawk is active, all normal ranged attacks have a 3% chance of increasing ranged attack speed by 30% for 12 sec.
    - Rank 4: While Aspect of the Hawk is active, all normal ranged attacks have a 4% chance of increasing ranged attack speed by 30% for 12 sec.
    - Rank 5: While Aspect of the Hawk is active, all normal ranged attacks have a 5% chance of increasing ranged attack speed by 30% for 12 sec.
- **Endurance Training** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the Health and Armor of your pets by 3%.
  - Rank 2: Increases the Health and Armor of your pets by 6%.
  - Rank 3: Increases the Health and Armor of your pets by 9%.
  - Rank 4: Increases the Health and Armor of your pets by 12%.
  - Rank 5: Increases the Health and Armor of your pets by 15%.
  - Classic (5 ranks):
    - Rank 1: Increases the Health of your pets by 3%.
    - Rank 2: Increases the Health of your pets by 6%.
    - Rank 3: Increases the Health of your pets by 9%.
    - Rank 4: Increases the Health of your pets by 12%.
    - Rank 5: Increases the Health of your pets by 15%.

#### Row 2 (5 points)

- **Focused Fire** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases all damage you deal by 1% while your pet is active.
  - Rank 2: Increases all damage you deal by 2% while your pet is active.
- **Improved Aspect of the Monkey** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the Dodge bonus of your Aspect of the Monkey by 2%. Additionally, your pet gains 50% of the effect of your Aspect of the Monkey ability.
  - Rank 2: Increases the Dodge bonus of your Aspect of the Monkey by 4%. Additionally, your pet gains 50% of the effect of your Aspect of the Monkey ability.
  - Rank 3: Increases the Dodge bonus of your Aspect of the Monkey by 6%. Additionally, your pet gains 50% of the effect of your Aspect of the Monkey ability.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Pathfinding** (2 ranks, column 3) [unchanged]
  - Rank 1: Increases the speed bonus of your Aspect of the Cheetah and Aspect of the Pack by 3%.
  - Rank 2: Increases the speed bonus of your Aspect of the Cheetah and Aspect of the Pack by 6%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Revive Pet** (2 ranks, column 4) [unchanged]
  - Rank 1: Revive Pet's casting time is reduced by 3 sec, mana cost is reduced by 20%, and increases the health your pet returns with by an additional 15%.
  - Rank 2: Revive Pet's casting time is reduced by 6 sec, mana cost is reduced by 40%, and increases the health your pet returns with by an additional 30%.

#### Row 3 (10 points)

- **Bestial Swiftness** (1 rank, column 2) [CHANGED]
  - Increases the movement speed of your pets by 30%.
  - Classic: Increases the outdoor movement speed of your pets by 30%.
- **Unleashed Fury** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage done by your pets and hawks by 3%.
  - Rank 2: Increases the damage done by your pets and hawks by 6%.
  - Rank 3: Increases the damage done by your pets and hawks by 9%.
  - Rank 4: Increases the damage done by your pets and hawks by 12%.
  - Rank 5: Increases the damage done by your pets and hawks by 15%.
  - Classic (5 ranks):
    - Rank 1: Increases the damage done by your pets by 4%.
    - Rank 2: Increases the damage done by your pets by 8%.
    - Rank 3: Increases the damage done by your pets by 12%.
    - Rank 4: Increases the damage done by your pets by 16%.
    - Rank 5: Increases the damage done by your pets by 20%.

#### Row 4 (15 points)

- **Improved Mend Pet** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives your Mend Pet spell a 15% chance of cleansing 1 Curse, Disease, Magic, or Poison effect from your pet each time it heals and reduces the Mana cost by 10%.
  - Rank 2: Gives your Mend Pet spell a 30% chance of cleansing 1 Curse, Disease, Magic, or Poison effect from your pet each time it heals and reduces the Mana cost by 20%.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Ferocity** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike chance of your pets and hawks by 2%.
  - Rank 2: Increases the critical strike chance of your pets and hawks by 4%.
  - Rank 3: Increases the critical strike chance of your pets and hawks by 6%.
  - Rank 4: Increases the critical strike chance of your pets and hawks by 8%.
  - Rank 5: Increases the critical strike chance of your pets and hawks by 10%.
  - Classic (5 ranks):
    - Rank 1: Increases the critical strike chance of your pets by 3%.
    - Rank 2: Increases the critical strike chance of your pets by 6%.
    - Rank 3: Increases the critical strike chance of your pets by 9%.
    - Rank 4: Increases the critical strike chance of your pets by 12%.
    - Rank 5: Increases the critical strike chance of your pets by 15%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Summon Hawk** (1 rank, column 4) [NEW]
  - Cost: 80 Mana, 35 yd range, Instant, 6 sec cooldown
  - Command a hawk to dive-bomb your targeted enemy, dealing 53 Physical damage and continuing its assault for 18 sec. Only 2 hawks can be active at once. Summon Hawk shares its cooldown with Arcane Shot.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 5 (20 points)

- **Spirit Bond** (2 ranks, column 1) [unchanged]
  - Rank 1: While your pet is active, you and your pet will regenerate 1% of total health every 10 sec.
  - Rank 2: While your pet is active, you and your pet will regenerate 2% of total health every 10 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Intimidation** (1 rank, column 2) [CHANGED]
  - Requires: Bestial Swiftness (1/1)
  - Cost: 84 Mana, 100 yd range, Instant, 1 min cooldown
  - Command your pet to Stun the target for 3 sec on its next successful attack, which also gains 100% increased critical strike chance. Generates high threat.
  - Classic: Command your pet to intimidate the target on the next successful melee attack, causing a high amount of threat and stunning the target for 3 sec.
- **Bestial Discipline** (2 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the Focus regeneration of your pets by 10% and allows 25% of your Mana regeneration to continue while casting.
  - Rank 2: Increases the Focus regeneration of your pets by 20% and allows 50% of your Mana regeneration to continue while casting.
  - Classic (2 ranks):
    - Rank 1: Increases the Focus regeneration of your pets by 10%.
    - Rank 2: Increases the Focus regeneration of your pets by 20%.

#### Row 6 (25 points)

- **Frenzy** (5 ranks, column 3) [unchanged]
  - Requires: Ferocity (5/5)
  - Rank 1: Gives your pet a 20% chance to gain a 30% attack speed increase for 8 sec after dealing a critical strike.
  - Rank 2: Gives your pet a 40% chance to gain a 30% attack speed increase for 8 sec after dealing a critical strike.
  - Rank 3: Gives your pet a 60% chance to gain a 30% attack speed increase for 8 sec after dealing a critical strike.
  - Rank 4: Gives your pet a 80% chance to gain a 30% attack speed increase for 8 sec after dealing a critical strike.
  - Rank 5: Gives your pet a 100% chance to gain a 30% attack speed increase for 8 sec after dealing a critical strike.

#### Row 7 (30 points)

- **Bestial Wrath** (1 rank, column 2) [unchanged]
  - Requires: Intimidation (1/1)
  - Cost: 125 Mana, 100 yd range, Instant, 2 min cooldown
  - Send your pet into a rage causing 50% additional damage for 18 sec. While enraged, the beast does not feel pity or remorse or fear and it cannot be stopped unless killed.

### Marksmanship

16 talents, 5 new. 31-point talent: Sniper Shot.

#### Row 1 (0 points)

- **Hawk Eye** (3 ranks, column 1) [unchanged]
  - Rank 1: Increases the range of your ranged weapons by 2 yards.
  - Rank 2: Increases the range of your ranged weapons by 4 yards.
  - Rank 3: Increases the range of your ranged weapons by 6 yards.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Concussive Shot** (5 ranks, column 2) [unchanged]
  - Rank 1: Gives your Concussive Shot a 4% chance to stun the target for 3 sec.
  - Rank 2: Gives your Concussive Shot a 8% chance to stun the target for 3 sec.
  - Rank 3: Gives your Concussive Shot a 12% chance to stun the target for 3 sec.
  - Rank 4: Gives your Concussive Shot a 16% chance to stun the target for 3 sec.
  - Rank 5: Gives your Concussive Shot a 20% chance to stun the target for 3 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Lethal Attacks** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your critical strike chance with all attacks by 1%.
  - Rank 2: Increases your critical strike chance with all attacks by 2%.
  - Rank 3: Increases your critical strike chance with all attacks by 3%.
  - Rank 4: Increases your critical strike chance with all attacks by 4%.
  - Rank 5: Increases your critical strike chance with all attacks by 5%.
  - Classic (5 ranks):
    - Rank 1: Increases your critical strike chance with ranged weapons by 1%.
    - Rank 2: Increases your critical strike chance with ranged weapons by 2%.
    - Rank 3: Increases your critical strike chance with ranged weapons by 3%.
    - Rank 4: Increases your critical strike chance with ranged weapons by 4%.
    - Rank 5: Increases your critical strike chance with ranged weapons by 5%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 2 (5 points)

- **Improved Stings** (3 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the damage of your Serpent Sting ability by 6%, reduces the cooldown of your Viper Sting ability by 2 sec, and increases the duration of your Scorpid Sting ability by 15 sec.
  - Rank 2: Increases the damage of your Serpent Sting ability by 12%, reduces the cooldown of your Viper Sting ability by 4 sec, and increases the duration of your Scorpid Sting ability by 30 sec.
  - Rank 3: Increases the damage of your Serpent Sting ability by 18%, reduces the cooldown of your Viper Sting ability by 6 sec, and increases the duration of your Scorpid Sting ability by 45 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Efficiency** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the Mana cost of your Shots, Stings, and melee abilities by 3%.
  - Rank 2: Reduces the Mana cost of your Shots, Stings, and melee abilities by 6%.
  - Rank 3: Reduces the Mana cost of your Shots, Stings, and melee abilities by 9%.
  - Rank 4: Reduces the Mana cost of your Shots, Stings, and melee abilities by 12%.
  - Rank 5: Reduces the Mana cost of your Shots, Stings, and melee abilities by 15%.
  - Classic (5 ranks):
    - Rank 1: Reduces the Mana cost of your Shots and Stings by 2%.
    - Rank 2: Reduces the Mana cost of your Shots and Stings by 4%.
    - Rank 3: Reduces the Mana cost of your Shots and Stings by 6%.
    - Rank 4: Reduces the Mana cost of your Shots and Stings by 8%.
    - Rank 5: Reduces the Mana cost of your Shots and Stings by 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Careful Aim** (5 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases your Attack Power by 20% of your Intellect.
  - Rank 2: Increases your Attack Power by 40% of your Intellect.
  - Rank 3: Increases your Attack Power by 60% of your Intellect.
  - Rank 4: Increases your Attack Power by 80% of your Intellect.
  - Rank 5: Increases your Attack Power by 100% of your Intellect.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Rapid Killing** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Reduces the cooldown on your Rapid Fire ability by 1 min. In addition, when you kill a non-trivial enemy or it dies while afflicted by your Serpent Sting, you gain Rapid Killing, increasing the damage of your next Shot ability within 20 sec by 10%.
  - Rank 2: Reduces the cooldown on your Rapid Fire ability by 2 min. In addition, when you kill a non-trivial enemy or it dies while afflicted by your Serpent Sting, you gain Rapid Killing, increasing the damage of your next Shot ability within 40 sec by 20%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Arcane Shot** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the cooldown of your Arcane Shot by 0.3 sec. Does not affect the cooldown of abilities which share a cooldown with Arcane Shot.
  - Rank 2: Reduces the cooldown of your Arcane Shot by 0.6 sec. Does not affect the cooldown of abilities which share a cooldown with Arcane Shot.
  - Rank 3: Reduces the cooldown of your Arcane Shot by 0.9 sec. Does not affect the cooldown of abilities which share a cooldown with Arcane Shot.
  - Rank 4: Reduces the cooldown of your Arcane Shot by 1.2 sec. Does not affect the cooldown of abilities which share a cooldown with Arcane Shot.
  - Rank 5: Reduces the cooldown of your Arcane Shot by 1.5 sec. Does not affect the cooldown of abilities which share a cooldown with Arcane Shot.
  - Classic (5 ranks):
    - Rank 1: Reduces the cooldown of your Arcane Shot by 0.2 sec.
    - Rank 2: Reduces the cooldown of your Arcane Shot by 0.4 sec.
    - Rank 3: Reduces the cooldown of your Arcane Shot by 0.6 sec.
    - Rank 4: Reduces the cooldown of your Arcane Shot by 0.8 sec.
    - Rank 5: Reduces the cooldown of your Arcane Shot by 1 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Lone Wolf** (1 rank, column 4) [NEW]
  - You deal 20% increased damage with all attacks while you do not have an active pet.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Trueshot Aura** (1 rank, column 2) [CHANGED]
  - Cost: 180 Mana, Instant
  - Increases the Ranged Attack Power of party members within 45 yards by 30. Lasts 30 min.
  - Classic: Increases the attack power of party members within 45 yards by 50. Lasts 30 min.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Mortal Shots** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Requires: Careful Aim (5/5)
  - Rank 1: Increases the critical strike damage bonus on all ranged abilities by 6%.
  - Rank 2: Increases the critical strike damage bonus on all ranged abilities by 12%.
  - Rank 3: Increases the critical strike damage bonus on all ranged abilities by 18%.
  - Rank 4: Increases the critical strike damage bonus on all ranged abilities by 24%.
  - Rank 5: Increases the critical strike damage bonus on all ranged abilities by 30%.
  - Classic (5 ranks):
    - Rank 1: Increases your ranged weapon critical strike damage bonus by 6%.
    - Rank 2: Increases your ranged weapon critical strike damage bonus by 12%.
    - Rank 3: Increases your ranged weapon critical strike damage bonus by 18%.
    - Rank 4: Increases your ranged weapon critical strike damage bonus by 24%.
    - Rank 5: Increases your ranged weapon critical strike damage bonus by 30%.

#### Row 5 (20 points)

- **Rapid Recuperation** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Requires: Rapid Killing (2/2)
  - Rank 1: Hitting a target with your Serpent Sting ability grants you 25% and consuming Rapid Killing grants you 50% of your Mana regeneration while casting for the next 15 sec.
  - Rank 2: Hitting a target with your Serpent Sting ability grants you 50% and consuming Rapid Killing grants you 100% of your Mana regeneration while casting for the next 30 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Barrage** (3 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage done by your Multi-Shot, Aimed Shot, and Volley abilities by 3%.
  - Rank 2: Increases the damage done by your Multi-Shot, Aimed Shot, and Volley abilities by 6%.
  - Rank 3: Increases the damage done by your Multi-Shot, Aimed Shot, and Volley abilities by 9%.
  - Classic (3 ranks):
    - Rank 1: Increases the damage done by your Multi-Shot and Volley spells by 5%.
    - Rank 2: Increases the damage done by your Multi-Shot and Volley spells by 10%.
    - Rank 3: Increases the damage done by your Multi-Shot and Volley spells by 15%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Scatter Shot** (1 rank, column 4) [unchanged]
  - Cost: 84 Mana, 15 yd range, Instant, 30 sec cooldown
  - A short-range shot that deals 50% weapon damage and disorients the target for 4 sec. Any damage caused will remove the effect. Turns off your attack when used.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 6 (25 points)

- **Ranged Weapon Specialization** (5 ranks, column 3) [unchanged]
  - Rank 1: Increases the damage you deal with ranged weapons by 1%.
  - Rank 2: Increases the damage you deal with ranged weapons by 2%.
  - Rank 3: Increases the damage you deal with ranged weapons by 3%.
  - Rank 4: Increases the damage you deal with ranged weapons by 4%.
  - Rank 5: Increases the damage you deal with ranged weapons by 5%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 7 (30 points)

- **Sniper Shot** (1 rank, column 2) [CHANGED]
  - Requires: Trueshot Aura (1/1)
  - Cost: 365 Mana, 8-35 yd range, 4 sec cast, 15 sec cooldown
  - A steady snipe that increases ranged damage by 160.
  - Classic: An aimed shot that increases ranged damage by 70.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

### Survival

18 talents, 8 new. 31-point talent: Lacerating Strikes.

#### Row 1 (0 points)

- **Improved Tracking** (5 ranks, column 2) [NEW] [ranks 2+ unverified]
  - Rank 1: While tracking Beasts, Demons, Dragonkin, Elementals, Giants, Humanoids, or Undead, all damage you deal to the tracked creature type is increased by 1%.
  - Rank 2: While tracking Beasts, Demons, Dragonkin, Elementals, Giants, Humanoids, or Undead, all damage you deal to the tracked creature type is increased by 2%.
  - Rank 3: While tracking Beasts, Demons, Dragonkin, Elementals, Giants, Humanoids, or Undead, all damage you deal to the tracked creature type is increased by 3%.
  - Rank 4: While tracking Beasts, Demons, Dragonkin, Elementals, Giants, Humanoids, or Undead, all damage you deal to the tracked creature type is increased by 4%.
  - Rank 5: While tracking Beasts, Demons, Dragonkin, Elementals, Giants, Humanoids, or Undead, all damage you deal to the tracked creature type is increased by 5%.
- **Deflection** (5 ranks, column 3) [unchanged]
  - Rank 1: Increases your Parry chance by 2%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Increases your Parry chance by 3%.
  - Rank 4: Increases your Parry chance by 4%.
  - Rank 5: Increases your Parry chance by 5%.

#### Row 2 (5 points)

- **Entrapment** (5 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: When your traps are triggered, all affected targets are Entrapped, preventing them from moving for 1 sec.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Rank 4: Rank 4 values not yet captured from footage.
  - Rank 5: Rank 5 values not yet captured from footage.
  - Classic (5 ranks):
    - Rank 1: Gives your Immolation Trap, Frost Trap, and Explosive Trap a 5% chance to entrap the target, preventing them from moving for 5 sec.
    - Rank 2: Gives your Immolation Trap, Frost Trap, and Explosive Trap a 10% chance to entrap the target, preventing them from moving for 5 sec.
    - Rank 3: Gives your Immolation Trap, Frost Trap, and Explosive Trap a 15% chance to entrap the target, preventing them from moving for 5 sec.
    - Rank 4: Gives your Immolation Trap, Frost Trap, and Explosive Trap a 20% chance to entrap the target, preventing them from moving for 5 sec.
    - Rank 5: Gives your Immolation Trap, Frost Trap, and Explosive Trap a 25% chance to entrap the target, preventing them from moving for 5 sec.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Savage Strikes** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike chance of all your melee abilities by 2%.
  - Rank 2: Increases the critical strike chance of all your melee abilities by 4%.
  - Classic (2 ranks):
    - Rank 1: Increases the critical strike chance of Raptor Strike and Mongoose Bite by 10%.
    - Rank 2: Increases the critical strike chance of Raptor Strike and Mongoose Bite by 20%.
  - Note: Rank 1 from footage (2% crit on all melee abilities, replacing Classic's 10% on Raptor Strike and Mongoose Bite). Rank 2 not shown, scaled like Classic (linear, 2x at max rank) to 4%.
- **Survivalist** (5 ranks, column 3) [unchanged]
  - Rank 1: Increases your total Health by 2%.
  - Rank 2: Increases total health by 4%.
  - Rank 3: Increases total health by 6%.
  - Rank 4: Increases total health by 8%.
  - Rank 5: Increases total health by 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Wing Clip** (3 ranks, column 4) [unchanged]
  - Rank 1: Gives your Wing Clip ability a 7% chance to immobilize the target for 5 sec.
  - Rank 2: Gives your Wing Clip ability a 8% chance to immobilize the target for 5 sec.
  - Rank 3: Gives your Wing Clip ability a 12% chance to immobilize the target for 5 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Clever Traps** (2 ranks, column 1) [unchanged]
  - Rank 1: Increases the duration of Freezing and Frost trap effects by 15% and the damage of Immolation and Explosive trap effects by 15%.
  - Rank 2: Increases the duration of Freezing and Frost trap effects by 30% and the damage of Immolation and Explosive trap effects by 30%.
- **Surefooted** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your hit chance by 1% and reduces the duration of movement impairing effects on you by 10%.
  - Rank 2: Increases your hit chance by 2% and reduces the duration of movement impairing effects on you by 20%.
  - Rank 3: Increases your hit chance by 3% and reduces the duration of movement impairing effects on you by 30%.
  - Classic (3 ranks):
    - Rank 1: Increases hit chance by 1% and increases the chance movement impairing effects will be resisted by an additional 5%.
    - Rank 2: Increases hit chance by 2% and increases the chance movement impairing effects will be resisted by an additional 10%.
    - Rank 3: Increases hit chance by 3% and increases the chance movement impairing effects will be resisted by an additional 15%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Deterrence** (1 rank, column 3) [unchanged]
  - Cost: Instant, 5 min cooldown
  - When activated, increases your Dodge and Parry chance by 25% for 10 sec.

#### Row 4 (15 points)

- **Survival Tactics** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases your chance to hit with your Trap and Feign Death abilities by 5%.
  - Rank 2: Increases your chance to hit with your Trap and Feign Death abilities by 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Predator's Edge** (5 ranks, column 2) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases your melee critical strike damage by 6% and your offhand weapon damage by 10%.
  - Rank 2: Increases your melee critical strike damage by 12% and your offhand weapon damage by 20%.
  - Rank 3: Increases your melee critical strike damage by 18% and your offhand weapon damage by 30%.
  - Rank 4: Increases your melee critical strike damage by 24% and your offhand weapon damage by 40%.
  - Rank 5: Increases your melee critical strike damage by 30% and your offhand weapon damage by 50%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Counterattack** (1 rank, column 3) [unchanged]
  - Requires: Deterrence (1/1)
  - Cost: 30 Mana, Melee Range, Instant, 5 sec cooldown
  - A strike that becomes active after parrying an opponent's attack. This attack deals 50% weapon damage plus 13 and immobilizes the target for 5 sec. Counterattack cannot be blocked, dodged, or parried.

#### Row 5 (20 points)

- **Resourcefulness** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Reduces the mana cost of your Trap abilities and melee abilities by 30%. In addition, your critical strikes have a 30% chance to allow 50% of your Mana regeneration to continue while casting for 30 sec.
  - Rank 2: Reduces the mana cost of your Trap abilities and melee abilities by 60%. In addition, your critical strikes have a 60% chance to allow 100% of your Mana regeneration to continue while casting for 60 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Expose Prey** (2 ranks, column 2) [NEW] [ranks 2+ unverified]
  - Rank 1: Your attacks against targets with Hunter's Mark have a 5% chance to activate your Mongoose Bite for 5 sec.
  - Rank 2: Your attacks against targets with Hunter's Mark have a 10% chance to activate your Mongoose Bite for 10 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Survivalist's Discipline** (2 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Reduces the cooldown of your Trap and Deterrence abilities by 20%.
  - Rank 2: Reduces the cooldown of your Trap and Deterrence abilities by 40%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Strider Kick** (1 rank, column 4) [NEW]
  - Cost: 61 Mana, Melee Range, Instant, 8 sec cooldown
  - A powerful kick that deals 100% melee weapon damage.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 6 (25 points)

- **Lightning Reflexes** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your Agility by 2%.
  - Rank 2: Increases your Agility by 4%.
  - Rank 3: Increases your Agility by 6%.
  - Rank 4: Increases your Agility by 8%.
  - Rank 5: Increases your Agility by 10%.
  - Classic (5 ranks):
    - Rank 1: Increases your Agility by 3%.
    - Rank 2: Increases your Agility by 6%.
    - Rank 3: Increases your Agility by 9%.
    - Rank 4: Increases your Agility by 12%.
    - Rank 5: Increases your Agility by 15%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 7 (30 points)

- **Lacerating Strikes** (1 rank, column 2) [NEW]
  - Requires: Expose Prey (2/2)
  - Your Mongoose Bite also causes the target to Bleed for damage over 21 sec equal to 40% of the damage done by Mongoose Bite.
