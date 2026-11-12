# Windfury Totem: which attacks roll it under Forever (2026-11-12)

The user's ruling of 2026-11-12, from their log audit of Nov 6-11 (`tools/logcheck/findings.md`, "Windfury Totem procs
from main hand abilities", 4,737 procs on 93 players): Windfury Totem has a 20% chance (100 ms cooldown, unchanged) on
white main hand swings and on every melee ability whose client spell requires a main hand weapon (SpellMisc
Attributes_3 0x400) or is on next swing (Attributes_0 0x4). Landed hits with no damage count. Nothing else rolls it:
no other ability, seal proc, Judgement, poison, item proc, off-hand swing, Auto Shot or extra swing. Classic keeps every
main hand hit. Standards 17; replaces the white-swings-only rule (0ce9dbf40, 2026-11-08) and the "nothing whose spell is
also a special attack" gate of earlier today.

## What the sim does

- `tools/gen_main_hand_spells.py` writes `sim/core/client_main_hand_spells_auto_gen.go`: the 317 spells of client
  1.60.1.70888+hf whose SpellMisc row has either flag.
- `RegisterSpell` sets `Spell.ClientMainHand` from the ActionID's spell. An ability whose ActionID isn't its client cast,
  or whose logs ever contradict its flag, sets the field itself (none needs to today).
- `RollsForeverWindfuryTotem` (`sim/core/buffs.go`): a white main hand swing that isn't an extra swing (the new
  `WeaponAttack.extraSwing`, set while extra attacks swing), or a main hand melee hit with `ClientMainHand`.
- Tests: `sim/windfury_totem_rule_test.go` pins the log list on both sides, on every spell the 68 report requests and
  13 tank requests register under Forever (white main hand swings roll; off-hand swings and Auto Shot don't), plus the
  client flags of Ghostly Strike, Kidney Shot, Shield Slam and Cheap Shot, which no build registers.
  `sim/warrior/tank_warrior/windfury_totem_test.go` lands Heroic Strike, Cleave, Sunder Armor, Revenge, Execute and Rend
  (about a fifth proc), Whirlwind (never), and Classic's Heroic Strike and Whirlwind (a fifth).

## Every melee ability against the flag

Every ability the logs name lands on its logged side by its flag; no ability's flag contradicts the logs. Checked on
every melee spell the report and tank requests register under Forever, plus the logged abilities no build registers
(Ghostly Strike, Kidney Shot, Shield Slam, Cheap Shot) by their client flag only. Mocking Blow isn't in the sim.

| Rolls (flag) | Never (no flag) |
|---|---|
| Logged: Heroic Strike (on next swing + main hand), Cleave (on next swing + main hand), Maul (on next swing), Raptor Strike (on next swing + main hand), Sunder Armor, Revenge, Execute (cast 20662; its damage spell 20647 isn't flagged, the sim's hit is under the cast's ID), Rend, Sinister Strike, Eviscerate, Wing Clip, Ghostly Strike, Kidney Shot, Mongoose Bite | Logged: Stormstrike, Holy Strike, Bloodthirst, Whirlwind, Shield Slam, Claw, Swipe, Rake, Cheap Shot, Seal of Command, Seal of Righteousness, Seal of Fury, Judgements, Windfury Weapon's attacks |
| Follow the flag (no log): Mortal Strike, Slam, Hamstring, Spearing Strike, Backstab, Ambush, Hemorrhage, Mutilate (main hand hit; the off-hand hit never), Rupture, Expose Armor, Gouge, Mocking Blow | Follow the flag (no log): Overpower, Pummel, Shield Bash, Riposte, Garrote, Feint, Death Mark, Shred, Ravage, Rip, Ferocious Bite, Bash, Lacerate (bear and hunter), Primal Bite, Strider Kick |

## Before and after

Before = tankheal HEAD 0af274ed6, after = the change. Paired: same seed (1), 3,000 iterations.

## Tank presets (tankbench, 3,000 iterations, seed 1, paired; mean over the 7 encounters in tools/encounters)

"Balanced score" is the objective at the Balanced weight (w = 1) for every request. The paladin (Holy Strike) and
Warden (Stormstrike) presets don't move.

| Request | SafeDPS before | after | change | TPS before | after | DPS before | after | Healing needed before | after | Balanced score before | after | change |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| druid_bear_balanced | 1508 | 1558 | +51 (+3.4%) | 1417 | 1474 | 527 | 554 | 687 | 676 | 992 | 1140 | +148 |
| druid_bear_max_survival | 1212 | 1247 | +35 (+2.9%) | 1171 | 1217 | 434 | 455 | 521 | 521 | 924 | 992 | +68 |
| druid_bear_max_threat | 1605 | 1662 | +57 (+3.6%) | 1496 | 1559 | 558 | 587 | 860 | 860 | 739 | 852 | +113 |
| paladin_protection_balanced | 1656 | 1656 | +0 (+0.0%) | 1478 | 1478 | 645 | 645 | 445 | 445 | 2032 | 2032 | +0 |
| paladin_protection_max_survival | 1584 | 1584 | +0 (+0.0%) | 1415 | 1415 | 620 | 620 | 405 | 405 | 2007 | 2007 | +0 |
| paladin_protection_max_threat | 1665 | 1665 | +0 (+0.0%) | 1488 | 1488 | 649 | 649 | 481 | 481 | 1944 | 1944 | +0 |
| shaman_warden_balanced | 1486 | 1486 | +0 (+0.0%) | 1323 | 1323 | 718 | 718 | 834 | 834 | 560 | 560 | +0 |
| shaman_warden_max_survival | 1443 | 1443 | +0 (+0.0%) | 1300 | 1300 | 705 | 705 | 806 | 806 | 561 | 561 | +0 |
| shaman_warden_max_threat | 1629 | 1629 | +0 (+0.0%) | 1455 | 1455 | 807 | 807 | 1082 | 1082 | 113 | 113 | +0 |
| warrior_protection | 1004 | 1041 | +36 (+3.6%) | 1221 | 1308 | 295 | 325 | 982 | 976 | -843 | -760 | +84 |
| warrior_protection_balanced | 2020 | 2187 | +167 (+8.3%) | 1876 | 2049 | 506 | 575 | 551 | 555 | 2390 | 2710 | +321 |
| warrior_protection_max_survival | 1682 | 1788 | +105 (+6.3%) | 1626 | 1763 | 444 | 498 | 416 | 418 | 2121 | 2322 | +201 |
| warrior_protection_max_threat | 2101 | 2310 | +209 (+9.9%) | 1932 | 2146 | 519 | 602 | 683 | 673 | 2127 | 2576 | +448 |

Per encounter (SafeDPS before -> after):

| Request | broodlord | chromaggus | golemagg | magmadar | onyxia | ragnaros | vaelastrasz |
|---|---|---|---|---|---|---|---|
| druid_bear_balanced | 1511 -> 1562 | 1406 -> 1459 | 1461 -> 1510 | 1468 -> 1517 | 1565 -> 1619 | 1431 -> 1476 | 1712 -> 1765 |
| druid_bear_max_survival | 1210 -> 1245 | 1130 -> 1168 | 1171 -> 1205 | 1172 -> 1209 | 1259 -> 1291 | 1150 -> 1184 | 1394 -> 1429 |
| druid_bear_max_threat | 1608 -> 1668 | 1492 -> 1544 | 1553 -> 1612 | 1557 -> 1621 | 1659 -> 1711 | 1513 -> 1568 | 1856 -> 1913 |
| paladin_protection_balanced | 1653 -> 1653 | 1554 -> 1554 | 1597 -> 1597 | 1629 -> 1629 | 1763 -> 1763 | 1430 -> 1430 | 1964 -> 1964 |
| paladin_protection_max_survival | 1584 -> 1584 | 1485 -> 1485 | 1527 -> 1527 | 1559 -> 1559 | 1688 -> 1688 | 1368 -> 1368 | 1877 -> 1877 |
| paladin_protection_max_threat | 1662 -> 1662 | 1561 -> 1561 | 1604 -> 1604 | 1636 -> 1636 | 1769 -> 1769 | 1445 -> 1445 | 1976 -> 1976 |
| shaman_warden_balanced | 1524 -> 1524 | 1250 -> 1250 | 1429 -> 1429 | 1455 -> 1455 | 1592 -> 1592 | 1320 -> 1320 | 1832 -> 1832 |
| shaman_warden_max_survival | 1469 -> 1469 | 1303 -> 1303 | 1392 -> 1392 | 1404 -> 1404 | 1528 -> 1528 | 1304 -> 1304 | 1700 -> 1700 |
| shaman_warden_max_threat | 1654 -> 1654 | 1454 -> 1454 | 1558 -> 1558 | 1585 -> 1585 | 1742 -> 1742 | 1450 -> 1450 | 1959 -> 1959 |
| warrior_protection | 988 -> 1028 | 1050 -> 1091 | 1006 -> 1043 | 995 -> 1026 | 1005 -> 1039 | 970 -> 1006 | 1017 -> 1053 |
| warrior_protection_balanced | 2011 -> 2167 | 1961 -> 2138 | 1987 -> 2150 | 1991 -> 2149 | 2070 -> 2239 | 1942 -> 2124 | 2181 -> 2344 |
| warrior_protection_max_survival | 1660 -> 1764 | 1685 -> 1805 | 1663 -> 1772 | 1661 -> 1756 | 1699 -> 1795 | 1639 -> 1759 | 1770 -> 1863 |
| warrior_protection_max_threat | 2118 -> 2319 | 2013 -> 2244 | 2078 -> 2280 | 2084 -> 2297 | 2171 -> 2366 | 1999 -> 2210 | 2247 -> 2455 |

## DPS report builds (tools/audit/run_reqs.py, 3,000 iterations, seed 1, each request's own encounter)

19 of 68 builds moved; the other 49 are identical to the hundredth. Feral (Shred, Claw, Rake, Rip, Ferocious Bite), Enhancement
(Stormstrike, Windfury Weapon) and Retribution (Holy Strike, seals) use no flagged ability, so the new rule leaves
them where the white-swings rule had them; leaving out extra swings moved no build at this precision either. Casters
and healers don't melee.

| Build | DPS before | after | change | % |
|---|---|---|---|---|
| warrior__arms_dragonkin | 923.2 | 1101.0 | +177.8 | +19.3% |
| warrior__arms | 865.9 | 1041.1 | +175.3 | +20.2% |
| warrior__arms-duty-battle-shout | 863.5 | 1037.8 | +174.3 | +20.2% |
| warrior__arms-duty-sunder | 825.0 | 999.2 | +174.2 | +21.1% |
| warrior__fury_2h-duty-battle-shout | 918.1 | 1029.1 | +111.0 | +12.1% |
| warrior__fury_2h | 923.5 | 1032.8 | +109.3 | +11.8% |
| warrior__fury_2h-duty-sunder | 897.7 | 997.2 | +99.5 | +11.1% |
| warrior__fury-duty-sunder | 939.1 | 1016.8 | +77.7 | +8.3% |
| warrior__fury-duty-battle-shout | 971.6 | 1043.8 | +72.2 | +7.4% |
| hunter__survival | 917.0 | 988.6 | +71.6 | +7.8% |
| hunter__survival-wolf | 891.7 | 961.1 | +69.3 | +7.8% |
| hunter__survival-tallstrider | 885.6 | 954.3 | +68.7 | +7.8% |
| warrior__fury | 984.1 | 1049.3 | +65.2 | +6.6% |
| rogue__combat-swords | 899.5 | 945.6 | +46.1 | +5.1% |
| rogue__seal-fate-swords | 894.6 | 938.3 | +43.8 | +4.9% |
| rogue__subtlety | 839.6 | 863.7 | +24.1 | +2.9% |
| rogue__seal-fate-daggers | 883.1 | 905.2 | +22.1 | +2.5% |
| rogue__combat-daggers | 915.1 | 935.4 | +20.3 | +2.2% |
| rogue__mutilate | 846.3 | 866.0 | +19.7 | +2.3% |
