# DPS report: dashboard data pipeline

These scripts produce every dataset behind the Forever level 60 DPS Report: the page in `ui/dps_report/`, its datasets in
`ui/dps_report/data/`, deployed to https://forever-dps.pages.dev (`make dps-report-deploy`).

```
binaries ─────► requests ──► sims ──────────► raw packs ◄──► datasets ──► report
 (bin/)          req/        out/ rv/ asm/     raw/<id>.tgz    ui/dps_report/data/   forever-dps.pages.dev
                             abl/ weights/ mana/
```

Each build's raw sim outputs are kept in `raw/<id>.tgz`, committed with the datasets they made, so a regeneration re-sims
only the builds that changed and still rebuilds every dataset, the cross-build ones included (see "Partial regenerations").

1. **Binaries.** `pipeline.sh bins` (`build_bins.py`) builds them from a clean worktree of this repository, with no
   patches (the sim reads the report switches itself, see "Report switches"), and stamps what they were built from.
   Every sim step refuses binaries whose stamp doesn't match the code.
2. **Requests.** `req/<id>.json` is the exact RaidSimRequest for each of the 36 level 60 presets. `req/<id>.meta.json` holds
   its class, spec, build name, preset race, talents, gear set, APL, published and verified DPS, summary, key talents, rotation,
   consumes and caveats.
3. **Sims.** `run_build.py`, `weights.py`, `run_races.py`, `ablate.py` and `mana_curves.py` run the binaries on the requests.
   After `build_data.py`, `events.py` sims the sample fights once more with their debug logs (see the datasets table).
4. **Datasets.** `build_data.py` (with `build_races.py`, `setup_active.py`, `rotation.py` and `tooltips.py`) writes `data/*.json`.
5. **Raw packs and publish.** `pipeline.sh publish` copies the datasets into `ui/dps_report/data/`, records provenance and
   writes the re-simmed builds' raw packs to `raw/`.

## Layout and configuration

| Path | What it is | In git |
|---|---|---|
| `dashpaths.py` | Where every script reads and writes. Uses the env vars below. | yes |
| `req/` | The 36 requests and their meta files (largest is 64 KB); `req/races/`, the races' own rotations and gear (see "Race overrides") | yes |
| `forever_tt_cache/` | Wowhead tooltip JSON cache, `forever/` and `classic/`, `item/` and `spell/` (2,126 files, 8.4 MB) | yes |
| `fortt_ids.json` | Every spell and item id the dashboard shows a tooltip for. Written by `whids.py`, read by `whcache.py`. | yes |
| `talent_ids.json` | Every spell a talent rank may link to (the Forever client's rank 1 spells and the trees' per rank spells). Written by `talent_links.py ids`, read by `whcache.py`. | yes |
| `tests/` | Unit tests for the timing shim, the probe key and skip, tankheal's resume decision, and the parallel parts, merge, planner and transport (no sims): `python3 -m unittest discover -s tools/dps_report/tests` | yes |
| `parallel.py` | A full regeneration over several cloud sessions at once: the plan, each worker's job, the transport and the merge (see "Parallel runs") | yes |
| `parallel/<run>/` | A parallel run's plan and the parts the merge fetched | no |
| `build_bins.py`, `test_build_bins.py` | Builds and stamps every binary, and the guard that refuses sims on binaries that don't match the code (see "Building the binaries"), with its tests | yes |
| `bin_override/` | Builds simmed with `DASH_ALLOW_BIN_MISMATCH`, and why (packed with them) | no |
| `raw/<id>.tgz` | Each build's raw sim outputs (about 2 MB for all), written by `rawstore.py pack` (see "Partial regenerations") | yes |
| `rawstore.py` | Packs, unpacks, cleans and checks the raw outputs, and copies the datasets into the report | yes |
| `probe.py`, `probe_cli.py` | The probe: every step's sims at a tenth of their iterations, fingerprinted, to tell which stale builds really moved (see "Probes") | yes |
| `probe/`, `probe_check/`, `probe_audit.json` | Probes of this run, of the last check, and the check's verdict (what publish records) | no |
| `timing.py`, `timing_cli.py` | Where a `pipeline.sh run` spent its time: every sim's CPU and wall time, logged by `timing_cli.py` standing in for the binaries and summed up by `timing.py` (see "Sim steps") | yes |
| `timing/` | One log per `pipeline.sh run` or `resume` (`<start>.jsonl`), the binary links the run used (`timing/bin`) and `phases.jsonl`, the wall time of each `hydrate`, `publish` and `tankheal.py bins` | no |
| `adaptive.py` | The luck goals and the fight counts sized to them: a pilot, then a top-up to what the paired spread needs (see "Luck goals") | yes |
| `vec/` | Per-fight DPS of the race, racial and assumption sims, keyed by request, which the top-ups extend; cleared per build by `run_build.py`, never packed | no |
| `changes.py`, `changes/` | The Changes tab: the updates, their public change lines and the results at each update's end (see "Changes tab") | yes |
| `rd/` | Headless check harness (Playwright) | scripts only |
| `bin/` | `wowsimcli`, `dashstats` (`tools/dashstats`), `tankbench`, `healbench` and `STAMP.json`, built by `build_bins.py` (see "Building the binaries") | no |
| `out/ rv/ abl/ weights/ names/ mana/` | Raw sim outputs | no |
| `icons/` | Download cache of wow.zamimg.com icons and talent-tree backgrounds | no |
| `data/` | The generated datasets | no |
| `page/` | The live page (`index.html`) plus `d3.min.js` for the harness. Fetched from the artifact; see "Checks harness" below. | no |

Environment variables:

- `DASH_DIR`: the base directory for `req/`, the cache and all outputs. Default: this directory.
- `DASH_BIN`: the directory with the binaries. Default: `$DASH_DIR/bin`.
- `DASH_TREE`: the merged tree. The binaries run with it as their cwd, and scripts read `assets/database/db.json`, talents,
  racials, `ui/` and `forever-wiki/` from it. Default: a `w-dash` checkout next to this repository (`../w-dash` from the repo root).
- `DASH_TT_CACHE`: overrides the tooltip cache directory.

If `DASH_DIR` points somewhere else, copy or symlink `req/` and `forever_tt_cache/` there.

Not included:

- **The raw outputs.** They are regenerated from the requests, as described below. Last full set: `out/` 4.6 MB, `rv/` 0.1 MB of
  results, `abl/ablate.json` 0.3 MB, `weights/` 0.15 MB, `names/` 0.36 MB.
- **`icons/`** (about 660 jpgs, 3.4 MB). `build_data.py` refetches any missing icon from
  `https://wow.zamimg.com/images/wow/icons/{medium,large}/<icon>.jpg`, and any missing tree background from the
  `backgroundUrl` in `ui/core/talents/trees/<class>.json`.
- **`spellname.csv`.** The scratch copy was an HTML error page, and no script reads it.

## Building the binaries

```bash
tools/dps_report/pipeline.sh bins           # or: python3 tools/dps_report/build_bins.py [--force] [--bin DIR]
```

`build_bins.py` builds every binary of the three modes into `$DASH_BIN` (default `$DASH_DIR/bin`) from a clean
worktree at `${SIM_COMMIT:-HEAD}`, so uncommitted edits never reach a binary. In that worktree it builds the packages
in the table at the top of `build_bins.py`, always regenerates the protos (`protoc`, else `npx protoc`, with
`protoc-gen-go` from `$(go env GOPATH)/bin`), and runs `go build -tags=with_db -trimpath -buildvcs=false` with
`CGO_ENABLED=0`, two jobs at a time (`BUILD_P`) under `nice`.

| Binary | Package | Used by |
|---|---|---|
| `wowsimcli` | `cmd/wowsimcli` | `run_build.py`, `run_races.py` (`DASH_RACIALS_OFF`), `run_assumptions.py` (`DASH_ASSUME`), `weights.py`, `ablate.py`, `mana_curves.py`, `events.py`, `verify.py` |
| `dashstats` | `tools/dashstats` | `run_build.py`, `race_gear.py` |
| `tankbench`, `healbench` | `tools/tankbench`, `tools/healbench` | `tankheal.py`, `bench_races.py` |

Every binary is built from the same code with no patches, so each reads every report switch (see "Report switches"):
one `wowsimcli` serves every DPS step, with the step's switches in its environment. A missing binary stops the step.

`bin/STAMP.json` records what the binaries were built from: the input hash, the commit, each binary's sha256, and the
Go, `protoc` and `protoc-gen-go` versions. The input hash covers exactly what goes into a binary: all of `sim/` but
`_test.go`, `.md`, `.results` and `testdata/`, `proto/`, `cmd/wowsimcli`, `go.mod`, `go.sum`, the Go files of
`tools/dashstats`, `tools/tankbench`, `tools/healbench` and `tools/tankscore`, all of `assets/database`, and the table
and flags themselves. It hashes git blob ids at the commit, as `stale.mjs` does, so a commit that changes none of these
gives the same hash and `bins` builds nothing. The build also checks with `go list -deps` that the binaries use no
package of this repository outside these inputs. With `-trimpath` and `-buildvcs=false` the same inputs give
byte-identical binaries at any commit, path or Go cache, so the race-vector cache of `bench_races.py`, keyed by the
binary's sha256, keeps hitting across commits.

On the shared 4-CPU box (2026-11-18) a cold build (empty Go cache) took 29s, a rebuild with a warm cache 6s, and a
`bins` with a matching stamp 0.2s. That is one worktree and one build per package, where the old recipe needed a
worktree for each patch group.

**The guard.** `pipeline.sh run` and `resume` (before the clean), `probe.py check` (`pipeline.sh probe`),
`tankheal.py run` (`pipeline.sh tankheal`) and `events.py` when it sims (`pipeline.sh publish <ids>`) refuse to start when:

- `$DASH_BIN/STAMP.json` (or `$TANKHEAL_BIN`'s) is missing,
- its input hash differs from the inputs at `${SIM_COMMIT:-HEAD}` (the message names what changed),
- a binary they run is missing or its sha256 differs from the stamp's,
- or the working tree has uncommitted changes under the stamped inputs.

`hydrate`, `plan` and a plain `publish` run no sims and aren't checked. To sim anyway, set
`DASH_ALLOW_BIN_MISMATCH=<reason>`: the run prints a loud warning and records the reason with what it simmed,
`$DASH_DIR/bin_override/<id>.json` for a build and `<mode>/bin_override.json` for a Tanking or Healing mode. The raw
packs carry these files (hydrate restores them), and `stale.mjs --record` and `--record-bench` refuse to mark those
builds or modes current until a `run` without the override re-sims them. A missing binary can't be overridden.

`mksettings` and `rotopt` make requests, not report numbers, so they aren't stamped. Build them from the checkout when
you need them: `go build -tags=with_db -o tools/dps_report/bin/mksettings ./tools/rotopt/mksettings`.

The binaries run with `$DASH_TREE` as their cwd, and the scripts read `db.json`, talents and racials from it. This
checkout works as `$DASH_TREE`. The `w-dash` tree of the first reports (every
class branch merged, the db regenerated) is no longer needed to build anything.

### Report switches

The sim reads these environment variables for the report (`sim/core/report_switches.go`). Each one does nothing unless
it is set, so every `wowsimcli`, `tankbench` and `healbench` built from master runs the plain sim without them. They
are read once when the binary starts.

| Variable | What it does | Set by |
|---|---|---|
| `DASH_NAMES=<file>` | Each sim's finalize appends every unit's auras to the file, one `spellId itemId otherId tag label` line each (one copy per sim split). | `run_build.py` (`names/<id>.tsv`, the aura and proc labels in `names`, `abilities`, `auras` and `timeline_*`) |
| `DASH_SPLITS=<n>` | Splits a sim's iterations into n concurrent parts instead of one per CPU. The splits' seeds continue each other, but averaging the parts moves the last bits, so a fixed n keeps the raw outputs the same on every machine. | `pipeline.sh` (4, every step), `tankheal.py` (4, every bench call), `rawstore.py` and `tankheal.py unpack` (refuse packs made with another count) |
| `DASH_RACIALS_OFF=all` or `=<part>[,<part>]` | Leaves those Forever racial parts out: spirit, ruthless, precision, beast, dodge, elune, shadowmeld, mind, eureka, command, bloodfury, touch, endurance, berserking, windblessed, insight, leyline, and the priest racial spells chastise, starshards, darksacrifice and shadowguard. | `run_races.py` (`off`, `off:<part>` and `weaponNoRacial` variants), `bench_races.py` |
| `DASH_RACIALS_OFF_CLASS=<Class>` | Keeps `DASH_RACIALS_OFF` to the players of one class (`Paladin`, `Priest`, ...), so a healing preset's racials go while the healbench raid's tank and damage dealers keep theirs. Alone it does nothing. | `bench_races.py` (the preset's class) |
| `DASH_ASSUME="key=value,..."` | Overrides the constants the assumption scenarios vary (`assumptions_def.py` lists the keys). Each knob defaults to the constant master sets. | `run_assumptions.py` |

`go test ./sim/core/ -run 'Report|Dash'` and the switch tests in `sim/priest/shadow` and `sim/warrior/dps_warrior`
check each switch. `run_assumptions.py --identical` checks that a run with `DASH_ASSUME` unset matches a plain binary.
`tools/dashstats` (the `stats` dataset's stat sheets) is a plain tool in the tree.

## Requests

Each `req/<id>.meta.json` `notes` field records how its request was made. The usual way is
`bin/mksettings -spec <spec> -gear <gearSet> -apl <apl> -talents <talents> -race <race>`, run with cwd `$DASH_TREE`, with the
duration set to 180 s and the preset's encounter variation. A few builds then get small JSON edits, such as a variant's own
consumes or the pet. The `-h` flag lists the specs. Rules: use the preset race and the build's own consumes, buffs, pet and imbues. Any
`simOptions` will do, because the scripts override them.

To refresh a request after a class change, rerun mksettings with the same arguments, reapply the edits listed in its notes, and
check it:

```bash
python3 verify.py req/<id>.json [iters=20000] [seed=1] [Race]   # must match the published 180 s DPS within ±2
```

Then update `meta.published` and `meta.verify`, and any `summary` / `caveats` / `consumes` text the class changed, from the
class PR bodies (bayov/forever-sim #1 to #9). An encounter preset (one with `encounter` and `baseBuild`) also has an
`advantage` line. It says in a sentence or two why the preset beats its normal build on that creature type of boss, like
Murder's 4% against Humanoids and Giants, and an `advantageShort` of 25 words at most that the page shows on hover over
the preset's "vs Humanoid" badge (the long line stays in the data). Its `advantageRefs` list the abilities and talents
`advantageShort` names, as game tooltip refs: `k:s10314` for a spell key, `tal:Rogue:0:4` for a talent by class, tree and
index in `talenttrees.json`. The badge's tooltip puts each one's icon in front of its name and its own tooltip below.
A duty preset (one with `duty` and `baseBuild`, standards 23) keeps a raid buff or debuff up itself, with the raid's copy
off. It gets `run_build.py` only (main, race sweep, lengths, samples): the page lists it in the Ranking while "Duty" is on,
with what it costs against its base, and `build_data.py` keeps it out of every per-preset detail dataset so the page stays
small. Its assumption factor is its base's.
A pet variant (one with `pet` and `baseBuild`, a hunter preset with another pet for what that pet gives the raid) is
handled as a duty preset is: `run_build.py` only, listed in the Ranking while "Duty" is on with its pet as the badge
and what the pet costs against its base, and its base's assumption factor.
Each assumption in `assumptions_def.py` is written for a player who hasn't read our notes: a question-style `label`
(8 words at most), a `short` line (20 words at most) the Ranking panel shows under it, setting labels that say what each
choice means with a `rationale` of 15 words at most, and a `desc` of 2 to 4 sentences plus its `src` (written to the
data as `source`) for the Method tab. The Healing page's assumptions are `HEAL` at the end of the same file, written the same
way with `builds` naming the Healing presets: one-value rows only, never simmed (an alternative's measured cost goes in
`desc`), which `tankheal.py build` writes as `heal_assumptions` and the Healing Ranking shows above its ranking.

### Race overrides (`req/races/`)

A build's race sweep sims every race with the preset's rotation and gear, except where a race has its own (standards 29):
`req/races/<id>.<Race>.json`, with `<Race>` the proto race name without `Race` (`Dwarf`, `NightElf`, `SkyborneHighOrder`).
It holds an APL `rotation` (the request's `player.rotation` shape), `bonusStats` (`proto.Stat` names, the sim's units,
added on top of the gear), a `note` for the Races tab, the `gain` over the preset's setup at that race (180 s, seed 1,
20,000 iterations, paired seeds; a file exists only for a gain of 2 standard errors or more) and `how` it was made. A
file has a rotation, bonus stats or both. Gear hit and crit are one pool under Forever (`unifyEquipHitAndCrit` adds the
melee and spell values, bonus stats included), so a file lowers or raises each once, as `MeleeHit` or `SpellHit`, never both.

`race_override.py` applies them, for every sim of a race other than the preset's (the preset race's request is the
preset): `run_build.py`'s race sweep, every `run_races.py` scenario (`off` and `off:<part>` too, so the breakdown measures
only the racials), and `verify.py <req> <iters> <seed> <Race>`. The other steps sim the preset race only. A duty or
encounter variant without a file of its own takes its base build's, but only what fits it: the rotation when the
variant's rotation is the base's (a duty preset's rotation keeps its duty up), the bonus stats when its gear is the
base's. `python3 race_override.py [ids]` lists what every build sims at each race. The race sweep records each override
it used in `out/<id>.json` (`races[].override`); `build_data.py` puts its note in `races` and `race_variants` as `tuned`,
and the Races tab shows it after the race's name and in the row's tooltip. A race's own rotation is also the
`race_rotation` dataset (`rotation.race_build`), which the Specs page's Rotation card shows for that race and weaves
into its All races view.

The report's words for a race's own setup are in `race_text.json`, keyed by override file (`<id>.<Race>`): its `note`
for the Races tab (in place of the file's own) and, for a rotation, a `why` for each line it changes
(`"<prepull|priority>:<spell key, or the action's name>"`). They live apart from the override files because the sims
read those and the staleness checks hash them, so rewording them would make the builds stale. A new override
needs a `note` and a `why` for each line it changes; `build_data.py` warns about a changed line with no `why`. `stale.mjs` and the packs' fingerprints
hash a build's override files (its base build's too), so adding, changing or deleting one makes the build stale.

`race_gear.py [ids] [--write] [--to-cap]` makes each race's gear files (standards 29): for every build and every race it
runs at other than the preset's, the race's hit (dashstats' final sheet, plus a priest's school talent) against the
target (the standards 4 cap, 16% spell for casters and 8% yellow for melee and hunters, or the preset race's own hit
where the preset carries more; `--to-cap`: the cap). Where the set has hit-trade steps the race wears the step that
lands it on the target, in either direction: a lower step (a race over it) is kept after a paired 20k sim where it gains
2 SE, a higher step (a race under it) by the rule. Without steps, the Dwarf alone re-spends its excess (at most 1%, at
most the gear's own hit) as bonus stats over the set's other non-Stamina, non-Spirit stats at the 1b prices, simmed the
same way. `--write` writes the kept ones (merging the gear into a file that holds the race's rotation) and removes the
gear of earlier runs that no longer apply; every build and race's numbers go to `$DASH_DIR/race_gear.json`. Run it with
`DASH_TREE` and `DASH_BIN` pointing at the tree and binaries built from it.

## Sim steps

Each step is restartable: it skips work whose output already exists. To redo a build, delete its output first.

| Step | Command | Sims | Writes | Runtime, 4 CPUs |
|---|---|---|---|---|
| main + races + lengths + samples | `runner.sh` (start up to 4) or `python3 run_build.py <id>...`; `--resample <id>` redoes only the sample fights | Per build: main 180 s, 20k iterations, seed 1, all values saved. Race sweep: every race the class can be, each with its race override if any, at the fights the race goal needs (2k pilot, up to 40k). Lengths 60/120/180/300/480 s, 5k. 30 one-iteration rolls, then 3 logged sample fights at the 10th, 50th and 90th percentile. dashstats stat sheet. | `out/<id>.json`, `names/<id>.tsv` | about 22 min for all 36 (2 to 5 min per class) |
| stat weights | `ls req/*.meta.json \| xargs -n1 basename \| sed 's/.meta.json//' \| xargs -P4 -n1 python3 weights.py` | Per build: the baseline plus 6 to 8 bonus-stat runs (hit also measured downward), 180 s, 40k iterations, seed 1 | `weights/<id>.json` | about 30 min |
| race scenarios | `python3 run_races.py [ids]` (`PAR=4` threads) | Per default build × race (encounter presets get only the race sweep): `off`, `off:<part>`, `beast`, `elemental`, `humanoid` (Priests only: Dwarf Chastise), and `weapon` / `weaponNoRacial` where the racial weapon is legal and neither hand holds it. The swap changes one weapon (the one whose speed changes least, the off hand on a tie) to the racial's type, at the nearest Phase 1 BiS speed of that type. Each race with its race override if any, except that a sim leaving out the racial a race's own gear is tuned around (`GEAR_PART`: Dwarf Artisan's Precision) wears the preset's gear, so the gear trade counts in that racial. Seed 1, at the fights the racial goal needs (1k pilot, up to 40k), paired with the race's sweep fights. The last full run was 1,084 sims. | appends to `rv/results.jsonl` (keyed build, race, variant) | about 35 min for all; the 458 beast and elemental reruns took about 16 min |
| mana curves | `python3 mana_curves.py [ids]` (`FIGHTS=100 PAR=8`, `--force` redoes builds that already have an output) | Per mana-using build (every class but warrior and rogue): 100 logged one-fight runs at a fixed 480 s with no duration variation, seeds 1001 and up, the player's mana read from the log every 5 s. Plus one 5k run for the sim's seconds out of mana, time to OOM and mana left at the end. | `mana/<id>.json` | about 17 s for the 6 shaman builds |
| setup ablation | `python3 ablate.py [ids]` (`ITERS=300 SEED=11 PAR=4`; `ABL_MERGE=1` keeps the other builds' entries) | Each preset with one consume, buff, debuff or profession left out, plus the non-stacking pairs. 1,957 sims. | `abl/ablate.json` | about 5 min |
| assumption scenarios | `python3 run_assumptions.py [ids]` (`PAR=4`) | Per build an assumption touches (`assumptions_def.py`): the base, every non-default setting alone at 180 s (at the fights the assumption goal needs: 1k pilot, up to 40k) and 120 / 300 s (5k), every combination at 180 s where a build has 3 or fewer assumptions, and 12 random combinations otherwise (the multiplicative estimate's check). About 1,100 sims. | appends to `asm/results.jsonl`; `build_data.py` (via `asm_data.py`) writes `assumptions` and `assumption_dps` | about 1.5 h for all |
| tooltip ids | `python3 whids.py` (after a first `build_data.py`) | none | `fortt_ids.json` | seconds |
| tooltip fetch | `python3 whcache.py fortt_ids.json` | none. Fetches from `nether.wowhead.com/{forever,classic}/tooltip/<kind>/<id>`, only ids missing from the cache. | `forever_tt_cache/` | about 13 min cold (about 1,400 requests at 0.35 s); seconds when warm |
| datasets | `python3 build_data.py` | none | `data/*.json`, `fortt_coverage.json` | about 10 s warm; about 1 min more when icons must be fetched |

Every sim the steps run sets `simOptions.useLabeledRands` (`RNG` in `dashpaths.py`, since 2026-11-04): each kind of
roll gets its own random stream, so two sims of the same seeds that differ in one setting still share every other
roll. That makes the paired comparisons (race against race, racials off, assumptions, weights, ablation) need about 9x
fewer fights for the same error, for about 24% more time per fight. It changed every published number by about its
error, so data from before it and after it don't mix: the switch came with a full run.

### Luck goals

Each comparison the page shows sims as many fights as it takes to meet its luck goal (`adaptive.py`, since
2026-11-04; the goals and why are in `orchestration/followups.md`, "DPS report luck goals"). The goal is one standard
error of the difference the reader compares, and every sim uses seed 1, so the two sides of a comparison fight the same
fights and the error is the paired one, sd(b_i - a_i) / sqrt(N):

| Comparison | Goal | Step | Pilot |
|---|---|---|---|
| A race against the next race of the same preset | 0.35 DPS | `run_build.py` race sweep: every race of the preset gets the same N, from the widest spread between races next to each other | 2k |
| A racial (or all of them) left out, another boss type, against the race | max(0.25 DPS, 10%) | `run_races.py`, each scenario against its race's sweep fights | 1k |
| An assumption setting or combination against the preset, 180 s | max(0.25 DPS, 5%) | `run_assumptions.py`, against the main run's fights | 1k |

From the pilot's paired sd, N = 1.2 x (sd / goal)^2, rounded up to a whole number of pilots, up to 40k (`ADAPT_PILOT`,
`ADAPT_RACE_PILOT`, `ADAPT_MAX`), and both sides are topped up to N once: the top-up is a run from seed 1 + pilot, so
the fights are exactly a seed-1 run of N. A row's DPS is its reference's (the main run's 20k mean, or the race's
sweep) plus the paired difference, so a setting that changes nothing gives the reference exactly, and the preset race's
row is the preset's DPS. The rows carry `iters` and the paired error: `pse` (a race against each other race, in
`races` and the `race_variants` baseline rows) and `pseBase` (a scenario against its race). The page's race ties use
it (`raceSe` in report.js). `run_build.py` prints a race pair still over its goal after the top-up. The spec DPS
(20k, error about 0.35), the weights (paired, their own top-up rule) and the fight lengths (5k) keep their counts.

The runtimes above are from the runs that first measured them, before later changes to the scripts. `pipeline.sh run`
and `resume` now measure every run: the binaries run through `timing_cli.py`, which logs each call's step, class, race,
iterations, fight length, targets, CPU seconds and wall seconds to `timing/<start>.jsonl`, and the run ends with
`timing.py`'s summary of that log. `tankbench` and `healbench` run through it too: `tankheal.py` logs the main runs
under the steps `tank` and `heal` and `bench_races.py` the race sims under `tank_races` and `heal_races`, with the bench's
role, boss, iterations, number of request files and fight length read from its flags. The shim prints nothing and passes
the bench's output and exit code through, because `tankheal.py` keeps that output as the mode's raw text. Whatever
hashes or checks a binary follows the shim to the real one in `TIMING_REAL_BIN` (`build_bins.real_bin` and `real_dir`):
the guard and its `STAMP.json`, `bench_races.py`'s cache key and the probe's `_key`. It shows:
- **Per step:** wall time, the CPU time of its sims, and how many cores those kept busy. A busy figure well under P
  means the step was waiting on its last builds or on its own Python. The `tankheal` step lists its four bench steps
  beneath it.
- **Sim kinds:** the costliest kinds of sim, grouped by step, binary, iterations, fight length and targets (benches:
  role, boss, iterations and request count).
- **Phases:** the wall time of the work that is no sim, from `timing/phases.jsonl`: `pipeline.sh hydrate` and `publish`
  and `tankheal.py bins` (the binary build) each log one line when they finish, so the latest of each shows here.
- **Per class:** core-minutes per base build. This is the unit of pipeline.sh's cost table `W`, which orders and shards
  builds, so the summary is what to refresh `W` from.

`python3 timing.py [log]` shows a past run's log again.

Diagnostics, which produce no datasets: `whdump.py` prints Forever and Classic tooltip text side by side for the setup chips,
and `wagocheck.py` compares the real items' ilvl and stats across Forever wowhead, the Forever client and the sim's db.

## Datasets

All are written by `build_data.py`. "out" means `out/<id>.json` from `run_build.py`.

| Dataset | Made by | Depends on | Runtime |
|---|---|---|---|
| builds | build_data.py | out (main 20k, race sweep, 120/300 s lengths), req meta, ranks across all builds | run_build |
| races | build_data.py | out race sweep: every legal race, 180 s, its fights; `pse` the paired error to each other race | run_build |
| lengths | build_data.py | out lengths: 60, 120, 180, 300 and 480 s at 5k | run_build |
| hist | build_data.py | out main run: all 20k fight DPS values, 5-DPS bins | run_build |
| abilities | build_data.py | out main run per-action metrics, names | run_build |
| auras | build_data.py | out main run aura uptimes and procs (player and target), names | run_build |
| timeline_\<class\> (×9) | build_data.py | out sample fights: 3 logged one-iteration fights per build, picked from 30 seeds. Only their auras and pre-pull casts, for the aura bars. | run_build |
| events/\<build\>.json.gz | events.py | the timeline datasets' seeds: each sample fight simmed again with its whole debug log, as a gzipped RaidSimRequest and RaidSimResult for the sim UI's events view | events.py (seconds) |
| event_names, event_icons | events.py | every spell and item id in the fights, their metrics and weapons, named from `db.json` and then the wowhead cache, and the icons that icons/icons_more lack | events.py |
| stats | build_data.py | out `stats`: dashstats `ComputeStats` final sheet | run_build |
| gear | build_data.py | the request's equipment, item names, icons and ilvl from `db.json` and wowhead item csv | none (no sims) |
| setup | build_data.py + setup_active.py | the request's consumes, buffs, debuffs and professions; `abl/ablate.json` for `active` / `why` | ablate.py |
| notes | build_data.py | req meta keyTalents, rotation, consumes and caveats, plus the gear rules text in build_data.py (snapshot-cleaned) | none |
| names | build_data.py | every key seen; `names/*.tsv` (DASH_NAMES), `db.json`, `assets/db_inputs/wowhead_*_tooltips.csv`, `forever-wiki/data/abilities.json`, talent trees, racials, the sim's rank arrays; colour from the icon (PIL) | run_build |
| rotation | build_data.py + rotation.py | the request's APL (prepull and priority), cast counts from abilities, names, `racials.json`; a talent a condition checks is `[[tal:<class>:<tree>:<index>\|<name>]]` in `when`, which the page draws with its icon and tooltip | none |
| race_rotation | build_data.py + rotation.py (`race_build`) | each race's own APL in `req/races/` (`race_override.load`), lined up with the preset's: `forRace`, `diff` (same, changed, added, removed), `base` (the preset line's order) and `why` (`race_text.json`). No casts | none |
| race_variants | build_races.py | `rv/results.jsonl` plus the out race sweep (baseline); `tuned`, the note of the race's override | run_races |
| race_breakdown | build_races.py | `rv/results.jsonl` (`off`, `off:<part>`, weapon and target deltas) | run_races |
| race_summary | build_races.py | race_variants by scenario, scope (overall or faction) and role | run_races |
| race_notes | build_races.py | scenario and part texts, a finding per class from race_variants | run_races |
| weights | build_data.py | `weights/*.json` (40k per stat) | weights.py |
| mana | build_data.py (`mana_curves.datasets()`) | `mana/*.json`: per build and 5 s step, the average mana % of max over 100 fights at 480 s, with the 25th and 75th percentile fight | mana_curves.py |
| mana_summary | build_data.py (`mana_curves.datasets()`) | `mana/*.json`: when mana first drops below 15% (`lowPct`) (median, p25, p75, null when that fight quantile never does), share of fights that do, mana % at 60/180/300/480 s, spend and regen per second, and the 5k run's `secondsOomAvg`, Tto and mana at the end | mana_curves.py |
| tooltips | build_data.py + tooltips.py | `forever_tt_cache` (Forever wowhead, Classic fallback), the tree's wowhead csvs, `db.json` (items, the tanks' and healers' synthetic items, enchants), `sim/*/item*sets_pve.go`, `sim/*/forever_synthetic_trinkets.go`, Forever wiki, `racials.json`, talent trees | whcache.py |
| talenttrees | build_data.py | `ui/core/talents/trees/<class>.json`, plus wowhead links from the tooltip cache | none |
| treeart | build_data.py | the tree icons from `tools/forever_talents/client/<class>.json`, and backgrounds fetched from each tree's `backgroundUrl` | icon fetch |
| icons | build_data.py | every icon used, from wow.zamimg.com (36 px medium, 56 px large for headers), plus `assets/item_slots/*.jpg` | icon fetch |

Side outputs: `fortt_coverage.json`, the count of tooltips by text source, for reporting. The page also expects every dataset
under 8 MB; the largest are icons and treeart at about 1 MB.

## Tanking and Healing

The report's Tanking and Healing modes come from `tankheal.py`, not from `run_build.py`: tools/tankbench and
tools/healbench print markdown tables, and `tankheal.py` keeps those outputs and parses them into datasets. Each
dataset has `"mode": "tank"` or `"heal"` in `datasets.json`; the ones every mode reads (names, icons, icons_more,
tooltips, talenttrees, treeart) have `"mode": "common"`, and no mode means DPS.

What is simmed (the docs say how, `tankheal.py` does not choose):

- Tanking: the 12 presets in `tools/tankbench/req/` (`warrior_protection.json` is the healbench's default tank, not a
  preset), `tankbench -iters 1000 -encounter tools/encounters/<boss>.json` for each of the 7 bosses, the reqs' own seed
  (`simOptions.randomSeed`). A score is the mean over the six scored bosses (Ragnaros is shown, not scored; docs/tank_sim.md),
  weights max threat 1/3, balanced 1, max survival 3.
- Healing: the presets in `tools/healbench/req/`, `-carry`, `-role tank` for the tank healer files and `-role raid` for
  `_raid`, 1,000 iterations, each file at its own length, healbench's fixed seed 101, and the mean over the 7 boss files
  (docs/healing_sim.md, the "Carry, 1,000 iterations" table and the shaman's "raid mean over these 7 files").
- Races (`bench_races.py`, into `<mode>/races.jsonl`; `--no-races` skips them): every race the class can be
  (`CLASS_RACES` in run_build.py), and each race with every racial left out (`off`) and, for a race with more than one
  part that can count, each part left out (`off:<part>`: the DPS parts plus Night Elf `dodge`; the beast and elemental
  parts are skipped, since no bench boss has a creature type). The benches write each fight's values (`-vectors`):
  tankbench its SafeDPS, the healing it lived through, DTPS and TPS; healbench, at a list of damage scales (`-scales`),
  each fight's deaths (`fight_deaths`), HPS, useful HPS and mana left. Every side runs seed 1 with labeled random
  streams (`-seed 1 -labeled`), so fight i is the same fight for every race and racial, and adaptive.py's `values`
  keeps the vectors (`vec/bench_tank`, `vec/bench_heal`) and tops them up. Fights are sized like the DPS race sims,
  with the paired pilots and `BENCH_GOALS` (adaptive.py's docstring says how they follow from the DPS goals): races
  pilot 2,000 fights, then go to what the widest gap between neighbouring races needs (tank up to `CEIL`, 40,000;
  heal up to `HEAL_CEIL`, 6,000), racials pilot 1,000 and go to what their difference needs. A tank's score per fight
  set is the bosses' mean of 2 x SafeDPS less the healer slots the 95th percentile of its healing needs; a healer's
  carry per boss walks the lattice of scales 1.03^j from the main run's carry and interpolates where the deaths cross
  the line (tank healer: tank deaths 2%; raid healer: any death 5%). The errors are a paired bootstrap (400 resamples
  of the fights, shared by every side). The build puts the races on the main run's numbers (each race moved by the
  preset race's gap) and splits a race's gap over its base stats into its parts (`_race_breakdown`: Shapley for two,
  scaled leave-one-out for three or more, like build_races.py).

Commands (`DASH_DIR` as for the DPS steps; with no `$DASH_DIR` tree they read the repository itself):

```sh
pipeline.sh tankheal [tank|heal]   # = TANKHEAL_BIN=<bin dir> tankheal.py run: sims into $DASH_DIR/{tank,heal}/, run.json, .fresh
pipeline.sh tankheal --stale       # the main benches whole, race sims only for the presets stale.mjs calls stale
tankheal.py run heal --presets priest_holy,priest_holy_raid   # the same for the presets named
tankheal.py run heal --resume      # a stopped run keeps the bench calls it finished (.progress.json), if HEAD is its commit or the bench inputs are the same
tankheal.py build                  # $DASH_DIR/{tank,heal}/ -> $DASH_DIR/data/{tank,heal}_*.json
tankheal.py pack | unpack          # tools/dps_report/raw/bench/{tank,heal}.tgz <-> $DASH_DIR/{tank,heal}/
tankheal.py fresh [--clear]        # the modes simmed since their last publish (<mode>:<stem,...> when some were kept)
pipeline.sh benchprobe [heal]      # = tankheal.py probe check: which stale presets really moved ("Bench probes" below)
tankheal.py probe heal [--presets <stem,...>]   # the probes the pack keeps, into $DASH_DIR/heal/probe/ (run makes them too)
tankheal.py run heal --main-only   # a part ("Parallel runs"): the main bench alone, part.json instead of run.json and .fresh
tankheal.py run heal --carries <part.tgz> --presets <stem,...>   # a part: those presets' race sims from another part's main bench outputs
tankheal.py run tank --part        # a part: the whole mode, merged later
tankheal.py part <mode> <dest.tgz> # pack a finished part (its outputs and part.json)
tankheal.py merge <mode> [--keep-rest] <part.tgz|dir>...   # parts -> $DASH_DIR/<mode>/ as one run writes it, run.json and .fresh
```

`--resume` reuses a mode's `.progress.json` when the run started at the current HEAD, or at another commit whose bench
inputs are the same: `stale.mjs --bench-fingerprint <mode> --commit <rev>` prints the per-preset fingerprints publish
records for the mode (`{<stem>: inputs}`, the same `benchFingerprint` that `--record-bench` stores), and `.progress.json`
keeps their hash (`inputs`) next to the commit the run started at. So a master that moved on
mid-run without touching the sim doesn't redo the run. The commit in `run.json` and `.fresh` is the one the run started
at. A progress file from before this, with no `inputs`, resumes only at its own commit, and so does any run when node
or the commit isn't available. A finished run (`run.json`) is always redone.

The raw outputs are kept in the repository like the DPS ones (`raw/bench/<mode>.tgz`, deterministic tars), outside
`raw/*.tgz` so `rawstore.py prune` never takes them for removed builds. The pipeline wires them in: `run` and `resume`
sim both modes after the DPS builds (all of them when no DPS build is stale or `TANKHEAL=1`; `resume` skips a mode
already marked fresh and passes `--resume` to the others), `hydrate` unpacks and builds them, and `publish` builds them before `rawstore.py copy`, then for a
fresh mode records its provenance (`stale.mjs --record-bench <modes> --iterations 1000`, into `provenance.json`'s
`benches`), packs it and clears the mark.

Staleness is per preset, like the DPS builds. `stale.mjs` reports the stale presets of each mode separately from the DPS
builds (`benches` in `--json`, so probe.py's ids are unchanged). A preset's fingerprint covers the shared sim code and
the top-level `sim/*.go`, the bench's sources (its `req/` left out), `tools/tankscore`, the seven boss files,
`tankheal.py`, `bench_races.py`, `adaptive.py`, `dashpaths.py` and `run_build.py` (its `CLASS_RACES`), then its own request and the code of every class in its raid with the db items those requests wear. A tank
preset's raid is the tank alone. A healing preset's raid is the healer and healbench's frost mage fillers, plus the
warrior tank for a tank healer (a raid healer's raid has no tank). So a `sim/shaman` change stales the shaman presets, a
`sim/mage` change every healing preset, and a `sim/core` change every preset.

A whole mode costs about 20min for Tanking and 1.5h to 4.3h for Healing, almost all of it the race sims. So
`tankheal.py run --stale` (or `--presets <stem,...>`) runs the main bench whole (one call covers every preset of a
role, about 15min to 20min), then the race sims only for the stale presets. Every other preset keeps its rows from the
mode's pack, but only when its main-run rows on every boss (what its race sims start from, `main_carries`) are the
same bytes as the pack's: a preset whose main rows moved is re-simmed too. The run sorts `races.jsonl` by preset, so
the pack is the same bytes for the same results, and `tankheal.py fresh` then prints `<mode>:<stem,...>`, the presets
simmed, so `publish` records only those (`stale.mjs --record-bench heal:priest_holy,...`). The kept presets keep their
records. A record from before per-preset staleness (one for the whole mode) is compared at its own commit, like an
old DPS record.

Datasets per mode: `<mode>_builds`, `_bosses`, `_encounters`, `_races` (with `pse`, the paired error of each race's
gap to every other race), `_race_breakdown` (base stats and each racial part, `v` and `se`, in the preset's score),
`_gear`, `_rotation`, `_setup` (the DPS shapes),
`_tooltips` (rows the common tooltips lack), `_notes`, `_method`, `_icons` and `_names` (what the common icons and names
lack), and for Healing `heal_assumptions` (`assumptions_def.HEAL`, the DPS `assumptions` shape).

The words players read come from `tankheal_text/tank.md` and `heal.md`, written for players, never from the docs
(docs/tank_*.md and healing_*.md are the developers' record and the facts behind the text). Each file has a summary
line per preset (`_builds` `summary`), the "Good to know" notes per class or preset (`_notes`), a line per boss
(`_encounters` `notes`, the boss list in Method), the Method cards (`_method`; the exact formula only in the last card),
and regex rewrites for the requests' rotation notes (`_rotation` `notes` and `why`). Their header comment has the
format; a missing preset or boss stops the build. Markdown goes through `md2html.py`, a small subset (no Markdown
library is installed). Then `tankheal.py`'s leak check reads every string the page can show and stops the build on a
tool, file or process word (`LEAK`: tools/, docs/, file extensions, bench, rotopt, seed, iterations, template,
request, APL, code spans, ...): reword the text file, don't widen `LEAK_SKIP`.

Building the benches: `pipeline.sh bins` (or `tankheal.py bins`, the same `build_bins.py`) builds them with every
other binary, from the same code. The race sims use the racials switches (`DASH_RACIALS_OFF` and
`DASH_RACIALS_OFF_CLASS`, see "Report switches"), and every bench run sets `DASH_SPLITS=4` as the DPS steps do.
`run.json` records the count, and `tankheal.py unpack` refuses a pack made with another one (a pack from before has
none and counts as 4, one part per CPU on the 4-CPU machine that made it).
`tankheal.py run` refuses benches the stamp doesn't match to the code (see "Building the binaries").

The main bench calls take about half a minute for tanking and a few minutes for healing; the race sims take most of
the run: on a 4-core machine about 15 minutes for tanking (2026-11-04: 12 minutes, some race fights already kept)
and about 3 hours for healing (most of it the raid healers' race sims, 25 to 95 minutes a preset).

### Bench probes

A `sim/core` change stales every Tanking and Healing preset, though most come out the same. `pipeline.sh benchprobe`
(`tankheal.py probe check`) finds the ones that don't, the way "Probes" below does for the DPS builds:

- It runs each stale preset's main bench (every boss) and its race sims in a scratch folder (`probe_work/`, so the race
  sims start with an empty vector store), at a tenth of the full sizes: the main bench at 100 fights, and every pilot
  and ceiling of `adaptive.py` divided by 10 (race pilot 200, pilot 100, ceiling 4000, healing ceiling 600). Each preset
  runs in its own child process, `TANKHEAL_PROBE_P` presets at a time (default 2), each sim split `DASH_SPLITS` ways.
- The probe of a preset is a fingerprint of each boss's main rows and of each race row (`probe_cli.py`'s `canon` and
  `digest`), plus the carry or score of each row. The sim reseeds every fight, so unchanged code and data give the same
  fingerprints. A probe is only ever compared with another probe, never with the full run: the healbench `-carry`
  bisection and the race walk pick their next scale from the deaths seen so far, so a shorter run sims other scales.
- The probe is optional. It is worth turning on only when a mode is mostly stale and a full re-sim of it would take well
  over 2h, so the next check can save that time. `tankheal.py run` probes the presets it ran the races for only with
  `TANKHEAL_PROBE=1`, into `<mode>/probe/<stem>.json`, and the pack keeps them. `tankheal.py probe <mode>` and
  `probe check` work on demand either way. With no probe in the pack, `probe check` says so and exits 1. A check compares the new probe with the pack's: the same means the
  full sims would come out the same, anything else names the bosses or the share of race rows that moved.
- Every probe carries a `_key`, the same idea as the DPS probe's: the probe settings (sizes, splits, architecture), the
  sha256 of the real bench binary (`build_bins.real_bin`, behind the timing links), the preset's request, the boss
  files, the report switches set in the environment (`DASH_NAMES`, `DASH_SPLITS`, `DASH_RACIALS_OFF`,
  `DASH_RACIALS_OFF_CLASS`, `DASH_ASSUME`), the healing raid's filler requests and the scripts the probe runs (`tankheal.py`, `bench_races.py`,
  `adaptive.py`, `dashpaths.py`, `run_build.py`, `probe_cli.py`). `run` keeps a probe whose key matches, copies one from
  `probe_check/bench/<mode>/` or the pack when that has it, and probes the rest (`PROBE_FORCE=1` probes all). `compare`
  ignores the key.
- A change only one fight in thousands reaches can hide from a probe. So a check picks one cleared preset per mode for
  a full re-sim anyway (the audit, picked at random with HEAD as the seed). `publish` then checks that the run kept the
  audit preset's main and race rows and every cleared preset's main rows (the main bench always runs whole) exactly as
  the pack had them. If one moved it stops: re-sim every cleared preset. Otherwise it records the cleared presets as
  current (`stale.mjs --record-bench heal:<stem,...> --by probe`), and their race rows stay as they are. Like
  `--by probe` for the DPS builds, it refuses a mode the check ran on mismatched binaries (`bin_override.json`).
- A pack without probes is re-simmed. To start the probes for a mode without re-simming it, run `tankheal.py probe
  <mode>` right after a `tankheal.py run` of the same code (it refuses a stale preset the last run didn't sim), then pack.

```bash
pipeline.sh hydrate
node stale.mjs                                   # stale Tanking and Healing presets
pipeline.sh benchprobe heal                      # prints what moved and the audit preset, writes probe_check/bench/heal/audit.json
tankheal.py run heal --presets <moved>,<audit>   # with TANKHEAL_BIN set, as pipeline.sh tankheal does
pipeline.sh publish                              # checks the audit, records the cleared presets --by probe
```

Measured on 4 cores (2026-11-18, one preset at a time, twice each with the same fingerprints): a raid healer
(`priest_holy_raid`) 565s wall and 1,500s CPU, a tank healer (`shaman_restoration`) 120s wall and 260s CPU, a tank
preset (`warrior_protection_balanced`) 80s wall and 145s CPU. So a check of a whole mode at the default 2 presets at a
time takes about 30min to 48min for Healing (4 raid and 5 tank healers, about 7,300s CPU), against about 87min for a
full run, and about 8min to 16min for Tanking (about 1,750s CPU), against about 50min. That is far more than a tenth:
each probe sim still starts the bench, loads the raid and walks as many bisection and race steps, only with fewer
fights. The probe pass at the end of `tankheal.py run` costs the same per preset it ran races for.

Every build's raw sim outputs live in `raw/<id>.tgz`, committed on master with the datasets and `provenance.json` they
made (design and audits: `orchestration/dash_raw_store.md`). `build_data.py` always rebuilds every dataset from every
build's raw outputs, so loading the packs first makes a partial run publish the same datasets as a full one, the
cross-build ones (ranks, `race_summary`, `race_notes`, `assumptions`) included.

A pack holds one build's `out/`, `names/` (sorted, deduped), `weights/` and `mana/` files, its lines of `rv/results.jsonl`
and `asm/results.jsonl` (sorted, each key once, only the keys the scripts still ask for), its `abl/ablate.json` entry with
that file's iterations and seed, and `fingerprint.json` (the sha256 of its request, its race overrides and assumption
rotations, the sim split count, the commit simmed at). Packs are deterministic, so an unchanged build's pack doesn't
change. `rawstore.py pack` refuses a build with a step missing or a failed sim, and `unpack` lists the packs simmed from
another request, race override or assumption rotation than the current one.

| Change | Commands (`$DASH_DIR` set, binaries built with `pipeline.sh bins`, run from the repo root) |
|---|---|
| A few builds changed (a class's code, presets, gear or rotation) | regenerate their requests, `pipeline.sh hydrate <ids>`, `pipeline.sh run <ids>`, `pipeline.sh publish <ids>` |
| Add a preset | write `req/<id>.json` and `.meta.json` (`verify.py`), then as above with its id |
| Remove a preset | delete `req/<id>.*` (and its duty and encounter variants), `pipeline.sh hydrate`, `pipeline.sh publish` |
| Text, tooltips, `build_data.py` or page data only | `pipeline.sh hydrate`, `pipeline.sh publish` (no sims) |
| New assumption combos | `pipeline.sh hydrate`, `python3 tools/dps_report/run_assumptions.py <ids>` (skips combos done), `pipeline.sh publish` |
| A shared sim input changed (`sim/core`, the protos, a game update): many builds stale | `pipeline.sh hydrate`, `pipeline.sh probe`, `pipeline.sh run <what it names>`, `pipeline.sh publish <the same ids>` (see "Probes") |
| Every build, no probe (packs without probes, or a probe audit failed) | `pipeline.sh run`, `pipeline.sh publish` (no hydrate) |

- `node tools/dps_report/stale.mjs` gives the scope: which builds *may* have moved. It hashes the shared inputs (the sim's
  packages and the top-level `sim/*.go` (`sim/core/report_switches.go` among them), the protos the sim compiles, the CLI
  and `go.mod`, `tools/dashstats` and the step scripts, `adaptive.py` among them), the class's `sim/<class>/`, the item
  database entries the build wears (its request's gear and its race overrides' own gear sets), and each build's
  request, the sim fields of its meta, its
  base build's request, race overrides (`req/races/`, its base build's too) and assumption rotations. Test files,
  goldens, testdata and docs are left out, and so are `tools/rotopt` and `ui/<spec>` (they make requests, which are
  hashed themselves). It lists new requests and removed presets too, and the stale Tanking and Healing presets.
  Tests: `node --test tools/dps_report/stale.test.mjs` and `python3 tools/dps_report/test_tankheal.py`.
- `hydrate [ids]` empties `$DASH_DIR`'s outputs, unpacks every pack, runs `build_data.py` and compares the sim numbers
  (`builds`, `races`, `lengths`, `hist`, `weights`, `mana`, `mana_summary`, `race_variants`, `assumption_dps`) with the
  published datasets. A build outside `[ids]` whose numbers differ stops it: re-sim it too, or find why the pack and the
  report disagree. Text, names, tooltips and icons may differ (a `build_data.py`, meta or tree change since the last
  publish); they go out with the next publish.
- `run <ids>` adds each id's duty and encounter variants (a preset whose `baseBuild` is in the list), cleans them from every
  raw output so every step re-sims them whatever its skip rule, then runs the sim steps. `resume <ids>` finishes a stopped
  run without cleaning.
- `publish [ids]` adds the ids' variants as `run` does, runs `build_data.py`, re-sims the sample fights of `[ids]` into `ui/dps_report/data/events/` (the event
  lookups always cover every build's fights), copies every dataset into `ui/dps_report/data/` (stamping `updated` in
  `datasets.json` only for changed files), drops removed presets' events, packs and provenance, records provenance for
  `[ids]` (`SIM_COMMIT=<rev>` if the sims didn't run at HEAD) and writes their packs (the others don't change).
- Commit `ui/dps_report/data`, `ui/dps_report/datasets.json`, `ui/dps_report/provenance.json` and `tools/dps_report/raw`
  together. If origin/master changed any of them meanwhile, don't hand-merge: take master's, hydrate again, re-check
  `stale.mjs`, re-sim what that adds and publish again (as with `db.json`). Deploy after the push, from the pushed commit:
  `deploy_pages.py` refuses a dirty report tree or a HEAD not on origin/master (`--unmerged` when the user asks for it).
- Shards don't hydrate: each runs `pipeline.sh run <its ids>` in an empty `$DASH_DIR` and hands back `pipeline.sh pack
  <tgz> <its ids>` on its slot (`parallel.py`, "Parallel runs"; `pipeline.sh shard` keeps each build with its
  variants). The orchestrator hydrates, `pipeline.sh unpack`s every shard (which drops those builds first, as `run`
  does) and publishes.
- **New ids to show** (new items, spells or talents): `whids.py`, then `whcache.py fortt_ids.json`, then publish.
- **Talent links**: each rank links to the Forever client's rank 1 spell on Forever's wowhead, because Forever keeps only that one (`talent_links.py`). Forever's wowhead adds spells over time, so `python3 talent_links.py ids`, then `python3 whcache.py --retry talent_ids.json` asks it again for what it didn't know. `python3 talent_links.py refresh` then rewrites the links in the shipped `talenttrees.json` without a full `build_data.py` run.
- `merge.mjs` (merge a partial `data/` folder into the report) is for hand repairs only now.

## Probes

A game update or a `sim/core` change flags most builds, though most come out the same. `pipeline.sh probe` (`probe.py
check`) finds the ones that don't, in about a tenth of a full run:

- It runs every step of the pipeline for each stale build in a scratch folder, with `probe_cli.py` standing in for the
  binaries. Every sim runs at a tenth of its iterations (one-fight runs, replays and runs of 50 fights or fewer as
  asked; the mana step logs 10 fights), and the probe keeps a fingerprint of each sim's whole result (lists sorted,
  empty fields dropped) keyed by the request the script asked for, plus a fingerprint of every output the steps write.
- The sim reseeds every fight, so unchanged code and data give the same fingerprints. `pipeline.sh run` ends by probing
  the builds it simmed (`probe.py run`), and the packs keep that probe (`probe.json`). A check compares the new probe
  with the pack's: the same means the full sims would come out the same; anything else names the step that moved.
- A change only one fight in thousands reaches can hide from a probe. So a check picks up to 3 of the builds it cleared
  for a full re-sim anyway (the audit; `PROBE_AUDIT`), and `publish` first checks that their new outputs equal their old
  packs exactly. If one doesn't, it stops: re-sim every cleared build. Otherwise it records the cleared builds as current
  (`stale.mjs --record <ids> --by probe`), and their packs and data stay as they are.
- Every probe record carries a `_key`: a hash of the probe settings, the sha256 of each binary it runs (`wowsimcli`,
  `_r`, `_assume`, `dashstats`, found behind the timing links), the sim's `db.json`, the build's `req/<id>.json`,
  `.meta.json` and `req/races/<id>.*`, `asm_apls/`, `assumptions_def.py` and the scripts the probe runs (`probe.py`,
  `probe_cli.py`, `adaptive.py`, `run_*.py`, `weights.py`, `mana_curves.py`, `ablate.py`, `rotation.py` and what they
  import, `race_override.py`, `dashpaths.py`). `probe.py run` skips a build whose `probe/<id>.json` has the key it
  would get now (so a stopped run restarts where it was), and copies `probe_check/<id>.json` when that has it (the
  check probed the same inputs, so `pipeline.sh run` after `pipeline.sh probe` doesn't probe the builds it just
  probed). Anything else is probed as before, and `PROBE_FORCE=1` probes everything. `compare` ignores the key.
- A pack without a probe (simmed before probes were kept) or from another split count, scale or architecture is re-simmed.

```bash
pipeline.sh hydrate
pipeline.sh probe                       # prints what moved and the audit; writes $DASH_DIR/probe_audit.json
pipeline.sh run <what moved> <audit>    # ends by probing them, for the packs
pipeline.sh publish <the same ids>      # checks the audit, records the cleared builds --by probe, packs the re-simmed
```

On 4 cores a probe of all 74 builds takes about 16 min (about 5,700 sims, most of them short, so process start-up is
much of it), against about 100 for a full run; `pipeline.sh run` spends the same share probing what it simmed. The probe
needs the same binaries a run uses (`$DASH_BIN`).

Checked on master's history (2026-11-03): from 251f112e1 to the hunter round's merge, `sim/testdata/report_hashes.txt`
moved 61 builds and left the 13 warlock builds alone. The probe cleared exactly those 13, and full re-sims of two of them
(and the variants `run` adds) on both binaries gave identical raw packs. On the moved builds it also names the steps:
the mage, Fury, Balance and Shadow builds moved in two `run_build` sims and the mana step only.

## Changes tab

The report's Changes tab shows the results before and after each update, with what changed (standards 30). It runs no
sims: an update's "before" is the report as the previous update left it, its "after" the report as it ended (the
current data while it is the live update). `changes.py` keeps it, in `changes/`:

- `updates.json` (hand-kept, oldest first): per update its `type` (`game`: a new client build or development notes;
  `sim`: our own changes after a game update was done), title, date, public source, `baselineData` (the commit whose
  data it starts from) and `endData` (null while live), and `lines`, the public change list. A line has a `title`, a
  `kind` (game, sim, rotation, gear, talents, consumes, race or method), an optional `note` and `source`, `moves: false` for a change that moves no number (a fix no preset reaches), and what it
  touches (`classes`, `builds`, `races`, racial `parts` as race_breakdown names them), which the page uses to show its
  effect: each racial's worth before and after for a line with `parts`, else the DPS of the presets it touches. A line
  that names nothing touches every preset. `commits` and `internal` (`{commit, why}`, commits left out of the public
  list on purpose) never ship; `check` reads them.
- `genesis.json`: the results at the first update's start, in full. `<id>.json`: a finished update's end, only the rows
  that changed. The page rebuilds every update's before and after from these, so a finished update needs no git history.

A snapshot keeps each build's DPS, error, ranks, race and best race, each build on each race, each racial part's DPS on
the generic boss (race_breakdown scenario `all`) and the race summary (baseline, all presets). The page shows a change
under 0.05 as unchanged (fights are seeded, so a build nothing touched keeps its numbers), puts ±2 standard errors of the
difference on the rest (unpaired, so it overstates the noise), marks a rank or best race change "about tied" when the
presets it passed are within that noise, and splits a race change into the old race's own change, the switch and the
sampling left over (the main run has twice a race row's fights).

| Command | What it does |
|---|---|
| `changes.py build` | Writes the datasets `changes` and `changes_snap` (`pipeline.sh publish` runs it last), and warns about lines naming unknown classes, presets, races or racials and about uncovered commits |
| `changes.py check` | Lists the commits since the live update's `baselineData` that change build inputs and no line covers (it needs that commit in the clone: `git fetch --shallow-since=<date>`) |
| `changes.py new game\|sim <id> "<title>" [--date D] [--source-label L --source-url U]` | Finishes the live update at the committed data and starts the next |
| `changes.py diff <old> [<new>]` | Prints what moved between two commits' report data (default new: the work tree), for writing lines |
| `changes.py freeze` | Rewrites `genesis.json` and the finished updates' files from git (a backfill or repair; needs their commits) |

When to start an update: before the report is republished with its first change, once the data that ends the previous
update is committed. A game update for a balance build or development notes (update-sources step 2, the dps-report full
pass); a sim/preset update for our first change of our own after a game update's full pass was deployed (dps-report
B11). Everything else joins the live update. `new` refuses while the four datasets a snapshot reads have uncommitted
changes, and warns when the finished update lists commits its data doesn't include yet. Commit what it writes. Lines are plain game terms, with only public links as sources, primary first (standards 30): Blizzard's posts, a
wago.tools page of the changed client rows or hotfix, foreverchanges.pro only when nothing else shows it. They ship.

## Parallel runs

`parallel.py` runs a full regeneration in several cloud sessions at once (the `regenerate-dps-report` skill's "Parallel
full regeneration" has the coordinator's steps). A one-session full run does DPS, Tanking and Healing one after the
other; in parallel it takes the longest session, about 25min since the 2026-11-18 engine speedups (heal-0, and heal-1
with the raid Restoration Druid), plus each session's binary build.

```sh
parallel.py plan --run <run> --commit <sha> [--dps-shards N] [--ids <id,...>] [--heal-groups N] [--heal-presets <stem,...>] [--modes dps,tank,heal] [--log <file>...] [--dry]
parallel.py worker <run> <session> --commit <sha> [--shard K/N] [--ids <id,...>] [--main] [--presets <stem,...>] [--carries-from heal-0]   # what each worker session runs
parallel.py merge <run>            # fetch, check and merge every part into $DASH_DIR, then print the publish command
parallel.py shard <k> <n> [ids]    # DPS shard k of n (pipeline.sh shard)
parallel.py slots                  # the run, session and commit each slot holds
```

- **The plan.** The DPS builds split by cost over `dps-0`.. (`pipeline.sh shard`: each build with its variants, the
  longest family first to the least loaded shard, a shard's time its share of the whole set's 80min). N is the fewest
  shards none of which outlasts the longest other session (4 for the 2026-11-18 set: about 20min each against heal-0's
  24min), or `--dps-shards N`; `--ids` shards only those builds. One session for Tanking, and the Healing presets split
  by cost: `heal-0` runs the main bench (7 bosses, tank and raid roles, 1,000 iterations, about 4min) and then its own
  presets, and every other Healing session starts its race sims from heal-0's main bench outputs. Each priest preset
  and each raid preset runs alone (priest_holy_raid is the longest, about 21min, then druid_restoration_raid); the rest
  are shared out so no session runs longer than the longest lone one. On 2026-11-18 the raid Druid and Shaman shared a
  session with the tank healers on the old model, which made it the longest at 31min. A preset's cost is its measured
  seconds in the 2026-11-18 run (`HEAL_S`), else its race rows times its role's rate from that run, or a run log's
  measured seconds with `--log` (`<mode> races <stem>: <n> rows, <s> s`).
- **A part.** A DPS shard is `pipeline.sh run <its ids>` in an empty `$DASH_DIR`, `pipeline.sh pack` of them and a
  manifest (`dps-<k>.json`: ids, commit, `DASH_SPLITS`, the binaries' stamp). `tankheal.py run <mode> --part`,
  `--main-only` or `--carries <part>` writes the mode's outputs and `part.json` instead of `run.json` and `.fresh`: the
  run settings (iterations, seed, race goals, pilot and ceiling), the commit, the bench inputs' fingerprint,
  `DASH_SPLITS`, the binaries' stamp (input hash, Go and each sha256), the presets raced and whether the main bench was
  simmed here or carried. `--carries` refuses a main bench made with any of these different.
- **The transport: slots.** The cloud sessions can't delete a branch (the git proxy answers 403 and the GitHub API
  delete is denied), so the runs share a fixed set of slot branches, `dash-raw/slot/<session>` (`plan`, `dps-0`..,
  `tank`, `heal-0`..), and leave them in place. Each push is one commit made with git plumbing (the checkout and master
  are never touched) whose parent is the slot's remote tip, so it is always a fast-forward, never a force-push. Its tree
  holds only this run's files and `slot.json` (run, session, commit simmed, files); its message names the run. A second
  push of the same run (heal-0's main bench, then its presets) keeps the first's files. A Healing worker waits for this
  run's `heal-main.tgz` on heal-0's slot by polling the remote in its background job, so it spends no turns, and waits
  past the slot's last run. One run at a time uses the slots.
- **The merge.** `parallel.py merge` reads only the planned sessions' slots and refuses a slot whose tip names another
  run (a session not done, or the slot's last run) or another commit than the plan's, a missing file, a DPS manifest
  whose ids aren't the plan's, and parts simmed with other binaries or `DASH_SPLITS`. It empties the outputs for a full
  DPS run (a partial one needs a hydrated `$DASH_DIR`), unpacks every DPS shard (`pipeline.sh unpack`, which first
  cleans the pack's builds as `run` does, so `$DASH_DIR` ends as one run of those builds leaves it) and runs
  `tankheal.py merge`. That refuses parts whose mode, commit, bench inputs (also against HEAD), splits, binaries or run
  settings differ, main bench outputs that aren't the same bytes in every part, a preset raced twice with different
  rows, rows of a preset its part didn't declare, and presets left out (unless `--keep-rest`, which keeps them from the
  pack as `run --presets` does). It writes the main outputs, `races.jsonl` sorted as one run sorts it, the probes,
  `run.json` (the run's timing keys from the parts, plus `parts`) and `.fresh`. Everything else is the same bytes as
  one run's, so the packs and datasets are too (`tests/test_parallel.py` checks this on a fake bench). The DPS shards
  were checked on a real run of four builds (2026-11-18): `pipeline.sh run` of all four in one empty `$DASH_DIR`
  against two shards each run in its own and unpacked into a third, then `publish` of each in its own worktree:
  `ui/dps_report/data` (but `events/`), `datasets.json`, `tools/dps_report/raw` and `changes` were the same bytes.
  What differs between any two publishes, sharded or not: `events/` (the Go sim's action order), `provenance.json`'s
  `recordedAt`, and in `$DASH_DIR` the line order of `rv`/`asm` results (thread-pool completion order) and the sim's
  action and names order, which `rawstore` canonicalises in the packs.

## Publishing

`pipeline.sh publish`, then `make dps-report-artifact`, the browser checks of the `regenerate-dps-report` skill, the
commit and push above, and `make dps-report-deploy`. The snapshot rule holds: the page and datasets show only final data,
with no session history, "round", "before/after", PR numbers, process words, or "shipped", except the Changes tab's
before and after (standards 30). `build_data.py` cleans notes
accordingly. No model names in any file.

## Checks harness (`rd/`)

`rd/build_harness.py` wraps the page in a mock dashboard that serves `data/*.json`. In `_str` mode it stringifies nested values
and numbers, the way the real dashboard does. With `ASYNC=1`, datasets arrive 0.06 to 1.4 s late.

The checks need Playwright on `NODE_PATH` (default `/opt/node22/lib/node_modules`) and Chromium at `PW_CHROME` (default
`/opt/pw-browsers/chromium-1194/chrome-linux/chrome`).

```bash
# page/index.html = the live files/index.html (ArtifactData get, collection files, doc index.html, out_dir)
# page/d3.min.js  = d3 v7.9.0, e.g. https://cdnjs.cloudflare.com/ajax/libs/d3/7.9.0/d3.min.js
python3 rd/build_harness.py page/index.html h            # rd/h_plain.html, rd/h_str.html
ASYNC=1 python3 rd/build_harness.py page/index.html ha   # late-arriving data
cd rd
node checklist2.js h plain > ck_h_plain.json   # feature counts per tab; diff against a baseline from the previous page
node checklist2.js h str   > ck_h_str.json
node races_check.js h plain                    # Races tab: every scenario toggle, overview, per-spec bars
node -r ./slow.js specswitch.js h ha plain     # every way to switch presets; each spec card must match a fresh load
node tips2.js h str light                      # tap one mark on every chart: tooltip + highlight (shots/tip_*)
node setup_check.js h plain light              # consumes and buffs card: faded, inactive chips, ordering, "Hide inactive"
node ttlink.js h plain light; node ttdesk.js h plain   # game tooltips and their Wowhead (Forever) links, phone and desktop
bash suite.sh h base                           # phone screenshots of every tab, light and dark (shots/)
bash spec_shots.sh h spec light plain          # spec page cards: timeline lanes and single lane, auras, rotation, gear, setup
```

Run the checks in plain, str and `ASYNC=1` (with `node -r ./slow.js`) modes, at phone width in both themes, before
publishing a page change. The older `checklist.js` and `tips.js` match earlier page versions. Use the `2` versions.
