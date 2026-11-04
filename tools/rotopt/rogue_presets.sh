#!/bin/sh
# Regenerates the rogue rotations in ui/rogue/apls from the rotopt templates. Run from the
# repo root with the settings file the optimizer should sim against (a RaidSimRequest or a UI
# settings export). The rotation itself does not depend on the settings, only the one sanity
# sim does.
#
#   tools/rotopt/rogue_presets.sh path/to/settings.json
set -e
settings=$1
gen() {
	template=$1
	name=$2
	knobs=$3
	shift 3
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 300 \
		-eval -confirm-iters 100 -out ui/rogue/apls -name "$name" -knobs "$knobs" "$@" >/dev/null
	mv "ui/rogue/apls/${name}_300s.apl.json" "ui/rogue/apls/$name.apl.json"
	echo "wrote ui/rogue/apls/$name.apl.json ($knobs)"
}
# The plain Sinister Strike and Backstab rotations, which the rogue Go tests run. The UI dropped
# them, their variants and the level 30 rotations on 2026-10-31.
gen rogue_ss combat_sinister_strike ""
# The Backstab build the talent search settled on has no Subtlety talents, so the
# Ghostly Strike line is left out (it would only warn).
# Blade Flurry no longer waits for Slice and Dice. Under Forever's continuous energy that
# is worth 2 DPS at 300 sec, with ticks it was a tie.
gen rogue_bs combat_backstab "ghostly=0,bfSnd=0"

# The Level 60 rotations, searched at 180 sec on the generic boss (mksettings' Level 60 target, no
# creature type) and the synthetic Phase 1 sets with the
# mksettings rogue_combat, rogue_assassination and rogue_subtlety setups (knob search,
# talent search, knob search again). These need no settings argument, each builds its own.
# Seal Fate Swords is the Combat Swords one with 4 point Eviscerates.
#
# Searched again on 2026-11-02 (the knob search over 120, 180 and 300 sec together, -combined, 3000
# iterations, on each build's report request tools/dps_report/req/rogue__<id>.json, seeded from the
# knobs below as they were with the builder change applied; the race knobs and the cooldown alignment
# knobs held, since the swing rule does not touch them). Each change was kept only when it beat the old
# knobs by 2 standard errors at 20000 iterations on seeds 1 and 1000003 (the mean of the three lengths)
# and lost at no length. Since 2026-10-30 an extra attack keeps the swing timer (standards 28), so timing
# the builder right after a main hand swing only delays it: every builder now goes out whenever it is
# affordable (ssAfterSwing=0 and ssEnergy=0, bsAfterSwing=0 and bsEnergy=0, the rogue_bs defaults). At
# 120 / 180 / 300 sec, seed 1 (seed 1000003 within 0.8 of the same gain, the same sign everywhere):
# Combat Swords +6.2 / +5.1 / +5.3, Seal Fate Swords +4.0 / +3.9 / +4.0, Seal Fate Daggers +1.7 / +1.5 /
# +1.7, Mutilate +1.0 / +1.1 / +1.1, Subtlety +0.4 / +0.8 / +1.1. Combat Daggers needed its 100 energy
# pool off too (the swing timing off alone lost 1.2) and then also gained from no Vanish and a Slice and
# Dice refresh with 1 sec left (sndRefresh=1; each alone passed on seed 1 only, the pair on both): +2.1 /
# +1.9 / +1.9 in all. Those two cost Seal Fate Daggers 0.3, so it has its own
# rotation again, the Combat Daggers one with the Vanish and the 3 sec refresh. The searches kept every
# other knob: Hemorrhage before Rupture ties when off (-0.1 to -0.3 on both seeds; Rupture's ticks read
# Hemorrhage as they land since 2026-10-23, and the Hemorrhage builder keeps it up anyway), Subtlety's
# Hemorrhage at 40 energy and Seal Fate Daggers' Backstab at 90 won at 3000 iterations and lost at 20000.
gen60() {
	spec=$1
	gear=$2
	template=$3
	name=$4
	talents=$5
	knobs=$6
	settings=$(mktemp)
	go run --tags=with_db ./tools/rotopt/mksettings -spec "$spec" -gear "$gear" -apl level60_combat_daggers \
		-talents "$talents" -race Undead -duration 180 -out "$settings"
	go run --tags=with_db ./tools/rotopt -template "$template" -settings "$settings" -durs 180 \
		-eval -confirm-iters 100 -out ui/rogue/apls -name "$name" -knobs "$knobs" >/dev/null
	mv "ui/rogue/apls/${name}_180s.apl.json" "ui/rogue/apls/$name.apl.json"
	rm "$settings"
	echo "wrote ui/rogue/apls/$name.apl.json ($knobs)"
}
# Round 70753 phase 5 (2026-11-04, hit trade, orchestration/round_70753/hit_trade/rogue.md): the six builds moved onto hit steps of
# their sets (talents and knobs unchanged): Combat Swords level60_phase1_synthetic_swords_hit+2, Combat Daggers
# level60_phase1_synthetic_combat_daggers_hit+2, Seal Fate Swords level60_phase1_synthetic_seal_fate_swords_hit+3, Seal Fate Daggers
# level60_phase1_synthetic_daggers_hit+3, Subtlety level60_phase1_synthetic_subtlety_hit+2, Mutilate level60_phase1_synthetic_mutilate_hit+3;
# use those as -gear in the gen60 lines below for them.
# Round 70753 (2026-11-03, the November 2 notes' rework; the tree is server hotfix push 113483's (1.60.1.70707+hf) since 2026-11-04, strings remapped to its order):
# talents searched again in batches of 1 to 5 point moves (scratch script, 150 valid neighbours a round at
# 3000 iterations, the best checked at 10000 and kept for 2 standard errors), alternated with the knob
# search over 120, 180 and 300 sec together (-combined) until neither moved. The knobs below are the whole
# set the search left, so the file regenerates exactly what the presets carry. Numbers in ui/rogue/presets.ts.
gen60 rogue_combat level60_phase1_synthetic_swords rogue_ss level60_combat_swords 00530320201-3200330112012321551 "arEnergy=39,arSnd=1,berserkingHoldForBf=0,bfHoldForAr=0,bfPrepull=1,bfSnd=1,bloodFuryHoldForAr=1,bloodFurySnd=0,cbPool=0,cbSnd=6,eluneHoldForAr=1,eurekaCp=5,eurekaSnd=8,evisCp=5,evisEnergy=79,ruptureCp=0,ruptureSnd=6,shankEnergy=75,sndCp=1,sndRefresh=3,sndRefreshCp=5,ssAfterSwing=0,ssEnergy=0,teaEnergy=10,trinketHoldForAr=0,trinketSnd=0"
gen60 rogue_combat level60_phase1_synthetic_combat_daggers rogue_bs level60_combat_daggers 005203002-3203230002012221551-005 "arEnergy=49,arSnd=1,berserkingHoldForBf=0,bfHoldForAr=0,bfPrepull=1,bfSnd=0,bloodFuryHoldForAr=1,bloodFurySnd=0,bsAfterSwing=0,bsEnergy=0,cbPool=0,cbSnd=6,deathMarkCp=0,eluneHoldForAr=1,eurekaCp=0,eurekaSnd=8,evisCp=4,evisEnergy=79,ghostly=0,ghostlyAmbush=0,ghostlyGarrote=0,hemoFiller=0,hemoForRupture=1,meldCp=3,mutilate=0,premedCp=9,ruptureCp=0,ruptureSnd=6,shankEnergy=75,sndCp=1,sndRefresh=1,sndRefreshCp=5,teaEnergy=10,trinketHoldForAr=0,trinketSnd=0,vanishCp=9,vanishPool=0"
gen60 rogue_combat level60_phase1_synthetic_daggers rogue_bs level60_seal_fate_daggers 0053032024105103-321303010201132 "arEnergy=59,arSnd=1,berserkingHoldForBf=0,bfHoldForAr=0,bfPrepull=0,bfSnd=0,bloodFuryHoldForAr=1,bloodFurySnd=0,bsAfterSwing=0,bsEnergy=0,cbPool=0,cbSnd=6,deathMarkCp=0,eluneHoldForAr=1,eurekaCp=0,eurekaSnd=8,evisCp=4,evisEnergy=79,ghostly=0,ghostlyAmbush=0,ghostlyGarrote=0,hemoFiller=0,hemoForRupture=1,meldCp=3,mutilate=0,premedCp=9,ruptureCp=0,ruptureSnd=6,shankEnergy=75,sndCp=1,sndRefresh=3,sndRefreshCp=5,teaEnergy=0,trinketHoldForAr=0,trinketSnd=0,vanishCp=3,vanishPool=0"
gen60 rogue_assassination level60_phase1_synthetic_mutilate rogue_mut level60_mutilate 00530320251451031-320303-005 "arEnergy=59,arSnd=0,berserkingHoldForBf=0,bfHoldForAr=0,bfPrepull=0,bfSnd=0,bloodFuryHoldForAr=1,bloodFurySnd=0,bsAfterSwing=0,bsEnergy=0,cbPool=0,cbSnd=6,deathMarkCp=0,eluneHoldForAr=1,eurekaCp=5,eurekaSnd=8,evisCp=4,evisEnergy=79,ghostly=0,ghostlyAmbush=0,ghostlyGarrote=0,hemoFiller=0,hemoForRupture=1,meldCp=3,mutilate=1,premedCp=9,ruptureCp=0,ruptureSnd=6,shankEnergy=75,sndCp=3,sndRefresh=1,sndRefreshCp=3,teaEnergy=10,trinketHoldForAr=0,trinketSnd=0,vanishCp=3,vanishPool=0"
gen60 rogue_subtlety level60_phase1_synthetic_subtlety rogue_sub level60_subtlety 00530301-320003-115003231320021151 "arEnergy=59,arSnd=0,berserkingHoldForBf=0,bfHoldForAr=0,bfPrepull=0,bfSnd=0,bloodFuryHoldForAr=1,bloodFurySnd=0,bsAfterSwing=0,bsEnergy=40,cbPool=0,cbSnd=6,deathMarkCp=5,eluneHoldForAr=1,eurekaCp=5,eurekaSnd=8,evisCp=5,evisEnergy=79,ghostly=0,ghostlyAmbush=0,ghostlyGarrote=0,hemoFiller=1,hemoForRupture=0,meldCp=3,mutilate=0,premedCp=3,ruptureCp=1,ruptureSnd=0,shankEnergy=75,sndCp=1,sndRefresh=1,sndRefreshCp=5,teaEnergy=10,trinketHoldForAr=0,trinketSnd=0,vanishCp=3,vanishPool=1"
gen60 rogue_combat level60_phase1_synthetic_seal_fate_swords rogue_ss level60_seal_fate_swords 0053032024105103-320033011201132 "arEnergy=39,arSnd=1,berserkingHoldForBf=0,bfHoldForAr=0,bfPrepull=1,bfSnd=1,bloodFuryHoldForAr=1,bloodFurySnd=0,cbPool=0,cbSnd=6,eluneHoldForAr=1,eurekaCp=5,eurekaSnd=8,evisCp=4,evisEnergy=79,ruptureCp=0,ruptureSnd=6,shankEnergy=75,sndCp=1,sndRefresh=3,sndRefreshCp=5,ssAfterSwing=0,ssEnergy=0,teaEnergy=10,trinketHoldForAr=0,trinketSnd=0"
