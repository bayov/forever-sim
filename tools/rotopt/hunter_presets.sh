#!/bin/sh
# Regenerates the hunter rotations in ui/hunter/apls from the hunter_ranged and hunter_sv
# templates, with the knobs the 2026-10-12 searches settled on (at 180 sec on the synthetic
# Phase 1 sets and the Level 60 talents). Run from the repo root with a settings file the
# optimizer should sim against, made with mksettings for the spec. The rotation itself does
# not depend on the settings, only the one sanity sim does.
#
#   go run --tags=with_db ./tools/rotopt/mksettings -spec hunter_bm -gear level60_phase1_synthetic_ranged \
#       -apl bm -talents 5020121505021241-00520503014 -race NightElf -duration 180 -out bm.json
#   tools/rotopt/hunter_presets.sh bm.json
set -e
settings=$1
gen() {
	template=$1
	name=$2
	knobs=$3
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 180 -combined \
		-eval -confirm-iters 100 -out ui/hunter/apls -name "$name" -knobs "$knobs" >/dev/null
	echo "wrote ui/hunter/apls/$name.apl.json ($knobs)"
}
# Beast Mastery and Marksmanship (and Lone Wolf, which plays the Marksmanship rotation): the
# knob searches on the searched builds kept the starting knobs. Summon Hawk while fewer than 2
# hawks are out (hawk=2) only matters with the talent, which neither build takes.
ranged="aimed=1,aimedMana=0,arcaneMana=20,hawk=2,intimidation=0,serpent=1,serpentMana=30,sniper=1,sniperFirst=0"
gen hunter_ranged bm "$ranged"
gen hunter_ranged mm "$ranged"
# Survival: Wing Clip in every free global cooldown, +31 at 180 sec.
gen hunter_sv sv "lacerate=1,lacerateMana=20,strikerMana=0,trap=1,trapMana=30,wingClip=1,wingClipMana=50"
