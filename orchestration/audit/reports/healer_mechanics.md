# healer_mechanics pass 16 (2026-11-15, master 25ce81de77)

Range 70ea588bb1..25ce81de77, 8 watched files. Client still 1.60.1.70938 (wago.tools latest); update ledger unchanged in range.

## What changed

| File | Change | Healer reach | Verdict |
|---|---|---|---|
| sim/druid/healing.go | Wild Growth tick coefficient 0.033 -> 0.066 (78c395984); comment cites SpellCategory 1237 shared CD | Restoration druid raid preset | match client 70938 SpellEffect 1072065/1235327/1235329 (tier 1); finding 15 resolved |
| sim/paladin/talents.go | 2H Weapon Spec Holy part moved from the school to melee-defense spells + Consecration (user ruling 2026-11-15) | none: damage multiplier only, Holy paladin heals untouched | n/a |
| sim/paladin/judgement.go, soc.go, sor.go, seal_of_fury.go, consecration.go | Judgement of Command on melee defense; seals take the spec via their own multiplier | Ret only | n/a |
| sim/shaman/windfury_weapon.go | Windfury imbue aura keyed for its ICD (Stormstrike gating) | Enhancement only | n/a |

## Tests

`go test -tags=with_db ./sim/druid/... ./sim/paladin/holy/...`: all ok. `healing_ranks_test.go:87` now pins 0.066, so
the Wild Growth coefficient has a failing test if it regresses. (Without `-tags=with_db` the balance/feral/tank suites panic
on synthetic item ids; that is the build tag, not a defect.)

## Open items

- `docs/healing_druid.md:115-117` still list Wild Growth at 0.033 (doc only; noted under finding 15).
- Findings 10 (test coverage) and 11 (Riptide mana) stay open.

No new findings.
