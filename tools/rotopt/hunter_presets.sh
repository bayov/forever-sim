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
# The ranged specs, from the knob searches on the standard gear (2026-10-12). Marksmanship and Lone
# Wolf put the Arcane Shot cooldown above Aimed Shot and drop the mana floors (Lone Wolf also casts
# Serpent Sting after Aimed Shot, +1.7). Beast Mastery keeps Aimed Shot first and the floors. All three
# cast Rapid Fire 0.1 sec before the pull (+2 at 120 sec), and Marksmanship and Lone Wolf cast Sniper
# Shot only while the mana covers 25 a second of the fight left (+8 and +11 at 300 sec, nothing at 180;
# both from the rotation auditor). Summon Hawk while fewer than 2 hawks are out (hawk=2) only matters
# with the talent, which no build takes.
ranged="aimed=1,aimedMana=0,hawk=2,intimidation=0,serpent=1,sniper=1,sniperFirst=0,rapidPrepull=1"
gen hunter_ranged bm "$ranged,arcaneFirst=0,arcaneMana=20,serpentLate=0,serpentMana=30"
gen hunter_ranged mm "$ranged,arcaneFirst=1,arcaneMana=0,serpentLate=0,serpentMana=0,sniperManaPerSec=25"
gen hunter_ranged lw "$ranged,arcaneFirst=1,arcaneMana=0,serpentLate=1,serpentMana=0,sniperManaPerSec=25"
# Survival: Wing Clip in every free global cooldown (+31 at 180 sec), held when Mongoose Bite or
# Strider Kick is about ready (+0.4).
gen hunter_sv sv "lacerate=1,lacerateMana=20,strikerMana=0,trap=1,trapMana=30,wingClip=1,wingClipMana=50,wingClipHold=1"
