#!/bin/bash
# confirm.sh id gearsuffix tpl newtalents newstep
S=${S:?set S to the folder holding rotopt and these scripts, with an out/ folder}
cd /home/user/forever-sim
id=$1 gs=$2 tpl=$3 nt=$4 st=$5
K=$(cat $S/out/${id}_knobs.txt)
G=ui/rogue/gear_sets/level60_phase1_synthetic_$gs
mkdir -p $S/apl
for d in 120 180 300; do
  echo "old $d $($S/rotopt -template $tpl -settings tools/dps_report/req/$id.json -knobs "$K" -durs $d -eval -confirm-iters 20000 -out $S/apl -name $id-old 2>&1 | grep -m1 'start')"
  echo "new $d $($S/rotopt -template $tpl -settings tools/dps_report/req/$id.json -knobs "$K" -talents $nt -durs $d -eval -confirm-iters 20000 -hit-steps $G -hit-cap melee:8 -hit-step $st -out $S/apl -name $id-new 2>&1 | grep -m1 'start')"
done
