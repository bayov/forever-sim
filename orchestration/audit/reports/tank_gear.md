# tank_gear pass 15 (master c7e3b3e, branch prot-racials-1124)

Changed since e29cfe00f on watched paths: nothing. No synthetic tank item file, tank gear_set, gear_targets.md or
standards.md change. The merged commits (racial timing knob, Elune's Light / Eureka! casts, Phase 0 DPS gear sets) touch
no tank gear. Re-ran tools/audit/tank_gear/{warrior,paladin,druid,shaman}.py: same results as passes 13 and 14 (pieces
within tolerance, trinkets within 1b/36 targets, set ids and enchants unchanged).

## Question for this pass: do race stats change what the warrior's gear should be?
Warrior set (warrior.py): hit 80 rating = 8.0% (the standards 4 cap, no warrior tank talent adds hit); Defense 107,
nothing racial touches Defense or the cap on Defense/crit immunity.
| Race stat | Effect on gear | Verdict |
|---|---|---|
| Dwarf +1% hit | set sits 1% over the 8% cap (docs/tank_warrior.md:126) | only race affected |
| Night Elf dodge, Human crit, Tauren health | no cap or threshold they move; Stamina/Defense picks unchanged | no gear change |
| Undead (the preset race) | no hit racial, set sits exactly at cap | preset gear correct |
Dwarf waste: 10 of about 860 weighted points, re-spent into crit/Strength it is worth well under 0.1% of the set
(estimate, not simmed). The Dwarf row (3,521.7 / 2,806.2 / 861.5) is 8th of 9 and 175 / 206 / 343 behind Undead, so
re-gearing cannot move any race's place; no tank has hit steps (hit_steps.json lists none), so standards 29/124's
re-spend does not run for tanks. Not filed: below the finding bar. If the user wants race rows exact, add a
`warrior_tank` hit-steps entry; the same holds for the paladin (Dwarf is one of its three races).

Findings: none new; none open.
