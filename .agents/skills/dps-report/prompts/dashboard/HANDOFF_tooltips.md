# Handoff: add game-style tooltips to the Forever DPS dashboard artifact

## The artifact
- URL: https://claude.ai/artifact/<DASHBOARD_ARTIFACT_ID> (a "Dashboard" type artifact; the user views it on a phone).
- Its content lives in the artifact's own store, written with the ArtifactData tool (load it with ToolSearch "select:ArtifactData"):
  - files/index.html  {text: <the page>}  (current version 7; every write to an existing doc needs if_version = the version you last read; read it first with action get to learn the version)
  - dash/meta, and datasets/<id> docs: {title, description, source: {kind: "file", url: "/_blob/<id>", name, format: "json"}, updated: {at, by}}
- Data files are uploaded with the Artifact tool: action publish, url = the artifact URL, asset: true, file_path = a local JSON file (one JSON file per call). The reply gives "/_blob/<id>"; then write/update the datasets/<id> doc to point at it.
- The dashboard type's rules (read once): Artifact action read, url above, paths ["SKILL.md", "artifact-type/reference/design-defaults.md"]. Key rules: every figure on the page comes from a dataset; marks (data-source / data-field / data-where) on values; scripts get the global `dash` (dash.data(id), dash.onData(fn), dash.calc, dash.colors, dash.setLink); d3 v7 is global; no network; images only as data: URLs; the page file must stay under 256 KB (it is ~108 KB now); each dataset file under 8 MB.
- IMPORTANT, learned the hard way: inside the real dashboard, nested arrays/objects and sometimes numbers in dataset rows arrive as TEXT. The page already normalises rows in rowsOf() (normVal: JSON-parses strings starting with [ or {, numbers, booleans). Use rowsOf() for any new dataset, or parse yourself. Also wrap each card's draw in try/catch (the page's `safe` helper) so one failure can't leave stale content.
- Also learned: `[hidden]` needs `.fv [hidden]{display:none!important}` (already in the page); `.style('fill', ...)` not `.attr('fill', ...)` for CSS variables in SVG.
- The type's instructions say not to render/screenshot the published dashboard. Testing your code LOCALLY in a headless browser against the real data files is fine and has caught real bugs: see the harness below.

## Local working copy
- Page source: $SCR/dash/page/index.html (edit this, then publish it as files/index.html: write {"text": <page>} to a JSON file and use ArtifactData set with file_path and if_version).
- Data builder: .../scratchpad/dash/build_data.py (reads .../dash/out/*.json and req/*.meta.json, writes .../dash/data/*.json). Datasets: builds, races, lengths, abilities, auras, hist, stats, gear, setup, notes, names, talenttrees, treeart, icons, weights, timeline_<class> (9).
  names.json rows: {key, name, rank, icon} where key is s<spellId>[t<tag>] / i<itemId> / oOtherAction...; abilities/auras rows carry `key`; gear rows carry itemId; talenttrees rows carry name, description (with {0} placeholders), ranks.
- Local headless test harness: .../dash/page/h_str.html shows the pattern (it inlines index.html into a mock page with a mock `dash` whose data() returns rows stringified like the real dashboard does, loads d3.min.js and bundle.js = all data files). Rebuild bundle.js from ../data/*.json, rebuild the harness with the new index.html, then run:
  /opt/pw-browsers/chromium-1194/chrome-linux/chrome --headless=new --no-sandbox --disable-gpu --allow-file-access-from-files --window-size=420,1000 --virtual-time-budget=6000 --dump-dom file://.../harness.html
  (or --screenshot=x.png; note the headless window is at least 500 px wide, so screenshots crop). Put a DIAG ... ENDDIAG block in the page via script to read results; window.onerror collects errors.
- Wide data sources for tooltip text:
  - /home/user/w-dash/assets/db_inputs/wowhead_spell_tooltips.csv: id,json with "name", "tooltip" (HTML), "buff" (HTML). Classic spell ids.
  - /home/user/w-dash/assets/db_inputs/wowhead_item_tooltips.csv: id,json with "name","quality","icon","tooltip" (HTML). Classic items.
  - /home/user/w-dash/assets/database/db.json: items (incl. our synthetic items, ids 990000+ ... 998xxx, with stats array, ilvl, quality, setName, hitRating/critRating) and enchants (effectId -> name, stats).
  - /home/user/w-dash/forever-wiki/data/abilities.json: Forever ability descriptions by class (name, desc, ranks[].d) - the best source for Forever-changed spells; match by name.
  - talenttrees.json already has each talent's description and per-rank values.
  Strip HTML to plain lines (keep line structure: name, rank, cost/cast lines, description). Never inject HTML from data into the page: build tooltip DOM with textContent.

## The task (user request: "improve our artifact to have tooltips (items, skills, talents, etc)")
Add WoW-style tooltips across the dashboard, phone first (tap to open, tap elsewhere to close; hover too on desktop):
1. Abilities and auras (damage breakdown rows, per-ability table, buff/debuff uptime rows, timeline cast bars/cooldowns/aura lanes, opener chips, event log names): name, rank, the spell's tooltip text (Forever wiki description when it exists for that name, else Wowhead classic tooltip), plus the preset's own numbers for that ability (DPS, share, casts, crit %, avg hit) where available.
2. Items (gear table, consumes/setup chips): name in quality colour, item level, slot, every stat line (from db.json for synthetic items, Wowhead tooltip for real ones), enchant line in green, set name and the 2-piece bonus if known.
3. Talents: already have a basic tooltip; restyle it to match (name, Rank x/y, current rank text, next rank text).
4. Races (race portraits/rows) and classes (crests): the racial traits from /home/user/w-dash/ui/core/forever/racials.json (name + description), for the class if variants differ.
5. Stat weights rows: a one-line explanation of what the number means.
Build a new dataset `tooltips` (rows {key, kind, title, sub, lines:[...], quality?}) from build_data.py (add a step there) so the page stays small; key by the same keys the page uses (names keys, item ids "i<id>", "talent:<class>:<tree>:<idx>", "race:<name>").
Style: a dark panel with a thin light border, gold title for spells, quality colour for items, green for "Equip:" / enchant lines, muted rank line, max width ~18rem, with the icon top-left; it must stay inside the viewport on a 375 px phone and in both light and dark themes (a dark tooltip is fine in both, like the game).
Keep everything else working. Test locally (both plain and stringified data), then publish: upload new/changed data files, update dataset docs (url + updated) and files/index.html, pinning if_version.

## Deliverable
Reply with a short summary: what got tooltips, data sources used, anything without text (e.g. Forever-only spells with no description), and the new files/index.html version. Don't message other sessions or touch git branches.
