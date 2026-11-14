# Results pass 89 (master 1bf5feb4b)

Shared run (68 builds, 20k, seed 1) vs published:
- builds.json: 68 of 68 within 1.5 of the run, including the four Enhancement builds (760d4c1e5 re-sim; last pass they lagged by +22 to +24).
- Metas: 41 of 68 off by more than 0.5, all in families already under results-20261104-1/-2 (warrior, Affliction, hunters, feral, rogues, Ret Shock crusader). Enhancement metas match.
- Drift (mksettings regen vs committed): equals the reference except the known feral sapper line (results-20261105-1). No Enhancement drift after the Phase 0 / presets.ts / magma.apl changes.
- Not run: unread_fields.py (needs generated pb.go; no build in this container).
- Watched changes since 3a86d4a6b are the Enhancement re-sim (req files, builds.json, presets, magma APL), the statweight step change (tooling, no sim effect), db.bin/db.json reveals and report_hashes; no unexplained move.
