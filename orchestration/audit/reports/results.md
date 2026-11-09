# Results audit, pass 81 (master 66a243cf8)

Changed since 9ba00c2dc: sim commits 0ce9dbf40 (Windfury Totem white main hand only) and 3b0983bc6 (Orc shaman base stats levels 2-43, no effect at 60), plus db regeneration 8051c4322. Shared run: 68 builds, 20k, seed 1.

| Build | published/previous | now | move |
|---|---|---|---|
| warrior arms | 1088.0 | 888.0 | -200 (-18.4%) |
| warrior fury_2h | 1151.0 | 961.3 | -190 |
| warrior fury | 1136.0 | 1036.6 | -99 |
| hunter survival | 1041.2 | 945.4 | -96 |
| paladin level_60 | 1092.1 (run 1103.2) | 1009.7 | -93 vs run |
| rogue combat-swords | 963.5 | 917.1 | -46 |
| shaman enhancement | 1146.7 | 1146.7 | none (own totem rules since 9ba00c2dc) |
| mage arcane, elemental, warlock, feral | | | unchanged by this commit (existing meta lag) |

Attribution: paired runs on binaries 9ba00c2dc and 66a243cf8 reproduce the whole move on warrior, hunter, paladin and rogue; enhancement and mage are identical. Cause is 0ce9dbf40 only. Finding results-20261109-1 (high).
Not run this pass: regen drift check and unread_fields (needs proto generation); the db and base-stat changes do not touch requests. Existing findings 20261104-1/-2 and 20261105-1 still open, unchanged.
