# Client update pass: 1.60.1.70450 -> 1.60.1.70532

- Pass 2026-10-21, master 599f9de74ba71d4ba48ace70af3e118841c88f9e.
- `wago.py latest --probe 300` returns 1.60.1.70532.

## Client diff

`table_summary.py 70450 70532`: all 39 tables in the `client_update` set are **identical**, with no added rows. That includes SpellEffect, Spell, SpellMisc, SpellPower, the cooldown and aura tables, all Trait* tables, ItemSparse, ItemEffect, ItemXItemEffect, SpellItemEnchantment and ChrRaces. Both downloads are 22 MB and SpellEffect.csv is byte-equal (`cmp`). There is no class, racial, talent, item or enchant change in the client.

## Sources

| Source | State |
|---|---|
| Blizzard dev notes (topic 2360696) | No new post, no edit since the post /6 edit of 2026-10-16 (diff empty) |
| foreverchanges.pro | Lists 70450 as the current build, not 70532 yet. It adds three unannounced hotfixes (10-16, 10-18, 10-20) |
| Wowhead | Not checked (blocked) |

## Server-side hotfixes (from foreverchanges, not in the client)

| Ledger row | Change | L60 DPS? | Sim status |
|---|---|---|---|
| PAL-U-HF1 | Champion of the Light back to 20/40/60% Int (sim 10/20/30) | yes | ambiguous (needs verify), finding -1 |
| MAGE-U-HF1 | Piercing Ice removed | yes | ambiguous (needs verify), finding -2; already in TASKS.md |
| MAGE-U-HF2 | Elemental Precision 3 ranks x 2% | yes | ambiguous (needs verify), finding -2 |
| MAGE-U-HF3 | Ice Barrier +10% absorb (+3% Frost damage per FC) | minor | ambiguous, in finding -2 |
| MAGE-U-HF5 | Frost Blast triggers Missile Barrage | minor | ambiguous, in finding -2 |
| MAGE-U-HF4, WL-U-HF1, PRI-U-HF1 | Frost Ward/Warding, Amplify Curse CoEx 20%, Holy Fire r1 53-65 | n/a | n/a |

The 70532 client still has the old values: Piercing Ice node 130502 is in the tree, and Elemental Precision entry 130507 has MaxRanks 5. The client doesn't back these, so per the rules no sim change goes in without verification. Nothing was sized, because the changes are unconfirmed.

## Ledger

Added a `## Build 1.60.1.70532 (2026-10-21)` section with 8 hotfix rows, and updated the Summary counts (726 rows; ambiguous 16 -> 21, n/a 217 -> 220).

For the orchestrator: set `.agents/skills/dps-report/state.json` `foreverClientBuild` to 1.60.1.70532.
