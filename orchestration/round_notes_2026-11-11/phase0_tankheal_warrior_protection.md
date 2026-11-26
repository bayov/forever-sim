# Phase 0 tank and heal: warrior_protection

Branch phase0-tankheal-warrior_protection. Code: phase0-tankheal through f219400ab (threat batch, items, final go, Shield Slam main-hand special, Premier sets dropped, Force Reactive Disk out).
The presets are Undead (Horde), Engineering and Blacksmithing, talents 05-05050005-250530020310001351.

## Scores (step 9)

The mean of the six scored bosses (magmadar, golemagg, onyxia, broodlord, vaelastrasz, chromaggus), in the preset's own column, at 4000 iterations from seed 1. That is more than the brief's 1000, for less noise. Synthetic is `git show origin/master:<req>` on the same code.

| Preset | Phase 0 | Synthetic | Gap |
|---|---|---|---|
| Balanced | 3379.7 | 3262.8 | +116.9 |
| Max threat | 4727.2 | 4110.0 | +617.2 |
| Max survival | 1203.8 | 1617.8 | -414.0 |

The Phase 0 set has more threat than the synthetic set but less mitigation and health. Survival loses most on onyxia, broodlord and chromaggus.
The search's own figures (3299.4, 4699.7, 1202.2) use the search's scoring, not the bench's.

## Picks

All three presets use Iceblade Hacker and Evergreen Shield. Force Reactive Disk was the shield until it went out (its schematic drops in Molten Core). Losing it cost 34 to 58 in the search.
- Max threat moved from Alcor's Sunrazor to Iceblade Hacker. Shield Slam no longer procs weapons, and the earlier tie between them was noise.
- Balanced and max survival wear 6 to 8 pieces of Battlegear of Heroism.
- Max threat wears 1 (the breastplate) and goes for hit, strength and agility.

### balanced
| Slot | Pick | Flag | Source |
|---|---|---|---|
| head | Lionheart Helm | doubtful | Blacksmithing 300 · Armorsmith · plans (12717) rare drop: UBRS Smolderthorn Headhunter 0.2%, Strat Plague Ghoul 0.05% · 80 Thorium Bars |
| neck | Horn of Ve'han | doubtful | Quest: Work Well Done |
| shoulder | Pauldrons of Heroism |  | Quest (NPC not in our data yet) · Spaulders of Valor (Rend, UBRS), Legplates of Valor (Rivendare, Strat), Boots of Valor (Kirtonos, Scholo)  |
| back | Howler's Furs |  | Pix Xizzix · 75 Reals |
| chest | Breastplate of Heroism | doubtful | Quest (NPC not in our data yet) · Helm of Valor (Gandling, Scholo) + Breastplate of Valor (Drakkisath, UBRS) · end of the Valthalak chain, s |
| wrist | Bracers of Heroism |  | Quest (NPC not in our data yet) · Bracers of Valor (BoE) + 15 Silithus Venom Samples + 20g · or ~9g at the vendor after · Forever Battlegear |
| hands | Gauntlets of Heroism |  | Quest (NPC not in our data yet) · Gauntlets of Valor (Ramstein, Strat) + Belt of Valor (BoE) · after the 20g bracer quest · or 47g at the ve |
| waist | Waistguard of Heroism |  | Quest (NPC not in our data yet) · Gauntlets of Valor (Ramstein, Strat) + Belt of Valor (BoE) · after the 20g bracer quest · or 23.5g at the  |
| legs | Handcrafted Mastersmith Leggings |  | Seen in game 2026-11-10 · source not in our data yet |
| feet | Battleboots of Heroism |  | Quest (NPC not in our data yet) · Spaulders of Valor (Rend, UBRS), Legplates of Valor (Rivendare, Strat), Boots of Valor (Kirtonos, Scholo)  |
| finger | Tracker's Signet of Vigilance | doubtful | Watchers Honored · no rep source found yet |
| finger | Watcher's Signet of Vigilance | doubtful | Watchers Revered · no rep source found yet |
| trinket | Chained Essence of Eranikus |  | Quest: The Essence of Eranikus, Sunken Temple |
| trinket | Adaptive Combat Assistant |  | Pix Xizzix · 125 Reals |
| mainhand | Iceblade Hacker |  | Ras Frostwhisper (Scholo) |
| offhand | Evergreen Shield | doubtful | Blacksmithing 300 · plans (279311) have no source in Forever data · reagents not in data |
| ranged | Blackcrow |  | Shadow Hunter Vosh'gajin · Blackrock Spire |
enchants: head Lesser Arcanum of Voracity, neck Enchant Necklace - Agility, back Enchant Cloak - Superior Defense, chest Enchant Chest - Greater Stats, wrist Enchant Bracer - Superior Strength, hands Enchant Gloves - Superior Strength, legs Lesser Arcanum of Voracity, feet Enchant Boots - Greater Agility, mainhand Enchant Weapon - Crusader, offhand Enchant Shield - Greater Stamina
### max_threat
| Slot | Pick | Flag | Source |
|---|---|---|---|
| head | Lionheart Helm | doubtful | Blacksmithing 300 · Armorsmith · plans (12717) rare drop: UBRS Smolderthorn Headhunter 0.2%, Strat Plague Ghoul 0.05% · 80 Thorium Bars |
| neck | Beads of Ogre Might |  | Quest: Falrin's Vendetta (Dire Maul) · 25 Ogre Warbeads from Dire Maul or Blackrock Spire ogres · step of Mokvar's Five Thunders chain |
| shoulder | Spaulders of Containment | doubtful | No source in our data · revealed Nov 8 |
| back | Howler's Furs |  | Pix Xizzix · 75 Reals |
| chest | Breastplate of Heroism | doubtful | Quest (NPC not in our data yet) · Helm of Valor (Gandling, Scholo) + Breastplate of Valor (Drakkisath, UBRS) · end of the Valthalak chain, s |
| wrist | Bracers of Heroism |  | Not in Forever's client data (the server calls it invalid) · Classic stats |
| hands | Raider Gauntlets |  | Darkspear Raiders Revered · Revered per the item data, not Exalted · Creeg Bothunk (Barrens) |
| waist | Brigam Girdle |  | General Drakkisath · Upper Blackrock Spire |
| legs | Titanic Leggings | doubtful | Blacksmithing 300 · Armorsmith · plans (22388) have no source in Forever data · 2 Flask of the Titans |
| feet | Furbolg Berserker's Boots | doubtful | Seen in game Nov 21 · drop source unknown (Furbolg rares like Sabella No-Eyes?) · unsure, used by standards 42 |
| finger | Watcher's Signet of Vigilance | doubtful | Watchers Revered · no rep source found yet |
| finger | Tracker's Signet of Vigilance | doubtful | Watchers Honored · no rep source found yet |
| trinket | Chained Essence of Eranikus |  | Quest: The Essence of Eranikus, Sunken Temple |
| trinket | Adaptive Combat Assistant |  | Pix Xizzix · 125 Reals |
| mainhand | Iceblade Hacker |  | Ras Frostwhisper (Scholo) |
| offhand | Evergreen Shield | doubtful | Blacksmithing 300 · plans (279311) have no source in Forever data · reagents not in data |
| ranged | Bloodseeker | doubtful | Alterac Valley quest · AV queue unconfirmed · unsure, used by standards 42 |
enchants: head Lesser Arcanum of Voracity, neck Enchant Necklace - Strength, back Enchant Cloak - Agility, chest Enchant Chest - Greater Stats, wrist Enchant Bracer - Superior Strength, hands Enchant Gloves - Superior Strength, legs Lesser Arcanum of Voracity, feet Enchant Boots - Greater Agility, mainhand Enchant Weapon - Crusader, offhand Enchant Shield - Greater Stamina
### max_survival
| Slot | Pick | Flag | Source |
|---|---|---|---|
| head | Faceguard of Heroism | doubtful | Quest (NPC not in our data yet) · Helm of Valor (Gandling, Scholo) + Breastplate of Valor (Drakkisath, UBRS) · end of the Valthalak chain, s |
| neck | Amulet of the Darkmoon |  | Quest: 1200 Tickets - Amulet of the Darkmoon · Faire weeks Nov 9-16, Dec 7-14 (Mulgore), Jan 4-11 |
| shoulder | Pauldrons of Heroism |  | Quest (NPC not in our data yet) · Spaulders of Valor (Rend, UBRS), Legplates of Valor (Rivendare, Strat), Boots of Valor (Kirtonos, Scholo)  |
| back | Shifting Cloak |  | Leatherworking 300, pattern from Knot Thimblejack's Cache, Dire Maul |
| chest | Breastplate of the Chromatic Flight |  | Quest: Breastplate of the Chromatic Flight (started by the Catalogue of the Wayward, Andorhal) · Chromatic Carapace (Gyth, UBRS, rare), 10 B |
| wrist | Wristguards of Heroism |  | Quest (NPC not in our data yet) · Bracers of Valor (BoE) + 15 Silithus Venom Samples + 20g · or ~9g at the vendor after · Forever Battlegear |
| hands | Handguards of Heroism |  | Quest (NPC not in our data yet) · Gauntlets of Valor (Ramstein, Strat) + Belt of Valor (BoE) · after the 20g bracer quest · or 47g at the ve |
| waist | Valiant's Waistguard |  | Pix Xizzix · 100 Reals · Warrior only |
| legs | Legguards of Heroism |  | Quest (NPC not in our data yet) · Spaulders of Valor (Rend, UBRS), Legplates of Valor (Rivendare, Strat), Boots of Valor (Kirtonos, Scholo)  |
| feet | Sabatons of Heroism |  | Quest (NPC not in our data yet) · Spaulders of Valor (Rend, UBRS), Legplates of Valor (Rivendare, Strat), Boots of Valor (Kirtonos, Scholo)  |
| finger | Watcher's Signet of Vigilance | doubtful | Watchers Revered · no rep source found yet |
| finger | Tracker's Signet of Vigilance | doubtful | Watchers Honored · no rep source found yet |
| trinket | Adaptive Combat Assistant |  | Pix Xizzix · 125 Reals |
| trinket | Royal Seal of Eldre'Thalas |  | Quest: Codex of Defense (Warrior), Dire Maul |
| mainhand | Iceblade Hacker |  | Ras Frostwhisper (Scholo) |
| offhand | Evergreen Shield | doubtful | Blacksmithing 300 · plans (279311) have no source in Forever data · reagents not in data |
| ranged | Eaglehorn Long Bow |  | Random world drop |
enchants: head Lesser Arcanum of Constitution, neck Enchant Necklace - Strength, back Enchant Cloak - Superior Defense, chest Enchant Chest - Greater Stats, wrist Enchant Bracer - Superior Deflection, hands Enchant Gloves - Superior Strength, legs Lesser Arcanum of Constitution, feet Enchant Boots - Greater Agility, mainhand Enchant Weapon - Agility, offhand Enchant Shield - Greater Stamina

The flags:
- Doubtful picks carry their reason in the source column.
- The Watchers rings have no rep source found yet.
- Evergreen Shield has no plans source in Forever data. It is the next shield after Force Reactive Disk went out.
- Lionheart Helm's plans are rare drops (0.2%).
- Titanic Leggings' plans have no source.
- Bloodseeker is an AV quest, and AV is unconfirmed.
- No pick is out or infeasible.

New tags this round:
- open: Enchant Cloak Defense, Gorewood Bow, Barrier Shield, Reiver Claws;
- out: Lord Blackwood's Blade (Scholomance summon, like the Buckler);
- open: Petrified Bark Shield, Avalanchion's Stony Hide, Tunnelweb Pincer, Enchant Weapon Striking (weapon re-rank top 5s)
- doubtful: Officer's Gloves 250551 (a Forever Blacksmithing recipe, plans source not in our data).

## Audit fix round (849024cbb)

tank_gear-1 was that the bench scored the lists 5 to 11% under their list score. The cause: the request APLs still cast the synthetic set's on-use trinkets (990617, 991018). The Phase 0 sets don't wear them, so tankbench never used a trinket. rotopt (setRotation, real_trinket.go) swaps them for the first worn trinket with a use, so the search was right and the requests were wrong.
- The requests now cast that trinket: Chained Essence of Eranikus on Balanced and Max threat, Adaptive Combat Assistant on Max survival.
- At Magmadar, tankbench SafeDPS went from 2254 to 2344. rotopt's is 2342.
- The bench now matches the lists: Balanced 3309.7 against 3299.4, Max threat 4698.5 against 4692.0, Max survival 1203.8 against 1218.7.
- No re-search was needed.

Follow-up (ecf1cbb50): rotopt now maps each synthetic id to its own worn on-use trinket.
- Re-ranking the trinket slots on all three presets and the Alliance row left the picks unchanged.
- The list scores went up: Balanced 3360.5 (+61.2), Max threat 4725.7 (+33.7), Alliance 4584.3 (+19.0). Max survival is unchanged at 1218.7, because Royal Seal of Eldre'Thalas has no use.
- The requests now cast both worn trinkets: 990617 is trinket 1's line and 991018 is trinket 2's.
- One catch: rotopt maps only synthetic ids. A request that casts real ids keeps them whatever trinket the search tries. So a trinket re-rank needs the synthetic ids put back in the request first, which is how this one ran.

Fixed (trinket mapping round): rotopt's withRealTrinket now gives each synthetic id its own worn on-use trinket (first synthetic id to the first trinket slot's, the second to the second's; TestRealTrinketPerSlot). The Balanced and Max threat requests' second trinket line casts Adaptive Combat Assistant. tankbench, 3000 iterations, first seed 1, labeled, six bosses: Balanced 3305.8 to 3380.0, Max threat 4700.3 to 4728.0. rotopt -eval on the same requests: 3364.0 and 4720.7. The list scores (3299.4, 4692.0) predate this and are 2.4% and 0.8% under; a re-rank would refresh them. Max survival is unchanged (Royal Seal of Eldre'Thalas has no modelled use).

tank_gear-2: Frozen Steel Vambraces 19112 is now out (the server calls it invalid). The Max threat wrist was re-ranked:
- Horde: Bracers of Heroism, list 4692.0 (-7.7).
- Alliance: Vambraces of the Sadist, list 4565.3 (-4.7).
- Gordok Bracers of Power 18533 tagged open (Dire Maul Tribute).

## Price check round (fc39e5000)

tools/audit/gear/phase0_check.py flagged three open items ranked above the pick. The slots were re-ranked with a 20000-iteration confirm. Swaps must clear 2 stderr.
- Balanced shoulder, Spaulders of Heroism: now ranks 1.3 below Pauldrons of Heroism. Clean.
- Max survival feet, Sabatons of Heroism: +6.6 (3.6) at 20000 iterations, then +4.3 (1.9) at 80000. Swapped in. The list score is 1218.7.
- Max survival neck, Talisman of Evasion: -18.6 (4.4) at 20000 iterations, so Amulet of the Darkmoon stays. The flag stands, because the check reads the 1000-iteration screen value (+5.2).

The scores and race rows above are on fc39e5000 code. Synthetic and the other presets did not move.

## Set trap

Two max threat searches started from full Battlegear of Heroism: 6 pieces with Heroism waist, and 4 pieces.
- The 6-piece start ended at 4695.7 with Heroism boots and bracers and Valiant's Waistguard.
- The 4-piece start ended at 4682.7.
- The main list scores 4733.2 with 1 Heroism piece and keeps its picks.

Balanced and max survival already pick 6 or more Heroism pieces.

## Talents and rotation (step 7)

No change. The knob search (tools/rotopt warrior_tank, 1000 iterations, 3000 to confirm, potion, wall and trinket knobs frozen) ran from the shipped knobs on each Phase 0 request. In all three, its best equalled its start:
- balanced 3345.4;
- max threat 4717.5;
- max survival 1231.0.

The talents are unchanged.

Thunder Clap check (code 3aeced173, Thunder Clap threat now 1x): Balanced on the Phase 0 request, paired evals at 20000 iterations, six bosses.

| Seed | Shipped (with Thunder Clap) | tc=0 | Delta |
|---|---|---|---|
| 1 | 3301.5 (4.1) | 3266.1 (4.4) | -35.4 |
| 1000003 | 3295.1 (3.9) | 3269.0 (4.3) | -26.1 |

Dropping Thunder Clap loses on both seeds, well beyond noise, so it stays. A likely reason, not measured: the Phase 0 set has less mitigation than the synthetic one, so Thunder Clap's attack speed slow counts for more. The rotation and the Horde Balanced row are unchanged. The scores and race rows above are from code before 3aeced173.

## Faction row (step 8)

Max threat's hands, Raider Gauntlets, need Darkspear Raiders Revered (Horde). The Alliance row is ranked at RaceHuman over the hands slot.
- The best Alliance hands are Gauntlets of Heroism.
- The Theramore Gauntlets would rank higher but are infeasible.
- The Alliance list scores 4565.3 in the search, against Horde's 4692.0. That includes the race change.

It is written to ui/tank_warrior/gear_sets/level60_phase0_warrior_protection_max_threat_alliance.gear.json. Balanced and max survival have no faction reward. The Watchers rings are for both factions.

## Hand check (step 5)

Sources are as in the tables. No public pre-raid list could be read:
- Icy Veins returned 403;
- Wowhead's classic guide page served no gear content;
- the Forever prot guide has no gear list.

So no outside comparison was made.

## Horde race ranking

This uses the tools/dps_report bench_races method:
- only `players[0].race` changes;
- `tankbench -iters 4000 -first-seed 1 -labeled` on the six bosses;
- the score is the preset's column, as the mean.

All four races wear the preset's gear.

| Run | Preset | Orc | Troll | Tauren | Undead |
|---|---|---|---|---|---|
| Synthetic, before the rage change (b4cab8498^) | Balanced | 3417.8 | 3398.5 | 3346.7 | **3480.2** |
| | Max threat | **4250.7** | 4204.2 | 4142.0 | 4247.0 |
| | Max survival | 1575.5 | 1642.7 | 1585.5 | **1840.5** |
| Synthetic, after (current code) | Balanced | 3199.0 | 3182.2 | 3127.7 | **3262.8** |
| | Max threat | **4114.3** | 4066.7 | 4004.7 | 4110.0 |
| | Max survival | 1349.8 | 1418.7 | 1360.7 | **1617.8** |
| Phase 0 set, after (current code) | Balanced | 3270.3 | 3292.8 | 3241.3 | **3379.7** |
| | Max threat | 4706.0 | 4701.0 | 4627.7 | **4727.2** |
| | Max survival | 854.0 | 974.8 | 903.3 | **1203.8** |
| Synthetic, rage stance rule off (final go code) | Balanced | 3217.2 | 3199.0 | 3144.7 | 3279.7 |
| | Max threat | 4124.0 | 4077.7 | 4013.5 | 4120.0 |
| | Max survival | 1371.3 | 1441.7 | 1380.8 | 1635.7 |

The ranking:
- Undead first, on all three presets with the Phase 0 set.
- Orc is level on max threat (within 13, which is noise at this size). The set trap ran with Force Reactive Disk on both sides, so its result holds.
- Then Troll and Orc.
- Tauren last.

The before to after drop also includes the threat and items batches. The stance rule alone, comparing "rage stance rule off" with "after" on the synthetic gear, costs about 10 to 20 per preset, the same for every race. Shield Slam's main-hand change moved the synthetic rows by less than 0.1, since that gear has no weapon procs on the bench.

Undead's lead comes from Touch of the Grave (sim/core/racials_forever.go, registerTouchOfTheGrave): a 20% proc on a 15s cooldown that heals 5% of max health. It is worth about 250 to 360 on max survival and 100 on balanced. The open question: the tooltip says "up to 5%" and no log has measured it. If it heals less, Undead's lead shrinks.

## Open

- Shield Slam's 2x block value may be reversed by the user. Every score here has it.
- Touch of the Grave's real heal size ("up to 5%").
- The doubtful picks above, chiefly the Watchers rep, the Evergreen Shield plans and the Lionheart Helm plans.


## Max survival vs synthetic

`tankbench -iters 1000`, mean of the six scoring bosses. Synthetic is the origin/master request. Stats are tankbench's sheet against a level 63 boss. Block value is (sheet value + 0.05 Str) x the block value multiplier. Coverage is miss + dodge + parry + block (102.4% on the sheet is uncrushable). Both sets clear the 315 Defense that stops Forever's 0.6% boss crit, so crit plays no part. Max survival is 2 x SafeDPS less 3 x 900 / 300 x the healing that holds deaths to 5%, so healing needed is the survival side.

| Stat | Phase 0 | Synthetic |
|---|---|---|
| Health | 10269 | 10229 |
| Stamina | 656 | 652 |
| Armor | 8824 | 9300 |
| Defense | 382 | 467 |
| Miss / dodge / parry / block | 9.7 / 22.3 / 12.7 / 12.7 | 13.1 / 20.5 / 16.1 / 16.1 |
| Block value | 115 | 154 |
| Coverage | 57.3% | 65.7% |
| Avoided / blocked / crushed | 44.3 / 45.9 / 3.4% | 49.5 / 43.5 / 3.0% |
| DTPS | 363 | 321 |
| Healing needed | 266 | 213 |
| Max survival | 1218.8 | 1617.0 |

No non-gear field differs except the APL's trinket item ids, which follow the worn trinkets. Shield Block keeps both nearly uncrushable. The drop is Defense (-85, 5.2 points less avoidance), armor (-476) and block value (-39), with health flat.
