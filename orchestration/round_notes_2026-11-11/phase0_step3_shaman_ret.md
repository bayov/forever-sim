# Phase 0 step 3: shaman_ret (Enhancement, Elemental, Retribution)

Rerun of 2026-11-23/24 (briefs/phase0_rerun_1123.md) on master 5e0a327fe and later: the item proc and
data fixes, Blackhand's Breadth and Eye of the Beast crit split, weapon skill, the faction check by race
and the Exalted rewards tagged infeasible (b1bbe4347). The dry runs found no new unmodelled or hidden
stat items, and nothing joined any pool. The procs now modelled only left the unmodelled lists.

Set checks and calibration are `rotopt -eval` at 180s and 10000 iterations. The search's own list `dps`
runs about 10 above `-eval` of the same set (uses scored on 90s).

Tooling: this container had no `jj`, which `bis_phase0.sh` calls for the sim hash, so a shim gave it the
git commit. The container was reclaimed twice while idle and killed the Elemental search each time, so
the third run was kept busy from a waiting subagent.

## Enhancement

List `ui/enhancement_shaman/bis/level60_phase0.bis.json`, talents 053303102-054030110051212141-03 (the
user's), Orc. Score 994.0 (was 1001.4).

| Slot | Pick | Source |
|---|---|---|
| Head | Black Dragonscale Helm | AH ~296g, or Leatherworking 300 |
| Neck | Beads of Ogre Might | Quest Falrin's Vendetta, Dire Maul |
| Shoulder | Spaulders of The Five Thunders | Mokvar, Anthion chain |
| Back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| Chest | Chain of The Five Thunders | Mokvar, end of the Valthalak chain (doubtful) |
| Wrist | Bands of The Five Thunders | Mokvar bracer quest, 20g |
| Hands | Gauntlets of The Five Thunders | Mokvar gloves and belt quest |
| Waist | Belt of the Archmage | Tailoring 295, pattern on AH ~131g |
| Legs | Kilt of The Five Thunders | Mokvar, Anthion chain |
| Feet | Treads of The Five Thunders | Mokvar, Anthion chain |
| Rings | Watcher's Signet of Brilliance, Channeler's Ring | Watchers Revered (doubtful); Pix Xizzix 75 Reals |
| Trinkets | Chained Essence of Eranikus, Adaptive Combat Assistant | Sunken Temple quest; Pix Xizzix 125 Reals |
| Weapon | Brain Hacker | AH ~215g |
| Totem | Totem of the Storm | Pix Xizzix, 50 Reals |

Enchants: Arcanum of Rapidity (head), Lesser Arcanum of Voracity Strength (legs), +5 Strength neck,
+5 Agility cloak, +4 stats chest, +16 spell power bracer, +15 Strength gloves, +7 Agility boots, Crusader.

The full search stopped at a Five Thunders 4pc with Devilsaur Gauntlets and Leggings (987.0). It swaps
one slot at a time, and each Five Thunders piece alone loses about 9 there, so it never reached the 6pc.
By `-eval` the 6pc was 7.7 better (985.0 against 977.3), so the list was ranked again (RANK_ONLY) from
the 6pc. That ranking kept it.

Flags:
- Chain of The Five Thunders is doubtful (after the summoning chain). Timbermaw Tunic in its place loses
  7.05. Timbermaw Tunic with Cord of The Five Thunders in the waist keeps the 6pc and loses about 2.6.
- Watcher's Signet of Brilliance is doubtful (no Watchers rep source found). It ties Channeler's Ring.
  The next ring that isn't doubtful or infeasible is Amalgam's Band, -1.52 (Kormok summon unconfirmed).
- Adaptive Combat Assistant (Forever, modelled now): a 450 absorb shield that hits Nature in a cone if
  broken early. It is worth 0.4 DPS (1.6 against 4 mobs), so it is filler. Blackhand's Breadth is 0.21
  behind and Darkmoon Card: Maelstrom (doubtful, card drops unconfirmed) 1.3 ahead, both inside the
  search's noise.
- Blackhand's Breadth: BRS quest reward on Wowhead Forever, quest not in our data.
- The Five Thunders 4pc proc rate (The Furious Storm) is unconfirmed (5% from a guide).
- No pick has an unmodelled effect or hidden stats.

Set check (`-eval`, search talents): 6pc as picked 984.1 (with Don Julio's Band, before it went
infeasible, 985.0). 4pc with Devilsaur Gauntlets and Leggings 977.3. 5pc with only the Kilt or only the
Gauntlets 967.7 and 968.2. The 6pc stands.

Comparison, timesaver.gg Forever Enhancement (data of 2026-11-13), from the list's own ranks:
- Devilsaur Gauntlets and Leggings came close (the search's 4pc) but lose to the 6pc. Savage Gladiator
  Chain -13.5. Savage Gladiator Shoulders, Slashclaw Bracers, Smolderweb's Eye, Postmaster's Seal and
  Blackstone Ring rank outside the top 25. Windreaver Greaves -12.3.
- Cape of the Black Baron -14.2, Might of the Timbermaw -6.9, Rune of the Guard Captain -6.6, Mask of the
  Unforgiven -12.7.
- Cleansed Vilebranch Medallion is out (Alliance only quest). Forest Defender's Axe is out (raid essence,
  Dec 9).

Calibration (report build, report talents): Phase 0 993.9, synthetic 1071.3, gap -77.4 (-7.2%).

Changed from the earlier list: Blackhand's Breadth went to Adaptive Combat Assistant (Breadth's crit is
melee only now, so it lost its shock crit, and the Assistant's proc is modelled now). Score fell 7.4 from
the proc and crit fixes. Picks otherwise the same.

## Retribution

List `ui/retribution_paladin/bis/level60_phase0_ret.bis.json`, report talents, Undead (Horde). Score
975.1 (was 973.8).

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

The full search picked Don Julio's Band. It went infeasible (Exalted, b1bbe4347) during the run, so the
rings were ranked again (SLOTS=finger).

Flags:
- Lionheart Helm is doubtful (plans 0.2% drop). Stalwart Helm loses 1.69.
- Watcher's Signet of Brilliance is doubtful. It ties Channeler's Ring. Without it the best pair is
  Channeler's Ring and Cutthroat's Signet, -2.97.
- Destiny: random world drop epic with no AH price. Warblade of Caer Darrow (Scholomance quest) loses
  5.54, Blackblade of Shahram 8.52, Brain Hacker 9.15.
- Soulforge: the quest NPC isn't in our data and the step gold is unknown. The 4pc proc (Crusader's
  Wrath) uses Classic Lightforge's rate, unconfirmed for Forever.
- Blackhand's Breadth: quest not in our data, as for Enhancement.
- Two ranked items had no content entry and got one: Warblade of Caer Darrow (open, Scholomance quest)
  and Furbolg Berserker's Boots (doubtful and included, drop source unknown). Neither beats its pick.
- No pick has an unmodelled effect or hidden stats.

Set checks (`-eval`): Soulforge 4pc as picked 957.3. 6pc (Soulforge Belt and Legplates) 947.2, 5pc
950.5 (belt) or 944.6 (legs), 2pc (Vambraces of the Sadist, Heartflame Gauntlets) 943.1. The 4pc stands.
Blessed Plate (timesaver's pick) lost 40 or more in the earlier check, and its item levels haven't moved.

Comparison, timesaver.gg Forever Retribution (data of 2026-11-13):
- Lionheart Helm matches. Their Blessed Plate pieces lose 40 or more as a set.
- Titanic Leggings -7.2 (doubtful, plans have no Forever source), Devilsaur Gauntlets -21.6, Cape of the
  Black Baron -15.3, Blackstone Ring -4.0, Rune of the Guard Captain -10.3. Beads of Ogre Might,
  Postmaster's Seal and Smolderweb's Eye rank outside the top 25.
- Cleansed Vilebranch Medallion and Forest Defender's Axe are out, as for Enhancement.

Calibration (report build): Phase 0 957.3, synthetic 1027.1, gap -69.8 (-6.8%).

Changed from the earlier list: no pick or enchant changed. The fixes only moved the ranks below the picks.

## Elemental

List `ui/elemental_shaman/bis/level60_phase0_ele.bis.json`, report talents, Windshaper (Horde). Score
870.6 (was 877.0).

| Slot | Pick | Source |
|---|---|---|
| Head | Spellweaver's Turban | General Drakkisath, UBRS |
| Neck | Orb of the Darkmoon | Darkmoon Faire, 1200 tickets (Faire weeks Nov 9-16, Dec 7-14) |
| Shoulder | Pauldrons of The Five Thunders | Mokvar, Anthion chain |
| Back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| Chest | Vest of The Five Thunders | Mokvar, end of the Valthalak chain (doubtful) |
| Wrist | Bindings of The Five Thunders | Mokvar bracer quest, 20g |
| Hands | Gauntlets of The Five Thunders | Mokvar gloves and belt quest |
| Waist | Girdle of The Five Thunders | Mokvar gloves and belt quest |
| Legs | Ironfeather Leggings | AH ~121g, or Leatherworking 300 |
| Feet | Slippers of The Five Thunders | Mokvar, Anthion chain |
| Rings | Rune Band of Wizardry, Channeler's Ring | Lord Valthalak (summon unconfirmed, included per standards 42); Pix Xizzix 75 Reals |
| Trinkets | Chained Essence of Eranikus, Eye of the Beast | Sunken Temple quest; BRS quest |
| Main hand | Hand of Edward the Odd | Random world drop |
| Off hand | Obsidian Bauble | Pusillin, Dire Maul |
| Totem | Burning Totem | Pix Xizzix, 50 Reals |

Enchants: Arcanum of Focus on head and legs, +6 spell power neck, Mystic Rugged Armor Kit on chest and
feet, +16 spell power bracer, +20 Nature damage gloves, Revelation on the weapon, +6 Intellect +5 Spirit
off hand. The list has no cloak enchant.

Flags:
- Vest of The Five Thunders is doubtful (after the summoning chain). Without it the best set is a 4pc
  with Alanna's Embrace and Belt of the Archmage, 854.0 against 859.4.
- Rune Band of Wizardry needs the Valthalak summon, unconfirmed in Forever. Eye of Orgrimmar in its place
  loses 1.1 (`-eval`), Watcher's Signet of Brilliance (doubtful) 1.0.
- Girdle and Cord of The Five Thunders tie (859.4 and 859.5 by `-eval`). Either works.
- Hand of Edward the Odd: random world drop with no AH price. Witchblade (Gandling) loses 7.54.
- Orb of the Darkmoon needs a Faire week. The next one is Dec 7-14. Barbed Thorn Necklace loses 1.38.
- Eye of the Beast: BRS quest reward, quest not in our data.
- Weakness Analyzer (Pix Xizzix, 125 Reals) and Royal Seal of Eldre'Thalas rank 0.68 and 0.16 above the
  trinket picks, inside the search's noise.
- No pick has an unmodelled effect or hidden stats.

Set check (Five Thunders, `-eval`): 6pc as picked 859.4. Alanna's Embrace for the Vest (5pc) 847.9, 6pc
with the Kilt for the Vest (Alanna's in the chest) 852.2, 4pc (Alanna's, Belt of the Archmage) 854.0.
The search's 6pc stands.

Comparison, timesaver.gg Forever Elemental (data of 2026-11-13, a leveling-gear page):
- Orb of the Darkmoon matches.
- Their Elements pieces give way to the Five Thunders 6pc. Vest of Elements -15.7, the Coif, Gauntlets
  and Boots, Ironfeather Shoulders and Brazecore Armguards rank outside our top 25.
- Barrage Girdle -14.9, Mastersmith's Hammer -8.5, Jewel of the Guard Captain -3.6. Postmaster's Seal,
  Blue Dragonscale Leggings and Observer's Shield rank outside the top 25.
- Cloak of Earth and Sky is doubtful (pattern has no known source). Don Mauricio's Band and Mindfang are
  infeasible (Exalted). Cleansed Vilebranch Icon is out (Alliance only quest).

Calibration (report build): Phase 0 859.4, synthetic 981.9, gap -122.5 (-12.5%).

Changed from the earlier list: Watcher's Signet of Brilliance went to Rune Band of Wizardry, and Cord to
Girdle of The Five Thunders (a tie). Score fell 6.4 from the proc fixes and Eye of the Beast's crit
split.

## Calibration summary

| Spec | Phase 0 | Synthetic | Gap | Earlier gap |
|---|---|---|---|---|
| Enhancement | 993.9 | 1071.3 | -77.4 (-7.2%) | -75.1 (-7.0%) |
| Elemental | 859.4 | 981.9 | -122.5 (-12.5%) | -121.2 (-12.3%) |
| Retribution | 957.3 | 1027.1 | -69.8 (-6.8%) | -69.1 (-6.7%) |

The gaps barely moved. The synthetic Phase 1 gear still sits 7% above the Phase 0 sets for the melee
specs and 12.5% for Elemental, so it flatters the casters by about 5% next to them.
