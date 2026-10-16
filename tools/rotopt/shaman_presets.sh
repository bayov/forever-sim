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
#
# The elemental rotations at the end are written the same way.
set -e
settings=$1
# Since the rotation auditor's pass 23 (2026-10-16) the Level 60 lines below write the same actions
# as the shipped Level 60 APLs in ui/*_shaman/apls (Lightning Shield before the pull and its recast
# gates, the totem recasts that stop 20 sec before the end, Searing or Magma Totem before the pull
# after the shield: totemMinTime, shield, prepullShield=2, prepullFireLead, waterTotemMinTime,
# shieldMinMana, shieldMinTime). Some shipped notes were shortened by hand and read differently. The
# templates pick each spell's top rank from the sim's tables, so the client 70425 Lava Burst and Mana
# Tide Totem top ranks (1326047, 1325256) come out right.
# Standards 23 (user ruling 2026-10-15): there is no shaman totem duty. Every shaman drops its own
# party's totems, like a warrior's Sunder: Strength of Earth, Windfury and Mana Spring for the melee
# party, Mana Spring for the caster party.
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
# From 2026-10-08 to 2026-10-16 the builds started the fight with their totems already down for
# free (the Enhancement option StartingTotems), so the rotations left them out of the prepull
# actions. The game has no such thing, so the option is gone and the totems are back in the
# prepull actions (prepullTotems=1). The apls/*.apl.json files put the shield after the totems,
# so the fire totem goes down last (Searing at -1.5 sec, Magma at -1 sec).
#
# On 2026-10-13 (the standards 1b gear round) the knob search moved the fire totem from
# Flametongue Totem to Searing Totem (with fireTotemMinTime 10): 801.8 against 795.9 at 180 sec
# with Windfury Totem kept for the group. Flametongue Totem holds the same weapon buff slot as
# Windfury Totem. Searing goes down at the pull, so the Level 60 builds have no fire starting totem.
# Since the rotation auditor's cooldown pass Rage of the Farseer waits for the on-use trinket when the
# wait costs it no use (rageHoldForTrinket, here and in magma and mana_tide): paired at 20000
# iterations +0.7 at 300 sec on Level 60 (+0.7 to +1.3 across the builds), within ±0.2 at 120 and 180.
# Fire Nova outside Clearcasting needs 12 mana a second left in the fight in place of the 20%
# floor (fireNovaManaPerSec=12, the rotation auditor, 2026-10-15): +0.4 / +2.2 / +1.7 on Level 60
# and +0.5 / +2.2 / +0.3 on Imp GW at 120 / 180 / 300 sec, paired at 20000 iterations.
# With the raid's Flametongue Totem (standards 25, the user's ruling 2026-10-16) Magma Totem beats
# Searing Totem on the Level 60 build (fireTotem=2, fireTotemMinTime=2: +28.5 / +19.0 / +12.7 at
# 120 / 180 / 300 sec; Trial by Fire's 20 sec on Magma Totem since client 70425 helps), and Earth Shock
# waits for the Stormstrike mark (earthShockHold=1, +0.3 to +1.5 more), paired at 20000 iterations on two
# seed sets. A knob search that could drop Windfury Totem dropped it; the builds keep it for the party.
# Lightning Bolt at 5 Maelstrom Weapon stacks after the shocks (earthShockBeforeBolt=1, rotation pass
# 2026-10-16): +0.2 to +1.2 on Level 60 and +0.1 to +1.0 on Imp GW at every length, two seed sets.
gen optimized "shock=4,airTotem=2,fireTotem=2,fireTotemMinTime=2,earthShockHold=1,fireNova=2,fireNovaMana=20,waterTotem=1,flameShockFirst=1,fireNovaClearcast=2,prepullTotems=1,rageHoldForTrinket=1,fireNovaManaPerSec=12,totemMinTime=20,shield=1,prepullShield=2,prepullFireLead=1.5,earthShockBeforeBolt=1" -level 60
# The Water Shield and Mana Tide talent builds have no Improved Fire Nova, so Magma Totem
# beats Searing Totem there (899.7 against 879.5 at 180 sec on the Water Shield build). A
# new Magma Totem goes down with as little as 10 sec left.
# Since the rotation auditor's pass 23 (2026-10-16) a new Magma Totem goes down with as little as 2 sec
# left (fireTotemMinTime=2): +1.17 / +0.88 / +0.55 at 120 / 180 / 300 sec on Water Shield and +0.63
# at 180 on Mana Tide, paired at 20000 iterations on two seed sets.
# Flame Shock ahead of the bolt (flameShockFirst=1) on the Water Shield build since the raid's
# Flametongue Totem: +3.0 to +3.5 / +2.3 to +2.9 / +1.6 to +2.3 on two seed sets.
# Since 2026-10-16 Flame Shock goes ahead of Stormstrike and the bolt after Earth Shock (flameShockFirst=2,
# earthShockBeforeBolt=1): +1.3 to +1.9 on Water Shield at every length, two seed sets.
gen magma "shock=4,airTotem=2,fireTotem=2,fireTotemMinTime=2,fireNova=1,fireNovaMana=20,waterTotem=1,earthShockFlameShockLeft=1,flameShockFirst=2,earthShockBeforeBolt=1,prepullTotems=1,rageHoldForTrinket=1,totemMinTime=20,shield=1,prepullShield=2,prepullFireLead=1" -level 60
# On 2026-10-12, on the searched Water Shield build (050031-053030031005112151-052030001), Fire
# Nova ahead of the shocks is +2.7 here too, and holding Earth Shock while Flame Shock has 1 sec
# or less left (so the shock cooldown is back when the DoT runs out) +1.0: 762.3 as the mean of
# 120, 180 and 300 sec, from 758.6. The hold is +1.1 on the Mana Tide build. On the Level 60
# build it does nothing, and longer holds lose (4 sec is 8.5 behind), because Earth Shock spends
# the Stormstrike mark.
# Both builds pay for Water Shield, so since 2026-10-12 they put it up before the pull (1.5 sec before the first totem) and
# again whenever it is down below 50% mana (the rotation auditor's finding). It only does
# anything when something hits the shaman: with the Enhancement option of 6 raid damage hits a
# minute it is +0.0 / +0.5 / +4.1 on the Water Shield build at 120 / 180 / 300 sec, and a wash
# on the Mana Tide build. Recast at any mana it is 1.4 behind at 120 sec. These two builds start
# with Magma Totem 1 sec before the pull: mksettings -spec enh writes Flametongue Totem, so set
# the starting fire totem to Magma Totem in the settings file to reproduce their numbers.
# Magma Totem with Mana Tide Totem at 20% mana, for the Mana Tide talent build. Fire Nova
# goes ahead of the shocks here (fireNova=1): +4.4 at 180 sec and +1.7 at 300 on 2026-10-08.
# On the Water Shield build that was +1.7 at 180 sec and even at 300, so magma keeps it after.
# Fire Nova after the shocks (fireNova=2) on the Mana Tide build since the raid's Flametongue Totem:
# +0.7 to +2.5 / +1.1 to +2.2 / +1.1 to +1.8 on two seed sets.
gen mana_tide "shock=4,airTotem=2,fireTotem=2,fireTotemMinTime=2,fireNova=2,fireNovaMana=20,waterTotem=1,manaTideMana=20,earthShockFlameShockLeft=1,prepullTotems=1,rageHoldForTrinket=1,totemMinTime=20,shield=1,prepullShield=2,prepullFireLead=1" -level 60
# Level 30 with 5 extra talent points (26 in all), tuned for the 120 sec Level 30
# encounter. Flame Shock when its DoT is down and Earth Shock otherwise (shock=4), because
# Earth Shock spends the Stormstrike mark (+20%). Windfury Weapon from level 30 and Mana Spring Totem from 26, and shocks only
# above 10% mana so the pool lasts the fight. Flametongue Totem (level 28) in the fire
# slot, and Fire Nova after the shocks only while the shaman is above 80% mana. Fire Nova
# at any mana was 9.6 behind at 120 sec, because the shocks then run dry.
gen level30p5 "airTotem=0,strengthOfEarth=1,fireTotemMinTime=10,shield=1,shock=4,shockMana=10,waterTotem=1,maelstromStacks=0,fireTotem=3,fireNova=2,fireNovaMana=80,prepullTotems=1" -level 30 -bonus-talents 5
# Level 30 PvP, with the PvP talents and a 60 sec fight against a level 30 player (the
# "Level 30 PvP" encounter: PvP mode, 30% of the fight out of melee range at the time, 70%
# since 2026-10-06). The same knobs
# as level30p5, but Fire Nova at any mana (+6.3 at 60 sec, the pool lasts a minute). A
# Lightning Bolt hard cast while out of melee range (lightningBolt=2) was 28 behind,
# because the 3.5 sec cast runs on after the shaman is back in range and holds the swing.
# At 70% out of melee range (2026-10-07) a knob search from these knobs kept every one.
gen level30_pvp "airTotem=0,strengthOfEarth=1,fireTotemMinTime=10,shield=1,shock=4,shockMana=10,waterTotem=1,maelstromStacks=0,fireTotem=3,fireNova=2,fireNovaMana=0" -level 30 -bonus-talents 5

# The elemental presets in ui/elemental_shaman/apls, from the shaman_ele template. Mana Spring and
# Searing Totem, Flame Shock when its DoT is down, Lava Burst on cooldown unless Flame Shock runs
# out before the cast ends (lavaBurstFlameShockLeft=2), Lightning Bolt, rank 8 below 20% mana, and
# no Chain Lightning (it loses on one target even above 80% mana). The fight opens with Mana
# Spring at -3.9 sec, a Lightning Bolt at -2.9 (it lands 0.1 sec into the fight) and Searing Totem
# at -0.9 (prepullBolt=1). Mana Tide Totem goes down at 20% mana on the Mana Tide talents. An Earth
# Shock ends the fight once a Lightning Bolt (cast plus about 1 sec of flight) would not land in time
# (earthShockFinish=1, the rotation auditor, 2026-10-15): +5.8 / +3.4 / +1.7 at 120 / 180 / 300 sec
# on Level 60 and +5.7 / +3.4 / +1.8 on Mana Tide, paired at 20000 iterations.
gen_ele() {
	name=$1
	knobs=$2
	go run --tags=with_db ./tools/rotopt -template shaman_ele -settings "$settings" -durs 300 \
		-eval -confirm-iters 100 -out ui/elemental_shaman/apls -name "$name" -knobs "$knobs" -level 60 >/dev/null
	mv "ui/elemental_shaman/apls/${name}_300s.apl.json" "ui/elemental_shaman/apls/$name.apl.json"
	echo "wrote ui/elemental_shaman/apls/$name.apl.json ($knobs)"
}
# Lightning Shield at -5.4 sec and again in the fight with half the mana and a minute left
# (shield=1, shieldMinMana=50, shieldMinTime=60, prepullShield=1), and Mana Spring not recast in the
# last 20 sec (waterTotemMinTime=20). Since the rotation auditor's pass 23 (2026-10-16) no new Flame
# Shock with less than 9 sec left (flameShockMinTime=9), and Mana Tide Totem at 10% mana: +2.43 /
# +1.06 / +0.30 at 120 / 180 / 300 sec on Mana Tide, and the Flame Shock gate +2.50 / +0.70 / +0.71
# on Level 60, paired at 20000 iterations on two seed sets.
ele60="chainLightning=0,lightningBoltDownrank=2,lavaBurstFlameShockLeft=2,prepullBolt=1,earthShockFinish=1,waterTotemMinTime=20,shield=1,shieldMinMana=50,shieldMinTime=60,prepullShield=1,flameShockMinTime=9"
gen_ele level60 "$ele60"
gen_ele level60_mana_tide "$ele60,manaTideMana=10"
