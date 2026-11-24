# Phase 0 step 4, rogue

Every rogue report build now runs on its own Phase 0 set (`tools/rotopt/bis_phase0.sh`), and the synthetic hit steps are gone.
All numbers are 180 ± 20s, 20,000 iterations, seed 1.

| Build | Race | Phase 0 dps | Synthetic dps (published) | Change | Phase 0 set | List dps | Melee hit |
|---|---|---|---|---|---|---|---|
| Combat Swords | Undead | 844.1 | 963.5 (`_swords_hit+2`) | -119.4 | `level60_phase0_rogue_combat_swords` | 852.42 | 10.7% |
| Combat Daggers | Undead | 853.7 | 946.6 (`_combat_daggers_hit+2`) | -92.9 | `level60_phase0_rogue_combat_daggers` (new) | 861.97 | 11.3% |
| Mutilate | Undead | 821.0 | 866.1 (`_mutilate_hit+3`) | -45.1 | `level60_phase0_rogue_assassination` | 830.21 | 8.7% |
| Seal Fate Daggers | Undead | 842.4 | 920.7 (`_daggers_hit+3`) | -78.3 | `level60_phase0_rogue_seal_fate_daggers` (new) | 851.62 | 8.7% |
| Seal Fate Swords | Undead | 864.2 | 957.0 (`_seal_fate_swords_hit+3`) | -92.8 | `level60_phase0_rogue_seal_fate_swords` (new) | 869.32 | 9.7% |
| Subtlety | Night Elf | 816.1 | 872.2 (`_subtlety_hit+2`) | -56.1 | `level60_phase0_rogue_subtlety` | 825.74 | 8.3% |

- **Phase 0 dps** is verify.py on the new request. It matches the final measure of its list run within 0.6 dps on every build.
- **List dps** is the search's own score (90s use values), not the request's dps.
- The synthetic numbers come from the last report run. Step 3 re-simmed three of them on today's sim: Combat Swords 948.4, Mutilate 870.3, Subtlety 875.4.

## Lists

- The three old lists (Combat Swords, Assassination, Subtlety) were re-ranked with RANK_ONLY.
  - Their picks are unchanged. Their list dps moved from 855.11, 833.81 and 829.01 because the slot costs changed (SLOTS numbers).
- Combat Daggers, Seal Fate Daggers and Seal Fate Swords had no list. Their weapon types or talents differ from their spec's list, so each got its own spec in `tools/rotopt/phase0_specs.json` and a RANK_ONLY run.
  - The dagger builds started from Assassination.
  - Seal Fate Swords started from Combat Swords.
- Common picks: Outlaw's Collar, Dawn Armor, Cape of the Black Baron, Darkmantle bracers, gloves, pants and boots, Watcher's Signet of Ferocity and Cutthroat's Signet.

| Build | Weapons | Other picks |
|---|---|---|
| Combat Swords | Sword of Zeal, Felstriker | Darkspear Pauldrons, Assassin's Waistguard, Mark of Fordring, Eranikus and Maelstrom trinkets, Deep Jungle Rifle |
| Combat Daggers | Felstriker, Greenhammer (mace) | Darkmantle Spaulders and Belt, Amulet of the Darkmoon, Maelstrom and Eranikus, Deep Jungle Rifle |
| Mutilate | Felstriker, Felstriker | Darkmantle Spaulders and Belt, Blackcrow |
| Seal Fate Daggers | Felstriker, Distracting Dagger | Darkmantle Spaulders and Belt, Mark of Fordring, Blackcrow |
| Seal Fate Swords | Sword of Zeal, Felstriker | Darkspear Pauldrons, Assassin's Waistguard, Blackhand's Breadth and Maelstrom, Deep Jungle Rifle |
| Subtlety | Felstriker, Felstriker | Darkmantle Spaulders and Belt, Deep Jungle Rifle |

## Stones by weapon type

The search scores every weapon with the request's stone, but Sharpening Stone fits only edged weapons and Weightstone fits only blunt ones (`ImbueFitsWeapon`).

- `bis_phase0.sh` now takes `MH_IMBUE` and `OH_IMBUE`.
- Each build's weapon slots were ranked again with the other stone, written to scratch lists.

The results:
- **Combat Swords off hand.** Felstriker on Sharpening Stone beats Greenhammer on Weightstone by 2.0.
  - The request's off hand is now Dense Sharpening Stone, an edit made after mksettings and noted in the meta.
  - Off-hand slot re-ranked: 852.42.
- **Combat Swords main hand.** On Weightstone, the best blunt weapon is Bloody Claws of Attack, about 7 below Sword of Zeal on Sharpening Stone. Kept.
- **Combat Daggers off hand.** Greenhammer on Weightstone wins at 861.97, 3.4 over the Sharpening Stone ranking.
  - The off hand is back to mksettings' Dense Weightstone. My earlier switch to Sharpening Stone was wrong.
- **Seal Fate Daggers, Seal Fate Swords and Subtlety off hands.** The best blunt weapon on Weightstone is 6 to 14 behind the edged pick. Kept.

## Hit and talents

Phase 0 melee hit, from dashstats on each request, is 8.3% to 11.3%. That is against 12% to 13% on the synthetic steps.
- The round 70753 hit trade (`orchestration/round_70753/hit_trade/rogue.md`) searched steps from the cap step (8%) up to +3, found no talent change at any of them, and kept Precision 3/3.
- Every Phase 0 hit lies inside that range, so the talents and rotations stay.
- Mutilate and Subtlety, which had the thinnest Precision margin (1.4 to 1.8), are now nearer the cap, where Precision is worth more.
- The Dwarf's +1% hit, traded back for Agility, loses or ties on every build (-0.05 to -0.65), so no Dwarf row trades it.

## Race rows

`race_gear.py` gives a race of the other faction `<gearSet>_<faction>.gear.json` when that file exists.

- Combat Swords and Seal Fate Swords pick Darkspear Pauldrons, which are Horde only.
  - A shoulder SLOTS run at Human picked Theramore Pauldrons.
  - The new `_alliance` sets are the Horde sets with that one swap.
  - Human on them: Combat Swords 836.9, Seal Fate Swords 858.8. That matches the Alliance list runs (836.1 with the old off hand, and 858.7).
  - The Alliance shoulder has 1% less hit, so Alliance rows sit at 9.7% hit.
- The other four builds pick no faction item, so they have no faction set.
- Race-specific weapon picks don't apply, since Forever's racials give no weapon skill.
- Removed `rogue__seal-fate-daggers.Dwarf.json`, an old hit trade that now loses.
- Kept the Gnome rotations of Mutilate and Seal Fate Swords. The Seal Fate Swords Gnome row now also wears the Alliance set. Its rotation gain, +4.7 on the old gear, is +3.6 on the Alliance Phase 0 set (863.6 against the preset rotation's 860.0, 20,000 iterations, seed 1; seed 1000003 also +3.6), so it stays (2026-11-24 publish check).

## Problems

- Deathmaw's Guise (282049) is not in `phase0_content.json`.
  - It ranks 5th on Seal Fate Swords' head, so it is not a pick.
  - The run stopped before writing the gear set, so I wrote it with GEAR_ONLY. The item still needs its Phase 0 check.
- Alliance races lose about 6 to 7 dps on the sword builds from the shoulder alone.
