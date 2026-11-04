# Round 70753, phase 5: warrior server hotfix push 113539

The phase 4 follow-ups session (`orchestration/round_70753/hotfix_113483_paladin.md`, README "Phase 4") ran
`import_db2.py 1.60.1.70707+hf <class>` for every class and found that warrior push 113539 is not in the sim: 6 structure,
1 number and 7 wording diffs. Raging Blows is removed; Reckless Swings and Improved Whirlwind are in the tree; Spearing
Strike goes from 40/80 to 75/75 (the sim still has 40/80); Booming Voice covers the Howls; Slam, Sweeping Strikes, Deep
Wounds and Concussion Blow are also touched.

Read the README, `orchestration/round_70753/warrior.md`, `orchestration/round_70753/hotfix_113483.md` (the rogue session
did the same job for push 113483: follow its method), `orchestration/standards.md` (source tiers: a server hotfix beats
the notes and the client; the sim is Forever only), and the `update-sources` skill for `tools/sources/hotfixes.py`.

1. Read every warrior row of push 113539 (and any other warrior rows in the same 70707 hotfix pushes) and write what each
   changes to `orchestration/round_70753/hotfix_113539_warrior.md`, with a ledger row per changed spell.
2. Land the values under Forever, gated and cited like the other hotfix values, including the tree's structure changes
   (`tools/forever_talents`). Where a hotfix value replaces a notes-only one, take the hotfix's and drop it from
   warrior.md's notes-only list. The user had an open question on Reckless Swings; say what the hotfix settles about it.
3. If the tree change makes any warrior preset illegal or opens a better point, re-pick talents and re-run the rotation
   and race checks the way the warrior worker did (warrior.md); otherwise keep the presets. The phase 5 hit-trade session
   (branch `phase5-hit-trade`, `briefs/phase5_hit_trade.md`) is changing the talent search so hit talents are scored with
   gear that trades hit away, and is re-searching every class, warrior included. If its tool is on master when you
   search, use it; if not, search the usual way, say so, and tell it (session_01TXnjySCSuMhjz5SQHToQZg, with
   send_message) which warrior picks changed so its warrior result is redone on the new tree. Merge origin/master before
   you publish; if the report data moved, take master's, hydrate, re-check stale and publish again. Never rebuild db.bin
   before merging master (two sessions have rebuilt it today).
4. Re-sim every warrior request at 120/180/300 s, rewrite `sim/testdata/report_hashes.txt` for the warrior rows, run the
   full test suite (`go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`; a fresh clone first
   needs `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`), regenerate the warrior builds
   in the report, add a public Changes-tab line (`tools/dps_report/changes.py check` clean), land on master and deploy
   (`make dps-report-deploy`).
5. Side item from the same follow-ups: paladin push 113542 confirms Vengeance 5 stacks and Two-Handed Weapon
   Specialization covering Holy. `pal_2h_spec_holy` (`tools/dps_report/assumptions_def.py`, `spec_assumptions.md:251`) and
   `forever-wiki/classes/paladin.md` still call these notes-only: cite hotfix 113542 instead. Also cite push 113527 in
   the warlock notes entries for Wrack and Shadow and Flame. No DPS moves.

Rules: never force-push; end every commit message with the trailer lines your session prompt specifies; no model names in
files; long runs in the background or in subagents, and keep your own context small (the repo's CLAUDE.md). When done,
add the old and new 180 s number of every warrior build to `hotfix_113539_warrior.md`, push it, and send one message to
`@parent` with `send_message` (claude-code-remote MCP): the master commit, the deploy, the warrior numbers, every talent
pick that changed, and anything that needs a user ruling. Then stop.
