# Phase 0 step 3: mage (2026-11-23, rerun on master 5e0a327)

Arcane, Fire, Frost and Frostfire, each a full `bis_phase0.sh` search at 180s (uses at 90s), rerun per
`briefs/phase0_rerun_1123.md` after the item data fixes (Dodge, Block, Parry and haste, Blackhand's Breadth and Eye of
the Beast crit types, weapon skill, faction by race, standards 42 flips, the 26 Forever item procs and the weapon
damage, resistance and class limit fixes). Lists: `ui/mage/bis/level60_phase0_mage_<spec>.bis.json` and
`.effects.json`.

## What changed from the earlier lists

- **Arcane** (846.0 to 855.2 by the search's score): Knowledge of the Timbermaw and Rugged Mantle of the Timbermaw
  (Tailoring, Meilosh) take waist and shoulder. Sorcerer's Robes replaces Robe of the Archmage to keep the 6pc. Star of
  Mystaria for Lady Maye's Pendant, Songstone of Ironforge for Band of the Damned.
- **Fire** (935.5 to 930.0): Blackhand's Breadth is out (its crit is melee and ranged only now), Chained Essence of
  Eranikus takes its place. Rune Band of Wizardry (in by standards 42) replaces Watcher's Signet. Knowledge of the
  Timbermaw on waist, Star of Mystaria 0.5 ahead of Orb of the Darkmoon.
- **Frost** (733.8 to 740.4): Knowledge of the Timbermaw on waist, Rune Band of Wizardry for Channeler's Ring.
- **Frostfire** (884.8 to 883.4): Blackhand's Breadth out for Chained Essence of Eranikus. Rune Band of Wizardry for
  Watcher's Signet. Knowledge of the Timbermaw on waist.

## PvP reputation check (b1bbe4347)

b1bbe4347 made Exalted rewards infeasible and put the other AV and AB items in by standards 42. DRY_RUN on it, done
by the warlock worker: the pools are unchanged (1275 items). No mage pick was an Exalted reward. Each newly included
item that fits a mage was already ranked and loses to the pick: Stormpike Cloth Girdle 14 to 23 DPS, Crackling Staff 8
to 30, Wand of Biting Cold (Frost) 2.8. So no slot was rerun and the lists and calibration stand.

## What changed on the way

- Golden Sandals (arcane spell power) was an unlisted arcane pick in the first, stopped run. It's a Tailoring 235 pattern
  from Jim'bek for 90 Merchant's Favor (Honored), like the listed Black Sandals. It and the mage school siblings
  (Netherfroth and Nethershine Cuffs, Fiery Waistcord, Frothing and Fiery Sandals) are now in `phase0_content.json`.
  None of them is a pick on the final lists.
- The dry runs on 5e0a327 found no new unmodelled or unlisted item worth a look: the Stormpike Insignias (no DPS
  effect) and three hidden-stat items (below). Sul'thraze and Demonshear left the unmodelled list.
- Container fixes, not in the repo: a `jj` shim for the sim stamp, protoc and protoc-gen-go for the Go protos.

## Picks

Calibration `-eval` matches each search's final line exactly (843.5, 921.1, 726.8, 872.4).

Enchants (all four): Arcanum of Focus on head and legs, +6 spell power necklace, Mystic Rugged Armor Kit on chest
and feet, Greater Spellpower on wrist, Spellblasting on the staff. Hands: Arcane Power for Arcane, the armor kit
for the rest.

| Slot | Arcane | Fire | Frost | Frostfire | Source |
|---|---|---|---|---|---|
| Head | Sorcerer's Crown (doubtful) | same | same | same | T0.5, end of the Valthalak chain |
| Neck | Star of Mystaria | Star of Mystaria | Orb of the Darkmoon | Star of Mystaria | Balnazzar / Faire 1200 tickets |
| Shoulder | Rugged Mantle of the Timbermaw (doubtful, hidden stats) | Sorcerer's Mantle | same as Fire | same as Fire | Meilosh, Tailoring / T0.5 |
| Back | Arcanoweave Cloak | same | same | same | Pix Xizzix, 75 Reals |
| Chest | Sorcerer's Robes (doubtful) | Robe of the Archmage | same as Fire | same as Fire | T0.5 / Tailoring 295, BRS pattern |
| Wrist | Sorcerer's Bindings | same | same | same | T0.5, bracer quest |
| Hands | Sorcerer's Gauntlets | same | same | same | T0.5, gloves and belt quest |
| Waist | Knowledge of the Timbermaw (hidden stats) | same | same | same | Meilosh, Tailoring, 20 Winterfall Spirit Beads |
| Legs | Sorcerer's Leggings | same | same | same | T0.5, Anthion chain |
| Feet | Sorcerer's Sandals | same | same | same | T0.5, Anthion chain |
| Rings | Channeler's Ring, Songstone of Ironforge | Channeler's Ring, Rune Band of Wizardry | Watcher's Signet of Brilliance (doubtful), Rune Band of Wizardry | same as Fire | Pix 75 Reals / Ironforge chain, Alliance / Valthalak / Watchers Revered |
| Trinkets | Chained Essence of Eranikus, Eye of the Beast | same | same | same | ST quest / BRS quest |
| Weapon | Headmaster's Charge | Elemental Mage Staff | same as Fire | same as Fire | Gandling, Scholo / world drop epic BoE |
| Wand | Bonecreeper Stylus | Ritssyn's Wand of Bad Mojo | same as Fire | same as Fire | Gandling / Rivendare |

All four keep the Sorcerer's 6pc: Arcane wears the Robes in place of the Mantle, the rest wear the Mantle.

## Hand check flags

- **Knowledge of the Timbermaw** (all four) and **Rugged Mantle of the Timbermaw** (Arcane): Forever still hides
  their stats, so the search uses the SoD values (25 and 27 spell power). The belt is 7.3 to 10.3 over Belt of the
  Archmage. If Forever's numbers are lower, Belt of the Archmage is the fallback. Both need Tailoring, which the
  report builds have. The Mantle is doubtful (Loom blueprint unseen), 6.3 over Mantle of the Timbermaw for Arcane.
- **Sorcerer's Crown** (doubtful, all four) and **Sorcerer's Robes** (doubtful, Arcane): as before, the end of the
  Valthalak chain. Robe of the Archmage is 5.2 behind for Arcane.
- **Rune Band of Wizardry** (Fire, Frost, Frostfire): Lord Valthalak, in by standards 42. Arcane has it 0.03 behind
  Songstone.
- **Watcher's Signet of Brilliance** (Frost, doubtful, no rep source found): Channeler's Ring is 0.06 behind.
- **Elemental Mage Staff** (Fire, Frost, Frostfire): epic BoE world drop, no AH price known.
- **Orb of the Darkmoon** (Frost): needs a Faire week (next Dec 7 to 14). Star of Mystaria is 1.9 behind.
- **Songstone of Ironforge** (Arcane): Alliance only. The report's race is Alliance, and Horde gets Eye of Orgrimmar.
- **Nacreous Shell Necklace** and **Warden's Signet of Brilliance** beat some picks by up to 2.8 but are infeasible
  (cost cap, Exalted).
- Uses and procs: Chained Essence of Eranikus is modelled. No pick has an unmodelled effect.

## Public lists

No public Arcane, Frost or Frostfire list. Fire lists (the fetch tool's summaries from the earlier pass):

- conquestcapped.com, Forever, T0.5: our Fire list now matches it on Rune Band of Wizardry too, besides head,
  shoulder, wrist, hands, legs, feet, back, chest, Channeler's Ring and Eye of the Beast. Differences: Sageblade
  (world boss plans, out), Nacreous Shell Necklace (over our cost cap), Belt of the Archmage (Knowledge of the
  Timbermaw beats it by 7.3 here, on hidden stats), Elemental Mage Staff for Sageblade plus Obsidian Bauble.
- mythicsim.com, Forever, BRS focus: Spellweaver's Turban and Mantle of the Timbermaw lose to the 6pc. Dryad's Wrist
  Bindings and Don Rodrigo's Band are Exalted (infeasible). Gloves of Spell Mastery is a raid pattern. Same staff.
- noobtoboss.com, generic Classic: older choices, all well below the set pieces here.

## Calibration

Report build at 180s, 10000 iterations, `rotopt -eval`, the Phase 0 set against the report's synthetic gear.

| Spec | Phase 0 | Synthetic | Gap | Earlier Phase 0 |
|---|---|---|---|---|
| Arcane | 843.5 | 962.9 | -119.4 (-12.4%) | 830.4 |
| Fire | 921.1 | 1052.8 | -131.7 (-12.5%) | 918.2 |
| Frost | 726.8 | 838.5 | -111.7 (-13.3%) | 719.7 |
| Frostfire | 872.4 | 1007.4 | -135.0 (-13.4%) | 874.9 |

The synthetic Phase 1 gear sits 12 to 13% above the Phase 0 sets for every mage spec, so the specs rank the same
either way.
