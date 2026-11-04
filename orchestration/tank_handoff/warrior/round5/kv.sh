#!/bin/bash
# kv.sh <preset> <name> <extra knobs> [talents] : req/<name>_<preset>.json = req/warrior_protection_<preset>.json with the APL
# rotopt writes for knobs/knobs_<preset> plus <extra knobs> (later duplicates win), and the talent string if given.
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round5
p=$1; n=$2; x=$3; t=$4
$R/opt.sh ${p//_/-} $R/apl/kv.txt $R/req/warrior_protection_$p.json -knobs "$(cat $R/knobs/knobs_$p)${x:+,$x}" -eval -iters 10 -confirm-iters 10 -out $R/apl -name ${n}_$p
python3 - $R/req/warrior_protection_$p.json $R/req/${n}_$p.json $R/apl/${n}_$p.apl.json "$t" <<'PY'
import json, sys
src, out, apl, tal = sys.argv[1:5]
d = json.load(open(src)); pl = d['raid']['parties'][0]['players'][0]
pl['rotation'] = json.load(open(apl))
if tal: pl['talentsString'] = tal
json.dump(d, open(out, 'w'), indent=1)
PY
