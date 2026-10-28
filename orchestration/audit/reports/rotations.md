# Rotations pass 37 (2026-10-28)

- **Master:** 1274e15b3 (last audited 224b8a4e0). Client 1.60.1.70630 data; the change is a server hotfix.
- **Result:** 1 new finding (low, report wording). No open findings carried.

## What changed

- 73cedc9a9, user ruling 2026-10-28 (client_update-20261028-2, resolved there): the launch realms' 2026-10-28
  hotfix. Trueshot Aura 15 / 22 / 30 / 40 / 50 (rank 5 was assumed 100); Aspect of the Hawk ranks 1-5
  lowered to 15 / 22 / 30 / 38 / 47, so a level 60 hunter now casts rank 6 (14322, 55 RAP) instead of rank 5 (90).
- `apl_logic_diff.py 224b8a4 1274e15`: the 6 BM/MM/LW (and `_turret`) APLs swap 14321 for 14322 in the
  pre-pull at -3 s and the keep-up line, nothing else. SV keeps Aspect of the Beast. The reqs carry 14322;
  `tools/rotopt/hunter_ranged.go` too. No 14321 left in any APL or req.

## Checks

- **Baseline** (`baseline.py`, 3,000 it, seed 424242, shared CLI at 1274e15b3) against pass 36's:

| preset | pass 36 | pass 37 | change |
|---|---:|---:|---:|
| hunter__beast-mastery | 1150.43 | 1132.03 | -18.40 |
| hunter__marksmanship | 1137.21 | 1115.72 | -21.49 |
| hunter__lone-wolf | 1119.56 | 1094.83 | -24.73 |
| hunter__survival | 1029.17 | 1028.22 | -0.95 |

  Nothing else moved by more than 0.05. The drops fit -50 RAP (Trueshot) and -35 RAP (Hawk); LW loses most
  since all of its damage is the hunter's. New baseline: `ref/rotations/baseline_1274e15_3000it_seed424242.tsv`.
- **Aspect choice still right:** Hawk (55 RAP + Deadly Aspects Quick Shots on Auto Shot) against Beast (110
  melee AP, Quick Strikes on melee white hits only) for the ranged specs: Beast gives a ranged hunter
  nothing, so no A/B was needed. Rank 6's mana (110 against 90) is one cast a fight.
- **Report text:** the BM/MM/LW caveats still describe the old assumptions (rotations-20261028-1).
- `vs_published.txt`: 0 of 58 moved more than 0.5 DPS, so the published numbers already carry the hotfix.

## Optional

- None new. The pass 26 list stands (`passes/rotations.md`).
