# Fairness pass 21 (2026-11-02, master 5a6e9c6)

Result: quiet (no new finding; fairness-20261101-3 resolved). Full pass: gear check, full race sweep, new baseline.

## Why it ran

Since c23ef04, 99 watched files: the launch realms' 2026-10-30 hotfix trees over client 1.60.1.70707 (paladin, druid,
mage, priest, warlock: 0e7c70c86, 2b2d19684; hunter: d2f12b8f0), the druid sim and presets (175e3fdb3, 4176687d3), the
paladin rules and Ret presets (1c83bc05c, 306f78678), the mage re-search (5a6e9c6a6).

## What I checked

- Gear: `audit1b.py` crashed on "Warlock (every spec)" a third time (the fix again never reached master). Re-fixed in
  `tools/audit/fairness/audit1b.py` (one line per warlock target). 19 of 20 pass; Demonology fails 5a (crit 10.4 vs
  14.1), the open fairness-20261031-1. **The run should commit `tools/audit/fairness/audit1b.py`.**
- Race sweep (5k, seed 1) against c23ef04. Every move has its commit; every other build is bit-equal:

| Builds | Move | Cause |
|---|---|---|
| Ret twist / sub-Holy | +64.2 / +59.9 | 306f78678 Ret re-search on the hotfix tree (Unsealed Wrath twisting, Conviction 2% all crit) |
| Ret Shock / deep-Holy Shock | +13.7 / -52.3 | same commit; the deep-Holy build loses Light's Vigil damage (heals only) |
| Feral | +56.3 | 4176687d3: Genesis / Naturalist 2% a point (+55.3), Predatory Discipline (+27.5), Berserk gone (-18.1) |
| Balance, Wrath | +44.5, +49.2 | 4176687d3: Genesis / Naturalist |
| Arcane, Frost | +43.6, +15.2 | 5a6e9c6a6 mage re-search on the hotfix and the Arcane Blast change |
| hunters, warlocks, priests, warriors, rogues, shamans, Fire, Frostfire | +0.0 | hunter tree swap is a string remap; priest / warlock text only: no leak |

- Shared assumptions on the changed requests (Ret, Feral, Balance, mages) against Fury: same raid buffs, debuffs and
  Sunder ramp (`pullSunders: 1` for non-warriors per standards 32), consumables by role. Nothing uneven is new.
  Predatory Discipline's literal reading is listed in spec_assumptions.md (`druid_predatory_discipline`, one value).
- Published against current: the metas are stale for 62 of 77 (results-20261101-1, open; I confirm independently).
  builds.json still carries the pre-hotfix paladin, druid and mage numbers and the old two-hand warrior numbers: the
  hotfix's report re-sim hasn't run ("The DPS report data isn't regenerated" in each commit). Until it does, the
  published ranking is off for Ret (about 975 -> 1039.2, 14th -> 9th), Feral (886 -> 942), Balance (846 -> 891),
  Arcane (896 -> 939) and Frost (929 -> 944). Not filed (the results auditor owns staleness).
- Best race against report race: best everywhere except Affliction (Orc +1.3, fairness-20261101-2, open) and Fire
  (Gnome +0.4, tie). Arcane Undead +0.1 over Gnome is now a tie: fairness-20261101-3 resolved.

## Ranking now (report race, 5k)

Lone Wolf 1087.8; Frostfire 1075.6; MM 1063.7; Fire 1056.7; Demonology 1046.5; BM 1045.2; Fury DW 1041.6; Fury 2H
1039.6; Ret 1039.2; Survival 1024.8; Affliction 1012.9; Arms 1008.7; Destruction 995.0; Destruction Fire 977.8; Shadow
948.7; Frost 944.0; Feral 942.4; Arcane 939.4; Balance 890.6; Enhancement 883.0; rogues 788.4 to 732.4, Elemental 769.0,
Smite 761.2. Top over bottom 1.49x (unchanged: the top and bottom didn't move).

## Findings

- None new. fairness-20261101-3 (Arcane race) resolved by 5a6e9c6a6.

## Open items carried

- fairness-20261101-1 (caster distance), -2 (Affliction race), 20261031-1 (Demonology set): no change on master.
- Ice Lance Low range, Flametongue Totem stacking, boss partial resist, Rage k / pet scaling: unchanged.
- Warriors, rogues and shamans have no hotfix rows at 70707; if a later hotfix push touches them, re-sweep.

## Baseline

`ref/fairness/sweep_baseline.log` and `table_baseline.md` replaced with this sweep (5a6e9c6).
