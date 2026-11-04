#!/bin/sh
# Regenerates the Level 60 warlock rotations in ui/warlock/apls from the warlock template
# (tools/rotopt/warlock.go). The knobs are what the knob search settled on for each build on
# 2026-10-12 and 2026-10-23 (see ui/warlock/presets.ts). Run from the repo root. The rotation does not depend
# on the settings, only the one sanity sim does, so any level 60 warlock settings will do:
#
#   go run --tags=with_db ./tools/rotopt/mksettings -spec demonology -gear level60_phase1_synthetic_demonology \
#       -apl level60_demonology -talents 051-0305003131221001351-0550004 -race Human -duration 180 -out lock60.json
#   tools/rotopt/warlock_presets.sh lock60.json
#
# The searches behind them (one talent search per starting build, then a knob search, then the
# talent search again on the searched knobs):
#
#   go run --tags=with_db ./tools/rotopt -template warlock -settings lock60.json -durs 180 -talent-search \
#       -require suppression \
#       -iters 2000 -knobs bane=2,corruption=1,immolate=1,siphonLife=1,conflagrate=1,shadowburn=2,soulFire=0,filler=2,soulLink=1,lifeTapMana=5
#   go run --tags=with_db ./tools/rotopt -template warlock -settings lock60.json -durs 180 -iters 3000 -knobs ...
set -e
settings=$1
gen() {
	name=$1
	knobs=$2
	go run --tags=with_db ./tools/rotopt -template warlock -settings "$settings" -durs 180 \
		-eval -confirm-iters 100 -out ui/warlock/apls -name "$name" -knobs "$knobs" >/dev/null
	mv "ui/warlock/apls/${name}_180s.apl.json" "ui/warlock/apls/$name.apl.json"
	echo "wrote ui/warlock/apls/$name.apl.json ($knobs)"
}
base="bane=2,corruption=1,immolate=1,lifeTapMana=5,shadowburn=0,soulFire=0,endInstants=1,berserking=1,doomAlign=1"
# Since 2026-11-02 every build holds the second Bane of Doom 1 sec past its cooldown (doomAlign=1), so
# its one tick, a minute later on the stats it lands on, comes after the trinket and Blood Fury are back
# and used, inside their window. 20000 iterations at 120 / 180 / 300 sec: Demonology +0.5 / +1.2 / +0.8,
# Destruction +0.6 / +1.6 / +1.2, Affliction +0.9 / +2.2 / +2.0, Fire Destruction +0.3 / +0.6 / +0.4;
# 2 sec is about the same, 3 and up lose (the third Doom misses the fight's end).
# Since 2026-10-28 every build holds Troll Berserking (berserking=1) for the casts its cast speed
# shortens, the filler with the GCD free, instead of the pull's bane, Corruption, Immolate and
# Conflagrate opener (standards 31). Troll, paired, 16 x 2500 iterations at 120 / 180 / 300 / 180+-20
# sec: Demonology +1.7 / +2.0 / +1.5 / +1.1, Fire Destruction +1.1 / +1.3 / +1.7 / +1.1, Affliction
# +1.8 / +1.7 / +1.0 / +0.3 (+0.6 +- 0.2 over 80000), Destruction +1.1 / +0.8 / +0.8 / -0.2 (+0.2 +- 0.2
# over 80000), the duty presets +0.2 to +3.8. The line does nothing for the other races.
# Since 2026-10-23 every build casts a new Immolate only with 15 sec of the fight left (Demonology
# and the fire build Corruption too): +0.3 to +2.2 at 120 / 180 / 300 sec on both seed sets.
# Demonology: Soul Link and a Life Tap before the pull, the damage over time spells recast 1.5 sec
# early.
gen level60_demonology "$base,filler=1,conflagrate=0,siphonLife=0,soulLink=1,prepullLifeTap=1,dotLead=1.5,prepullAt=2.5,corruptionMinTime=15,immolateMinTime=15"
# Destruction (Shadow Bolt): Conflagrate on cooldown while Immolate is up.
# Destruction and Affliction recast Immolate once it has less left than its cast time.
gen level60_destruction "$base,filler=1,conflagrate=1,siphonLife=0,soulLink=0,dotLead=0,immolateLead=1,prepullAt=2.5,immolateMinTime=15"
# Affliction (since round 70753, the 2026-11-02 notes doubled Wrack): deep Affliction with Siphon Life
# and Wrack, the Imp sacrificed for 15% Shadow damage; Wrack channelled while Corruption and the bane
# are up, Shadow Bolt between. Knob searches at 120, 180 and 300 sec from these moved nothing (300
# sec alone wanted doomAlign=6, +1).
gen level60_affliction "$base,filler=1,conflagrate=0,siphonLife=1,wrack=1,soulLink=0,dotLead=0,immolateLead=1,prepullAt=2.5,immolateMinTime=15"
# Shadow Bolt Affliction (the Affliction preset before round 70753), the Succubus out.
gen level60_affliction_shadow_bolt "$base,filler=1,conflagrate=0,siphonLife=0,soulLink=0,dotLead=0,immolateLead=1,prepullAt=2.5,immolateMinTime=15"
# Destruction fire: Incinerate while Immolate is up, Conflagrate on cooldown, Immolate and
# Corruption recast 1.5 sec early.
gen level60_destruction_fire "$base,filler=2,conflagrate=1,siphonLife=0,soulLink=0,dotLead=1.5,prepullAt=1.5,corruptionMinTime=15,immolateMinTime=4"

# The raid curse duty presets (standards 23): the same rotations with the warlock keeping Curse of
# the Elements (curse=1) or Curse of Recklessness (curse=2) up for the raid itself. A warlock holds one curse duty, so there is one
# preset per build. Under Forever the curse has its own slot, so the bane stays.
K_demonology="$base,filler=1,conflagrate=0,siphonLife=0,soulLink=1,prepullLifeTap=1,dotLead=1.5,prepullAt=2.5,corruptionMinTime=15,immolateMinTime=15"
K_destruction="$base,filler=1,conflagrate=1,siphonLife=0,soulLink=0,dotLead=0,immolateLead=1,prepullAt=2.5,immolateMinTime=15"
K_affliction="$base,filler=1,conflagrate=0,siphonLife=1,wrack=1,soulLink=0,dotLead=0,immolateLead=1,prepullAt=2.5,immolateMinTime=15"
K_destruction_fire="$base,filler=2,conflagrate=1,siphonLife=0,soulLink=0,dotLead=1.5,prepullAt=1.5,corruptionMinTime=15,immolateMinTime=4"
for spec in demonology destruction affliction destruction_fire; do
	eval dutyKnobs=\$K_$spec
	gen "level60_${spec}_duty_curse" "$dutyKnobs,curse=1"
	gen "level60_${spec}_duty_recklessness" "$dutyKnobs,curse=2"
done

