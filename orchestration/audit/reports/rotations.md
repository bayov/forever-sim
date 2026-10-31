# Rotations pass 41 (2026-10-31)

- **Master:** e05fd732e (last audited a2a3e66e1). Client 1.60.1.70707.
- **Result:** quiet. No new findings; all three open findings landed on master (e1fcd57b7) as written
  and were already moved to Resolved.

## What changed

- `apl_logic_diff.py a2a3e66e1 origin/master`: 30 files. 24 deleted (old Era presets removed, 39906cdaa;
  not report builds). The Level 60 logic edits are exactly the pass 40 fixes:
  - Ret twist lines lose the `autoTimeToNext < 1.5s` gate in level60, level60_sub_holy,
    level60_crusader_duty, level60_wisdom_duty (rotations-20261030-1). level60_undead keeps 1.5 s
    (the worker measured the free window losing there, 64faff678).
  - Hamstring 7373 r3 from 10 Rage in the four Arms and three Fury 2H APLs (rotations-20261030-2).
  - Enhancement leveling plans changed (leveling only; not reached by raid presets).
- Sim `.go` the presets reach:
  - Efficiency now covers Strider Kick (client 70707 mask). Mana-cost only; no survival preset moved >0.5.
  - `threat_check.go` and its hooks (rage.go, spell_result.go, unit.go, metrics): observational, active
    only with a raid threat check, uses no RNG. compsearch/resolve.go is compsearch only.

## Baseline (3000 it, seed 424242, shared CLI e05fd732e) vs fb14858

| preset | old | new | diff |
|---|---|---|---|
| paladin__level_60 | 961.0 | 975.8 | +14.8 |
| paladin__level_60-duty-crusader / -wisdom | 954.4 / 947.3 | 967.3 / 960.3 | +12.9 / +13.0 |
| paladin__level_60_sub_holy (+ crusader / wisdom) | 965.0 | 977.8 | +12.9 (+14.5 / +14.2) |
| warrior__arms / -battle-shout / -sunder / dragonkin | 995.5 / 993.6 / 951.9 / 1039.3 | 1000.2 / 997.6 / 954.9 / 1040.9 | +4.7 / +4.0 / +3.0 / +1.6 |
| warrior__fury_2h / -battle-shout / -sunder | 1027.3 / 1026.1 / 994.6 | 1030.8 / 1029.1 / 996.6 | +3.5 / +3.0 / +2.0 |
| mage__frost | 909.4 | 911.6 | +2.2 (Frost Undead req change, e1fcd57b7) |

Every other row within 0.5. All moves match the applied fixes and their measured gains, so
`regress.py` was not needed. The loop's shared run: 0 of 77 moved >0.5 vs published.

## Open items

- The pass doc's settled item "Paladin twist: 1.5 s before the swing is a sharp optimum" is now stale
  (true only for sub_holy_undead). The orchestrator should reword it in `passes/rotations.md`; this
  pass doesn't own that doc.
- Carried over, untouched: Balance Night Elf vs Tauren, Elemental Mastery holding, warlock duty-curse
  screens, and the optional list in the pass doc.
