# Consumables pass 101 (quiet)

Master 7e0a01a8d9, previous audited 2de7021bf71.

- Findings: none new. Open items unchanged (consumables-20261110-1, consumables-20261105-1, both low).
- Why due: orchestration/standards.md changed (rule 42, Phase 0 gear with unsure availability is used, user ruling 2026-11-23). It concerns Phase 0 gear lists, not Level 60 consumes, buffs or enchants, so nothing here changes.
- Watched paths (presets.ts, mksettings, level60 gear, sim/core consumes/buffs/debuffs/totems, sim/common, forever_enchants.go, assumptions_def.py, staff imbue): `git diff 2de7021bf71 HEAD` is empty.
- Wago: latest 1.60 build 1.60.1.71037, same as the last pass.
- Class branches: changes.sh lists them as thousands of commits ahead, all stale copies (gotcha 1); not audited.
