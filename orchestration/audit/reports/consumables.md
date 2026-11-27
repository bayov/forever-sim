# Consumables pass 115 (master 34994816e)

Result: quiet, no new findings. Open findings 20261125-1, 20261125-2 and 20261110-1 are unchanged.

## Checked
- changes since f9fa79917: 18 files. 17 are Level 60 Phase 0 gear sets, with item swaps only (trinkets and gear ids).
  The only enchant changes are in holy_paladin (out of scope: 1505 to 1503, kit 1255074 to Rugged Armor Kit 1843).
  DPS spec enchants are unchanged.
- sim/common/item_effects.go: Cloud of Chained Essence ticks now crit for 1.5x at a fixed 57/2151 chance, from the
  2026-11-26 log audit. Not a consumable or buff value; nothing to check against the client.
- scan.sh: banned hits are only the Phase 0 sets (gear pass owns legality). Counts as expected, minor-armor effects at 3 (known, finding 20261125-2).
- wago latest: 1.60.1.71139, the build already audited.
