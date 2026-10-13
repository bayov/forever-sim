#!/bin/sh
# Regenerates the druid Level 60 rotations in ui/balance_druid/apls and ui/feral_druid/apls from
# the druid_balance and druid_feral templates, with the knobs the searches settled on. Run from
# the repo root with a settings file per spec (tools/rotopt/mksettings -spec balance / -spec feral).
# The rotations do not depend on the settings, only the sanity sim does.
#
#   tools/rotopt/druid_presets.sh balance.json feral.json
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
# Balance on the Level 60 talents (5532220115501351--505): a Starfire cast 3 sec before the pull
# (it lands at 0), Insect Swarm and Moonfire whenever they are down, a free Starfire on
# Clearcasting, Wrath once Eclipse is down to 2 charges or fewer with no Wrath flying, Starfire
# otherwise. Innervate is left to its own trigger. The search kept that at 120, 180 and 300 sec.
# cdStack (rotation auditor): the potion and Elune's Light 0.1 sec before the pull too, Elune's
# later uses held for the trinket (Balance 2.2 / 1.5 / 1.6 ahead, Wrath 1.2 / 0.8 / 1.0); for the
# cat, Berserk held for the Strength trinket and Elune's Light kept out of Berserk (2.5 / 0.9 / 2.7).
# Both Balance rotations use the synthetic spell power trinket 0.1 sec before the pull (1.8 / 1.1 /
# 0.7 ahead of using it at 0 at 120 / 180 / 300 sec; 1.0 / 0.7 / 0.4 for the Wrath build).
gen druid_balance "$balance" ui/balance_druid/apls level60 "innervateMana=0,filler=3,eclipseStacks=2,prepull=1,prepullTime=3,prepullTrinket=1,cdStack=1"
# The same with Wrath as the only filler and a Wrath 1.5 sec before the pull, 6.1% behind (776.5
# against 827.0 at 180 sec), on its own talents (5502220115501351-03-505) and gear (Natural Power
# gloves).
gen druid_balance "$balance" ui/balance_druid/apls level60_wrath "innervateMana=0,filler=1,prepull=2,prepullTime=1.5,prepullTrinket=1,cdStack=1"
# Feral cat on the Level 60 talents (05002-05200332021031212051-05501): Rip at 4 combo points
# whenever it is down, Ferocious Bite at 5 while Rip has 6 sec left, Shred, no Rake (keeping it
# up costs 33), Shifting Power at 50 Energy or less but not while a Bite is due (0.6 to 0.9
# ahead: its Energy goes into Shreds, not the Bite), Tiger's Fury and Berserk on cooldown. The
# search on the 1b gear kept that at 120 and 180 sec (at 300 sec Shifting Power at 40 Energy is
# 0.3 ahead).
# At the end of the fight: no new Rip in the last 8 sec, and Ferocious Bite at any combo points
# in the last 2 sec (0.6 to 1.4 together, the rogue's end-of-fight Eviscerate idea).
gen druid_feral "$feral" ui/feral_druid/apls level60 "rake=0,ripCP=4,shiftEnergy=50,shiftHoldForBite=1,endBite=2,ripMinTime=8,biteRipTime=6,cdStack=1"
