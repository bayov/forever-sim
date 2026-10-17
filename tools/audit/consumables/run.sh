#!/bin/bash
# usage: AUDIT_TREE=<checkout> run.sh <request.json>...   Runs wowsimcli on each request and prints the average DPS.
# Make requests with: go run --tags=with_db ./tools/rotopt/mksettings -spec X -gear G -apl A -talents T -race R -duration 180 -out f.json
set -e
cd "${AUDIT_TREE:?set AUDIT_TREE to the forever-sim checkout}"
for f in "$@"; do
  f=$(realpath "$f")
  go run --tags=with_db ./cmd/wowsimcli sim --infile "$f" --outfile "${f%.json}_out.json" >/dev/null 2>&1
  python3 -c "import json,sys;d=json.load(open(sys.argv[1]));print(sys.argv[2],round(d['raidMetrics']['dps']['avg'],1))" "${f%.json}_out.json" "$(basename "$f")"
done
