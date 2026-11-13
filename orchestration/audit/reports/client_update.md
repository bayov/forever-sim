# client_update pass 13: build 1.60.1.70938 (2026-11-13)

Master audited: 97da1cca5c34. Previous build 1.60.1.70888. Sim runs: none (no change needed sizing).

## What 70938 is

The client copy of the server hotfixes of November 10 and 11 (foreverchanges.pro: no Blizzard notes for it; nothing reads differently from the hotfixes already live). `table_summary.py 70888 70938`: 98 spells changed, ItemSparse 1,042 changed (cosmetic fields), 1 enchant (Spirit stat type fix); trait tables, ChrRaces, cast times and durations identical. Also read: Blizzard's 11-11 and 11-12 hotfix notes on foreverchanges (server side, not in this client).

## DPS-relevant changes and sim status

| Ledger row | Change | Sim |
|---|---|---|
| WL-HF-02 | Wrack: DoTs tick 100% faster while channelled (no more 10% damage taken), own damage x0.7 | landed (`sim/warlock/wrack.go`, `talents.go` applyWrack) |
| WL-HF-03 | Soul Siphon counts only Corruption, Siphon Life, Bane of Agony | landed |
| SHA-HF-02 | Lightning Overload rows now exactly half of Lightning Bolt and Chain Lightning (180.7 / 261.6 at 60, coef 0.357) | landed (x0.5 copy); settles SHA-HF-01 |
| SHA-HF-04..07 | 11-12 Enhancement hotfix: Stormstrike 100% / 8 s, Mental Dexterity 33/66/100%, Maelstrom Weapon 20% a stack at 6-10 PPM, imbues and weapon totems both proc | landed |
| HUN-HF-05 | Volley activates no procs, breaks on movement | landed (2fcb6fe2a), minor |
| DRU-U-19 | Idol of Ferocity no longer discounts Mangle | landed (never did) |
| SHR-U-23 | Poised's trigger restructured in the client (helper auras); 10 Energy unchanged | landed |

n/a: Bane of Havoc GCD, Imp Firebolt ranks 1-4, Flametongue Totem rank 2, Rend and Tear on Pounce, healing and priest buff masks, low-level items.

## Findings

- No new findings.
- client_update-20261110-1 (Arcane Concentration split) resolved by ad6884211.
- client_update-20261110-2 (ilvl 65 trinkets) partly fixed: Adaptive Combat Assistant updated in the db, Weakness Analyzer and Serenity Field still stale. Stays open.

## For the orchestrator

- `.agents/skills/dps-report/state.json` `foreverClientBuild` -> 1.60.1.70938.
- Ledger section "Build 1.60.1.70938" added (14 rows: 10 landed, 4 n/a); Summary counts updated.
