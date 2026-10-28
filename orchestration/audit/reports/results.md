# Results audit, pass 60 (master dd663e652)

Checked with the loop's shared run (20k, seed 1, 180 s, all 58 requests) and a fresh mksettings/wowsimcli build at dd663e652.

- Why due: 126 watched files changed since c9d93dd66 (mostly tools/dps_report/req/* re-written by the report's Dwarf gear-trade, Smite Chastise and priest racial re-sims). sim/, proto/, mksettings, tools/database and assets/database are unchanged in the range.
- Published vs mine: 56 of 58 within 0.15. Two metas lag builds.json (which matches my run): mage__frostfire meta 1087.5 vs run 1088.4 (+0.9), shaman__elemental-level60 meta 772.2 vs run 772.7 (+0.5). Cause: 8624ba126 / 3ad93a86b changed those requests after the meta numbers were written. Tiny and stale-meta only; not filed, will clear on the next re-sim.
- Moves against the reference req_summary.txt: only those two rows (+0.9, +0.5); every other row identical.
- builds.json vs the run: all 58 within 0.15 except the same two (builds.json is right).
- Drift: drift_report.txt identical to the reference (open finding results-20261025-1 stays open, nothing new).
- Unread fields: only `notes`.
