# tank_rotations pass 18

Master 8219f1083 (2026-11-18), against pass 17's baseline (master 2f7e255af). 1,000 iterations, default seed, every
tank request on all seven boss files. Quiet: no new finding.

## What changed

Since c198d81e3 (gate: sim/druid/hurricane.go). Tank-relevant diff: `sim/core/dot.go` (engine) only; the rest is
Hurricane, Volley and their tests.

- `dot.go` adds `HasteReducesDuration` and Forever's hasted-tick rule (`hastedTicks`). For a dot without
  `AffectedByCastSpeed` it returns the base tick length and 0 extra ticks, as before. Only Hurricane and Volley set
  the flag (grep of `sim/`), and no tank rotation casts either. So no tank code path should move.

## Baseline

`ref/tank_bench/summary.tsv` rewritten for 8219f1083: all 104 data rows bit-identical to pass 17's (diff empty apart
from the header line). Confirms the reading above.

## Rotation lines

No tank APL, request, rotopt template, tank ability, talent or encounter changed, and the baseline didn't move, so
no new one-line A/B was warranted. Pass 17's checks still hold.

## Open findings

- 20261104-11 (warrior Max survival armor trinket, Broodlord): user call, unchanged.
- 20261104-3 (paladin Bulwark 13 s, Broodlord): user ruling, unchanged.

## Notes for the orchestrator

- The pass doc's open item (main-hand timing for Windfury) was done for all four tanks in pass 15; it can still be
  removed from passes/tank_rotations.md.
- `go build ./tools/tankbench` in the repo checkout failed on a missing `sim/core/proto`; built it instead in the
  shared CLI worktree (/tmp/audit-cache/cli/8219f1083fee/tree, protos generated), same commit.
