# Results pass 57, master 5d27b0015 (2026-10-28)

Trigger: sim/core/racials*.go (ae8522c14: Blood Fury CD 2 min, Berserking 15 s, Touch of the Grave ICD 15 s, per dev notes post 8, server side; client 70634 still reads 180000/25000) and tools/dps_report/req/* (race re-picks, standards 29).

| Check | Result |
|---|---|
| Shared run, 58 requests, 20k seed 1 | matches ui/dps_report/data/builds.json on all 58 (max diff 0.1) |
| Metas' published.dps vs run | 34 of 58 off by more than 0.15 (stale metas; open finding results-20261027-1) |
| Drift (regen.py vs ref drift_report.txt) | identical to the reference |
| Unread fields | only notes |
| Racial code diff | consistent with comments and dev-notes citations; no modelling error seen |

Largest meta gaps: mage__frost -35.5, warlock__affliction -17.2, mage__fire -12.0, mage__frostfire -8.9, rogue Mutilate/Seal Fate +5.7..6.0.
No new finding. Baseline files not replaced (req_summary.txt is not refreshed this pass).
