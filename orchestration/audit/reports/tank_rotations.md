# tank_rotations pass 30

Master 0404484fb6 (2026-11-25), against pass 29's master 5536db4bbb and the baseline of eceb090196.
Quiet: no new finding.

## What changed

Gate (5536db4 .. 0404484): `sim/core/rage.go` and `sim/warrior/stances.go` (b4cab8498: Rage from hits taken now
follows the stance's damage taken modifier, 0.9 in Defensive Stance, and twice the Rage for non-Physical hits,
periodic ticks and Physical specials), `sim/core/target.go` (Encounter.outside_raid, world buffs off in raid
instances: 3275a60b1, 9701cff08), and comment-only edits to `sim/shaman/air_totems.go` and `fire_totems.go`
(5dbc01412). Also in range: f1942793d (Arcane Brilliance and Divine Spirit overwrite the elixirs of their stat) and
75dee7458 (Armor Shatter / Puncture Armor share a slot). No tank request, APL, rotopt template, encounter,
tankbench or tankscore change. No tank request ticks a world buff.

## Baseline (1,000 iterations, default seed; `ref/tank_bench/summary.tsv` rewritten)

Attributed with a second full baseline from cdd0b2482 (just before the Rage change, `tools/audit/build_cli.sh`).

| request (scoring mean) | old | before Rage change | now | cause |
|---|---|---|---|---|
| warrior_protection_balanced | 3454.2 | 3454.2 | 3450.5 | Rage rule, noise |
| warrior_protection_max_threat | 4233.3 | 4233.3 | 4233.0 | Rage rule, noise |
| warrior_protection_max_survival | 1832.8 | 1832.8 | 1820.0 | Rage rule (Broodlord -95 ± 70, Vael -70 ± 45, Onyxia +83 ± 60) |
| warrior_protection (legacy) | -14.2 | -14.2 | -49.0 | Rage rule (TPS -9) |
| druid_bear_balanced | 2130.0 | 2130.0 | 2152.8 | Rage rule (Onyxia +106, healing needed -1.2%) |
| druid_bear_max_threat | 3131.8 | 3131.8 | 3133.7 | noise |
| druid_bear_max_survival | 320.7 | 320.7 | 431.8 | Rage rule (more Rage from specials and ticks: healing needed -3.4%, Onyxia +262, Broodlord +152) |
| shaman_warden_balanced | 1264.7 | 1238.0 | 1238.0 | f1942793d |
| shaman_warden_max_threat | 2621.7 | 2599.7 | 2599.7 | f1942793d |
| shaman_warden_max_survival | -2263.5 | -2283.2 | -2283.2 | f1942793d |
| paladin (all three) | 2809.0 / 3243.7 / 1787.2 | same | same | bit-identical |

Warden: every boss moves 20 to 30 the same way (TPS -0.7 to -1.0%). The cause is f1942793d: Arcane Brilliance in
the raid buffs now overwrites the request's Elixir of Greater Intellect. Removing the elixir on master scores
0.0 ± 0.0 at 1,000 iterations (all bosses identical), so the elixir slot is dead. That is a consumables question,
not a rotation one, and there's no finding here. The Rage rule moves the warrior and bear rows as it should (an
intended mechanic change, already ruled on), with no preset's mean moving more than 0.7% except bear Max survival,
whose small mean (431.8) makes +111 look large.

## Rotation lines (Rage thresholds re-checked under the new Rage rule)

Paired abtest on the scoring mean, 3,000 iterations, seeds 1 and 2, B minus preset, at the preset's weight:

| change | seed 1 | seed 2 | verdict |
|---|---|---|---|
| warrior Balanced Heroic Strike 40 -> 30 | -1.2 ± 5.7 | +7.4 ± 6.1 | noise, keep |
| warrior Balanced Heroic Strike 40 -> 50 | -9.1 ± 5.7 | -8.3 ± 6.2 | loses |
| warrior Max threat Heroic Strike 30 -> 20 | -1.0 ± 1.9 | +2.5 ± 2.2 | noise, keep |
| warrior Max threat Heroic Strike 30 -> 40 | -9.2 ± 2.0 | -7.5 ± 2.2 | loses |
| warrior Max survival Heroic Strike 40 -> 30 | -8.2 ± 13.4 | -0.6 ± 11.7 | noise, keep |
| warrior Max survival Heroic Strike 40 -> 50 | -40.4 ± 13.0 | -39.8 ± 11.8 | loses |
| warrior Balanced Bloodrage under 30 -> 20 | -1.1 ± 2.6 | -0.5 ± 2.8 | noise |
| warrior Balanced Bloodrage under 30 -> 40 | +0.1 ± 1.0 | -1.8 ± 1.3 | noise |
| bear Balanced Maul 20 -> 10 | -0.1 ± 3.5 | +0.9 ± 2.8 | noise |
| bear Balanced Maul 20 -> 30 | -3.0 ± 4.6 | -7.9 ± 3.8 | loses |
| bear Max survival Maul 40 -> 30 | +14.1 ± 18.0 | -1.4 ± 19.5 | noise |
| bear Max survival Maul 40 -> 50 | -73.6 ± 18.6 | -125.5 ± 21.4 | loses |
| bear Max survival Swipe 80 -> 70 | -110.2 ± 17.1 | -125.8 ± 17.7 | loses |
| bear Max survival Swipe 80 -> 90 | +5.5 ± 17.2 | -13.7 ± 17.5 | noise |

Every Rage threshold holds under the new rule. Variants: `$AUDIT_WORK/tank_rotations/v/` (made with
`mk_warrior.py ... setcond <line> <cond>`; abtest built in the master CLI tree).

## Open findings

- 20261122-1 (paladin Balanced Consecration floor): paladin bit-identical, still open, not re-tested.
- 20261104-11, 20261104-3 (Broodlord-dependent trinket hold and Bulwark timer): user calls, untouched.

## Left for a later pass

- The Windfury main-hand timing open item in `passes/tank_rotations.md` (user request 2026-11-15) was not done this
  pass; it stays open.
