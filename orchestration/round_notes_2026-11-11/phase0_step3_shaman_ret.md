# Phase 0 step 3: shaman_ret (Enhancement, Elemental, Retribution)

Full searches of 2026-11-23 on 4 cores (about 2h15m a spec). The dry runs found no new unmodelled or
hidden stat items since step 2. The pools only shrank (items tagged out since then). Soulforge Armor
(set 1673) went into `phase0_content.json` before the Retribution search, by the parent's ruling: the
18 rows the client cache marks valid. No Soulforge head or chest row is valid, so Retribution can reach
6 pieces only with belt and legs. The Five Thunders head and chest rows are now doubtful, as TASKS.md
2026-11-20 says.

Set checks and calibration are `rotopt -eval` at 180s and 10000 iterations. Set swaps put each slot's
best non-set item in, with the old enchant and no random suffix. The search's own list `dps` runs 10
to 15 above `-eval` of the same set, as the mage notes found (uses scored on 90s).

Tooling: `bis_phase0.sh` doesn't create `ui/<dir>/bis/`, so the first Elemental search ran 2h and then
failed to write its list. Fixed with a `mkdir -p` before the run.

## Enhancement

List `ui/enhancement_shaman/bis/level60_phase0.bis.json`, talents 053303102-054030110051212141-03 (the
user's), Orc. Score 1001.4.

| Slot | Pick | Source |
|---|---|---|
| Head | Black Dragonscale Helm | AH ~296g, or Leatherworking 300 |
| Neck | Beads of Ogre Might | Quest Falrin's Vendetta, Dire Maul |
| Shoulder | Spaulders of The Five Thunders | Mokvar, Anthion chain |
| Back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| Chest | Chain of The Five Thunders | Mokvar, end of the Valthalak chain (doubtful) |
| Wrist | Bands of The Five Thunders | Mokvar bracer quest, 20g |
| Hands | Gauntlets of The Five Thunders (caster variant) | Mokvar gloves and belt quest |
| Waist | Belt of the Archmage | Tailoring 295, pattern on AH ~131g |
| Legs | Kilt of The Five Thunders (caster variant) | Mokvar, Anthion chain |
| Feet | Treads of The Five Thunders | Mokvar, Anthion chain |
| Rings | Channeler's Ring, Watcher's Signet of Brilliance | Pix Xizzix 75 Reals; Watchers Revered (doubtful) |
| Trinkets | Blackhand's Breadth, Chained Essence of Eranikus | BRS quest; Sunken Temple quest |
| Weapon | Brain Hacker | World drop, AH ~215g |
| Totem | Totem of the Storm | Pix Xizzix, 50 Reals |

Enchants: Arcanum of Rapidity (head, Pristine Black Diamond ~75g), Lesser Arcanum of Voracity Strength
(legs), +5 Strength neck, +5 Agility cloak, +4 stats chest, +16 spell power bracer, +15 Strength gloves,
+7 Agility boots, Crusader.

Flags:
- Chain of The Five Thunders is doubtful (after the summoning chain). Without it, Timbermaw Tunic plus
  Cord of The Five Thunders in the waist keeps the 6pc: 983.8 against 987.0.
- Watcher's Signet of Brilliance is doubtful (no Watchers rep source found). It ties Channeler's Ring.
  The best ring that isn't doubtful loses 1.95 (Eye of Orgrimmar).
- Blackhand's Breadth: BRS quest reward on Wowhead Forever, quest not in our data.
- The Five Thunders 4pc proc rate (The Furious Storm) is unconfirmed (5% from a guide).
- No pick has an unmodelled effect or hidden stats.

Set check (search talents): 6pc as picked 987.0, 4pc (Skyfury Gauntlets, Pristine Scorpid Leggings)
981.9, 2pc (also Timbermaw Tunic, Bloodmail Boots) 985.3. The search's 6pc stands.

Comparison, timesaver.gg Forever Enhancement (data of 2026-11-13, before the Five Thunders reveal of
2026-11-17):
- Their Savage Gladiator Shoulders and Chain, Slashclaw Bracers, Devilsaur Gauntlets and Leggings and
  Windreaver Greaves give way to the Five Thunders 6pc. Ours lose 10.6 to 17.5 each, or rank outside the
  top 25.
- Cape of the Black Baron -14.3, Might of the Timbermaw -7.9, Smolderweb's Eye -13.5, Rune of the Guard
  Captain -13.3. Mask of the Unforgiven, Postmaster's Seal and Blackstone Ring rank outside the top 25.
  Ours are Forever items (Reals vendor) or value the spell power the Five Thunders proc and the shocks use.
- Cleansed Vilebranch Medallion is out (Alliance only quest). Forest Defender's Axe is out (raid essence,
  Dec 9).

Calibration (report build, report talents): Phase 0 996.2, synthetic 1071.3 (report 1070.1), gap -75.1
(-7.0%).

## Retribution

List `ui/retribution_paladin/bis/level60_phase0_ret.bis.json`, report talents, Undead. Score 973.8.

| Slot | Pick | Source |
|---|---|---|
| Head | Lionheart Helm | Blacksmithing 300 Armorsmith, plans 0.2% drop (doubtful) |
| Neck | Mark of Fordring | Quest In Dreams, Plaguelands |
| Shoulder | Soulforge Spaulders | Soulforge quest, Anthion chain |
| Back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| Chest | Dawn Armor | AH ~350g, or Leatherworking |
| Wrist | Soulforge Bracers | Soulforge bracer quest |
| Hands | Soulforge Gauntlets | Soulforge gloves and belt quest |
| Waist | Belt of the Archmage | Tailoring 295, pattern on AH ~131g |
| Legs | Pristine Scorpid Leggings | Leatherworking 300, pattern on AH ~134g |
| Feet | Soulforge Warboots | Soulforge quest, Anthion chain |
| Rings | Watcher's Signet of Brilliance, Channeler's Ring | Watchers Revered (doubtful); Pix Xizzix 75 Reals |
| Trinkets | Chained Essence of Eranikus, Blackhand's Breadth | Sunken Temple quest; BRS quest |
| Weapon | Destiny | Random world drop |
| Libram | Libram of Law | Pix Xizzix |

Enchants: Arcanum of Rapidity on head and legs, +6 spell power neck, +5 Agility cloak, +4 stats chest,
+16 spell power bracer, +15 Strength gloves, +7 Agility boots, Spellblasting (+65 spell power) on the
weapon.

Flags:
- Lionheart Helm is doubtful (plans 0.2% drop). Stalwart Helm loses 1.89.
- Watcher's Signet of Brilliance is doubtful. The best ring that isn't loses 3.0 (Cutthroat's Signet).
- Destiny: random world drop epic with no AH price. Blackblade of Shahram (Drakkisath) loses 8.45 and
  Brain Hacker 8.91.
- Soulforge: the quest NPC isn't in our data and the step gold is unknown. The 4pc proc (Crusader's
  Wrath) uses Classic Lightforge's rate, unconfirmed for Forever.
- Blackhand's Breadth: quest not in our data, as for Enhancement.
- No pick has an unmodelled effect or hidden stats.

Set checks:
- Soulforge: 4pc as picked 958.0. 6pc (Soulforge Belt and Legplates) 948.0, 5pc 951.3 (belt) or 945.4
  (legs), 2pc (Vambraces of the Sadist, Heartflame Gauntlets) 943.9. The search's 4pc stands.
- Blessed Plate (set 1968, Forever Blacksmithing, timesaver's pick): their 5pc 918.6, with legs 906.2,
  with head 903.3. 3pc next to our Soulforge 4pc 916.1. All lose 40 or more, since the pieces are item
  level 52 to 61.

Comparison, timesaver.gg Forever Retribution (data of 2026-11-13):
- Lionheart Helm matches.
- Their 5 Blessed Plate pieces lose 40 as a set (above).
- Titanic Leggings -7.1 (doubtful, plans have no Forever source), Devilsaur Gauntlets -21.4, Cape of the
  Black Baron -15.2, Blackstone Ring -4.1, Rune of the Guard Captain -11.0. Beads of Ogre Might,
  Postmaster's Seal and Smolderweb's Eye rank outside the top 25.
- Cleansed Vilebranch Medallion and Forest Defender's Axe are out, as for Enhancement.

Calibration (report build): Phase 0 958.0, synthetic 1027.1 (report 1026.4), gap -69.1 (-6.7%).

## Elemental

List `ui/elemental_shaman/bis/level60_phase0_ele.bis.json`, report talents, Windshaper. Score 877.0.

| Slot | Pick | Source |
|---|---|---|
| Head | Spellweaver's Turban | General Drakkisath, UBRS |
| Neck | Orb of the Darkmoon | Darkmoon Faire, 1200 tickets (Faire weeks Nov 9-16, Dec 7-14) |
| Shoulder | Pauldrons of The Five Thunders | Mokvar, Anthion chain |
| Back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| Chest | Vest of The Five Thunders | Mokvar, end of the Valthalak chain (doubtful) |
| Wrist | Bindings of The Five Thunders | Mokvar bracer quest, 20g |
| Hands | Gauntlets of The Five Thunders | Mokvar gloves and belt quest |
| Waist | Cord of The Five Thunders | Mokvar gloves and belt quest |
| Legs | Ironfeather Leggings | AH ~121g, or Leatherworking 300 |
| Feet | Slippers of The Five Thunders | Mokvar, Anthion chain |
| Rings | Channeler's Ring, Watcher's Signet of Brilliance | Pix Xizzix 75 Reals; Watchers Revered (doubtful) |
| Trinkets | Eye of the Beast, Chained Essence of Eranikus | BRS quest; Sunken Temple quest |
| Main hand | Hand of Edward the Odd | Random world drop |
| Off hand | Obsidian Bauble | Pusillin, Dire Maul |
| Totem | Burning Totem | Pix Xizzix, 50 Reals |

Enchants: Arcanum of Focus on head and legs, +6 spell power neck, Mystic Rugged Armor Kit on chest and
feet, +16 spell power bracer, +20 Nature damage gloves, Revelation on the weapon, +6 Intellect +5 Spirit
off hand. The list has no cloak enchant.

Flags:
- Vest of The Five Thunders is doubtful (after the summoning chain). Without it the best set is a 4pc
  with Alanna's Embrace and Belt of the Archmage, 855.4 against 860.7.
- Watcher's Signet of Brilliance is doubtful. The best ring that isn't loses 2.96 (Maiden's Circle).
- Hand of Edward the Odd: random world drop with no AH price. Witchblade (Gandling) loses 7.35. Its
  melee proc doesn't matter to a caster.
- Orb of the Darkmoon needs a Faire week. The next one is Dec 7-14. Barbed Thorn Necklace loses 1.63.
- Eye of the Beast: BRS quest reward, quest not in our data.
- No pick has an unmodelled effect or hidden stats.

Set check (Five Thunders): 6pc as picked 860.7. 6pc with Kilt for the doubtful Vest (Alanna's Embrace in
the chest) 853.3, 4pc (Alanna's Embrace, Belt of the Archmage) 855.4, 2pc (also Adventurer's Boots of
Nature's Wrath, Skyfury Gauntlets) 828.0. The search's 6pc stands.

Comparison, timesaver.gg Forever Elemental (data of 2026-11-13). Its URL is a leveling-gear page, so it
may not mean to be a level 60 list:
- Orb of the Darkmoon matches.
- Their Elements pieces (Coif, Vest, Gauntlets, Boots), Ironfeather Shoulders and Brazecore Armguards
  give way to the Five Thunders 6pc and rank outside our top 25.
- Barrage Girdle -13.8, Postmaster's Seal -6.5, Mastersmith's Hammer -8.6. Blue Dragonscale Leggings and
  Observer's Shield rank outside the top 25.
- Cloak of Earth and Sky is doubtful (pattern has no known source). Don Mauricio's Band is infeasible
  (379g a DPS). Cleansed Vilebranch Icon is out (Alliance only quest).
- Heart of the Primeval is item level 40, below the pool's floor of 50, so it was never simmed.
  Jewel of the Guard Captain ranks outside the top 25.

Calibration (report build): Phase 0 860.7, synthetic 981.9 (report 981.8), gap -121.2 (-12.3%).

## Calibration summary

| Spec | Phase 0 | Synthetic | Gap |
|---|---|---|---|
| Enhancement | 996.2 | 1071.3 | -75.1 (-7.0%) |
| Elemental | 860.7 | 981.9 | -121.2 (-12.3%) |
| Retribution | 958.0 | 1027.1 | -69.1 (-6.7%) |

The synthetic Phase 1 gear sits 7% above the Phase 0 sets for the two melee specs and 12% for
Elemental, close to the mages' 13 to 14%. So the synthetic gear flatters the casters by about 5% next to
Enhancement and Retribution. On Phase 0 gear the melee specs would rank about 5% higher against them.
