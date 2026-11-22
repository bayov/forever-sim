# Results pass 98, master 7a6d1beb9d (2026-11-22)

Shared run: 68 requests, 20k, seed 1 (`/tmp/audit-work/runs/master/summary.tsv`), CLI built from 7a6d1beb9d.

## builds.json (30314ad60) against the run
57 of 68 within 1.5. The 11 off are all explained by sim commits after that re-sim:

| Build | builds.json | run | move | cause |
|---|---|---|---|---|
| mage frost | 945.4 | 834.6 | -110.8 (-11.7%) | 4192d6e49 Ice Lance coefficient 0.429 -> 0.11 (user ruling 2026-11-22) |
| mage frostfire | 1064.3 | 1006.0 | -58.3 (-5.5%) | same |
| hunter survival / -wolf / -tallstrider | | | -4.6 / -3.4 / -3.9 | 051a21cd3 pet keeps summon attack power |
| hunter BM / -wolf / -tallstrider | | | -4.0 / -2.8 / -3.2 | same |
| hunter MM / -wolf / -tallstrider | | | -2.1 / -2.7 / -1.7 | same |

Paired checks (20k, seed 1, master binary): request rotation 834.6 / 1006.0 (Frost / Frostfire); the `asm_apls` no-Ice-Lance rotation 681.8 / 993.9. Ice Lance still pays at 0.11 in both; the Frost APL file is stale.
All other builds are within 1.5 of the run: 30314ad60 caught them up.

## Reference moves
`ref/results/req_summary.txt` is old (baseline 0facd9428), so most rows show as moved; not refreshed. Metas lag builds.json: 62 of 68 off by more than 0.5 (results-20261104-2).

## Drift and unread fields
Drift report equals the reference except the known feral sapper line (results-20261105-1), now on both feral requests. Unread fields: none printed.

## Ranking note
Frost mage is now 834.6, below Balance (956.1); Frostfire 1006.0 sits below Fire 1053.0. The pass doc's "Ice Lance coefficient 0.429" under Settled is retired by the 2026-11-22 ruling (passes/results.md is not owned by this pass).
