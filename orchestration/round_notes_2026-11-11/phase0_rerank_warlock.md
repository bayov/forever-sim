# Phase 0 re-rank: warlock (2026-11-25)

`RANK_ONLY=1 SLOTS=mainhand,offhand,ranged,trinket` on master 34fe148 for affliction, demonology and destruction
(warlocks wear a wand, no relic). The SLOTS run re-ranks the set's enchants too.

## Slot changes

| List | Change | List dps (use_dur 90) |
|---|---|---|
| Affliction | Main hand enchant Mighty Intellect to Revelation (Blade of the New Moon), +3.0 | 976.65 -> 977.14 |
| Demonology | none | 887.95 -> 889.44 |
| Destruction | Trinket 2 Weakness Analyzer to Eye of the Beast, +1.5 (confirmed at 20000) | 873.82 -> 875.87 |
| Affliction Alliance copy (Human) | Finger 2 Songstone of Ironforge to Tracker's Signet of Brilliance (0.2 ahead, noise), enchant as Affliction | 975.41 -> 970.74 |

Close calls not taken: Destruction main hand Darkspear Insurgent's Spellblade -0.5, off hand Scepter of Interminable
Focus -0.9. The Alliance list's drop comes from the Human run itself (956.2 at 180s against 962.6 Undead).

The Alliance copy now wears two Watchers signets (Watcher's and Tracker's of Brilliance), as Fury DW already does. If the
Watchers rings turn out to be one upgrade chain, Songstone of Ironforge is the pick at -0.2.

## Requests

The 13 requests and the four Affliction race files (Human, Gnome, base and Shadow Bolt) wear the new sets. Every
request's equipment equalled its old set exactly, so the item lines were swapped in place (same result as mksettings
with the same arguments). `race_gear.py --write` on every warlock build: no other row changes.

verify.py, 180s, 20000 iterations, seed 1, same binary for before and after:

| Build | Before | After | List (180s) |
|---|---|---|---|
| Affliction | 959.6 | 962.5 | 962.6 |
| Affliction, Human | | 955.9 | 956.2 |
| Affliction curse duty | | 953.4 | |
| Affliction recklessness duty | | 948.2 | |
| Shadow Bolt Affliction | 793.4 | 820.5 | |
| Demonology | 877.0 | 877.0 (no change) | 877.3 |
| Destruction | 868.2 | 868.0 | 868.7 |
| Fire Destruction | 825.5 | 828.9 | |

Every request sims to its list within noise. Revelation is worth +27 to Shadow Bolt Affliction, which casts far more
Shadow Bolts. Destruction's trinket swap is flat at 180s: Weakness Analyzer's use effect is what the list's use_dur 90
score values, Eye of the Beast wins there. The metas' published numbers are the parent's to refresh (pipeline run).

## Set trap

Sets a warlock can wear in Phase 0: Dreadmist Raiment (set 183, 2pc resistances, 3pc +18 spell power, 6pc +8 MP5) and
Deathmist Raiment (Forever, 2pc resistances, 3pc +8 MP5, 6pc +23 spell power). Dreadmist had not been checked before.
Each piece alone against the set, then the best 2, 3, 4, 6 and 8 pieces, no random suffix, the slot's enchant kept,
10000 iterations at 180s on the request:

| Spec | List set | Best Dreadmist | Best Deathmist |
|---|---|---|---|
| Affliction | 962.6 | 3pc 933.4 (-29.1) | 2pc 944.1 (-18.5), 6pc 911.9 (-50.7) |
| Demonology (wears Deathmist Wraps) | 877.3 | 3pc 861.4 (-15.9) | 2pc 876.1 (-1.3), 6pc 856.8 (-20.5) |
| Destruction (wears Deathmist Wraps) | 868.7 | 3pc 856.7 (-12.0) | 2pc 867.8 (-0.9), 6pc 850.3 (-18.3) |

No set wins at any count. The lists stand.

## Open items

None for warlock.
