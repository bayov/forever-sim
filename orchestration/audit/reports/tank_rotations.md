# tank_rotations pass 21

Master e671e881e (2026-11-19), against pass 20's baseline (master 8f388c7f3). 1,000 iterations, default seed, every
tank request (13) on all seven boss files. Quiet: no new finding.

## What changed

Gate: 1 watched file, `sim/core/tank_score.go` (d73c17568, "Tank race sims reuse twin races and racial rows, faster
least-HPS search"). `lives(hps)` became `replayFrom(hps, from, h0)`, which also returns the last event before any
death where health was set to max health; `leastHps` restarts each later replay (higher HPS) from that checkpoint
instead of from event 0.

Read for correctness: each step (damage, stand-in heal units, real heals, the max-health clamp) is monotone in the HPS,
float rounding included, so a higher HPS is also at max health at the checkpoint and the replay from there is the
same, bit for bit. The checkpoint is only taken from a replay at an HPS below every one tried after it (the
`lo` side of the doubling and the bisection, and the HPS 0 replay for the +Inf test). The clamp rules are unchanged:
damage clamps to `min(maxHealth, health)` first, and heals clamp to the event's max health. No tank APL, request,
rotopt template, encounter file, tankbench or tankscore change since 8f388c7f3.

## Baseline

`ref/tank_bench/summary.tsv` rewritten for e671e881e. All 104 rows (91 boss rows plus 13 scoring means) match pass 20's
exactly in every column, healing needed and death floor included, which are the columns `leastHps` feeds. No rotation
lines tried: nothing a rotation reads changed.

## Open findings

- 20261104-11 (warrior Max survival armor trinket, Broodlord): user call, unchanged (warrior rows identical).
- 20261104-3 (paladin Bulwark 13 s, Broodlord): user ruling, unchanged (paladin rows identical).

## Notes for the orchestrator

- The pass doc's open item (main-hand timing for Windfury) was done for all four tanks in pass 15; it can still be
  removed from passes/tank_rotations.md.
- `tools/audit/tank_rotations/bench_summary.py` still writes a stale header (master f62bfc74); this pass fixed line 1
  of the tsv with sed afterwards.
- This pass's first attempt ran in a shell without `$AUDIT_WORK` set, which left a stray `/tank_rotations/` directory
  (a build log and empty bench outputs) at the filesystem root. A safety check blocked removing it; delete it by hand.
