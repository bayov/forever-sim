# Healer mechanics: pass 30

2026-11-25, master 0404484fb6 (last 5536db4bc), client 1.60.1.71091 (unchanged). Result: quiet.

## Why due

master: `sim/core/spell_result.go` (b4cab8498), `sim/shaman/air_totems.go` and `sim/shaman/fire_totems.go` (5dbc01412).

## What changed, read in full

| Commit | File | Change | Healer reach | Verdict |
|---|---|---|---|---|
| b4cab8498 | sim/core/spell_result.go | New `PreAbsorbDamage`, reset in `NewResult`, set to `Damage` just before `absorbDamage` | Read only by warrior Rage from hits taken (`rage.go`); absorbs (Power Word: Shield, Divine Aegis) and health booking unchanged | no healing effect |
| b4cab8498 | sim/raid_healing_shaman_test.go | Ancestral Fortitude test 120 -> 180 s, as the tank's Rage now moves the seed's rolls | the test still checks the buff comes and goes | fine |
| 5dbc01412 | sim/shaman/air_totems.go, fire_totems.go | Comments only: Windfury and Flametongue Totem are raid area auras (effect 65) since client 70425 | none (DPS totems) | no code change |
| 5dbc01412 | docs/raid_sim.md | Scope table: Mana Spring (10494), Mana Tide (1325257), Healing Stream (10461) party area auras (35) | matches the verdicts' party reach | consistent |

## Get current

No new client build (71091) and no newer Blizzard post than 2026-11-21 recorded; nothing reopened.

## Open findings (unchanged)

- healer_mechanics-20261124-17 (Mana Well MP5 under Innervate/Energized)
- healer_mechanics-20261104-11 (Riptide mana)
- healer_mechanics-20261104-10 (untested heal paths)
