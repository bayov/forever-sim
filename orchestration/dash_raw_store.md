# DPS Report: keep the raw sim outputs, so a partial regeneration is cheap (design, 2026-10-26)

Status: built on 2026-10-26, after three audits (data correctness against the code, operations and git, workflow and simplicity). The commands are in `tools/dps_report/pipeline.sh` and `tools/dps_report/rawstore.py`, the steps in `tools/dps_report/README.md` ("Partial regenerations"). What the build changed from the plan below:

- `publish <ids>` packs only the re-simmed builds (and adds their variants, as `run` does). A pack records the commit it was simmed at, so repacking an unchanged build would rewrite it for nothing.
- `resume <ids>` finishes a stopped run without cleaning: it skips the builds `run_build.py`, `mana_curves.py` and `ablate.py` already finished (they redo work otherwise), and the other steps skip done work themselves.
- The shard branches keep their `dash-raw/<run>/<k>` names: there is no `dash-raw` branch to collide with.
- Bootstrap: the full run of 2026-10-26 (58 builds, every step) published through `publish`. The other 57 builds' values matched the published ones exactly; their ranks moved for the new build, numbers `merge.mjs` wrote as `1` are now `1.0`, and the sample fights differ only in the sim's action order.

## The problem

Adding one preset (Level 60 Shock (Ret side), 2026-10-26) took a full re-sim of all 58 builds, about 95 min on 4 cores, though the other 57 builds could not change. Three things force that:

1. **The raw sim outputs are thrown away.** `out/`, `names/`, `weights/`, `mana/`, `rv/results.jsonl`, `asm/results.jsonl` and `abl/ablate.json` live only in the `$DASH_DIR` of the machine that ran them. Only the derived datasets in `ui/dps_report/data/` are committed.
2. **Some datasets span every build.** `build_data.py` rebuilds every dataset from the raw outputs it finds. `race_summary`, `race_notes`, `assumptions` and the `rank` / `rankAll` columns of `builds` are computed across builds (build_data.py:590-593), and `merge.mjs` replaces the first three whole. A partial `$DASH_DIR` gives them as if the report had only the re-simmed builds. So does a text-only change (meta, `build_data.py`, tooltips) today, which also needs every build's raw outputs.
3. **Staleness is coarse.** `tools/rotopt` is a shared input, so one class's `<class>_presets.sh` edit marks every build stale (that is what made this one full). That part is a separate change (section F).

Re-simming a build also means clearing its outputs in seven places with different skip rules (`run_build.py` overwrites, `weights.py` and the reuse in `run_races.py` / `run_assumptions.py` compare mtimes, `rv` and `asm` skip done keys, `mana_curves.py` needs `--force`, `ablate.py` never skips but `ABL_MERGE=1` keeps the other builds).

## Goal

- Re-sim N builds: those N builds' sims only (a few minutes a build on 4 cores, against 95 for all 58), plus `build_data.py` (seconds).
- Add a build: write its request, sim it alone. Remove one: delete its request, no sims.
- A text or page-data change: no sims at all.
- Every published dataset, including the cross-build ones, is what a full run at the same inputs would give.

## Design

### A. Raw packs on master, one per build

`tools/dps_report/raw/<id>.tgz` holds one build's raw outputs, committed on master in the same commit as the datasets and `provenance.json` it produced:

```
out/<id>.json        run_build.py
names/<id>.tsv       run_build.py (DASH_NAMES), lines sorted and deduped (patch 0001 writes one copy per sim split)
weights/<id>.json    weights.py
mana/<id>.json       mana_curves.py (mana users)
rv.jsonl             that build's run_races.py lines, sorted by key, each key once
asm.jsonl            that build's run_assumptions.py lines, sorted by key, each key once, only keys run_assumptions.tasks() still lists
abl.json             that build's ablate.py entry, with its own iters and seed
fingerprint.json     sha256 of req/<id>.json and asm_apls/<id>.*, of the wowsimcli binary, and the sim split count (section D)
```

Packed deterministically (`tar --sort=name --mtime=@0 --owner=0 --group=0 --numeric-owner`, `gzip -n`), so an unchanged build's pack is byte-identical and git stores it once. Sizes (an `out/` file goes from 135 KB to 25 KB gzipped): about 2 MB for all 58, about 30 KB for a one-build change. Master already takes about 6 MB of new objects per full regeneration (`ui/dps_report/data` is 13 MB, half of it the per-build `events/*.json.gz`), so history grows about 30% faster.

Why on master and not a branch or R2: the datasets, their raw outputs and `provenance.json` land in one commit, so they can't disagree, and the master push is the only lock. No new branch (a `dash-raw` branch also can't coexist with the shard branches `dash-raw/<run>/<k>`), no push permission, no pointer from master to a store, nothing to fetch: a fresh clone has it. Two regenerations of different builds merge cleanly; of the same build, they conflict on exactly that pack. R2 with content-addressed keys is the fallback if history size ever matters. A single `raw.tgz` (always rewritten, always conflicting) and git LFS are out.

`pipeline.sh pack` / `unpack` already split and join per build; they gain `--per-build <dir>` to write and read this layout.

### B. Pipeline commands

- `hydrate`: start from an empty `$DASH_DIR` (only `req/` and the caches symlinked from the checkout), unpack every `raw/*.tgz`, and refuse a pack whose `fingerprint.json` names another split count than this machine's run will use. Then run `build_data.py` and compare with `ui/dps_report/data/`:
  - sim-derived numbers of a build (its `builds` row's DPS columns, `races`, `lengths`, `weights`, `mana_summary`, `race_variants`) differ: stop. The pack and the published data disagree; re-sim that build (add it to the scope).
  - anything else differs (text, names, tooltips, icons, from a `build_data.py`, meta or tree change since the last publish): expected, reported, and it goes out with this publish.
- `clean <ids>`: drop those builds from every raw output (files, jsonl lines, the `abl` entry), so every script re-sims them whatever its skip rule.
- `run <ids>`: adds every preset whose `baseBuild` is in `<ids>` (duty and encounter variants) and says so, calls `clean`, then the sim steps as now. `run_build.py` goes first for each build, since `run_races.py` and `run_assumptions.py` read its `out/<id>.json`.
- `publish [<ids>]`:
  1. `build_data.py` from the full `$DASH_DIR` into an empty `$DASH_DIR/data`, then copy every dataset into `ui/dps_report/data/`. No partial merge.
  2. `events.py --out ui/dps_report/data/events <ids>`, so the `event_names` / `event_icons` lookups cover every build's files, not only the new ones.
  3. Delete `events/` and `raw/` files of builds with no request.
  4. `stale.mjs --record <ids> --commit <the commit the sims ran at>`.
  5. `pack --per-build tools/dps_report/raw`, after checking every build has its outputs: `run_races.tasks()` and `run_assumptions.tasks()` all present, no `abl` errors, weights and mana where they apply.
  6. Stamp `updated` in `datasets.json` only for files whose content changed.

Scenarios:
- Add a preset: `req/<id>.*` + `verify.py`, `hydrate`, `run <id>`, `publish <id>`.
- Remove one: `rm req/<id>.*` (and its variants; `build_data.py` fails on a variant whose base is gone), `hydrate`, `publish`.
- A class change makes 6 builds stale: regenerate their requests, `hydrate`, `run <the 6>`, `publish <the 6>`.
- Text, tooltip or `build_data.py` change: `hydrate`, `publish`. No sims.
- New assumption combos: `hydrate`, `run_assumptions.py <ids>` alone (it skips combos already done), `publish`.
- Everything stale (a `sim/core` change): skip `hydrate`, `run` all, `publish`. Same cost as today.
- `merge.mjs` stays only for hand repairs.

### C. Landing it and deploying

Work on a branch from origin/master. Before pushing, merge origin/master and, if it changed anything under `ui/dps_report/data`, `tools/dps_report/raw` or `provenance.json`, never hand-merge those: take master's, re-hydrate, re-run `stale.mjs` (master may have made more builds stale), re-sim what that adds, and `publish` again, as with `db.json`. Deploy only after the push, from a clean checkout of the pushed commit; `deploy_pages.py` refuses a dirty tree or a HEAD not on origin/master and passes the commit to Pages.

### D. Reproducibility

- **The split count.** The sim splits a run over `runtime.NumCPU()` and averages the parts (sim/core/sim_concurrent.go:120, 519), so the last bits of averages and stdevs depend on the machine. Patch 0001 pins the split count with `DASH_SPLITS` (default 4, what this report's runs use) and `fingerprint.json` records it. A shard on a bigger machine still uses 4 splits, run in parallel.
- **build_data.py's unsorted globs** get `sorted()`: `names/*.tsv` (first label wins, build_data.py:141-152), `go_ranks` (:133) and `talent_names` (:92). `out_pre1b` (:310) is read but unused: delete that code.
- **Icons.** A failed icon or tree background fetch is silently dropped today (build_data.py:778-801, 859-871), which changes `names.icon` and the colours. Make it an error. The icon cache isn't stored: it isn't a sim output.
- **Skip rules by mtime** (weights.py:117, run_races.py:200, run_assumptions.py:164) become arbitrary after an unpack. `clean` covers the re-sim path; the follow-up gives each output the request hash it came from and compares that instead.

### E. Shards

Shards don't hydrate: every sim step is per build, and an empty `$DASH_DIR` skips nothing. Each shard runs its ids and pushes `pack <its ids>`, as now. The orchestrator hydrates, unpacks the shards over it and runs `publish`.

### F. Staleness: a separate, later change

The first draft dropped `tools/rotopt` from the shared inputs and hashed only the requests. That misses inputs that change the sims without changing any request: `asm_apls/<id>.*` (generated from the rotopt templates, read by `run_assumptions.py`), patches 0001-0004, `dashstats`, `assumptions_def.py`, the step scripts' constants (iterations, lengths, races, stats, durations, ablation entries) and their env knobs (`ITERS`, `FIGHTS`, `SEED`...). For now:

- `stale.mjs` also hashes `req/<id>.json` and `asm_apls/<id>.*` per build, and adds `tools/dps_report/patches` to the shared inputs.
- `tools/rotopt` stays shared. A false positive costs a re-record without a re-sim, which is already the practice.
- Later: a fingerprint per step and build (binary, script, its `assumptions_def` entries), and a structured `edits` field in each meta instead of prose `notes`, so requests can be regenerated and checked by a script.

### G. Bootstrap

The full run of 2026-10-26 (58 builds, 4 splits) is packed into `tools/dps_report/raw/` with the report data it produces, so the first partial regeneration has a store without another 95 min run.

## Checks before it ships

1. `hydrate` then `publish` with no ids reproduces the published datasets (the expected text differences aside, none on a fresh publish).
2. `clean` plus `run` of one unchanged build reproduces its pack byte for byte.
3. Remove a build, then add it back: the datasets end identical to the start.
4. A two-build partial run gives the same datasets as a full run at the same commit.

## What the skills say afterwards

regenerate-dps-report and dps-report B7: "Never re-sim every build for a change that touches only some. `stale.mjs` gives the scope; `hydrate`, `run <ids>`, `publish <ids>`. A full run is for shared sim inputs only, and skips `hydrate`. Text or page-data changes: `hydrate`, `publish`. Commit the datasets, `raw/` and `provenance.json` together, merge master by regenerating rather than hand-merging, then deploy from the pushed commit."
