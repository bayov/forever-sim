#!/bin/sh
# Regenerates the retribution paladin presets in ui/retribution_paladin/apls that come
# from the paladin_ret template. The level 20 one is the knobs the level 20 search settled
# on: Seal of Command (the 11 point build), Judgement of the Crusader up first, Holy
# Strike on cooldown and Consecration down to a 20% mana floor (dumping every last point
# into it is 0.4 better at 60 sec and 0.5 worse at 120, the floor keeps the seal and
# judgement going). Run from the repo root with the settings file the optimizer should
# sim against (a RaidSimRequest or a UI settings export). The rotation itself does not
# depend on the settings, only the one sanity sim does.
#
#   tools/rotopt/paladin_presets.sh path/to/settings.json
set -e
settings=$1
gen() {
	name=$1
	knobs=$2
	shift 2
	go run --tags=with_db ./tools/rotopt -template paladin_ret -settings "$settings" -durs 60 \
		-eval -confirm-iters 100 -out ui/retribution_paladin/apls -name "$name" -knobs "$knobs" "$@" >/dev/null
	mv "ui/retribution_paladin/apls/${name}_60s.apl.json" "ui/retribution_paladin/apls/$name.apl.json"
	echo "wrote ui/retribution_paladin/apls/$name.apl.json ($knobs $*)"
}
gen level20 "seal=2,consecration=1,consecrationMana=20" -level 20 -bonus-talents 0
