# Round 70753, phase 4: Chain Lightning's coefficient and the shaman leftovers

The user ruled on 2026-11-04 (orchestration/standards.md 39): Chain Lightning's spell power coefficient is the standard
one for its new 2.5 s cast, 2.5 / 3.5 = 0.714 (the sim kept 0.571). It's a notes-only value (standards 37). Also new:
the sim is Forever only (standards 38), so no Classic fallback is needed. Read `orchestration/round_70753/README.md`,
`orchestration/round_70753/shaman.md` (its Chain Lightning notes at about line 225), and the `regenerate-dps-report` skill.

1. **Chain Lightning 0.714** for every rank the notes' 2.5 s cast covers, cited and listed in shaman.md's "Notes-only
   values". At about 1.27 x a Lightning Bolt it may enter Elemental's rotation: search the Elemental rotations (and the
   Mana Tide build) with and without Chain Lightning lines on single target, the way the shaman worker searched
   (shaman.md), and keep what wins on two seed sets at 120/180/300 s. Enhancement too if it casts Chain Lightning. Update
   the requests, metas (`published.dps`), presets.ts, the wiki and shaman.md. Update the assumption rows that name the
   0.571 coefficient (`tools/dps_report/assumptions_def.py`).
2. **The six shaman req metas** still carry the pre-round `published.dps` (results-20261102-1): set them from the new
   finals.
3. **compsearch TestReportRaid flake** (README "Phase 3 to-do"): shaman__enhancement-imp-gw does about 6 DPS less in the
   20-player raid than alone (-6.0 against an allowed +-5.9 at 4000 iterations). Find why (a raid-only interaction: a
   totem, a buff overwritten, Windfury from another shaman) and fix the sim or the request. Don't widen the tolerance
   unless the gap is noise, and then show it is.
4. Rewrite `sim/testdata/report_hashes.txt` for the shaman rows, run the full test suite
   (`go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`; a fresh clone first needs
   `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`), regenerate the shaman builds in the
   DPS report (hydrate, `run <ids>`, `publish <ids>`), add a Changes-tab line, land on master and deploy
   (`make dps-report-deploy`). Another session (rogue hotfix 113483) may be publishing rogue builds at the same time:
   merge origin/master right before you publish, and if the report data moved, follow the README (take master's,
   hydrate, re-check stale, publish again).

Rules: never force-push; end every commit message with the trailer lines your session prompt specifies; no model names in
files; long runs in the background or in subagents, and keep your own context small (the repo's CLAUDE.md). When done,
write `orchestration/round_70753/shaman_phase4.md` (old and new 180 s number of every shaman build, the raid gap's cause),
push it, and send one message to `@parent` with `send_message` (claude-code-remote MCP): the master commit, the deploy,
the numbers, anything that needs a user ruling. Then stop.
