#!/bin/sh
# Regenerates the retribution paladin presets in ui/retribution_paladin/apls that come
# from the paladin_ret template. The level 30 one is the knobs the level 20 and 30 searches
# settled on at 60 sec: Judgement of the Crusader up first, then Consecration ahead of the seal
# and judgement lines down to the last point of mana, Seal of Command and Holy Strike on
# cooldown. Over 120 sec a 20% mana floor for Consecration is 1.0 better on one target,
# because it keeps the seal and judgement going.
#
# Run from the repo root with the settings file the optimizer should sim against (a
# RaidSimRequest or a UI settings export). The rotation itself does not depend on the
# settings, only the one sanity sim does.
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
# Level 30 with 5 extra talent points (26 in all). The old level 20 rotation still wins,
# and Seal of Command stays ahead of Seal of Righteousness on Corpsemaker.
gen level30p5 "seal=2,consecration=1,consecrationFirst=1,consecrationMana=0" -level 30 -bonus-talents 5 -talents --502240312013201
