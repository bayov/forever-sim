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
gen level60 "seal=2,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistMana=20,twistWindow=15,hammerOfWrathMana=0,prepullTwo=1,judgeRighteousnessWindow=25" -level 60
# Level 60 Shock, for the Holy Shock talents (no Twist of Light, so no twisting): Consecration
# first down to the last point of mana (2 ahead of a 20% floor), Judgement on cooldown on Seal of
# Righteousness, Holy Strike on cooldown (90), Light's Vigil above 40% mana just before Holy
# Shock is ready (64), so the Holy Shock after it detonates it and starts no cooldown, Holy Shock
# after Holy Strike (148, 6.7 ahead of before it), Hammer of Wrath at any mana (23).
# From the rotation audit (2026-10-12): Light's Vigil before the pull, and then whenever it is
# off the target above 30% mana with Holy Shock waiting for it (it used to wait for Holy
# Shock's cooldown), so nearly every Holy Shock detonates a Vigil and starts no cooldown:
# +22.8 at 180 sec. The floor has to stay over the Vigil's 1340 mana.
gen level60_shock "seal=1,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,lightsVigilMana=30,hammerOfWrathMana=0,prepullTwo=1,vigilChain=1" -level 60
# Level 30 with 5 extra talent points (26 in all). The old level 20 rotation still wins,
# and Seal of Command stays ahead of Seal of Righteousness on Corpsemaker.
gen level30p5 "seal=2,consecration=1,consecrationFirst=1,consecrationMana=0" -level 30 -bonus-talents 5 -talents --502240312013201
