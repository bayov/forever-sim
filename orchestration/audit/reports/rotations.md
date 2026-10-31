# Rotations pass 42 (2026-10-31)

- **Master:** d96adcce1 (last audited e05fd732e). Client 1.60.1.70707.
- **Result:** quiet. No open findings; none new.

## What changed

- `apl_logic_diff.py e05fd732e origin/master`: 7 rogue APLs touched (combat_backstab,
  combat_sinister_strike, level60_combat_daggers, level60_combat_swords, level60_mutilate,
  level60_seal_fate_swords, level60_subtlety). All notes-only: the Gnome Eureka! line's note now says
  it lasts 8 sec (10% more damage) instead of "the next 3 abilities". Conditions unchanged.
- `tools/rotopt/rogue_common.go`: the same note text in the template; no knob or condition change.
- `tools/rotopt/mksettings/main.go` and `tools/dps_report/req/*.json` (152 files): the raid's Sunder
  Armor becomes a ramp from the pull (`sunderRamp: {tank: true, pullSunders: 1}` in place of
  `sunderArmor: true`), standard 32. Its rule that pull cooldowns don't wait for the ramp is a ruling,
  not re-flagged. Rotations embedded in the reqs are unchanged.
- No sim `.go` changed. Baseline (`baseline_e05fd73_3000it_seed424242.tsv`) not moved.

## Optional

- Carried from earlier passes (see the pass doc's open items); nothing re-measured this pass.
