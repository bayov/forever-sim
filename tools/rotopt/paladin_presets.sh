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
# From the rotation audit (2026-10-13): the spell power on-use trinket and the Goblin Sapper
# Charge at the top of the list instead of on the autocast line, which a GCD line above it
# nearly always beat, so they went out 3 to 4.5 sec late and drifted from there.
gen level60 "seal=2,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistMana=20,twistWindow=15,hammerOfWrathMana=0,prepullTwo=1,judgeRighteousnessWindow=0,cooldownsFirst=1,trinketExecute=1" -level 60
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
gen level60_shock "seal=1,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,lightsVigilMana=30,hammerOfWrathMana=0,prepullTwo=1,vigilChain=1,cooldownsFirst=1,trinketExecute=1" -level 60
# Level 60 (vs Undead), for an Undead or Demon boss (standards 19, 2026-10-13): Exorcism and Holy
# Wrath now land. Searched on an Undead boss over 120, 180 and 300 sec together from each twist
# build, with a talent search in between, until neither moved. Seal of Righteousness stays up and
# Twist of Light twists to Seal of Command before each swing at any mana, Judgement on whichever
# seal is up, Holy Strike, Hammer of Wrath and Exorcism on cooldown, Holy Wrath above 80% mana,
# Consecration last above 20%. The searches on the twist and sub Holy talents ended within 2 DPS
# of each other summed over both builds, and this one keeps more mana at 300 sec.
gen level60_undead "seal=1,jotc=0,consecration=1,consecrationFirst=0,consecrationMana=20,twist=1,twistMana=0,twistWindow=15,hammerOfWrathMana=0,prepullTwo=1,judgeRighteousnessWindow=0,cooldownsFirst=1,holyWrathMana=80,trinketExecute=1" -level 60
# Level 60 Shock (vs Undead): the Shock rotation, with Light's Vigil above 30% mana in fights up
# to 240 sec and above 60% in longer ones, to leave mana for Exorcism (rotation audit pass 13: a
# flat 60% left 1300 to 1700 mana unspent at 120 and 180 sec; +9.7 / +10.5 / 0 at 120 / 180 /
# 300 sec; any switch point from 210 to 270 sec is the same). On the generic boss the longer
# floor only loses, so level60_shock keeps a flat 30%. Holy Wrath does not make it: its 2 sec cast is worth less
# than the Holy Shocks and Holy Strikes it pushes back.
gen level60_shock_undead "seal=1,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,lightsVigilMana=30,vigilLongFight=240,vigilLongMana=60,hammerOfWrathMana=0,prepullTwo=1,vigilChain=1,cooldownsFirst=1,trinketExecute=1" -level 60
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
gen level60_crusader_duty "seal=2,jotc=1,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistMana=20,twistWindow=15,hammerOfWrathMana=0,prepullTwo=1,judgeRighteousnessWindow=0,cooldownsFirst=1,trinketExecute=1" -level 60
gen level60_shock_crusader_duty "seal=1,jotc=1,sealFirst=1,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,holyStrikeFirst=1,lightsVigilMana=30,hammerOfWrathMana=0,prepullTwo=1,vigilChain=1,cooldownsFirst=1,trinketExecute=1" -level 60
# Level 30 with 5 extra talent points (26 in all). The old level 20 rotation still wins,
# and Seal of Command stays ahead of Seal of Righteousness on Corpsemaker.
gen level30p5 "seal=2,consecration=1,consecrationFirst=1,consecrationMana=0" -level 30 -bonus-talents 5 -talents --502240312013201
