# Tank round 2: warrior

You are the warrior tank worker in forever-sim, round 2 after the first tank audit. The shared engine fixes are on master (5668ea5): every scored tank faces the boss (boss parries, blocks), encounter files keep the logs' swing gap with parryHaste off, survival cooldowns never read health in scoring runs, and every tank carries Elixir of the Phalanx and Greater Fortitude. Your job: apply your class's findings, then re-optimize your three presets (Max threat, Balanced, Max survival) on the new engine, as the first round did.

Read the repo's CLAUDE.md first and follow it, above all "Keep the context small": hand sim runs, rotopt searches, test runs and wide reads to subagents and keep only their conclusions. Background: orchestration/brief_tanks.md (your class's row and sections T2/T3), docs/tank_sim.md, orchestration/standards.md (it overrides older docs; item 40 has the user's tank rulings from today), your docs/tank_warrior.md and orchestration/tank_handoff/warrior.md. The findings, with evidence, repro commands and drafted patches, are in orchestration/audit/findings/tank_{mechanics,rotations,gear,fairness}.md (patches in orchestration/audit/ref/tank_mechanics/patches/; the rotations auditor's paired A/B harness is tools/audit/tank_rotations/abtest).

Your findings:
- tank_rotations-20261104-2 (high): Shield Wall in the last 18 s, not 12 (Balanced, Max survival; the rotopt wallAt grid is too coarse). Check Vaelastrasz, which loses, and 16 s.
- tank_rotations-20261104-7 (low): Max threat's Sunder filler at 15 Rage, not 30.
- tank_mechanics-20261104-4: user ruling: Revenge takes 0.25 attack power on every rank, with a to-verify note in code and doc (standards 40).
- tank_mechanics-20261104-6 (medium): Unbridled Wrath procs on Heroic Strike, but client mask 4 is auto attacks only. This moves Fury and Arms too, so run dps_warrior's tests and regenerate its .results. Report the DPS change per DPS preset.
- tank_gear-20261104-2 (medium): the warrior wears the attack power on-use trinket 991018 with no A/B against the armor trinket 990617 (standards 36). Run it at each preset's weight and keep the winner.
- tank_gear-20261104-3 (low): stale Classic enchant values: Shield Greater Stamina +7 (should be +9), Bracer Deflection understated in enchant_overrides.go, Superior Deflection unevaluated. It's a shared file: fix the values for every tank, and say which other tanks' numbers move.
- tank_gear-20261104-5 (low, your part): the warrior header numbers are stale.
- tank_rotations-20261104-8 (low, your part): the Defensive Stance note.
- From the wowsims/forever comparison (no finding yet): wowsims/forever 841a787806 now takes "% of base health" costs (Bloodrage 20%, Reckless Swings 25%) from a base that counts base Stamina (Orc warrior at 60: 2629, not 1689), with no evidence cited. Keep our class base (the emulator convention), with a to-verify note (a level 60 Bloodrage in a combat log settles it). They also stopped counting self costs as damage taken. Keep them in healing needed, since a healer refills them. If the score or report keeps DTPS or TMI apart from healing needed, keep self costs out of those, and say what you did.
- tank_fairness-20261104-5 (low): leave warrior_protection.json alone (the healing session decides). Say in your doc that it's legacy.

Then:
1. Re-optimize the three presets on the new engine (talents, rotation, gear choices the standards allow) with rotopt and tankbench: 1,000 iterations a row, 3,000 to settle a gain under twice its paired error, scored as the mean of the six scoring bosses at the preset's weight (1/3, 1, 3). Ragnaros is benched, not scored. A preset whose best build equals another's says so in its note (tank_fairness-20261104-4); don't force a difference.
2. Each mechanic you change gets a test that fails when it breaks (tank_mechanics-20261104-11 lists the untested ones; do your class's). Fix your class's rows in tank_mechanics-20261104-10 (doc tier labels, stale rows).
3. Update docs/tank_warrior.md (numbers, calls for the user) and the presets in ui/tank_warrior and tools/tankbench/req/warrior_protection_{max_threat,balanced,max_survival}.json. Re-bench your three presets and rewrite only your rows of orchestration/audit/ref/tank_bench/summary.tsv (three other class workers do theirs at the same time: on a rebase conflict there, keep every other row as master has it).
4. Under each finding you fixed, add "- fixed: <date>, <commit>". Where you leave one open, say why under it.
5. Run the whole test suite (go test -tags=with_db over ./sim/... minus sim/practice and sim/web, plus npx tsc --noEmit -p .). Regenerate .results or report hashes your change moves, and say which and why.

Work on branch tank-warrior (create it from master) and push after each commit. When everything is green, merge origin/master in, run the tests again if master changed under you, then merge into master and push (no PR, per CLAUDE.md). Don't regenerate or deploy the DPS report: another session is restructuring it. If your change moves DPS presets' numbers, say by how much. Commits end with:
Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>

When done, send the result to session session_01TXnjyP5nUQs8zySrhUvYoi with send_message (claude-code-remote), in at most 12 lines: each finding fixed or left (one line each), the three presets' talents and scores at their own weight, before and after, the master commit, and every call left for the user.

