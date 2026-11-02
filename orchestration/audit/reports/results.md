# Results audit, pass 71, master 796495d6f

Why: 209 watched files changed since b647a21ff (db, protos, sim/core; 123 commits, incl. the 2026-11-02 re-searches and Demonology set).

| Check | Result |
|---|---|
| Shared run (73 requests, 20k, seed 1) vs metas' published.dps | 72 of 73 within 0.15; shaman__elemental-level60-mana-tide 752.83 vs 753.2 |
| Shared run vs ui/dps_report/data/builds.json | 73 of 73 within 0.15 |
| Run vs previous master run (compare_runs) | 0 of 73 moved more than 0.5 |
| Drift (regen.py) vs reference | identical (127 lines) |
| Unread request fields | none printed |

Conclusions: the sim changes in this range moved no published number. One low finding (results-20261102-1): one stale meta.
compare.py builds still lists OFF rows against the old req_summary.txt (pre-hotfix numbers); req_summary.txt was not rebuilt this pass.
