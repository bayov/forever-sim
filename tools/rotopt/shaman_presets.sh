#!/bin/sh
# Regenerates the enhancement shaman presets in ui/enhancement_shaman/apls that come from
# the shaman_enh template. The level 60 one is the optimizer's best on the Phase 2 gear
# and the Level 60 talents. The level 20 one is the same template with the knobs the
# level 20 search settled on (no Maelstrom Weapon yet, Frost Shock, Lightning Shield up
# because the boss is hitting the shaman). The hand written air totem presets are not
# touched. Run from the repo root with the settings file the optimizer should sim
# against (a RaidSimRequest or a UI settings export). The rotation itself does not
# depend on the settings, only the one sanity sim does.
#
#   tools/rotopt/shaman_presets.sh path/to/settings.json
set -e
settings=$1
gen() {
	name=$1
	knobs=$2
	shift 2
	go run --tags=with_db ./tools/rotopt -template shaman_enh -settings "$settings" -durs 300 \
		-eval -confirm-iters 100 -out ui/enhancement_shaman/apls -name "$name" -knobs "$knobs" "$@" >/dev/null
	mv "ui/enhancement_shaman/apls/${name}_300s.apl.json" "ui/enhancement_shaman/apls/$name.apl.json"
	echo "wrote ui/enhancement_shaman/apls/$name.apl.json ($knobs $*)"
}
gen optimized "shock=4" -level 60
gen level20 "airTotem=0,strengthOfEarth=0,fireTotemMinTime=10,shield=1,shock=3,maelstromStacks=0" -level 20
# At 25 Stormstrike is in reach (a 15 point talent under Forever), Earth Shock beats
# Frost Shock by 7 DPS and the rank 2 Strength of Earth is worth its mana.
gen level25 "airTotem=0,strengthOfEarth=1,fireTotemMinTime=10,shield=1,shock=1,maelstromStacks=0" -level 25
