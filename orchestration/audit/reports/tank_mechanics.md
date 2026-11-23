# tank_mechanics: pass 28

- master: 5518f4873f (last pass 7a6d1beb9d), 2026-11-23
- result: quiet
- client: 1.60.1.71037 (latest on wago.tools, unchanged since pass 25)

## What changed (watched paths)

| commit | file | change | reaches tanks? |
|---|---|---|---|
| 79c688d0c | sim/core/attack.go | `WeaponAttack.swing` sets `swingAt = NeverExpires` when `SwingSpeed <= 0` (empty slot), so a hunter with no ranged weapon no longer loops at one instant | No |

Why not: a tank's shield off-hand has SwingSpeed 0, but `IsDualWielding` is false for it, so the
off-hand swing is never scheduled (attack.go:544, :717) and this branch is never reached. Tank main
hands and the boss's weapon have SwingSpeed > 0. A weapon swap re-sets `mh.swingAt` (attack.go:1056),
so no slot stays parked at NeverExpires after one. Engine behaviour for tanks (attack table, Rage
from damage, threat, block, tank sheet, score replay) untouched.

## Tests

`go test -tags=with_db ./sim/warrior/... ./sim/paladin/... ./sim/druid/tank/... ./sim/shaman/warden/... ./sim/core/`: all pass.

## Open findings (unchanged)

- tank_mechanics-20261117-1 (low, user ruling): Unbridled Wrath skips Sunder Armor.
- tank_mechanics-20261108-1 (low, worker fix): no test pins Warden shock threat under Stormstrike.
- tank_mechanics-20261107-1 (low, worker fix): stale "client disagrees" labels (warrior, bear, Warden).
