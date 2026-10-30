# Comp search: 10 players, 1 tanks, 2 healers

Generic boss, 180 ± 20 s, in the natural flask zone. 12736 single-player sims at 1000 iterations (contenders at 10000), 6971 from the cache, 983 s. Sim version 6c652f4a55a41916.

## Horde, camps off

**Best: 7259.9 raid DPS in the raid sims** (±2.1 at 95%; score 7262.0), Protection paladin, Restoration shaman, Discipline priest.

A score is the sum of the members' single-player sims in their seats; a raid sim runs the whole raid.

1. **7259.9** in the raid sims (the best; score 7262.0, the best score): 5× mage__fire, 1× warlock__demonology-duty-curse (CurseOfTheElements), 1× druid__balance; Protection paladin, Restoration shaman, Discipline priest
2. **7243.4** in the raid sims (16.5 behind the best, ±3.1 at 95%; score 7242.6, 19.4 behind the best score, ±6.7 at 95%): 1× warlock__demonology-duty-curse (CurseOfTheElements), 1× druid__balance, 5× mage__fire; Protection paladin, Holy paladin, Discipline priest
3. score **7242.2** (19.8 behind the best score, ±6.4 at 95%): 5× mage__fire, 1× warlock__demonology-duty-curse (CurseOfTheElements), 1× druid__balance; Protection paladin, Restoration shaman, Holy priest
4. score **7222.8** (39.2 behind the best score, ±7.6 at 95%): 1× warlock__demonology-duty-curse (CurseOfTheElements), 1× druid__balance, 5× mage__fire; Protection paladin, Holy paladin, Holy priest
5. score **7206.7** (55.3 behind the best score, ±6.3 at 95%): 5× mage__fire, 1× warlock__demonology-duty-curse (CurseOfTheElements), 1× priest__shadow; Protection paladin, Restoration shaman, Restoration druid

| Party | Member | Race | DPS | Duties and choices | Gets |
|---|---|---|---|---|---|
| 1 | Restoration shaman |  |  | StrengthOfEarth, ManaSpring, Flametongue, GraceOfAir |  |
| 1 | mage__fire | Undead | 1106.6 |  | Kings, Restoration, Power Infusion |
| 1 | mage__fire | Undead | 1086.8 |  | Kings, Restoration |
| 1 | mage__fire | Undead | 1086.8 |  | Kings, Restoration |
| 1 | mage__fire | Undead | 1086.8 |  | Kings, Restoration |
| 2 | Protection paladin |  |  | judges Wisdom |  |
| 2 | Discipline priest |  |  |  |  |
| 2 | warlock__demonology-duty-curse | Orc | 1005.5 | CurseOfTheElements (costs 9.8) | Kings |
| 2 | druid__balance | SkyborneWindshaper | 845.4 |  | Kings |
| 2 | mage__fire | Undead | 1044.2 |  | Kings |

**Roles** at the winning set (re-filled without each, or with it when it's missing):

| Role | Gain | Provider | Provider cost | Net |
|---|---|---|---|---|
| Battle Shout (missing) | | value of adding | | -465.3 |
| Furious Howl (missing) | | value of adding | | -260.0 |
| Sunder upkeep (missing) | | value of adding | | -340.0 |
| Minor armor slot (missing) | | value of adding | | -29.4 |
| Curse of the Elements | 658.5 | warlock__demonology-duty-curse | 38.6 | 619.9 |
| Judgement of the Crusader (missing) | | value of adding | | -256.1 |
| Judgement of Wisdom | | placeholders (not measured: they give it at no cost) | | not measured |
| Moonkin Aura (with Gift of the Wild) (its giver also gives Gift of the Wild, and its cost counts in each) | 286.9 | druid__balance | 197.3 | 89.6 |
| Leader of the Pack (missing) | | value of adding | | -436.4 |
| Trueshot Aura (missing) | | value of adding | | -237.6 |
| Arcane Brilliance | 308.4 | mage__fire | 0.0 | 308.4 |
| Gift of the Wild (with Moonkin Aura) (its giver also gives Moonkin Aura, and its cost counts in each) | 286.9 | druid__balance | 197.3 | 89.6 |
| Prayer of Spirit | | placeholders (not measured: they give it at no cost) | | not measured |
| Enhancement's totems (Strength of Earth, Windfury) (missing) | | value of adding | | -336.3 |
| a shaman placeholder's Grace of Air over Windfury Totem | 0.0 | placeholders | 0.0 | 0.0 |

**Fillers** under the winning set, without a package:

| Build | DPS | ±SE | Assumption range | |
|---|---|---|---|---|
| mage__fire | 1043.8 | 2.2 | 1008 to 1050 | interchangeable with the best |
| mage__frostfire | 1025.5 | 2.6 | 942 to 1031 |  |
| warlock__demonology | 1015.1 | 1.5 | 935 to 1033 |  |
| warlock__affliction | 985.6 | 1.7 | 911 to 1006 |  |
| warlock__destruction_fire | 970.1 | 1.0 | 900 to 987 |  |
| warlock__destruction | 969.9 | 1.5 | 904 to 989 |  |
| priest__shadow | 951.1 | 0.8 | 938 to 955 |  |
| mage__frost | 907.8 | 1.7 | 844 to 924 |  |
| priest__shadow-holy | 890.8 | 1.0 | 879 to 895 |  |
| mage__arcane | 885.9 | 1.4 | 834 to 898 |  |
| druid__balance | 846.5 | 1.0 | 829 to 853 |  |
| hunter__beast-mastery | 822.7 | 0.8 | 781 to 824 |  |
| druid__balance-wrath | 810.6 | 0.9 | 796 to 822 |  |
| shaman__elemental-level60 | 767.7 | 0.9 | 759 to 771 |  |
| shaman__elemental-level60-mana-tide | 757.0 | 0.8 | 747 to 760 |  |
| priest__smite | 672.1 | 0.9 | 657 to 676 |  |
| hunter__survival | 648.5 | 0.7 | 596 to 649 |  |
| warrior__fury | 559.9 | 1.1 | 480 to 612 |  |
| warrior__arms | 500.3 | 0.9 | 461 to 526 |  |
| rogue__combat-daggers | 497.7 | 0.6 | 461 to 498 |  |
| warrior__fury_2h | 488.3 | 1.0 | 404 to 544 |  |
| rogue__subtlety | 482.1 | 0.5 | 450 to 482 |  |
| rogue__combat-swords | 470.6 | 0.6 | 443 to 471 |  |
| rogue__seal-fate-daggers | 469.4 | 0.6 | 442 to 469 |  |
| rogue__mutilate | 461.5 | 0.5 | 434 to 462 |  |
| rogue__seal-fate-swords | 449.1 | 0.6 | 418 to 449 |  |

**Member values** (one row per distinct member):

| Member | Own DPS | Contribution | Replacement margin | Best replacement | vs the report |
|---|---|---|---|---|---|
| warlock__demonology-duty-curse | 1005.3 | 1569.8 | 25.6 | warlock__affliction-duty-curse | -46.5 (the report's request in this sim: +0.8; race (Orc): +0.0; this fight (boss, length, flask zone): +0.0; the raid's buff set: -37.6; its duty (Curse of the Elements): -9.8) |
| druid__balance | 846.5 | 1087.6 | 35.9 | druid__balance-wrath | -4.7 (the report's request in this sim: -0.1; race (SkyborneWindshaper): -0.7; this fight (boss, length, flask zone): +0.0; the raid's buff set: -3.9) |
| mage__fire | 1106.8 | 1087.2 | 6.9 | mage__frostfire | +17.8 (the report's request in this sim: +1.0; race (Undead): -3.9; this fight (boss, length, flask zone): +0.0; the raid's buff set: +20.7) |
| mage__fire (×3) | 1087.2 | 1087.2 | 6.9 | mage__frostfire | -1.8 (the report's request in this sim: +1.0; race (Undead): -3.9; this fight (boss, length, flask zone): +0.0; the raid's buff set: +1.2) |
| mage__fire | 1043.8 | 1043.8 | 18.3 | mage__frostfire | -45.2 (the report's request in this sim: +1.0; race (Undead): -3.9; this fight (boss, length, flask zone): +0.0; the raid's buff set: -42.2) |

**Value of adding** a build that isn't in the comp (swapped for the member it beats most):

- mage__frostfire: -6.9
- warlock__affliction: -25.6
- druid__balance-wrath: -35.9
- warlock__destruction: -42.5
- warlock__destruction_fire: -45.0
- priest__shadow: -136.1
- mage__frost: -178.2
- priest__shadow-holy: -178.3

**Tank and healer combos** (best first, by their best comp's screening score):

| Combo | Best comp's score | Behind |
|---|---|---|
| Protection paladin, Restoration shaman, Discipline priest | 7264.0 | 0.0 |
| Protection paladin, Restoration shaman, Holy priest | 7244.4 | 19.6 |
| Protection paladin, Holy paladin, Discipline priest | 7244.0 | 20.0 |
| Protection paladin, Holy paladin, Holy priest | 7224.5 | 39.5 |
| Protection paladin, Restoration shaman, Restoration druid | 7211.3 | 52.7 |
| Protection paladin, Holy paladin, Restoration shaman | 7179.8 | 84.2 |
| Protection paladin, Restoration shaman, Restoration shaman | 7173.0 | 91.0 |
| Warden shaman, Holy paladin, Discipline priest | 7171.0 | 93.0 |
| Protection warrior, Holy paladin, Restoration shaman | 7165.9 | 98.1 |
| Protection paladin, Holy paladin, Restoration druid | 7163.9 | 100.1 |
| … 50 more | | |

**Each preset's DPS value**: at the winner's combo (its screening score less the best with that slot changed), and the best combo with it less the best without it:

| Preset | At the winner's combo | With minus without |
|---|---|---|
| Protection warrior | n/a | -98.1 |
| Feral tank | n/a | -106.2 |
| Protection paladin | 217.1 | 93.0 |
| Warden shaman | n/a | -93.0 |
| Holy paladin | n/a | -20.0 |
| Restoration shaman | 20.0 | 20.0 |
| Discipline priest | 19.6 | 19.6 |
| Holy priest | n/a | -19.6 |
| Restoration druid | n/a | -52.7 |

**Buffs and debuffs**:

- Curse of the Elements (the boss, cast, 7 members): warlock__demonology 4
- Judgement of Wisdom (the boss, given, 7 members): Protection paladin
- Strength of Earth (raid, given, 7 members): Restoration shaman
- Grace of Air (raid, given, 7 members): Restoration shaman
- Flametongue Totem (raid, given, 7 members): Restoration shaman
- Moonkin Aura (raid, given, 7 members): druid__balance 5
- Arcane Brilliance (raid, given, 7 members): mage__fire 6, mage__fire 7, mage__fire 8, mage__fire 9, mage__fire 10
- Gift of the Wild (raid, given, 7 members): druid__balance 5
- Prayer of Spirit (raid, given, 7 members): Discipline priest
- Fortitude (raid, given, 7 members): Discipline priest
- Mana Spring ×1.25 and Mana Tide (party, given, 4 members): Restoration shaman (party 1)

**What the raid goes without**: nobody gives Battle Shout; nobody gives Furious Howl; nobody gives Sunder upkeep; nobody gives Minor armor slot; nobody gives Judgement of the Crusader; nobody gives Leader of the Pack; nobody gives Trueshot Aura; Druids get 1 of the 3 blessings; Mages get 1 of the 3 blessings; Warlocks get 1 of the 3 blessings.

The model scored 130176 buff sets; 140 were simmed exactly, where it missed by 117.0 at most (20.2 on average).

9 of 43 builds survive pruning: hunter__beast-mastery, hunter__beast-mastery-tallstrider, hunter__lone-wolf, hunter__marksmanship, hunter__marksmanship-tallstrider, mage__fire, mage__frostfire, priest__shadow, warlock__demonology.

60 tank and healer combos in 35 signatures.

Sims and seconds by phase: model 3847 sims 269 s, ranking buff sets 0 sims 1 s, signatures 1888 sims 147 s, exact sets 0 sims 0 s, local moves 419 sims 32 s, combos 2 sims 1 s, contenders 101 sims 46 s, done 0 sims 0 s, confirmation 4 sims 11 s, extras 267 sims 50 s.

## Alliance, camps off

**Best: 7289.6 raid DPS in the raid sims** (±2.1 at 95%; score 7288.7), Protection paladin, Holy paladin, Discipline priest.

A score is the sum of the members' single-player sims in their seats; a raid sim runs the whole raid.

The first 2 comps below are a tie (within 2 standard errors or 0.1%).

1. **7289.6** in the raid sims (1.9 behind the best, ±3.4 at 95%; score 7288.7, 7.0 behind the best score, ±8.4 at 95%): 1× warlock__demonology-duty-curse (CurseOfTheElements), 1× druid__balance, 5× mage__fire; Protection paladin, Holy paladin, Discipline priest
2. **7291.5** in the raid sims (the best; score 7295.8, the best score): 4× mage__frostfire, 1× warlock__demonology-duty-curse (CurseOfTheElements), 1× druid__balance, 1× mage__fire; Protection paladin, Restoration shaman, Discipline priest
3. score **7275.4** (20.4 behind the best score, ±9.5 at 95%): 4× mage__frostfire, 1× warlock__demonology-duty-curse (CurseOfTheElements), 1× druid__balance, 1× mage__fire; Protection paladin, Restoration shaman, Holy priest
4. score **7268.1** (27.6 behind the best score, ±9.1 at 95%): 1× warlock__demonology-duty-curse (CurseOfTheElements), 1× druid__balance, 5× mage__fire; Protection paladin, Holy paladin, Holy priest
5. score **7222.1** (73.7 behind the best score, ±8.0 at 95%): 5× mage__fire, 1× warlock__demonology-duty-curse (CurseOfTheElements), 1× druid__balance; Warden shaman, Holy paladin, Discipline priest

| Party | Member | Race | DPS | Duties and choices | Gets |
|---|---|---|---|---|---|
| 1 | warlock__demonology-duty-curse | Gnome | 1014.9 | CurseOfTheElements (costs 9.7) | Kings, Wisdom |
| 1 | druid__balance | NightElf | 848.1 |  | Kings, Wisdom |
| 1 | mage__fire | Gnome | 1101.6 |  | Kings, Wisdom, Power Infusion |
| 1 | mage__fire | Gnome | 1081.0 |  | Kings, Wisdom |
| 1 | mage__fire | Gnome | 1081.0 |  | Kings, Wisdom |
| 2 | Protection paladin |  |  | judges Wisdom |  |
| 2 | Holy paladin |  |  | judges Wisdom |  |
| 2 | Discipline priest |  |  |  |  |
| 2 | mage__fire | Gnome | 1081.0 |  | Kings, Wisdom |
| 2 | mage__fire | Gnome | 1081.0 |  | Kings, Wisdom |

**Roles** at the winning set (re-filled without each, or with it when it's missing):

| Role | Gain | Provider | Provider cost | Net |
|---|---|---|---|---|
| Battle Shout (missing) | | value of adding | | -532.4 |
| Furious Howl (missing) | | value of adding | | -334.1 |
| Sunder upkeep (missing) | | value of adding | | -415.7 |
| Minor armor slot (missing) | | value of adding | | -28.4 |
| Curse of the Elements | 669.9 | warlock__demonology-duty-curse | 68.8 | 601.2 |
| Judgement of the Crusader (missing) | | value of adding | | 0.0 |
| Judgement of Wisdom | | placeholders (not measured: they give it at no cost) | | not measured |
| Moonkin Aura (with Gift of the Wild) (its giver also gives Gift of the Wild, and its cost counts in each) | 292.3 | druid__balance | 234.6 | 57.7 |
| Leader of the Pack (missing) | | value of adding | | -495.3 |
| Trueshot Aura (missing) | | value of adding | | -300.6 |
| Arcane Brilliance | 359.9 | mage__fire | 0.0 | 359.9 |
| Gift of the Wild (with Moonkin Aura) (its giver also gives Moonkin Aura, and its cost counts in each) | 292.3 | druid__balance | 234.6 | 57.7 |
| Prayer of Spirit | | placeholders (not measured: they give it at no cost) | | not measured |
| Enhancement's totems (Strength of Earth, Windfury) (missing) | | value of adding | | -370.5 |

**Fillers** under the winning set, without a package:

| Build | DPS | ±SE | Assumption range | |
|---|---|---|---|---|
| mage__fire | 1083.4 | 1.9 | 1047 to 1089 | interchangeable with the best |
| mage__frostfire | 1068.1 | 2.6 | 984 to 1074 |  |
| warlock__demonology | 1024.2 | 1.6 | 944 to 1042 |  |
| warlock__affliction | 996.9 | 1.7 | 922 to 1017 |  |
| warlock__destruction | 978.0 | 1.6 | 912 to 997 |  |
| warlock__destruction_fire | 975.8 | 1.1 | 906 to 993 |  |
| priest__shadow | 948.4 | 0.8 | 935 to 952 |  |
| mage__frost | 910.4 | 1.7 | 847 to 927 |  |
| mage__arcane | 897.8 | 1.3 | 846 to 910 |  |
| priest__shadow-holy | 871.6 | 1.1 | 859 to 875 |  |
| druid__balance | 848.8 | 1.0 | 832 to 856 |  |
| druid__balance-wrath | 810.6 | 0.9 | 796 to 822 |  |
| hunter__beast-mastery | 786.4 | 0.7 | 745 to 787 |  |
| shaman__elemental-level60 | 764.3 | 0.9 | 755 to 768 |  |
| shaman__elemental-level60-mana-tide | 751.7 | 0.9 | 742 to 755 |  |
| priest__smite | 643.6 | 0.9 | 628 to 647 |  |
| hunter__survival | 610.8 | 0.6 | 558 to 611 |  |
| warrior__fury | 534.5 | 1.1 | 454 to 587 |  |
| rogue__combat-daggers | 481.6 | 0.6 | 445 to 482 |  |
| warrior__arms | 479.3 | 0.9 | 440 to 505 |  |
| warrior__fury_2h | 468.3 | 1.0 | 384 to 524 |  |
| rogue__subtlety | 467.5 | 0.5 | 435 to 467 |  |
| rogue__seal-fate-daggers | 452.2 | 0.6 | 425 to 452 |  |
| rogue__combat-swords | 451.2 | 0.6 | 424 to 451 |  |
| rogue__mutilate | 443.4 | 0.5 | 415 to 443 |  |
| rogue__seal-fate-swords | 429.6 | 0.6 | 399 to 430 |  |

**Member values** (one row per distinct member):

| Member | Own DPS | Contribution | Replacement margin | Best replacement | vs the report |
|---|---|---|---|---|---|
| warlock__demonology-duty-curse | 1014.6 | 1586.0 | 26.1 | warlock__affliction-duty-curse | -37.2 (the report's request in this sim: +0.8; race (Gnome): -3.9; this fight (boss, length, flask zone): +0.0; the raid's buff set: -24.5; its duty (Curse of the Elements): -9.7) |
| druid__balance | 848.8 | 1098.3 | 38.2 | druid__balance-wrath | -2.4 (the report's request in this sim: -0.1; race (NightElf): +0.0; this fight (boss, length, flask zone): +0.0; the raid's buff set: -2.4) |
| mage__fire | 1103.7 | 1083.4 | 15.2 | mage__frostfire | +14.7 (the report's request in this sim: +1.0; race (Gnome): +2.5; this fight (boss, length, flask zone): +0.0; the raid's buff set: +11.3) |
| mage__fire (×4) | 1083.4 | 1083.4 | 15.2 | mage__frostfire | -5.6 (the report's request in this sim: +1.0; race (Gnome): +2.5; this fight (boss, length, flask zone): +0.0; the raid's buff set: -9.1) |

**Value of adding** a build that isn't in the comp (swapped for the member it beats most):

- mage__frostfire: -15.2
- warlock__affliction: -26.1
- druid__balance-wrath: -38.2
- warlock__destruction: -45.2
- warlock__destruction_fire: -47.1
- priest__shadow: -135.0
- mage__frost: -172.9
- mage__arcane: -179.7

**Tank and healer combos** (best first, by their best comp's screening score):

| Combo | Best comp's score | Behind |
|---|---|---|
| Protection paladin, Restoration shaman, Discipline priest | 7305.5 | 0.0 |
| Protection paladin, Holy paladin, Discipline priest | 7300.5 | 5.0 |
| Protection paladin, Restoration shaman, Holy priest | 7285.1 | 20.4 |
| Protection paladin, Holy paladin, Holy priest | 7280.1 | 25.4 |
| Warden shaman, Holy paladin, Discipline priest | 7230.9 | 74.6 |
| Protection paladin, Restoration shaman, Restoration druid | 7220.6 | 84.9 |
| Warden shaman, Holy paladin, Holy priest | 7210.6 | 94.9 |
| Protection paladin, Discipline priest, Restoration druid | 7205.7 | 99.8 |
| Protection paladin, Holy paladin, Restoration druid | 7203.5 | 102.0 |
| Protection paladin, Holy paladin, Restoration shaman | 7203.4 | 102.2 |
| … 50 more | | |

**Each preset's DPS value**: at the winner's combo (its screening score less the best with that slot changed), and the best combo with it less the best without it:

| Preset | At the winner's combo | With minus without |
|---|---|---|
| Protection warrior | n/a | -116.4 |
| Feral tank | n/a | -125.0 |
| Protection paladin | 69.6 | 74.6 |
| Warden shaman | n/a | -74.6 |
| Holy paladin | -5.0 | -5.0 |
| Restoration shaman | n/a | 5.0 |
| Discipline priest | 20.3 | 20.4 |
| Holy priest | n/a | -20.4 |
| Restoration druid | n/a | -84.9 |

**Buffs and debuffs**:

- Curse of the Elements (the boss, cast, 7 members): warlock__demonology 4
- Judgement of Wisdom (the boss, given, 7 members): Protection paladin, Holy paladin
- Moonkin Aura (raid, given, 7 members): druid__balance 5
- Arcane Brilliance (raid, given, 7 members): mage__fire 6, mage__fire 7, mage__fire 8, mage__fire 9, mage__fire 10
- Gift of the Wild (raid, given, 7 members): druid__balance 5
- Prayer of Spirit (raid, given, 7 members): Discipline priest
- Fortitude (raid, given, 7 members): Discipline priest

**What the raid goes without**: nobody gives Battle Shout; nobody gives Furious Howl; nobody gives Sunder upkeep; nobody gives Minor armor slot; nobody gives Judgement of the Crusader; nobody gives Strength of Earth; nobody gives Grace of Air; nobody gives Leader of the Pack; nobody gives Trueshot Aura; Druids get 2 of the 3 blessings; Mages get 2 of the 3 blessings; Warlocks get 2 of the 3 blessings.

The model scored 130176 buff sets; 140 were simmed exactly, where it missed by 92.5 at most (23.1 on average).

10 of 43 builds survive pruning: hunter__beast-mastery, hunter__beast-mastery-tallstrider, hunter__lone-wolf, hunter__marksmanship, hunter__marksmanship-tallstrider, mage__fire, mage__frostfire, priest__shadow, warlock__demonology, warlock__destruction_fire.

60 tank and healer combos in 35 signatures.

Sims and seconds by phase: model 3847 sims 187 s, ranking buff sets 0 sims 1 s, signatures 1694 sims 120 s, exact sets 0 sims 0 s, local moves 386 sims 35 s, combos 2 sims 1 s, contenders 81 sims 36 s, done 0 sims 0 s, confirmation 4 sims 10 s, extras 202 sims 36 s.

## Assumptions

- Each build at its best race in the faction, from the report's race tables.
- Mana Spring from two shamans doesn't stack, and a party has one package (shaman_audit 8.45, unverified).
- Furious Howl and the minor armor slot: the strongest counts, nothing stacks (until the launch checks).
- Hunter's Mark: each hunter marks for itself; the shared mark is left to the raid sims.
- Power Infusion fires as soon as it's ready; Innervate below a mana threshold.
- Everyone is in range of every raid-wide buff.
- Moonkin Aura takes the whole Camp Chair (unverified).

**Left out** (no build can do it yet): a DPS shaman's totem choice: Enhancement and Elemental drop their builds' fixed totems (the engine's assigned totems aren't in their rotations yet); Dark Inspiration as a party package (a Shadow priest's own Spirit is in its sims); the hunter pet variants' cat swing speed and Cobra Reflexes on the Tallstrider and wolf (an assumption).

**Not modelled**: healing, tanking, threat, pet deaths, movement, boss mechanics, the tanks' and healers' own damage.
