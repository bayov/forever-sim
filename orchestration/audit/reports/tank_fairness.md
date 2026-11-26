# tank_fairness pass 35: 2026-11-26, master 2af970d748

Due: 8 watched files changed since ab3fff7007, all warrior leveling (48902239f, c377aaf95, 6a447e231): leveling plan
APLs, Fury Prot shield weights, leveling presets in `ui/tank_warrior/presets.ts` (the raid tank presets unchanged), a
`weights_build` proto field, and a Barov servant guardian swing guard. No change to `tools/tankbench`, its requests,
`tools/encounters` or tank mechanics.

## Method

tankbench built in the master worktree (/tmp/audit-cache/cli/2af970d7489d/tree, `--tags=with_db`), `bench.sh` on the
12 preset requests x six scoring bosses, 3,000 iterations, seed 1; `summary.py --tsv` diffed against
ref/tank_fairness/summary.tsv: all 84 rows identical (the legacy warrior_protection request, finding 20261104-5, is
outside the baseline). Baseline kept.

## Ranking (unchanged from pass 33)

| weight | order | scores |
|---|---|---|
| Max threat | Bear, Warrior, Paladin, Warden | 4537 / 4097 / 3282 / 2606 |
| Balanced | Bear, Warrior, Paladin, Warden | 3342 / 3251 / 2863 / 1244 |
| Max survival | Paladin, Warrior, Bear, Warden | 1875 / 1617 / 1301 / -2277 |

TPS (Balanced): bear 2371, warrior 1987, paladin 1531, Warden 1367. Healing needed (Balanced): paladin 233, warrior 341,
bear 468, Warden 664. The explanations and the comparison with the Anniversary logs in pass 33 stand
(orchestration/fairness_tanks.md).

## What each preset rewards

Unchanged from pass 33: every preset wins its own weight except the bear's Balanced, which ties its Max threat build
at w 1 (finding -20261125-1). The three builds remain distinct for every spec.

## Same terms

Requests unchanged; the open unequal terms stand: -20261105-1 (paladin potion), -20261105-2 (Warden races),
-20261124-1 (Stoneform on paladin and Warden Dwarf rows).

## Findings

- New: none.
- Open, unchanged: -20261125-1, -20261124-1, -20261105-2, -20261105-1, -20261104-2, -20261104-5.
