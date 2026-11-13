# tank_rotations pass 12

Master 97da1cca5 (2026-11-13), against pass 11's 71562a8b0 baseline. 1,000 iterations, default seed; A/Bs paired
(abtest) at the preset's own weight on the six scoring bosses, Ragnaros reported apart.

## What changed

Since 71562a8b0 (gate: 2 watched files):

- a1a1a81bb (sim/core/attack.go): Windfury Totem no longer rolls on extra attacks; stored extra attacks keep their
  proc as source.
- 33b13e310 (sim/paladin/unsealed_wrath.go): Nov 11 hotfix, Windfury Totem swings store no Unsealed Wrath echoes;
  Ret paladin re-tuned on it (tools/rotopt/paladin_presets.sh: seal=1, twistBack=3). The tank template
  tools/rotopt/paladin_tank.go is unchanged.
- No change to tank APLs, tools/tankbench requests, tools/encounters, tools/tankbench/main.go or tools/tankscore.

## Baseline (scoring mean, own weight; pass 11 ref -> now; full grid in ref/tank_bench/summary.tsv, rewritten)

| request | score | SafeDPS | DTPS | healing needed |
|---|---|---|---|---|
| paladin Max threat | 3266.7 -> 3248.8 | 1744.8 -> 1739.0 | 308.7 -> 308.2 | 264.8 -> 272.3 |
| paladin Balanced | 2837.5 -> 2809.3 | 1716.5 -> 1710.8 | 308.0 -> 307.7 | 238.0 -> 243.8 |
| paladin Max survival | 1810.7 -> 1780.2 | 1637.2 -> 1629.3 | 287.8 -> 287.5 | 198.3 -> 201.5 |
| bear (3), Warden (3), warrior (4 incl. legacy) | bit-identical | | | |

Only the paladin rows moved, all the same way: SafeDPS -6 to -9 on every boss (same seed, so effectively paired),
healing needed +3 to +8. That is the intended Windfury rule change (fewer Windfury swings / echoes for the
paladin), not a regression. The other ten requests are bit-identical, so the engine change touched nothing else for
them. No finding.

## Changes tried (paired 1,000, 7 bosses)

The Ret paladin flipped its seal on this hotfix, so the tank's seal knob was re-checked:

| change | preset | scoring mean diff | paired err | verdict |
|---|---|---|---|---|
| Seal of Righteousness (20293) for Seal of Fury (20423) | Balanced | -90.2 | 3.0 | loses |
| same | Max threat | -42.6 | 1.2 | loses |
| same | Max survival | -187.5 | 7.8 | loses |

Seal of Fury holds at all three weights.

## Open findings

New: none. Open, unchanged (user calls): 20261104-11 (warrior Max survival armor trinket held 35 s, Broodlord),
20261104-3 (paladin Bulwark 13 s, Broodlord).

Raw output: `/tmp/audit-work/tank_rotations/` (base/, ab_sor_*.txt, old.tsv).

## Note for the next pass

Build both tools with `-tags=with_db` (abtest panics in NewRaid without it):
`go build -tags=with_db -o /tmp/audit-work/abtest ./tools/audit/tank_rotations/abtest`. If the tankbench build says
sim/core/proto is missing, re-run `source tools/audit/setup.sh` (it generates the protos) and build again.
bench_summary.py writes a fixed header line; fix line 1 of summary.tsv to the audited master after running it.
