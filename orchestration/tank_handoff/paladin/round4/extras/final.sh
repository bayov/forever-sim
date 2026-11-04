#!/bin/bash
# final.sh : the finals. Paired abtest before (var/base_*) vs after (after/), 10,000 iterations, seeds 1 and 2, and
# tankbench (../bench.py) on before/ and after/ at 10,000 on both seeds for the absolute own-weight means.
cd /home/user/forever-sim; E=orchestration/tank_handoff/paladin/round4/extras
TB=/tmp/claude-0/paladin_tank/tankbench
for s in 1 2; do
  ./$E/run.sh $s 10000 _10k base f_sc:max_threat f_sp:balanced f_sp:max_survival
  for w in before after; do python3 $E/../bench.py $TB 10000 $s $E/$w/paladin_protection_{max_threat,balanced,max_survival}.json > $E/${w}_10k_seed$s.txt; done
done
