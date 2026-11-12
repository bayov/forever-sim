# Results pass 85, master 8306f1d52

- Shared run (68 builds, 20k, seed 1, /tmp/audit-work/runs/master) vs builds.json: all 68 within 1.5. The republishes 2b62e304a (druid, Rake's hit takes attack power) and bff0169e4 (hunter pet rules, Windfury Totem) brought builds.json up to a0b4d3be9 and later. results-20261112-1 is resolved.
- Vs the metas' published.dps: 45 of 68 off by more than 0.5. The republishes did not touch the metas. Largest: fury_2h -118.7, paladin level_60_sub_holy -92.1, level_60 -82.4, fury -84.7, Ret Shock -59.3, arms -45.2, Affliction +36.7, Survival -27.1, hunters -9 to -22, rogues -10 to -19, feral-cat +14.5. Logged under results-20261104-2 (update 2026-11-12).
- Code changes since a0b4d3be9 in watched paths: sim/core/pet.go, unit.go, druid/rake.go, hunter/pet*.go (explain the druid and hunter moves, already republished), report_hashes.txt, and mksettings (leveling target for levels other than 20 and 30 only; level 60 output unchanged, drift confirms).
- Drift (regen.py): only the known feral Goblin Sapper difference (results-20261105-1), now on both feral-cat and feral-cat-duty-ff (reference had it only on the duty one). No new drift.
- Unread fields: none printed (run with the master tree; protos present).
- Not done: no paired before/after sims for the druid and hunter moves; builds.json equality with the run is the check, and the commits state the causes.
- Open: results-20261105-1 (low), results-20261104-2 (low, metas), results-20261104-1 (low).
