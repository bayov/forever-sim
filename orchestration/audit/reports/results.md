# Results audit, pass 82 (master ca5e35c3f)

Changed since 91ad19535: sim commit bcc5d6cd4 (Shaman: Maelstrom Weapon stacks on every landed melee hit under Forever,
Totem of the Storm 50% a bolt, Flametongue Weapon rank 5 cut by 7), 04bdcdb9e (enchants 41 and 66 take the client values:
Minor Health +2 Stamina, Minor Stamina +3 Stamina), a778e524b (db fold-in of items seen in game). No request, preset,
mksettings, builds.json or meta changed.

Method: shared run (68 builds, 20k, seed 1, CLI ca5e35c3f) against a fresh run of all 68 on a 91ad19535 binary, same
settings (`run_reqs.py`), then the enhancement builds on a bcc5d6cd4 binary to split the commits.

| Build | 91ad19535 | bcc5d6cd4 | ca5e35c3f | move |
|---|---|---|---|---|
| shaman enhancement-imp-gw | 1132.22 | 1129.85 | 1129.85 | -2.4 (-0.21%) |
| shaman enhancement-water-shield | 1099.62 | 1097.28 | 1097.28 | -2.3 (-0.21%) |
| shaman enhancement-level60 | 1146.67 | 1145.64 | 1145.64 | -1.0 (-0.09%) |
| shaman enhancement-mana-tide | 1042.00 | 1042.12 | 1042.12 | +0.1 |

The other 64 builds moved 0.5 or less (`compare_runs.py`: 3 of 68 moved).

Attribution: the whole Enhancement move is bcc5d6cd4; 04bdcdb9e and a778e524b give identical DPS (the enchant change is
Stamina only, so the four moved report hashes carry no DPS). At 5/5 the old 20 PPM already gave a main hand white swing
on a 3.x speed weapon more than 100%, so "every hit" changes little for white swings (not broken down further), and the bolt rule is
0.5 x 1 either way; a 1 to 2.4 DPS change (part of it a changed random stream) is in line with that. Level 60 uses
Flametongue rank 6, which keeps its points, so the rank 5 cut does not reach the report. Explained; no finding.

Not re-run: regen drift and unread_fields (no request, mksettings or proto change in this range). Published vs run is
unchanged from pass 81 (`/tmp/audit-work/runs/vs_published.txt`, 41 of 68 off by more than 0.5, all from
results-20261109-1 and the meta lag in 20261104-1/-2). Ranking unchanged. Open findings unchanged.
