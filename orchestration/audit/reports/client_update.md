# Client update pass: 1.60.1.70532 -> 1.60.1.70630

- Pass 2026-10-27, master d3df9247f7dc9317eceabda2f4b9b649d92178f4. `wago.py latest --probe 300` returns 1.60.1.70630.
- For `.agents/skills/dps-report/state.json`: `foreverClientBuild` → 1.60.1.70630 (the orchestrator updates it).

## Client diff

`table_summary.py 70532 70630`: real changes. Spell 55 changed / 9 added, SpellEffect 27 changed / 30 added / 2 removed, SpellMisc 67 changed, Trait* (the mage tree), SpellCooldowns, SpellCategories, ItemSparse 28 changed. Identical: SpellItemEnchantment, ChrRaces, SpellCastTimes, TraitDefinition, TraitTree, SpellRange, SpellReagents. `db2diff.py`: 132 spells (`ref/client_update/diffs/70532-70630.md`). `items_diff.py`: no added or removed items; the 28 changes are flags, sheathe types and low-level item levels (Dark Iron Boots only gains a flag). `enchant_diff`: no SpellItemEnchantment change.

Most of the build is the 10-16 to 10-21 server hotfixes now in the client: the mage Frost tree, Elemental Precision, Ice Barrier / Frigid Power / Frost Ward / Frost Warding, Winter's Chill, Hunter's Mark rank 4 143, Lava Burst rank 3's Flame Shock bonus, Holy Fire rank 1, Amplify Curse on Curse of Exhaustion. The sim already had all of them (landed 2026-10-23).

## Sources

| Source | State |
|---|---|
| Blizzard dev notes (topic 2360696) | **New post /7**, 2026-10-26 20:35 (edited 21:09): racial pass. Not in 70630. Baseline saved as `sources/devnotes_post7_2026-10-26.txt` |
| foreverchanges.pro | Still lists 70532 as current; says the beta closed 2026-10-23. No new hotfix since the 10-23 Righteous Fury one |
| Wowhead | Not checked (blocked) |

## DPS-relevant changes

| Ledger row | Change | Sim status | Sized (180 s, 20k, paired seeds) |
|---|---|---|---|
| MAGE-U-15 | Missile Barrage tooltip drops the Arcane Missiles hit proc (adds Frost Blast) | ambiguous, finding -1 (high) | mage__arcane −77.6 (−7.9%); frostfire 0 |
| WL-U-16 | Spellstone is Shadow only (was Fire + Shadow) | wrong, finding -2 (low) | affliction −1.5, demonology −1.4, destruction −2.8 (−0.13 to −0.28%) |
| WL-U-17 | Amplify Curse no longer touches Bane of Agony | wrong, finding -3 (low) | not sized: no preset casts it |
| SHR-U-15 | Touch of the Grave drain loses "ignore caster modifiers" | ambiguous, finding -4 (low) | not sized |
| SHR-P7-01..05 | Post /7 racials: Human 1% crit and 10% Spirit, Dwarf 1% hit instead of crit damage, Tauren and Orc lose their crit, Orc Command, Skyborne 2% haste | ambiguous (not in the client), finding -5 (medium) | estimated: about ±1% for each race |
| MAGE-U-16, -17, SHA-U-20 | Arcane Wrath category cooldown, Revelation mask, Elemental Mastery cooldown on cast | landed | — |
| MAGE-U-18, -19 | Frost Blast coefficient 0.343 → 0.271, Flamestrike attribute | not simulated / n/a | — |

Ledger: new section "Build 1.60.1.70630" (14 rows), Summary counts updated, Rulings needed items 5 and 6 added.

## Open items

- Champion of the Light (PAL-U-HF1) still can't be checked in the client: its curve tables aren't downloaded.
- Post /7 will need a new pass when a build or hotfix carries it. Post /7 might also be cut short: it covers only Human, Dwarf, Tauren, Orc and Skyborne.
- `dps_beforeafter.sh` takes a directory holding `wowsimcli`, not the binary itself, and prints empty rows when given the binary. MANIFEST's `-o $AUDIT_WORK/bin/after` goes with that only when `after` is a folder.
