# Results audit, pass 67, master dd053d24c

Why due: Goblin Sapper Charge and Land Mine reach only 10 yd (consumes.go `explosiveBlastRadius`), warlocks and priests cast from 25 yd, so requests swap the casters' filler Land Mine for EZ-Thro Dark Bomb and Sapper is dropped where out of reach; report hashes updated; 38fb06ea9 re-simmed 32 builds.

## Checks
| Check | Result |
|---|---|
| Shared run (77 requests, 20k, seed 1) vs builds.json | all 77 within 0.1 |
| Shared run vs metas' published.dps | 74 of 77 off (0.5 to 8.9), all negative: stale metas, finding results-20261101-1 |
| Largest moves vs metas | warlocks -5.8 to -8.9 (-0.6 to -0.9%), hunters -2.4 to -6.9, Balance -4.6, mages -3.8 to -4.7, Ele -4.0 to -4.2, Smite -5.4, Shadow -4.1, Feral -3.0 to -4.1 |
| Explanation | the explosive reach and 25 yd caster range (bb21e3b32, 905787afd, 9eb1fad6b); lost Sapper/Land Mine damage matches the sizes |
| Drift (regen.py vs ref/drift_report.txt) | identical |
| Unread fields | none |

## Conclusion
Not quiet only for the stale metas (low). Numbers moved by the ruled explosive-reach change; ranking at this size is not re-ranked here. req_summary.txt in ref/ is replaced with this run (pub = builds.json).
