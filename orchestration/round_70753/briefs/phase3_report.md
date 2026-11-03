# Round 70753, phase 3: regenerate and deploy the DPS report

All nine class workers of round 70753 (the November 2 development notes, client build 1.60.1.70753) landed on master.
You regenerate the DPS report from it and deploy it. Read `orchestration/round_70753/README.md` (the round, its
"Phase 3 to-do" list), each class's final summary in `orchestration/round_70753/<class>.md`, then follow the repo skill
`.claude/skills/regenerate-dps-report/SKILL.md` and `tools/dps_report/README.md`. The dps-report skill's B7, B9 and B11
(`.claude/skills/dps-report/SKILL.md`) say what this pass owes.

1. **Shared cleanup first** (one commit each, on a branch from origin/master):
   - Blessing of Salvation is gone in game: drop it from `tools/rotopt/mksettings/main.go` (line ~74) and from every
     Level 60 request's raid buffs (regenerate the requests the way the README says; check with `verify.py` that only
     the Salvation line moved). It is threat only, 0 DPS: confirm on two requests.
   - The assumption rows the round README lists under "Phase 3 to-do" (B9): shaman `sham_lb_31plus` and
     `sham_es7_tuning` out (now 1.0), `sham_mw_ppm` to 20, Chain Lightning's coefficient row if there is one; mage
     `mage_ice_lance_cd` to one value; hunter `hunter_tsa_hm_rank` to Trueshot Aura 100, `hunter_shot_ap_shares` to 15% / 5% and the stale caveat numbers in the hunter metas; paladin `pal_2h_spec_holy`
     wording (Two-Handed Weapon Specialization covers all Holy damage). Also, each class's round file may list more;
     read their "Open questions" and apply what is an assumption-row change, leave the rest.
   - New build `warlock__affliction_wrack` (the warlock worker added its request): add it to the report the way the
     skill's "Add a preset" says (DISPLAY, assumptions_def.py lists). Check every class file for other added or removed
     builds. Known: the user dropped the six rogue "(vs Humanoid)" builds (their requests, presets and report_hashes
     lines are gone; remove them from ui/dps_report/data, provenance.json and ui/raid_sim/data/builds.json too); rogue's
     class default is now Combat Swords and warrior's is two-hand Fury (check the report's defaults follow).
   -  still has the old warrior strings: bring every class's rows there up to the round.
2. **Full regeneration.** Race base stats (sim/core) changed, so every build is stale: `pipeline.sh run` then
   `pipeline.sh publish` with no hydrate (about 95 min on 4 cores; run it in the background and poll its log with short
   checks, not sleeps). Check each build's 180 s number against its class round file (within the sim's noise; a bigger
   gap is a question, not something to paper over). Then B9: the assumption scenarios for every build.
3. **Changes tab** (standards 30): the live update is `2026-11-02-build-70753`. Write its public lines from the round
   files, in plain game terms (`tools/dps_report/changes.py`'s rules: no commit ids, dates, session names, file paths),
   and make `python3 tools/dps_report/changes.py check` list nothing uncovered (internal entries with a reason for
   commits that move no build, such as the healing, tank and practice-room work).
4. **Checks and land.** Every check in the dps-report skill's B8 (Playwright on the wrapped dist file at phone and desktop
   width, both themes, no console errors). Merge origin/master last; commit `ui/dps_report/data`, `datasets.json`,
   `provenance.json` and `tools/dps_report/raw` together; push to master (the repo's CLAUDE.md); if master changed the
   report data meanwhile, follow the README (take master's, hydrate, re-check stale, re-sim, publish again). Then deploy
   from the pushed commit: `make dps-report-deploy` (api.cloudflare.com is authenticated by this environment's proxy).
5. **State** (B11): update `.agents/skills/dps-report/state.json` and `.claude/skills/dps-report/state.json` (client
   build, report data commit, deploy, date, a note on this round, open items).

The user asked the mage and rogue workers for exploratory follow-ups: Arcane already landed again (94b59b37f), and a
rogue search is still running and may land a new build. Other sessions may still land fixes from the audit pass on master while you run. If one moves a build after your run,
re-sim just those builds (`hydrate` + `run <ids>` + `publish <ids>`) before you land, and say so.

Rules: never force-push; end every commit message with the trailer lines your session prompt specifies; no model names in
files. When deployed, write a short summary at `orchestration/round_70753/report.md` (the ranking at 180 s, every
build's old and new number, race changes, the deploy id), push it, and send one message to `@parent` with
`send_message` (claude-code-remote MCP) giving the master commit, the deploy and the top of the ranking. Then stop.
