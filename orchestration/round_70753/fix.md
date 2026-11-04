# Round 70753, phase 3: fix round (`briefs/phase3_fix.md`)

## 1. Client 1.60.1.70755 (d57aa24, 47dab06)
- All 39 client_update tables and the talent importer's tables are byte-identical to 70753. db2diff finds 0 spells
  changed, and import_db2 finds 0 changes for every class. Details are in `client_70755.md`.
- No notes-only value is in 70755, so every list and gate stays as it was.
- The rogue, warrior and shaman reworks are not in 70755, so those classes stay on `notes/<class>.json`. Every preset
  is still legal.
- Committed: build restamps of the inputs and `tools/forever_talents/client/*.json` regenerated. No item changed.
- Builds moved: none. Tests: `go test --tags=with_db ./sim/...` passes. sim/web is skipped (needs binary_dist), and
  compsearch passes with a 40-minute timeout.
- `state.json` `foreverClientBuild` still says 70753. Updating it is the orchestrator's job.

## 2. Gating (13c34f2; findings d7bd147, af2c445)
Outside Forever, each value now takes Classic's. Where Classic doesn't have the mechanic, it keeps the client's
pre-notes value.
- **Rogue:**
  - Classic values for Lethality, Relentless Strikes, Seal Fate, Dual Wield Specialization, Aggression and Improved
    Eviscerate.
  - Improved Sinister Strike's cut to Hemorrhage's cost, Shank, the Ghostly Strike opener and Hemorrhage on Garrote
    are now Forever only.
  - Mutilate keeps 0.75 and Flawless Execution 10 outside Forever (neither is in Classic).
- **Hunter:** Classic ranks for Arcane Shot, Serpent Sting and Aimed Shot, with no AP share. Sniper Shot keeps
  160/225/295.
- **Paladin:** Classic Vengeance.
- **Shaman:** Era Lightning Bolt ranks 4-10, Classic Chain Lightning, and Flurry at 10-30%.
- **Warlock (found in the sweep):** demon health stays at 7.5 per Stamina point outside Forever.
- Forever DPS: unchanged. All 68 computed report hashes are identical with and without this commit. Classic goldens
  moved: none, because every `.results` golden runs under Forever.
- Tests: all `sim/...` pass (except sim/web), compsearch included.

## 3. Tank threat (0dcb0d8; finding 0a367fa)
Under Forever:
- Bear and Dire Bear Form x2.15 (notes: from 50% to 115%).
- Defensive Stance x1.8.
- Defiance 7/13/20%. The stance and Defiance multiply, as two separate client auras, so 3/3 with a shield is x2.16.

Classic keeps x1.5, x1.3 and 5% a point. The values are listed in druid.md and warrior.md under "Notes-only values".
- New tests: TestForeverBearFormThreat and TestForeverDefensiveStanceThreat.
- Golden moved: TestP1TankWarrior.results, its TPS lines only, each x1.445. No bear golden records threat.
- DPS moved: none.

## 4. Enhancement leveling weights (f6ca172)
`ui/enhancement_shaman/leveling_weights.json` was re-measured on the new tree, at levels 5 (no spec), 10 ... 55 and 59.
- Mobs per 20-minute session:

  | Level | Before | After |
  |---|---|---|
  | 5 | 70.4 | 70.2 |
  | 10 | 58.4 | 58.8 |
  | 30 | 61.5 | 74.6 |
  | 50 | 56.3 | 72.2 |
  | 59 | 52.1 | 65.9 |

- Intellect, Spirit and MP5 weights are up.
- The rotation still works on the new tree, and the template's knobs are untouched.
- This was one round, not the documented mean of three.
- To do: bayovui's `weights.py` (wow repo).

## Open
- **TestRaidForeverReportHashesStable fails on master.** All 68 builds show as moved. It already failed on 0a988c9 (the
  report session's re-sim), before 13c34f2. `sim/testdata/report_hashes.txt` needs regenerating from the report's
  final requests.
- **For a user ruling:**
  - Outside Forever, Forever-only mechanics fall back to the client's pre-notes value.
  - Older ungated beta values remain in the hunter and shaman files: cast times, coefficients, Lightning Bolt mana.
  - Should the stance and Defiance multiply (x2.16, as now) or add (x2.0)?
  - Is one round of leveling weights enough, or should it be the mean of three?
