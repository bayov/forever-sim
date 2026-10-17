#!/bin/bash
# usage: AUDIT_TREE=<checkout> dps_beforeafter.sh <before-cli> <after-cli> <request-id>...
# DPS of tools/dps_report/req/<id>.json at 180 s, 20k iterations, seeds 1 and 1000003, with two
# wowsimcli binaries (build each with: go build --tags=with_db -o <out> ./cmd/wowsimcli after
# regenerating protos). Prints a markdown table for the PR body.
set -u
: "${AUDIT_TREE:?set AUDIT_TREE}"
BEFORE=$1; AFTER=$2; shift 2
cd "$AUDIT_TREE"
echo "| Request | Seed | Before | After | Delta |"
echo "|---|---|---|---|---|"
for id in "$@"; do
  for seed in 1 1000003; do
    b=$(DASH_BIN=$BEFORE DASH_TREE=$PWD python3 tools/dps_report/verify.py tools/dps_report/req/$id.json 20000 $seed | awk '{print $1}')
    a=$(DASH_BIN=$AFTER DASH_TREE=$PWD python3 tools/dps_report/verify.py tools/dps_report/req/$id.json 20000 $seed | awk '{print $1}')
    d=$(python3 -c "print(f'{$a-$b:+.2f} ({($a-$b)/$b*100:+.2f}%)')")
    echo "| $id | $seed | $b | $a | $d |"
  done
done
