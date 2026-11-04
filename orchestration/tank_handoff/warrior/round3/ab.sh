#!/bin/bash
# ab.sh <out> <iters> <seed> <preset> A.json B.json: paired abtest on the six scoring bosses (B - A).
cd /home/user/forever-sim
out=$1; it=$2; sd=$3; p=$4; shift 4
"$BIN/abtest" -preset $p -iters $it -seed $sd "$@" 2>&1 | grep -v -e "Running .* iterations" -e "sims finished" > "$out"; echo "exit $?" >> "$out"
