# Tasks

Work for forever-sim that needs a person or a local Claude Code session. The daily WoW Forever
update task appends new items to the end of this file when it finds work it can't finish on its
own, such as a new client build whose wago.tools data needs to be pulled and processed.

Tick an item when it's done.

Item format:

```
- [ ] YYYY-MM-DD — <what to do> (<why / source>)
  - <details: build number, wago.tools links, spell or item IDs, sim files, notes.md or audit items>
```


- [ ] 2026-10-13 — Check Baron Charr's Sceptre's fire proc, 35 → 90 on the Forever tooltip (revealed Oct 11–12, client tooltip only)
  - The sim keeps 35: `BaronCharrsSceptre` in `sim/common/item_effects.go`, proc spell 13442. Its ilvl, required level and damage went in with PR #12.
  - https://wago.tools/db2/SpellEffect?build=1.60.1.70338&filter[SpellID]=13442 and https://wago.tools/db2/ItemEffect?build=1.60.1.70338&filter[ParentItemID]=18671
  - https://foreverchanges.pro/item/18671
- [ ] 2026-10-13 — Decide how "+X healing and +Y spell damage" items go into the DB, then add the ones waiting on it (wowhead's Forever data mixes two conventions)
  - Staff of Nobles 3902: `spldmg 7 + splheal 22` for a 29-healing tooltip. Grimlok's Tribal Vestments 9415: `splpwr 10 + splheal 19` for a 29-healing tooltip.
  - Waiting on it: Stone Keeper Spire 274432 (+103 Healing, +34 Spell Damage, +19 Sta, +180 Armor, 81–122 / 3.20, req 44), https://foreverchanges.pro/item/274432
  - Papal Fez 9431: the DB has the Classic healing-only stat. Forever shows +17 Int, +9 Spi, +22 Healing, +8 Spell Damage. https://foreverchanges.pro/item/9431
- [ ] 2026-10-13 — Revealed items left out: unmodelled effects or stats (foreverchanges.pro hidden items)
  - Abomination Chain 274385 (neck, req 37): +10 Sta, +6 AP, +6 AP vs Humanoids, +3 Defense Rating. It needs an item effect like Dull Sawblade's in `sim/common/item_effects.go`, and the Defense Rating conversion is unknown. https://foreverchanges.pro/item/274385
  - Oilskin Leggings 9414 (req 42): new Feral Combat +1 stat (Icy Veins, datamined, Oct 13), not modelled. https://foreverchanges.pro/item/9414
- [ ] 2026-10-13 — Give the items revealed Oct 10–13 a source once one is known, so the gear search can use them (PR #12 adds them with no source)
  - 284064 Cliffbreaker's Pauldrons, 284285 Feardred's Gauntlets, 281661 Well-Worn Pitchfork, 283545 Moonchaser's Headdress, 281896 Surveyor's Gloves, 286569 Firebird Leggings, 286640 High Priestess's Kilt
- [ ] 2026-10-13 — Restorative Totems should also raise Healing Stream Totem's healing from healing power after the next beta build (~Oct 15) (forever-bugs #291, dev-marked fixed)
  - `sim/shaman/water_totems.go` scales only the base heal. Re-check on the new build's SpellEffect data for HST, and on the beta.
