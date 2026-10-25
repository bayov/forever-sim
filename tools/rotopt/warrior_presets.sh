#!/bin/sh
# Regenerates the Level 60 warrior rotations in ui/warrior/apls from the warrior_fury and
# warrior_arms templates, with the knobs the searches settled on (2026-10-13 with Forever's normalized Rage at s = 0.5, 180 sec,
# re-searched and kept on 2026-10-23 with the 8% hit sets, the natural flasks and the Charge opener,
# and re-searched on 2026-10-25 with flat Rage (s = 0, the user's ruling; standards 21): Rage-starved,
# Fury queues Heroic Strike from 80 Rage (55 before) and in the execute phase too, Bloodrages under
# 20, holds Death Wish for the last 33 sec and the trinket's last use for the last 30; two-hand Fury
# takes Berserker Rage under 30 for its Rage (Improved Berserker Rage 2) and Slams from 45; Arms
# Bloodrages under 20 and Hamstrings from 30;
# the synthetic Phase 1 sets and the Level 60 talents, see ui/warrior/presets.ts). The settings
# come from tools/rotopt/mksettings (spec fury or arms). The rotation itself does not depend on
# them, only the one sanity sim does.
#
#   go run --tags=with_db ./tools/rotopt/mksettings -spec fury -gear level60_phase1_synthetic_fury \
#       -apl level60_fury -talents 30305013002-25053005142010501 -race Human -duration 180 -out fury.json
#   tools/rotopt/warrior_presets.sh fury.json
#
# mineSync=0 since 2026-10-24: the presets take no Goblin Land Mine (standards 11, explosives only
# when they gain), so the Sapper goes with the other cooldowns (within 0.6 of mineSync=1 without the
# Land Mine at 180 sec on every build).
#
# The searches: tools/rotopt -talent-search and the knob search, one after the other until
# neither moved, from these talents and knobs.
set -e
settings=$1
gen() {
	template=$1
	name=$2
	talents=$3
	knobs=$4
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 180 -combined \
		-eval -confirm-iters 100 -out ui/warrior/apls -name "$name" -talents "$talents" -knobs "$knobs" >/dev/null
	echo "wrote ui/warrior/apls/$name.apl.json ($knobs)"
}
gen warrior_fury level60_fury 30305013002-25053005142010501 \
	battleShout=0,berserkerRageRage=0,bloodrageRage=20,deathWish=2,dwEnd=33,executeFirst=0,fillerGap=0,hamstringRage=10,hsExecute=1,hsRage=80,mineSync=0,potionAt=20,potionHoldForDw=2,potionPull=1,racialAt=25,racialHoldForDw=1,recklessAt=25,recklessness=3,slamAfterSwing=0,slamRage=0,spearDance=0,trinketAt=30,wwBtGap=0,wwFirst=1,wwRage=25
# level60_fury on Battle Shout duty (standards 23): shouts before the pull and refreshes it.
gen warrior_fury level60_fury_battle_shout 30305013002-25053005142010501 \
	battleShout=1,berserkerRageRage=0,bloodrageRage=20,deathWish=2,dwEnd=33,executeFirst=0,fillerGap=0,hamstringRage=10,hsExecute=1,hsRage=80,mineSync=0,potionAt=20,potionHoldForDw=2,potionPull=1,racialAt=25,racialHoldForDw=1,recklessAt=25,recklessness=3,slamAfterSwing=0,slamRage=0,spearDance=0,trinketAt=30,wwBtGap=0,wwFirst=1,wwRage=25
gen warrior_fury level60_fury_2h 30305013032-25053000142210501 \
	battleShout=0,berserkerRageRage=30,bloodrageRage=20,deathWish=2,dwEnd=36,executeFirst=0,fillerGap=0,hamstringRage=20,hsExecute=0,hsRage=90,mineSync=0,potionAt=20,potionHoldForDw=2,potionPull=0,racialAt=30,racialHoldForDw=2,recklessAt=30,recklessness=2,slamAfterSwing=1,slamRage=45,spearDance=0,trinketAt=34,wwBtGap=0,wwFirst=0,wwRage=25
# level60_fury_2h on Battle Shout duty (standards 23): shouts before the pull and refreshes it. The duty
# search takes Bloodrage below 20 Rage (+1.6 at 180 sec), since the pull's Bloodrage goes before the shout.
gen warrior_fury level60_fury_2h_battle_shout 30305013032-25053000142210501 \
	battleShout=1,berserkerRageRage=30,bloodrageRage=20,deathWish=2,dwEnd=36,executeFirst=0,fillerGap=0,hamstringRage=20,hsExecute=0,hsRage=90,mineSync=0,potionAt=20,potionHoldForDw=2,potionPull=0,racialAt=30,racialHoldForDw=2,recklessAt=30,recklessness=2,slamAfterSwing=1,slamRage=45,spearDance=0,trinketAt=34,wwBtGap=0,wwFirst=0,wwRage=25
gen warrior_arms level60_arms 20305213032515201-55050000001 \
	battleShout=0,berserkerRageRage=0,bloodrageRage=20,deathWish=2,dwEnd=36,executeFirst=0,hamstringRage=30,hsExecute=0,hsRage=90,mineSync=0,msFirst=1,overpowerRage=5,potionAt=15,potionHoldForDw=2,potionPull=0,racialAt=25,racialHoldForDw=0,reckSwap=2,recklessAt=25,recklessness=2,rend=1,slamAfterSwing=0,slamRage=20,spearingRage=0,stance=1,trinketAt=0,wwRage=25
# level60_arms on Battle Shout duty (standards 23): shouts before the pull and refreshes it.
gen warrior_arms level60_arms_battle_shout 20305213032515201-55050000001 \
	battleShout=1,berserkerRageRage=0,bloodrageRage=20,deathWish=2,dwEnd=36,executeFirst=0,hamstringRage=30,hsExecute=0,hsRage=90,mineSync=0,msFirst=1,overpowerRage=5,potionAt=15,potionHoldForDw=2,potionPull=0,racialAt=25,racialHoldForDw=0,reckSwap=2,recklessAt=25,recklessness=2,rend=1,slamAfterSwing=0,slamRage=20,spearingRage=0,stance=1,trinketAt=0,wwRage=25
# Arms against a Giant or Dragonkin boss (Onyxia), where Spearing Strike does 80% more weapon damage:
# the talent and knob searches on a Dragonkin boss take Spearing Strike and Improved Bloodrage for two
# points of Booming Voice and one of Improved Execute.
gen warrior_arms level60_arms_dragonkin 20305213132515201-3505-2 \
	battleShout=0,berserkerRageRage=0,bloodrageRage=20,deathWish=2,dwEnd=36,executeFirst=0,hamstringRage=30,hsExecute=0,hsRage=90,mineSync=0,msFirst=1,overpowerRage=5,potionAt=15,potionHoldForDw=2,potionPull=0,racialAt=25,racialHoldForDw=0,reckSwap=2,recklessAt=25,recklessness=2,rend=1,slamAfterSwing=0,slamRage=20,spearingRage=15,stance=1,trinketAt=0,wwRage=25
