# tank_rotations pass 31 (2026-11-25, phase0-tankheal 1624c3d56, bear rows on 95962a7d3)

Forced pass on the Phase 0 tank branch. Baseline: every request on every boss, 1000 iterations, default seed;
ref/tank_bench/summary.tsv rewritten (bear rows on 95962a7d3, the Lacerate fix).

## Baseline, scoring mean (old = pass 30, master 0404484fb, synthetic gear)

| Request | Score old -> new | TPS | Healing needed | SafeDPS |
|---|---|---|---|---|
| druid_bear_balanced | 2153 -> 3646 | 1628 -> 2546 | 467 -> 687 | 1714 -> 2687 |
| druid_bear_max_survival | 432 -> 1015 | 1315 -> 1815 | 284 -> 340 | 1333 -> 1791 |
| druid_bear_max_threat | 3134 -> 5387 | 1739 -> 2797 | 635 -> 909 | 1854 -> 3023 |
| paladin_protection_balanced | 2809 -> 2915 | 1516 -> 1845 | 240 -> 448 | 1707 -> 2080 |
| paladin_protection_max_survival | 1787 -> 800 | 1446 -> 1623 | 201 -> 344 | 1627 -> 1826 |
| paladin_protection_max_threat | 3244 -> 3913 | 1543 -> 1972 | 268 -> 572 | 1735 -> 2218 |
| shaman_warden_balanced | 1238 -> 2026 | 1364 -> 1738 | 662 -> 699 | 1534 -> 1980 |
| shaman_warden_max_survival | -2283 -> -1337 | 1333 -> 1492 | 637 -> 573 | 1480 -> 1675 |
| shaman_warden_max_threat | 2600 -> 3418 | 1513 -> 1856 | 858 -> 884 | 1696 -> 2115 |
| warrior_protection_balanced | 3451 -> 3064 | 2066 -> 2135 | 346 -> 518 | 2215 -> 2301 |
| warrior_protection_max_survival | 1820 -> 1096 | 1761 -> 1726 | 214 -> 281 | 1804 -> 1771 |
| warrior_protection_max_threat | 4233 -> 4468 | 2175 -> 2394 | 472 -> 777 | 2348 -> 2623 |

Every row moved for a known cause: the Phase 0 gear in the requests (the round notes' Phase 0 against synthetic
tables), 2f42f4a44 (threat batch: bear TPS +46% from Lacerate and Primal Bite, warrior Shield Slam), the procs and
items batches, and 95962a7d3 (bear +170 to +185). No unexplained move. The report's Tanking mode has none of this
yet: tank_rotations-20261125-3.

## Lines tried (paired, 1000 iterations, six scoring bosses)

| Change | Balanced | Max threat | Max survival | Verdict |
|---|---|---|---|---|
| Bear: Faerie Fire (Feral) before the Lacerate filler (258 threat a cast now) | -208.8 ± 18.9 | -312.9 ± 6.9 | -39.9 ± 32.2 | keep out |
| Warrior: Shield Slam ahead of Revenge (Shield Slam now 2 x block value) | -5.4 ± 6.1 | -4.6 ± 2.1 | -21.8 ± 23.0 | keep |

## Open

- The Windfury main-hand timing item (passes/tank_rotations.md) is still not done.
- The paladin and Warden lines were not re-tried this pass. Their presets were searched today on the final sim code
  (the round notes).
