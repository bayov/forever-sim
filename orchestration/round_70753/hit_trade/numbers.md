# Round 70753 phase 5, hit trade: DPS before and after the regears

The 22 Level 60 builds the hit-trade phase moved, each simmed from its own report request
(`tools/dps_report/req/<id>.json`) with `wowsimcli sim` through `run_build.run` (the dashboard's path: the request's
180 ± 20 s fight with only the base duration set to 120, 180 or 300, `useLabeledRands`), 20,000 iterations, seed 1,
on the binaries built from this tree. "Old" is the request as of commit `75e1e9a13` (the phase 5 brief), "new" the
request on this branch; both run on the same binary, so they differ only by what changed in the request (gear step,
and the talents where the column says so). DPS ± standard error, at 120 / 180 / 300 s. D180 is new minus old at 180 s.

| Build | Step (set, old -> new) | Talents changed? | Old 120 / 180 / 300 | New 120 / 180 / 300 | D180 |
|---|---|---|---|---|---|
| druid__feral-cat | level60_phase1_synthetic -> level60_phase1_synthetic_hit-4 | yes: 05002-53200032023012125001-05502 -> 050022-53200032023012125001-055 | 1150.3±0.3 / 1139.1±0.3 / 1124.4±0.2 | 1176.6±0.3 / 1165.3±0.3 / 1146.7±0.2 | +26.2 |
| druid__feral-cat-duty-ff | level60_phase1_synthetic -> level60_phase1_synthetic_hit-4 | yes: 05002-53200032023012125001-05502 -> 050022-53200032023012125001-055 | 1143.2±0.3 / 1134.1±0.3 / 1120.2±0.2 | 1169.3±0.3 / 1160.1±0.3 / 1143.8±0.2 | +26.0 |
| mage__arcane | level60_phase1_synthetic_arcane -> level60_phase1_synthetic_arcane_hit-1 | yes: 050205203100311531-5551-003 -> 050205203100311531-5550000000001-003 | 985.5±0.4 / 962.1±0.3 / 944.6±0.3 | 992.9±0.4 / 969.3±0.3 / 952.3±0.3 | +7.2 |
| paladin__level_60 | level60_phase1_synthetic -> level60_phase1_synthetic_hit-3 | no | 1094.6±0.5 / 1091.7±0.4 / 1080.9±0.3 | 1106.6±0.5 / 1103.4±0.4 / 1092.2±0.3 | +11.7 |
| paladin__level_60-duty-crusader | level60_phase1_synthetic_crusader_duty -> level60_phase1_synthetic_crusader_duty_hit-3 | no | 1078.3±0.5 / 1083.4±0.4 / 1076.5±0.3 | 1090.0±0.5 / 1094.9±0.4 / 1087.9±0.3 | +11.5 |
| paladin__level_60-duty-wisdom | level60_phase1_synthetic -> level60_phase1_synthetic_hit-3 | no | 1069.5±0.5 / 1075.3±0.4 / 1068.5±0.3 | 1081.0±0.5 / 1086.6±0.4 / 1079.9±0.3 | +11.4 |
| rogue__combat-daggers | level60_phase1_synthetic_combat_daggers -> level60_phase1_synthetic_combat_daggers_hit+2 | no | 975.4±0.3 / 949.9±0.2 / 926.1±0.2 | 980.1±0.3 / 954.6±0.2 / 931.1±0.2 | +4.8 |
| rogue__combat-swords | level60_phase1_synthetic_swords -> level60_phase1_synthetic_swords_hit+2 | no | 990.5±0.3 / 959.0±0.3 / 932.2±0.2 | 995.8±0.3 / 964.5±0.3 / 937.8±0.2 | +5.5 |
| rogue__mutilate | level60_phase1_synthetic_mutilate -> level60_phase1_synthetic_mutilate_hit+3 | no | 873.2±0.3 / 858.3±0.2 / 847.2±0.2 | 878.0±0.3 / 863.4±0.2 / 852.4±0.2 | +5.1 |
| rogue__seal-fate-daggers | level60_phase1_synthetic_daggers -> level60_phase1_synthetic_daggers_hit+3 | no | 919.6±0.3 / 903.9±0.2 / 889.7±0.2 | 924.4±0.3 / 909.0±0.2 / 894.6±0.2 | +5.0 |
| rogue__seal-fate-swords | level60_phase1_synthetic_seal_fate_swords -> level60_phase1_synthetic_seal_fate_swords_hit+3 | no | 959.3±0.4 / 940.6±0.3 / 924.3±0.2 | 963.6±0.4 / 945.2±0.3 / 928.9±0.2 | +4.6 |
| rogue__subtlety | level60_phase1_synthetic_subtlety -> level60_phase1_synthetic_subtlety_hit+2 | no | 896.3±0.3 / 881.5±0.2 / 867.6±0.2 | 900.8±0.3 / 885.9±0.2 / 872.0±0.2 | +4.4 |
| shaman__enhancement-level60 | level60_phase1_synthetic -> level60_phase1_synthetic_hit-3 | yes: 353003102-055030110051012251-02 -> 353003102-055030110051012151-03 | 1167.1±0.6 / 1157.8±0.5 / 1140.3±0.4 | 1176.8±0.6 / 1164.6±0.5 / 1144.3±0.4 | +6.9 |
| shaman__enhancement-mana-tide | level60_phase1_synthetic_burning_totem -> level60_phase1_synthetic_burning_totem_hit-4 | yes: 03-055030310051012051-053051001001 -> 03-055030210051012151-053111031001 | 1064.5±0.5 / 1054.6±0.4 / 1046.5±0.3 | 1068.3±0.5 / 1059.5±0.4 / 1054.4±0.3 | +4.9 |
| shaman__enhancement-water-shield | level60_phase1_synthetic -> level60_phase1_synthetic_hit-4 | yes: 052002-054030210051012151-053020001 -> 052002-055030110051012151-053020001 | 1119.3±0.5 / 1111.3±0.4 / 1107.6±0.3 | 1124.6±0.6 / 1116.7±0.5 / 1113.1±0.3 | +5.5 |
| warlock__destruction_fire | level60_phase1_synthetic -> level60_phase1_synthetic_hit+1 | yes: 252-0050003201-0050345003101351 -> 242-0050003201-0050355003101351 | 1002.4±0.3 / 996.4±0.2 / 976.0±0.2 | 1002.7±0.3 / 997.3±0.2 / 977.5±0.2 | +0.9 |
| warlock__destruction_fire-duty-curse | level60_phase1_synthetic -> level60_phase1_synthetic_hit+1 | yes: 252-0050003201-0050345003101351 -> 242-0050003201-0050355003101351 | 986.9±0.3 / 986.8±0.2 / 968.9±0.2 | 987.4±0.3 / 987.7±0.2 / 970.0±0.2 | +0.9 |
| warlock__destruction_fire-duty-recklessness | level60_phase1_synthetic -> level60_phase1_synthetic_hit+1 | yes: 252-0050003201-0050345003101351 -> 242-0050003201-0050355003101351 | 984.5±0.3 / 980.4±0.2 / 962.6±0.2 | 985.0±0.3 / 981.5±0.2 / 963.8±0.2 | +1.1 |
| warrior__fury | level60_phase1_synthetic_fury -> level60_phase1_synthetic_fury_hit+2 | yes: 30313113002-0505300551120100321 -> 30313113002-050530055120100231-1 | 894.9±0.5 / 897.2±0.4 / 880.9±0.3 | 1181.7±0.5 / 1137.5±0.4 / 1111.6±0.3 | +240.4 |
| warrior__fury-duty-battle-shout | level60_phase1_synthetic_fury -> level60_phase1_synthetic_fury_hit+2 | yes: 30313113002-0505300551120100321 -> 30313113002-050530055120100231-1 | 888.2±0.5 / 890.7±0.4 / 870.8±0.3 | 1172.1±0.5 / 1126.9±0.4 / 1097.1±0.3 | +236.2 |
| warrior__fury-duty-sunder | level60_phase1_synthetic_fury -> level60_phase1_synthetic_fury_hit+2 | yes: 30313113002-0505300551120100321 -> 30313113002-050530055120100231-1 | 872.9±0.5 / 870.2±0.4 / 854.7±0.3 | 1141.4±0.5 / 1100.1±0.4 / 1077.0±0.3 | +229.9 |

**Warrior Fury (and its two duties) is not a pure regear.** The request at `75e1e9a13` is from before master's hotfix
merge (tree, Spearing Strike, Raging Blows out, Reckless Swings on the GCD: commits `7cd399e88`, `e07adf11b`), so its
D180 of about +230 to +240 is mostly that hotfix's rotation and tree, not the hit step. The request just before the
regear (`cd3f44cf2^`, hotfix in, set `level60_phase1_synthetic_fury`, same binary) gives the clean comparison:

| Build | Before the regear 120 / 180 / 300 | New 120 / 180 / 300 | D180 (regear only) |
|---|---|---|---|
| warrior__fury | 1179.3±0.5 / 1135.8±0.4 / 1110.9±0.3 | 1181.7±0.5 / 1137.5±0.4 / 1111.6±0.3 | +1.7 |
| warrior__fury-duty-battle-shout | 1168.7±0.5 / 1124.2±0.4 / 1096.4±0.3 | 1172.1±0.5 / 1126.9±0.4 / 1097.1±0.3 | +2.7 |
| warrior__fury-duty-sunder | 1138.4±0.5 / 1097.4±0.4 / 1075.4±0.3 | 1141.4±0.5 / 1100.1±0.4 / 1077.0±0.3 | +2.7 |

Other notes: the Warlock Destruction Fire builds also changed talents (Fire 2/5 and one Destruction point moved); their
gain is small (+0.9 to +1.1). Shaman Mana Tide and Water Shield, the Feral cats and the Arcane mage re-spent talents as
shown. The rogues, paladins keep their talents; their numbers match the class result files within noise.
