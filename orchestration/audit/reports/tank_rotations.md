# tank_rotations pass 33

Master 2af970d748 (2026-11-26), against pass 32's master ab3fff7007. Quiet: no new finding.

## What changed

The 8 watched files the gate listed are warrior leveling only: `ui/tank_warrior/apls/leveling_plan_*.json` (five
plans), `leveling_weights_shield_fury_prot.json`, and `presets.ts` / `sim.ts`, which add the leveling builds (Prot as
the default plan, 48902239f, c377aaf95) and the Fury Prot shield EP weights. The rest of the range is Phase 0 DPS
re-rank work (non-tank requests, rotopt mksettings for Enh/Feral/Hunter, `sim/common/guardians/barov_servant.go`).
No tank request, level 60 UI apl, encounter, tankbench, tankscore, rotopt tank template or tank sim code changed.

## Baseline (1,000 iterations, default seed; `ref/tank_bench/summary.tsv` rewritten)

All 104 rows (13 requests x 7 bosses, plus scoring means) are bit-identical to pass 32 on score, SafeDPS, TPS,
DTPS and healing needed, so nothing in the range reaches the level 60 tank sims. Scoring means:

| request | weight | score | SafeDPS | TPS | DTPS | healing needed |
|---|---|---|---|---|---|---|
| druid_bear_balanced | 1 | 3338.3 | 2309.0 | 2372.5 | 586.7 | 469.5 |
| druid_bear_max_survival | 3 | 1260.7 | 1754.0 | 1793.7 | 556.2 | 283.5 |
| druid_bear_max_threat | 0.333 | 4541.7 | 2558.5 | 2557.2 | 613.7 | 636.3 |
| paladin_protection_balanced | 1 | 2857.7 | 1720.8 | 1530.2 | 302.7 | 233.7 |
| paladin_protection_max_survival | 3 | 1858.0 | 1639.0 | 1456.8 | 283.8 | 196.7 |
| paladin_protection_max_threat | 0.333 | 3280.3 | 1749.7 | 1557.2 | 302.8 | 261.8 |
| shaman_warden_balanced | 1 | 1245.7 | 1534.3 | 1365.2 | 660.0 | 660.3 |
| shaman_warden_max_survival | 3 | -2234.0 | 1482.5 | 1336.2 | 670.8 | 630.3 |
| shaman_warden_max_threat | 0.333 | 2608.2 | 1697.0 | 1513.5 | 683.0 | 851.3 |
| warrior_protection | 1* | -399.2 | 854.5 | 1206.5 | 636.2 | 755.5 |
| warrior_protection_balanced | 1 | 3241.2 | 2108.8 | 1989.0 | 374.8 | 344.2 |
| warrior_protection_max_survival | 3 | 1617.0 | 1697.3 | 1687.2 | 321.0 | 213.0 |
| warrior_protection_max_threat | 0.333 | 4102.7 | 2281.2 | 2125.0 | 418.3 | 469.8 |

## Rotation lines

No rotation changed and the baseline is identical, so pass 31's line checks stand; no A/B re-run. The five open
findings (20261125-1 bear Roar, 20261125-2 warrior Balanced Thunder Clap, 20261122-1 paladin Consecration floor,
20261104-11 and 20261104-3 Broodlord user calls) are untouched on master and get a pass 33 note.

## Open items

- Windfury main-hand ability timing (pass doc, user request 2026-11-15): not done this pass either. It needs APL
  conditions on the swing timer for each tank request that takes the totem; left for a pass with room for it.

## Setup note

A fresh session needs `make sim/core/proto/api.pb.go` (Go protos) before `go build -tags=with_db ./tools/tankbench`;
`tools/audit/setup.sh` installs protoc and protoc-gen-go but doesn't generate them. The full bench (91 runs,
4 parallel) takes about 15 s at full parallelism.
