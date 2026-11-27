# Findings: Fairness between specs

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### fairness-20261126-3: Blackblade of Shahram's Might of Shahram stuns the stun-immune boss, and since 8f5b59325 a stunned target can't dodge, parry or block, so seven Ret paladin builds gain about 1.2 DPS
- severity: low (0.13%; no place in the ranking changes)
- affects: paladin__level_60, paladin__level_60_sub_holy, paladin__level_60_sub_holy_undead and their four duty variants (every request wearing Blackblade of Shahram 12592); not Shock Ret (other weapon)
- found: 2026-11-26, master ab3fff700
- evidence: no paladin file changed in bfbee78fc..ab3fff700, yet the 5k sweep moved Ret +1.1 and Sub Holy +1.3 at every race (Shock Ret, warlocks, rogues 0.0). Paired 20k, seeds 1 / 1000003, same request, bfbee78fc vs ab3fff700 binaries: paladin__level_60 930.2 / 930.6 -> 931.5 / 931.7 (+1.20), sub_holy 902.3 / 901.7 -> 903.5 / 903.0 (+1.22), shock_ret +0.00, Affliction +0.00. Cause: 8f5b59325 added `if result.Target.PseudoStats.Stunned { return false }` to applyAttackTableBlock / Dodge / Parry (sim/core/spell_outcome.go); the only Level 60 stun on the boss is Might of Shahram (sim/common/item_effects.go:545-565), which sets Stunned on every ActiveTargetUnit with no StunImmune check, though the boss is StunImmune (sim/core/target.go:192; EV-0088: a stun misses a raid boss whole, as consumes.go:1438 does for explosives). Fairly sure of the cause (only stun setter at Level 60, only Blackblade builds moved); not confirmed with a patched binary. Script: $AUDIT_WORK/fairness/shahram/paired.py (scratch).
- 2026-11-26 (pass 63, 2af970d74): no report request wears Blackblade of Shahram any more (518f23165: Ret to Destiny, Sub Holy and vs Undead back to Warblade), so the leak no longer touches the report; the missing StunImmune check in item_effects.go is still there and still applies to anyone who sims the blade. Kept open (low).
- proposed fix: in Might of Shahram's ApplyEffects skip targets with StunImmune when sim.IsForever() (as the explosives do); expected -1.2 on the seven builds.
- needs: orchestrator fix

### fairness-20261126-2: Arcane's best race is Undead (+2.2 over Gnome) and Fury DW's is Undead (+1.6 over Orc)
- severity: low (0.2% to 0.3%; no place in the ranking changes)
- affects: mage__arcane (report race Gnome), warrior__fury (report race Orc)
- found: 2026-11-26, master bfbee78fc
- evidence: paired 20k, seeds 1 / 1000003, report request with only the race changed (Arcane Undead on its committed Horde race file req/races/mage__arcane.Undead.json): Arcane Gnome 833.9 / 833.9, Undead 836.1 / 836.0 (+2.15); Fury DW Orc 983.6 / 984.0, Undead 985.5 / 985.2 (+1.55). 5k sweep agrees (Arcane Undead 836.2 vs Gnome 834.3; Fury DW Undead 986.2 vs Orc 984.2). Both came with the Phase 0 re-rank (at 85691655a Gnome and Orc were best). Script: $AUDIT_WORK/fairness/race.py (scratch).
- 2026-11-26 (pass 62, ab3fff700): 5k sweep Arcane Undead 836.2 vs Gnome 834.3 (+1.9, unchanged); Fury DW after its regear Undead 994.6 vs Orc 993.9 (+0.7, now noise). Open for Arcane.
- 2026-11-26 (pass 63, 2af970d74): Arcane unchanged (Undead +1.9 at 5k); Fury DW after the round 2 trinkets Undead 987.5 vs Orc 986.3 (+1.2, near noise). Open for Arcane.
- proposed fix: re-pick the report race (standards 16 / 29) with a race rotation check, or record why Gnome / Orc stay.
- needs: orchestrator fix

### fairness-20261125-2: The explosive re-pick after EV-0088 reached only the hunters and Feral; every other melee build still takes no filler explosive and loses 2.7 to 7.3 DPS to Goblin Land Mine
- severity: high (0.3% to 0.9%; Mutilate with the mine (802.1) passes Subtlety (798.1), which already carries it)
- affects: warrior__fury, warrior__fury_2h, rogue__mutilate, rogue__combat-swords, rogue__combat-daggers, rogue__seal-fate-swords, paladin__level_60 (and shock_ret, sub_holy), shaman__enhancement-level60 (and imp-gw), hunter__survival; their duty / race variants likely the same (not simmed)
- found: 2026-11-25, master 85691655a
- evidence: standards 11 (22d2cf952) says a thrown explosive now keeps the swing timer and "the picks are due a re-check"; 22d2cf952 re-picked only hunter (Dense Dynamite) and feral (Goblin Land Mine) in mksettings / presets (results-20261125-1 covers those requests' drift). Rogue Subtlety and Seal Fate Daggers already carry the Land Mine (-7.0 / -6.6 without it). Paired, report request with only fillerExplosive changed, 10k iterations, seeds 1 / 1000003, CLI 85691655a (tools/audit/fairness/explosives.py): Goblin Land Mine over none: Fury DW +3.3, Fury 2H +2.7, Mutilate +6.8, Combat Swords +6.5, Combat Daggers +6.2, Seal Fate Swords +6.9, Ret +7.3, Shock Ret +6.8, Sub Holy +7.1, Enhancement +6.3, Enh imp-gw +5.0, Survival +4.6, Survival wolf +5.2; Arms dragonkin +0.3 (noise); Dense Dynamite 0.6 to 1.2 below the mine on each; EZ-Thro Dark Bomb loses on all (stun-immune boss). Arms keeps none (mine -1.4, Dynamite -2.1). Casters gain nothing (Fire, Affliction, Elemental: mine +0.0, Dynamite -5 to -6). Sure of the numbers.
- 2026-11-26 (pass 61, bfbee78fc): Survival and the four Enhancement requests now carry Goblin Land Mine. Still none on the rest; paired 10k, seeds 1 / 1000003, tools/audit/fairness/explosives.py, mine / Dynamite over none: Fury DW +4.1 / +3.4, Fury 2H +2.1 / +1.4, Mutilate +6.8 / +5.8, Combat Swords +6.5 / +5.7, Combat Daggers +6.1 / +5.2, Seal Fate Swords +6.5 / +5.8, Ret +6.5 / +5.7, Shock Ret +7.7 / +6.8; Arms -1.2 / -1.9 (stays none). Still open for Fury, four rogues and Ret.
- 2026-11-26 (pass 62, ab3fff700): Fury, Fury 2H, Mutilate, Combat, Seal Fate Swords and the Ret requests still have no fillerExplosive (the Fury DW regear did not add one). Not re-measured. Still open.
- 2026-11-26 (pass 63, 2af970d74): after the prices round 2 re-ranks, Fury, Fury 2H, Mutilate, Combat Swords, Combat Daggers, Seal Fate Swords, Ret, Shock Ret and Sub Holy still carry no fillerExplosive (Seal Fate Daggers and Subtlety do). Not re-measured. Still open.
- proposed fix: set fillerExplosive Goblin Land Mine in the listed requests, their presets.ts and mksettings (Fury, rogues, Ret, Enhancement, Survival), confirm the duty variants with explosives.py, and re-sim; Arms and the casters stay on none.
- needs: orchestrator fix

### fairness-20261125-3: Demonology's report race (Orc) is no longer its best; Undead sims 5.7 higher (Destruction +2.2)
- severity: medium (0.6% on warlock__demonology; no place in the ranking changes)
- affects: warlock__demonology (and its duty variants, not simmed), warlock__destruction (low)
- found: 2026-11-25, master 85691655a
- evidence: paired 20k, seeds 1 / 1000003, report request with only the race changed (no Undead race rotation exists): Demonology Orc 885.2 / 884.5, Undead 890.4 / 890.7 (+5.7); Destruction Orc 870.0 / 869.9, Undead 871.9 / 872.4 (+2.2). 5k sweep agrees (Demonology Undead 891.1, Orc 885.4; Destruction Undead 871.7, Orc 870.2). At 0404484fb Human / Orc / Undead were within 1.6 on Demonology. The likely cause is efeb06b95 (resist constant 8L+19 raises the boss's resist, which makes Undead's racial worth more; Shadow priest's Undead also overtook Troll), not checked further. Script: $AUDIT_WORK/fairness/paired/race.py (scratch).
- 2026-11-26 (pass 61, bfbee78fc): after the warlock re-rank, paired 20k: Demonology Orc 877.6 / 876.5, Undead 882.6 / 883.3 (+5.9); Destruction Orc 868.4 / 867.9, Undead 869.8 / 870.2 (+1.85). Still open.
- 2026-11-26 (pass 62, ab3fff700): no warlock change; 5k sweep Demonology Undead 882.1 vs Orc 877.9, Destruction 870.9 vs 869.2. Still open.
- 2026-11-26 (pass 63, 2af970d74): no warlock change; 5k sweep unchanged (Demonology Undead +4.2, Destruction +1.7). Still open.
- proposed fix: move Demonology (and Destruction if the +2.2 holds with a race rotation check, standards 29) to Undead and re-sim; or record why Orc stays.
- needs: orchestrator fix

### fairness-20261125-1: Marksmanship's report race (Orc) is no longer its best on the Phase 0 set; Skyborne High Order sims 6.1 higher
- severity: medium (0.56% on hunter__marksmanship, 0.65% on the wolf variant; MM keeps 2nd place)
- affects: hunter__marksmanship (report race Orc), hunter__marksmanship-wolf, hunter__marksmanship-tallstrider
- found: 2026-11-25, master 63129ce35
- evidence: paired 20k, seeds 1 / 1000003, CLI /tmp/audit-cache/cli/63129ce35290: hunter__marksmanship Orc 1090.4 / 1090.3, High Order on its committed race file req/races/hunter__marksmanship.SkyborneHighOrder.json (Alliance set) 1096.4 / 1096.5 (+6.1; the same on the Horde set); hunter__marksmanship-wolf Orc 1070.4, High Order 1077.4 (+7.0, seed 1). 5k sweep: MM High Order 1096.3 vs Orc 1090.0, tallstrider +6.8, wolf +7.5. Beast Mastery is a tie (Orc 1009.5 / 1009.4, High Order 1009.5 / 1009.1). Came with the hunter Phase 0 regear (8e04994c5, merged 3fdc2dca4); on the synthetic set Orc was within 1.0 of the best. Script: $AUDIT_WORK/fairness/paired/p.py (scratch). Standards 16 says presets use the best race.
- 2026-11-25 (pass 59, 0404484fb): 5k sweep MM High Order 1088.8 vs Orc 1082.2 (+6.6) after the elixir overwrite; variants +7.1 / +7.7. Still open.
- 2026-11-25 (pass 60, 85691655a): 5k sweep MM High Order 1100.9 vs Orc 1095.9 (+5.0); wolf +6.7, tallstrider +6.4. Still open.
- 2026-11-26 (pass 61, bfbee78fc): after the hunter re-rank, 5k sweep MM High Order 1087.5 vs Orc 1082.4 (+5.1); wolf 1067.5 vs 1060.9 (+6.6), tallstrider 1059.5 vs 1052.9 (+6.6). Still open.
- 2026-11-26 (pass 62, ab3fff700): no MM request change in range, still Orc; 5k sweep unchanged (High Order 1087.5 vs Orc 1082.4, variants +6.6). Still open.
- 2026-11-26 (pass 63, 2af970d74): still Orc; 5k sweep unchanged (High Order +5.1, variants +6.6). Still open.
- proposed fix: move the three Marksmanship requests, metas and preset to Skyborne High Order on level60_phase0_hunter_mm_alliance (or check a High Order race rotation first, standards 29) and re-sim; or record why Orc stays.
- needs: orchestrator fix

### fairness-20261124-2: On the Phase 0 sets, Fire's best race is Undead (+2.2 over Human) and Shadow's is Troll (+1.9 over Undead)
- severity: low (0.2% each; no place in the ranking changes)
- affects: mage__fire (report race Human), priest__shadow (report race Undead)
- found: 2026-11-24, master eceb09019
- evidence: paired 20k, seeds 1 / 1000003, report request with only the race changed (no race rotation exists for these races): Fire Human 920.8 / 920.2, Undead 922.8 / 922.5 (+2.2), High Order with req/races/mage__fire.SkyborneHighOrder 906.4 / 906.6; Shadow Undead 855.2 / 855.2, Troll 857.3 / 856.9 (+1.9). The 5k sweep agrees (Fire Undead 923.3 vs Human 920.4, Shadow Troll 857.3 vs Undead 855.5). Both came with the Phase 0 regear (8b5137c6d sweep: Fire Human best, Shadow Undead best). Scripts: $AUDIT_WORK/fairness/paired/run.py (scratch).
- 2026-11-25 (pass 56, 63129ce35): Shadow on its new talents (a2d195d46): paired 20k Undead 857.6 / 857.7, Troll 859.5 / 859.4 (+1.8). Fire unchanged (5k sweep Undead 923.3, Human 920.4). Still open.
- 2026-11-25 (pass 59, 0404484fb): 5k sweep after the elixir overwrite: Fire Undead 913.9 vs Human 913.1 (+0.8, now near noise; needs a paired check); Shadow Troll 856.0 vs Undead 854.3 (+1.7). Still open.
- 2026-11-25 (pass 60, 85691655a): Shadow fixed by the rules, not the request: paired 20k Undead 869.2 / 869.2, Troll 865.5 / 865.5 (Undead, the report race, best by 3.7). Fire: paired 20k Human 918.5 / 917.8, Gnome 918.9 / 918.3 (+0.45), Undead 918.7 / 919.7 (+1.05), at the edge of noise. Open for Fire only (low).
- 2026-11-26 (pass 61, bfbee78fc): after the mage re-rank, 5k sweep Fire Gnome 918.9, Human 917.8 (+1.1), Undead 916.5. At the edge of noise; still open (low).
- proposed fix: re-pick the report race (standards 16 / 29), Fire to Undead and Shadow to Troll, with a race rotation check; or note in the metas why Human / Undead stay.
- needs: orchestrator fix

### fairness-20261118-1: Balance's report race (Night Elf) is no longer its best; High Order sims 3.7 to 6.1 higher since haste shortens the GCD
- severity: medium (0.4% to 0.7% on all four Balance builds; Balance keeps its class-level place between Elemental and Shadow)
- affects: druid__balance (default), druid__balance-duty-ff, druid__balance-wrath, druid__balance-wrath-duty-ff
- found: 2026-11-18, master a6677362f
- evidence: the published report's own race table (ui/dps_report/data/builds.json at a6677362f, raceIters 10000) now names High Order as bestRace for all four: Balance 955.6 vs 959.4 (+3.8), FF duty 920.0 vs 924.9 (+4.9), Wrath 865.8 vs 871.9 (+6.1), Wrath FF duty 833.9 vs 840.0 (+6.1). At 8219f1083 Night Elf was best for all four. Paired check, shared CLI a6677362fc06, 20000 iterations, seeds 1 / 1000003, report request with only the race changed: druid__balance Night Elf 956.1 / 956.2, High Order with its race rotation (req/races/druid__balance.SkyborneHighOrder.json) 959.8 / 959.8 (+3.7), High Order on the preset rotation 948.8 / 949.0; druid__balance-wrath Night Elf 866.0 / 866.1, High Order (no race rotation) 871.9 / 871.9 (+5.9). Sure of the numbers. Cause: ee35da473 / c2efe4ae3 (standards 27, haste shortens a spell GCD to 1s) makes the Skyborne 2% haste racial worth more; the 87ce834 regeneration re-simmed at the old race. Standards 16 / 29 say presets re-pick to the best race; this is outside noise.
- 2026-11-22 (pass 49, 7a6d1beb9): no druid or race change in range; Balance requests still Night Elf, shared 20k run Balance 956.1 / Wrath 866.0 (baseline Wrath High Order 872.5). Still open.
- 2026-11-23 (pass 50, 5518f4873): no druid or race change in range; builds.json race table still High Order +3.8 to +6.1 for all four. Still open.
- 2026-11-23 (pass 51, 2de7021bf): no druid or race change in range; builds.json race table unchanged. Still open.
- 2026-11-24 (pass 53, 566067a39): no druid Balance or race change in range (be773f5eb touches Rip and Rake only); shared 20k run Balance 956.1 unchanged. Still open.
- 2026-11-24 (pass 55, eceb09019): the four Balance requests moved to the Phase 0 set (still Night Elf). Paired 20k, seeds 1 / 1000003, report request with only the race changed, race rotation where req/races has one: Balance NE 822.3 / 822.1, High Order (race rotation) 826.5 / 826.1 (+4.1), Windshaper 825.7 / 825.4; Wrath NE 730.1 / 730.4, High Order 735.1 / 736.0 (+5.3); Balance FF duty NE 789.0 / 788.3, High Order (race rotation) 793.5 / 793.5 (+4.9); Wrath FF duty NE 702.8 / 702.7, High Order 708.5 / 708.5 (+5.8). The 5k sweep (base rotation only) shows Night Elf best for default Balance, but with the committed race rotation High Order still wins all four. Still open.
- 2026-11-25 (pass 56, 63129ce35): no druid change in range; 5k sweep Balance unchanged (Night Elf 822.9, Wrath High Order 734.5 vs Night Elf 730.5). Still open.
- 2026-11-25 (pass 59, 0404484fb): no druid Balance change; 5k sweep (base rotation) Balance Night Elf 818.8, Wrath High Order 732.4 vs Night Elf 727.7. Still open.
- 2026-11-25 (pass 60, 85691655a): no Balance code change; 5k sweep (base rotation) Night Elf 816.1, Wrath High Order 735.3 vs Night Elf 731.5. Still open.
- 2026-11-26 (pass 61, bfbee78fc): Balance re-ranked (a60eb2ca1, 6a6ea270c), still Night Elf. 5k sweep (base rotation): Balance Night Elf 802.4 best; Wrath High Order 718.3 vs Night Elf 714.2 (+4.1). Race rotations not re-checked this pass. Still open.
- proposed fix: set the four Balance requests' and presets' race to Skyborne High Order (Balance and Balance FF duty with the race-tuned rotation the race table already uses) and re-sim them; or record in standards why Night Elf stays.
- needs: orchestrator fix

## Resolved

### fairness-20261127-1: The Armaments of Battle removal and the Enhancement trinket re-rank changed nine DPS gear sets, but no report request followed; 37 published builds still wear out-of-game rewards or an old pick, Affliction by +17 and Demonology by +10
- severity: high (Affliction -1.8% drops below Survival and Enhancement imp-gw; Demonology -1.1% drops below Destruction and Enhancement mana-tide; Enhancement +0.6%)
- affects: Affliction (4 requests), Demonology (3), Destruction and Destruction Fire (6), Balance (4), Feral (2), Fury 2H (3), Elemental (2), Smite, Enhancement (4)
- found: 2026-11-27, master 34994816e
- evidence: 287da91da / 6bd2d60a2 put the 9 Armaments of Battle rewards out (no quests at Light's Hope, user Nov 26) and changed the Phase 0 gear sets of Balance, Elemental, Feral, Fury, Smite, Affliction, Demonology, Destruction (Talisman of Ascendance 22678, Bracers of Subterfuge 22668, Verimonde's Last Resort 22688 out); 3698a73d7 moved the Enhancement set from Chained Essence 10455 to Darkmoon Card: Maelstrom 19289 after the 2.7% tick crit fix (5591bd019). Only paladin__level_60_shock_ret.meta.json changed under tools/dps_report/req; phase0_check.py now flags 37 requests "its equipment is not <set>" and the out items at every race file. Paired 10k, seeds 1 / 1000003, 34994816e CLI, request with only its equipment set to the current gear set (tools/audit/fairness/gearset.py f9fa79917 <id>): Affliction 962.8 -> 945.6 (-17.2), duty-curse -17.7, duty-recklessness -18.4, Shadow Bolt -10.2; Demonology 876.8 -> 867.2 (-9.6), duties -10.5 / -9.3; Destruction 868.4 -> 866.8 (-1.5), duties -2.5 / -2.6; Destruction Fire -0.7 / -0.7 / -1.2; Balance 801.8 -> 796.2 (-5.7), FF duty -5.1, Wrath -2.7, Wrath FF -1.9; Feral 1057.5 -> 1054.1 (-3.3), FF duty -3.4; Fury 2H 994.5 -> 991.7 (-2.8), duties -2.6 / -2.6; Elemental +1.6, mana-tide +0.8; Smite -0.0; Enhancement 966.4 -> 972.5 (+6.1), imp-gw +6.5, mana-tide +6.0, water-shield +5.4. The warlock APL change in the same commit does not move these (current APL plus gear set gives the same numbers). The published ranking now credits eight specs with gear they cannot get. Sure of the numbers.
- 2026-11-27 (pass 66, 5ab589d85): the Talisman of Ascendance every-cast fix (413f06cb0) makes the Talisman worth less, so the regear costs less; no request or gear set changed. Paired 10k, seeds 1 / 1000003, 5ab589d85 CLI, tools/audit/fairness/gearset.py (now committed): Affliction 962.8 -> 945.6 (-17.2), duty-curse 948.1 -> 935.7 (-12.4), Shadow Bolt 818.4 -> 810.4 (-8.0), Demonology 873.8 -> 867.2 (-6.6), Destruction 866.6 -> 866.8 (+0.2), Balance -5.6, Elemental +1.7, Enhancement +6.1. Rank effect now: Affliction still below Survival and Enhancement imp-gw; Demonology below Enhancement mana-tide, level with Destruction. phase0_check.py still flags the same requests. Still open.
- overlap: gear-20261127-1 (same pass run) names the out items on 29 requests; this finding adds the measured DPS, the ranking effect and the four Enhancement requests (Maelstrom), which are not about out items. One fix covers both.
- proposed fix: regenerate the listed requests (and their race files and metas) from the current gear sets, then re-sim and re-publish them together.
- needs: orchestrator fix
- 2026-11-27 (pass 67, branch phase0-regear 313069ce, not master): 94d6cc79 put the current Phase 0 gear set on the 29 requests (race rows and verify 48cc3aed), 313069ce re-simmed and re-published them. phase0_check.py no longer flags any DPS request ("equipment is not" and out items gone; the only out item left is the restoration druid neck, a healer list). builds.json moves match the measured sizes: Affliction 962.5 -> 945.1 (-17.4, measured -17.2), duty-curse -12.8 (-12.4), Shadow Bolt -7.2 (-8.0), Demonology -6.8 (-6.6), Destruction -0.2 (+0.2), Balance -6.2 (-5.6), Feral -3.5 (-3.3), Fury 2H -2.9 (-2.8), Elemental +1.9 (+1.7), Enhancement +6.3 (+6.1), Smite -0.1 (-0.0). Each of the 29 requests simmed at 313069ce (10k, seed 1) is within 0.9 of builds.json, and its meta.published.dps equals builds.json. No other build moved.
- resolved: 2026-11-27, 313069ce (branch phase0-regear; resolved on master when it merges)

### fairness-20261126-4: Ret leaves out Darkmoon Card: Maelstrom as "doubtful" though standards 42 keeps unsure gear in and Arms, Feral, Fury DW and Seal Fate Swords wear it; Ret loses 3.8 DPS
- severity: medium (0.4% on paladin__level_60, 0.3% on Sub Holy; Ret keeps 11th)
- affects: paladin__level_60 (and its two duty variants), paladin__level_60_sub_holy (and duty variants); Shock Ret likely (not simmed)
- found: 2026-11-26, master 2af970d74
- evidence: phase0_check.py flags "ret trinket: open Darkmoon Card: Maelstrom (19289) sits 1.87 above the pick" (ui/retribution_paladin/bis/level60_phase0_ret.bis.json). The round notes (518f23165, phase0_rerank_shaman_ret.md:91) say "Maelstrom (doubtful) came out 0.4 and 0.9 over Blackhand's Breadth, not confirmed", measured before Ret moved to Destiny. Paired 20k, seeds 1 / 1000003, 2af970d74 CLI, report request with one trinket swapped ($AUDIT_WORK/fairness/swap.py, scratch): paladin__level_60 932.8 / 932.6 -> Maelstrom for Chained Essence of Eranikus 936.6 / 936.6 (+3.83), for Blackhand's Breadth 936.6 / 935.8 (+3.42); paladin__level_60_sub_holy Eranikus -> Maelstrom 905.3 / 905.1 -> 907.4 / 908.0 (+2.52). Standards 42 says an unsure item is used, not asked about; the same card is a pick for Arms (dac79e5f6), Feral (f4e2b6e46), Fury DW and Seal Fate Swords (8c5ebbb30). Sure of the numbers.
- 2026-11-27 (pass 65, 34994816e): after the Chained Essence 2.7% crit fix (5591bd019) Enhancement moved to Maelstrom (3698a73d7), Ret did not. Paired 10k, seeds 1 / 1000003, Chained Essence -> Maelstrom on paladin__level_60_shock_ret: 899.9 -> 902.8 (+2.9). Rogues, Feral and Survival already wear both. Still open for Ret, Sub Holy and Shock Ret.
- proposed fix: Ret and Sub Holy take Darkmoon Card: Maelstrom in place of Chained Essence of Eranikus (list, gear set, requests, race rows), check Shock Ret the same way, and re-sim.
- needs: orchestrator fix
- 2026-11-27 (pass 67, branch phase0-regear 313069ce): ee28c064 re-ranked the Ret trinkets with Maelstrom in (still tagged doubtful, but ranked, not left out): the rank put it at -1.85, rotopt's 20k confirm read -0.9 (0.5) and the picks stayed. Why the confirm and the report fight disagree, measured on the list's own build paladin__level_60 (phase0_specs.json "ret"), paired 10k, seeds 1 / 1000003, 313069ce CLI, labeled rands, itemUsesOff for the "use off" runs (/tmp/claude-0/aud-fair/ret_twopart*.txt, helper bin/run.py, scratch):
  - report fight (180 +/- 20 s, uses pressed): set 931.75 / 932.16; Maelstrom for Chained Essence 935.81 / 936.37 (+4.1), for Blackhand's Breadth 934.23 / 934.74 (+2.5).
  - the list's two-part score (standards 43: 180 s with every use off, plus each use's value on 90 s): set 920.74 + 16.23 (Chained Essence use) + 6.58 (Breadth use) = 943.55, seed 2 943.66; Maelstrom for Breadth 926.40 + 15.85 = 942.25 / 942.77 (-1.1, the confirm's -0.9); Maelstrom for Chained Essence 932.73 + 6.14 = 938.87 / 939.61 (-4.4).
  - the cause: a once-a-fight use is worth about twice as much on 90 s as in the 180 s report fight. Chained Essence's use is 16.2 on 90 s and 8.0 on 180 s (931.75 - 923.75), Breadth's 6.6 and 3.2. Maelstrom is passive and worth the same on both, so the list picks the use trinkets and the report fight prefers Maelstrom.
  - judgement: Ret's trinkets now follow standards 43 like every other DPS list (all 26 have use_dur 90 on 180 s), and the doubtful tag no longer keeps Maelstrom out, which was this finding's point. The report-fight gap is the ruled scoring, not a Ret defect, and it is not Ret's alone: from the lists' own cooldown losses, about 20 DPS specs pick a use trinket whose 90 s use value is more than twice the gap to their best passive one (rough estimate, not simmed; table in reports/fairness.md). phase0_check.py's "Maelstrom sits 1.85 above the pick" is the rank's number, which the confirm overrides.
- resolved: 2026-11-27, ee28c064 (branch phase0-regear): settled by standards 43's two-part score; whether the report should show the 180 s-best trinkets instead is a question about that rule for every spec, not this finding

### fairness-20261126-1: The Phase 0 lists of the six rogues, the three warriors and Survival were searched on a sim without the proc, PPM, set-proc and explosive rules every other list had
- severity: medium (pick effect not measured; those rules moved the rogues 11.6 to 26.4 DPS and Fury DW 13, so set and weapon picks may differ)
- affects: rogue__mutilate, rogue__combat-swords, rogue__combat-daggers, rogue__seal-fate-swords, rogue__seal-fate-daggers, rogue__subtlety, warrior__arms, warrior__fury, warrior__fury_2h, hunter__survival (and their variants)
- found: 2026-11-26, master bfbee78fc
- evidence: the lists' `sim` stamps (phase0_check.py, ui/<class>/bis/level60_phase0_*.bis.json): rogues ac75336346de / c6fd1275d90e / 8c9f0e3b77dc / 10d171144cea, warriors c6fd1275d90e / 6c2211c6d565, Survival 6c2211c6d565; none has 295206ddf (Forever weapon PPMs, dungeon set procs), b663145ad (proc rules), efeb06b95 or 22d2cf952 as an ancestor (`git merge-base --is-ancestor`), and their effects.json stamps (ab9f67d6, 179335bf, c7e3b3e1, d28dd52c, 741ce1da) don't either. Mage, warlock, druid, priest, shaman, Ret and the other hunter lists are on 34fe148a5 / c815dd28f / d38dae9a9, which have no sim/ diff to master. 51a4f44a3 says "some on the pre-merge sim, to be redone"; these were not redone. Standards 43 / the pass doc: no spec's search may run on an older sim than the rest.
- 2026-11-26 (pass 62, ab3fff700): the Arms, Fury, Fury DW and Survival lists were re-ranked on 34fe148a5 (7a3b753cb, f79e810e1, 5f50e8841, 02c06a709; phase0_check.py), which has the batch rules; their requests moved (Survival +13.0, Fury DW +8.4, Arms +2.3 at 5k). The six rogue lists are still on ac75336346de / c6fd1275d90e / 8c9f0e3b77dc / 10d171144cea. Open for the rogues only.
- proposed fix: re-run the Phase 0 search for these ten specs on the current sim (with fairness-20261125-2's explosive in place) and re-sim the changed requests.
- needs: orchestrator fix
- resolved: 2026-11-26, 2af970d74: the six rogue lists were re-ranked on 6a5cacef8 (e9da0ddcd, f3bf14c7c, 26e4300f1), which has 295206ddf, b663145ad, efeb06b95, 22d2cf952 and 34fe148a5 as ancestors and no sim/ diff to master; every DPS list is now on 34fe148a5 or later (phase0_check.py).

### fairness-20261124-1: Report requests mix Phase 0 real gear (7 classes) with Phase 1 synthetic sets (hunter, rogue); hunters and rogues rank about 8 to 9% too high
- severity: high (moves hunters and rogues by 8 to 9% and reorders the ranking: Lone Wolf, MM and BM sit 1 to 3 only because they still wear the stronger set)
- affects: all 10 hunter requests (993xxx / 9931xx synthetic items) and all 6 rogue requests (994xxx); against every other class's request
- found: 2026-11-24, master eceb09019
- evidence: first equipment ids in tools/dps_report/req at eceb09019: hunter__* 993001.., hunter__survival* 993101.., rogue__* 994001.. / 994101.. (the 1b synthetic sets); druid, mage, paladin, priest, shaman, warlock and warrior now wear real Phase 0 items (e.g. mage__fire 226935, warrior__fury 12640). The Phase 0 regear cut those classes 5.8% to 23% against the 8b5137c6d baseline (Feral -11.0%, Fury -8.6%, Destruction -17.5%, Frost -23.0%; 5k sweep, seed 1, report race), while hunters and rogues moved 0.0 to -0.6%. The unmerged origin/phase0-step4-hunter metas give the hunters' Phase 0 numbers: Lone Wolf 1291.8 -> 1174.0 (-9.1%), MM 1200.6 -> 1090.5 (-9.2%), Survival 1007.3 -> 926.7 (-8.0%). origin/phase0-step4-rogue has not regeared the rogue requests yet. builds.json is still the Phase 1 publication for every class, so the published page is consistent today; a re-sim and republish from current master would not be. Sure of the gear state; the rogue size is an estimate from the other classes.
- proposed fix: merge phase0-step4-hunter and phase0-step4-rogue (rogue requests regeared) before any report re-sim or republish of master; or, if the report republishes first, re-sim only from a tree where every class wears the same phase.
- needs: orchestrator fix
- resolved: 2026-11-25, 63129ce35 (hunter requests 8e04994c5 / ea670f170 merged 3fdc2dca4; rogue requests a82c8ce6a merged 6fdbcc951). No request in tools/dps_report/req carries a 99xxxx synthetic item now; every meta's published.dps is within 1.0 of the 5k sweep at its report race. Hunters moved -8.0% to -9.3%, rogues -5.7% to -11.0%.

### fairness-20261122-1: Published report lags five classes' fixes after the 30314ad60 re-sim; 28 of 68 builds off by 1.7 to 111 DPS and six specs change place
- severity: high (Fury 2H +2.5%, Frost -11.7%; the published order is wrong in six places)
- affects: warrior (all 11 builds, +5.7 to +27.4), warlock Affliction (3 builds, +11.9 to +12.8), shaman Enhancement (4 builds, -8.9 to -9.5), mage Frost / Frostfire (-110.8 / -58.3), hunter pet builds (9 builds, -1.7 to -4.6)
- found: 2026-11-22, master 7a6d1beb9
- evidence: `ui/dps_report/data/builds.json` at 7a6d1beb9 (last re-sim 30314ad60) against the loop's shared 20k run on the 7a6d1beb9 CLI (`/tmp/audit-work/runs/master`, report requests unchanged): 40 of 68 builds within 1.5 (every class whose code did not move), 28 off, each traced to a sim commit merged after 30314ad60: b7edf19bb (Execute rolls Windfury Totem twice, no Flurry charge or Bloodthrill off the totem's swing) Fury 2H 1104.2 -> 1131.6, Fury 1099.2 -> 1115.7, Arms 1072.2 -> 1080.3; 3960dd1f6 (Wrack half-point ticks) Affliction 1076.4 -> 1089.2; 840029bd8 (no Maelstrom stack from the totem) plus bd7029410 (Judgement of Wisdom on landed swings) Enhancement 1080.0 -> 1070.9; 4192d6e49 (Ice Lance 0.11 SP) Frost 945.4 -> 834.6, Frostfire 1064.3 -> 1006.0; 051a21cd3 (pet keeps summon AP) BM 1115.4 -> 1111.4, Survival -4.6, MM -2.1. Place changes in the 28-spec table: Fury 2H 5 -> 4 and Fury 6 -> 5 (both pass BM), Affliction 9 -> 7 (passes Arms and Enhancement), Enhancement 7 -> 9, Frostfire 11 -> 15, Frost 22 -> 27. The race picks hold: Frost / Frostfire re-swept at 5k seed 1 every race, High Order 834.0 best for Frost, Frostfire Human 1007.1 vs High Order 1005.5. results.md's 2026-11-22 update on results-20261104-2 names only the mage and hunter rows ("57 of 68 match"); the warrior, Affliction and Enhancement rows are not in it.
- proposed fix: re-sim the five classes' builds in the report (`regenerate-dps-report`; it re-runs the builds whose inputs changed) and commit builds.json and the metas together, so one class's fix never ships without the others' re-sim.
- needs: orchestrator fix
- resolved: 2026-11-23, 6ce43d3c8 (re-sim of the five classes) and 98dc8871e (Frost / Frostfire re-sim after the b8b89d6a8 re-tune); pass 50 on 5518f4873: builds.json within 1.5 of the shared 20k run on 68 of 68 builds (largest 0.8), the order matches.

### fairness-20261112-1: Combat Daggers is now the rogue Combat default, but Combat Swords sims 10 DPS higher at master
- severity: medium (the default build shown for Combat; no change to the cross-class ranking of the best rogue)
- affects: rogue__combat-daggers / rogue__combat-swords metas (isDefault), ui/rogue presets and sim.ts default, forever-wiki rogue page
- found: 2026-11-12, master a0b4d3be9
- evidence: 7462c17c8 (user ruling 2026-11-12, "switch to best") set Combat Daggers as default because it was the best 180 s Combat build then (707ae92 sweep: Daggers 915.4, Swords 899.6). The next commit that moved rogues, a0b4d3be9 (Windfury Totem on main hand abilities) plus e2322baa1 (Venom on Instant Poison), lifted Swords +46.5 and Daggers +21.2. At master, report race Undead, 180 s requests: Swords 946.1 vs Daggers 936.1 (shared run `$AUDIT_WORK/runs/master`, and the 5k seed 1 sweep, 946.1 vs 936.0; noise about 1.5). Daggers' best race (Gnome 936.6) does not close it.
- proposed fix: apply the ruling again on the current numbers: isDefault back to rogue__combat-swords, with its UI preset defaults and the wiki line, once results-20261112-1's re-sim confirms the order at 20k. Fury DW over 2H (1050.2 vs 1032.0), Affliction over Destruction (1071.1 vs 1063.2) and Lone Wolf over BM still hold.
- needs: orchestrator fix
- 2026-11-12, master 8306f1d52: still open after the re-sim (bff0169e4). builds.json at 20k: Combat Swords 946.4, Combat Daggers 936.2 (isDefault); 5k sweep 946.1 / 936.0. Order confirmed.
- resolved: 2026-11-13, 8f5de1c87 (Combat Swords isDefault in the metas, UI default gear/talents/rotation; builds.json e9a8857e7 Swords 946.4 default over Daggers 936.2)

### fairness-20261109-1: Windfury Totem's white-swing-only rule lets Heroic Strike and Cleave proc it but not Raptor Strike
- severity: high (Survival +24.9 DPS, +2.6%, if Raptor Strike gets the same treatment as Heroic Strike; it passes Arms in the ranking)
- affects: hunter survival (3 builds); warrior arms / fury / fury_2h (the other way); the comment on the rule
- found: 2026-11-09, master 66a243cf8
- evidence: 0ce9dbf40 sets the totem's proc mask to ProcMaskMeleeMHAuto under Forever (sim/core/buffs.go:2166). Heroic Strike and Cleave carry `ProcMaskMeleeMHSpecial | ProcMaskMeleeMHAuto` (sim/warrior/heroic_strike_cleave.go:25, :79), so they still proc the totem. Raptor Strike, also an on-next-swing attack, deals its damage through a hit spell with `ProcMaskMeleeMHSpecial` only (sim/hunter/raptor_strike.go:85), so it no longer procs. The commit comment says the rule "leaves out" on-next-swing attacks like Heroic Strike, but the code leaves them in for warriors and out for hunters. Neither case was checked in logs. Measured with an env-gated scratch hack, request as published, 5000 iterations, seed 1:
  - Raptor Strike hit also procs (adds MHAuto to its mask): survival 945.7 -> 970.5 (+24.9), survival-wolf 919.3 -> 944.1 (+24.8), survival-tallstrider 911.3 -> 936.8 (+25.5).
  - Heroic Strike / Cleave don't proc (skip MHSpecial in the totem's handler): arms 887.6 -> 882.9 (-4.8), fury 1038.1 -> 1021.5 (-16.6), fury_2h 961.8 -> 958.3 (-3.5).
- proposed fix: treat every on-next-swing attack the same way. Option A, all of them proc (give the Raptor Strike hit the MHAuto bit, or make the totem accept a queued-swing flag): Survival +25. Option B, none proc: Fury -16.6, Fury 2H -3.5, Arms -4.8. Either way fix the comment in buffs.go. A log check of Heroic Strike with the totem up decides between them. results-20261109-1 already covers the re-sim and the Stormstrike-to-all-specials question.
- needs: user ruling (A or B), then orchestrator fix
- resolved: 2026-11-12, a0b4d3be9 (the totem rolls on main hand abilities by client flag: Heroic Strike, Cleave and Raptor Strike 2973/14260-14266 all roll it, by spell ID, so the Raptor Strike hit counts; Survival +70.8, Arms +171.0 at 5k on the full change)

### fairness-20261101-2: Affliction's report race Human is 1.4 under Orc
- severity: low (0.14%, no rank change)
- affects: warlock__affliction (and its duty variants, same race)
- found: 2026-11-01, master dd053d24c
- evidence: `ITERS=20000 python3 tools/audit/fairness/sweep.py warlock__affliction` (seed 1, race overrides): Orc 1013.5, Human 1012.1 (the report race), Undead 1011.5. At 5k: Orc 1014.2, Human 1012.9. At d96adcc the two tied (1019.7 / 1019.6); the 25 yd move (905787afd) cost Human 0.4 more than Orc. Fairly sure it is real (same at 5k and 20k), but small.
- update 2026-11-02, master 796495d6f (pass 23): at 30 yd (standards 33) the gap shrinks. `ITERS=20000 sweep.py warlock__affliction`: Orc 1013.6, Human 1012.7 (+0.9), Undead 1010.8; 5k: Orc 1013.7, Human 1013.1. Now under the 1.0 note threshold at 20k; keep open for the race pass, close if it stays under 1.
- proposed fix: re-pick the Affliction race on the race pass (Orc), or keep Human and say it is within 1.5.
- needs: orchestrator fix
- resolved: 2026-11-02, b33ad307d. Affliction and its two curse duty builds are Orc: at 180 sec (20k, seed 1) Orc 1013.6, Human 1012.7 (still the best Alliance race), Undead 1010.8, Gnome 1009.9, Troll 1006.6; the duty builds 1003.7 -> 1005.2 and 997.4 -> 999.6.

### fairness-20261101-1: casters stand at 20, 25 or 30 yd with no standard; Elemental's 20 yd is worth +6.5 over the others' 25
- severity: medium (0.85% on Elemental; at 25 yd it would sit 0.6 above Smite instead of 7.1)
- affects: shaman__elemental-level60 (and -mana-tide) at 20 yd, druid__balance* at 30 yd; mages, warlocks, priests and the ranged hunters at 25 yd
- found: 2026-11-01, master dd053d24c
- evidence: requests' `distanceFromTarget` (mksettings spec_ele.go:38 = 20, spec_balance.go:39 = 30, the mage / warlock / priest / hunter specs 25 since 905787afd, which says only "like the mages"). Lightning Bolt, Wrath, Shadow Bolt, Frostbolt and Fireball all carry a MissileSpeed (sim/shaman/lightning_bolt.go:79 = 20), so distance is flight time. Paired, seed 1, 20k iterations, 180 +- 20 s (`python3 tools/audit/fairness/distance.py <cli> 25 <ids>`): Elemental 768.5 at 20 yd -> 762.0 at 25 (-6.5); Balance (Starfire) 846.6 -> 846.6 (0.0); Balance Wrath 806.5 -> 808.8 (+2.3). standards.md has no line on caster distance. Sure of the numbers; whether 20 yd is Elemental's real need (shock range) is not checked against the client.
- proposed fix: (a) one standard, every ranged spec at 25 yd: Elemental -6.5, Wrath +2.3, no rank change; (b) each ranged spec at the longest range its rotation needs (Elemental 20 for its shocks if the client says 20; others at 30 or their spells' range, re-measured); (c) keep, and write the per-spec distances and why into standards.md.
- needs: user ruling
- resolved: 2026-11-02, fa2d49997 (user ruling, standards 33: "Range the rotation needs"). Each caster stands at the longest range all its spells reach (orchestration/caster_distance.md); Elemental keeps 20 yd for its shocks.

### fairness-20261031-1: Demonology wears the Destruction / Affliction set and fails its own 5a target (crit 10.4 vs 14.1); its own mix sims +9.2
- severity: medium (0.9%; Demonology 1051.7 -> 1060.9 would sit 0.1 under Fire, so it stays 6th)
- affects: warlock__demonology and its duty-curse / duty-recklessness variants
- found: 2026-10-31, master d96adcce1
- evidence: `orchestration/gear_targets.md` ("Warlock: one set for Destruction, Affliction and Fire; Demonology keeps its own") and `orchestration/tools/gear_targets_v3.json` still give Demonology its own target (Sta 13.5, Int 15.5, SP 38.8, Crit 14.8), but 159443615 deleted `level60_phase1_synthetic_demonology` and every warlock request wears `level60_phase1_synthetic` (it already did before). `audit1b.py` on that set against the Demonology target: `5a FAIL Crit 10.4 vs 14.1` (hit-adjusted; Int 16.0 vs 14.7, SP 38.3 vs 36.9 within tolerance; 1b passes). Measured as bonus stats on the report request (the target mix minus the set's: +2.44% spell crit, -15 SP, -12 Int, -8 Sta, same linear budget), paired, seed 1, 20k iterations, 180 s: 1051.7 -> 1060.9 (+9.2, SE about 0.35). Crit alone at the cost of SP: 1044.4 (-7.3), so the gain is from the mix, not just crit. Script: $AUDIT_WORK/fairness/demo.py (bonus stats via tools/dps_report/race_override.stat_index). Sure of the 5a fail; the +9.2 is a bonus-stat proxy, not a built set.
- proposed fix: (a) rebuild a Demonology set at its own target (standards 1b/5a, as the mage Arcane set does): about +9 DPS, no rank change; or (b) rule that every warlock spec shares one set and drop the Demonology row from gear_targets.md / gear_targets_v3.json (then `audit1b.py` drops its Demonology line).
- needs: user ruling
- resolved: 2026-11-02, be5126ccd (user ruling, standards 34: "Own Demonology set"). Demonology wears level60_phase1_synthetic_demonology at its own target: 1045.5 -> 1052.7 at 180 s.

### fairness-20261101-3: Arcane's report race Gnome is 2.8 under Undead after the Arcane Blast hotfix
- severity: low (0.3%, no rank change)
- affects: mage__arcane
- found: 2026-11-01, master c23ef0440
- evidence: `ITERS=20000 python3 tools/audit/fairness/sweep.py mage__arcane` (seed 1, race overrides): Undead 896.4, Orc 894.1, Gnome 893.6 (the report race), Human 891.0. At 5k: Undead 895.9, Gnome 894.1. At dd053d2 they tied (911.4 / 911.3); f619fb648 (Arcane Blast 12% Int / 100% mana / 18 sec) moved Arcane -15.5 and Gnome more than Undead. Fairly sure (same order at 5k and 20k). The report hasn't taken in the hotfix yet (results pass 68), so this is about the coming re-sim.
- proposed fix: on the October 30 hotfix's report pass, re-pick Arcane's race (Undead), or keep Gnome and say it is within 3.
- needs: orchestrator fix
- resolved: 2026-11-02, 5a6e9c6a6 (the mage re-search on the hotfix kept Gnome; at 5k seed 1 Undead 939.5, Gnome 939.4, a tie)

### fairness-20261030-1: Frost's report race (High Order) is no longer its best; Undead sims 3.3 DPS higher after the 70707 talent re-pick
- severity: low (0.36%; Frost stays 17th either way)
- affects: mage__frost (report race SkyborneHighOrder)
- found: 2026-10-30, master fb1485888
- evidence: paired seed 1, 20k iterations, the report request with only the race changed (`ITERS=20000 python3 tools/audit/fairness/sweep.py mage__frost`): Undead 911.6, Gnome 909.9, Orc 909.3, High Order 908.3, Human 907.4, Troll 904.4. At d50909a High Order led (912.1 against Undead 911.8, 5k); cacd15f83 re-picked the mage talents on client 70707 and kept the race. Frost has no req/races/ override for any race. Best Alliance is Gnome (+1.6).
- proposed fix: set mage__frost's race to Undead (or Gnome if the report wants an Alliance pick) and re-sim it; or record why High Order stays.
- needs: orchestrator fix
- resolved: 2026-10-31, e1fcd57b7: mage__frost is Undead (request, meta and the UI preset); re-simmed.

### fairness-20261027-1: report races are stale after the 2026-10-26 racial pass (13 builds lose 1 to 12 DPS; Survival should pass Affliction)
- severity: high (BM +12.4 DPS, 1.1%; Survival's Orc moves it from #8 to #7, above Affliction)
- affects: hunters (BM, MM, Survival), Balance (both builds), Arcane, Enhancement, Arms, all rogue builds but Subtlety, the Ret Shock variants
- found: 2026-10-27, master 06a8987
- evidence: 500e3a147 applied the racial pass (Human 1% crit and Spirit x1.10, Orc Savage out and Command in, Dwarf 1% hit, Tauren crit damage out, Skyborne haste 2%); 4273ee51d re-simmed every build at its old race, but standards 16 says presets use the best race. Paired seeds 1 / 1000003, 20000 iterations each, only the race changed in the report request (best vs report race): BM Orc 1146.0 vs Human 1133.8 (+12.2); Survival Orc 1024.6 vs Human 1016.2 (+8.4, above Affliction Human 1018.3); MM Orc 1132.3 vs Human 1126.2 (+6.1); Balance Night Elf 850.2 vs Tauren 841.7 (+8.5); Balance Wrath High Order 810.9 vs Tauren 800.7 (+10.2); Seal Fate Swords Undead 726.0 vs Orc 722.0 (+4.0); Seal Fate Daggers Undead 729.4 vs Orc 725.2 (+4.2); Mutilate Undead 717.6 vs Windshaper 715.1 (+2.5); Combat Daggers Undead 767.5 vs Windshaper 765.4 (+2.1); Arcane Gnome 901.6 vs Human 899.2 (+2.4); Arms Undead 901.0 vs Orc 898.7 (+2.3); Ret Shock (Ret side) Undead 927.7 vs Human 925.6 (+2.1); Enhancement Windshaper 882.2 vs Orc 880.9 (+1.3). The 5000-iteration sweep (orchestration/audit/ref/fairness/sweep_baseline.log) adds, unconfirmed: Combat Swords Undead +1.2, Fire Gnome +1.1, Fury DW Human +1.0, Shock / Sub Holy Undead +1.3 to +2.3, Enhancement variants Windshaper +1.2 to +2.7. Orc Command reaches the pets (BM pet 333 vs 317 Undead, Demonology imp 85 vs 81). Reproduce: `tools/audit/fairness/sweep.py <id>`, or edit the race in a scratch copy of the request.
- proposed fix: rerun each class's race table at 120/180/300 sec (the racial pass changed every row), switch the presets above to their best race, regenerate the requests and re-publish. Read the rogue picks and the hunters' Night Elf / Orc split should be re-read from the new table, not from this list.
- needs: orchestrator fix
- resolved: 2026-10-28, master 5d27b0015 (race re-picks 86282541d, a22cb0295, b505cf238, eca80e9d9 / 67edf5d54, 466124303, 86d2ece1a / 621e2feb9, 12e348a85, 254a4d7d4, b57bd205c). Sweep with the standards 29 overrides: every report race is the best race; leftover gaps are ties (Fire Gnome +0.7, Affliction Orc +0.6, 20k paired seeds).

### fairness-20261025-1: Marksmanship's report race (Skyborne Windshaper) loses 2.5 DPS to the Human
- severity: low (0.2%; Marksmanship stays #2 either way, 1170.2 Human vs 1167.6 Windshaper, Lone Wolf 1159)
- affects: hunter Marksmanship (base and duty-mark builds); Lone Wolf checked and stays Windshaper
- found: 2026-10-25, master 54878fa3d04f
- evidence: tools/dps_report/req/hunter__marksmanship.json with only the race changed, 20000 iterations: Human 1170.2 / 1170.1, Windshaper 1167.6 / 1167.7 (seeds 1 / 1000003), +2.5. hunter__marksmanship-duty-mark: Human 1168.8 / 1168.0, Windshaper 1166.1 / 1165.6, +2.6. Lone Wolf: Windshaper 1159.3 / 1158.8 vs Human 1157.8 / 1158.3 (Windshaper +1.0, keep). The fairness sweep at 5000 iterations agrees (Human 1169.4, Windshaper 1168.0). The race table in forever-wiki/sim-status.md:318 (2026-10-23) had Windshaper ahead by 2.4 at 180 sec (1168.9 vs 1166.5), so a later hunter commit in the round (the mana-gated weave rotations or the finals) moved Human up about 3.7 and the race was not re-checked. Reproduce: the race sweep (`tools/audit/fairness/sweep.py hunter__marksmanship`) or edit the race in a scratch copy of the request.
- update 2026-10-25, master 016daa654fd5 (pass 5): still open. The duty-mark builds are gone and the base request now casts its own mark (e71b90f5), still Windshaper. 20000 iterations on the new request: Human 1168.8 / 1168.0, Windshaper 1166.1 / 1165.6 (seeds 1 / 1000003), +2.6 / +2.3. The 5000-iteration sweep: Human 1168.4, Windshaper 1166.7.
- proposed fix: re-run the hunter race sweep at 120/180/300 sec on the current Marksmanship presets; if Human still leads, switch the Marksmanship presets (and their duty builds) to Human, regenerate the requests and re-publish (+2.5 DPS).
- needs: orchestrator fix
- resolved: 2026-10-26, cdceb56c4 (Marksmanship plays a Human; report at 74b3a5f: Human 1139.6, best race, published 1139.4)

