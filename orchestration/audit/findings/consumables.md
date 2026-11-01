# Findings: Consumables, buffs and debuffs

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### consumables-20261101-1: Ranged hunters keep the Goblin Sapper Charge, worth +0.2 now, comments say +4.8
- severity: low
- affects: ui/hunter/presets.ts (DefaultConsumes, comments), tools/rotopt/mksettings/spec_hunter.go (hunter_bm, hunter_mm, hunter_lw)
- found: 2026-11-01, master dd053d24c
- evidence: the new 10 yd reach (sim/core/consumes.go explosiveBlastRadius, commit bb21e3b32) gates the Sapper on the static DistanceFromTarget; ranged hunters stand at 25 yd and only get inside 10 yd during a melee weave. tools/dps_report/req/hunter__beast-mastery.json at 180 s: 300 iterations cast the Sapper 98 times (about a third of fights); 20000 iterations with vs without it: 1125.55 vs 1125.40 (+0.16, same seed). The comments (spec_hunter.go, presets.ts) still say Sapper +4.8 / "goes off in a weave". Dark Bomb checked: 1125.55 vs 1122.42 with none (+3.1), matches the claimed +2.8 to +3.9.
- proposed fix: drop SapperExplosive from the ranged hunter consumes (or let the weave be planned so it fires), fix the +4.8 comment. Survival (5 yd) is unaffected.
- needs: orchestrator fix

## Resolved

### consumables-20261030-1: Elixir of Wisdom gives +3 Intellect in client 70707, sim has 6
- severity: low
- affects: sim/core/consumes.go (intellectElixirs, ElixirOfWisdom); no Level 60 preset uses it
- found: 2026-10-30, master 36a772fd1
- evidence: builddiff.py 70685 -> 70707: spell 3166 (item 3383 Elixir of Wisdom) Aura 29 base points 6 -> 3 (misc 3 = Intellect). consumes.go:956 has 6. Same diff: Well Fed 1249521 changed to a different aura (30/356) and Fengus' Ferocity 22817 and Missile Barrage 400589 changed; none used by a preset or the sim (git grep). Unconfirmed whether ElixirOfLesserIntellect (also 6) maps to the same spell.
- proposed fix: set ElixirOfWisdom to 3 (check Lesser Intellect against its own spell with q.py)
- needs: orchestrator fix
- resolved: 2026-10-30, sim/core/consumes.go ElixirOfWisdom 6 -> 3 (Lesser Intellect is its own spell 1250988, still 6 in 70707)
