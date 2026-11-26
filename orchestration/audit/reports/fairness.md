# Fairness pass 61 (2026-11-26, master bfbee78fc)

Full pass: range 85691655a..bfbee78fc is the Phase 0 gold-line re-rank (every DPS list and all 68 requests changed;
no sim/, proto, database or standards.md change). 5k race sweep (seed 1, every race, the report's own requests with
only the race changed) on the shared bfbee78fc CLI, table.py against the 85691655a baseline, phase0_check.py, the
loop's 20k run against published, paired 20k race checks and a paired 10k explosive check
(tools/audit/fairness/explosives.py, now committed).

## What moved, and why

Every build moved with its own request (gear re-ranked under the gold line); no class moved without a request change
and no shared code changed, so nothing leaked. Most specs fell 0.3% to 1.9% (Smite -1.9%, Balance -1.7%, Fury DW
-1.6%, Seal Fate Daggers -1.6%); Enhancement +0.9% (re-rank plus Land Mine), Combat Swords +0.9%, Feral +0.6%.
Top over bottom 1.68x (was 1.66x).

## Published against current (results-20261104-2; not filed again)

94d6cd2e4 re-simmed the report before the last re-rank commits (4a39a50be shaman / Ret, c15219593 warlock,
1a6600121 Elemental). Gaps at the report race, 5k: Ret +6.4, Sub Holy +6.6, Shock Ret +3.5, Elemental +3.9,
mana-tide +4.0, Affliction +3.4, Enhancement +2.3 (variants +3.5 to +5.2), Destruction Fire -3.2, Affliction Shadow Bolt
+26.3 (at Orc). The loop's 20k comparison agrees (25 of 68 off by >0.5). No place in the ranking changes. Fix: re-sim.

## Gear (standards 43)

- phase0_check.py: every DPS request wears its spec's set; two '!' close calls (Destruction Spellblade +1.75, Enhancement
  Blackrock Slicer +1.77) already noted by the gear pass.
- **New, fairness-20261126-1 (medium):** the rogue (6), warrior (3) and Survival lists were searched on sims without
  295206ddf / b663145ad / efeb06b95 / 22d2cf952 (proc rules, weapon PPMs, set procs, explosives); every other list is on
  a sim equal to master. Their picks are due a re-search.
- List `dps` sits 1.2 to 16.5 above the report race's sweep DPS (gear pass: the list number is the search fight with
  uses). The gap doesn't follow the stale-sim split (BM 1.8, Fury 2H 1.5 vs MM 14.4, Shadow 16.5), so not a finding here.

## Explosives (fairness-20261125-2, still open)

Survival and Enhancement now carry Goblin Land Mine. Paired 10k, mine / Dynamite over none:

| Build | Mine | Dynamite |
|---|---|---|
| Fury DW | +4.1 | +3.4 |
| Fury 2H | +2.1 | +1.4 |
| Mutilate | +6.8 | +5.8 |
| Combat Swords | +6.5 | +5.7 |
| Combat Daggers | +6.1 | +5.2 |
| Seal Fate Swords | +6.5 | +5.8 |
| Ret | +6.5 | +5.7 |
| Shock Ret | +7.7 | +6.8 |
| Arms | -1.2 | -1.9 |

## Race (paired 20k unless noted)

| Build | Report race | Best | Gain | Finding |
|---|---|---|---|---|
| Demonology | Orc 877.1 | Undead 883.0 | +5.9 | fairness-20261125-3 |
| Destruction | Orc 868.2 | Undead 870.0 | +1.85 | fairness-20261125-3 |
| Marksmanship (5k) | Orc 1082.4 | High Order 1087.5 | +5.1 | fairness-20261125-1 |
| MM wolf / tallstrider (5k) | Orc | High Order | +6.6 / +6.6 | fairness-20261125-1 |
| Arcane | Gnome 833.9 | Undead 836.05 | +2.15 | fairness-20261126-2 (new) |
| Fury DW | Orc 983.8 | Undead 985.35 | +1.55 | fairness-20261126-2 (new) |
| Fire (5k) | Human 917.8 | Gnome 918.9 | +1.1 | fairness-20261124-2 |
| Balance Wrath (5k) | Night Elf 714.2 | High Order 718.3 | +4.1 | fairness-20261118-1 |

Shadow, Balance (base), Lone Wolf (+0.3), BM, Survival, Feral, the rogues and the paladins are at their best race.

## Baseline

sweep_baseline.log and table_baseline.md replaced with this sweep (bfbee78fc). Ranking in orchestration/fairness_report.md §1.
