#!/bin/bash
# ba.sh <seed> : before (base/ = the requests before this pass) vs after (final/ = the new tools/tankbench/req), 10,000 iterations -> ab/ba10k_s<seed>_<preset>.txt
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round4/raid; B=${ABTEST:-/tmp/claude-0/warrior_tank/abtest}
for p in max_threat balanced max_survival; do
 $B -preset $p -iters 10000 -seed $1 $R/base/warrior_protection_$p.json $R/final/warrior_protection_$p.json 2>&1 | grep -v -e "Running .* iterations" -e "sims finished" > $R/ab/ba10k_s$1_$p.txt
done
