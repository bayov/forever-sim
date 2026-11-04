#!/bin/bash
# run1.sh <seed> <iters> <suffix> <variant>:<preset>... : one preset per variant
cd /home/user/forever-sim
R=orchestration/tank_handoff/paladin/round4; V=${VDIR:-$R/var}; B=${ABTEST:-abtest}
seed=$1; iters=$2; suf=$3; shift 3
ENC="tools/encounters/magmadar.json tools/encounters/golemagg.json tools/encounters/onyxia.json tools/encounters/broodlord.json tools/encounters/vaelastrasz.json tools/encounters/chromaggus.json tools/encounters/ragnaros.json"
for vp in "$@"; do v=${vp%%:*}; p=${vp##*:}
  $B -iters $iters -seed $seed -preset $p $V/base_$p.json $V/${v}_$p.json $ENC > $R/ab/s${seed}_${v}_${p}${suf}.txt 2>&1
done
