# Findings: Results and spec strength

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

## Resolved

### results-20261025-1: mksettings and the committed requests disagree on consumes (potion, flask, filler explosive)
- severity: medium (Ret Shock -13.7 DPS = -1.6% if regenerated; others under 2.5)
- affects: paladin__level_60_shock* (defaultPotion MajorManaPotion vs DragonfirePotion), shaman__elemental-level60-mana-tide (vs CausticSmogPotion), all rogues (flask FlaskOfNaturalAccuracy vs FlaskOfNaturalPrecision on combat-daggers; fillerExplosive ExplosiveGoblinLandMine absent in requests), priest__smite and feral duty-ff (fillerExplosive)
- found: 2026-10-25, master 54878fa3d
- evidence: `python3 tools/audit/results/regen.py` vs `ref/results/drift_report.txt` (previous reference). Paired 20k seed 1, regen JSON vs committed: Ret Shock 855.3 vs 869.0; Elemental mana-tide 759.9 vs 762.1; Smite filler explosive part of an already-standing 4-field drift; rogue combat-daggers 768.7 vs 770.3, combat-swords 762.0 vs 764.0, mutilate 718.7 vs 718.9, seal-fate-swords 728.8 vs 728.9. Published numbers are unaffected (all 64 requests within 0.15 of published).
- proposed fix: make mksettings defaults match the standards 26 consumes the requests carry (or regenerate the requests); the Ret/Ele potion and rogue flask are the ones that move DPS.
- needs: orchestrator fix
- resolved: 2026-10-31 (user approved the fix): the potion and flask lines were regen.py's spec mapping, not mksettings (it ran every paladin as `ret` and every Combat rogue as `rogue_combat`; mksettings has `ret_shock` with the Major Mana Potion and `rogue_combat_daggers` with Natural Accuracy). regen.py now takes a variant's spec from its base build, then `-spec` from the meta's notes, then the id. The real drift is gone: mksettings has `rogue_combat_swords` and `rogue_combat_daggers` without the Goblin Land Mine, `rogue_assassination` without it, `ele_mana_tide` with the Major Mana Potion, and `shadow_holy` and `smite` with their own consumes, so a regenerated request equals the committed one. Left by design: the feral Faerie Fire duty preset drops the base's EZ-Thro Dark Bomb, a duty edit its notes record, like every duty preset's debuff toggles. ref/results/drift_report.txt is regenerated. No request changed, so no published number moves.

### results-20261030-1: 46 of 77 metas' published.dps are stale after the 2026-10-30 re-sim (cd1f6b2d0, fb1485888)
- severity: low (no published number is wrong: builds.json matches my 20k seed-1 run on all 77 within 0.15)
- affects: tools/dps_report/req/*.meta.json (46); runreq.sh and verify.py report them as OFF published
- found: 2026-10-30, master fb1485888
- evidence: the re-sim commits update ui/dps_report/data/builds.json but not the metas (warrior__arms meta 905.6, builds.json 994.9, run 994.9; hunter__survival 1028.9 vs 1072.6; mage__fire 1089.0 vs 1061.0). Same pattern as results-20261027-1.
- proposed fix: set each meta's published.dps and source from builds.json, per tools/dps_report/README.md "update meta.published".
- needs: orchestrator fix
- resolved: 2026-10-31: every meta's published.dps and source set from builds.json after the re-sim, in the same commit as the data.

### results-20261027-1: all 58 metas' published.dps are stale after the 2026-10-26 racial-pass re-sim
- severity: low (no DPS number is wrong: builds.json matches my 20k seed-1 run on all 58 within 0.1)
- affects: tools/dps_report/req/*.meta.json (all 58); runreq.sh / verify.py compare against them
- found: 2026-10-27, master 06a898751
- evidence: 4273ee51d re-simmed every build into ui/dps_report/data/builds.json but left `published.dps` in the metas at the 10-25 values (warrior__fury meta 986.6, builds.json 975.2, run 975.19; mage__arcane 982.5 vs 899.1, -8.5%; Ret Shock metas were updated only by 3d9aa585b). Every request is "OFF published" by 0.15 to 83 DPS until they are updated.
- proposed fix: set each meta's `published.dps` (and `source`) from builds.json, as tools/dps_report/README.md "update meta.published" says.
- update 2026-10-28, master 5d27b0015: still open after post 8's re-sim (73e31ec4c, 9857d17e3). 34 of 58 metas differ from builds.json by more than 0.15 (mage__frost 946.6 vs 911.1, warlock__affliction 1035.3 vs 1018.1); my 20k seed-1 run matches builds.json on all 58 within 0.1.
- update 2026-10-28, master 224b8a4e0: still 34 of 58 after the racial-audit deploy.
- needs: orchestrator fix
- resolved: 2026-10-28, every req/*.meta.json published.dps set from the 2026-10-28 full re-sim's builds.json.

