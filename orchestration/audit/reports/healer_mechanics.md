# Healer mechanics, pass 14 (2026-11-13)

Master 3a86d4a6bf against last pass cad23ae1a9. Client 1.60.1.70938 (unchanged). Result: quiet. Findings 10, 11, 14, 15 still open.

## Get current

| Source | Seen | Result |
|---|---|---|
| Newest client (`wago.py latest`) | 1.60.1.70938, same as pass 13 | no new rows to check |
| Blizzard notes | thread t/2403040 | not re-read this pass (no new build; last pass hit the rate limit) |

## What changed (cad23ae1a9..3a86d4a6bf, watched paths)

| File | Change | Healer reach | Verdict |
|---|---|---|---|
| `sim/shaman/flametongue_weapon.go` | Flametongue Weapon rank 6 cut 0 -> 5.5 (log-derived, Forever) | none: no healer preset or healbench request uses Flametongue (all use Wizard Oil or Lesser Mana Oil) | n/a |
| `sim/core/buffs.go` | Windfury Weapon attacks roll Windfury Totem under Forever | melee only | n/a |
| `sim/core/consumes.go` | Shadow Oil / Frost Oil proc spell IDs 350026 / 350025 under Forever | ID only, DPS oils | n/a |

No heal rank, healing talent, heal-target, regen or raid-damage code changed.

## Open findings re-checked on master

- 15 (Wild Growth coef): `sim/druid/healing.go:174` still 0.033; open.
- 11 (Riptide cost): `sim/shaman/healing.go:117` `RiptideNotesManaScale` 0.6 still in place; open.
- 10, 14: no related change; open.
