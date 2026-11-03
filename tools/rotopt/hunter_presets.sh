#!/bin/sh
# Regenerates the hunter rotations in ui/hunter/apls from the hunter_ranged and hunter_sv
# templates, with the knobs the searches settled on (last on 2026-11-03, round 70753; before that 2026-11-02, at 120 / 180 / 300 sec on
# the synthetic Phase 1 sets with the 8% hit cap, the Level 60 talents and Classic's crit per
# Agility at 60; Beast Mastery last on 2026-10-25).
# Run from the repo root with a settings file the optimizer should sim against, made with
# mksettings for the spec. The rotation itself does not depend on the settings, only the one
# sanity sim does.
#
#   go run --tags=with_db ./tools/rotopt/mksettings -spec hunter_bm -gear level60_phase1_synthetic_ranged \
#       -apl bm -talents 5021201505021241-00510504005 -race Orc -duration 180 -out bm.json
#   tools/rotopt/hunter_presets.sh bm.json
set -e
settings=$1
# Hunter's Mark (standards 23, user ruling 2026-10-25): there is no Hunter's Mark duty preset. Every
# hunter keeps up its own mark, with the raid's off in the request (mksettings' hunter specs), so
# every rotation here casts it (the templates' markPrepull, 1.5 by default). It goes up 1.5 sec before
# the pull (it doesn't pull; user ruling 2026-10-15), after the pre-pull Aspect's global cooldown, so
# the boss has it from the first shot (rotation audit pass 20: +4.7 / +3.3 / +3.4 / +1.4 at 180 sec on
# Beast Mastery / Marksmanship / Lone Wolf / Survival over the first global cooldown after the pull,
# 20000 iterations), and again in its last 3 sec. It is not held back for the hunter's own DPS (user
# ruling). Against the raid's mark it costs 0.5 to 2.5 DPS at 180 sec.
gen() {
	template=$1
	name=$2
	knobs=$3
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 180 -combined \
		-eval -confirm-iters 100 -out ui/hunter/apls -name "$name" -knobs "$knobs" >/dev/null
	echo "wrote ui/hunter/apls/$name.apl.json ($knobs)"
}
# The ranged specs, from the knob searches on the standard gear (2026-10-12). Marksmanship and Lone
# Wolf put the Arcane Shot cooldown above Aimed Shot, drop the mana floors and cast Serpent Sting
# after Aimed Shot (+1.7 Lone Wolf, +2.1 Marksmanship with Forever's consumables). Beast Mastery keeps Aimed Shot first and the floors. All three
# cast Rapid Fire 0.1 sec before the pull (+2 at 120 sec), and Marksmanship and Lone Wolf cast Sniper
# Shot only while the mana covers 25 a second of the fight left (+8 and +11 at 300 sec, nothing at 180;
# both from the rotation auditor). Summon Hawk while fewer than 2 hawks are out (hawk=2) only matters
# with the talent, which no build takes.
ranged="aimed=1,aimedMana=0,hawk=2,intimidation=0,serpent=1,sniper=1,sniperFirst=0,rapidPrepull=1"
# The regear and rotation round (2026-10-23: no hit suppression, the 8% cap, the Level 60 talents of
# that day), knob searches at 120, 180 and 300 sec combined, 3000 iterations, confirmed at 20000.
# Beast Mastery now puts the Arcane Shot cooldown above Aimed Shot and drops the Serpent Sting and
# Arcane Shot mana floors (+4.4 combined); it keeps Serpent Sting last (serpentLate=2: +2.2 with its
# own Hunter's Mark over first). Marksmanship and Lone Wolf cast Aimed Shot only above 40% mana
# (+1.4, +1.5) and hold Sniper Shot back sooner, while the mana covers 10 a second of the fight left
# (+2.5 and +1.9 over 25; Lone Wolf at range +1.3 and +1.0 over 20 on the two seed sets). All three cast Rapid Fire 0.1 sec before
# the pull (+2 at 120 sec), and Summon Hawk while fewer than 2 hawks are out (hawk=2) only matters
# with the talent, which no build takes (the hawk Beast Mastery build is 11.8 behind at 180 sec).
bm="$ranged,arcaneFirst=1,arcaneMana=0,serpentLate=2,serpentMana=0,holdElune=1"
mmBase="$ranged,arcaneFirst=1,arcaneMana=0,serpentLate=1,serpentMana=0,aimedMana=40,holdRapid=1,holdElune=1"
mm="$mmBase,sniperManaPerSec=10"
# The recheck after the hunter's crit per Agility became Classic's at 60 (2026-10-25, knob searches
# at 120, 180 and 300 sec combined, confirmed at 20000 iterations on seeds 1 and 1000003): the talents
# and the other knobs hold (the searches' one other move, weaveReact 0.1, is faster than a player),
# and Marksmanship and Lone Wolf now put Multi-Shot on the cooldown it shares with Aimed Shot
# (aimed=0). Marksmanship on its Human +2.9 / +1.8 / +1.2 at 120 / 180 / 300 sec (seed 1000003 +3.3 /
# +2.2 / +0.8), Lone Wolf +1.7 / +1.6 / +1.2 (+0.5 / +2.6 / +1.5). Beast Mastery keeps Aimed Shot.
# The recheck of 2026-11-02 (knob searches at 120, 180 and 300 sec combined, 3000 iterations, the
# report's requests: Marksmanship on its Orc, Lone Wolf on its Night Elf, seeded from those knobs
# and from aimed=1, frozen: the reaction times, markPrepull, leylineMana, hawk, intimidation) takes
# Aimed Shot back (aimed=1): it goes on the shared cooldown above 40% mana and Multi-Shot below.
# From the old knobs the searches' first move was aimed=1, and from there the searches are the same.
# They took that one move and no other (aimedMana stays 40: 0 to 30 lose 0.1 to 1.0 combined). At
# 20000 iterations Marksmanship +6.0 / +2.7 / -0.3 at 120 / 180 / 300 sec (seed 1000003 +6.5 / +3.0
# / +0.1), Lone Wolf +7.6 / +2.6 / +0.2 (+7.5 / +2.9 / -0.2): 1081.8 / 1066.4 / 1039.1 and 1107.6 /
# 1090.0 / 1055.5. Aimed Shot's +166 over Multi-Shot's shot is worth its 2 sec cast (Auto Shot keeps
# firing through it) and its 310 mana while the mana is high; a 300 sec fight runs out of mana 8 sec
# sooner with it, which takes the gain back there. (The searches' fights are the requests' 20 sec
# either way.) The races keep their order (180 sec, 20000 iterations, seed 1, each with its race
# override): Marksmanship Orc 1066.4, Night Elf 1061.3, both Skyborne 1059.9, Human 1057.8, Troll
# 1055.9, Dwarf 1053.5, Tauren 1046.4; Lone Wolf Night Elf 1090.0, both Skyborne 1089.5, Orc 1088.6,
# Human 1085.8, Troll 1085.4, Dwarf 1081.5, Tauren 1073.5.
# The melee weave (2026-10-23, after the weave mechanics of wowsims/forever): between Auto Shots the
# hunter walks from 12 yd into melee for a main hand swing Raptor Strike is ready for, when that holds
# the next Auto Shot back 1.2 sec or less and Arcane Shot and the Aimed Shot cooldown stay down until
# it is back but 0.75 sec. It keeps the player's 0.2 sec reaction before the walk in and after the
# walk out (the search's 0.1 sec is +3.3 more, but faster than a player). Beast Mastery weaves (+9.5
# combined over standing at range, 20000 iterations). Marksmanship and Lone Wolf weave only while the
# mana covers 25 a second of the fight left (weaveManaPerSec): weaving wins a short fight (+4 at 120
# sec) and costs a long one the mana (-11 at 180 sec, -15 at 300 sec without the gate). With the gate
# Marksmanship reads 1184.8 / 1167.6 / 1140.2 at 120 / 180 / 300 sec, against 1180.1 / 1168.9 / 1142.0
# at range, and stays above the 2026-10-23 rotation at all three lengths, which the turret rotation
# doesn't (-4 at 120 sec).
weave="weave=1,weaveRaptor=1,weaveDelay=1.2,weaveShotSlack=0.75"
mmWeave="$weave,weaveManaPerSec=25"
# Round 70753 (2026-11-03, after the 2026-11-03 server hotfix: Aimed Shot +350, Arcane Shot 280 and 15%
# of attack power, Serpent Sting 675 and 5% a tick, Sniper Shot +590, Barrage 15%, Hawk 110, Trueshot
# Aura 100). Knob searches at 120, 180 and 300 sec combined, 3000 iterations, the report's requests,
# frozen: the reaction times, markPrepull, leylineMana, hawk, intimidation. Confirmed at 20000
# iterations on seeds 1 and 1000003 (120 / 180 / 300 sec):
# - Marksmanship casts Aimed Shot at any mana (aimedMana=0: its +350 is worth its 310 mana now) and
#   Serpent Sting first again (serpentLate=0, its 675 and 5% a tick): 1268.9 / 1254.3 /
#   1217.6, +1.3 / +10.0 / +12.8 (seed 1000003 +0.9 / +9.0 / +11.9).
# - Lone Wolf casts Aimed Shot at any mana (aimedMana=0) and keeps Serpent Sting after it:
#   1332.4 / 1313.7 / 1264.7, +1.8 / +10.0 / +9.8 (+2.1 / +10.2 / +10.7).
# - Beast Mastery walks into melee only when Arcane Shot and the Aimed Shot cooldown stay down until
#   it is back at range (weaveShotSlack=0, was 0.75 sec of slack: the bigger shots aren't delayed), on its new talents:
#   1197.3 / 1195.6 / 1176.3, the knob +1.3 / +1.6 / +2.0 over the talents alone (+0.7 / +1.5 / +2.2).
mm70753="$ranged,arcaneFirst=1,arcaneMana=0,serpentLate=0,serpentMana=0,aimedMana=0,holdRapid=1,holdElune=1,sniperManaPerSec=10"
lw70753="$ranged,arcaneFirst=1,arcaneMana=0,serpentLate=1,serpentMana=0,aimedMana=0,holdRapid=1,holdElune=1,sniperManaPerSec=10"
gen hunter_ranged bm "$bm,$weave,weaveShotSlack=0"
gen hunter_ranged mm "$mm70753,$mmWeave"
gen hunter_ranged lw "$lw70753,$mmWeave"
# The Beast Mastery rotation at range, which the hunter Go tests run. The UI dropped the turret
# rotations on 2026-10-31.
gen hunter_ranged bm_turret "$bm"
# Marksmanship and Lone Wolf hold Rapid Fire and Elune's Light for the on-use trinket when the wait
# costs no use, Beast Mastery only Elune's Light (rotation auditor, 2026-10-13: +2.3 / +1.7 / +0.5
# at 300 sec, nothing at 120 and 180).
# Survival melees and doesn't weave (2026-10-23, 180 sec, 20000 iterations): stepping out to 12 yd for
# each Auto Shot after a main hand swing is 874.0 against its 997.2, and the ranged weave rotation on
# its build and gear 850.1.
# Survival: Wing Clip in every free global cooldown (+31 at 180 sec), held when Mongoose Bite or
# Strider Kick is about ready (+0.4).
# Since the trap scales with ranged attack power (hunter_trap_ap) it goes right after Strider Kick,
# and Wing Clip holds only for an opened Mongoose Bite, not for Strider Kick (rotation audit pass 23,
# 2026-10-16: +2.8 / +3.0 / +3.1 at 120 / 180 / 300 sec; dropping the Mongoose Bite hold too loses 0.55).
# The recheck of 2026-11-02 (120, 180 and 300 sec combined, 3000 iterations, the report's Orc request)
# moves nothing. Survival ends a fight with about 3,700 of its 3,750 mana: it is out of global
# cooldowns, not mana (Wing Clip already fills every free one, and Resourcefulness makes the melee
# abilities cheap), so its mana floors never bind: every floor from 0 to 60% for Lacerate and the
# trap and 0 to 80% for Wing Clip sims the same to the cent, also at 20000 iterations with all
# three at 0 (1031.9 / 1024.6 / 1016.3 at 120 / 180 / 300 sec, seed 1).
sv="lacerate=1,lacerateMana=20,strikerMana=0,trap=1,trapMana=30,wingClip=1,wingClipMana=50,wingClipHold=2,trapFirst=1"
gen hunter_sv sv "$sv"
# No Scorpid Sting duty (standards 23, user ruling 2026-10-25): Scorpid Sting shares Insect Swarm's
# 2% hit reduction slot, which Balance druids keep up for its own damage, so a hunter on it would give
# up Serpent Sting's slot (about 85 to 100 DPS) for nothing.
