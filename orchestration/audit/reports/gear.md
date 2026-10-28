# Gear pass 44 (master 5d27b00150)

- Watched change: f6116e453 added standards item 29 (race-tuned rotations and gear). No other ruling commits since 37a8a8ccd.
- Item 29 gear part: Dwarf (+1% hit racial) sets at the hit cap re-spend the excess hit (at most 10 points) into other non-Stamina, non-Spirit stats. Modelled as bonus stats on the preset gear, not new items. So no synthetic item, gear.json set or enchant is affected; the 1b budget checks are unchanged.
- Gear files (synthetic items, level60 gear.json, enchants, trinkets, items.go) changed since 37a8a8ccd: none. Full audit not needed.
- Open: if a later pass sees race-specific gear as real items or sets, check them against 1b prices and the hit-adjusted rule in gear_targets.md.
