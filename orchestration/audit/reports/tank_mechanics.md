# Tank mechanics, pass 21

- Master: a6677362fc (from 8219f1083f). Client: 1.60.1.70959, unchanged (wago latest, 2026-11-18).
- Result: quiet.

## What changed (watched files)

| Commit | File | Tank effect |
|---|---|---|
| c8b5275b1 | sim/core/rage.go | None by default. `ForeverNormalizedRageLevelFactor` and `ForeverDamageTakenRage` read `reportSwitches.rageScaling` / `.damageTakenRageFactor`, which `parseReportSwitches` (sim/core/report_switches.go:86-87) defaults to `ForeverNormalizedRageScaling` and `ForeverDamageTakenRageFactor` unless a DASH_ASSUME knob is set. Standards 21 values unchanged. |
| c8b5275b1 | debuffs.go, druid/wrath.go, paladin/consecration.go, holy_strike.go, soc.go, warrior/rend.go | DPS Report assumption knobs (`Assume`/`AssumeOk`), identity with no env. Consecration's knob only scales ranks of level 40+ when set. No default change. |
| bbfd014e4 | core/health.go | `healthAt` (read only by heal_comm.go:25, the stand-in healer's view) is now a binary search over the time-ordered ring. Read against the old walk: same answer (oldest change after t's `before`; all-after-t reads the oldest kept; ring unset slots sit only on the old side, so the predicate is monotone). `sightEpoch++` on health changes and target death: cache invalidation only. |
| bbfd014e4 | rage.go, water_totems.go, sweeping_strikes.go, target_death.go | `SpellMetrics` field -> `SpellMetrics()` accessor (marks metrics taken for the per-iteration reset). Rage-gain threat metrics unchanged in value. |

Also in range, not in the watched list: ee35da473 "Haste shortens a spell's GCD under Forever" (cast.go
hastedGCD, standards 27; Warchief's Blessing +15% cast speed under Forever). Moved TestProtection and
TestWardenShaman goldens. No tank preset (ui/protection_paladin, tank_warrior, feral_tank_druid,
warden_shaman) carries Warchief's Blessing, so the presets' GCDs move only with other spell haste. Rule
reach for tanks: Warden shocks and Lightning Shield, Holy Shield, Consecration, Faerie Fire (Feral) are
hasted; Stormstrike, Holy Strike, Judgement (sim ranged/melee defense type), Maul/Swipe and all warrior
abilities keep 1.5 s. Client 20271 Judgement is DefenseType 0, but the cast's damage proc is ranged per
the Nov 11 hotfix; consistent with standards 27's "physical abilities keep theirs". No finding.

## Tests

`go test -tags=with_db` on sim/warrior/..., sim/paladin/..., sim/druid/..., sim/shaman/warden, sim/core:
all pass at a6677362fc. (Without `-tags=with_db` the class tests panic on missing items: a test-setup
note for the next pass, not a finding.)

## Findings

No new findings. Open, unchanged (none fixed in this range): 20261117-1 (UW `result.Damage > 0` still at
sim/warrior/talents.go:238), 20261108-1, 20261107-1 (stances.go:110, bear.go:57 still stale), 20261112-2,
20261112-1, 20261111-1, 20261104-15, 20261104-16.
