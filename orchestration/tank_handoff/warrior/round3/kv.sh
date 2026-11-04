#!/bin/bash
# kv.sh <preset> <name> <extra knobs>: req/<name>_<preset>.json = current request with the APL for its knobs plus extra
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round3; S=${S:?set S to a scratch dir holding knobs_<preset> files and apl/}
p=$1; n=$2; x=$3
$R/opt.sh ${p//_/-} $S/kv.txt $R/req/cur_$p.json -knobs "$(cat $S/knobs_$p),$x" -eval -iters 10 -confirm-iters 10 -out $S/apl -name ${n}_$p
python3 $R/mkvar.py $R/req/cur_$p.json $R/req/${n}_$p.json apl=$S/apl/${n}_$p.apl.json
