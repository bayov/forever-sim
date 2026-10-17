#!/bin/bash
# usage: AUDIT_WORK=<scratch> wago_fetch.sh <build>...   (build = the last number, e.g. 70450)
# Downloads every table in tables.txt for each build to $AUDIT_WORK/db2/<Table>.<build>.csv,
# skipping files already there. About 20 MB a build; ItemSparse and Spell are the big ones.
set -u
: "${AUDIT_WORK:?set AUDIT_WORK}"
D="$AUDIT_WORK/db2"; mkdir -p "$D"
T=$(cat "$(dirname "$0")/tables.txt")
for b in "$@"; do
  for t in $T; do
    [ -s "$D/$t.$b.csv" ] && continue
    for i in 1 2 3 4; do
      code=$(curl -s --max-time 300 -o "$D/$t.$b.csv" -w '%{http_code}' "https://wago.tools/db2/$t/csv?build=1.60.1.$b")
      [ "$code" = 200 ] && break
      rm -f "$D/$t.$b.csv"; sleep $((i*3))
    done
    echo "$b $t $code"
  done
done
