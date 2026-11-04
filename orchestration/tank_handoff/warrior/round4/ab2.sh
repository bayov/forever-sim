#!/bin/bash
# ab2.sh <preset> <iters> <seed> <variant>...: like ab.sh on the round-4 base (Natural Precision flask + Elemental Sharpening Stone): A = base2/, B = req2/<variant>; out b2_s<seed>_<variant>_<preset>[_10k].txt
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round4
p=$1; it=$2; sd=$3; shift 3; suf=""; [ $it -ge 10000 ] && suf=_10k
for v in "$@"; do
 /tmp/claude-0/bin/abtest -preset $p -iters $it -seed $sd $R/base2/warrior_protection_$p.json $R/req2/${v}_$p.json 2>&1 | grep -v -e "Running .* iterations" -e "sims finished" > $R/b2_s${sd}_${v}_$p$suf.txt
done
