# Fairness pass 1 (2026-10-17 15:04 UTC, master 0a204d4)

Result: quiet. No new findings, none open.

## Why it ran

Since 1735382: item database regenerated from a Wowhead refresh (b1b07631) and shaman heal rank
families (b803e856, `sim/shaman/healing.go`, `shaman.go`). Both are on the full-pass list (item
database, class code), so a full pass was done.

## What was checked

| Check | Result |
|---|---|
| `audit1b.py` (19 gear sets) | all `5a PASS | 1b PASS` |
| Synthetic items (ids 99xxxx) in the db.json diff | none changed (context lines only) |
| Shared run vs published (`runs/vs_published.txt`) | 0 of 66 moved more than 0.5 DPS |
| Full race sweep, 5000 it, seed 1, vs 1735382 baseline | byte-identical per build and race |
| Published vs current | every build within 1.2 DPS |
| Best race vs report race | report race is best everywhere (Arcane, Survival exact ties) |
| `findings/results.md` | no open findings to confirm |

The shaman change only registers heal ranks as rank families (Healing Wave, Lesser Healing Wave,
Chain Heal); no DPS spell is touched, and Elemental / Enhancement sweep numbers are unchanged.
The `.results` churn in rogue and shaman tests comes from real items in the test gear, not the
report's synthetic sets.

## Ranking

Unchanged from the baseline: Demonology 1177.8 top, Subtlety 656.7 bottom, 1.79×. Full table in
`ref/fairness/table_baseline.md` and `orchestration/fairness_report.md` §1.

## Baseline

Refreshed `ref/fairness/sweep_baseline.log` (same numbers) and `table_baseline.md` (Before column now
1735382); `state/fairness.json` baseline = 0a204d4.

## Open items carried over (no change)

Hunter end-of-fight test; Ice Lance Low range ruling; `sham_ftt_stack` To Verify; boss partial resist
log; Warrior Rage k and pet scaling ranges. Client build not re-checked on wago.tools this pass.
