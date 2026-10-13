---
name: regenerate-dps-report
description: Regenerate the level 60 DPS Report (ui/dps_report) from fresh sims, re-running only the builds whose inputs changed unless asked for more, then republish its phone artifact. Use when the user asks to regenerate, refresh or update the DPS report or its numbers.
---

# Regenerate the DPS Report

The DPS Report ranks every level 60 DPS preset and explains each one (races, talents,
gear, rotation, consumes, timelines). It has one source and two outputs:

- `ui/dps_report/report.html`: the page. `runtime.js` gives it its `dash.data` /
  `dash.onData` / `dash.calc` API, and `theme.css` its colours.
- `ui/dps_report/data/<id>.json`: the datasets, one file each, described in
  `ui/dps_report/datasets.json`. Rows of per-build datasets carry `build` (the preset
  id, `<class>__<preset>`; `builds.json` calls it `id`).
- Output 1, sim UI: `/classic/dps_report/` (vite builds `ui/dps_report/index.html`).
- Output 2, artifact: `node tools/dps_report/build_standalone.mjs` writes
  `dist/dps_report.html` (everything inlined), published to
  https://claude.ai/artifact/RUzPysb5dSk2RAXzXij4oR.

This normally runs in a Claude Code web session, because a full refresh spins up a dozen
or more subagents for hours. Work on a branch, never directly on `master`.

## 1. Decide the scope

```bash
node tools/dps_report/stale.mjs
```

It compares each build's inputs (git tree hashes of the shared sim inputs, `sim/<class>/`
and the class's `ui/<spec>/` folders) with `ui/dps_report/provenance.json`, the record of
when its data was last made. Then:

- **Default (incremental):** re-sim only the stale builds it lists.
- **The user names classes or builds:** re-sim those, plus anything stale if they agree.
- **The user asks for everything, or a shared input changed** (`sim/core`, encounters,
  the item database, mksettings): every build is stale; re-sim all of them.
- A build whose inputs didn't change but whose data is wrong (a pipeline bug) can be
  forced: name it.

Tell the user the scope in one line before starting long runs.

## 2. Generate data for the scoped builds

The pipeline lives in `tools/dps_report/` (see its README for scripts and order). Run it
for the scoped builds only, writing into a scratch folder (not `ui/dps_report/data/`)
with the same file names. Split the builds across subagents by class; each one writes
its own class's rows.

The standards in `orchestration/standards.md` decide the settings: 20,000+ iterations,
the generic boss, 120 / 180 / 300 s, binaries rebuilt after the last gear or DB change.

Whole-report datasets (`race_summary`, `race_notes`, the headline text in `notes`) are
summaries across every build. Recompute them after merging (step 3), from the merged
data, not from the partial run alone.

## 3. Merge, then record provenance

```bash
node tools/dps_report/merge.mjs <scratch-folder>
node tools/dps_report/stale.mjs --record <build ids...> --iterations 20000
```

`merge.mjs` replaces only the rows of the builds in the scratch folder, merges lookup
tables (icons, names, tooltips, talent trees) by key, and stamps `updated` in
`datasets.json`. Record only the builds actually re-simmed, at the commit their sims
ran on (`--commit <rev>` if it isn't `HEAD`).

## 4. Check both outputs

- Rebuild the artifact file: `node tools/dps_report/build_standalone.mjs`. It must stay
  under 16 MB.
- Open `dist/dps_report.html` in the pre-installed Chromium (Playwright with
  `executablePath: '/opt/pw-browsers/chromium'`; wrap the file in
  `<!doctype html><html><head><meta charset="utf-8"></head><body>...</body></html>`
  for a local look). Click through Ranking, Races, Specs, Compare and Method at phone
  width, and look for console errors.
- If you changed `report.html`, `runtime.js` or `index.ts`, also check
  `npx vite serve` at `/classic/dps_report/`.

## 5. Publish and commit

- Publish `dist/dps_report.html` with the Artifact tool, passing
  `url: https://claude.ai/artifact/RUzPysb5dSk2RAXzXij4oR` so the user's phone link
  keeps working. From a new session the tool first needs a `read` of that URL. Never
  create a second artifact for the report.
- Commit `ui/dps_report/data/`, `datasets.json`, `provenance.json` and any pipeline
  changes on the branch, push, and ask the user before opening or merging a PR.
- Reply with the artifact link, which builds were re-simmed and why, and the biggest
  ranking changes.

## Editing the page itself

Change `ui/dps_report/report.html`, never the published artifact. It only reads data
through `dash.data(id)` inside `dash.onData`. A new dataset needs an entry in
`datasets.json` and a file in `data/`.
