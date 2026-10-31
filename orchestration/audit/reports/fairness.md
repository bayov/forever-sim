# Fairness pass 17 (2026-10-31, master e05fd73)

Result: quiet. Full pass: gear check, full race sweep, new baseline.

## Why it ran

Since a2a3e66, 40 commits, 164 watched files. Sim-relevant: e1fcd57b7 (the 2026-10-30 audit fixes: Ret twisting
without the 1.5 sec window, Hamstring from 10 Rage, Frost Undead, Efficiency on Strider Kick), its re-sim 3d4cdda4a,
the threat check in sim/core (e05fd732e, threat only), and the preset cleanup (159443615, 39906cdaa: Other presets
and old gear sets deleted, report requests unchanged). standards.md changes are the 2026-10-30 rulings already
swept last pass (WF Totem 100 ms, extra attacks, Eviscerate 4% AP).

## What I checked

- Gear: `audit1b.py` passes 5a and 1b on all 18 lines. Tool fix: 159443615 renamed the warlock target to
  "Warlock (every spec)", which isn't in `gear_targets_v3.json`, so the script crashed. Now uses
  "Warlock (Destruction / Affliction / Fire)". The deleted Demonology gear set was unused: every warlock request
  already wore the shared set (same items at a2a3e66 and fb14858).
- Race sweep (5k, seed 1, race overrides applied), against the fb14858 baseline:

| Moved | Change | Cause |
|---|---|---|
| Retribution | +14.4 (960.9 -> 975.3) | e1fcd57b7 twisting, no swing window |
| Sub Holy Ret variant | +12.6 | same |
| Arms | +4.0 | e1fcd57b7 Hamstring from 10 Rage |
| Fury two-hand | +2.9 | same |

  Every other build at every race is unchanged to 0.0. Survival correctly doesn't move from Efficiency on Strider
  Kick (its build has no Efficiency). Shock Ret variants unchanged. No shared change leaked.
- Published against current: every report race within 1.0 of `meta.published.dps` except Affliction +1.6 (the
  known 5k seed-1 offset); agrees with the loop's vs_published (0 of 77 moved). The metas are current again, so
  results-20261030-1's stale metas are fixed on master.
- Best race against report race: best everywhere. Ties: Fire Gnome +0.6, Arcane Undead +0.2, Affliction Orc +0.1.
  Arms now sims best as Undead, its report race.

## Tool fix (fifth time)

`tools/audit/fairness/sweep.py` had lost the race-override step again; re-applied in the working tree with a
KEEP comment, for the run's commit.

## Ranking

Top BM Orc 1131.8, bottom Subtlety Undead 735.8: 1.54x (unchanged). Full table in fairness_report.md §1 and
ref/fairness/table_baseline.md. The baseline is now e05fd73.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, Flametongue Totem with the Windfury imbue, boss partial resists,
Rage k and pet scaling (see passes/fairness.md).
