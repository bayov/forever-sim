#!/bin/bash
# usage: latest_build.sh  -> prints the newest 1.60.x Forever build wago.tools knows, e.g. 1.60.1.70450
# The Forever beta is listed under several products (wow_classic_beta, wow_cn_beta, wowf), and
# wowf lags, so take the highest 1.60 version across all of them.
set -euo pipefail
curl -sf --max-time 60 https://wago.tools/api/builds | python3 -I -c '
import json, sys
d = json.load(sys.stdin)
v = {b["version"] for bs in d.values() for b in bs if str(b.get("version", "")).startswith("1.60.")}
print(max(v, key=lambda s: tuple(map(int, s.split(".")))))
'
