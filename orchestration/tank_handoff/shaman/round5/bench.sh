#!/bin/bash
# bench.sh: tankbench (1k) on the six round 5 requests over the seven bosses into bench/<request>__<boss>.txt, then
# bench_summary.py over them into bench/summary.tsv (its druid_bear_* and shaman_warden_* rows go into the ref).
BIN=/tmp/claude-0/-home-user-forever-sim/9040db86-6de0-5226-9d88-b0db7baa031d/scratchpad/bin
R=orchestration/tank_handoff/shaman/round5
mkdir -p $R/bench
for req in shaman_warden_max_threat shaman_warden_balanced shaman_warden_max_survival druid_bear_max_threat druid_bear_balanced druid_bear_max_survival; do
  for b in broodlord chromaggus golemagg magmadar onyxia vaelastrasz ragnaros; do
    "$BIN/tankbench" -iters 1000 -encounter tools/encounters/$b.json tools/tankbench/req/$req.json > $R/bench/${req}__$b.txt 2>&1 &
  done
  wait
done
python3 $R/bench_summary.py $R/bench $R/bench/summary.tsv
echo DONE
