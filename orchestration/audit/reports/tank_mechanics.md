# tank_mechanics pass 34 (2026-11-25, phase0-tankheal 1624c3d56, fix 95962a7d3)

Forced pass (gate.py --force) on the Phase 0 tank branch. Client build and Blizzard posts not re-checked: pass 33
did both at 13:10 today (1.60.1.71091, hotfix thread post 15).

## What changed since pass 33

2f42f4a44 (tank threat batch), b663145ad / 295206ddf (procs, items), efeb06b95 (casters), 22d2cf952 (six rulings
to wowsims/forever's log), and the Phase 0 gear in every tank request.

## Bear threat batch (2f42f4a44) against wowsims/forever 59b8622 and the bench

| Rule | Ours | Theirs | Bench threat / damage | Verdict |
|---|---|---|---|---|
| Lacerate hit and ticks | 10/3, no flat | 10/3 (EV-0034, EV-0037) | 7.17 = 10/3 x 2.15 | right (their logs) |
| Primal Bite | 3x under Forever | 3x (EV-0034) | 6.45 = 3 x 2.15 | right (their logs) |
| Faerie Fire in Bear / Dire Bear | flat 2 x level, x form (258 at 60), none in Cat | the same | not cast by any bear APL | right; casting it before the Lacerate filler loses 40 to 313 (A/B below) |
| Blood Frenzy Rage threat | unscaled | unscaled (EV-0035) | - | right |
| Maul, Swipe | form only (2.15) | 1.75 on top, no log | Maul 2.15 | open call (verdicts.md) |
| Lacerate hit damage | was 10% x stacks before the cast | 10% x stacks after the cast (EV-0040) | - | **fixed, 95962a7d3** |

The bear's TPS jump is Lacerate and Primal Bite: on Magmadar (Balanced, 500 iterations) Lacerate is 797 TPS and
Primal Bite 571 of 2530; at the old 1x and 2x they would be 239 and 381, about 750 TPS less. The rule values match
theirs and the tests pin each (TestLacerate, TestPrimalBite, TestBearFaerieFireThreat, TestBearBloodFrenzyRageThreat).

Lacerate's hit (tank_mechanics-20261125-1): +170 to +185 score and +45 to +51 TPS on the three bear presets.

## Warrior

Shield Slam (shield_slam.go): 625-685 + 2 x block value, no attack power, flat 255, SuppressWeaponProcs and no
Windfury (EV-0004, EV-0050, EV-0155, user ruling 2026-11-25). Matches the commit's sources. Revenge still beats
Shield Slam in the priority (swap: -5.4 / -4.6 / -21.8, inside the error).

## Max survival on Phase 0 gear against synthetic: gear, not a bug

Paladin Max survival, same talents and APL, 2000 iterations:

| Boss | Gear | Max surv. | SafeDPS | Heal needed | DTPS | Health | Armor | Def | Miss / Dodge / Parry / Block | Crushes taken |
|---|---|---|---|---|---|---|---|---|---|---|
| Broodlord | Phase 0 | -1205 | 1792 | 532 (1.40x) | 381 | 8271 | 9391 | 382 | 9.7 / 18.4 / 12.7 / 7.7 | 9.8% |
| Broodlord | synthetic | 463 | 1602 | 305 (1.08x) | 283 | 8691 | 9816 | 463 | 12.9 / 20.3 / 15.9 / 17.9 | 2.5% |
| Chromaggus | Phase 0 | -2120 | 1667 | 612 | 502 | | | | | |
| Chromaggus | synthetic | -588 | 1501 | 403 | 363 | | | | | |

The synthetic set is a Phase 1 (Molten Core) budget: 127 Defense and 5% block on the gear, a Tier 1 2-piece (+7
Defense). The Phase 0 list has 53 Defense and no block chance, so block falls from 17.9% to 7.7%, avoidance plus block
from 67% to 48.5%, and crushing blows reach 9.8% of swings when Holy Shield is down (2.5% before). DTPS rises 35% and
the hits come spikier (1.40x DTPS for healing needed). At w = 3 the score is 2 x SafeDPS - 9 x healing needed, so
+190 SafeDPS cannot cover +227 healing needed. Block 7.7% = 5 + 82 x 0.04 - 0.6 (boss skill) checks by hand; no item
is missing from the database. Warrior (Def -87, armor -476) and bear (armor 9473 against 10159, dodge 35.0 against
38.1) on Chromaggus show the same: DTPS +17% and +9%, with no sim or scoring step out of line.

## Tests

go test for sim/druid/..., sim/warrior/..., sim/paladin/..., sim/shaman/..., sim/core and tools/tankscore: all pass.
