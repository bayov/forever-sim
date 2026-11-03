# Mechanics pass 60 (2026-11-03, master a0b306984)

Range audited: 796495d6f..a0b306984 (9 non-merge commits on watched paths). Client 1.60.1.70707 -> 1.60.1.70753.
Shared sims: `compare_runs`: 0 of 73 builds moved more than 0.5 DPS against the published report.

## Get current

- Client: 1.60.1.70753 is new. Fetched (mechanics set) and diffed against 70707 (`wago.py diff --by-spell`).
- Dev notes thread 2360696: still 8 posts, edit times unchanged (post 8 edited 2026-10-27T16:46:55Z).
- Blue tracker: newest post 2026-10-31 02:17 (Midnight hotfixes); nothing new for Forever since 10-29.
- No `forever-update/*` merge and no daily-update TASKS.md item in range.

## Client 70753

| What | Rows | Verdict |
|---|---|---|
| SpellMisc Attributes_14 0x40 | 1401 spells, nothing else changed on them | No mechanic. |
| Moonkin Aura / Leader of the Pack / Camp Chair / Conviction | aura 52 or 57 -> 290 | Hotfix bake; the sim already gives all crit. Matches. |
| Winter's Chill 12579 | 5 -> 20 | Matches the sim (20% a stack). |
| Berserking, Human Spirit, Ruthless, Endurance, Wind Blessed, Energized, Touch of the Grave | 10 -> 15 sec; 5 -> 10%; 2 -> 1%; crit damage effect removed; 3 -> 2%; 100 -> 50%, 30 -> 15 sec; ICD 25 -> 15 sec | Dev notes post 7 / 8 values the sim already takes. Matches. |
| Read Leyline 1259705 | cast 2000 -> 500 ms | New. Sim casts in 2 sec: mechanics-20261103-1 (low, no preset casts it). |
| Minor Speed, Cultivation, 2H Weapon Spec 20111 threat, stance attributes, Execute Attr0 0x100 | | Movement, profession, threat or flags. No DPS. |
| SpellItemEnchantment | 8214 Defense only | Tank only. |
| CurvePoint / TraitDefinitionEffectPoints | talent-tree rows | Talents auditor. |
| Class rows | Ice Lance (300%, category 19 cd on every rank), Arcane Blast, Frigid Power, Shifting Power, Primal Bite, paladin and druid talents, masks | Class / rank tuning auditors. Mage rows match the sim. |

Open item: a full "70707 + hotfixes vs 70753" diff (to split new rows from baked hotfixes for class rows) was not finished; the hotfix fetch for 70707 spell tables takes over 15 minutes and was cut short.

## Code range

| Commit | What | Verdict |
|---|---|---|
| d7541708d | Ice Lance shares Fire Blast's category 19 cooldown; Fingers of Frost buff 400669 | 70753 puts category 19 / 8000 ms on every Ice Lance rank (tier 1), confirming the 2026-11-02 ruling. 400669 is aura 262, 15 sec. Matches. |
| 407f959a5 | Warrior Blood Craze (Forever triggers: Bloodthirst, hit > 20% max health) | 16487 e1 bp 20 triggers 16488 (6 sec, 2 sec period) on 70753. Healing only, no DPS. |
| b33ad307d | Ranged hunters drop the Goblin Sapper; Affliction is an Orc | Preset choice (consumables-20261101-1), not a mechanic. |
| 82b2b31b5, 42f982678, e41f082eb, fb9ae4b79, 374c437aa | Healing sim, raid damage shares, tank scores | `Share` path and tank-only skip apply only to healing / tank sims. No DPS preset moves (0 of 73). |
| 9530cc9d1 | Practice room preloads | No mechanic. |

## Open findings

New: mechanics-20261103-1. Still open: mechanics-20261031-1 (Efficiency / Lacerate; 19416 unchanged on 70753), mechanics-20261027-2 (Amplify Curse; 18288 unchanged).

## Result

findings (1 new, low).
