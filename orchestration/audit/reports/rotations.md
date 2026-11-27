# Rotations pass 87 (2026-11-27, master f4c45c3f9)

Result: quiet. Range 5ab589d85..f4c45c3f9: Phase 0 re-gear of 29 report requests (94d6cc79f, 48cc3aed9), Talisman of
Ascendance comment (8d45b49e6), rogue filler talent points to utility (447e6e8f1, 04631ff49), report republish.

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py 5ab589d85 f4c45c3f9` | 0 APL files touched |
| Non-test sim `.go` changed | `sim/common/item_effects.go`: comment lines only (Talisman charges come from the tooltip) |
| Req rotations changed | 10 warlock reqs: item lines swap 22678 for the worn 13968 / 10455; each req's priorityList now equals its `ui/warlock/apls/level60_*.apl.json` (spot-checked 5) |
| Warlock trinkets worn vs APL lines | Demonology wears 13968+10455, lines name 10455; Destruction / Fire wear 13968+272438, lines name 13968; Affliction wear 272438+10455, autocast only (Shadow Bolt variant names 10455) |
| Talisman 22678 anywhere | no req or warlock APL names it now |
| Rogue talents | Deflection / Riposte / Setup filler moved to utility; no rotation line depends on them |
| Shared run vs published | all within 1.5 DPS (largest: lone-wolf -1.5, fury_2h-duty-sunder +1.3) |
| Open findings | 20261127-1 resolved as moot (Talisman no longer worn). 20261126-3, 20261122-1, 20261112-1 still open, untouched |
| Baseline | not refreshed (`baseline_ab3fff7_3000it_seed424242.tsv`); no sim logic change |

## Optional

None this pass.
