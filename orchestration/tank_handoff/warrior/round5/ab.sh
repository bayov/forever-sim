#!/bin/bash
# ab.sh <out name> <iters> <seed> <preset> A.json B.json : paired abtest -preset <preset> (own weight) on the six scoring
# bosses, B - A -> ab/<out name>.txt. ABTEST = the abtest binary (go build -tags=with_db ./tools/audit/tank_rotations/abtest).
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round5; B=${ABTEST:-/tmp/claude-0/warrior_r5/abtest}
out=$1; it=$2; sd=$3; p=$4; shift 4
$B -preset $p -iters $it -seed $sd "$@" 2>&1 | grep -v -e "Running .* iterations" -e "sims finished" > $R/ab/$out.txt
