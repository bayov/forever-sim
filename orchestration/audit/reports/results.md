# Results pass 80, master 9ba00c2dc (2026-11-08)

Method: shared run (/tmp/audit-work/runs/master, 68 builds, 20k, seed 1, CLI 9ba00c2dc473) vs metas and builds.json; paired
runs of the old Enhancement requests; drift check (regen.py, mksettings built at 9ba00c2dc); unread_fields.py. Client build 1.60.1.70862.

| Check | Result |
|---|---|
| Run vs metas | 28 of 68 off by more than 0.5. Same set as pass 79 plus the four Enhancement builds (-11.7 to -17.6) |
| Run vs builds.json | 64 within 1.5; the four Enhancement builds -18.3 to -19.3 (report not re-simmed after 9ba00c2dc) |
| Enhancement move | 9ba00c2dc: Windfury and Flametongue Totem replace each other; the old request on the new binary gives identical DPS (1146.67, 1132.22), so the move is the sim rule |
| Other commits (2a60326a1 whole-point spell rolls, 0193721bc downranking on heals only) | no build moved more than 1.1 outside the known regear lag |
| Drift vs reference | only the open feral sapper difference (results-20261105-1) |
| Unread fields | none |

Conclusion: no modelling error; no new finding. Enhancement stale numbers are noted under results-20261104-2. Reference
req_summary.txt not replaced.
