# tank_rotations pass 25

Master 5518f4873f (2026-11-23), against pass 24's baseline (master 7a6d1beb9d). 1,000 iterations, default seed,
every tank request (13) on all seven boss files (91 runs). Quiet: no new finding.

## What changed

Gate: 1 watched file (7a6d1beb9d .. 5518f4873f):

- `sim/core/attack.go` (79c688d0c): `WeaponAttack.swing` now parks a weapon with `SwingSpeed <= 0` (an empty slot,
  e.g. a hunter without a ranged weapon) at `NeverExpires` instead of swinging forever in the same instant. Every
  tank swings a real main hand (bear form has its own fixed speed) and no tank's off hand auto-attacks, so no tank
  path reaches the new branch.
- Nothing else watched moved: no tank apl, request, rotopt template, encounter, `tools/tankbench` or
  `tools/tankscore` change; the only `sim/warrior` change is a DPS-warrior test file.

## Baseline

All 104 rows of `ref/tank_bench/summary.tsv` (13 requests x 7 bosses, plus each request's scoring mean) are
byte-identical to pass 24 in every column (score, error, SafeDPS, TPS, DTPS, healing needed, death floor, TMI,
death %, biggest hit, DPS). Only the header line changed. Expected: the guard draws no RNG and no tank hits it.

## Rotation lines

No rotation changed, and the engine change can't shift any line's value, so no new one-line A/B this pass. The
three open findings (20261122-1, 20261104-11, 20261104-3) stand on unchanged numbers.

## Open items carried

- Windfury main-hand ability timing (user request 2026-11-15, pass doc Open items): still not done. It needs `ab.py`
  runs on both seed sets for every totem-taking tank request; out of scope for this gate-only pass.

## Notes

- First `go build` of tankbench failed with "no required module provides package .../sim/core/proto"; a retry a
  moment later built cleanly (likely a concurrent proto regeneration in the tree). Not a finding.
- `tools/audit/tank_rotations/bench_summary.py` writes a placeholder header line now; edit line 1 of the tsv
  after running it.
