# Rotations pass 68 (2026-11-17, master 5f20dfdf4)

Last audited master: 1a85df1d7. Result: quiet (no new finding). One finding still open
(rotations-20261112-1, rogue poison swap: not on master).

## What changed (1a85df1d7..5f20dfdf4, 23 commits)

`apl_logic_diff.py`: 0 APL files touched. No req rotation changed.

| commit | change | reaches presets? |
|---|---|---|
| ad1d7f469, 3e1214a69 | push 114097: Unbridled Wrath procs on damaging abilities (not Whirlwind); Cleave uses one Reckless Swings charge | every DPS warrior (+3 to +11) |
| 2ebdc14d6 | an avoided Execute keeps its extra Rage | warriors (execute phase) |
| b075ae64d | pet white swings take none of the hunter's crit | every hunter (BM -69 to -76, MM -30 to -33, SV -20 to -23) |
| 4a8541e46, eada6763e | Predator's Edge counts twice on Raptor Strike; a queued Raptor Strike keeps the DW miss penalty | SV |
| 29481dd0e | Immolation/Explosive Trap roll with spell hit | SV (trap hits more) |
| cf95c3e32 | Death Mark always hits, cap uses sheet crit | Subtlety +5.9 |
| 6ea6e47b4, 6875e3b1e, 39b94e364 | Manual Crowd Pummeler ICD, Five Thunders set, rotopt gear search | no preset moved |

`regress.py` (old CLI 1a85df1d7 vs 5f20dfdf4, 3000 it, seed 424242): 20 presets moved past 1 DPS, all
hunters, warriors and Subtlety, as above. Nothing else moved.

## Baseline

New `ref/rotations/baseline_5f20dfd_3000it_seed424242.tsv` (68 rows), replacing `baseline_25ce81d`.

## Re-tuning tested after the sim change

Unbridled Wrath on abilities gives warriors more Rage, so the Heroic Strike Rage gates might want to come
down. Paired `ab.py`, 20k it, S0 1e6, B-A at 120 / 180 / 300 s:

| preset | variant (APL index) | B-A | A at 180 |
|---|---|---|---|
| warrior__fury | HS gate 60 -> 50 (line 14) | -2.02 / -1.10 / -1.72 | 1101.0 |
| warrior__fury | HS gate 60 -> 70 (line 14) | -0.05 / +0.29 / -1.27 | 1101.0 |
| warrior__fury_2h | HS gate 90 -> 75 (line 16) | -3.07 / -5.75 / -4.75 | 1112.9 |
| warrior__arms | HS gate 90 -> 75 (line 16) | -2.72 / -2.86 / -1.08 | 1079.1 |

All lose or sit under the bar, so no second seed set was run. Current gates hold.

Hunters: not re-tested. The pet-crit loss doesn't touch the hunter's own priorities. Raptor Strike gained
(Predator's Edge) more than it lost (DW penalty), and the traps gained hit, so dropping either can only lose.

## Open items carried

- Windfury Totem ability timing (user request 2026-11-15) is still not done; see the pass doc.
- The other open items in `passes/rotations.md` are unchanged.
