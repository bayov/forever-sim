#!/bin/bash
# opt.sh <preset|weight> <out file> <settings> [rotopt flags...]: rotopt -template warrior_tank on the six scoring
# encounters (their mean), 1,000 iterations a candidate and 3,000 to confirm unless the flags say otherwise (round 3's).
# BIN: directory with the rotopt binary (default /tmp/claude-0/warrior_r5).
preset=$1; out=$2; settings=$3; shift 3
E=tools/encounters; BIN=${BIN:-/tmp/claude-0/warrior_r5}
cd /home/user/forever-sim
"$BIN/rotopt" -template warrior_tank -settings "$settings" -tank "$preset" \
  -tank-encounter $E/magmadar.json,$E/golemagg.json,$E/onyxia.json,$E/broodlord.json,$E/vaelastrasz.json,$E/chromaggus.json \
  -iters 1000 -confirm-iters 3000 "$@" > "$out" 2>&1
echo "exit $?" >> "$out"
