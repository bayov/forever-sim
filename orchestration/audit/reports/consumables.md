# Consumables pass 114 (master f9fa79917)

Result: quiet, no new findings. Open: 20261125-1, -2, 20261110-1 (unchanged).

## Moved since 2af970d74 (25 files)
- Tank, healer and restoration level60_phase0* gear sets (feral_tank, healing_priest, holy_paladin, protection_paladin, restoration_shaman, tank_warrior, warden_shaman): new files, out of scope for this pass.
- ui/{protection_paladin,tank_warrior,warden_shaman}/presets.ts: tank consumable swaps (Zanza over Scorpok Assay, Flask of the Titans over Supreme Power) and race text; out of scope.
- orchestration/standards.md: tanks and healers now wear Phase 0 lists (1b, 8a, 35, 36, 43, 29); no consumable rule changed.
- No DPS presets.ts, mksettings, sim/core, enchant or assumptions file moved.

## Verified
- scan.sh: banned list shows only the Phase 0 sets' 2544 / 1248805 (baseline, gear pass owns). Counts: minor-armor effects 3 (known, in 20261125-2); air/weapon totems unchanged.
- Client build 1.60.1.71139, unchanged per the gate.
