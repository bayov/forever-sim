# tank_rotations pass 20

Master 8f388c7f3 (2026-11-19), against pass 19's baseline (master a6677362f). 1,000 iterations, default seed, every
tank request (13) on all seven boss files. Quiet: no new finding.

## What changed

Gate: 1 watched file, `sim/druid/healing.go` (65f0ea9d6). `swiftmendHot` now loops over the Rejuvenation and Regrowth
lists directly, skipping nil entries, instead of building a filtered slice with `balanceSpells` (talents.go:82, which
does Flatten + nil filter). Same iteration order, same tie rule, so the result is unchanged. No tank casts Swiftmend.
No tank APL, request, rotopt template, encounter file, tankbench or tankscore change since a6677362f.

## Baseline

`ref/tank_bench/summary.tsv` rewritten for 8f388c7f3. All 104 rows (91 boss rows plus 13 scoring means) are
bit-identical to pass 19's in score, SafeDPS, DTPS and healing needed. No rotation lines tried: nothing a rotation
reads changed.

## Open findings

- 20261104-11 (warrior Max survival armor trinket, Broodlord): user call, unchanged (warrior rows identical).
- 20261104-3 (paladin Bulwark 13 s, Broodlord): user ruling, unchanged (paladin rows identical).

## Notes for the orchestrator

- The pass doc's open item (main-hand timing for Windfury) was done for all four tanks in pass 15; it can still be
  removed from passes/tank_rotations.md.
- `tools/audit/tank_rotations/bench_summary.py`: its header line now says master 8f388c7f3, pass 20 (the next pass
  edits it the same way, or passes the sha in).
- The first tankbench build failed with a missing `sim/core/proto` package, likely while another pass was regenerating
  the protos; a rebuild a minute later worked.
