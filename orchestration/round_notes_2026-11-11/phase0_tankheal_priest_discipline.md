# Phase 0 tank healer: priest_discipline (2026-11-25)

Tank healer role only (healRoles tank, no raid request). Sim at origin/phase0-tankheal a1904ac (all six sim batches and
the Premier drop merged).

## Score (step 9)

`healbench -carry -role tank -iters 1000` on the seven bosses, damage carried:

| Boss | Phase 0 | Synthetic | Mana left (P0 / syn) |
|---|---|---|---|
| Magmadar | 202.1% | 215.9% | 6532 / 5896 |
| Golemagg | 194.1% | 203.3% | 6643 / 6014 |
| Onyxia | 135.5% | 135.9% | 6693 / 5782 |
| Broodlord | 92.6% | 93.8% | 6715 / 5862 |
| Vaelastrasz | 89.1% | 93.1% | 7079 / 6145 |
| Chromaggus | 86.9% | 82.8% | 374 / 327 |
| Ragnaros | 60.1% | 61.6% | 5746 / 4393 |
| **Mean** | **122.9%** | **126.6%** | |

Phase 0 is 3.7 points (2.9%) under synthetic. The search's own score (rotopt) on the final set: 122.7 (0.2).

## Picks

| Slot | Pick | Source | Flags |
|---|---|---|---|
| Head | Virtuous Crown | Forever Dungeon Set 2 copy, end of the Valthalak chain | doubtful |
| Neck | Animated Chain Necklace | Ramstein, Stratholme | |
| Shoulder | Virtuous Mantle | Dungeon Set 2 copy (Anthion chain) | |
| Back | Hide of the Wild | Leatherworking craft | source note thin ("open" by the 11-12 check) |
| Chest | Devout Robe | Drakkisath, UBRS | |
| Wrist | Virtuous Bracers | Dungeon Set 2 copy | |
| Hands | Virtuous Mitts | Dungeon Set 2 copy | |
| Waist | Belt of the Archmage | Tailoring 295, pattern ~131g | |
| Legs | Virtuous Skirt | Dungeon Set 2 copy | |
| Feet | Virtuous Sandals | Dungeon Set 2 copy | |
| Finger | Band of Mending | Sothos and Jarien summon | summon unconfirmed, included by standards 42 |
| Finger | Fordring's Seal | Quest: In Dreams | |
| Trinket | Second Wind | Argelmach, BRD | |
| Trinket | Serenity Field | Pix Xizzix, 125 Reals | |
| Main hand | Glowing Brightwood Staff | random world drop, epic BoE | AH price not checked |
| Ranged | Ritssyn's Wand of Bad Mojo | Rivendare, Stratholme | |

Enchants: Lesser Arcanum of Voracity (Int), Necklace Healing Power, Mystic Rugged Armor Kit (chest, legs, feet), Mystic
Heavy Armor Kit (hands), Bracer Healing Power (Argent Dawn Revered formula), 2H Spellblasting. Enchant Weapon Healing
Power is out (Molten Core formula).

No pick has an unmodelled effect or hidden stats. The Virtuous healer pieces (226945 to 226952) share set 1675 with the
shadow pieces the table starts from.

## How the list was made

- The first full search (1000 iterations, from the table's start) made two swaps and stopped. A swap must beat 2 standard
  errors, and at 1000 iterations that is about 2 points of carry, while the healer score is flat (most slot gaps 0.3 to
  1.5). The 5000 iteration ranking showed better items in most slots.
- Second search: from a start of each slot's ranked best (open items over doubtful ones within noise, the Virtuous 6pc
  kept) at 2000 iterations. It kept that start and took Hide of the Wild. A full RANK_ONLY on the merged sim batches,
  then RANK_ONLY chest,waist after the Premier drop. Every pick is within 0.3 of its slot's top (noise).
- Worth telling the parent: on a healer list the search's acceptance threshold leaves most of the work to the start.
- Tags: 38 items and 2 enchants added to phase0_content.json and phase0_enchants.json. Out: Benediction, Enchant Weapon
  Healing Power, the three Conviction pieces. Doubtful: Argent Crusader, Lesser Eternal Wand, Torch of Light, Premier
  Satin Cowl, Lorespinner, Wisdom of Ursol, Mooncloth Circlet, Earthenweave Cuffs. Unsure: Mooncloth Boots recipe source,
  Rosewine Circle (Urok summon), Lost Researcher's Hood quest zone (Alliance).

## Set trap

Vestments of the Virtuous (2/3/4/6). Both search starts held 6 pieces. Two more searches at 2000 iterations:

- A, 6pc with no doubtful head (Virtuous Belt in the waist, Lost Researcher's Hood on the head): started at 121.8, put
  Virtuous Crown back (+0.8), ended at the list's set (Virtuous Belt in place of Belt of the Archmage, a tie). So the
  doubtful Crown is worth about 0.8 points, and the best open 6pc is about 0.8 under the list.
- B, 4pc (Desert Bloom Gloves, Swarmtender's Footpads): no swaps, 122.5 (0.1) against the list's 122.7 (0.2). A tie, and
  Desert Bloom is doubtful, so the list keeps 6pc.

## Talents and rotation

Unchanged. Mana is ample on six bosses. On Chromaggus the healer ends near empty in both sets (374 and 327 mana left), so
it is the fight, not the gear.

## Faction row

None. The preset is Human. No pick or picked enchant is Alliance only (a Troll dry run drops none of them).

## Public list

Not compared: Icy Veins returned 403 and the other guide 404. Wowhead was not tried.
