# Phase 0 healer: priest_holy (2026-11-26)

Tank and raid healer roles. Sim at origin/phase0-tankheal 72079b0 (merged after the list was done). Human, talents
and APL unchanged.

## Method

Restarts killed the long runs, so I skipped the 2000-iteration full search. Instead:

- The list was seeded by hand from the Disc picks and the dry-run pool.
- Pass 1 ranked every slot. Pass 2 started each slot from its rank 1 and ranked it again with RANK_ONLY, one slot per
  chunk. Each finished slot was committed.
- All picks are within noise of their slot's top. The largest gap is +0.3 (Amulet of the Dawn, neck), not confirmed at
  5000 iterations.
- Pass 2 swaps:
  - Belt of the Archmage over Knowledge of the Timbermaw.
  - Adaptive Combat Assistant over Draconic Infused Emblem.
  - Gift of the Elven Magi + Brightly Glowing Stone over Staff of Metanoia: +0.5, confirmed.
  - Hide of the Wild over Drape of Recovery: +0.2, confirmed.
- Pass 1 put the old trinkets (Chained Essence of Eranikus, Weakness Analyzer) last, 0.5-0.8 under the picks.

rotopt list score 133.59. phase0_check.py: no priest_holy findings. Line 2214g per point.

## Picks

| Slot | Pick | Source | Flags |
|---|---|---|---|
| Head | Virtuous Crown (1509) | Dungeon Set 2 copy, set 1675 | doubtful. Open fallbacks: Virtuous Cowl (keeps 6pc) or Devout Crown, both -0.17 |
| Neck | Emberfury Talisman (Healing Power) | Dungeon drop | |
| Shoulder | Virtuous Mantle | Set 1675 | |
| Back | Hide of the Wild | Leatherworking craft | |
| Chest | Devout Robe (Mystic Rugged Armor Kit) | Drakkisath, UBRS | |
| Wrist | Virtuous Bracers (2566) | Set 1675 | |
| Hands | Virtuous Mitts (Mystic Heavy Armor Kit) | Set 1675 | |
| Waist | Belt of the Archmage | Tailoring 295, pattern on AH | |
| Legs | Virtuous Skirt (Mystic Rugged Armor Kit) | Set 1675 | Magister's Leggings +0.16, noise |
| Feet | Virtuous Sandals (Mystic Rugged Armor Kit) | Set 1675 | |
| Finger | Blessed Band of Light | Pix Xizzix, Reals | |
| Finger | Fordring's Seal | Quest: In Dreams | |
| Trinket | Adaptive Combat Assistant | Pix Xizzix, 125 Reals | |
| Trinket | Skarr's Broken Gauntlet | Skarr the Unbreakable, Feralas | |
| Main hand | Gift of the Elven Magi (2568) | Balnazzar, Stratholme | |
| Off hand | Brightly Glowing Stone (Wisdom) | King Gordok, DM North | |
| Ranged | Lunar Wand of the Channeler | World drop | Torch of Light +0.1 to +0.2, not confirmed |

Class check: none of the 82 items on the start row or the final list exclude priests.

New Phase 0 tags (all top 5s, none picked unless listed above):

- Dungeon drops: Necropile Robe, Leggings and Mantle, Devout Sandals, Crown and Mantle, Dreadmist Leggings, Boots of
  the Full Moon, Faith Healer's Boots, The Hammer of Grace, Magmus Stone, Masterwork Crimson Hammer, Cape of the Fire
  Salamander, Staff of Metanoia, Gravetoll Gavel, Soulwoven Drape, Mantle of the Scarlet Crusade, Gordok's Handwraps,
  Arena Wristguards.
- Doubtful: Mooncloth Shoulders, Venomspew Cord, Ooze Rummagers, Royal Seal of Alexis, Faintly Glowing Orb.
- Out: Band of Icy Depths (Hydraxis chain needs Molten Core).
- Caretaker's Cape is Alliance-only.
- Enchants: Enchant Cloak - Resistance, Medium Armor Kit.

## Set trap

- The Virtuous 6pc holds.
- 5pc single swaps from the ranking are all within noise: Devout Crown -0.17, Magister's Leggings +0.16.
- A 4pc mix (Devout Crown + Magister's Leggings) scores 131.2 vs 131.7 for the 6pc in healbench.

## Rotation and mana

Unchanged. Mana is ample: on Ragnaros in the raid role the set ends with 7034 mana, and running out is only projected
at 1057s.

## Faction row

Skipped. No pick is a faction reward.

## Hand check

No public Phase 0 holy priest list compared. Wowhead's bot check blocks it and I did not bypass it.

## Score (step 9)

`healbench -carry -role tank|raid -iters 1000`, damage carried:

| Boss | Tank P0 | Tank syn | Raid P0 | Raid syn |
|---|---|---|---|---|
| Magmadar | 201.1% | 214.4% | 200.0% | 207.4% |
| Golemagg | 186.1% | 200.0% | 251.6% | 246.2% |
| Ragnaros | 56.7% | 58.9% | 133.0% | 134.6% |
| Onyxia | 131.6% | 138.5% | 97.1% | 96.3% |
| Vaelastrasz | 84.1% | 91.4% | 125.3% | 122.4% |
| Broodlord | 88.1% | 92.2% | 90.4% | 92.4% |
| Chromaggus | 87.8% | 91.4% | 110.8% | 106.7% |
| **Mean** | **119.4%** | **126.7%** | **144.0%** | **143.7%** |

Score 131.7 vs synthetic 135.2: -3.5 points (-2.6%). The whole gap is in the tank role. Raid is even.

## For the user

- Virtuous Crown availability is doubtful. The open fallbacks cost about 0.2.
- The doubtful Mooncloth Shoulders, Royal Seal of Alexis and Faintly Glowing Orb are top-5 only.

## Cleanup (2026-11-27): list vs bench, and Chromaggus

**The list sat 1.4% over the bench** (133.59 against 131.70). Not the scoring: on the same set rotopt's reference
carries are the bench's to the third decimal on all 14 slots. The list was ranked at 652488b, before the merge that
gave the bench's Frost mage fillers (tools/dps_report/req/mage__frost.json) a new ring and wand (68fe2ee9, the DPS
re-rank). The raid role fails when anyone dies, fillers too, so every raid carry fell: the Holy raid mean 1.479 at
652488b, 1.4943 against 1.440 on three bosses with the old filler on today's code, and 1.440 today. The tank role was
the same to the third decimal. The other healer lists were ranked after that merge and sit 0.1 to 0.4% from the
bench. Fix: `RANK_ONLY=1 SLOTS=ranged` re-scored the list on today's bench (picks unchanged, Lunar Wand of the
Channeler still within noise of the top): 133.59 -> 132.29 with the paced rotation below. phase0_check.py now
flags a healer list more than 1% off the Healing pack.

**Chromaggus** (the tank preset ends with 412 mana): the fight, not the gear or the bench. See docs/healing_priest.md,
Mana on Chromaggus. The Flash Heal first lines now also need mana % at least the fight's remaining %: tank mean
1.1936 -> 1.1996 (seed 202, 2000 iterations: 1.1951 -> 1.1990), Chromaggus 0.878 -> 0.912. Going dry there is real
at Phase 0 gear: one priest can't keep the main tank up for 115 s at the logs' damage.
