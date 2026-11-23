# Phase 0 gear, step 3: warlock (2026-11-23)

Affliction, Demonology and Destruction, from `briefs/phase0_step3_1122.md`. Lists in
`ui/warlock/bis/level60_phase0_<spec>.bis.json`, effects files next to them.

## Runs

- Dry runs matched `phase0_pool/` but for 24 items since tagged out. No new unmodelled or unlisted item.
- The first two searches lost their lists because `ui/warlock/bis/` didn't exist yet. Fixed by
  creating it.
- The parent's ruling on the Forever class sets came mid run: Deathmist Raiment (set 1681,
  226905-226912) is now in `phase0_content.json` as "include", head and chest "doubtful", notes after
  the Five Thunders ones. All three searches ran again with it in the pool (1278 items).
- No search stopped on unlisted picks, and no pick's tag changed on master since.
- The search's own scores (930.65, 876.97, 879.41) run its settings. The sims below run the report
  build (`tools/dps_report/req/warlock__<spec>.json`) at 180s, 10000 iterations, so they read a
  little lower (920.1, 858.3, 865.8).

## Picks

Shared by all three:

| Slot | Item | Source |
|---|---|---|
| Head | Spellweaver's Turban, Arcanum of Focus | Drakkisath, UBRS |
| Neck | Orb of the Darkmoon, Spell Power | Darkmoon Faire, 1200 tickets |
| Shoulder | Mantle of the Timbermaw | Tailoring, Timbermaw Revered (doubtful) |
| Back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| Chest | Robe of the Void, armor kit | Tailoring 295, pattern from Gandling |
| Waist | Belt of the Archmage | Tailoring 295, pattern on AH |
| Finger | Watcher's Signet of Brilliance | Watchers Revered (doubtful) |
| Off hand | Obsidian Bauble, Wisdom | Pusillin, Dire Maul |
| Wand | Bonecreeper Stylus | Gandling, Scholomance |

By spec:

| Slot | Affliction | Demonology | Destruction |
|---|---|---|---|
| Wrist | Eternal Bindings of Shadow Wrath (BoE world drop) | Flameweave Cuffs of Shadow Wrath (Incendius, BRD) | High Councillor's Bracers of Magic (BoE world drop) |
| Hands | Felcloth Gloves (Tailoring 280) | Deathmist Wraps (set quest) | Deathmist Wraps (set quest) |
| Legs | Eternal Sarong of Shadow Wrath (Tortheldrin, DM West, or BoE) | Eternal Sarong of Shadow Wrath | Skyshroud Leggings (Omokk, LBRS) |
| Feet | Eternal Boots of Shadow Wrath (BoE world drop) | High Councillor's Boots of Shadow Wrath (BoE world drop) | Master's Boots of Magic (BoE world drop) |
| Finger 2 | Eye of Orgrimmar (BRD quest, Horde) | Channeler's Ring (Pix, 75 Reals) | Channeler's Ring |
| Trinkets | Weakness Analyzer (Pix, 125 Reals), Chained Essence of Eranikus (Sunken Temple quest) | Eye of the Beast (UBRS quest), Chained Essence of Eranikus | Eye of the Beast, Weakness Analyzer |
| Main hand | Blade of the New Moon (Immol'thar, DM), Mighty Intellect | Blade of the New Moon, Revelation | Witchblade (Gandling, Scholo), Revelation |

## Flags

- Doubtful: Mantle of the Timbermaw (Loom blueprint not seen dropping, not on our AH) and Watcher's
  Signet of Brilliance (no Watchers rep source found). Both in all three lists.
- Pix Xizzix Reals: Arcanoweave Cloak, Channeler's Ring and Weakness Analyzer. We know no Reals
  source rate. Affliction needs 200 Reals, Demonology 150, Destruction 275.
- Orb of the Darkmoon: 1200 tickets, Faire weeks only (Nov 9-16, Dec 7-14).
- Robe of the Void has an unmodelled use effect. No pick has hidden stats.
- Eye of the Beast: the quest isn't named in our data.
- Weakness Analyzer is "unique with ACA": that is Adaptive Combat Assistant, a melee trinket, so no
  clash with Arcanoweave Cloak.
- Enchants on Merchant's Favor rep (neck, bracer, armor kits) and Off-Hand Wisdom (Enchanting 300).

## Deathmist set check

The search is slot by slot, so the set was simmed at 2 to 8 pieces in its best slots (report build,
180s, 10000 iterations). It never wins. The 6 piece +23 spell power doesn't pay for the stats it
gives up.

| Pieces | Affliction | Demonology | Destruction |
|---|---|---|---|
| Picks | 920.1 (0 pieces) | 858.3 (1, Wraps) | 865.8 (1, Wraps) |
| 2 | 907.7 | 856.0 | 863.6 |
| 3 | 901.2 | 851.4 | 858.0 |
| 4 | 890.4 | 847.2 | 853.0 |
| 6 | 890.0 | 848.7 | 855.9 |
| 8 | 854.5 | 819.2 | 824.3 |

No spec gets a start set.

## Public list comparison

wowtbc.gg's WoW Forever pre-raid lists (no epics, no BoE world drops). Icy Veins refused the fetch.

- Ours rightly leaves out Sageblade, Leggings of Torment, Songstone of Ironforge and Rune Band of
  Wizardry (tagged out: world boss plans, patch 1.10, Alliance only).
- Tome of Shadow Force and Therazane's Touch (Frostwolf Exalted) rank first in our off hand but are
  infeasible while the AV queue is unconfirmed: 2.6 to 5.4 DPS each.
- Omnicast Boots: 2.5 to 2.7 DPS behind our BoE Shadow Wrath or Magic boots. The public list leaves out
  BoE world drops, hence the gap.
- Affliction: Dark Advisor's Pendant is 0.8 behind the Orb, an easy swap if the Faire is missed.
  Maiden's Circle is 1.2 behind, a sure alternative to the doubtful Signet. Skul's Ghastly Touch
  ranks within noise of Bonecreeper Stylus (0.35).
- Destruction and Demonology: Nacreous Shell Necklace ranks first (0.1 and 2.1 over the Orb) but
  costs about 280g, over the gold line. Royal Seal of Eldre'Thalas is 3.3 and 6.3 behind.
  Sublime Wristguards are well down.
- PvP rep items (Dryad's Wrist Bindings, Mindfang, Warden's Signet) would beat the picks by 1 to 14
  DPS but are infeasible.

## Calibration

Report build, 180s, 10000 iterations:

| Spec | Phase 0 | Synthetic | Gap |
|---|---|---|---|
| Affliction | 920.1 | 1089.1 | -169.0 (-15.5%) |
| Demonology | 858.3 | 1056.0 | -197.7 (-18.7%) |
| Destruction | 865.8 | 1069.3 | -203.5 (-19.0%) |

Gear totals, items and enchants only:

| | Phase 0 | Synthetic |
|---|---|---|
| Spell power, shadow | 499 (aff), 452 (demo), 436 (destro) | 561 (aff, destro), 545 (demo) |
| Hit | 2% (aff), 4% | 10% |
| Crit | 3.5 to 4.5% | 7.8 to 10.2% |
| Int | 93 to 113 | 144 to 156 |

The synthetic gear is set well above Phase 0: 10% hit against 2 to 4%, about 60 to 125 more spell
power and twice the crit. So the warlock report's synthetic gear reads as a raid level set, not pre-raid.
