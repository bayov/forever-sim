# healer_mechanics pass 23 (2026-11-20, master e07e8fdb8)

Range e671e881e..e07e8fdb8, 3 watched files, 3 commits touching them (c04e51c13, 617a63635, 449bff490). Client
1.60.1.70959 (wago latest, unchanged). Update ledger unchanged in range. Blizzard posts not rechecked (last seen
t/2403040, 2026-11-17T02:18Z).

## What changed

| Source | Change | Healer effect | Verdict |
|---|---|---|---|
| sim/core/spell_result.go `critChanceAgainst` / `AddCritChanceAgainst` (c04e51c13) | per-target crit bonus (a caster's mark on a unit, by global `UnitIndex`), added to `PhysicalCritChance` and `SpellCritChance` | direct heals roll `HealingCritChance()`, which has no target term: unchanged. HoT snapshots (`SnapshotHeal`, :723) call `SpellCritChance(friendly)`, but the only writer (Blackhand's Breadth, `sim/common/item_effects.go:3206`) marks enemies via `NewEnemyAuraArray`, and `UnitIndex` is one index space over targets + raid (`environment.go:98-103`), so a friendly's entry stays 0. No collision of enemy and friendly indexes. | no healer effect |
| sim/paladin/items.go (617a63635) | `RegisterItemConditions(TenetsOfTheSilverHand, Undead)` | damage libram condition only; Holy presets don't use it for healing | no healer effect |
| sim/shaman/items.go Totem of Rage (449bff490) | Forever: +2% additive damage on Earth/Flame/Frost Shock instead of +30 bonus damage | shock damage only; no Restoration heal touched | no healer effect (DPS mechanic, for `mechanics`) |

Nothing in `ref/healer_mechanics/verdicts.md` reopened.

## Tests

`go test -tags=with_db ./sim/core/... ./sim/shaman/... ./sim/paladin/...` and
`go test -tags=with_db ./sim/ -run 'Heal|Shaman|Paladin|Priest|Druid'`: all pass.

## Findings

None new. Open: 10 (untested paths, low), 11 (Riptide cost 147/180/231 vs client 150/180/230, low), unchanged.
Low-weight note, not a finding: `SpellCritChance(target)` in HoT snapshots would pick up a future friendly-target
mark; a test pinning heal crit to `HealingCritChance` would guard it (fits finding 10's list).
