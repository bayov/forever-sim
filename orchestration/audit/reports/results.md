# Results audit, pass 49 (quiet)

Master 1ab3ec28a0, previous audited 753d81da74.

- Watched paths changed since the last pass: `ui/enhancement_shaman/presets.ts` (Level 30 Leveling preset replaced by Leveling 10-60; encounter and build presets for the leveling mode only). `tools/database/forever_wowhead/class_spells.py` is a scraper only, no regenerated `assets/database/`.
- No request in `tools/dps_report/req/`, no sim, proto, mksettings or database change.
- Shared run (20k, seed 1): `compare_runs` reports 0 of 66 moved more than 0.5 DPS against published.
- Open findings: none. Baseline unchanged.
