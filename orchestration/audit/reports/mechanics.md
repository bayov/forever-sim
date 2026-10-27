# Mechanics pass 41 (2026-10-27, master d3df9247f)

Range audited: 99bd53ada..d3df9247f. No commit touched `sim/**/*.go`, mksettings, TASKS.md or the update ledger, and no `forever-update/*` merge landed.

## New client build 1.60.1.70630

Fetched the mechanics set and diffed it against 70532 (`wago.py diff --by-spell`, 111 spell rows). 20 of 25 tables moved; SpellItemEnchantment, SpellCastTimes, ChrRaces, CharBaseInfo and TraitDefinition are identical.

| Area | Change in 70630 | Sim | Verdict |
|---|---|---|---|
| Spellstone (1237162/4/5) | Misc 36 -> 32, Shadow only (was a hotfix) | Fire + Shadow | open, mechanics-20261025-1 |
| Touch of the Grave drain | hotfix attributes baked in; Attributes_6 0x20000000 kept | no caster mods | holds |
| Racials, raid buffs, debuffs, consumes, enchants | no row changed | - | hold |
| Revelation buff 1248808 | mask + Arcane Wrath, Arcane Barrage | all direct casts | holds |
| Amplify Curse 18288 | e0 mask drops Bane of Agony; e1 -20 on Curse of Exhaustion | Bane x1.5 | finding mechanics-20261027-2 (low) |
| Elemental Mastery 16166 | COOLDOWN_ON_EVENT bit removed | 3 min from cast | now matches |
| Elemental Focus 16164 | ProcChance 10 -> 100, base points 10 | 10% | assumed same, unverified |
| Lava Burst rank 3 dummy | per-level 4.2 -> 0 | flat 20% Flame Shock bonus | holds |
| Hunter's Mark, Bloodthrill, mage Frost rows | hotfixes baked in | ported earlier | hold |
| Talent tables | Frost Warding node 3 ranks, 5 curve rows (the 10-20 mage hotfix) | ported (TASKS 119) | hold |
| Others | Holy Fire R1, Frost Blast coef, Dark Inspiration aura, Feral Swiftness, CC and attribute-only rows | not used at 60 or not simulated | no action |

## Blizzard

Dev notes thread has a new post 7 (Kaivax, 2026-10-26 20:35, edited 21:09): a racial pass before launch. Human Ruthless 1%, Human Spirit 10%; Dwarf Earth Fury removed, Artisan's Precision +1% hit; Tauren Endurance only 5% health; Orc Savage removed, Command +5% pet damage back; Skyborne haste 2%, plus Read Leyline / Skysight utility changes. Client 70630 doesn't have these rows yet and master doesn't model them: finding mechanics-20261027-1 (high, orchestrator fix). The DPS sizes are estimates because measuring needs a code change.

## Notes for the next pass

- `tools/audit/mechanics/hotfixes.py` (named in state and in mechanics-20261025-1) is not in the repo. wago.tools/hotfixes pages parse (Inertia `data-page` JSON), but the `build` filter is ignored and `search=` matches substrings, so filtering by build needs a real tool. No hotfix rows for 70630 were listed yet.
- Re-check post 7's racials against the next client build (new spells for Artisan's Precision and Command).
