#!/bin/bash
# Usage: sim.sh <spec> <gear> <apl> <talents> <race> <duration> <iterations> <tag> [extra mksettings args]
# Builds a request with mksettings from AUDIT_TREE's presets, sims it and prints summ.py's breakdown. With iterations <= 2 it
# turns on the combat log (debug). For hypothesis tests on a single build; the pass itself uses runreq.sh on the committed requests.
# Output: $AUDIT_WORK/runs/<tag>.json and <tag>.out.json.
# Env: AUDIT_TREE (default: this repo), AUDIT_WORK (default ${TMPDIR:-/tmp}/audit-results), AUDIT_BIN (default $AUDIT_WORK/bin).
D=$(cd "$(dirname "$0")" && pwd)
AUDIT_TREE=${AUDIT_TREE:-$(cd "$D/../../.." && pwd)}
AUDIT_WORK=${AUDIT_WORK:-${TMPDIR:-/tmp}/audit-results}
AUDIT_BIN=${AUDIT_BIN:-$AUDIT_WORK/bin}
R=$AUDIT_WORK/runs; mkdir -p $R
cd "$AUDIT_TREE" || exit 1
"$AUDIT_BIN/mksettings" -spec $1 -gear $2 -apl $3 -talents "$4" -race $5 -duration $6 -out $R/$8.json "${@:9}" || exit 1
python3 - $R/$8.json $7 <<'P'
import json,sys
f,it=sys.argv[1],int(sys.argv[2]); d=json.load(open(f)); d['simOptions']=d.get('simOptions',{}); d['simOptions']['iterations']=it
if it<=2: d['simOptions']['debug']=True
json.dump(d,open(f,'w'))
P
"$AUDIT_BIN/wowsimcli" sim --infile $R/$8.json --outfile $R/$8.out.json >/dev/null 2>&1 || echo SIMFAIL
python3 "$D/summ.py" $R/$8.out.json
