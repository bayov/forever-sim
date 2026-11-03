# Pass: fairness between healers

Follow `orchestration/audit/PASS.md` around this doc. Written 2026-11-03, as `fairness.md` is for the DPS specs.

You check that the healer specs are compared on equal terms, so the healing sim's rankings mean something:
the Restoration shaman, Holy paladin, Restoration druid, and Holy and Discipline priest, as their presets ship.

## The pass

1. **Same terms**: each preset's request (`tools/healbench/req/*.json`) against the others: level, the synthetic
   set on its healer row (standards 35), consumables (the same tier: flask, elixirs, food, mana potion, rune, oil),
   raid buffs and debuffs, professions, race (the spec's best, or the same rule for all), heal assignment, reaction
   time and cancelling (healbench's defaults unless the doc says why). Any difference is a finding unless a
   standard or a user ruling says why.
2. **The numbers**: healbench on every boss file in both roles for every healer (or the healer_rotations pass's
   `ref/healer_bench/summary.tsv` when it ran on this master), and their spread. A healer far ahead or behind
   (say 15% carry) gets an explanation: a mechanic (send it to `healer_mechanics`' findings as evidence), a
   rotation, or the gear. Check it against what real raids had: the Anniversary logs' healer classes and HPS in
   each boss file's `raidHealing` (a raid of 40 with 11 healers: about 5 priests, 3 shamans, 2 paladins and 1
   druid), and say where the sim's ranking and the logs' disagree.
3. **What a role rewards**: tank and raid roles judged at the right lengths; a preset for one role not benched in
   the other as if it were for it.
4. Report: the comparison table, the explanations, new findings; owned doc `orchestration/fairness_healers.md`
   (the healer ranking, its reasons and the open questions), kept current.
