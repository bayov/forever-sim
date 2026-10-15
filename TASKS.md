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
- [ ] 2026-10-14 — Optimize for a leveling journey instead of a single encounter (user idea)
  - The score would be kill speed over a whole level range, not encounter DPS. It counts downtime: how long we sit between pulls to get HP and mana back. It also counts how many mobs we can pull at once and survive.
  - Talents are picked at each level, so the search finds the best order to spend points in, not just the final build.
  - It would let us compare paths, for example whether respeccing at some level beats keeping one build the whole way.
- [ ] 2026-10-14 — Whole raid sim (user idea)
  - Sim every member of a raid together, so we can try out raid comps and see the total raid damage.
  - Then optimize the comp itself: which classes and specs, and how buffs, debuffs and totems are spread across groups.
- [ ] 2026-10-14 — Compare build 1.60.1.70374's item tables with 70338's (new client build, first seen Oct 13 23:25 UTC; no patch notes)
  - The spell tables are already byte-identical to 70338's (`orchestration/rank_tuning.md`, check 2026-10-14 00:30 UTC). Trackers say talents, spells and items are unchanged; wowclassicforever.info lists only ManifestInterfaceData and ManifestInterfaceItemIcon as changed tables.
  - Still to diff: https://wago.tools/db2/ItemSparse?build=1.60.1.70374 , https://wago.tools/db2/ItemEffect?build=1.60.1.70374 , https://wago.tools/db2/SpellItemEnchantment?build=1.60.1.70374 , https://wago.tools/db2/Talent?build=1.60.1.70374
  - The fix build expected Thursday Oct 15 (forever-bugs "fixed on Thursday" list) is the one to watch: re-run the full table diff then.
  - 2026-10-15 17:10 UTC: no newer build yet. wago.tools lists 1.60.1.70374 (wow_cn_beta, Oct 13) as the newest, plus a new `wowf` product at 1.60.1.70245 (first seen Oct 13, probably the live client's product).
- [x] 2026-10-14 — Decide whether Searing and Magma Totem crits should stop proccing Elemental Devastation (forever-bugs #334, Oct 13, NAB, tracker)
  - Settled, no change needed: Searing and Magma Totem's attack spells have `ProcMaskEmpty` in `sim/shaman/fire_totems.go`, so they never match the talent's `ProcMaskSpellDamage` check, and Fire Nova is excluded by spell code. No totem attack procs Elemental Devastation in the sim, as #334 says.
  - AidanZMoon: "None of your Totem or Totem originating attacks cause Elemental Devastation"; the tooltip is updated for Thursday's build. `applyElementalDevastation` in `sim/shaman/talents.go` excludes only Fire Nova (shaman_audit.md 4.13).
  - Check on the beta after Oct 15: does a Searing Totem crit give buff 30165? https://github.com/ClassicWoWCommunity/forever-bugs/issues/334
- [ ] 2026-10-14 — Give Crown of Maldazzar 284065 and Triune Amulet 7722 attention once PR #20 merges (revealed Oct 12–14)
  - Crown of Maldazzar has no known source (BoE); add one when it's known. https://foreverchanges.pro/item/284065
- [ ] 2026-10-15 — Model the item procs revealed on Oct 14–15 (client, foreverchanges.pro hidden items). The stats went in with PR #37.
  - Galgann's Firehammer 9419 (main hand mace, +54 spell power, req 42): "Chance on harmful spell cast to blast a target for 64 to 90 Fire damage (100ms cooldown)". The proc rate is unknown. https://wago.tools/db2/ItemEffect?build=1.60.1.70374&filter[ParentItemID]=9419 , https://foreverchanges.pro/item/9419
  - Galgann's Fireblaster 9412 (gun, 44–84 / 2.60, req 42): its ranged chance-on-hit Fire Blast went from 12–18 to 22–34. The sim doesn't model the proc. https://foreverchanges.pro/item/9412
  - Suggestion: compare Firehammer with the Elemental level 60 synthetic staff (55 SP) plus an off-hand in the gear search.
- [ ] 2026-10-15 — On the next beta build, check these tracker fixes against the sim (forever-bugs, tracker only)
  - #331: hunter pets use the hunter's **ranged** crit "starting Thursday" (Oct 15). `sim/hunter/pet.go:243` inherits `MeleeCrit`. https://github.com/ClassicWoWCommunity/forever-bugs/issues/331
  - #20: the Bear Form armor multiplier will no longer apply to temporary armor buffs ("fixed in the next build"). Examples: Elixir of Minor Defense gives 140 armor in Bear Form vs 50; Mark of the Wild rank 1 gives 95 vs 34. Check whether the feral tank sim multiplies buff armor. https://github.com/ClassicWoWCommunity/forever-bugs/issues/20
  - #344 (Shaman, Oct 14): Eye of the Storm's pushback text is now Vanilla-style. It's a wording fix only; check that the sim's pushback model matches.
