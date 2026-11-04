# Round 70753, phase 5: full DPS report re-sim and deploy

Since the last deploy, master has landed:
- the hit-trade gear and hit talent re-search (`fdea38921`; `orchestration/round_70753/hit_trade/numbers.md`, `hit_sweep.md`);
- the warrior hotfix 113539 (`f9a03d6e7`; `orchestration/round_70753/hotfix_113539_warrior.md`);
- the shaman launch-damage hotfix 113380 (`8358d14bd`);
- labeled random streams (`ca0ca1285`);
- adaptive fight counts (`18ec491ec`);
- the new warrior assumption `warr_rs_landed` (`72c7d150f`).

You re-sim every build, publish, land and deploy. First read:
- `orchestration/round_70753/README.md`;
- `.claude/skills/regenerate-dps-report/SKILL.md`;
- `tools/dps_report/README.md`;
- the dps-report skill's B7, B8, B9 and B11 (`.claude/skills/dps-report/SKILL.md`).

A fresh clone first needs `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`. The `.pb.go`
files are gitignored, and stale ones silently sim old field numbers.

1. **Full run.** Run `pipeline.sh run` with no ids, so that every build is re-simmed, then `pipeline.sh publish` for all
   builds. Run it in the background and check its log briefly; don't sleep-poll.
   - Packs from before `ca0ca1285` (labeled RNG) or `18ec491ec` (adaptive counts) aren't comparable, so don't mix any in.
   - The race, race-scenario and 180 s assumption steps now sim a pilot (2k fights for races, 1k otherwise), then top up
     to their luck goal, up to 40k fights. Expect warrior and warlock races to go past 10k.
   - Per-fight vectors go to `$DASH_DIR/vec`. That folder is gitignored and isn't packed, but it can take a few hundred
     MB, so watch the disk and delete it once the publish is done.
   - Check each build's 180 s number:
     - the 21 regeared builds against `hit_trade/numbers.md`;
     - the warriors against `hotfix_113539_warrior.md`;
     - the rest against the live report.
     A gap beyond the sim's noise is a question to report, not something to paper over.
2. **Front end.** `report.js` and `layout.html` changed in `18ec491ec` and aren't deployed yet. This publish and deploy
   ship them together with the new data. Run the B8 Playwright checks on the wrapped dist file:
   - phone and desktop width;
   - both themes;
   - no console errors.
3. **Changes tab** (standards 30, `tools/dps_report/changes.py`'s rules). Write public lines in plain game terms for:
   - the hit-trade gear and the builds it moved;
   - the warrior hotfix 113539;
   - the re-sim on labeled random streams;
   - adaptive fight counts;
   - the new assumption `warr_rs_landed`.

   Also fill in the DPS delta on the existing shaman 113380 line in `tools/dps_report/changes/updates.json`. The user ruled
   out a rotopt for shaman 113380, so don't run one. `python3 tools/dps_report/changes.py check` must list nothing
   uncovered.
4. **Small fix.** The Fury Sunder duty meta's rotation text is stale. Bring it in line with its current APL.
5. **Checks and land.**
   - Rewrite `sim/testdata/report_hashes.txt` for anything that moved, and run the full suite:
     `go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`. It takes about 17 minutes; give it
     to a subagent.
   - Merge origin/master last. Commit `ui/dps_report/data`, `datasets.json`, `provenance.json` and
     `tools/dps_report/raw` together, and push to master.
   - If master moved a build meanwhile, re-sim just those builds (`hydrate`, then `run <ids>`, then `publish <ids>`).
   - Deploy from the pushed commit with `make dps-report-deploy` (this environment's proxy authenticates
     api.cloudflare.com).
   - Update both `state.json` files (B11).
6. **Timing.** Send the timing summary that `pipeline.sh run` prints at the end
   (`tools/dps_report/timing/<start>.jsonl`, summed per step) to `session_01TXnjyPE5JfDBjP3yqKyxXE` with `send_message`.
   That session refreshes its W cost table from it. It needs no reply.

Open user rulings may still move a default after you land (the Demonology Suppression pin, the Ret default). If one
arrives while you run, the parent will tell you. Otherwise leave every pick as master has it.

Rules:
- Never force-push.
- End every commit message with the trailer lines your session prompt specifies.
- Keep model names out of files.
- Keep long runs in the background or in subagents, so your own context stays small (the repo's CLAUDE.md).

When deployed:
- Write a short summary at `orchestration/round_70753/report_phase5.md` with the 180 s ranking, every build's old (live)
  and new 180 s number, the race-table changes and the deploy id, and push it.
- Send one message to `@parent` with `send_message` (claude-code-remote MCP) giving the master commit, the deploy, every
  build's old and new 180 s number, and anything that needs a ruling.

Then stop.
