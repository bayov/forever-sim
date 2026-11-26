# Phase 0 re-rank: warrior (2026-11-26)

RANK_ONLY SLOTS=mainhand,offhand,ranged,trinket on master 34fe148 (the item, proc, caster and log-conflict batches).
Requests verified with verify.py, 180s ± 20s, 20000 iterations, seed 1, SE 0.4 to 0.5. "Before" is each meta's verify
on master, not the step 4 table (Fury DW moved since then).

## Per build

| Spec | Slot changes | List dps before -> after |
|---|---|---|
| Arms | Trinket: Darkmoon Card: Maelstrom -> Heart of Wyrmthalak (+0.17, noise, but Maelstrom is doubtful) | 1018.51 -> 1017.05 |
| Fury 2H | none | 996.34 -> 999.30 |
| Fury DW | Main hand: Ironwood Blade -> Thrash Blade. Trinket: Blackhand's Breadth -> Heart of Wyrmthalak | 988.86 -> 993.76 |

Fury DW: Felstriker went from -8.65 over the off hand pick to 8.93 under it. The main hand imbue stays a Dense Sharpening
Stone (Thrash Blade is a sword).

| Request | Before | After |
|---|---|---|
| Arms | 1011.08 | 1012.14 |
| Arms, Battle Shout duty | 1008.88 | 1009.68 |
| Arms, Sunder duty | 962.47 | 963.85 |
| Arms, Dragonkin | 1038.61 | 1039.44 |
| Fury 2H | 994.41 | 994.16 |
| Fury 2H, Battle Shout duty | 989.65 | 990.27 |
| Fury 2H, Sunder duty | 961.15 | 959.75 |
| Fury DW | 983.62 | 992.33 |
| Fury DW, Sunder duty | 948.48 | 960.15 |
| Fury DW, Battle Shout duty | 979.67 | 988.19 |

Requests against lists: Arms -4.9, Fury 2H -5.1, Fury DW -1.4. The lists score at exactly 180s with uses on 90s,
and Arms showed the same 5.0 gap at step 4, so within what we expect. Fury 2H's request didn't change.

Faction copies: `level60_phase0_arms_alliance` and `level60_phase0_fury_dw_alliance` take the same swaps (their faction
slots are hands and shoulders). race_gear.py rerun: Alliance rows wear the new items. Fury DW hit at Orc is now 10%
(was 11%, Heart of Wyrmthalak for Blackhand's Breadth), so the race hit steps moved by 1. The Fury DW Dwarf trade loses
(-1.77 ± 0.24), no bonus stats. Fury 2H Dwarf keeps its trade, +4.52 ± 0.09.

## Set trap

Evals at 180s, 10000 iterations, against the gear set (SE 0.6). Best subset of each set's pieces, delta vs the set:

| Set | Arms | Fury 2H | Fury DW |
|---|---|---|---|
| Devilsaur 2pc | -10.3 | -12.5 | -11.9 |
| Black Dragon Mail 2pc | -22.4 | -26.1 | -35.1 |
| Black Dragon Mail 3pc | -46.0 | -49.1 | -42.9 |
| Imperial Plate 3pc | -22.9 | -23.1 | -33.4 |
| Imperial Plate 5pc | -38.9 | -45.3 | -57.4 |
| Imperial Plate 6pc | -55.8 | -65.3 | -76.7 |
| Blessed Plate 3pc | -17.5 | -21.3 | -26.1 |
| Blessed Plate 5pc | -27.4 | -36.5 | -52.7 |
| Valor 3pc | -19.0 | -18.3 | -22.3 |
| Stormshroud 4pc | | -72.8 | |

Heroism was checked at step 3 and lost. No set comes within 10, so the lists stay.

## Other

- Dragon's Call (10847, Shade of Eranikus, Sunken Temple) was unlisted in the Fury DW off hand top 5. Tagged open in
  `phase0_content.json`. Its whelp proc is modelled.
- No open items for this group.
- Container note: a fresh container needs `protobuf-compiler` (apt) and `protoc-gen-go` before bis_phase0.sh builds.

## Prices round 2 (master 6a5cace)

Heart of Wyrmthalak is over the 30g line for Arms (155.2g per DPS) and Fury DW (51.9), Satyr's Bow for Fury DW (31.4).
RANK_ONLY SLOTS=trinket (Arms) and SLOTS=trinket,ranged (Fury DW).

| Spec | Slot changes | List dps | Request (verify) |
|---|---|---|---|
| Arms | Trinket: Heart of Wyrmthalak -> Darkmoon Card: Maelstrom | 1017.05 -> 1016.65 | 1012.14 -> 1010.49 |
| Fury DW | Trinkets: Heart of Wyrmthalak + Maelstrom -> Maelstrom + Blackhand's Breadth. Ranged: Satyr's Bow -> Bloodseeker | 993.76 -> 991.53 | 992.33 -> 986.50 |

Duties: Arms Battle Shout 1008.06, Sunder 962.35, Dragonkin 1037.88. Fury DW Sunder 952.83, Battle Shout 981.43.
Fury 2H unchanged (994.16, no drift). Requests vs lists: Arms -6.2, Fury DW -5.0, in line with the 5 DPS gap Arms has
shown since step 4 (lists at exactly 180s, requests 180s ± 20s).

Alliance copies follow. race_gear: Fury DW hit at Orc is 9% (was 10%). The Fury Dwarf trade loses (-1.28), so no bonus
stats. Arms Dwarf trade kept, +5.79. Fury 2H Dwarf kept, +4.52.
