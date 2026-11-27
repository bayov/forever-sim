# Results audit report

## Pass 114 (master f4c45c3f9, 2026-11-27)
- Trigger: 41 watched files since 313069ce: rogue talent strings (ui/rogue/presets.ts, 5 lines, 04631ff49), rogue requests and metas (f4c45c3f9 republish), metas re-gear refresh (87364fe2a), rotopt tie-break (no published effect).
- Run: shared run (68 requests, 20k, seed 1) from wowsimcli f4c45c3f9.
- builds.json vs run: 68 of 68 within 1.5 (max 1.32). Rogue rows: combat-daggers 829.2 vs 828.7, combat-swords 840.5 vs 839.9, mutilate 786.7 vs 786.6, seal-fate-daggers 819.6 vs 819.5, seal-fate-swords 848.7 vs 848.3, subtlety 786.5 vs 786.5.
- Metas published.dps vs run: 28 of 68 differ by more than 0.5, all within 1.5 (noise band at 180 s: hunters, paladins, warriors, warlocks); none stale. The rogue metas now match.
- Drift report equals the reference (115 lines); unread fields: none.
- orchestration/audit/ref/results/req_summary.txt is the 2026-10-17 reference and no longer comparable (compare.py builds shows 60+ OFF); not refreshed this pass.
- Findings: none new. Resolved: results-20261127-1, -2, -3 (metas, duty verify blocks and Changes line fixed in 87364fe2a), results-20261104-2 and -1 (metas lag closed).
- Not checked: Tanking and Healing modes; the 39 hand-recorded builds.
