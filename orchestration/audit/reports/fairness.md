# Fairness pass 55 (2026-11-24, master eceb09019)

Result: findings (2 new, 1 updated). Range e29cfe00f..eceb09019 (84 commits): Phase 0 step 4 regears of druid,
mage, paladin, priest, shaman, warlock and warrior requests; sim/core racial and base-stat changes (e19d2df85 Skyborne
base offsets, 9c08e16a2 Blood Fury spell damage, 7a2311b36 Troll Regeneration, ed1df1d51 Innervate MP5, 0352d95dd
shaman class row); client 1.60.1.71091. Full pass.

## What I ran

- 5k race sweep, every race, seed 1, report requests (CLI /tmp/audit-cache/cli/eceb090196be). New baseline:
  `ref/fairness/sweep_baseline.log`, `table_baseline.md` (eceb09019).
- audit1b: 19/19 "5a PASS | 1b PASS".
- Paired 20k (seeds 1 / 1000003) for each race flag, with the committed race rotation or race set where one exists.

## Findings

| id | sev | what |
|---|---|---|
| fairness-20261124-1 | high | Hunters (10 requests) and rogues (6) still wear Phase 1 synthetic sets; every other class wears Phase 0 real gear. Others dropped 5.8 to 23%, hunters/rogues 0 to -0.6%. Hunter step 4 branch: Lone Wolf 1174.0, MM 1090.5, SV 926.7 (-8 to -9%). Merge hunter/rogue step 4 before any re-sim/republish. |
| fairness-20261124-2 | low | Fire Undead 922.7 vs Human 920.5 (+2.2); Shadow Troll 857.1 vs Undead 855.2 (+1.9). |
| fairness-20261118-1 (update) | medium | Balance High Order (race rotation) still +4.1 / +5.3 / +4.9 / +5.8 over Night Elf on the Phase 0 set. |

## Race checks (paired 20k, mean of two seeds)

| Build | Report race | Challenger | Delta |
|---|---|---|---|
| mage__fire | Human 920.5 | Undead 922.7 | +2.2 |
| mage__fire | Human 920.5 | High Order (race rotation) 906.5 | -14.0 |
| priest__shadow | Undead 855.2 | Troll 857.1 | +1.9 |
| warrior__fury | Orc 1007.2 | High Order (race set) 1006.8 / (Orc set) 1008.7 | -0.4 / +1.5 |
| druid__balance | Night Elf 822.2 | High Order (race rotation) 826.3 | +4.1 |
| druid__balance-wrath | Night Elf 730.3 | High Order 735.6 | +5.3 |
| druid__balance-duty-ff | Night Elf 788.7 | High Order (race rotation) 793.5 | +4.9 |
| druid__balance-wrath-duty-ff | Night Elf 702.8 | High Order 708.5 | +5.8 |

Fury stays Orc: High Order's legal race set is below Orc. MM Night Elf +1.0 is the known kept call.

## Movement against the 8b5137c6d baseline

Every class that moved had its requests regeared to Phase 0 in the range (Feral -11.0%, Arms -5.8%, Fury -8.6%,
Fury 2H -10.8%, Enhancement -8.3%, Affliction -10.3%, Ret -6.8%, Fire -12.4%, Demonology -15.3%, Destruction -17.5%,
Frostfire -17.9%, Elemental -12.4%, Shadow -9.7%, Arcane -12.5%, Destruction Fire -16.3%, Balance -13.9%, Frost -23.0%,
Smite -10.9%). Hunters and rogues (requests unchanged): rogues 0.0, hunters -0.7 to -5.6 (already in pass 53, from
b82ad9a1e / earlier hunter commits). No unexplained move; the sim/core racial changes show up only through the races.

## Published vs current

Meta gaps (Lone Wolf -21.9, combat-swords -15.0, seal-fate-swords -14.4, seal-fate-daggers -12.8, Survival -4.8,
BM -2.7) are results-20261104-2's meta lag (its 2026-11-24 update lists the same). Independent confirmation, not refiled.
builds.json is the Phase 1 publication for all classes; a full re-sim is due once hunters and rogues are on Phase 0.

## Open items

Balance High Order (fairness-20261118-1), Ice Lance Low range (ruled 0.11), FTT stacking, boss partial resists,
Rage k / pet scaling: unchanged.
