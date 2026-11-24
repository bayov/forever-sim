# Phase 0 step 4: shaman_ret (Enhancement, Elemental, Retribution)

Brief `briefs/phase0_step4_1124.md`. Every DPS shaman and paladin build of the report now wears its spec's Phase 0
set. No `_hit±N` set is left in a request. DPS is `verify.py` at 180s, 20000 iterations, seed 1. "Synthetic" is the
meta's earlier published number. Each Phase 0 number matches `-eval` of its set within noise.

## Enhancement (set `level60_phase0`, Orc)

| Build | Phase 0 | Synthetic | Talents |
|---|---|---|---|
| Level 60 | 990.7 | 1080.0 | unchanged |
| Imp GW | 982.4 | 1066.3 | now the list's `053303102-054030110051212141-03`, +22 |
| Water Shield | 937.5 | 1025.7 | unchanged |
| Mana Tide | 894.2 | 977.2 | now `03-055030310051012051-053141001001`, +2.3 |

- The set has 5% melee hit. With Tidal Focus that puts Level 60 at the 8% cap, Imp GW at 6% before its change (Tidal
  Focus 3 brings it to 8%), Water Shield and Mana Tide at 10%.
- Mana Tide: Elemental Weapons 3 out for Improved Stormstrike 1, Totemic Focus 4 out for Healing Focus 3. It wears
  the Phase 0 set with Burning Totem (the set's own totem pick is Totem of the Storm, the two tie within 2).
- Rotation knobs held (Fire Nova 16, Mana Tide 20 mana a second).
- Level 60's talents beat the list's (the user's) by about 9 on the Phase 0 set. A RANK_ONLY of the list on
  the Level 60 talents (quick-check iterations: 300, confirm 2000, enchants 1000) keeps all 16 picks and enchants,
  list dps 1002.5 against 993.2, so the list stays on the user's talents and every build wears its set.
- Searches: rotopt `-talent-search` with the build's pins, mean of 120/180/300s, confirmed at 20000 iterations.

## Elemental (set `level60_phase0_ele`, Windshaper)

| Build | Phase 0 | Synthetic | Talents |
|---|---|---|---|
| Level 60 | 859.2 | 976.9 | unchanged |
| Level 60 + Mana Tide | 844.4 | 959.6 | unchanged |

- 12% spell hit with Tidal Focus 5/5, under the 16% cap. No hit talent change wins. Downrank knobs (30 and 20 mana
  a second) held.
- Neck re-rank (the new checker's Nacreous Shell Necklace flag): Orb of the Darkmoon stays. Nacreous ranks 1.12
  above it but costs 217g a DPS, over the gold line, so the list marks it infeasible. The checker saw it because
  5e556ff1e ran `phase0_prices.py --write`, which prices from the Enhancement list only and wiped the Elemental-only
  price of 22403 from `phase0_content.json`. Restored. The script will wipe it again on its next run.

## Retribution (set `level60_phase0_ret`, Undead)

RANK_ONLY rerun for the stale SLOTS numbers: no pick or enchant moved, list dps 975.1 to 972.9, gear set unchanged.

| Build | Phase 0 | Synthetic |
|---|---|---|
| Twist | 956.9 | 1023.0 |
| Twist, Crusader duty | 943.8 | 1016.3 |
| Twist, Wisdom duty | 945.2 | 1009.9 |
| Sub Holy | 928.1 | 1011.6 |
| Sub Holy, Crusader duty | 914.7 | 1006.0 |
| Sub Holy, Wisdom duty | 915.3 | 999.7 |
| Sub Holy, vs Undead | 1055.7 | 1151.0 |
| Shock | 922.4 | 1009.9 |
| Shock, Crusader duty | 914.5 | 1008.8 |
| Shock, Wisdom duty | 914.8 | 999.6 |

- The set has 5% melee and spell hit. Twist with Precision 3 sits exactly on the 8% cap, so Precision is no longer
  wasted: Precision -1 for Reverence now costs 13.0 (6.0 on synthetic). Sub Holy and Shock sit at 5%. Sub Holy's
  Divine Precision -2 for Divine Intellect loses 10.6. Talent searches on the Phase 0 set keep every build's talents.
- Twist stays the best build, about 29 ahead of sub Holy (14 on synthetic).
- Crusader duties: Libram of Fervor (23203) has no known Forever source, so they keep Libram of Law. Fervor would be
  worth 7.5 (twist), 7.4 (sub Holy), 9.7 (Shock).
- vs Undead: own set `level60_phase0_ret_undead` with Tenets of the Silver Hand (249397, open: formula from the
  Forever enchant vendors, 45 of currency 3402), +7.9 over Libram of Law, and Rune of the Dawn (Stratholme quest)
  for Blackhand's Breadth, +11.9 by verify. Seal of the Dawn ranks 4.25 behind the picks.

## Race rows

- Shaman (Dwarf, Orc, Tauren, Troll, Windshaper): no Phase 0 pick is faction-tagged, so Dwarf needs no gear swap. The
  bonus-stats trade loses or ties on every Enhancement build and Elemental Dwarf is under the cap: the three Dwarf
  files are removed. Both Elemental Troll (Berserking) overrides now lose about 3, and about 5 on the old synthetic
  set too, so the cause is an older sim or rotation change, not the gear. Removed.
- Paladin (Human, Dwarf, Undead): no ret pick is faction-tagged. Forever racials carry no weapon skill, so Destiny
  stands for every race. Twist Dwarf (9% hit) keeps a bonus-stats file, +2.58. Sub Holy and Shock Dwarf are under
  the cap, files removed. Alliance-only items that might beat a pick were not checked.
- `race_text.json` lost the entries of the removed files.

## Weapon imbues

Enhancement and Ret use Shadow Oil (Enhancement plus Windfury Weapon), Elemental Wizard Oil. Oils apply to every
weapon type, so the search's one imbue for all candidates judges them fairly here.
