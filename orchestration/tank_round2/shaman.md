# Tank round 2: shaman

You are the shaman tank worker in forever-sim, round 2 after the first tank audit. The shared engine fixes are on master (5668ea5): every scored tank faces the boss (boss parries, blocks), encounter files keep the logs' swing gap with parryHaste off, survival cooldowns never read health in scoring runs, and every tank carries Elixir of the Phalanx and Greater Fortitude. Your job: apply your class's findings, then re-optimize your three presets (Max threat, Balanced, Max survival) on the new engine, as the first round did.

Read the repo's CLAUDE.md first and follow it, above all "Keep the context small": hand sim runs, rotopt searches, test runs and wide reads to subagents and keep only their conclusions. Background: orchestration/brief_tanks.md (your class's row and sections T2/T3), docs/tank_sim.md, orchestration/standards.md (it overrides older docs; item 40 has the user's tank rulings from today), your docs/tank_shaman.md and orchestration/tank_handoff/shaman.md. The findings, with evidence, repro commands and drafted patches, are in orchestration/audit/findings/tank_{mechanics,rotations,gear,fairness}.md (patches in orchestration/audit/ref/tank_mechanics/patches/; the rotations auditor's paired A/B harness is tools/audit/tank_rotations/abtest).

Your findings:
- tank_rotations-20261104-4 (high): Earth Shock above Stormstrike (Max threat +62). Balanced and Max survival are noisy on Broodlord, so rotopt-eval those at 3,000 iterations.
- tank_mechanics-20261104-5: ruling (standards 40): Rockbiter keeps 72 × weapon speed on specials. Add the to-verify note to sim/shaman/rockbiter_weapon.go and your doc. No value change.
- tank_gear-20261104-6 (low): the boots take Greater Stamina, and Mystic Rugged Armor Kit is unevaluated (about +0.2% weighted). Evaluate both and keep the winners.
- tank_rotations-20261104-8 (low, your part): the shield note.
- tank_fairness-20261104-3: user ruling (standards 40): the Warden's 2 to 3 times higher healing need is a real class gap (mail, little built-in defense; it won't necessarily be a viable raid tank). Don't compensate for it. Check only that its gear and mechanics are correct, and say so in your doc. The gear auditor found it over the hit cap by its sim: check that too.

Then:
1. Re-optimize the three presets on the new engine (talents, rotation, gear choices the standards allow) with rotopt and tankbench: 1,000 iterations a row, 3,000 to settle a gain under twice its paired error, scored as the mean of the six scoring bosses at the preset's weight (1/3, 1, 3). Ragnaros is benched, not scored. A preset whose best build equals another's says so in its note (tank_fairness-20261104-4); don't force a difference.
2. Each mechanic you change gets a test that fails when it breaks (tank_mechanics-20261104-11 lists the untested ones; do your class's). Fix your class's rows in tank_mechanics-20261104-10 (doc tier labels, stale rows).
3. Update docs/tank_shaman.md (numbers, calls for the user) and the presets in ui/warden_shaman and tools/tankbench/req/shaman_warden_{max_threat,balanced,max_survival}.json. Re-bench your three presets and rewrite only your rows of orchestration/audit/ref/tank_bench/summary.tsv (three other class workers do theirs at the same time: on a rebase conflict there, keep every other row as master has it).
4. Under each finding you fixed, add "- fixed: <date>, <commit>". Where you leave one open, say why under it.
5. Run the whole test suite (go test -tags=with_db over ./sim/... minus sim/practice and sim/web, plus npx tsc --noEmit -p .). Regenerate .results or report hashes your change moves, and say which and why.

Work on branch tank-shaman (create it from master) and push after each commit. When everything is green, merge origin/master in, run the tests again if master changed under you, then merge into master and push (no PR, per CLAUDE.md). Don't regenerate or deploy the DPS report: another session is restructuring it. If your change moves DPS presets' numbers, say by how much. Commits end with:
Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>

When done, send the result to session session_01TXnjyP5nUQs8zySrhUvYoi with send_message (claude-code-remote), in at most 12 lines: each finding fixed or left (one line each), the three presets' talents and scores at their own weight, before and after, the master commit, and every call left for the user.

