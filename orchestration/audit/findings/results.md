# Findings: Results and spec strength

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### results-20261027-1: all 58 metas' published.dps are stale after the 2026-10-26 racial-pass re-sim
- severity: low (no DPS number is wrong: builds.json matches my 20k seed-1 run on all 58 within 0.1)
- affects: tools/dps_report/req/*.meta.json (all 58); runreq.sh / verify.py compare against them
- found: 2026-10-27, master 06a898751
- evidence: 4273ee51d re-simmed every build into ui/dps_report/data/builds.json but left `published.dps` in the metas at the 10-25 values (warrior__fury meta 986.6, builds.json 975.2, run 975.19; mage__arcane 982.5 vs 899.1, -8.5%; Ret Shock metas were updated only by 3d9aa585b). Every request is "OFF published" by 0.15 to 83 DPS until they are updated.
- proposed fix: set each meta's `published.dps` (and `source`) from builds.json, as tools/dps_report/README.md "update meta.published" says.
- update 2026-10-28, master 5d27b0015: still open after post 8's re-sim (73e31ec4c, 9857d17e3). 34 of 58 metas differ from builds.json by more than 0.15 (mage__frost 946.6 vs 911.1, warlock__affliction 1035.3 vs 1018.1); my 20k seed-1 run matches builds.json on all 58 within 0.1.
- needs: orchestrator fix

### results-20261025-1: mksettings and the committed requests disagree on consumes (potion, flask, filler explosive)
- severity: medium (Ret Shock -13.7 DPS = -1.6% if regenerated; others under 2.5)
- affects: paladin__level_60_shock* (defaultPotion MajorManaPotion vs DragonfirePotion), shaman__elemental-level60-mana-tide (vs CausticSmogPotion), all rogues (flask FlaskOfNaturalAccuracy vs FlaskOfNaturalPrecision on combat-daggers; fillerExplosive ExplosiveGoblinLandMine absent in requests), priest__smite and feral duty-ff (fillerExplosive)
- found: 2026-10-25, master 54878fa3d
- evidence: `python3 tools/audit/results/regen.py` vs `ref/results/drift_report.txt` (previous reference). Paired 20k seed 1, regen JSON vs committed: Ret Shock 855.3 vs 869.0; Elemental mana-tide 759.9 vs 762.1; Smite filler explosive part of an already-standing 4-field drift; rogue combat-daggers 768.7 vs 770.3, combat-swords 762.0 vs 764.0, mutilate 718.7 vs 718.9, seal-fate-swords 728.8 vs 728.9. Published numbers are unaffected (all 64 requests within 0.15 of published).
- proposed fix: make mksettings defaults match the standards 26 consumes the requests carry (or regenerate the requests); the Ret/Ele potion and rogue flask are the ones that move DPS.
- needs: orchestrator fix

## Resolved
