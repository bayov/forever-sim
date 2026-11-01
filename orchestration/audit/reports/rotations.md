# Rotations pass 43 (2026-11-01)

- **Master:** dd053d24c (last audited d1935640f). Client 1.60.1.70707.
- **Result:** quiet. No open findings; none new.

## What changed

- `apl_logic_diff.py d1935640f dd053d24c`: 0 APL files touched. Rotations embedded in the reqs are
  unchanged.
- `sim/core/consumes.go` (bb21e3b): Goblin Sapper Charge and Goblin Land Mine get an
  `ExtraCastCondition` of `DistanceFromTarget <= 10` (blast radius). Thrown explosives unchanged.
- Reqs / mksettings (905787a and before): warlocks and priests move from 5 to 25 yd; ranged casters
  and ranged hunters swap Goblin Land Mine for EZ-Thro Dark Bomb; ranged hunters keep the Sapper,
  which now fires only when the Raptor Strike weave puts them in melee.

## Regression

`regress.py` (CLI d1935640f vs dd053d24c, new reqs, 3,000 it, seed 424242). Only the sim change moves:

| preset | old | new | diff |
|---|---|---|---|
| hunter__lone-wolf | 1092.0 | 1086.5 | -5.5 |
| hunter__marksmanship | 1112.9 | 1108.3 | -4.6 |
| hunter__marksmanship-tallstrider | 1073.9 | 1069.3 | -4.5 |
| hunter__beast-mastery | 1128.9 | 1124.6 | -4.3 |
| hunter__marksmanship-wolf | 1081.4 | 1077.6 | -3.8 |
| hunter__beast-mastery-tallstrider | 1057.8 | 1054.9 | -2.9 |
| hunter__beast-mastery-wolf | 1070.8 | 1068.8 | -2.0 |

New baseline `baseline_dd053d2_3000it_seed424242.tsv` (replaces e05fd73). Against e05fd73, 56 of 77
moved over 1 DPS: the explosive reach plus the req moves above (casters -4 to -8), and the Sunder ramp
from pass 42 (melee -2 to -4), which that quiet pass hadn't re-baselined.

## Checked: hunter Sapper in the weave

In the shared run (20k, seed 1) the Sapper fires in 31% of BM fights, 17% of MM, 14% of LW (SV and
melee: 100%). It shares the 1 min explosive timer with the Dark Bomb, which fires on cooldown at range.
Variant B1: explicit `castSpell {itemId: 10646}` gated on `inMeleeRange` just above "Weave: back out"
(line 5), paired, S0 = 1e6:

| preset | 120 s | 180 s | 300 s |
|---|---|---|---|
| beast-mastery | +0.06 ± 0.11 | -0.13 ± 0.11 | -0.05 ± 0.09 |
| marksmanship | -0.31 ± 0.09 | -0.22 ± 0.03 | -0.00 |
| lone-wolf | -0.27 ± 0.08 | -0.03 ± 0.05 | -0.00 |

No gain, so no second seed set. A Sapper (~1,300 a cast here) only replaces one Dark Bomb on the
shared timer, so forcing it gains nothing; keeping the Sapper in the consumes can't lose. No ranged
APL has an explicit Sapper or Land Mine line that the reach change makes dead.

## Optional

- None new. Carried items are in the pass doc's open items.
