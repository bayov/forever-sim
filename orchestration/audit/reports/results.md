# Results audit, pass 53 (master f796d9953, quiet)

- Why due: 23 meta.json files changed (rogue, shaman Enhancement, warrior). The diff is notes wording only (flask and hit sentences); no sim, proto, mksettings, database, preset or request change since e2316bebe.
- Shared run (57 requests, 20k, seed 1): 0 moved against `ref/results/req_summary.txt` (tolerance 0.05).
- Published vs mine: 53 of 57 within 0.15. Over: druid feral-cat -1.4, feral-cat-duty-ff -1.5, rogue subtlety -1.0, subtlety-humanoid -1.1. Same as pass 52 and the reference: the metas' published.dps lag the 2026-10-24 rounds (meta lag, not a sim gap). Re-ran feral-cat and subtlety with the shared CLI: 866.7 and 714.9, reproducible.
- Drift and unread fields: not re-run (no mksettings or request change); standing finding results-20261025-1 stays open.
- builds.json unchanged since e2316bebe.
