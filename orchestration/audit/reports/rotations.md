# Rotations and resources: pass 32 (2026-10-26T00:58:23Z)

Master 74b3a5f7c (last audited 016daa654). Result: quiet, 0 new findings.

## What changed

- 107874b75: white-hit Rage is flat with level (`ForeverNormalizedRageScaling` 0.5 -> 0, user ruling 2026-10-25).
- 44fdd8599: warrior Level 60 rotations and talents re-searched at flat Rage (Fury DW: HS from 80 and in execute,
  Bloodrage <20, Death Wish / trinket end holds; 2H: Berserker Rage <30, HS 90, Slam 45; Arms: Bloodrage <20,
  Hamstring 30). Duty variants carry the same knobs.
- 01afcb1d3: hunter crit per Agility is Classic's at 60 (scale 1.22 -> per-level table, 1 at 60).
- 92484f3c8: Forever base stats for Dwarf, Night Elf, Gnome, High Order.
- cdceb56c4: MM, LW and mm_turret drop the Aimed Shot line (Multi-Shot takes the shared cooldown); MM plays a Human.
  This supersedes the pass doc's settled note "Multi-Shot never fires in BM, MM or LW" for MM/LW (BM unchanged).

## Checks

| Check | Result |
|---|---|
| `apl_logic_diff.py 016daa654 74b3a5f7c` | 10 files: 7 warrior (knob values as in the commit), 3 hunter (Aimed Shot line removed). |
| Embedded req rotations vs APL files (7 warrior, MM, LW) | All identical. |
| `baseline.py` vs `baseline_016daa6` | Warrior -51 to -85, hunter -27 to -30 (the two rulings); feral -1.5, Subtlety -1.1/-1.2 (Night Elf base stats); frost mage +0.2. Nothing unexplained. New baseline `baseline_74b3a5f_3000it_seed424242.tsv`. |
| Shared run vs published | Warrior and hunter rows match published; only feral/Subtlety moved (-1.0 to -1.5, base stats). |
| Open findings to verify | None open. |

## Verification A/B (paired, 20k it, 8 blocks; B-A at 120 / 180 / 300 s)

| Preset | B variant | s0=1e6 | s0=2e8 | Verdict |
|---|---|---|---|---|
| hunter__marksmanship (A 1139.3 @180) | Aimed Shot line restored above Multi-Shot | -3.40 / -1.68 / -1.01 | -2.63 / -1.94 / -0.61 | master's Multi-Shot holds |
| warrior__fury (A 984.8 @180) | HS ≥80 gated off in execute (old form) | -0.24 / -1.10 / +0.04 | -0.20 / -1.03 / +0.04 | master's HS-in-execute holds |

Knob-level moves (thresholds, holds) were searched by the worker with tools/rotopt on paired seeds and are
not re-searched here.

## Optional (under the bar)

None new.

## Open items

- Pass doc "Settled" note on Multi-Shot is stale for MM/LW (orchestrator to update passes/rotations.md).
- Carried from earlier passes: Balance Night Elf + Elune's Light, Elemental Mastery hold, warlock duty-curse screens.
