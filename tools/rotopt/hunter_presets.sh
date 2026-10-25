#!/bin/sh
# Regenerates the hunter rotations in ui/hunter/apls from the hunter_ranged and hunter_sv
# templates, with the knobs the searches settled on (last on 2026-10-23, at 120 / 180 / 300 sec on
# the synthetic Phase 1 sets with the 8% hit cap and the Level 60 talents).
# Run from the repo root with a settings file the optimizer should sim against, made with
# mksettings for the spec. The rotation itself does not depend on the settings, only the one
# sanity sim does.
#
#   go run --tags=with_db ./tools/rotopt/mksettings -spec hunter_bm -gear level60_phase1_synthetic_ranged \
#       -apl bm -talents 5020121505021241-00520503005 -race SkyborneWindshaper -duration 180 -out bm.json
#   tools/rotopt/hunter_presets.sh bm.json
set -e
settings=$1
# Hunter's Mark (standards 23, user ruling 2026-10-25): there is no Hunter's Mark duty preset. Every
# hunter keeps up its own mark, with the raid's off in the request (mksettings' hunter specs), so
# every rotation here casts it (the templates' markPrepull, 1.5 by default). It goes up 1.5 sec before
# the pull (it doesn't pull; user ruling 2026-10-15), after the pre-pull Aspect's global cooldown, so
# the boss has it from the first shot (rotation audit pass 20: +4.7 / +3.3 / +3.4 / +1.4 at 180 sec on
# Beast Mastery / Marksmanship / Lone Wolf / Survival over the first global cooldown after the pull,
# 20000 iterations), and again in its last 3 sec. It is not held back for the hunter's own DPS (user
# ruling). Against the raid's mark it costs 0.5 to 2.5 DPS at 180 sec.
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
# The regear and rotation round (2026-10-23: no hit suppression, the 8% cap, the Level 60 talents of
# that day), knob searches at 120, 180 and 300 sec combined, 3000 iterations, confirmed at 20000.
# Beast Mastery now puts the Arcane Shot cooldown above Aimed Shot and drops the Serpent Sting and
# Arcane Shot mana floors (+4.4 combined); it keeps Serpent Sting last (serpentLate=2: +2.2 with its
# own Hunter's Mark over first). Marksmanship and Lone Wolf cast Aimed Shot only above 40% mana
# (+1.4, +1.5) and hold Sniper Shot back sooner, while the mana covers 10 a second of the fight left
# (+2.5 and +1.9 over 25; Lone Wolf at range +1.3 and +1.0 over 20 on the two seed sets). All three cast Rapid Fire 0.1 sec before
# the pull (+2 at 120 sec), and Summon Hawk while fewer than 2 hawks are out (hawk=2) only matters
# with the talent, which no build takes (the hawk Beast Mastery build is 11.8 behind at 180 sec).
bm="$ranged,arcaneFirst=1,arcaneMana=0,serpentLate=2,serpentMana=0,holdElune=1"
mmBase="$ranged,arcaneFirst=1,arcaneMana=0,serpentLate=1,serpentMana=0,aimedMana=40,holdRapid=1,holdElune=1"
mm="$mmBase,sniperManaPerSec=10"
# The melee weave (2026-10-23, after the weave mechanics of wowsims/forever): between Auto Shots the
# hunter walks from 12 yd into melee for a main hand swing Raptor Strike is ready for, when that holds
# the next Auto Shot back 1.2 sec or less and Arcane Shot and the Aimed Shot cooldown stay down until
# it is back but 0.75 sec. It keeps the player's 0.2 sec reaction before the walk in and after the
# walk out (the search's 0.1 sec is +3.3 more, but faster than a player). Beast Mastery weaves (+9.5
# combined over standing at range, 20000 iterations). Marksmanship and Lone Wolf weave only while the
# mana covers 25 a second of the fight left (weaveManaPerSec): weaving wins a short fight (+4 at 120
# sec) and costs a long one the mana (-11 at 180 sec, -15 at 300 sec without the gate). With the gate
# Marksmanship reads 1184.8 / 1167.6 / 1140.2 at 120 / 180 / 300 sec, against 1180.1 / 1168.9 / 1142.0
# at range, and stays above the 2026-10-23 rotation at all three lengths, which the turret rotation
# doesn't (-4 at 120 sec).
weave="weave=1,weaveRaptor=1,weaveDelay=1.2,weaveShotSlack=0.75"
mmWeave="$weave,weaveManaPerSec=25"
gen hunter_ranged bm "$bm,$weave"
gen hunter_ranged mm "$mm,$mmWeave"
gen hunter_ranged lw "$mm,$mmWeave"
gen hunter_ranged bm_turret "$bm"
gen hunter_ranged mm_turret "$mm"
gen hunter_ranged lw_turret "$mm"
# Marksmanship and Lone Wolf hold Rapid Fire and Elune's Light for the on-use trinket when the wait
# costs no use, Beast Mastery only Elune's Light (rotation auditor, 2026-10-13: +2.3 / +1.7 / +0.5
# at 300 sec, nothing at 120 and 180).
# Survival melees and doesn't weave (2026-10-23, 180 sec, 20000 iterations): stepping out to 12 yd for
# each Auto Shot after a main hand swing is 874.0 against its 997.2, and the ranged weave rotation on
# its build and gear 850.1.
# Survival: Wing Clip in every free global cooldown (+31 at 180 sec), held when Mongoose Bite or
# Strider Kick is about ready (+0.4).
# Since the trap scales with ranged attack power (hunter_trap_ap) it goes right after Strider Kick,
# and Wing Clip holds only for an opened Mongoose Bite, not for Strider Kick (rotation audit pass 23,
# 2026-10-16: +2.8 / +3.0 / +3.1 at 120 / 180 / 300 sec; dropping the Mongoose Bite hold too loses 0.55).
sv="lacerate=1,lacerateMana=20,strikerMana=0,trap=1,trapMana=30,wingClip=1,wingClipMana=50,wingClipHold=2,trapFirst=1"
gen hunter_sv sv "$sv"
# Scorpid Sting duty (standards 23), for the ranged hunters only (user ruling 2026-10-15, not melee
# Survival): Scorpid Sting kept up in place of Serpent Sting, for the presets with the raid's Scorpid
# Sting off (mksettings specs hunter_{bm,mm,lw}_scorpid). Without Serpent Sting Marksmanship and Lone
# Wolf have the mana for Sniper Shot at 5 and 10 a second of the fight left (2026-10-23 knob search,
# 120 / 180 / 300 sec combined, 20000 iterations: +5.5 and +6.5 over 15), and weave as their bases do.
# Beast Mastery refreshes the sting 4 sec before it drops, not 2: with its
# busier global cooldowns the 2 sec refresh lets it lapse (+3.6 over 2, +1.6 over 3). It keeps Beast
# Mastery's rotation from before 2026-10-23 (Aimed Shot first, the 30% and 20% mana floors on Serpent
# Sting and Arcane Shot): the new one is +0.3 to +0.6 at 120 sec but 0.3 to 0.4 behind at 180 and 300.
bmOld="$ranged,arcaneFirst=0,arcaneMana=20,serpentLate=0,serpentMana=30,holdElune=1"
gen hunter_ranged bm_scorpid "$bmOld,scorpid=1,scorpidLead=4,$weave"
gen hunter_ranged mm_scorpid "$mmBase,scorpid=1,sniperManaPerSec=5,$mmWeave"
gen hunter_ranged lw_scorpid "$mmBase,scorpid=1,sniperManaPerSec=10,$mmWeave"
