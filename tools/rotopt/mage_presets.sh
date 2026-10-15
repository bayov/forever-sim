#!/bin/sh
# Regenerates the mage's Level 60 rotations in ui/mage/apls from the mage_arcane, mage_fire,
# mage_frost and mage_frostfire templates, with the knobs the searches over 120, 180 and 300 sec together settled on. Run
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
# Arcane Blast to 2 stacks, then Arcane Missiles on a Missile Barrage proc. Without one, Arcane
# Blast goes on to 4 stacks and the missiles go anyway. The missiles end the stacks and take
# their bonus. Below 20% mana one Arcane Blast, then the missiles. No Evocation. An Arcane Blast
# precast lands at the pull, and the Demonic Rune takes the cooldown the mana gems share. Arcane
# Power and Presence of Mind wait for the on-use trinket unless it won't be back in time.
gen mage_arcane level60_arcane "blastStacks=2,missilesNeedBarrage=1,barrageFirst=1,conserveMana=20,evocationMana=0,filler=1,gems=0,prepull=1,powerWithTrinket=1"
# Scorch to 5 Fire Vulnerability stacks and before they run out (5 sec), Pyroblast at 3 Heating
# Up stacks (a 1.5 sec cast), Arcane Missiles on Missile Barrage, Fire Blast above 20% mana,
# Fireball. Evocation below 20% mana with 90 sec of the fight left. Combustion and the on-use
# trinket wait for the 5th stack, so its crits go on spells that take all 15%. No mana gems:
# they share the Demonic Rune's 2 min cooldown, and the rune gives more.
gen mage_fire level60_fire "scorchRefresh=5,pyroStacks=3,fireBlastMana=20,fireBlastHeating=0,missileBarrage=1,filler=1,evocationMana=20,evocationMinTime=90,cooldownsAfterScorch=1,trinketAfterScorch=1,gems=0"
# Frostbolt, with Ice Lance on every Fingers of Frost charge (4 times the damage on a Frozen
# target, with SoD's 42.9% spell power coefficient: +175 at 180 sec over none, +157 over only the
# last charge). The 2026-10-15 search at 180 sec with Ice Lance kept every other knob. Frost rarely
# runs low on mana, so Evocation (below 20% with 90 sec left) almost never goes.
gen mage_frost level60_frost "iceLance=1,missileBarrage=1,fireBlastMana=0,evocationMana=20,evocationMinTime=90"
# Scorch to 5 Fire Vulnerability stacks, Pyroblast at 3 Heating Up stacks, Ice Lance on every
# Fingers of Frost charge (with SoD's 42.9% coefficient, +66 at 180 sec over none, +7 over only the
# last charge), Frostfire Bolt. Fire Blast on cooldown loses: its mana is worth more as Frostfire
# Bolts. Ice Lance's mana moves Evocation up to below 30% with 60 sec of the fight left (the
# 2026-10-15 search at 180 sec: +6.2 there over 10% and 30 sec, level at 120, -3.9 at 300). Once
# Evocation is spent (or under 60 sec are left) and the mana is under 60 a second of the fight
# left, rank 2 Frostfire Bolt (285 mana for 9% less damage). No mana gems, for the Demonic Rune.
# Fire Blast only with the mana the fight cannot use (above 150 a second of the fight left). The
# on-use trinket waits for the 5th Fire Vulnerability stack.
gen mage_frostfire level60_frostfire "scorchRefresh=5,pyroStacks=3,pyroMana=0,iceLance=1,missileBarrage=1,fireBlastMana=0,evocationMana=30,evocationMinTime=60,gems=0,lowManaRank=60,fireBlastSurplus=150,trinketAfterScorch=1"
