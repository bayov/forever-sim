# Tank mechanics, pass 23

- Master: e671e881e (from 8f388c7f3c). Client: 1.60.1.70959, unchanged (wago latest, 2026-11-19). No new update_ledger entries.
- Result: quiet. Open findings unchanged (20261117-1, 20261108-1, 20261107-1).

## What changed

| commit | file | what |
|---|---|---|
| d73c17568 | sim/core/tank_score.go | `leastHps` restarts each replay from the last point the lower bound's health was capped at max health (`replayFrom`); `lives` wraps it from event 0 |

## Engine check: the tank score's replay

| check | verdict |
|---|---|
| Per-event arithmetic same as before (`min(maxHealth, health) - damage`, `min(maxHealth, health + hps*units)`, `min(maxHealth, health + heal)`) | same, rewritten as branches |
| Checkpoint valid for every higher HPS | yes: health at each event is non-decreasing in HPS (IEEE add and mul are monotone; maxHealth per event does not depend on HPS), so where the lower bound caps at max health a higher HPS caps too, and the state after it is identical |
| Checkpoint only advanced on a failed replay (`lo` moves up), and later probes are all above `lo` | yes (doubling loop and bisection) |
| +Inf probe from the HPS-0 checkpoint | valid (Inf >= maxHealth caps; units > 0, so no Inf*0) |
| Nothing in a scoring run reads the tank's health | unchanged; the replay uses only recorded events |
| Measured equivalence | temporary test, 20000 random event lists (200-600 events, max-health drops, flat heals): old and new `leastHps` bit-identical in all 20000. Test file removed |

## Tests

`go test -tags=with_db` on sim/core (Tank*), sim/warrior/..., sim/paladin/..., sim/druid/tank and sim/shaman/warden: all pass, and the goldens are unchanged.
TestTankScoreReplayCapsAtMaxHealthDrop covers `lives`. No committed test exercises the checkpoint path through `leastHps` with a failing low bound. This is not a finding: the code is right and the goldens pin the score end to end.

Note for the next pass: the Go protos are not checked in. Run `make sim/core/proto/api.pb.go` (needs protoc-gen-go on PATH, `$HOME/go/bin`) before `go test`.
