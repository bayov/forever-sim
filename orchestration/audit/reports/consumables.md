# Consumables audit, pass 102

Master c78d983aca, previous audited b48605aa38. Client 1.60.1.71037 (wago latest, unchanged). Result: quiet, no findings.

## Moved
- sim/common/forever_item_procs.go (new): Forever chance-on-hit, equip and use effects for 26 weapons, gauntlets and trinkets. Item ids are not in any preset or level60 gear set (`git grep` on presets.ts and level60 gear: no hits).
- sim/common/item_effects.go: Forever values for Argent Avenger, Bonechill Hammer, undead/demon/elemental/beast AP items, Wraith Scythe, Heart of Wyrmthalak, +1% ranged hit on three items. None are in a Level 60 DPS preset. No consumable, buff, debuff or enchant value changed.
- Open findings (Minor Health name, feral mksettings sapper) unchanged; neither file moved.

## scan.sh
- banned, Revelation, crit elixirs: empty.
- NONSYN: ui/enhancement_shaman/gear_sets/level60_phase0.gear.json trinkets 13965, 10455. New Phase 0 BiS set (2026-11-22/23), real items on purpose; standards 42 says use unsure Phase 0 gear. Not a finding. Source of the two trinkets not checked this pass.
- Count "minor-armor exclusive effects" is 3 vs expected 2 (ref 2): debuffs.go now has the exclusive effect on CoR, FF and a third site (lines 1106, 1135, 1172); debuffs.go not moved since the last pass, so the scan counts a pre-existing form. Not looked into further.
- Totem lines match the rulings.

## Open
No new items. Suggest the next pass update scan.sh's NONSYN allowance for Phase 0 sets.
