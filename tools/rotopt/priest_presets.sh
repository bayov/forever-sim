#!/bin/sh
# Regenerates the shadow priest Level 60 rotation in ui/shadow_priest/apls from the
# priest_shadow template, with the knobs the searches settled on (2026-10-12): the
# template's defaults, which are the searched values at 120, 180 and 300 sec.
#
# The settings, talent and knob searches behind the Level 60 presets:
#
#   go run --tags=with_db ./tools/rotopt/mksettings -spec shadow -gear level60_phase1_synthetic \
#       -apl level60 -talents 005030211305--500321501201302051 -race Human -duration 180 -out sp.json
#   go run --tags=with_db ./tools/rotopt -template priest_shadow -settings sp.json -durs 180 \
#       -talent-search -iters 3000 -confirm-iters 20000
#   go run --tags=with_db ./tools/rotopt -template priest_shadow -settings sp.json -durs 120,180,300 \
#       -iters 3000 -confirm-iters 20000
#
# Run from the repo root with the settings file. The rotation itself does not depend on the
# settings, only the one sanity sim does.
#
#   tools/rotopt/priest_presets.sh path/to/settings.json
set -e
settings=$1
go run --tags=with_db ./tools/rotopt -template priest_shadow -settings "$settings" -durs 180 \
	-eval -confirm-iters 100 -out ui/shadow_priest/apls -name level60 >/dev/null
mv ui/shadow_priest/apls/level60_180s.apl.json ui/shadow_priest/apls/level60.apl.json
echo "wrote ui/shadow_priest/apls/level60.apl.json"
