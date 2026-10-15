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
  - Design and defaults settled 2026-10-15: `docs/leveling_sim.md` (score: mobs killed in a 20 min session; Enhancement shaman first). The groundwork items below come first.
- [ ] 2026-10-15 — Leveling groundwork 1: damage taken from mobs (`docs/leveling_sim.md`, Groundwork)
  - A table of normal mob HP, armor, damage and swing speed by level (Classic averages to start).
  - Check the player's defenses at low level (dodge per Agility for every class, block, armor), spell pushback (`sim/core/cast.go`) and parry and block from in front.
  - Calibrate against the user's logs: damage taken per mob and the mob's miss, dodge and crit rates. A "solo mob" DPS encounter that reports HP lost per kill and deaths.
  - 2026-10-15: the sim side is done (mob table from the 1.12 world database, death without a healing model, health fight length, pushback cap, tests). Still open: run `go run --tags=with_db ./tools/logcheck -mobs` on the user's machine and compare Forever's mobs with the table; the low-level defense checks; the UI preset. `docs/leveling_sim.md`, Groundwork 1.
- [ ] 2026-10-15 — Leveling groundwork 2: separate mob death (`docs/leveling_sim.md`, Groundwork)
  - Each mob has its own HP, dies when it runs out and stops attacking. The fight ends when all are dead. Today `use_health` only ends the fight on the targets' combined HP.
  - 2026-10-15: done (`sim/core/target_death.go`, `sim/core/aoe.go`), with AoE taking only living targets and helpers for hard and soft caps. Which kind each AoE is in Forever goes to the Forever mechanics auditor (orchestration/auditor_forever.md, AoE scaling). A cast at a mob that dies during it stops unpaid.
- [ ] 2026-10-15 — Leveling groundwork 3: self-healing (`docs/leveling_sim.md`, Groundwork)
  - Shaman first: Healing Wave and Lesser Healing Wave by rank and level. Then Priest (Lesser Heal, Heal, Flash Heal, Greater Heal, Renew; today commented out or missing), Paladin (Holy Light, Flash of Light) and Druid (Healing Touch, Rejuvenation, Regrowth).
  - Absorbs in the core, for Power Word: Shield, Mana Shield and Sacrifice.
  - 2026-10-15: shaman done (`sim/shaman/healing.go`, shaman_audit.md 11.5), with the healing talents. Still open: Priest, Paladin, Druid and absorbs.
- [ ] 2026-10-15 — Leveling mode step 2: move the pull and rest decisions into the rotation (`docs/leveling_sim.md`, The leveling mode)
  - Step 1, the session engine, is done (`sim/core/leveling.go`): pack size, eat_below and drink_below are options for now.
  - APL actions: pull a mob (with a spell or a face pull), chain pull, eat and drink to a threshold, bandage. Values: mobs in combat, mobs running in, in combat.
- [ ] 2026-10-15 — Leveling: measure search, run-in and chain-pull times and pack sizes from the user's logs (`docs/leveling_sim.md`, From the logs)
  - Needs `~/forever-logs` on the user's machine. Drop gaps over about 60 s (AFK) and use the fast quartile as the default.
- [ ] 2026-10-14 — Whole raid sim (user idea)
  - Sim every member of a raid together, so we can try out raid comps and see the total raid damage.
  - Then optimize the comp itself: which classes and specs, and how buffs, debuffs and totems are spread across groups.
- [ ] 2026-10-14 — Compare build 1.60.1.70374's item tables with 70338's (new client build, first seen Oct 13 23:25 UTC; no patch notes)
  - The spell tables are already byte-identical to 70338's (`orchestration/rank_tuning.md`, check 2026-10-14 00:30 UTC). Trackers say talents, spells and items are unchanged; wowclassicforever.info lists only ManifestInterfaceData and ManifestInterfaceItemIcon as changed tables.
  - Still to diff: https://wago.tools/db2/ItemSparse?build=1.60.1.70374 , https://wago.tools/db2/ItemEffect?build=1.60.1.70374 , https://wago.tools/db2/SpellItemEnchantment?build=1.60.1.70374 , https://wago.tools/db2/Talent?build=1.60.1.70374
  - The fix build expected Thursday Oct 15 (forever-bugs "fixed on Thursday" list) is the one to watch: re-run the full table diff then.
- [x] 2026-10-14 — Decide whether Searing and Magma Totem crits should stop proccing Elemental Devastation (forever-bugs #334, Oct 13, NAB, tracker)
  - Settled, no change needed: Searing and Magma Totem's attack spells have `ProcMaskEmpty` in `sim/shaman/fire_totems.go`, so they never match the talent's `ProcMaskSpellDamage` check, and Fire Nova is excluded by spell code. No totem attack procs Elemental Devastation in the sim, as #334 says.
  - AidanZMoon: "None of your Totem or Totem originating attacks cause Elemental Devastation"; the tooltip is updated for Thursday's build. `applyElementalDevastation` in `sim/shaman/talents.go` excludes only Fire Nova (shaman_audit.md 4.13).
  - Check on the beta after Oct 15: does a Searing Totem crit give buff 30165? https://github.com/ClassicWoWCommunity/forever-bugs/issues/334
- [ ] 2026-10-14 — Give Crown of Maldazzar 284065 and Triune Amulet 7722 attention once PR #20 merges (revealed Oct 12–14)
  - Crown of Maldazzar has no known source (BoE); add one when it's known. https://foreverchanges.pro/item/284065
