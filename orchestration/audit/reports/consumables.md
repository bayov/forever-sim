# Consumables pass 86 (master 95289e879)

No new findings.

## Moved
- sim/common/item_effects.go (+197): three new Forever trinket effects (Bolt's Charged Eyeball 280032, Chained Essence of Eranikus 10455, Everlook Pathcarver 274759).

## Verified (client 1.60.1.70938, q.py)
| Spell | Client | Sim |
|---|---|---|
| Grounded Electroshock 1315996 | 60 Nature | 60 |
| Grounded Electroshocks 1311505 | proc trigger of 1315996 | 15% / 3 sec ICD (per code comment) |
| Cloud of Chained Essence 1300240 | 140 per tick, periodic | 140 x 9 ticks / 2 sec |
| Prototype Pathcarver 1295270 | 607 Physical, 4 targets | 607 split over up to 4 |

- No preset, level60 gear set, mksettings file or dps_report request references the three items (git grep), so no DPS effect yet.
- scan.sh on master: banned, Revelation, non-synthetic trinkets and crit-elixir sections all empty.
- wago latest 1.60 build is 70938, already audited. Class branches are stale copies, not audited.

## Open items
The two older low findings stay open (Inferior Stamina naming, feral mksettings Sapper).
