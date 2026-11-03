# Client 1.60.1.70755

Phase 3 fix round, item 1 (`briefs/phase3_fix.md`). Checked 2026-11-03 with the `update-sources` skill, step 2 and 6.

## What 70755 changed

Nothing in the game data the sim reads. wago.tools lists 1.60.1.70755 (product `wowf`, created 2026-11-03 18:16 UTC,
its own build config), and serves its tables (a build it doesn't know answers 400, so these are 70755's own rows).

- `table_summary.py 1.60.1.70753 1.60.1.70755`: all 39 tables of the `client_update` set identical (Spell, SpellName,
  SpellEffect, SpellAuraOptions, SpellCooldowns, SpellCategories, SpellCategory, SpellEquippedItems, SpellLevels,
  SpellMisc, SpellPower, SpellClassOptions, SpellDuration, SpellCastTimes, SpellProcsPerMinute, the restriction and
  interrupt tables, SpellShapeshift, SkillLineAbility, every Trait table, ItemSparse, ItemEffect, ItemXItemEffect,
  SpellItemEnchantment, ChrRaces, SpellLabel, SpellXSpellVisual, SpellRadius, SpellRange, SpellReagents). The cached
  CSVs are byte-identical.
- The talent importer's other tables (TalentTab, SkillLine, ChrClasses, CurvePoint, SpellDescriptionVariables,
  SpellXDescriptionVariables, SpellShapeshiftForm, ItemSubClass, ManifestInterfaceData) are byte-identical too.
- `db2diff.py`: 0 spells changed. Its one lookup line, "SpellScaling missing in one build", is wago.tools serving no
  SpellScaling for either build; nothing reads it.
- `import_db2.py 1.60.1.70755`: 0 structure, 0 numbers, 0 wording for all nine classes; its printout is the same
  as for 70753, line for line.

So 70755 is a client rebuild with no data change (likely the launch build, 2026-11-04).

## What I did with each line

| Line | Done |
| --- | --- |
| Spell and item tables | Nothing to port. `itemsparse.py`, `enchants.py`, `racials.py --write` and `abilities.py` rerun on 70755: only their `meta` build and fetch time moved. |
| Notes-only values (standards 37), every class file: druid, hunter, mage, paladin, rogue, shaman, warlock, warrior, shared | None arrives: 70755's spell tables are 70753's, which the round already found without them. Every value stays notes-only and gated under Forever; the lists are unchanged. Their "client 1.60.1.70753 unchanged" citations hold for 70755 as well. |
| Rogue, warrior and shaman talent reworks | Not in 70755: the Trait tables are 70753's, the rogue tree is still the old 53 nodes. The three classes stay on `tools/forever_talents/notes/<class>.json`; no preset string changes, so none can become illegal. |
| Talent trees, other classes | `import_db2.py 1.60.1.70755 --write` rewrote `tools/forever_talents/client/<class>.json`: the 70755 build stamp, plus the notes lines the importer lays over them from `tools/forever_talents/notes/<class>.json`, which the round's class sessions added or extended after these files were last written (rogue's removed talents, shaman's Lava Burst, the druid, warlock and warrior notes text). `import_client.py` on the new files gives the same `data/<class>.json` content (only `generatedAt` and the source line differ, left out), so the sim's trees and talent strings don't move. |
| Item database | `gen_db -gen=db` regenerated: no item changed. The `spells` icon list follows the rewritten client trees (Murder, the old Improved Eviscerate and Improved Sinister Strike ranks out; Sword Specialization, Relentless Strikes Effect, Sticky Tar, Focused Rage in). UI icons only. |
| `foreverClientBuild` in `.agents/skills/dps-report/state.json` | Left at 1.60.1.70753 for the orchestrator (`passes/client_update.md`); 70755 adds nothing to audit against. |
| Hotfix caches (`dbcache.py`), Wowhead, foreverchanges suffixes | Not part of this item; `dbcache.py` needs the local client. |

## DPS

No Level 60 report build moves: `sim/testdata/report_hashes.txt` is unchanged (`TestRaidForeverReportHashesStable`
passes on the new data).
