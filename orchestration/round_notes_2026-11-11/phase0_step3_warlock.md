# Phase 0 gear, step 3: warlock (2026-11-23)

Affliction, Demonology and Destruction, from `briefs/phase0_step3_1122.md`. Lists in
`ui/warlock/bis/level60_phase0_<spec>.bis.json`, effects files next to them.

## Rerun (2026-11-23, `briefs/phase0_rerun_1123.md`)

On master 5e0a327 (item procs, weapon damage, resistances, class limits on top of 9457811ae).

- Dry runs: 1292 items a spec. The pools match `phase0_pool/` but for Sul'thraze and Demonshear, whose procs
  are now modelled. No new unmodelled or unlisted item.
- No search stopped on unlisted picks.
- Search scores: 966.23, 904.82, 907.58 (were 930.65, 876.97, 879.41). The sims below run the report build
  at 180s and 10000 iterations, so they read a little lower.

What changed and why:

- All three: Rugged Mantle of the Timbermaw and Knowledge of the Timbermaw (the Meilosh Tailoring upgrades,
  new in the pool) take shoulder and waist. Mindfang takes the main hand now that standards 42 includes it.
- Affliction (+35.7): Talisman of Ascendance (standards 42) for Chained Essence, Tearfall Bracers and
  Master's Boots for the BoE bracers and boots, Skul's Ghastly Touch for Bonecreeper Stylus (was within noise).
- Demonology (+29.7): Talisman of Ascendance for Eye of the Beast, Elegant Bracers and Master's Boots of
  Shadow Wrath for the Flameweave Cuffs and High Councillor's Boots.
- Destruction (+28.2): Rune Band of Wizardry (now included) for Channeler's Ring, Sandworm Skin Gloves for
  Deathmist Wraps.

## Picks

Shared by all three:

| Slot | Item | Source |
|---|---|---|
| Head | Spellweaver's Turban, Arcanum of Focus | Drakkisath, UBRS |
| Neck | Orb of the Darkmoon, Spell Power | Darkmoon Faire, 1200 tickets |
| Shoulder | Rugged Mantle of the Timbermaw | Meilosh, Tailoring, from Mantle of the Timbermaw (doubtful) |
| Back | Arcanoweave Cloak | Pix Xizzix, 75 Reals |
| Chest | Robe of the Void, armor kit | Tailoring 295, pattern from Gandling |
| Waist | Knowledge of the Timbermaw | Meilosh, Tailoring, from Wisdom of the Timbermaw |
| Finger | Watcher's Signet of Brilliance | Watchers Revered (doubtful) |
| Main hand | Mindfang | Defilers Exalted (doubtful) |
| Off hand | Obsidian Bauble, Wisdom | Pusillin, Dire Maul |

By spec:

| Slot | Affliction | Demonology | Destruction |
|---|---|---|---|
| Wrist | Tearfall Bracers of Shadow Wrath (The Unforgiven, Strat) | Elegant Bracers of Shadow Wrath (BoE world drop) | High Councillor's Bracers of Magic (BoE world drop) |
| Hands | Felcloth Gloves (Tailoring 280) | Deathmist Wraps (set quest) | Sandworm Skin Gloves (Silithus quest, doubtful) |
| Legs | Eternal Sarong of Shadow Wrath (Tortheldrin, DM West, or BoE) | Eternal Sarong of Shadow Wrath | Skyshroud Leggings (Omokk, LBRS) |
| Feet | Master's Boots of Shadow Wrath (BoE world drop) | Master's Boots of Shadow Wrath | Master's Boots of Magic (BoE world drop) |
| Finger 2 | Eye of Orgrimmar (BRD quest, Horde) | Channeler's Ring (Pix, 75 Reals) | Rune Band of Wizardry (Valthalak, Dungeon Set 2 chain) |
| Trinkets | Talisman of Ascendance (Insignia quest, doubtful), Weakness Analyzer (Pix, 125 Reals) | Talisman of Ascendance, Chained Essence of Eranikus (Sunken Temple quest) | Eye of the Beast (UBRS quest), Weakness Analyzer |
| Wand | Skul's Ghastly Touch (Skul, Strat) | Bonecreeper Stylus (Gandling, Scholo) | Bonecreeper Stylus |

Weapon enchants: Mighty Intellect (Affliction), Revelation (Demonology, Destruction).

## Flags

- Hidden stats: Rugged Mantle and Knowledge of the Timbermaw. Forever hides their stats, so the sim uses
  the older version's. Both in all three lists.
- Doubtful, included by standards 42: Rugged Mantle (its Loom blueprint not seen dropping), Watcher's
  Signet (no Watchers rep source found), Mindfang (Arathi Basin unconfirmed), Talisman of Ascendance
  (Naxx-era Insignia quest), Sandworm Skin Gloves (Roman'khan chain), Rune Band of Wizardry (Valthalak
  summon unconfirmed).
- Pix Xizzix Reals, rate unknown: Affliction 200, Demonology 150, Destruction 200.
- Orb of the Darkmoon: 1200 tickets, Faire weeks only (Nov 9-16, Dec 7-14).
- Robe of the Void has an unmodelled use effect. Talisman of Ascendance's use is modelled.
- Eye of the Beast: the quest isn't named in our data.
- Enchants on Merchant's Favor rep (neck, bracer, armor kits) and Off-Hand Wisdom (Enchanting 300).

## Deathmist set check

From the first run, not repeated. The new lists keep one piece at most (Demonology's Wraps).

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

wowtbc.gg's WoW Forever pre-raid lists (no epics, no BoE world drops), from the first run. Icy Veins refused the fetch.

- Rune Band of Wizardry now matches the public list for Destruction. Sageblade, Leggings of Torment and
  Songstone of Ironforge stay out (world boss plans, patch 1.10, Alliance only).
- Mindfang, a PvP rep item the first run left out, is now our main hand. The public list has no PvP items.
- Talisman of Ascendance is a standard pre-raid caster trinket and now leads two lists.
- Tome of Shadow Force and Therazane's Touch (Frostwolf Exalted) would beat Obsidian Bauble by 0.4 to 5.0 DPS
  (Destruction's Tome is 0.6 behind) but stay no-pick, AV queue unconfirmed. Mindfang (Defilers, Arathi Basin
  unconfirmed) is included by standards 42, so the two PvP reputations are tagged differently.
- Omnicast Boots trail our BoE Master's Boots. The public list leaves out BoE world drops.
- Affliction: Dark Advisor's Pendant is an easy swap for the Orb if the Faire is missed. Maiden's Circle is
  a sure alternative to the doubtful Signet.
- Destruction and Demonology: Nacreous Shell Necklace costs about 280g, over the gold line.

## Calibration

Report build, 180s, 10000 iterations, `rotopt -eval`. The synthetic numbers match the first run's.

| Spec | Phase 0 | Synthetic | Gap | First Phase 0 list |
|---|---|---|---|---|
| Affliction | 955.8 | 1089.1 | -133.3 (-12.2%) | 920.1 |
| Demonology | 888.0 | 1056.0 | -168.0 (-15.9%) | 858.3 |
| Destruction | 894.0 | 1069.3 | -175.3 (-16.4%) | 865.8 |

The synthetic gear still reads as a raid level set, 12 to 16% above Phase 0, with 10% hit against 2 to 4%.
