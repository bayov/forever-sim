# Hit sweep, DPS report builds (round 70753)

Method. For each of the 68 builds in ui/dps_report/data/builds.json, the sim request is tools/dps_report/req/<id>.json
(gear, talents string, race, buffs, consumes; <id>.meta.json has presentation data only). I fed each request to
core.ComputeStats through tools/dps_report/dashstats (the helper the report's stats.json comes from), which prints the
player's stat layers base, gear, talents, buffs, consumes and final (cumulative). Total hit = final MeleeHit and final
SpellHit (percent; 1 unit = 1%; ranged attacks use MeleeHit). Gear-only hit = items (HitRating/10 plus stat hit) plus
enchants, summed straight from the item database (core.ItemsByID and EnchantsByEffectID) for each request; the gear
layer minus that sum is the set bonus (+1 for the mage, elemental shaman and warlock synthetic 2pc sets). Forever pays gear
hit into both melee and spell, so gear melee = gear spell. "Other" = total minus items+enchants, so it includes the set
bonus. Talent names and ranks come from decoding each build's talent string against ui/dps_report/data/talenttrees.json;
school-only talents are not in the stat sheet and are listed separately and added to the cap check. No build has a racial
hit bonus (the Dwarf's Artisan's Precision is not in any build) and no buff or party aura adds hit, except one consume.

Caps vs a level 63 boss in this sim: 8% yellow melee and ranged, 27% dual-wield white, 16% spell. Which caps apply:
single-weapon melee (2H Arms/Fury, Ret, Enhancement as built (no off-hand in its request), Feral): 8%. Dual wielders
(Fury 1H, all rogues, Survival hunter with an off-hand): 8% yellow and 27% white. Marksmanship/Beast Mastery: ranged 8%
(Survival also melee). Casters (Elemental, Balance, mage, warlock, priest): 16%; warlock pets inherit the owner's spell
hit. Hybrids (Ret, Enhancement): melee 8% and spell 16%. Rogue poisons roll on spell hit (Precision gives spell hit too),
so the poison spell line is shown as information. Feral with Faerie Fire duty rolls FF on spell hit (informational).
Column format: "total / cap (over+ or under-)". Total and other columns are melee / spell.

| Build | Spec | Race | Gear hit (items+ench, same for melee and spell) | Other melee / spell: sources | Total melee / spell | Caps: total/cap (diff) |
|---|---|---|---|---|---|---|
| warrior__fury_2h | Fury | Orc | 8 | 0 / 0: none | 8 / 8 | yellow/white 8/8 (+0) |
| warrior__fury_2h-duty-battle-shout | Fury | Orc | 8 | 0 / 0: none | 8 / 8 | yellow/white 8/8 (+0) |
| warrior__fury | Fury | Orc | 10 | 0 / 0: Furious Precision 3/3: +10% off-hand only (melee) | 10 / 10 | yellow MH 10/8 (+2); white MH (DW) 10/27 (-17); white OH (DW) 20/27 (-7) |
| warrior__fury-duty-battle-shout | Fury | Orc | 10 | 0 / 0: Furious Precision 3/3: +10% off-hand only (melee) | 10 / 10 | yellow MH 10/8 (+2); white MH (DW) 10/27 (-17); white OH (DW) 20/27 (-7) |
| warrior__arms_dragonkin | Arms | Orc | 8 | 0 / 0: none | 8 / 8 | yellow/white 8/8 (+0) |
| warrior__fury_2h-duty-sunder | Fury | Orc | 8 | 0 / 0: none | 8 / 8 | yellow/white 8/8 (+0) |
| warrior__fury-duty-sunder | Fury | Orc | 10 | 0 / 0: Furious Precision 3/3: +10% off-hand only (melee) | 10 / 10 | yellow MH 10/8 (+2); white MH (DW) 10/27 (-17); white OH (DW) 20/27 (-7) |
| warrior__arms | Arms | Orc | 8 | 0 / 0: none | 8 / 8 | yellow/white 8/8 (+0) |
| warrior__arms-duty-battle-shout | Arms | Orc | 8 | 0 / 0: none | 8 / 8 | yellow/white 8/8 (+0) |
| warrior__arms-duty-sunder | Arms | Orc | 8 | 0 / 0: none | 8 / 8 | yellow/white 8/8 (+0) |
| rogue__combat-swords | Combat | Undead | 6 (+set 1) | 3 / 3: Precision 3/3 (3% all); set 2pc +1 | 10 / 10 | yellow 10/8 (+2); white (DW) 10/27 (-17); poison spell 10/16 (-6) |
| rogue__combat-daggers | Combat | Undead | 5 (+set 1) | 8 / 8: Precision 3/3 (3% all); set 2pc +1; Flask of Natural Accuracy +5 | 14 / 14 | yellow 14/8 (+6); white (DW) 14/27 (-13); poison spell 14/16 (-2) |
| rogue__seal-fate-swords | Seal Fate | Undead | 6 (+set 1) | 3 / 3: Precision 3/3 (3% all); set 2pc +1 | 10 / 10 | yellow 10/8 (+2); white (DW) 10/27 (-17); poison spell 10/16 (-6) |
| rogue__seal-fate-daggers | Seal Fate | Undead | 6 (+set 1) | 3 / 3: Precision 3/3 (3% all); set 2pc +1 | 10 / 10 | yellow 10/8 (+2); white (DW) 10/27 (-17); poison spell 10/16 (-6) |
| rogue__subtlety | Subtlety | Night Elf | 7 (+set 1) | 3 / 3: Precision 3/3 (3% all); set 2pc +1 | 11 / 11 | yellow 11/8 (+3); white (DW) 11/27 (-16); poison spell 11/16 (-5) |
| rogue__mutilate | Assassination | Undead | 6 (+set 1) | 3 / 3: Precision 3/3 (3% all); set 2pc +1 | 10 / 10 | yellow 10/8 (+2); white (DW) 10/27 (-17); poison spell 10/16 (-6) |
| hunter__lone-wolf | Marksmanship | Night Elf | 8 | 0 / 0: none | 8 / 8 | ranged 8/8 (+0) |
| hunter__marksmanship | Marksmanship | Orc | 8 | 0 / 0: none | 8 / 8 | ranged 8/8 (+0) |
| hunter__marksmanship-wolf | Marksmanship | Orc | 8 | 0 / 0: none | 8 / 8 | ranged 8/8 (+0) |
| hunter__marksmanship-tallstrider | Marksmanship | Orc | 8 | 0 / 0: none | 8 / 8 | ranged 8/8 (+0) |
| hunter__beast-mastery | Beast Mastery | Orc | 8 | 0 / 0: none | 8 / 8 | ranged 8/8 (+0) |
| hunter__beast-mastery-wolf | Beast Mastery | Orc | 8 | 0 / 0: none | 8 / 8 | ranged 8/8 (+0) |
| hunter__beast-mastery-tallstrider | Beast Mastery | Orc | 8 | 0 / 0: none | 8 / 8 | ranged 8/8 (+0) |
| hunter__survival | Survival | Orc | 5 | 3 / 3: Surefooted 3/3 (3% all); Survival Tactics 2/2: +10% Trap/Feign hit only; traps 18 | 8 / 8 | ranged 8/8 (+0); melee yellow 8/8 (+0); melee white (DW) 8/27 (-19) |
| hunter__survival-wolf | Survival | Orc | 5 | 3 / 3: Surefooted 3/3 (3% all); Survival Tactics 2/2: +10% Trap/Feign hit only; traps 18 | 8 / 8 | ranged 8/8 (+0); melee yellow 8/8 (+0); melee white (DW) 8/27 (-19) |
| hunter__survival-tallstrider | Survival | Orc | 5 | 3 / 3: Surefooted 3/3 (3% all); Survival Tactics 2/2: +10% Trap/Feign hit only; traps 18 | 8 / 8 | ranged 8/8 (+0); melee yellow 8/8 (+0); melee white (DW) 8/27 (-19) |
| paladin__level_60_sub_holy_undead | Retribution | Undead | 8 | 0 / 0: Divine Precision 2/2: +8% on Consecration, Exorcism, Holy Shock, Holy Wrath (spell) and Holy Strike (melee) only, not seals/Judgement; those total 16 spell, 16 Holy Strike | 8 / 8 | melee 8/8 (+0); spell 8/16 (-8) |
| paladin__level_60 | Retribution | Undead | 8 | 3 / 3: Precision 3/3 (3% all) | 11 / 11 | melee 11/8 (+3); spell 11/16 (-5) |
| paladin__level_60_sub_holy | Retribution | Undead | 8 | 0 / 0: Divine Precision 2/2: +8% on Consecration, Exorcism, Holy Shock, Holy Wrath (spell) and Holy Strike (melee) only, not seals/Judgement; those total 16 spell, 16 Holy Strike | 8 / 8 | melee 8/8 (+0); spell 8/16 (-8) |
| paladin__level_60-duty-crusader | Retribution | Undead | 8 | 3 / 3: Precision 3/3 (3% all) | 11 / 11 | melee 11/8 (+3); spell 11/16 (-5) |
| paladin__level_60_sub_holy-duty-crusader | Retribution | Undead | 8 | 0 / 0: Divine Precision 2/2: +8% on Consecration, Exorcism, Holy Shock, Holy Wrath (spell) and Holy Strike (melee) only, not seals/Judgement; those total 16 spell, 16 Holy Strike | 8 / 8 | melee 8/8 (+0); spell 8/16 (-8) |
| paladin__level_60-duty-wisdom | Retribution | Undead | 8 | 3 / 3: Precision 3/3 (3% all) | 11 / 11 | melee 11/8 (+3); spell 11/16 (-5) |
| paladin__level_60_sub_holy-duty-wisdom | Retribution | Undead | 8 | 0 / 0: Divine Precision 2/2: +8% on Consecration, Exorcism, Holy Shock, Holy Wrath (spell) and Holy Strike (melee) only, not seals/Judgement; those total 16 spell, 16 Holy Strike | 8 / 8 | melee 8/8 (+0); spell 8/16 (-8) |
| paladin__level_60_shock_ret | Retribution | Undead | 8 | 0 / 0: Divine Precision 2/2: +8% on Consecration, Exorcism, Holy Shock, Holy Wrath (spell) and Holy Strike (melee) only, not seals/Judgement; those total 16 spell, 16 Holy Strike | 8 / 8 | melee 8/8 (+0); spell 8/16 (-8) |
| paladin__level_60_shock_ret-duty-crusader | Retribution | Undead | 8 | 0 / 0: Divine Precision 2/2: +8% on Consecration, Exorcism, Holy Shock, Holy Wrath (spell) and Holy Strike (melee) only, not seals/Judgement; those total 16 spell, 16 Holy Strike | 8 / 8 | melee 8/8 (+0); spell 8/16 (-8) |
| paladin__level_60_shock_ret-duty-wisdom | Retribution | Undead | 8 | 0 / 0: Divine Precision 2/2: +8% on Consecration, Exorcism, Holy Shock, Holy Wrath (spell) and Holy Strike (melee) only, not seals/Judgement; those total 16 spell, 16 Holy Strike | 8 / 8 | melee 8/8 (+0); spell 8/16 (-8) |
| shaman__enhancement-level60 | Enhancement | Orc | 8 | 2 / 2: Tidal Focus 2/5 (2% all) | 10 / 10 | melee 10/8 (+2); spell 10/16 (-6) |
| shaman__enhancement-imp-gw | Enhancement | Orc | 8 | 1 / 1: Tidal Focus 1/5 (1% all) | 9 / 9 | melee 9/8 (+1); spell 9/16 (-7) |
| shaman__enhancement-water-shield | Enhancement | Orc | 8 | 5 / 5: Tidal Focus 5/5 (5% all) | 13 / 13 | melee 13/8 (+5); spell 13/16 (-3) |
| shaman__enhancement-mana-tide | Enhancement | Orc | 8 | 5 / 5: Tidal Focus 5/5 (5% all) | 13 / 13 | melee 13/8 (+5); spell 13/16 (-3) |
| shaman__elemental-level60 | Elemental | Windshaper | 10 (+set 1) | 5 / 5: Tidal Focus 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| shaman__elemental-level60-mana-tide | Elemental | Windshaper | 10 (+set 1) | 5 / 5: Tidal Focus 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| druid__feral-cat | Feral | Night Elf | 8 | 0 / 0: none | 8 / 8 | melee 8/8 (+0) |
| druid__feral-cat-duty-ff | Feral | Night Elf | 8 | 0 / 0: none | 8 / 8 | melee 8/8 (+0); Faerie Fire (spell) 8/16 (-8) |
| druid__balance | Balance | Night Elf | 11 (+set 1) | 4 / 4: Nature's Reach 2/2 (4% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| druid__balance-duty-ff | Balance | Night Elf | 11 (+set 1) | 4 / 4: Nature's Reach 2/2 (4% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| druid__balance-wrath | Balance | Night Elf | 11 (+set 1) | 4 / 4: Nature's Reach 2/2 (4% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| druid__balance-wrath-duty-ff | Balance | Night Elf | 11 (+set 1) | 4 / 4: Nature's Reach 2/2 (4% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| mage__frostfire | Frostfire | Human | 9 (+set 1) | 0 / 6: Precision 3/3 (6% all); set 2pc +1 | 10 / 16 | spell 16/16 (+0) |
| mage__fire | Fire | Human | 9 (+set 1) | 0 / 6: Precision 3/3 (6% all); set 2pc +1 | 10 / 16 | spell 16/16 (+0) |
| mage__arcane | Arcane | Gnome | 10 (+set 1) | 0 / 6: Precision 3/3 (6% all); set 2pc +1 | 11 / 17 | spell 17/16 (+1) |
| mage__frost | Frost | High Order | 9 (+set 1) | 0 / 6: Precision 3/3 (6% all); set 2pc +1 | 10 / 16 | spell 16/16 (+0) |
| warlock__destruction | Destruction | Orc | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__demonology | Demonology | Orc | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__destruction-duty-curse | Destruction | Orc | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__destruction-duty-recklessness | Destruction | Orc | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__demonology-duty-curse | Demonology | Orc | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__demonology-duty-recklessness | Demonology | Orc | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__affliction | Affliction | Undead | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__affliction-duty-curse | Affliction | Undead | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__affliction-duty-recklessness | Affliction | Undead | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__affliction_shadow_bolt | Affliction | Orc | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__destruction_fire | Destruction | Undead | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__destruction_fire-duty-curse | Destruction | Undead | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| warlock__destruction_fire-duty-recklessness | Destruction | Undead | 10 (+set 1) | 5 / 5: Suppression 5/5 (5% all); set 2pc +1 | 16 / 16 | spell 16/16 (+0) |
| priest__shadow | Shadow | Undead | 10 (+set 1) | 0 / 0: set 2pc +1; Shadow Focus 5/5: +5% Shadow spells only; Shadow total 16 | 11 / 11 | Shadow spells 16/16 (+0) |
| priest__shadow-holy | Shadow | Undead | 10 (+set 1) | 0 / 0: set 2pc +1; Shadow Focus 5/5: +5% Shadow spells only; Shadow total 16 | 11 / 11 | Shadow spells 16/16 (+0) |
| priest__smite | Smite | Undead | 0 | 0 / 0: Shadow Focus 5/5: +5% Shadow spells only; Shadow total 5; Holy Precision 3/3: +18% Smite, Holy Fire, Penance only; total 18 | 0 / 0 | spell (generic) 0/16 (-16); Shadow spells 5/16 (-11); Smite/Holy Fire/Penance 18/16 (+2) |

Reading notes. Fury 1H (warrior__fury*) has Furious Precision 3/3, +10% on the off hand only, so the off hand white line
is melee total + 10. Paladin Divine Precision (sub-holy and shock builds) adds +8% on Consecration, Exorcism, Holy Shock,
Holy Wrath and Holy Strike only, so those spells total 16 (Holy Strike melee 16) but the seals and Judgements stay at 8.
Priest Shadow Focus 5/5 is +5% on Shadow spells only (so Shadow = 16 for both Shadow builds, exact cap); the generic
spell figure is 11 for them. Priest Holy Precision 3/3 is +18% on Smite, Holy Fire and Penance only. Survival Tactics 2/2
is +10% on traps and Feign Death only. Mage Precision 3/3 is +6% on all spells (renamed from Elemental Precision; no
school limit in the code, sim/mage/talents.go). Nature's Reach, Tidal Focus, Suppression, Precision (rogue, paladin),
Surefooted give generic hit (both melee and spell), so they are already in the totals.

## Off cap (by 0.5% or more, relevant caps only)

Over a cap (wasted hit):
- Rogues, yellow 8%: combat-swords 10 (+2), combat-daggers 14 (+6, Flask of Natural Accuracy gives 5 of it), seal-fate-swords 10 (+2), seal-fate-daggers 10 (+2), subtlety 11 (+3), mutilate 10 (+2). Precision 3/3 is in all six. (They are far under the 27% white cap, so the extra is not wasted on white hits, only on yellow.)
- Warrior Fury 1H (fury, fury-duty-battle-shout, fury-duty-sunder), yellow 10 (+2).
- Ret Precision builds (paladin__level_60, -duty-crusader, -duty-wisdom), melee 11 vs 8 (+3).
- Enhancement shaman, melee vs 8: enhancement-level60 10 (+2), enhancement-imp-gw 9 (+1), enhancement-water-shield 13 (+5), enhancement-mana-tide 13 (+5).
- Arcane mage 17 vs 16 (+1; Gnome gear gets 11, Precision 3/3 +6), under the 0.5% rule only just over.
- Smite priest, Smite/Holy Fire/Penance 18 vs 16 (+2, Holy Precision 3/3 on 0 gear hit).

Under a cap:
- Dual-wield white 27%: all Fury 1H (MH white 10, OH white 20 with Furious Precision), all rogues (10 to 14), Survival hunters (8, three builds) are far below 27.
- Ret and Enhancement spell 16%: paladin__level_60 and its duties 11 (-5); sub-holy and shock builds 8 (-8, except the Divine Precision spells at 16); Enhancement 10/9/13/13 (-6/-7/-3/-3).
- Rogue poison spell hit 16: swords/seal-fate/mutilate 10 (-6), subtlety 11 (-5), combat-daggers 14 (-2).
- Feral cat duty-ff, Faerie Fire spell hit 8 vs 16 (-8).
- Smite priest: generic spell hit 0 (-16, no hit gear at all); Shadow spells 5 (-11); only Smite, Holy Fire and Penance are covered.

On cap (within 0.5): Arms and Fury 2H (8/8), Ret sub-holy and shock builds, melee (8/8), Marksmanship and Beast Mastery ranged (8/8), Survival ranged/yellow 8/8, Feral cat melee 8/8, Elemental (16/16, both builds), Balance (16/16, four builds), Frost, Fire, Frostfire mage (16/16), all warlock builds 16/16, Shadow priest (Shadow 16/16).

## Hit talents in play (rank taken, by which builds)

- Precision (rogue, Combat tree) 3/3: +3% melee and spell: all 6 rogue builds.
- Furious Precision (warrior, Fury) 3/3: +10% off-hand only: warrior__fury, -duty-battle-shout, -duty-sunder.
- Surefooted (hunter, Survival) 3/3: +3% melee and spell: the 3 Survival builds. Survival Tactics 2/2 (+10% traps only): same 3.
- Precision (paladin, Protection) 3/3: +3% melee and spell: paladin__level_60 and its crusader and wisdom duties.
- Divine Precision (paladin, Holy) 2/2: +8% on Holy spells (school-limited): the sub_holy builds (3 Undead and duties) and the 3 shock_ret builds.
- Tidal Focus (shaman, Restoration) generic hit: Enhancement level60 2/5 (+2), imp-gw 1/5 (+1), water-shield 5/5 (+5), mana-tide 5/5 (+5); both Elemental builds 5/5 (+5).
- Nature's Reach (druid, Balance) 2/2: +4% melee and spell: the 4 Balance builds.
- Precision (mage, Frost) 3/3: +6% all spells: all 4 mage builds.
- Suppression (warlock, Affliction) 5/5: +5% spells and melee: all 13 warlock builds.
- Shadow Focus (priest, Shadow) 5/5: +5% Shadow spells only: priest__shadow, shadow-holy and smite. Holy Precision 3/3 (+18% Smite/Holy Fire/Penance only): priest__smite.
- No hit talent: Arms (all 4 + dragonkin), Fury 2H, Marksmanship, Beast Mastery, Feral (cat), Ret shock builds beyond Divine Precision.
