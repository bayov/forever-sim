# DPS report: dashboard data pipeline

These scripts produce every dataset behind the Forever level 60 DPS Report: the page in `ui/dps_report/`, its datasets in
`ui/dps_report/data/`, deployed to https://forever-dps.pages.dev (`make dps-report-deploy`).

```
merged tree ──► requests ──► sims ──────────► raw packs ◄──► datasets ──► report
 (w-dash)        req/        out/ rv/ asm/     raw/<id>.tgz    ui/dps_report/data/   forever-dps.pages.dev
                             abl/ weights/ mana/
```

Each build's raw sim outputs are kept in `raw/<id>.tgz`, committed with the datasets they made, so a regeneration re-sims
only the builds that changed and still rebuilds every dataset, the cross-build ones included (see "Partial regenerations").

1. **Merged tree.** One local checkout with every class branch and the consumes branch merged in, the item DB regenerated, and
   the local dashboard patches (`patches/`) applied. The binaries are built from it.
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
| `patches/` | Local-only sim patches for the merged tree (see below) | yes |
| `raw/<id>.tgz` | Each build's raw sim outputs (about 2 MB for all), written by `rawstore.py pack` (see "Partial regenerations") | yes |
| `rawstore.py` | Packs, unpacks, cleans and checks the raw outputs, and copies the datasets into the report | yes |
| `dashstats/main.go` | The `dashstats` helper. Copy it to `tools/dashstats/` in the merged tree, or let patch 0001 add it. | yes |
| `rd/` | Headless check harness (Playwright) | scripts only |
| `bin/` | `wowsimcli`, `dashstats` and optionally `wowsimcli_r`, built from the merged tree | no |
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

## Building the merged tree

```bash
# once: a worktree on a local-only branch (never push it)
git worktree add ../w-dash -b dashdata <the class branches' common base, or the previous dashdata>
cd ../w-dash
git fetch origin
for c in warrior paladin hunter rogue priest shaman mage warlock druid consumes; do
  git merge --no-edit origin/claude/wow-forever-dps-presets-m9u15h-$c || { DASH_TREE=$PWD bash ../forever-sim/tools/dps_report/resolve.sh; git status; }
done
```

Conflict rules:

- `assets/database/db.{json,bin}`: take ours and regenerate after all the merges. `resolve.sh` does this.
- `*.md`: union of both sides. `resolve.sh` does this too.
- Code: take the class's side for that class's files. For `tools/rotopt/run.go`, the hunter side won last time.
- If `tools/rotopt` ends up with two `func manaAbove`, delete the one in `priest_shadow.go`. Patch 0001 does this.

Then regenerate the DB, apply the patches and build:

```bash
go run tools/database/gen_db/*.go -outDir=./assets -gen=db      # every class's synthetic items and trinkets, the spell renames
git commit -am "dash: regenerate db"                               # local only
git am -3 ../forever-sim/tools/dps_report/patches/000[1-3]-*.patch  # `git am --skip` a patch the merged heads already contain
git apply ../forever-sim/tools/dps_report/patches/0004-*.patch    # a plain diff, not a mail
git apply ../forever-sim/tools/dps_report/patches/0005-*.patch    # DASH_SPLITS, a plain diff too
protoc -I=./proto --go_out=./sim/core ./proto/*.proto              # only if sim/core/proto/*.pb.go are missing (protoc-gen-go in ~/go/bin on PATH)
B=../forever-sim/tools/dps_report/bin; mkdir -p $B
go build -tags=with_db -o $B/wowsimcli ./cmd/wowsimcli
go build -tags=with_db -o $B/mksettings ./tools/rotopt/mksettings
go build -tags=with_db -o $B/rotopt ./tools/rotopt
go build -tags=with_db -o $B/dashstats ./tools/dashstats
cp $B/wowsimcli $B/wowsimcli_r   # optional; run_races.py uses wowsimcli when wowsimcli_r is missing
```

After regenerating the DB, check that every class's synthetic item ids (990000 and up, from
`tools/database/forever_synthetic_items_<class>.go`) are in `db.json`.

### Local patches (`patches/`)

These were exported with `git format-patch` from the local `dashdata` branch, against the merged class heads. They are
dashboard-only and never go to a class branch.

| Patch | What it does | Needed by |
|---|---|---|
| `0001` DASH_NAMES dump, dashstats, manaAbove dedupe | In `sim/core/environment.go` (finalize): when `DASH_NAMES=<file>` is set, appends every unit's aura `spellId itemId otherId tag label` to that file. Also adds `tools/dashstats/main.go`, and drops the duplicate `manaAbove` from `tools/rotopt/priest_shadow.go`. | `run_build.py` (`names/<id>.tsv`, used for aura and proc labels in `names`, `abilities`, `auras` and `timeline_*`). `stats` needs dashstats. rotopt needs the dedupe to build. |
| `0002` DASH_RACIALS_OFF | In `sim/core/racials_forever.go` and the priest racials (Chastise, Dark Sacrifice, Starshards): `DASH_RACIALS_OFF=all` or `=<part>[,<part>]` leaves those racial parts out. Parts: spirit, ruthless, earthfury, savage, beast, dodge, elune, shadowmeld, mind, eureka, bloodfury, touch, endurance, berserking, windblessed, insight, leyline, chastise, starshards, darksacrifice. | `run_races.py` (`off`, `off:<part>` and `weaponNoRacial` variants) |
| `0003` target racials damage only | Beast Slaying, Big Game Hunter and Elemental Insight no longer also raise CritMultiplier. Client 70338 gives them aura 168 at 5%. | nothing in the report: master has the Forever half since PR #94, and the patch keeps only the Classic `racials.go` line. |
| `0004` DASH_ASSUME knobs | `sim/core/assumptions.go`: `DASH_ASSUME="key=value,..."` overrides the constants the assumption scenarios vary (`assumptions_def.py` lists the keys). A plain diff against master with 0001 to 0003 applied; apply with `git apply`. A knob on a constant master already sets (the caster rank scales, Ice Lance's coefficient, Rend's attack power share, the hunter's crit per Agility, Raptor Strike rank 8, the pet hit rule, the top Trueshot Aura and Hunter's Mark ranks) defaults to that constant, so a run with no knob reproduces master. Rebased 2026-10-23 onto the wowsims/forever alignment round: the knobs of the assumptions it settled are gone (orchestration/alignment_round_2026-10-23.md). Build it as `$DASH_BIN/wowsimcli_assume`. | `run_assumptions.py` |
| `0005` DASH_SPLITS | `sim/core/sim_concurrent.go`: `DASH_SPLITS=<n>` sets how many parts a sim's iterations are split into (default one per CPU). `pipeline.sh` sets 4, so the raw outputs are the same on every machine: the splits' seeds already continue each other, but averaging the parts moves the last bits. A plain diff; apply with `git apply`. | `pipeline.sh` (every step), `rawstore.py` (refuses packs made with another count) |

The db regeneration commit isn't exported because it's binary; rerun gen_db instead. Merge commits aren't exported either:
their conflict resolutions follow the rules above. With no env var set, the patches leave the Forever sim's behaviour
unchanged. `run_assumptions.py --identical`
checks that a 0004 build is bit-identical to the plain one with `DASH_ASSUME` unset.

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
Each assumption in `assumptions_def.py` is written for a player who hasn't read our notes: a question-style `label`
(8 words at most), a `short` line (20 words at most) the Ranking panel shows under it, setting labels that say what each
choice means with a `rationale` of 15 words at most, and a `desc` of 2 to 4 sentences plus its `src` (written to the
data as `source`) for the Method tab.

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
into its All races view. `stale.mjs` and the packs' fingerprints
hash a build's override files (its base build's too), so adding, changing or deleting one makes the build stale.

`race_gear.py [ids] [--write]` makes the Dwarf's gear files: for every build a Dwarf can be, the Dwarf's hit (dashstats'
final sheet, plus a priest's school talent) against its standards 4 cap (16% spell for casters, 8% yellow for melee and
hunters), the excess (at most 1%, at most the gear's own hit) re-spent over the set's other non-Stamina, non-Spirit stats
in its own shares at the 1b prices, then a paired 20k sim with and without. `--write` writes the files that gain 2 SE;
every build's numbers go to `$DASH_DIR/race_gear.json`.

## Sim steps

Each step is restartable: it skips work whose output already exists. To redo a build, delete its output first.

| Step | Command | Sims | Writes | Runtime, 4 CPUs |
|---|---|---|---|---|
| main + races + lengths + samples | `runner.sh` (start up to 4) or `python3 run_build.py <id>...`; `--resample <id>` redoes only the sample fights | Per build: main 180 s, 20k iterations, seed 1, all values saved. Race sweep: every race the class can be, 10k, each with its race override if any. Lengths 60/120/180/300/480 s, 5k. 30 one-iteration rolls, then 3 logged sample fights at the 10th, 50th and 90th percentile. dashstats stat sheet. | `out/<id>.json`, `names/<id>.tsv` | about 22 min for all 36 (2 to 5 min per class) |
| stat weights | `ls req/*.meta.json \| xargs -n1 basename \| sed 's/.meta.json//' \| xargs -P4 -n1 python3 weights.py` | Per build: the baseline plus 6 to 8 bonus-stat runs (hit also measured downward), 180 s, 40k iterations, seed 1 | `weights/<id>.json` | about 30 min |
| race scenarios | `python3 run_races.py [ids]` (`PAR=4` threads) | Per default build × race (encounter presets get only the race sweep): `off`, `off:<part>`, `beast`, `elemental`, `humanoid` (Priests only: Dwarf Chastise), and `weapon` / `weaponNoRacial` where the racial weapon is legal and neither hand holds it. The swap changes one weapon (the one whose speed changes least, the off hand on a tie) to the racial's type, at the nearest Phase 1 BiS speed of that type. Each race with its race override if any. 10k iterations, seed 1. The last full run was 1,084 sims. | appends to `rv/results.jsonl` (keyed build, race, variant) | about 35 min for all; the 458 beast and elemental reruns took about 16 min |
| mana curves | `python3 mana_curves.py [ids]` (`FIGHTS=100 PAR=8`, `--force` redoes builds that already have an output) | Per mana-using build (every class but warrior and rogue): 100 logged one-fight runs at a fixed 480 s with no duration variation, seeds 1001 and up, the player's mana read from the log every 5 s. Plus one 5k run for the sim's seconds out of mana, time to OOM and mana left at the end. | `mana/<id>.json` | about 17 s for the 6 shaman builds |
| setup ablation | `python3 ablate.py [ids]` (`ITERS=300 SEED=11 PAR=4`; `ABL_MERGE=1` keeps the other builds' entries) | Each preset with one consume, buff, debuff or profession left out, plus the non-stacking pairs. 1,957 sims. | `abl/ablate.json` | about 5 min |
| assumption scenarios | `python3 run_assumptions.py [ids]` (`PAR=4`) | Per build an assumption touches (`assumptions_def.py`): the base, every non-default setting alone at 180 s (20k) and 120 / 300 s (5k), every combination at 180 s where a build has 3 or fewer assumptions, and 12 random combinations otherwise (the multiplicative estimate's check). About 1,100 sims. | appends to `asm/results.jsonl`; `build_data.py` (via `asm_data.py`) writes `assumptions` and `assumption_dps` | about 1.5 h for all |
| tooltip ids | `python3 whids.py` (after a first `build_data.py`) | none | `fortt_ids.json` | seconds |
| tooltip fetch | `python3 whcache.py fortt_ids.json` | none. Fetches from `nether.wowhead.com/{forever,classic}/tooltip/<kind>/<id>`, only ids missing from the cache. | `forever_tt_cache/` | about 13 min cold (about 1,400 requests at 0.35 s); seconds when warm |
| datasets | `python3 build_data.py` | none | `data/*.json`, `fortt_coverage.json` | about 10 s warm; about 1 min more when icons must be fetched |

Diagnostics, which produce no datasets: `whdump.py` prints Forever and Classic tooltip text side by side for the setup chips,
and `wagocheck.py` compares the real items' ilvl and stats across Forever wowhead, the Forever client and the sim's db.

## Datasets

All are written by `build_data.py`. "out" means `out/<id>.json` from `run_build.py`.

| Dataset | Made by | Depends on | Runtime |
|---|---|---|---|
| builds | build_data.py | out (main 20k, race sweep 10k, 120/300 s lengths), req meta, ranks across all builds | run_build |
| races | build_data.py | out race sweep: every legal race, 180 s, 10k | run_build |
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
| rotation | build_data.py + rotation.py | the request's APL (prepull and priority), cast counts from abilities, names, `racials.json` | none |
| race_rotation | build_data.py + rotation.py (`race_build`) | each race's own APL in `req/races/` (`race_override.load`), lined up with the preset's: `forRace`, `diff` (same, changed, added, removed) and `base` (the preset line's order). No casts | none |
| race_variants | build_races.py | `rv/results.jsonl` plus the out race sweep (baseline); `tuned`, the note of the race's override | run_races |
| race_breakdown | build_races.py | `rv/results.jsonl` (`off`, `off:<part>`, weapon and target deltas) | run_races |
| race_summary | build_races.py | race_variants by scenario, scope (overall or faction) and role | run_races |
| race_notes | build_races.py | scenario and part texts, a finding per class from race_variants | run_races |
| weights | build_data.py | `weights/*.json` (40k per stat) | weights.py |
| mana | build_data.py (`mana_curves.datasets()`) | `mana/*.json`: per build and 5 s step, the average mana % of max over 100 fights at 480 s, with the 25th and 75th percentile fight | mana_curves.py |
| mana_summary | build_data.py (`mana_curves.datasets()`) | `mana/*.json`: when mana first drops below 15% (`lowPct`) (median, p25, p75, null when that fight quantile never does), share of fights that do, mana % at 60/180/300/480 s, spend and regen per second, and the 5k run's `secondsOomAvg`, Tto and mana at the end | mana_curves.py |
| tooltips | build_data.py + tooltips.py | `forever_tt_cache` (Forever wowhead, Classic fallback), the tree's wowhead csvs, `db.json` (synthetic items, enchants), `sim/*/item*sets_pve.go`, `sim/*/forever_synthetic_trinkets.go`, Forever wiki, `racials.json`, talent trees | whcache.py |
| talenttrees | build_data.py | `ui/core/talents/trees/<class>.json`, plus wowhead links from the tooltip cache | none |
| treeart | build_data.py | the tree icons from `tools/forever_talents/client/<class>.json`, and backgrounds fetched from each tree's `backgroundUrl` | icon fetch |
| icons | build_data.py | every icon used, from wow.zamimg.com (36 px medium, 56 px large for headers), plus `assets/item_slots/*.jpg` | icon fetch |

Side outputs: `fortt_coverage.json`, the count of tooltips by text source, for reporting. The page also expects every dataset
under 8 MB; the largest are icons and treeart at about 1 MB.

## Partial regenerations

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

| Change | Commands (`$DASH_DIR` set, binaries built from the tree, run from the repo root) |
|---|---|
| A few builds changed (a class's code, presets, gear or rotation) | regenerate their requests, `pipeline.sh hydrate <ids>`, `pipeline.sh run <ids>`, `pipeline.sh publish <ids>` |
| Add a preset | write `req/<id>.json` and `.meta.json` (`verify.py`), then as above with its id |
| Remove a preset | delete `req/<id>.*` (and its duty and encounter variants), `pipeline.sh hydrate`, `pipeline.sh publish` |
| Text, tooltips, `build_data.py` or page data only | `pipeline.sh hydrate`, `pipeline.sh publish` (no sims) |
| New assumption combos | `pipeline.sh hydrate`, `python3 tools/dps_report/run_assumptions.py <ids>` (skips combos done), `pipeline.sh publish` |
| A shared sim input changed (`sim/core`, the DB, the patches): every build stale | `pipeline.sh run`, `pipeline.sh publish` (no hydrate) |

- `node tools/dps_report/stale.mjs` gives the scope. It hashes the shared inputs, the class's `sim/` and `ui/` folders, and
  each build's request, race overrides (`req/races/`, its base build's too) and assumption rotations, and lists new
  requests and removed presets.
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
  shard<k>.tgz <its ids>`. The orchestrator hydrates, `pipeline.sh unpack`s every shard and publishes.
- **New ids to show** (new items, spells or talents): `whids.py`, then `whcache.py fortt_ids.json`, then publish.
- **Talent links**: each rank links to the Forever client's rank 1 spell on Forever's wowhead, because Forever keeps only that one (`talent_links.py`). Forever's wowhead adds spells over time, so `python3 talent_links.py ids`, then `python3 whcache.py --retry talent_ids.json` asks it again for what it didn't know. `python3 talent_links.py refresh` then rewrites the links in the shipped `talenttrees.json` without a full `build_data.py` run.
- `merge.mjs` (merge a partial `data/` folder into the report) is for hand repairs only now.

## Publishing

`pipeline.sh publish`, then `make dps-report-artifact`, the browser checks of the `regenerate-dps-report` skill, the
commit and push above, and `make dps-report-deploy`. The snapshot rule holds: the page and datasets show only final data,
with no session history, "round", "before/after", PR numbers, process words, or "shipped". `build_data.py` cleans notes
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
