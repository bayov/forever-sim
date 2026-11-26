# tank_fairness pass 34: 2026-11-26, master ab3fff7007

Due: 38 watched files changed since bfbee78. They are the warrior leveling batch (565f32155, 270731faa, 143bf4f72,
8f5b59325, 1223e764e and others): warrior spells take the rank known at the character's level (`knownRank`; at 60 the
same top ranks as before: Shield Slam 23925, Revenge 11601, Sunder 11597, Thunder Clap 11581), Defensive Stance only once
known, and core rules that apply only inside a leveling session (swings from behind can't be avoided, Cannibalize,
kill hooks). Stunned targets can no longer block, dodge or parry (`sim/core/spell_outcome.go`); no tank preset is
stunned on the six scoring bosses. No change to `tools/tankbench`, its requests or `tools/encounters`.

## Method

tankbench built in the master worktree (/tmp/audit-cache/cli/ab3fff700700/tree, `--tags=with_db`), `bench.sh` on the
12 preset requests x six scoring bosses, 3,000 iterations, seed 1; `summary.py --tsv` diffed against
ref/tank_fairness/summary.tsv: all 84 rows (72 boss rows and 12 means) identical. Baseline kept.

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
