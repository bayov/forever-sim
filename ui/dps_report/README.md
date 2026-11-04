# DPS Report

Every level 60 preset, ranked and explained, in three modes: DPS, Tanking and Healing. It looks like the rest of the sim UI and reuses its styles: the page
shell, header tabs and content blocks, the talent picker, the gear pane, the character sheet and tippy tooltips (Alt+click
pins one, as everywhere in the sim UI).

## Modes and addresses

A mode bar in the header picks the report's mode. Each mode has its own tabs, which read only that mode's data. `MODES`
in `report.js` lists them. A mode whose datasets are not in `datasets.json` is left out of the bar, and with one mode
left the bar is hidden.

| Mode | Tabs | Datasets |
|---|---|---|
| DPS (default) | Ranking, Specs, Races, Compare, Changes, Method | those with no `mode` in `datasets.json` |
| Tanking | Ranking (score at a threat/survival weight, per boss or the six-boss mean, or any metric), Specs (per-boss table, talents, rotation, gear, setup, notes), Races, Method | `mode: "tank"` (`tank_*`) |
| Healing | the same four, ranking the healers' carry as tank healer or raid healer | `mode: "heal"` (`heal_*`) |

Raid is not a mode yet. The modes share the `mode: "common"` datasets (icons, names, races, tooltips).

Every DPS address is the same as before the modes: `#ranking`, `#spec`, `#races`, `#compare`, `#changes`,
`#changes/<update>`, `#method`, `#<build>`, and an empty hash for the Ranking. The other modes put theirs first:
`#tank`, `#tank/<tab>` (`ranking`, `spec`, `races`, `method`) and `#tank/<build>` for a preset's Specs page, and the
same with `#heal`. Each tab change is a history entry, so Back and Forward work across modes, and each mode keeps the
tab and preset it was last on.

## Outputs and how the data loads

One source, three outputs. Each loads only the data of the modes the reader opens:

| Output | How | Data |
|---|---|---|
| Sim UI page, `/forever/dps_report/` | Built by vite with the rest of the UI (`make rundevserver`, `npx vite serve`, `make dist`). Linked from the homepage. | `data_source.ts` fetches each `data/<id>.json` when a tab first asks for it. |
| Site, `dist/dps_report_site/` | `make dps-report-artifact` writes it; `make dps-report-deploy` builds it and deploys it to Cloudflare Pages at https://forever-dps.pages.dev (`tools/dps_report/deploy_pages.py`, which uploads every file and checks each one live). | `index.html` holds the page, styles, images and fonts; `data/<mode>.<hash>.json` holds one mode's datasets (`dps`, `tank`, `heal`, `common`), and `data/events/<build>.<hash>.json.gz` holds one preset's sample fights. `data_source.site.ts` fetches a mode's file the first time the page needs one of its datasets. So a DPS link fetches `dps` and `common`, and `#tank` fetches `tank` and `common`. |
| Single file, `dist/dps_report.html` (about 15 MB) | Written by the same `make dps-report-artifact`. It opens anywhere with no server, e.g. as a claude.ai artifact. `deploy_pages.py <file>` still deploys a single file. | All of it is inside (`data_source.inline.ts`). |

Both the site and the single file are built by `tools/dps_report/build_artifact.mjs` with `vite.dps_report.config.mts`
(the site with `--mode site`). They differ from the sim UI build in where the datasets come from, and in that the
`/forever/assets/...` images and fonts become data URIs. An artifact may load scripts only from a few CDNs and
stylesheets only from Google Fonts, so d3 comes from cdnjs and nothing else is loaded from outside. Font Awesome, which
the events view uses, is inlined too. Cloudflare Pages takes a file of up to 25 MiB; the build fails when a file is
bigger than that.

Files:

- `index.html`, `index.ts`: the page. `index.ts` puts `layout.html` in the page and starts `report.js`.
- `layout.html`: the markup, in the sim UI shell (`.sim-ui`, header tabs with the title, `.tab-pane` pages, `.content-block`
  cards).
- `report.js`: the page logic. It draws every tab with d3 and reads data only through `dash.data(id)` inside
  `dash.onData`, an API it kept from its first life as a Claude dashboard artifact.
- `data.ts`: that `dash` API.
- `events.ts`: the Specs page's sample fights in the sim UI's own events view (timeline, chart, log and replay). It gives
  the view the report's names and icons in place of the sim UI's database and wowhead, which the single file can't reach.
- `data_source.ts`, `data_source.site.ts`, `data_source.inline.ts`: where the datasets come from, in the sim UI, the
  site and the single file.
- `../scss/dps_report/`: the styles. `index.scss` brings in the sim UI's; `_report.scss` maps the report's own tokens onto
  the sim UI palette and styles the charts and controls.
- `datasets.json`: title, description, last update and report mode (`mode`: `tank`, `heal`, `common`, or none for
  DPS) of each dataset.
- `data/<id>.json`: the datasets. Rows of per-build datasets carry `build` (`<class>__<preset>`; `builds.json` calls it `id`).
- `data/events/<build>.json.gz`: each preset's typical, unlucky and lucky fight with its whole debug log, made by
  `tools/dps_report/events.py`. The single file leaves out the lines only the log uses (the rotation waiting and the
  `[DEBUG]` damage formulas) to keep the file small. See `vite.dps_report.config.mts`.
- `data/mana.json`, `data/mana_summary.json`: mana % of max every 5 s over a 480 s fight per mana-using preset, and when
  it first drops below 15%. Made by `tools/dps_report/mana_curves.py`.
- `data/changes.json`, `data/changes_snap.json`: the Changes tab, the updates with their public change lines and the
  results at each update's end, made by `tools/dps_report/changes.py` (`tools/dps_report/README.md`, "Changes tab").
- `provenance.json`: per build, the inputs its data was made from (written by `tools/dps_report/stale.mjs`).
- The raw sim outputs these datasets were built from are in `tools/dps_report/raw/<id>.tgz`, one pack per build, committed with them (`tools/dps_report/README.md`, "Partial regenerations").

To regenerate the data, use the `regenerate-dps-report` skill (`.agents/skills/regenerate-dps-report/`).
