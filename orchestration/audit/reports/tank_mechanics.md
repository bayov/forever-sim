# tank_mechanics pass 34 (2026-11-25, phase0-tankheal 1624c3d56, fix 95962a7d3)

Forced pass (gate.py --force) on the Phase 0 tank branch. Client build and Blizzard posts not re-checked: pass 33
did both at 13:10 today (1.60.1.71091, hotfix thread post 15).

## What changed since pass 33

2f42f4a44 (tank threat batch), b663145ad / 295206ddf (procs, items), efeb06b95 (casters), 22d2cf952 (six rulings
to wowsims/forever's log), and the Phase 0 gear in every tank request.

## Bear threat batch (2f42f4a44) against wowsims/forever 59b8622 and the bench

| Rule | Ours | Theirs | Bench threat / damage | Verdict |
|---|---|---|---|---|
| Lacerate hit and ticks | 10/3, no flat | 10/3 (EV-0034, EV-0037) | 7.17 = 10/3 x 2.15 | right (their logs) |
| Primal Bite | 3x under Forever | 3x (EV-0034) | 6.45 = 3 x 2.15 | right (their logs) |
| Faerie Fire in Bear / Dire Bear | flat 2 x level, x form (258 at 60), none in Cat | the same | not cast by any bear APL | right; casting it before the Lacerate filler loses 40 to 313 (A/B below) |
| Blood Frenzy Rage threat | unscaled | unscaled (EV-0035) | - | right |
| Maul, Swipe | form only (2.15) | 1.75 on top, no log | Maul 2.15 | open call (verdicts.md) |
| Lacerate hit damage | was 10% x stacks before the cast | 10% x stacks after the cast (EV-0040) | - | **fixed, 95962a7d3** |

The bear's TPS jump is Lacerate and Primal Bite: on Magmadar (Balanced, 500 iterations) Lacerate is 797 TPS and
Primal Bite 571 of 2530; at the old 1x and 2x they would be 239 and 381, about 750 TPS less. The rule values match
theirs and the tests pin each (TestLacerate, TestPrimalBite, TestBearFaerieFireThreat, TestBearBloodFrenzyRageThreat).

Lacerate's hit (tank_mechanics-20261125-1): +170 to +185 score and +45 to +51 TPS on the three bear presets.

## Warrior

Shield Slam (shield_slam.go): 625-685 + 2 x block value, no attack power, flat 255, SuppressWeaponProcs and no
Windfury (EV-0004, EV-0050, EV-0155, user ruling 2026-11-25). Matches the commit's sources. Revenge still beats
Shield Slam in the priority (swap: -5.4 / -4.6 / -21.8, inside the error).

## Max survival on Phase 0 gear against synthetic: gear, not a bug

Paladin Max survival, same talents and APL, 2000 iterations:

| Boss | Gear | Max surv. | SafeDPS | Heal needed | DTPS | Health | Armor | Def | Miss / Dodge / Parry / Block | Crushes taken |
|---|---|---|---|---|---|---|---|---|---|---|
| Broodlord | Phase 0 | -1205 | 1792 | 532 (1.40x) | 381 | 8271 | 9391 | 382 | 9.7 / 18.4 / 12.7 / 7.7 | 9.8% |
| Broodlord | synthetic | 463 | 1602 | 305 (1.08x) | 283 | 8691 | 9816 | 463 | 12.9 / 20.3 / 15.9 / 17.9 | 2.5% |
| Chromaggus | Phase 0 | -2120 | 1667 | 612 | 502 | | | | | |
| Chromaggus | synthetic | -588 | 1501 | 403 | 363 | | | | | |

The synthetic set is a Phase 1 (Molten Core) budget: 127 Defense and 5% block on the gear, a Tier 1 2-piece (+7
Defense). The Phase 0 list has 53 Defense and no block chance, so block falls from 17.9% to 7.7%, avoidance plus block
from 67% to 48.5%, and crushing blows reach 9.8% of swings when Holy Shield is down (2.5% before). DTPS rises 35% and
the hits come spikier (1.40x DTPS for healing needed). At w = 3 the score is 2 x SafeDPS - 9 x healing needed, so
+190 SafeDPS cannot cover +227 healing needed. Block 7.7% = 5 + 82 x 0.04 - 0.6 (boss skill) checks by hand; no item
is missing from the database. Warrior (Def -87, armor -476) and bear (armor 9473 against 10159, dodge 35.0 against
38.1) on Chromaggus show the same: DTPS +17% and +9%, with no sim or scoring step out of line.

## Tests

go test for sim/druid/..., sim/warrior/..., sim/paladin/..., sim/shaman/..., sim/core and tools/tankscore: all pass.

---

## Master's pass, carried over from origin/master: Tank mechanics pass 36 (2026-11-26, master ab3fff700)

Merged in from master's audit loop (synthetic-gear tank/healer rows; numbers there are master's baseline, not the
Phase 0 one above). Kept for the verdicts and finding notes that exist only on master.

Quiet: no new finding. Due for 32 watched files changed (bfbee78fc..ab3fff700), all from the warrior leveling batch.

### Get current

- Client: still 1.60.1.71139 (`wago.py latest`); mechanics set fetched for the rank checks below.
- Blizzard: hotfix thread t/2403040 last post 2026-11-25T23:25Z (seen last pass); dev notes t/2360696 last post
  2026-11-04. Update ledger: only the audit loop's own commit.

### What changed (6 commits: 143bf4f72, 8f5b59325, 30bb3a590, 81a2e51fc, 1d20f40d4, adb69fdc7)

Warrior leveling sessions: every ability on rank tables (`sim/warrior/ranks.go`, `core.HighestKnownRank`,
`Player.untrained_spells`), stances by level (`KnowsStance`), reaches (`meleeReach` 5, Whirlwind/Thunder Clap 8,
Demoralizing Shout 10, Charge 8-25), and leveling-only Intercept, Victory Rush, Concussion Blow, Gore Drinker,
Rage decay. Core: `SunderArmorAuraOfRank`, `killTarget` calls the leveling `onKill`.

### Level 60 tank warrior: unchanged

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

### Lower ranks vs client 71139 (tier 1; leveling sessions only)

| Ability | Client | Sim | Verdict |
|---|---|---|---|
| Sunder Armor 7386-11597 | -90..-450 armor, E63 threat 34/75/117/158/206 | 90 x rank, same threat + 5% AP | ok |
| Demoralizing Shout 1160-11556 | E63 11/19/27/35/43 | same + 0.8/level for 10 levels (EV formula carried down) | ok (scaling below 54 is the rank-5 fit, tier 1.5) |
| Revenge 6572-11601 | 22/34/48/82/121, Variance .154/.2/.214/.208/.197 | 19-25, 27-41, 38-58, 65-99, 97-145 | ok |
| Shield Slam 23922-23925 | 430/505/580/655, Variance .043-.046 | 411-449, 483-527, 554-606, 625-685 | ok |
| Thunder Clap 6343-11581 | 10/23/37/55/82/103, 10/14/18/22/26/30 s | same | ok |

Tests: the batch adds `sim/warrior/tank_warrior/leveling_test.go` and `leveling_heals_test.go`; lower-rank
values have no per-rank pin, but they reach no tank preset (all level 60).

### Open findings

Unchanged: 20261124-4 (Stoneform autocast), 20261124-3 (racial tests), 20261117-1 (Unbridled Wrath/Sunder),
20261108-1 (Warden shock threat test), 20261107-1 (stale "client disagrees" comments).
