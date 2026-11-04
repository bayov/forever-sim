# Results audit, pass 78 (master 0facd9428, 2026-11-04)

Trigger: 340 watched files (phase 5 hit-trade regears, shaman hotfix 113380, tank/healer rounds, UI presets).

- Shared run (68 builds, 20k, seed 1) vs metas: 31 builds differ by more than 0.15; vs builds.json 33 differ. All trace to request regears or sim commits (finding results-20261104-2).
- Moves (run vs meta): feral-cat +26 (74cc2b8e3), Ret level_60 family +11 (3f4d44d4a), Arcane +7.6, shaman +6.5 to +7.4 (87922bb77, 3e4796fe7), rogues -12 to +7.9 (31c10b61c, aa4551a04), warlock fire +0.8 to +1.1 (c8a7b4ac3; matches its meta note), warriors <=1.6 off metas but 6.6 to 25 below builds.json (cd3f44cf2).
- Drift (regen.py): identical to the reference. Unread fields: no output.
- Not checked by paired runs: the sizes of individual moves were taken from the commits' own notes; no unexplained move found.
- Result: findings (1 new low, results-20261104-2; results-20261104-1 still open). req_summary.txt replaced.
