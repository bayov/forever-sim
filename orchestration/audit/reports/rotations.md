# Rotations pass 58 (2026-11-10, master ca5e35c3f)

Last audited master 91ad19535. 10 commits; one reaches sim code the raid presets use:

| commit | change | reaches raid presets |
|---|---|---|
| bcc5d6cd4 | Maelstrom Weapon gives a stack on every landed melee hit under Forever (was 4 PPM a point); Totem of the Storm 50% a bolt | yes (Enhancement) |
| bcc5d6cd4 | Flametongue Weapon rank 5 takes 7 off its points | no (rank 5 is a level 47 rank; L60 uses rank 6, cut 0) |
| 04bdcdb9e, a778e524b | enchants 41/66 client values, items seen in game | not in the L60 baseline (no move outside Enhancement) |
| others | log audit, logcheck, audit loop | no |

`apl_logic_diff.py 91ad19535 ca5e35c3f`: 0 APL files touched. Req files unchanged.

## Baseline move (3000 it, seed 424242, same reqs)

New baseline `ref/rotations/baseline_ca5e35c_3000it_seed424242.tsv` (replaces `baseline_66a243c_...`).
Only three presets moved by more than 1 DPS, all within noise of a 3000-iteration run:

| preset | old | new | diff |
|---|---|---|---|
| shaman__enhancement-level60 | 1147.3 | 1144.7 | -2.6 |
| shaman__enhancement-mana-tide | 1040.2 | 1041.6 | +1.4 |
| shaman__enhancement-water-shield | 1098.8 | 1097.7 | -1.1 |

At 5/5 the old 20 PPM already filled the stack before the Lightning Bolt line came up, so the
every-hit rule changes little at Level 60. (The -14 against the published report in
`vs_published.txt` predates 91ad19535; it isn't from this commit.)

## Lightning Bolt re-screen under the new Maelstrom rule

`ab.py`, shared CLI ca5e35c3f219, shaman__enhancement-level60, 20k it, 8 blocks, S0 1e6, 180 s only (screen):

| variant | B-A at 180 s |
|---|---|
| LB line (`level60` req index 12) above Flame Shock | +0.28 ± 0.42 |
| LB above Stormstrike | -2.40 ± 0.61 |
| LB at 4 stacks | -29.38 ± 0.64 |
| LB at 3 stacks | -72.15 ± 0.61 |

The 5-stack gate and its place hold. Nothing reached the bar, so the second seed set wasn't run.

## Open findings

- rotations-20261109-1 (Arms/Fury 2H HS 90 line) and rotations-20261105-1 (feral FB gate 65): not
  landed (no APL change on master). Still open.

## Optional

- None new. Mana Tide and Water Shield builds weren't re-screened for the bolt; they share the gate.
