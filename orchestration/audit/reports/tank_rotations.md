# tank_rotations pass 22

Master e07e8fdb8 (2026-11-20), against pass 21's baseline (master e671e881e). 1,000 iterations, default seed, every
tank request (13) on all seven boss files. Quiet: no new finding.

## What changed

Gate: 3 watched files, all from the BiS / item-conditions work (617a63635 .. e07e8fdb8):

- `sim/core/target.go`: `Encounter.OutdoorTerrain` and `Target.MobFamily` are copied from the proto. New fields only;
  nothing in the tank path reads them, and the boss files don't set them.
- `sim/paladin/items.go`: `RegisterItemConditions(TenetsOfTheSilverHand, Undead)`, a tag for the BiS ranker. The
  item's effect is unchanged.
- `sim/shaman/items.go`: Totem of Rage under Forever is now +2% shock damage (additive multiplier) instead of +30
  bonus damage. No tank request equips it (27859 appears in no `tools/tankbench/req/*.json`).

No tank APL, request, rotopt tank template, encounter file, tankbench or tankscore change since e671e881e.

## Baseline

`ref/tank_bench/summary.tsv` rewritten for e07e8fdb8. All 104 rows (91 boss rows plus 13 scoring means) match pass
21's exactly in every column. No rotation lines tried: nothing a rotation reads changed.

## Open findings

- 20261104-11 (warrior Max survival armor trinket, Broodlord): user call, unchanged (warrior rows identical).
- 20261104-3 (paladin Bulwark 13 s, Broodlord): user ruling, unchanged (paladin rows identical).

## Notes for the orchestrator

- `tools/audit/tank_rotations/bench_summary.py` now takes `<master sha> <pass number>` and writes the right header;
  no more sed on line 1.
- The pass doc's open item (main-hand timing for Windfury) was done for all four tanks in pass 15; remove it from
  passes/tank_rotations.md.
- The stray `/tank_rotations/` directory at the filesystem root (a shell without `$AUDIT_WORK`) is still there;
  delete it by hand. Each Bash call is a fresh shell, so source setup.sh in every call that uses `$AUDIT_WORK`.
- The first tankbench build failed with "no required module provides package .../sim/core/proto" while another
  auditor was regenerating the protos at the same moment; a rebuild a minute later worked.
