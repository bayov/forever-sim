# Results audit, pass 66, master d96adcce1

Why due: standard 32 (the raid's Sunder Armor ramps from the pull) changed 77 requests' `sunderArmor` to `sunderRamp`, plus mksettings `-sunder` and the sim UI presets. No other sim code changed.

## Checks
| Check | Result |
|---|---|
| Shared run (77 requests, 20k, seed 1) vs published | 36 of 77 moved more than 0.5, all physical builds, all negative |
| Size of moves | Feral -4.1 (-0.46%), Subtlety -3.7, hunters -2.4 to -3.1, warriors -2.3 to -2.9, other rogues -2.0 to -2.2, Ret Crusader duty -0.5 |
| Standard 32's stated range | Feral 3-4, rogues 2-4, hunters 2-3, warriors 0-3, Enhancement 1, Ret under 1, casters 0: matches (hunter BM 3.1 is 0.1 over the stated range) |
| Casters | none moved (no mage, warlock, priest, shaman Ele, Balance in the moved list) |
| Drift (regen.py vs ref/drift_report.txt) | identical, so mksettings's default `-sunder ramp` reproduces the committed requests |
| Unread fields | none (sunderRamp is read by the sim) |

## Conclusion
Quiet. The moves are the ruled Sunder ramp and match its measured cost. builds.json and the metas' published.dps are still pre-ramp, so requests read OFF published by 0.5 to 4.1 until the report re-sims; that is the expected lag, not a bug. Ranking order is unchanged at this size, except where spec pairs sit within about 4 DPS (not re-ranked here).
