#!/bin/sh
# Regenerates the shaman BiS lists in ui/*_shaman/bis (docs/bis_lists.md). For each spec we
# build its fight with mksettings, run the gear search on it, and rank every slot's options
# against the set the search settles on (-gear-rank-out, bis.go). Run from the repo root:
#
#   tools/rotopt/bis_lists.sh
#
# Level 30 is the beta's level cap, so it stands in for endgame until Forever releases. The
# fight is mksettings' level 30 one, stretched to the 120 sec of the UI's Level 30
# encounter: Interrogator Vishas (level 32), soloed with Blessing of Kings, no consumes.
# The search leaves out items that need a level above 30, and the synthetic Phase 1 items.
#
# Each list takes a while, because the pools are big at level 30 (about 2000 backs, off
# hands and chests once every random suffix counts). Every slot is screened at 500
# iterations and the best 20 are scored again at 3000. On 2026-10-16 enhancement took 14 min
# (two search rounds, then 5 min of ranking) and elemental 13 min (three rounds, then 3 min),
# on 16 cores with another sim job running next to it.
set -e

# The commit the lists are simmed on, with a "+" when the working copy has changes on top.
sim="$(jj log --no-graph -r @- -T 'commit_id.short()')$(jj log --no-graph -r @ -T 'if(empty, "", "+")')"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# bis spec template settings out knobs [rotopt flags...]
bis() {
	spec=$1
	template=$2
	settings=$3
	out=$4
	knobs=$5
	shift 5
	mkdir -p "$(dirname "$out")"
	echo "== $out"
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 120 -knobs "$knobs" \
		-gear-search -gear-screen-iters 500 -gear-screen-keep 20 \
		-gear-rank-out "$out" -gear-rank-name "Level 30" -gear-rank-spec "$spec" -gear-rank-sim "$sim" "$@"
}

# Enhancement starts from the old Level 30 + 5 gear and talents, with the knobs its rotation had
# (the level30p5 line shaman_presets.sh had until 2026-10-31). The UI dropped its Level 30 presets
# that day, so the set is written out here: mksettings takes the Level 60 set and we replace it.
# The rotation comes from the template and the knobs, so the APL we name doesn't matter. Manual
# Crowd Pummeler (9449) is left out, because the sim gives it its haste every 30 sec while its 3
# charges never come back in 1.12.
go run --tags=with_db ./tools/rotopt/mksettings -spec enh -gear level60_phase1_synthetic -apl optimized \
	-talents -0540320010051122 -race Orc -level 30 -bonus-talents 5 -duration 120 -out "$tmp/enh60gear.json"
jq '.raid.parties[0].players[0].equipment = {items: [{id: 3020}, {id: 13084, enchant: 1249057}, {id: 6697},
	{id: 7460, randomSuffix: 936, enchant: 13419}, {id: 273026, enchant: 866}, {id: 9428, randomSuffix: 176, enchant: 1248498},
	{id: 277223, enchant: 13887}, {id: 252461}, {id: 6690, enchant: 1255124}, {id: 15142, randomSuffix: 942, enchant: 13637},
	{id: 281635}, {id: 13097}, {}, {}, {id: 280604, enchant: 13695}, {}, {id: 249398}]}' \
	"$tmp/enh60gear.json" >"$tmp/enh.json"
bis enh shaman_enh "$tmp/enh.json" ui/enhancement_shaman/bis/level30.bis.json \
	"airTotem=0,strengthOfEarth=1,fireTotemMinTime=10,shield=1,shock=4,shockMana=10,waterTotem=1,maelstromStacks=0,fireTotem=3,fireNova=2,fireNovaMana=80,prepullTotems=1" \
	-gear-exclude 9449

# Elemental has no level 30 preset. The talents are the Level 60 build
# 5504301300103051-052-05305 cut to 26 points, Elemental first, the way bayovui's
# levelweights.py cuts it. The knobs are the Level 60 ones from shaman_presets.sh without
# Lightning Shield, since nothing hits a shaman at range in this fight. mksettings has no level
# 30 elemental gear, so we take its Level 60 set and replace it with only a main hand and an
# off hand: Soulsplatter Mace (271804) and Omega Orb (7749).
#
# From an empty set the search took Greatstaff of the Necrokhans in round 1, and from there it
# never tried a one hander with an off hand, since nothing was in the off hand to keep. The
# ranking then found Soulsplatter Mace with Omega Orb 1.9 ahead of the staff (2026-10-16). With
# both hands filled at the start, each two hander is weighed against the one hander and off
# hand together, so the search can still pick a staff when one wins.
go run --tags=with_db ./tools/rotopt/mksettings -spec ele -gear phase_1 -apl default \
	-talents 550430130010301 -race Orc -level 30 -bonus-talents 5 -duration 120 -out "$tmp/ele60gear.json"
jq '.raid.parties[0].players[0].equipment = {items: ([range(14) | {}] + [{id: 271804}, {id: 7749}])}' \
	"$tmp/ele60gear.json" >"$tmp/ele.json"
bis ele shaman_ele "$tmp/ele.json" ui/elemental_shaman/bis/level30.bis.json \
	"chainLightning=0,lightningBoltDownrank=2,lavaBurstFlameShockLeft=2,prepullBolt=1,earthShockFinish=1,waterTotemMinTime=20,flameShockMinTime=9"
