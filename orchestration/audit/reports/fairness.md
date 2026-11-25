# Fairness pass 56 (2026-11-25, master 63129ce35)

Result: findings (1 new, 1 resolved, 2 updated). Range eceb09019..63129ce35: Phase 0 step 4 regears of the
hunter (10 requests, merged 3fdc2dca4) and rogue (6 requests, merged 6fdbcc951) requests with faction race sets;
Shadow priest talents re-searched (a2d195d46: Wand Specialization 2, Mental Strength 1, Meditation 1). No sim/,
proto, database or standards change. Full pass.

## What I ran

- 5k race sweep, every race, seed 1, report requests (CLI /tmp/audit-cache/cli/63129ce35290). New baseline:
  `ref/fairness/sweep_baseline.log`, `table_baseline.md` (63129ce35).
- audit1b: 19/19 "5a PASS | 1b PASS" (it checks the synthetic sets, which no report request wears any more).
- Check that no request carries a 99xxxx synthetic item: none does.
- Paired 20k (seeds 1 / 1000003) for the new race flags, with the committed race file where one exists.
- The loop's 20k vs_published comparison: 17 of 68 builds move more than 0.5, at most 1.1 (warrior__arms).

## Findings

| id | sev | what |
|---|---|---|
| fairness-20261125-1 | medium | MM: High Order 1096.5 vs Orc 1090.4 (+6.1, paired 20k); wolf variant +7.0. BM a tie. |
| fairness-20261124-1 | resolved | Every class now on Phase 0 gear. Hunters -8.0 to -9.3%, rogues -5.7 to -11.0%. |
| fairness-20261124-2 (update) | low | Shadow on its new talents: Troll 859.5 vs Undead 857.6 (+1.8). Fire unchanged (+2.9 in 5k). |
| fairness-20261118-1 (update) | medium | Balance unchanged; still open. |

## Race checks (paired 20k, mean of two seeds)

| Build | Report race | Challenger | Delta |
|---|---|---|---|
| hunter__marksmanship | Orc 1090.4 | High Order (race file, Alliance set) 1096.5 | +6.1 |
| hunter__marksmanship-wolf | Orc 1070.4 (seed 1) | High Order 1077.4 | +7.0 |
| hunter__beast-mastery | Orc 1009.5 | High Order (race file) 1009.3 | -0.2 |
| priest__shadow | Undead 857.7 | Troll 859.5 | +1.8 |

## Ranking (5k, report race)

Lone Wolf 1174.0, MM 1090.0, Feral 1055.4, Arms 1016.5, Fury 1007.5, BM 1009.3, Fury 2H 992.8, Enh 989.9,
Affliction 965.7, Ret 957.0, Survival 926.8, Fire 920.4, ... Smite 695.7. Top over bottom 1.69x (1.86x last pass,
when hunters still wore the Phase 1 sets). Full table: `ref/fairness/table_baseline.md` and fairness_report.md §1.

## Movement

- Every move is a hunter or rogue request regear, or Shadow's talents (+2.7 at Undead, a2d195d46). Every other
  build is identical to the eceb09019 sweep (same seed): no shared change leaked.
- Published (meta) against current: every base build and variant within 1.0 except Arms +1.7 and Fury 2H +2.5,
  both unchanged from last pass (warrior metas, results auditor's area).

## Notes

- BM variants (wolf, tallstrider) show High Order +2.2 / +2.4 in the 5k sweep; the base BM is a paired tie, so not
  flagged.
- The hunter notes (round_notes_2026-11-11/phase0_step4_hunter.md "Problem") already say BM/MM/SV Dwarf rows keep
  0.6 to 1% excess hit because race_gear.py skips the trade on faction sets. Not re-flagged; the Dwarf is not near
  the best race for any hunter (MM Dwarf 5k well below High Order).
