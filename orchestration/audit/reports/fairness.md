# Fairness pass 16 (2026-10-30, master fb14858)

Result: 1 new finding (low). Full pass: gear check, imbue fit check, full race sweep, new baseline.

## Why it ran

Since bbbaf2e, 17 commits. Shared: f3017897f (Forever extra attacks keep the swing timer, Windfury Totem's 100 ms
proc cooldown, Eviscerate 4% AP, a stone or oil only on a weapon it fits; Combat rogues' off-hand mace takes a
Weightstone, Fury two-hand drops its off-hand stone). Mage: client 1.60.1.70707 talents re-picked (31054ead2,
cacd15f83) and Arcane's four Arcane Blast stacks at any mana (56695e4e5). Threat-only class changes (e2c54a743:
warrior shouts, druid Subtlety, Righteous Fury, Searing Pain, Rockbiter). Report re-sims cd1f6b2d0 and fb1485888.

## What I checked

- Gear: `audit1b.py` passes 5a and 1b on all 19 lines.
- Imbues against weapons (new rule in f3017897f): every report request's main- and off-hand stone or oil fits its
  weapon type (script `$AUDIT_WORK/fairness/imbues.py`, not kept). No spec silently loses a stone.
- Race sweep (5k, seed 1, race overrides applied), against the d50909a baseline:

| Moved | Change | Cause |
|---|---|---|
| Arms | +90.7 (+10.0%), 16th -> 12th | f3017897f extra attacks / WF Totem |
| Fury DW / 2H | +61.3 / +63.5 | same |
| Survival (+ pet variants) | +43.9 (+43) | same |
| Feral | +22.7 | same |
| Rogues (6) | +12.3 to +21.2 | same, plus Eviscerate 4% AP |
| Retribution (+ Shock / Sub Holy variants) | +16.8 (+15 to +30) | same (stored extra attacks) |
| Fire / Frostfire / Arcane / Frost | -28.2 / -39.4 / +13.8 / +0.3 | mage 70707 talents, Arcane stacks |

  Every other build at every race is unchanged to 0.0 (hunters ranged, warlocks, priests, Enhancement, Balance,
  Elemental). No shared change leaked. Enhancement's Windfury Weapon attacks are yellow attacks and already kept the
  swing timer, so it correctly does not move; every melee spec gets the extra-attack change, so it is even.
- Published against current: the sweep matches `ui/dps_report/data/builds.json` within 0.6 at every report race
  (Affliction +1.6, the known 5k seed-1 offset). The metas' `published.dps` are stale on 46 builds: independent
  confirmation of results-20261030-1 (not re-filed).
- Best race against report race: Frost Undead +3.3 over High Order at 20k (fairness-20261030-1). Ties: Fire Gnome
  +0.6, Arms Orc +0.3, Arcane (20k: Gnome 915.5, Undead 915.2), Affliction Orc +0.1.

## Note for the report's assumptions

The extra-attack swing-timer ruling is "assumed until a log settles it" and is now the largest single assumption for
warriors (+62 to +91). It is applied evenly, so no fairness finding, but it belongs among §3's assumption ranges.

## Tool fix (fourth time)

`tools/audit/fairness/sweep.py` had again lost the race-override step; re-applied (with a comment) in the working
tree for the run's commit. Without it Dwarf hunters, High Order mages and other override rows read 6 to 15 low.

## Ranking

Top BM Orc 1131.8, bottom Subtlety Undead 735.8: 1.54x (was 1.58x). Full table in fairness_report.md §1 and
ref/fairness/table_baseline.md. The baseline is now fb14858.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, Flametongue Totem with the Windfury imbue, boss partial resists,
Rage k and pet scaling (see passes/fairness.md).
