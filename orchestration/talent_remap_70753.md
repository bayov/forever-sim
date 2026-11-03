# Talent remap for client 1.60.1.70753

Written 2026-11-03 by the sources pass (agent A1).

## Result: nothing to remap

`tools/forever_talents/import_db2.py 1.60.1.70753` finds no structure, number or wording change from the trees the sim already has, in any of the nine classes. So no talent moved, no talent was added or removed, and no rank count changed between the sim's trees and client 70753. Every preset talent string in `sim/<class>/**`, `ui/<class>/**/presets.ts`, `tools/rotopt`, `tools/dps_report/req/*.json` and the test fixtures compiles and spends the same points in the same talents as before. No preset lost a point.

| Class | Tree source before | Tree source now | Remapped | Points dropped |
|---|---|---|---|---|
| Druid | 1.60.1.70707+hf | 1.60.1.70753 (same tree) | none | none |
| Hunter | 1.60.1.70707+hf | 1.60.1.70753 (same tree) | none | none |
| Mage | 1.60.1.70707+hf | 1.60.1.70753 (same tree) | none | none |
| Paladin | 1.60.1.70707+hf | 1.60.1.70753 (same tree) | none | none |
| Priest | 1.60.1.70707+hf | 1.60.1.70753 (same tree) | none | none |
| Rogue | 1.60.1.70630 | 1.60.1.70753 (same tree) | none | none |
| Shaman | 1.60.1.70630 | 1.60.1.70753 (same tree) | none | none |
| Warlock | 1.60.1.70707+hf | 1.60.1.70753 (same tree) | none | none |
| Warrior | 1.60.1.70630 | 1.60.1.70753 (same tree) | none | none |

Why: the November 2 development notes describe two kinds of talent change.

- Paladin, druid, hunter, mage, priest and warlock: the reworks the notes list (Heart of the Wild and Furor swapped, Lacerate and Predatory Discipline on row 4, Bloom, Oath of Light, Boundless Light, Shieldsworn, Stoicism, Turn Evil, Arcane Focus and Piercing Ice out, the hunter tree's remap, and the rest) reached the realms as the 2026-10-30/31 hotfix. The sim took them on 2026-11-02 as `1.60.1.70707+hf`, and its presets were remapped then. Client 70753 bakes those rows into its tables, so its trees equal what we have.
- Rogue, warrior and shaman: the reworks the notes list are **not** in client 70753. Its tables have no Death Mark, Boot Knife, Blood Splatter, Composure, Flash of Steel, Quick Getaway, Reckless Swings or Improved Whirlwind spell. Murder, Quietus, Cutthroat, Thousand Cuts, Improved Distract and Lava Burst are still in the trees, and Lethality, Relentless Strikes, Dual Wield Specialization, Camouflage, Opportunity, Improved Tactical Mastery, Bloodthrill and Flurry keep their old rank counts. wago.tools had no hotfix rows for build 70753 at all on 2026-11-03 13:00Z, and foreverchanges.pro's "hotfix to the November 2 build" covers only the hunter. These trees can't be imported yet. Building them by hand from the notes would guess the columns, since the notes give rows only.

## Per class

- Tree rework not in client 70753: rogue, left to the class worker.
- Tree rework not in client 70753: warrior, left to the class worker.
- Tree rework not in client 70753: shaman, left to the class worker.
- Druid, hunter, mage, paladin, priest, warlock: client 70753 has the notes' tree changes (taken as 70707+hf on 2026-11-02). Ported, and nothing to remap.

## Adding notes-only talents: `tools/forever_talents/notes/<class>.json`

Under standards 37, a class worker can build a rework from the notes. It goes in `tools/forever_talents/notes/<class>.json`, and `import_db2.py` lays that file over the tree it generates from the client on every run, so a later client import doesn't wipe it. The ops are `add` (a whole talent in the `client/` shape, with row, col, max, desc per rank and req), `remove` (the talent goes to the tree's `removed` list) and `set` (change named fields; moving a talent is a `set` of row and col). Each entry has a `source` naming the announcement and saying what was guessed. When the client catches up, the importer prints `... the entry can go` for each entry the client now matches, and the client's version wins for an `add`. Two talents on one square stop the import. The format and the regeneration chain (`import_db2.py <build> <class> --write`, `import_client.py <class>`, `import_talents.py <class> --write`, then the proto message and the preset remap) are in `tools/forever_talents/notes/README.md`. No notes files exist yet.

## When the rogue, warrior and shaman trees arrive

Run `tools/sources/hotfixes.py fetch 1.60.1.70753 --set talents --set spelltext`, then `tools/forever_talents/import_db2.py 1.60.1.70753+hf rogue warrior shaman`, or use the next client build. These presets spend points that the notes remove or re-rank. The list below comes from scanning the repo's talent strings, so it may miss strings stored in other formats:

| Class | Talent string | Points in talents the notes remove or re-rank | Where |
|---|---|---|---|
| Rogue | `0053031033150103-302303202015-002` | Relentless Strikes 1, Lethality 3, Seal Fate 3, Dual Wield Specialization 5, Opportunity 2 | `sim/rogue/dps_rogue/mutilate_forever_test.go`, `tools/dps_report/req/rogue__mutilate.json`, `tools/dps_report/req/rogue__mutilate.meta.json`, `ui/rogue/presets.ts` |
| Rogue | `0053031035140004--3021203310113011` | Relentless Strikes 1, Lethality 3, Seal Fate 4, Camouflage 3, Opportunity 2 | `sim/rogue/dps_rogue/forever_mechanics_test.go`, `tools/dps_report/req/rogue__subtlety.json`, `tools/dps_report/req/rogue__subtlety.meta.json`, `ui/rogue/presets.ts` |
| Rogue | `005303104-30230320201515231-002` | Relentless Strikes 1, Lethality 4, Dual Wield Specialization 5, Opportunity 2 | `tools/dps_report/req/rogue__combat-daggers.json`, `tools/dps_report/req/rogue__combat-daggers.meta.json`, `ui/rogue/presets.ts` |
| Rogue | `0053031052150005-322303200013-002` | Relentless Strikes 1, Lethality 5, Seal Fate 5, Dual Wield Specialization 3, Opportunity 2 | `tools/dps_report/req/rogue__seal-fate-daggers.json`, `tools/dps_report/req/rogue__seal-fate-daggers.meta.json`, `ui/rogue/presets.ts` |
| Rogue | `0053231013150103-302303202015-002` | Murder 2, Relentless Strikes 1, Lethality 1, Seal Fate 3, Dual Wield Specialization 5, Opportunity 2 | `tools/dps_report/req/rogue__mutilate-humanoid.json`, `tools/dps_report/req/rogue__mutilate-humanoid.meta.json`, `ui/rogue/presets.ts` |
| Rogue | `0053231015140004--3021203310113011` | Murder 2, Relentless Strikes 1, Lethality 1, Seal Fate 4, Camouflage 3, Opportunity 2 | `tools/dps_report/req/rogue__subtlety-humanoid.json`, `tools/dps_report/req/rogue__subtlety-humanoid.meta.json`, `ui/rogue/presets.ts` |
| Rogue | `005323102-30230320201515231-002` | Murder 2, Relentless Strikes 1, Lethality 2, Dual Wield Specialization 5, Opportunity 2 | `sim/rogue/dps_rogue/dps_rogue_test.go`, `tools/dps_report/req/rogue__combat-daggers-humanoid.json`, `tools/dps_report/req/rogue__combat-daggers-humanoid.meta.json`, `ui/rogue/presets.ts` |
| Rogue | `0053231053120005-322303200013-002` | Murder 2, Relentless Strikes 1, Lethality 5, Seal Fate 5, Dual Wield Specialization 3, Opportunity 2 | `tools/dps_report/req/rogue__seal-fate-daggers-humanoid.json`, `tools/dps_report/req/rogue__seal-fate-daggers-humanoid.meta.json` |
| Warrior | `20305213132515201-3505-2` | Improved Tactical Mastery 5, Bloodthrill 5 | `tools/dps_report/req/warrior__arms_dragonkin.json`, `tools/dps_report/req/warrior__arms_dragonkin.meta.json`, `ui/warrior/presets.ts` |
| Shaman | `5501331003001351-04-053150001001` | Lava Burst 1 | `tools/dps_report/req/shaman__elemental-level60-mana-tide.json`, `tools/dps_report/req/shaman__elemental-level60-mana-tide.meta.json`, `ui/elemental_shaman/presets.ts` |
| Shaman | `5501331003001351-055000101-053` | Lava Burst 1 | `sim/shaman/elemental/elemental_test.go`, `tools/dps_report/req/shaman__elemental-level60.json`, `tools/dps_report/req/shaman__elemental-level60.meta.json`, `ui/elemental_shaman/presets.ts` |
| Shaman | `550533100000135--553302001` | Lava Burst 1 | `sim/shaman/elemental/elemental_test.go` |

What the notes change for these talents: Murder, Quietus, Cutthroat, Thousand Cuts and Improved Distract are removed. Lethality goes from 5 ranks to 2, Relentless Strikes from 1 to 5, Seal Fate to 3, Dual Wield Specialization (rogue) from 5 to 2, Camouflage from 5 to 3 and Opportunity from 2 to 5. On the warrior, Improved Tactical Mastery goes from 5 to 3, Bloodthrill to 2 and Flurry from 5 to 3. On the shaman, Lava Burst is replaced by Call of Thunder. Many other talents only move row, which changes the string positions but not the points.
