# Healer mechanics, pass 11 (2026-11-12)

Master 8306f1d52 against last pass a0b4d3be9. Result: quiet.

## Get current

| Source | Seen | Result |
|---|---|---|
| Newest client (wago.py latest) | 1.60.1.70888 | Same as pass 10, no verdict reopened |
| Blizzard notes | Nov 10 hotfixes (already seen) | Nothing newer |
| orchestration/update_ledger.md | no commits since a0b4d3be9 | Nothing new |

## What changed (sim/, a0b4d3be9..8306f1d52)

| Commit | Files | Healer impact |
|---|---|---|
| c380bbaef Druid: Rake's direct hit adds 6.5% AP | sim/druid/rake.go, feral tests/results | None. Cat form melee ability; the Restoration druid presets never cast Rake, and the change is limited to Rake's hit base damage |
| 368b2e181 Hunter pet melee speed | sim/core/pet.go, sim/core/unit.go, sim/hunter/* | None. Only pets with FollowsOwnerMeleeSpeed (the hunter pet); healers' pets (none) and health, heals, regen untouched |

Steps 3 to 5 (spells, engine, tests): no heal, healing talent, set bonus or healing engine file changed, so
the pass 10 tables and verdicts in ref/healer_mechanics/verdicts.md stand unchanged.

## Findings

No new findings. Open: healer_mechanics-20261104-11 (Riptide mana), healer_mechanics-20261104-10 (untested
heal paths); neither touched by these commits.
