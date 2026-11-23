# Phase 0 step 3: druid and priest (2026-11-23)

Balance, Feral, Shadow and Smite, from `tools/rotopt/bis_phase0.sh` on master. Lists in
`ui/balance_druid/bis/`, `ui/feral_druid/bis/` and `ui/shadow_priest/bis/` (`level60_phase0_<spec>`).

## Summary

| Spec | Phase 0 | Synthetic | Gap |
|---|---|---|---|
| Balance | 821.4 | 956.1 | -134.7 (-14.1%) |
| Feral | 1066.5 | 1186.1 | -119.6 (-10.1%) |
| Shadow | 847.2 | 948.4 | -101.2 (-10.7%) |
| Smite | 686.2 | 781.1 | -94.9 (-12.1%) |

Calibration: each spec's report build at 180s (the request's 20s variation kept), 10000 iterations,
seed 1, on the list's picks with their enchants and on the report's synthetic gear. The synthetic
gear is 10 to 14% above what Phase 0 offers for all four, so it reads as a Phase 1 set, not a pre-raid one.

What changed in the shared files:

- The dry run pools matched `phase0_pool/` but for fewer items (more now tagged), with no new
  unmodelled or hidden stat items.
- Bracers of the Fallen Son (21594) "out": item level 88 BoP leather with no source, Classic
  Ahn'Qiraj loot. It was Balance's first wrist pick, so Balance ran again.
- Fel Cape (279269) and Mana-infused Cord (279267) "doubtful": Forever's Tailoring 300 crafts with
  no known pattern source. Their siblings (Robes of Fiery Devastation, Rime-encrusted Boots) need
  the raid essence. Fel Cape was 1 DPS under Shadow's back pick.
- Feralheart Raiment (set 1667, 31 rows) and Vestments of the Virtuous (set 1675, 14 rows, no
  chest) "include", by the parent's ruling, with head and chest "doubtful" like Five Thunders.
  All four specs ran again after it. The set code covers the IDs (`setId` on every row).

A false alarm, for the record: a hand test said Virtuous in 6 slots beat Shadow's picks by 41 DPS.
It had kept the old items' random suffixes on three swapped slots. Without them the 6 pieces sim
841.6 against the list's 847.2, and a full rerun from that start ended at 841.6 too. The list
stands. Anyone who swaps items in a request by hand should drop `randomSuffix` on those slots.

## Balance

Feralheart 6 pieces (head, shoulder, chest, wrist, hands, feet).

| Slot | Pick | Source |
|---|---|---|
| head | Feralheart Cowl (doubtful) | Set quest: end of the Valthalak chain, step gold unknown, or the vendor after |
| neck | Orb of the Darkmoon | Quest: 1200 Tickets - Orb of the Darkmoon, Faire weeks Nov 9-16, Dec 7-14 (Mulgore), Jan 4-11 |
| shoulder | Feralheart Spaulders | Set quest: Anthion chain (Strat, Dire Maul), step gold unknown, or the vendor after |
| back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| chest | Feralheart Vest (doubtful) | Set quest: end of the Valthalak chain, step gold unknown, or the vendor after |
| wrist | Feralheart Wraps | Set quest: bracer quest, or the vendor after |
| hands | Feralheart Hands | Set quest: gloves and belt quest, after the bracer quest, or the vendor after |
| waist | Belt of the Archmage | Tailoring 295, pattern on AH ~131g, Loom on AH ~10g |
| legs | Ironfeather Leggings | On AH ~121g, or Leatherworking 300, pattern on AH ~16g |
| feet | Feralheart Galoshes | Set quest: Anthion chain (Strat, Dire Maul), step gold unknown, or the vendor after |
| finger | Watcher's Signet of Brilliance (doubtful) | Watchers Revered, no rep source found yet |
| finger | Channeler's Ring | Pix Xizzix, 75 Reals, owned |
| trinket | Chained Essence of Eranikus | Quest: The Essence of Eranikus, Sunken Temple |
| trinket | Eye of the Beast | Quest reward, Blackrock Spire (Wowhead Forever source), quest not named in our data |
| mainhand | Hand of Edward the Odd | Random world drop |
| offhand | Obsidian Bauble | Pusillin, Dire Maul |
| ranged | Idol of the Moon (doubtful) | No known source, may be late content |

Flags:
- Night Elf build: the report's Balance and Feral builds are Night Elf, but Eye of the Beast (and
  Feral's Blackhand's Breadth) are Horde quest rewards in Classic ("For The Horde!", Rend). The item
  data has no faction flag on them, so the search keeps them. User call.
- Doubtful head and chest (Feralheart, end of the Valthalak chain): without both, the best is 4
  pieces with Spellweaver's Turban and Alanna's Embrace, 806.7 DPS (-14.7).
- Watcher's Signet of Brilliance (doubtful, no Watchers rep source): Maiden's Circle is 3.0 DPS under.
- Idol of the Moon (doubtful, no known source) is the only idol the list ranks. Without it the slot is empty.
- No pick has an unmodelled effect or hidden stats.

Public list (wowtbc.gg Balance pre-raid): it shares Eye of the Beast and the Belt of the
Archmage slot idea (Ban'thok Sash there, 2.7 DPS under ours). The big differences are the Forever
items (Feralheart set, Arcanoweave Cloak, Channeler's Ring, Orb of the Darkmoon) and Hand of Edward
the Odd with Obsidian Bauble over Witchblade and Spirit of Aquementas. Those guides predate Forever's
items, so that's expected. Robes of the Royal Crown, Hands of Power and Skyshroud Leggings lose to
the set pieces and Ironfeather Leggings.

## Feral

No set: only the Feralheart Cap (1 piece).

| Slot | Pick | Source |
|---|---|---|
| head | Feralheart Cap (doubtful) | Set quest: end of the Valthalak chain, step gold unknown, or the vendor after |
| neck | Amulet of the Darkmoon | Quest: 1200 Tickets - Amulet of the Darkmoon, Faire weeks Nov 9-16, Dec 7-14 (Mulgore), Jan 4-11 |
| shoulder | Darkspear Pauldrons |  |
| back | Cape of the Black Baron | Drop: Baron Rivendare, Stratholme |
| chest | Timbermaw Tunic | On AH ~251g, or Leatherworking 300, pattern on AH ~256g |
| wrist | Warpwood Bracers | Quest: Lethtendris's Web (Dire Maul) |
| hands | Theramore Gloves |  |
| waist | Shifter's Belt | Pix Xizzix, Reals (price not in data), Druid only |
| legs | Warbear Woolies |  |
| feet | Drudge Boots (doubtful) | Quest: The Nightmare Manifests |
| finger | Cutthroat's Signet | Pix Xizzix, 75 Reals |
| finger | Opal Ring of the Tiger | Random world drop |
| trinket | Chained Essence of Eranikus | Quest: The Essence of Eranikus, Sunken Temple |
| trinket | Blackhand's Breadth | Quest reward, Blackrock Spire (Wowhead Forever source), quest not named in our data |
| mainhand | Crystal Spiked Maul (doubtful) | No known source |
| ranged | Idol of Ferocity | Lord Roccor, Blackrock Depths |

Flags:
- Crystal Spiked Maul: doubtful (no known source) and its stats are hidden in Forever. Bonecrusher
  is 1.8 DPS under. The Unstoppable Force ranks 7.7 above, infeasible (Frostwolf Exalted).
- Feralheart Cap (doubtful): Outlaw's Collar is 5.2 under. Drudge Boots (doubtful, Silithus):
  Supreme Shoes of the Tiger 1.7 under.
- Blackhand's Breadth on a Night Elf build, as above.
- Darkspear Pauldrons, Theramore Gloves and Warbear Woolies were left open by the Enhancement check
  without a note. Wowhead's Forever data sells the first two in The Barrens. Not confirmed in game.

Public list (wowtbc.gg Feral pre-raid): Cape of the Black Baron matches. Manual Crowd Pummeler is
its weapon. Ours ranks it 23rd, 24.5 under: its use is modelled (+46.7 DPS at 90s) but its item
level 34 stats lose that and more. Mark of Fordring (5.8 under), Swiftwalker Boots (2.7),
Cloudrunner Girdle (8.5), Hand of Justice (6.7) and the Devilsaur pieces (12 to 13) all rank under
the Forever items and Pix Xizzix gear. No mistake found.

## Shadow

Two Virtuous pieces (head, shoulder). More pieces lose (see the false alarm above).

| Slot | Pick | Source |
|---|---|---|
| head | Virtuous Cowl (doubtful) | Set quest: end of the Valthalak chain, step gold unknown, or the vendor after |
| neck | Dark Advisor's Pendant | Drop: Vectus, Scholomance |
| shoulder | Virtuous Epaulets | Set quest: Anthion chain (Strat, Dire Maul), step gold unknown, or the vendor after |
| back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| chest | Robes of Dark Summons (doubtful) | Impaan the Imprisoner, Blackmaw Hold, not in guides yet |
| wrist | Eternal Bindings of Shadow Wrath | Random world drop, BoE |
| hands | Felcloth Gloves | Tailoring 280, pattern from Knot Thimblejack's Cache (Dire Maul) 16%, BoP, 12 Runecloth bolts, 20 Felcloth, 8  |
| waist | Belt of the Archmage | Tailoring 295, pattern on AH ~131g, Loom on AH ~10g |
| legs | Master's Leggings of Shadow Wrath | Random world drop, BoE |
| feet | Eternal Boots of Shadow Wrath | Random world drop, BoE |
| finger | Watcher's Signet of Brilliance (doubtful) | Watchers Revered, no rep source found yet |
| finger | Channeler's Ring | Pix Xizzix, 75 Reals, owned |
| trinket | Chained Essence of Eranikus | Quest: The Essence of Eranikus, Sunken Temple |
| trinket | Eye of the Beast | Quest reward, Blackrock Spire (Wowhead Forever source), quest not named in our data |
| mainhand | Hand of Edward the Odd | Random world drop |
| offhand | Obsidian Bauble | Pusillin, Dire Maul |
| ranged | Ritssyn's Wand of Bad Mojo | Baron Rivendare, Stratholme |

Flags:
- Doubtful head (Virtuous Cowl) and chest (Robes of Dark Summons): with Spellweaver's Turban and Robe
  of Winter Night instead it sims 847.9, the same within noise. So the doubtful picks cost nothing.
- Watcher's Signet (doubtful): Eye of Orgrimmar is level with it.
- No unmodelled effects or hidden stats among the picks. Robe of Winter Night, Felcloth Gloves and
  Belt of the Archmage need the build's Tailoring.

Public list (wowtbc.gg Shadow pre-raid): it shares Dark Advisor's Pendant, Belt of the Archmage's slot,
Eye of the Beast and the Robe of Winter Night (our open fallback). Ritssyn's Wand over Skul's
Ghastly Touch, and Hand of Edward the Odd with Obsidian Bauble over Scepter of the Unholy with
Spirit of Aquementas. Shadow Wrath world drops over Maleki's Footwraps and Skyshroud Leggings.

## Smite

Virtuous 6 pieces (all but chest and waist).

| Slot | Pick | Source |
|---|---|---|
| head | Virtuous Cowl (doubtful) | Set quest: end of the Valthalak chain, step gold unknown, or the vendor after |
| neck | Orb of the Darkmoon | Quest: 1200 Tickets - Orb of the Darkmoon, Faire weeks Nov 9-16, Dec 7-14 (Mulgore), Jan 4-11 |
| shoulder | Virtuous Epaulets | Set quest: Anthion chain (Strat, Dire Maul), step gold unknown, or the vendor after |
| back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| chest | Eternal Chestguard of Holy Wrath | Random world drop, BoE |
| wrist | Virtuous Wraps | Set quest: bracer quest, or the vendor after |
| hands | Virtuous Hands | Set quest: gloves and belt quest, after the bracer quest, or the vendor after |
| waist | Belt of the Archmage | Tailoring 295, pattern on AH ~131g, Loom on AH ~10g |
| legs | Virtuous Leggings | Set quest: Anthion chain (Strat, Dire Maul), step gold unknown, or the vendor after |
| feet | Virtuous Slippers | Set quest: Anthion chain (Strat, Dire Maul), step gold unknown, or the vendor after |
| finger | Maiden's Circle | On AH ~80g |
| finger | Watcher's Signet of Brilliance (doubtful) | Watchers Revered, no rep source found yet |
| trinket | Chained Essence of Eranikus | Quest: The Essence of Eranikus, Sunken Temple |
| trinket | Weakness Analyzer | Pix Xizzix, 125 Reals, unique with ACA |
| mainhand | Hand of Edward the Odd | Random world drop |
| offhand | Tome of the Lost | General Drakkisath, Blackrock Spire |
| ranged | Ritssyn's Wand of Bad Mojo | Baron Rivendare, Stratholme |

Flags:
- Virtuous Cowl (doubtful): without it the best is Spellweaver's Turban with Virtuous Cord to keep 6
  pieces, 677.5 DPS (-8.7).
- Watcher's Signet (doubtful): Eye of Orgrimmar is level with it.
- Maiden's Circle and Weakness Analyzer are priced (80g, 125 Reals). Nothing unmodelled or hidden.

Public list: no Smite pre-raid list found. A forum post suggests the mage list. Ours shares the
caster items with Shadow, with Holy Wrath suffixes and Tome of the Lost in the off hand.

## User calls

- Night Elf druid builds wearing Horde quest rewards (Eye of the Beast, Blackhand's Breadth).
- The Feralheart and Virtuous head and chest (end of the Valthalak chain): Balance -14.7, Smite -8.7
  without them, Shadow and Feral about 0 and -5.
- Watcher's Signet of Brilliance (Watchers Revered, no rep source) in three lists, worth 0 to 3 DPS.
- Crystal Spiked Maul (no source, hidden stats) and Idol of the Moon (no source).
