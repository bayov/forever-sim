#!/bin/bash
# search.sh <preset> <round n> : talent search (t_<n>_<preset>.txt) with knobs/knobs_<preset>, then the knob search
# (k_<n>_<preset>.txt) with the talents it found, on req/warrior_protection_<preset>.json, own weight, 1,000 / 3,000.
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round5
p=$1; n=$2; req=$R/req/warrior_protection_$p.json
tal=${TALENTS:-}
$R/opt.sh ${p//_/-} $R/t_${n}_$p.txt $req -knobs "$(cat $R/knobs/knobs_$p)" ${tal:+-talents $tal} -talent-search
best=$(grep -m1 '^  talents ' $R/t_${n}_$p.txt | awk '{print $2}')
$R/opt.sh ${p//_/-} $R/k_${n}_$p.txt $req -knobs "$(cat $R/knobs/knobs_$p)" -talents $best
