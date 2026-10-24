#!/bin/sh
# Regenerates the priest Level 60 rotations in ui/shadow_priest/apls from the priest_shadow
# and priest_smite templates, with the knobs the searches settled on (2026-10-12, rerun on
# 2026-10-23 with the wand knobs, which keep every value): the
# templates' defaults, which are the searched values at 120, 180 and 300 sec.
#
# The settings, talent and knob searches behind the Level 60 presets:
#
#   go run --tags=with_db ./tools/rotopt/mksettings -spec shadow -gear level60_phase1_synthetic \
#       -apl level60 -talents 00530400102-3-50052101231302251 -race Human -duration 180 -out sp.json
#   go run --tags=with_db ./tools/rotopt -template priest_shadow -settings sp.json -durs 120,180,300 \
#       -talent-search -require silentResolve -iters 3000 -confirm-iters 20000
#
# (Silent Resolve is a must-take, standards 26. The talent search scores all three lengths since
# 2026-10-23: at 180 sec alone it takes Meditation down to 1, which costs 9 DPS at 300 sec.)
#   go run --tags=with_db ./tools/rotopt -template priest_shadow -settings sp.json -durs 120,180,300 \
#       -iters 3000 -confirm-iters 20000
#
# Run from the repo root with the settings file. The rotation itself does not depend on the
# settings, only the one sanity sim does.
#
#   tools/rotopt/priest_presets.sh path/to/settings.json
set -e
settings=$1
gen() {
	template=$1
	name=$2
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 180 \
		-eval -confirm-iters 100 -out ui/shadow_priest/apls -name "$name" >/dev/null
	mv "ui/shadow_priest/apls/${name}_180s.apl.json" "ui/shadow_priest/apls/$name.apl.json"
	echo "wrote ui/shadow_priest/apls/$name.apl.json"
}
gen priest_shadow level60
gen priest_smite level60_smite
