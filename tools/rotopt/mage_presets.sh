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
# Client 1.60.1.70425. Arcane Blast's stacks are 15% of Intellect as spell damage each for 15 sec
# from the first, so Arcane keeps all 4 up: four Arcane Blasts at the start of each 15 sec (the
# last costs 6.25 times the first), and once the mana is under 20% a single Arcane Blast instead.
# Arcane Missiles on a Missile Barrage proc (free and twice as fast) goes ahead of the Arcane
# Blasts, so a proc during the ramp isn't rolled over (+40.5 at 180 sec, the rotation auditor's pass 23), Frostbolt as the filler (it
# beats Arcane Missiles once the missiles aren't free). No Evocation, no mana gems (the Demonic
# Rune takes their cooldown). An Arcane Blast precast lands at the pull. Arcane Power and Presence
# of Mind wait for the on-use trinket unless it won't be back in time. When a Frostbolt would not
# land before the end (its cast plus 0.75 sec), Fire Blast and Arcane Explosion instead.
gen mage_arcane level60_arcane "blastBuff=4,blastRefresh=0,barrageBeforeBlast=1,arcaneWrath=0,conserveMana=20,filler=2,missileBarrage=1,barrageFirst=0,evocationMana=0,gems=0,prepull=1,powerWithTrinket=1,pomWrath=0,finisher=1,finishPad=0.75"
# Scorch to 5 Fire Vulnerability stacks and before they run out (5 sec), Pyroblast at 3 Heating
# Up stacks (a 1.5 sec cast), Arcane Missiles on Missile Barrage, Fire Blast above 10% mana (since
# 70425 every Fire Blast adds a Heating Up stack, two on a crit), Fireball. Evocation below 20%
# mana with 90 sec of the fight left. Combustion and the on-use trinket wait for the 5th stack, so
# its crits go on spells that take all 15%. No mana gems: they share the Demonic Rune's 2 min
# cooldown, and the rune gives more. When a Fireball would not land before the end (its cast plus
# 1.04 sec), Fire Blast and Scorch instead.
gen mage_fire level60_fire "scorchRefresh=5,pyroStacks=3,fireBlastMana=10,fireBlastHeating=0,missileBarrage=1,filler=1,evocationMana=20,evocationMinTime=90,cooldownsAfterScorch=1,trinketAfterScorch=1,gems=0,blastBuff=0,arcaneWrath=0,finisher=1,finishPad=1.04"
# Frostbolt, with Ice Lance on every Fingers of Frost charge (3 times the damage on a Frozen
# target since 70425, with SoD's 42.9% spell power coefficient), and one Arcane Blast every 15 sec
# for 15% of Intellect as spell damage. Frost rarely runs low on mana, so Evocation (below 20%
# with 90 sec left) almost never goes. When a Frostbolt would not land before the end (its cast
# plus 1 sec), Ice Lance instead.
gen mage_frost level60_frost "iceLance=1,missileBarrage=1,fireBlastMana=0,evocationMana=20,evocationMinTime=90,blastBuff=1,blastRefresh=9,arcaneWrath=0,finisher=1,finishPad=1"
# Scorch to 5 Fire Vulnerability stacks and again with under 7 sec left on it, Pyroblast at 3
# Heating Up stacks, Ice Lance on every Fingers of Frost charge (ahead of Scorch and Pyroblast),
# Frostfire Bolt. Since 70425 every Fire Blast adds a Heating Up stack, so Fire Blast goes on
# cooldown above 80% mana, and below that with the mana the fight can't use (above 70 a second of
# the fight left). Evocation below 20% with 45 sec of the fight left (the last three from the
# rotation auditor's pass 23). Once Evocation is spent (or
# under 60 sec are left) and the mana is under 60 a second of the fight left, rank 2 Frostfire
# Bolt (285 mana for 9% less damage). No mana gems, for the Demonic Rune. The on-use trinket waits
# for the 5th Fire Vulnerability stack. When a Frostfire Bolt would not land before the end (its
# cast plus 1.5 sec), Ice Lance instead.
gen mage_frostfire level60_frostfire "scorchRefresh=7,pyroStacks=3,pyroMana=0,iceLance=1,missileBarrage=1,fireBlastMana=80,evocationMana=20,evocationMinTime=45,gems=0,lowManaRank=60,fireBlastSurplus=70,trinketAfterScorch=1,iceLanceFirst=1,blastBuff=0,arcaneWrath=0,finisher=1,finishPad=1.5"
