# tank_rotations: pass 2 (2026-11-04, master b38ff16)

Forced by hand after tank round 2 landed (b38ff16): the four class workers fixed pass 1's findings and re-searched
every tank's Max threat / Balanced / Max survival presets with new rotopt knobs (warrior wallAt grid and Sunder
count, paladin Bulwark step, bear roarAfterLacerate and in-fight Enrage, Warden shockFirst).

## What changed (f62bfc74..b38ff16)
All 13 tankbench requests, all 12 preset APLs and presets.ts, rotopt's four tank templates, tools/tankbench/main.go,
tools/tankscore, sim/core/tank_score.go and the class sim files of round 2 (71 watched files).

## Baseline (ref/tank_bench/summary.tsv, header updated to b38ff16)
Every request on every boss, 1,000 iterations, default seed. Every row equals the table the class workers wrote after
round 2, so no row moved without a rotation change. Scoring mean (six bosses without Ragnaros), score at own weight:

| request | w | score | SafeDPS | DTPS | healing needed | death floor |
|---|---|---|---|---|---|---|
| druid_bear_max_threat | 1/3 | 2837.3 | 1685.0 | 577.8 | 592.7 | 0 |
| druid_bear_balanced | 1 | 1867.5 | 1603.3 | 575.8 | 484.5 | 0 |
| druid_bear_max_survival | 3 | -750.2 | 1519.8 | 576.7 | 463.2 | 0 |
| paladin_protection_max_threat | 1/3 | 2785.0 | 1493.7 | 319.7 | 238.5 | 0 |
| paladin_protection_balanced | 1 | 2397.7 | 1457.3 | 301.2 | 208.3 | 0 |
| paladin_protection_max_survival | 3 | 1363.3 | 1457.3 | 301.2 | 208.3 | 0 |
| shaman_warden_max_threat | 1/3 | 2292.3 | 1592.8 | 806.8 | 954.0 | 0 |
| shaman_warden_balanced | 1 | 547.8 | 1551.0 | 791.7 | 905.2 | 0 |
| shaman_warden_max_survival | 3 | -4356.7 | 1482.8 | 754.0 | 858.0 | 0 |
| warrior_protection_max_threat | 1/3 | 3694.8 | 2051.8 | 416.8 | 416.8 | 0 |
| warrior_protection_balanced | 1 | 3008.5 | 1963.5 | 366.8 | 324.0 | 0 |
| warrior_protection_max_survival | 3 | 1205.0 | 1851.0 | 353.5 | 295.8 | 0 |
| warrior_protection (legacy) | 1* | -14.2 | 1048.2 | 637.3 | 757.0 | 0 |

Drift: in all 12 preset requests, priorityList and prepullActions equal the ui APL exactly, and talents and gear
match presets.ts.

## Pass 1's findings (paired, 3k, six-boss mean; fix reverted minus preset)
| finding | state | check |
|---|---|---|
| 1 bear Roar below Lacerate | resolved | -234.9 ± 8.3 / -212.6 ± 11.6 / -120.3 ± 30.0 (MT / Bal / MS) |
| 2 warrior Shield Wall 17 s | resolved | Bal 12 s -27.4 ± 6.1, 16 s -8.8 ± 4.2, 18 s -5.6 ± 5.0; MS 12 s -89.9, 16 s -16.2, 18 s -20.6 |
| 3 paladin Bulwark 13 s | open (user ruling) | without Broodlord 13 s loses -3.6 (5 seeds), -4.4 / -4.2 at 10k; rests on Vaelastrasz vs Onyxia |
| 4 Warden shock first | resolved | -99.0 ± 6.1 / -88.1 ± 11.6 / -95.9 ± 34.8 |
| 5 bear Enrage | resolved | in-fight under 40 Rage beats pre-pull -1 s by 60 (5 seeds and 10k) |
| 6 paladin autocast top | resolved | -12.9 ± 0.2 / -12.3 ± 0.4 / -11.0 ± 1.3 |
| 7 warrior Sunder filler 15 | resolved | 30 Rage -6.9 ± 1.6; -5.7 at 10k |
| 8 notes | resolved | all three parts present (the prepull stance action still has no note: a nit) |

## Seed and iteration count (the druid worker's report)
- rotopt -seed does reach the sim, but iteration i runs on seed+i (sim/core/sim.go:428), so -seed 2 replays 2,999 of
  seed 1's 3,000 iterations, and a 10k confirm at the same seed contains the 3k search. That's why it looked inert.
- The 3k-vs-10k gap is luck in seed 1's first 3,000 iterations (+30 for the bear, 2.7 SD), not an N-dependent
  bias: healing needed is the empirical 95% quantile (sim/core/tank_score.go:203-229); bear Balanced at 30k is
  1843.0 vs 1844.5 ± 3.9 over 25 independent 1k seeds; warrior 3005.0 vs 3006.8 ± 2.8.
- Bosses share iteration seeds, so the six-boss mean's error is about 1.25× what rotopt and abtest assume.
- Paired calls hold: Shield Wall 16 vs 17 s, Sunder 30 vs 15, Bulwark 13 vs 10 (without Broodlord), bear Enrage
  keep their sign on 5 disjoint 3k seeds and two 10k seeds, and the paired ± covers the seed spread.
- Tooling asks are in finding 13 (hash the iteration seed, disjoint confirm seed, per-boss seed offsets or a
  1.25× error, fixed count for absolute scores, a 2× paired-error rule on a second seed).

## Line-by-line probe (about 45 A/Bs, 1k, wins settled at 3k)
New findings: 9 (Warden Max survival armor trinket held for the last 30 s, +328.5 ± 17.0), 10 (Warden Balanced
armor trinket held instead of 990015, +95.6 ± 7.3), 11 (warrior Max survival's held armor trinket wins only on
Broodlord), 12 (bear Max threat Enrage at the knob's edge, +3.0 ± 0.5).

Tried, no win (1k, paired): bear Barkskin in the last 15 s (Bal -44.8, MS -123.5), from the pull (Bal +8.2 ± 10.6),
none at MT (-34.5); bear Balanced Enrage under 60 (-8.1 ± 12.0) or 20 (-23.9); bear potion on a Rage condition
(-0.5 to -1.5); bear armor trinket 990815 held (MS +12.8 ± 18.2 at 30 s; Bal -179.9); Warden shockFirst=0 (Bal
-94.8, MS -119.7); Warden Balanced autocast to the top (+0.5 ± 7.1) or 990015 held 35 s (-23.2); paladin
Consecration mana 0.5/0.7 (all within ±1 error); warrior Bloodrage on cooldown (Bal +1.6 ± 3.6, MS -10.0); warrior
Sunder keep-up to 3 (MT +0.4, Bal -2.6), no keep-up at MT (+0.6 ± 1.5).
Per-ability output (Golemagg): no Rage capping; bear and warrior use every GCD; the paladin's ~15 idle GCDs are
Consecration held by mana (thresholds flat); Holy Shield about every 10 s.

## Open items for the next pass
- A defensive held to the fight's end wins every time, mostly on Broodlord and Onyxia (findings 3, 9-11, Shield
  Wall). Whether the replay's healing need at a fight's end is over-weighted is a question for tank_mechanics,
  untested here.
- The Warden uses about 52 of 77 GCDs on Chromaggus at Balanced; not tested why.
- Scratch (gone with the session): $AUDIT_WORK/tank_rotations/{bench,noise,probe}/.
