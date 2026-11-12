# tank_rotations pass 9

Master cb0d928fd (2026-11-12), against pass 8's 707ae92cc (state 33db32b36). 1,000 iterations, default seed.

## What changed

Since 33db32b36: 209b355e5 (audit log), 934c698a2 (new items, no tank request touches them), cb0d928fd (talent
tooltips). The only sim edits in range:
- sim/shaman/talents.go: `Talents.ImprovedFireNova` -> `Talents.TrialByFire`, a proto field rename on the same
  field number 9 (proto/shaman.proto:18), so talent strings and behaviour are unchanged.
- sim/druid/talents.go: comment only (Rend and Tear).
- sim/priest/priest.go: TalentTreeSizes Shadow 18 -> 17 (not a tank; never read).

No tank APL, request, rotopt template, encounter, tankbench or tankscore change.

## Baseline

Spot check instead of the full grid, since nothing behavioural moved: Onyxia, 1,000 iterations, default seed.

| request | pass 8 score / SafeDPS / DTPS | pass 9 |
|---|---|---|
| shaman_warden_balanced | 874 / 1448 / 553 | 874 ± 37 / 1448 / 553 (identical) |
| druid_bear_balanced | 1495 / 1460 / 485 | 1495 ± 31 / 1460 / 485 (identical) |

Bit-identical, as expected from a rename. ref/tank_bench/summary.tsv (pass 8, 707ae92cc) stays the baseline.

## Changes tried

None: no rotation-relevant input changed, and pass 8's tests stand.

## Findings

No new findings. Open: 20261111-1 (Warden re-optimization after the Nov 10 shaman tuning, worker fix),
20261110-1 (Windfury Totem and Maul, user ruling), 20261104-11 and 20261104-3 (Broodlord-timer calls, user).

Note for the next pass: a fresh session needs `protoc -I=./proto --go_out=./sim/core ./proto/*.proto` (with
`$(go env GOPATH)/bin` on PATH) before `go build ./tools/tankbench`.
