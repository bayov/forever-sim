# Consumables pass 91 (master c198d81e3)

Result: quiet, no new findings.

## Moved watched files (2f7e255af..c198d81e3)
| File | Change | Verdict |
|---|---|---|
| tools/dps_report/assumptions_def.py | b77ffa52a removes the dead warr_exec_wf2 assumption (7 lines) | Not a consumable or buff value; nothing to check |

## Checks
- Presets, mksettings, sim/core consumes/buffs/debuffs and forever_enchants.go: no diff since the last audited master.
- wago latest build is 1.60.1.70959, same as the last pass; no builddiff needed.
- Class branches show as ahead only through stale divergence; not audited.
- Note: `oil_proc_damage` is no longer found in tools/dps_report/*.py (the pass doc names it); the doc watch item may be stale.

No open items beyond the two low findings in findings/consumables.md.
