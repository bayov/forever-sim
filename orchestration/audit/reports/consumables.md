# Consumables pass 97 (master 8b5137c6d6, client 1.60.1.71037)

Result: quiet, no new findings. Wago has no build newer than 71037.

Checked (diff b4b68c52e..8b5137c6d):
- Enhancement Phase 0 gear set (new, level60_phase0.gear.json): enchants 2543, 1249019, 1219587, 1891, 1220624, 1248640, 1506, 1887, 1900 all exist in the enchant tables and match the Phase 0 BiS list. Real (non-synthetic) trinkets 10455 and 13968 trip scan.sh NONSYN; expected, since Phase 0 is real pre-raid gear, not a Phase 1 synthetic set (standards 1b applies to Phase 1). Arcanum of Rapidity (item 18329) shows NO SOURCE in the db2 tables, so its Phase 0 availability is unverified (not guessed).
- Enhancement presets.ts: only the Phase 0 gear, talents and build added; consumes, buffs unchanged.
- Ret presets.ts: comment only. buffs.go: Windfury Totem comment only. assumptions_def.py: paladin assumptions removed, not buff related. item_effects.go (Barov Peasant Caller, Vilerend Slicer 225) and crafted.go (Stormshroud 3% / 18 energy): item effects, not consumables; match the cited hotfixes.
- scan.sh: banned, Revelation, crit-elixir checks empty. Minor-armor exclusive effect count reads 3 (Faerie Fire, CoR and the Dust Cloud aura); debuffs.go unchanged since the last pass, so not new.
- Air/weapon totem counts as the standing rules.

No open items beyond the two low findings already in findings/consumables.md.
