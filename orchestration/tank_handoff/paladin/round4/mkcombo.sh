#!/bin/bash
# mkcombo.sh : writes the stacked-consumable variants c_<flask|fl0>_<food>_<zanza>_<sap|nos>_<rum|norum>_<st|nost> per preset
cd /home/user/forever-sim; R=orchestration/tank_handoff/paladin/round4; V=$R/var
declare -A F=([nf]=FoodNightfinSoup [tw]=FoodTenderWolfSteak [dg]=FoodDirgesKickChimaerokChops)
declare -A Z=([sc]=GroundScorpokAssay [sp]=SpiritOfZanza [lu]=LungJuiceCocktail)
mk() { # name preset flask food zanza sap rum st
  a=(); [ $3 = 1 ] && a+=(flask=FlaskOfSupremePower); a+=(food=${F[$4]} zanzaBuff=${Z[$5]}); [ $6 = 1 ] && a+=(sapperExplosive=SapperGoblinSapper)
  [ $7 = 1 ] && a+=(alcohol=AlcoholRumseyRumBlackLabel); [ $8 = 1 ] && a+=(mainHandImbue=ElementalSharpeningStone)
  python3 $R/mkcons.py $V/base_$2.json $V/$1_$2.json "${a[@]}"; }
for p in max_threat balanced max_survival; do
 for f in nf tw dg; do for z in sc sp lu; do mk c_1_${f}_${z}_1_1_1 $p 1 $f $z 1 1 1; done; done
 mk c_0_nf_sc_1_1_1 $p 0 nf sc 1 1 1; mk c_1_nf_sc_0_1_1 $p 1 nf sc 0 1 1; mk c_1_nf_sc_1_0_1 $p 1 nf sc 1 0 1; mk c_1_nf_sc_1_1_0 $p 1 nf sc 1 1 0
done
# second pass: Max survival keeps the Flask of the Titans (Supreme Power loses there); leave-outs on the best
mk c_0_tw_sp_1_1_1 max_survival 0 tw sp 1 1 1; mk c_0_dg_sp_1_1_1 max_survival 0 dg sp 1 1 1; mk c_0_tw_sc_1_1_1 max_survival 0 tw sc 1 1 1
mk c_0_tw_sp_0_1_1 max_survival 0 tw sp 0 1 1; mk c_0_tw_sp_1_0_1 max_survival 0 tw sp 1 0 1; mk c_0_tw_sp_1_1_0 max_survival 0 tw sp 1 1 0
mk c_0_dg_sp_0_1_1 max_survival 0 dg sp 0 1 1
