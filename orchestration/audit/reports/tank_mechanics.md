# tank_mechanics pass 18

- master: 5f20dfdf4c (from 1a85df1d75)
- client: 1.60.1.70959 (was 70938; wago.tools latest; bare-build diff, hotfix tables not fetched)
- result: findings (1 new, low)

## What changed on master

| Commit | File | Change | Tank impact |
|---|---|---|---|
| ad1d7f469, 3e1214a69 | sim/warrior/talents.go | Push 114097 (SpellAuraOptions 124699 ProcTypeMask 4 -> 20): Unbridled Wrath procs on landed melee specials with damage > 0, Whirlwind excepted | Prot warrior: all three Level 60 presets take Unbridled Wrath 5/5 (`05-05050005-...`), so Shield Slam, Revenge and Heroic Strike now proc it; Sunder Armor (no damage) doesn't, see 20261117-1. Thunder Clap is ProcMaskSpellDamage, no proc |
| ad1d7f469 | sim/warrior/reckless_swings.go | One charge a Cleave cast | none: no tank preset uses Reckless Swings |
| ad1d7f469 | talents.go (Flurry) | comment only | none |
| 2ebdc14d6 | sim/warrior/execute.go | Avoided Execute keeps its extra Rage | none: no tank APL casts Execute |
| 6875e3b1e | sim/shaman/item_sets_pve.go | Five Thunders (1679) set | none: no Warden preset wears it (golden adds only an AllItems row) |

## Client 70938 -> 70959 (tier 1, bare build)

| Row | Change | Sim | Verdict |
|---|---|---|---|
| Judgement of Fury / Righteousness / Crusader / Light / Wisdom / Justice (20183-20188, 20280-20286, 20411-20414, 1311650/5) | SpellCategories DefenseType 2 -> 3 (ranged) | ranged specials (rangedJudgement, pass 13) | match; pass 13's notes-only label is now client tier 1. Command (20467...) unchanged = melee, matches pass 17 |
| Stormstrike 17364 | #2 150 -> 100, cooldown 6 -> 8 s | 100% / 8 s (pass 10) | match, now client tier 1 |
| Windfury Totem 8515/8516, 10608-10612 | AP aura, extra-attack effect, 2 charges, 1 s duration leave the buff spells; AP moves to raid auras | sim/core/buffs.go WindfuryTotemBuffAura unchanged | open: already in TASKS.md for a user ruling (5f20dfdf4); every tank preset carries the raid totem. Nov 16 hotfix: totem attacks now generate Rage (prot warrior) |
| Berserker Rage 18499 | new #2 mechanic immunity (aura 77) | n/a | no tank effect |
| Gloves of Holy Might 1330855 | proc mask | not modelled | no tank preset; in TASKS |
| Unsealed Wrath 1310735 | 100 ms proc recovery | no tank takes it | none |
| 12322 / 12966 / 1329555 | not in the bare build | push 114097 per commit | rests on the commit's hotfix citation |

## Measurement (20261117-1)

Patch `$AUDIT_WORK/tank_mechanics/uw_sunder.patch` drops `result.Damage > 0`. tankbench, 3000 iters, -labeled, seed 1:

| Preset | Objective (own weight) base -> patch | SafeDPS | TPS | Sunder casts |
|---|---|---|---|---|
| Max threat | 3854 ± 8 -> 3866 ± 7 | 2108 -> 2113 | 2049 -> 2053 | 58.3 |
| Balanced | 3140 ± 9 -> 3153 ± 8 | 1981 -> 1986 | 1929 -> 1934 | 47.3 |

## Engine and tests

- `go test -tags=with_db ./sim/warrior/...` passes. TestForeverTankUnbridledWrath pins Heroic Strike and Cleave at 0.6; nothing pins Shield Slam, Revenge or Sunder (covered by the generic `procsOnAbility`).

## Findings

New: 20261117-1 (low). Open: 20261108-1, 20261107-1.
