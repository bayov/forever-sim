# Results pass 111 (master 34994816e)

Range f9fa79917..34994816e (29 watched files). Sim change: sim/common/item_effects.go (Cloud of Chained Essence ticks roll a fixed 2.7% crit, 5591bd019).
Gear/APL files: Phase 0 lists without the Armaments of Battle rewards (6bd2d60a2), Enhancement trinket rerank (3698a73d7). Request change: none; only paladin__level_60_shock_ret.meta.json notes (Grayston Shockadin gap).
No change to proto, mksettings, database, assets/database, requests or builds.json.

- Shared run (68, 20k, seed 1) vs metas: 31 listed, all within 1.5 (largest -1.5 feral-cat, Lone Wolf), noise-level; none moved by a commit. Only seal-fate-swords carries the trinket; it is not among the moved rows.
- builds.json unchanged; matches the run on the warrior rows checked (fury-duty-sunder 952.8 vs 953.3). compare.py builds against the old Phase 1 req_summary.txt still flags the warrior rows, expected (stale reference).
- Drift: mksettings and requests unchanged, so equal to the reference (regen.py and unread_fields not re-run).
- Open finding results-20261104-2 unchanged (metas within 1.5).
