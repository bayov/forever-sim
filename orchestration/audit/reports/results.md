# Results audit, pass 83, master 707ae92cc (client 1.60.1.70888)

Shared run: 68 requests, 20k, seed 1, CLI 707ae92cc50b. Previous audited master 27d0d59d6.

## What was checked
- Run vs metas (`vs_published.txt`): 45 of 68 off by more than 0.5. Run vs builds.json: 43 of 68 off by more than 1.5.
- Drift (mksettings vs committed requests): only the open feral Goblin Sapper difference (results-20261105-1); no new drift.
- Unread request fields: none printed.
- Paired 20k seed-1 sims on binaries 27d0d59d6, 1516752ac, ca07df649, 032692409 and its parent, f20d5938a and its parent, a34d22451, 333643570.

## Moves, with causes
| Build(s) | Meta / ref | Run | Cause |
|---|---|---|---|
| Warrior arms/fury families | 1088 / 1136 / 1151 | 872 / 999 / 926 | Windfury white-swing rule (open finding results-20261109-1), plus flask |
| Hunter survival | 1041 | 917 | rule (-96) and flask (-28) |
| Paladin Ret, Sub Holy, Shock | 1103 / 1091 / 1030 | 1010 / 999 / 971 | rule (Shock Ret -59 from the rule alone) |
| 28 natural-flask presets | n/a | -17 to -45 | 032692409: Stamina 60 -> 30, zone bonus 5 -> 2 (client 70862). Feral 1169.5 -> 1151.5, BM 1195.6 -> 1174.6, Arms 1088.3 -> 1069.0, Fury 1137.1 -> 1092.5, Subtlety 885.6 -> 866.4 |
| Affliction (3) | 1034.1 | 1070.8 | hotfix 113766: 1022.7 -> 1077.1; f20d5938a (no dot tick after expiry) -> 1070.8 |
| Enhancement (4), Elemental (2) | now match | 1016 to 1031 | re-simmed in f8a37c8e8 and c61845d5c |
| Mage Arcane/Fire, Frost, Frostfire, Balance, Smite, Shadow | | unchanged | Arcane/Fire gaps (-6.6/-4.3) predate 27d0d59d6 |

Shaman Enhancement (1016 to 1031) now sits below Ret and Arms-class builds; the Stormstrike/imbue/totem tuning (cb651085c) and the Windfury rule explain it, and the report's own why notes were re-simmed.

## Conclusion
No new finding. Everything off is either the known Windfury rule (needs ruling) or waiting on the full re-sim and republish that brief 1516752ac asks for. Appended evidence to results-20261109-1 and results-20261104-2. The reference files were not replaced: the run is not a baseline until the re-sim lands.
