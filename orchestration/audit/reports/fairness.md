# Fairness pass 18 (2026-10-31, master d96adcc)

Result: findings (1 new). Full pass: gear check, full race sweep, new baseline.

## Why it ran

Since e05fd73, 8 commits, 172 watched files: standard 32 (the raid's Sunder Armor ramps from the pull, pull
cooldowns don't wait for it), its request and preset edits (fedb73949), the re-sim 3c46da089 and report_hashes.
No sim code changed.

## What I checked

- Ramp applied evenly: 67 non-warrior requests carry `sunderRamp {tank: true, pullSunders: 1}`; the 7 warrior
  builds `{tank: true}` and their rotations cast "Sunder Armor once at the pull" (the DPS warrior's stack); the 3
  Sunder duty builds keep only their own. Casters carry it too and are unaffected.
- Race sweep (5k, seed 1, race overrides), against the e05fd73 baseline: every physical build drops at every race
  (Feral -4.0, Subtlety -3.4, hunters -2.2 to -3.1, warriors -2.3 to -2.9, rogues -1.9 to -2.6, Enhancement
  -0.7 to -1.6, Ret -0.3, warlock pets -0.1 to -0.2). Casters 0.0. Matches standard 32's measured cost; no leak.
- Best race against report race: best everywhere; ties within 0.6 (Fire Gnome, Arcane Undead, Affliction Orc).
- Published against current: `ui/dps_report/data/builds.json` matches the shared 20k run on all 77 (within 0.15)
  and this sweep within 0.8 (Affliction +1.6, the known offset). The 36 physical metas' `published.dps` are
  still pre-ramp (feral 889.7 vs 885.6): only the metas are stale, not builds.json. This corrects results'
  report this run, which says builds.json is pre-ramp too. Results' area (same pattern as results-20261030-1).
- Gear: `audit1b.py` 18 of 19 lines pass. The new Demonology line fails 5a: fairness-20261031-1.

## Finding

fairness-20261031-1 (medium, user ruling): Demonology wears the shared Destruction / Affliction / Fire set
since its own was deleted (159443615), but gear_targets.md still gives it its own target. Against it the set
fails 5a (crit 10.4 vs 14.1). As bonus stats (+2.44% crit, -15 SP, -12 Int, -8 Sta), paired 20k seed 1:
1051.7 -> 1060.9 (+9.2); it would stay 6th, 0.1 under Fire.

## Tool fixes (again)

- `tools/audit/fairness/sweep.py`: the race-override step was missing for the sixth time; re-applied (KEEP
  comment). The earlier five fixes never reached master: the run's commit should include `tools/audit/fairness/`.
- `tools/audit/fairness/audit1b.py`: target "Warlock (every spec)" isn't in gear_targets_v3.json (crash; same as
  pass 17, also lost); now "Warlock (Destruction / Affliction / Fire)" plus a Demonology line on the same set.

## Ranking

Order unchanged. Top BM Orc 1128.7, bottom Subtlety Undead 732.4: 1.54x. Full table in fairness_report.md §1
and ref/fairness/table_baseline.md. Baseline is now d96adcc.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, Flametongue Totem with the Windfury imbue, boss partial resists,
Rage k and pet scaling (see passes/fairness.md).
