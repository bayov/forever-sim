# Rotations pass 70 (2026-11-18, master a6677362f)

Last audited master: 8219f1083. Result: quiet (no new finding). One finding still open
(rotations-20261112-1, rogue poison swap; no req changed in this range).

## What changed (8219f1083..a6677362f, 46 commits)

`apl_logic_diff.py`: 0 APL files touched. No `tools/dps_report/req/*.json` change.
`standards.md`: rule 27 rewritten (haste shortens a spell's GCD, 1s floor).

| commit | change | reaches presets? |
|---|---|---|
| ee35da473, c2efe4ae3 | Spell haste shortens a spell's GCD to 1.5/(1+haste), floor 1s | casters and paladins (see baseline) |
| c8b5275b1, cc5a55f47 | report patches folded into the code as switches; faster start-up | no (melee/hunter/rogue unchanged at 3000 it) |
| 7199c8f9c, bbfd014e4 | healer idle-skip wake, healbench speedups | healers only |

## Baseline (3000 it, seed 424242)

`baseline.py` on a6677362f vs `baseline_2f7e255`: 31 of 68 moved past 0.05 DPS, every move
fits the hasted GCD. Warlocks +0.3 to +0.9%, Balance +0.3 to +0.6%, Elemental +0.4%, Frost +0.2%,
paladins -0.1 to -0.4% (the open paladin re-timing item, user ruling 2026-11-18). Warriors,
rogues, hunters, feral, Enhancement (within 0.8), Fire, priests: unchanged.

| preset | 2f7e255 | a667736 | delta |
|---|---|---|---|
| warlock__destruction-duty-curse | 1051.6 | 1060.9 | +9.3 |
| warlock__affliction | 1071.0 | 1076.8 | +5.8 |
| druid__balance-duty-ff | 916.1 | 921.3 | +5.1 |
| shaman__elemental-level60 | 976.9 | 980.7 | +3.8 |
| paladin__level_60_sub_holy | 1012.4 | 1007.9 | -4.5 |
| paladin__level_60_shock_ret-duty-crusader | 1009.6 | 1006.4 | -3.2 |

New baseline: `ref/rotations/baseline_a667736_3000it_seed424242.tsv` (old one removed).
No `regress.py` / `ab.py` needed: no APL change, and the only sim change reaching presets is the ruled GCD rule.

## Note for the dps-report pass (not a rotations finding)

`runs/vs_published.txt` shows Affliction +4.1%, rogues -1.1 to -2.1%, Lone Wolf -1.7% against the
published report, but the rotation baseline shows none of those presets moved in this range beyond
the GCD effect (Affliction +0.5%, rogues/Lone Wolf 0). The gap is the report's switches/assumptions
vs the plain CLI, not a rotation change.

## Not done this pass

- Caster rotations (warlock, Balance, Elemental, mage) weren't re-searched for the hasted GCD; no
  ruling asks for it yet (only paladins have an open re-timing item).
- Windfury ability-timing item: unchanged since pass 69 (Fury BT, Combat SS, Arms MS null); remaining
  abilities still to test.
