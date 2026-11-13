# Healer rotations pass 12 (quiet: no new findings; -20261104-2 still open)

- Master: 3a86d4a6b (`last.master` was cad23ae1a). Audited 2026-11-13.
- Bench: `tools/healbench` built from this checkout, default seed, `-carry`, 1,000 iterations, each request on each
  of the 7 boss files at its own length in its role (same rows as `tools/audit/healer_rotations/baseline.sh`, run in
  parallel into a private directory).

## What changed since cad23ae1a

- Gate: one watched file, `sim/shaman/flametongue_weapon.go`: Flametongue Weapon rank 6 now has a 5.5 point cut
  (`FlametongueWeaponCut`), from the Nov 12-13 logs. Weapon imbue damage only; neither Restoration shaman request
  imbues Flametongue or melees in the bench.
- Also on master, not in the gate: `sim/common/item_effects.go` (Frightskull Shaft 3.9 PPM and Forever disease),
  `sim/core/buffs.go` (Windfury Weapon attacks proc Windfury Totem), `sim/core/consumes.go` (Forever spell IDs for
  Shadow/Frost Oil procs). All melee or weapon procs.
- No healer APL, healbench request, encounter file or `tools/healbench/main.go` changed.

## Baseline

All 63 rows (9 requests x 7 boss files) are identical to `ref/healer_bench/summary.tsv` in every column (carry,
EHPS, overheal, EH/mana, mana left, OOM time, deaths). Max carry delta 0. The reference is unchanged, so it was not
rewritten.

Process note: `tools/audit/setup.sh` left protoc-gen-go uninstalled in this session (silent), so
`make sim/core/proto/api.pb.go` was needed before the build. A second session ran baseline.sh into the shared
`$AUDIT_WORK/healer_rotations/summary.tsv` at the same time and the two runs interleaved, so this pass's numbers come
from a private copy.

## Rotation lines

No rotation changed and no input to them changed, so the pass 10 analysis stands (see git history of this file
for its tables and A/B screens). No new one-line changes were tried.

## Findings

- New: none.
- Open: healer_rotations-20261104-2 (Discipline tank healer OOM on Chromaggus at its own length and Golemagg at
  300 s): unchanged, as nothing in the Disc request or mana engine moved.
