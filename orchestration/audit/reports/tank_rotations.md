# tank_rotations: pass 4 (2026-11-05, master a07941531)

Due on 32 watched files changed since dc303b0: tank rounds 4 and 5 (full raid consumables: Zanza, Rumsey Rum,
Goblin Sapper Charge, potions; talents and knobs re-searched), rotopt's tank templates (shaman `fireLast`,
warrior `heldForLast` and the sapper line), all 12 preset requests, the four classes' UI APLs and presets,
sim/core consumes / major_cooldown / spell_result, sim/druid/ferocious_bite.go (cat, not bear).

## Baseline (ref/tank_bench/summary.tsv, header now a07941531)
Every request on every boss, 1,000 iterations, default seed. All 91 rows equal, to the decimal, the round-5 ref
the workers committed, so nothing moved without a rotation, gear or consumable change. Scoring mean (six bosses
without Ragnaros) at own weight; the pass-3 column is from the last report. The rises are the round 4/5
consumables and re-searches, each with its commit (not split per change).

| request | w | score (pass 3 -> 4) | SafeDPS | DTPS | healing needed | death floor |
|---|---|---|---|---|---|---|
| druid_bear_balanced | 1 | 1852 -> 2498 | 1883 | 584 | 456 | 0.0 |
| druid_bear_max_survival | 3 | -634 -> 238 | 1649 | 544 | 372 | 0.0 |
| druid_bear_max_threat | 0.333 | 2842 -> 3492 | 2024 | 610 | 614 | 0.0 |
| paladin_protection_balanced | 1 | 2437 -> 2835 | 1724 | 314 | 241 | 0.0 |
| paladin_protection_max_survival | 3 | 1500 -> 1910 | 1611 | 295 | 180 | 0.0 |
| paladin_protection_max_threat | 0.333 | 2804 -> 3249 | 1731 | 305 | 254 | 0.0 |
| shaman_warden_balanced | 1 | 767 -> 1786 | 1857 | 677 | 694 | 0.0 |
| shaman_warden_max_survival | 3 | -3971 -> -1988 | 1788 | 686 | 674 | 0.0 |
| shaman_warden_max_threat | 0.333 | 2430 -> 3170 | 1949 | 721 | 795 | 0.0 |
| warrior_protection | 1* | -14 -> -14 | 1048 | 637 | 757 | 0.0 |
| warrior_protection_balanced | 1 | 3008 -> 3880 | 2386 | 376 | 315 | 0.0 |
| warrior_protection_max_survival | 3 | 1205 -> 2139 | 2059 | 336 | 240 | 0.0 |
| warrior_protection_max_threat | 0.333 | 3700 -> 4652 | 2535 | 415 | 427 | 0.0 |

Drift: in all 12 preset requests priorityList and prepullActions equal the UI APL.

## Open findings
- 3 (paladin Bulwark 13 s) and 11 (warrior Max survival armor trinket held 35 s): user calls, unchanged on
  master (Bulwark 10 s, 990617 held 35 s); not re-tested.

## Line-by-line probe (36 A/Bs, paired, 3,000 iterations, seed 21, scoring mean, B minus preset)
Aimed at what rounds 4-5 added (the sapper, the potions, the opener) plus a few thresholds.

| change | Max threat | Balanced | Max survival |
|---|---|---|---|
| warrior: sapper to the top line | +1.5 ± 2.9 | -12.2 ± 5.9 | 0.0 (no change) |
| warrior: sapper above the potion | +1.6 ± 2.8 | -11.4 ± 5.8 | 0.0 |
| warrior: Bloodrage with no Rage condition | +1.1 ± 0.7 | +0.2 ± 2.1 | +1.0 ± 3.9 |
| warrior: Bloodrage before the pull (-2 s) | -19.5 ± 2.1 | -22.9 ± 5.2 | -26.8 ± 12.4 |
| paladin: explicit sapper line on top | 0.0 | 0.0 | 0.0 (autocast already fires it at the pull) |
| paladin: Holy Shield above the Judgements | 0.0 | 0.0 | 0.0 (Judgements are off the GCD) |
| bear: explicit sapper line on top | -5.5 ± 1.6 | -1.5 ± 6.4 | -2.6 ± 5.8 |
| bear: Maul at 30 Rage | +0.2 ± 0.9 | -0.5 ± 4.0 | n/a (no such line) |
| bear: Maul at 10 Rage | -0.1 ± 0.7 | -5.8 ± 3.1 | n/a |
| Warden: explicit sapper line on top | -3.4 ± 0.2 | -3.3 ± 0.6 | 0.0 |
| Warden: no Greater Stoneshield line | -151.0 ± 1.6 | -373.8 ± 3.9 | -1124.1 ± 14.3 |
| Warden: Fire Nova at 25% mana (from 40%) | -2.9 ± 1.1 | -1.8 ± 3.3 | -1.5 ± 2.0 |

No change wins by twice its error; the best, warrior Max threat
Bloodrage unconditioned, +1.1 ± 0.7 (0.02%), is a tie. Note the potion row: removing the explicit potion line
also stops the potion (autocast doesn't take it), so the line is load-bearing.

## Conclusion
Quiet. No new findings; findings 3 and 11 stay open on their user calls. Scratch: /tmp/audit-work/tank_rotations/p4
(variant maker mk.py, outputs in out/).
