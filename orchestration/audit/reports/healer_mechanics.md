# Healer mechanics: pass 4 (2026-11-05 UTC, master bba052f750, client 1.60.1.70790)

Quiet pass. Due because a watched file (tools/healbench/main.go) changed on master since pass 3 (a079415316).

## Get current

| Source | Seen | Healer content |
|---|---|---|
| wago.tools latest build | 1.60.1.70790 (same as pass 3) | none new |
| Dev notes thread 2360696 | not re-fetched (pass 3 read post 10 at 01:33Z today) | none |
| update_ledger.md | only the audit-loop commit 5fc045ccb (70790 entry, already read) | none |

## What changed on master

| File | Commit | Healer effect | Verdict |
|---|---|---|---|
| tools/healbench/main.go | cb4065412, bba052f75 | new flags: `-seed` (default 101, same as the old hard-coded seed), `-labeled` (UseLabeledRands), `-scales` + `-vectors` (appends per-fight deaths, HPS, useful HPS, mean healer mana left per scale, for tools/dps_report/tankheal.py race pairing). Healers are picked by `b.slots` as in `run`; HPS and EHPS summed over healers, mana averaged | OK, bench tooling; default runs unchanged |
| sim/core/raid.go (doneHealingFight) | cb4065412 | appends a per-fight flag (1 tank died, 2 other died) to `HealingFight.FightDeaths`; counters unchanged | OK, metrics only |
| sim/core/sim_concurrent.go | cb4065412 | concatenates `FightDeaths` in AddResult order, the same order as each DistributionMetrics.AllValues (line 148), so deaths and HPS vectors stay aligned per fight | OK |

No heal rank, healing talent, set bonus or healing-engine rule changed, so steps 3 and 4 had nothing new to check.
The verdicts in `ref/healer_mechanics/verdicts.md` stand unchanged.

## Tests

No new heal or engine test needed for this range. `FightDeaths` has no test of its own; it is reporting
for the race tables, and its counts are cross-checkable against FightsTankDeath/FightsOtherDeath. Not a new
finding (covered in spirit by healer_mechanics-20261104-10).

## Open findings

- healer_mechanics-20261104-11 (Riptide's mana): still open.
- healer_mechanics-20261104-10 (untested heals, talents, engine paths): still open.

## New findings

None.
