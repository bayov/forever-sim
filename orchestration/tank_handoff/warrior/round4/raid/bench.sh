#!/bin/bash
# bench.sh : tankbench -iters 1000 (default seed) on the three final requests, each of the seven encounters -> bench/<request>__<boss>.txt
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round4/raid; B=${TANKBENCH:-/tmp/claude-0/warrior_tank/tankbench}; mkdir -p $R/bench
for p in max_threat balanced max_survival; do for e in broodlord chromaggus golemagg magmadar onyxia ragnaros vaelastrasz; do
 $B -iters 1000 -encounter tools/encounters/$e.json tools/tankbench/req/warrior_protection_$p.json > $R/bench/warrior_protection_${p}__$e.txt 2>&1
done; done
