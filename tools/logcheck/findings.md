# Log audit findings

What the log audit found that we haven't settled yet. The `log-audit` skill fills it in every
hour while its session is open, and we go through it together on `/log-audit` (the skill's
sections 2 and 6 have the format).

## Fix in the sim

## Collect more data

## logcheck issues

## Watching

### Sheet Attack power and Dodge a bit high at level 6

- **Flag**: `sheet` (Attack power +0.50, Dodge % +0.12).
- **Windows**:
  - `WoWCombatLog-101626_165126.txt 23:45:55-23:59:16` (`Player-4620-01872808`, level 6). The sheet has AP 97 and Dodge 4.43%, but the sim has 97.5 and 4.55%. Every other sheet row matches, Armor 170 included.
- **What we think**: the sim's own base stats at low level, not logcheck. The half point of AP may be the game rounding AP down. The dodge may be the Agility to dodge ratio at level 6 (`sim/core/base_stats.go`). The AP moves Melee raw damage by about 0.1, and our dodge doesn't touch our damage rows.
- **Next step**: see if it comes back at other levels. If it does, check the AP rounding and the dodge per Agility by level.
- **Status**: open, first seen 2026-10-18, last updated 2026-10-18.

## Done

### Low level gear missing from the sim's database

- **Flag**: `sheet`, and `Melee: raw` at mob levels 6 and 7.
- **Windows**:
  - `WoWCombatLog-101626_165126.txt 23:45:55-23:59:16` (Bayov-Oom, `Player-4620-01872808`, level 6). The sheet has Armor 170 and AP 97, but the sim has 38 and 48. Melee raw damage is 28.9 avg in 32 hits at level 6 and 30.4 in 18 hits at level 7. The sim's is 3.0 at both levels.
- **What we think**: logcheck built the wrong character. All 9 worn items are left off as "not in the sim's database": Primitive Mantle (154), Ragged Leather Vest (1364), Belt (1369), Boots (1367) and Bracers (1370), Dust-covered Leggings (4921), Battleworn Leather Gloves (4914), Flimsy Chain Cloak (2652) and Primitive Walking Stick (5778). So the sim swings unarmed (3 damage), with too little armor and AP. The sim's item DB leaves out the grey and white starter gear a new character wears. `character.equip` in `tools/logcheck/export.go` skips any item missing from `core.ItemsByID`. Every new character's first windows will flag like this, so their Melee rows don't tell us anything about the sim. The outcome rows and Earth Shock fit.
- **Next step**: When an item is missing from the DB, have `equip` build it from the client's item tables (logcheck already loads them for the wago build). It needs the weapon damage, speed, armor and stats. Another option is to add the low level items to the sim's DB. Until then, the audit could skip a window whose weapon is missing instead of flagging it.
- **Done**: 2026-10-18. gen_db already builds these items, but it writes the grey and white gear and anything below item level 10 to `assets/database/leftover_db.bin`, which only the UI loads. logcheck now adds the leftover items to the sim's item table at startup (`addLeftoverItems` in `tools/logcheck/export.go`). They match the sheet: Kolkar Hammer 5-11 at 2.4 speed, and Armor 170 = 170 at 23:45. With `-redo` the window's `Melee: raw` flag is gone. Only Primitive Mantle (154) is still left off. It's a shirt with no stats.
