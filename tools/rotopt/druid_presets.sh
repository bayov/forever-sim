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
# otherwise. Innervate is left to its own trigger. The search kept that at 120 and 180 sec. At
# 300 sec the mana runs out, and the old weave (Wrath until Eclipse has a charge) without the
# Clearcasting Starfire is 12 DPS ahead.
gen druid_balance "$balance" ui/balance_druid/apls level60 "innervateMana=0,filler=3,eclipseStacks=2,prepull=1,prepullTime=3"
# The same with Wrath as the only filler and a Wrath 1.5 sec before the pull, 8.5% behind (751.5
# against 821.3 at 180 sec), on its own talents (5502220115501351-03-505).
gen druid_balance "$balance" ui/balance_druid/apls level60_wrath "innervateMana=0,filler=1,prepull=2,prepullTime=1.5"
# Feral cat on the Level 60 talents (05002-05200332021031212051-05501): Rip at 4 combo points
# whenever it is down, Ferocious Bite at 5 while Rip has 4 sec left, Shred, no Rake (keeping it
# up costs 33), Shifting Power at 50 Energy or less, Tiger's Fury and Berserk on cooldown. The
# search kept that at 120 and 180 sec; at 300 sec Shifting Power at 40 Energy is 1 ahead.
gen druid_feral "$feral" ui/feral_druid/apls level60 "rake=0,ripCP=4,shiftEnergy=50"
