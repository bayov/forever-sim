#!/usr/bin/env bash
# Usage: tools/audit/mechanics/probe_builds.sh <first n> <last n>   (e.g. 70451 70750; takes a few minutes, run it in the background)
# The wago.tools builds API lags for Forever beta builds (it never listed 70425 or 70450), so this asks for
# SpellName at each 1.60.1.<n> and prints the builds that answer 200. Then fetch_db2.sh the new build and bdiff2.py it.
# (tools/audit/wago_latest.py --probe N does the same 20 at a time; the audit loop's gate runs it.)
set -u
for n in $(seq "${1:?first}" "${2:?last}"); do
  c=$(curl -s -o /dev/null -w '%{http_code}' --max-time 4 "https://wago.tools/db2/SpellName/csv?build=1.60.1.$n")
  [ "$c" = 200 ] && echo "1.60.1.$n"
done
curl -s https://wago.tools/api/builds | python3 -c "import json,sys;d=json.load(sys.stdin);[print('api',k,[b['version'] for b in v[:3]]) for k,v in d.items() if any(b['version'].startswith('1.60') for b in v[:3])]"
