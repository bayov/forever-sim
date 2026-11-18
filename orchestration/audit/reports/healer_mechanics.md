# healer_mechanics pass 21 (2026-11-18, master a6677362f)

Range 8219f1083..a6677362f, 20 watched files, 4 commits touching them (7199c8f9c, c8b5275b1, f7963e0ca merge,
bbfd014e4). Client 1.60.1.70959 (wago latest, unchanged). Blizzard posts not rechecked (last seen t/2403040,
2026-11-17T02:18Z).

## What changed

| Source | Change | Healer effect | Verdict |
|---|---|---|---|
| heal_comm.go (7199c8f9c) | seenHealthChangesAt: the idle skip (apl_idle.go:111) wakes a reaction time after the first hit or heal on a candidate the healer doesn't see yet | healers react at the right time when skipping; same fights as no skipping (tools/healbench/main_test.go) | OK, fixes a determinism bug |
| heal_sight.go, heal_targets.go, heal_comm.go (bbfd014e4) | per-check cache of other healers' heals in flight and the Most Hurt / Lowest Health pick; dropped on sightEpoch | sightEpoch bumps on health gain/loss, deaths (player, target, pet), stat change (unit.go), cast start/end/move (cast.go, gcd.go), auras (aura.go). Same terms in the same order, so sums are bit-equal | OK |
| health.go (bbfd014e4) | healthAt: binary search over the 32-change ring | predicate "set and after t" is monotone going back (time-ordered, unset tail; all-after-t returns oldest, as before). health_history_test.go added | OK |
| spell_result.go, mana.go, dot.go, priest/healing.go, water_totems.go, shield.go, absorb.go | SpellMetrics[] -> SpellMetrics() accessor | metrics only | no effect |
| druid/wrath.go, priest/smite.go, priest/talents.go (Shadowform), paladin consecration/holy_strike/soc | DPS Report assumption knobs (core.Assume / AssumeOk) | defaults equal the old constants; knobs unset outside the report; damage spells only | no healer effect |

No heal rank, talent, cost, coefficient or HoT changed in the range. Added an engine row to
ref/healer_mechanics/verdicts.md.

## Tests

At a6677362f: `go test -tags=with_db ./sim/core/ ./sim/shaman/... ./sim/paladin/holy/... ./sim/druid/... ./sim/priest/...`
ok; `go test -tags=with_db ./sim/ -run 'Heal|Raid'` ok; `go test -tags=with_db ./tools/healbench/` ok.
Untested (finding 10): no unit test that a stale sight is dropped when a cast starts mid-check; healbench's
idle-skip equivalence test covers it end to end.

## Findings

- New: none.
- Resolved: none.
- Still open: 10 (test coverage), 11 (Riptide mana 147/180/231 vs client 150/180/230; sim/shaman/healing.go unchanged).
