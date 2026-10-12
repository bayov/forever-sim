#!/bin/sh
# Regenerates the mage's Level 60 rotations in ui/mage/apls from the mage_arcane, mage_fire,
# mage_frost and mage_frostfire templates, with the knobs the 180 sec searches settled on. Run
# from the repo root with a settings file for the spec (tools/rotopt/mksettings -spec mage_<spec>).
# The rotation itself does not depend on the settings, only the one sanity sim does.
#
#   tools/rotopt/mage_presets.sh path/to/settings.json
set -e
settings=$1
gen() {
	template=$1
	name=$2
	knobs=$3
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 180 \
		-eval -confirm-iters 100 -out ui/mage/apls -name "$name" -knobs "$knobs" >/dev/null
	mv "ui/mage/apls/${name}_180s.apl.json" "ui/mage/apls/$name.apl.json"
	echo "wrote ui/mage/apls/$name.apl.json ($knobs)"
}
# Three Arcane Blasts, then Arcane Missiles, which ends the stacks and takes their bonus. A
# Missile Barrage proc goes on the missiles as soon as the stacks are up. Below 30% mana one
# Arcane Blast, then the missiles. Four stacks and waiting for Missile Barrage both lost.
gen mage_arcane level60_arcane "blastStacks=3,missilesNeedBarrage=0,barrageFirst=1,conserveMana=30,evocationMana=20,filler=1"
# Scorch to 5 Fire Vulnerability stacks and before they run out (5 sec), Pyroblast at 3 Heating
# Up stacks (a 1.5 sec cast), Arcane Missiles on Missile Barrage, Fire Blast above 20% mana,
# Fireball. No Evocation: the mage does not run dry in 3 minutes.
gen mage_fire level60_fire "scorchRefresh=5,pyroStacks=3,fireBlastMana=20,fireBlastHeating=0,missileBarrage=1,filler=1,evocationMana=0"
# Frostbolt, and Ice Lance on every Fingers of Frost charge. Frost never runs low on mana, so
# no Evocation.
gen mage_frost level60_frost "iceLance=1,missileBarrage=1,fireBlastMana=0,evocationMana=0"
# The Fire Mage rotation with Frostfire Bolt as the filler and Ice Lance on Fingers of Frost.
# Evocation at 30% mana.
gen mage_frostfire level60_frostfire "scorchRefresh=5,pyroStacks=3,iceLance=1,missileBarrage=1,fireBlastMana=20,evocationMana=30"
