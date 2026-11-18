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

It also lists the stale Tanking and Healing presets in a second block (`benches` in
`--json`), per preset like the DPS builds: a preset's fingerprint covers the shared sim
code, the encounters, the bench and its scripts, its request and the code of every class in
its raid (healing raids hold frost mage fillers, and a warrior tank for tank healers). So a
`sim/shaman` change stales the shaman presets, a `sim/mage` change every healing preset.
Re-sim them with `pipeline.sh tankheal --stale` (or `tankheal.py run <mode> --presets
<stem,...>`): the main bench runs whole (about 15min to 20min a mode), then the race sims,
most of a mode's cost (Healing 1.5h to 4.3h on 4 cores in full), only for the stale presets.
The others keep their pack's race rows when their main-run rows match the pack's bytes, else
they are re-simmed too. `publish` records only the presets simmed. With no flag, a mode
re-sims whole. Run it in the background with its log in a file.

**Many stale presets** (a `sim/core` change stales them all): probe them first (README
"Bench probes"). `pipeline.sh hydrate`, then `pipeline.sh benchprobe [tank|heal]`. It
prints per mode the presets that moved, the cleared ones and one audit preset, and the
`tankheal.py run <mode> --presets <moved>,<audit>` to run next. Then `pipeline.sh publish`:
it checks the audit and records the cleared presets (`--record-bench --by probe`). If the
audit fails, re-sim every cleared preset. On 4 cores a check of every preset takes about
30min to 48min for Healing (a raid healer about 565s, a tank healer about 120s) and 8min to
16min for Tanking (about 80s a preset), against about 87min and 50min for full runs. So probe
when most of a mode is stale, and just re-sim when only a few presets are. A pack without
probes can't be checked (`benchprobe` says so and exits 1): re-sim the mode. The bench probe
is optional: `tankheal.py run` makes it only with `TANKHEAL_PROBE=1`, and that is worth it only
when a mode is mostly stale and a full re-sim of it would take well over 2h. It spends the same
time again probing what it simmed.

Tell the user the scope in one line before long runs.

## 2. Generate data

Commit any sim or data change first, then build the binaries with
`tools/dps_report/pipeline.sh bins` (`build_bins.py`: every binary of the three modes from a
clean worktree at `${SIM_COMMIT:-HEAD}`, stamped in `bin/STAMP.json`, with no patches: the sim
reads the report switches itself, README "Report switches"). It
builds nothing when the stamp already matches. `run`, `resume`, `probe`, `tankheal` and
`publish <ids>` refuse binaries whose stamp doesn't match the code, and uncommitted changes
under the sim's inputs: rebuild, don't override. `DASH_ALLOW_BIN_MISMATCH=<reason>` is for
throwaway checks only, since `stale.mjs --record` then refuses those builds (README
"Building the binaries"). Then run `tools/dps_report/pipeline.sh` with `DASH_DIR` set
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
  `publish <every id>` (the line `run` prints; a bare `publish` records, packs and
  samples none of them), no hydrate (about 95 min on 4 cores). With Tanking and Healing too, prefer
  "Parallel full regeneration" below. For more than one class's worth, shard
  over remote sessions: `pipeline.sh shard <k> <n> <ids>` splits by cost; each session, on
  the same master commit with binaries built, runs
  `pipeline.sh run <its ids>` in an empty `$DASH_DIR`, then
  `pipeline.sh pack shard<k>.tgz <its ids>`, and pushes the pack to `dash-raw/<run>/<k>`.
  Fetch the packs, `pipeline.sh unpack` each into a hydrated `$DASH_DIR`, then publish.

A stopped run finishes with `pipeline.sh resume <ids>` (its probe step skips what it probed, and the Tanking and Healing
modes resume from `.progress.json` even if master moved, as long as the sim inputs `stale.mjs --bench-fingerprint <mode>` hashes
are the same). `pipeline.sh run` after `pipeline.sh probe` copies each build's check probe instead of probing it again.
The end-of-run `timing.py` summary now covers the Tanking and Healing benches and the time `hydrate`, `publish` and
`tankheal.py bins` took. `publish` writes every output
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
datasets. `pipeline.sh bins` builds `tankbench` and `healbench` too, from the same code as
`wowsimcli`, so they read every report switch (the race sims use `DASH_RACIALS_OFF` and
`DASH_RACIALS_OFF_CLASS`, and every bench run sets `DASH_SPLITS=4`). Their text
(preset lines, notes, boss lines, Method) is written in `tools/dps_report/tankheal_text/{tank,heal}.md`,
for players, not taken from the docs: a new preset or boss needs its entry there, and the build
stops on a missing one or on a tool or file word in anything the page shows (`tankheal.py` LEAK).

Settings come from `orchestration/standards.md`: 20,000+ iterations, the generic boss,
120 / 180 / 300 s, binaries rebuilt after the last gear or DB change (the stamp check enforces it).

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

### Parallel full regeneration

For a full regeneration, or one that re-sims most of Healing: run the three modes, and Healing's presets, in
separate cloud sessions at once (`tools/dps_report/parallel.py`, README "Parallel runs"). It takes about 1.5h of
sims instead of 3h to 6.5h, and the merged outputs are the same bytes as one session's. A few builds, one mode's
stale presets or a no-sims publish keep the one-session path above: the workers' binary builds and the merge cost
more than they save.

The coordinator (this session):

1. Commit and push to master everything the sims read. Pin that SHA and pick a run id, `<date>-<sha7>`.
2. `python3 tools/dps_report/parallel.py plan --run <run> --commit <sha>`, with `--heal-presets <stem,...>` when only
   some Healing presets are stale and `--modes` to leave a mode out. It prints one worker command per session (dps,
   tank, heal-0 with the Healing main bench, heal-1.. with every priest preset alone and the rest balanced by
   measured cost) and pushes `plan.json` to `dash-raw/<run>/plan`. `--dry` prints only.
3. Read your own session id (`get_session` with no id). For each planned session, `create_session` with the repo as
   `source_url`, `source_revision=<sha>`, and a prompt that says: run `<its command>` in the background
   (`run_in_background`, output to a log file), wait for it to exit without polling, then `send_message` the
   coordinator `<session id>` its last 20 log lines and exit code. Do not publish, deploy, commit, merge or push
   anything else, and do not touch master (the worker pushes only its `dash-raw/<run>/<session>` branch, with git
   plumbing). On a failure, report it the same way and stop.
4. Meanwhile run `pipeline.sh bins` here (publish sims the sample fights with them), then wait for the hand-backs. A
   hand-back wakes the session; don't poll. Schedule at most one fallback check-in (`send_later`, 90 min or more
   ahead) that reads `get_session` of each worker before any transcript.
5. When every worker has reported success: with `DASH_DIR` set, `python3 tools/dps_report/parallel.py merge <run>`.
   It fetches every branch, refuses parts made at another commit, other sim inputs, other binaries (STAMP.json
   inputHash and sha256s), another `DASH_SPLITS` or a Healing preset set other than the plan's, and refuses to merge
   with binaries here that another input hash would build. For a full DPS run it empties the outputs first; a
   partial Healing plan needs a hydrated `$DASH_DIR` (`pipeline.sh hydrate`), whose pack keeps the other presets.
   It prints the publish command: `SIM_COMMIT=<sha> pipeline.sh publish <every id>`.
6. Then steps 3 to 5 below as for one run (check, commit, push master, `make dps-report-deploy`), then
   `parallel.py clean <run>`. The cloud git proxy refuses branch deletes, so it may list the branches for the user
   to delete instead.

A worker that fails can be re-run alone with its same command in a new session; the others' branches stay.

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
