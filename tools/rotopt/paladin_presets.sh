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
# November 11 hotfix (Judgements roll as ranged attacks, 2026-11-12): knob searches over 120, 180 and 300
# sec, each choice confirmed at 20,000 iterations on seeds 1 and 1000003. Seal of Righteousness is now the
# seal that stays up (seal=1) and the twist goes back once Command has stored all 3 copies (twistBack=3);
# the duties and Undead swap on 3 copies. Gains at 120 / 180 / 300 sec, seed 1: twist +1.9 / +1.7 / +3.8,
# sub Holy +2.7 / +2.1 / +3.1, Crusader duty +2.9 / +3.6 / +10.0, Wisdom duty (plus holyStrikeFirst=1)
# +4.2 / +4.1 / +10.6, sub Holy Crusader duty +2.0 / +2.3 / +2.8, sub Holy Wisdom duty +4.1 / +3.9 / +4.2,
# Undead +6.3 / +4.6 / +5.3, Ret Shock (plus holyShock=2) +6.5 / +5.1 / +4.8, its Wisdom duty +5.8 / +5.0 /
# +4.2. Seed 1000003 within 1 of these. The Shock Crusader duty keeps its knobs (seal=1 and holyShock=2
# tie or lose). The talent searches keep the twist, sub Holy and Ret Shock strings (twist: Precision -1 for
# Reverence +1 loses 5.2); vs Undead trades Reverence 1 for Divine Intellect 3 (+1.1 to +1.6 in every cell).
# November 13 hotfixes (Judgement of Command melee again, rolling Windfury Totem and on-hit procs; Windfury
# Totem on every landed main hand melee hit, 2026-11-15): knob searches over 120, 180 and 300 sec (Judgement of
# the Crusader or Wisdom frozen on the duties), each choice confirmed at 20,000 iterations on seeds 1 and
# 1000003. Seal of Righteousness stays up on the twist builds, but the swap away now goes on a single copy:
# twist twistStacks 2 -> 1, twistWindow 0 -> 2.5 sec, twistThriftRate 30 -> 0 (+4.1 / +4.4 / +5.2 at 120 /
# 180 / 300 sec, seed 1000003 +3.7 / +4.0 / +5.0); sub Holy 2 -> 1 (+3.6 / +3.1 / +3.4, +2.6 / +3.1 / +3.3);
# both duty rotations 3 -> 1 (twist Crusader +7.4 / +7.6 / +7.5, Wisdom +7.8 / +8.3 / +7.4; sub Holy Crusader
# +7.5 / +8.3 / +7.4, Wisdom +6.3 / +7.0 / +6.2; seed 1000003 within 1.4); vs Undead 3 -> 2 (+1.6 / +1.0 /
# +1.4, +1.2 / +1.2 / +1.3; 1 copy loses 4.4 at 300). Ret Shock keeps Seal of Command up (seal 1 -> 2): +29.2
# / +29.8 / +29.9 and +28.7 / +29.8 / +29.9, its Wisdom duty +25.4 / +26.7 / +28.1 and +25.6 / +27.5 / +29.0
# (its Crusader duty already had it). Seal of Command kept up on twist loses 18 at every length; judging
# Righteousness only far from the swing (judgeRighteousnessWindow) -1.7 to -2.8. Windfury timing gates (new
# knobs): holding Holy Strike and the Judgement of Command while the totem's 100 ms internal cooldown runs
# (wfIcdGate) is +0.0 to +0.1, holding them for the next swing within 100 ms (swingGap) -0.1 to -1.1; not kept.
# Talent searches at 180 sec kept every string (Undead's search moved to 1500520022-3-33233001201301351,
# which lost -0.9/+0.0/-15.8 and -0.2/+0.3/-15.0 on mana).
gen level60 "seal=1,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistStacks=1,twistBack=3,twistMana=0,twistWindow=25,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,swiftJudgement=0,holyStrikeFirst=1" -level 60
# Sub Holy (level60_sub_holy): the knob search kept the twist knobs. It has the mana (Reverence,
# Divine Intellect) to swap on 2 copies at any length: the mana rule loses 2.3 at 300 sec there.
gen level60_sub_holy "seal=1,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistStacks=1,twistBack=3,twistMana=0,twistWindow=0,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,swiftJudgement=0,holyStrikeFirst=1" -level 60
# Judgement duties (standards 23), shared by twist and sub Holy: the sub Holy knobs plus Seal of the
# Crusader (jotc=1) or Seal of Wisdom (jotc=2) at -1.5 sec and its Judgement at the pull. Under Forever
# the judgement lasts 40 sec and the paladin's own melee hits refresh it. The mana rule loses 2.8 at
# 300 sec on the sub Holy Crusader duty.
gen level60_crusader_duty "seal=1,jotc=1,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistStacks=1,twistBack=3,twistMana=0,twistWindow=0,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,swiftJudgement=0" -level 60
gen level60_wisdom_duty "seal=1,jotc=2,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistStacks=1,twistBack=3,twistMana=0,twistWindow=0,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,swiftJudgement=0,holyStrikeFirst=1" -level 60
# Sub Holy vs Undead (level60_undead, standards 19): Exorcism lands, so the copies wait for all 3 (it
# needs the mana: -0.2 / -0.8 / +2.2 at 120 / 180 / 300 against the mana rule), no Holy Wrath (its 2
# sec cast resets the swing). Round 70753 (Two-Handed Weapon Specialization on all Holy damage, Vengeance
# 5 stacks; knobs searched over 120, 180 and 300 sec, every other rotation kept its knobs): Consecration
# first and Hammer of Wrath in the execute phase at the top (consecrationFirst 0 -> 1, hammerOfWrathFirst
# 3 -> 1), +8.4 / +7.0 / +4.0 at 120 / 180 / 300 sec (seed 1, 20000 iterations; +8.5 / +7.2 / +5.0 on
# seed 1000003).
gen level60_undead "seal=1,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=20,twist=1,twistStacks=2,twistBack=3,twistMana=0,twistWindow=0,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=1,exorcism=1,holyWrathMana=100,swiftJudgement=0" -level 60
# Ret Shock (level60_shock_ret, 20 Holy for Holy Shock, so no Unsealed Wrath): the knob search kept
# its knobs. Seal of Command up, Consecration first at any mana, Judgement, Holy Strike and Holy Shock
# on cooldown, Hammer of Wrath above Holy Strike. Its duties put the seal back up ahead of
# Consecration, since the opener's judgement leaves it down (sealFirst: +0.7 / +1.8 / +2.1).
gen level60_shock_ret "seal=2,jotc=0,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=2,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=3,swiftJudgement=0" -level 60
gen level60_shock_ret_crusader_duty "seal=2,jotc=1,sealFirst=1,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=3,swiftJudgement=0" -level 60
gen level60_shock_ret_wisdom_duty "seal=2,jotc=2,sealFirst=1,consecration=1,consecrationFirst=1,consecrationMana=0,holyShock=1,hammerOfWrathMana=0,cooldownsFirst=1,trinketExecute=1,hammerOfWrathFirst=3,swiftJudgement=0" -level 60
# Ret Shock (deep Holy), the 31 point Holy Shock build (level60_shock, its duties and level60_shock_undead),
# was dropped by the user's ruling (2026-11-02, "drop deep holy shock"); its rotations went with it.
