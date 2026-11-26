# tank_rotations pass 32

Master ab3fff7007 (2026-11-26), against pass 31's master bfbee78fcc. Quiet: no new finding.

## What changed

The 38 watched files the gate listed are the warrior leveling work (143bf4f72, 8f5b59325, 30bb3a590, 1223e764e,
c747e7ccd, 270731faa, 565f32155 and their tests): every rank at the level it's learned, leveling hooks (charges,
stuns, kills, Rage decay, abilities that reach only nearby mobs), Rage from hits taken by stance in leveling
sessions, and `ui/tank_warrior` leveling plan, weights and presets. `sim/core/debuffs.go` and `target_death.go`
changed with them. No tank request, level 60 UI apl, encounter, tankbench, tankscore or rotopt tank template
changed. The `ui/tank_warrior/presets.ts` and `sim.ts` diffs add only the leveling plan and its EP weights.

## Baseline (1,000 iterations, default seed; `ref/tank_bench/summary.tsv` rewritten)

All 104 rows (13 requests x 7 bosses, plus scoring means) are bit-identical to pass 31 on score, SafeDPS, TPS,
DTPS and healing needed, so the leveling code doesn't reach the level 60 tank sims. Scoring means:

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
20261104-11 and 20261104-3 Broodlord user calls) are untouched on master and get a pass 32 note.

## Open items

- Windfury main-hand ability timing (pass doc, user request 2026-11-15): not done this pass either. It needs APL
  conditions on the swing timer for each tank request that takes the totem; left for a pass with room for it.

## Setup note

A fresh session needs `make sim/core/proto/api.pb.go` (Go protos) before `go build -tags=with_db ./tools/tankbench`;
`tools/audit/setup.sh` installs protoc and protoc-gen-go but doesn't generate them. The full bench (91 runs,
4 parallel) takes about 25 s.
