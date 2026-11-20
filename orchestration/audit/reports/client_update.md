# client_update pass 15: build 1.60.1.71037 (2026-11-20)

Master audited: e6eaa94706. Previous build 1.60.1.70959. Sim runs: 2,000-iteration runs of the three Retribution requests and Subtlety on the master CLI, only for damage shares and cast counts (no before/after; fixes need code).

## What 71037 is

The client copy of Blizzard's November 17 hotfixes (t/2403040 post 12, 2026-11-18 03:24Z) and part of push 114097. Post 13 (November 19, 2026-11-20 04:00Z) was read too; its Retribution line is server side. Dev-notes thread 2360696: nothing after post 10 (11-04). foreverchanges.pro: no 71037 yet. `table_summary.py 70959 71037`: SpellEffect +16 / -2 / 25 changed, SpellMisc 74 changed, ItemSparse 74 added / 17 changed; Trait*, SpellPower, ChrRaces identical. Mostly camping, Reagent Economy, NPC spells and new dungeon set items.

## DPS-relevant changes and sim status

| Ledger row | Change | Sim |
|---|---|---|
| PAL-HF-07 | Blizzard 11-19: Two-Handed Weapon Specialization on all Holy damage with a 2H | **wrong** vs ruling 2026-11-15; already filed as mechanics-20261120-1. Damage shares (2k it): HoW + JoR 8.5-9% of `paladin__level_60`, HoW + Holy Shock 16.6% of shock_ret, so about +0.5% / +1% at 6% |
| ROG-U-02 | Rogue T1 5P cuts Death Mark's cost too | **missing**; already filed as mechanics-20261120-2. Subtlety casts 9.5 Death Marks a 180 s fight: 48 energy |
| SHA-HF-10 | Blizzard 11-17: Windfury Totem attacks don't spend Flurry | **wrong**; already filed as mechanics-20261118-1 |
| WAR-U-16 | Execute dummy Always Hit + suppress procs | ambiguous; new client_update-20261120-1 (low) |
| HUN-U-14 | Hunter T1 5P reaches Multi-Shot (push 114065 in client) | landed |
| SHR-HF-05 | Darkmoon Card: Blue Dragon 3% (now in client) | landed (3ec610c40); 20261117-3 resolved |
| WAR-HF-01..03 | Flurry / Unbridled Wrath / Reckless Swings mask 4 -> 20 in the client | unchanged (UW, RS landed; Flurry 20261117-2 still open) |
| SHR-HF-04 | Windfury Totem procs get category 2619, 100 ms | landed |
| SHR-U-25 | Staffs, Onyxia Blood Talisman, Spiritcaller pieces, 74 new dungeon set items | missing in db, minor (update on 20261117-4) |

n/a: Reagent Economy, camping, world buff death/player flags, Vampiric Embrace, Filcher's Cowl, Sap, Genesis on Pounce Bleed, Darkmoon Heroism/Maelstrom, Howler's Cry, Lash of the Dark Rider, NPC spells.

## Findings

- New: client_update-20261120-1 (low, needs verify).
- Resolved: client_update-20261117-3.
- Updated: -20261117-1 (WF category), -20261117-2 (mask in client), -20261117-4 (more items).
- Not filed again (the mechanics auditor has them): mechanics-20261120-1, -20261120-2, -20261118-1.

## For the orchestrator

- `.agents/skills/dps-report/state.json` `foreverClientBuild` -> 1.60.1.71037.
- Ledger: new section "Build 1.60.1.71037", Summary +6 rows (819), ruling 8 (PAL-HF-07).
