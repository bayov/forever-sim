# tank_mechanics pass 17

- master: 25ce81de77 (from 70ea588bb1)
- client: 1.60.1.70938 (unchanged; wago.tools latest)
- result: quiet

## What changed

| Commit | File | Change | Tank impact |
|---|---|---|---|
| 9ec69044f, 6ce77e5ef | sim/paladin/soc.go, judgement.go | Judgement of Command is a melee special again (Nov 13 hotfix t/2403040, client SpellCategories Melee; user ruling 2026-11-15) | none: prot paladin presets judge Fury, which stays ranged |
| 9ec69044f | sim/paladin/sor.go, seal_of_fury.go, soc.go (+ sim/core/buffs.go) | Windfury Totem rolls on every landed main hand melee hit; the seal procs carry SpellFlagNoWindfuryTotem | prot paladin (Seal of Fury, Holy Strike, raid Windfury Totem): Holy Strike now rolls it, the Seal of Fury proc doesn't. Consistent with the notes ("all melee abilities"); seal procs are on-hit effects, not abilities (tier 1 notes, reading) |
| baa363615 | sim/paladin/talents.go, consecration.go, seal_of_fury.go, sor.go | Two-Handed Weapon Specialization's Holy part follows defense type (user ruling 2026-11-15) | none: prot paladins hold a shield, multiplier 1 |
| 78c395984, 98291cfa3 | sim/druid/healing.go | Wild Growth tick 0.066 (client 70938), ranks share category 1237 | none: no bear preset casts it |
| 9f8a71d2d | sim/shaman/windfury_weapon.go | Windfury Imbue aura exposes its Icd via ActionIDForProc (aura.go:461, lookup only) | none: Warden presets use Rockbiter |
| f2a5dfe25 | tools/tankbench/req/warrior_protection_{balanced,max_survival}.json | Revenge before Shield Slam, a 2s -> 8s and 50 -> 40 threshold | rotation tuning, no mechanic |

## Engine and tests

- `windfuryTotemRollsOnForever` (sim/core/buffs.go:2216): MH proc mask, not flagged, not an extra attack (pass 14). Covered by sim/paladin/retribution/windfury_totem_test.go (TestJudgementsAndWindfuryTotem, seal procs).
- No new tank ability or talent without a test.

## Findings

None new. Open: 20261108-1 (Warden shock threat under Stormstrike, untested), 20261107-1 (client 70862 threat comments).
