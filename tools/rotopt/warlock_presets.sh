#!/bin/sh
# Regenerates the Level 60 warlock rotations in ui/warlock/apls from the warlock template
# (tools/rotopt/warlock.go). The knobs are what the knob search settled on for each build on
# 2026-10-12 (see ui/warlock/presets.ts). Run from the repo root. The rotation does not depend
# on the settings, only the one sanity sim does, so any level 60 warlock settings will do:
#
#   go run --tags=with_db ./tools/rotopt/mksettings -spec demonology -gear level60_phase1_synthetic \
#       -apl level60_demonology -talents 011-0305003131221001351-0550005003 -race Human -duration 180 -out lock60.json
#   tools/rotopt/warlock_presets.sh lock60.json
#
# The searches behind them (one talent search per starting build, then a knob search, then the
# talent search again on the searched knobs):
#
#   go run --tags=with_db ./tools/rotopt -template warlock -settings lock60.json -durs 180 -talent-search \
#       -iters 2000 -knobs bane=2,corruption=1,immolate=1,siphonLife=1,conflagrate=1,shadowburn=2,soulFire=0,filler=2,soulLink=1,lifeTapMana=5
#   go run --tags=with_db ./tools/rotopt -template warlock -settings lock60.json -durs 180 -iters 3000 -knobs ...
set -e
settings=$1
gen() {
	name=$1
	knobs=$2
	go run --tags=with_db ./tools/rotopt -template warlock -settings "$settings" -durs 180 \
		-eval -confirm-iters 100 -out ui/warlock/apls -name "$name" -knobs "$knobs" >/dev/null
	mv "ui/warlock/apls/${name}_180s.apl.json" "ui/warlock/apls/$name.apl.json"
	echo "wrote ui/warlock/apls/$name.apl.json ($knobs)"
}
base="bane=2,corruption=1,immolate=1,filler=1,lifeTapMana=5,siphonLife=0,shadowburn=0,soulFire=0"
# Demonology: Soul Link before the pull.
gen level60_demonology "$base,conflagrate=0,soulLink=1"
# Destruction: Conflagrate on cooldown while Immolate is up.
gen level60_destruction "$base,conflagrate=1,soulLink=0"
gen level60_affliction "$base,conflagrate=0,soulLink=0"
