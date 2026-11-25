# Rotations pass 78 (2026-11-25, master 5536db4bc)

Result: quiet. Range 226ada86b..5536db4bc (6 commits).

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py 226ada86b 5536db4bc` | 0 APL files touched |
| Non-test sim `.go` changed | none |
| `tools/dps_report/req/*.json` changed | none |
| Watched-file trigger `sim/compsearch/drops.json` | compsearch drop table regenerated after the trinket-line fix (86fbcf456); not read by the raid presets' rotations |
| Other changes | DPS Report assumption settings (assumptions_def.py, asm_data.py, ui data), Phase 0 prices brief, audit state/reports |

## Open findings

Unchanged: rotations-20261122-1 (Shock Ret ordering) and rotations-20261112-1 (rogue poison hands) have not landed
on master (no APL or request change in range).

## Open items

Carried over from passes/rotations.md "Open items and pending rulings"; none advanced this pass.
