#!/bin/bash
# run.sh <seed> <iters> <suffix> <A> <variant>[:<preset>]... : paired abtest, var/<A>_<p>.json vs var/<variant>_<p>.json,
# all three presets unless :<preset> is given; output ab/s<seed>_<variant>_<preset><suffix>.txt
cd /home/user/forever-sim; E=orchestration/tank_handoff/paladin/round4/extras; B=${ABTEST:-/tmp/claude-0/paladin_tank/abtest}
seed=$1; iters=$2; suf=$3; A=$4; shift 4
ENC="tools/encounters/magmadar.json tools/encounters/golemagg.json tools/encounters/onyxia.json tools/encounters/broodlord.json tools/encounters/vaelastrasz.json tools/encounters/chromaggus.json tools/encounters/ragnaros.json"
for vp in "$@"; do v=${vp%%:*}; ps="max_threat balanced max_survival"; [ "$vp" != "$v" ] && ps=${vp##*:}
  for p in $ps; do $B -iters $iters -seed $seed -preset $p $E/var/${A}_$p.json $E/var/${v}_$p.json $ENC > $E/ab/s${seed}_${A}_vs_${v}_${p}${suf}.txt 2>&1; done
done
