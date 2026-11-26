# Consumables pass 112 (master ab3fff700)

Result: quiet, no new findings. Open findings 20261125-1, -2, -3, 20261110-1, 20261105-1 unchanged.

## Moved since bfbee78fc (11 files)
- sim/common/leveling_enchants.go (new): Weapon - Recovery (8721) and Iron/Mithril/Thorium shield spikes. Tank or leveling only; no DPS preset uses them.
- forever_enchants.go: entries for those enchants. No DPS preset affected.
- sim/core/consumes.go: explosives, Dragonbreath Chili, Dragonfire and Caustic Smog now use TargetsNear(radius). Single-target presets unchanged (stun-immune rule kept).
- debuffs.go: SunderArmorAuraOfRank added (warrior lower ranks); Sunder values unchanged.
- ui/tank_warrior/presets.ts: out of scope.
- Hunter SV and warrior Arms/Fury Phase 0 gear: trinket/weapon swaps only, enchants unchanged. Heart of Wyrmthalak (22321, LBRS drop), Thrash Blade (17705, Desolace vendor) are not raid-only. The nonexistent Blackhand's Breadth (13965) is gone. Gear pass owns legality.

## scan.sh
Banned list: only Phase 0 enchants 2544/Revelation (gear pass owns), as last pass. Revelation-outside-shamans, trinkets, crit-elixir lists empty. Minor-armor count 3 (known, 20261125-2). Totem and Sunder/MobType lines unchanged.
