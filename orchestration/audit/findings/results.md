# Findings: Results and spec strength

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### results-20261025-1: mksettings and the committed requests disagree on consumes (potion, flask, filler explosive)
- severity: medium (Ret Shock -13.7 DPS = -1.6% if regenerated; others under 2.5)
- affects: paladin__level_60_shock* (defaultPotion MajorManaPotion vs DragonfirePotion), shaman__elemental-level60-mana-tide (vs CausticSmogPotion), all rogues (flask FlaskOfNaturalAccuracy vs FlaskOfNaturalPrecision on combat-daggers; fillerExplosive ExplosiveGoblinLandMine absent in requests), priest__smite and feral duty-ff (fillerExplosive)
- found: 2026-10-25, master 54878fa3d
- evidence: `python3 tools/audit/results/regen.py` vs `ref/results/drift_report.txt` (previous reference). Paired 20k seed 1, regen JSON vs committed: Ret Shock 855.3 vs 869.0; Elemental mana-tide 759.9 vs 762.1; Smite filler explosive part of an already-standing 4-field drift; rogue combat-daggers 768.7 vs 770.3, combat-swords 762.0 vs 764.0, mutilate 718.7 vs 718.9, seal-fate-swords 728.8 vs 728.9. Published numbers are unaffected (all 64 requests within 0.15 of published).
- proposed fix: make mksettings defaults match the standards 26 consumes the requests carry (or regenerate the requests); the Ret/Ele potion and rogue flask are the ones that move DPS.
- needs: orchestrator fix

## Resolved
