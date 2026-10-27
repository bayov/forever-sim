# Fairness pass 9 (2026-10-27, master 99bd53a)

Result: quiet (no new findings). Targeted pass: one new variant, no sim code changed.

## Why it ran

Watched files changed since f62cb79: the new preset Level 60 Shock (Ret side) (51442388b, 878000154):
`tools/dps_report/req/paladin__level_60_shock_ret{,.meta}.json`, `ui/retribution_paladin/presets.ts`,
`ui/retribution_paladin/apls/level60_shock_ret.apl.json`. Nothing under sim/, proto/, assets/database or
standards.md changed, so the other builds can't have moved; only the paladin requests were re-swept.

## What I checked

- Gear: `audit1b.py`, all 19 sets "5a PASS | 1b PASS". The new build uses the shared `level60_phase1_synthetic` Ret set.
- Shared assumptions: the new request, normalised without talents and rotation, is byte-identical to
  `paladin__level_60.json` (buffs, debuffs, raid hits, consumes, gear, Dragonfire Potion, Seal of Command).
  Against `paladin__level_60_shock` it differs only in defaultPotion (Dragonfire vs Major Mana; the meta notes
  Dragonfire is 5.7 better at 180 s, so each build takes its own best potion) and primarySeal.
- Race sweep (5000 it, seed 1): Shock (Ret side) Human 932.3, Undead 928.5, Dwarf 925.3. Report race Human is best.
- Published against current: builds.json 932.4, meta 932.4, shared run 932.4, sweep 932.3. Current.
- Other paladin builds re-swept: level_60 947.0, sub_holy 949.8, shock 869.2, identical to the f62cb79 baseline.
- The shared run's `vs_published.txt` gaps (Affliction -5.9 etc.) are the stale req metas noted in pass 8; builds.json is current.

| Variant | Best race | DPS | Published | Report race |
|---|---|---|---|---|
| paladin__level_60_shock_ret | Human | 932.3 | 932.4 | Human |

## Conclusions

- The new variant is simmed under the same conditions as every other Ret build; it sits 1.6% under Level 60 Ret
  (947.0), below the base build, so the cross-spec ranking (Ret #16) is unchanged.
- Baselines: added the Shock (Ret side) line to `ref/fairness/sweep_baseline.log` and `table_baseline.md`; every
  other line stays valid (same sim code). `orchestration/fairness_report.md` §1 got an addendum and the variant row.
- results-20261025-1 (mksettings vs request consumes) is the results auditor's; the new meta's notes already say
  primarySeal was set by hand to match the shipped preset, consistent with that finding. Nothing to add.
