#!/bin/sh
# Regenerates the rogue Sinister Strike presets in ui/rogue/apls from the rogue_ss
# template. The first one is the optimizer's best. The others are the same template
# with knobs set to the alternatives people ask about, so they can be compared in the
# UI. Run from the repo root with the settings file the optimizer should sim against
# (a RaidSimRequest or a UI settings export).
#
#   tools/rotopt/rogue_presets.sh path/to/settings.json
set -e
settings=$1
gen() {
	name=$1
	knobs=$2
	go run --tags=with_db ./tools/rotopt -template rogue_ss -settings "$settings" -durs 300 \
		-eval -confirm-iters 100 -out ui/rogue/apls -knobs "$knobs" >/dev/null
	mv ui/rogue/apls/rogue_ss_300s.apl.json "ui/rogue/apls/$name.apl.json"
	echo "wrote ui/rogue/apls/$name.apl.json ($knobs)"
}
gen combat_sinister_strike ""
gen combat_sinister_strike_on_cooldown "arSnd=0,bfSnd=0,cbSnd=0,eurekaSnd=0,bloodFuryHoldForAr=0,eluneHoldForAr=0"
gen combat_sinister_strike_hold_for_ar "bfHoldForAr=1,berserkingHoldForBf=1,trinketHoldForAr=1"
gen combat_sinister_strike_snd_opener "arSnd=15,bfHoldForAr=1"
gen combat_sinister_strike_snd_window "bloodFurySnd=20,bfSnd=20,cbSnd=18,trinketSnd=20"
