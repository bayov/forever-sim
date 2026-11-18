# healer_mechanics pass 20 (2026-11-18, master 8219f1083)

Range c198d81e3..8219f1083, 2 watched files (sim/core/dot.go, sim/druid/hurricane.go), one commit (8219f1083).
Client 1.60.1.70959 (wago latest, unchanged since pass 17). Blizzard posts not rechecked this pass (last seen
t/2403040 edit 2026-11-17T02:18Z).

## What changed

| Source | Change | Rule / source | Sim | Verdict | Tier |
|---|---|---|---|---|---|
| sim/core/dot.go (8219f1083) | Under Forever a dot flagged AffectedByCastSpeed ticks at base x cast speed, rounded to the ms, read at application; HasteReducesDuration keeps the ticks (Volley), otherwise keeps the duration and adds ticks rounded half to even (Hurricane). MaxTicksRemaining adds hasteTicks | wowsims/forever b83fe9df5f, user ruling 2026-11-18, to verify against a log | hastedTicks returns (TickLength, 0) for every unflagged dot, so RecomputeAuraDuration is unchanged for them | no healer effect | other sim (ruling) |
| sim/druid/hurricane.go | AffectedByCastSpeed under Forever | as above | damage spell only | out of scope (DPS mechanics pass) | - |

Healer reach: `grep AffectedByCastSpeed sim/` sets it true only on Hurricane and Volley (Mind Flay, Siphon Life
and one item effect set it false). No HoT or heal channel (Renew, Rejuvenation, Regrowth, Tranquility, Wild Growth,
Riptide, Healing Stream) is flagged, so their ticks, durations and MaxTicksRemaining (read by priest
healing_talents.go:270) are unchanged. Added an engine row to ref/healer_mechanics/verdicts.md.

## Tests

At 8219f1083 (cached worktree with protos): `go test -tags=with_db ./sim/priest/healing/ ./sim/paladin/holy/
./sim/core/` ok; `go test -tags=with_db ./sim/ -run 'Healing|Heal'` ok. No healing golden moved.

## Findings

- New: none.
- Resolved: none.
- Still open: 10 (test coverage), 11 (Riptide mana 147/180/231 vs client 150/180/230).
