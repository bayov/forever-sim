# Pass: fairness between tanks

Follow `orchestration/audit/PASS.md` around this doc. Written 2026-11-03, as `fairness.md` is for the DPS specs.

You check that the tank specs are compared on equal terms, so the tank sim's rankings mean something: the
Protection warrior, Protection paladin, Feral tank and Warden shaman, as their presets ship.

## The pass

1. **Same terms**: each preset's request (`tools/tankbench/req/*.json`) against the others: level, the synthetic
   set on its Tanks row (standards 36), consumables (the same tier: flask, elixirs, food, potions, weapon stones and
   oils, armor kits), raid buffs and debuffs (the tank's own debuffs only through `tankscore.OwnDebuffs`), Salvation
   off, professions, race (the spec's best by the objective, or the same rule for all), reaction time. Any
   difference is a finding unless a standard or a user ruling says why.
2. **The numbers**: tankbench on every boss file for every tank preset (or the tank_rotations pass's
   `ref/tank_bench/summary.tsv` when it ran on this master), and their spread at each weight. A tank far ahead or
   behind (say 15% on the score or 25% on TPS or healing needed) gets an explanation: a mechanic (send it to
   `tank_mechanics`' findings as evidence), a rotation, or the gear. Check it against what real raids had: the
   Anniversary logs' main tanks in each boss file (`healing.notes` names their classes) and their damage taken, and
   say where the sim's ranking and the logs disagree.
3. **What a preset rewards**: Max threat, Balanced and Max survival each ahead at its own weight, and the three
   not the same build (a sign the weights don't separate them).
4. Report: the comparison table, the explanations, new findings; owned doc `orchestration/fairness_tanks.md` (the
   tank ranking, its reasons and the open questions), kept current.
