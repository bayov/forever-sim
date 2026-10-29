# Raid sim M0: what each role is worth (rough)

Written by `tools/compsearch/m0.py` from the DPS Report's one-buff-at-a-time runs (300 iterations, 180 s,
the report's full set, no camps). Additive, no sims: M1's own role sims replace it. A role a build's
request never had shows 0 for it (marked *).

## The fillers, by published DPS

| Build | DPS | Dominated by |
|---|---|---|
| hunter__beast-mastery | 1131.8 | **survives** |
| hunter__marksmanship | 1115.7 | **survives** |
| hunter__lone-wolf | 1094.3 | **survives** |
| mage__fire | 1089.0 | **survives** |
| mage__frostfire | 1087.5 | **survives** |
| warlock__demonology | 1051.8 | **survives** |
| hunter__survival | 1028.9 | hunter__beast-mastery |
| warlock__affliction | 1018.1 | warlock__demonology |
| warlock__destruction | 1001.6 | warlock__affliction |
| warlock__destruction_fire | 986.4 | warlock__affliction |
| warrior__fury | 982.2 | hunter__beast-mastery |
| warrior__fury_2h | 964.0 | hunter__beast-mastery |
| priest__shadow | 952.8 | warlock__affliction |
| paladin__level_60_sub_holy | 948.8 | warlock__affliction |
| paladin__level_60 | 944.0 | warlock__affliction |
| paladin__level_60_shock_ret | 931.6 | warlock__affliction |
| mage__frost | 911.1 | mage__fire |
| priest__shadow-holy | 910.4 | priest__shadow |
| warrior__arms | 905.6 | hunter__beast-mastery |
| mage__arcane | 901.6 | mage__fire |
| paladin__level_60_shock | 887.8 | warlock__affliction |
| shaman__enhancement-level60 | 884.2 | hunter__beast-mastery |
| shaman__enhancement-imp-gw | 880.1 | hunter__beast-mastery |
| druid__feral-cat | 866.8 | hunter__beast-mastery |
| druid__balance | 851.2 | warlock__affliction |
| shaman__enhancement-water-shield | 836.3 | hunter__beast-mastery |
| druid__balance-wrath | 811.1 | druid__balance |
| shaman__enhancement-mana-tide | 796.6 | hunter__beast-mastery |
| rogue__combat-daggers | 773.3 | hunter__beast-mastery |
| shaman__elemental-level60 | 772.2 | warlock__affliction |
| priest__smite | 766.8 | priest__shadow-holy |
| rogue__combat-swords | 765.4 | hunter__beast-mastery |
| shaman__elemental-level60-mana-tide | 757.0 | warlock__affliction |
| rogue__seal-fate-daggers | 734.9 | hunter__beast-mastery |
| rogue__seal-fate-swords | 731.8 | hunter__beast-mastery |
| rogue__mutilate | 723.3 | hunter__beast-mastery |
| rogue__subtlety | 717.2 | hunter__beast-mastery |

**6 of 37 survive** the dominance rule at a 1.5 DPS margin: hunter__beast-mastery, hunter__lone-wolf, hunter__marksmanship, mage__fire, mage__frostfire, warlock__demonology.

## 10 players (7 DPS slots)

The fill under the full set is 7 × hunter__beast-mastery. Gain is what the role is worth to the other 6; cost is its cheapest provider's DPS below the filler (duty included).

| Role | Per member | Gain | Provider | Cost | Net |
|---|---|---|---|---|---|
| Sunder Armor | 229.4 | 1376 | placeholders only | - | - |
| Minor armor slot | 64.4 | 387 | druid__feral-cat-duty-ff | 269.6 | 117 |
| Battle Shout | 1.8 | 11 | warrior__fury-duty-battle-shout | 151.4 | -141 |
| Curse of the Elements | 21.7 | 130 | warlock__demonology-duty-curse | 89.0 | 41 |
| Judgement of Wisdom | 56.1 | 336 | placeholders only | - | - |
| Judgement of the Crusader | -0.0 | -0 | paladin__level_60_sub_holy-duty-crusader | 190.6 | -191 |
| Kings | 48.1 | 288 | paladin__level_60_sub_holy | 183.0 | 105 |
| Might | 2.1 | 13 | paladin__level_60_sub_holy | 183.0 | -170 |
| Wisdom | 3.2 | 19 | paladin__level_60_sub_holy | 183.0 | -164 |
| Gift of the Wild | 15.2 | 91 | druid__feral-cat | 265.0 | -174 |
| Arcane Brilliance | 8.5 | 51 | mage__fire | 42.8 | 8 |
| Prayer of Spirit | 0.8 | 5 | priest__shadow | 179.0 | -174 |
| Strength of Earth | 0.0* | 0 | shaman__enhancement-level60 | 247.6 | -248 |
| Grace of Air | 52.9 | 317 | placeholders only | - | - |
| Weapon totem | 0.0* | 0 | shaman__enhancement-level60 | 247.6 | -248 |
| Trueshot Aura | 10.8 | 65 | hunter__marksmanship | 16.1 | 49 |
| Leader of the Pack | 28.2 | 169 | druid__feral-cat | 265.0 | -96 |
| Moonkin Aura | 0.2 | 1 | druid__balance | 280.6 | -280 |
| Mana Spring (party) | 1.4 | 9 | - | - | - |

## 20 players (14 DPS slots)

The fill under the full set is 14 × hunter__beast-mastery. Gain is what the role is worth to the other 13; cost is its cheapest provider's DPS below the filler (duty included).

| Role | Per member | Gain | Provider | Cost | Net |
|---|---|---|---|---|---|
| Sunder Armor | 229.4 | 2982 | placeholders only | - | - |
| Minor armor slot | 64.4 | 838 | druid__feral-cat-duty-ff | 269.6 | 568 |
| Battle Shout | 1.8 | 24 | warrior__fury-duty-battle-shout | 151.4 | -128 |
| Curse of the Elements | 21.7 | 282 | warlock__demonology-duty-curse | 89.0 | 193 |
| Judgement of Wisdom | 56.1 | 729 | placeholders only | - | - |
| Judgement of the Crusader | -0.0 | -0 | paladin__level_60_sub_holy-duty-crusader | 190.6 | -191 |
| Kings | 48.1 | 625 | paladin__level_60_sub_holy | 183.0 | 442 |
| Might | 2.1 | 27 | paladin__level_60_sub_holy | 183.0 | -156 |
| Wisdom | 3.2 | 42 | paladin__level_60_sub_holy | 183.0 | -141 |
| Gift of the Wild | 15.2 | 198 | druid__feral-cat | 265.0 | -67 |
| Arcane Brilliance | 8.5 | 111 | mage__fire | 42.8 | 68 |
| Prayer of Spirit | 0.8 | 10 | priest__shadow | 179.0 | -169 |
| Strength of Earth | 0.0* | 0 | shaman__enhancement-level60 | 247.6 | -248 |
| Grace of Air | 52.9 | 687 | placeholders only | - | - |
| Weapon totem | 0.0* | 0 | shaman__enhancement-level60 | 247.6 | -248 |
| Trueshot Aura | 10.8 | 141 | hunter__marksmanship | 16.1 | 125 |
| Leader of the Pack | 28.2 | 366 | druid__feral-cat | 265.0 | 101 |
| Moonkin Aura | 0.2 | 2 | druid__balance | 280.6 | -278 |
| Mana Spring (party) | 1.4 | 19 | - | - | - |

## 40 players (29 DPS slots)

The fill under the full set is 29 × hunter__beast-mastery. Gain is what the role is worth to the other 28; cost is its cheapest provider's DPS below the filler (duty included).

| Role | Per member | Gain | Provider | Cost | Net |
|---|---|---|---|---|---|
| Sunder Armor | 229.4 | 6422 | placeholders only | - | - |
| Minor armor slot | 64.4 | 1804 | druid__feral-cat-duty-ff | 269.6 | 1535 |
| Battle Shout | 1.8 | 51 | warrior__fury-duty-battle-shout | 151.4 | -101 |
| Curse of the Elements | 21.7 | 608 | warlock__demonology-duty-curse | 89.0 | 519 |
| Judgement of Wisdom | 56.1 | 1570 | placeholders only | - | - |
| Judgement of the Crusader | -0.0 | -0 | paladin__level_60_sub_holy-duty-crusader | 190.6 | -191 |
| Kings | 48.1 | 1345 | paladin__level_60_sub_holy | 183.0 | 1162 |
| Might | 2.1 | 59 | paladin__level_60_sub_holy | 183.0 | -124 |
| Wisdom | 3.2 | 91 | paladin__level_60_sub_holy | 183.0 | -92 |
| Gift of the Wild | 15.2 | 426 | druid__feral-cat | 265.0 | 161 |
| Arcane Brilliance | 8.5 | 239 | mage__fire | 42.8 | 196 |
| Prayer of Spirit | 0.8 | 22 | priest__shadow | 179.0 | -157 |
| Strength of Earth | 0.0* | 0 | shaman__enhancement-level60 | 247.6 | -248 |
| Grace of Air | 52.9 | 1480 | placeholders only | - | - |
| Weapon totem | 0.0* | 0 | shaman__enhancement-level60 | 247.6 | -248 |
| Trueshot Aura | 10.8 | 304 | hunter__marksmanship | 16.1 | 288 |
| Leader of the Pack | 28.2 | 789 | druid__feral-cat | 265.0 | 524 |
| Moonkin Aura | 0.2 | 5 | druid__balance | 280.6 | -276 |
| Mana Spring (party) | 1.4 | 40 | - | - | - |

## Per build, per role (DPS lost when the role is missing)

| Build | Sunder Armor | Minor armor slot | Battle Shout | Curse of the Elements | Judgement of Wisdom | Judgement of the Crusader | Kings | Might | Wisdom | Gift of the Wild | Arcane Brilliance | Prayer of Spirit | Strength of Earth | Grace of Air | Weapon totem | Trueshot Aura | Leader of the Pack | Moonkin Aura | Mana Spring (party) |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| druid__balance | -0.0 | -0.0 | -0.0 | 77.1 | 3.6 | -0.0 | 5.2 | 0.0* | 0.5 | 2.4 | 5.4 | -0.3 | 0.0* | 0.0* | 0.0* | -0.0 | 0.0* | 0.0* | 0.5 |
| druid__balance-wrath | -0.0 | -0.0 | -0.0 | 73.7 | 3.2 | -0.0 | 4.3 | 0.0* | 0.2 | 2.4 | 4.2 | 0.1 | 0.0* | 0.0* | 0.0* | -0.0 | 0.0* | 0.0* | 1.5 |
| druid__feral-cat | 143.0 | 40.2 | 25.5 | 1.3 | 56.9 | -0.0 | 46.5 | 29.5 | 2.6 | 17.3 | 0.5 | 0.0 | 22.3 | 42.5 | 56.2 | -0.0 | 0.0* | 0.0* | 0.0* |
| hunter__beast-mastery | 229.4 | 64.4 | 1.8 | 21.7 | 56.1 | -0.0 | 48.1 | 2.1 | 3.2 | 15.2 | 8.5 | 0.8 | 0.0* | 52.9 | 0.0* | 10.8 | 28.2 | 0.2 | 1.4 |
| hunter__lone-wolf | 202.7 | 57.0 | 0.9 | 28.0 | 111.3 | -0.0 | 60.5 | 1.1 | 30.8 | 21.6 | 16.4 | 1.1 | 0.0* | 63.5 | 0.0* | -0.0 | 28.1 | 0.2 | 13.3 |
| hunter__marksmanship | 219.7 | 61.7 | 0.8 | 23.6 | 94.5 | -0.0 | 54.1 | 1.0 | 22.8 | 20.1 | 12.8 | -0.3 | 0.0* | 59.0 | 0.0* | -0.0 | 29.4 | 0.2 | 11.9 |
| hunter__survival | 234.2 | 66.1 | 27.0 | 7.8 | 99.3 | -0.0 | 39.7 | 31.3 | -0.0 | 18.5 | 9.1 | -0.0 | 10.8 | 38.4 | 81.0 | 1.0 | 27.0 | 1.2 | -0.0 |
| mage__arcane | -0.0 | -0.0 | -0.0 | 82.4 | 25.4 | -0.0 | 19.3 | 0.0* | 8.5 | 9.1 | -0.0 | 7.0 | 0.0* | 0.0* | 0.0* | -0.0 | -0.0 | 21.1 | 5.8 |
| mage__fire | -0.0 | -0.0 | -0.0 | 99.5 | 58.8 | -0.0 | 22.9 | 0.0* | 23.0 | 10.0 | -0.0 | 8.0 | 0.0* | 0.0* | 0.0* | -0.0 | -0.0 | 32.5 | 11.4 |
| mage__frost | -0.0 | -0.0 | -0.0 | 83.0 | 6.4 | -0.0 | 10.2 | 0.0* | 0.9 | 4.5 | -0.0 | -0.4 | 0.0* | 0.0* | 0.0* | -0.0 | -0.0 | 21.0 | 1.1 |
| mage__frostfire | -0.0 | -0.0 | -0.0 | 98.9 | 50.3 | -0.0 | 25.5 | 0.0* | 32.6 | 8.2 | -0.0 | 5.2 | 0.0* | 0.0* | 0.0* | -0.0 | -0.0 | 35.9 | 19.6 |
| paladin__level_60 | 60.1 | 17.0 | 18.0 | 63.1 | 50.7 | 41.4 | 27.4 | 20.8 | -0.8 | 17.9 | 8.6 | 0.2 | 15.9 | 26.8 | 101.3 | -0.0 | 18.9 | 2.5 | -1.2 |
| paladin__level_60_shock | 39.2 | 11.0 | 10.1 | 65.7 | 63.6 | 59.7 | 15.9 | 11.7 | 10.5 | 12.2 | 5.8 | 1.4 | 9.0 | 17.7 | 67.1 | -0.0 | 12.4 | 4.7 | 5.0 |
| paladin__level_60_shock_ret | 52.7 | 15.0 | 17.0 | 64.5 | 38.6 | 44.6 | 22.9 | 19.6 | 0.1 | 14.7 | 9.9 | -3.6 | 15.1 | 23.9 | 77.9 | -0.0 | 16.5 | 3.6 | -1.6 |
| paladin__level_60_sub_holy | 60.2 | 17.1 | 18.1 | 63.8 | 25.0 | 41.5 | 29.6 | 21.0 | 2.6 | 19.2 | 9.4 | 3.0 | 16.1 | 28.3 | 104.2 | -0.0 | 19.4 | 2.4 | 1.1 |
| priest__shadow | -0.0 | -0.0 | -0.0 | 85.4 | 0.4 | -0.0 | 6.4 | 0.0* | -0.2 | 3.0 | 4.7 | -0.0 | 0.0* | 0.0* | 0.0* | -0.0 | -0.0 | 23.4 | -0.4 |
| priest__shadow-holy | -0.0 | -0.0 | -0.0 | 81.6 | 25.5 | -0.0 | 6.3 | 0.0* | 5.5 | 3.0 | 5.1 | -0.0 | 0.0* | 0.0* | 0.0* | -0.0 | -0.0 | 22.6 | 2.8 |
| priest__smite | -0.0 | -0.0 | -0.0 | 68.5 | 39.0 | 61.8 | 13.1 | 0.0* | 18.3 | 6.6 | 6.6 | -0.0 | 0.0* | 0.0* | 0.0* | -0.0 | -0.0 | 9.9 | 10.0 |
| rogue__combat-daggers | 161.6 | 44.6 | 26.6 | 4.7 | -0.0 | -0.0 | 36.1 | 30.7 | -0.0 | 11.7 | -0.0 | -0.0 | 10.6 | 36.2 | 47.0 | -0.0 | 17.9 | 0.8 | 0.0* |
| rogue__combat-swords | 161.1 | 44.4 | 28.5 | 4.5 | -0.0 | -0.0 | 39.3 | 32.9 | -0.0 | 12.8 | -0.0 | -0.0 | 11.4 | 38.1 | 77.1 | -0.0 | 18.1 | 0.7 | 0.0* |
| rogue__mutilate | 160.8 | 45.3 | 24.0 | 7.1 | -0.0 | -0.0 | 33.5 | 27.7 | -0.0 | 11.3 | -0.0 | -0.0 | 9.6 | 33.7 | 38.8 | -0.0 | 17.3 | 1.1 | 0.0* |
| rogue__seal-fate-daggers | 163.3 | 46.2 | 23.8 | 7.0 | -0.0 | -0.0 | 32.9 | 27.5 | -0.0 | 10.9 | -0.0 | -0.0 | 9.5 | 32.4 | 41.9 | -0.0 | 16.1 | 1.2 | 0.0* |
| rogue__seal-fate-swords | 163.2 | 46.1 | 25.4 | 6.8 | -0.0 | -0.0 | 37.5 | 29.4 | -0.0 | 13.0 | -0.0 | -0.0 | 10.2 | 37.0 | 60.2 | -0.0 | 19.5 | 1.0 | 0.0* |
| rogue__subtlety | 130.2 | 36.1 | 26.2 | 7.4 | -0.0 | -0.0 | 33.6 | 30.4 | -0.0 | 10.5 | -0.0 | -0.0 | 10.5 | 35.0 | 42.2 | -0.0 | 16.7 | 1.1 | 0.0* |
| shaman__elemental-level60 | -0.0 | -0.0 | -0.0 | 70.4 | 26.0 | -0.0 | 6.1 | 0.0* | 7.4 | 3.0 | 4.9 | 1.0 | 0.0* | 0.0* | 0.0* | -0.0 | -0.0 | 18.0 | 0.0* |
| shaman__elemental-level60-mana-tide | -0.0 | -0.0 | -0.0 | 68.8 | 6.4 | -0.0 | 4.7 | 0.0* | 1.8 | 2.4 | 4.0 | 0.7 | 0.0* | 0.0* | 0.0* | -0.0 | -0.0 | 17.1 | 0.0* |
| shaman__enhancement-imp-gw | 117.1 | 33.3 | 18.7 | 38.2 | 36.3 | -0.0 | 25.9 | 21.6 | 6.6 | 18.4 | 13.6 | 1.9 | 0.0* | 22.4 | 17.1 | -0.0 | 14.3 | 8.1 | 0.0* |
| shaman__enhancement-level60 | 117.7 | 33.5 | 18.7 | 38.3 | 34.2 | -0.0 | 31.3 | 21.6 | 9.4 | 19.1 | 14.1 | 2.6 | 0.0* | 23.9 | 16.9 | -0.0 | 14.4 | 8.3 | 0.0* |
| shaman__enhancement-mana-tide | 113.6 | 32.4 | 17.9 | 31.6 | 4.9 | -0.0 | 26.6 | 20.7 | -0.6 | 17.2 | 10.1 | 1.1 | 0.0* | 24.7 | 19.5 | -0.0 | 16.8 | 4.7 | 0.0* |
| shaman__enhancement-water-shield | 116.1 | 32.8 | 18.2 | 35.1 | 9.0 | -0.0 | 26.3 | 21.0 | -0.1 | 13.8 | 9.1 | 0.6 | 0.0* | 29.0 | 22.5 | -0.0 | 20.0 | 5.3 | 0.0* |
| warlock__affliction | 13.9 | 3.9 | 0.0* | 87.5 | 10.5 | -0.0 | 9.7 | 0.0* | 6.8 | 4.3 | 6.6 | 3.3 | 0.0* | 0.0* | 0.0* | 0.0* | 0.0* | 31.1 | 2.5 |
| warlock__demonology | 16.6 | 4.7 | 0.0* | 89.9 | 18.8 | -0.0 | 12.3 | 0.0* | 10.4 | 6.8 | 10.5 | 4.0 | 0.0* | 0.0* | 0.0* | 0.0* | 0.0* | 29.8 | 6.4 |
| warlock__destruction | 14.5 | 4.1 | 0.0* | 86.2 | 17.7 | -0.0 | 13.2 | 0.0* | 7.1 | 4.3 | 9.7 | 7.9 | 0.0* | 0.0* | 0.0* | 0.0* | 0.0* | 32.3 | 7.2 |
| warlock__destruction_fire | -0.0 | -0.0 | 0.0* | 88.6 | 21.1 | -0.0 | 10.8 | 0.0* | 11.3 | 6.3 | 8.5 | 3.4 | 0.0* | 0.0* | 0.0* | 0.0* | 0.0* | 21.2 | 6.5 |
| warrior__arms | 194.0 | 56.1 | 24.1 | 0.6 | -0.0 | -0.0 | 38.5 | 27.9 | 0.0* | 15.9 | -0.0 | -0.0 | 19.3 | 28.3 | 173.2 | -0.0 | 21.5 | 0.6 | 0.0* |
| warrior__fury | 227.8 | 65.7 | 32.5 | 0.6 | -0.0 | -0.0 | 55.6 | 37.5 | 0.0* | 23.2 | -0.0 | -0.0 | 26.0 | 46.0 | 122.7 | -0.0 | 31.6 | -0.1 | 0.0* |
| warrior__fury_2h | 221.9 | 64.2 | 30.2 | 0.6 | -0.0 | -0.0 | 49.3 | 34.9 | 0.0* | 17.5 | -0.0 | -0.0 | 24.2 | 47.8 | 226.8 | -0.0 | 29.2 | 0.8 | 0.0* |
