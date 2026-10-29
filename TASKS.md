# Tasks

Work for forever-sim that needs a person or a local Claude Code session. The daily WoW Forever
update task appends new items to the end of this file when it finds work it can't finish on its
own, such as a new client build whose wago.tools data needs to be pulled and processed.

Tick an item when it's done.

Item format:

```
- [ ] YYYY-MM-DD — <what to do> (<why / source>) [source: <tag>]
  - <details: build number, wago.tools links, spell or item IDs, sim files, notes.md or audit items>
```

The tag is one of `wago build <N>`, `wowhead`, `foreverchanges`, `blizzard` or `tooltip`
(tools/forever_update/ROUTINE.md says which). `tools/sources/stale.py` finds unprocessed client
builds by the `wago build` tag. Items from before 2026-10-17 have none, and it matches their
wording instead.


- [x] 2026-10-13 — Check Baron Charr's Sceptre's fire proc, 35 → 90 on the Forever tooltip (revealed Oct 11–12, client tooltip only)
  - Done 2026-10-23: Firebolt 13442 is 90 in wowsims/forever's client store, and the sim gives 90 under Forever (`CreateWeaponCoHProcDamageByRuleset`). The item's entry states no rate, so 1 PPM stays.
  - The sim keeps 35: `BaronCharrsSceptre` in `sim/common/item_effects.go`, proc spell 13442. Its ilvl, required level and damage went in with PR #12.
  - https://wago.tools/db2/SpellEffect?build=1.60.1.70338&filter[SpellID]=13442 and https://wago.tools/db2/ItemEffect?build=1.60.1.70338&filter[ParentItemID]=18671
  - https://foreverchanges.pro/item/18671
- [ ] 2026-10-13 — Add the "+X healing and +Y spell damage" items waiting on the convention, now decided
  - Decided 2026-10-18 (`spellStats` in `tools/database/forever_wowhead.go`): healing is `splpwr + splheal` and damage is `splpwr + spldmg`. All 184 Forever items with spell stats that BayovCore's item sweep recorded match it.
  - Staff of Nobles 3902 is the one doubt: `spldmg 7 + splheal 22`, noted on 2026-10-13 as a 29-healing tooltip, which the rule makes 22. Nothing on record says where the 29 came from. The next full sweep reads Classic IDs too, and bayovui's `seen_items.py` will list it if the game disagrees. Grimlok's Tribal Vestments 9415 (`splpwr 10 + splheal 19`, 29 healing) fits the rule.
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
  - 2026-10-17: the Leveling Plan's Talent search (`docs/leveling_plan.md`, Talent search) scores whole talent routes by the plan's total hours, and the best respec at 20, 30, 40 and 50 against keeping the route. Still open: rotations searched along with the talents.
  - 2026-10-18: the default enhancement plan goes from 10 to 60 (grinding up to 59), with the talents on to the Level 60 + Imp GW build. Still open: rotations and stat weights past 30. A rotation search has to know the spells and ranks of each level (Chain Lightning at 32, new ranks mid range).
  - 2026-10-18: the Talent search tunes the rotations too (Rotations too, `docs/leveling_plan.md`, Rotation search), a range at each level with a new ability or rank, in turns with the route. The leveling template moved to `tools/rotopt/aplkit` and has rules past 30 (Maelstrom Weapon bolts, Chain Lightning, air totems). Still open: stat weights past 30, templates for the other specs.
- [ ] 2026-10-15 — Leveling groundwork 1: damage taken from mobs (`docs/leveling_sim.md`, Groundwork)
  - A table of normal mob HP, armor, damage and swing speed by level (Classic averages to start).
  - Check the player's defenses at low level (dodge per Agility for every class, block, armor), spell pushback (`sim/core/cast.go`) and parry and block from in front.
  - Calibrate against the user's logs: damage taken per mob and the mob's miss, dodge and crit rates. A "solo mob" DPS encounter that reports HP lost per kill and deaths.
  - 2026-10-15: the sim side is done (mob table from the 1.12 world database, death without a healing model, health fight length, pushback cap, tests). Still open: the low-level defense checks; the UI preset. `docs/leveling_sim.md`, Groundwork 1.
  - 2026-10-16: spell pushback checked in the logs (casts match the sim's rule, `tools/combat_log/pushback.py`), and channels now take pushback too.
  - 2026-10-16: `logcheck -mobs` on the user's logs. Forever's mobs have up to 30% more health than 1.12's (the table uses it now), and they crit us far less than Classic's 5% (Need to Verify). `logcheck -pacing` measured the session pacing, and the search default is now 2 to 7 s (`docs/leveling_sim.md`, From the logs).
- [ ] 2026-10-15 — Leveling groundwork 2: separate mob death (`docs/leveling_sim.md`, Groundwork)
  - Each mob has its own HP, dies when it runs out and stops attacking. The fight ends when all are dead. Today `use_health` only ends the fight on the targets' combined HP.
  - 2026-10-15: done (`sim/core/target_death.go`, `sim/core/aoe.go`), with AoE taking only living targets and helpers for hard and soft caps. Which kind each AoE is in Forever goes to the Forever mechanics auditor (orchestration/auditor_forever.md, AoE scaling). A cast at a mob that dies during it stops unpaid.
- [ ] 2026-10-15 — Leveling groundwork 3: self-healing (`docs/leveling_sim.md`, Groundwork)
  - Shaman first: Healing Wave and Lesser Healing Wave by rank and level. Then Priest (Lesser Heal, Heal, Flash Heal, Greater Heal, Renew; today commented out or missing), Paladin (Holy Light, Flash of Light) and Druid (Healing Touch, Rejuvenation, Regrowth).
  - Absorbs in the core, for Power Word: Shield, Mana Shield and Sacrifice.
  - 2026-10-15: shaman done (`sim/shaman/healing.go`, shaman_audit.md 11.5), with the healing talents. Still open: Priest, Paladin, Druid and absorbs.
- [x] 2026-10-15 — Leveling mode step 4: move the rest decisions into the rotation (`docs/leveling_sim.md`, The leveling mode)
  - Done (2026-10-15): Rest (Leveling) and Bandage (Leveling) actions, Bandage Ready, the ready state between pulls; health regen from Wrath's game tables, a third more sitting, Troll Regeneration.
- [ ] 2026-10-15 — Leveling: open items after steps 1 to 4 (`docs/leveling_sim.md`)
  - Measure a shaman's out of combat health ticks on the beta (notes.md, Need to Verify): two Spirit values under 50 and one over, at two levels, standing and sitting.
  - A bolt's flight time to a chain pull target that isn't our current target uses our current target's distance.
  - Stoneskin Totem's client data also has a 20 yard effect next to the 30 yard one; we use 30 for all of it.
- [ ] 2026-10-15 — Leveling: measure search, run-in and chain-pull times and pack sizes from the user's logs (`docs/leveling_sim.md`, From the logs)
  - Needs `~/forever-logs` on the user's machine. Drop gaps over about 60 s (AFK) and use the fast quartile as the default.
- [ ] 2026-10-14 — Whole raid sim (user idea)
  - Sim every member of a raid together, so we can try out raid comps and see the total raid damage.
  - Then optimize the comp itself: which classes and specs, and how buffs, debuffs and totems are spread across groups.
  - 2026-10-29: design in `docs/raid_sim.md`, with the user's first rulings (own page, members from the spec pages, tanks and healers and threat out of scope for now, configurable tank and healer slots, no debuff limit, assignments, comp search as the goal). Groundwork steps 1 to 7; four open decisions left.
  - 2026-10-29: reviewed (buffs and debuffs against client 70685, engine, comp search with measurements, page). The doc now has the provider table, the engine bugs found by running raids of report builds (aura cap at 20 players, two of one build, gaps), measured speed and noise, a model-then-confirm comp search, and groundwork steps 1 to 11. Three open decisions: range, hunter pet families, where big searches run.
- [ ] 2026-10-14 — Compare build 1.60.1.70374's item tables with 70338's (new client build, first seen Oct 13 23:25 UTC; no patch notes)
  - The spell tables are already byte-identical to 70338's (`orchestration/rank_tuning.md`, check 2026-10-14 00:30 UTC). Trackers say talents, spells and items are unchanged; wowclassicforever.info lists only ManifestInterfaceData and ManifestInterfaceItemIcon as changed tables.
  - Still to diff: https://wago.tools/db2/ItemSparse?build=1.60.1.70374 , https://wago.tools/db2/ItemEffect?build=1.60.1.70374 , https://wago.tools/db2/SpellItemEnchantment?build=1.60.1.70374 , https://wago.tools/db2/Talent?build=1.60.1.70374
  - The fix build expected Thursday Oct 15 (forever-bugs "fixed on Thursday" list) is the one to watch: re-run the full table diff then.
  - 2026-10-15 17:10 UTC: no newer build yet. wago.tools lists 1.60.1.70374 (wow_cn_beta, Oct 13) as the newest, plus a new `wowf` product at 1.60.1.70245 (first seen Oct 13, probably the live client's product).
- [x] 2026-10-14 — Decide whether Searing and Magma Totem crits should stop proccing Elemental Devastation (forever-bugs #334, Oct 13, NAB, tracker)
  - Settled, no change needed: Searing and Magma Totem's attack spells have `ProcMaskEmpty` in `sim/shaman/fire_totems.go`, so they never match the talent's `ProcMaskSpellDamage` check, and Fire Nova is excluded by spell code (since 2026-10-23 its hits are a proc, `ProcMaskSpellDamageProc`, so the check skips them without a special case; wowsims/forever 3c39b8afe6). No totem attack procs Elemental Devastation in the sim, as #334 says.
  - AidanZMoon: "None of your Totem or Totem originating attacks cause Elemental Devastation"; the tooltip is updated for Thursday's build. `applyElementalDevastation` in `sim/shaman/talents.go` excluded only Fire Nova (shaman_audit.md 4.13); since 2026-10-23 Fire Nova's hits carry a proc mask instead.
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
  - #344 (Shaman, Oct 14): Eye of the Storm's pushback text is now Vanilla-style. It's a wording fix only; check that the sim's pushback model matches. 2026-10-16: it does. The sim's pushback rule (1 sec, then 0.8 and so on) fits the beta logs (`tools/combat_log/pushback.py`), and Eye of the Storm's chance applies per hit.
- [ ] 2026-10-16 — Give the 12 rares revealed October 15 to 16 a source once one is known, so the gear search can use them (PR #107 adds them with no source)
  - 286983 Ancient Cleaver, 286996 City Guard's Tarnished Blade, 287068 Excavated Dagger, 287070 Skypierce Javelin, 286984 Mark of the Marsh, 286995 Guardian's Emblem, 286994 Tattered Kirin Tor Cloak, 286992 Construct's Grips, 286990 Arcane Remnant Leggings, 286993 Suffused Bindings, 286985 Thicket Bindings, 286986 Abandoned Leggings
  - All bind on equip, levels 26 to 40. The Kirin Tor and city guard names point to a new Forever dungeon (City of Dalaran?), unconfirmed.
  - Then re-run the Level 30 gear searches: Ancient Cleaver (2H axe 80–120 / 3.60, +12 Str, +14 SP) for Enhancement and Ret, Mark of the Marsh and Thicket Bindings for Combat.
- [ ] 2026-10-16 — Revealed items left out of PR #107 (foreverchanges.pro hidden items, October 15 to 16)
  - Unearthed Barrier 286982 (shield, 699 armor, 15 block, +8 Sta, req 27): "Deals 6 to 38 Nature damage every time you block. Damage is doubled in swamps." (wowhead: 6 to 39). The block proc is not modelled. https://foreverchanges.pro/item/286982
  - Roused Nature Footwraps 286991 (leather feet, 73 armor, +6 Sta, +9 Spi, +13 Healing, +4 Spell Damage, req 27): waits on the "+X healing and +Y spell damage" item above. https://foreverchanges.pro/item/286991
  - Low level thrown knives with no damage on foreverchanges (wowhead may have it): 287059, 287060, 287061, 287062, 287063, 287064, 287066, 287067, 287069 Yeti Skinner (+7 AP), 287071 Goblin Switchblade (+3 Agi, +10 AP vs Beasts). https://foreverchanges.pro/item/287071
- [ ] 2026-10-16 — shaman_audit.md: Restorative Totems now scales all of Healing Stream Totem's heal under Forever (official, Oct 15 development notes, client 70425; implemented in PR #107)
  - Update the Healing Stream Totem note once #107 merges, then tick the 2026-10-13 Restorative Totems item above.
- [ ] 2026-10-16 — Grounding Totem breaks on non-damaging spells in 70425; a beta hotfix is promised (forever-bugs #354, AidanZMoon "Will be hotfixed into beta", tracker)
  - Not simulated (PvP and utility). Check after the hotfix whether anything in the sim or the docs assumes the old rule. https://github.com/ClassicWoWCommunity/forever-bugs/issues/354
  - Also new on the tracker, no dev reply yet: #370 Elemental Reach lost its Flame Shock range (the 70425 tooltip agrees), #358 Ghost Wolf, Cure Poison and Cure Disease use up Water Shield charges.
- [ ] 2026-10-17 — Give Ankle Slicers 284587 a source once one is known, and decide on its slow proc (foreverchanges.pro hidden items, October 15 to 16 group, revealed after PR #107; client)
  - Epic thrown, ilvl 63, req 58, BoE: 51–96 / 1.80, +9 Sta, Equip: +12 Attack Power. Chance on hit: Sliced Ankle (spell 1316346), 30% movement slow for 6 sec, not modelled (no DPS effect).
  - Its stats are already in db.json from the Oct 17 Wowhead refresh (b1b0763), with no source. Ranged-slot option for Rogue and Warrior at 60 (+12 AP). https://foreverchanges.pro/item/284587
- [ ] 2026-10-20 — Decide on Pendulum of Doom 9425's Forever proc and new equip (revealed October 19 to 20, client tooltip only; left out of PR #115) [source: foreverchanges]
  - Chance on hit: Classic's 250–350 → 259–363 Physical, and now "Deals 50% increased damage to targets below 25% health". `PendulumOfDoom` in `sim/common/item_effects.go` still has `250, 100` (proc spell 10373). Wowhead's Forever tooltip agrees.
  - Proc done 2026-10-23: Fatal Wound 10373 is 311 with variance 0.333 in wowsims/forever's client store, rolled the whole variance either side as a physical hit (their 0dfbb7b009, measured on Revenge), 207–415, with 50% more below 25% health; the tooltip's 259–363 has the same average. The item still states no rate (0.5 PPM stays). The equip below is open.
  - New Equip: "Each swing alternates your attack speed between 100% faster and 50% slower" (4.00 speed, so 2.00 then 6.00 in turn). Needs new code; average speed stays 4.00, but swing timing and Windfury/procs per swing change.
  - https://foreverchanges.pro/item/9425 , https://wago.tools/db2/ItemEffect?build=1.60.1.70450&filter[ParentItemID]=9425 , https://wago.tools/db2/SpellEffect?build=1.60.1.70450&filter[SpellID]=10373
- [ ] 2026-10-20 — Axe of Inevitable Ruin 274484: model its chance on hit and give it a source once known (revealed October 19 to 20; stats in PR #115) [source: foreverchanges]
  - 2H axe 89–134 / 3.00, req 37, BoP. Chance on hit: "Wounds the target for 140 to 196 damage" (proc rate unknown). Use: combines with Axe of Lingering Dread 274485 (still hidden, ilvl 44, req 39) into Pendulum of Doom. Probably Uldaman, unconfirmed.
  - https://foreverchanges.pro/item/274484 , https://wago.tools/db2/ItemEffect?build=1.60.1.70450&filter[ParentItemID]=274484
- [x] 2026-10-21 — Check the Oct 20 mage Frost talent hotfix against the client and apply it (foreverchanges.pro "Hotfix", first seen October 20 on build 70450; not in Blizzard's notes) [source: foreverchanges]
  - Elemental Precision: 5 ranks of 1% → 3 ranks of 2% (6% max). `sim/mage/talents.go` · `applyFrostTalents` uses 1% per point; `ui/core/talents/trees/mage.json` has maxPoints 5 (spells 29438, 29439, 29440).
  - Piercing Ice (11151, 12952, 12953, 2/4/6% Frost damage) removed from the tree. Frost Channeling moves row 3 column 2 → 1, Improved Blizzard row 3 column 4 → 2.
  - Ice Barrier rank 1 absorb 448 → 493; Frost Ward rank 1 162 → 178, plus 10% of Frost spell power. Frost Warding raises both absorbs by 10/20%, and its Frost Ward chance now negates instead of reflecting.
  - Done 2026-10-23: confirmed in wago.tools' hotfix table (push of 2026-10-20 09:51 on 70450: TraitNode 105773 deleted, 105771/105772/105774/105776 moved, TraitNodeEntry 130507 MaxRanks 3, curve 2/4/6, Spell and SpellEffect rows of 11189, 11426-13033 and the new 1326891 Frigid Power) and in wowsims/forever 983ea6fcc1. Frostbite and Permafrost also swapped columns, which foreverchanges didn't list. The tree, proto, `client/mage.json` (by hand, the client tables don't have it), Elemental Precision, Ice Barrier (a real absorb with 3% while shielded, Frigid Power 3% for 1 min, Frost Warding +10/20%) and the presets are updated, and the Frost rotation casts Ice Barrier (forever-wiki/classes/mage.md). Frost Ward isn't simulated. The same push also changed Winter's Chill (11180, 12579): 5% Cone of Cold damage taken and 7% of spell power a stack, modelled with Cone of Cold on 2026-10-23 (`sim/mage/cone_of_cold.go`; the Frost preset stays at range, the melee-range variant is the pending assumption `mage_cone_of_cold`).
  - Wowhead's Forever tooltips (dataEnv 16) still show the old values today, and wago.tools returned 503. Check https://wago.tools/db2/Talent?build=1.60.1.70450 , https://wago.tools/db2/SpellEffect?build=1.60.1.70450&filter[SpellID]=29438 and the hotfix table, then re-import the mage tree (`tools/forever_talents/import_client.py`). https://foreverchanges.pro/patch-notes
- [x] 2026-10-22 — Decide on the Oct 21 Hunter's Mark hotfix: rank 4 gives 143 ranged attack power, not 71 (foreverchanges.pro "Hotfix", first seen October 21 on the October 16 build; not in Blizzard's notes) [source: foreverchanges]
  - The sim assumes 140 (`ForeverHuntersMarkRank4` in `sim/core/debuffs.go`, report assumption hunter_tsa_hm_rank, user ruling 2026-10-15). If the hotfix holds, set it to 143 and rerun the hunter and physical-ranged presets (+3 RAP against the marked target).
  - Same hotfix: Lava Burst rank 3 now gains 20% from Flame Shock instead of 54%, like the other ranks. The sim already uses 20% for every rank (`LavaBurstFlameShockBonus`), so nothing to change.
  - Client 70532 is identical to 70450, so the change is server side. Check https://wago.tools/db2/SpellEffect?build=1.60.1.70532&filter[SpellID]=14325 and the hotfix table, or the mark's tooltip in game. https://foreverchanges.pro/patch-notes
  - Done 2026-10-23: the user approved 143. `ForeverHuntersMarkRank4` is 143; hunter_tsa_hm_rank (spec_assumptions.md), standards 8a, notes.md and forever-wiki/classes/hunter.md say so. Trueshot Aura rank 5 stays at the assumed 100. Level 60 presets at 180 sec: Lone Wolf +0.8 (the base presets take the mark from the raid's debuff; the golden suites have no Hunter's Mark, so they don't move).
- [ ] 2026-10-23 — Decide on Coldrage Dagger 10761's Forever proc: 20–30 → 34–50 Frost damage (revealed October 21 to 23, foreverchanges.pro and Wowhead Forever tooltips; the client's spell data doesn't match either) [source: foreverchanges]
  - The sim keeps Classic's 20–30: `itemhelpers.CreateWeaponCoHProcDamage(ColdrageDagger, ..., 13439, ..., 20, 10, ...)` in `sim/common/item_effects.go`. The dagger's other stats (30–57 / 1.50, +3 Frost Resistance, req 38) already match db.json.
  - Client 70532 SpellEffect for spell 13439, effect 1 (school damage): EffectBasePointsF 50, Variance 0.4, which reads as 40–60, not the tooltip's 34–50. Settle it with a proc in a combat log, or the hotfix table. https://wago.tools/db2/SpellEffect?build=1.60.1.70532&filter[SpellID]=13439 , https://foreverchanges.pro/item/10761
- [ ] 2026-10-23 — Paladins learn Righteous Fury at level 10 instead of 16 (foreverchanges.pro "Hotfix" to the October 22 build, first seen October 23; not in Blizzard's notes) [source: foreverchanges]
  - The sim has no level gate on it (`sim/paladin/righteous_fury.go`, an option). `forever-wiki/data/abilities.json` still says "Level 16" (from the client's tables, so the change is server side). Check the trainer level in the next `class_spells.py` scrape, for BayovCore's Training page. https://foreverchanges.pro/patch-notes
- [ ] 2026-10-25 — Decide how to read The Rockpounder 9413's Forever "+28 Critical Strike Rating" (revealed October 21 to 23; Classic had "+2% crit"; foreverchanges.pro, seen in game) [source: foreverchanges]
  - Everything else matches Classic (126–190 / 3.70, +5 Str, ilvl 49, req 44, Archaedas, Uldaman). The DB has Classic's 2% melee crit (`StatMeleeCrit` 2, hiddenStats).
  - If Forever converts 14 rating to 1% at level 60, 28 is the same 2% and nothing changes; at lower levels a rating would be worth more. Like Abomination Chain 274385's "+3 Defense Rating" (2026-10-13 item), the conversion is unknown, so it waits on one.
  - Wowhead has no Forever tooltip for it yet. https://foreverchanges.pro/item/9413
- [ ] 2026-10-27 — Test Amplify Curse on Bane of Agony in game before any preset or rotation takes it (client 70630 is contradictory; user: "something we need to test eventually") [source: wago build 70630]
  - 70630: 18288 effect 0 (+50%) mask 33792 → 32768 drops Agony (Curse of Weakness only), but 18288's own text and Wowhead's Forever tooltip still say "Curse of Weakness or Bane of Agony by 50%", and every Agony rank (980, 1014, 6217, 11711, 11712, 11713) gains a base-0 effect 1, aura 271 (damage taken from the caster), which Agony's new debuff text multiplies by (`$o1*($w2/100+1)`). So the +50% likely moved onto Agony as a server-filled value.
  - Test: an Amplify Curse'd Bane of Agony vs a plain one, at two spell power values. Is it +50%, of base only or of the whole tick (spell power included), or nothing?
  - Sim keeps +50% of base damage (`sim/warlock/curses.go:74`). No preset or APL casts Amplify Curse, so the report doesn't move. Open findings client_update-20261027-3 / mechanics-20261027-2; ledger WL-U-17.
- [ ] 2026-10-27 — Pull wago.tools DB2 data for build 1.60.1.70634 and process it into the sim (new since 70630; the routine's diff found nothing sim-relevant) [source: wago build 70634]
  - `db2diff.py 70630 70634`: 0 spells changed. Spell, SpellEffect, SpellMisc, SpellAuraOptions, SpellCooldowns, Item and ItemSparse are identical. Bump `foreverClientBuild` and the enchant, itemsparse and racials meta builds when the next source pull runs.
  - It still reads Blood Fury 20572 cooldown 180000, Berserking 26297 10 sec, and Touch of the Grave 1260189/1260201 ProcCategoryRecovery 25000, so the 10-27 notes below are server side. https://wago.tools/db2/SpellCooldowns?build=1.60.1.70634&filter[SpellID]=20572 , https://wago.tools/db2/SpellAuraOptions?build=1.60.1.70634&filter[SpellID]=1260189
- [x] 2026-10-27 — Carry the development notes' post 8 racial additions into the racial text, wiki and DPS report (Blood Fury 2 min, Berserking 15 sec, Touch of the Grave 15 sec ICD; sim side in PR #117) [source: blizzard]
  - `tools/sources/racials_hotfix.json`: add overlays for 20572 (2 min cooldown), 26297 (15 sec) and Touch of the Grave (15 sec), then re-run `racials.py` for `ui/core/forever/racials.json`. `forever-wiki/racials.md` lines 105, 118, 137 and 156 still say 3 min, 25 sec and 10 sec.
  - Re-sim the Orc, Troll and Undead builds of the DPS report once #117 merges. https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-%E2%80%93-updated-october-26/2360696
  - Done 2026-10-27: #117 merged (master 9857d17), overlays and wiki in 2bd9a60, report texts in the next commit, full DPS report re-sim and a second race re-pick (standards 29) deployed.
- [ ] 2026-10-28 — Pull wago.tools DB2 data for build 1.60.1.70658 and process it into the sim (new since 70634; client-identical to it, see orchestration/update_ledger.md) [source: wago build 70658]
  - The ledger's `table_summary.py 70634 70658`: all 39 `client_update` tables byte-identical, no added rows. foreverchanges.pro: 2 of 137 tables differ, both creature models. Bump `foreverClientBuild` and the enchant, itemsparse and racials meta builds when the next source pull runs (this supersedes the 70634 item above).
  - The Oct 28 launch-realm hotfix (Aspect of the Hawk, Trueshot Aura) is not in these tables and already landed by ruling (HUN-HF-01/02). https://wago.tools/db2/SpellEffect?build=1.60.1.70658 , https://wago.tools/db2/SpellCooldowns?build=1.60.1.70658&filter[SpellID]=20572
- [ ] 2026-10-28 — Decide on Kang the Decapitator 2291's Forever proc: Classic's bleed 560 over 30 sec → 500 over 15 sec (revealed October 23 to 28, foreverchanges.pro; client 70658 agrees) [source: foreverchanges]
  - Everything else matches Classic and db.json: 2H axe 136–205 / 3.60, ilvl 49, req 44, BoE, epic. Proc spell Rend 17153: 70658 SpellEffect 694561 is aura 3, 100 every 3000 ms, DurationIndex 8 (15 sec), Physical; Classic 56 every 3 sec for 30 sec.
  - The sim doesn't model Kang's proc at all (no 2291 or 17153 under sim/), Classic or Forever. The item has no ItemEffect row in the client (still hidden in its files), so no proc rate is known; Classic's chance on hit has none either. Adding it means a new `item_effects.go` entry with an assumed PPM. Levelling Ret, Enhancement and Arms weapon (44–51ish).
  - `assets/database/hidden_items.json` still marks it hidden ("simStats": "classic", so the UI's orange mark): re-run `tools/sources/hidden.py` locally after the next cache read. https://foreverchanges.pro/item/2291 , https://wago.tools/db2/SpellEffect?build=1.60.1.70658&filter[SpellID]=17153
- [ ] 2026-10-29 — Pull wago.tools DB2 data for build 1.60.1.70685 and process it into the sim (new since 70658, first seen on wago.tools 2026-10-28 22:57 UTC; live on every region's wowf version server) [source: wago build 70685]
  - The audit loop already diffed it (orchestration/update_ledger.md "Build 1.60.1.70685"): only Adventurer's Tent (1325145, item 238896, new SpellCategory 2607) and a new Camping spell (1328862). No DPS row moved. What's left is the bookkeeping: bump `foreverClientBuild` in `.agents/skills/dps-report/state.json` and the enchant, itemsparse and racials meta builds at the next source pull. This supersedes the 70658 item above.
  - Before ticking it, settle the mage item below: foreverchanges.pro lists mage talent changes under 70685 that wago.tools' 70685 tables don't show. https://wago.tools/db2/TraitDefinition?build=1.60.1.70685 , https://wago.tools/db2/TraitNodeEntry?build=1.60.1.70685 , https://wago.tools/db2/SpellCooldowns?build=1.60.1.70685
- [ ] 2026-10-29 — Confirm foreverchanges.pro's mage talent changes for build 70685 before the sim takes them (foreverchanges.pro/patch-notes "Build 1.60.1.70685", which itself says "None of this is confirmed"; wago.tools' 70685 tables disagree) [source: foreverchanges]
  - What foreverchanges says changed: Improved Fireball → Accelerant (now also Arcane Wrath, 0.1 to 0.5 sec); Elemental Precision → Precision (2/4/6% hit with all spells, was Frost and Fire); Incineration 3 → 5 ranks (6% → 10% crit on Fire Blast, Ice Lance, Arcane Blast and Scorch); Critical Mass 5 → 3 ranks (10% → 6% Fire crit); Ice Lance no cooldown → 8 sec; Arcane Focus (1 to 5% Arcane hit) removed; Frost Blast out of the spellbook; Improved Fire Blast, Arcane Resilience, Missile Barrage and Improved Channeling moved in the tree.
  - What wago.tools' 70685 has (checked 2026-10-29 by the daily routine): TraitDefinition, TraitNode, TraitNodeEntry, TraitEdge, TraitNodeXTraitNodeEntry and TraitDefinitionEffectPoints byte-identical to 70658's; spell 29438 still reads "Improves your chance to hit with Frost and Fire spells", 11222 Arcane Focus still a talent (TraitDefinition 133103/135344); SpellCooldowns differs only on Adventurer's Tent; SpellName has no "Accelerant". So the changes may be server data (a hotfix cache) or another file set. Check https://foreverchanges.pro/talents/mage and wago.tools/hotfixes (product wowf) once it filters by product, and the client cache at launch.
  - If real, the sim moves in `sim/mage/talents.go` (ArcaneFocus 102, Incineration 184, CriticalMass 195, ElementalPrecision 214), `sim/mage/fireball.go:76` and `frostfire_bolt.go:59` (Accelerant on Arcane Wrath too), Ice Lance's cooldown, `ui/core/talents/trees/mage.json` and the mage presets' talent strings (rank counts change). Fire mage loses 4% Fire crit at full Critical Mass and Arcane loses its hit talent, so presets and the DPS report move.
