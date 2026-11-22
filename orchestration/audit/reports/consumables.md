# Consumables pass 98 (master 7a6d1beb9d, client 1.60.1.71037)

Result: quiet, no new findings. Wago has no build newer than 71037.

Checked (diff 8b5137c6d..7a6d1beb9d, the 7 watched files):
- ui/enhancement_shaman/gear_sets/level60_phase0.gear.json: reranked Phase 0 set (Five Thunders pieces, Skyfury Gauntlets, Spiritcaller Leggings, Beads of Ogre Might, Brain Hacker). Enchants 2543 and 1249019 as before; 1254965 (Forceful Rugged Armor Kit, boots) exists as spell 1254965 in the 71037 tables and is in forever_enchants.go (level 50 kit). Trinkets 13965 (Blackhand's Breadth) and 10455 (Chained Essence of Eranikus) are real, so scan.sh NONSYN lists both; accepted as in pass 97 (Phase 0 is real pre-raid gear, not the Phase 1 synthetic set). Both are in Item.csv of 71037 but not ItemSparse; the Phase 0 content file cites the Wowhead Forever source. wowhead gave no source page from this host (srcid.py: NO SOURCE LISTED), so not independently checked here.
- sim/core/buffs.go: Windfury Totem now rolls on the Execute cast as well (warrior only, assumption warr_exec_wf2, cites wowsims/forever EV-0155); no consumable value changed.
- sim/core/debuffs.go: Judgement of Wisdom needs a landed melee auto under Forever (cites wowsims/forever EV-0018); not a number from the consumable tables.
- tools/dps_report/assumptions_def.py: warr_exec_wf2 added, Ice Lance coefficient set to 11% (user ruling 2026-11-22); neither is a buff value this pass owns.
- sim/common/item_effects.go (Linken's Boomerang, Cloak of Fire, armor use notes) and item_sets/forever_dungeon_*: item and set effects, not consumables or enchants; not in any preset.
- scan.sh: banned, Revelation outside shamans, crit-elixir checks empty. Minor-armor exclusive effect count reads 3 (as in pass 97, debuffs.go unchanged there). Air/weapon totem counts as the standing rules.

Open items: the two low findings in findings/consumables.md.
