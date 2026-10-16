# Shaman

Source: wowforevertalents.com dataset (BlizzCon 2026 footage overlaid on Classic Era 1.15.9 client tables), data version 2026-09-13/14. See [../README.md](../README.md) for caveats. Numbers on new talents were read off level 38 demo characters, so absolute damage and mana values are not level 60 values.

Talents: 50 total, 15 new, 25 changed, 10 unchanged (verified).

Row N needs 5*(N-1) points spent in that tree. Row 1 is the top row. Rows 3, 4, 5 and 7 hold the 11, 16, 21 and 31 point talents.

Warning on multi-rank NEW and CHANGED talents: the footage showed rank 1 only. The dataset scaled the other ranks linearly from rank 1, which sometimes produces nonsense (for example Maelstrom Weapon rank 5 reading 'stacks up to 25 times'). Treat rank 1 as observed and the rest as a guess until the beta client is out. The [ranks settled by the 70338 trait curves] tag marks every multi-rank new or changed talent. Read the Note line, it says when a higher rank was actually seen on a BlizzCon slide.

## Client 1.60.1.70425 (2026-10-16)

From wago.tools, diffing every shaman spell-class spell, the talent tree (TraitNode, TraitNodeEntry, TraitDefinition, TraitCond, TraitEdge) and the trait curves between builds 70338 and 70425, with the 15 October development notes. The talent sections below are the BlizzCon transcription and predate this; `tools/forever_talents/client/shaman.json` and `ui/core/talents/trees/shaman.json` have the 70425 tree.

- Elemental tree: Lava Burst moved from row 7 to row 5 (the 20-point gate) and Elemental Mastery (16166) is the new row 7 talent, needing Lava Burst. Call of Thunder is gone from the tree (its spell has no level row, so it is not baseline either). Elemental Alacrity moved to row 2, Elemental Warding to row 3, Earthbound to row 4, and the columns were rearranged. Lightning Overload now needs Elemental Alacrity and Elemental Fury needs Lightning Overload. Restoration: Tidal Focus and Totemic Focus swapped rows (Tidal Focus is row 1 now). Every talent string with Elemental or Restoration points changes; the old Enhancement Water Shield build is no longer legal.
- Elemental Mastery: 15 sec, 3 min cooldown, off the global cooldown. Lightning Bolt and Lava Burst cast 0.5 sec faster and Chain Lightning has no cooldown.
- Call of Flame also lengthens Flame Shock by 3 / 6 / 9 sec (one tick a point). Trial by Fire (the sim's Improved Fire Nova) adds 10 / 20 sec to Magma Totem. Elemental Reach lost Flame Shock's extra range and cuts Totemic Projection's cooldown by 10 / 20 sec.
- Lava Burst: the three ranks moved ten levels down (30 / 40 / 50: 123 + 0.8, 164 + 0.9, 196 + 1.1 a level, 135 / 165 / 230 mana), and a new rank 4 (1326047) at 60 has what was rank 3 (220 + 1.3 a level, 265 mana). At 60 Lava Burst is unchanged under a new spell ID.
- Mana Tide Totem: ranks at 1 / 38 / 48 / 58 (88 / 133 / 197 / 290 mana a tick, 10 / 20 / 30 / 60 cost), the new rank 4 (1325256) at 58 being the old rank 3. Unchanged at 60 under a new spell ID.
- Fire Nova: coefficient 0.143 (was 0.214), base 33 / 55 / 130 / 179 / 239 plus 0.4 / 0.6 / 0.9 / 1.1 / 1.3 a level (was 50 / 102 / 183 / 284 / 403 plus 1.1 / 1.6 / 2.2 / 2.8 / 3.4), and cheaper (65 / 120 / 210 / 290 / 375 mana). Rank 5 hits 232 to 259 at 57 where it was 397 to 443.
- Chain Lightning costs 35% less (145 / 200 / 250 / 315). Stormstrike costs 8% of base mana (121.6 at 60, was 125).
- Stormstrike's mark is no longer used up by a spell that misses (development notes). Water Shield's globes also come from our damage spells' crits (ProcTypeMask 0x262a8 to 0x362a8).
- Strength of Earth, Grace of Air, Windfury Totem and Flametongue Totem became raid auras (effect 35 to 65), and Grace of Air lost the exclusivity flag it shared with Windfury Totem, so the two stack (master #77, standards 25).
- Unchanged and cosmetic: Eye of the Storm (23 / 47 / 70% not to lose cast time), Elemental Alacrity's and Elemental Warding's curves (the client swapped which definition holds them), Elemental Devastation (direct harmful spells), the shocks' new Attributes_6 bit 21 (mechanic shown as combat text), Restorative Totems' modifier (healing only).

The presets after the port, with the searched talents (as the mean of 120, 180 and 300 sec, 2 raid hits a minute), 20000 iterations, master before the port to after, 120 / 180 / 300 sec:

| Preset | Talents | 120 s | 180 s | 300 s |
|---|---|---|---|---|
| Elemental Level 60 (Windshaper Skyborne, was Orc) | `5501331003001351-055000101-053` | 691.6 → 732.3 | 684.9 → 724.9 | 671.5 → 715.3 |
| Elemental Level 60 + Mana Tide (Windshaper Skyborne, was Orc) | `5501331003001351-04-053150001001` | 685.0 → 722.1 | 678.0 → 714.6 | 665.2 → 699.4 |
| Enhancement Level 60 | `053503002-055030011005112151-02` | 865.7 → 835.5 | 852.9 → 825.3 | 834.8 → 809.1 |
| Enhancement Level 60 + Imp GW | `053503102-054032001005112151-01` | 862.0 → 827.4 | 848.7 → 818.4 | 830.7 → 811.5 |
| Enhancement Level 60 + Water Shield | `051003-054030021005112151-053020001` | 813.8 → 803.7 | 804.2 → 795.3 | 792.4 → 793.3 |
| Enhancement Level 60 + Mana Tide | `03-055030031005112051-053051001001` | 790.3 → 769.0 | 782.1 → 760.5 | 778.0 → 756.7 |

Elemental gains from Elemental Mastery (+22 as the mean of the three lengths), Eye of the Storm 3 in place of 3 Reverberation points (+8, now that raid hits push casts back) and Call of Flame's longer Flame Shock, against Call of Thunder's 3% crit. Enhancement loses mostly to the Fire Nova cut. Chain Lightning, even at the new cost and with no cooldown under Elemental Mastery, loses 15 to 55 on one target, so the Elemental rotations still leave it out.

Races at 180 sec on the new builds: Elemental Windshaper Skyborne 724.9, Orc 721.4, Dwarf 719.3, Tauren 719.1, Troll 710.1 (Mana Tide 714.6 / 711.2 / 708.7 / 708.6 / 700.1); Enhancement Level 60 Orc 825.3, Windshaper Skyborne 822.6, Tauren 817.1, Dwarf 816.9, Troll 813.7 (the other Enhancement builds rank the same). Best overall and best Horde: Windshaper Skyborne for Elemental, Orc for Enhancement. Best Alliance: Dwarf (the only Alliance shaman).

Rotation audit pass 23 (2026-10-16): no new Flame Shock with less than 9 sec left on both Elemental builds, Mana Tide Totem at 10% mana on Elemental Mana Tide, and a new Magma Totem with as little as 2 sec left on Water Shield and Enhancement Mana Tide: +0.7 / +1.1 / +0.8 / +0.6 at 180 sec, no loss at any length (paired, 20000 iterations, two seed sets). The rotopt templates and `tools/rotopt/shaman_presets.sh` now write the shipped Level 60 APLs' actions. The report requests were rebuilt with mksettings: since the starting totems were dropped (97b76aa9) they carry the rotations' prepull totems, which moves Enhancement +3 to +13 against the requests on master before.

The raid's Flametongue Totem (user ruling 2026-10-16, standards 25): the raid's second shaman's fire slot holds Flametongue Totem, which only Enhancement benefits from, because its own Windfury Weapon turns its own Windfury Totem's buff off for it. With it, Windfury Weapon is still the imbue (Rockbiter 25 to 30 and Frostbrand 36 to 45 behind at 180 sec), and the rotations moved: Magma Totem in place of Searing Totem with a 2 sec gate and the Earth Shock hold on Level 60 and Imp GW, Flame Shock first on Water Shield, Fire Nova after the shocks on Mana Tide. 180 sec: Enhancement 848.1 to 886.4, Imp GW 841.6 to 886.9, Water Shield 819.3 to 840.4, Mana Tide 781.7 to 802.8. Since the rotation pass of 2026-10-16, Lightning Bolt at 5 Maelstrom Weapon stacks waits for the shocks on Level 60, Imp GW and Water Shield, and Water Shield puts Flame Shock up before Stormstrike: Enhancement 887.0, Imp GW 887.8, Water Shield 842.2 at 180 sec. Level 60 stays the default, because Imp GW's lead at 180 sec (+0.8 and +0.1 on two seed sets) is within noise. A talent search on the new order found no build ahead at all of 120, 180 and 300 sec. In game, check that a warrior keeps Windfury with both totems down and that an Enhancement shaman with Windfury Weapon gets Flametongue Totem (notes.md).

## Client 1.60.1.70338 and the 7 October deep dive

From wago.tools (build 1.60.1.70338, 2026-10-10, the newest so far), foreverchanges.pro's 70338 talent calculator and the [Mage and Shaman class deep dive](https://worldofwarcraft.blizzard.com/en-us/news/24302097) of 7 October. The Elemental tree matches the sim rank for rank (shaman_audit.md 8.8 to 8.16).

- Tier 1 sets, one per spec, in ItemSet and ItemSetSpell. The items carry no stats yet.
  - The Spiritcaller's Storm (2111, Elemental): 2 pieces +1% hit, 3 pieces Earthbind Totem cooldown -2 sec, 4 pieces +21 spell damage against Elementals, 5 pieces Lava Burst cooldown -1 sec. The crafted Spiritcaller Kilt, Gloves, Mantle and Boots (276538 to 276541, item level 61) belong to it.
  - The Spiritcaller's Rage (2110, Enhancement): 2 pieces 1% attack and cast speed, 3 pieces Air and Earth totem radius +10 yd, 4 pieces +36 attack power against Elementals, 5 pieces Stormstrike cooldown -0.5 sec. The same bonuses as the crafted set of build 70205.
  - The Spiritcaller (2109, Restoration): 2 pieces +10 Spirit, 3 pieces Fire and Water totem radius +10 yd, 4 pieces +26 healing and +9 damage, 5 pieces Riptide cooldown -1 sec.
- Forever shaman relics (ItemEffect): Burning Totem (Flame Shock 3 sec longer), Totem of the Storm (Lightning Bolt can proc Maelstrom Weapon at half the chance), Tidal Totem (Healing Wave cost), Totem of Urgency (Lesser Healing Wave crit, Enchanting) and Polished Driftwood Icon (8% of mana regen while casting).
- The deep dive removes Elemental Mastery ("We could not increase Elemental Shaman damage to appropriate values in PvE and keep this talent") and Nature's Guidance, and says Elemental Focus gives Clearcasting less often on low ranks. Totems take a 1 sec global cooldown and the persistent ones last 5 min. Weapon imbues last 60 min and stack with oils and stones.

## Sim re-checks for the Level 60 presets (2026-10-13)

- Forever mechanics, against client 70338 (wago.tools), wowhead Forever's tooltips (dataEnv 17), the Mage and Shaman deep dive and foreverchanges.pro's patch notes: the sim already had the imbue and totem buff rules (a main-hand Windfury or Flametongue Weapon turns off our own benefit from the totem of the same kind, and the weapon holds a totem buff, our imbue and an oil together), the totems, shocks, weapon rules and Lava Burst. With Windfury Totem kept for the group, Windfury Weapon is still the best imbue on the Level 60 build (813.0 at 3 min, Flametongue Weapon 811.3, Rockbiter 810.9, Frostbrand 786.0). wowhead's Lava Burst rank 3 tooltip (203 to 258, 54% with Flame Shock) is the level 68 value: the client's dummy is 20 with 4.2 a level from 60, so 20% at 60. Rage of the Farseer has been attack speed only since 70009 (`rage_of_the_farseer_test.go`). Maelstrom Weapon's talent aura (408498) has ProcChance 100 and no PPM row, so its rate (2 a minute a point) stays a guess.
- Downranking: under Forever a rank keeps 1 - 5% for each level past its learn level + 17 of its spell power coefficient (Blizzard's 2026-10-01 development notes, the formula from foreverchanges.pro/downranking). Lightning Bolt and Chain Lightning take it (`downrank_test.go`), so the elemental low-mana fallback is rank 8 (learned at 44) now, not rank 7 (75% at 60).
- Talent verification (every preset talent, from sim runs on the Level 60 and Level 30 builds): 108 entries over 32 talents, 92 verified, 4 Maelstrom Weapon entries partly (the proc rate), 12 utility or inert, no bugs. Not verifiable in the sim: Maelstrom Weapon's proc rate, Elemental Focus's trigger set, Water Shield and Improved Stormstrike's dodge and parry reset (inert at 0 hits), two dummy effects.

## Totems in the presets (standards 23, user ruling 2026-10-15)

- There is no shaman totem duty: every shaman drops its own party's totems, like a warrior's Sunder, and the raid has none of its own. Enhancement puts down the melee party's Strength of Earth, Windfury Totem and Mana Spring (both are air totems, so one shaman puts down one of them; our own Windfury Weapon turns Windfury Totem's buff off for us, so it is the party's); Elemental puts down the caster party's Mana Spring, its only DPS totem. Searing or Magma Totem and Mana Tide Totem are the shaman's own as before.
- Under Forever the totems last 5 min and a starting totem costs no mana. The Enhancement builds put theirs down 4, 3 and 2 sec before the pull so they last a 300 sec fight. Against the raid's totems the cost is 0.0 to 0.7 DPS at 3 min and 0.9 to 2.4 at 5 min (sim-status.md has the numbers).

## Beta build 1.60.1.70009 (2026-09-24)

The [development notes of 24 September](https://www.wowhead.com/blue-tracker/topic/us/wow-forever-beta-development-notes-updated-september-24-2360696), with the numbers from the client data on foreverchanges.pro (build 1.60.1.70009, downrank calculator and talent calculator). The sim follows all of it.

- Lightning Bolt rank 3 is 45 average at level 14 (was 35) and rank 4 is 56 at level 20 (was 50), so every rank is an upgrade. Growth per level, coefficients, mana and cast times are unchanged.
- Lava Burst rank 1 is 164 average at level 40 (was 113) and rank 2 is 196 at level 50 (was 179), about 10% above the Lightning Bolt rank of the same level.
- Elemental Fury (now row 6, after Call of Thunder) and Elemental Alacrity (now row 3, no prerequisite) swapped places. Call of Thunder now needs Elemental Alacrity. An enhancement build can no longer reach Elemental Fury.
- Rage of the Farseer is attack speed only (it was attack and cast speed).
- Windfury, Grace of Air and Tranquil Air Totem no longer stacked in a party, even from different shamans (development notes, 24 September). Client 70425 undid it for Grace of Air: "Grace of Air Totem is no longer exclusive with Windfury Totem and Flametongue Totem" (development notes, 15 October; Grace of Air's Attributes_11 lost 1024). Windfury and Flametongue Totem stay exclusive. One shaman still has one air totem, so Windfury Totem with Grace of Air takes two shamans, and both totems are raid-wide since 70425.
- Flametongue Totem has no duration, does not stack with itself, with Flametongue Weapon or with Windfury Totem. The sim models it since 2026-10-01: a fire totem from level 28 (90 Mana, 5 min) that gives every main hand hit 548 / 25 Fire damage per 4 sec of weapon speed (21.9 on a 4.0 speed weapon, rank 1 does not grow with level), at 10% of spell power. Under Forever a weapon has three slots that all work together: a totem buff (Windfury or Flametongue Totem, one at a time), the shaman's own imbue (Windfury, Rockbiter, Flametongue or Frostbrand Weapon) and an oil or stone. The sim models this since 2026-10-03, for the shaman's own totems and for another shaman's (the raid buffs' Totem Weapon Buff). When both totems stand, the buff of the one placed last holds the slot.

## Trait curves, client 1.60.1.70338 (2026-10-12)

Forever's talents are traits. TraitDefinitionEffectPoints gives each effect a per-rank curve that overrides the base points on the rank 1 spell, so these are the real per-rank values. Every shaman talent with more than one rank has one, read from wago.tools (TraitDefinition, TraitDefinitionEffectPoints, CurvePoint, build 1.60.1.70338). The sim matched all of them, so nothing changed:

- Elemental: Convection 2% a rank, Concussion 1%, Elemental Warding 3 / 7 / 10%, Reverberation 0.2 sec, Call of Flame 5%, Elemental Devastation 3%, Elemental Alacrity 0.17 / 0.33 / 0.50 sec, Improved Fire Nova 10% and 2 sec, Eye of the Storm 23 / 47 / 70%, Elemental Reach 3 / 6 and 8 / 15 yards, Lightning Overload 3 / 7 / 10%, Elemental Fury 20%.
- Enhancement: Earth's Grasp 25 / 50% and 10 / 20%, Thundering Strikes 1%, Ancestral Knowledge 2%, Guardian Totems 10% and 1 sec, Mental Dexterity 33 / 67 / 100%, Improved Ghost Wolf 1 / 3 sec, Improved Lightning Shield 5%, Elemental Weapons 7 / 13 / 20% Rockbiter, 5 / 10 / 15% Flametongue and Frostbrand, 13 / 27 / 40% Windfury, Anticipation 2%, Toughness 2%, Flurry 5% attack speed a rank, Mental Quickness 15%, Improved Stormstrike 50%, Maelstrom Weapon 4% a rank per stack.
- Restoration: Improved Healing Wave 0.1 sec, Totemic Focus 5%, Mindfulness 17 / 33 / 50%, Natural Grace 5%, Tidal Focus 1% (cost and hit), Improved Reincarnation, Ancestral Healing and Healing Way 8 / 17 / 25%, Healing Focus 23 / 47 / 70%, Tidal Mastery 1%, Restorative Totems 5% Mana Spring and 10% Healing Stream, Purification 2%.

## Beta client 1.60.1.69876 (2026-09-16)

The beta client's shaman tree, read through [hyjal.cc](https://hyjal.cc/updates/1.60.1.69876), against the BlizzCon data below. Rank counts and prerequisites held. Tidal Mastery and Totemic Focus swapped places (Totemic Focus is row 1, Tidal Mastery row 4). Ranks 2 and up are now real numbers, and `tools/forever_talents/data/shaman.json` in the sim repo is regenerated from the client tree (`client/shaman.json`), so the per-talent text below is superseded where it differs. The sim follows the client.

- Stormstrike: 125 Mana, 8 sec cooldown, and it "increases the damage you deal to the target with your next Lightning Bolt, Chain Lightning, or Earth Shock spell by 20% for 12 sec". One spell, the shaman's own, instead of all Nature damage on a 20 sec cooldown at 21% of base mana.
- Lightning Overload is 3 / 7 / 10%. Elemental Alacrity is 0.17 / 0.33 / 0.50 sec. Mental Dexterity is 33 / 67 / 100% of Intellect. Elemental Weapons rank 2 is 13% Rockbiter.
- Lava Burst at 60: 106 to 135 Fire damage, 165 Mana, 2.5 sec cast, 10 sec cooldown, 20% more with Flame Shock up.
- Improved Fire Nova (20% / 4 sec at rank 2), Improved Stormstrike (100% at rank 2) and Maelstrom Weapon (4% per rank per stack) scale linearly, as the sim assumed.
- Rage of the Farseer: 30% melee and cast speed for 25 sec, 3 min cooldown (the sim assumed 3 min).
- Tidal Focus reads "improves your chance to hit by X%". Mindfulness is 17 / 33 / 50%, Healing Way 8 / 17 / 25%, Healing Focus 23 / 47 / 70%, Improved Reincarnation rank 2 doubles rank 1, Riptide 486 to 534 plus 445 over 15 sec, 245 Mana, 6 sec cooldown, Mana Tide Totem 88 mana every 3 sec.
- Spirit Weapons (Enhancement row 4, 30% less threat) is a must for a raiding enhancement shaman even though the sim scores no threat, so the shipped level 60 build pins it.

## Spellbook changes (trainer abilities)

10 of 51 trainer abilities confirmed from footage. Abilities not listed below are either verified unchanged or still Classic placeholders (unverified).

### New abilities

- **Totemic Projection** (enhancement, 230 Mana, 30 yd range, Instant, 1 min cooldown)
  - Relocates your active totems to the specified location.
- **Call of the Ancestors** (elemental-combat)
  - Listed in the Elemental Combat tab of the level 38 demo shaman. Exact effect not yet captured from footage.
  - Note: Listed without a rank. Tooltip not yet captured from footage.
- **Call of the Elements** (elemental-combat)
  - Listed in the Elemental Combat tab of the level 38 demo shaman. Exact effect not yet captured from footage.
  - Note: Listed without a rank. Tooltip not yet captured from footage.
- **Fire Nova** (elemental-combat)
  - Listed in the Elemental Combat tab of the level 38 demo shaman as Rank 3. Exact effect not yet captured from footage.
  - Note: Takes the place of Fire Nova Totem, which is not in the spellbook, Rank 3 known at level 38, matching the old totem's rank schedule. Tooltip not yet captured.
  - Beta build 1.60.1.70009 (ForeverChanges spellbook, wowhead spell 408341 to 408345): "Instantly inflicts 48 to 56 fire damage to enemies within 10 yd of your active Fire totem." Instant, 30 yd range, 10 sec cooldown, 95 / 170 / 280 / 395 / 520 Mana at levels 12 / 22 / 32 / 42 / 52. The damage is the old totem's (Classic's 53 to 62 up to 413 to 459, reached five levels after the rank is learned, 10% of spell power at rank 1 and 14.3% after). Totemic Focus no longer applies, Call of Flame and Improved Fire Nova do. The sim casts it this way since 2026-10-01. Totems take a 1 sec global cooldown.
- **Totemic Recall** (elemental-combat)
  - Listed in the Elemental Combat tab of the level 38 demo shaman. Exact effect not yet captured from footage.
  - Note: Listed without a rank. Tooltip not yet captured from footage.

### Changed abilities

- **Nature Resistance Totem** (enhancement, Level 30, 75 Mana, Instant)
  - Summons a Nature Resistance Totem with 5 health at the feet of the caster for 5 min that increases the nature resistance of party and raid members within 30 yards by 30.
  - Classic: Summons a Nature Resistance Totem with 5 health at the feet of the caster for 2 min that increases the nature resistance of party members within 20 yards by 60.
  - Note: Now lasts 5 minutes, reaches 30 yards and covers raid members. Only rank 1 seen.
- **Windfury Weapon** (enhancement, Level 30, 90 Mana, Instant)
  - Imbue the Shaman's weapon with wind. Each hit has a 20% chance of granting you 2 extra attacks with 103 extra melee attack power. When applied to main hand, disables any benefit you personally receive from Windfury Totem. Lasts for 60 minutes.
  - Classic: Imbue the Shaman's weapon with wind. Each hit has a 20% chance of granting you 2 extra attacks with 333 extra melee attack power. Lasts for 5 minutes.
  - Note: Lasts 60 minutes and a main-hand imbue no longer stacks with your own Windfury Totem. Only rank 1 seen, on a level 38 character.
- **Windfury Totem** (enhancement, Level 32, 115 Mana, Instant)
  - Summons a Windfury Totem with 5 health at the feet of the caster. The totem enhances the melee attacks of all party members within 0 yards. Each main hand hit has a 20% chance of granting the attacker 1 extra attack with 95 extra melee attack power. Lasts 5 min.
  - Classic: Summons a Windfury Totem with 5 health at the feet of the caster. The totem enchants all party members main-hand weapons with wind, if they are within 20 yards. Each hit has a 20% chance of granting the attacker 1 extra attack with 315 extra melee attack power. Lasts 2 min.
  - Note: Lasts 5 minutes and procs on main-hand hits only. The tooltip in this build read "within 0 yards", which looks like a display bug. Only rank 1 seen, on a level 38 character.
- **Earth Shock** (elemental-combat, Level 36, 240 Mana, 20 yd range, Instant, 6 sec cooldown)
  - Instantly shocks the target with concussive force, causing 162 to 171 Nature damage. It also interrupts spellcasting and prevents any spell in that school from being cast for 2 sec. Causes a high amount of threat.
  - Classic: Instantly shocks the target with concussive force, causing 517 to 545 Nature damage. It also interrupts spellcasting and prevents any spell in that school from being cast for 2 sec. Causes a high amount of threat.
  - Note: Rank 5 read 162 to 171 damage on the level 38 demo shaman, where the Classic client gives 229 to 243 at that level. Other ranks not seen.

### Removed abilities

- **Fire Nova Totem** (elemental-combat, Level 12, 520 Mana, Instant, 15 sec cooldown)
  - Summons a Fire Nova Totem that has 5 health and lasts 5 sec. Unless it is destroyed within 4 sec., the totem inflicts 413 to 459 fire damage to enemies within 10 yd.
  - Note: Not in the level 38 demo shaman's Elemental Combat tab, a ranked Fire Nova spell is listed instead.

### Not yet verified (Classic text assumed)

Rockbiter Weapon, Stoneskin Totem, Lightning Shield, Flametongue Weapon, Strength of Earth Totem, Frostbrand Weapon, Ghost Wolf, Water Breathing, Frost Resistance Totem, Far Sight, Fire Resistance Totem, Flametongue Totem, Water Walking, Astral Recall, Grounding Totem, Sentry Totem, Windwall Totem, Grace of Air Totem, Healing Wave, Ancestral Spirit, Cure Poison, Tremor Totem, Healing Stream Totem, Lesser Healing Wave, Cure Disease, Poison Cleansing Totem, Mana Spring Totem, Reincarnation, Disease Cleansing Totem, Chain Heal, Mana Tide Totem, Tranquil Air Totem, Lightning Bolt, Earthbind Totem, Stoneclaw Totem, Flame Shock, Searing Totem, Purge, Frost Shock, Magma Totem, Chain Lightning

## Talents

### Elemental Combat

16 talents, 5 new. 31-point talent: Lava Burst.

#### Row 1 (0 points)

- **Convection** (5 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Reduces the mana cost of your Shock, Lightning Bolt, Lava Burst, and Chain Lightning spells by 2%.
  - Rank 2: Reduces the mana cost of your Shock, Lightning Bolt, Lava Burst, and Chain Lightning spells by 4%.
  - Rank 3: Reduces the mana cost of your Shock, Lightning Bolt, Lava Burst, and Chain Lightning spells by 6%.
  - Rank 4: Reduces the mana cost of your Shock, Lightning Bolt, Lava Burst, and Chain Lightning spells by 8%.
  - Rank 5: Reduces the mana cost of your Shock, Lightning Bolt, Lava Burst, and Chain Lightning spells by 10%.
  - Classic (5 ranks):
    - Rank 1: Reduces the mana cost of your Shock, Lightning Bolt and Chain Lightning spells by 2%.
    - Rank 2: Reduces the mana cost of your Shock, Lightning Bolt and Chain Lightning spells by 4%.
    - Rank 3: Reduces the mana cost of your Shock, Lightning Bolt and Chain Lightning spells by 6%.
    - Rank 4: Reduces the mana cost of your Shock, Lightning Bolt and Chain Lightning spells by 8%.
    - Rank 5: Reduces the mana cost of your Shock, Lightning Bolt and Chain Lightning spells by 10%.
- **Concussion** (5 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases the damage done by your Lightning Bolt, Chain Lightning, and Earth Shock spells by 1%.
  - Rank 2: Increases the damage done by your Lightning Bolt, Chain Lightning, and Earth Shock spells by 2%.
  - Rank 3: Increases the damage done by your Lightning Bolt, Chain Lightning, and Earth Shock spells by 3%.
  - Rank 4: Increases the damage done by your Lightning Bolt, Chain Lightning, and Earth Shock spells by 4%.
  - Rank 5: Increases the damage done by your Lightning Bolt, Chain Lightning, and Earth Shock spells by 5%.
  - Classic (5 ranks):
    - Rank 1: Increases the damage done by your Lightning Bolt, Chain Lightning and Shock spells by 1%.
    - Rank 2: Increases the damage done by your Lightning Bolt, Chain Lightning and Shock spells by 2%.
    - Rank 3: Increases the damage done by your Lightning Bolt, Chain Lightning and Shock spells by 3%.
    - Rank 4: Increases the damage done by your Lightning Bolt, Chain Lightning and Shock spells by 4%.
    - Rank 5: Increases the damage done by your Lightning Bolt, Chain Lightning and Shock spells by 5%.

#### Row 2 (5 points)

- **Elemental Warding** (3 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Reduces damage taken from Fire, Frost, and Nature effects by 3%.
  - Rank 2: Reduces damage taken from Fire, Frost, and Nature effects by 6%.
  - Rank 3: Reduces damage taken from Fire, Frost, and Nature effects by 9%.
  - Classic (3 ranks):
    - Rank 1: Reduces damage taken from Fire, Frost and Nature effects by 4%.
    - Rank 2: Reduces damage taken from Fire, Frost and Nature effects by 7%.
    - Rank 3: Reduces damage taken from Fire, Frost and Nature effects by 10%.
- **Reverberation** (5 ranks, column 2) [unchanged]
  - Rank 1: Reduces the cooldown of your Shock spells by 0.2 sec.
  - Rank 2: Reduces the cooldown of your Shock spells by 0.4 sec.
  - Rank 3: Reduces the cooldown of your Shock spells by 0.6 sec.
  - Rank 4: Reduces the cooldown of your Shock spells by 0.8 sec.
  - Rank 5: Reduces the cooldown of your Shock spells by 1 sec.
- **Call of Flame** (3 ranks, column 3) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases the damage done by your Fire Totems and by your Flame Shock, Fire Nova, and Lava Burst spells by 5%.
  - Rank 2: Increases the damage done by your Fire Totems and by your Flame Shock, Fire Nova, and Lava Burst spells by 10%.
  - Rank 3: Increases the damage done by your Fire Totems and by your Flame Shock, Fire Nova, and Lava Burst spells by 15%.
  - Note: Slot from the third-party dataset, not yet verified in footage by us.
- **Elemental Devastation** (3 ranks, column 4) [unchanged]
  - Rank 1: Your offensive spell critical strikes will increase your chance to get a critical strike with melee attacks by 3% for 10 sec.
  - Rank 2: Your offensive spell crits will increase your chance to get a critical strike with melee attacks by 6% for 10 sec.
  - Rank 3: Your offensive spell crits will increase your chance to get a critical strike with melee attacks by 9% for 10 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Elemental Focus** (1 rank, column 2) [unchanged]
  - Gives you a 10% chance to enter a Clearcasting state after casting any Fire, Frost, or Nature damage spell. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
- **Elemental Fury** (5 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
  - Requires: Call of Flame (3/3)
  - Rank 1: Increases the critical strike damage bonus of your Searing and Magma Totems and your Fire, Frost, and Nature spells by 20%.
  - Rank 2: Increases the critical strike damage bonus of your Searing and Magma Totems and your Fire, Frost, and Nature spells by 40%.
  - Rank 3: Increases the critical strike damage bonus of your Searing and Magma Totems and your Fire, Frost, and Nature spells by 60%.
  - Rank 4: Increases the critical strike damage bonus of your Searing and Magma Totems and your Fire, Frost, and Nature spells by 80%.
  - Rank 5: Increases the critical strike damage bonus of your Searing and Magma Totems and your Fire, Frost, and Nature spells by 100%.
  - Classic: Increases the critical strike damage bonus of your Searing, Magma, and Fire Nova Totems and your Fire, Frost, and Nature spells by 100%.
  - Note: The arrow from Call of Flame ends here, third row third slot.

#### Row 4 (15 points)

- **Improved Fire Nova** (2 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases the damage done by your Fire Nova spell by 10% and reduces its cooldown by 2 sec.
  - Rank 2: Increases the damage done by your Fire Nova spell by 20% and reduces its cooldown by 4 sec.
  - Classic (2 ranks):
    - Rank 1: Reduces the delay before your Fire Nova Totem activates by 1 sec. and decreases the threat generated by your Magma Totem by 25%.
    - Rank 2: Reduces the delay before your Fire Nova Totem activates by 2 sec. and decreases the threat generated by your Magma Totem by 50%.
- **Eye of the Storm** (3 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Reduces the pushback suffered from damaging attacks while casting Lightning Bolt, Chain Lightning, and Lava Burst by 23%.
  - Rank 2: Reduces the pushback suffered from damaging attacks while casting Lightning Bolt, Chain Lightning, and Lava Burst by 46%.
  - Rank 3: Reduces the pushback suffered from damaging attacks while casting Lightning Bolt, Chain Lightning, and Lava Burst by 69%.
  - Classic (3 ranks):
    - Rank 1: Gives you a 33% chance to gain the Focused Casting effect that lasts for 6 sec after being the victim of a melee or ranged critical strike. The Focused Casting effect prevents you from losing casting time when taking damage.
    - Rank 2: Gives you a 66% chance to gain the Focused Casting effect that lasts for 6 sec after being the victim of a melee or ranged critical strike. The Focused Casting effect prevents you from losing casting time when taking damage.
    - Rank 3: Gives you a 100% chance to gain the Focused Casting effect that lasts for 6 sec after being the victim of a melee or ranged critical strike. The Focused Casting effect prevents you from losing casting time when taking damage.
  - Note: Now also covers Lava Burst.
- **Call of Thunder** (1 rank, column 3) [CHANGED]
  - Requires: Elemental Fury (5/5)
  - Increases the critical strike chance of your Lightning Bolt and Chain Lightning spells by 3%.
  - Classic (5 ranks):
    - Rank 1: Increases the critical strike chance of your Lightning Bolt and Chain Lightning spells by an additional 1%.
    - Rank 2: Increases the critical strike chance of your Lightning Bolt and Chain Lightning spells by an additional 2%.
    - Rank 3: Increases the critical strike chance of your Lightning Bolt and Chain Lightning spells by an additional 3%.
    - Rank 4: Increases the critical strike chance of your Lightning Bolt and Chain Lightning spells by an additional 4%.
    - Rank 5: Increases the critical strike chance of your Lightning Bolt and Chain Lightning spells by an additional 6%.
  - Note: Slot from the third-party dataset, not yet verified in footage by us.

#### Row 5 (20 points)

- **Elemental Reach** (2 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases the range of your Lightning Bolt, Chain Lightning, Fire Nova, and Lava Burst spells by 3 yards, and increases the range of your Flame Shock spell by 8 yards.
  - Rank 2: Increases the range of your Lightning Bolt, Chain Lightning, Fire Nova, and Lava Burst spells by 6 yards, and increases the range of your Flame Shock spell by 16 yards.
  - Classic (2 ranks):
    - Rank 1: Increases the range of your Lightning Bolt and Chain Lightning spells by 3 yards.
    - Rank 2: Increases the range of your Lightning Bolt and Chain Lightning spells by 6 yards.
  - Note: Now also extends Fire Nova, Lava Burst and Flame Shock. Rank 2 extrapolated.
- **Lightning Overload** (3 ranks, column 2) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: Gives your Lightning Bolt and Chain Lightning spells a 3% chance to cast a second, similar spell on the same target at no additional cost that causes half damage and no threat.
  - Rank 2: Gives your Lightning Bolt and Chain Lightning spells a 6% chance to cast a second, similar spell on the same target at no additional cost that causes half damage and no threat.
  - Rank 3: Gives your Lightning Bolt and Chain Lightning spells a 9% chance to cast a second, similar spell on the same target at no additional cost that causes half damage and no threat.
  - Note: Ranks 2 and 3 extrapolated. The client's trait curve says 3 / 7 / 10%.
- **Earthbound** (1 rank, column 4) [NEW]
  - Your Earthbind Totem immobilizes nearby targets for 5 sec when cast.
  - Note: Slot from the third-party dataset, not yet verified in footage by us.

#### Row 6 (25 points)

- **Elemental Alacrity** (3 ranks, column 3) [NEW] [ranks settled by the 70338 trait curves]
  - Requires: Call of Thunder (1/1)
  - Rank 1: Reduces the cast time of your Lightning Bolt, Chain Lightning, and Lava Burst spells by 0.17 sec.
  - Rank 2: Reduces the cast time of your Lightning Bolt, Chain Lightning, and Lava Burst spells by 0.34 sec.
  - Rank 3: Reduces the cast time of your Lightning Bolt, Chain Lightning, and Lava Burst spells by 0.51 sec.
  - Note: Slot from the third-party dataset, not yet verified in footage by us.

#### Row 7 (30 points)

- **Lava Burst** (1 rank, column 2) [NEW]
  - Requires: Lightning Overload (3/3)
  - Cost: 165 Mana, 30 yd range, 2.5 sec cast, 10 sec cooldown
  - You hurl molten lava at the target, dealing 158 to 187 Fire damage. If your Flame Shock is on the target, Lava Burst deals 20% increased damage.
  - Note: 165 Mana, 30 yd range, 2.5 sec cast, 10 sec cooldown. Capstone, the arrow from Lightning Overload leads here.

### Enhancement

18 talents, 7 new. 31-point talent: Rage of the Farseer.

#### Row 1 (0 points)

- **Earth's Grasp** (2 ranks, column 1) [unchanged]
  - Rank 1: Increases the health of your Stoneclaw Totem by 25% and the radius of your Earthbind Totem by 10%.
  - Rank 2: Increases the health of your Stoneclaw Totem by 50% and the radius of your Earthbind Totem by 20%.
  - Note: Moved from Elemental Combat to the first Enhancement slot, text unchanged.
- **Thundering Strikes** (5 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Improves your chance to get a critical strike with all spells and attacks by 1%.
  - Rank 2: Improves your chance to get a critical strike with all spells and attacks by 2%.
  - Rank 3: Improves your chance to get a critical strike with all spells and attacks by 3%.
  - Rank 4: Improves your chance to get a critical strike with all spells and attacks by 4%.
  - Rank 5: Improves your chance to get a critical strike with all spells and attacks by 5%.
  - Note: Higher-rank values taken from the nikftw dataset.
- **Ancestral Knowledge** (5 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases your Intellect by 2%.
  - Rank 2: Increases your Intellect by 4%.
  - Rank 3: Increases your Intellect by 6%.
  - Rank 4: Increases your Intellect by 8%.
  - Rank 5: Increases your Intellect by 10%.
  - Note: Higher-rank values taken from the nikftw dataset.

#### Row 2 (5 points)

- **Guardian Totems** (2 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases the amount of damage reduced by your Stoneskin Totem and Windwall Totem by 10% and reduces the cooldown of your Grounding Totem by 1 sec.
  - Rank 2: Increases the amount of damage reduced by your Stoneskin Totem and Windwall Totem by 20% and reduces the cooldown of your Grounding Totem by 2 sec.
  - Note: Slot from the third-party dataset, not yet verified in footage by us.
- **Mental Dexterity** (3 ranks, column 2) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases your Attack Power by an amount equal to 33% of your Intellect.
  - Rank 2: Increases your Attack Power by an amount equal to 66% of your Intellect.
  - Rank 3: Increases your Attack Power by an amount equal to 99% of your Intellect.
- **Improved Ghost Wolf** (2 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Reduces the cast time of your Ghost Wolf spell by 1.0 sec, and Ghost Wolf may be used indoors.
  - Rank 2: Reduces the cast time of your Ghost Wolf spell by 2 sec, and Ghost Wolf may be used indoors.
  - Classic (2 ranks):
    - Rank 1: Reduces the cast time of your Ghost Wolf spell by 1 sec.
    - Rank 2: Reduces the cast time of your Ghost Wolf spell by 2 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Lightning Shield** (3 ranks, column 4) [unchanged]
  - Rank 1: Increases the damage done by your Lightning Shield orbs by 5%.
  - Rank 2: Increases the damage done by your Lightning Shield orbs by 10%.
  - Rank 3: Increases the damage done by your Lightning Shield orbs by 15%.

#### Row 3 (10 points)

- **Elemental Weapons** (3 ranks, column 1) [unchanged]
  - Rank 1: Increases the melee attack power bonus of your Rockbiter Weapon by 7%, your Windfury Weapon effect by 13% and increases the damage caused by your Flametongue Weapon and Frostbrand Weapon by 5%.
  - Rank 2: Increases the melee attack power bonus of your Rockbiter Weapon by 14%, your Windfury Weapon effect by 27% and increases the damage caused by your Flametongue Weapon and Frostbrand Weapon by 10%.
  - Rank 3: Increases the melee attack power bonus of your Rockbiter Weapon by 20%, your Windfury Weapon effect by 40% and increases the damage caused by your Flametongue Weapon and Frostbrand Weapon by 15%.
- **Shamanistic Focus** (1 rank, column 3) [NEW]
  - Reduces the mana cost of your Shock and Lightning Shield spells by 45%.
- **Anticipation** (3 ranks, column 4) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases your chance to dodge by an additional 2%.
  - Rank 2: Increases your chance to dodge by an additional 4%.
  - Rank 3: Increases your chance to dodge by an additional 6%.
  - Classic (5 ranks):
    - Rank 1: Increases your chance to dodge by an additional 1%.
    - Rank 2: Increases your chance to dodge by an additional 2%.
    - Rank 3: Increases your chance to dodge by an additional 3%.
    - Rank 4: Increases your chance to dodge by an additional 4%.
    - Rank 5: Increases your chance to dodge by an additional 5%.

#### Row 4 (15 points)

- **Toughness** (5 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases your Stamina by 2%.
  - Rank 2: Increases your Stamina by 4%.
  - Rank 3: Increases your Stamina by 6%.
  - Rank 4: Increases your Stamina by 8%.
  - Rank 5: Increases your Stamina by 10%.
  - Classic (5 ranks):
    - Rank 1: Increases your armor value from items by 2%.
    - Rank 2: Increases your armor value from items by 4%.
    - Rank 3: Increases your armor value from items by 6%.
    - Rank 4: Increases your armor value from items by 8%.
    - Rank 5: Increases your armor value from items by 10%.
- **Flurry** (5 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Requires: Mental Dexterity (3/3)
  - Rank 1: Increases your attack speed by 5% for your next 3 swings after dealing a melee critical strike.
  - Rank 2: Increases your attack speed by 10% for your next 6 swings after dealing a melee critical strike.
  - Rank 3: Increases your attack speed by 15% for your next 9 swings after dealing a melee critical strike.
  - Rank 4: Increases your attack speed by 20% for your next 12 swings after dealing a melee critical strike.
  - Rank 5: Increases your attack speed by 25% for your next 15 swings after dealing a melee critical strike.
  - Classic (5 ranks):
    - Rank 1: Increases your attack speed by 10% for your next 3 swings after dealing a critical strike.
    - Rank 2: Increases your attack speed by 15% for your next 3 swings after dealing a critical strike.
    - Rank 3: Increases your attack speed by 20% for your next 3 swings after dealing a critical strike.
    - Rank 4: Increases your attack speed by 25% for your next 3 swings after dealing a critical strike.
    - Rank 5: Increases your attack speed by 30% for your next 3 swings after dealing a critical strike.
- **Stormstrike** (1 rank, column 3) [CHANGED]
  - Cost: 125 Mana, Melee Range, Instant, 8 sec cooldown
  - Instantly strike for normal weapon damage and increase Nature damage you deal to the target by 20% for 12 sec.
  - Classic: Gives you an extra attack. In addition, the next 2 sources of Nature damage dealt to the target are increased by 20%. Lasts 12 sec.

#### Row 5 (20 points)

- **Spirit Weapons** (1 rank, column 1) [NEW]
  - Gives a chance to parry enemy melee attacks, reduces all threat generated by your attacks by 30% while Rockbiter Weapon is not active, and increases all threat generated by 30% while Rockbiter Weapon is active.
- **Mental Quickness** (2 ranks, column 2) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases your spell damage and healing by up to 15% of your Intellect.
  - Rank 2: Increases your spell damage and healing by up to 30% of your Intellect.
- **Improved Stormstrike** (2 ranks, column 3) [NEW] [ranks settled by the 70338 trait curves]
  - Requires: Stormstrike (1/1)
  - Rank 1: When you Stormstrike, you have a 50% chance to gain 50% mana regeneration while casting spells for 15 sec, and Stormstrike's cooldown has a 50% chance to reset each time you Dodge or Parry.
  - Rank 2: When you Stormstrike, you have a 100% chance to gain 100% mana regeneration while casting spells for 30 sec, and Stormstrike's cooldown has a 100% chance to reset each time you Dodge or Parry.

#### Row 6 (25 points)

- **Maelstrom Weapon** (5 ranks, column 3) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: When you deal damage with a melee attack, you have a chance to reduce the cast time and Mana cost of your next Lightning Bolt spell by 4%. Stacks up to 5 times. Lasts 30 sec.
  - Rank 2: When you deal damage with a melee attack, you have a chance to reduce the cast time and Mana cost of your next Lightning Bolt spell by 8%. Stacks up to 10 times. Lasts 60 sec.
  - Rank 3: When you deal damage with a melee attack, you have a chance to reduce the cast time and Mana cost of your next Lightning Bolt spell by 12%. Stacks up to 15 times. Lasts 90 sec.
  - Rank 4: When you deal damage with a melee attack, you have a chance to reduce the cast time and Mana cost of your next Lightning Bolt spell by 16%. Stacks up to 20 times. Lasts 120 sec.
  - Rank 5: When you deal damage with a melee attack, you have a chance to reduce the cast time and Mana cost of your next Lightning Bolt spell by 20%. Stacks up to 25 times. Lasts 150 sec.

#### Row 7 (30 points)

- **Rage of the Farseer** (1 rank, column 2) [NEW]
  - Requires: Mental Quickness (2/2)
  - Cost: Instant, 3 min cooldown
  - Increases your melee attack speed and spell casting speed by 30% for 25 sec.
  - Note: Bloodlust-style haste capstone. Replaced ability_hunter_pet_wolf (Feral Spirit / Ghost Wolf) with spell_nature_bloodlust.

### Restoration

16 talents, 3 new. 31-point talent: Riptide.

#### Row 1 (0 points)

- **Improved Healing Wave** (5 ranks, column 2) [unchanged]
  - Rank 1: Reduces the casting time of your Healing Wave spell by 0.1 sec.
  - Rank 2: Reduces the casting time of your Healing Wave spell by 0.2 sec.
  - Rank 3: Reduces the casting time of your Healing Wave spell by 0.3 sec.
  - Rank 4: Reduces the casting time of your Healing Wave spell by 0.4 sec.
  - Rank 5: Reduces the casting time of your Healing Wave spell by 0.5 sec.
- **Tidal Mastery** (5 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases the critical effect chance of your healing spells by 1%.
  - Rank 2: Increases the critical effect chance of your healing spells by 2%.
  - Rank 3: Increases the critical effect chance of your healing spells by 3%.
  - Rank 4: Increases the critical effect chance of your healing spells by 4%.
  - Rank 5: Increases the critical effect chance of your healing spells by 5%.
  - Classic (5 ranks):
    - Rank 1: Increases the critical effect chance of your healing and lightning spells by 1%.
    - Rank 2: Increases the critical effect chance of your healing and lightning spells by 2%.
    - Rank 3: Increases the critical effect chance of your healing and lightning spells by 3%.
    - Rank 4: Increases the critical effect chance of your healing and lightning spells by 4%.
    - Rank 5: Increases the critical effect chance of your healing and lightning spells by 5%.
  - Note: Slot from the third-party dataset, not yet verified in footage by us.

#### Row 2 (5 points)

- **Mindfulness** (3 ranks, column 1) [NEW] [ranks settled by the 70338 trait curves]
  - Rank 1: Allows 17% of your Mana regeneration to continue while casting.
  - Rank 2: Allows 34% of your Mana regeneration to continue while casting.
  - Rank 3: Allows 51% of your Mana regeneration to continue while casting.
- **Natural Grace** (3 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Reduces the threat generated by your spells by 5%.
  - Rank 2: Reduces the threat generated by your spells by 10%.
  - Rank 3: Reduces the threat generated by your spells by 15%.
  - Classic (3 ranks):
    - Rank 1: Reduces the threat generated by your healing spells by 5%.
    - Rank 2: Reduces the threat generated by your healing spells by 10%.
    - Rank 3: Reduces the threat generated by your healing spells by 15%.
  - Note: No arrow into this talent in the footage.
- **Tidal Focus** (5 ranks, column 3) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Reduces the Mana cost of your healing spells by 1% and increases your chance to hit with all spells and attacks by 1%.
  - Rank 2: Reduces the Mana cost of your healing spells by 2% and increases your chance to hit with all spells and attacks by 2%.
  - Rank 3: Reduces the Mana cost of your healing spells by 3% and increases your chance to hit with all spells and attacks by 3%.
  - Rank 4: Reduces the Mana cost of your healing spells by 4% and increases your chance to hit with all spells and attacks by 4%.
  - Rank 5: Reduces the Mana cost of your healing spells by 5% and increases your chance to hit with all spells and attacks by 5%.
  - Classic (5 ranks):
    - Rank 1: Reduces the Mana cost of your healing spells by 1%.
    - Rank 2: Reduces the Mana cost of your healing spells by 2%.
    - Rank 3: Reduces the Mana cost of your healing spells by 3%.
    - Rank 4: Reduces the Mana cost of your healing spells by 4%.
    - Rank 5: Reduces the Mana cost of your healing spells by 5%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Reincarnation** (2 ranks, column 4) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Reduces the cooldown of your Reincarnation spell by 10 min, increases your maximum health by 2%, and increases the amount of health and Mana you reincarnate with by an additional 10%.
  - Rank 2: Reduces the cooldown of your Reincarnation spell by 20 min, increases your maximum health by 4%, and increases the amount of health and Mana you reincarnate with by an additional 20%.
  - Classic (2 ranks):
    - Rank 1: Reduces the cooldown of your Reincarnation spell by 10 min and increases the amount of health and mana you reincarnate with by an additional 10%.
    - Rank 2: Reduces the cooldown of your Reincarnation spell by 20 min and increases the amount of health and mana you reincarnate with by an additional 20%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Ancestral Healing** (3 ranks, column 1) [unchanged]
  - Rank 1: Increases your target's armor value by 8% for 15 sec after getting a critical effect from one of your healing spells.
  - Rank 2: Increases your target's armor value by 16% for 15 sec after getting a critical effect from one of your healing spells.
  - Rank 3: Increases your target's armor value by 25% for 15 sec after getting a critical effect from one of your healing spells.
- **Healing Focus** (3 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Gives you a 23% chance to avoid interruption caused by damage while casting any healing spell.
  - Rank 2: Gives you a 46% chance to avoid interruption caused by damage while casting any healing spell.
  - Rank 3: Gives you a 69% chance to avoid interruption caused by damage while casting any healing spell.
  - Classic (5 ranks):
    - Rank 1: Gives you a 14% chance to avoid interruption caused by damage while casting any healing spell.
    - Rank 2: Gives you a 28% chance to avoid interruption caused by damage while casting any healing spell.
    - Rank 3: Gives you a 42% chance to avoid interruption caused by damage while casting any healing spell.
    - Rank 4: Gives you a 56% chance to avoid interruption caused by damage while casting any healing spell.
    - Rank 5: Gives you a 70% chance to avoid interruption caused by damage while casting any healing spell.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Water Shield** (1 rank, column 3) [NEW]
  - Cost: Instant, 15 sec cooldown
  - The caster is surrounded by 3 globes of water. When a spell, melee, or ranged attack hits the caster or when one of the caster's healing spells gets a critical result, 2% of maximum mana is restored to the caster, expending one water globe. Only one globe will activate every few seconds. Lasts 10 min. Only one Elemental Shield can be active on the Shaman at any one time.
  - Note: Instant, 15 sec cooldown.

#### Row 4 (15 points)

- **Totemic Focus** (5 ranks, column 1) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Reduces the Mana cost of your totems and any spells that summon or move them by 5%.
  - Rank 2: Reduces the Mana cost of your totems and any spells that summon or move them by 10%.
  - Rank 3: Reduces the Mana cost of your totems and any spells that summon or move them by 15%.
  - Rank 4: Reduces the Mana cost of your totems and any spells that summon or move them by 20%.
  - Rank 5: Reduces the Mana cost of your totems and any spells that summon or move them by 25%.
  - Classic (5 ranks):
    - Rank 1: Reduces the Mana cost of your totems by 5%.
    - Rank 2: Reduces the Mana cost of your totems by 10%.
    - Rank 3: Reduces the Mana cost of your totems by 15%.
    - Rank 4: Reduces the Mana cost of your totems by 20%.
    - Rank 5: Reduces the Mana cost of your totems by 25%.
  - Note: Slot from the third-party dataset, not yet verified in footage by us.
- **Restorative Totems** (5 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Rank 1: Increases the effect of your Mana Spring Totem by 5% and increases the effect of your Healing Stream Totem by 10%.
  - Rank 2: Increases the effect of your Mana Spring Totem by 10% and increases the effect of your Healing Stream Totem by 20%.
  - Rank 3: Increases the effect of your Mana Spring Totem by 15% and increases the effect of your Healing Stream Totem by 30%.
  - Rank 4: Increases the effect of your Mana Spring Totem by 20% and increases the effect of your Healing Stream Totem by 40%.
  - Rank 5: Increases the effect of your Mana Spring Totem by 25% and increases the effect of your Healing Stream Totem by 50%.
  - Classic (5 ranks):
    - Rank 1: Increases the effect of your Mana Spring and Healing Stream Totems by 5%.
    - Rank 2: Increases the effect of your Mana Spring and Healing Stream Totems by 10%.
    - Rank 3: Increases the effect of your Mana Spring and Healing Stream Totems by 15%.
    - Rank 4: Increases the effect of your Mana Spring and Healing Stream Totems by 20%.
    - Rank 5: Increases the effect of your Mana Spring and Healing Stream Totems by 25%.
- **Mana Tide Totem** (1 rank, column 3) [CHANGED]
  - Cost: 10 Mana, Instant, 5 min cooldown, Tools: Water Totem
  - Summons a Mana Tide Totem with 5 health at the feet of the caster for 12 sec that restores 88 mana every 3 seconds to group members within 30 yards.
  - Classic: Summons a Mana Tide Totem with 5 health at the feet of the caster for 12 sec that restores 170 mana every 3 seconds to group members within 20 yards.
  - Note: Slot per the third-party dataset, not yet hovered in footage we have.

#### Row 5 (20 points)

- **Healing Way** (3 ranks, column 2) [CHANGED] [ranks settled by the 70338 trait curves]
  - Requires: Restorative Totems (5/5)
  - Rank 1: Increases the amount healed by your Healing Wave spell by 8%.
  - Rank 2: Increases the amount healed by your Healing Wave spell by 16%.
  - Rank 3: Increases the amount healed by your Healing Wave spell by 24%.
  - Classic (3 ranks):
    - Rank 1: Your Healing Wave spells have a 33% chance to increase the effect of subsequent Healing Wave spells on that target by 6% for 15 sec. This effect will stack up to 3 times.
    - Rank 2: Your Healing Wave spells have a 66% chance to increase the effect of subsequent Healing Wave spells on that target by 6% for 15 sec. This effect will stack up to 3 times.
    - Rank 3: Your Healing Wave spells have a 100% chance to increase the effect of subsequent Healing Wave spells on that target by 6% for 15 sec. This effect will stack up to 3 times.
  - Note: Slot per the third-party dataset, the arrow from Restorative Totems runs into it.
- **Nature's Swiftness** (1 rank, column 3) [unchanged]
  - Cost: Instant, 3 min cooldown
  - When activated, your next Nature spell with a casting time less than 10 sec. becomes an instant cast spell.

#### Row 6 (25 points)

- **Purification** (5 ranks, column 3) [unchanged]
  - Rank 1: Increases the effectiveness of your healing spells by 2%.
  - Rank 2: Increases the effectiveness of your healing spells by 4%.
  - Rank 3: Increases the effectiveness of your healing spells by 6%.
  - Rank 4: Increases the effectiveness of your healing spells by 8%.
  - Rank 5: Increases the effectiveness of your healing spells by 10%.

#### Row 7 (30 points)

- **Riptide** (1 rank, column 2) [NEW]
  - Requires: Healing Way (3/3)
  - Cost: 245 Mana, 40 yd range, Instant, 6 sec cooldown
  - Heals a friendly target for 479 to 528, an additional 499 over 15 sec, and increases the effectiveness of your Chain Heal casts directly on that target by 25%.
  - Note: 245 Mana, 40 yd range, instant, 6 sec cooldown. Capstone, the arrow from the slot above leads here.
