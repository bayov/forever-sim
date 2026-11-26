# healer_gear pass 17

Branch phase0-tankheal 3e9748641 (last: master 85691655a). First pass on the Phase 0 healer lists (standards 43).
druid_restoration is still being searched: no list yet, expected, not flagged.

## Mechanical checks

phase0_check.py: priest_holy (Human, 133.59), priest_discipline (Human, 122.23), paladin_holy (Human, 123.91),
shaman_restoration (Skyborne Windshaper, 127.63); no open item above a pick on a healer row.

## Requests wear the lists

priest_holy and priest_holy_raid, priest_discipline, paladin_holy and paladin_holy_raid, shaman_restoration and
shaman_restoration_raid all wear their gear set at party 0 player 0 (ids and enchants equal); every pick is in its set.

## List against the bench (100 x mean carry, roles alike)

| spec | list | bench |
|---|---|---|
| priest_holy | 133.59 | 131.7 |
| priest_discipline | 122.23 | 122.8 |
| paladin_holy | 123.91 | 122.2 |
| shaman_restoration | 127.63 | 127.5 |

Holy priest and Holy paladin are 1.4% high (-20261126-1).

## Faction

shaman_restoration (Horde) wears Battle Healer's Cloak (Warsong) and Raider Handwraps (Darkspear); its Alliance row
(Dwarf) differs only in back (Hide of the Wild) and hands (Desert Bloom Gloves, a quest) and holds no
Horde pick. The Human lists have no faction pick.

## Picks by hand

Doubtful picks as the notes say: Virtuous Crown, Sanctified Leather Helm, Eternity Pauldrons, Talisman of
Ascendance, Band of Piety (Argent Dawn armaments / Insignia quests, standards 42), Five Thunders Crown and Chain,
Watcher's Signet, Evergreen Shield, Totem of Urgency. Relics modelled: Libram of Economy (sim/paladin/healing.go),
Totem of Urgency (sim/shaman/items.go). Trinket effects modelled (Burst of Knowledge, Smokey's Lighter, Second Wind,
Adaptive Combat Assistant in sim/common; Serenity Field in forever_item_procs.go). priest_holy's Skarr's Broken
Gauntlet is a plain 15 Fire Resistance trinket (the raid fire damage values it). Stale "g per DPS" notes on the healer
lists: tank_gear-20261126-3. Unpriced picks (Eternity Pauldrons, Totem of Urgency, Evergreen Shield, the Alliance
row's Hide of the Wild) and Emberfury Talisman's missing source note: tank_fairness-20261126-2.

## New findings

- healer_gear-20261126-1 (low): Holy priest and Holy paladin lists 1.4% above the bench.
