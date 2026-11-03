# Round 70753, phase 3: the whole audit suite

You run the forever-sim audit loop by hand, every DPS auditor at once, on master after all nine class workers of
round 70753 (the November 2 development notes, client build 1.60.1.70753) landed. Read
`orchestration/round_70753/README.md` (the round, standards 37: notes-only values), the class round files
`orchestration/round_70753/<class>.md` (each class's final summary and its "Notes-only values"), and
`orchestration/audit/README.md`.

Do "A run" from `orchestration/audit/README.md` with `python3 tools/audit/gate.py --force results rotations gear
consumables mechanics fairness client_update` instead of the plain gate (README "Running a pass by hand"). That means the
shared sims once (step 4), then one subagent per auditor with the model `auditors.json` names (step 5), record and commit
(step 6). Healer and tank auditors stay off.

Points for the passes (put them in each subagent's brief):
- Standards 37 (user ruling 2026-11-03): a value Blizzard announced that neither the client nor a server hotfix carries
  is modelled at the notes' value. A notes-only value is not a finding by itself. Do check that each one is gated, cited
  and listed in its class's round file, and that the value matches the notes' text.
- The rogue, warrior and shaman talent trees come from the notes through `tools/forever_talents/notes/<class>.json`
  (client 70753 doesn't have their rework): check them against the notes' rows, ranks and prerequisites.
- The report data (`ui/dps_report/data`) is being regenerated at the same time by other sessions, so the results
  auditor compares the requests with each class's round file numbers, not with builds.json, and fairness reads the
  new numbers from the shared sims (step 4), not the report.
- Known and planned (not findings): Salvation still in the raid buffs (being dropped in this phase), the assumption
  rows listed under "Phase 3 to-do" in the round README. The six rogue "(vs Humanoid)" builds were dropped by the user's ruling
  (their entries in ui/dps_report/data go with the regeneration). A rogue exploratory search may land a follow-up while
  you run; audit what is on master when each pass starts.

Rules: never force-push; end every commit message with the trailer lines your session prompt specifies; no model names in
files; findings go in `orchestration/audit/findings/<id>.md` as usual; land on master (the repo's CLAUDE.md). Don't fix
findings yourself. When done, put a summary at `orchestration/round_70753/audit.md`: per auditor, the new findings with
severity and `needs:`, grouped by class. Push it to master and send one message to `@parent` with `send_message`
(claude-code-remote MCP) giving the commit and the count of findings per class. Then stop.
