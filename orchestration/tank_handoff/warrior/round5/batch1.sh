#!/bin/bash
# batch1.sh : the round-1 candidates at 10,000 iterations, seeds 1 and 2, A = the current request (req/cur_<p>.json).
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round5
for s in 1 2; do
 for v in tal talwall12; do $R/ab.sh o10_${v}_max_threat_s$s 10000 $s max_threat $R/req/cur_max_threat.json $R/req/${v}_max_threat.json; done
 for v in su25 tal talsu25; do $R/ab.sh o10_${v}_balanced_s$s 10000 $s balanced $R/req/cur_balanced.json $R/req/${v}_balanced.json; done
 for v in su25 tal; do $R/ab.sh o10_${v}_max_survival_s$s 10000 $s max_survival $R/req/cur_max_survival.json $R/req/${v}_max_survival.json; done
done
echo done > $R/ab/batch1.done
