#!/bin/sh
# Regenerates the level 60 Phase 0 Enhancement BiS list, ui/enhancement_shaman/bis/level60_phase0.bis.json
# (docs/bis_lists.md). Run from the repo root:
#
#   tools/rotopt/bis_phase0.sh
#   TALENTS=053303102-055030110051012151-03 tools/rotopt/bis_phase0.sh
#   RANK_ONLY=1 tools/rotopt/bis_phase0.sh
#
# RANK_ONLY=1 skips the gear search and ranks the picks of the list we
# have (FROM, the list itself by default) again, enchants included. It is for a rerun after a sim
# or item data change that shouldn't move the picks much, since the ranking still follows an item
# that beats its pick (docs/bis_lists.md). OUT writes the list somewhere else, and any arguments go
# on to rotopt after ours, so they win: a quick check can run at fewer iterations with
#
#   RANK_ONLY=1 OUT=/tmp/check.bis.json tools/rotopt/bis_phase0.sh -iters 300 -gear-rank-confirm-iters 2000 -gear-rank-enchant-iters 1000
#
# SLOTS (with RANK_ONLY) ranks only those lists, with their enchants, and copies every other list
# from FROM as it is (-gear-rank-slots). The lists are head, neck, shoulder, back, chest, wrist,
# hands, waist, legs, feet, finger, trinket, mainhand, offhand and ranged:
#
#   RANK_ONLY=1 SLOTS=trinket tools/rotopt/bis_phase0.sh
#
# EFFECTS_ONLY=1 skips the search and the ranking and only writes the effects file (EFFECTS_OUT,
# ui/enhancement_shaman/bis/level60_phase0.effects.json by default) for the picks of FROM: what
# each item's effects add on the set, for BayovCore's stat score (docs/bis_lists.md, Effect
# values). A full run and a RANK_ONLY run write it too, after the list, and a SLOTS run only when
# EFFECTS_OUT is set. EFFECTS_OUT=none skips it.
#
#   EFFECTS_ONLY=1 tools/rotopt/bis_phase0.sh
#
# Phase 0 is what Forever has before its raids open on 2026-12-09: dungeons, quests, crafting,
# reputation, world drops and PvP reputation (tools/rotopt/phase0.go has the rules). The fight
# is the DPS Report's default Enhancement build (tools/dps_report/req/shaman__enhancement-level60.json):
# Orc, Engineering and Leatherworking, its rotation, enchants, buffs and consumes.
#
# A set is scored in two parts (user, 2026-11-20, -use-dur in use_values.go). Everything but
# the uses counts on a DUR second fight (180 by default) with every use off, because between
# fights there's little or no time to regen mana. Each use counts for what it adds on a USE_DUR
# second fight (90 by default), pressed at the pull. The user's level 60 dungeon fights are short
# (the logs of 2026-11-20: trash fights over 20s had a median of 32 to 38s, bosses 42s), an addon
# swaps trinkets between fights, and swapping a usable item in puts it on a 30s equip cooldown,
# so sometimes its use never gets pressed. 90s takes that into account roughly. When the raids
# open (2026-12-09) their bosses and trash may need both lengths measured again from raid logs.
#
# Each item with a use also gets an "on cooldown" rank, without its use value
# (-gear-rank-cooldowns): the swap addon takes it off while its use is on cooldown.
#
# And a "vs 4 mobs" conditional rank (user, 2026-11-20, -use-aoe-targets in bis_aoe.go), because
# a dungeon pull is often several mobs and some uses hit them all, like Smokey's Lighter's fire
# cone. Only the use counts against the mobs: its value on USE_DUR is measured against AOE_TARGETS
# copies of the boss (4 by default), while the DUR part stays on the one boss. An item gets the
# line, and its effect value with it, only when its use is worth more than 1 DPS more against the
# mobs. AOE_TARGETS=0 leaves it out.
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
#     (Sothos and Jarien, Lord Valthalak, Kormok, Isalien, Mor Grayhoof, Theldren). They are "out"
#     in phase0_content.json. Forever turned the chain into Mokvar's Five Thunders upgrades
#     (2026-11-17), so its steps without a summon are open, like Beads of Ogre Might from
#     Falrin's Vendetta. The head and chest upgrades come at the end, after the Valthalak
#     summons, so they are doubtful until we know the summons work.
#   - -gear-min-ilvl 50: nothing under item level 50 competes at 60, and the random suffix rolls
#     of those greens make the pool three times as big. An "include" item joins anyway, like
#     Manual Crowd Pummeler (item level 34, its on-use haste is what counts).
#   - Forever items with no source at item level 66 and up, likely Forever's raid loot:
#     Leafre's Ring of Precise Spell Power (276765) and Band of the Better Half (284716).
#     The Tracker's Signets (275972 to 275975) used to be on this list too, but their tooltips
#     say "Requires The Watchers - Honored", the same reputation as the Warden's Signets
#     (Exalted). So they're candidates now, doubtful because we know of no reputation source
#     yet. The Warden's Signets need Exalted, so they're infeasible (below).
#   - Every item tools/rotopt/phase0_content.json marks "out". That file is the hand check of
#     what the game has open (2026-11-12): "out" means not before the raids, "doubtful" means
#     we don't know of a way yet, and "infeasible" is below. Everything else is open, and its
#     note says what it takes.
#
# What it ranks but never picks: the items phase0_content.json marks "infeasible" (user,
# 2026-11-20, -gear-no-pick). Those need Exalted reputation, which is a very long grind: the
# Frostwolf rewards (The Unstoppable Force, Don Julio's Band), Warsong Gulch, the Defilers, the
# Darkspear Raiders, the Watchers (the Warden's Signets) and Timbermaw (the Stormcarver plans).
# The search never puts one in the set, so each slot's pick is the best item we can get. But
# each slot's list still ranks them on that set, so a player who has the reputation sees where
# the item goes. Each list is in order of loss, so one that beats the pick sits above it, with a
# negative loss. Only the "bis" flag marks the picks (user, 2026-11-20).
# An item that needs Exalted and also isn't known to be open yet (the Defilers' rewards) is
# "infeasible", not "doubtful".
#
# What it lets in anyway: the items phase0_content.json marks "include", whatever -gear-phase
# and -gear-no-raids say (-gear-include). Our item data takes its phases from Classic, where
# these came out later, but Forever has them from the start:
#   - The four Darkmoon cards (Heroism, Blue Dragon, Maelstrom and Twisting Nether) and the
#     Faire's ticket rewards (Amulet and Orb of the Darkmoon). Forever has the Darkmoon Faire
#     open, and the cards are the rewards of its decks.
#   - Totem of Rage (Magmus in Blackrock Depths) and Totem of Sustaining (Scholomance). Classic
#     added them in patch 1.10, but Wowhead's Forever data has them on those drops, and the game
#     knows both items (forever_revealed_in_game.json).
#   - The Arathi Basin rewards of the Defilers (Rutherford Twing in Hammerfall). Wowhead's
#     Forever data sells them there, but we don't know yet that Arathi Basin is open, so they're
#     doubtful.
#   - Every other item the content check keeps (not "out") that our data puts after phase 2,
#     such as the Silithus Cenarion Circle rewards. Their status shows on the list, so a
#     doubtful one is marked rather than left out without a word.
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
# Items whose effect grows under a condition (Blackhand's Breadth vs Orcs and Ogres, Seal of the Dawn
# vs Undead, Golden Banana in forest or grassland) are ranked again with it met, and an item that
# makes the list only that way joins it (-gear-rank-conditions, docs/bis_lists.md).
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
# The use values at 90s (2026-11-20) add their runs to it, so expect about 20 to 25 min. A
# RANK_ONLY SLOTS=trinket run took 2 min, and EFFECTS_ONLY 1.5 min. The uses against 4 mobs
# (2026-11-20) made a RANK_ONLY SLOTS=trinket,mainhand run 3 min, and EFFECTS_ONLY 2.5 min.
set -e

sim="$(jj log --no-graph -r @- -T 'commit_id.short()')$(jj log --no-graph -r @ -T 'if(empty, "", "+")')"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

talents="${TALENTS:-053303102-054030110051212141-03}"
# The two fight lengths (user, 2026-11-20, see above). Everything but the uses counts on 180s,
# because there's little or no time to regen mana between fights. The uses count on 90s, for the
# short fights and the 30s equip cooldown a swapped in usable item gets. Revisit both for raids.
dur="${DUR:-180}"
use_dur="${USE_DUR:-90}"
# How many mobs the uses are also valued against, for the "vs 4 mobs" lines (see above).
aoe_targets="${AOE_TARGETS:-4}"
effects_out="${EFFECTS_OUT:-ui/enhancement_shaman/bis/level60_phase0.effects.json}"
# A SLOTS run is a quick rerun of a few lists, so it writes the effects file only when asked.
if [ -n "$SLOTS" ] && [ -z "$EFFECTS_OUT" ]; then
	effects_out=none
fi
req=tools/dps_report/req/shaman__enhancement-level60.json
content=tools/rotopt/phase0_content.json
out="${OUT:-ui/enhancement_shaman/bis/level60_phase0.bis.json}"
from="${FROM:-ui/enhancement_shaman/bis/level60_phase0.bis.json}"
exclude="276765,284716,$(jq -r '[to_entries[] | select(.value.avail == "out") | .key] | join(",")' "$content")"
include="$(jq -r '[to_entries[] | select(.value.include) | .key] | join(",")' "$content")"
nopick="$(jq -r '[to_entries[] | select(.value.avail == "infeasible") | .key] | join(",")' "$content")"
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
rank_only=""
if [ -n "$EFFECTS_ONLY" ]; then
	RANK_ONLY=1
	if [ -n "$SLOTS" ]; then
		echo "SLOTS doesn't go with EFFECTS_ONLY" >&2
		exit 1
	fi
fi
if [ -n "$RANK_ONLY" ]; then
	# A rerun copies lists from FROM or ranks against its numbers, so FROM must have been ranked
	# on the same fight lengths. A 45s list with 180s ranks in a few slots reads as one list but
	# isn't. MIX_DURS=1 goes on anyway.
	from_durs="$(jq -r '"\(.durations[0] // "none")/\(.use_dur // "none")"' "$from")"
	if [ "$from_durs" != "$dur/$use_dur" ] && [ -z "$MIX_DURS" ]; then
		echo "$from was ranked at durations[0]/use_dur $from_durs, not DUR/USE_DUR $dur/$use_dur." >&2
		echo "Rebuild it in full, or set MIX_DURS=1 to rerun on it anyway." >&2
		exit 1
	fi
	# The list's picks, with their suffix and enchant, in the order of the equipment list. An
	# empty slot (the off hand next to a two hander) has no pick.
	jq --slurpfile list "$from" '$list[0].slots as $s
		| def pick($k; $i): ([$s[$k][]? | select(.bis)][$i]) as $p
			| if $p == null then {} else {id: $p.id} + (if $p.suffix then {randomSuffix: $p.suffix} else {} end)
				+ (if $p.enchant then {enchant: $p.enchant} else {} end) end;
		.raid.parties[0].players[0].equipment.items = [pick("head"; 0), pick("neck"; 0), pick("shoulder"; 0), pick("back"; 0),
			pick("chest"; 0), pick("wrist"; 0), pick("hands"; 0), pick("waist"; 0), pick("legs"; 0), pick("feet"; 0),
			pick("finger"; 0), pick("finger"; 1), pick("trinket"; 0), pick("trinket"; 1), pick("mainhand"; 0), pick("offhand"; 0), pick("ranged"; 0)]' \
		"$tmp/enh.json" >"$tmp/picks.json"
	mv "$tmp/picks.json" "$tmp/enh.json"
	rank_only=-gear-rank-only
	if [ -n "$SLOTS" ]; then
		rank_only="$rank_only -gear-rank-slots $SLOTS -gear-rank-from $from"
	fi
elif [ -n "$SLOTS" ]; then
	echo "SLOTS needs RANK_ONLY=1" >&2
	exit 1
fi

# What we write: the list (not with EFFECTS_ONLY) and the effects file (not with EFFECTS_OUT=none).
outputs="-gear-rank-out $out"
if [ -n "$EFFECTS_ONLY" ]; then
	outputs=""
fi
if [ "$effects_out" != none ]; then
	outputs="$outputs -gear-effects-out $effects_out"
fi

go run --tags=with_db ./tools/rotopt -apl "$tmp/enh.apl.json" -settings "$tmp/enh.json" -durs "$dur" -iters 2000 \
	-gear-search -gear-no-raids -gear-phase 2 -gear-min-ilvl 50 \
	-gear-exclude "$exclude" -gear-include "$include" -gear-no-pick "$nopick" \
	-gear-screen-iters 300 -gear-screen-keep 40 \
	-gear-rank-name "Phase 0" -gear-rank-keep 25 -gear-rank-two-hand-only \
	-gear-rank-spec enh -gear-rank-sim "$sim" \
	-gear-rank-enchants tools/rotopt/phase0_enchants.json -gear-rank-enchant-iters 10000 \
	-gear-rank-confirm-iters 20000 -gear-rank-conditions -gear-rank-cooldowns -gear-effects-iters 10000 \
	-use-dur "$use_dur" -use-iters 20000 -use-aoe-targets "$aoe_targets" $outputs $rank_only "$@"

if [ -n "$EFFECTS_ONLY" ]; then
	exit 0
fi

# The content check's status on every entry.
jq --slurpfile content "$content" '.slots |= map_values(map(
	. as $e | ($content[0][$e.id | tostring] // {}) as $c
	| $e + (if $c.avail then {avail: $c.avail} else {} end) + (if $c.note then {note: $c.note} else {} end)))' \
	"$out" >"$tmp/list.json"
mv "$tmp/list.json" "$out"
