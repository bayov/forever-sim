#!/bin/bash
# kv.sh <class> <preset> <name> <extra knobs> [talents]: req/<class>_<name>_<preset>.json = the tankbench request with the APL
# rotopt writes for k/<class>_<preset> plus <extra knobs> (later duplicates win), and the talent string if given.
W=/tmp/claude-0/-home-user-forever-sim/16f56354-4609-5851-a77a-5815d0058063/scratchpad/work
c=$1; p=$2; n=$3; x=$4; t=$5
cd /home/user/forever-sim/.claude/worktrees/agent-a8d0dcc3b585ee4a7
src=tools/tankbench/req/warrior_protection_$p.json; [ $c = druid ] && src=tools/tankbench/req/druid_bear_$p.json
$W/opt.sh $c $p $W/apl/kv_${c}_${n}_$p.txt -knobs "$(cat $W/k/${c}_$p)${x:+,$x}" -eval -iters 10 -confirm-iters 10 -out $W/apl -name ${c}_${n}_$p
python3 - $src $W/req/${c}_${n}_$p.json $W/apl/${c}_${n}_$p.apl.json "$t" <<'PY'
import json, sys
src, out, apl, tal = sys.argv[1:5]
d = json.load(open(src)); pl = d['raid']['parties'][0]['players'][0]
pl['rotation'] = json.load(open(apl))
if tal: pl['talentsString'] = tal
json.dump(d, open(out, 'w'), indent=1)
PY
