#!/bin/bash
# run3.sh: the combined candidates (gen2.py). out3/<request>__final_s<seed>.txt: A = the round 4 request, B = the
# candidate (before -> after); out3/<request>__<check>_s<seed>.txt: A = the candidate, B = one check. 10k, seeds 1 and 2.
BIN=/tmp/claude-0/-home-user-forever-sim/9040db86-6de0-5226-9d88-b0db7baa031d/scratchpad/bin
R=orchestration/tank_handoff/shaman/round5
REQ=/tmp/claude-0/-home-user-forever-sim/9040db86-6de0-5226-9d88-b0db7baa031d/scratchpad/r5req3
OUT=$R/out3
mkdir -p $OUT
python3 $R/gen2.py $REQ
ab() { "$BIN/abtest" -preset "$1" -iters 10000 -seed "$2" "$3" "$4" 2>&1 | grep -v -e "Running .* iterations" -e "sims finished" > "$5"; }
for sd in 1 2; do
  for pre in shaman_warden druid_bear; do
    for p in max_threat balanced max_survival; do
      ab $p $sd tools/tankbench/req/${pre}_$p.json $REQ/final/${pre}_$p.json $OUT/${pre}_${p}__final_s$sd.txt
      for f in $REQ/loo/$pre/*_$p.json; do
        n=$(basename $f _$p.json)
        ab $p $sd $REQ/final/${pre}_$p.json $f $OUT/${pre}_${p}__${n}_s$sd.txt
      done
    done
  done
done
echo DONE
