# Tank round 3: warrior

In bayov/forever-sim, you are the round-3 warrior tank worker. Tank auditors' second pass (master 06a1f170b and later) left findings for your class; fix them and re-optimize your three presets (Max threat, Balanced, Max survival). Read first: orchestration/standards.md item 40 (user rulings), orchestration/tank_round2/warrior.md (your round-2 brief, same procedure and deliverables), docs/tank_warrior.md, and the findings files orchestration/audit/findings/tank_{gear,rotations,fairness,mechanics}.md. Follow CLAUDE.md: keep your context small (sim runs, test runs and wide reads go to subagents). Develop on branch tank-warrior-3; merge into master and push (no PR); if master moves, merge it in, never force-push.

Order of work:
A. Now: the non-score items below (gear sets, UI presets, enchant ids, docs, tooltips).
B. A separate engine session is changing mana-gain threat (tank_mechanics -12: it will be unmultiplied by the threat multiplier) and making tank-tool iteration seeds disjoint (tank_rotations -13). Before any re-optimization or final benches, wait until origin/master's orchestration/audit/findings/tank_mechanics.md marks -12 fixed (git fetch every ~10 min; use the Monitor tool with an until-loop, not sleep). Then merge master in and do the score work on top of it.
C. Re-optimize the three presets on the new engine with the same search as round 2, plus the items below. Confirm every preset choice that rests on a small paired difference on a disjoint seed (per tank_rotations -13). Commit the raw search output (tank_fairness -8: only the warrior did last round) under tools/tank_bench/ or where round 2's warrior put it.
D. Update docs/tank_warrior.md (preset table, calls), the UI presets and tooltips, tools/tank_bench/summary.tsv rows for your class, tankbench reqs; full go test suite and tsc green; merge to master. Mark each finding you fixed in its findings file with the commit.

Shared items for every class:
- tank_gear -3 / -7: shield enchant — 929 no longer exists in the DB (CLI gives +7 by effect id; the web UI loads no enchant). Use 1886 (+9) on shields where your class has one. Wrists: A/B Superior Deflection (1248665) against your current 1886 like the warrior did. Cloak: A/B Agility +5 (1219587) against Superior Defense +70 armor (1889) at each preset. Record the A/Bs in the doc.
- tank_fairness -6: the UI presets must match the tankbench reqs exactly (buffs incl. Kings and Might, consumables, professions).
- tank_fairness -9: stale tooltips and numbers for your class.
- User calls stay calls: don't decide them, list them in the doc.

Class items:
- tank_fairness -6: the warrior UI uses other professions than its bench; match them.
- tank_rotations -11 is a user call (warrior Max survival's held armor trinket wins only on Broodlord; +71 for 991018 on the pull without Broodlord): keep the current choice, report the numbers both ways on the new engine, and list it as a call next to paladin Bulwark (-3).
- The warrior gains no mana, so expect the engine change to move nothing; still re-bench to confirm.

When done, send a result of at most 10 lines to session session_01TXnjyP5nUQs8zySrhUvYoi with send_message (claude-code-remote): master commit, each finding fixed (id + commit), each preset's build and own-weight six-boss mean before → after (with error), and calls for the user.

