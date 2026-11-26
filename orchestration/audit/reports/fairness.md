# Fairness pass 62 (2026-11-26, master ab3fff700)

Full pass: range bfbee78fc..ab3fff700 (29 commits) is the warrior and hunter Phase 0 re-rank (Arms, Fury DW, Survival
requests and race rows) plus warrior leveling-mode work that touched sim/core and sim/warrior. 5k race sweep (seed 1,
every race, the report's own requests with only the race changed) on the shared ab3fff700 CLI, table.py against the
bfbee78fc baseline, phase0_check.py, the loop's 20k run against published, and paired 20k binary checks
(bfbee78fc / 26ecec3f1 vs ab3fff700, same requests).

## What moved, and why

| Build | Before | Now | Cause |
|---|---|---|---|
| Survival (and wolf / tallstrider) | 950.5 | 963.5 (+1.4%) | own requests: 02c06a709, 7fb108796, 9c1cf0b4d (new trinkets) |
| Fury DW | 986.2 | 994.6 (+0.9%) | own request: f79e810e1 (Thrash Blade, Heart of Wyrmthalak) |
| Arms | 1010.6 | 1012.9 (+0.2%) | own request: 7a3b753cb / bb218ef46 |
| Ret, Sub Holy | 930.0 / 903.9 | 931.1 / 905.2 | **leak**, no paladin change: fairness-20261126-3 |

Every other build is identical to the baseline (same seed, so exactly 0.0). The warrior and core leveling code is gated:
paired 20k, 26ecec3f1 (re-rank branch before the leveling merge) vs ab3fff700 binaries, Fury, Arms, Fury 2H and Arms
sunder duty move 0.00. TargetsNear, InExecuteRange and Cannibalize all fall back to the Level 60 behaviour outside a
leveling session (sim/core/aoe.go:33, leveling_melee.go:161, :197).

## New finding

- **fairness-20261126-3 (low):** 8f5b59325 made a stunned target unable to dodge, parry or block
  (spell_outcome.go). Blackblade of Shahram's Might of Shahram stuns every target with no StunImmune check
  (item_effects.go:545-565), so the stun-immune boss now loses its avoidance for 5 s per proc. Paired 20k:
  paladin__level_60 +1.20, sub_holy +1.22, shock_ret 0.00 (other weapon). Fix: skip StunImmune targets under Forever,
  as the explosives do.

## Published against current (results-20261104-2; not filed again)

The re-ranked warriors and hunters match their metas (<=1.6). Still lagging at the report race, 5k: Ret +7.5 (now
including the leak), Elemental +3.9, Affliction +3.4, Enhancement +2.3, Destruction Fire -3.2, Affliction Shadow Bolt
+26 (at Orc). No place in the ranking changes.

## Gear (standards 43)

- phase0_check.py: every DPS request wears its spec's set; the same two close calls the gear pass holds (Destruction
  Spellblade +1.75, Enhancement Blackrock Slicer +1.77).
- fairness-20261126-1: Arms, Fury, Fury DW and Survival lists are now on 34fe148a5 (batch rules in). Open for the six
  rogue lists only.

## Explosives and race (open findings, no change in range)

- fairness-20261125-2: Fury, Fury 2H, four rogues and Ret still carry no filler explosive (not re-measured).
- Race, 5k: Demonology Undead +4.2 (paired +5.9 last pass), MM High Order +5.1 (variants +6.6), Balance Wrath High
  Order +4.1, Arcane Undead +1.9, Destruction Undead +1.7, Fire Gnome +1.1, Fury DW Undead +0.7 (now noise after its
  regear). Survival tallstrider High Order +0.5 (noise).

## Ranking

Top over bottom 1.68x (Lone Wolf 1140.5 over Smite 679.5). Survival rises from 10th to 9th (past Affliction); Fury DW stays 7th, 0.2 behind Fury 2H.
Full table in orchestration/fairness_report.md §1.

## Baseline

sweep_baseline.log and table_baseline.md replaced with this sweep (ab3fff700).

Note: origin/master had moved to 6a447e231 (warrior leveling presets only) during the pass; ab3fff700 was audited as
instructed.
