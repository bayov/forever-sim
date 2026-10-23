# Class worker brief, regear and rotation round (2026-10-23)

Read `<Class>` as your class name and `<class>` as its lower-case form.

You are the **<Class>** class worker for the WoW Forever level 60 DPS report in the forever-sim repo (github.com/bayov/forever-sim, a wowsims/classic fork, Forever ruleset). Another session (the orchestrator) started you; it will merge your branch.

## Branch and rules
- Start from branch `claude/keen-brown-o2g32n` at commit 24a17e198f or later. Create and push only to `claude/keen-brown-o2g32n-<class>` (`git push -u origin claude/keen-brown-o2g32n-<class>`). Never force-push. Don't open a PR.
- End every commit message with exactly these lines:
  Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
  Claude-Session: https://claude.ai/code/session_01S9W3rHHkMdAc24iprSm6vo
- No model names in files.
- Stay inside your class: `sim/<class>/`, `ui/<your specs>/`, your class's `tools/rotopt` templates, your `tools/dps_report/req/<class>__*.json` requests, your synthetic items in `tools/database/forever_synthetic_items_<class>*.go` (if any), and your class's sections of `forever-wiki/` and `notes.md`.
- Don't edit `sim/core`, `sim/common`, `orchestration/standards.md`, or `tools/rotopt/mksettings/main.go`'s shared raid/encounter settings. If you need a shared change or a cross-class ruling, put it in your final report instead.
- If you change synthetic items, regenerate the DB (`go run ./tools/database/gen_db -outDir=assets -gen=db`) and commit it; the orchestrator regenerates on merge conflicts.

## Setup (fresh container)
```
pip install -q grpcio-tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6
PATH=$PATH:$HOME/go/bin python3 -m grpc_tools.protoc -I=./proto --go_out=./sim/core ./proto/*.proto
go build -tags with_db -o /tmp/wowsimcli ./cmd/wowsimcli
```
Tests: `go test --tags=with_db $(go list ./sim/... | grep -v sim/web) ./tools/...`. Goldens: a failing suite writes `X.results.tmp`. Check the deltas, then bless by moving each `.tmp` over its `.results` by hand. Never run `make update-tests`: it deletes every golden.

## Read first
1. `orchestration/alignment_round_2026-10-23.md`: this round's sim changes, the user decisions, and the per-class follow-up list. Your class's bullet is your core task list.
2. `orchestration/standards.md`: cross-class rulings. Gear budget is 1b, hit is 4, consumables 10/11 (natural flasks: every preset takes its best flask with the natural flask zone on), and the pre-pull rule is 13.
3. `orchestration/handoff_orchestrator.md`: the working rules for rotation, talent and race changes.
4. `orchestration/gear_targets.md`, `orchestration/stat_weights.md`, your class rows in `orchestration/spec_assumptions.md`, `orchestration/brief_<class>.md` (method and files; its gear numbers are superseded), and `forever-wiki/classes/<class>.md`.
5. `tools/dps_report/README.md`: how requests in `tools/dps_report/req/` are made from presets, and `tools/rotopt/` for rotation templates and knobs.

## Your task, for every Level 60 preset of your class
1. **Regear.** The melee/ranged hit cap is now 8% yellow / 27% dual-wield white against a level 63 boss (no hit suppression). Physical presets built to 9% now sit 1% over. Remove the excess hit from the synthetic gear and reallocate that budget by standards 1b and 5a, keeping each piece's weighted totals and shape check. Casters: check the 16% spell cap is unchanged (the 1% spell miss floor is still in). Re-check the presets' consumables, relics and enchants under the new sim (natural flask choice included).
2. **Talents and rotation.** Re-run the talent and rotation/knob searches (`tools/rotopt`) with the new gear under the new sim. Apply your class's follow-ups from the round document.
3. **Race sweep.** Best overall, Horde and Alliance. Switch a preset's race only with a lead over 0.5 DPS on both seed sets.
4. **Finals** at 120/180/300 s, 20k iterations, generic boss, on seeds 1 and 1000003. A rotation change ships only if it's positive at all three durations on both seed sets (handoff rules).
5. **Update everything that depends on it:**
   - `ui/<spec>/presets.ts`, APLs and gear sets.
   - The `tools/dps_report/req/<class>__*.json` requests: regenerate them from the presets as the README says, with `meta.published.dps` set to your 180 s finals.
   - Goldens.
   - `forever-wiki/classes/<class>.md` and the class section of `forever-wiki/sim-status.md`.
6. **Test before pushing.** The full test suite must be green.

## Report
When done, send ONE message to the orchestrator with the claude-code-remote `send_message` tool (session_id `@parent`). Keep it under 400 words:
- Branch head.
- Per preset: race, and DPS at 120/180/300 before → after.
- What changed and why.
- Any shared change or ruling you need.

Then stop. No check-ins.

## Your class's extra notes
The orchestrator gives these in your session prompt.
