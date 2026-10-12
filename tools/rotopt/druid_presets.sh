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
# they are down, a free Starfire on Clearcasting, Wrath once Eclipse is down to 2 charges or
# fewer with no Wrath flying, Starfire otherwise. Innervate is left to its own trigger. The
# search kept that at 120 and 180 sec. At 300 sec, with Major Spellblasting Potion in place of
# a mana potion, Innervate at 40% and the old weave (Wrath until Eclipse has a charge) are 6
# DPS ahead, because the mana runs out.
gen druid_balance "$balance" ui/balance_druid/apls level60 "innervateMana=0,filler=3,eclipseStacks=2"
# The same with Wrath as the only filler, 7% behind (692 against 744 at 180 sec).
gen druid_balance "$balance" ui/balance_druid/apls level60_wrath "innervateMana=0,filler=1"
# Feral cat on the Level 60 talents (05002-05200332021032212041-05501): Rip at 4 combo points
# whenever it is down, Ferocious Bite at 5 while Rip has 4 sec left, Shred, no Rake (40 Energy
# for 160 damage, a Shred does 740), Shifting Power at 50 Energy or less, Tiger's Fury and
# Berserk on cooldown. The search kept that at 120 and 180 sec, and Shifting Power at 40
# Energy at 300 sec (2 DPS, the mana runs out sooner).
gen druid_feral "$feral" ui/feral_druid/apls level60 "rake=0,ripCP=4,shiftEnergy=50"
