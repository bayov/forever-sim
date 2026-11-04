---
name: regenerate-dps-report
description: Regenerate the level 60 DPS Report from fresh sims, re-running only the builds whose inputs changed, and deploy it to forever-dps.pages.dev. Use when the user asks to regenerate or refresh the report's numbers, or to add, remove or re-sim presets in it.
---

# Regenerate the DPS Report

One source, two outputs:

- `ui/dps_report/`: the page, styles in `ui/scss/dps_report/` (README).
- `ui/dps_report/data/<id>.json`: the datasets, described in `ui/dps_report/datasets.json`.
  Per-build rows carry `build`, the preset id `<class>__<preset>` (`id` in `builds.json`).
- `tools/dps_report/raw/<id>.tgz`: each build's raw sim outputs, committed with its
  datasets.
- The Tanking and Healing modes: `<mode>_*.json` datasets (`"mode"` in datasets.json),
  raw bench outputs in `tools/dps_report/raw/bench/{tank,heal}.tgz` (README "Tanking and
  Healing").
- Output 1: the sim UI at `/forever/dps_report/`.
- Output 2: `make dps-report-artifact` writes the site, `dist/dps_report_site/` (the page in
  `index.html`, one data file per mode, `data/{dps,tank,heal,common}.<hash>.json`, and one per
  preset's sample fights in `data/events/`, each fetched when the page first needs it), and the
  same report as one self-contained file, `dist/dps_report.html`. `make dps-report-deploy`
  uploads the site directory to https://forever-dps.pages.dev (Cloudflare Pages project
  `forever-dps`) and checks every file live. Never publish the report as a Claude artifact.

Work on a local branch, never directly on `master`.

## 1. Decide the scope

```bash
node tools/dps_report/stale.mjs
```

It lists stale builds (inputs changed since `ui/dps_report/provenance.json`), new requests
and removed presets. A stale build *may* have moved; the probe says whether it did.

- **Default:** re-sim only the stale builds, never every build for a change that touches
  only some.
- **User names classes or builds:** those, plus anything stale if they agree.
- **Many stale** (game update, `sim/core` or proto change): `pipeline.sh hydrate`, then
  `pipeline.sh probe` (README "Probes"). Re-sim only what it says moved, plus the audit
  builds it picks, and publish those ids. If the audit fails, re-sim every cleared build.
- **Text, tooltips or `build_data.py` only:** no sims.
- **A false positive** that plainly can't feed the build, with no binaries for a probe:
  `stale.mjs --record <ids> --by hand`, reason in the commit message. Prefer the probe.

It also lists stale Tanking and Healing modes in a second block (`benches` in `--json`):
their fingerprint covers the shared sim code, the classes' code, the encounters and the
benches. A stale mode re-sims whole (`pipeline.sh tankheal tank`, about TANK_TIME;
`pipeline.sh tankheal heal`, about HEAL_TIME, most of each the race sims); there is no probe
for it. Run it in the background with its log in a file.

Tell the user the scope in one line before long runs.

## 2. Generate data

Build the merged tree (master plus `patches/`) and the binaries per
`tools/dps_report/README.md`, then run `tools/dps_report/pipeline.sh` with `DASH_DIR` set
(README "Partial regenerations"):

- **Some builds:** `pipeline.sh hydrate <ids>`. It checks the packs reproduce the published
  numbers for every build outside `<ids>`; if one differs, add it to the scope or find out
  why. Then `pipeline.sh run <ids>`, `pipeline.sh publish <ids>`.
- **Add a preset:** write `req/<id>.json` (mksettings) and `req/<id>.meta.json`, check with
  `verify.py`, add it to `DISPLAY` in `build_data.py` and to the assumptions in
  `assumptions_def.py` that list builds by id, then as above.
- **Remove a preset:** delete `req/<id>.*` and its variants' requests, then as no sims.
- **No sims:** `hydrate`, `publish`.
- **After a probe:** `pipeline.sh run <the ids it names>`, then `publish` the same ids. Run
  the probe with the binaries the run will use, after the requests are regenerated.
- **Every build:** only when packs have no probes or an audit failed: `run`, then
  `publish`, no hydrate (about 95 min on 4 cores). For more than one class's worth, shard
  over remote sessions: `pipeline.sh shard <k> <n> <ids>` splits by cost; each session, on
  the same master commit with `patches/` applied and binaries built, runs
  `pipeline.sh run <its ids>` in an empty `$DASH_DIR`, then
  `pipeline.sh pack shard<k>.tgz <its ids>`, and pushes the pack to `dash-raw/<run>/<k>`.
  Fetch the packs, `pipeline.sh unpack` each into a hydrated `$DASH_DIR`, then publish.

A stopped run finishes with `pipeline.sh resume <ids>`. `publish` writes every output
(datasets, Changes data, sample fights, provenance, raw packs, removed presets dropped);
set `SIM_COMMIT=<rev>` if the sims didn't run at HEAD. `merge.mjs` is for hand repairs only.

`run` and `resume` also sim the Tanking and Healing modes when nothing DPS is stale or with
`TANKHEAL=1`; `publish` builds their datasets, and for a mode simmed since its last publish
records its provenance and repacks `raw/bench/<mode>.tgz`. Their settings are the docs'
(docs/tank_sim.md, docs/healing_sim.md): 1,000 iterations, the 7 boss encounters, the reqs'
seeds (healbench: 101). Their races (`bench_races.py`) sim every race, its racials off and each
racial part off on paired fights (seed 1, labeled streams), sized by adaptive.py's `BENCH_GOALS`
(the DPS luck goals with floors 2.5 times looser, in each score's unit; healing capped at
6,000 fights), into `<mode>/races.jsonl` and the `<mode>_races` and `<mode>_race_breakdown`
datasets. Build `tankbench` and `healbench` with `tankheal.py bins` (HEAD plus patches 0002 and
0006, the racials switch); a bench without the switch stops the run. Their text
(preset lines, notes, boss lines, Method) is written in `tools/dps_report/tankheal_text/{tank,heal}.md`,
for players, not taken from the docs: a new preset or boss needs its entry there, and the build
stops on a missing one or on a tool or file word in anything the page shows (`tankheal.py` LEAK).

Settings come from `orchestration/standards.md`: 20,000+ iterations, the generic boss,
120 / 180 / 300 s, binaries rebuilt after the last gear or DB change.

### The Changes tab

Every change needs a public line (standards 30). `python3 tools/dps_report/changes.py check`
lists commits since the live update's start that change build inputs and no line covers.
Add lines to the live update in `tools/dps_report/changes/updates.json`, or mark a commit
`internal` with a reason when no player would care. Then `changes.py build`.

The lines ship: plain game terms, no commit ids, dates of our work, session or audit names,
ruling references or file paths. Sources are public links, primary first: Blizzard's posts,
a wago.tools page of the changed rows or hotfix, foreverchanges.pro only when nothing else
shows it.

Start a new update before the report is republished with its first change, once the data
ending the previous one is committed: `changes.py new game <id> "<title>" --source-label
<label> --source-url <url>` for a balance build or dev notes; `changes.py new sim <id>
"<title>"` for our own first change after a game update's full pass was deployed. Commit
what it writes (`tools/dps_report/changes/`, the two datasets, `datasets.json`).

## 3. Check both outputs

- `make dps-report-artifact`. It fails when any file is over 25 MiB (Cloudflare's per-file
  limit). Its last line lists the site's files and size; a mode whose datasets are missing has
  no `data/<mode>.*.json` and is left out of the page's mode bar.
- Serve `dist/dps_report_site/` (`python3 -m http.server -d dist/dps_report_site`) and open it
  in Playwright with `executablePath: '/opt/pw-browsers/chromium'` and the session's proxy so
  d3 loads from cdnjs. In each mode (DPS, `#tank`, `#heal`) click through its tabs at phone and
  desktop width; look for console errors, and check in the network log that a mode fetches only
  its own data file and `common`.
- Open `dist/dps_report.html` the same way (it holds the whole page, so a `file://` URL or any
  static server works) and spot-check one page per mode.
- Page changed: also check `npx vite serve` at `/forever/dps_report/`.
- Run `go run -tags=with_db ./tools/raid_sim_data` and commit `ui/raid_sim/data/` with the
  report. Then `go run -tags=with_db ./tools/compsearch -stale` lists comp searches whose
  builds changed: name them in the reply, re-run only if the user asks.

## 4. Commit

Commit `ui/dps_report/data/`, `datasets.json`, `provenance.json`, `tools/dps_report/raw/`
and `tools/dps_report/changes/` together, with any pipeline changes, on a local branch from
origin/master. If origin/master changed any of them meanwhile, don't hand-merge: take
master's, `hydrate` again, re-run `stale.mjs`, re-sim what it adds, and `publish` again.
Except `tools/dps_report/changes/updates.json`: merge it by hand, keeping both sides' lines,
and let `publish` regenerate the rest.

## 5. Push, then deploy

- Push straight to `master`, no PR, merging origin/master first as in step 4.
- From the pushed commit: `make dps-report-deploy`. It refuses uncommitted report changes
  or a HEAD not on origin/master (`--unmerged` only when the user asks). It needs
  `CLOUDFLARE_ACCOUNT_ID` and a Pages Edit token, which the cloud environment provides.
  Not plain `wrangler pages deploy` (the cloud proxy breaks its upload). `--branch <name>`
  gives a preview URL.
- Reply with https://forever-dps.pages.dev, which builds were re-simmed and why, the
  biggest ranking changes, and the live update's public lines.

## Editing the page itself

Change `ui/dps_report/` and `ui/scss/dps_report/`, never the deployed file. The dps-report
skill's B8 has the style and data rules.
