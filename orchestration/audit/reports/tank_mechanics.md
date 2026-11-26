# Tank mechanics pass 36 (2026-11-26, master ab3fff700)

Quiet: no new finding. Due for 32 watched files changed (bfbee78fc..ab3fff700), all from the warrior leveling batch.

## Get current

- Client: still 1.60.1.71139 (`wago.py latest`); mechanics set fetched for the rank checks below.
- Blizzard: hotfix thread t/2403040 last post 2026-11-25T23:25Z (seen last pass); dev notes t/2360696 last post
  2026-11-04. Update ledger: only the audit loop's own commit.

## What changed (6 commits: 143bf4f72, 8f5b59325, 30bb3a590, 81a2e51fc, 1d20f40d4, adb69fdc7)

Warrior leveling sessions: every ability on rank tables (`sim/warrior/ranks.go`, `core.HighestKnownRank`,
`Player.untrained_spells`), stances by level (`KnowsStance`), reaches (`meleeReach` 5, Whirlwind/Thunder Clap 8,
Demoralizing Shout 10, Charge 8-25), and leveling-only Intercept, Victory Rush, Concussion Blow, Gore Drinker,
Rage decay. Core: `SunderArmorAuraOfRank`, `killTarget` calls the leveling `onKill`.

## Level 60 tank warrior: unchanged

| Check | Result |
|---|---|
| Rank picked at 60 | top non-AQ rank for every ability, as before (HighestKnownRank with no untrained spells) |
| Heroic Strike / Cleave threat | 2.5 x 56 = 140 / 1.6 x 60 = 96, as before |
| Demoralizing Shout threat | rank 5: floor(43 + 0.8 x (min(L, 64) - 54)), same formula at 60 |
| Sunder | rank 5, 450 armor a stack, 206 + 5% AP threat, ActionID 11597 |
| Area abilities (TC, Cleave, WW) | `AreaTargets` outside leveling = AoETargetCount + NextTargetUnit, the old loop |
| Melee MaxRange = 5 | `Spell` range check and APL wait read `LevelingDistance`, 0 outside leveling |
| castBySwing (HS/Cleave) | nil outside leveling |
| Intercept, Victory Rush, Concussion Blow, Gore Drinker, Rage decay | registered only when `IsLeveling()` |
| Retaliation | registered at 20+ in raids too, APL-only (no MCD); no tank APL names it |
| Anger Management | the in-combat gate is leveling-only |
| Unbridled Wrath | now also skips Victory Rush and Retaliation strikes (leveling-only spells); open 20261117-1 untouched |
| Golden test | `TestP1TankWarrior.results` not changed in the range; `go test ./sim/warrior/... ./sim/core` green |

## Lower ranks vs client 71139 (tier 1; leveling sessions only)

| Ability | Client | Sim | Verdict |
|---|---|---|---|
| Sunder Armor 7386-11597 | -90..-450 armor, E63 threat 34/75/117/158/206 | 90 x rank, same threat + 5% AP | ok |
| Demoralizing Shout 1160-11556 | E63 11/19/27/35/43 | same + 0.8/level for 10 levels (EV formula carried down) | ok (scaling below 54 is the rank-5 fit, tier 1.5) |
| Revenge 6572-11601 | 22/34/48/82/121, Variance .154/.2/.214/.208/.197 | 19-25, 27-41, 38-58, 65-99, 97-145 | ok |
| Shield Slam 23922-23925 | 430/505/580/655, Variance .043-.046 | 411-449, 483-527, 554-606, 625-685 | ok |
| Thunder Clap 6343-11581 | 10/23/37/55/82/103, 10/14/18/22/26/30 s | same | ok |

Tests: the batch adds `sim/warrior/tank_warrior/leveling_test.go` and `leveling_heals_test.go`; lower-rank
values have no per-rank pin, but they reach no tank preset (all level 60).

## Open findings

Unchanged: 20261124-4 (Stoneform autocast), 20261124-3 (racial tests), 20261117-1 (Unbridled Wrath/Sunder),
20261108-1 (Warden shock threat test), 20261107-1 (stale "client disagrees" comments).
