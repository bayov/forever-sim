# Phase 0 step 4: warlock (2026-11-24)

The 13 warlock report requests (five builds and their curse duty variants) now wear the Phase 0 sets: Affliction
and Shadow Bolt Affliction `level60_phase0_affliction`, Demonology `level60_phase0_demonology`, Destruction and Fire
Destruction `level60_phase0_destruction` (its picks are school generic "of Magic" and Fire Destruction ranks no list of
its own). Requests come from mksettings with the meta's race, talents and APL and `-gear level60_phase0_<set>`, the duty
edits on top. Rerun with the old sets and talents, the script gave the old 13 requests exactly.

## Per build

verify.py, 180s ± 20s, 20000 iterations, seed 1. List dps is the list's own 180s figure from its RANK_ONLY log.

| Build | Race | Phase 0 | Synthetic it replaces | List (180s) |
|---|---|---|---|---|
| Affliction | Undead | 965.7 ± 0.2 | 1034.1 | 965.8 |
| Affliction curse duty | Undead | 956.6 | 1025.1 | |
| Affliction recklessness duty | Undead | 951.5 | 1019.9 | |
| Shadow Bolt Affliction | Orc | 810.4 ± 0.4 | 1015.6 | |
| Demonology | Orc | 893.7 ± 0.4 | 1053.9 | 894.0 |
| Demonology curse duty | Orc | 885.2 | 1043.6 | |
| Demonology recklessness duty | Orc | 880.1 | 1036.6 | |
| Destruction | Orc | 880.5 ± 0.4 | 1063.4 | 881.2 |
| Destruction curse duty | Orc | 873.3 | 1052.7 | |
| Destruction recklessness duty | Orc | 868.1 | 1046.4 | |
| Fire Destruction | Undead | 835.0 ± 0.3 | 996.7 | |
| Fire Destruction curse duty | Undead | 827.5 | 987.0 | |
| Fire Destruction recklessness duty | Undead | 822.3 | 980.7 | |

Every request sims to its list within noise. The lists' headline `dps` (981.0, 904.9, 888.8) reads higher because uses
are scored on a 90s fight (`use_dur` 90), as for the other classes.

## Hit, talents and rotations

Total spell hit on the Phase 0 sets at the preset race (race_gear.py): Affliction 6%, Demonology 9%, Destruction 10%,
against the 16% cap. The synthetic sets sat at the cap with `_hit±N` steps, and those steps are gone. Talent search
(`-talent-search`, 2000 iterations, 180s, the build's pins) and knob search (3000 iterations) on each build's set,
the moves confirmed at 20000 iterations, 120 / 180 / 300s:

- **Fire Destruction:** Suppression back to 5/5, `252-0050003201-0050345003101351` (Aftermath 4), was 242 from the
  hit trade's hit+1 step. +4.9 / +5.0 / +4.8. `ui/warlock/presets.ts` and the three metas.
- **Demonology and Affliction:** no Immolate (`immolate=0` in `tools/rotopt/warlock_presets.sh`, base and duty APLs).
  Demonology +10.6 / +9.6 / +10.0, Affliction +13.2 / +14.0 / +32.7. Affliction ended 180s with about 3300 mana unused,
  and Immolate's mana was worth less than its GCDs. The Immolate line is gone from their metas' rotation text.
- **Affliction talents:** `05500023035001351-0050003201-034` (Fel Concentration 1 to 3, Aftermath 4 into Improved Shadow
  Bolt 3), was `05500021035001351-0050003201-005004`. +3.7 / +3.4 / +2.8. Not a hit talent: the search found it once
  the build ran without Immolate on the re-ranked set. Presets, test and metas, golden refreshed, list re-ranked.
- Demonology (Suppression 5/5 pinned, standards 26), Destruction and Shadow Bolt Affliction: nothing moved.

The first talent searches ran on the sets before the list rerun (Affliction 8%, Destruction 9%). The rerun moved the
sets' hit, so all five were searched again on the final sets with the final knobs. Destruction and Fire Destruction were searched once more after the neck fix below (10%): no move.

## Lists

- **Stale SLOTS:** RANK_ONLY on all three lists with the new requests (Phase 0 gear, new APLs, new talents), so every
  slot is ranked against the set. Gear set changes since step 3:
  - Affliction: Spellweaver's Turban to Eternal Crown of Shadow Wrath, Arcanoweave Cloak to Fel Cape, Obsidian Bauble
    to Eternal Rod of Shadow Wrath.
  - Demonology: Deathmist Wraps to Felcloth Gloves, Chained Essence of Eranikus to Eye of the Beast, Bonecreeper
    Stylus to Skul's Ghastly Touch, Orb of the Darkmoon to Dark Advisor's Pendant.
  - Destruction: Sandworm Skin Gloves to Deathmist Wraps (one piece, no set bonus), Eye of the Beast to Talisman of
    Ascendance.
- **Nacreous Shell Necklace:** the full reruns picked it for Demonology and Destruction. Its price (282.6g, 217.4g a
  DPS, over the 30g line) was gone from `tools/rotopt/phase0_content.json`: the Enhancement price pass of `5e556ff1e`
  rewrote goldPerDps from its own list and dropped the items not on it. Restored, and the neck re-ranked with SLOTS
  (Demonology takes Dark Advisor's Pendant, Destruction keeps the Orb, both within 0.5 of the necklace, and 1% less
  hit). Other items that pass dropped (Brain Hacker, Painweaver Band) are melee, so no warlock list is affected, but
  `phase0_prices.py` should keep the prices of items it doesn't see.
- The Affliction list was re-ranked once more on its new talents. The set didn't change.
- **Deathmist trap:** none. Deathmist Raiment (Forever: 2pc resistances, 3pc +8 MP5, 6pc +23 spell power, the 4pc
  isn't modelled so 5pc equals 3pc) loses at every count, 10000 iterations at 180s, the pieces with no random suffix:

  | Spec | List set | Best 3pc | Best 6pc | 8pc |
  |---|---|---|---|---|
  | Affliction | 953.9 | 919.6 | 905.2 | 863.6 |
  | Demonology | 882.9 | 877.2 | 863.5 | 827.7 |
  | Destruction | 879.2 | 869.8 | 861.3 | 823.9 |

  Demonology's 3pc is 4 pieces, since its set already wears the Wraps. A first pass that left the replaced items'
  Shadow Wrath suffix on the Deathmist pieces showed the set up to 52 ahead. That was the suffix, not the set.

## Races

- **Faction:** Eye of Orgrimmar (12545, Horde) is Affliction's only faction pick, and no faction enchant is picked.
  RANK_ONLY SLOTS=finger at Human gives `level60_phase0_affliction_alliance.gear.json` with Songstone of Ironforge
  (12543) in its place, the finger slot only. Human sims 964.9 and Gnome 960.0 on it (Undead 965.7), and its list reads 979.9 against 981.0. The list is
  `phase0_step4_warlock/level60_phase0_affliction_alliance.bis.json`.
- **race_gear.py** (master's faction set support, from the mage session): Human and Gnome on Affliction and Shadow Bolt
  Affliction wear the Alliance set, kept by the rule, written to `req/races/` (merged with the Gnome rotation file of
  Shadow Bolt Affliction). Every other warlock row has no swap, since the Phase 0 sets have no hit steps and no warlock
  race is a Dwarf. Duty variants take their base build's files. `race_gear.json` gains the 20 warlock rows.
