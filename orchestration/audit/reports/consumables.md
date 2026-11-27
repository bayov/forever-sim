# Consumables pass report

Pass 116, master 5ab589d85. Result: quiet (no new findings).

- Moved: sim/common/item_effects.go (Talisman of Ascendance now charges on every APL cast; comment cites classic 1.15.7.61186 data; no consumable, buff or enchant value) and the new ui/restoration_druid Phase 0 gear set (healer, out of scope).
- Client: wago latest is still 1.60.1.71139; no build diff needed.
- scan.sh: banned list shows only Phase 0 set enchants (2544, 1248805), which the gear pass owns; minor-armor exclusive count 3 is the known by-design change (consumables-20261125-2).
- Open items: consumables-20261125-1, -2 and consumables-20261110-1 remain open.
