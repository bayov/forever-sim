# client_update pass 16 (2026-11-24, master eceb090196)

Build: **1.60.1.71091** (wago.tools `latest --probe 300`), baseline 71037. Note for `.agents/skills/dps-report/state.json` `foreverClientBuild`: 1.60.1.71091.

## What was checked

- `table_summary.py 71037 71091`: 75 spells changed (SpellEffect +15/-5/~42, SpellMisc ~34), ItemSparse +1/~3, ItemEffect -1/~1. Trait*, SpellItemEnchantment, ChrRaces identical.
- `db2diff.py 71037 71091` (337 lines) and `items_diff.py 71037 71091` (no request item, nothing at ilvl 55+).
- Sources: foreverchanges.pro (71091 current, "No item reads differently"; Blizzard's November 23 hotfix notes and "Found in the game"). Blizzard forum JSON rate-limited ("performed this action too many times"); the Nov 23 notes were read through foreverchanges.
- Each player-relevant change grepped in `sim/` and read; open branches `forever-update/2026-11-22` (PR #137) and `forever-update/2026-11-24-71091` checked.

## What 71091 is

The client copy of server pushes 114268, 114283, 114309, 114313, 114314 and 114320 (Nov 19-23), plus a new druid-themed NPC encounter (~45 NPC spells). No talent, enchant, racial or item-stat change.

## DPS-relevant rows (ledger section "Build 1.60.1.71091", 12 rows)

| Row | Change | Status |
|---|---|---|
| WAR-HF-05 | Execute casts: Suppress Weapon Procs, Suppress Caster Procs off (push 114313) + Blizzard "fewer weapon procs" | ambiguous; update on client_update-20261120-1 (`warr_exec_wf2`) |
| ROG-HF-07 | Mutilate casts suppress weapon procs | landed (no separate cast in sim) |
| ROG-HF-08 | Shank procs again, Boot Knife as a target proc aura | landed |
| PAL-HF-08 | Judgement 20271 suppresses weapon procs | landed |
| PAL-HF-09 | JoL/JoW/JoJ 40 s | landed |
| PRI-HF-01 | Chastise in Shadowform | in flight (PR #137) |
| PRI-HF-02 | Dark Inspiration Spirit % via Mind Flay | landed |
| WL-HF-06 | Intensity on Incinerate | in flight, n/a |
| DRU-HF-03 | Feralheart 4P / Wildheart 5P Nature's Bounty gives Energy | in flight (3f1c2708d) |
| SHR-HF-07 | Fengus' Ferocity 115 SP | in flight (3f1c2708d); update on client_update-20261030-4 |
| SHR-HF-08 | Darkmoon Faire buffs "twice as much power" | ambiguous; new finding client_update-20261124-1 (low) |
| SHR-HF-09 | Stormshroud 3P 3% / 18 Energy now in client | landed (PR #136) |

Also in the client: Two-Handed Weapon Specialization 20111 Attributes_5 0x10000 (PAL-HF-07, landed).

## Conclusions

- One new low finding (Darkmoon Faire values unverified). Nothing missing or wrong at yes-level beyond what is in flight or already open.
- Open findings: none resolved on master since pass 15 (11 sim commits, none touching them).
- Left open: Blizzard forum thread 2403040 not re-read (rate limit); the Nov 23 notes came from foreverchanges.
