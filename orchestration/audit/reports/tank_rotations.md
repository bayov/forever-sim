# tank_rotations pass 10

Master 4da607cbc (branch tankheal, 2026-11-12), against pass 9's cb0d928fd / state 71994fca6. Forced re-audit after
the tank fixes. 1,000 iterations, default seed unless noted; gains are paired (abtest / rotopt) at the preset's own
weight on the six scoring bosses.

## What changed

- ee4a3df64: Warden re-tune (Max threat talents, Searing / Fire Nova all fight; Max survival Healing Stream), flasks
  re-picked (Titans at warrior / bear / Warden Max survival and Warden Balanced, Supreme Power at paladin Max
  survival, Wizard Oil for all paladins).
- f1c07e13a: Judgement is a ranged attack (sim/paladin/judgement.go, seals).
- 814decce0: Greater Agility boots (warrior, Warden), Strength wrists (warrior Max threat / Balanced), Warden
  Agility cloak (Max threat / Balanced) and Max threat wrists.
- 65b5d49cf: healer-side health reads (no tank effect). No encounter, tankbench or tankscore change.
- e89962738: user ruling, only white swings roll Windfury Totem (standards 17); code still lets Heroic Strike roll it
  (note under 20261110-1).

## Baseline (scoring mean, own weight; pass 8 ref -> now)

| request | score | SafeDPS | DTPS | healing needed |
|---|---|---|---|---|
| bear Max threat / Balanced | 2654.8 / 1717.0, identical | | | |
| bear Max survival | -578.3 -> -468.7 | 1343 -> 1265 | 545 -> 547 | 393 -> 361 |
| paladin Max threat | 3171.8 -> 3191.2 | 1694 -> 1702 | 306 -> 306 | 257 -> 251 |
| paladin Balanced | 2753.2 -> 2781.7 | 1689 -> 1694 | 314 -> 314 | 245 -> 238 |
| paladin Max survival | 1554.0 -> 1744.2 | 1504 -> 1621 | 298 -> 296 | 197 -> 200 |
| Warden Max threat | 2495.7 -> 2539.5 | 1626 -> 1660 | 718 -> 691 | 823 -> 840 |
| Warden Balanced | 1124.7 -> 1146.0 | 1557 -> 1516 | 675 -> 674 | 717 -> 681 |
| Warden Max survival | -2736.0 -> -2381.8 | 1503 -> 1467 | 685 -> 685 | 691 -> 644 |
| warrior Max threat | 3957.7 -> 3970.8 | 2196 -> 2209 | 415 -> 420 | 443 -> 456 |
| warrior Balanced | 3226.5 -> 3239.7 | 2101 -> 2114 | 379 -> 382 | 343 -> 347 |
| warrior Max survival | 1561.8 -> 1605.3 | 1822 -> 1734 | 337 -> 335 | 252 -> 227 |

Every move lines up with an intended request change (flask, imbue, enchants, Warden re-tune) or the ranged
Judgement; rows without one (bear Max threat / Balanced, legacy warrior_protection) are identical. New
ref/tank_bench/summary.tsv written (master 4da607cbc).

## Changes tried

Rotopt knob searches (1,000 a candidate, 3,000 to confirm) from each preset's knobs, after checking that `-eval` writes
an APL equal to the request's (`$AUDIT_WORK/tank_rotations/match.sh`):

| preset | result |
|---|---|
| paladin Max threat / Balanced / Max survival (bulwarkAt frozen) | every knob kept (judgementMana, consecration, Holy Shield, seal) |
| paladin Max survival talent search (Supreme Power, Wizard Oil) | `04-25300102513501531-5023` kept |
| warrior Max threat | kept |
| warrior Balanced | `tcRefresh=6`: +5.8 +- 2.1 / +4.3 +- 2.1 at 10k (finding 4) |
| warrior Max survival | `wallAt=13`: +21.9 +- 5.5 / +16.5 +- 6.5 at 10k, Vaelastrasz +215 (finding 3) |
| Warden Max threat | kept |
| Warden Balanced | `waterTotem=2`: +17.2 +- 4.9 / +28.8 +- 4.6 at 10k (finding 2) |

Bear Max survival (the template has no potion line, so hand A/Bs on the request with mk_bear.py, 1,000):

| change | diff |
|---|---|
| FR at 90 / 80 / 70 / 60 / 40 Rage (now 100) | +143.5 / +283.7 / +404.4 / +411.8 / -44.8 |
| FR at 70, 3,000 seed 1 / 2 | +385.3 +- 18.2 / +361.1 +- 17.8 |
| FR at 60, 3,000 seed 1 / 2 | +393.0 +- 22.4 / +367.8 +- 25.6 |
| FR at 60 on the pre-Titans request (71994fca6) | +339.5 +- 39.3 |
| no FR / no Enrage / no potion / Lacerate filler | -897.0 / -268.5 / -516.7 / -225.8 |
| Enrage under 20 / 60, Maul at 15 / 20 | 0.0 / +0.4, +13.4 +- 6.3 / +17.2 +- 11.3 |
| bear Balanced FR 60 / 80 / 90; Max threat FR 60 added | -130.9 / -28.8 / +12.2; -149.4 |

Not run this pass: talent searches for warrior, bear and Warden Balanced (the Warden's ran on ee4a3df64), and the
bear knob search through rotopt (no potion knob in druid_tank).

## Findings

New: 20261112-1 (high) bear Max survival FR at 70 Rage, +361 to +385; 20261112-2 (medium) Warden Balanced Healing
Stream, +17 to +29; 20261112-3 (medium) warrior Max survival Shield Wall 13 s, +16 to +22 (Vaelastrasz); 20261112-4
(low) warrior Balanced tcRefresh 6, +4 to +6. Open, unchanged: 20261110-1 (now ruled; code still lets Heroic Strike
roll the totem), 20261104-11, 20261104-3.

Raw output: `$AUDIT_WORK/tank_rotations/` (knobs/, bear/, ab10k_*.txt, talents_pal_ms.txt).
