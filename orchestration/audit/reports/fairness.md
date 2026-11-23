# Fairness pass 50 (2026-11-23, master 5518f4873)

Result: quiet (no new finding; fairness-20261122-1 resolved). Range 7a6d1beb9..5518f4873, 12 watched files.
Economy pass: no full race re-sweep. Current numbers from the loop's shared 20k run (report race, 5518f4873 CLI)
against builds.json at 5518f4873; baseline left at 8b5137c6d.

## What changed in the watched paths

| Change | Commit | Effect at the report race |
|---|---|---|
| DPS Report re-sim of warrior, Affliction, Enhancement, mage, hunter builds | 6ce43d3c8 | builds.json catches up with the pass 49 sim fixes |
| Frost / Frostfire re-tune for Ice Lance 0.11 (talents and APL only; race, gear, consumables, buffs unchanged) | b8b89d6a8, re-sim 98dc8871e | Frost 834.6 -> 838.6, Frostfire 1006.0 -> 1007.3 |
| Helm of Fire, Circle of Flame, Robe of the Archmage, Headmaster's Charge uses; Axe of the Deep Woods; Flame Wrath shield (sim/common) | efe009fe1, 65f071726, 078699747, 592444f24 | no report request wears them (synthetic gear, 1b) |
| Empty weapon slot doesn't swing; APL wait guard (sim/core/attack.go, apl_actions_timing.go) | 79c688d0c | no request moves |
| Item database (reveals, Onyxia Blood Talisman Parry, Phase 0 items) | 2ed0e68ba, 56cafb3b7, 604b0516d | none in the synthetic sets |

## Checks

- Published vs current: builds.json within 1.5 of the shared run on 68 of 68 builds (largest 0.8, Frostfire and
  Enhancement). fairness-20261122-1 resolved. Independent confirmation of results-20261104-2's 2026-11-23 update
  (68 of 68). The req metas' `published.dps` still lag (61 of 68 off by more than 0.5, the stale mage metas
  included); that is results-20261104-2's point, not refiled.
- Movement: every non-mage row within 0.8 of pass 49's shared run, so no shared change leaked.
- Race (builds.json race table, 10k): Balance High Order +3.8 to +6.1 for all four builds (fairness-20261118-1,
  open). Frost High Order and Frostfire Human remain best. Below the bar for a finding: MM wolf / tallstrider pet variants
  show Night Elf +1.6 / +1.8 over Orc while base MM keeps Orc, and Fire shows Gnome +1.2 over Human. Both are near
  race-table noise. Worth a look if they persist.
- Gear: `audit1b.py` on 5518f4873, 19 of 19 "5a PASS | 1b PASS".
- Shared assumptions: the mage requests changed only talents and rotation; no consumable, buff or debuff drift.

## Ranking (shared 20k run, report race)

Unchanged order from pass 49. Top Lone Wolf 1291.3, bottom Smite 781.0, 1.65x. Full table in
`orchestration/fairness_report.md` §1.
