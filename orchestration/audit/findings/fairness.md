# Findings: Fairness between specs

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### fairness-20261030-1: Frost's report race (High Order) is no longer its best; Undead sims 3.3 DPS higher after the 70707 talent re-pick
- severity: low (0.36%; Frost stays 17th either way)
- affects: mage__frost (report race SkyborneHighOrder)
- found: 2026-10-30, master fb1485888
- evidence: paired seed 1, 20k iterations, the report request with only the race changed (`ITERS=20000 python3 tools/audit/fairness/sweep.py mage__frost`): Undead 911.6, Gnome 909.9, Orc 909.3, High Order 908.3, Human 907.4, Troll 904.4. At d50909a High Order led (912.1 against Undead 911.8, 5k); cacd15f83 re-picked the mage talents on client 70707 and kept the race. Frost has no req/races/ override for any race. Best Alliance is Gnome (+1.6).
- proposed fix: set mage__frost's race to Undead (or Gnome if the report wants an Alliance pick) and re-sim it; or record why High Order stays.
- needs: orchestrator fix

## Resolved

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

