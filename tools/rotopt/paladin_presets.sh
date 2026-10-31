#!/bin/sh
# Regenerates the retribution paladin presets in ui/retribution_paladin/apls that come
# from the paladin_ret template. The level 30 one is the knobs the level 20 and 30 searches
# settled on at 60 sec: Judgement of the Crusader up first, then Consecration ahead of the seal
# and judgement lines down to the last point of mana, Seal of Command and Holy Strike on
# cooldown. Over 120 sec a 20% mana floor for Consecration is 1.0 better on one target,
# because it keeps the seal and judgement going.
#
# Run from the repo root with the settings file the optimizer should sim against (a
# RaidSimRequest or a UI settings export). The rotation itself does not depend on the
# settings, only the one sanity sim does.
#
#   tools/rotopt/paladin_presets.sh path/to/settings.json
set -e
settings=$1
gen() {
	name=$1
	knobs=$2
	shift 2
	go run --tags=with_db ./tools/rotopt -template paladin_ret -settings "$settings" -durs 60 \
		-eval -confirm-iters 100 -out ui/retribution_paladin/apls -name "$name" -knobs "$knobs" "$@" >/dev/null
	mv "ui/retribution_paladin/apls/${name}_60s.apl.json" "ui/retribution_paladin/apls/$name.apl.json"
	echo "wrote ui/retribution_paladin/apls/$name.apl.json ($knobs $*)"
}
# From the cooldown-stacking audit (2026-10-13): all four Level 60 rotations use the on-use
# trinket at the pull and then hold it for the 20% execute phase whenever that loses no use
# (trinketExecute): +1.3 / +1.3 / +1.1 at 180 sec on twist / sub Holy / Shock, +0.7 to 0.9 vs
# Undead, 0 at 120 sec and up to 0.2 at 300.
# Level 60 on the synthetic Phase 1 set (2026-10-12, the search over 120, 180 and 300 sec
# together on the Level 60 talents). Judgement of the Crusader comes from the raid, our own
# Seal of the Crusader opener is 11 behind. Consecration first down to 20% mana (+1.5 over
# after the seal lines, a 60% floor is 8 behind), Seal of Command up with Twist of Light
# twisting to Seal of Righteousness and back in the last 1.5 sec before each swing, down to
# 20% mana (126 ahead of no twisting, a 0.5 sec window is 35 behind), Holy Strike on
# cooldown (99), Hammer of Wrath below 20% health at any mana (29). Seal of Righteousness
# as the primary seal is 1.5 behind, since both seals proc every swing either way.
# From the rotation audit (2026-10-12): Seal of Righteousness at -3 sec and Seal of Command at
# -1.5, so the first swing gets the Echo too, and Judgement on Seal of Command, on Seal of
# Righteousness only when the next swing is more than 2.5 sec off (Judgement of Command hits
# harder): +5.8 at 180 sec on the new consumes. The search keeps them.
# From the rotation audit (2026-10-30), once extra attacks kept the swing timer: twisting whenever
# no Echo is waiting (twistWindow=0, no swing gate) beats the 1.5 sec window by +13.9 at 180 sec
# (+11.9 at 120, +4.8 at 300) on level60 and +14.9 on sub Holy. Before that change the same
# variant lost 4.0, and 1.5 sec was the sharp optimum.
# From the rotation audit (2026-10-13): the spell power on-use trinket and the Goblin Sapper
# Charge at the top of the list instead of on the autocast line, which a GCD line above it
# nearly always beat, so they went out 3 to 4.5 sec late and drifted from there.
# Regear and rotation round (2026-10-23, knob search over 120, 180 and 300 sec together, then
# paired at 20000 iterations on seeds 1 and 1000003): twist at any mana (twistMana=0, was 20):
# +0.0 / +0.1 / +5.7 at 120 / 180 / 300 sec (the 20% floor never binds at 120), the Crusader duty
# copy +0.0 / +0.1 / +6.2.
gen level60 "seal=2,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistMana=0,twistWindow=0,hammerOfWrathMana=0,prepullTwo=1,judgeRighteousnessWindow=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1" -level 60
# Level 60 Shock, for the Holy Shock talents (no Twist of Light, so no twisting): Consecration
# first down to the last point of mana (2 ahead of a 20% floor), Judgement on cooldown on Seal of
# Righteousness, Holy Strike on cooldown (90), Light's Vigil above 40% mana just before Holy
# Shock is ready (64), so the Holy Shock after it detonates it and starts no cooldown, Holy Shock
# after Holy Strike (148, 6.7 ahead of before it), Hammer of Wrath at any mana (23).
# From the rotation audit (2026-10-12): Light's Vigil before the pull, and then whenever it is
# off the target above 30% mana with Holy Shock waiting for it (it used to wait for Holy
# Shock's cooldown), so nearly every Holy Shock detonates a Vigil and starts no cooldown:
# +22.8 at 180 sec. The floor has to stay over the Vigil's 1340 mana.
# From the rotation audit (2026-10-13): the spell power on-use trinket and the Goblin Sapper
# Charge at the top of the list instead of on the autocast line, which a GCD line above it
# nearly always beat, so they went out 3 to 4.5 sec late and drifted from there.
# Regear and rotation round (2026-10-23): Light's Vigil only when the next swing is more than 2 sec
# off (vigilSwing=20). Its 1.5 sec cast resets the swing when it completes (216d430cbf), so a Vigil
# started just before a swing cost a white hit: +45.6 / +35.4 / +26.3 at 120 / 180 / 300 sec (1.5
# sec is +47.6 / +34.4 / +24.3, 2.5 sec +39.8 / +32.8 / +25.3; a 10, 50 or 70% Vigil floor with it
# is 8 to 14 behind 30% at 180 sec). Shock vs Undead +52.8 / +40.4 / +17.5 and the Shock duty
# +37.0 / +32.2 / +24.3 with it (paired, seeds 1 and 1000003).
# User ruling 2026-10-27: Consecration after the seal, Judgement and Holy Strike lines instead of
# first (consecrationFirst=0), Light's Vigil only when the next swing is more than 2.5 sec off
# (vigilSwing=25), and in fights longer than 210 sec a 60% Vigil floor (vigilLongFight=210; with it
# Holy Shock also goes out whenever the Vigil is held, the swing check included). rotopt -eval, 20000
# iterations, Human: 875.8 / 863.6 / 838.2 / 819.2 -> 894.1 / 881.2 / 840.2 / 827.2 at 120 / 180 / 300
# / 480 sec; Undead 879.5 / 866.6 / 840.1 / 820.7 -> 897.0 / 883.2 / 841.9 / 828.9. The same three
# changes lose on Shock vs Undead (-4.7 / -3.7 / -8.4 / -3.0 at Undead) and on the Shock duty at 300
# sec (+6.4 / +7.4 / -4.7 / +5.1), so those keep their knobs.
# Standards 31 (2026-10-28): the 60% Vigil floor in fights longer than 210 sec (vigilLongFight) is gone
# from Shock and its Crusader duty. It keyed on the fight's length, not on the mana, and it only lost:
# a flat 30% is +0.0 at 120 / 180 / 180±20 sec and +12.6 / +8.4 / +11.5 at 220 / 240 / 300 sec (duty
# +11.1 / +3.9 / +11.9), paired, 8 seeds x 5000. A mana-per-second rule in its place (vigilManaPerSec)
# never helps: 10 or 15 mana a second never binds, 20 or more loses up to 13 at 220 to 300 sec. Holy
# Shock still goes out whenever the Vigil is held, the swing check included.
gen level60_shock "seal=1,jotc=0,consecration=1,consecrationFirst=0,consecrationMana=0,holyShock=1,lightsVigilMana=30,hammerOfWrathMana=0,prepullTwo=1,vigilChain=1,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,vigilSwing=25" -level 60
# Level 60 Shock (called Shock (Ret side) on 2026-10-26; the deep Holy build is Level 60 Shock (deep Holy)
# since 2026-10-27), for the mostly Retribution Holy Shock talents (23/0/28: Holy Shock and
# Consecrated Ground, Champion of the Light, no Light's Vigil and no Twist of Light). Searched over
# 120, 180 and 300 sec together, then a talent search, then the knobs again (no move): Seal of
# Command up with no twisting (+6 over Seal of Righteousness), Consecration first down to the last
# point of mana, Judgement, Holy Strike and Holy Shock on cooldown, Hammer of Wrath above Holy
# Strike (hammerOfWrathFirst=3). Only Seal of Command before the pull: with no Twist of Light a
# Seal of Righteousness first (prepullTwo) is twisting for nothing, and the same DPS (-0.3 / 0 / +0.1).
gen level60_shock_ret "seal=2,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,lightsVigilMana=100,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=3" -level 60
# Level 60 (sub Holy) gets its own rotation since the regear round (2026-10-23): the twist one with
# Seal of Righteousness as the seal that stays up, twisting to Seal of Command (seal=1): +1.1 /
# +0.5 / +0.9 at 120 / 180 / 300 sec on its new talents (paired, two seed sets, all positive). The
# twist talents keep Seal of Command, and the sub Holy Crusader duty is 0.4 behind with seal=1, so
# it keeps the twist duty rotation.
gen level60_sub_holy "seal=1,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistMana=0,twistWindow=0,hammerOfWrathMana=0,prepullTwo=1,judgeRighteousnessWindow=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1" -level 60
# Level 60 (vs Undead), for an Undead or Demon boss (standards 19, 2026-10-13): Exorcism and Holy
# Wrath now land. Searched on an Undead boss over 120, 180 and 300 sec together from each twist
# build, with a talent search in between, until neither moved. Seal of Righteousness stays up and
# Twist of Light twists to Seal of Command before each swing at any mana, Judgement on whichever
# seal is up, Holy Strike, Hammer of Wrath and Exorcism on cooldown, Holy Wrath above 80% mana,
# Consecration last above 20%. The searches on the twist and sub Holy talents ended within 2 DPS
# of each other summed over both builds, and this one keeps more mana at 300 sec.
# Regear and rotation round (2026-10-23): no Holy Wrath (holyWrathMana=100, was 80). Its 2 sec
# cast resets the swing and costs more than it adds: +9.6 / +6.8 / +8.3 at 120 / 180 / 300 sec on the
# new vs Undead talents (paired, two seed sets).
gen level60_undead "seal=1,jotc=0,consecration=1,consecrationFirst=0,consecrationMana=20,twist=1,twistMana=0,twistWindow=0,hammerOfWrathMana=0,prepullTwo=1,judgeRighteousnessWindow=0,cooldownsFirst=1,holyWrathMana=100,trinketExecute=1,hammerOfWrathFirst=3" -level 60
# Level 60 Shock (vs Undead): the Shock rotation, with Light's Vigil above 30% mana while the mana
# lasts the fight and only above 60% when it won't. Holy Wrath does not make it: its 2 sec cast is
# worth less than the Holy Shocks and Holy Strikes it pushes back.
# Rotation audit pass 13 had a fight-length switch here (30% in fights up to 240 sec, 60% in longer
# ones; a flat 60% left 1300 to 1700 mana unspent at 120 and 180 sec). Standards 31 (2026-10-28)
# states what it bought instead: against Undead the Vigil and its Holy Shock (1340 mana) are the
# least mana-efficient casts next to Exorcism, Consecration and Hammer of Wrath, so once the mana
# can't keep up they get only the top of the bar. Spending the top keeps the Major Mana Potion and
# Demonic Rune coming on cooldown (they wait for a 2250 mana gap): a per-second rule on the bar
# alone held the Vigil at the pull of long fights, the first potion came late and the third never
# (-8 to -12 at 240 to 300 sec for 20 to 50 mana a second, whatever the rate). So "the mana lasts"
# counts the bar plus the potions and runes still to come with 15 sec left to spend them, at 55 mana
# a second of the fight left (vigilManaPerSec=55), 75 from 10 sec before the execute phase, where
# Hammer of Wrath needs it (vigilHammerManaPerSec=20, vigilExecuteLead=10). Below that the Vigil
# needs 60% (vigilTopMana=60). Searched by hand over 120 to 300 sec (rates 50 to 70, Hammer 0 to 80,
# lead 0 to 20, counting potions or not, with 15 to 60 sec left). Against the switch, paired, 8 seeds
# x 5000: +0.46 / -0.11 / -0.10 / +4.19 / +0.43 / +1.44 at 120 / 150 / 180 / 210 / 260 / 300 sec;
# 20 seeds x 5000: -0.16+-0.07 at 180+-20 sec and -0.13+-0.02 at 240 sec, the switch's own tuning
# point (the report's request, seed 1 x 20000: 973.11 -> 973.17).
gen level60_shock_undead "seal=1,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,lightsVigilMana=30,vigilManaPerSec=55,vigilTopMana=60,vigilHammerManaPerSec=20,vigilExecuteLead=10,hammerOfWrathMana=0,prepullTwo=1,vigilChain=1,cooldownsFirst=1,trinketExecute=1,vigilSwing=25" -level 60
# Judgement of the Crusader duty (standards 23): the base rotation plus Seal of the Crusader at
# -1.5 sec after the base pre-pull and its Judgement at the pull (jotc=1). Under Forever the
# judgement adds 161 Holy damage taken for 40 sec and the paladin's own melee hits refresh it, so
# one cast lasts the fight. The duty requests have the raid's Judgement of the Crusader off. A
# knob search on the duty request kept the twist knobs; Shock (sub Holy shares level60) is best
# with Holy Strike ahead of Holy Shock there (+1.3 / +0.6 / +1.9 at 120 / 180 / 300 sec; the base
# Shock rotation loses 0.1 to 1.1 with it, so it keeps its order).
# Rotation audit pass 20: the Shock duty puts Seal of Righteousness back up ahead of Consecration
# (sealFirst), since the opener's Judgement of the Crusader leaves it down: +6.0 / +3.0 / +0.5 at 120
# / 180 / 300 sec (base Shock, which never loses its seal early, is 1.6 behind with it at 180 sec).
gen level60_crusader_duty "seal=2,jotc=1,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistMana=0,twistWindow=0,hammerOfWrathMana=0,prepullTwo=1,judgeRighteousnessWindow=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1" -level 60
gen level60_shock_crusader_duty "seal=1,jotc=1,sealFirst=1,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,holyStrikeFirst=1,lightsVigilMana=30,hammerOfWrathMana=0,prepullTwo=1,vigilChain=1,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=2,vigilSwing=25" -level 60
# Rotation audit pass 23 (2026-10-16): Hammer of Wrath above Consecration, so in the execute phase
# it isn't held behind a Consecration that ticks for less (hammerOfWrathFirst=1; 2 puts it at the
# very top, which is where the Shock duty list gains most). Paired over two seed sets at 20k, 120 /
# 180 / 300 sec: twist +2.1 / +2.2 / +2.2, sub Holy +2.0 / +2.5 / +1.6, their duty +2.2 to 2.7 /
# +2.4 to 2.5 / +1.7 to 2.2, Shock +0.8 / +1.4 / +0.5, Shock duty +1.9 / +1.1 / +0.8. vs Undead
# already has Consecration last, and Shock vs Undead gains under 1.1, so both keep their order.
# Rotation audit pass 24 (2026-10-16, after the partial resists went): vs Undead puts Hammer of
# Wrath above Holy Strike instead (hammerOfWrathFirst=3): +2.0 / +2.4 / +1.4 at 120 / 180 / 300 sec
# paired over two seed sets, sub Holy vs Undead +1.4 / +2.2 / +1.9.
# Level 30 with 5 extra talent points (26 in all). The old level 20 rotation still wins,
# and Seal of Command stays ahead of Seal of Righteousness on Corpsemaker.
gen level30p5 "seal=2,consecration=1,consecrationFirst=1,consecrationMana=0" -level 30 -bonus-talents 5 -talents --502240312013201
