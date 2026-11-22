# healer_mechanics pass 26 (2026-11-22, master 7a6d1beb9d)

Range 8b5137c6d6..7a6d1beb9d. Gate: 4 watched files (sim/core/dot.go, sim/paladin/item_sets_pve.go,
sim/shaman/flame_shock.go, sim/shaman/talents.go). Client still 1.60.1.71037 (wago latest). Blizzard t/2403040: no post
after post 14 (2026-11-21T00:18:29Z, no healer line). TASKS.md / daily update: Nov 20 server pushes 114204-114326
(Execute, Judgement procs, Chastise, Intensity, Nature's Bounty, Spiritcaller) name no heal, healing talent or
healing-engine row. Update ledger unchanged.

## Changed files

| File | Change | Healer effect | Verdict |
|---|---|---|---|
| sim/core/dot.go | `SetTickPeriod` retimes on the dot's expiry grid (3960dd1f6); CleanUp change is a comment | only caller is Wrack (`sim/warlock/talents.go:331,333`); no HoT calls it; CleanUp logic unchanged, so pass 9's "last tick at expiry lands" holds | no healer effect |
| sim/paladin/item_sets_pve.go | Soulforge (Classic) 4pc reuses an existing "Crusader's Wrath" aura so 4 + 4 Forever Soulforge pieces share one buff | spell power proc on melee; Holy presets wear `level60_phase1_synthetic` gear, no Soulforge | none |
| sim/shaman/flame_shock.go | Forever: haste shortens Flame Shock's GCD | damage spell; Restoration preset casts no Flame Shock in healing fights | none |
| sim/shaman/talents.go | Maelstrom Weapon: no stack from Windfury Totem's swing under Forever | Enhancement talent | none |
| sim/common/item_sets/forever_dungeon_set_1_reveal.go (unwatched) | 8 Dungeon Set 1 reveal sets wired, incl. Vestments of the Virtuous (priest) and Soulforge (paladin) | no healer preset wears them (all healer gear is synthetic); Deathmist 4pc self-heal not modelled, a DPS set | none |

## Tests

`go test -tags=with_db ./sim/paladin/holy/ ./sim/priest/... ./sim/shaman/... ./sim/druid/... ./sim/core/`: all green (holy, priest
healing/shadow, shaman elemental/enhancement/warden, druid, balance, feral, tank, core). (Without `-tags=with_db` the holy suite panics on a missing item 272436; that's the build tag, not a bug.)

## Findings

None new. Open: 10 (untested paths, low), 11 (Riptide cost 147/180/231 vs client 150/180/230, low), unchanged (no
restoration code in range).
