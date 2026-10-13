# Handoff: redesign the Forever DPS dashboard (clearer, cleaner, more interactive, more modern)

## The user's request (verbatim)
"make the presets artifact more sexy with better, cleaner design. Take screenshots and find ways to improve. Focus on clarity, interactivity, modern look, use textures where it makes sense."
They read it on a PHONE. Phone first, but it must also look good on desktop.

## The artifact
- URL: https://claude.ai/artifact/<DASHBOARD_ARTIFACT_ID>, a "Dashboard" type artifact.
- Its page is the doc files/index.html ({text: <page>}) in the artifact's store. Write it with the ArtifactData tool (load via ToolSearch "select:ArtifactData"). Read the doc first to learn its version, then pin if_version.
- Data: datasets/<id> docs pointing at uploaded JSON blobs. These are builds, races, lengths, abilities, auras, hist, stats, gear, setup, notes, names, talenttrees, treeart, icons (36 px icons as data URLs), weights, tooltips, and timeline_<class> for 9 classes.
- Read the type's rules before you start: Artifact tool, action read, url above, paths ["SKILL.md", "artifact-type/reference/design-defaults.md"] (and any other reference files SKILL.md points to). Follow them:
  - every figure on the page comes from a dataset;
  - marks (data-source / data-field / data-where) go on values;
  - scripts get the global `dash` (dash.data(id), dash.onData(fn), dash.calc, dash.colors, dash.setLink);
  - d3 v7 is global;
  - no network: images only as data: URLs, fonts as the type allows;
  - the page file stays under 256 KB.
  The type's instructions also say not to render or screenshot the PUBLISHED dashboard. All screenshots are LOCAL, with the harness below.
- Also load the `artifact-design` and `dataviz` skills (Skill tool) for design and chart guidance, and apply what fits inside the dashboard type's rules.

## Local working copy and harness
- Page source: $SCR/dash/page/index.html, about 123 KB. SCR=$SCR.
  Work on a copy, $SCR/dash/page/index_v2.html, and only replace index.html at the end.
- Data: $SCR/dash/data/*.json. A separate job is regenerating these files with new sim results right now, and will upload them to the artifact itself. It has been told not to touch index.html. Your page must work with whatever data is there, now and after the refresh. The data SHAPE won't change.
- Harness: HANDOFF_tooltips.md in $SCR (section "Local working copy") explains it. In short:
  - $SCR/dash/page/make_tt_harness.py and h_str.html show how a mock `dash` serves the data files, with nested values stringified like the real dashboard does.
  - Build both plain and stringified variants.
- IMPORTANT, learned the hard way: inside the real dashboard, nested arrays/objects and sometimes numbers in dataset rows arrive as TEXT.
  - The page normalises rows in rowsOf(); keep using it.
  - Wrap each card's draw in the `safe` helper so one failure can't leave stale content.
  - `[hidden]` needs `.fv [hidden]{display:none!important}`.
  - Use `.style('fill', ...)` not `.attr` for CSS variables in SVG.
- Screenshots: use Playwright (node or python), already installed. Use executablePath /opt/pw-browsers/chromium-1194/chrome-linux/chrome if needed; never `playwright install`.
  - iPhone-like viewport 390x844 at deviceScaleFactor 2 or 3, and desktop 1280x900.
  - Light and dark (colorScheme).
  - Full-page shots of each tab, plus key interactions (tooltip open, spec drilldown, timeline, talent tree, compare).
  Look at your screenshots with the Read tool, critique them honestly like a picky product designer, fix, and repeat. Do several rounds.

## What to improve (examples; use your judgement)
Clarity:
- a clear visual hierarchy, and a hero that answers "which spec is best?" at a glance;
- consistent spacing, type scale and number formatting;
- fewer competing colours; legends where needed;
- readable at 375 px with no horizontal page scroll (wide charts may scroll inside their card).
Interactivity:
- smooth tab and drilldown transitions, sticky tab bar, filter chips (class, role, faction, fight length);
- sort toggles, tap-to-highlight in charts, a spec picker that's fast on a phone;
- swipe or segmented controls where natural, and state remembered in the URL hash or localStorage (try/catch).
Modern look:
- refined cards, subtle depth, good dark mode, class colours used with restraint;
- game-flavoured textures where they help, kept tasteful and subtle: parchment/stone/leather, the talent tree backgrounds (treeart), the class crests and spell icons already in the data;
- textures as CSS gradients/noise or small data URLs. Keep the page under 256 KB: generate noise with CSS or an inline SVG filter rather than embedding big images.
Keep every existing feature working:
- Ranking, Races, Spec detail (DPS spread, timelines, ability and aura tables, gear, setup, talents, stat weights), Compare, Method;
- tooltips everywhere;
- class crests and talent trees.

## Publish
When you're happy:
1. Copy index_v2.html to index.html.
2. Test both harness modes once more with zero script errors.
3. Write {"text": <page>} to a JSON file and call ArtifactData set on collection "files", doc_id "index.html", with file_path and if_version.
4. Read it back and confirm it matches.
If $SCR/dash/method_text_1b.md exists when you publish (the data job writes it), put its content in the Method tab, marked up to fit your design. Otherwise keep the Method tab's text and say so.

## Rules
- No model names in any file. Don't message other sessions, and don't touch git or /home/user/forever-sim.
- Don't upload or alter dataset docs or data files: the other job owns them. If you truly need a new data field, tell me instead.

## Deliverable
Reply briefly:
- what you changed and why, as a short before/after list;
- the screenshot file paths of the final version (phone light, phone dark, desktop) so I can show the user;
- the new files/index.html version;
- anything you couldn't do.
