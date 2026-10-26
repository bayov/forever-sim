---
name: regenerate-dps-report
description: Regenerate the level 60 DPS Report (ui/dps_report) from fresh sims, re-running only the builds whose inputs changed (each build's raw outputs are kept, so the others need no sims), then deploy it to https://forever-dps.pages.dev. Use when the user asks to regenerate, refresh or update the DPS report or its numbers, or to add, remove or re-sim presets in it.
---

# Regenerate the DPS Report

The DPS Report ranks every level 60 DPS preset and explains each one (races, talents,
gear, rotation, consumes, timelines). It has one source and two outputs:

- `ui/dps_report/`: the page (`layout.html`, `report.js`, styles in `ui/scss/dps_report/`), in
  the sim UI's look. See its README.
- `ui/dps_report/data/<id>.json`: the datasets, one file each, described in
  `ui/dps_report/datasets.json`. Rows of per-build datasets carry `build` (the preset
  id, `<class>__<preset>`; `builds.json` calls it `id`).
- `tools/dps_report/raw/<id>.tgz`: each build's raw sim outputs, committed with the datasets they
  made, so a regeneration re-sims only what changed and still rebuilds every dataset.
- Output 1, sim UI: `/classic/dps_report/` (vite builds `ui/dps_report/index.html`).
- Output 2, the site: `make dps-report-artifact` writes `dist/dps_report.html` (everything
  inlined), deployed to Cloudflare Pages at https://forever-dps.pages.dev (project
  `forever-dps`). The old Claude artifact (RUzPysb5dSk2RAXzXij4oR) is retired: don't
  publish the report as an artifact (user ruling, 2026-10-25).

A partial refresh takes minutes; a full one (a shared sim input changed) about 95 min on 4
cores, or sharded over remote sessions. Work on a branch, never directly on `master`.

## 1. Decide the scope

```bash
node tools/dps_report/stale.mjs
```

It compares each build's inputs (git tree hashes of the shared sim inputs, `sim/<class>/`, the class's `ui/<spec>/`
folders, and the build's own request and assumption rotations) with `ui/dps_report/provenance.json`, the record of when
its data was last made, and lists new requests (a preset not in the report yet) and removed presets. Then:

- **Default (incremental):** re-sim only the stale builds it lists. Never re-sim every build for a change that touches
  only some: each build's raw outputs are kept in `tools/dps_report/raw/<id>.tgz`, so the others don't need it.
- **The user names classes or builds:** re-sim those, plus anything stale if they agree.
- **A shared sim input changed** (`sim/core`, encounters, the item database, the local patches): every build is stale;
  re-sim all of them.
- **Text, tooltips or `build_data.py` only:** no sims.
- **A false positive** (an input changed that can't feed the build, like a test-only commit or a comment): re-record it
  (`stale.mjs --record <ids>`) instead of re-simming, and give the reason in the commit message.

Tell the user the scope in one line before starting long runs.

## 2. Generate data for the scoped builds

The pipeline is `tools/dps_report/README.md`: build the merged tree (master plus `patches/`) and the binaries, then
`tools/dps_report/pipeline.sh` with `DASH_DIR` set (its "Partial regenerations" section has the table):

- **Some builds:** `pipeline.sh hydrate <ids>` loads every build's raw pack and checks that the packs reproduce the
  published sim numbers for every build outside `<ids>`. If one differs, add it to the scope or find out why before going
  on. Then `pipeline.sh run <ids>` (it adds their duty and encounter variants and clears their old outputs first) and
  `pipeline.sh publish <ids>`.
- **Add a preset:** write `req/<id>.json` (mksettings) and `req/<id>.meta.json`, check it with `verify.py`, add it to
  `DISPLAY` in `build_data.py` and to the assumptions it shares in `assumptions_def.py` where they list builds by id, then
  as above with its id.
- **Remove a preset:** delete `req/<id>.*` and its variants' requests, `pipeline.sh hydrate`, `pipeline.sh publish`.
- **No sims:** `pipeline.sh hydrate`, `pipeline.sh publish`.
- **Every build:** `pipeline.sh run`, then `pipeline.sh publish`, with no hydrate. It takes about 95 min on 4 cores; for
  more than one class's worth, fan out instead of running it here: one remote session per shard (`pipeline.sh shard <k>
  <n> <ids>` balances them by cost), each on the same master commit with `patches/` applied and the binaries built,
  running `pipeline.sh run <its ids>` in an empty `$DASH_DIR` and `pipeline.sh pack shard<k>.tgz <its ids>`, and pushing
  the pack to `dash-raw/<run>/<k>`. Fetch the packs, `pipeline.sh unpack` each into a hydrated `$DASH_DIR`, then publish.

`run` plans first (it counts every step's sims, catches a script that can't enumerate its work and prints an estimate).
The sims seed every iteration and `pipeline.sh` pins the split count (`DASH_SPLITS=4`, patch 0005), so a build's numbers
don't depend on the machine. A run that stops can be finished with `pipeline.sh resume <ids>`.

`publish` writes everything: every dataset into `ui/dps_report/data/` (its `build_data.py` reads every build's raw
outputs, so ranks, the race summary and the assumptions cover every build), the sample fights of the re-simmed builds
(`events.py`, which needs the binary the numbers came from), provenance for them (`SIM_COMMIT=<rev>` if the sims didn't
run at HEAD), the raw packs, and drops removed presets everywhere. `merge.mjs` is for hand repairs only.

The standards in `orchestration/standards.md` decide the settings: 20,000+ iterations,
the generic boss, 120 / 180 / 300 s, binaries rebuilt after the last gear or DB change.

## 3. Check both outputs

- Build the artifact file: `make dps-report-artifact` (or
  `node tools/dps_report/build_artifact.mjs`). It fails if the file is over 16 MB.
- Open `dist/dps_report.html` in the pre-installed Chromium (Playwright with
  `executablePath: '/opt/pw-browsers/chromium'`; wrap the file in
  `<!doctype html><html><head><meta charset="utf-8"></head><body>...</body></html>`
  for a local look, and pass the session's proxy so d3 loads from cdnjs). Click through
  Ranking, Races, Specs, Compare and Method at phone and desktop width, and look for
  console errors.
- If you changed the page itself, also check `npx vite serve` at `/classic/dps_report/`.

## 4. Commit

Commit `ui/dps_report/data/`, `datasets.json`, `provenance.json` and `tools/dps_report/raw/` together, with any pipeline
changes, on a local branch from origin/master. If origin/master changed any of them meanwhile, don't hand-merge them:
take master's, `hydrate` again, re-check `stale.mjs` (master may have made more builds stale), re-sim what that adds,
and `publish` again.

## 5. Push, then deploy

- Push straight to `master` with no PR (user ruling, 2026-10-14), merging origin/master first as in step 3.
- Deploy from the pushed commit: `make dps-report-deploy` (or `python3 tools/dps_report/deploy_pages.py` after
  building). It refuses a report tree with uncommitted changes, or a HEAD that isn't on origin/master (`--unmerged` only
  when the user asks to deploy unmerged work), and records the commit in the deployment. It needs
  `CLOUDFLARE_ACCOUNT_ID` and a Pages Edit token, which the cloud environment provides (the `CLOUDFLARE_API_TOKEN`
  variable, or a network secret the proxy adds). Plain `wrangler pages deploy` fails in cloud sessions, because the
  proxy breaks its asset upload; the script uploads those directly. It checks the live site serves the new file byte
  for byte and rolls production back if not. To try a page change first, pass `--branch <name>` for a preview URL.
  Never publish the report as a Claude artifact.
- Reply with the link https://forever-dps.pages.dev, which builds were re-simmed and why, and the biggest
  ranking changes.

## Editing the page itself

Change `ui/dps_report/` and `ui/scss/dps_report/`, never the deployed file. Keep the
sim UI's look: use its classes and scss variables rather than new colours or fonts. The
page reads data only through `dash.data(id)` inside `dash.onData`. A new dataset needs an
entry in `datasets.json` and a file in `data/`.
