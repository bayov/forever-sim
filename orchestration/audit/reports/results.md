# Results audit, pass 59 (master 1274e15b3)

Checked with the loop's shared run (20k, seed 1, 180 s, all 58 requests) and a fresh mksettings build.

- Published vs mine: all 58 within 0.15 (the metas' published.dps were refreshed by 5b019f33c; no stale gaps).
- Moves since the last audited build (224b8a4e0, builds.json): Beast Mastery 1150.4 -> 1131.8 (-18.6), Marksmanship 1137.3 -> 1115.7 (-21.6), Lone Wolf 1118.9 -> 1094.3 (-24.6), Survival 1029.8 -> 1028.9 (-0.9). Every other row is unchanged.
- Cause: 73cedc9a9, the launch realms' 2026-10-28 hunter hotfix (user ruling). Aspect of the Hawk now casts rank 6 (55 ranged AP) instead of rank 5 (90 AP before the hotfix, 47 after), and the three requests moved from spell 14321 to 14322. A drop of 35 AP on the hunter's ranged and pet-less white and shot damage is the right size for -1.6 to -2.2%. Trueshot Aura changed in the same commit (Survival keeps its rank).
- Drift: drift_report.txt identical to the reference (open finding results-20261025-1 stays open, no new difference; the hunter requests have 0 drift).
- Unread fields: only `notes`.
- Reference req_summary.txt replaced with this run.
