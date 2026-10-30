# Results audit, pass 63 (master bbbaf2e3c, client build 1.60.1.70707)

Why due: 33 watched files changed since 36a772fd1 (shaman spell ranks and untrained-rank option, Elixir of Wisdom +3 Int,
Evocation/Innervate, compsearch, db regen, enhancement presets.ts).

## Checked
- Shared run (20k, seed 1, every request, 77 of them): 75 within 0.15 of published. Two are over, both identical to the
  reference run, so not moves: mage__frostfire 1088.4 vs 1087.5 (+0.9), shaman__elemental-level60 772.7 vs 772.2 (+0.5).
  Both are within seed noise of the published re-sim; neither request carries Elixir of Wisdom.
- Moves against req_summary.txt: none. builds.json matches the run on all 77.
- Drift report (regen.py with the master mksettings): identical to the reference (155 lines). Open consumes drift
  (results-20261025-1) is unchanged.
- Unread fields: none reported (not even notes).
- Rank-training and shaman rank changes did not move any shaman number: the requests keep max ranks.

## Housekeeping
- regen.py mapped hunter ids without the -tallstrider/-wolf suffixes; fixed so the drift check runs on the 77 requests.
- ref/results/req_summary.txt replaced with this run (77 lines); drift_report.txt unchanged.

## Conclusion
Quiet. No new findings.
