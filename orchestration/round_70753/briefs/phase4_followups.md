# Round 70753, phase 4: follow-ups from the rogue hotfix session

The rogue hotfix 113483 session landed on master 9ff134b0b and deployed. Its report
(`orchestration/round_70753/hotfix_113483.md`) left these. Read `orchestration/round_70753/README.md`,
`orchestration/standards.md` (source tiers 37 to 39: a server hotfix beats the notes and the client; the sim is Forever
only) and the `regenerate-dps-report` skill. A shaman session (phase 4 shaman) may publish shaman builds at the same
time: merge origin/master right before you publish, and if the report data moved, take master's, hydrate, re-check
stale and publish again.

1. **TestRaidForeverReportHashesStable fails on master**: "No item with id: 998001" in the druid Balance requests.
   d57aa24 (client 70755) last rewrote db.bin without the synthetic Balance gear. Find where the synthetic items come
   from, rebuild db.bin so it has them (and check no other synthetic or report item went missing), and get the test
   green with no Forever DPS moving.
2. **compsearch TestDropsAreUnread fails** for warlock__affliction and its two duty variants (already failing at
   ca22004). Find why and fix the request or the code.
3. **Paladin rows of hotfix push 113483**: `import_db2.py 1.60.1.70707+hf paladin` shows 4 number and 3 wording diffs
   (Oath of Light, Vengeance, Two-Handed Weapon Specialization) that were never written. Read each row, land the
   values under Forever, gated and cited, with a ledger row each, as the rogue session did. Re-sim the paladin builds
   they move at 120/180/300 s; if a talent's value changed enough to change a pick, re-check the talents the way the
   paladin worker did (`orchestration/round_70753/paladin.md`). Write `orchestration/round_70753/hotfix_113483_paladin.md`.
   Also check every other class's rows in the same push the same way (`import_db2.py ... <class>`), and say which
   showed diffs.
4. **stale.mjs reports all 68 builds stale by hash** after the rogue publish. Find out why (item 1's db.bin? a hash
   input that moved?). If the builds really are stale, re-sim only what changed; if the hash is wrong, fix it. Remove
   the leftover `../w-dash` worktree if it's yours to remove (nothing else uses it).
5. **Warrior note**: the Fury numbers in the warrior line of the Changes tab (or its note) may be out of date after
   7b4dc5d. Check them against the published report and fix the text.

Then rewrite `sim/testdata/report_hashes.txt` for the rows that moved, run the full test suite
(`go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`; a fresh clone first needs
`PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`), regenerate the moved builds in the
report, add Changes-tab lines (`tools/dps_report/changes.py check` clean), land on master and deploy
(`make dps-report-deploy`).

Rules: never force-push; end every commit message with the trailer lines your session prompt specifies; no model names in
files; long runs in the background or in subagents, and keep your own context small (the repo's CLAUDE.md). When done,
send one message to `@parent` with `send_message` (claude-code-remote MCP): the master commit, the deploy, the old and
new 180 s numbers of any build that moved, and anything that needs a user ruling. Then stop.
