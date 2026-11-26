# Phase 0 re-rank, rogue (2026-11-26)

Main hand, off hand, ranged and trinkets re-ranked on master 34fe148 (`RANK_ONLY=1 SLOTS=mainhand,offhand,ranged,trinket`), then the set trap.
Request dps is verify.py, 180 ± 20s, 20,000 iterations, seed 1. "Before" is the prices round (`phase0_prices_1125.md`).

| Build | Slot changes | List dps | Request dps |
|---|---|---|---|
| Combat Swords | trinket Chained Essence of Eranikus -> Heart of Wyrmthalak, ranged Bloodseeker -> Blackcrow | 850.43 -> 854.71 | 840.96 -> 853.07 |
| Combat Daggers | trinket Maelstrom -> Heart of Wyrmthalak, hands Darkmantle Grips -> Devilsaur Gauntlets, legs Darkmantle Pants -> Devilsaur Leggings | 840.86 -> 839.23 | 828.81 -> 830.58 |
| Seal Fate Swords | trinket Maelstrom -> Heart of Wyrmthalak | 859.55 -> 857.21 | 848.83 -> 850.31 |
| Mutilate | trinket Maelstrom -> Heart of Wyrmthalak, Devilsaur pair as Combat Daggers | 798.0 -> 797.00 | 785.09 -> 788.45 |
| Subtlety | trinket Maelstrom -> Heart of Wyrmthalak, Devilsaur pair | 792.8 -> 798.55 | 786.46 -> 788.79 |
| Seal Fate Daggers | trinket Maelstrom -> Heart of Wyrmthalak, Devilsaur pair | 825.18 -> 829.16 | 818.47 -> 821.30 |

- Every request sims to its list's final measure within 0.7.
- Alliance rows (Human, on the `_alliance` sets, Theramore Pauldrons): Combat Swords 846.94, Seal Fate Swords 844.57. Those sets are the new Horde sets with the one shoulder swap.
- Heart of Wyrmthalak's proc is modelled (Forever Flame Lash, 1.95 a minute, EV-0133). The procs batch lifted it over Maelstrom and Eranikus on every build.
- List dps moves with the sim batches as well as the picks, so a list can fall while its picks gain.
- Weapons stay: Sword of Zeal and Greenhammer (Combat Swords), Glacial Blade and Greenhammer (Combat Daggers), Sword of Zeal and Distracting Dagger (Seal Fate Swords), Glacial Blade or Electrified Dagger with Distracting Dagger (the dagger builds).
- Blackcrow (Combat Swords) and the Devilsaur pair have no rogue `goldPerDps` for those specs yet. The prices run decides whether they stay under the line. Devilsaur Leggings are about 56g on the AH, the Gauntlets 11g.

## Set trap

Sims at 5000 iterations, seed 1, against the build's new set. Sets a rogue can wear in Phase 0: Darkmantle (Forever 1676, 6 pieces in the pool, the Cap and Tunic are not), Shadowcraft, Devilsaur, Cadaverous, Stormshroud, Dal'Rend's, The Defiler's Purpose.

| Build | Darkmantle 6 | Shadowcraft 2 / 4 / 6 / 8 | Devilsaur 2 | Dal'Rend's 2 |
|---|---|---|---|---|
| Combat Swords (853.1, wears Darkmantle 4) | -5.5 | - / -38.4 / -46.2 / -52.7 | -4.2 | -10.7 |
| Seal Fate Swords (851.2, Darkmantle 4) | -8.0 | - / -43.6 / -51.6 / -58.6 | -9.5 | -10.6 |
| Combat Daggers (830.4, Darkmantle 6) | worn | -22.3 / -29.8 / -36.7 / -43.2 | +0.8 | |
| Mutilate (787.2, Darkmantle 6) | worn | -22.7 / -29.5 / -36.9 / -42.9 | +1.3 | |
| Subtlety (787.8, Darkmantle 6) | worn | -21.1 / -27.5 / -34.3 / -40.4 | +1.2 | |
| Seal Fate Daggers (820.1, Darkmantle 6) | worn | -24.4 / -31.2 / -38.7 / -45.5 | +2.0 | |

- The Devilsaur pair (+2% hit) beats Darkmantle 6 on the four dagger builds, which sat at 8.7% hit. Confirmed on Combat Daggers at 20,000 iterations (829.8 -> 830.6).
- With Devilsaur on, Darkmantle drops to 4 pieces. Taking the sword builds' Darkspear Pauldrons or Assassin's Waistguard as well loses 2.3 to 12.8 against the Devilsaur set (Mutilate, Seal Fate Daggers), so the Darkmantle shoulder and belt stay.
- Hands and legs ranked again from the Devilsaur start (`SLOTS=hands,legs`): Devilsaur tops both on all four, list +0.65 to +2.32. Raider Gloves and Outrider's Leather Pants are next, 5 to 6 behind.
- Dagger builds now sit at 10.3% melee hit. The Dwarf hit trade still loses (-0.05 to -0.94), so no Dwarf row trades it.
- Cadaverous (defense, 10 AP, resists), Stormshroud and The Defiler's Purpose (1% crit at 3 pieces) cost 30 to 60 in pieces for little, not simmed.
- Not done: a full re-rank of every other slot at the new hit. Only hands and legs were ranked from the Devilsaur start.

## Open item

- **Bloodseeker** (Alterac Valley quest, AV queue unconfirmed): by standards 42 an unsure source stays in, with a note, and `phase0_content.json` already has it as "doubtful", include. It stays the ranged pick of Combat Daggers, Mutilate and Seal Fate Daggers. On Combat Swords it fell to 2.5 behind Blackcrow.

## Sim fix

- Subtlety's trinket use values panicked: Servant of Alexi Barov "cannot wait negative time". `sim/common/guardians/barov_servant.go` now guards its swing wait like the other guardians (`core_hound.go`). Rogue and item set tests pass.

## Other

- The container had no `protoc`. Installed `protobuf-compiler` and `protoc-gen-go` v1.36.6 to build.
- Effects files refreshed with `EFFECTS_ONLY` for all six specs.
