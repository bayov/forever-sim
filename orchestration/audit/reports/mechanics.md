# Mechanics pass 62 (2026-11-03, master f23ea25d2)

Quiet: no findings.

## Range

f5c99ad01..f23ea25d2. One non-test code commit under the watched paths:

| Commit | What | Mechanics check |
|---|---|---|
| 734fe0e93 | Affliction (Wrack) preset sacrifices the Imp, no demon out; Shadow Bolt Affliction keeps the Succubus (`tools/rotopt/mksettings/spec_warlock.go`, `sim/compsearch/pool.go` left-out list) | Imp sacrifice = Burning Shadow 18789: aura 79, misc 32 (Shadow), 15, 7200000 ms on client 70753 (tier 1). `sim/warlock/talents.go` applyDemonicSacrifice gives 15% Shadow for 2 h. Matches. |

`sim/testdata/report_hashes.txt` is test data. f23ea25d2 (TASKS.md) adds the Nov 2 notes' tank / heal values (already
mechanics-20261103-2) and hotfix push 113483's rogue poison / Shank rows (class rows, for the rogue and rank tuning
auditors; not shared mechanics).

## Getting current

| Source | Result |
|---|---|
| Client 1.60.1.70755 | Fetched the mechanics set (27 tables); every file is byte-identical to 70753's, so `wago.py diff --by-spell` is empty. All verdicts hold. |
| Hotfixes (wago.tools, product wowf) | Newest spell push 113574 (70707, region 91, 2026-11-03 18:29): Spell / SpellName text of the NPC Holy Strike ranks only. No 70753 / 70755 spell rows past 113557. |
| Dev notes thread 2360696 | Still 9 posts; post 9 created and edited 2026-11-03T03:20:33Z, post 1 edited 03:20:46Z (as last pass). |
| Blue tracker | One new post: "WoW: Forever—What's New and What Feels Familiar" (2026-11-03 18:00, worldofwarcraft.com/en-us/news/24315821). An overview (presets, dungeons, Skyborne, Legacy points, camping, rulesets); it states no mechanic values. |

## Open items

None new. Open findings unchanged (mechanics-20261103-2, tank threat values).
