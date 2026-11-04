#!/bin/bash
# run.sh <seed> <iters> <A> <variant>[:<preset>]... : paired abtest var/<A>_<p>.json vs var/<variant>_<p>.json at the
# preset's own weight, all three presets unless :<preset> is given; output ab/s<seed>_<A>_vs_<variant>_<preset>_<iters>.txt
cd /home/user/forever-sim; R=orchestration/tank_handoff/paladin/round5; B=/tmp/claude-0/paladin_r5/abtest
seed=$1; iters=$2; A=$3; shift 3
ENC="tools/encounters/magmadar.json tools/encounters/golemagg.json tools/encounters/onyxia.json tools/encounters/broodlord.json tools/encounters/vaelastrasz.json tools/encounters/chromaggus.json tools/encounters/ragnaros.json"
for vp in "$@"; do v=${vp%%:*}; ps="max_threat balanced max_survival"; [ "$vp" != "$v" ] && ps=${vp##*:}
  for p in $ps; do [ -f $R/var/${v}_$p.json ] || continue
    $B -iters $iters -seed $seed -preset $p $R/var/${A}_$p.json $R/var/${v}_$p.json $ENC > $R/ab/s${seed}_${A}_vs_${v}_${p}_${iters}.txt 2>&1; done
done
