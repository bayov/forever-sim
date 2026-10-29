# Client update pass: 1.60.1.70658 -> 1.60.1.70685

- Pass 5, 2026-10-29, master dcd92749092615f781a700d8b5dc24b7cd853895. `wago.py latest --probe 300` returns 1.60.1.70685.
- For `.agents/skills/dps-report/state.json`: `foreverClientBuild` -> 1.60.1.70685 (the orchestrator updates it).
- Result: quiet. Nothing DPS-relevant changed; no findings.

## Client diff

`table_summary.py 70658 70685`: Trait*, SpellPower, SpellClassOptions, SpellAuraOptions, SkillLineAbility, ItemSparse, SpellItemEnchantment and ChrRaces identical. `db2diff.py` finds two spells only:

| Spell | Change | L60 DPS |
|---|---|---|
| Adventurer's Tent (1325145) | cosmetic tent summon -> rested-XP tent (party nearby, rested XP up to 20% of a level, needs a campfire), 1 h category cooldown (new SpellCategory 2607); ItemEffect 238896 item cooldown 900000 -> -1 | n/a (leveling) |
| Camping (1328862) | new helper spell | n/a |

`items_diff.py`: 0 added, 0 removed, 0 changed items.

## Sources

| Source | State |
|---|---|
| Blizzard dev notes (topic 2360696) | Not read: forum rate limit ("performed this action too many times"). Next pass: diff against `sources/devnotes_post7_2026-10-26.txt` |
| foreverchanges.pro | Does not list 70685 yet; newest hotfix is still the 10-28 one (HUN-HF-01/02, landed) |
| Wowhead | Not checked (blocked) |

## Ledger

Added `## Build 1.60.1.70685 (first seen 2026-10-29)`: one paragraph, no rows.

Open findings: client_update-20261027-3 (Amplify Curse, parked for an in-game test) unchanged.
