#!/bin/bash
# run5.sh <beforedir>: before -> after on the written requests (write_final.py). A = <beforedir>/<request>.json (the
# round 4 requests), B = tools/tankbench/req/<request>.json; out5/<request>__final_s<seed>.txt, 10k, seeds 1 and 2.
BIN=/tmp/claude-0/-home-user-forever-sim/9040db86-6de0-5226-9d88-b0db7baa031d/scratchpad/bin
R=orchestration/tank_handoff/shaman/round5
mkdir -p $R/out5
for sd in 1 2; do
  for pre in shaman_warden druid_bear; do
    for p in max_threat balanced max_survival; do
      "$BIN/abtest" -preset $p -iters 10000 -seed $sd "$1/${pre}_$p.json" tools/tankbench/req/${pre}_$p.json 2>&1 \
        | grep -v -e "Running .* iterations" -e "sims finished" > $R/out5/${pre}_${p}__final_s$sd.txt
    done
  done
done
echo DONE
