# Warlock

Source: wowforevertalents.com dataset (BlizzCon 2026 footage overlaid on Classic Era 1.15.9 client tables), data version 2026-09-13/14. See [../README.md](../README.md) for caveats. Numbers on new talents were read off level 38 demo characters, so absolute damage and mana values are not level 60 values.

Talents: 52 total, 23 new, 26 changed, 3 unchanged (verified).

Row N needs 5*(N-1) points spent in that tree. Row 1 is the top row. Rows 3, 4, 5 and 7 hold the 11, 16, 21 and 31 point talents.

Warning on multi-rank NEW and CHANGED talents: the footage showed rank 1 only. The dataset scaled the other ranks linearly from rank 1, which sometimes produces nonsense (for example Maelstrom Weapon rank 5 reading 'stacks up to 25 times'). Treat rank 1 as observed and the rest as a guess until the beta client is out. The [ranks 2+ unverified] tag marks every multi-rank new or changed talent. Read the Note line, it says when a higher rank was actually seen on a BlizzCon slide.

## Spellbook changes (trainer abilities)

0 of 65 trainer abilities confirmed from footage. Abilities not listed below are either verified unchanged or still Classic placeholders (unverified).

### Not yet verified (Classic text assumed)

Demon Skin, Summon Imp, Create Healthstone (Minor), Summon Voidwalker, Health Funnel, Unending Breath, Create Soulstone (Minor), Demon Armor, Ritual of Summoning, Summon Incubus, Summon Succubus, Create Healthstone (Lesser), Eye of Kilrogg, Sense Demons, Detect Lesser Invisibility, Banish, Create Firestone (Lesser), Create Soulstone (Lesser), Subjugate Demon, Summon Felhunter, Shadow Ward, Create Healthstone, Create Firestone, Create Spellstone, Detect Invisibility, Create Soulstone, Summon Felsteed, Create Firestone (Greater), Create Healthstone (Greater), Create Spellstone (Greater), Create Soulstone (Greater), Detect Greater Invisibility, Inferno, Create Firestone (Major), Create Healthstone (Major), Create Soulstone (Major), Create Spellstone (Major), Ritual of Doom, Summon Dreadsteed, Corruption, Curse of Weakness, Life Tap, Curse of Agony, Fear, Drain Soul, Curse of Recklessness, Drain Life, Drain Mana, Curse of Tongues, Curse of the Elements, Curse of Idiocy, Siphon Life, Howl of Terror, Death Coil, Curse of Shadow, Dark Pact, Curse of Doom, Immolate, Shadow Bolt, Searing Pain, Rain of Fire, Shadowburn, Hellfire, Conflagrate, Soul Fire

## Talents

### Affliction

17 talents, 8 new. 31-point talent: Drain Hope.

#### Row 1 (0 points)

- **Improved Life Tap** (2 ranks, column 1) [unchanged]
  - Rank 1: Increases the amount of Mana awarded by your Life Tap spell by 10%.
  - Rank 2: Increases the amount of Mana awarded by your Life Tap spell by 20%.
- **Suppression** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your chance to hit with all spells and attacks by 1% and reduces all threat you generate by 4%.
  - Rank 2: Increases your chance to hit with all spells and attacks by 2% and reduces all threat you generate by 8%.
  - Rank 3: Increases your chance to hit with all spells and attacks by 3% and reduces all threat you generate by 12%.
  - Rank 4: Increases your chance to hit with all spells and attacks by 4% and reduces all threat you generate by 16%.
  - Rank 5: Increases your chance to hit with all spells and attacks by 5% and reduces all threat you generate by 20%.
  - Classic (5 ranks):
    - Rank 1: Reduces the chance for enemies to resist your Affliction spells by 2%.
    - Rank 2: Reduces the chance for enemies to resist your Affliction spells by 4%.
    - Rank 3: Reduces the chance for enemies to resist your Affliction spells by 6%.
    - Rank 4: Reduces the chance for enemies to resist your Affliction spells by 8%.
    - Rank 5: Reduces the chance for enemies to resist your Affliction spells by 10%.
- **Improved Corruption** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the casting time of your Corruption spell by 0.4 sec and increases the damage it deals by 2%.
  - Rank 2: Reduces the casting time of your Corruption spell by 0.8 sec and increases the damage it deals by 4%.
  - Rank 3: Reduces the casting time of your Corruption spell by 1.2 sec and increases the damage it deals by 6%.
  - Rank 4: Reduces the casting time of your Corruption spell by 1.6 sec and increases the damage it deals by 8%.
  - Rank 5: Reduces the casting time of your Corruption spell by 2 sec and increases the damage it deals by 10%.
  - Classic (5 ranks):
    - Rank 1: Reduces the casting time of your Corruption spell by 0.4 sec.
    - Rank 2: Reduces the casting time of your Corruption spell by 0.8 sec.
    - Rank 3: Reduces the casting time of your Corruption spell by 1.2 sec.
    - Rank 4: Reduces the casting time of your Corruption spell by 1.6 sec.
    - Rank 5: Reduces the casting time of your Corruption spell by 2 sec.

#### Row 2 (5 points)

- **Malediction** (5 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases all periodic damage done by your Warlock spells by 1%.
  - Rank 2: Increases all periodic damage done by your Warlock spells by 2%.
  - Rank 3: Increases all periodic damage done by your Warlock spells by 3%.
  - Rank 4: Increases all periodic damage done by your Warlock spells by 4%.
  - Rank 5: Increases all periodic damage done by your Warlock spells by 5%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Soul Harvesting** (2 ranks, column 2) [NEW] [ranks 2+ unverified]
  - Rank 1: You gain Soul Harvest for 10 sec if a victim is killed while afflicted with your Drain Soul. Soul Harvest allows your Mana to regenerate at 50% of normal speed while you are casting spells, and grants a 50% increase to your Mana regeneration.
  - Rank 2: You gain Soul Harvest for 20 sec if a victim is killed while afflicted with your Drain Soul. Soul Harvest allows your Mana to regenerate at 100% of normal speed while you are casting spells, and grants a 100% increase to your Mana regeneration.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Drains** (3 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the damage done or health drained by your Drain Life and Drain Soul spells by 2% per each of your other Affliction effects active on the target, up to a maximum increase of 6%. When your Drain Soul damages targets below 20% health, this bonus is tripled. Additionally, your Drain Life range is extended by 3 yards.
  - Rank 2: Increases the damage done or health drained by your Drain Life and Drain Soul spells by 4% per each of your other Affliction effects active on the target, up to a maximum increase of 12%. When your Drain Soul damages targets below 40% health, this bonus is tripled. Additionally, your Drain Life range is extended by 6 yards.
  - Rank 3: Increases the damage done or health drained by your Drain Life and Drain Soul spells by 6% per each of your other Affliction effects active on the target, up to a maximum increase of 18%. When your Drain Soul damages targets below 60% health, this bonus is tripled. Additionally, your Drain Life range is extended by 9 yards.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Improved Bane of Agony** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the damage done by your Bane of Agony by 5%.
  - Rank 2: Increases the damage done by your Bane of Agony by 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Fel Concentration** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives you a 23% chance to avoid interruption caused by damage while channeling or casting your Drain Life, Drain Mana, or Drain Soul spells.
  - Rank 2: Gives you a 46% chance to avoid interruption caused by damage while channeling or casting your Drain Life, Drain Mana, or Drain Soul spells.
  - Rank 3: Gives you a 69% chance to avoid interruption caused by damage while channeling or casting your Drain Life, Drain Mana, or Drain Soul spells.
  - Classic (5 ranks):
    - Rank 1: Gives you a 14% chance to avoid interruption caused by damage while channeling the Drain Life, Drain Mana, or Drain Soul spell.
    - Rank 2: Gives you a 28% chance to avoid interruption caused by damage while channeling the Drain Life, Drain Mana, or Drain Soul spell.
    - Rank 3: Gives you a 42% chance to avoid interruption caused by damage while channeling the Drain Life, Drain Mana, or Drain Soul spell.
    - Rank 4: Gives you a 56% chance to avoid interruption caused by damage while channeling the Drain Life, Drain Mana, or Drain Soul spell.
    - Rank 5: Gives you a 70% chance to avoid interruption caused by damage while channeling the Drain Life, Drain Mana, or Drain Soul spell.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Amplify Curse** (1 rank, column 3) [CHANGED]
  - Cost: Instant, 3 min cooldown
  - Increases the effect of your next Curse of Weakness or Bane of Agony by 50%, or your next Curse of Exhaustion by 20%. Lasts 30 sec.
  - Classic: Increases the effect of your next Curse of Weakness or Curse of Agony by 50%, or your next Curse of Exhaustion by 20%. Lasts 30 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Pandemic** (3 ranks, column 4) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike damage bonus of your Corruption, Bane of Agony, Bane of Doom, Drain Soul, Drain Life, Siphon Life, and Drain Hope spells by 33%.
  - Rank 2: Increases the critical strike damage bonus of your Corruption, Bane of Agony, Bane of Doom, Drain Soul, Drain Life, Siphon Life, and Drain Hope spells by 67%.
  - Rank 3: Increases the critical strike damage bonus of your Corruption, Bane of Agony, Bane of Doom, Drain Soul, Drain Life, Siphon Life, and Drain Hope spells by 100%.
  - Note: Slot from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Malevolence** (5 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the critical effect chance of your Shadow spells by 1%.
  - Rank 2: Increases the critical effect chance of your Shadow spells by 2%.
  - Rank 3: Increases the critical effect chance of your Shadow spells by 3%.
  - Rank 4: Increases the critical effect chance of your Shadow spells by 4%.
  - Rank 5: Increases the critical effect chance of your Shadow spells by 5%.
- **Nightfall** (2 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Requires: Fel Concentration (3/3)
  - Rank 1: Gives your Corruption, Drain Soul, and Drain Life spells a 2% chance to cause you to enter a Shadow Trance after damaging the opponent. The Shadow Trance reduces the casting time of your next Shadow Bolt spell by 100%.
  - Rank 2: Gives your Corruption, Drain Soul, and Drain Life spells a 4% chance to cause you to enter a Shadow Trance after damaging the opponent. The Shadow Trance reduces the casting time of your next Shadow Bolt spell by 200%.
  - Classic (2 ranks):
    - Rank 1: Gives your Corruption and Drain Life spells a 2% chance to cause you to enter a Shadow Trance state after damaging the opponent. The Shadow Trance state reduces the casting time of your next Shadow Bolt spell by 100%.
    - Rank 2: Gives your Corruption and Drain Life spells a 4% chance to cause you to enter a Shadow Trance state after damaging the opponent. The Shadow Trance state reduces the casting time of your next Shadow Bolt spell by 100%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Curse of Exhaustion** (1 rank, column 3) [CHANGED]
  - Requires: Amplify Curse (1/1)
  - Cost: 69 Mana, 30 yd range, Instant
  - Reduces the target's movement speed by 30% for 12 sec. Only one Curse per Warlock can be active on any one target.
  - Classic: Reduces the target's movement speed by 10% for 12 sec. Only one Curse per Warlock can be active on any one target.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 5 (20 points)

- **Siphon Life** (1 rank, column 2) [unchanged]
  - Requires: Nightfall (2/2)
  - Cost: 150 Mana, 30 yd range, Instant
  - Transfers 15 health from the target to the caster every 3 sec. Lasts 30 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Soul Siphon** (3 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the rate at which your Drain Life and Drain Soul deal damage by 17%, but reduces your healing from Drain Life by 10%.
  - Rank 2: Increases the rate at which your Drain Life and Drain Soul deal damage by 34%, but reduces your healing from Drain Life by 20%.
  - Rank 3: Increases the rate at which your Drain Life and Drain Soul deal damage by 51%, but reduces your healing from Drain Life by 30%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 6 (25 points)

- **Shadow Mastery** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage dealt or life drained by your Shadow spells by 1%.
  - Rank 2: Increases the damage dealt or life drained by your Shadow spells by 2%.
  - Rank 3: Increases the damage dealt or life drained by your Shadow spells by 3%.
  - Rank 4: Increases the damage dealt or life drained by your Shadow spells by 4%.
  - Rank 5: Increases the damage dealt or life drained by your Shadow spells by 5%.
  - Classic (5 ranks):
    - Rank 1: Increases the damage dealt or life drained by your Shadow spells by 2%.
    - Rank 2: Increases the damage dealt or life drained by your Shadow spells by 4%.
    - Rank 3: Increases the damage dealt or life drained by your Shadow spells by 6%.
    - Rank 4: Increases the damage dealt or life drained by your Shadow spells by 8%.
    - Rank 5: Increases the damage dealt or life drained by your Shadow spells by 10%.

#### Row 7 (30 points)

- **Drain Hope** (1 rank, column 2) [NEW]
  - Requires: Siphon Life (1/1)
  - Cost: 200 Mana, 30 yd range, Channeled
  - Drains all hope from the target, dealing 52 Shadow damage every 1 sec and increasing all other Shadow damage over time you deal to that target by 10%. Lasts 6 sec.

### Demonology

19 talents, 9 new. 31-point talent: Demonic Pact.

#### Row 1 (0 points)

- **Improved Health Funnel** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the amount of health transferred by your Health Funnel spell by 20%, reduces its health cost by 15%, and reduces all threat your Health Funnel generates by 50%. Allows Health Funnel to be used regardless of your demon's health.
  - Rank 2: Increases the amount of health transferred by your Health Funnel spell by 40%, reduces its health cost by 30%, and reduces all threat your Health Funnel generates by 100%. Allows Health Funnel to be used regardless of your demon's health.
  - Classic (2 ranks):
    - Rank 1: Increases the amount of Health transferred by your Health Funnel spell by 10%.
    - Rank 2: Increases the amount of Health transferred by your Health Funnel spell by 20%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Imp** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the damage of your Imp's Firebolt spell by 10% and the effect of its Fire Shield spell by 10%.
  - Rank 2: Increases the damage of your Imp's Firebolt spell by 20% and the effect of its Fire Shield spell by 20%.
  - Rank 3: Increases the damage of your Imp's Firebolt spell by 30% and the effect of its Fire Shield spell by 30%.
  - Classic (3 ranks):
    - Rank 1: Increases the effect of your Imp's Firebolt, Fire Shield, and Blood Pact spells by 10%.
    - Rank 2: Increases the effect of your Imp's Firebolt, Fire Shield, and Blood Pact spells by 20%.
    - Rank 3: Increases the effect of your Imp's Firebolt, Fire Shield, and Blood Pact spells by 30%.
- **Demonic Embrace** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases your total Stamina by 3%.
  - Rank 2: Increases your total Stamina by 6%.
  - Rank 3: Increases your total Stamina by 9%.
  - Rank 4: Increases your total Stamina by 12%.
  - Rank 5: Increases your total Stamina by 15%.
  - Classic (5 ranks):
    - Rank 1: Increases your total Stamina by 3% but reduces your total Spirit by 1%.
    - Rank 2: Increases your total Stamina by 6% but reduces your total Spirit by 2%.
    - Rank 3: Increases your total Stamina by 9% but reduces your total Spirit by 3%.
    - Rank 4: Increases your total Stamina by 12% but reduces your total Spirit by 4%.
    - Rank 5: Increases your total Stamina by 15% but reduces your total Spirit by 5%.
- **Unholy Power** (5 ranks, column 4) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases all damage done by your Imp, Voidwalker, Succubus, Incubus, and Felhunter pets by 2%.
  - Rank 2: Increases all damage done by your Imp, Voidwalker, Succubus, Incubus, and Felhunter pets by 4%.
  - Rank 3: Increases all damage done by your Imp, Voidwalker, Succubus, Incubus, and Felhunter pets by 6%.
  - Rank 4: Increases all damage done by your Imp, Voidwalker, Succubus, Incubus, and Felhunter pets by 8%.
  - Rank 5: Increases all damage done by your Imp, Voidwalker, Succubus, Incubus, and Felhunter pets by 10%.
  - Classic (5 ranks):
    - Rank 1: Increases the damage done by your Voidwalker, Succubus, Incubus, and Felhunter's melee attacks by 4%.
    - Rank 2: Increases the damage done by your Voidwalker, Succubus, Incubus, and Felhunter's melee attacks by 8%.
    - Rank 3: Increases the damage done by your Voidwalker, Succubus, Incubus, and Felhunter's melee attacks by 12%.
    - Rank 4: Increases the damage done by your Voidwalker, Succubus, Incubus, and Felhunter's melee attacks by 16%.
    - Rank 5: Increases the damage done by your Voidwalker, Succubus, Incubus, and Felhunter's melee attacks by 20%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 2 (5 points)

- **Demonic Aegis** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the effectiveness of your Demon Skin and Demon Armor spells by 15%.
  - Rank 2: Increases the effectiveness of your Demon Skin and Demon Armor spells by 30%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Voidwalker** (3 ranks, column 2) [unchanged]
  - Rank 1: Increases the effectiveness of your Voidwalker's Torment, Consume Shadows, Sacrifice and Suffering spells by 10%.
  - Rank 2: Increases the effectiveness of your Voidwalker's Torment, Consume Shadows, Sacrifice and Suffering spells by 20%.
  - Rank 3: Increases the effectiveness of your Voidwalker's Torment, Consume Shadows, Sacrifice and Suffering spells by 30%.
- **Fel Vitality** (3 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the maximum health and Mana of your Imp, Voidwalker, Succubus, Incubus, and Felhunter by 5%, and increases your maximum Mana by 5%.
  - Rank 2: Increases the maximum health and Mana of your Imp, Voidwalker, Succubus, Incubus, and Felhunter by 10%, and increases your maximum Mana by 10%.
  - Rank 3: Increases the maximum health and Mana of your Imp, Voidwalker, Succubus, Incubus, and Felhunter by 15%, and increases your maximum Mana by 15%.
- **Demonic Energies** (2 ranks, column 4) [NEW] [ranks 2+ unverified]
  - Rank 1: You heal your pet for 8% of all spell damage you deal. When you gain Mana from Life Tap, your summoned demon gains 50% of the Mana you gain.
  - Rank 2: You heal your pet for 16% of all spell damage you deal. When you gain Mana from Life Tap, your summoned demon gains 100% of the Mana you gain.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Improved Sayaad** (3 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the effect of your Succubus' and Incubus' Lash of Pain and Soothing Kiss spells by 10%, and increases the duration of your Succubus' and Incubus' Seduction and Lesser Invisibility spells by 10%.
  - Rank 2: Increases the effect of your Succubus' and Incubus' Lash of Pain and Soothing Kiss spells by 20%, and increases the duration of your Succubus' and Incubus' Seduction and Lesser Invisibility spells by 20%.
  - Rank 3: Increases the effect of your Succubus' and Incubus' Lash of Pain and Soothing Kiss spells by 30%, and increases the duration of your Succubus' and Incubus' Seduction and Lesser Invisibility spells by 30%.
- **Demonic Sacrifice** (1 rank, column 2) [CHANGED]
  - Cost: 100 yd range, Instant
  - When activated, sacrifices your summoned Demon to enhance the opposing aspect of your power, granting you an effect that lasts 2 hrs. The effect is canceled if any Demon is summoned. Imp: Increases your Shadow damage by 15%. Voidwalker: Restores 2% of your total Mana every 4 sec. Succubus/Incubus: Increases your Fire damage by 15%. Felhunter: Restores 3% of your total Health every 4 sec.
  - Classic: When activated, sacrifices your summoned demon to grant you an effect that lasts 30 min. The effect is canceled if any Demon is summoned.

Imp: Increases your Fire damage by 15%.

Voidwalker: Restores 3% of total Health every 4 sec.

Succubus/Incubus: Increases your Shadow damage by 15%.

Felhunter: Restores 2% of total Mana every 4 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Master Summoner** (2 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the casting time of your Imp, Voidwalker, Succubus, Incubus, and Felhunter Summoning spells by 2 sec and the Mana cost by 20%.
  - Rank 2: Reduces the casting time of your Imp, Voidwalker, Succubus, Incubus, and Felhunter Summoning spells by 4 sec and the Mana cost by 40%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Decimation** (2 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Reduces the cooldown of your Soul Fire spell by 45%. When you cast Shadow Bolt or Searing Pain on an enemy below 35% health, they deal 3% increased damage, and for the next 10 sec your Soul Fire spell has its cast time reduced by 20% and costs no Soul Shards.
  - Rank 2: Reduces the cooldown of your Soul Fire spell by 90%. When you cast Shadow Bolt or Searing Pain on an enemy below 70% health, they deal 6% increased damage, and for the next 20 sec your Soul Fire spell has its cast time reduced by 40% and costs no Soul Shards.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Fel Domination** (1 rank, column 3) [CHANGED]
  - Requires: Master Summoner (2/2)
  - Cost: Instant, 5 min cooldown
  - Your next Imp, Voidwalker, Succubus, Incubus, or Felhunter Summon spell has its casting time reduced by 5.5 sec and its Mana cost reduced by 50%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Demonic Brand** (3 ranks, column 4) [NEW] [ranks 2+ unverified]
  - Rank 1: Your Searing Pain generates 17% less threat and brands the target for 10 sec. Your pet's next 2 attacks against the target generate high threat and deal 39 to 42 Fire or Shadow damage based on the pet.
  - Rank 2: Your Searing Pain generates 34% less threat and brands the target for 20 sec. Your pet's next 4 attacks against the target generate high threat and deal 78 to 84 Fire or Shadow damage based on the pet.
  - Rank 3: Your Searing Pain generates 51% less threat and brands the target for 30 sec. Your pet's next 6 attacks against the target generate high threat and deal 117 to 126 Fire or Shadow damage based on the pet.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 5 (20 points)

- **Improved Felhunter** (3 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the Attack Power reduction of your Felhunter's Tainted Blood, the healing of its Devour Magic, and the detection level of its Paranoia by 10%, and reduces the cooldown of its Spell Lock by 2 sec.
  - Rank 2: Increases the Attack Power reduction of your Felhunter's Tainted Blood, the healing of its Devour Magic, and the detection level of its Paranoia by 20%, and reduces the cooldown of its Spell Lock by 4 sec.
  - Rank 3: Increases the Attack Power reduction of your Felhunter's Tainted Blood, the healing of its Devour Magic, and the detection level of its Paranoia by 30%, and reduces the cooldown of its Spell Lock by 6 sec.
- **Soul Link** (1 rank, column 2) [CHANGED]
  - Requires: Demonic Sacrifice (1/1)
  - Cost: 173 Mana, 100 yd range, Instant
  - When active, 30% of all damage taken by the caster is taken by your Imp, Voidwalker, Succubus, Incubus, or Felhunter Demon instead. In addition, both the Demon and the master will inflict 3% more damage. Lasts as long as the Demon is active.
  - Classic: When active, 30% of all damage taken by the caster is taken by your Imp, Voidwalker, Succubus, Incubus, or Felhunter demon instead. In addition, both the demon and master will inflict 3% more damage. Lasts as long as the demon is active.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Demonic Knowledge** (3 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases your spell damage and healing by up to 33% of your level while you have a summoned Demon pet active.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Note: Rank 1 reads 33% of your level in the footage, the earlier 3% came from a screenshot misread. Ranks 2-3 not yet seen.

#### Row 6 (25 points)

- **Master Demonologist** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Requires: Fel Domination (1/1)
  - Rank 1: Grants both the Warlock and the summoned demon an effect as long as that demon is active. Imp - Increases Fire damage done by 2%. Voidwalker - Reduces Physical damage taken by 2%. Succubus/Incubus - Increases Shadow damage done by 2%. Felhunter - Reduces Magic damage taken by 2%.
  - Rank 2: Grants both the Warlock and the summoned demon an effect as long as that demon is active. Imp - Increases Fire damage done by 4%. Voidwalker - Reduces Physical damage taken by 4%. Succubus/Incubus - Increases Shadow damage done by 4%. Felhunter - Reduces Magic damage taken by 4%.
  - Rank 3: Grants both the Warlock and the summoned demon an effect as long as that demon is active. Imp - Increases Fire damage done by 6%. Voidwalker - Reduces Physical damage taken by 6%. Succubus/Incubus - Increases Shadow damage done by 6%. Felhunter - Reduces Magic damage taken by 6%.
  - Rank 4: Grants both the Warlock and the summoned demon an effect as long as that demon is active. Imp - Increases Fire damage done by 8%. Voidwalker - Reduces Physical damage taken by 8%. Succubus/Incubus - Increases Shadow damage done by 8%. Felhunter - Reduces Magic damage taken by 8%.
  - Rank 5: Grants both the Warlock and the summoned demon an effect as long as that demon is active. Imp - Increases Fire damage done by 10%. Voidwalker - Reduces Physical damage taken by 10%. Succubus/Incubus - Increases Shadow damage done by 10%. Felhunter - Reduces Magic damage taken by 10%.
  - Classic (5 ranks):
    - Rank 1: Grants both the Warlock and the summoned demon an effect as long as that demon is active.

Imp - Reduces threat caused by 4%.

Voidwalker - Reduces physical damage taken by 2%.

Succubus/Incubus - Increases all damage caused by 2%.

Felhunter - Increases all resistances by .2 per level.
    - Rank 2: Grants both the Warlock and the summoned demon an effect as long as that demon is active.

Imp - Reduces threat caused by 8%.

Voidwalker - Reduces physical damage taken by 4%.

Succubus/Incubus - Increases all damage caused by 4%.

Felhunter - Increases all resistances by .4 per level.
    - Rank 3: Grants both the Warlock and the summoned demon an effect as long as that demon is active.

Imp - Reduces threat caused by 12%.

Voidwalker - Reduces physical damage taken by 6%.

Succubus/Incubus - Increases all damage caused by 6%.

Felhunter - Increases all resistances by .6 per level.
    - Rank 4: Grants both the Warlock and the summoned demon an effect as long as that demon is active.

Imp - Reduces threat caused by 16%.

Voidwalker - Reduces physical damage taken by 8%.

Succubus/Incubus - Increases all damage caused by 8%.

Felhunter - Increases all resistances by .8 per level.
    - Rank 5: Grants both the Warlock and the summoned demon an effect as long as that demon is active.

Imp - Reduces threat caused by 20%.

Voidwalker - Reduces physical damage taken by 10%.

Succubus/Incubus - Increases all damage caused by 10%.

Felhunter - Increases all resistances by 1 per level.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 7 (30 points)

- **Demonic Pact** (1 rank, column 2) [NEW]
  - Requires: Soul Link (1/1)
  - Your Demonic Sacrifice effect is no longer cancelled by summoning a different Demon pet. Resummoning the sacrificed pet will still cancel the effect.
  - Note: Capstone slot per the third-party dataset.

### Destruction

16 talents, 6 new. 31-point talent: Incinerate.

#### Row 1 (0 points)

- **Destructive Reach** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the range of your damaging spells by 10%.
  - Rank 2: Increases the range of your damaging spells by 20%.
  - Classic (2 ranks):
    - Rank 1: Increases the range of your Destruction spells by 10%.
    - Rank 2: Increases the range of your Destruction spells by 20%.
- **Improved Shadow Bolt** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Your Shadow Bolt critical strikes increase Shadow damage taken by the target from your attacks by 4% for 12 sec.
  - Rank 2: Your Shadow Bolt critical strikes increase Shadow damage taken by the target from your attacks by 8% for 24 sec.
  - Rank 3: Your Shadow Bolt critical strikes increase Shadow damage taken by the target from your attacks by 12% for 36 sec.
  - Rank 4: Your Shadow Bolt critical strikes increase Shadow damage taken by the target from your attacks by 16% for 48 sec.
  - Rank 5: Your Shadow Bolt critical strikes increase Shadow damage taken by the target from your attacks by 20% for 60 sec.
  - Classic (5 ranks):
    - Rank 1: Your Shadow Bolt critical strikes increase Shadow damage dealt to the target by 4% until 4 non-periodic damage sources are applied. Effect lasts a maximum of 12 sec.
    - Rank 2: Your Shadow Bolt critical strikes increase Shadow damage dealt to the target by 8% until 4 non-periodic damage sources are applied. Effect lasts a maximum of 12 sec.
    - Rank 3: Your Shadow Bolt critical strikes increase Shadow damage dealt to the target by 12% until 4 non-periodic damage sources are applied. Effect lasts a maximum of 12 sec.
    - Rank 4: Your Shadow Bolt critical strikes increase Shadow damage dealt to the target by 16% until 4 non-periodic damage sources are applied. Effect lasts a maximum of 12 sec.
    - Rank 5: Your Shadow Bolt critical strikes increase Shadow damage dealt to the target by 20% until 4 non-periodic damage sources are applied. Effect lasts a maximum of 12 sec.
- **Bane** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the casting time of your Shadow Bolt, Immolate, and Incinerate spells by 0.1 sec and your Soul Fire spell by 0.4 sec.
  - Rank 2: Reduces the casting time of your Shadow Bolt, Immolate, and Incinerate spells by 0.2 sec and your Soul Fire spell by 0.8 sec.
  - Rank 3: Reduces the casting time of your Shadow Bolt, Immolate, and Incinerate spells by 0.3 sec and your Soul Fire spell by 1.2 sec.
  - Rank 4: Reduces the casting time of your Shadow Bolt, Immolate, and Incinerate spells by 0.4 sec and your Soul Fire spell by 1.6 sec.
  - Rank 5: Reduces the casting time of your Shadow Bolt, Immolate, and Incinerate spells by 0.5 sec and your Soul Fire spell by 2 sec.
  - Classic (5 ranks):
    - Rank 1: Reduces the casting time of your Shadow Bolt and Immolate spells by 0.1 sec and your Soul Fire spell by 0.4 sec.
    - Rank 2: Reduces the casting time of your Shadow Bolt and Immolate spells by 0.2 sec and your Soul Fire spell by 0.8 sec.
    - Rank 3: Reduces the casting time of your Shadow Bolt and Immolate spells by 0.3 sec and your Soul Fire spell by 1.2 sec.
    - Rank 4: Reduces the casting time of your Shadow Bolt and Immolate spells by 0.4 sec and your Soul Fire spell by 1.6 sec.
    - Rank 5: Reduces the casting time of your Shadow Bolt and Immolate spells by 0.5 sec and your Soul Fire spell by 2 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 2 (5 points)

- **Molten Skin** (5 ranks, column 1) [NEW] [ranks 2+ unverified]
  - Rank 1: Reduces all damage taken by 2%.
  - Rank 2: Reduces all damage taken by 4%.
  - Rank 3: Reduces all damage taken by 6%.
  - Rank 4: Reduces all damage taken by 8%.
  - Rank 5: Reduces all damage taken by 10%.
  - Note: Rank 1 (2%) and rank 5 (10%) seen on screen, 2% per rank.
- **Cataclysm** (3 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Reduces the Mana cost of your Destruction spells by 3%.
  - Rank 2: Reduces the Mana cost of your Destruction spells by 6%.
  - Rank 3: Reduces the Mana cost of your Destruction spells by 9%.
  - Classic (5 ranks):
    - Rank 1: Reduces the Mana cost of your Destruction spells by 1%.
    - Rank 2: Reduces the Mana cost of your Destruction spells by 2%.
    - Rank 3: Reduces the Mana cost of your Destruction spells by 3%.
    - Rank 4: Reduces the Mana cost of your Destruction spells by 4%.
    - Rank 5: Reduces the Mana cost of your Destruction spells by 5%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Aftermath** (5 ranks, column 3) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the initial damage of your Immolate spell by 10% and your Conflagrate spell has a 20% chance to Daze the target, reducing the target's movement speed by 50% for 5 sec.
  - Rank 2: Increases the initial damage of your Immolate spell by 20% and your Conflagrate spell has a 40% chance to Daze the target, reducing the target's movement speed by 100% for 10 sec.
  - Rank 3: Increases the initial damage of your Immolate spell by 30% and your Conflagrate spell has a 60% chance to Daze the target, reducing the target's movement speed by 150% for 15 sec.
  - Rank 4: Increases the initial damage of your Immolate spell by 40% and your Conflagrate spell has a 80% chance to Daze the target, reducing the target's movement speed by 200% for 20 sec.
  - Rank 5: Increases the initial damage of your Immolate spell by 50% and your Conflagrate spell has a 100% chance to Daze the target, reducing the target's movement speed by 250% for 25 sec.
  - Classic (5 ranks):
    - Rank 1: Gives your Destruction spells a 2% chance to daze the target for 5 sec.
    - Rank 2: Gives your Destruction spells a 4% chance to daze the target for 5 sec.
    - Rank 3: Gives your Destruction spells a 6% chance to daze the target for 5 sec.
    - Rank 4: Gives your Destruction spells a 8% chance to daze the target for 5 sec.
    - Rank 5: Gives your Destruction spells a 10% chance to daze the target for 5 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Ruin** (5 ranks, column 2) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike damage bonus of your Destruction spells by 20%.
  - Rank 2: Increases the critical strike damage bonus of your Destruction spells by 40%.
  - Rank 3: Increases the critical strike damage bonus of your Destruction spells by 60%.
  - Rank 4: Increases the critical strike damage bonus of your Destruction spells by 80%.
  - Rank 5: Increases the critical strike damage bonus of your Destruction spells by 100%.
  - Classic: Increases the critical strike damage bonus of your Destruction spells by 100%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Shadowburn** (1 rank, column 3) [CHANGED]
  - Cost: 105 Mana, 30 yd range, Instant, 15 sec cooldown, Reagents: Soul Shard
  - Instantly blasts the target for 102 to 111 Shadow damage. If a non-trivial target dies within 8 sec of being hit with Shadowburn, the caster gains a Soul Shard.
  - Classic: Instantly blasts the target for 92 to 104 Shadow damage. If the target dies within 5 sec of Shadowburn, and yields experience or honor, the caster gains a Soul Shard.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Intensity** (3 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Rank 1: Gives you a 23% chance to resist interruption caused by damage while casting or channeling any Destruction spell.
  - Rank 2: Gives you a 46% chance to resist interruption caused by damage while casting or channeling any Destruction spell.
  - Rank 3: Gives you a 69% chance to resist interruption caused by damage while casting or channeling any Destruction spell.
  - Classic (2 ranks):
    - Rank 1: Gives you a 35% chance to resist interruption caused by damage while channeling the Rain of Fire, Hellfire or Soul Fire spell.
    - Rank 2: Gives you a 70% chance to resist interruption caused by damage while channeling the Rain of Fire, Hellfire or Soul Fire spell.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Agonizing Flames** (3 ranks, column 2) [NEW] [ranks 2+ unverified]
  - Rank 1: Increases the critical strike chance of your Searing Pain spell by 3% and the damage done by all your Destruction spells by 3%.
  - Rank 2: Increases the critical strike chance of your Searing Pain spell by 6% and the damage done by all your Destruction spells by 6%.
  - Rank 3: Increases the critical strike chance of your Searing Pain spell by 9% and the damage done by all your Destruction spells by 9%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Conflagrate** (1 rank, column 3) [CHANGED]
  - Requires: Shadowburn (1/1)
  - Cost: 100 Mana, 30 yd range, Instant, 10 sec cooldown
  - Ignites a target that is already afflicted by your Immolate spell, dealing 109 to 132 Fire damage and consuming your Immolate effect.
  - Classic: Ignites a target that is already afflicted by Immolate, dealing 250 to 316 Fire damage and consuming the Immolate spell.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 5 (20 points)

- **Pyroclasm** (2 ranks, column 1) [CHANGED] [ranks 2+ unverified]
  - Requires: Intensity (3/3)
  - Rank 1: Gives your Soul Fire spell a 13% chance to Stun the target for 3 sec, and your Rain of Fire and Hellfire spells a 13% chance over their duration to Stun targets they damage for 3 sec.
  - Rank 2: Gives your Soul Fire spell a 26% chance to Stun the target for 6 sec, and your Rain of Fire and Hellfire spells a 26% chance over their duration to Stun targets they damage for 6 sec.
  - Classic (2 ranks):
    - Rank 1: Gives your Rain of Fire, Hellfire, and Soul Fire spells a 13% chance to stun the target for 3 sec.
    - Rank 2: Gives your Rain of Fire, Hellfire, and Soul Fire spells a 26% chance to stun the target for 3 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Bane of Havoc** (1 rank, column 2) [NEW]
  - Cost: 43 Mana, 30 yd range, Instant
  - Afflicts the target for 5 min, causing 15% of all damage done by the Warlock to other targets to also be dealt to the cursed target. Bane of Havoc is limited to 1 target, and only one Bane per Warlock can be active on any one target.
  - Note: 43 Mana, 30 yd range, instant. Slot per the third-party dataset.
- **Fire and Brimstone** (3 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Requires: Conflagrate (1/1)
  - Rank 1: Increases the critical strike chance of your Conflagrate spell by 8%.
  - Rank 2: Increases the critical strike chance of your Conflagrate spell by 17%.
  - Rank 3: Increases the critical strike chance of your Conflagrate spell by 25%.
  - Note: Rank 1 (8%) and rank 3 (25%) seen on screen, rank 2 interpolated.

#### Row 6 (25 points)

- **Shadow and Flame** (5 ranks, column 3) [NEW] [ranks 2+ unverified]
  - Rank 1: Hitting an enemy with Conflagrate increases all Shadow damage you deal by 2% for 20 sec, and hitting an enemy with Shadowburn increases all Fire damage you deal by 2% for 20 sec. In addition, Conflagrate has a 20% chance not to consume Immolate, and Shadowburn has a 20% chance to instantly refund a Soul Shard.
  - Rank 2: Hitting an enemy with Conflagrate increases all Shadow damage you deal by 4% for 20 sec, and hitting an enemy with Shadowburn increases all Fire damage you deal by 4% for 20 sec. In addition, Conflagrate has a 40% chance not to consume Immolate, and Shadowburn has a 40% chance to instantly refund a Soul Shard.
  - Rank 3: Hitting an enemy with Conflagrate increases all Shadow damage you deal by 6% for 20 sec, and hitting an enemy with Shadowburn increases all Fire damage you deal by 6% for 20 sec. In addition, Conflagrate has a 60% chance not to consume Immolate, and Shadowburn has a 60% chance to instantly refund a Soul Shard.
  - Rank 4: Hitting an enemy with Conflagrate increases all Shadow damage you deal by 8% for 20 sec, and hitting an enemy with Shadowburn increases all Fire damage you deal by 8% for 20 sec. In addition, Conflagrate has a 80% chance not to consume Immolate, and Shadowburn has a 80% chance to instantly refund a Soul Shard.
  - Rank 5: Hitting an enemy with Conflagrate increases all Shadow damage you deal by 10% for 20 sec, and hitting an enemy with Shadowburn increases all Fire damage you deal by 10% for 20 sec. In addition, Conflagrate has a 100% chance not to consume Immolate, and Shadowburn has a 100% chance to instantly refund a Soul Shard.
  - Note: Rank 4 and 5 confirmed on screen, ranks 2 and 3 follow the same 2% / 20% steps. Duration is 20 sec at every rank.

#### Row 7 (30 points)

- **Incinerate** (1 rank, column 2) [NEW]
  - Requires: Bane of Havoc (1/1)
  - Cost: 205 Mana, 30 yd range, 2.5 sec cast
  - Deals 125 to 140 Fire damage to your target and an additional 25% damage if the target is afflicted by Immolate.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
