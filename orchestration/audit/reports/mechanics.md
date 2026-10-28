# Mechanics pass 45 (2026-10-28, master 224b8a4e0)

Range audited: 5d27b0015..224b8a4e0. No commit touches `sim/**/*.go` or `tools/rotopt/mksettings/**`; no TASKS.md / ledger commit, no `forever-update/*` merge.

## Get current

- Client: new build 1.60.1.70658 (from the gate). Fetched the mechanics set for 70634 and 70658 (wago.tools transfers kept getting cut off; took six fetch rounds). All 25 tables byte-identical:
  SpellEffect, SpellName, SpellMisc, SpellAuraOptions, SpellCooldowns, SpellDuration, SpellLevels, SpellPower, SpellCastTimes, SpellCategories, SpellClassOptions, Spell, SkillLineAbility, ItemEffect, ItemXItemEffect, ItemSparse, SpellItemEnchantment, SpellShapeshiftForm, SpellShapeshift, ChrRaces, CharBaseInfo, TraitDefinition, TraitDefinitionEffectPoints, TraitNodeEntry, CurvePoint.
  Sanity: SpellDuration for 70658 answers 200 and a non-build (70657) answers 400, so the 70658 tables are that build's, not a fallback. Every verdict holds on 70658; verdicts.md baseline updated.
- Dev notes thread 2360696: still 8 posts; post 7 (edited 2026-10-27 16:47:14) and post 8 (edited 16:46:55) unchanged.
- Blue posts since 2026-10-28 00:44:50: Kaivax 01:28 (Forever restricted name list), nothing mechanical. Aggrend, Linxy, Gnedi, Brenoch: no Forever posts.
- Server hotfixes: `tools/audit/mechanics/hotfixes.py` is still missing from the tree; not re-read.

## Findings

- No new findings. mechanics-20261027-2 (Amplify Curse vs Bane of Agony) stays open; no warlock commit in range, and 70658 does not change those rows.
