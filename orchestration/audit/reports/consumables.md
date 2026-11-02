# Consumables pass 62 (master 5a6e9c6a6c)

Quiet on new findings. Client build unchanged (1.60.1.70707, wago latest probe).

## Moved files since c23ef0440 (11)
| File | Change | Consumables impact |
|---|---|---|
| ui/retribution_paladin/presets.ts, tools/rotopt/mksettings/spec_ret.go | ret_shock drops Major Mana Potion; ShockConsumes = DefaultConsumes (Dragonfire Potion) | presets.ts and mksettings agree; pot choice sim-backed in comments (+5.3 to +6.0) |
| ui/balance_druid, feral_druid presets.ts | hotfix talent strings, feral APL tooltip | none (no consumes touched) |
| ui/hunter/presets.ts | BM talent string, comment | none; open Sapper finding unchanged |
| ui/mage/presets.ts | tooltip rename (Undead -> Skyborne High Order), comments | none |
| ui/feral_tank_druid, restoration_druid, holy_paladin, protection_paladin | out of scope | n/a |
| tools/dps_report/assumptions_def.py | new pal_unsealed_wrath and druid_predatory_discipline assumptions | no buff/oil assumption changed |

## Verified
- scan.sh on master: all ban sections empty (banned, Revelation, trinkets, crit elixirs).
- Counts as before; minor-armor exclusive-effect count prints 3 (ref expects 2), unchanged in sim files this pass (no debuffs.go change), left as-is.
- Class branches show "ahead" only through stale copies of shared files; no branch-own commits audited.

## Open items
- consumables-20261101-1 (ranged hunter Sapper, low) still open.
