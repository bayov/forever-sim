# Protection Warrior racials deep dive (2026-11-24)

The user is starting a Horde Protection Warrior now. A quick pass (0f567e196) found that the level 60
prot APLs never cast Blood Fury or Berserking, fixed that, and ranked the Horde races on the Balanced
preset (six scored bosses, 10k iterations, seed 1):

| Race | Score | vs best |
|---|---|---|
| Undead | 3475 | 0 |
| Orc | 3411 | -1.8% |
| Troll | 3356 | -3.4% |
| Tauren | 3343 | -3.8% |
| Windshaper | 3306 | -4.9% |

Max threat: Orc first, Undead -0.1%. Max survival: Undead first by about 14%. The quick pass's logs are
gone, rerun what you need.

## Task

Bring the prot warrior race results to the DPS report's standard. Horde races first, Alliance after.

1. **Touch of the Grave.** Undead's lead rests on its self-heal hitting the "up to 5% of max health"
   cap, which `docs/tank_warrior.md` flags as a call. Find what it really does in Forever: client spell
   data under the sources, foreverchanges, the user's combat logs if any are in the repo (log-audit
   skill), Wowhead. Model it on the evidence and say how sure it is.
2. **Fear and Will of the Forsaken.** Neither is modelled, and Magmadar's Panic is left out. Find which
   scored bosses fear the tank, and model the least that is honest: the fear, the tank's answer
   (Berserker Rage stance dance, Will of the Forsaken, Fear Ward if the report assumes a priest), and
   what a fear costs (lost threat, a loose boss). If a fair model is out of reach, say so in the report's
   prot notes rather than guess.
3. **Racial timing.** Blood Fury and Berserking go on cooldown at the pull. Search when to use them per
   preset (pull, after N seconds, with Death Wish or trinkets, at low health for Berserking), with the
   same tooling the report uses for rotation searches.
4. **Every racial works.** Check each race's racials for prot warrior: passives applied, actives cast,
   and the score reads the stat they move. Fix what's broken.
5. Re-sim every prot preset per race at the report's usual iterations, regenerate the tank race tables,
   run the tank auditors (tank_mechanics, tank_rotations, tank_gear in `orchestration/audit/auditors.json`,
   brief `orchestration/brief_tanks.md`), do the fix round, and `make dps-report-deploy`.

## Coordination

Phase 0 step 4 runs at the same time: class workers change DPS requests, gear sets and `tools/rotopt/phase0*`
and the parent publishes the DPS data at the end. Don't touch DPS requests or Phase 0 files, and publish
only the tank ids, so your deploy ships no half-done DPS data. Fetch and merge master before each push and
before the deploy. Never force-push.

## Hand back

The user is waiting to pick a race, so send the parent a first message as soon as steps 1 to 3 have
numbers, before the audit: the Horde ranking per preset (Balanced, Max threat, Max survival) with what
changed and why, under 200 words. Then a final message after the deploy: the master commit, the final
Horde and Alliance tables, the audit findings and fixes, under 300 words. Then stop.
