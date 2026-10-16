#!/bin/bash
# usage: wago_db2_fetch.sh <dir>: every table below for every beta build with a balance change
cd "$1"
T="Spell SpellName SpellEffect SpellAuraOptions SpellCooldowns SpellCategories SpellCategory SpellEquippedItems SpellLevels SpellMisc SpellPower SpellClassOptions SpellDuration SpellCastTimes SpellProcsPerMinute SpellProcsPerMinuteMod SpellAuraRestrictions SpellTargetRestrictions SpellInterrupts SpellShapeshift SkillLineAbility TraitDefinition TraitDefinitionEffectPoints TraitNode TraitNodeEntry TraitNodeXTraitNodeEntry TraitCond TraitEdge TraitTree TraitNodeGroup ItemSparse ItemEffect SpellItemEnchantment ChrRaces SpellLabel SpellXSpellVisual SpellRadius SpellRange SpellReagents SpellScaling"
for b in 69893 70009 70170 70291 70334 70338 70374 70425; do
  for t in $T; do
    [ -s "$t.$b.csv" ] && continue
    for i in 1 2 3 4; do
      code=$(curl -s --max-time 300 -o "$t.$b.csv" -w '%{http_code}' "https://wago.tools/db2/$t/csv?build=1.60.1.$b")
      [ "$code" = 200 ] && break
      rm -f "$t.$b.csv"; sleep $((i*3))
    done
    echo "$b $t $code"
  done
done
echo DONE
