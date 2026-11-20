# Tank mechanics, pass 24

- Master: e07e8fdb8 (from e671e881e). Client: 1.60.1.70959, unchanged (wago latest, 2026-11-20). No new update_ledger entries.
- Result: quiet. Open findings unchanged (20261117-1, 20261108-1, 20261107-1).

## What changed

| commit | file | what | tank effect |
|---|---|---|---|
| c04e51c13 | sim/core/target.go | `Encounter.OutdoorTerrain` and `Target.MobFamily` copied from the proto | none: new fields, read only by item effects (Golden Banana, Frozen Heart); no attack table, threat, Rage or scoring path reads them |
| 617a63635 | sim/paladin/items.go | `RegisterItemConditions(TenetsOfTheSilverHand, Undead)` | none on the sim: a BiS-ranking condition; the effect itself is unchanged |
| 449bff490 | sim/shaman/items.go | Totem of Rage (22395) on Forever: +2% to Earth/Flame/Frost Shock via `DamageMultiplierAdditive`, flat +30 kept for non-Forever | none on tank presets: every Warden gear set uses totem 272432; 22395 appears only in Enhancement BiS/gear and rotopt content. If a Warden ever equips it, threat follows damage through ThreatFromDamage (also under the Stormstrike mark, stormstrike.go) |

Totem of Rage's client description (spell 27859, aura 108 on mask 0x90100000 plus Flame Shock DoT 2%) is the DPS `mechanics` pass's to verify; not a tank mechanic.

## Tests

`go test -tags=with_db` on sim/core (Tank*), sim/warrior/..., sim/paladin/..., sim/druid/tank and sim/shaman/warden: all pass, goldens unchanged.
