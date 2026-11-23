# Phase 0 step 3: druid and priest (rerun of 2026-11-23)

Balance, Feral, Shadow and Smite, from `tools/rotopt/bis_phase0.sh` on master 5e0a327fe (item procs,
weapon damage, resistances, class limits), then b1bbe4347 (Exalted rewards infeasible, AV and AB items
in). Lists in `ui/balance_druid/bis/`, `ui/feral_druid/bis/` and `ui/shadow_priest/bis/`
(`level60_phase0_<spec>`).

## Summary

| Spec | Phase 0 | Synthetic | Gap | Earlier Phase 0 |
|---|---|---|---|---|
| Balance | 822.7 | 956.1 | -133.4 (-14.0%) | 821.4 |
| Feral | 1068.5 | 1186.1 | -117.6 (-9.9%) | 1066.3 |
| Shadow | 855.3 | 948.4 | -93.1 (-9.8%) | 847.2 |
| Smite | 687.8 | 781.1 | -93.3 (-11.9%) | 686.2 |

Calibration: each spec's report build at 180s, 10000 iterations, seed 1, `rotopt -eval` on the list's
picks with their enchants and on the report's synthetic gear. "Earlier Phase 0" is the earlier list
simmed the same way on today's data. The synthetic gear stays 10 to 14% above Phase 0, a Phase 1 set.

What changed:

- Balance: Rune Band of Wizardry for Channeler's Ring, within noise (Channeler's ranks 1.4 above).
- Feral: Theramore Pauldrons, Bracers of Subterfuge, Timbermaw Brawlers and Protector's Band in, from
  the newly modelled data and the faction check. Don Julio's Band was picked, then went infeasible
  (b1bbe4347), so the finger slots ran again and Cutthroat's Signet came back.
- Shadow: the Virtuous pieces are out. Spellweaver's Turban, Rugged Mantle of the Timbermaw, Robe of
  Winter Night, Knowledge of the Timbermaw and other Shadow Wrath rolls in. Mindfang and Don Julio's
  Band were picked, then went infeasible, so finger, main hand and off hand ran again: Hand of Edward
  the Odd and Rune Band of Wizardry.
- Smite: Master's Vest of Holy Wrath, Knowledge of the Timbermaw, Jereman's Bauble and Spirit of
  Aquementas for the Eternal Chestguard, Belt of the Archmage, Watcher's Signet and Tome of the Lost.

Process: the dry runs matched `phase0_pool/` but for three effects now modelled (Force of Magma,
Shimmering Platinum Warhammer, Taran Icebreaker) and, after b1bbe4347, six fewer unlisted items in the
priest pools. No new unmodelled or hidden stat items. The faction check leaves 134 Horde items out of the
druid pools and 84 Alliance items out of the priest pools. The runs took 36 min (Balance) to 2h (Feral)
on 4 cores. This container has no `jj`, so a git shim stood in for the script's sim hash.

## Balance

Feralheart 6 pieces (head, shoulder, chest, wrist, hands, feet). List score 839.7.

| Slot | Pick | Source |
|---|---|---|
| head | Feralheart Cowl (doubtful) | Set quest: end of the Valthalak chain, or the vendor after |
| neck | Orb of the Darkmoon | Quest: 1200 Tickets, Faire weeks Nov 9-16, Dec 7-14 (Mulgore), Jan 4-11 |
| shoulder | Feralheart Spaulders | Set quest: Anthion chain (Strat, Dire Maul), or the vendor after |
| back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| chest | Feralheart Vest (doubtful) | Set quest: end of the Valthalak chain, or the vendor after |
| wrist | Feralheart Wraps | Set quest: bracer quest, or the vendor after |
| hands | Feralheart Hands | Set quest: gloves and belt quest, or the vendor after |
| waist | Belt of the Archmage | Tailoring 295, pattern on AH ~131g |
| legs | Ironfeather Leggings | On AH ~121g, or Leatherworking 300 |
| feet | Feralheart Galoshes | Set quest: Anthion chain, or the vendor after |
| finger | Watcher's Signet of Brilliance (doubtful) | Watchers Revered, no rep source found yet |
| finger | Rune Band of Wizardry | Lord Valthalak, summon unconfirmed in Forever (include, standards 42) |
| trinket | Chained Essence of Eranikus | Quest: The Essence of Eranikus, Sunken Temple |
| trinket | Eye of the Beast | Quest reward, Blackrock Spire, no side in our data |
| mainhand | Hand of Edward the Odd | Random world drop |
| offhand | Obsidian Bauble | Pusillin, Dire Maul |
| ranged | Idol of the Moon (doubtful) | No known source |

Flags:
- Doubtful head and chest (Feralheart, end of the Valthalak chain): Spellweaver's Turban is 11.1 under,
  Robes of Dark Summons 13.4 under (doubtful too).
- Rune Band of Wizardry and Channeler's Ring are within noise. Idol of the Moon is the only idol ranked.
- No pick has an unmodelled effect or hidden stats.

Public list (wowtbc.gg Balance pre-raid): as before. It shares Eye of the Beast and Belt of the Archmage's
slot (Ban'thok Sash there, 3.1 under). The differences are Forever items, which those guides predate.

## Feral

Feralheart Cap only, no set. List score 1082.4.

| Slot | Pick | Source |
|---|---|---|
| head | Feralheart Cap (doubtful) | Set quest: end of the Valthalak chain, or the vendor after |
| neck | Amulet of the Darkmoon | Quest: 1200 Tickets, Faire weeks |
| shoulder | Theramore Pauldrons | Theramore rep, Alliance only |
| back | Cape of the Black Baron | Baron Rivendare, Stratholme |
| chest | Timbermaw Tunic | On AH ~251g, or Leatherworking 300 |
| wrist | Bracers of Subterfuge (doubtful) | Naxx-era Insignia quest, Eastern Plaguelands per Wowhead's Forever data |
| hands | Timbermaw Brawlers | Leatherworking 300, pattern from Meilosh (Timbermaw Revered) |
| waist | Shifter's Belt | Pix Xizzix, Reals, Druid only |
| legs | Warbear Woolies | No source in our data, left open by the Enhancement check |
| feet | Drudge Boots (doubtful) | Quest: The Nightmare Manifests |
| finger | Protector's Band | Silverwing Sentinels Honored, Alliance only |
| finger | Cutthroat's Signet | Pix Xizzix, 75 Reals |
| trinket | Chained Essence of Eranikus | Quest, Sunken Temple |
| trinket | Blackhand's Breadth | Quest reward, Blackrock Spire, no side in our data |
| mainhand | Crystal Spiked Maul (doubtful) | No known source, stats hidden in Forever |
| ranged | Idol of Ferocity | Lord Roccor, Blackrock Depths |

Flags:
- Bracers of Subterfuge (doubtful): Forest Stalker's Bracers rank 4.0 above it at the final iterations,
  so the wrist is a toss up either way.
- Crystal Spiked Maul (doubtful, hidden stats): as before. The Unstoppable Force ranks 7.6 above,
  infeasible. Feralheart Cap (doubtful): Outlaw's Collar 4.4 under. Drudge Boots: Dunestalker's 1.4 under.
- Don Julio's Band ranks 1.6 above the rings, infeasible (Frostwolf Exalted, Horde).

Public list (wowtbc.gg Feral pre-raid): as before. Cape of the Black Baron matches, Manual Crowd
Pummeler ranks far under on its item level 34 stats. No mistake found.

## Shadow

No set: Virtuous Cowl ranks 0.2 above Spellweaver's Turban, within noise. List score 875.2.

| Slot | Pick | Source |
|---|---|---|
| head | Spellweaver's Turban | General Drakkisath, Upper Blackrock Spire |
| neck | Dark Advisor's Pendant | Vectus, Scholomance |
| shoulder | Rugged Mantle of the Timbermaw (doubtful) | Meilosh, Tailoring, Loom blueprint unseen |
| back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| chest | Robe of Winter Night | Tailoring 245, pattern from Cobalt Mageweaver 0.8% |
| wrist | Flameweave Cuffs of Shadow Wrath | Lord Incendius, Blackrock Depths |
| hands | Felcloth Gloves | Tailoring 280, pattern from Knot Thimblejack's Cache |
| waist | Knowledge of the Timbermaw | Meilosh, Tailoring, Mooncloth and 20 Winterfall Spirit Beads |
| legs | Eternal Sarong of Shadow Wrath | Prince Tortheldrin, Dire Maul West, or world drop |
| feet | Master's Boots of Shadow Wrath | Random world drop, BoE |
| finger | Watcher's Signet of Brilliance (doubtful) | Watchers Revered, no rep source found yet |
| finger | Rune Band of Wizardry | Lord Valthalak, summon unconfirmed (include) |
| trinket | Chained Essence of Eranikus | Quest, Sunken Temple |
| trinket | Eye of the Beast | Quest reward, Blackrock Spire |
| mainhand | Hand of Edward the Odd | Random world drop |
| offhand | Obsidian Bauble | Pusillin, Dire Maul |
| ranged | Ritssyn's Wand of Bad Mojo | Baron Rivendare, Stratholme |

Flags:
- Rugged Mantle of the Timbermaw (doubtful): Virtuous Epaulets 4.2 under.
- Mindfang ranks 4.5 above the main hand, Tome of Shadow Force 3.9 above the off hand, both infeasible.
- Tailoring picks: Robe of Winter Night, Felcloth Gloves and both Timbermaw pieces need the build's Tailoring.
- No unmodelled effects or hidden stats among the picks.

Public list (wowtbc.gg Shadow pre-raid): now closer, sharing Dark Advisor's Pendant, Robe of Winter
Night, Eye of the Beast and the Shadow Wrath rolls. Ritssyn's Wand over Skul's Ghastly Touch, Hand of
Edward the Odd with Obsidian Bauble over Scepter of the Unholy with Spirit of Aquementas.

## Smite

Virtuous 6 pieces (all but chest and waist). List score 697.6.

| Slot | Pick | Source |
|---|---|---|
| head | Virtuous Cowl (doubtful) | Set quest: end of the Valthalak chain, or the vendor after |
| neck | Orb of the Darkmoon | Quest: 1200 Tickets, Faire weeks |
| shoulder | Virtuous Epaulets | Set quest: Anthion chain, or the vendor after |
| back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| chest | Master's Vest of Holy Wrath | Random world drop, BoE, also General Drakkisath |
| wrist | Virtuous Wraps | Set quest: bracer quest, or the vendor after |
| hands | Virtuous Hands | Set quest: gloves and belt quest, or the vendor after |
| waist | Knowledge of the Timbermaw | Meilosh, Tailoring |
| legs | Virtuous Leggings | Set quest: Anthion chain, or the vendor after |
| feet | Virtuous Slippers | Set quest: Anthion chain, or the vendor after |
| finger | Maiden's Circle | On AH ~80g |
| finger | Jereman's Bauble (doubtful) | Quest: Return to Jereman |
| trinket | Chained Essence of Eranikus | Quest, Sunken Temple |
| trinket | Weakness Analyzer | Pix Xizzix, 125 Reals |
| mainhand | Hand of Edward the Odd | Random world drop |
| offhand | Spirit of Aquementas | Quest: It's Dangerous to Go Alone |
| ranged | Ritssyn's Wand of Bad Mojo | Baron Rivendare, Stratholme |

Flags:
- Virtuous Cowl (doubtful): Virtuous Crown keeps the set, 7.8 under.
- Jereman's Bauble (doubtful): Eye of Orgrimmar and Band of the Damned are level with it, so it costs nothing.
- Mindfang 1.6 under the main hand, infeasible now anyway. Nothing unmodelled or hidden.

Public list: no Smite pre-raid list found. It shares the caster items with Shadow.

## Open questions (unchanged)

- Night Elf druids wear Eye of the Beast and Blackhand's Breadth, Classic Horde quest rewards with no
  side in our data, kept for both factions by `phase0_content.json`.
- Feralheart and Virtuous head and chest at the end of the Valthalak chain, and Idol of the Moon and
  Crystal Spiked Maul with no known source.
