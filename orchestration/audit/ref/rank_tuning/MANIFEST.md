# Rank tuning audit: reference data

## In this directory

- `baseline_d2633f4.csv`: the last full sim baseline, made on master d2633f4 (client 70450) at 20k iterations and 180 s with `tools/audit/rank_tuning/rt_sensitivity.patch` applied.
  - It has one row per preset and tag: `base` (unpatched behaviour) and each `jobs.txt` sensitivity tag, with mean DPS and stdev.
  - Compare a new run against it with `python3 tools/audit/rank_tuning/summarize.py $AUDIT_WORK --compare orchestration/audit/ref/rank_tuning/baseline_d2633f4.csv`.
  - Seeds are fixed, so an unchanged tree reproduces these numbers exactly. For example, a fresh build on b389688 gave mage__frost 843.3 / fb 855.6.

## Kept only locally (re-create, don't commit)

**Built binary:** `$AUDIT_WORK/cli_master`. Re-create it with `AUDIT_TREE=<scratch checkout of origin/master> AUDIT_WORK=<dir> tools/audit/rank_tuning/build.sh`.

**20k sim inputs and outputs:** `$AUDIT_WORK/m2_<preset>.in.json` and `m2_<preset>_<tag>.out.json`, about 60 KB each, ~150 files. Re-create them with `tools/audit/rank_tuning/run_all.sh` (all presets) or `run_all.sh <preset> ...`.

**Client db2 tables:** the shared cache `tools/sources/wago.py` keeps, from wago.tools. The tracked tables are its `rank_tuning` set: CurvePoint, SkillLineAbility, SpellCastTimes, SpellCooldowns, SpellDuration, SpellEffect, SpellItemEnchantment, SpellLevels, SpellMisc, SpellName, SpellPower, TraitCond, TraitDefinition, TraitDefinitionEffectPoints, TraitNode, TraitNodeEntry, TraitNodeGroupXTraitCond, TraitNodeGroupXTraitNode and TraitNodeXTraitNodeEntry.
- The baseline is build 1.60.1.70450 (70425 is identical apart from cosmetic additions).
- Re-create it with `python3 tools/sources/wago.py fetch 1.60.1.70450 --set rank_tuning`.
- For a new build N, run `wago.py fetch 1.60.1.N --set rank_tuning`, then `python3 -I tools/sources/wago.py diff 70450 N --set rank_tuning`.

**Older builds and tables, not needed for the recurring pass:**
- The first-pass analysis also kept 69876–70334 SpellEffect/SpellLevels/SpellName, Era 1.15.9.70003 (E_*) and 70338 (F_*, N_*) copies, plus one-off ratio and rank-inversion scripts.
- Their method and results are written up in `orchestration/rank_tuning.md`.
- Re-download any build with `wago.py fetch <build> <Table>` (Era: build 1.15.9.70003).
