# Fairness pass 19 (2026-11-01, master dd053d2)

Result: findings (2 new). Full pass: gear check, full race sweep, new baseline.

## Why it ran

Since d193564, 11 commits, 72 watched files: Goblin Sapper Charge and Goblin Land Mine reach only 10 yd
(bb21e3b32, sim/core/consumes.go), warlocks and priests moved from mksettings' default 5 yd to 25 yd (905787afd),
their request / preset edits, report_hashes and the 32-build re-sim (38fb06ea9).

## What I checked

- Gear: `audit1b.py` crashed on "Warlock (every spec)" (159443615 renamed the set line; gear_targets_v3.json
  still has two warlock targets). Fixed in `tools/audit/fairness/audit1b.py`: one line per target. 19 of 20 pass;
  Demonology fails 5a (crit 10.4 vs 14.1), the open fairness-20261031-1.
- Race sweep (5k, seed 1, race overrides) against the d96adcc baseline: only the 32 touched builds move
  (hunters -3.2 to -4.2, mages -4.4 to -5.3, warlocks -5.3 to -9.6, Shadow -4.0, Smite -5.5, Balance -4.9,
  Elemental -3.5). Survival, all melee, Enhancement and Ret are +0.0: no leak. Arms passes Destruction (11th).
  Top over bottom 1.54x.
- Published against current: the metas' `published.dps` are stale for 74 builds (gaps -1.9 to -8.6 here);
  builds.json matches the shared 20k run. Already results-20261101-1; I confirm it independently (sweep gaps match
  vs_published.txt within the 5k noise), not re-filed.
- Best race against report race: best everywhere except Affliction. 20k: Orc 1013.5, Human 1012.1 (report).
- Explosives fairness: melee and Survival (5 yd) keep the Sapper; ranged hunters keep it through melee weaves
  and throw the Dark Bomb; casters take none. Not unfair: a caster walking in for one Sapper per 5 min would lose
  more casting than it deals. Not filed.
- Distances: Elemental 20 yd, Balance 30 yd, everyone else ranged 25 yd; no standard says which. Every main
  caster nuke has a MissileSpeed, so distance is flight time. Paired 20k, seed 1
  (`tools/audit/fairness/distance.py <cli> 25 <ids>`, new helper):

| Build | Now | DPS now | At 25 yd | Delta |
|---|---|---|---|---|
| shaman__elemental-level60 | 20 yd | 768.5 | 762.0 | -6.5 |
| druid__balance | 30 yd | 846.6 | 846.6 | +0.0 |
| druid__balance-wrath | 30 yd | 806.5 | 808.8 | +2.3 |

## Findings

- fairness-20261101-1 (medium, user ruling): no caster distance standard; Elemental's 20 yd is +6.5 over 25.
  Options: all ranged at 25 yd; each at the range its rotation needs; or keep and write it down.
- fairness-20261101-2 (low, orchestrator): Affliction's report race Human is 1.4 under Orc.

## Open items carried

- fairness-20261031-1 (Demonology set) and the older open findings: no change on master.
- Ice Lance Low range, Flametongue Totem stacking, boss partial resist, Rage k / pet scaling: unchanged.

## Baseline

`ref/fairness/sweep_baseline.log` and `table_baseline.md` replaced with this sweep (dd053d2).
