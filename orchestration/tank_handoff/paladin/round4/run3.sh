#!/bin/bash
cd /home/user/forever-sim; R=orchestration/tank_handoff/paladin/round4; V=$R/var3
ENC="tools/encounters/magmadar.json tools/encounters/golemagg.json tools/encounters/onyxia.json tools/encounters/broodlord.json tools/encounters/vaelastrasz.json tools/encounters/chromaggus.json tools/encounters/ragnaros.json"
seed=$1; iters=$2; suf=$3; shift 3
for v in "$@"; do $ABTEST -iters $iters -seed $seed -preset max_survival $V/base_max_survival.json $V/${v}_max_survival.json $ENC > $R/ab/s${seed}_r3_${v}_max_survival${suf}.txt 2>&1; done
