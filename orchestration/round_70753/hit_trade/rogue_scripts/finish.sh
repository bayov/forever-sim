#!/bin/bash
# finish.sh id gearsuffix tpl  -- after the climb: step scan of the final talents, then confirm
S=${S:?set S to the folder holding rotopt and these scripts, with an out/ folder}
cd /home/user/forever-sim
id=$1 gs=$2 tpl=$3
K=$(cat $S/out/${id}_knobs.txt)
G=ui/rogue/gear_sets/level60_phase1_synthetic_$gs
nt=$(tail -1 $S/out/${id}_final.txt)
cs=$($S/rotopt -template $tpl -settings tools/dps_report/req/$id.json -knobs "$K" -talents $nt -durs 180 -eval -confirm-iters 100 -hit-steps $G -hit-cap melee:8 -print-stats -out $S/apl -name t 2>&1 | grep 'hit step')
echo "capnote $cs" > $S/out/${id}_scan.txt
c=$(echo "$cs" | sed -E 's/.*cap step ([+-]?[0-9]+).*/\1/')
for st in $(seq -7 3); do
  [ -f ${G}_hit$(printf '%+d' $st).gear.json ] || { [ $st = 0 ] || continue; }
  r=$($S/rotopt -template $tpl -settings tools/dps_report/req/$id.json -knobs "$K" -talents $nt -combined -durs 120,180,300 -eval -confirm-iters 10000 -hit-steps $G -hit-cap melee:8 -hit-step $st -out $S/apl -name t 2>&1 | grep -m1 'start')
  echo "step $st $r" >> $S/out/${id}_scan.txt
done
best=$(grep '^step' $S/out/${id}_scan.txt | sed -E 's/^step ([+-]?[0-9]+) +start ([0-9.]+).*/\2 \1/' | sort -rn | head -1 | cut -d' ' -f2)
echo "best $best" >> $S/out/${id}_scan.txt
$S/confirm.sh $id $gs $tpl $nt $best > $S/out/${id}_confirm.txt 2>&1
echo done >> $S/out/${id}_confirm.txt
pins=$(cat $S/out/${id}_pins.txt)
python3 $S/precheck.py $id $tpl "$K" $G "$pins" $nt $S/out
