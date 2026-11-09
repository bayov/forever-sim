# Consumables pass 76 (master 66a243cf8, client 1.60.1.70862, unchanged)

Findings: none new. Open: consumables-20261107-1 (Natural flasks halved in the client; sim still models 60 Stamina and 4-5 bonuses) and consumables-20261105-1 (feral mksettings lacks the Sapper).

Verified since 9ba00c2dc
- Moved audited files: sim/core/buffs.go and tools/dps_report/assumptions_def.py only.
- buffs.go: under Forever the Windfury Totem buff procs from white main hand swings only (ProcMaskMeleeMHAuto), user ruling 2026-11-08, backed by the Nov 5-8 logs (882 procs, none from ~2,770 Stormstrikes). Raid totem presets (melee keep Windfury Totem, Enhancement Grace of Air only) are unchanged and agree between presets.ts and mksettings. No number changed.
- assumptions_def.py: removed level_partial_resist and sham_ftt_stack; no dangling references in tools/dps_report, ui or standards.md.
- scan.sh on master: every section above "== counts" empty. Minor-armor exclusive-effect grep counts 3 (pattern also hits the category variable; no debuff change on master).
- wago: no build newer than 1.60.1.70862. Class and consumes branches listed as ahead are stale diverged copies (gotcha 1); not audited.
- Pass doc note (not edited): the Enhancement standing check still mentions Flametongue Totem rank 4; Enhancement now gets Grace of Air only.

No sims run.
