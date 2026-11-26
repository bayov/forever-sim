#!/bin/sh
# Regenerates a level 60 Phase 0 BiS list (docs/bis_lists.md), Enhancement's
# (ui/enhancement_shaman/bis/level60_phase0.bis.json) by default. Run from the repo root:
#
#   tools/rotopt/bis_phase0.sh
#   SPEC=rogue_combat_swords tools/rotopt/bis_phase0.sh
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
# from FROM (-gear-rank-slots). The list's dps is the final set's, measured in the run, and each
# copied entry keeps its loss but gets its dps from that measure. The lists are head, neck,
# shoulder, back, chest, wrist, hands, waist, legs, feet, finger, trinket, mainhand, offhand and
# ranged:
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
# SPEC picks the spec from tools/rotopt/phase0_specs.json, one entry per DPS spec of the DPS
# Report: the report build the fight comes from, the talents, the starting gear, the weapon
# setup and where the lists go. Most specs start with no gear, and the search fills each slot.
# A hunter can't: with no ranged weapon its DPS is NaN, nothing beats NaN, so the search swapped
# nothing (2026-11-23). The hunters start from the set Beast Mastery's first search found from
# timesaver.gg's Forever list (2026-11-20, open items, Eaglehorn Long Bow for its Core Marksman
# Rifle), with the report's enchants. Its shoulders (Theramore Pauldrons, Alliance only) and second
# ring (a random suffix) are now the ones Beast Mastery's search settled on (2026-11-24), so a
# container restart costs a search one round less. What follows on talents, starting gear and
# weapons is Enhancement's entry.
#
# GEAR_ONLY=1 sims nothing either. It writes each spec's gear set (the table's gearOut, GEAR_OUT
# to write it elsewhere) from the picks of its list (FROM, the list itself by default), with
# their enchants, the way a full or RANK_ONLY run writes it after the list. A pick that is a
# no-pick item now gives its place to the best entry that isn't (see picks below). SPEC is one
# spec or a comma separated list, every spec in the table by default:
#
#   GEAR_ONLY=1 SPEC=affliction tools/rotopt/bis_phase0.sh
#
# RACE (a proto race, RaceHuman) runs the spec at another race than its report build's, for a row
# of the other faction: its pool drops the build's faction rewards and takes its own (faction.go).
# With RANK_ONLY, SLOTS, FROM and OUT it ranks the build's faction slots for that faction, and
# GEAR_ONLY with that list and GEAR_OUT writes the set's faction copy, <gearOut minus .gear.json>_<horde|alliance>.gear.json,
# which tools/dps_report/race_gear.py gives the other faction's races (2026-11-24).
#
# MH_IMBUE and OH_IMBUE (a proto weapon imbue, DenseWeightstone) put another imbue on a hand. The
# whole pool is ranked with the request's imbue, and a stone only fits its weapon type (a
# Sharpening Stone edged, a Weightstone blunt, consumes.go ImbueFitsWeapon), so a dual wield list
# that mixes edged and blunt weapons ranks the other type again with its own stone.
#
# DRY_RUN=1 sims nothing. For each spec of SPEC (every spec in the table by default, or a comma
# separated list) it writes what the search's pool would hold (-gear-pool-report) to
# DRY_RUN_DIR (orchestration/round_notes_2026-11-11/phase0_pool by default): the pool's size,
# the items with use or proc effects the sim doesn't model, the items whose stats Forever still
# hides, and the items the content check doesn't list. That is the hand check's input before a
# spec's first search.
#
#   DRY_RUN=1 tools/rotopt/bis_phase0.sh
#
# It also lists what the faction check leaves out of each pool (below).
#
# Phase 0 is what Forever has before its raids open on 2026-12-09: dungeons, quests, crafting,
# reputation, world drops and PvP reputation (tools/rotopt/phase0.go has the rules). The fight
# is the spec's DPS Report build, for Enhancement the default one (tools/dps_report/req/shaman__enhancement-level60.json):
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
#     what the game has open (2026-11-12): "out" means we know it isn't open before the raids
#     (orchestration/standards.md item 42, an unsure item is "include" with a note), "doubtful"
#     means we don't know of a way yet, and "infeasible" is below. Every other item it lists is open,
#     and its note says what it takes. An item it doesn't list hasn't been checked: when a pick
#     or one of a list's top 5 is one, the run prints them all and fails (below).
#   - What only the other faction can get, by the build's race (2026-11-23, faction.go): an
#     item our data gives a faction, one whose every source is one faction's reputation,
#     quest or vendor (Theramore and Darkspear in the Barrens), and one the content check
#     tags with a "faction" (Courier's Clasp, a Horde quest reward our data has no quest
#     side for). It comes off the starting set too. phase0_enchants.json tags enchants the same
#     way. The dry run lists what it leaves out (factionDropped).
#
# What it ranks but never picks: the items phase0_content.json marks "infeasible" (user,
# 2026-11-20, -gear-no-pick). Those need Exalted reputation, which is a very long grind: the
# Frostwolf rewards (The Unstoppable Force, Don Julio's Band), Warsong Gulch, the Defilers, the
# Watchers (the Warden's Signets) and Timbermaw (the Stormcarver plans). The Darkspear Raiders'
# gloves and shoulders need only Revered (the item data, 2026-11-24), so they're open.
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
# is empty (-gear-rank-two-hand-only). We don't list a one hand setup next to it. That is the
# table's "two-hand" weapons. It leaves the search's pool as it was (one handers and off hands
# in), so Enhancement's search is the same as before the table. The other setups hold the pool
# to them (-gear-weapons, applyWeaponRules in gear.go): "dual-wield" for the rogues and Fury
# dual wield, with daggers only where the rotation needs them (mhTypes, ohTypes), and "any" for
# the casters and hunters, which weigh a staff or a two hander against one hand and off hand.
#
# Then every enchant goes on the set, slot by slot, and the best one stays (bis_enchants.go). The
# items are ranked on the set with those enchants, so the list's DPS has them. The enchants are a
# few DPS apart, so they're ranked at 10000 iterations. tools/rotopt/phase0_enchants.json is the
# hand check of what's open, by spell ID: "out" leaves an enchant out (raid formulas), "faction"
# leaves it out of the other faction's builds (an Alliance quest's), and its avail and note go on
# the list's entries. Each slot keeps its 10 best enchants.
#
# Items whose effect grows under a condition (Blackhand's Breadth vs Orcs and Ogres, Seal of the Dawn
# vs Undead, Golden Banana in forest or grassland) are ranked again with it met, and an item that
# makes the list only that way joins it (-gear-rank-conditions, docs/bis_lists.md).
#
# Each slot keeps its 25 best items. After the search we copy `avail` and `note` from
# phase0_content.json onto every entry of the list, so a rerun keeps them.
#
# Then the unchecked items: every pick and every item in the top 5 of a slot's list that
# phase0_content.json doesn't list, and every pick and top 5 enchant phase0_enchants.json
# doesn't list. The run prints them and fails, after writing the list, so the hand check knows
# what to look at. ALLOW_UNLISTED=1 prints them and goes on. The Enhancement list's items and
# enchants that the 2026-11-12 check left open by not listing them are listed now, with only
# their name (2026-11-22).
#
# Tanks and healers (2026-11-25): a spec with bench in the table is one of the Tanking and
# Healing report's presets (its reqFile in tools/tankbench/req or tools/healbench/req), and its
# sets are scored by the score that report publishes, not DPS:
#   - tank: the preset's tank score (tools/tankscore, rotopt -tank with its tankPreset weight),
#     the mean over the six scored bosses (magmadar to chromaggus in tools/encounters), each at
#     its own length and healing. The synthetic tank sets differ by preset, so each preset has
#     its own list (Balanced, Max threat, Max survival).
#   - heal: the damage carried in percent (tools/healscore, rotopt -heal-roles), the mean over
#     the seven bosses in each of its roles (healRoles), the roles weighed alike. The raid role
#     sims the raid preset (raidReqFile) under the set. The synthetic healer sets are one a spec,
#     so a healer has one list. docs/healing_sim.md has the search's quicker carry estimate.
# The bosses' index takes the fight length's place (DUR 0), and the uses count inside the boss
# fights (USE_DUR 0, no "vs 4 mobs" lines). There's no effects file, since BayovCore's stat
# score is DPS. A score is six or fourteen boss runs, so the search runs 1000 iterations (the
# report's), screens at 300 and keeps 15 for the full sims, and confirms at 5000. Before that the
# pool keeps each slot's best 25 by the score's own stat weights, measured on the preset's
# synthetic set (-gear-ep-keep, docs/bis_lists.md), with every item with an effect or of a set.
# The tanks start from real sets of their armor: Valor for the warrior, and the Phase 0 sets of
# Retribution, Feral and Enhancement for the paladin, bear and Warden, with a one hander and
# Draconian Deflector for the shield tanks (Vilerend Slicer, Annihilator for the Warden). With no
# weapon a tank dies past what healing saves on some bosses, and the score is the reject value. The healers start from the Phase 0 sets of Smite
# (priests), Elemental (paladin and shaman) and Balance (druid).
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

# The sim the list comes from: the last commit, with a + when the files differ from it. jj where
# it is, git where it isn't (the cloud containers have no jj).
if command -v jj >/dev/null 2>&1 && jj root >/dev/null 2>&1; then
	sim="$(jj log --no-graph -r @- -T 'commit_id.short()')$(jj log --no-graph -r @ -T 'if(empty, "", "+")')"
else
	sim="$(git rev-parse --short=12 HEAD)$(if [ -n "$(git status --porcelain)" ]; then echo +; fi)"
fi
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

specs=tools/rotopt/phase0_specs.json
spec="${SPEC:-enh}"
if [ -n "$DRY_RUN" ] || [ -n "$GEAR_ONLY" ]; then
	spec="${SPEC:-$(jq -r '.specs | keys_unsorted | join(",")' "$specs")}"
fi
# One field of a spec's entry, empty when it has none.
field() {
	jq -r --arg spec "$1" --arg f "$2" '.specs[$spec][$f] // empty' "$specs"
}
for s in $(echo "$spec" | tr , ' '); do
	if [ "$(jq --arg spec "$s" '.specs | has($spec)' "$specs")" != true ]; then
		echo "SPEC $s isn't in $specs: $(jq -r '.specs | keys_unsorted | join(", ")' "$specs")" >&2
		exit 1
	fi
done
# The two fight lengths (user, 2026-11-20, see above). Everything but the uses counts on 180s,
# because there's little or no time to regen mana between fights. The uses count on 90s, for the
# short fights and the 30s equip cooldown a swapped in usable item gets. Revisit both for raids.
dur="${DUR:-180}"
use_dur="${USE_DUR:-90}"
# A tank or healer spec (bench in the table, 2026-11-25, see Tanks and healers below) scores its
# sets on the report's bosses, each at its own length, so the bosses' index takes the fight
# length's place (0), and the uses count inside those fights (USE_DUR 0, no mobs).
bench="$(field "$(echo "$spec" | cut -d, -f1)" bench)"
if [ -n "$bench" ]; then
	dur="${DUR:-0}"
	use_dur="${USE_DUR:-0}"
	AOE_TARGETS="${AOE_TARGETS:-0}"
fi
# How many mobs the uses are also valued against, for the "vs 4 mobs" lines (see above).
aoe_targets="${AOE_TARGETS:-4}"
# The weapon flags of a spec (see above).
weapon_flags() {
	weapons="$(field "$1" weapons)"
	if [ "$weapons" = two-hand ]; then
		echo -gear-rank-two-hand-only
	else
		echo -gear-weapons "$weapons"
	fi
	mh="$(field "$1" mhTypes)"
	oh="$(field "$1" ohTypes)"
	if [ -n "$mh" ]; then
		echo -gear-mh-types "$mh"
	fi
	if [ -n "$oh" ]; then
		echo -gear-oh-types "$oh"
	fi
}
# The list's spec field: the table's rankSpec, or the spec's key.
rank_spec() {
	r="$(field "$1" rankSpec)"
	echo "${r:-$1}"
}
# The request a spec's fight comes from: the table's reqFile (a tank or healer spec's preset,
# tools/tankbench/req or tools/healbench/req), or its report build tools/dps_report/req/<req>.json.
req_file() {
	r="$(field "$1" reqFile)"
	echo "${r:-tools/dps_report/req/$(field "$1" req).json}"
}
# The scoring flags of a tank or healer spec (see Tanks and healers above), none for a DPS spec.
# A tank's score is the tank report's: its preset (tankPreset) on the six scored bosses, the mean.
# A healer's is the healer report's damage carried, in its roles (healRoles), on the seven bosses,
# the raid role simmed with the raid preset's request (raidReqFile) under the search's gear.
tank_bosses="$(for b in magmadar golemagg onyxia broodlord vaelastrasz chromaggus; do printf 'tools/encounters/%s.json,' "$b"; done | sed 's/,$//')"
heal_bosses="$tank_bosses,tools/encounters/ragnaros.json"
bench_flags() {
	case "$(field "$1" bench)" in
	tank) echo -tank "$(field "$1" tankPreset)" -tank-encounter "$tank_bosses" ;;
	heal)
		echo -heal-roles "$(field "$1" healRoles)" -heal-encounter "$heal_bosses"
		r="$(field "$1" raidReqFile)"
		if [ -n "$r" ]; then
			settings "$1" "$r" >"$tmp/$1.raid.json"
			echo -heal-raid-settings "$tmp/$1.raid.json"
		fi
		;;
	esac
}
# The settings a spec starts from: its report build with the table's talents and starting gear.
# A "req" slot keeps the report's item, and startEnchants keeps the report's enchant on every
# filled slot but the off hand. TALENTS wins over the table.
settings() {
	t="${TALENTS:-$(field "$1" talents)}"
	jq --arg talents "$t" --arg race "${RACE:-}" --arg mhi "${MH_IMBUE:-}" --arg ohi "${OH_IMBUE:-}" --argjson start "$(jq -c --arg spec "$1" '.specs[$spec].start // [range(17) | null]' "$specs")" \
		--argjson ench "$(jq --arg spec "$1" '.specs[$spec].startEnchants // false' "$specs")" \
		'(if $talents != "" then .raid.parties[0].players[0].talentsString = $talents else . end)
	| (if $race != "" then .raid.parties[0].players[0].race = $race else . end)
	| (if $mhi != "" then .raid.parties[0].players[0].consumes.mainHandImbue = $mhi else . end)
	| (if $ohi != "" then .raid.parties[0].players[0].consumes.offHandImbue = $ohi else . end)
	| .raid.parties[0].players[0].equipment.items as $old
	| [range(17) as $i | if $start[$i] == "req" then $old[$i].id else $start[$i] end] as $ids
	| .raid.parties[0].players[0].equipment.items = [range(17) as $i
		| if $ids[$i] == null then {} else {id: $ids[$i]} + (if $ench and $old[$i].enchant and $i != 15 then {enchant: $old[$i].enchant} else {} end) end]' \
		"${2:-$(req_file "$1")}"
}

req="$(req_file "$spec")"
dir="$(field "$spec" dir)"
# Where a spec's list goes: the table's out, or ui/<dir>/bis/level60_phase0_<spec>.bis.json.
list_path() {
	o="$(field "$1" out)"
	echo "${o:-ui/$(field "$1" dir)/bis/level60_phase0_$1.bis.json}"
}
def_out="$(list_path "$spec")"
def_effects="$(field "$spec" effectsOut)"
def_effects="${def_effects:-ui/$dir/bis/level60_phase0_$spec.effects.json}"
effects_out="${EFFECTS_OUT:-$def_effects}"
# The effects file values items in DPS for BayovCore's stat score, which has no tank or healer side.
if [ -n "$bench" ] && [ -z "$EFFECTS_OUT" ]; then
	effects_out=none
fi
# A SLOTS run is a quick rerun of a few lists, so it writes the effects file only when asked.
if [ -n "$SLOTS" ] && [ -z "$EFFECTS_OUT" ]; then
	effects_out=none
fi
content=tools/rotopt/phase0_content.json
enchants=tools/rotopt/phase0_enchants.json
out="${OUT:-$def_out}"
from="${FROM:-$def_out}"
exclude="276765,284716,$(jq -r '[to_entries[] | select(.value.avail == "out") | .key] | join(",")' "$content")"
include="$(jq -r '[to_entries[] | select(.value.include) | .key] | join(",")' "$content")"
# The most gold an item may cost for each DPS it adds (user, 2026-11-22). An item over it is
# infeasible like an Exalted reputation item: ranked, but never a pick. tools/rotopt/phase0_prices.py
# writes each item's goldPerDps into the content check from our auction house scans, one number
# per spec's list ({"mage_fire": 160}, since an item adds different DPS on each), and should run
# again after a rerun moves the picks. nopick_of gives a spec's no-pick items.
max_gpd="${MAX_GOLD_PER_DPS:-30}"
# A tank or healer list's line is in its score (2026-11-25): 30g per DPS on a build near 1000 DPS
# is about 3000g per 1% of DPS, so an item over MAX_GOLD_PER_PCT (3000) for each 1% of the spec's
# score it adds is infeasible. 1% is of the table's scoreRef: a tank class's Balanced score on the
# synthetic set for all three of its presets (Max survival scores can be near 0 or below), a
# healer's damage carried on it. goldPerDps by spec then reads as gold per score unit.
max_gold_per_pct="${MAX_GOLD_PER_PCT:-3000}"
nopick_of() {
	local max="$max_gpd"
	if [ -n "$(field "$1" bench)" ]; then
		max="$(jq -n --argjson g "$max_gold_per_pct" --argjson r "$(field "$1" scoreRef)" '$g * 100 / $r')"
	fi
	jq -r --arg s "$1" --argjson max "$max" '[to_entries[] | select(.value.avail == "infeasible"
		or ((.value.goldPerDps | if type == "object" then .[$s] else null end) // 0) > $max) | .key] | join(",")' "$content"
}
nopick="$(nopick_of "$(echo "$spec" | cut -d, -f1)")"
# The gear search's pool, without the search: DRY_RUN=1 (see above).
if [ -n "$DRY_RUN" ]; then
	dry_dir="${DRY_RUN_DIR:-orchestration/round_notes_2026-11-11/phase0_pool}"
	mkdir -p "$dry_dir"
	go build --tags=with_db -o "$tmp/rotopt" ./tools/rotopt
	for s in $(echo "$spec" | tr , ' '); do
		jq '.raid.parties[0].players[0].rotation' "$(req_file "$s")" >"$tmp/$s.apl.json"
		settings "$s" >"$tmp/$s.json"
		# shellcheck disable=SC2046
		"$tmp/rotopt" -apl "$tmp/$s.apl.json" -settings "$tmp/$s.json" -durs "$dur" \
			-gear-search -gear-no-raids -gear-phase 2 -gear-min-ilvl 50 \
			-gear-exclude "$exclude" -gear-include "$include" -gear-no-pick "$(nopick_of "$s")" \
			$(weapon_flags "$s") $(bench_flags "$s") -gear-pool-spec "$s" \
			-gear-content "$content" -gear-pool-report "$dry_dir/$s.json" "$@" >/dev/null
		jq -r '"\(.spec): \(.items) items, \(.unmodelled | length) unmodelled, \(.hiddenStats | length) hidden stats, \(.unlisted) unlisted, \(.factionDropped | length) left out as \(if .faction == "Horde" then "Alliance" else "Horde" end) only"' "$dry_dir/$s.json"
	done
	exit 0
fi

# The list's picks, with their suffix and enchant, in the order of the equipment list, as a jq
# expression on $list[0]. An empty slot (the off hand next to a two hander) has no pick.
#
# A pick that is a no-pick item now ($nopick, say one that crossed the gold line since the list was
# made) gives its place to the best entry of its list that isn't one, and isn't the slot's other
# pick for rings and trinkets. It keeps the pick's enchant when it has none of its own. rotopt
# takes a no-pick item off the starting set, so without this a RANK_ONLY run started with the
# slot empty. On 2026-11-22 Treant's Bane went over the line and the run started at 563 DPS with
# no weapon.
picks='$list[0].slots as $s
	| ($nopick | split(",") | map(select(. != "") | tonumber)) as $np
	| def start($k): [$s[$k][]? | select(.bis)] as $p
		| reduce ($p[] | select(.id as $i | $np | index($i))) as $gone ([$p[] | select(.id as $i | $np | index($i) | not)];
			[.[].id] as $have
			| . + [first($s[$k][] | select(.id as $i | ($np | index($i) | not) and ($have | index($i) | not)))
				| if .enchant == null and $gone.enchant then .enchant = $gone.enchant else . end]);
	def pick($k; $i): (start($k)[$i]) as $p
		| if $p == null then {} else {id: $p.id} + (if $p.suffix then {randomSuffix: $p.suffix} else {} end)
			+ (if $p.enchant then {enchant: $p.enchant} else {} end) end;
	[pick("head"; 0), pick("neck"; 0), pick("shoulder"; 0), pick("back"; 0),
		pick("chest"; 0), pick("wrist"; 0), pick("hands"; 0), pick("waist"; 0), pick("legs"; 0), pick("feet"; 0),
		pick("finger"; 0), pick("finger"; 1), pick("trinket"; 0), pick("trinket"; 1), pick("mainhand"; 0), pick("offhand"; 0), pick("ranged"; 0)]'
# The gear set of a list's picks, without a search or a ranking: GEAR_ONLY=1 (see above).
if [ -n "$GEAR_ONLY" ]; then
	for s in $(echo "$spec" | tr , ' '); do
		g="${GEAR_OUT:-$(field "$s" gearOut)}"
		if [ -z "$g" ] || [ "$g" = none ]; then
			echo "SPEC $s has no gearOut in $specs" >&2
			exit 1
		fi
		f="${FROM:-$(list_path "$s")}"
		mkdir -p "$(dirname "$g")"
		jq -n --slurpfile list "$f" --arg nopick "$nopick" "$picks"' | {items: .}' >"$g"
		echo "wrote $g from $f"
	done
	exit 0
fi

jq '.raid.parties[0].players[0].rotation' "$req" >"$tmp/enh.apl.json"
settings "$spec" >"$tmp/enh.json"
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
	# Tank and healer lists rank on encounters, not fight lengths, so the check skips them.
	if [ "$from_durs" != "$dur/$use_dur" ] && [ -z "$MIX_DURS" ] && [ -z "$bench" ]; then
		echo "$from was ranked at durations[0]/use_dur $from_durs, not DUR/USE_DUR $dur/$use_dur." >&2
		echo "Rebuild it in full, or set MIX_DURS=1 to rerun on it anyway." >&2
		exit 1
	fi
	jq --slurpfile list "$from" --arg nopick "$nopick" "$picks"' as $items | .raid.parties[0].players[0].equipment.items = $items' \
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
mkdir -p "$(dirname "$out")"
outputs="-gear-rank-out $out"
if [ -n "$EFFECTS_ONLY" ]; then
	outputs=""
fi
if [ "$effects_out" != none ]; then
	outputs="$outputs -gear-effects-out $effects_out"
fi

# A tank or healer set's score is many sims (six or fourteen boss runs), each at the report's 1000
# iterations or more, so it runs fewer iterations and keeps fewer items for the full sims
# (Tanks and healers above). Arguments after ours still win.
iters="-iters 2000 -gear-screen-iters 300 -gear-screen-keep 40"
confirm="-gear-rank-confirm-iters 20000 -gear-rank-enchant-iters 10000"
if [ -n "$bench" ]; then
	iters="-iters 1000 -gear-screen-iters 300 -gear-screen-keep 15 -gear-ep-keep 25 -gear-ep-settings $req"
	confirm="-gear-rank-confirm-iters 5000 -gear-rank-enchant-iters 3000"
fi
# The conditional and cooldown ranks need the uses valued on their own (USE_DUR), and a boss file
# sets the fight a condition would change, so a tank or healer list has neither.
rank_extra="-gear-rank-conditions -gear-rank-cooldowns"
if [ -n "$bench" ]; then
	rank_extra=""
fi
# shellcheck disable=SC2046,SC2086
go run --tags=with_db ./tools/rotopt -apl "$tmp/enh.apl.json" -settings "$tmp/enh.json" -durs "$dur" $iters \
	-gear-search -gear-no-raids -gear-phase 2 -gear-min-ilvl 50 \
	-gear-exclude "$exclude" -gear-include "$include" -gear-no-pick "$nopick" -gear-content "$content" \
	-gear-rank-name "Phase 0" -gear-rank-keep 25 $(weapon_flags "$spec") $(bench_flags "$spec") \
	-gear-rank-spec "$(rank_spec "$spec")" -gear-rank-sim "$sim" \
	-gear-rank-enchants tools/rotopt/phase0_enchants.json $confirm \
	$rank_extra -gear-effects-iters 10000 \
	-use-dur "$use_dur" -use-iters 20000 -use-aoe-targets "$aoe_targets" $outputs $rank_only "$@"

if [ -n "$EFFECTS_ONLY" ]; then
	exit 0
fi

# The content check's status on every entry. An item over the gold line is infeasible, and
# its note says what a DPS (a tank or healer: a score unit) costs on it.
spec1="$(echo "$spec" | cut -d, -f1)"
list_max="$max_gpd"
unit="DPS"
if [ -n "$bench" ]; then
	list_max="$(jq -n --argjson g "$max_gold_per_pct" --argjson r "$(field "$spec1" scoreRef)" '$g * 100 / $r')"
	unit="point of score"
fi
jq --slurpfile content "$content" --argjson max "$list_max" --arg spec "$spec1" --arg unit "$unit" '.slots |= map_values(map(
	. as $e | ($content[0][$e.id | tostring] // {}) as $c
	| ($c.goldPerDps | if type == "object" then .[$spec] else null end) as $gpd
	| (($gpd // 0) > $max) as $costly
	| ([$c.note // empty] + (if $gpd then ["\($gpd | round)g per \($unit)"] else [] end) | join(" · ")) as $note
	| $e + (if $c.avail then {avail: $c.avail} elif $costly then {avail: "infeasible"} else {} end)
		+ (if $note != "" then {note: $note} else {} end)))' \
	"$out" >"$tmp/list.json"
mv "$tmp/list.json" "$out"

# The unchecked items and enchants: picks and top 5s the content checks don't list (see above).
unlisted="$(jq -r --slurpfile content "$content" --slurpfile enchants "$enchants" '
	[.slots | to_entries[] | .key as $slot | .value | to_entries[]
		| select(.key < 5 or .value.bis) | .value | select($content[0][.id | tostring] == null)
		| "item \(.id) \(.name) (\($slot)\(if .bis then ", pick" else "" end))"]
	+ [(.enchants // {}) | to_entries[] | .key as $slot | .value | to_entries[]
		| select(.key < 5 or .value.bis) | .value | select($enchants[0][.spell | tostring] == null)
		| "enchant \(.spell) \(.name) (\($slot)\(if .bis then ", pick" else "" end))"]
	| unique[]' "$out")"
if [ -n "$unlisted" ]; then
	echo "Not in $content or $enchants, so not checked for Phase 0:" >&2
	echo "$unlisted" >&2
	if [ -z "$ALLOW_UNLISTED" ]; then
		echo "Check them and add them, or set ALLOW_UNLISTED=1. $out is written." >&2
		exit 1
	fi
fi

# The picks as the spec's Phase 0 gear set (gearOut), when we wrote the real list: the sim UI's
# "Phase 0" gear preset (ui/enhancement_shaman/presets.ts) and the DPS Report's gear for the
# spec. GEAR_OUT=none skips it.
gear_out="${GEAR_OUT:-$(field "$spec" gearOut)}"
gear_out="${gear_out:-none}"
if [ -z "$OUT" ] && [ "$gear_out" != none ]; then
	jq -n --slurpfile list "$out" --arg nopick "$nopick" "$picks"' | {items: .}' >"$gear_out"
fi
