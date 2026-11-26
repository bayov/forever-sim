# Phase 0 re-rank: mage (2026-11-25)

`RANK_ONLY=1 SLOTS=mainhand,offhand,ranged,trinket` on master 34fe148 for the four mage lists (mages have no relic).
List dps is the set's score (180s with uses on 90s), request dps is verify.py at 180s ± 20s, 20000 iterations, seed 1.

## Per build

| Build | Slot changes | List dps before / after | Request dps before / after |
|---|---|---|---|
| Arcane | Staff enchant Spellblasting -> Revelation (items unchanged) | 841.13 / 847.17 | 830.65 / 834.05 |
| Fire | none | 924.57 / 931.04 | 917.89 / 917.89 (list's 180s measure 917.2) |
| Frost | none | 736.06 / 738.76 | 725.42 / 725.42 (725.0) |
| Frostfire | none | 869.54 / 875.37 | 862.57 / 862.57 (862.7) |

- No weapon, off hand, wand or trinket pick moved. Chained Essence of Eranikus and Eye of the Beast stay the trinkets
  everywhere. The lists rise 2.7 to 6.5 because the whole set scores higher on the new sim.
- **Arcane:** Enchant Weapon - Revelation now ranks 4.1 over Enchant 2H Weapon - Spellblasting (it was below it before
  the procs batch). Gear set, request (mksettings with the meta's arguments gives the old request plus this enchant
  only), meta, the Horde copy and the Orc, Undead and Troll rows follow. The request sims 834.05 (0.29) against the
  list's plain 180s measure 833.7 (0.4).
- **Horde copy:** RACE=RaceTroll SLOTS=finger from the new list keeps Channeler's Ring and Eye of Orgrimmar (list
  840.07 at Troll). The set differs from Alliance's in the finger only, as before.
- **High Order rotation (Arcane)** on the new enchant: +4.3 (825.74 against 821.44 with the preset's rotation, same
  seed), was +8.1. Kept.
- **Ties, not swaps:** Fire takes Revelation -0.46 against Spellblasting (enchant ranking, within noise). Frostfire's
  Dragon Finger of Fiery Wrath reads -1.0 over Ritssyn's Wand on the ranking, but the 20000-iteration confirm puts it
  +0.2 (0.3), not confirmed. Arcane's Wit's Edge + Scepter of Interminable Focus reads -0.22 over Staff of Sorrow
  (Wit's Edge has no known source).
- **Unlisted:** Cloud Stone (17737, Arcane off hand list, not a pick) wasn't in `phase0_content.json`. It drops from
  Landslide in Maraudon, added.

## Set trap

None. All four sets wear Sorcerer's Regalia (1671) at 6 pieces, its last bonus (6pc +23 spell power, no 8pc). The
other sets a mage can wear in Phase 0 all lose, `rotopt -eval` on each request, 180s, 10000 iterations, seed 1,
swapping the cheapest slots:

| Spec | Current | Magister's 3 / 5 / 6 / 8 | Dreadmist 3 / 6 / 8 | Necropile 3 / 5 | Postmaster 4 / 5 |
|---|---|---|---|---|---|
| Arcane | 830.2 | 789.3 / 774.4 / 761.6 / 747.5 | 784.6 / 747.0 / 725.1 | 784.0 / 773.9 | 787.9 / 779.1 |
| Fire | 917.2 | 860.7 / 837.6 / 831.1 / 811.0 | 847.1 / 818.9 / 794.1 | 866.5 / 844.6 | 852.3 / 840.8 |
| Frost | 725.0 | 674.2 / 664.6 / 655.9 / 642.3 | 671.8 / 657.4 / 639.7 | 686.9 / 681.9 | 675.8 / 670.2 |
| Frostfire | 862.7 | 804.0 / 778.6 / 772.9 / 752.0 | 794.9 / 753.7 / 729.5 | 806.1 / 783.1 | 792.8 / 778.4 |

## Open items

None for mage.

## Other

- A fresh container has no Go protos: `apt-get install protobuf-compiler`, `go install
  google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6`, `make sim/core/proto/api.pb.go` before bis_phase0.sh.
