#!/bin/bash
# Usage: runreq.sh <tools/dps_report/req/<id>.json> <outdir> <iterations>
# Sims one DPS report request as committed (iterations overridden, seed "1" unless the request sets one) and prints
#   <id>  mine <dps>  pub <published.dps from <id>.meta.json next to the request>  diff <mine - pub>
# Env: AUDIT_TREE (checkout the binary runs in; default: this repo), AUDIT_WORK (scratch; default ${TMPDIR:-/tmp}/audit-results),
#      AUDIT_BIN (dir holding wowsimcli; default $AUDIT_WORK/bin, built by build.sh).
set -u
D=$(cd "$(dirname "$0")" && pwd)
AUDIT_TREE=${AUDIT_TREE:-$(cd "$D/../../.." && pwd)}
AUDIT_WORK=${AUDIT_WORK:-${TMPDIR:-/tmp}/audit-results}
AUDIT_BIN=${AUDIT_BIN:-$AUDIT_WORK/bin}
req=$(cd "$(dirname "$1")" && pwd)/$(basename "$1")
id=$(basename "$req" .json); out=$(mkdir -p "$2" && cd "$2" && pwd)
python3 - "$req" "$out/$id.in.json" "$3" <<'P'
import json,sys
d=json.load(open(sys.argv[1])); d.setdefault('simOptions',{}); d['simOptions']['iterations']=int(sys.argv[3]); d['simOptions']['randomSeed']=d['simOptions'].get('randomSeed','1')
json.dump(d,open(sys.argv[2],'w'))
P
(cd "$AUDIT_TREE" && "$AUDIT_BIN/wowsimcli" sim --infile "$out/$id.in.json" --outfile "$out/$id.out.json" >/dev/null 2>&1) || { echo "$id SIMFAIL"; exit; }
python3 - "$out/$id.out.json" "${req%.json}.meta.json" "$id" <<'P'
import json,sys
o=json.load(open(sys.argv[1])); m=json.load(open(sys.argv[2]))
dps=o['raidMetrics']['dps']['avg']; pub=m.get('published',{}).get('dps')
print("%-44s mine %7.1f  pub %7s  diff %+5.1f"%(sys.argv[3],dps,pub,dps-(pub or 0)))
P
