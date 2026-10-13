# Handoff: assumption scenarios on the dashboard ("Maelstrom Weapon PPM low / expected / high", etc.)

## The user's request
- Original wording: "add a section about each assumption we made about each spec. E.g. warlock pet scaling, Maelstrom Weapon PPM. For each, provide multiple reasonable settings for what the actual effect will be on release ... Then, we can select different configurations to see their effect. In the rankings tab, those should be top-level settings, categorized per class, above the dps ranking widget. Remember user selection for each in cookie".
- They've now asked where it is, so it's wanted now.

## Plan and inputs
- Plan: /home/user/forever-sim/orchestration/dashboard_next_round.md, steps 2-5 (git pull branch claude/wow-forever-dps-presets-m9u15h first).
- Inventory (step 1, done): /home/user/forever-sim/orchestration/spec_assumptions.md. 31 assumptions, each with:
  - key, builds, "Sim now", Low / Expected / High values and rationale;
  - sensitivity;
  - where the constant lives and how to make it a knob.
- Settled items, not scenarios: Windfury Totem ICD 1.5 s, Rage of the Farseer 3 min, Touch of the Grave.
- hunter_tsa_hm_rank: Low = sim now, because the strongest-rank change still waits on the user.
- Druid form weapon scaling is being implemented as a real fix by the druid session, so it is NOT a scenario.

## Steps
1. Knobs. In the merged local tree /home/user/w-dash (branch dashdata, local only), add one dashboard-only override mechanism. For example, an env var DASH_ASSUME="key=value,key=value" parsed once in sim/core, with a small helper each constant site calls: `core.Assume("sham_mw_ppm", 10.0)` returns the override or the default.
   - Wire every assumption in spec_assumptions.md at the file:line it names, with defaults EXACTLY equal to today's values.
   - Prove it: with no env var, a sim of every build gives bit-identical DPS (same seed).
   - Commit locally on dashdata, never push. Rebuild wowsimcli (with_db) into $SCR.
   - Before rebuilding, merge the latest class heads (git fetch origin; merge each claude/wow-forever-dps-presets-m9u15h-<class>; resolve as earlier handoffs describe).
2. Runs. For every build that an assumption touches, and every setting other than "Sim now", sim the build with only that one knob changed: 180 s, 20k iterations, seed 1, the build's preset race and request ($SCR/dash/req/<id>.json, regenerated if the class heads changed; see HANDOFF_dash_refresh.md).
   - Also run 120 s and 300 s at 5k iterations, so the fight-length view can use them.
   - Combinations: where a class has 3 or fewer assumptions touching a build, sim the full grid; otherwise store per-assumption deltas and combine them multiplicatively. Spot-check about 10 combos against real sims and report the error.
   - Write a script ($SCR/dash/run_assumptions.py) the full dashboard refresh can re-run, and make build_data.py emit the datasets.
3. Data:
   - `assumptions`: class, key, label, description (plain words), builds, settings [{id: low / expected / high / now, label, value text, rationale}], defaultSetting = the one equal to "Sim now", expectedSetting.
   - `assumption_dps`: build, key (or combo key), setting(s), dps120, dps180, dps300, se.
   Snapshot rule: no session history, no "shipped".
4. Page (Ranking tab).
   - Above the DPS ranking widget, an "Assumptions" panel grouped by class: collapsible per class, each assumption a segmented control (Low / Expected / High), with "Sim default" marked and an ⓘ tooltip (description, values, rationale, sensitivity).
   - The ranking, the headline tiles, the Specs page headline DPS and Compare recompute from the selections.
   - A visible "Custom assumptions" badge and a "Reset" button appear when anything differs from the defaults.
   - The Method tab lists every assumption.
   - Persistence: the user asked for a cookie. Check the dashboard type's rules (Artifact read, url below, paths ["SKILL.md"]) for storage. The page already stores some per-viewer state in localStorage behind try/catch; use the same, mirror to the URL hash if allowed, and tell me what you used.
5. Dashboard: https://claude.ai/artifact/<DASHBOARD_ARTIFACT_ID>.
   - The page doc files/index.html is edited by several jobs. Another job is currently converting tooltips to Forever Wowhead text.
   - Do all page work LAST: read the live files/index.html, merge your changes into it, run the rd/ regression checks (checklist2.js, races_check.js, specswitch.js, tips2.js, setup_check.js) in plain, text and ASYNC=1 modes plus a new check for the panel, then write with if_version pinned to the version you read. If the write is refused, re-read, re-merge and retry.
   - Upload data files as assets and create or update the datasets docs (updated.by "Claude (forever-sim)").

## Rules
- SCR=$SCR. Earlier handoffs there describe the pipeline, harness and publishing: HANDOFF_dash_refresh.md, HANDOFF_specpage.md, HANDOFF_setup_active.md.
- No model names in files. Don't message other sessions. Never push or touch /home/user/forever-sim's working tree (git pull there is fine, read only).
- No new artifacts.

## Deliverable
Reply briefly with:
- the knobs wired, and any you couldn't make;
- the measured sensitivities per assumption (Low / High at 180 s);
- the combination error;
- the persistence used;
- the checks result;
- 2 phone screenshots of the panel;
- the new files/index.html version.
