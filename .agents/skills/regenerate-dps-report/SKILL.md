---
name: regenerate-dps-report
description: Regenerate the level 60 DPS Report (ui/dps_report) from fresh sims, re-running only the builds whose inputs changed unless asked for more, then republish its phone artifact. Use when the user asks to regenerate, refresh or update the DPS report or its numbers.
---

# Regenerate the DPS Report

The DPS Report ranks every level 60 DPS preset and explains each one (races, talents,
gear, rotation, consumes, timelines). It has one source and two outputs:

- `ui/dps_report/`: the page (`layout.html`, `report.js`, styles in `ui/scss/dps_report/`), in
  the sim UI's look. See its README.
- `ui/dps_report/data/<id>.json`: the datasets, one file each, described in
  `ui/dps_report/datasets.json`. Rows of per-build datasets carry `build` (the preset
  id, `<class>__<preset>`; `builds.json` calls it `id`).
- Output 1, sim UI: `/classic/dps_report/` (vite builds `ui/dps_report/index.html`).
- Output 2, artifact: `make dps-report-artifact` writes `dist/dps_report.html` (everything
  inlined), published to https://claude.ai/artifact/RUzPysb5dSk2RAXzXij4oR.

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

The pipeline is `tools/dps_report/README.md` and its Python scripts (on
`claude/wow-forever-dps-presets-m9u15h` until that branch is merged). Follow it: build the
merged tree, refresh the scoped builds' requests, rerun the sim steps for those builds only
(its "Rerunning only the stale parts" section says which outputs to delete), then
`python3 build_data.py`, which writes every dataset to `$DASH_DIR/data/`. Split the builds
across subagents by class; each one runs its own class's sims.

The sim steps include `mana_curves.py`, which sims every mana-using build (all but warrior
and rogue) through 100 fixed 480 s fights and reads its mana from the debug logs. It writes
`mana/<id>.json`, and `build_data.py` turns those into the `mana` (mana % every 5 s) and
`mana_summary` (when mana runs low) datasets. It takes seconds per build. Rerun it with
`--force <ids>` for every re-simmed build that uses mana.

After `build_data.py`, run `python3 events.py --data "$DASH_DIR/data" <ids>` for the
re-simmed builds. The Specs page shows each preset's typical, unlucky and lucky fight in the
sim UI's events view (timeline, chart, log, replay), which needs the whole debug log of each
fight. `events.py` sims those fights again from the seeds in the `timeline_<class>` datasets
and writes `events/<id>.json.gz`, plus the `event_names` and `event_icons` lookups. It says
so when a fight's DPS differs from the dataset's (then the binary isn't the one the numbers
came from). It takes seconds.

Skip the README's "Publishing" section and its `rd/` harness: they are for the old
Dashboard-type artifact. Steps 3 to 5 here replace them.

The standards in `orchestration/standards.md` decide the settings: 20,000+ iterations,
the generic boss, 120 / 180 / 300 s, binaries rebuilt after the last gear or DB change.

## 3. Merge, then record provenance

```bash
node tools/dps_report/merge.mjs "$DASH_DIR/data"
node tools/dps_report/stale.mjs --record <build ids...> --iterations 20000
```

`merge.mjs` replaces only the rows of the builds in that folder, merges lookup tables
(icons, names, tooltips, talent trees) by key, replaces whole-report datasets
(`race_summary`, `race_notes`, ...), and stamps `updated` in `datasets.json`. `build_data.py`
rebuilds every dataset from all the sim outputs, so its whole-report datasets already
cover every build. Record only the builds actually re-simmed, at the commit their sims
ran on (`--commit <rev>` if it isn't `HEAD`).

## 4. Check both outputs

- Build the artifact file: `make dps-report-artifact` (or
  `node tools/dps_report/build_artifact.mjs`). It fails if the file is over 16 MB.
- Open `dist/dps_report.html` in the pre-installed Chromium (Playwright with
  `executablePath: '/opt/pw-browsers/chromium'`; wrap the file in
  `<!doctype html><html><head><meta charset="utf-8"></head><body>...</body></html>`
  for a local look, and pass the session's proxy so d3 loads from cdnjs). Click through
  Ranking, Races, Specs, Compare and Method at phone and desktop width, and look for
  console errors.
- If you changed the page itself, also check `npx vite serve` at `/classic/dps_report/`.

## 5. Publish and commit

- Publish `dist/dps_report.html` with the Artifact tool, passing
  `url: https://claude.ai/artifact/RUzPysb5dSk2RAXzXij4oR` so the user's phone link
  keeps working. From a new session the tool first needs a `read` of that URL. Never
  create a second artifact for the report.
- Commit `ui/dps_report/data/`, `datasets.json`, `provenance.json` and any pipeline
  changes on a local branch, merge origin/master in, then push straight to `master`
  with no PR (user ruling, 2026-10-14).
- Reply with the artifact link, which builds were re-simmed and why, and the biggest
  ranking changes.

## Editing the page itself

Change `ui/dps_report/` and `ui/scss/dps_report/`, never the published artifact. Keep the
sim UI's look: use its classes and scss variables rather than new colours or fonts. The
page reads data only through `dash.data(id)` inside `dash.onData`. A new dataset needs an
entry in `datasets.json` and a file in `data/`.
