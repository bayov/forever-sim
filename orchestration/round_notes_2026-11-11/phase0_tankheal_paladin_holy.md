# Phase 0 gear: paladin_holy (2026-11-26)

Branch `phase0-tankheal-paladin_holy`. Brief: `briefs/phase0_tankheal_spec_1125.md`. Merged with
`origin/phase0-tankheal`, including every sim batch and the Premier drop.

## Score

Here the score is 100 × the mean carry of the healbench bench, taken over 7 bosses × 2 roles at 1000 iterations.

| | Phase 0 | Synthetic (master) |
|---|---|---|
| Score | **123.4** | **127.1** |
| Gap | −3.7 (−2.9%) | |

The rotopt ranking gives the same set 123.4(0.1) at 5000 iterations. scoreRef is 126.9.

| Boss | Tank P0 | Tank syn | Raid P0 | Raid syn |
|---|---|---|---|---|
| Magmadar | 1.788 | 1.876 | 1.757 | 1.872 |
| Golemagg | 1.676 | 1.800 | 2.535 | 2.580 |
| Onyxia | 1.201 | 1.253 | 0.964 | 0.937 |
| Broodlord | 0.869 | 0.872 | 0.906 | 0.918 |
| Vaelastrasz | 0.785 | 0.786 | 1.000 | 1.042 |
| Chromaggus | 0.834 | 0.898 | 1.128 | 1.113 |
| Ragnaros | 0.548 | 0.569 | 1.288 | 1.281 |

On every boss the Phase 0 set ends with about 1500 less mana. Tank Chromaggus is the tightest at 1123 left
(2512 on synthetic), but no fight runs out of mana.

## Picks

| Slot | Item | Rank | Source | Flag |
|---|---|---|---|---|
| Head | Sanctified Leather Helm (22689), Lesser Arcanum of Resilience | 1 | Argent Dawn Armaments quest | doubtful |
| Neck | Tempest Talisman (18317) | 2 | open, from an earlier check | |
| Shoulder | Eternity Pauldrons (279255) | 1 | Leatherworking 300, the pattern has no known source | doubtful |
| Back | Cloak of Fire (14134), Greater Resistance | 1 | Tailoring 235 | |
| Chest | Crownroar Chestplate (280027), Greater Stats | 1 | Vitrix (Forever rare), seen in game Nov 8 | |
| Wrist | Bracers of Prosperity (18525), Healing Power | 1 | King Gordok, Dire Maul | |
| Hands | Gordok's Handwraps (18369), Heavy Armor Kit | 1 | Dire Maul quest | quest not identified in our data |
| Waist | Belt of the Archmage (18405) | 2 | Tailoring 295, the pattern is on the AH | |
| Legs | Devout Skirt (16694), Mystic Rugged Armor Kit | 3 | Baron Rivendare | |
| Feet | Boots of the Full Moon (18507), Mystic Rugged Armor Kit | 1 | Captain Kromcrush, Dire Maul | |
| Finger | Band of Rumination (18103) | 1 | Rend, UBRS | |
| Finger | Amalgam's Band (22326) | 2 | Kormok, Dungeon Set 2 summon | summon unconfirmed |
| Trinket | Talisman of Ascendance (22678) | 1 | Argent Dawn Insignia quest | doubtful |
| Trinket | Burst of Knowledge (11832) | 2 | Ambassador Flamelash, BRD | |
| Main hand | Demon's Claw of Healing (15240), Mighty Intellect | 5 | world drop | |
| Off hand | Exalted Shield of the Hierophant (14982), Law of Nature | 1 | world drop | |
| Libram | Libram of Economy (272436) | 1 | vendor Pix Xizzix, STV | |

Ranks 2 to 5 in a slot sit within 0.3 of the pick, inside 2 standard errors, so the ranking kept the pick.

- **Hidden stats:** no pick has hidden stats or an unmodelled effect.
- **Doubtful picks:** Sanctified Leather Helm, Eternity Pauldrons and Talisman of Ascendance are doubtful but still pickable.

## How the list was found

1. **Start set fault.** The spec row's start was the shaman's Elemental set, with Burning Totem in ranged. The
   first run panicked on the totem, so I ran with "req" in ranged, and the parent fixed the row. Five Thunders is
   shaman only (classAllowlist 7), but the start set isn't class filtered. The full search therefore kept six
   Five Thunders pieces: a swap must beat 2 standard errors (about 2 points), and it made only 2 swaps
   (Premier head, Postmaster's Seal).
2. **Ranked tops.** As in the Discipline Priest fix, I reran from each slot's rank 1, leaving out Premier
   items. This was RANK_ONLY over all slots at 2000 iterations on the merged code. That set held, with no
   further swaps.
3. **Unlisted items.** I tagged 28 items and 3 enchants in `phase0_content.json` and `phase0_enchants.json`.
   None is out, and only Sentinel's Lamellar Legguards (not a pick) is infeasible.

## Set trap

The holy pieces of the Forever Soulforge Armor (set 1673) are open quest rewards:

| Bonus | Effect |
|---|---|
| 3 | mana regen |
| 4 | Crusader's Wrath |
| 6 | +23 spell power |

I scored each start with RANK_ONLY SLOTS=ranged at 5000 iterations:

| Start | Score |
|---|---|
| List as is | 123.4(0.1) |
| 3pc (Leggings, Cord, Bindings) | 123.5(0.1) |
| 6pc (+ Epaulets, Fists, Treads) | 123.4(0.2) |

This is a tie, so the list stays without the set. A 3pc or 6pc Soulforge set is an equal choice for a player
who has it. Lightforge (Dungeon Set 1) loses to its own Soulforge upgrade. Five Thunders is shaman only.

## Talents and rotation

Unchanged. Mana ends at 1100 or more on every boss, and the carry here isn't mana bound.

## Faction row

No pick or enchant is a faction reward, so I skipped it (the Undead set is the same).

## For the user

- **Doubtful picks:** head, shoulder and trinket 1 are doubtful (Argent Dawn Armaments and Insignia quests,
  and the Eternity Pauldrons pattern). The second ring's Kormok summon is unconfirmed.
- **Gordok's Handwraps:** the quest is not identified in our data.
- **The gap:** the 3.7 point gap to synthetic is largest on Magmadar and Golemagg tank, and the Phase 0 set
  ends every fight with about 1500 less mana.
