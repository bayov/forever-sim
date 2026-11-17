# client_update pass 14: build 1.60.1.70959 (2026-11-17)

Master audited: 5f20dfdf4c. Previous build 1.60.1.70938. Sim runs: none (the two medium findings need code changes before they can be sized).

## What 70959 is

The client copy of Blizzard's hotfixes of November 13 and 16 (t/2403040 posts 10 and 11, both read here; foreverchanges: no notes for the build). `table_summary.py 70938 70959`: Spell 21, SpellEffect +8/-3/12, SpellAuraOptions -23/2, SpellMisc 115 (mostly a logging bit on consumables), SpellCategories 33, ItemSparse 38 changed; ChrRaces, enchants, trait tables (one curve row added), cast times and durations identical. Maelstrom Weapon's PPM curve 128299 reads 6/7/8/9/10, Stormstrike 100% / 8 s: the 11-12 Enhancement hotfix, already landed.

## DPS-relevant changes and sim status

| Ledger row | Change | Sim |
|---|---|---|
| SHR-HF-01 | Windfury Totem proc becomes a weapon-damage strike (no extra white swing, no 2-charge AP buff); Blizzard: armor and ICD fixed | **wrong** (20261117-1) |
| SHR-HF-02..04 | WF Totem rage, one roll per main hand hit (no Mutilate/Execute double, no shield abilities), 100 ms ICD | landed |
| WAR-HF-02 | Flurry no longer spent by Heroic Strike or Cleave (Blizzard 11-16) | **wrong** (20261117-2) |
| WAR-HF-01, 03 | Unbridled Wrath and Reckless Swings from Heroic Strike and Cleave | landed (ad1d7f469) |
| PAL-HF-05, 06 | Judgement of Command melee with on-hit effects; other judgements ranged defense | landed |
| PAL-U-14 | Unsealed Wrath 100 ms proc cooldown | landed (no DPS change) |
| SHA-HF-08, 09 | Maelstrom Weapon one cast; Thundering Strikes without a weapon | landed |
| SHR-HF-05 | Darkmoon Card: Blue Dragon 2% -> 3% | missing, minor (20261117-3) |
| SHR-U-24 | 11-13 item stat corrections (Darkspear Signet Mp5, Bad Hat, Channeler's Ring...) | missing in db, minor (20261117-4) |

n/a: Berserker Rage vs Sap, Siphon Life silence, Wild Growth, Genesis on Pounce, Dark Inspiration, Vampiric Embrace, poison max rank, Shank, consumable logging bits, quest and dungeon lines.

## Findings

- New: client_update-20261117-1 (medium), -2 (medium), -3 (low), -4 (low).
- Open ones not re-checked beyond: 70959 does not touch the 20261110-2 trinkets.

## For the orchestrator

- `.agents/skills/dps-report/state.json` `foreverClientBuild` -> 1.60.1.70959.
- Ledger section "Build 1.60.1.70959" added (16 rows: 10 landed, 2 wrong, 2 missing, 2 n/a); Summary counts updated.
