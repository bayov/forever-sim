#!/usr/bin/env bash
# Usage: AUDIT_WORK=<scratch> tools/audit/mechanics/fetch_db2.sh <build, e.g. 1.60.1.70450> [Table ...]
# Downloads wago.tools db2 CSVs into $AUDIT_WORK/db2/<build>/<Table>.csv. With no tables, fetches the
# set the mechanics pass uses. Era reference: build 1.15.9.69722.
set -euo pipefail
build=${1:?build}; shift || true
tables=("$@")
[ ${#tables[@]} -eq 0 ] && tables=(SpellEffect SpellName SpellMisc SpellAuraOptions SpellCooldowns SpellDuration SpellLevels SpellPower SpellCastTimes SpellCategories SpellClassOptions Spell SkillLineAbility ItemEffect ItemXItemEffect ItemSparse SpellItemEnchantment SpellShapeshiftForm SpellShapeshift ChrRaces CharBaseInfo TraitDefinition TraitDefinitionEffectPoints TraitNodeEntry CurvePoint)
out="${AUDIT_WORK:?set AUDIT_WORK}/db2/$build"; mkdir -p "$out"
for t in "${tables[@]}"; do
  curl -sfS -o "$out/$t.csv" "https://wago.tools/db2/$t/csv?build=$build" || echo "failed: $t" >&2
done
echo "$out"
