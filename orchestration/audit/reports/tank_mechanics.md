# Tank mechanics, pass 22

- Master: 8f388c7f3c (from eebd00b2bd). Client: 1.60.1.70959, unchanged (wago latest, 2026-11-19).
- Result: quiet.

## What changed (watched files)

| Commit | File | Tank effect |
|---|---|---|
| 65f0ea9d6 | sim/druid/healing.go | `swiftmendHot` now walks `Rejuvenation` then `Regrowth` with an inline nil skip instead of `balanceSpells(...)` (sim/druid/talents.go:82, Flatten + nil filter). Same order, same nil filter, same "least time left" pick: no behaviour change. Restoration-only; no tank preset casts Swiftmend. |

Other commits in range (d82ab73d8, 4654ee8f5, d67d3eab7, 8f388c7f3, 1c328d8c5) are healing-race tooling,
DPS Report data and the audit loop; none touches a tank ability, talent or engine file.

## Checks

- Client build and ledger: 1.60.1.70959 still latest; no `orchestration/update_ledger.md` change in range.
  No verdict in `ref/tank_mechanics/verdicts.md` reopened.
- Tests: `go test -tags=with_db ./sim/druid/... ./sim/warrior/... ./sim/paladin/... ./sim/shaman/... ./sim/core/
  -run 'Tank|Bear|Feral|Swiftmend|Threat|Warden|Protection'`: all 13 packages ok.
- Shared comparison (`runs/vs_published.txt`): no tank rows move.

## Findings

None new. Open findings unchanged: 20261117-1, 20261108-1, 20261107-1.
