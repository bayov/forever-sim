#!/bin/sh
# Regenerates the Level 60 warrior rotations in ui/warrior/apls from the warrior_fury and
# warrior_arms templates, with the knobs the searches settled on (2026-10-12, Human, 180 sec,
# the synthetic Phase 1 sets and the Level 60 talents, see ui/warrior/presets.ts). The settings
# come from tools/rotopt/mksettings (spec fury or arms). The rotation itself does not depend on
# them, only the one sanity sim does.
#
#   go run --tags=with_db ./tools/rotopt/mksettings -spec fury -gear level60_phase1_synthetic_fury \
#       -apl level60_fury -talents 30305013002-25053005112110501-2 -race Human -duration 180 -out fury.json
#   tools/rotopt/warrior_presets.sh fury.json
#
# The searches: tools/rotopt -talent-search and the knob search, one after the other until
# neither moved, from these talents and knobs.
set -e
settings=$1
gen() {
	template=$1
	name=$2
	talents=$3
	knobs=$4
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 180 -combined \
		-eval -confirm-iters 100 -out ui/warrior/apls -name "$name" -talents "$talents" -knobs "$knobs" >/dev/null
	echo "wrote ui/warrior/apls/$name.apl.json ($knobs)"
}
gen warrior_fury level60_fury 30305013002-25053005112110501-2 \
	battleShout=0,berserkerRageRage=0,bloodrageRage=10,deathWish=2,dwEnd=33,executeFirst=1,fillerGap=0,hamstringRage=10,hsExecute=0,hsRage=20,potionHoldForDw=0,racialHoldForDw=1,recklessAt=25,recklessness=2,slamRage=0,wwBtGap=0,wwFirst=0,wwRage=25
gen warrior_fury level60_fury_2h 30305013032-25053000152110501 \
	battleShout=0,berserkerRageRage=0,bloodrageRage=70,deathWish=2,dwEnd=30,executeFirst=1,fillerGap=0,hamstringRage=10,hsExecute=0,hsRage=70,potionHoldForDw=1,racialHoldForDw=1,recklessAt=25,recklessness=2,slamRage=0,wwBtGap=0,wwFirst=0,wwRage=25
gen warrior_arms level60_arms 20305213032515201-55050000001 \
	battleShout=0,berserkerRageRage=0,bloodrageRage=70,deathWish=2,dwEnd=36,executeFirst=0,hsExecute=0,hsRage=50,msFirst=1,overpowerRage=5,potionHoldForDw=1,racialHoldForDw=1,recklessAt=20,recklessness=2,rend=1,slamAfterSwing=0,slamRage=30,spearingRage=0,stance=1,wwRage=25
