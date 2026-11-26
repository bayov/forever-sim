# Consumables pass 113 (master 2af970d74)

Result: quiet, no new findings. Resolved: 20261125-3, 20261105-1. Open: 20261125-1, -2, 20261110-1.

## Moved since ab3fff700 (29 files)
- spec_hunter.go: Survival filler explosive Goblin Land Mine, ranged Dense Dynamite. Matches ui/hunter/presets.ts and the dps_report req files.
- spec_feral.go: SapperGoblinSapper added (comment fixed); the feral filler is the Land Mine as in presets.ts and the requests. Resolves 20261105-1 and, with the req files, 20261125-3.
- spec_enh.go: FillerExplosive Goblin Land Mine; matches presets.ts and the shaman requests.
- barov_servant.go: guardian wait guard (crash fix); no consumable effect.
- level60_phase0* gear sets: item swaps only (trinkets, weapons, rings); no enchant lines changed. Legality is the gear pass's.
- ui/tank_warrior/presets.ts: out of scope.
- scan.sh banned list shows only the Phase 0 sets' 2544 / 1248805, unchanged baseline (gear pass owns).
- Client build 1.60.1.71139, unchanged per the gate.
