# Rotations pass 61 (2026-11-12, master 8306f1d52)

Last audited master a0b4d3be9. Two commits reach the raid presets:

| commit | change | reaches raid presets |
|---|---|---|
| 368b2e181 | Hunter pet swings at its owner's melee and attack speed under Forever (`Pet.FollowsOwnerMeleeSpeed`, `sim/core/pet.go`); pet autocast passes an ability that is up but short of focus | yes (every hunter with a pet) |
| c380bbaef | Rake's direct hit adds 6.5% of attack power, outside the Nov 2 scale (`sim/druid/rake.go`) | yes (feral) |

Other watched files (`tools/rotopt/*`, `sim/testdata/report_hashes.txt`) don't reach the presets.
`apl_logic_diff.py a0b4d3be9 8306f1d52`: 0 APL files. No req file, preset or APL changed.

## Baseline move (3000 it, seed 424242, CLI 8306f1d52875)

New baseline `ref/rotations/baseline_8306f1d_3000it_seed424242.tsv`; every other preset unchanged (|diff| < 0.05).

| build | a0b4d3b | 8306f1d | move |
|---|---|---|---|
| druid__feral-cat | 1123.40 | 1152.74 | +29.34 |
| druid__feral-cat-duty-ff | 1118.07 | 1147.41 | +29.34 |
| hunter__survival / -tallstrider / -wolf | 986.8 / 956.2 / 961.8 | 1015.1 / 981.0 / 990.0 | +28.3 / +24.8 / +28.2 |
| hunter__beast-mastery / -tallstrider / -wolf | 1174.9 / 1117.9 / 1131.1 | 1182.5 / 1127.4 / 1140.9 | +7.6 / +9.5 / +9.8 |
| hunter__marksmanship / -tallstrider / -wolf | 1233.1 / 1200.0 / 1208.0 | 1234.0 / 1204.4 / 1212.7 | +0.9 / +4.5 / +4.7 |
| hunter__lone-wolf | | | 0 (no pet) |

## A/B (paired, 20k it, 8 blocks, both seed sets)

Feral cat (Rake hit now scales with AP):

| variant | 120 / 180 / 300 s pooled | verdict |
|---|---|---|
| Ferocious Bite 5-CP gate 45 -> 65 | +2.12 / +2.74 / +4.59 (duty-ff +1.80 / +1.99 / +3.88) | open finding rotations-20261105-1 grows (was +1.9); updated |
| Ferocious Bite gate 75 (set 1) | +0.28 / +1.28 / +4.16 | worse than 65 |
| Rake end-of-fight gate 6 s -> 3 s | +0.52 / +0.31 / +0.22 | optional, under the bar |
| Rake end-of-fight gate removed (set 1) | -0.06 / -0.09 / -0.04 | no |

Hunters: the pet change is not a player-rotation lever (Rapid Fire already fires on cooldown and now hastes the
pet as well). Not A/B-screened this pass; open item if a Rapid Fire hold/sync with pet cooldowns is wanted.

## Optional

- Feral Rake end-of-fight gate 3 s (about +0.3 at 180 s).
- Earlier optional list (pass doc) unchanged.

## Open findings

None landed on master since pass 60 (no req/APL/preset diff): rotations-20261112-1 (rogue poison swap, high),
rotations-20261109-1 (Fury 2H 90-rage HS, low), rotations-20261105-1 (feral FB gate 65, low, updated).
