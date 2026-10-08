#!/bin/sh
# Regenerates the enhancement shaman presets in ui/enhancement_shaman/apls that come from
# the shaman_enh template. The level 60 one is the optimizer's best on the Phase 2 gear
# and the Level 60 talents. The level 30 one is the same template with the knobs the
# level 30 search settled on (no Maelstrom Weapon yet, Lightning Shield up because the boss
# is hitting the shaman). The level 20 presets were dropped on 2026-10-05.
# The hand written air totem presets are not touched. Run from the repo root with the
# settings file the optimizer should sim against (a RaidSimRequest or a UI settings
# export). The rotation itself does not depend on the settings, only the one sanity sim
# does.
#
#   tools/rotopt/shaman_presets.sh path/to/settings.json
set -e
settings=$1
gen() {
	name=$1
	knobs=$2
	shift 2
	go run --tags=with_db ./tools/rotopt -template shaman_enh -settings "$settings" -durs 300 \
		-eval -confirm-iters 100 -out ui/enhancement_shaman/apls -name "$name" -knobs "$knobs" "$@" >/dev/null
	mv "ui/enhancement_shaman/apls/${name}_300s.apl.json" "ui/enhancement_shaman/apls/$name.apl.json"
	echo "wrote ui/enhancement_shaman/apls/$name.apl.json ($knobs $*)"
}
# Windfury Totem for the group, Flametongue Totem for the shaman (Windfury Weapon turns
# Windfury Totem's buff off for the shaman, and leaves the weapon's totem slot to
# Flametongue Totem), Fire Nova after the shocks above 20% mana and Mana Spring Totem:
# 839.2 at 180 sec on 2026-10-07, where Searing Totem was 792.0. Fire Nova is a spell with
# its own cooldown under Forever.
#
# On 2026-10-06 we gave Fire Nova the free cast from Clearcasting (Elemental Focus): a line
# right after Stormstrike, and Fire Nova below the 20% floor too while Clearcasting is up.
# We also put Flame Shock ahead of the Maelstrom Weapon Lightning Bolt. On the synthetic set
# with potions that is +0.9 at 180 sec, +4.6 at 300 and +6.8 at 600 (900.1 at 300 sec), and
# +8 to +12 without potions. Holding the bolt or Earth Shock for the next Stormstrike mark
# lost 2 to 11 at every length, because the mark already lands on about 20 of the 22
# Stormstrikes. The Water Shield and Mana Tide builds below have no Elemental Focus, and
# Flame Shock first is within noise there, so they keep the old order.
#
# Since 2026-10-08 the builds start the fight with their totems already down (the
# Enhancement option StartingTotems, see the presets), so the rotations leave them out of
# the prepull actions (prepullTotems=0). A starting totem costs no mana. The Level 30 PvP
# rotation still puts them down before the pull.
gen optimized "shock=4,airTotem=2,fireTotem=3,fireNova=2,fireNovaMana=20,waterTotem=1,flameShockFirst=1,fireNovaClearcast=2,prepullTotems=0" -level 60
# The Water Shield and Mana Tide talent builds have no Improved Fire Nova, so Magma Totem
# beats Searing Totem there (899.7 against 879.5 at 180 sec on the Water Shield build). A
# new Magma Totem goes down with as little as 10 sec left.
gen magma "shock=4,airTotem=2,fireTotem=2,fireTotemMinTime=10,fireNova=2,fireNovaMana=20,waterTotem=1,prepullTotems=0" -level 60
# Magma Totem with Mana Tide Totem at 20% mana, for the Mana Tide talent build. Fire Nova
# goes ahead of the shocks here (fireNova=1): +4.4 at 180 sec and +1.7 at 300 on 2026-10-08.
# On the Water Shield build that was +1.7 at 180 sec and even at 300, so magma keeps it after.
gen mana_tide "shock=4,airTotem=2,fireTotem=2,fireTotemMinTime=10,fireNova=1,fireNovaMana=20,waterTotem=1,manaTideMana=20,prepullTotems=0" -level 60
# Level 30 with 5 extra talent points (26 in all), tuned for the 120 sec Level 30
# encounter. Flame Shock when its DoT is down and Earth Shock otherwise (shock=4), because
# Earth Shock spends the Stormstrike mark (+20%). Windfury Weapon from level 30 and Mana Spring Totem from 26, and shocks only
# above 10% mana so the pool lasts the fight. Flametongue Totem (level 28) in the fire
# slot, and Fire Nova after the shocks only while the shaman is above 80% mana. Fire Nova
# at any mana was 9.6 behind at 120 sec, because the shocks then run dry.
gen level30p5 "airTotem=0,strengthOfEarth=1,fireTotemMinTime=10,shield=1,shock=4,shockMana=10,waterTotem=1,maelstromStacks=0,fireTotem=3,fireNova=2,fireNovaMana=80,prepullTotems=0" -level 30 -bonus-talents 5
# Level 30 PvP, with the PvP talents and a 60 sec fight against a level 30 player (the
# "Level 30 PvP" encounter: PvP mode, 30% of the fight out of melee range at the time, 70%
# since 2026-10-06). The same knobs
# as level30p5, but Fire Nova at any mana (+6.3 at 60 sec, the pool lasts a minute). A
# Lightning Bolt hard cast while out of melee range (lightningBolt=2) was 28 behind,
# because the 3.5 sec cast runs on after the shaman is back in range and holds the swing.
# At 70% out of melee range (2026-10-07) a knob search from these knobs kept every one.
gen level30_pvp "airTotem=0,strengthOfEarth=1,fireTotemMinTime=10,shield=1,shock=4,shockMana=10,waterTotem=1,maelstromStacks=0,fireTotem=3,fireNova=2,fireNovaMana=0" -level 30 -bonus-talents 5
