# Client update pass: 1.60.1.70685 -> 1.60.1.70707

- Pass 6, 2026-10-30, master 36a772fd13853013772a7a48119361c14e979da0. `wago.py latest --probe 300` returns 1.60.1.70707.
- For `.agents/skills/dps-report/state.json`: `foreverClientBuild` -> 1.60.1.70707 (the orchestrator updates it).
- Result: findings. 70707 writes the 10-27 to 10-29 server hotfixes into the client, and that confirms the mage talent rework TASKS.md parked on 2026-10-29 (the user's condition was "a newer client build").

## Client diff

`table_summary.py 70685 70707`: 201 spells changed; Trait* changes in the mage tree (1112) only; ItemSparse +10 / ~43; ItemEffect +11 / ~3; ChrRaces and SpellItemEnchantment identical. `items_diff.py`: no stat change on a level 60 or preset item (new ilvl 10 copies of potions and Molten Core rings, Watcher's Signet renames, consumable flags). `enchant_diff.py`: nothing.

## DPS-relevant rows

| Row | Change | Sim status |
|---|---|---|
| MAGE-U-20 | Critical Mass 5 -> 3 ranks (max 6% Fire crit) | missing (finding -1) |
| MAGE-U-21 | Incineration 3 -> 5 ranks (max 10%) | missing (-1) |
| MAGE-U-22 | Arcane Focus removed | missing (-1) |
| MAGE-U-23 | Precision: 2/4/6% hit with all spells | missing (-1) |
| MAGE-U-24 | Accelerant also on Arcane Wrath | missing (-1) |
| MAGE-U-25 | Ice Lance 8 s cooldown (rank 1 in the client only) | missing / ambiguous (-1) |
| MAGE-U-26 | Tree moves, Missile Barrage and Arcane Resilience swap | missing: mage.json, talent strings (-1) |
| MAGE-U-27 | Missile Barrage's tick-period effect becomes a dummy | ambiguous (finding -2) |
| HUN-U-13 | Efficiency covers Strider Kick | missing (finding -3) |
| SHR-U-16 | Fengus' Ferocity effects become dummies | ambiguous (finding -4) |
| SHR-U-17 | +hit item spells gain spell hit | landed (Forever gear hit is universal) |
| SHR-U-18 | Endurance crit damage mask, Earth Fury weapon rule | not simulated on purpose (10-26 notes, settled) |
| HUN-HF-01/02, SHR-P8-01 | Hawk, Trueshot Aura, Blood Fury now in the client | landed already |

Measured on the fire request (scratch edit, 180 s, 20k iterations, seeds 1 / 1000003):

| Variant | Seed 1 | Seed 1000003 | Delta |
|---|---|---|---|
| master | 1089.03 | 1088.66 | |
| Critical Mass 3 | 1051.59 | 1050.76 | −37.7 (−3.5%) |
| Critical Mass 3, Incineration 5 | 1063.87 | 1062.52 | −25.7 (−2.4%) |

These are before re-spending the freed points (Arcane Focus 4, Critical Mass 2). Arcane is not measured: Precision's 6% on every school replaces Arcane Focus's 5% Arcane hit, so it should gain a little. Frost's Ice Lance cooldown needs a code change to measure. In the client only rank 1 has the cooldown, so it may not touch the rank 6 the APLs cast.

## Sources

| Source | State |
|---|---|
| Blizzard dev notes (topic 2360696) | Not read: forum rate limit again. Next pass: diff against `sources/devnotes_post7_2026-10-26.txt` |
| foreverchanges.pro | Lists 70707 (2026-10-30): the client now has the 10-27 to 10-29 hotfixes (mage rework, Ice Lance cooldown, Blood Fury, Revenge AP), and the tent is cosmetic again. No hotfix newer than the 10-28 one |
| Wowhead | Not checked (blocked) |

## Ledger

Added `## Build 1.60.1.70707 (first seen 2026-10-30)`: one paragraph for the n/a and already-landed changes, 13 rows (MAGE-U-20..27, HUN-U-13, SHR-U-16..18, WL-U-18). Summary counts updated.

Open findings: client_update-20261027-3 (Amplify Curse, parked) has new 70707 evidence; 20261030-1 to -4 are new.
