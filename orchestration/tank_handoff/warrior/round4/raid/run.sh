#!/bin/bash
# run.sh <base dir> <var dir> <seed> <iters> <tag> <variant>[:<preset>]... : paired abtest A = <base>/warrior_protection_<p>.json, B = <var>/<variant>_<p>.json
# (all three presets unless :<preset>); out ab/<tag>_s<seed>_<variant>_<p>.txt. ABTEST = the abtest binary (go build -tags=with_db ./tools/audit/tank_rotations/abtest).
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round4/raid; B=${ABTEST:-/tmp/claude-0/warrior_tank/abtest}
base=$1; var=$2; seed=$3; it=$4; tag=$5; shift 5
for vp in "$@"; do v=${vp%%:*}; ps="max_threat balanced max_survival"; [ "$vp" != "$v" ] && ps=${vp##*:}
 for p in $ps; do $B -preset $p -iters $it -seed $seed $base/warrior_protection_$p.json $var/${v}_$p.json 2>&1 | grep -v -e "Running .* iterations" -e "sims finished" > $R/ab/${tag}_s${seed}_${v}_$p.txt; done
done
