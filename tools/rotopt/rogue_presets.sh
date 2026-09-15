#!/bin/sh
# Regenerates the rogue presets in ui/rogue/apls from the rotopt templates. The plain
# Sinister Strike and Backstab ones are the optimizer's best. The others are the same
# template with knobs set to the alternatives people ask about, so they can be compared
# in the UI. Run from the repo root with the settings file the optimizer should sim
# against (a RaidSimRequest or a UI settings export). The rotation itself does not
# depend on the settings, only the one sanity sim does.
#
#   tools/rotopt/rogue_presets.sh path/to/settings.json
set -e
settings=$1
gen() {
	template=$1
	name=$2
	knobs=$3
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 300 \
		-eval -confirm-iters 100 -out ui/rogue/apls -name "$name" -knobs "$knobs" >/dev/null
	mv "ui/rogue/apls/${name}_300s.apl.json" "ui/rogue/apls/$name.apl.json"
	echo "wrote ui/rogue/apls/$name.apl.json ($knobs)"
}
gen rogue_ss combat_sinister_strike ""
gen rogue_ss combat_sinister_strike_on_cooldown "arSnd=0,bfSnd=0,cbSnd=0,eurekaSnd=0,bloodFuryHoldForAr=0,eluneHoldForAr=0"
gen rogue_ss combat_sinister_strike_hold_for_ar "bfHoldForAr=1,berserkingHoldForBf=1,trinketHoldForAr=1"
gen rogue_ss combat_sinister_strike_snd_opener "arSnd=15,bfHoldForAr=1"
gen rogue_ss combat_sinister_strike_snd_window "bloodFurySnd=20,bfSnd=20,cbSnd=18,trinketSnd=20"

# The Backstab build the talent search settled on has no Subtlety talents, so the
# Ghostly Strike line is left out of the shipped presets (it would only warn).
gen rogue_bs combat_backstab "ghostly=0"
gen rogue_bs combat_backstab_rupture "ghostly=0,ruptureCp=5"
gen rogue_bs combat_backstab_on_cooldown "ghostly=0,arSnd=0,bfSnd=0,cbSnd=0,eurekaSnd=0,bloodFuryHoldForAr=0,eluneHoldForAr=0"
gen rogue_bs combat_backstab_snd_window "ghostly=0,bloodFurySnd=20,bfSnd=20,cbSnd=18,trinketSnd=20"
