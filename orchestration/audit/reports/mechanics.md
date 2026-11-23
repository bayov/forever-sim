# Mechanics pass 100 (2026-11-23)

Master b48605aa3..c78d983ac. Client 1.60.1.71037 still the newest (wago probe 300). Result: quiet.

## Get current

| Source | Newest read | New since pass 99 |
|---|---|---|
| Dev notes thread 2360696 | post 8, 2026-11-04 (post 1 edited 11-03) | none |
| Hotfix thread 2403040 | post 14, 2026-11-21 (post 1 edited 11-21) | none |
| Blue posts (Kaivax, Aggrend, Gnedi, Linxy, Brenoch) | Kaivax 2026-11-23, topic 2426335 | lockbox behaviour only, no combat mechanic |
| Daily update | 0f1d8ebcd TASKS.md: process 71037 | already covered by pass 99 (Windfury 100 ms category CD folds into mechanics-20261118-1) |

## Commits

| Commit | Area | Verdict |
|---|---|---|
| aff06f74e | `sim/core/ruleset.go` unifyEquipHitAndCrit: items flagged KindLimitedCrit keep crit to their own kind (Blackhand's Breadth melee/ranged, Eye of the Beast spells) | Sound: the flagged crit is pulled out of the shared pool and added back to its own stat only; multipliers applied as for the pool. Item-level tooltip source, not a shared formula change. |
| 6e128175e | `sim/core/racials.go` GetFaction -> RaceFaction | Refactor, behaviour unchanged. |
| 57ea8cb6b | Weapon skills from the 70938 item sweep; 1% ranged-only hit on Marksman Bands, Stoneshatter, Sash of the Grand Hunt; not passed to the hunter pet | Matches the Biznick's Accurascope treatment; pet comment updated. |
| 31e5077ab | Forever item effect values and 26 Forever-only procs (`sim/common/forever_item_procs.go`) | Spell data spot-checked on 71037, all match (table below). PPMs are not in the client (ProcChance 101, no SpellProcsPerMinute); the 1 PPM default is a stated assumption, not a finding. |

## Spot check, client 1.60.1.71037 (`tools/audit/mechanics/sp.py`)

| Spell | Sim | Client |
|---|---|---|
| 17483 Shadow Bolt (Demonshear) | 150 + 34 / 2 s for 6 s | bp 150; aura 3 bp 34 per 2000, dur 6000 |
| 18086 Firebolt (Force of Magma) | 182 | bp 182 |
| 1299631 Smite (Lightcrest Hammer) | 105, x2 vs Undead | bp 105; eff 3 bp 2 |
| 19874 Lightning Bolt (Shimmering Platinum Warhammer) | 252 | bp 252 |
| 16415 Fireball (Taran Icebreaker) | 300 + 25 / 2 s for 8 s | bp 300; bp 25 per 2000, dur 8000 |
| 1298499 Warblade of Caer Darrow | 27.72 | bp 27.72 |
| 1300783 Slaughter (Angerforge's) | 109, x2 vs Orc/Troll | bp 109; eff 3 bp 2 |
| 1315878 Flamebreak Talisman | 33 Agi (66 vs Elemental), 30 s, 5 min / 30 s category | aura 29 bp 33 x2, dur 30000, cd 300000 / 30000 |
| 1291103 Serenity (Serenity Field) | 72 Spirit, 15 s | aura 29 misc 4 bp 72, dur 15000 |

## Open findings

None of the six open findings was touched by these commits; all stay open.
