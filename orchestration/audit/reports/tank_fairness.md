# tank_fairness pass 38

Master 4a6dad5c2 (last: f4c45c3f9). Due: 2 watched files, `ui/tank_warrior/presets.ts` and
`ui/tank_warrior/apls/leveling_plan_prot.json`. Quiet, no sims run.
- dc7eb8c8f: the warrior Leveling mode's Prot plan now plays a 5/15/31 talent order without Defiance by 20; the
  presets.ts change is that plan's comment and the Leveling: Prot build tooltip (96.4h, 26 deaths). Leveling only: no
  level-60 tank preset, gear set, talent string, APL or `tools/tankbench/req/*.json` request changes.
- No change under `tools/tankbench`, `tools/tankscore`, `tools/encounters`, `sim/` or the other tank UIs in the range.
- So the pass-35 tables below and ref/tank_fairness/summary.tsv still hold. All open findings stay open.

## Pass 35 (master f9fa79917), still current

Master f9fa79917 (last: 3e9748641, the Phase 0 tank batch now merged). Due: 80 watched files. Tank-relevant: the
Phase 0 merge's request changes (bear Balanced re-searched from the Max threat list, b7b411a01; paladin Max survival
Titans, 52572a1d7, and Max threat Spirit of Zanza, 3ae9da70a; Warden races Troll / Troll / Tauren, 3ea8c0ecc, and its
gear re-searched on them, eae898b24, 12e6132b2; warrior requests cast the worn on-use trinkets, 0741191ab, 2f20719e4).
The rest is the warrior leveling batch (rank by level, leveling-only rules), Water Shield's 4th globe with Totem of
Sustaining (no tank wears it) and a leveling kill hook in `sim/core/target_death.go`.

Method: tankbench built `--tags=with_db`, `tools/audit/tank_fairness/bench.sh` on all 12 presets x six scoring bosses,
3,000 iterations, seeds 1 and 2; the paladin Balanced / Max threat pair again at 10,000 on both seeds.
Baseline refreshed: ref/tank_fairness/summary.tsv (seed 1).

## Same terms

- Level 60, Engineering / Blacksmithing, reaction 200 ms, raid buffs and debuffs (same hash on all 12), Kings and
  improved Might: equal. The Warden alone has improved Wisdom (a mana tank; the paladin gains nothing from it, settled).
- Races now all legal: warrior and paladin Undead, bear Night Elf / Night Elf / Tauren, Warden Troll / Troll / Tauren.
  -20261105-2 resolved.
- Consumables: same tier for all; flask, food, Zanza and potion differ per preset by each worker's search. Paladin Max
  survival now on Titans (-20261126-1 resolved). The paladin still has no potion action (open -20261105-1).
- On-use trinkets: everyone wears Chained Essence of Eranikus 10455 and/or Adaptive Combat Assistant 272437. The
  warrior casts them by explicit APL lines (no autocast); the other three rely on `autocastOtherCooldowns`, which fires
  them. Equal. Before 0741191ab the warrior cast the synthetic set's ids, so it went without its trinkets: that is most
  of its +267 / +297 / +131 this pass.
- Stoneform: only the warrior's APLs cast it (open -20261124-1; the Dwarf race-table rows only).

## The numbers (two-seed mean, own column; pass 34 in brackets)

| weight | order | scores |
|---|---|---|
| Max threat (1/3) | Bear, Warrior, Paladin, Warden | 5603 / 4720 / 3930 / 3413 (5603 / 4455 / 3913 / 3400) |
| Balanced (1) | Bear, Warrior, Paladin, Warden | 3775 / 3363 / 2952 / 1996 (3696 / 3058 / 2952 / 2024) |
| Max survival (3) | Warrior, Bear, Paladin, Warden | 1192 / 1047 / 899 / -1214 (1079 / 1047 / 780 / -1401) |

Seed 1 against the pass-34 baseline: bear Max threat, Max survival and paladin Balanced identical; bear Balanced +89
(re-search), paladin Max survival +108 (Titans), paladin Max threat +22 (Zanza), warrior +267 / +297 / +131 (trinkets),
Warden +5 / -20 / +171 (legal races and gear re-search; Max survival on Tauren gains).

TPS at Balanced: bear 2823, warrior 2210, paladin 1859, Warden 1735. Healing needed at Balanced: paladin 447,
warrior 504, Warden 704, bear 761.

- The bear leads Balanced by 12% over the warrior (21% in pass 34): the warrior's trinkets closed most of the gear gap.
  The bear's TPS lead (28%) is Lacerate and Primal Bite threat (sourced, pass 33) plus its Phase 0 weapon.
- Warrior first at Max survival, 14% over the bear; the paladin third (899) with Titans, its 8271 health against the
  warrior's 10269.
- The Warden: 47% behind the bear at Balanced, below 0 at Max survival. A class gap by user ruling (standards 40).
- Logs: boss files unchanged since the last comparison; no new disagreement to report.

## What each preset rewards (two-seed mean, 3,000 iterations; rows = build, columns = weight)

| build | Max threat | Balanced | Max survival |
|---|---|---|---|
| bear Balanced / Max threat / Max survival | 5210 / 5603 / 3536 | 3775 / 3889 / 2914 | -530 / -1250 / 1047 |
| paladin | 3780 / 3930 / 3078 | 2952 / 2971 / 2533 | 468 / 93 / 899 |
| warrior | 4352 / 4720 / 3262 | 3363 / 3215 / 2744 | 396 / -1301 / 1192 |
| Warden | 3302 / 3413 / 2752 | 1996 / 1761 / 1761 | -1924 / -3195 / -1214 |

Each preset wins its own weight except two Balanced presets:
- bear: Max threat build ahead at w 1 by 114 (3%), narrower than 193 (open -20261125-1, new evidence).
- paladin: Max threat build ahead at w 1 by 31 at 10,000 iterations (2977.7 / 2948.0 and 2971.7 / 2939.3), since its
  Spirit of Zanza (new -20261126-3).
The three builds differ for every spec.

## Not re-checked

Price coverage (-20261126-2): 12e6132b2 notes the Warden's final-price flags; the lists weren't re-walked this pass.

## New findings

- tank_fairness-20261126-3 (low): paladin Max threat build beats Balanced at w 1 (+31, 1%).

Resolved: -20261126-1 (52572a1d7), -20261105-2 (3ea8c0ecc).
