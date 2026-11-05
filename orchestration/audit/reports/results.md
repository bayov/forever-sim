# Results pass 79, master a079415316 (2026-11-05)

Method: shared run (/tmp/audit-work/runs/master, 68 builds, 20k, seed 1, CLI a0794153163a) vs the metas' published.dps and
ui/dps_report/data/builds.json; drift check (regen.py, mksettings built at a07941531); unread_fields.py.

| Check | Result |
|---|---|
| Run vs builds.json | all 68 within 1.3 (noise between runs); the report republish caught up (warrior, feral included) |
| Run vs metas | 28 of 68 off by more than 0.5, same set as pass 78 (open results-20261104-2); only feral-cat moved further (+30.5 vs +26 before, Ferocious Bite hotfix 113403 and the sapper) |
| Drift vs reference | one new difference: feral-cat and duty-ff carry sapperExplosive that mksettings does not write (results-20261105-1, -3.6 DPS) |
| Unread fields | none (not even notes) |

Conclusion: no modelling error found. The metas need the same republish step as before. Reference req_summary.txt not replaced
(metas, not the reference, are the open gap).
