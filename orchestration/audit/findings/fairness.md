# Findings: Fairness between specs

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### fairness-20261112-1: Combat Daggers is now the rogue Combat default, but Combat Swords sims 10 DPS higher at master
- severity: medium (the default build shown for Combat; no change to the cross-class ranking of the best rogue)
- affects: rogue__combat-daggers / rogue__combat-swords metas (isDefault), ui/rogue presets and sim.ts default, forever-wiki rogue page
- found: 2026-11-12, master a0b4d3be9
- evidence: 7462c17c8 (user ruling 2026-11-12, "switch to best") set Combat Daggers as default because it was the best 180 s Combat build then (707ae92 sweep: Daggers 915.4, Swords 899.6). The next commit that moved rogues, a0b4d3be9 (Windfury Totem on main hand abilities) plus e2322baa1 (Venom on Instant Poison), lifted Swords +46.5 and Daggers +21.2. At master, report race Undead, 180 s requests: Swords 946.1 vs Daggers 936.1 (shared run `$AUDIT_WORK/runs/master`, and the 5k seed 1 sweep, 946.1 vs 936.0; noise about 1.5). Daggers' best race (Gnome 936.6) does not close it.
- proposed fix: apply the ruling again on the current numbers: isDefault back to rogue__combat-swords, with its UI preset defaults and the wiki line, once results-20261112-1's re-sim confirms the order at 20k. Fury DW over 2H (1050.2 vs 1032.0), Affliction over Destruction (1071.1 vs 1063.2) and Lone Wolf over BM still hold.
- needs: orchestrator fix
- 2026-11-12, master 8306f1d52: still open after the re-sim (bff0169e4). builds.json at 20k: Combat Swords 946.4, Combat Daggers 936.2 (isDefault); 5k sweep 946.1 / 936.0. Order confirmed.

## Resolved

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

