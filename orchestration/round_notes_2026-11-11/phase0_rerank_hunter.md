# Phase 0 re-rank: hunters (2026-11-26)

Brief: `briefs/phase0_rerank_group_1125.md`. Specs hunter_bm, hunter_mm, hunter_lw, hunter_sv.
`RANK_ONLY=1 SLOTS=mainhand,offhand,ranged,trinket` on master 34fe148. Request DPS is wowsimcli on the report
request, 180 ± 20s, seed 1.

## Per build

| Spec | Slot changes | List dps | Request dps |
|---|---|---|---|
| Beast Mastery | none | 1040.11 -> 1051.32 | 1038.45, checked 1038.43 |
| Marksmanship | none | 1096.77 -> 1096.77 | 1082.63, checked 1082.38 |
| Lone Wolf | none | 1144.28 -> 1155.45 | 1140.13, checked 1140.44 |
| Survival | trinkets Chained Essence of Eranikus + Darkmoon Card: Maelstrom -> Maelstrom + Heart of Wyrmthalak | 955.17 -> 963.70 | 950.39 -> 963.18 |

- A list's dps is its ranking score. The run measured the final set again, so it moves without a pick change.
- The checked values come from 10,000 iterations. The Survival requests ran at 20,000. The run measured the lists at
  180s at 1038.4 (BM), 1082.4 (MM), 1140.4 (LW) and 963.2 (SV), so every request matches its list within noise.
- The Survival pet builds' requests moved too: wolf 927.82 -> 941.19 and tallstrider 918.73 -> 931.59.
- Main hand, off hand and ranged keep their picks on all four lists. Beastslayer's Trident is the ranged main hand.
  Hyper Deluxe is the BM and MM ranged pick, Eaglehorn is LW's, and SV keeps Sword of Zeal, Distracting Dagger and
  Ancient Bone Bow.
- Tagged the unlisted MM and LW entries as include, "Random world drop · rare BoE, AH price not checked":
  Warmonger (13052) and Assassination Blade (13036). Neither is a pick.

## Survival follow-up

- `GEAR_ONLY`: the run already wrote the set.
- The Survival requests now use the new trinkets. mksettings gives the same equipment.
- The Alliance copy's shoulder was re-ranked at Human from the new list. It is still Theramore Epaulets and now has the
  new trinkets.
- race_gear.py `--write` was run on the three Survival builds. Dwarf's 1% excess-hit trade is +0.02 ± 0.15, so it is
  dropped.

## Set trap

The Beastmaster set (Forever, 1669) bonuses are: 4pc 5% on auto attack for 200 mana, 6pc +40 AP. Devilsaur 2pc is
already worn on every list. Results from 10,000 iterations on the report request:

| Build | Base | Tunic + Belt (4pc) | + Mantle (6pc) | + Pants (6pc, no Devilsaur) |
|---|---|---|---|---|
| Marksmanship | 1082.4 | 1062.9 | 1056.2 | 1056.5 |
| Beast Mastery | 1038.4 | 1017.2 | 1010.7 | 1009.8 |
| Lone Wolf | 1140.4 | 1115.9 | 1108.4 | 1111.7 |

Survival wears 2 pieces. It scored 950.6 base, 941.4 with Cap + Belt (4pc) and 929.7 with Cap + Belt + Tunic + Mantle
(6pc).

Every variant loses 9 to 32. Beaststalker needs 5 pieces, 3 of them in Beastmaster slots, at 3 to 19 loss each, so it
is not a trap. The lists stay.

## Open items

- **Hit:** no hit item is missing under the 30g per DPS line. The hit items left out are:
  - Exalted grind: Outrider's Chain Leggings, Don Julio's Band and Warden's Signet.
  - Well over the line: Windreaver Greaves (88 to 2064 g/DPS), Deep Jungle Rifle, Blackcrow and Satyr's Bow (92 to 118
    g/DPS) and Bloodmail Boots (875).
  - Accurascope: its schematic drops in Molten Core.
  - Forever ranged-hit gear, which is ranked and loses 1 to 6: Marksman Bands, Sash of the Grand Hunt, Stoneshatter and
    Marksman's Girdle.
- **Hyper Deluxe Sniper Rifle Mk XVII:** include, doubtful, under standard 42. It is an Engineering 300 craft. Its
  schematic (279322) has no known source, but nothing shows it is raid loot. The note says so. BM's backup is Dwarven
  Hand Cannon at 2.39 behind, a sure source at 77.8g.
