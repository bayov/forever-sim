# Orchestrator handoff (2026-10-17)

The long-running DPS report orchestrator session was retired on 2026-10-17, along with the nine class workers. The
auditors moved to scheduled fresh-session runs (`orchestration/audit/`, `tools/audit/`). This file holds what that session
knew that wasn't already in the repo. Read it after `.agents/skills/dps-report/SKILL.md` and `state.json`.

> 2026-10-26: the pipeline notes below predate the raw packs. A regeneration now re-sims only what changed (`pipeline.sh hydrate`, `run <ids>`, `publish <ids>`, tools/dps_report/README.md), and `publish` runs `events.py` itself.

## State at handoff

- **Master:** d2633f4a4. b389688f7 came after it, and it only touches the Wowhead scraper.
- **Report:** artifact version 35, https://claude.ai/artifact/RUzPysb5dSk2RAXzXij4oR. Published from 437b92161. (Since 2026-10-25 the report is deployed to https://forever-dps.pages.dev instead; the artifact is retired.)
  - All 66 builds were re-simmed on 787da848b. Provenance is recorded at 1735382cd, because the only commits in between
    are the Leveling Plan page and its `ui.proto` message, which no sim reads.
  - `node tools/dps_report/stale.mjs` showed 0 of 66 builds stale at handoff.
- **PRs:** none open. #114 (Enhancement Level 60 talents) was the last one, merged at 787da848b.
- **Class and fix branches:** no branch is ahead of master, except the coordination branch.
- **Coordination branch:** `claude/wow-forever-dps-presets-m9u15h` is 9 commits ahead of master (head 1e1d201af).
  - All 9 are `orchestration/rank_tuning.md` updates from the rank tuning auditor, documentation only.
  - Merge them to master when convenient. The rank tuning auditor's latest findings live only there.
- **Last audit passes:** every auditor reported nothing new on master d2633f4:
  - results, pass 42;
  - fairness, at 1735382;
  - rotations, pass 26;
  - consumables, pass 32;
  - mechanics, pass 27;
  - rank tuning, 03:25 and 09:25 checks;
  - gear, pass 36.
- **Forever client:** 1.60.1.70450, which is cosmetic against 70425 (ledger #113). 70425 remains the spell data reference.
  The mechanics auditor probed builds 70451–70700 and all returned 404.

## Open threads

Nothing was in flight. These are parked items with no owner now:

1. **The Ice Lance Low range (user's call, never asked).** The fairness auditor kept it open. The `mage_ice_lance_coef`
   assumption's Low setting is "None (client)": no spell power, and the rotation drops Ice Lance. The open question is
   whether that range stays as it is. Raise it with the user if mage assumptions come up.
2. **Hunter end-of-fight gate (parked).** Rotation pass 19c found it worth +0.7 at 180 s, under the 1 DPS bar. It was
   parked unless the user asks. The Mutilate window tweak (+0.8) is parked the same way.
3. **Hunter Auto Shot sim bug (shared code, left alone).** If a hunter ends its pre-pull walk at range, Auto Shot keeps
   firing after it walks into melee, and melee swings never come back. The Survival Scorpid duty avoids it by walking
   into melee before the pull. The 12 yd vs 8 yd ranged minimum check is unresolved too. Update 2026-10-23: the
   minimum is 11.67 yd under Forever (the 8 yd row from both combat reaches, wowsims/forever 8a192337d8), and the
   weaving presets walk out of melee and back many times a fight with Auto Shot and the swings coming back each
   time (sim/hunter/forever_weave_test.go); the bug as described didn't reproduce.
4. **`state.json` openItems** are still current: Gnome Eureka! spell mask, APL duration × fractional constant
   truncation, beta Rage and Shadow Oil measurements, and Smite's consumes from mksettings.
5. **Rank tuning effects after the Enhancement rise.** Recorded only in the auditor's last message, so here:
   - Ret twist at −1.4% (876.9) now falls below Imp GW Enhancement too.
   - Earth Shock ×1.16 (901.4) ties Ret with Shock (901.7) within noise.
   - Frost at +1.5% no longer passes any Enhancement preset.

Nothing was promised to the user beyond the published v35.

## Rules the old session applied that aren't in standards.md

These were working rules. They are not user rulings, but every merged class PR followed them:

- **A rotation change** ships only if it's positive at 120, 180 and 300 s on two seed sets: seed 1 and seed 1000003.
  Iteration i uses seed + i, paired at 20,000 iterations.
  So a second seed set must start more than the iteration count away: seeds 1 and 2 share all but one iteration and
  give nearly the same average by construction (checked 2026-10-23: the CLI honours the seed; twist paladin at 2,000
  iterations reads 935.18, 935.14 and 935.19 on seeds 1, 2 and 3, and 935.01 on 1000003).
- **Talents** are judged at 180 s. A build that loses at 120 but wins at 180 and 300 ships, as in Enhancement #114.
- **The class default preset** is the one highest at 180 s.
- **A race switch** needs a lead of more than 0.5 DPS on both seed sets.
- **Report-only changes** (`ui/dps_report`, `tools/dps_report`, docs) go straight to master. Sim, preset and class changes
  go through PRs that the orchestrator merges after verifying them. User ruling, 2026-10-14.

## How a PR was verified before merging

1. Fetch the PR: `git fetch origin pull/N/head:prN`. Confirm the diff only touches what the PR claims.
2. Merge it onto origin/master in a scratch worktree.
   - Don't put the worktree in a `mktemp -d` directory named `tmp.*`: `tsc` stops with "No inputs were found in config
     file" there.
3. Regenerate the protos: `protoc -I=./proto --go_out=./sim/core ./proto/*.proto`, with `$(go env GOPATH)/bin` on PATH.
4. Run the tests and the type check:
   - `go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) ./tools/...`
   - `npx tsc --noEmit -p .`
5. For every changed `tools/dps_report/req/*.json`:
   - set `"iterations"` to 20000;
   - sim it with a `wowsimcli` built from the merge (`go build --tags=with_db ./cmd/wowsimcli`);
   - compare with `meta.published.dps`. It must agree within 2× the standard error.
6. Merge with the PR head's full 40-character SHA as `expectedHeadSha`.

**`assets/db.bin` conflicts:** regenerate with `go run tools/database/gen_db/*.go -outDir=./assets -gen=db` and push a
merge commit to the worker's branch. Never rebase someone else's branch.

**Request JSON:** add fields to `req/*.json` textually, for example with a Python string replace. Loading and
re-serializing the JSON reformats the whole file and buries the real diff.

## Report pipeline lessons (supplementing the regenerate-dps-report skill)

- **Dash tree.**
  - Keep one worktree with `patches/0001–0003` applied as commits, plus 0004 as a commit.
  - To refresh it, discard the uncommitted DB regeneration first (`git checkout -- assets`), then
    `git rebase origin/master`.
  - Then regenerate the protos and the DB and rebuild the binaries into a new bin directory.
  - Make a new `DASH_DIR`, copying `icons/`, `names/`, `fortt_*.json` and the `forever_tt_cache` link from the last one,
    plus a fresh `req/` from the tree.
- **Time.**
  - A full 66-build `pipeline.sh run` took 108 min on 4 cores (the estimate says 95):
    - `run_build` 23 min;
    - weights 29;
    - races 15;
    - mana 6;
    - ablate 4;
    - assumptions 30.
  - Background shell jobs are killed at 2 h, so start it with the maximum timeout.
  - If it's killed, `run_races.py` and `run_assumptions.py` resume where they stopped. Re-run the other steps only for the
    builds they hadn't finished.
- **Weights.**
  - `weights.py` re-runs a build only when its weights file is older than its request. `WEIGHTS_FORCE=1` re-runs them all.
  - Each row records its own `iters`: 5000, or 40000 when its paired error was worse than the old method's.
  - On v35, 147 of 311 rows stopped at 5000. The weights agreed with v32 within 3 standard errors, except where a rotation
    had changed.
- **events.py.**
  - It must run after `build_data.py` and before `merge.mjs`. v31 and v32 shipped without it once and had to be
    republished.
  - "no typical fight in the dataset" for the two warrior `-duty-battle-shout` builds is expected: duty presets have no
    timelines.
- **Provenance.**
  - Record at the commit the sims ran on.
  - If master has since gained only commits no sim reads (UI pages, a `ui.proto` message for the UI), recording at HEAD is
    fine, and keeps the next `stale.mjs` run from flagging every build. Say so in the commit message, as 437b92161 does.
- **Artifact publish conflicts.** Other sessions publish the same artifact too.
  - On a conflict, read the live version.
  - Check that your build contains its changes; only the data should differ, and master has the page source.
  - Then republish from the same file path.
- **Paths.**
  - The old session built the artifact in a scratch worktree of master (`wt-rep`) and published `dist/dps_report.html`
    from there.
  - The browser check wraps that file in a minimal HTML document and clicks every tab at 390 and 1280 px.

## Who was who (for reading old transcripts)

These sessions hold the history behind each class's choices. Their scratchpads are gone once they're archived, so read
their transcripts with `list_events` if a question comes up.

- **Class workers:** shaman session_01Pe2HfoGHLpExHKyZsSE43D, druid session_012e53DVbJdcEwgUK3m15hrz. The other seven
  are in this orchestrator's transcript, session_01AGmMcVxGoubbDodiQDQwhz.
- **Auditors:**
  - fairness session_01QnE9gpZ5kp8fwRANBsMoqP;
  - results session_01X9dJY1kQbnkLdLwRbqh6GT;
  - mechanics session_016BhZyigQBmZb3KuabGUP88;
  - rank tuning session_017sYpPk8z2WHj1cXtiJKrL6;
  - rotations session_01TACD9s7qZpSiUwEWmDHEHo;
  - consumables session_01GteY69D2ZNZnHnzGwb3qPu;
  - gear session_01Q6QNZiT7pTQwT7t9yKvWMo.
