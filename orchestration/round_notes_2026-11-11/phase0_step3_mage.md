# Phase 0 step 3: mage (2026-11-23)

Arcane, Fire, Frost and Frostfire, each a full `bis_phase0.sh` search at 180s (uses at 90s). Frost and Frostfire
ran after the re-tune (b8b89d6a8). Lists: `ui/mage/bis/level60_phase0_mage_<spec>.bis.json` and `.effects.json`.

## What changed on the way

- **Forever's class sets weren't in any pool.** Sorcerer's Regalia 1671 (226929 to 226936) sits in the SoD item ID
  range, and `isSodItem` drops those items unless `phase0_content.json` marks them `include`. The parent's ruling
  (Forever's Dungeon Set 2 copies, like Five Thunders) added the 8 valid mage rows, with doubtful on head and chest
  only. The first arcane and fire lists (no set) were 811 and 903. With the set they're 846 and 936.
- Unlisted picks, now listed: Sparklance Wand (Bolt or Vitrix, seen in game Nov 21), Magister's Bindings
  (Blackrock Spire trash), Spritecaster Cape (Houndmaster Grebmar, BRD, ranked only), Wand of Biting Cold (AV quest,
  doubtful, ranked only).
- The dry runs found no new unmodelled, hidden-stat or unlisted items, only a smaller pool from step 2's "out" tags.
- Container fixes, not in the repo: no `jj` here (a shim gives the script the git hash for its sim stamp), and the
  Go protos had to be generated (`make sim/core/proto/api.pb.go`). `ui/mage/bis/` didn't exist, so the first run's
  writes failed and it ran again.

## Picks

The search's set scores (list `dps`) are 846.0 Arcane, 935.5 Fire, 733.8 Frost and 884.8 Frostfire. They run
10 to 17 above a plain `-eval` of the same set (Calibration). The settings and APL are the same byte for byte, so the
gap is in how the search scores a set (uses off at 180s plus use values at 90s), not in the gear.

Enchants (all four): Arcanum of Focus on head and legs, +6 spell power necklace, Mystic Rugged Armor Kit on chest
and feet, Greater Spellpower on wrist, Spellblasting on the staff. Hands: Arcane Power for Arcane, the armor kit
for the rest.

| Slot | Arcane | Fire | Frostfire | Source |
|---|---|---|---|---|
| Head | Sorcerer's Crown (doubtful) | same | same | T0.5, end of the Valthalak chain |
| Neck | Lady Maye's Pendant (doubtful) | Orb of the Darkmoon | Star of Mystaria | world drop / Faire 1200 tickets / Balnazzar |
| Shoulder | Sorcerer's Mantle | same | same | T0.5, Anthion chain |
| Back | Arcanoweave Cloak | same | same | Pix Xizzix, 75 Reals |
| Chest | Robe of the Archmage | same | same | Tailoring 295, BRS pattern |
| Wrist | Sorcerer's Bindings | same | same | T0.5, bracer quest |
| Hands | Sorcerer's Gauntlets | same | same | T0.5, gloves and belt quest |
| Waist | Belt of the Archmage | same | same | Tailoring 295, AH pattern |
| Legs | Sorcerer's Leggings | same | same | T0.5, Anthion chain |
| Feet | Sorcerer's Sandals | same | same | T0.5, Anthion chain |
| Rings | Channeler's Ring, Band of the Damned | Channeler's Ring, Watcher's Signet of Brilliance (doubtful) | same as Fire | Pix 75 Reals / Vectus, Scholo / Watchers Revered |
| Trinkets | Chained Essence of Eranikus, Eye of the Beast | Eye of the Beast, Blackhand's Breadth | same as Fire | ST quest / BRS quests |
| Weapon | Headmaster's Charge | Elemental Mage Staff | same as Fire | Gandling, Scholo / world drop epic BoE |
| Wand | Bonecreeper Stylus | Ritssyn's Wand of Bad Mojo | same as Fire | Gandling / Rivendare |

Frost has the same armor as Frostfire, with Orb of the Darkmoon, Chained Essence of Eranikus and Eye of the Beast,
Elemental Mage Staff and Ritssyn's Wand. Its rings are Watcher's Signet of Brilliance (doubtful) and Channeler's Ring.

## Hand check flags

- **Sorcerer's Crown** (doubtful, all four): worth 15 DPS to Arcane, 10 to Fire and 3 to Frost on its own, since it's
  the sixth piece. Without it, the best 6pc (Spellweaver's Turban plus Sorcerer's Belt) sims 815.0 Arcane, 908.0
  Fire, 716.4 Frost.
- **Elemental Mage Staff** (Fire, Frost, Frostfire): Forever made it an epic, item level 61, random world drop, BoE.
  It's 15.6 DPS over the next weapon for Fire, 20.6 for Frostfire and 6.2 for Frost. No AH price is known, so the 30g per
  DPS cap can't check it. Arcane takes Headmaster's Charge.
- **Lady Maye's Pendant** (Arcane, doubtful, no Forever source): Star of Mystaria is 1.7 behind.
- **Watcher's Signet of Brilliance** (Fire, Frost, Frostfire, doubtful, no rep source found): Flaming Band of
  Intellect is 1.7 behind for Fire, even for Frostfire.
- **Orb of the Darkmoon** (Fire, Frost): needs a Faire week (next Dec 7 to 14).
- **Nacreous Shell Necklace** beats every neck pick by 0.7 to 2.3, but 217g per DPS is over the cap.
- Uses and procs: Chained Essence of Eranikus (Arcane) is modelled. None of the picks have an unmodelled effect or
  hidden stats. The Sorcerer's 5pc (when struck) isn't modelled and doesn't matter for DPS.
- Robe of the Archmage and Belt of the Archmage are BoP tailoring: a non-tailor loses them. Sorcerer's Robes and
  Belt are the set's own fallbacks.

## Set check

Each spec's picks at 180s, 10000 iterations, with set pieces swapped out for each slot's best non-set item
(Spellweaver's Turban, or Eye of Flame (doubtful) for Frostfire, on head, Mantle of the Timbermaw (doubtful) on
shoulder, the next legs and feet).

| | 6pc | 4pc | 2pc |
|---|---|---|---|
| Arcane | 830.4 | 814.3 | 803.4 |
| Fire | 918.2 | 909.4 | 884.2 |
| Frostfire | 874.9 | 860.9 | 835.4 |
| Frost (first search took 2pc) | 719.7 | 715.8 | 718.1 |
| Frost 6pc without the Crown (Sorcerer's Belt) | 716.4 | | |

Frost's first search stopped at 2 pieces (730.8 by its score, 718.1 by `-eval`), since a 4pc is worth less than
the 2pc picks for a frost mage who never runs out of mana. Frost now starts from the 6pc set (`start` in
`phase0_specs.json`) and its search kept all 6 (733.8 by its score, 719.7 by `-eval`). The 1.6 DPS comes from the
6pc's +23 spell power and needs the doubtful Crown.

## Public lists

No public Arcane or Frostfire list. Fire lists found (the fetch tool's summaries, not read directly):

- conquestcapped.com, Forever, T0.5: our Fire list matches it on head, shoulder, wrist, hands, legs, feet, back,
  chest, waist, Channeler's Ring and Eye of the Beast. Differences: Rune Band of Wizardry and Sageblade are "out" in
  our tags (Valthalak, and world boss plans). Nacreous Shell Necklace is over our cost cap. Weakness Analyzer ranks
  6.5 behind Blackhand's Breadth. They use Sageblade plus Obsidian Bauble, we use Elemental Mage Staff.
- mythicsim.com, Forever, BRS focus: Spellweaver's Turban and Mantle of the Timbermaw lose to the 6pc. Dryad's Wrist
  Bindings (WSG exalted) and Don Rodrigo's Band (AV exalted) are infeasible in our tags. Gloves of Spell Mastery is
  "out" (raid pattern). Pendant of Celerity and Rune Band of Wizardry are "out". Same staff as ours.
- noobtoboss.com, generic Classic: older choices (Hands of Power, Skyshroud Leggings, Clutch of Andros), all ranked
  well below the set pieces here.

One thing to check: Rune Band of Wizardry and Pendant of Celerity are tagged "out" as Dungeon Set 2 chain items
(Lord Valthalak). If Forever's T0.5 chain is open, Valthalak may be too.

## Calibration

Report build at 180s, 10000 iterations, `rotopt -eval`, the Phase 0 set against the report's synthetic gear. The
synthetic numbers match the published ones (Arcane 962.9 against 962.3).

| Spec | Phase 0 | Synthetic | Gap |
|---|---|---|---|
| Arcane | 830.4 | 962.9 | -132.5 (-13.8%) |
| Fire | 918.2 | 1052.8 | -134.6 (-12.8%) |
| Frost | 719.7 | 838.5 | -118.8 (-14.2%) |
| Frostfire | 874.9 | 1007.4 | -132.5 (-13.2%) |

The synthetic Phase 1 gear sits 13 to 14% above the Phase 0 sets for every mage spec, so the specs are ranked about
the same either way.
