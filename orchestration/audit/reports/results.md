# Results audit, pass 55 (master 99bd53ada)

Trigger: new request paladin__level_60_shock_ret (Level 60 Shock, Ret side) plus its APL, presets and builds.json row.

| Check | Result |
|---|---|
| Shared run (20k, seed 1) vs reference | 57 existing requests: 0 moved (>0.06); 1 new |
| paladin__level_60_shock_ret | mine 932.38, published 932.4 (exact); builds.json 932.4 |
| Request vs level_60_shock | differs only in talents, APL and the raid party block; same gear, consumes, 180 s |
| builds.json | 58 builds; diff vs previous is the id/format rewrite plus the new row |
| Standing gaps vs published (druid/warlock/priest/hunter, -1 to -6) | identical to the reference, not new |
| Drift (regen.py) | not run: no mksettings binary in the shared cache; new id maps to ret |

Ranking: Ret Shock (Ret side) 932.4 sits above Ret Shock 868.8 and below sub Holy per the preset comment (1.8% behind at 180 s); it avoids Light's Vigil and Twist of Light, so the gain over Shock comes from Seal of Command, Vengeance, Conviction and Champion of the Light. Not re-measured.
