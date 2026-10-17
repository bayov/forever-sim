#!/bin/bash
# Usage: wago_builds.sh [known build, default 1.60.1.70450]
# Lists the Forever (1.60.x) client builds wago.tools offers, newest last. A build newer than the
# handoff's baseline means a db2 diff is due. wago.tools sometimes returns 503 (Cloudflare): retry later,
# and if it refuses the request, don't reach it any other way.
b=${1:-1.60.1.70450}
curl -s "https://wago.tools/db2/SpellEffect?build=$b" | grep -o '1\.60\.[0-9]\.[0-9]*' | sort -t. -k4 -n -u
