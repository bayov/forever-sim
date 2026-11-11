package main

import "github.com/wowsims/classic/sim/core/proto"

// What -gear-no-raids leaves out, for a search over what Forever has before its raids open.
//
// Forever opened on 2026-11-04 with no raid at all. Its first three raids (Barrow Deeps, Hyjal
// Summit and Onyxia's Lair) open on 2026-12-09, and Molten Core has no date yet. Until then a
// player gears from dungeons, quests, crafting, reputation, world drops and PvP. We call that
// Phase 0.
//
// Forever's own raid loot has no source in our data yet, so these rules can't catch it. A
// Forever item with no source and an item level above what dungeons drop (66 and up) is
// probably raid loot, and the search that wants Phase 0 should leave it out by ID.

// raidZones are the Classic raids: Molten Core, Onyxia's Lair, Blackwing Lair, Zul'Gurub,
// Ruins of Ahn'Qiraj, Ahn'Qiraj and Naxxramas.
var raidZones = []int32{2717, 2159, 2677, 1977, 3429, 3428, 3456}

// worldBosses are the outdoor raid bosses: Azuregos, Lord Kazzak and the four Emerald dragons.
var worldBosses = []int32{6109, 12397, 14887, 14888, 14889, 14890}

// raidQuestRewards are the rewards for turning in Onyxia's head: Onyxia Tooth Pendant,
// Onyxia Blood Talisman and Dragonslayer's Signet. The Blackwing Lair, Zul'Gurub and
// Ahn'Qiraj quest rewards are Classic phase 3 and later, so -gear-phase 2 already leaves
// them out.
var raidQuestRewards = []int32{18404, 18406, 18403}

// raidMaterialCrafts are the crafted items (item level 50 and up) whose recipe needs a raid
// material: Molten Core's Fiery Core, Lava Core, Core Leather and Sulfuron Ingot, Zul'Gurub's
// Bloodvine, Souldarite and primal leathers, Ahn'Qiraj's obsidian shards, Naxxramas's Frozen
// Rune and the emerald dragons' Dreamscale. The Scourge Invasion's Dark Rune is in too, since
// Forever has no invasion. Forever moved some of these recipes (Plans: Nightfall now comes
// from the Thorium Brotherhood at Exalted), but the recipe still takes 8 Lava Core and 5
// Fiery Core.
//
// The list comes from the crafted sources in db.json and wago.tools' SpellReagents table for
// client 1.60.1.70888, read on 2026-11-11.
var raidMaterialCrafts = []int32{
	16979, 16980, 16982, 16983, 16984, 16988, 16989, 17013, 17014, 17015,
	17016, 17193, 18263, 18282, 19148, 19149, 19156, 19157, 19162, 19163,
	19164, 19165, 19166, 19167, 19168, 19169, 19170, 19682, 19683, 19684,
	19685, 19686, 19687, 19688, 19689, 19690, 19691, 19692, 19693, 19694,
	19695, 19998, 19999, 20039, 20380, 20537, 20538, 20539, 20549, 20550,
	20551, 22191, 22194, 22195, 22196, 22197, 22198, 22384, 22652, 22654,
	22655, 22658, 22660, 22661, 22662, 22663, 22664, 22665, 22666, 22669,
	22670, 22671, 22756, 22759, 22762, 22763,
}

// raidOnly says whether an item needs a raid to get: a legendary, a reward for a raid boss's
// head, a craft with a raid material, or an item whose every source is a raid or world boss
// drop. An item a raid boss drops but that can also be crafted, bought or looted in a dungeon
// stays.
func raidOnly(item *proto.UIItem) bool {
	if item.Quality >= proto.ItemQuality_ItemQualityLegendary {
		return true
	}
	if contains(raidQuestRewards, item.Id) || contains(raidMaterialCrafts, item.Id) {
		return true
	}
	if len(item.Sources) == 0 {
		return false
	}
	for _, src := range item.Sources {
		drop := src.GetDrop()
		if drop == nil || !contains(raidZones, drop.ZoneId) && !contains(worldBosses, drop.NpcId) {
			return false
		}
	}
	return true
}
