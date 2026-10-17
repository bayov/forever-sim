#!/bin/bash
# Usage: AUDIT_TREE=<sim checkout> AUDIT_WORK=<scratch dir> run2.sh <req name> <tag> [VAR=value ...]
# Sims tools/dps_report/req/<name>.json from AUDIT_TREE at 20k iterations with the given env vars
# (tag "base" = no vars; tag "prep" only writes the 20k input). Output: $AUDIT_WORK/m2_<name>_<tag>.out.json
set -uo pipefail
: "${AUDIT_TREE:?}"; : "${AUDIT_WORK:?}"
n=$1; tag=$2; shift 2
R="$AUDIT_TREE/tools/dps_report/req"; W="$AUDIT_WORK"
[ -f "$W/m2_$n.in.json" ] || python3 -c "
import json
d=json.load(open('$R/$n.json'))
def f(o):
  if isinstance(o,dict):
    for k in o:
      if k=='iterations': o[k]=20000
      else: f(o[k])
  elif isinstance(o,list):
    for x in o: f(x)
f(d); json.dump(d,open('$W/m2_$n.in.json','w'))"
[ "$tag" = prep ] && exit 0
env "$@" "$W/cli_master" sim --infile "$W/m2_$n.in.json" --outfile "$W/m2_${n}_$tag.out.json" >/dev/null 2>&1
python3 -c "import json;d=json.load(open('$W/m2_${n}_$tag.out.json'));print('$n $tag', round(d['raidMetrics']['dps']['avg'],1) if not d.get('error') else 'ERROR '+d['error']['message'][:80])"
