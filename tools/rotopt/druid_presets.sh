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
# Balance on the Level 60 talents (5532220115501351--505): Insect Swarm and Moonfire whenever
# they are down, a free Starfire on Clearcasting, then Wrath until Eclipse has a charge and
# Starfire with it. Innervate is left to its own trigger. The search kept every other knob at
# 120, 180 and 300 sec.
gen druid_balance "$balance" ui/balance_druid/apls level60 "innervateMana=0"
# The same with Wrath as the only filler, 8% behind (658 against 716 at 180 sec).
gen druid_balance "$balance" ui/balance_druid/apls level60_wrath "innervateMana=0,filler=1"
