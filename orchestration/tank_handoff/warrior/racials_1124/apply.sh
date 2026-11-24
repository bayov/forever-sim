#!/bin/bash
# apply.sh <class> <preset> [talents]: write the UI APL from k/<class>_<preset> with rotopt and put it (and talents) in the tankbench request
W=/tmp/claude-0/pr
c=$1; p=$2; t=$3
cd /home/user/forever-sim
if [ $c = warrior ]; then ui=ui/tank_warrior/apls/level60_$p.apl.json; req=tools/tankbench/req/warrior_protection_$p.json
else ui=ui/feral_tank_druid/apls/$p.apl.json; req=tools/tankbench/req/druid_bear_$p.json; fi
$W/opt2.sh $c $p $W/apl/final_${c}_$p.txt -knobs "$(cat $W/k/${c}_$p)" -eval -iters 10 -confirm-iters 10 -out $W/apl -name final_${c}_$p
cp $W/apl/final_${c}_$p.apl.json $ui
python3 - $req $ui "$t" <<'PY'
import json, sys
req, ui, tal = sys.argv[1:4]
s = open(req).read(); d = json.loads(s); pl = d['raid']['parties'][0]['players'][0]
pl['rotation'] = json.load(open(ui))
if tal: pl['talentsString'] = tal
o = json.dumps(d, indent=2)
open(req, 'w').write(o + ('\n' if s.endswith('\n') else ''))
PY
