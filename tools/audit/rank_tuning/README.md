# Rank tuning audit tools

Each script starts with a usage line. The pass that uses them is in `orchestration/audit/passes/rank_tuning.md`.

**Environment:**
- `AUDIT_TREE`: a throwaway checkout of origin/master to patch and sim.
- `AUDIT_WORK`: scratch space for the binary, the 20k inputs and outputs, and the db2 tables.

**Data files:**
- `rt_sensitivity.patch`: the env-var hooks (`RT_BANE_COEF`, `RT_SMITE_MULT`, `RT_RIP_MULT`, `RT_WRATH_MULT`, `RT_FB_MULT`, `RT_HS_MULT`, `RT_CONS_MULT`, `RT_LB_MULT`, `RT_LVB_MULT`, `RT_ES_MULT`). `build.sh` applies it to `AUDIT_TREE`. Never commit it to the sim.
- `jobs.txt`: one sensitivity per line, as `<preset> <tag> VAR=value...`, the arguments to `run2.sh`.
- `allreq.txt`: the 66 preset names at the baseline. `run_all.sh` regenerates it from the req dir.

**Scripts:**
- `build.sh`: apply the patch, regenerate the protos, build `cli_master`.
- `run2.sh`: sim one preset at 20k iterations with optional env vars.
- `run_all.sh`: sim every preset's base plus all sensitivities, or only the named presets.
- `summarize.py`: print each sensitivity with its place, write or compare the reduced baseline CSV.
- `wago_builds.sh`: list the Forever client builds on wago.tools.
- `db2_fetch.sh`: download the 19 tracked db2 tables for a build.
- `db2_rowdiff.py`: row-diff every table between two builds.
