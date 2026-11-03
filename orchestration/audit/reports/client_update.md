# Client updates: pass 9 (2026-11-03, master f23ea25d2)

## Sources

| Source | Result |
|---|---|
| wago.tools (`wago.py latest --probe 300`) | 1.60.1.70755 |
| `table_summary.py 70753 70755` | all 39 tables identical: a technical build |
| Dev notes thread 2360696 | posts /1-/9, nothing after /9 (2026-11-03T03:20Z) |
| foreverchanges.pro | lists 70755 as "a technical build, game data the same as 70753"; its newest hotfix is still the 11-03 hunter one (HUN-HF-03..10, all landed) |
| wago.tools/hotfixes push 113483 | new, 2026-11-03 18:29Z, region 91, 1418 rows, rogue: finding client_update-20261103-3 (high) |

## Conclusions

- 70755 changes nothing DPS-relevant in the client. The ledger's 70755 section is added; `foreverClientBuild` should become 1.60.1.70755 (for the orchestrator).
- Server hotfix push 113483 (rogue tree: Improved Distract removed and two new nodes; per TASKS.md also poisons +33-60%, Mutilate's flat bonus, Shank and Flash of Steel) isn't read in full. The pass was stopped during the `hotfixes.py fetch --refresh` of the 70707 Spell/Trait rows. The next step is in the finding.
- Open findings 20261103-2, 20261030-2 and 20261030-4: no change on master since pass 8.
