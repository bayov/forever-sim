# Results pass 99, master 5518f4873f (2026-11-23)

Full pass on the shared run (/tmp/audit-work/runs/master, 68 requests, 20k, seed 1).

- builds.json vs run: 68 of 68 within 1.5 (mage Frost 838.4 vs 838.6, Frostfire 1008.1 vs 1007.3, after 98dc8871e).
- Metas vs run: 61 of 68 off by more than 0.5. Known lag (results-20261104-2); mage metas still 944.7 and 1064.3.
- Sim changes since 7a6d1beb9d (forever_armor_uses.go, item_effects.go, no-weapon swing and wait guards): no request moves.
- Drift (regen.py): only the known feral sapper line (results-20261105-1).
- Result: quiet; no new finding.
