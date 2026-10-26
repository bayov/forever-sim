# Consumables pass 43 (quiet)

Master f796d99536 (previous audited e2316be). Client 1.60.1.70532 is still the latest (wago probe).

## Findings
None.

## Checked
- Watched file moved: tools/dps_report/assumptions_def.py. The diff is wording only: user-ruling mentions dropped
  from src/desc/rationale text, the natural_flask_zone label reworded (same values and the no_flask_zone edit), and a
  new optional `warn` field with GCD_ROT_WARN. oil_proc_damage, hunter_tsa_hm_rank values and other buff settings are unchanged.
- No presets.ts, mksettings, sim/core buff files, forever_enchants.go or level60 gear changed (only a request meta.json).

## Open items
None.
