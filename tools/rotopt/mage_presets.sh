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
# last costs 6.25 times the first). Until client 70707 it fell back to a single Arcane Blast below
# 20% mana; since Precision covers Arcane, keeping all four is +1.6 at 120 / 180 / 300 sec together.
# Arcane Missiles on a Missile Barrage proc (free and twice as fast) goes ahead of the Arcane
# Blasts, so a proc during the ramp isn't rolled over (+40.5 at 180 sec, the rotation auditor's pass 23), Frostbolt as the filler (it
# beats Arcane Missiles once the missiles aren't free). No Evocation, no mana gems (the Demonic
# Rune takes their cooldown). An Arcane Blast precast lands at the pull. Arcane Power and Presence
# of Mind wait for the on-use trinket unless it won't be back in time. When a Frostbolt would not
# land before the end (its cast plus 0.75 sec), Fire Blast and Arcane Explosion instead.
gen mage_arcane level60_arcane "blastBuff=4,blastRefresh=0,barrageBeforeBlast=1,arcaneWrath=0,conserveMana=0,filler=2,missileBarrage=1,barrageFirst=0,evocationMana=0,gems=0,prepull=1,powerWithTrinket=1,pomWrath=0,finisher=1,finishPad=0.75,wand=1"
# Scorch to 5 Fire Vulnerability stacks and before they run out (5 sec), Pyroblast at 3 Heating Up
# stacks (a 1.5 sec cast), Arcane Missiles on Missile Barrage, Fire Blast above 10% mana (since
# 70425 every Fire Blast adds a Heating Up stack, two on a crit), Fireball. Evocation below 20% mana
# with 90 sec of the fight left. Combustion and the on-use trinket wait for the 5th stack, so its
# crits go on spells that take all 15%, and so do Berserking and Blood Fury (the Scorch ramp's 1.5
# sec casts are held by the GCD, so haste buys nothing there; Troll +2.5 / +2.5 / +1.4 and Orc +2.8
# / +2.2 / +1.5 at 120 / 180 / 300 sec, nothing for other races). No mana gems: they share the
# Demonic Rune's 2 min cooldown, and the rune gives more. When a Fireball would not land before the
# end (its cast plus 1.04 sec), Fire Blast and Scorch instead.
gen mage_fire level60_fire "scorchRefresh=5,pyroStacks=3,fireBlastMana=10,fireBlastHeating=0,missileBarrage=1,filler=1,evocationMana=20,evocationMinTime=90,cooldownsAfterScorch=1,trinketAfterScorch=1,racialAfterScorch=1,gems=0,blastBuff=0,arcaneWrath=0,finisher=1,finishPad=1.04,wand=1"
# Frostbolt, with Ice Lance on a Fingers of Frost charge whenever its cooldown is up (3 times the
# damage on a Frozen target since 70425, with SoD's 42.9% spell power coefficient; 8 sec cooldown
# since 70707), and three Arcane Blasts every 15 sec for 45% of Intellect as spell damage. Since
# Ice Lance's cooldown the mana goes further: three stacks and no Evocation (blastBuff 3,
# evocationMana 0) are +4.8 at 120 / 180 / 300 sec together over the 2026-10-23 rotation's two
# stacks with Evocation below 20%. A charge granted during a Frostbolt's cast is spent a spell batch
# after it completes, so that Frostbolt and the Ice Lance right after it both read Frozen; waiting
# for a Frostbolt before every Ice Lance (iceLance 3 or 4) lost 11 to 23 at 180 sec before the
# cooldown. When a Frostbolt would not land before the end (its cast plus 1 sec), Ice Lance
# instead. The build stops at Fingers of Frost, so no Ice Barrier, and the wand only when out of
# mana.
gen mage_frost level60_frost "iceLance=1,missileBarrage=1,fireBlastMana=0,evocationMana=0,evocationMinTime=90,blastBuff=3,blastRefresh=9,arcaneWrath=0,finisher=1,finishPad=1,iceBarrier=0,wand=1"
# Scorch to 5 Fire Vulnerability stacks and again with under 7 sec left on it, Pyroblast at 3
# Heating Up stacks, Ice Lance on the last Fingers of Frost charge (ahead of Scorch and Pyroblast;
# the first charge goes on a Frostfire Bolt, which the charge makes a Shatter crit), Frostfire Bolt.
# Since 70425 every Fire Blast adds a Heating Up stack, so Fire Blast goes on cooldown above 80%
# mana, and below that with the mana the fight can't use (above 70 a second of the fight left).
# Evocation below 10% with more than 75 sec of the fight left. Once Evocation is spent or no longer
# worth it (75 sec or less left, the same cutoff, so the two lines are one decision) and the mana is
# under 30 a second of the fight left, rank 2 Frostfire Bolt (285 mana for 9% less damage). The
# 2026-10-23 knob search (120, 180 and 300 sec together, after the Fingers of Frost batch timing)
# moved Ice Lance to the last charge, Evocation from 20% and 45 sec, and the rank 2 line from 60 a
# second: +15 / +44 / +39 at 120 / 180 / 300 sec on both seed sets. No mana gems, for the Demonic
# Rune. The rank 2 line read 60 sec while Evocation read 90 sec, so with 60 to 90 sec left and
# Evocation ready neither went (full-rank bolts drained the mana, the wand finished); one cutoff
# searched at 180 +- 20 sec (2026-10-28, 8 seeds x 2500): 70 to 90 tie there (+0.5 to +0.9), 60
# loses 3.1, and 75 gains most at 210 / 240 sec (+10.2 / +7.4, 0 at 120 / 150 / 300). The on-use
# trinket waits for the 5th Fire Vulnerability stack. When a Frostfire Bolt would not land before
# the end (its cast plus 1.5 sec), Ice Lance instead, and the wand when out of mana.
gen mage_frostfire level60_frostfire "scorchRefresh=7,pyroStacks=3,pyroMana=0,iceLance=2,missileBarrage=1,fireBlastMana=80,evocationMana=10,evocationMinTime=75,gems=0,lowManaRank=30,fireBlastSurplus=70,trinketAfterScorch=1,iceLanceFirst=1,blastBuff=0,arcaneWrath=0,finisher=1,finishPad=1.5,wand=1"
