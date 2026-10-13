# Handoff: Spec detail page upgrades (timeline, uptime, rotation, gear, consumes)

Do every item in $SCR/QUEUED_timeline.md. They are the user's requests, quoted there with a plan for each:
1. Timeline colours that match each spell or item icon; resource colours (mana blue, rage red, energy green, ...).
2. A "single lane" timeline mode.
3. Remove the Opener card. Put pre-pull casts on the timeline (negative time, with a clear pull marker at 0).
4. Buff and debuff uptime bars become mini timelines.
5. A much nicer Rotation card, built from the real APL, with icons, plain-English conditions and precasts.
6. A gear card modelled on the sim UI's gear picker, plus a character-stats panel.
7. A consumes and buffs card modelled on the sim UI's icon pickers.

## Context and rules (same as the previous dashboard jobs; read these first)
- Dashboard artifact: https://claude.ai/artifact/<DASHBOARD_ARTIFACT_ID> ("Dashboard" type). Its page is the files/index.html doc, currently version 16. Its data is in datasets/<id> docs pointing at uploaded JSON blobs.
- The type's rules: Artifact tool, action read, url above, paths ["SKILL.md", "artifact-type/reference/design-defaults.md"].
- Previous handoffs in $SCR (SCR=$SCR) describe the pipeline, harness, data quirks and publishing: HANDOFF_tooltips.md, HANDOFF_redesign.md, HANDOFF_dash_refresh.md, HANDOFF_races.md.
  - Pipeline: $SCR/dash (run_build.py, build_data.py, build_races.py, tooltips.py, req/, out/, data/, page/index.html = v16).
  - Harness: $SCR/dash/rd (build_harness.py, shoot.js, suite.sh, checklist.js, tips.js, tips2.js, races_check.js).
  - Merged local tree: /home/user/w-dash (branch dashdata, local only, never push).
- Data quirk: inside the real dashboard, nested arrays/objects and sometimes numbers arrive as TEXT. Use the page's rowsOf(), and wrap each card in `safe`.
- Start from page/index.html (v16) and change it incrementally. Keep every existing feature: tabs, tooltips, icons, the races work, filters.
  - Before changing anything, take baseline screenshots and run the feature checklist (rd/checklist.js and the races check).
  - After your changes, prove no regressions in both data modes, at phone 390 px light and dark, and on desktop.
  - Removing the Opener card is intended. Its information moves onto the timeline.
- Snapshot rule (user): the page shows only the final data, with no session history or process text (no "round", "before/after", "previous", PR numbers, "ruling"...).
- Publish only to the same artifact; never create a new artifact.
  - Upload changed or new data files as assets (Artifact publish, url, asset: true, one JSON file per call).
  - Create or update their datasets/<id> docs with updated.by = "Claude (forever-sim)", and read each doc's version first.
  - Write files/index.html with ArtifactData set, if_version pinned (currently 16).
- No model names in files. Don't message other sessions. Never push or touch /home/user/forever-sim; local commits on dashdata are fine.
- Page size must stay under 256 KB (now about 182 KB). If you get close, move static data (e.g. humanised rotation text, icon colours) into datasets.

## Deliverable
Reply briefly with:
- what changed per item;
- anything you couldn't do, e.g. APL node types you couldn't humanise;
- the checklist result;
- 5-6 phone screenshot paths: timeline lanes, timeline single lane, uptime mini timelines, rotation, gear, consumes;
- the new files/index.html version.
