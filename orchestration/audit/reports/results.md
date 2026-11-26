# Results pass 107 (master bfbee78fc)

Range 85691655a..bfbee78fc. Watched changes: Phase 0 re-rank commits (warlock, mage, hunter, druid, priest, shaman, Ret requests,
metas, presets), 94d6cd2e4 (report re-sim, builds.json), tools/rotopt gear tooling. No sim/, proto, mksettings or assets/database change,
so the binary is behaviourally the one of pass 106. Shared run: 68 builds, 20k, seed 1.

## Run vs metas (published.dps)
25 of 68 off by more than 0.5, largest +27.1. All are request moves from the re-rank commits since the last re-sim, not sim changes:
| Build | meta | run | cause |
|---|---|---|---|
| warlock affliction_shadow_bolt | 793.6 | 820.8 (+3.4%) | c15219593 / 140f0f767 warlock re-rank |
| paladin level_60 family (8) | | +4.8 to +9.1 (0.5 to 1.0%) | 4a39a50be, 1a6600121, 8e536aa0d Ret weapons, trinkets, Undead |
| shaman enhancement (4), elemental (2) | | +2.6 to +4.7 | 4a39a50be, c3884acc6 (Imp GW talents) |
| warlock destruction_fire / destruction duties | | -1.0 to -3.6 | warlock re-rank |
Everything else within 0.5. builds.json: 28 of 68 differ from the run by more than 1.5 (the same families), waiting on the re-sim.

## Drift (regen.py vs reference)
Hunter and feral Dark Bomb drift (results-20261125-1) is gone: requests carry Dense Dynamite / Goblin Land Mine now. New: Survival (3) and
Enhancement (4) carry Land Mine that mksettings omits (results-20261126-1, about -5 DPS). Feral sapper line unchanged (-0.7 now).
Reference drift_report.txt not replaced (orchestrator fix pending). Unread fields: only the known set (script prints NOT-IN-SUMMARY lines for the run dir layout, no UNREAD besides notes).
