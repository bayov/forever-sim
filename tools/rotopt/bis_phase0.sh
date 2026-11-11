#!/bin/sh
# Regenerates the level 60 Phase 0 Enhancement BiS list, ui/enhancement_shaman/bis/level60_phase0.bis.json
# (docs/bis_lists.md). Run from the repo root:
#
#   tools/rotopt/bis_phase0.sh
#
# Phase 0 is what Forever has before its raids open on 2026-12-09: dungeons, quests, crafting,
# reputation, world drops and PvP reputation (tools/rotopt/phase0.go has the rules). The fight
# is the DPS Report's default Enhancement build (tools/dps_report/req/shaman__enhancement-level60.json):
# Orc, Engineering and Leatherworking, its talents, rotation, enchants, buffs and consumes, at
# 180 sec. Its gear is the synthetic Phase 1 set, and the search keeps a slot's starting item
# until something beats it, so we start from a real pre-raid set instead: timesaver.gg's level
# 60 Horde Enhancement list (data of 2026-11-06), with the report's enchants on it.
#
# What the search leaves out:
#   - -gear-no-raids: raid and world boss drops, Onyxia's head rewards, legendaries and crafts
#     that need a raid material.
#   - -gear-phase 2: Classic phase 3 and later (Blackwing Lair, Zul'Gurub and Ahn'Qiraj era
#     rewards, Dungeon Set 2). Dire Maul (phase 2) is open in Forever.
#   - -gear-min-ilvl 50: nothing under item level 50 competes at 60, and the random suffix rolls
#     of those greens make the pool three times as big.
#   - Forever items with no source at item level 66 and up, likely Forever's raid loot:
#     Tracker's Signets of Ferocity, Brilliance, Resolve and Vigilance (275972 to 275975),
#     Leafre's Ring of Precise Spell Power (276765) and Band of the Better Half (284716).
#
# On 2026-11-11 the search took 16 min on 16 cores (two rounds, then the ranking).
set -e

sim="$(jj log --no-graph -r @- -T 'commit_id.short()')$(jj log --no-graph -r @ -T 'if(empty, "", "+")')"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

req=tools/dps_report/req/shaman__enhancement-level60.json
jq '.raid.parties[0].players[0].rotation' "$req" >"$tmp/enh.apl.json"
# Head, neck, shoulder, back, chest, wrist, hands, waist, legs, feet, two rings, two trinkets,
# main hand, off hand and the relic, in the order of the equipment list. The report's enchant
# stays on each slot, and the relic is the report's (Totem of the Storm).
jq '.raid.parties[0].players[0].equipment.items as $old
	| ([252605, 22150, 15067, 19121, 16666, 279256, 15063, 18504, 15062, 16670, 281635, 18402, 19120, 249469, 279260, null, $old[16].id]) as $ids
	| .raid.parties[0].players[0].equipment.items = [range(17) as $i
		| if $ids[$i] == null then {} else {id: $ids[$i]} + (if $old[$i].enchant and $i != 15 then {enchant: $old[$i].enchant} else {} end) end]' \
	"$req" >"$tmp/enh.json"

go run --tags=with_db ./tools/rotopt -apl "$tmp/enh.apl.json" -settings "$tmp/enh.json" -durs 180 -iters 2000 \
	-gear-search -gear-no-raids -gear-phase 2 -gear-min-ilvl 50 \
	-gear-exclude 275972,275973,275974,275975,276765,284716 \
	-gear-screen-iters 100 -gear-screen-keep 20 \
	-gear-rank-out ui/enhancement_shaman/bis/level60_phase0.bis.json -gear-rank-name "Phase 0" \
	-gear-rank-spec enh -gear-rank-sim "$sim"
