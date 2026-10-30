# Results audit, pass 62 (master 36a772fd1)

Checked with the loop's shared run (20k, seed 1, 180 s, all 77 requests) against published, builds.json and the reference.

- Why due: 25 watched files since d50909a37: the raid sim's compsearch (not read by the report's sims), sim/core buffs.go (Furious Howl rank by level), debuffs.go (rank labels for Expose Armor and Curse of Recklessness), sim_concurrent.go (empty party slot guard), sim/warlock/curses.go (level-ranked Recklessness, Forever only), sim/web/main.go, report_hashes.txt, healer presets (ui/holy_paladin, ui/restoration_druid). No change under tools/rotopt, tools/dps_report, tools/database, assets or ui/dps_report, so drift (mksettings vs requests) is unchanged and was not re-run.
- Published vs mine: 76 of 77 within 0.15. mage__frostfire meta 1087.5 vs run 1088.4 (+0.9), the known stale meta of pass 61 (builds.json already carries the run's value). Not filed.
- Moves against the reference: none. The level-60 changes (howl rank 24597 at 83, top-rank labels) reproduce the old numbers exactly.
- Unread fields, drift: unchanged (no mksettings or request changes).
- Findings: none new. results-20261025-1 stays open.
