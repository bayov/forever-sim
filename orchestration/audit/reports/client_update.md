# Client updates: pass 11 (2026-11-07, master 12aa2796fc)

## Sources

| Source | Result |
|---|---|
| wago.tools (`wago.py latest --probe 300`) | 1.60.1.70862 (was 70790) |
| `table_summary.py 70790 70862` | 461 spells changed, 395 of them added (NPC, dungeon, portals, "Endless" consumables); SpellEffect 37 changed; ItemSparse 22 changed (flags, recipe quality, ilvl 10 -> 30 level-1 copies); 1 enchant added (quest arcanum); ChrRaces, trait trees, SpellCastTimes, SpellDuration identical |
| foreverchanges.pro | 70862 "Not in Blizzard's notes": 7 items; "The Hyjal flasks' 30 Stamina and 2% were already on their item pages from the server's hotfixes". 11-06 hotfix: tooltips, Bloom, Arcane Blast 18 s, Soul Siphon/Wrack, Voidwalker health, Orc Command. 11-06 "Announced for November 10" (pets, Wrack vs DoTs, Enhancement Strength): not live, no rows yet |
| Dev notes thread 2360696 | not re-fetched (the hotfix lines come through foreverchanges) |

## DPS-relevant changes and sim status

| Row | Change | Sim status |
|---|---|---|
| SHR-U-19 | Natural flasks (1293740-1293743): Stamina 60 -> 30, zone bonus 5/4/5/5 -> 2 | **missing**: `sim/core/consumes.go` applyNaturalFlask still 60 and 5%/4%/5%/1.05. Already filed today as mechanics-20261107-1 and consumables-20261107-1, so no duplicate finding here |
| SHR-U-20 | Orc Command warlock 20575 effect 1 (+5% Firebolt / Lash of Pain) -> 0 | landed (the sim never had it) |
| MAGE-HF-01 | Arcane Blast lasts its full 18 s (server fix) | landed (`arcane_blast.go:31`) |
| WL-HF-01 | Soul Siphon no longer counts Wrack for Wrack's own damage (server fix) | landed (sim never counted Wrack) |
| DRU-U-17, WAR-U-15 | Bear Form and Defensive Stance threat 50 -> 115, 30 -> 80 | n/a (threat) |

Measured cost of SHR-U-19 (scratch before/after build, 180 s, 20k iterations, seeds 1 and 1000003; both seeds agree within 0.05%):

| Request | Flask | Before | After | Delta |
|---|---|---|---|---|
| druid__feral-cat | Aggression | 1168.78 | 1150.99 | -1.52% |
| hunter__marksmanship | Aggression | 1253.96 | 1231.99 | -1.75% |
| hunter__survival | Precision | 1040.08 | 1006.62 | -3.22% |
| rogue__combat-daggers | Accuracy | 954.64 | 936.22 | -1.93% |
| rogue__mutilate | Precision | 863.36 | 846.50 | -1.95% |
| warrior__arms | Aggression | 1087.13 | 1068.42 | -1.72% |
| warrior__fury | Precision | 1136.41 | 1092.76 | -3.84% |

The scratch edit: Stamina 30, hit 2/2, crit 2/2, expertise 2, speeds 1.02. With the bonus at 2%, standards item 11's flask choice per spec should be rechecked (Supreme Power or Titans may now win for some).

## Ledger

Added `## Build 1.60.1.70862 (first seen 2026-11-07)` with rows SHR-U-19, SHR-U-20, MAGE-HF-01, WL-HF-01, DRU-U-17, WAR-U-15; Summary 767 -> 773 rows. For the orchestrator: `.agents/skills/dps-report/state.json` `foreverClientBuild` -> 1.60.1.70862.

## Open findings

client_update-20261105-1 (feral exact rank values) still open: `nov2BleedScale` is still on Bite, Rake and Rip.
