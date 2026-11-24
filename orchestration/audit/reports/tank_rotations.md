# tank_rotations pass 26

Master 566067a394 (2026-11-24), against pass 25's baseline (master c78d983aca). 1,000 iterations, default seed,
every tank request (13) on all seven boss files (91 runs). Quiet: no new finding.

## What changed

Gate: 6 watched files (c78d983aca .. 566067a394), two commits:

- `cd71e78c2` (`sim/warrior/talents.go`, `reckless_swings.go`, `heroic_strike_cleave.go`): from wowsims/forever's
  Forever logs, Shield Slam, Revenge and Shield Bash no longer proc Unbridled Wrath (next to Whirlwind); a Heroic
  Strike or Cleave spends one Flurry and one Reckless Swings charge a cast. The prot presets run Unbridled Wrath
  5/5 (talents `05-05050005-...`), so their rage income drops a little. A measured Forever rule (user ruling
  2026-11-22), not a rotation question.
- `be773f5eb` (`sim/druid/rake.go`, `rip.go`, `tuning_70753.go`): Rip and Rake tick AP shares. No bear request
  casts either; the bear rows are byte-identical.
- No tank apl, request, rotopt template, encounter, `tools/tankbench` or `tools/tankscore` change.

## Baseline

80 of 104 rows of `ref/tank_bench/summary.tsv` are byte-identical. Only the three warrior preset requests moved
(the legacy `warrior_protection` request, Unbridled Wrath 0, is identical). Scoring means, old -> new:

| request | score | TPS | DTPS |
|---|---|---|---|
| warrior max_threat (w 1/3) | 4221.3 -> 4233.3 | 2175.3 -> 2177.0 | 419.0 -> 419.2 |
| warrior balanced (w 1) | 3470.2 -> 3454.2 | 2074.5 -> 2070.3 | 375.3 -> 376.0 |
| warrior max_survival (w 3) | 1847.5 -> 1832.8 | 1772.7 -> 1766.0 | 322.7 -> 322.3 |

TPS moves -0.2% to +0.1%, in the expected direction for the Balanced and Max survival presets; per-boss swings
(Broodlord Balanced -149 ± 56, Onyxia Max survival -111 ± 82, Vaelastrasz Max survival +98 ± 45) are the RNG
stream diverging, within about two errors. No row moved past its error without the cause above; nothing for
`tank_mechanics`.

## Rotation lines

Less rage could shift the warriors' Heroic Strike floor, so each preset's HS line was tried ±5 (abtest, paired,
3,000 iterations, scoring mean at the preset's weight, B minus preset):

| preset | HS floor | seed 1 | seed 2 |
|---|---|---|---|
| Balanced (40) | 35 | -0.2 ± 4.6 | |
| Balanced (40) | 45 | -9.8 ± 4.6 | |
| Max threat (30) | 25 | +0.9 ± 1.4 | |
| Max threat (30) | 35 | -0.8 ± 1.4 | |
| Max survival (40) | 35 | +23.8 ± 9.9 | +9.2 ± 10.2 |
| Max survival (40) | 45 | +5.8 ± 9.3 | |

Max survival at 35 does not hold on seed 2 (pooled about +16 ± 7, about 0.9% of a noisy w = 3 score, Broodlord
+75 / -44); not filed. Worth one 10k check if a later pass touches warrior rage again. The other floors stay.

## Open findings

20261122-1 (paladin Consecration mana floor), 20261104-11 and 20261104-3 (Broodlord user calls) stand: no
paladin change, warrior trinket lines untouched.

## Open items carried

- Windfury main-hand ability timing (user request 2026-11-15, pass doc Open items): still not done; out of scope for
  this gate pass.

## Repro

`tools/audit/tank_rotations/mk_warrior.py <req> B.json setcond 11 '<currentRage >= 35>'` then
`abtest -iters 3000 [-seed 2] <req> B.json` (Max threat's HS line is index 10).
