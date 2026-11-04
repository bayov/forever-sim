#!/bin/bash
# final.sh : paired abtest before (round 4's requests) vs f (round 5's), 10,000, seeds 1 and 2, own weight; and bench.py
# on the before and after requests at 10,000 on both seeds for the absolute own-weight means.
cd /home/user/forever-sim; R=orchestration/tank_handoff/paladin/round5; TB=/tmp/claude-0/paladin_r5/tankbench
for s in 1 2; do
  $R/run.sh $s 10000 before f
  python3 $R/bench.py $TB 10000 $s $R/req_before/paladin_protection_{max_threat,balanced,max_survival}.json > $R/before_10k_seed$s.txt
  python3 $R/bench.py $TB 10000 $s tools/tankbench/req/paladin_protection_{max_threat,balanced,max_survival}.json > $R/after_10k_seed$s.txt
done
