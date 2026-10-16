#!/bin/sh
# Regenerates the hunter rotations in ui/hunter/apls from the hunter_ranged and hunter_sv
# templates, with the knobs the 2026-10-12 searches settled on (at 180 sec on the synthetic
# Phase 1 sets and the Level 60 talents), confirmed on the standards 1b gear on 2026-10-13.
# Run from the repo root with a settings file the optimizer should sim against, made with
# mksettings for the spec. The rotation itself does not depend on the settings, only the one
# sanity sim does.
#
#   go run --tags=with_db ./tools/rotopt/mksettings -spec hunter_bm -gear level60_phase1_synthetic_ranged \
#       -apl bm -talents 5020121505021241-00520503014 -race SkyborneWindshaper -duration 180 -out bm.json
#   tools/rotopt/hunter_presets.sh bm.json
set -e
settings=$1
gen() {
	template=$1
	name=$2
	knobs=$3
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 180 -combined \
		-eval -confirm-iters 100 -out ui/hunter/apls -name "$name" -knobs "$knobs" >/dev/null
	echo "wrote ui/hunter/apls/$name.apl.json ($knobs)"
}
# The ranged specs, from the knob searches on the standard gear (2026-10-12). Marksmanship and Lone
# Wolf put the Arcane Shot cooldown above Aimed Shot, drop the mana floors and cast Serpent Sting
# after Aimed Shot (+1.7 Lone Wolf, +2.1 Marksmanship with Forever's consumables). Beast Mastery keeps Aimed Shot first and the floors. All three
# cast Rapid Fire 0.1 sec before the pull (+2 at 120 sec), and Marksmanship and Lone Wolf cast Sniper
# Shot only while the mana covers 25 a second of the fight left (+8 and +11 at 300 sec, nothing at 180;
# both from the rotation auditor). Summon Hawk while fewer than 2 hawks are out (hawk=2) only matters
# with the talent, which no build takes.
ranged="aimed=1,aimedMana=0,hawk=2,intimidation=0,serpent=1,sniper=1,sniperFirst=0,rapidPrepull=1"
bm="$ranged,arcaneFirst=0,arcaneMana=20,serpentLate=0,serpentMana=30,holdElune=1"
mm="$ranged,arcaneFirst=1,arcaneMana=0,serpentLate=1,serpentMana=0,sniperManaPerSec=25,holdRapid=1,holdElune=1"
gen hunter_ranged bm "$bm"
gen hunter_ranged mm "$mm"
gen hunter_ranged lw "$mm"
# Marksmanship and Lone Wolf hold Rapid Fire and Elune's Light for the on-use trinket when the wait
# costs no use, Beast Mastery only Elune's Light (rotation auditor, 2026-10-13: +2.3 / +1.7 / +0.5
# at 300 sec, nothing at 120 and 180).
# Survival: Wing Clip in every free global cooldown (+31 at 180 sec), held when Mongoose Bite or
# Strider Kick is about ready (+0.4).
sv="lacerate=1,lacerateMana=20,strikerMana=0,trap=1,trapMana=30,wingClip=1,wingClipMana=50,wingClipHold=1"
gen hunter_sv sv "$sv"
# Hunter's Mark duty (standards 23): the base rotation plus the mark before the pull and before it drops,
# for the presets with the raid's mark off (mksettings specs hunter_*_mark).
# The mark goes up 1.5 sec before the pull (it doesn't pull; user ruling 2026-10-15), after the
# pre-pull Aspect's global cooldown, so the boss has it from the first shot (rotation audit pass 20:
# +4.7 / +3.3 / +3.4 / +1.4 at 180 sec on Beast Mastery / Marksmanship / Lone Wolf / Survival, 20000
# iterations). It is not held back for the hunter's own DPS (user ruling).
gen hunter_ranged bm_mark "$bm,mark=1,markPrepull=1.5"
gen hunter_ranged mm_mark "$mm,mark=1,markPrepull=1.5"
gen hunter_ranged lw_mark "$mm,mark=1,markPrepull=1.5"
gen hunter_sv sv_mark "$sv,mark=1,markPrepull=1.5"
# Scorpid Sting duty (standards 23), for the ranged hunters only (user ruling 2026-10-15, not melee
# Survival): Scorpid Sting kept up in place of Serpent Sting, for the presets with the raid's Scorpid
# Sting off (mksettings specs hunter_{bm,mm,lw}_scorpid). Without Serpent Sting Marksmanship and Lone
# Wolf have the mana for Sniper Shot at 15 a second of the fight left (+15 and +20 at 300 sec over 25,
# nothing at 120 and 180); the other knobs moved nothing (2026-10-15, 6000 iterations).
# Beast Mastery refreshes the sting 4 sec before it drops, not 2: with its busier global cooldowns
# the 2 sec refresh let it lapse (rotation audit pass 20; +2.6 / +2.7 / +3.0 at 120 / 180 / 300 sec).
# Marksmanship and Lone Wolf lose 1.0 to 1.4 at 180 and 300 sec with 4, so they keep 2.
gen hunter_ranged bm_scorpid "$bm,scorpid=1,scorpidLead=4"
gen hunter_ranged mm_scorpid "$mm,scorpid=1,sniperManaPerSec=15"
gen hunter_ranged lw_scorpid "$mm,scorpid=1,sniperManaPerSec=15"
