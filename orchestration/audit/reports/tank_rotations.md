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

---

## Master's pass, carried over from origin/master: tank_rotations pass 32

Merged in from master's audit loop (synthetic-gear tank/healer rows; numbers there are master's baseline, not the
Phase 0 one above). Kept for the verdicts and finding notes that exist only on master.

Master ab3fff7007 (2026-11-26), against pass 31's master bfbee78fcc. Quiet: no new finding.

### What changed

The 38 watched files the gate listed are the warrior leveling work (143bf4f72, 8f5b59325, 30bb3a590, 1223e764e,
c747e7ccd, 270731faa, 565f32155 and their tests): every rank at the level it's learned, leveling hooks (charges,
stuns, kills, Rage decay, abilities that reach only nearby mobs), Rage from hits taken by stance in leveling
sessions, and `ui/tank_warrior` leveling plan, weights and presets. `sim/core/debuffs.go` and `target_death.go`
changed with them. No tank request, level 60 UI apl, encounter, tankbench, tankscore or rotopt tank template
changed. The `ui/tank_warrior/presets.ts` and `sim.ts` diffs add only the leveling plan and its EP weights.

### Baseline (1,000 iterations, default seed; `ref/tank_bench/summary.tsv` rewritten)

All 104 rows (13 requests x 7 bosses, plus scoring means) are bit-identical to pass 31 on score, SafeDPS, TPS,
DTPS and healing needed, so the leveling code doesn't reach the level 60 tank sims. Scoring means:

| request | weight | score | SafeDPS | TPS | DTPS | healing needed |
|---|---|---|---|---|---|---|
| druid_bear_balanced | 1 | 3338.3 | 2309.0 | 2372.5 | 586.7 | 469.5 |
| druid_bear_max_survival | 3 | 1260.7 | 1754.0 | 1793.7 | 556.2 | 283.5 |
| druid_bear_max_threat | 0.333 | 4541.7 | 2558.5 | 2557.2 | 613.7 | 636.3 |
| paladin_protection_balanced | 1 | 2857.7 | 1720.8 | 1530.2 | 302.7 | 233.7 |
| paladin_protection_max_survival | 3 | 1858.0 | 1639.0 | 1456.8 | 283.8 | 196.7 |
| paladin_protection_max_threat | 0.333 | 3280.3 | 1749.7 | 1557.2 | 302.8 | 261.8 |
| shaman_warden_balanced | 1 | 1245.7 | 1534.3 | 1365.2 | 660.0 | 660.3 |
| shaman_warden_max_survival | 3 | -2234.0 | 1482.5 | 1336.2 | 670.8 | 630.3 |
| shaman_warden_max_threat | 0.333 | 2608.2 | 1697.0 | 1513.5 | 683.0 | 851.3 |
| warrior_protection | 1* | -399.2 | 854.5 | 1206.5 | 636.2 | 755.5 |
| warrior_protection_balanced | 1 | 3241.2 | 2108.8 | 1989.0 | 374.8 | 344.2 |
| warrior_protection_max_survival | 3 | 1617.0 | 1697.3 | 1687.2 | 321.0 | 213.0 |
| warrior_protection_max_threat | 0.333 | 4102.7 | 2281.2 | 2125.0 | 418.3 | 469.8 |

### Rotation lines

No rotation changed and the baseline is identical, so pass 31's line checks stand; no A/B re-run. The five open
findings (20261125-1 bear Roar, 20261125-2 warrior Balanced Thunder Clap, 20261122-1 paladin Consecration floor,
20261104-11 and 20261104-3 Broodlord user calls) are untouched on master and get a pass 32 note.

### Open items

- Windfury main-hand ability timing (pass doc, user request 2026-11-15): not done this pass either. It needs APL
  conditions on the swing timer for each tank request that takes the totem; left for a pass with room for it.

### Setup note

A fresh session needs `make sim/core/proto/api.pb.go` (Go protos) before `go build -tags=with_db ./tools/tankbench`;
`tools/audit/setup.sh` installs protoc and protoc-gen-go but doesn't generate them. The full bench (91 runs,
4 parallel) takes about 25 s.
