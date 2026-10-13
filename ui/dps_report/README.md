# DPS Report

Every level 60 DPS preset, ranked and explained. One source, two outputs:

| Output | How |
|---|---|
| Sim UI page, `/classic/dps_report/` | Built by vite with the rest of the UI (`make rundevserver`, `npx vite serve`, `make dist`). Linked from the homepage. |
| Single-file HTML for a Claude artifact | `node tools/dps_report/build_standalone.mjs` writes `dist/dps_report.html` with everything inlined. Published at https://claude.ai/artifact/RUzPysb5dSk2RAXzXij4oR. |

Files:

- `report.html`: the page (styles, markup, one script). It was written for the Claude dashboard artifact type, so
  it reads data through a small `dash` API.
- `runtime.js`: provides that API (`dash.data`, `dash.onData`, `dash.calc`, `dash.colors`, `dash.setLink`) in both
  outputs.
- `theme.css`: the colour, font and spacing tokens the page uses, light and dark.
- `index.html`, `index.ts`: the sim UI page. Datasets are fetched on first use.
- `datasets.json`: title, description and last update of each dataset.
- `data/<id>.json`: the datasets.
- `provenance.json`: per build, the inputs its data was made from (written by `tools/dps_report/stale.mjs`).

To regenerate the data, use the `regenerate-dps-report` skill (`.claude/skills/regenerate-dps-report/`).
