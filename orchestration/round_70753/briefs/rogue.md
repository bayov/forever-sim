You are the **Rogue worker** in a full pass of forever-sim (github.com/bayov/forever-sim, a wowsims/classic fork for WoW Forever, level 60 DPS presets and the DPS report at https://forever-dps.pages.dev). The orchestrator is the session that created you. Blizzard published a very large update: the **November 2 development notes**, which go with client build **1.60.1.70753**. They raise damage for classes across the game and rework many talent trees. The full text is in the repo at `orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt`. Your class's section is "- Rogue" under "Classes", and the "Races" and "Bug Fixes" sections matter too. Your job: get the Rogue DPS specs fully up to date with it, re-optimize every Rogue preset, and audit them.

Read first:
1. `orchestration/standards.md`: cross-class rulings that override everything older.
2. `orchestration/brief_rogue.md` (method, files; its gear numbers are superseded by standards 1b), `orchestration/gear_targets.md` and `orchestration/spec_assumptions.md` (your rows), `orchestration/forever_class_changes.md` and `orchestration/update_ledger.md` (your class), `forever-wiki/classes/rogue.md`, and `orchestration/handoff_orchestrator.md` (the working rules for rotation, talent and race changes).
3. The open findings for your class in `orchestration/audit/findings/*.md`.
4. Your class's report builds: `tools/dps_report/req/rogue__*.json` (12 builds), with `tools/dps_report/README.md`, and the optimizer `tools/rotopt/` (your `rogue_*.go` templates and `rogue_presets.sh`).

Who owns what (several agents and sessions push to master at once):
- **Agent A1** ports client build 70753: the client tables and their diff (it commits a copy under `orchestration/audit/ref/client_update/`), the talent trees for every class (it remaps every preset's talent string to the new tree positions and lists the points that removed talents freed in `orchestration/talent_remap_70753.md`), the ability text and the item DB.
- **Agent A2** owns the shared sim/core changes: racials, race base stats, Throw's windup, and the shared part of "talents no longer have a low-rank proc penalty".
- **You** own sim/rogue/, ui/rogue/ (only the DPS presets), tools/rotopt/rogue_* and rogue_presets.sh, tools/dps_report/req/rogue__*, and forever-wiki/classes/rogue.md.
- **Live sessions** "healing" and "tanking" own the healer and tank specs and their sim files. Don't edit healer or tank files. Threat and healing rows are n/a for you; list them in your ledger as "tank/healer sessions".
- **The orchestrator** owns `ui/dps_report/data`, `tools/dps_report/raw`, `provenance.json`, the Changes tab and the report deploy. Don't touch those; the orchestrator regenerates the report after every class lands.
- **Any other shared change** (sim/core, sim/common, proto, tools/database, tools/rotopt shared files): keep it as small as possible and gated (`Env.IsForever()` plus a class-only flag or caller), and name it in your report. If it can't be gated, don't make it; describe it in your report.

**Phase A (start now): research and mechanics.**
- For every line of your class's notes section (and Poised, Shank, racial or bug-fix lines that hit your class), check what the sim does now. The 10-30 launch-realm hotfix already landed many reworks, so some lines may already be in.
- Check the 70753 client data on wago.tools (`python3 tools/sources/wago.py fetch 1.60.1.70753 --set client_update`; the CSVs come from https://wago.tools/db2/<Table>/csv?build=1.60.1.70753) and any newer server hotfix (wago.tools/hotfixes, foreverchanges.pro/patch-notes).
- Source tiers, highest first: server hotfixes > client data (newest build) > Wowhead Forever (tooltip API, dataEnv 16) > Blizzard dev notes > foreverchanges.pro > forever-wiki. Where the client or a hotfix disagrees with the notes, the higher tier wins. Note the disagreement. An example: hunter Trueshot Aura was hotfixed to 15/22/30/40/50 on 10-28, and the notes say 20/35/50/75/100.
- Write `orchestration/round_70753/rogue.md` with a ledger table: note line | tier and evidence | sim status (landed / missing / wrong / partial / n/a) | fix.
- Then fix the sim's mechanics with a test per rule. Do everything that doesn't depend on the new talent trees now.
- Don't regenerate talent trees or edit preset talent strings in Phase A: A1 is doing that. Work on branch `round-70753-rogue` from origin/master and push it often.
- When Phase A is done, push the branch and end your turn. The orchestrator messages you "sources landed" once A1 and A2 are on master.
- If you see on master that `orchestration/talent_remap_70753.md` exists and `.agents/skills/dps-report/state.json` says foreverClientBuild 1.60.1.70753, the sources have landed: go straight on with Phase B.

Class-specific pointers (check them, they're not orders): a near-total rework.
- New baseline abilities:
  - Shank (level 24 and 44): 5 Energy, 10 s cooldown, off-hand 50% weapon damage, +15/+25 Energy, can't miss.
  - Poised (level 18): open finding client_update-20261103-1, yours for rogues. Model it the same way as the druid worker: a one-charge aura at the pull, −10 Energy on the next ability, only if combat starts at max Energy.
- Deadly Poison AP coefficient 0.009 and Instant Poison 0.01 (doubled).
- Assassination: Murder removed. Lethality 2 points 10/20%. Vigor row 3. Blood Splatter (Rupture +6 s, Garrote +9 s). Relentless Strikes 5 points, 4–20% a combo point for 25 Energy. Venom is now a passive (Shank → +30% poison damage and +20% poison chance for 6 s). Seal Fate 3 points 33/66/100%. Mutilate row 7, 90% weapon damage plus more. Improved Expose Armor moved (standards 18: Sunder, never Expose).
- Combat: Coup de Grace (was Improved Eviscerate) 7/13/20% Eviscerate and Death Mark. Efficient Strikes −3/5 Sinister Strike and Hemorrhage. Composure (+5/10 Poised). Flawless Execution 15. Aggression 5/10/15% Sinister Strike and Eviscerate (no Backstab). Dual Wield Specialization 2 points 25/50% off-hand damage. Flash of Steel (Shank +3–15 Energy, −0.4–2.0 s cooldown). Weapon Expertise and Aggression row 4.
- Subtlety: Hemorrhage and Ghostly Strike's dagger bug fixed. Opportunity 5 points 4–20% (Backstab, Garrote, Ambush, Mutilate). Ghostly Strike lets the next Ambush or Garrote skip Stealth. Hemorrhage now buffs Rupture and Garrote 15%. Quietus, Cutthroat and Thousand Cuts removed. Boot Knife (Shank adds 2 combo points, 20–100%). Death Mark: a 31-point active Curse finisher that stores damage (twice from Garrote and Rupture) and deals it after 6 s, capped by AP and crit; model it carefully, from the client data.
- With this much change, consider whether each of the current builds (Combat Swords/Daggers, Mutilate, Seal Fate and the humanoid variants) still makes sense, and whether a new one (Shank or Death Mark Subtlety) should replace one. Put that in your report as a recommendation for the user. Don't remove a build without a ruling; adding one is fine.

**Phase B (after "sources landed"): re-optimize and audit.**
1. `git merge origin/master`. Read `orchestration/talent_remap_70753.md` for your class, then finish the talent tree mechanics: new, moved and changed talents.
2. Re-optimize each preset per the brief and handoff rules: the talent search and every rotation knob with rotopt, with gear fixed to the synthetic set; new abilities into the rotation; removed ones out. Then the finals at 120 / 180 / 300 s (20k iterations, generic boss). Then the race sweep (best overall / Horde / Alliance; standards 29: each race's own rotation and gear in tools/dps_report/req/races/), and re-pick the preset race if it changed beyond noise.
3. Regenerate your report requests from the presets (tools/dps_report/README.md) and check each one's 180 s number. Update presets.ts, goldens (`make update-tests`, and check that only intended values moved), the wiki, and the ledger file.
4. **Audit your class** before landing. For each auditor in `orchestration/audit/passes/` that covers DPS (results, rotations, gear, consumables, mechanics, client_update; fairness is cross-class, so skip it), read its pass file and apply its checks to your class's builds. File what you find and fix in your round file. Standards 31: rotation lines state their intent, checked at 120/180/300 s and 180±20.
5. Land it. Run `go build ./...` and `go test --tags=with_db $(go list ./sim/... | grep -v /sim/web)` (a fresh clone first needs `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`). Merge origin/master last, re-run the tests, then `git push origin HEAD:master` (the repo's CLAUDE.md: finished work goes into master without a PR). Also push your branch.
   - If the push is rejected: fetch, merge, re-test, push again.
   - On conflicts in `assets/database/db.json` / `db.bin`, regenerate them with `go run ./tools/database/gen_db -outDir=assets -gen=db`; never hand-merge them.

Rules:
- Never force-push. Commit only to your branch and master.
- End every commit message with exactly these two lines:
  Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
  Claude-Session: https://claude.ai/code/session_01TXnjyRFPyrcjiD1fa7XFBV
- Don't put model names in files.
- Keep sim output and long diffs in files, and read summaries.
- Cross-class judgement calls (for example a rule that would change the standards) go in your report as a question with your recommendation. Don't decide them silently.

**Report.** When Phase B is landed, write the final summary at the top of `orchestration/round_70753/rogue.md` and push it to master. Include:
- what changed, with its source and DPS effect;
- per build: old and new DPS at 120/180/300 s, talents and race;
- best race overall / Horde / Alliance;
- the audit results;
- open questions and shared changes.
Then send the orchestrator one message with `send_message` to `@parent` (claude-code-remote MCP; if that tool is missing, the file is enough) saying the class is landed and giving the master commit. Then stop.
