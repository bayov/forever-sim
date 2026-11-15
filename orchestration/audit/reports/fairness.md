# Fairness pass 42 (2026-11-15, master 70ea588)

Result: quiet (no new finding). Watched files changed since d630df2: sim/druid/ferocious_bite.go and sim/druid/forms.go
(0cc2c3c8b: an avoided Ferocious Bite refunds 80% of its 35 Energy under Forever; Sharpened Claws adds spell crit in
the feral forms), the Feral test results and report_hashes.txt (only druid__feral-cat and its duty-ff variant moved),
and the Feral re-sim (8c75031e2). Scoped pass: only druid code changed, so the druid base requests were re-swept at every
race (5k, seed 1, the loop's 70ea588 CLI and tree). Client 1.60.1.70938.

## What moved against the 1bf5feb baseline

| Build | Before | Now | Cause |
|---|---|---|---|
| Feral (Night Elf) | 1153.8 | 1156.5 (+2.7, +0.2%) | Bite refund (0cc2c3c8b); the commit's 20k re-sim says +2.5, consistent |
| Feral other races | Windshaper 1136.2, Tauren 1128.2 | 1138.9, 1130.4 | same, race order unchanged |
| Balance, Balance Wrath | 952.1, 862.2 | identical | Sharpened Claws spell crit is feral-form only |

No shared leak: the only code change is in sim/druid, and report_hashes.txt changes only the two feral builds.

## What I checked

- Fairness of the rule: the 80% refund on an avoided Bite now matches Rip, Claw, Shred, Ravage and the bear attacks
  (sim/druid), and the rogue finishers (finisherMissRefund, sim/rogue/rogue.go:122) and builders. Symmetric.
- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Best race: Night Elf is still Feral's best and report race (+17.6 over Windshaper).
- Published: the Feral meta still says 1139.0 (gap now +17.5); 8c75031e2 updated builds.json/updates. Same stale-meta
  pattern as results-20261104-2 / -20261112-1; not refiled. Other gaps as in pass 41 (`/tmp/audit-work/runs/vs_published.txt`).

## Ranking now (report race, 5k)

Unchanged order: Feral stays 4th at 1156.5. Top over bottom 1.65x. Full table: `ref/fairness/table_baseline.md` and
`orchestration/fairness_report.md` §1.

## Baseline

`ref/fairness/sweep_baseline.log` feral line and `table_baseline.md` feral row replaced; the other rows are the 1bf5feb
full sweep. Baseline sha 70ea588 (see `baseline_note` in state).

## Open for the next pass

- Next full re-sweep on any sim/core, request or database change.
- Other metas' published.dps: re-check when results-20261104-1/-2 and -20261112-1 are fixed.
