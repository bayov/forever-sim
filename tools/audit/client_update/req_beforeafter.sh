#!/bin/bash
# usage: AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> req_beforeafter.sh <cli> <request-id>...
# Same binary, origin/master's request vs the checkout's request (for preset or request-only
# changes); 180 s, 20k iterations, seeds 1 and 1000003.
set -u
: "${AUDIT_TREE:?set AUDIT_TREE}"; : "${AUDIT_WORK:?set AUDIT_WORK}"
CLI=$1; shift
cd "$AUDIT_TREE"; mkdir -p "$AUDIT_WORK/oldreq"
echo "| Request | Seed | Before | After | Delta |"
echo "|---|---|---|---|---|"
for id in "$@"; do
  git show origin/master:tools/dps_report/req/$id.json > "$AUDIT_WORK/oldreq/$id.json"
  for seed in 1 1000003; do
    b=$(DASH_BIN=$CLI DASH_TREE=$PWD python3 tools/dps_report/verify.py "$AUDIT_WORK/oldreq/$id.json" 20000 $seed | awk '{print $1}')
    a=$(DASH_BIN=$CLI DASH_TREE=$PWD python3 tools/dps_report/verify.py tools/dps_report/req/$id.json 20000 $seed | awk '{print $1}')
    d=$(python3 -c "print(f'{$a-$b:+.2f} ({($a-$b)/$b*100:+.2f}%)')")
    echo "| $id | $seed | $b | $a | $d |"
  done
done
