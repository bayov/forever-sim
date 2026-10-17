# Mechanics audit: local data (not committed)

The pass reads wago.tools db2 CSVs. They are large (SpellEffect alone is about 7 MB a build), so they stay outside the repo. Re-create them with the scripts in `tools/audit/mechanics/`.

| Build | What it is | Tables kept |
|---|---|---|
| 1.60.1.70450 | Current Forever beta client (2026-10-16). The spell data equals 70425's apart from 22 cosmetic rows. | The default set in `fetch_db2.sh` |
| 1.60.1.70425 | The last balance build (2026-10-16). The sim's racials, raid buffs and debuffs follow it. | The default set |
| 1.60.1.70338 | The previous reference. Many code comments cite it. Spell data equals 70374's. | The default set |
| 1.15.9.69722 | Classic Era, for Era-vs-Forever comparisons (cmp.py). | SpellEffect, SpellName, SpellMisc, SpellLevels |

## Re-create

```sh
export AUDIT_WORK=$HOME/audit-work            # any scratch dir
for b in 1.60.1.70450 1.60.1.70425 1.60.1.70338; do tools/audit/mechanics/fetch_db2.sh $b; done
tools/audit/mechanics/fetch_db2.sh 1.15.9.69722 SpellEffect SpellName SpellMisc SpellLevels
export AUDIT_DB=$AUDIT_WORK/db2/1.60.1.70450 AUDIT_ERA_DB=$AUDIT_WORK/db2/1.15.9.69722
```

Enum names for aura types and attributes come from TrinityCore's headers. Fetch them when you need a name:

```sh
mkdir -p $AUDIT_WORK/ref
curl -sS -o $AUDIT_WORK/ref/SpellAuraDefines.h https://raw.githubusercontent.com/TrinityCore/TrinityCore/master/src/server/game/Spells/Auras/SpellAuraDefines.h
curl -sS -o $AUDIT_WORK/ref/SharedDefines.h  https://raw.githubusercontent.com/TrinityCore/TrinityCore/master/src/server/game/Miscellaneous/SharedDefines.h
```

## Column notes

- Forever's SpellEffect stores the real value in `EffectBasePointsF`. Era uses EffectBasePoints + DieSides.
- SpellLevels' column order differs between Forever and Era. Read the columns by name.
- SpellItemEnchantment renamed four columns in 70425 (`Field_12_1_5_*` became `Condition_ID` / `ItemSquishEraID`). Skip them when diffing.
