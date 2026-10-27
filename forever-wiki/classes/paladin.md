# Paladin

Source: wowforevertalents.com dataset (BlizzCon 2026 footage overlaid on Classic Era 1.15.9 client tables), data version 2026-09-13/14. See [../README.md](../README.md) for caveats. Numbers on new talents were read off level 38 demo characters, so absolute damage and mana values are not level 60 values.

Talents: 52 total, 20 new, 21 changed, 11 unchanged (verified).

Row N needs 5*(N-1) points spent in that tree. Row 1 is the top row. Rows 3, 4, 5 and 7 hold the 11, 16, 21 and 31 point talents.

Warning on multi-rank NEW and CHANGED talents (now settled for every talent in the 70338 tree, see Trait curves below): the footage showed rank 1 only. The dataset scaled the other ranks linearly from rank 1, which sometimes produces nonsense (for example Maelstrom Weapon rank 5 reading 'stacks up to 25 times'). Treat rank 1 as observed and the rest as a guess until the beta client is out. The [ranks 2+ unverified] tag marks every multi-rank new or changed talent. Read the Note line, it says when a higher rank was actually seen on a BlizzCon slide.

## Client build 1.60.1.70338 (checked 2026-10-12)

Every retribution spell and talent checked against the client tables on [wago.tools](https://wago.tools/db2) (SpellEffect, SpellMisc, SpellPower, SpellCooldowns, SpellLevels, SpellAuraOptions, SpellClassOptions, ItemSet, ItemSetSpell for build 1.60.1.70338) and the talent and patch note pages on foreverchanges.pro. The seals, judgements, Exorcism, Hammer of Wrath, Holy Wrath, Holy Shock, Consecration, the blessings and Retribution Aura (Vanilla's 5 / 8 / 12 / 16 / 20 plus 6% of spell power since the 2026-10-08 build) already matched the sim. What the client showed:

- Holy Strike is built like Backstab: a normalized weapon damage effect that carries the flat Holy part (93 at rank 8, a 25% spread, 42.9% of spell power), then a weapon damage percent (50% at rank 8). The server adds the flat part before it takes the percent, so the flat part and its spell power are halved too, though the tooltip shows "plus 93". The normalized speed is Forever's 3.6 for a two-hander (2.7 for a one-hander), as the beta's Stormstrike showed for the same effect.
- Improved Seals is a damage modifier on the seal procs and judgements themselves, so it scales the spell power part of Seal of Righteousness too. Its spell carries 3%, the talent tooltips 5 / 10 / 15%.
- Holy Power's 1% per point is a crit modifier on a list of spells: Consecration, Exorcism, Holy Wrath, Hammer of Wrath, both seal procs and both judgements (Holy Shock and Holy Strike get 3%). Corrected 2026-10-23: the list (effect 0's mask) has the Seal of Command proc 20424 but not the Seal of Righteousness procs (25713-25742), see the 2026-10-23 section.
- Precision gives spell hit as well as melee hit.
- Benediction's spell list leaves out Hammer of Wrath, Twist of Light's covers the three seals.
- Twist of Light's Echo is two spells, Echo of Command (1311703, "a chance to activate Seal of Command") and Echo of Righteousness (1311704), one charge each and no duration, used up by the next auto attack.
- Reckoning can trigger once every 1.5 sec, Holy Shield blocks 30%, Redoubt is 4% a point. Improved Seal of Fury's mana grows with the paladin's level.
- Justice Battlegear (item set 2106) is the Tier 1 retribution set: head, shoulders, chest, legs, hands and feet in plate. 2 pieces: 1% attack and casting speed. 3: 5 sec off Hammer of Justice. 4: 36 attack power against Undead. 5: 0.5 sec off Judgement. The pieces have no stats in the client yet. Justice Armor (Holy, 2107) and Justice Battleplate (Protection, 2108) are the other two: 2 pieces +10 Spirit / +7 Defense, 3 a utility cooldown (Blessing of Protection / Turn Undead), 4 +26 healing and +9 spell damage / 12 expertise rating (1.2%), 5 1 sec off Holy Shock / 10 sec off Forbearance. The sim registers all three by their set IDs (since 2026-10-23, as wowsims/forever's item_sets.go has them); the synthetic Phase 1 set keeps Justice Battlegear's bonuses.

## Beta build 1.60.1.70009 (2026-09-24)

The [development notes of 24 September](https://www.wowhead.com/blue-tracker/topic/us/wow-forever-beta-development-notes-updated-september-24-2360696), with the numbers checked against the client tooltips on foreverchanges.pro (build 1.60.1.70009). The sim follows all of it except Retribution Aura's spell power scaling.

- Holy Strike is 25 / 29 / 32 / 36 / 39 / 43 / 46 / 50% weapon damage by rank (was 25 / 25 / 30 / 30 / 35 / 35 / 40 / 40) and its cooldown is 10 sec baseline (was 12). The flat Holy part is unchanged.
- Improved Holy Strike is gone from the Holy tree and Crusade from the Retribution tree (both checked in game). Their slots stay empty, so every Holy talent string loses its first digit and the Retribution strings shift after Sacred Arbiter.
- Holy Power now also gives Holy Strike 3% crit per point.
- Sacred Arbiter is +20% Holy Strike damage (was 10%).
- Vengeance stacks 3 times (was 5), and only a non-periodic crit adds a stack.
- Two-Handed Weapon Specialization is 2 / 4 / 6% (was 3 / 6 / 9%).
- Twist of Light also takes 20% off the Mana cost of the seals.
- Retribution Aura scales with the caster's spell power. The coefficient is unknown. The rank 1 tooltip stayed at 7 damage with 18 spell power on, so it is small, and the sim keeps the flat value.
- Righteous Fury is 60% Holy threat (was 90%). Client 70425 put it back to 90%.
- Consecration had targeting fixes, and Echo of Justice (the Twist of Light echo) had its proc chance fixed. Neither changes the sim.

## Client 1.60.1.70425 (2026-10-16)

Diffed against 70338 for every paladin spell and trait curve (wago.tools SpellEffect, SpellLevels and TraitDefinitionEffectPoints). The sim takes all of it:

- Judgement of Righteousness is 42.9% of spell power on every rank (was 50%), and Judgement of Fury 42.9% (was 45%).
- Champion of the Light is 10 / 20 / 30% of Intellect as spell damage (trait curve 110350, was 20 / 40 / 60%). The 2026-10-20 server hotfix put it back to 20 / 40 / 60%, see the 2026-10-23 section.
- Consecration's Forever ranks (1280345 to 1280353) are 4.2% of spell power a tick (was 9.5%), Era's value. The 15 October beta notes give 0.042 too, and say Judgement of Righteousness and Judgement of Fury now have "correct spell power coefficients for instant cast spells".
- Righteous Fury is 90% Holy threat again (60% in 70338).
- Shield Specialization's mana is 4% of maximum mana a block (1310925, was 6%).
- Reckoning stores its extra attacks as charges, up to 4 (new spell 1326084), and the next auto attack spends them all; 70338 swung at once (trigger 20178).
- Seal of Fury rank 1 (20163) is learned at level 26 (was 25). Seal of Fury is not in the sim.
- The Seal of Righteousness proc spells (25713 and the lower ranks) lost their 10% spell power coefficient (0 in 70425). The seal's tooltip (auras 20154, 20287 to 20293, 21084) still adds `$1h` = 7.65% and `$2h` = 12% of spell power (description variables 1023) to each hit's s3/87 to s3/25, so the bonus reads as moved server side. Blizzard's 15 October beta notes confirm it: "Fixed an issue causing Seal of Righteousness to gain almost twice as much benefit from Spell Damage than intended." The sim adds it once: 12% with a two-hander, flat per swing (10% with a one-hander since 2026-10-23, see below). It used to count 2 x 11%, the Season of Discovery tests; that would be 50 DPS more on twist at 180 sec, and the proc's 0 would be 57 less.
- Improved Seals stores 0 on its spell (was 3); its trait curve is still 5 / 10 / 15%.
- Blessing of Salvation is now Salvation (1038), a threat buff outside the blessing slot, and Greater Blessing of Salvation is Mass Salvation (25895). No DPS change.

## Aligned with wowsims/forever, and the 2026-10-20 hotfix (2026-10-23)

Where wowsims/forever has in-game measurements or the game files on its side, the sim now follows it (all under the Forever ruleset):

- Judgement of Righteousness is a melee attack (client 20286 and the lower ranks are DefenseType 2 with No Active Defense): it rolls melee hit and crit, crits for double, can't be dodged, parried or blocked, and One- and Two-Handed Weapon Specialization raise it (wowsims/forever e10ca32c3d, d35afca4da). It used to roll spell hit and crit at 1.5x. As a main hand ability it now hears weapon procs, Windfury Totem and Vindication, as Judgement of Command already did; white-hit-only effects still don't.
- Judgement of Command rolls once on the melee table (20968 and 20966 are both melee), not spell hit first and then melee crit (e10ca32c3d, 447d83b595).
- Sanctified Judgement refunds its share of the seal's base cost with the seal's percent cut added on top: Benediction 5/5 makes a 90 mana seal cost 81 and 3/3 return 60% of 99 = 59.4, measured at Benediction 0, 1, 3 and 5 (04b40b17b8). It used to return a share of the discounted cost. With Twist of Light's 20% the two cuts multiply before the flip (Seal of Righteousness rank 8: 200 x 1.28 x 0.6 = 153.6), as theirs has it; that part is unmeasured.
- Seal of Righteousness with a one-hander: 10% of spell power (the tooltip's 7.65% fits none of their 1.8, 2.2 and 2.8 speed hits) and a weapon factor of 0.85 - 0.2 x max(0, 2.0 - speed), so 0.83 at 1.9 and 0.79 at 1.7 (18 weapons from 1.7 to 3.7, 49f94c9326). Two-handers keep 1.2 and 12%.
- Holy Power follows effect 0's mask on 5923: Consecration's ticks get it (they crit under Forever, 2f953756bd and fd449cd4eb), the Seal of Righteousness procs don't.
- Holy Shield's damage takes 8% of spell power (20925, 20927 and 20928, effect 1; it was Classic's 5%). Its outcome stays a spell crit roll that can't miss: theirs rolls spell hit with no crit but cites nothing.
- Vindication procs 5 times a minute on the main hand's speed (Seal of the Crusader's faster swings each carry 1/1.4 of the chance), measured at 1.6 and 3.6 speed (54edb1d2e2); it used to proc on every landed melee hit. The Seal of Righteousness proc doesn't trigger it (a proc, which the talent can't proc from). Its debuff's own spell hit roll isn't modelled because the debuff isn't scored.
- Consecration's extra damage and its 4.2% of spell power go to the first 4 enemies only (the rank's periodic dummy names 4), and so does the mark (client 1280353) that carries Consecrated Ground. Later enemies take the 12 a tick base alone (theirs consecration.go).
- Hammer of Wrath keeps Precision (20189 is an unscoped A_MOD_HIT_CHANCE) and gets no Conviction (20117 is melee weapon crit), as theirs conviction_test.go has it.
- Librams and gloves (2026-10-23, as theirs item_librams.go and item_armor.go have them): Tenets of the Silver Hand (249397, spell 1302545, A_MOD_DAMAGE_DONE_VERSUS on mask 32) is 1% more damage against Undead, damage only. Libram of Invocation (249442) takes 5% off the seals, multiplied with Benediction and Twist of Light, and Sanctified Judgement's refund flips that cut too (Seal of Righteousness rank 8 with Benediction 5/5 and Twist of Light: 136.8 mana, 157.92 back). The Judgement of the Crusader PvP gloves (16410, 16471, 23274; spell 23300) add 20 to the judgement's Holy damage taken.
- Twist of Light keeps one Echo (2026-10-23): a new Echo replaces the one that was up, and casting a seal drops its own Echo (theirs a9392a4b5c, e7f7cad34e, a modelling reading). The tooltip (1310735) has one Echo that "your next melee attack" consumes, which we read the same way. It only matters when a third seal comes between the twists: the Crusader duty rotations lose 3.0 (twist) and 2.7 (sub Holy) at 180 sec, the other presets nothing.
- Champion of the Light is 20 / 40 / 60% again: the server hotfix pushed on 2026-10-20 to build 70450 (wago.tools hotfix push 112842, CurvePoint 331369-331371) sets curve 110350 to 20 / 40 / 60. That curve is Champion of the Light's (TraitDefinition 142634, spell 1311084, TraitDefinitionEffectPoints 24621). The 70532 client tables still read 10 / 20 / 30, as they still have the old mage tree from the same push. foreverchanges.pro's 2026-10-16 hotfix list says the same.

At 180 sec (20,000 iterations, seed 1), the Level 60 presets: twist 877.2 -> 935.2, sub Holy 871.1 -> 935.0, Shock 780.3 -> 827.7, and the undead and Crusader duty variants alike. Champion of the Light is +34 to +39 on the builds that take it, Judgement of Righteousness on the melee table +22 to +25 (+40 to +44 on Shock, which judges Righteousness), Judgement of Command +4 to +7, Hammer of Wrath -1.5 to -2, Consecration's Holy Power +2 on Shock, the Seal of Righteousness proc's lost Holy Power -2.6 on Shock, Sanctified Judgement's larger refund -1.7 to +5.0 (noise, except +1.9 and +5.0 on the mana-short Shock builds) and Vindication under 0.5.

## Regear and rotation round (2026-10-23)

With no melee hit suppression under Forever the yellow cap against a level 63 boss is 8% (standards 4), and the Level 60 presets were re-tuned against it (20,000 iterations, seeds 1 and 1000003, paired at 120 / 180 / 300 sec):

- Gear: the shoulders' 10 hit rating went to 5 Strength and 5 crit rating (80 rating, 8% from gear), +2.3 to +2.8 on every build. Precision now puts the twist build's melee hit 3% over the cap (it still adds spell hit).
- Talents: the sub Holy and both vs Undead builds take `550030022--05225331001330321` (no Protection points: Divine Strength 5, Divine Intellect 5, Improved Seals 3, Reverence 2, Purifying Power 2, Holy Conduit 2): vs Undead +29.0 at 180 sec, sub Holy vs Undead +20.7, sub Holy +1.0 (+12.6 at 300). Twist and Shock keep theirs.
- Rotations: Light's Vigil waits until the next swing is more than 2 sec off, since its 1.5 sec cast resets the swing when it completes (+35 / +40 / +32 at 180 sec on Shock / Shock vs Undead / Shock duty); the vs Undead rotation drops Holy Wrath (2 sec cast, +6.8); sub Holy keeps Seal of Righteousness up, twisting to Seal of Command (+0.5 to +1.1); the twist goes on at any mana (+0.1 at 180 sec, +5.7 at 300).
- Relics and consumables: Libram of Fervor (+33 to Judgement of the Crusader's Holy damage taken, a world drop) for the Crusader duty builds (+5.9 twist, +7.7 Shock); Libram of Law stays elsewhere (Libram of Invocation is -2.1 / -2.3 on twist / sub Holy at 180 sec, +3 and +2 at 300). The Shock builds take the Major Mana Potion (+12.8 to +18.4 over the Dragonfire Potion), the twist builds keep the Dragonfire Potion. The Goblin Land Mine costs 1 to 6 on every build and stays (standards 11).
- Human stays the best race on every build (2.4 to 6.0 over Undead at 180 sec). At 180 sec: sub Holy 944.5 (the class default now), twist 943.3, Shock 865.5, vs Undead 1047.8, Shock vs Undead 965.2.

### Must-take talent and explosives (2026-10-24)

- Pursuit of Justice 2 (movement speed) is a must-take talent on every Level 60 build (standards 26). We put it in each string by hand and re-ran the talent searches at 180 sec with it pinned (`tools/rotopt -require pursuitOfJustice`), then paired the results at 120 / 180 / 300 sec. What it costs: twist -3.7 / -3.4 / -3.5 (Precision 3 to 2, Vindication 3 to 2: `51003-502-05225231201330321`); sub Holy nothing (it takes Purifying Power's 2 points, which do nothing on the generic boss: `55003002--05225331201330321`); Shock -2.3 / -2.3 / -2.1 (Vindication 2 to 0: `55003003201121251--05225030201`, the same on Shock vs Undead and the Shock duty); sub Holy vs Undead -4.1 at each length (Vindication 3 to 1: `550030022--05225131201330321`; the search's Reverence and Vindication point each was 2 ahead at 120 and 180 sec but 9 behind at 300). The duty builds keep their base talents. The twist duty search moved to the sub Holy side (4.6 ahead at 180 sec), the same DPS as the sub Holy duty build.
- Level 60 (vs Undead) is gone: it had the same talents, gear and rotation as Level 60 (sub Holy, vs Undead), and the search from the twist talents ended 8 behind on the Protection side.
- Level 60 (Crusader duty), the twist duty copy, is gone too (user ruling 2026-10-24): it tied Level 60 (sub Holy, Crusader duty), which keeps the duty for the twisting playstyle.
- No filler explosive (standards 11, only when it gains): no explosive is 4.9 to 7.0 ahead of the Goblin Land Mine and the EZ-Thro Dark Bomb (the same 450 average Fire damage) at 180 sec, and 5.7 to 7.7 ahead of the EZ-Thro Thorium Grenade. Each one's 1 sec cast resets the swing. The Goblin Sapper Charge (no cast) stays, +4.4 to 5.1.
- At 180 sec (seed 1, seed 1000003 within 0.7): sub Holy 949.1 (the class default), twist 945.1, Shock 869.0, sub Holy vs Undead 1050.8, Shock vs Undead 968.4.

### Holy Shock on the Retribution side (2026-10-26)

Blizzard's Paladin Class Deep Dive (15 October 2026, https://worldofwarcraft.blizzard.com/en-us/news/24302099) placed Holy Shock "as an alternative to Twist of Light ... allowing for players to opt out of Seal Twisting gameplay for a hybrid build utilizing Holy talents": Retribution can "seal twist with Twist of Light, or focus on Holy damage by hybridizing into Holy Shock". Holy Shock needs 20 Holy points, so Twist of Light (31 Retribution) is out of reach with it. Level 60 Shock (deep Holy) is the deep Holy version (Light's Vigil, 31/0/20; called Level 60 Shock until 2026-10-27), and Level 60 Shock is the Retribution one (called Level 60 Shock (Ret side) on 2026-10-26). The report calls them Ret Shock (deep Holy) and Ret Shock.

- Talents: a search at 180 sec with Holy Shock, Champion of the Light and Pursuit of Justice pinned, from a 21/0/30 seed (872.1 before the search): `550030032011012--052150312012303` (23/0/28). Holy: Divine Strength 5, Divine Intellect 5, Improved Seals 3, Reverence 3, Purifying Power 2, Illumination 1, Divine Favor, Holy Shock, Consecrated Ground 2. Retribution: Benediction 5, Improved Judgement 2, Holy Conduit 1, Conviction 5, Sanctified Judgement 3, Seal of Command, Pursuit of Justice 2, Sacred Arbiter, Two-Handed Weapon Specialization 2, Vengeance 3, Champion of the Light 3.
- Rotation (`level60_shock_ret`, the paladin_ret template searched over 120, 180 and 300 sec before and after the talent search): Seal of Command up with no twisting (+6 over Seal of Righteousness), Consecration first at any mana, Judgement, Holy Strike and Holy Shock on cooldown, Hammer of Wrath above Holy Strike.
- Consumes: the twist builds' (Dragonfire Potion): +5.5 / +5.7 / -10.3 against the Major Mana Potion at 120 / 180 / 300 sec.
- Human, 20000 iterations, seed 1, 120 / 180 / 300 sec: 939.5 / 932.4 / 910.7, against sub Holy 956.5 / 949.1 / 941.0, twist 951.6 / 945.1 / 932.8 and Shock 881.2 / 868.8 / 842.8. It is 1.8% behind sub Holy at 180 sec and 6.7% ahead of the deep Holy Shock build.

## Trait curves, client 1.60.1.70338 (2026-10-12)

Forever's talents are traits: TraitDefinitionEffectPoints gives every talent effect a per-rank curve, and the curve overrides the base points on the rank 1 spell, which can be a Classic leftover. The paladin tree (TraitTree 1100) per rank:

| Talent | Curve | Sim |
| --- | --- | --- |
| Divine Strength, Divine Intellect | 2 / 4 / 6 / 8 / 10% | same |
| Toughness | 2 to 10% armor | same |
| Benediction | -2 to -10% cost | same |
| Improved Seals | 5 / 10 / 15% (the spell stores 3) | same |
| Precision | 1 / 2 / 3% melee and spell hit | same |
| Improved Judgement | -1 / -2 sec | same |
| Holy Conduit | -20 / -40% cost | same |
| Conviction | 1 to 5% crit | same |
| Reverence | 10 / 20 / 30% regen while casting | same |
| Purifying Power | -17 / -33% Exorcism and Holy Wrath cooldown (and -10 / -20% Cleanse cost) | same (cooldown) |
| Sacred Duty | 2 / 4% Stamina | same |
| Vindication | 1 / 2 / 3% | same |
| Sanctified Judgement | 33 / 66 / 100% chance, 20 / 40 / 60% refund | same (of the seal's base cost with its cut flipped on top since 2026-10-23) |
| Divine Precision | 6 / 12 / 18% Holy hit | same |
| Consecrated Ground | 5 / 10% | same |
| Two-Handed Weapon Specialization | 2 / 4 / 6% | same |
| One-Handed Weapon Specialization | 3 / 7 / 10% | same |
| Vengeance | 1 / 2 / 3% a stack (the spell stores 5) | same |
| Holy Power | 1 to 5% crit, 3 to 15% on its second effect | same (Holy Strike and Holy Shock 3% a point) |
| Champion of the Light | 20 / 40 / 60% of Intellect (10 / 20 / 30% in 70425, 20 / 40 / 60% again by the 2026-10-20 hotfix) | 20 / 40 / 60% |
| Instrument of Law | -0.5 / -1.0 sec Hammer of Wrath cast | same |
| Iron Creed | 5 to 25% threat | same |
| Redoubt | 4 to 20% block | same |
| Reckoning | 8 to 40% (its only curve) | the sim's 8% a point on blocks; its 20% a point on crits comes from the beta notes |

No paladin value changed. Ranks 2 and up of every talent in the tree are settled by these curves; Improved Holy Strike and Crusade below are not in the 70338 tree.

## Beta client 1.60.1.69876 (2026-09-16)

The beta client's paladin tree, read through [hyjal.cc](https://hyjal.cc/updates/1.60.1.69876), against the BlizzCon data below. Positions, rank counts and prerequisites all held. Ranks 2 and up are now real numbers, and `tools/forever_talents/data/paladin.json` in the sim repo is regenerated from the client tree (`client/paladin.json`) since 2026-09-20, so the per-talent text below is superseded where it differs. The sim follows the client.

- Vengeance is 1 / 2 / 3% per stack (5 stacks, 30 sec), Redoubt 6 to 30% block, Iron Creed 5 to 25% threat and 2 to 10% damage taken, Shield Specialization 10 / 20 / 30% and 33 / 66 / 100% mana chance, Sacred Duty 2 / 4% and 30 / 60 sec, Instrument of Law 0.5 / 1.0 sec: the multi-rank new talents scale linearly where the transcription had rank 1 repeated.
- One-Handed Weapon Specialization is 3 / 7 / 10%, Champion of the Light 20 / 40 / 60% (33 / 66 / 100% before the 2026-10-01 development notes), Sanctified Judgement 33 / 66 / 100%, Pursuit of Justice 8 / 15%, Purifying Power 17 / 33%.
- Vindication takes 66.8 / 133.7 / 200.5 attack power off the target (42 in the footage, level 38).
- Consecrated Ground stays at the first 4 enemies at rank 2 (the transcription scaled it to 8). Improved Seal of Fury restores 60 mana (38 at level 38).
- Precision reads "Improves your chance to hit by 1%" (the transcription said "with all spells and attacks").

Wowhead's Forever database (2026-09-20) has the trainer abilities, see [../sim-status.md](../sim-status.md) for the numbers the sim took from it (Holy Strike ranks, Judgement, blessings, Consecration, Retribution Aura, Righteous Fury, Holy Shield, Holy Shock). Its talent tooltips show one rank each and disagree with the client on a few values (Improved Seals 3%, Vengeance 5%, One-Handed Weapon Specialization 2%), so the sim keeps the client's.

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
  - Note (sim, 2026-10-15): rank 6 is 161 Holy damage taken at level 60 (140 x 1.15). Judgement of Righteousness and Judgement of Command leave no debuff, so a Retribution paladin keeps its Judgement of the Crusader up while judging its damage seals; the Level 60 presets get it from the raid, and the "Crusader duty" presets cast it themselves (standards 23).
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
  - Note (client 1.60.1.70425): renamed **Salvation** (spell 1038): "Reduces all threat generated by target party member by 30% for 1 hour." The blessing wording is gone, so it no longer takes the paladin's blessing slot on the target and stacks with a Blessing. Mass Salvation (25895) likewise drops "Greater Blessing of Salvation". The sim doesn't model threat for DPS, so this changes no numbers.
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

- **Healing Light** (3 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases the amount healed by your Holy Light, Flash of Light, and Holy Shock spells by 4%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Classic (3 ranks):
    - Rank 1: Increases the amount healed by your Holy Light and Flash of Light spells by 4%.
    - Rank 2: Increases the amount healed by your Holy Light and Flash of Light spells by 8%.
    - Rank 3: Increases the amount healed by your Holy Light and Flash of Light spells by 12%.
  - Note: Rank 1 reads 4% in the footage (our earlier 12% was a maxed-rank screenshot). Ranks 2-3 not yet seen.
- **Spiritual Focus** (2 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Gives your Flash of Light, Holy Light, and Light's Vigil spells a 70% chance to not lose casting time when you take damage.
  - Rank 2: Gives your Flash of Light, Holy Light, and Light's Vigil spells a 100% chance to not lose casting time when you take damage.
  - Classic (5 ranks):
    - Rank 1: Gives your Flash of Light and Holy Light spells a 14% chance to not lose casting time when you take damage.
    - Rank 2: Gives your Flash of Light and Holy Light spells a 28% chance to not lose casting time when you take damage.
    - Rank 3: Gives your Flash of Light and Holy Light spells a 42% chance to not lose casting time when you take damage.
    - Rank 4: Gives your Flash of Light and Holy Light spells a 56% chance to not lose casting time when you take damage.
    - Rank 5: Gives your Flash of Light and Holy Light spells a 70% chance to not lose casting time when you take damage.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Improved Seals** (3 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases the damage done by your Seals and Judgements by 5%.
  - Rank 2: Increases the damage done by your Seals and Judgements by 10%.
  - Rank 3: Increases the damage done by your Seals and Judgements by 15%.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Unyielding Faith** (2 ranks, column 4) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases your chance to resist Fear and Disorient effects by an additional 5%.
  - Rank 2: Increases your chance to resist Fear and Disorient effects by an additional 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Voice of Truth** (1 rank, column 1) [NEW]
  - Cost: Instant, 3 min cooldown
  - Grants you immunity to Silence and Interrupt effects. Lasts 6 sec.
  - Note: Slot from the third-party dataset, not yet verified in footage by us.
- **Reverence** (3 ranks, column 2) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: Allows 10% of your Mana regeneration to continue while casting.
  - Rank 2: Allows 20% of your Mana regeneration to continue while casting.
  - Rank 3: Allows 30% of your Mana regeneration to continue while casting.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Purifying Power** (2 ranks, column 3) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: Reduces the mana cost of your Cleanse and Purify spells by 10% and reduces the cooldown of your Exorcism and Holy Wrath spells by 17%.
  - Rank 2: Reduces the mana cost of your Cleanse and Purify spells by 20% and reduces the cooldown of your Exorcism and Holy Wrath spells by 34%.
  - Note: Higher-rank values taken from the nikftw dataset.

#### Row 4 (15 points)

- **Infusion of Light** (2 ranks, column 1) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: Your Holy Shock and Flash of Light critical hits reduce the cast time of your next Holy Light cast within 15 sec by 0.5 sec.
  - Rank 2: Your Holy Shock and Flash of Light critical hits reduce the cast time of your next Holy Light cast within 15 sec by 1.0 sec.
  - Note: Rank 1 read in footage, rank 2 verified on the BlizzCon slide.
- **Illumination** (5 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
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

- **Divine Precision** (3 ranks, column 1) [NEW] [ranks settled by the 70338 trait curves]
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
- **Consecrated Ground** (2 ranks, column 3) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: Gives your Holy spells 5% increased damage against the first 4 enemies that enter your Consecration.
  - Rank 2: Gives your Holy spells 10% increased damage against the first 4 enemies that enter your Consecration.
  - Note: Rank 2 verified on the BlizzCon slide, rank 1 from the nikftw screenshots.

#### Row 6 (25 points)

- **Holy Power** (5 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
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

- **Toughness** (5 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases your armor value from items by 2%.
  - Rank 2: Increases your armor value from items by 4%.
  - Rank 3: Increases your armor value from items by 6%.
  - Rank 4: Increases your armor value from items by 8%.
  - Rank 5: Increases your armor value from items by 10%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Redoubt** (5 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
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

- **Precision** (3 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases your chance to hit with all spells and attacks by 1%.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Classic (3 ranks):
    - Rank 1: Increases your chance to hit with melee weapons by 1%.
    - Rank 2: Increases your chance to hit with melee weapons by 2%.
    - Rank 3: Increases your chance to hit with melee weapons by 3%.
  - Note: Now applies to spells as well as attacks. Ranks 2-3 not yet seen.
- **Guardian's Favor** (2 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
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
- **Improved Righteous Fury** (3 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: While Righteous Fury is active, all damage taken is reduced by 2%.
  - Rank 2: While Righteous Fury is active, all damage taken is reduced by 4%.
  - Rank 3: While Righteous Fury is active, all damage taken is reduced by 6%.
  - Classic (3 ranks):
    - Rank 1: Increases the amount of threat generated by your Righteous Fury spell by 16%.
    - Rank 2: Increases the amount of threat generated by your Righteous Fury spell by 33%.
    - Rank 3: Increases the amount of threat generated by your Righteous Fury spell by 50%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Shield Specialization** (3 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
  - Requires: Redoubt (5/5)
  - Rank 1: Increases the amount of damage absorbed by your shield by 10%, and gives your blocks a 33% chance to restore 6% of your maximum Mana. May only occur once every 3 sec.
  - Rank 2: Increases the amount of damage absorbed by your shield by 20%, and gives your blocks a chance to restore 6% of your maximum Mana (rank 2 chance not yet captured, 33% at rank 1, 100% at rank 3). May only occur once every 3 sec.
  - Rank 3: Increases the amount of damage absorbed by your shield by 30%, and gives your blocks a 100% chance to restore 6% of your maximum Mana. May only occur once every 3 sec.
  - Classic (3 ranks):
    - Rank 1: Increases the amount of damage absorbed by your shield by 10%.
    - Rank 2: Increases the amount of damage absorbed by your shield by 20%.
    - Rank 3: Increases the amount of damage absorbed by your shield by 30%.
  - Note: Every rank carries the block-to-Mana clause Classic lacked: a 33% chance at rank 1 (footage) and 100% at rank 3 (BlizzCon slide), the rank 2 chance has not been seen.
- **Sacred Duty** (2 ranks, column 4) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases your total Stamina by 2% and reduces the cooldown of your Divine Shield, Divine Protection, and Templar's Bulwark spells by 30 sec.
  - Rank 2: Increases your total Stamina by 4% and reduces the cooldown of your Divine Shield, Divine Protection, and Templar's Bulwark spells by 60 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Swift Judgement** (1 rank, column 1) [NEW]
  - Requires: Improved Seal of Fury (1/1)
  - Cost: Instant, 1 min cooldown
  - Finishes the remaining cooldown on your Judgement ability and reduces the Mana cost of your next Judgement by 100%.
  - Note: Text verified on the BlizzCon slide. The cooldown read 1 min on the show-floor demo (footage and screenshot) but 50 sec on the panel slide, the demo value is shown.
- **One-Handed Weapon Specialization** (3 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
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
- **Reckoning** (5 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Gives you a 8% chance to gain an extra attack after Blocking a melee attack and a 20% chance to gain an extra attack after being the victim of a non-periodic critical strike.
  - Rank 2: Gives you a 16% chance to gain an extra attack after Blocking a melee attack and a 40% chance to gain an extra attack after being the victim of a non-periodic critical strike.
  - Rank 3: Gives you a 24% chance to gain an extra attack after Blocking a melee attack and a 60% chance to gain an extra attack after being the victim of a non-periodic critical strike.
  - Rank 4: Gives you a 32% chance to gain an extra attack after Blocking a melee attack and a 80% chance to gain an extra attack after being the victim of a non-periodic critical strike.
  - Rank 5: Gives you a 40% chance to gain an extra attack after Blocking a melee attack and a 100% chance to gain an extra attack after being the victim of a non-periodic critical strike.
  - Note: Rank 5 verified on the BlizzCon slide, ranks 2-4 from the nikftw dataset.

#### Row 6 (25 points)

- **Iron Creed** (5 ranks, column 3) [NEW] [ranks settled by the 70338 trait curves]
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
- **Benediction** (5 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
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
- **Holy Conduit** (2 ranks, column 2) [NEW] [ranks settled by the 70338 trait curves]
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

- **Vindication** (3 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Gives your damaging melee attacks a chance to reduce the target's Attack Power by 42, and increase your Attack Power by 1% for 30 sec.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Gives your damaging melee attacks a chance to reduce the target's Attack Power by 204, and increase your Attack Power by 3% for 30 sec.
  - Classic (3 ranks):
    - Rank 1: Gives the Paladin's damaging melee attacks a chance to reduce the target's Strength and Agility by 5% for 10 sec.
    - Rank 2: Gives the Paladin's damaging melee attacks a chance to reduce the target's Strength and Agility by 10% for 10 sec.
    - Rank 3: Gives the Paladin's damaging melee attacks a chance to reduce the target's Strength and Agility by 15% for 10 sec.
  - Note: Redesigned: an Attack Power debuff plus a self buff instead of Classic's Strength and Agility reduction. Rank 3 from the level 60 BlizzCon slide, rank 1 was read at level 38 from the nikftw screenshots, so the two are not on one scale, rank 2 not yet seen.
- **Sanctified Judgement** (3 ranks, column 2) [NEW] [ranks settled by the 70338 trait curves]
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

- **Eye for an Eye** (2 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
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

- **Two-Handed Weapon Specialization** (3 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases the damage you deal with two-handed melee weapons by 3%.
  - Rank 2: Increases the damage you deal with two-handed melee weapons by 6%.
  - Rank 3: Increases the damage you deal with two-handed melee weapons by 9%.
  - Classic (3 ranks):
    - Rank 1: Increases the damage you deal with two-handed melee weapons by 2%.
    - Rank 2: Increases the damage you deal with two-handed melee weapons by 4%.
    - Rank 3: Increases the damage you deal with two-handed melee weapons by 6%.
- **Vengeance** (3 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
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

- **Champion of the Light** (3 ranks, column 2) [NEW] [ranks settled by the 70338 trait curves]
  - Requires: Vengeance (3/3)
  - Rank 1: Increases your spell damage and healing by up to 33% of your Intellect.
  - Rank 2: Increases your spell damage and healing by up to 66% of your Intellect.
  - Rank 3: Increases your spell damage and healing by up to 100% of your Intellect.
  - Note: Rank 3 verified on the BlizzCon slide (100%, not the 99% read from the nikftw screenshot), rank 1 read in footage.
- **Instrument of Law** (2 ranks, column 3) [NEW] [ranks settled by the 70338 trait curves]
  - Requires: Repentance (1/1)
  - Rank 1: Reduces the cast time of your Hammer of Wrath by 0.5 sec, and reduces the threat you generate by 10% while Righteous Fury is not active.
  - Rank 2: Reduces the cast time of your Hammer of Wrath by 1.0 sec, and reduces all threat you generate by 20% while Righteous Fury is not active.
  - Note: Rank 1 verified in footage, rank 2 on the BlizzCon slide.

#### Row 7 (30 points)

- **Twist of Light** (1 rank, column 2) [NEW]
  - Requires: Champion of the Light (3/3)
  - When you replace your Seal of Command, Seal of Righteousness, Seal of Fury, or Seal of Justice with a different Seal, gain an Echo. Your next melee attack applies the replaced Seal's effects, consuming the Echo.
  - Note: Retribution capstone, text verified on the BlizzCon slide. The demo tooltip showed Requires Level 40.
