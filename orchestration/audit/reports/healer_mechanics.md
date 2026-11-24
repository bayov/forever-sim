# healer_mechanics pass 27 (master 566067a394, 2026-11-24)

Result: quiet.

## Get current
| Source | Seen | Change |
|---|---|---|
| Client build (wago.tools latest) | 1.60.1.71037 | none since pass 26 |
| Blizzard t/2403040 (Nov 20 hotfixes) | last post 2026-11-21T00:18:29Z | none newer |
| orchestration/update_ledger.md | no commits c78d983aca..566067a394 | none |

## What changed (watched paths, c78d983aca..566067a394)
| Commit | Files | Healer reach |
|---|---|---|
| be773f5eb Druid: Rip and Rake tick AP shares from Forever logs | sim/druid/rake.go, rip.go, tuning_70753.go (comment) | none |

- Rake tick AP share 0.055 x nov2BleedScale -> 0.065 flat; Rip AP share 0.01 x scale -> 0.0114 a point, outside the scale.
- `nov2BleedScale`, `rakeTickAPCoefficient`, `ripTickAPCoefficient` are referenced only in rake.go, rip.go,
  ferocious_bite.go and tuning_70753.go; none in sim/druid/healing.go or sim/druid/restoration/. No Restoration
  druid heal, talent or engine path is touched. Feral DPS accuracy belongs to the mechanics pass.

## Engine / tests
No engine or healing file changed; checks from pass 26 stand. Open: healer_mechanics-20261104-10 (untested paths),
-11 (Riptide mana).

## New findings
None.
