---
name: regenerate-dps-report
description: Regenerate the level 60 DPS Report from fresh sims, re-running only the builds whose inputs changed, and deploy it to forever-dps.pages.dev. Use when the user asks to regenerate or refresh the report's numbers, or to add, remove or re-sim presets in it.
---

# Regenerate the DPS Report

The DPS Report ranks every level 60 DPS preset and explains each one (races, talents,
gear, rotation, consumes, timelines). It has one source and two outputs:

- `ui/dps_report/`: the page (`layout.html`, `report.js`, styles in
  `ui/scss/dps_report/`), in the sim UI's look. See its README.
- `ui/dps_report/data/<id>.json`: the datasets, one file each, described in
  `ui/dps_report/datasets.json`. Rows of per-build datasets carry `build` (the preset id,
  `<class>__<preset>`; `builds.json` calls it `id`).
- `tools/dps_report/raw/<id>.tgz`: each build's raw sim outputs, committed with the
  datasets they made, so a regeneration re-sims only what changed and still rebuilds
  every dataset.
- Output 1, sim UI: `/forever/dps_report/` (vite builds `ui/dps_report/index.html`).
- Output 2, the site: `make dps-report-artifact` writes `dist/dps_report.html` (everything
  inlined), deployed to Cloudflare Pages at https://forever-dps.pages.dev (project
  `forever-dps`). Never publish the report as a Claude artifact; the old one
  (RUzPysb5dSk2RAXzXij4oR) is retired.

A partial refresh takes minutes. A full one (a shared sim input changed) takes about
95 min on 4 cores, or is sharded over remote sessions. Work on a branch, never directly on
`master`.

## 1. Decide the scope

```bash
node tools/dps_report/stale.mjs
```

It compares each build's inputs (the sim's packages and protos, the step scripts, `sim/<class>/`, the item database
entries the build wears, and the build's own request, meta and assumption rotations; README "Partial regenerations")
with `ui/dps_report/provenance.json`, the record of when its data was last made, and lists new requests (a preset not in
the report yet) and removed presets. A stale build *may* have moved; the probe says whether it did. Then:

- **Default (incremental):** re-sim only the stale builds it lists. Never re-sim every
  build for a change that touches only some: the others' raw outputs are kept in
  `tools/dps_report/raw/<id>.tgz`.
- **The user names classes or builds:** re-sim those, plus anything stale if they agree.
- **Many builds stale** (a game update, a `sim/core` or proto change): probe first. `pipeline.sh hydrate`, then
  `pipeline.sh probe` re-runs each stale build's sims at a tenth of their iterations and compares them with the probe in
  its raw pack (README "Probes"). Re-sim only what it says moved, plus the audit builds it picks, and publish those ids:
  publish checks the audit and records the cleared builds (`--by probe`). If the audit fails, re-sim every cleared build.
- **Text, tooltips or `build_data.py` only:** no sims.
- **A false positive** the probe can't run on (no binaries at hand) and that plainly can't feed the build: re-record it
  (`stale.mjs --record <ids> --by hand`) instead of re-simming, and give the reason in the commit message. Prefer the
  probe: it checks instead of arguing.

Tell the user the scope in one line before starting long runs.

## 2. Generate data for the scoped builds

The pipeline is `tools/dps_report/README.md`: build the merged tree (master plus
`patches/`) and the binaries, then run `tools/dps_report/pipeline.sh` with `DASH_DIR` set
(its "Partial regenerations" section has the table):

- **Some builds:** `pipeline.sh hydrate <ids>` loads every build's raw pack and checks
  that the packs reproduce the published sim numbers for every build outside `<ids>`. If
  one differs, add it to the scope or find out why before going on. Then
  `pipeline.sh run <ids>` (it adds their duty and encounter variants and clears their old
  outputs first) and `pipeline.sh publish <ids>`.
- **Add a preset:** write `req/<id>.json` (mksettings) and `req/<id>.meta.json`, check it
  with `verify.py`, add it to `DISPLAY` in `build_data.py` and to the assumptions it
  shares in `assumptions_def.py` where they list builds by id, then as above with its id.
- **Remove a preset:** delete `req/<id>.*` and its variants' requests, then
  `pipeline.sh hydrate`, `pipeline.sh publish`.
- **No sims:** `pipeline.sh hydrate`, `pipeline.sh publish`.
- **After a probe:** `pipeline.sh run <the ids it names>` (they include the audit builds), then `pipeline.sh publish
  <the same ids>`. Run the probe with the binaries the run will use, after the requests are regenerated.
- **Every build:** only when the packs have no probes or an audit failed: `pipeline.sh run`, then `pipeline.sh publish`,
  with no hydrate. It takes about 95 min on 4 cores; for
  more than one class's worth, fan out instead of running it here: one remote session per shard (`pipeline.sh shard <k>
  <n> <ids>` balances them by cost), each on the same master commit with `patches/` applied and the binaries built,
  running `pipeline.sh run <its ids>` in an empty `$DASH_DIR` and `pipeline.sh pack shard<k>.tgz <its ids>`, and pushing
  the pack to `dash-raw/<run>/<k>`. Fetch the packs, `pipeline.sh unpack` each into a hydrated `$DASH_DIR`, then publish.

`run` plans first (it counts every step's sims, catches a script that can't enumerate its work and prints an estimate).
The sims seed every iteration and `pipeline.sh` pins the split count (`DASH_SPLITS=4`, patch 0005), so a build's numbers
don't depend on the machine. A run that stops can be finished with `pipeline.sh resume <ids>`.

`publish` writes everything: every dataset into `ui/dps_report/data/` (its `build_data.py` reads every build's raw
outputs, so ranks, the race summary and the assumptions cover every build), the Changes tab's data (`changes.py build`), the sample fights of the re-simmed builds
(`events.py`, which needs the binary the numbers came from), provenance for them (`SIM_COMMIT=<rev>` if the sims didn't
run at HEAD), the raw packs, and drops removed presets everywhere. `merge.mjs` is for hand repairs only.

The standards in `orchestration/standards.md` decide the settings: 20,000+ iterations,
the generic boss, 120 / 180 / 300 s, binaries rebuilt after the last gear or DB change.

### The Changes tab

The Changes tab compares the live update's start with the new data (standards 30;
`tools/dps_report/README.md`, "Changes tab"). It runs no sims, but every change needs a
public line. Run `python3 tools/dps_report/changes.py check`, which lists the commits since
the update's start that change build inputs and that no line covers. Add lines to the live
update in `tools/dps_report/changes/updates.json`, or mark a commit `internal` with a
reason when no player would care (a refactor that moves no numbers).

The lines ship, so write them in plain game terms, with no commit ids, dates of our work,
session or audit names, ruling references or file paths. Sources are public links, primary
first (standards 30): Blizzard's posts, a wago.tools page of the changed client rows or
hotfix, and foreverchanges.pro only when nothing else shows the change. Then run
`python3 tools/dps_report/changes.py build`.

Start a new update before the report is republished with its first change, once the
report data that ends the previous one is committed:
`changes.py new game <id> "<title>" --source-label <label> --source-url <url>` for a
balance build or development notes, `changes.py new sim <id> "<title>"` for our own first
change after a game update's full pass was deployed. Commit what it writes
(`tools/dps_report/changes/`, the two datasets, `datasets.json`).

## 3. Check both outputs

- Build the artifact file: `make dps-report-artifact` (or
  `node tools/dps_report/build_artifact.mjs`). It fails if the file is over 25 MiB,
  Cloudflare Pages' limit for one file.
- Open `dist/dps_report.html` in the pre-installed Chromium (Playwright with
  `executablePath: '/opt/pw-browsers/chromium'`; wrap the file in
  `<!doctype html><html><head><meta charset="utf-8"></head><body>...</body></html>`
  for a local look, and pass the session's proxy so d3 loads from cdnjs). Click through
  Ranking, Races, Specs, Compare, Changes and Method at phone and desktop width, and look
  for console errors.
- If you changed the page itself, also check `npx vite serve` at `/forever/dps_report/`.
- The raid sim reads the same builds. Run `go run -tags=with_db ./tools/raid_sim_data` and
  commit `ui/raid_sim/data/` with the report. Then
  `go run -tags=with_db ./tools/compsearch -stale` lists the committed comp searches whose
  builds changed. Name them in the reply, but don't re-run them unless the user asks.

## 4. Commit

Commit `ui/dps_report/data/`, `datasets.json`, `provenance.json`, `tools/dps_report/raw/`
and `tools/dps_report/changes/` together, with any pipeline changes, on a local branch
from origin/master. If origin/master changed any of them meanwhile, don't hand-merge
them: take master's, `hydrate` again, re-check `stale.mjs` (master may have made more
builds stale), re-sim what that adds, and `publish` again. The exception is
`tools/dps_report/changes/updates.json`, which is hand-kept: merge it by hand (keep both
sides' lines and updates), and let `publish` (`changes.py build`) regenerate the rest of
`changes/`'s outputs and the two datasets.

## 5. Push, then deploy

- Push straight to `master` with no PR, merging origin/master first as in step 4.
- Deploy from the pushed commit: `make dps-report-deploy` (or
  `python3 tools/dps_report/deploy_pages.py` after building). It refuses a report tree
  with uncommitted changes, or a HEAD that isn't on origin/master (pass `--unmerged` only
  when the user asks to deploy unmerged work), and records the commit in the deployment.
  It needs `CLOUDFLARE_ACCOUNT_ID` and a Pages Edit token, which the cloud environment
  provides (the `CLOUDFLARE_API_TOKEN` variable, or a network secret the proxy adds).
  Don't use plain `wrangler pages deploy`: the cloud proxy breaks its asset upload, so the
  script uploads them directly. It checks that the live site serves the new file byte for
  byte and rolls production back if not. To try a page change first, pass
  `--branch <name>` for a preview URL.
- Reply with the link https://forever-dps.pages.dev, which builds were re-simmed and why,
  the biggest ranking changes, and the live update's public lines.

## Editing the page itself

Change `ui/dps_report/` and `ui/scss/dps_report/`, never the deployed file. Keep the sim
UI's look: use its classes and scss variables, not new colours or fonts. The page reads
data only through `dash.data(id)` inside `dash.onData`. A new dataset needs an entry in
`datasets.json` and a file in `data/`.
