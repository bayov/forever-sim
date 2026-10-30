# Findings: Consumables, buffs and debuffs

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### consumables-20261030-1: Elixir of Wisdom gives +3 Intellect in client 70707, sim has 6
- severity: low
- affects: sim/core/consumes.go (intellectElixirs, ElixirOfWisdom); no Level 60 preset uses it
- found: 2026-10-30, master 36a772fd1
- evidence: builddiff.py 70685 -> 70707: spell 3166 (item 3383 Elixir of Wisdom) Aura 29 base points 6 -> 3 (misc 3 = Intellect). consumes.go:956 has 6. Same diff: Well Fed 1249521 changed to a different aura (30/356) and Fengus' Ferocity 22817 and Missile Barrage 400589 changed; none used by a preset or the sim (git grep). Unconfirmed whether ElixirOfLesserIntellect (also 6) maps to the same spell.
- proposed fix: set ElixirOfWisdom to 3 (check Lesser Intellect against its own spell with q.py)
- needs: orchestrator fix


## Resolved
