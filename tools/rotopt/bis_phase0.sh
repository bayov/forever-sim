#!/bin/sh
# Regenerates the level 60 Phase 0 Enhancement BiS list, ui/enhancement_shaman/bis/level60_phase0.bis.json
# (docs/bis_lists.md). Run from the repo root:
#
#   tools/rotopt/bis_phase0.sh
#   TALENTS=053303102-055030110051012151-03 tools/rotopt/bis_phase0.sh
#
# Phase 0 is what Forever has before its raids open on 2026-12-09: dungeons, quests, crafting,
# reputation, world drops and PvP reputation (tools/rotopt/phase0.go has the rules). The fight
# is the DPS Report's default Enhancement build (tools/dps_report/req/shaman__enhancement-level60.json):
# Orc, Engineering and Leatherworking, its rotation, enchants, buffs and consumes, at 180s.
#
# The talents are the ones the user plays (2026-11-14), not the report's: TALENTS sets them, and
# the default is 053303102-054030110051212141-03. Next to the report's build it takes Improved
# Ghost Wolf 2 for one point out of Ancestral Knowledge (4) and one out of Maelstrom Weapon (4).
#
# The report's gear is the synthetic Phase 1 set, and the search keeps a slot's starting item
# until something beats it, so we start from a real pre-raid set instead: timesaver.gg's level
# 60 Horde Enhancement list (data of 2026-11-06), with the report's enchants on it. Two of its
# items aren't out before the raids (Tranquil Wristguards and Forest Defender's Axe), so we
# start with Wildheart Bracers and Treant's Bane there. An excluded starting item can't survive
# anyway, because rotopt takes it off before the search.
#
# What the search leaves out:
#   - -gear-no-raids: raid and world boss drops, Onyxia's head rewards, legendaries and crafts
#     that need a raid material.
#   - -gear-phase 2: Classic phase 3 and later (Blackwing Lair, Zul'Gurub and Ahn'Qiraj era
#     rewards, Dungeon Set 2). Dire Maul (phase 2) is open in Forever.
#   - Classic patch 1.10's Dungeon Set 2 chain, which our item data puts in phase 1 or 2: what
#     the bosses summoned with the Brazier of Beckoning and the Banner of Provocation drop
#     (Sothos and Jarien, Lord Valthalak, Kormok, Isalien, Mor Grayhoof, Theldren), and the
#     chain's quest rewards (Beads of Ogre Might from Falrin's Vendetta). They are "out" in
#     phase0_content.json.
#   - -gear-min-ilvl 50: nothing under item level 50 competes at 60, and the random suffix rolls
#     of those greens make the pool three times as big.
#   - Forever items with no source at item level 66 and up, likely Forever's raid loot:
#     Tracker's Signets of Ferocity, Brilliance, Resolve and Vigilance (275972 to 275975),
#     Leafre's Ring of Precise Spell Power (276765) and Band of the Better Half (284716).
#   - Every item tools/rotopt/phase0_content.json marks "out". That file is the hand check of
#     what the game has open (2026-11-12): "out" means not before the raids, "doubtful" means
#     we don't know of a way yet, and "grind" means open but slow. Items not in it are open now.
#
# What it lets in anyway: the items phase0_content.json marks "include", whatever -gear-phase
# and -gear-no-raids say (-gear-include). Those are the four Darkmoon cards (Heroism, Blue
# Dragon, Maelstrom and Twisting Nether). Our item data puts them in phase 3, but Forever has
# the Darkmoon Faire open, and the cards are the rewards of its decks.
#
# The pick is a two hander, so the main hand list has only two handers and the off hand list
# is empty (-gear-rank-two-hand-only). We don't list a one hand setup next to it.
#
# Then every enchant goes on the set, slot by slot, and the best one stays (bis_enchants.go). The
# items are ranked on the set with those enchants, so the list's DPS has them. The enchants are a
# few DPS apart, so they're ranked at 10000 iterations. tools/rotopt/phase0_enchants.json is the
# hand check of what's open, by spell ID: "out" leaves an enchant out (raid formulas, Alliance
# quests), and its avail and note go on the list's entries. Each slot keeps its 10 best enchants.
#
# Each slot keeps its 25 best items. After the search we copy `avail` and `note` from
# phase0_content.json onto every entry of the list, so a rerun keeps them.
#
# Each round of the search screens a slot's items at 300 iterations and keeps the best 40 for the
# full sims. It used to be 100 and 20, but on 2026-11-14 that kept Devilsaur Gauntlets (the
# starting item) while Bloodmail Gauntlets were 3.9 DPS better at 20000 iterations: a 100
# iteration screen is several DPS off, so the better gloves never got past it.
#
# On 2026-11-13 the whole script took 16 min on 16 cores (two rounds, then the ranking of 25 a slot).
set -e

sim="$(jj log --no-graph -r @- -T 'commit_id.short()')$(jj log --no-graph -r @ -T 'if(empty, "", "+")')"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

talents="${TALENTS:-053303102-054030110051212141-03}"
req=tools/dps_report/req/shaman__enhancement-level60.json
content=tools/rotopt/phase0_content.json
out=ui/enhancement_shaman/bis/level60_phase0.bis.json
exclude="275972,275973,275974,275975,276765,284716,$(jq -r '[to_entries[] | select(.value.avail == "out") | .key] | join(",")' "$content")"
include="$(jq -r '[to_entries[] | select(.value.include) | .key] | join(",")' "$content")"
jq '.raid.parties[0].players[0].rotation' "$req" >"$tmp/enh.apl.json"
# Head, neck, shoulder, back, chest, wrist, hands, waist, legs, feet, two rings, two trinkets,
# main hand, off hand and the relic, in the order of the equipment list. The report's enchant
# stays on each slot, and the relic is the report's (Totem of the Storm).
jq --arg talents "$talents" '.raid.parties[0].players[0].talentsString = $talents
	| .raid.parties[0].players[0].equipment.items as $old
	| ([252605, 22150, 15067, 19121, 16666, 16714, 15063, 18504, 15062, 16670, 281635, 18402, 19120, 249469, 18538, null, $old[16].id]) as $ids
	| .raid.parties[0].players[0].equipment.items = [range(17) as $i
		| if $ids[$i] == null then {} else {id: $ids[$i]} + (if $old[$i].enchant and $i != 15 then {enchant: $old[$i].enchant} else {} end) end]' \
	"$req" >"$tmp/enh.json"

go run --tags=with_db ./tools/rotopt -apl "$tmp/enh.apl.json" -settings "$tmp/enh.json" -durs 180 -iters 2000 \
	-gear-search -gear-no-raids -gear-phase 2 -gear-min-ilvl 50 \
	-gear-exclude "$exclude" -gear-include "$include" \
	-gear-screen-iters 300 -gear-screen-keep 40 \
	-gear-rank-out "$out" -gear-rank-name "Phase 0" -gear-rank-keep 25 -gear-rank-two-hand-only \
	-gear-rank-spec enh -gear-rank-sim "$sim" \
	-gear-rank-enchants tools/rotopt/phase0_enchants.json -gear-rank-enchant-iters 10000

# The content check's status on every entry.
jq --slurpfile content "$content" '.slots |= map_values(map(
	. as $e | ($content[0][$e.id | tostring] // {}) as $c
	| $e + (if $c.avail then {avail: $c.avail} else {} end) + (if $c.note then {note: $c.note} else {} end)))' \
	"$out" >"$tmp/list.json"
mv "$tmp/list.json" "$out"
