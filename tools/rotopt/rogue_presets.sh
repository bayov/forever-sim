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
	shift 3
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 300 \
		-eval -confirm-iters 100 -out ui/rogue/apls -name "$name" -knobs "$knobs" "$@" >/dev/null
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
# Blade Flurry no longer waits for Slice and Dice. Under Forever's continuous energy that
# is worth 2 DPS at 300 sec, with ticks it was a tie.
gen rogue_bs combat_backstab "ghostly=0,bfSnd=0"
gen rogue_bs combat_backstab_rupture "ghostly=0,bfSnd=0,ruptureCp=5"
gen rogue_bs combat_backstab_on_cooldown "ghostly=0,arSnd=0,bfSnd=0,cbSnd=0,eurekaSnd=0,bloodFuryHoldForAr=0,eluneHoldForAr=0"
gen rogue_bs combat_backstab_snd_window "ghostly=0,bloodFurySnd=20,bfSnd=20,cbSnd=18,trinketSnd=20"

# The low level presets are the same templates with the knobs the level 20 search settled
# on, at the 11 talent points a level 20 has. The lines name the level 60 spell ranks and
# the sim resolves them to the rank the level knows. The level is passed so the template
# leaves out Adrenaline Rush and Blade Flurry, which the points cannot reach. Rupture at 3
# points beats Eviscerate at this level. With finishers at 3 points the rotation never
# sits on 5, so Eureka! goes on cooldown (+3 for a Gnome) and, for anyone who picks Cold
# Blood, the points are held for its Eviscerate while it is ready (cbPool, free for builds
# without it).
#
# A fresh Slice and Dice waits for 4 points (sndCp, +1.4 for Backstab, +0.4 for Sinister
# Strike): a 1 point one runs out before the next point arrives, so it would take every
# point and Rupture would never go out. Ghostly Strike and Hemorrhage are out of the
# Backstab build's reach, so their lines are left out.
low="cbPool=1,cbSnd=0,eurekaCp=0"
gen rogue_ss level20_sinister_strike "ruptureCp=3,ruptureSnd=0,ssEnergy=10,sndCp=4,$low" -level 20 -bonus-talents 0
gen rogue_bs level20_backstab "ruptureCp=3,ruptureSnd=0,ghostly=0,hemoForRupture=0,sndCp=4,$low" -level 20 -bonus-talents 0
