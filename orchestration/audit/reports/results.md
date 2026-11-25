# Results pass 102, master 63129ce35

Quiet. Since eceb090196 only Phase 0 step 4 requests/metas/presets changed (hunters, shadow priest, rogues); no sim, proto, mksettings or database change.

- Shared run (68 requests, 20k, seed 1) vs metas: 24 within 0.15, 17 off by 0.5 to 1.1, none worse. Rogue and hunter metas now match the run (was -22 to +4).
- builds.json (Phase 1 publication) is 1.5 or more from the run on all 68 (-64 to -188): the Phase 0 gear drop, waiting on the full report re-sim (results-20261104-2).
- Drift vs reference: only the feral sapper line (results-20261105-1); the reference lists it for duty-ff only, the regen shows it on feral-cat too.
- Hunter, rogue and priest requests match mksettings output.
