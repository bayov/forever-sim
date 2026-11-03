# Results and spec strength: pass 72, master a0b306984d

- Shared run (73 requests, 20k, seed 1): 72 of 73 within 0.15 of published; 0 moved more than 0.5 vs published builds.
- Only exception: shaman__elemental-level60-mana-tide, 752.83 vs meta 753.2 (open finding results-20261102-1, unchanged).
- Warlock re-sims (3fce3491c, 58a1c3593 Bane of Doom hold) and ranged hunters without Sapper: metas and builds.json agree with the run.
- Drift check (regen.py at a0b306984d): identical to ref/results/drift_report.txt.
- Mage, shaman Restoration, warrior Blood Craze, apl and proto changes: no published number moved.
- Not run: unread_fields.py.
