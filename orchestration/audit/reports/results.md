# Results pass 92, master f30078da1 (2026-11-16)

Quiet. Changes since cfc7f9382 in watched paths:
- sim/core/sim_concurrent.go: adds a player's own database once before the splits (race fix); no DPS effect. Shared run, 68 builds, 20k seed 1.
- warrior metas: only isDefault flipped (Fury 2H is the default, ruling 2026-11-15); ui/warrior/presets.ts defaults follow. No request changed.
- assets/database, hidden_items, forever_seen.go: item database only; no request references changed items, so no number moved.
- builds.json vs shared run: 68 of 68 within 1.5. Metas off by more than 0.5: 28 of 68, same families as results-20261104-2.
- Requests and mksettings unchanged, so the drift report is unchanged (feral sapper line, results-20261105-1).
