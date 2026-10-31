#!/bin/sh
# Regenerates the rogue rotations in ui/rogue/apls from the rotopt templates. Run from the
# repo root with the settings file the optimizer should sim against (a RaidSimRequest or a UI
# settings export). The rotation itself does not depend on the settings, only the one sanity
# sim does.
#
#   tools/rotopt/rogue_presets.sh path/to/settings.json
set -e
settings=$1
gen() {
	template=$1
	name=$2
	knobs=$3
	shift 3
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 300 \
		-eval -confirm-iters 100 -out ui/rogue/apls -name "$name" -knobs "$knobs" "$@" >/dev/null
	mv "ui/rogue/apls/${name}_300s.apl.json" "ui/rogue/apls/$name.apl.json"
	echo "wrote ui/rogue/apls/$name.apl.json ($knobs)"
}
# The plain Sinister Strike and Backstab rotations, which the rogue Go tests run. The UI dropped
# them, their variants and the level 30 rotations on 2026-10-31.
gen rogue_ss combat_sinister_strike ""
# The Backstab build the talent search settled on has no Subtlety talents, so the
# Ghostly Strike line is left out (it would only warn).
# Blade Flurry no longer waits for Slice and Dice. Under Forever's continuous energy that
# is worth 2 DPS at 300 sec, with ticks it was a tie.
gen rogue_bs combat_backstab "ghostly=0,bfSnd=0"

# The Level 60 rotations, searched at 180 sec on the generic boss (mksettings' Level 60 target, no
# creature type) and the synthetic Phase 1 sets with the
# mksettings rogue_combat, rogue_assassination and rogue_subtlety setups (knob search,
# talent search, knob search again). These need no settings argument, each builds its own.
# Seal Fate Daggers uses the Combat Daggers rotation, the search kept every knob. Seal Fate Swords
# is the Combat Swords one with 4 point Eviscerates.
gen60() {
	spec=$1
	gear=$2
	template=$3
	name=$4
	talents=$5
	knobs=$6
	settings=$(mktemp)
	go run --tags=with_db ./tools/rotopt/mksettings -spec "$spec" -gear "$gear" -apl level60_combat_daggers \
		-talents "$talents" -race Undead -duration 180 -out "$settings"
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 180 \
		-eval -confirm-iters 100 -out ui/rogue/apls -name "$name" -knobs "$knobs" >/dev/null
	mv "ui/rogue/apls/${name}_180s.apl.json" "ui/rogue/apls/$name.apl.json"
	rm "$settings"
	echo "wrote ui/rogue/apls/$name.apl.json ($knobs)"
}
gen60 rogue_combat level60_phase1_synthetic_swords rogue_ss level60_combat_swords 00530310501-32003311201515231 "arEnergy=39,bfPrepull=1"
# Combat Daggers and Seal Fate Daggers are Gnome since the race-tuning round (2026-10-27, standards 29
# and 16), so Eureka! goes on cooldown (eurekaCp=0) instead of waiting for a 5 point Eviscerate these 4
# point Eviscerate builds seldom reach: +8.9 / +5.4 at 180 sec for a Gnome, 0.0 for any other race.
gen60 rogue_combat level60_phase1_synthetic_combat_daggers rogue_bs level60_combat_daggers 005303104-30230320201515231-002 "arEnergy=39,bfPrepull=1,bfSnd=0,bsAfterSwing=0.25,bsEnergy=100,evisCp=4,ghostly=0,vanishCp=3,eurekaCp=0"
gen60 rogue_assassination level60_phase1_synthetic_mutilate rogue_mut level60_mutilate 0053031033150103-302303202015-002 "bsAfterSwing=0.75,sndRefresh=1,evisCp=4,vanishCp=3"
gen60 rogue_subtlety level60_phase1_synthetic_subtlety rogue_sub level60_subtlety 0053031035140004--3021203310113011 "bsAfterSwing=0.25,hemoFiller=1,ruptureCp=3,ruptureSnd=0,vanishCp=3,vanishPool=1"
gen60 rogue_combat level60_phase1_synthetic_seal_fate_swords rogue_ss level60_seal_fate_swords 0053031034150005-3220132020141 "arEnergy=39,bfPrepull=1,evisCp=4"
