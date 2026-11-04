#!/bin/sh
# Regenerates the druid Level 60 rotations in ui/balance_druid/apls and ui/feral_druid/apls from
# the druid_balance and druid_feral templates, with the knobs the searches settled on. Run from
# the repo root with a settings file per spec (tools/rotopt/mksettings -spec balance / -spec feral).
# The rotations do not depend on the settings, only the sanity sim does.
#
#   tools/rotopt/druid_presets.sh balance.json feral.json
# Balance is Night Elf since the race-tuning round (2026-10-27, standards 29 and 16), with Elune's
# Light held for the spell power trinket (elune=1): 1.1 ahead at 180 sec of leaving it to the autocast
# line (+1.0 on seed 1000003), 0.0 for a Tauren. The Wrath build is Night Elf too since 2026-11-02 (its race
# re-pick at 36 yd) and holds it the same way (+0.9 / +0.6 / +1.1 at 120 / 180 / 300 sec; the same for
# the other races, which have no Elune's Light). The Feral presets are Night Elf and keep it in.
set -e
balance=$1
feral=$2
gen() {
	template=$1
	settings=$2
	dir=$3
	name=$4
	knobs=$5
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 180 \
		-eval -confirm-iters 100 -out "$dir" -name "$name" -knobs "$knobs" >/dev/null
	mv "$dir/${name}_180s.apl.json" "$dir/$name.apl.json"
	echo "wrote $dir/$name.apl.json ($knobs)"
}
# Balance on the Level 60 talents (5232220115501351--505003 on the 2026-10-30 hotfix trees, Subtlety 3 a must-take): a Starfire cast 3 sec before the pull
# (it lands at 0), Insect Swarm and Moonfire whenever they are down, a free Starfire on
# Clearcasting, Wrath once Eclipse is down to 2 charges or fewer with no Wrath flying, Starfire
# otherwise. Innervate is left to its own trigger. The search kept that at 120, 180 and 300 sec,
# and again on the hotfix trees (2026-11-01, both Balance builds).
# cdStack (rotation auditor): the potion and Elune's Light 0.1 sec before the pull too, Elune's
# later uses held for the trinket (Balance 2.2 / 1.5 / 1.6 ahead, Wrath 1.2 / 0.8 / 1.0); for the
# cat, Elune's Light held for the Strength trinket (until the 2026-10-30 hotfix took Berserk out of
# the tree, Berserk held for the trinket and Elune's Light out of Berserk, 2.5 / 0.9 / 2.7).
# Both Balance rotations use the synthetic spell power trinket 0.1 sec before the pull (1.8 / 1.1 /
# 0.7 ahead of using it at 0 at 120 / 180 / 300 sec; 1.0 / 0.7 / 0.4 for the Wrath build).
# wrathFallback (2026-11-03 exploration): Wrath after the Starfire lines, so the druid casts Wrath
# when Starfire's mana is not there instead of waiting for it; the Clearcasting Starfire line goes
# (the Starfire line below takes the free cast anyway). +0.9 / +0.7 / +2.0 at 120 / 180 / 300 sec
# (seed 1000003 +0.9 / +0.5 / +1.7, 20,000 iterations).
# endMoonfire (rotation auditor): Moonfire once the filler would not land before the fight ends,
# 4.4 / 2.9 / 1.4 ahead for Balance (when a Starfire would not land), 5.9 / 3.8 / 2.2 for the
# Wrath build (when a Wrath plus 1.8 sec would not, its flight from 36 yd; 1.5 from 30 yd before
# 2026-11-02, +0.5 / +0.4 / +0.2 at 120 / 180 / 300 sec at 36 yd).
gen druid_balance "$balance" ui/balance_druid/apls level60 "innervateMana=0,filler=3,eclipseStacks=2,prepull=1,prepullTime=3,prepullTrinket=1,cdStack=1,endMoonfire=1,elune=1,wrathFallback=1,clearcastStarfire=0"
# The Faerie Fire duty (standards 23): the same rotation keeping Faerie Fire on the target, cast
# first after the pull and again once it has under 1.5 sec left. The knob search kept the rest.
gen druid_balance "$balance" ui/balance_druid/apls level60_ff "innervateMana=0,filler=3,eclipseStacks=2,prepull=1,prepullTime=3,prepullTrinket=1,cdStack=1,endMoonfire=1,faerieFire=1,elune=1,wrathFallback=1,clearcastStarfire=0"
# The same with Wrath as the only filler and a Wrath 1.5 sec before the pull, 6.1% behind (776.5
# against 827.0 at 180 sec), on its own talents (5502220115501351--505003) and gear (Natural Power
# gloves).
gen druid_balance "$balance" ui/balance_druid/apls level60_wrath "innervateMana=0,filler=1,prepull=2,prepullTime=1.5,prepullTrinket=1,cdStack=1,endMoonfire=2,endMoonfireSlack=1.8,elune=1"
# The Wrath build's Faerie Fire duty, the same way.
gen druid_balance "$balance" ui/balance_druid/apls level60_wrath_ff "innervateMana=0,filler=1,prepull=2,prepullTime=1.5,prepullTrinket=1,cdStack=1,endMoonfire=2,endMoonfireSlack=1.8,faerieFire=1,elune=1"
# Feral cat on the Level 60 talents of the launch realms' 2026-10-30 hotfix trees
# (05002-53200032023012125001-05502, Feral Swiftness 2 a must-take), Night Elf: Rip at 4 combo points
# whenever it is down, Ferocious Bite at 5 with 45 Energy or more while Rip has 4 sec left (Predatory
# Discipline keeps the extra Energy, which still adds damage), Shadowmeld into Ravage on Clearcasting
# or in the last 10 sec, Rake whenever it is down with more than 6 sec left, Shred, Shifting Power
# (8 sec cooldown since the hotfix) at 60 Energy or less, Bite due or not, Elune's Light held for the
# Strength trinket unless waiting would lose a use, no Tiger's Fury (Forever removed it,
# sim/druid/tigers_fury.go). No new Rip in the last 8 sec, and Ferocious Bite at any combo points in
# the last 2 sec. The knob search over 120, 180 and 300 sec together (2026-11-01) moved ripCP 5 -> 4,
# biteRipTime 6 -> 4, biteEnergy 35 -> 45, shiftEnergy 50 -> 60 and shiftHoldForBite 1 -> 0: +9.2 /
# +9.4 / +8.9 at 120 / 180 / 300 sec (20,000 iterations, seeds 1 and 1000003 within 0.4). Berserk is
# out of the tree, so cdStack holds Elune's Light for the trinket instead of out of Berserk.
# The 2026-11-03 exploration added bomb=1 (the filler explosive on its own line, at 60 Energy or less
# right after a main hand swing, instead of the autocast line throwing it on cooldown), rakeAt5=1
# (Rake at 5 combo points too) and meldAt5=1 (Shadowmeld Ravage on Clearcasting at 5 combo points
# too): +8.2 / +8.0 / +7.6 at 120 / 180 / 300 sec (seed 1000003 +7.6 / +7.7 / +7.5, 20,000
# iterations), almost all from the bomb line. With that line the Faerie Fire duty build takes the
# bomb too (+4.2 / +3.7 / +3.2, seed 1000003 +4.9 / +4.3 / +3.5).
gen druid_feral "$feral" ui/feral_druid/apls level60 "rake=1,rakeMinTime=6,shadowmeld=1,shadowmeldEnd=10,meldBeforeRake=1,ripCP=4,shiftEnergy=60,shiftHoldForBite=0,endBite=2,ripMinTime=8,biteRipTime=4,biteEnergy=45,cdStack=1,tigersFury=0,elune=1,bomb=1,bombEnergy=60,rakeAt5=1,meldAt5=1"
# The cat's Faerie Fire duty: the caster Faerie Fire from Cat Form (Forever has no Faerie Fire
# (Feral); it's free in Cat Form, 1.5 sec global cooldown, no cooldown), first after the cooldowns and
# again once it has under 7 sec left. That was the old 6 sec cooldown's second try; with no cooldown
# 2, 3, 5 and 7 sec all keep it up 99.9% of the fight and are within 1 DPS (2026-10-29, 20000
# iterations, 180 sec), so 7 stayed. After Rip, or at 3 sec, was within the noise.
gen druid_feral "$feral" ui/feral_druid/apls level60_ff "rake=1,rakeMinTime=6,shadowmeld=1,shadowmeldEnd=10,meldBeforeRake=1,ripCP=4,shiftEnergy=60,shiftHoldForBite=0,endBite=2,ripMinTime=8,biteRipTime=4,biteEnergy=45,cdStack=1,tigersFury=0,faerieFire=1,ffRefresh=7,elune=1,bomb=1,bombEnergy=60,rakeAt5=1,meldAt5=1"
