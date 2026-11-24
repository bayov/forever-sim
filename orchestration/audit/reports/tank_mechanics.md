# tank_mechanics pass 29 (2026-11-24, master 566067a394)

Why due: 6 watched files changed since c78d983aca (cd71e78c2 warrior, be773f5eb druid).
Client: 1.60.1.71037 still newest (wago.py latest). Update ledger: nothing since the last pass.

## Changes read

| Commit | File | Tank effect | Verdict | Tier |
|---|---|---|---|---|
| cd71e78c2 | sim/warrior/talents.go `applyUnbridledWrath` | Shield Slam, Revenge, Shield Bash (and Whirlwind) no longer proc Unbridled Wrath; white, totem, Heroic Strike, Cleave still do | Matches wowsims/forever EV-0148, a level 60 Protection warrior log (Shield Slam 0/977, Revenge 0/601, Shield Bash 0/44; read 2026-11-24) | log (tier 1.5) |
| cd71e78c2 | sim/warrior/reckless_swings.go | Heroic Strike / Cleave spend one charge a cast, landed or not; white swings only when landed | Matches EV-0138 (Fury log); prot takes Reckless Swings, so an avoided Heroic Strike now spends a charge | log |
| cd71e78c2 | sim/warrior/talents.go Flurry consume, heroic_strike_cleave.go `isQueuedSwing` | Flurry: Fury only; prot presets don't take Flurry | n/a for tanks | - |
| be773f5eb | sim/druid/rake.go, rip.go, tuning_70753.go | Cat Rake/Rip AP shares | Bear never casts Rake or Rip; Lacerate untouched | n/a |

## Tests

`go test -tags=with_db ./sim/warrior/... ./sim/druid/... ./sim/core/`: all pass.
Exclusion pinned by `TestForeverUnbridledWrathOnAbilities` (dps_warrior/forever_mechanics_test.go:747),
Reckless Swings by `TestForeverRecklessSwingsHeroicStrikeAvoided` / `...CleaveOnce`.
(Without `-tags=with_db` the warrior packages panic on item 991014: a build-tag need, not a bug.)

## Findings

- No new findings.
- 20261117-1 (Sunder and Unbridled Wrath): new evidence added. EV-0148 doesn't count Sunder, but every
  other prot special reads 0, so the sim's exclusion of Sunder is the likely server behaviour; suggest
  ruling (b).
- 20261108-1, 20261107-1: unchanged, still open.
