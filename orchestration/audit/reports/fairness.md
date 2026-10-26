# Fairness pass 7 (2026-10-26, master f796d99)

Result: quiet (no new findings). No sweep: nothing that feeds a sim changed.

## Why it ran

23 watched files changed since e2316be, all `tools/dps_report/req/*.meta.json` (rogue x12, Enhancement x4, warrior x7).

## What I checked

- Every changed meta file differs only in `caveats` text (shorter wording, standards numbers and ruling dates dropped; the rogue flask caveat split in two). The `.json` requests, sim, proto, presets, APLs, item database and standards.md are untouched, so no build can move.
- The caveats still say the same thing as before for the shared assumptions (natural flask zone, Goblin Land Mine off, level 30 Rage rate, Flametongue Totem next to Windfury Weapon), applied to the same specs as before.
- Shared 20k run against published (`runs/vs_published.txt`): the same 4 of 57 over 0.5 as pass 6 (Night Elf Feral -1.4/-1.5, Subtlety -1.0/-1.1), within noise; the results auditor owns it.
- The pass 6 ranking and baseline (74b3a5f) below still stand.

---

# Last full pass: Fairness pass 6 (2026-10-26, master 74b3a5f)

Result: quiet (no new findings). fairness-20261025-1 resolved (cdceb56c4: Marksmanship plays a Human). Full pass.

## Why it ran

37 watched files changed since 016daa6: flat warrior Rage k (sim/core/rage.go s = 0, standards 21 ruling 2026-10-25) with re-searched warrior presets and APLs; hunter crit per Agility at Classic's value (sim/hunter/hunter.go), Multi-Shot in the Marksmanship / Lone Wolf APLs, Human Marksmanship; Forever race base stats for Dwarf, Night Elf, Gnome and High Order (sim/core/base_stats*.go); the full report regeneration 117d18af.

## What I checked

- Gear: `audit1b.py`, all 19 sets "5a PASS | 1b PASS".
- Race sweep: 36 base requests and variants, every race, 5000 iterations, seed 1, on the 74b3a5f tree. Baseline refreshed (`ref/fairness/`, sha 74b3a5f).
- 20k paired seeds (1 / 1000003) for the near ties the sweep flagged.
- The shared 20k run against published (`runs/vs_published.txt`): 4 of 57 move more than 0.5, all Night Elf builds (Feral -1.4, Feral FF duty -1.5, Subtlety -1.0, Subtlety vs Humanoid -1.1).
- Fixed `tools/audit/fairness/table.py`: meta.json races with a space ("Night Elf") crashed the Feral row; it now strips the space.

## Ranking (report race, 5000 iterations)

| # | Spec | DPS | Before (016daa6) |
|---|---|---|---|
| 1 | Beast Mastery | 1141.4 | 1171.9 |
| 2 | Marksmanship | 1139.6 | 1168.4 |
| 3 | Lone Wolf | 1127.4 | 1156.6 |
| 4 | Fire | 1100.8 | 1100.8 |
| 5 | Frostfire | 1098.0 | 1098.0 |
| 6 | Demonology | 1064.9 | 1064.9 |
| 7 | Affliction | 1036.4 | 1036.4 |
| 8 | Survival | 1025.5 | 1055.7 |
| 11 | Fury | 987.3 | 1069.1 |
| 13 | Fury 2H | 965.1 | 1034.5 |
| 17 | Arms | 906.7 | 956.5 |
| 19 | Feral | 867.1 | 868.5 |
| 28 | Subtlety | 714.6 | 715.8 |

Every other spec is unchanged to the decimal. Full table: `ref/fairness/table_baseline.md` and `orchestration/fairness_report.md` §1. Top over bottom 1.59x (was 1.64x).

## Conclusions

- Every move is attributed: warriors to the flat Rage ruling and their re-search, hunters to the crit-per-Agility change and Multi-Shot rotations, Night Elf / High Order builds to the race base stats. No class whose code and inputs didn't change moved, so no shared change leaked.
- Published against current: all within 1.9 at 5000 iterations. The req metas for Feral and Subtlety are 1.0 to 1.5 stale (the regen commit notes it, inside verify.py's tolerance of 2), but the dashboard's builds.json has the re-simmed 866.7 / 714.9, so the ranking people see is current. Not a finding.
- Best race: Marksmanship Human is the best race (resolves fairness-20261025-1). Subtlety Orc vs Night Elf: 715.2 / 714.7 vs 714.9 / 714.7 at 20k, a tie. Lone Wolf Human vs Windshaper: 1127.1 / 1127.0 vs 1127.0 / 1127.6, a tie. Rogue Combat Daggers High Order = Windshaper (770.5). Keep all.
- results-20261025-1 (consumes mismatch in mksettings) is the results auditor's; nothing to add.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range (now 0.143 per the 2026-10-25 rulings in the report), Flametongue Totem stacking, partial resist vs bosses: see passes/fairness.md. Warrior Rage k is now one value (flat), so it is no longer an assumption-driven range.
