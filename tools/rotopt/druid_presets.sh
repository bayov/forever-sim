#!/bin/sh
# Regenerates the druid Level 60 rotations in ui/balance_druid/apls and ui/feral_druid/apls from
# the druid_balance and druid_feral templates, with the knobs the searches settled on. Run from
# the repo root with a settings file per spec (tools/rotopt/mksettings -spec balance / -spec feral).
# The rotations do not depend on the settings, only the sanity sim does.
#
#   tools/rotopt/druid_presets.sh balance.json feral.json
# Balance is Night Elf since the race-tuning round (2026-10-27, standards 29 and 16), with Elune's
# Light held for the spell power trinket (elune=1): 1.1 ahead at 180 sec of leaving it to the autocast
# line (+1.0 on seed 1000003), 0.0 for a Tauren. The Wrath build is Windshaper Skyborne and leaves
# it out (elune=0). The Feral presets are Night Elf and keep it in.
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
# Balance on the Level 60 talents (5232220115501351--5053, Subtlety 3 a must-take): a Starfire cast 3 sec before the pull
# (it lands at 0), Insect Swarm and Moonfire whenever they are down, a free Starfire on
# Clearcasting, Wrath once Eclipse is down to 2 charges or fewer with no Wrath flying, Starfire
# otherwise. Innervate is left to its own trigger. The search kept that at 120, 180 and 300 sec.
# cdStack (rotation auditor): the potion and Elune's Light 0.1 sec before the pull too, Elune's
# later uses held for the trinket (Balance 2.2 / 1.5 / 1.6 ahead, Wrath 1.2 / 0.8 / 1.0); for the
# cat, Berserk held for the Strength trinket and Elune's Light kept out of Berserk (2.5 / 0.9 / 2.7).
# Both Balance rotations use the synthetic spell power trinket 0.1 sec before the pull (1.8 / 1.1 /
# 0.7 ahead of using it at 0 at 120 / 180 / 300 sec; 1.0 / 0.7 / 0.4 for the Wrath build).
# endMoonfire (rotation auditor): Moonfire once the filler would not land before the fight ends,
# 4.4 / 2.9 / 1.4 ahead for Balance (when a Starfire would not land), 5.9 / 3.8 / 2.2 for the
# Wrath build (when a Wrath plus 1.5 sec would not).
gen druid_balance "$balance" ui/balance_druid/apls level60 "innervateMana=0,filler=3,eclipseStacks=2,prepull=1,prepullTime=3,prepullTrinket=1,cdStack=1,endMoonfire=1,elune=1"
# The Faerie Fire duty (standards 23): the same rotation keeping Faerie Fire on the target, cast
# first after the pull and again once it has under 1.5 sec left. The knob search kept the rest.
gen druid_balance "$balance" ui/balance_druid/apls level60_ff "innervateMana=0,filler=3,eclipseStacks=2,prepull=1,prepullTime=3,prepullTrinket=1,cdStack=1,endMoonfire=1,faerieFire=1,elune=1"
# The same with Wrath as the only filler and a Wrath 1.5 sec before the pull, 6.1% behind (776.5
# against 827.0 at 180 sec), on its own talents (5502220115501351--5053) and gear (Natural Power
# gloves).
gen druid_balance "$balance" ui/balance_druid/apls level60_wrath "innervateMana=0,filler=1,prepull=2,prepullTime=1.5,prepullTrinket=1,cdStack=1,endMoonfire=2,endMoonfireSlack=1.5,elune=0"
# The Wrath build's Faerie Fire duty, the same way.
gen druid_balance "$balance" ui/balance_druid/apls level60_wrath_ff "innervateMana=0,filler=1,prepull=2,prepullTime=1.5,prepullTrinket=1,cdStack=1,endMoonfire=2,endMoonfireSlack=1.5,faerieFire=1,elune=0"
# Feral cat on the Level 60 talents (05002-55200032021030212051-055, Feral Swiftness 2 a must-take), Night Elf: Rip at 5 combo
# points whenever it is down (2.0 to 5.6 ahead of 4 with Idol of the Dream's longer Rip), Ferocious
# Bite at 5 while Rip has 6 sec left, Shadowmeld into Ravage on Clearcasting or in the last 10 sec,
# Rake whenever it is down with more than 6 sec left (with Ferocity 5; 38 to 42 ahead of no Rake
# under the 2026-10-23 sim, Rake's attack power and Rend and Tear on its ticks), Shred, Shifting
# Power at 50 Energy or less but not while a Bite is due, Berserk held for the Strength trinket,
# Elune's Light outside Berserk, no Tiger's Fury (Forever removed it, sim/druid/tigers_fury.go). No
# new Rip in the last 8 sec, and Ferocious Bite at any combo points in the last 2 sec. The knob
# search kept the rest at 120, 180 and 300 sec.
gen druid_feral "$feral" ui/feral_druid/apls level60 "rake=1,rakeMinTime=6,shadowmeld=1,shadowmeldEnd=10,meldBeforeRake=1,ripCP=5,shiftEnergy=50,shiftHoldForBite=1,endBite=2,ripMinTime=8,biteRipTime=6,cdStack=1,tigersFury=0,elune=1"
# The cat's Faerie Fire duty: Faerie Fire (Feral) from Cat Form, first after the cooldowns and
# again once it has under 7 sec left, so a missed cast (its 6 sec cooldown) gets a second try before
# the debuff drops (rotation auditor: 99.7% uptime against 98.7% at 1 sec, same DPS). After Rip, or
# at 3 sec, is within the noise.
gen druid_feral "$feral" ui/feral_druid/apls level60_ff "rake=1,rakeMinTime=6,shadowmeld=1,shadowmeldEnd=10,meldBeforeRake=1,ripCP=5,shiftEnergy=50,shiftHoldForBite=1,endBite=2,ripMinTime=8,biteRipTime=6,cdStack=1,tigersFury=0,faerieFire=1,ffRefresh=7,elune=1"
