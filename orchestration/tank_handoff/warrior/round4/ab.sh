#!/bin/bash
# ab.sh <preset> <iters> <seed> <variant>...: paired abtest, A = the request, B = the request with the variant; out s<seed>_<variant>_<preset>[_10k].txt
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round4
p=$1; it=$2; sd=$3; shift 3; suf=""; [ $it -ge 10000 ] && suf=_10k
for v in "$@"; do
 /tmp/claude-0/bin/abtest -preset $p -iters $it -seed $sd tools/tankbench/req/warrior_protection_$p.json $R/req/${v}_$p.json 2>&1 | grep -v -e "Running .* iterations" -e "sims finished" > $R/s${sd}_${v}_$p$suf.txt
done
