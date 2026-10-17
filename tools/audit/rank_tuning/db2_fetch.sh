#!/bin/bash
# Usage: AUDIT_WORK=<scratch dir> db2_fetch.sh <build e.g. 1.60.1.70450> [table ...]
# Downloads the tracked db2 tables as $AUDIT_WORK/db/B_<buildno>_<Table>.csv (default: all 19 tracked).
set -uo pipefail
: "${AUDIT_WORK:?}"; build=$1; shift; no=${build##*.}
T=${*:-"CurvePoint SkillLineAbility SpellCastTimes SpellCooldowns SpellDuration SpellEffect SpellItemEnchantment SpellLevels SpellMisc SpellName SpellPower TraitCond TraitDefinition TraitDefinitionEffectPoints TraitNode TraitNodeEntry TraitNodeGroupXTraitCond TraitNodeGroupXTraitNode TraitNodeXTraitNodeEntry"}
mkdir -p "$AUDIT_WORK/db"
for t in $T; do
  curl -s -f -o "$AUDIT_WORK/db/B_${no}_$t.csv" "https://wago.tools/db2/$t/csv?build=$build" || echo "FAIL $t"
  head -c 3 "$AUDIT_WORK/db/B_${no}_$t.csv" | grep -q '<' && echo "HTML instead of CSV for $t (site down?)"
done
