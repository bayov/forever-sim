# Phase 0 step 3: hunter (rerun of 2026-11-23)

Beast Mastery, Marksmanship and Survival, from `tools/rotopt/bis_phase0.sh` on master 566067a (item
procs, weapon damage, resistances, class limits, Exalted rewards infeasible, the faction check and the
pet's pre-pull Aspect of the Hawk). Survival ran on a47a463 and was ranked again on 566067a with
`RANK_ONLY`, and no pick changed. Lists in `ui/hunter/bis/` (`level60_phase0_hunter_<spec>`).

## Summary

| Spec | Phase 0 | Synthetic | Gap | Earlier Phase 0 |
|---|---|---|---|---|
| Beast Mastery | 1014.6 | 1112.9 | -98.3 (-8.8%) | 1018.7 |
| Marksmanship | 1102.2 | 1200.1 | -97.9 (-8.2%) | none |
| Survival | 937.2 | 1007.3 | -70.1 (-7.0%) | 937.8 |

Calibration: each spec's report build at 180s, 10000 iterations, `rotopt -eval` on the list's picks
with their enchants and on the report's synthetic gear. "Earlier Phase 0" is the list each spec had
before b1bbe4347, simmed the same way on today's data. Its higher DPS comes from items that are now
infeasible. Marksmanship had no finished earlier list.

What changed:

- Beast Mastery: Don Julio's Band went infeasible, so Eagle's Head Signet took the ring. Beastmaster's
  Bindings and Darkspear Pauldrons gave way to Bracers of the Eclipse and Darkspear Epaulets.
- Marksmanship: first list.
- Survival: the Defiler's shoulders, girdle and boots and both Exalted rings went infeasible. Darkspear
  Epaulets, Ranger's Belt, Windreaver Greaves, Watcher's Signet of Ferocity and Cutthroat's Signet in.

Beastmaster set (1669) check: each list with 2, 4 and 6 pieces swapped in where they lose least
(wrist and feet, then waist and head, then chest and legs), random suffixes dropped and enchants kept,
at 180s and 10000 iterations. Every variant loses, and more with each piece:

| Spec | 2 pieces | 4 pieces | 6 pieces |
|---|---|---|---|
| Beast Mastery | -6.0 | -16.5 | -38.2 |
| Marksmanship | -5.6 | -18.4 | -40.0 |
| Survival | -3.6 | -17.2 | -31.5 |

The bonuses are resistances, mana regen, an auto attack mana proc and 40 attack power, and the pet
lines aren't modelled. No start set.

Process: a hunter can't start from no gear (no ranged weapon gives NaN DPS), so the hunters start
from the set in `phase0_specs.json`. Theramore Pauldrons are Alliance only and left it, and the set
now carries the shoulders and second ring Beast Mastery's search settled on. Unlisted picks tagged
include: Iceblade Hacker, Defias Enforcer's Mask, Beastslayer's Trident, Peacemaker, Bear Poking
Stick, Belgrom's Revenge and Dal'Rend's Tribal Guardian. No list was shipped with ALLOW_UNLISTED.
Container restarts killed searches many times. The full runs took about 3h on 4 cores. This container has
no `jj`, so a git shim stood in for the script's sim hash.

Public list: timesaver.gg's Forever hunter BiS (2026-11-20), the same for all three specs. Shared
picks: Dawn Armor, Devilsaur Gauntlets and Leggings, Windreaver Greaves and Beastslayer's Trident.
Its Core Marksman Rifle (Molten Core) isn't in our pools. Where its other items rank on our lists:

| Item | BM loss | MM loss | SV loss |
|---|---|---|---|
| Warpwood Binding | 1.5 | 1.3 | 0.7 |
| Shifting Cloak | 3.5 | 4.1 | 4.0 |
| Horn of Ve'han | 4.0 | 4.5 | not listed |
| Hyper Deluxe Sniper Rifle | 8.7 | 17.9 | not listed |
| Smolderweb's Eye | 10.8 | 11.0 | 13.7 |
| Blackcrow | 15.6 | not listed | -0.7 (tie) |
| Mask of the Unforgiven | not listed | not listed | 11.8 |

Slashclaw Bracers rank 2.5 behind on Survival and Ring of Demonic Guile 3.9. Savage Gladiator
Shoulders, The Postmaster's Seal and Frozen Heart of the Mountain aren't on any list. Our lists favour
Agility and haste (Ranger's Belt, Cutthroat's Signet, two Arcanums of Rapidity) and Black Dragonscale
Helm.

## Beast Mastery

1026.4 at the slot ranking, 1014.6 at 180s.

| Slot | Pick | Source | Enchant |
|---|---|---|---|
| Head | Black Dragonscale Helm | AH or Leatherworking 300 | Arcanum of Rapidity |
| Neck | Pendant of Celerity | Lord Valthalak, Dungeon Set 2 chain | Agility |
| Shoulder | Darkspear Epaulets | Drop, NPC 263570, the Barrens, Horde only | |
| Back | Cape of the Black Baron | Baron Rivendare, Stratholme | Agility |
| Chest | Dawn Armor | AH or Leatherworking | Greater Stats |
| Wrist | Bracers of the Eclipse | Prince Tortheldrin, Dire Maul | Superior Agility |
| Hands | Devilsaur Gauntlets | Leatherworking | Minor Haste |
| Waist | Ranger's Belt | Pix Xizzix, 100 Reals, hunter only | |
| Legs | Devilsaur Leggings | AH or Leatherworking 290 | Arcanum of Rapidity |
| Feet | Windreaver Greaves | Kirtonos the Herald, Scholomance | Greater Agility |
| Finger | Cutthroat's Signet | Pix Xizzix, 75 Reals | |
| Finger | Eagle's Head Signet | Quest: Blackmaw Reclamation | |
| Trinket | Chained Essence of Eranikus | Quest, Sunken Temple | |
| Trinket | Blackhand's Breadth | Quest, Blackrock Spire | |
| Two hand | Beastslayer's Trident | Quest reward | Iron Counterweight |
| Ranged | Assassin's Throwing Axe | Random world drop | |

Flags:

- Assassin's Throwing Axe: random world drop with no source Wowhead knows.
- Pendant of Celerity: the Valthalak summon is unconfirmed in Forever.
- Blackhand's Breadth and Beastslayer's Trident: quest rewards with no quest named in our data.
- Enchant Bracer - Superior Agility: no known source.
- Beastmaster pet lines (pet damage and crit) aren't modelled.

## Marksmanship

1115.7 at the slot ranking, 1102.2 at 180s. The same set as Beast Mastery but for Darkspear Pauldrons
(shoulder) and Ankle Slicers (ranged), with the same enchants.

Flags:

- Ankle Slicers: no known source.
- Pendant of Celerity, Blackhand's Breadth, Beastslayer's Trident and Superior Agility as above.

## Survival

946.4 at the slot ranking, 937.2 at 180s.

| Slot | Pick | Source | Enchant |
|---|---|---|---|
| Head | Black Dragonscale Helm | AH or Leatherworking 300 | Arcanum of Rapidity |
| Neck | Mark of Fordring | Quest: In Dreams, Plaguelands | Agility |
| Shoulder | Darkspear Epaulets | Drop, NPC 263570, the Barrens, Horde only | |
| Back | Cape of the Black Baron | Baron Rivendare, Stratholme | Agility |
| Chest | Dawn Armor | AH or Leatherworking | Greater Stats |
| Wrist | Bracers of the Eclipse | Prince Tortheldrin, Dire Maul | Superior Agility |
| Hands | Devilsaur Gauntlets | Leatherworking | Minor Haste |
| Waist | Ranger's Belt | Pix Xizzix, 100 Reals, hunter only | |
| Legs | Devilsaur Leggings | AH or Leatherworking 290 | Lesser Arcanum of Voracity (Agility) |
| Feet | Windreaver Greaves | Kirtonos the Herald, Scholomance | Greater Agility |
| Finger | Watcher's Signet of Ferocity | Watchers, Revered | |
| Finger | Cutthroat's Signet | Pix Xizzix, 75 Reals | |
| Trinket | Chained Essence of Eranikus | Quest, Sunken Temple | |
| Trinket | Darkmoon Card: Maelstrom | Darkmoon Faire deck | |
| Main hand | Iceblade Hacker | Ras Frostwhisper, Scholomance | Fiery Weapon |
| Off hand | Felstriker | Rend Blackhand, Blackrock Spire | Agility |
| Ranged | Satyr's Bow | Zevrim Thornhoof, Dire Maul | |

Flags:

- Watcher's Signet of Ferocity: no reputation source found yet.
- Darkmoon Card: Maelstrom: Forever card drops unconfirmed.
- Enchant Bracer - Superior Agility: no known source.
- Blackcrow ties Satyr's Bow (0.7 ahead in the ranking, within the noise).
