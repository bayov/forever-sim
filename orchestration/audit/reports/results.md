# Results audit, pass 70, master fa865a3b8

Why: 83 watched files changed since 5a6e9c6a6: metas refreshed (3fd9b46d2), deep Holy Shock paladin build dropped (231d549), Frostfire mage default (sim/compsearch drops/pool, report_hashes).

| Check | Result |
|---|---|
| Shared run (73 requests, 20k, seed 1) vs metas' published.dps | 73 of 73 within 0.15 |
| Shared run vs ui/dps_report/data/builds.json | 73 of 73 within 0.15 |
| Run vs previous master run (compare_runs) | 0 of 73 moved more than 0.5 |
| Drift (regen.py) vs reference | only differences: the 4 dropped paladin__level_60_shock* builds gone; reference refreshed |
| Unread request fields | none printed |

Conclusions: no sim change, no number moved. Finding results-20261101-1 (stale metas) is fixed by 3fd9b46d2 and resolved.
compare.py builds still lists OFF rows against the old req_summary.txt (pre-hotfix paladin, druid, mage, two-hand warrior numbers; the 2026-11-02 report moved them and the earlier passes verified the hotfix). req_summary.txt was not rebuilt this pass.
