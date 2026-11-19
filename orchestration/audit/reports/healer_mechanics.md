# healer_mechanics pass 22 (2026-11-19, master 8f388c7f3)

Range eebd00b2b..8f388c7f3, 1 watched file, 1 commit touching it (65f0ea9d6). Client 1.60.1.70959 (wago latest,
unchanged). Blizzard posts not rechecked (last seen t/2403040, 2026-11-17T02:18Z).

## What changed

| Source | Change | Healer effect | Verdict |
|---|---|---|---|
| sim/druid/healing.go:666 swiftmendHot (65f0ea9d6) | the `balanceSpells(Rejuvenation, Regrowth)` flatten+filter replaced by a nested loop over the two rank lists, skipping nil | same ranks visited in the same order (Rejuvenation ranks, then Regrowth), same nil skip (talents.go:82), same strict `<` on ExpiresAt so ties still keep the first found; only drops the allocation. Callers healing.go:715 (cast check), :723, :735 unchanged | OK, equivalent |

Swiftmend's rule (consume the Rejuvenation or Regrowth HoT with the least time left) stays as verified in
`ref/healer_mechanics/verdicts.md`; nothing reopened.

## Tests

`go test -tags=with_db ./sim/druid/...` and `go test -tags=with_db ./sim/ -run 'Druid|Swiftmend|Heal'`: all pass
(Swiftmend covered by `sim/raid_healing_druid_test.go`, `sim/raid_healing_druid_more_test.go`). Note: a fresh
checkout needs `make sim/core/proto/api.pb.go` (protoc-gen-go on PATH) and the `with_db` tag, or the suites fail at
setup / "No item with id".

## Findings

None new. Open: 10 (untested paths, low), 11 (Riptide cost 147/180/231 vs client 150/180/230, low), unchanged.
