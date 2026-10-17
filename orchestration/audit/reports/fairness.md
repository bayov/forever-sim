# Fairness pass 2 (2026-10-17 18:55 UTC, master 1e2f7ec)

Result: quiet. No new findings, none open.

## Why it ran

Since 0a204d4, 14 watched files changed. The item database was regenerated with enchant values and
item sets taken from the Forever client's tables (cb3f64fa). There were also protos, `sim/core/spell_ranks.go`,
`sim/core/unit.go` and `sim/shaman/healing.go` (top-rank IDs for "highest rank" rotations, bdf075cc),
and `ui/enhancement_shaman/presets.ts` (`rotationsNameTopRanks` on the leveling plan). The item
database and `sim/core` are on the full-pass list, so a full pass was done.

## What was checked

| Check | Result |
|---|---|
| `audit1b.py` (19 gear sets) | all `5a PASS | 1b PASS` |
| db.json diff | set names, the enchant/set import and random-suffix entries (990xxx "of the Physician/Prophet"); no report-request item changed |
| Shared run vs published (`runs/vs_published.txt`) | 0 of 66 moved more than 0.5 DPS |
| Full race sweep, 5000 it, seed 1, vs 0a204d4 baseline | byte-identical per build and race |
| `table.py` | no `<--` rows |
| Published vs current | every build within 1.2 DPS |
| Best race vs report race | report race is best everywhere (Arcane, Survival exact ties) |
| `findings/results.md` | no open findings to confirm |

The code changes add `highest_rank_id` metadata, and rotations that name a top rank now cast the
highest rank the unit knows. At level 60 every report build knows its top ranks, so nothing changes
there, and the sweep confirms it. The racials in 1e2f7ec5 regenerate only `ui/core/forever/racials.json`
text; the sim's racial code didn't change.

## Ranking

Unchanged: Demonology 1177.8 top, Subtlety 656.7 bottom, 1.79×. Full table in
`ref/fairness/table_baseline.md` and `orchestration/fairness_report.md` §1.

## Baseline

Kept `ref/fairness/sweep_baseline.log` (same bytes). Refreshed `table_baseline.md` (its Before
column is now 0a204d4). `state/fairness.json` baseline = 1e2f7ec5.

## Open items carried over (no change)

Hunter end-of-fight test; Ice Lance Low range ruling; `sham_ftt_stack` To Verify; boss partial resist
log; Warrior Rage k and pet scaling ranges. Client build not re-checked on wago.tools this pass.
