#!/bin/bash
# usage: AUDIT_WORK=<scratch> fetch_db2.sh <build e.g. 1.60.1.70450>   Downloads the db2 CSVs the audit uses into $AUDIT_WORK/db2/<build>.
# latest builds: curl -sS https://wago.tools/api/builds | python3 -c "import json,sys;d=json.load(sys.stdin);print(sorted({b['version'] for v in d.values() for b in v if b['version'].startswith('1.60')})[-3:])"
set -e
B=${1:?build}; D="${AUDIT_WORK:?set AUDIT_WORK}/db2/$B"; mkdir -p "$D"
for t in Item ItemSparse ItemEffect ItemXItemEffect SpellEffect SpellName Spell SpellMisc SpellLevels SpellItemEnchantment SpellCategories SpellCategory SpellDuration SpellAuraOptions; do
  [ -s "$D/$t.csv" ] || curl -sS -o "$D/$t.csv" "https://wago.tools/db2/$t/csv?build=$B" -w "$t %{http_code} %{size_download}\n"
done
