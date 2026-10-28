# DPS Report

Every level 60 DPS preset, ranked and explained. It looks like the rest of the sim UI and reuses its styles: the page
shell, header tabs and content blocks, the talent picker, the gear pane, the character sheet and tippy tooltips (Alt+click
pins one, as everywhere in the sim UI).

One source, two outputs:

| Output | How |
|---|---|
| Sim UI page, `/forever/dps_report/` | Built by vite with the rest of the UI (`make rundevserver`, `npx vite serve`, `make dist`). Linked from the homepage. |
| Single-file HTML for the site | `make dps-report-artifact` writes `dist/dps_report.html` (about 15 MB) with the script, styles, images, fonts and every dataset inside. `make dps-report-deploy` builds it and deploys it to Cloudflare Pages at https://forever-dps.pages.dev (`tools/dps_report/deploy_pages.py`). |

The single file is built by `tools/dps_report/build_artifact.mjs` with `vite.dps_report.config.mts`. It differs from the
sim UI build only in where the datasets come from (`data_source.inline.ts` bundles them instead of fetching them) and in
being one file: the `/forever/assets/...` images and fonts become data URIs. An artifact may load scripts only from a few
CDNs and stylesheets only from Google Fonts, so d3 comes from cdnjs and nothing else is loaded from outside. Font
Awesome, which the events view uses, is inlined too. The file must stay under 16 MB; the build fails above that.

Files:

- `index.html`, `index.ts`: the page. `index.ts` puts `layout.html` in the page and starts `report.js`.
- `layout.html`: the markup, in the sim UI shell (`.sim-ui`, header tabs with the title, `.tab-pane` pages, `.content-block`
  cards).
- `report.js`: the page logic. It draws every tab with d3 and reads data only through `dash.data(id)` inside
  `dash.onData`, an API it kept from its first life as a Claude dashboard artifact.
- `data.ts`: that `dash` API.
- `events.ts`: the Specs page's sample fights in the sim UI's own events view (timeline, chart, log and replay). It gives
  the view the report's names and icons in place of the sim UI's database and wowhead, which the single file can't reach.
- `data_source.ts`, `data_source.inline.ts`: where the datasets come from, in the sim UI and in the single file.
- `../scss/dps_report/`: the styles. `index.scss` brings in the sim UI's; `_report.scss` maps the report's own tokens onto
  the sim UI palette and styles the charts and controls.
- `datasets.json`: title, description and last update of each dataset.
- `data/<id>.json`: the datasets. Rows of per-build datasets carry `build` (`<class>__<preset>`; `builds.json` calls it `id`).
- `data/events/<build>.json.gz`: each preset's typical, unlucky and lucky fight with its whole debug log, made by
  `tools/dps_report/events.py`. The single file leaves out the lines only the log uses (the rotation waiting and the
  `[DEBUG]` damage formulas) to stay under 16 MB. See `vite.dps_report.config.mts`.
- `data/mana.json`, `data/mana_summary.json`: mana % of max every 5 s over a 480 s fight per mana-using preset, and when
  it first drops below 15%. Made by `tools/dps_report/mana_curves.py`.
- `data/changes.json`, `data/changes_snap.json`: the Changes tab, the updates with their public change lines and the
  results at each update's end, made by `tools/dps_report/changes.py` (`tools/dps_report/README.md`, "Changes tab").
- `provenance.json`: per build, the inputs its data was made from (written by `tools/dps_report/stale.mjs`).
- The raw sim outputs these datasets were built from are in `tools/dps_report/raw/<id>.tgz`, one pack per build, committed with them (`tools/dps_report/README.md`, "Partial regenerations").

To regenerate the data, use the `regenerate-dps-report` skill (`.agents/skills/regenerate-dps-report/`).
