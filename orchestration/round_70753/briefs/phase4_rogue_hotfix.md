# Round 70753, phase 4: rogue server hotfix push 113483

The audit loop's client_update pass found a high finding after round 70753 closed: client_update-20261103-3 in
`orchestration/audit/findings/client_update.md`. Server hotfix push 113483 (logged 2026-11-03 18:29Z, region 91, build
70707) changes the rogue tree and rogue spells, and the sim doesn't read most of it. Client 70755 has none of it. The
release launches 2026-11-04, so this is the one open item that moves DPS.

Read the finding, `orchestration/round_70753/README.md`, `orchestration/round_70753/rogue.md` (the rogue tree comes from
the notes via `tools/forever_talents/notes/rogue.json`), `orchestration/standards.md` (source tiers: a server hotfix
beats the notes and the client), and the `update-sources` skill for `tools/sources/hotfixes.py`.

1. Finish the fetch the finding's proposed fix names (`hotfixes.py fetch 1.60.1.70707 ... --refresh`), then `apply` and
   `import_db2.py 1.60.1.70707+hf rogue`. Read every row of push 113483 and write what it changes to
   `orchestration/round_70753/hotfix_113483.md`. Give each changed spell a ledger row, as the finding says.
2. Land the values under Forever, gated and cited like the other hotfix values: the poisons (Instant, Deadly, Wound,
   Mind-numbing: find what each effect row drives), Mutilate's flat bonus, Shank's real spell ids and Flash of Steel, and
   the tree's two new nodes and deleted node (Improved Distract). Where a hotfix value replaces a notes-only one, take
   the hotfix's and drop it from rogue.md's "Notes-only values".
3. If the tree change makes any rogue preset's talents illegal or opens a better point, re-pick talents and re-run the
   rotation and race checks the way the rogue worker did (rogue.md); otherwise keep the presets. Re-sim every rogue
   request at 120/180/300 s, rewrite `sim/testdata/report_hashes.txt` for the rogue rows, and run the full test suite
   (`go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`; a fresh clone first needs
   `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`).
4. Regenerate the DPS report for the rogue builds (`regenerate-dps-report` skill: hydrate, `run <rogue ids>`,
   `publish <rogue ids>`), add a public Changes-tab line for the rogue hotfix (`tools/dps_report/changes.py check` clean),
   land on master and deploy (`make dps-report-deploy`). Mark the finding resolved as `orchestration/audit/README.md` says.

Rules: never force-push; end every commit message with the trailer lines your session prompt specifies; no model names in
files; long runs in the background or in subagents, and keep your own context small (the repo's CLAUDE.md). When done,
add the old and new 180 s number of every rogue build to `hotfix_113483.md`, push it, and send one message to `@parent`
with `send_message` (claude-code-remote MCP): the master commit, the deploy, the rogue numbers, and anything that needs a
user ruling. Then stop.
