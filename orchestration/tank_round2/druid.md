# Tank round 2: druid

You are the druid tank worker in forever-sim, round 2 after the first tank audit. The shared engine fixes are on master (5668ea5): every scored tank faces the boss (boss parries, blocks), encounter files keep the logs' swing gap with parryHaste off, survival cooldowns never read health in scoring runs, and every tank carries Elixir of the Phalanx and Greater Fortitude. Your job: apply your class's findings, then re-optimize your three presets (Max threat, Balanced, Max survival) on the new engine, as the first round did.

Read the repo's CLAUDE.md first and follow it, above all "Keep the context small": hand sim runs, rotopt searches, test runs and wide reads to subagents and keep only their conclusions. Background: orchestration/brief_tanks.md (your class's row and sections T2/T3), docs/tank_sim.md, orchestration/standards.md (it overrides older docs; item 40 has the user's tank rulings from today), your docs/tank_druid.md and orchestration/tank_handoff/druid.md. The findings, with evidence, repro commands and drafted patches, are in orchestration/audit/findings/tank_{mechanics,rotations,gear,fairness}.md (patches in orchestration/audit/ref/tank_mechanics/patches/; the rotations auditor's paired A/B harness is tools/audit/tank_rotations/abtest).

Your findings:
- tank_rotations-20261104-1 (high): Demoralizing Roar takes the pull's first GCD. Move it below the Lacerate keep-up line (about +200 at every weight).
- tank_rotations-20261104-5 (medium): Max survival's pre-pull Enrage at -1 s, not -3 s (+45.6 on top of the first fix).
- tank_mechanics-20261104-2 (high): Primal Bite's threat ×2 (the notes: "threat roughly doubled"), notes-only under standards 37 (+23% Balanced in the auditor's run).
- tank_mechanics-20261104-7 (medium): the bear's white hits use Classic's damage Rage. Under standards 21 it's the flat 11.25, doubled on crits.
- tank_gear-20261104-4 (low): the bear chest 990805 is -1.3% linear, and the set's Stamina share is 35.1 against 33.4 ±1.5. Bring both to target.
- tank_gear-20261104-5 (low, your part): comments name trinket 990815 where 998115 is worn.
- tank_rotations-20261104-8 (low, your part): the Enrage and Frenzied Regeneration wording.
- The core worker found that Max survival (1185) beats Balanced (1114) at the Balanced weight. Balanced must be the best build at weight 1, so re-optimize it.

Then:
1. Re-optimize the three presets on the new engine (talents, rotation, gear choices the standards allow) with rotopt and tankbench: 1,000 iterations a row, 3,000 to settle a gain under twice its paired error, scored as the mean of the six scoring bosses at the preset's weight (1/3, 1, 3). Ragnaros is benched, not scored. A preset whose best build equals another's says so in its note (tank_fairness-20261104-4); don't force a difference.
2. Each mechanic you change gets a test that fails when it breaks (tank_mechanics-20261104-11 lists the untested ones; do your class's). Fix your class's rows in tank_mechanics-20261104-10 (doc tier labels, stale rows).
3. Update docs/tank_druid.md (numbers, calls for the user) and the presets in ui/feral_tank_druid and tools/tankbench/req/druid_bear_{max_threat,balanced,max_survival}.json. Re-bench your three presets and rewrite only your rows of orchestration/audit/ref/tank_bench/summary.tsv (three other class workers do theirs at the same time: on a rebase conflict there, keep every other row as master has it).
4. Under each finding you fixed, add "- fixed: <date>, <commit>". Where you leave one open, say why under it.
5. Run the whole test suite (go test -tags=with_db over ./sim/... minus sim/practice and sim/web, plus npx tsc --noEmit -p .). Regenerate .results or report hashes your change moves, and say which and why.

Work on branch tank-druid (create it from master) and push after each commit. When everything is green, merge origin/master in, run the tests again if master changed under you, then merge into master and push (no PR, per CLAUDE.md). Don't regenerate or deploy the DPS report: another session is restructuring it. If your change moves DPS presets' numbers, say by how much. Commits end with:
Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>

When done, send the result to session session_01TXnjyP5nUQs8zySrhUvYoi with send_message (claude-code-remote), in at most 12 lines: each finding fixed or left (one line each), the three presets' talents and scores at their own weight, before and after, the master commit, and every call left for the user.

