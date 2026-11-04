#!/bin/sh
# Regenerates the retribution paladin presets in ui/retribution_paladin/apls that come
# from the paladin_ret template.
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
# Re-searched for the launch realms' October 30 hotfix trees (1.60.1.70707+hf, 2026-11-01), which
# replaced Twist of Light with Unsealed Wrath and made Light's Vigil heal only. Each build: a talent
# search at 180 sec (tools/rotopt -talent-search -require pursuitOfJustice, 3000 iterations), then the
# knobs searched over 120, 180 and 300 sec together (-combined), then each choice checked at 20000
# iterations, seed 1, at 120 / 180 / 300 sec (standards 31). Kept from before: the on-use trinket and
# the Goblin Sapper Charge at the top (cooldownsFirst, the trinket held for the execute phase when
# that loses no use), Hammer of Wrath in the execute phase at any mana, Judgement on cooldown.
# Every build has Swift Judgement off: the 16 Protection points it takes cost the twist build 60 at
# 180 sec (979 with it against 1039), though the free judgement itself is worth 26.
#
# Twist (level60): Seal of Command stays up; once it has stored 2 Unsealed Wrath copies, Seal of
# Righteousness goes up for the swing that spends them (and stores one of its own), and Seal of
# Command goes back once it has. Copies a swap, 120 / 180 / 300 sec: 1 copy -7.5 / -11.6 / -36.7, 3
# copies -4.6 / -4.3 / +1.9 against 2, the twist off -161. The 3 copies win at 300 sec only because
# each swap is a seal's mana, so the twist waits for 3 when the bar won't cover 30 mana a second of
# the fight left (twistThriftRate=30): 1043.4 / 1039.3 / 1028.9, against 2 always +0.0 / +0.3 / +2.9
# and 3 always +4.6 / +4.6 / +1.0. Seal of Righteousness as the seal that stays up -11.5, Consecration
# after the seal lines -7.6, twisting only within 0.5 to 3 sec of the swing -1 to -15.
# Round 70753 exploration (2026-11-03): Holy Strike ahead of the seal and twist lines (holyStrikeFirst=1)
# on twist and sub Holy, 20000 iterations at 120 / 180 / 300 sec: twist +1.2 / +1.2 / +0.4 (seed 1) and
# +2.4 / +1.3 / +1.9 (seed 1000003), sub Holy +1.0 / +1.4 / +1.5 and +2.5 / +2.2 / +2.5. The others keep
# it off: the Crusader duty loses (-5.4 / -4.0 / -1.5), the Wisdom duty, Ret Shock and vs Undead gain
# under 1 at 180 sec (+0.3 to +1.3).
# Round 70753 phase 5 (2026-11-04, hit trade): twist keeps its talents and knobs on the set's hit-3 step
# (Precision 3 put it 3% over the 8% melee cap; 30 hit rating re-spent on crit): +12.2 / +11.1 / +13.0 at
# 120 / 180 / 300 sec, its Crusader and Wisdom duties +11 to +12 (orchestration/round_70753/hit_trade/paladin.md).
gen level60 "seal=2,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistStacks=2,twistThriftRate=30,twistBack=1,twistMana=0,twistWindow=0,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,swiftJudgement=0,holyStrikeFirst=1" -level 60
# Sub Holy (level60_sub_holy): the knob search kept the twist knobs. It has the mana (Reverence,
# Divine Intellect) to swap on 2 copies at any length: the mana rule loses 2.3 at 300 sec there.
gen level60_sub_holy "seal=2,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistStacks=2,twistBack=1,twistMana=0,twistWindow=0,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,swiftJudgement=0,holyStrikeFirst=1" -level 60
# Judgement duties (standards 23), shared by twist and sub Holy: the sub Holy knobs plus Seal of the
# Crusader (jotc=1) or Seal of Wisdom (jotc=2) at -1.5 sec and its Judgement at the pull. Under Forever
# the judgement lasts 40 sec and the paladin's own melee hits refresh it. The mana rule loses 2.8 at
# 300 sec on the sub Holy Crusader duty.
gen level60_crusader_duty "seal=2,jotc=1,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistStacks=2,twistBack=1,twistMana=0,twistWindow=0,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,swiftJudgement=0" -level 60
gen level60_wisdom_duty "seal=2,jotc=2,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistStacks=2,twistBack=1,twistMana=0,twistWindow=0,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,swiftJudgement=0" -level 60
# Sub Holy vs Undead (level60_undead, standards 19): Exorcism lands, so the copies wait for all 3 (it
# needs the mana: -0.2 / -0.8 / +2.2 at 120 / 180 / 300 against the mana rule), no Holy Wrath (its 2
# sec cast resets the swing). Round 70753 (Two-Handed Weapon Specialization on all Holy damage, Vengeance
# 5 stacks; knobs searched over 120, 180 and 300 sec, every other rotation kept its knobs): Consecration
# first and Hammer of Wrath in the execute phase at the top (consecrationFirst 0 -> 1, hammerOfWrathFirst
# 3 -> 1), +8.4 / +7.0 / +4.0 at 120 / 180 / 300 sec (seed 1, 20000 iterations; +8.5 / +7.2 / +5.0 on
# seed 1000003).
gen level60_undead "seal=2,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistStacks=3,twistBack=1,twistMana=0,twistWindow=0,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,exorcism=1,holyWrathMana=100,swiftJudgement=0" -level 60
# Ret Shock (level60_shock_ret, 20 Holy for Holy Shock, so no Unsealed Wrath): the knob search kept
# its knobs. Seal of Command up, Consecration first at any mana, Judgement, Holy Strike and Holy Shock
# on cooldown, Hammer of Wrath above Holy Strike. Its duties put the seal back up ahead of
# Consecration, since the opener's judgement leaves it down (sealFirst: +0.7 / +1.8 / +2.1).
gen level60_shock_ret "seal=2,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=3,swiftJudgement=0" -level 60
gen level60_shock_ret_crusader_duty "seal=2,jotc=1,sealFirst=1,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=3,swiftJudgement=0" -level 60
gen level60_shock_ret_wisdom_duty "seal=2,jotc=2,sealFirst=1,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=3,swiftJudgement=0" -level 60
# Ret Shock (deep Holy), the 31 point Holy Shock build (level60_shock, its duties and level60_shock_undead),
# was dropped by the user's ruling (2026-11-02, "drop deep holy shock"); its rotations went with it.
