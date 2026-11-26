# tank_gear pass 19

Branch phase0-tankheal 3e9748641 (last pass: master 85691655a). First pass on the Phase 0 tank lists (standards 43,
passes/tank_gear.md rewritten 2026-11-25). Scratch helpers: phase0_check.py, a wear check (request = gearOut = list
picks) and tankbench at 3,000 iterations, seeds 1 and 2.

## Mechanical checks

`python3 tools/audit/gear/phase0_check.py`: every tank row has a list. 21 "open item above the pick" flags on tank rows;
all are re-tested in the spec notes (phase0_tankheal_*.md) as within bench noise, so none is a finding on its own
(tool finding -20261126-4).

## Requests wear the lists

All 12 tank requests wear their preset's gear set at `.raid.parties[0].players[0]` (ids and enchants equal), and every
list pick is in its gear set. No synthetic item left.

## List score against the bench (own column, six bosses)

| preset | race | list | bench s1 | bench s2 |
|---|---|---|---|---|
| warrior Balanced | Undead | 3299.4 | 3061.5 | 3055.3 |
| warrior Max threat | Undead | 4699.7 | 4454.7 | 4455.8 |
| warrior Max survival | Undead | 1218.7 | 1090.8 | 1067.2 |
| paladin Balanced | Undead | 2965.3 | 2958.7 | 2944.8 |
| paladin Max threat | Undead | 3912.6 | 3917.0 | 3909.8 |
| paladin Max survival | Undead | 800.2 | 793.2 | 766.0 |
| bear Balanced | Night Elf | 3686.0 | 3694.2 | 3697.8 |
| bear Max threat | Night Elf | 5606.7 | 5607.3 | 5597.8 |
| bear Max survival | Tauren | 1047.9 | 1081.8 | 1012.2 |
| Warden Balanced | Night Elf | 1962.8 | 2029.3 | 2019.7 |
| Warden Max threat | Night Elf | 3418.1 | 3410.3 | 3388.7 |
| Warden Max survival | Undead | -1336.9 | -1373.3 | -1428.3 |

The warrior is 5 to 11% off (finding -20261126-1); the others within about 1% except Warden Balanced (+3%).

## Faction rows

Other-faction sets differ only in faction-reward slots: warrior Max threat Alliance (hands, Raider Gauntlets to
Gauntlets of Heroism), paladin Max survival Alliance (shoulder, Darkspear to Theramore Mantle), bear Balanced Horde
(shoulder, Theramore to Darkspear Pauldrons), Warden Max survival Alliance (shoulder and feet). No pick of the preset's
faction is left in them. Presets with no faction pick have no other-faction set (checked: no Alliance/Horde/rep note on
their picks or picked enchants).

## Picks by hand

Doubtful picks are as the notes list (Watchers rings, Evergreen Shield, Lionheart Helm, Heroism set quest pieces,
Onyxia Scale Breastplate, Five Thunders, Steadfast Libram, Idol of the Ursine Twins). New here: Frozen Steel Vambraces
(-20261126-2) and stale "infeasible" tags on Warden picks (-20261126-3). Relics are modelled (sim/paladin/items.go
279247, sim/druid/items.go 279251, sim/shaman/items.go 22395). Unpriced crafted picks are in tank_fairness-20261126-2.

## New findings

- tank_gear-20261126-1 (medium): warrior lists sim 5 to 11% below their list score.
- tank_gear-20261126-2 (low): Frozen Steel Vambraces, called invalid by the server, is a pick.
- tank_gear-20261126-3 (low): stale price annotations: Warden picks tagged "infeasible", "g per DPS" notes on 15 tank and healer lists.
- tank_gear-20261126-4 (low): phase0_check's noise is in DPS and flags tank rows on noise.
