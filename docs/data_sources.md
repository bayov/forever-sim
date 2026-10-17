# Data sources

This is the registry of where forever-sim's game data comes from. Each source has one entry:
what we take, the URLs, the fetcher, its cache, the committed input, what generates the
outputs, who reads them, how often to refresh, and how we tell it's behind.

`python3 tools/sources/stale.py` checks freshness. `python3 tools/sources/stale.py --registry`
fails when a file in `assets/db_inputs/`, or a script under `tools/` or `orchestration/` that
downloads something, isn't named in this file. `make test` runs it, so add a new input or
fetcher here in the same commit.

The `update-sources` skill (`.agents/skills/update-sources/`) is the procedure that runs the
fetchers below in order, regenerates the database and tests.

bayovui (the WoW folder's addons) keeps its own registry in `bayovui/docs/data_sources.md`.
It reads four things from us: `assets/database/db.json`, `ui/*/bis/*.bis.json`,
`assets/db_inputs/forever_wowhead_items.json` and `forever_wowhead_recipes.json`.

## Which source wins

When two sources disagree, the one higher in this list wins:

1. The Forever client and beta itself: its DB2 tables (wago.tools), its tooltips, and what we
   see in game and in combat logs.
2. Season of Discovery.
3. Anniversary, then 2019 Classic.

Never cite vmangos, and never change the sim on a tooltip alone. Wowhead and foreverchanges.pro
report what the client says, so they rank with it when they show client data. The daily routine
(`tools/forever_update/ROUTINE.md`) and the audit passes rank sources by this list.

## Provenance

- **Scraped inputs carry a `meta` block** as their first key: source, tool, fetch date (UTC),
  and Wowhead's dataEnv where it applies. `tools/sources/meta.py` writes and reads it, and
  `stale.py` reads a scrape's age from it rather than from the file's modification time.
  Today that's the four scrapes we wrote ourselves: `forever_wowhead_items.json`,
  `forever_wowhead_recipes.json`, `foreverchanges_loot.json` and
  `foreverchanges_suffixes.json`. Their first blocks (2026-10-17) took the date of the commit
  that last changed each file.
- **The 2024 Classic snapshots and the gametables carry nothing.** Their build is only in
  this file and in the fetching code. Phase 3 of the data sources plan replaces them with
  the Forever client's own tables.
- **Hand-typed data names its source in a comment**, like `// source: wago
  SpellItemEnchantment 1.60.1.70425`, as bayovui does. The talent files already do better:
  each `tools/forever_talents/client/<class>.json` has a `source` string and a `build` per
  talent.

## The game client's tables

### wago.tools DB2, the Forever client (1.60.1.x)

- **What:** every table of each Forever beta build: spells, talents, items, enchants and the
  rest. We use them for the client diffs, the audit loop's checks, talent trees, racials,
  logcheck's spell and enchant names, and the regen gametables.
- **URLs:** `https://wago.tools/db2/<Table>/csv?build=<version>` and `https://wago.tools/api/builds`.
- **Fetcher:** `tools/sources/wago.py`, for every reader in this repo. `latest [--probe N]` for
  the newest build, `fetch <build> --set <reader>` for a reader's tables (the sets are in its
  `SETS`), `diff <old> <new>` to compare two builds, and `tools/sources/db2diff.py` for the
  report grouped by class. logcheck has the same fetcher in Go, `tools/logcheck/wago.go`.
- **Cache (not committed):** `~/.cache/wow-data/wago/<version>/<Table>.csv`, shared with
  bayovui's build scripts and logcheck. `WOW_DATA_CACHE` moves it, and `tools/audit/setup.sh`
  points it into `$AUDIT_WORK` for a cloud pass.
- **Committed results:** the diffs `orchestration/audit/ref/client_update/diffs/<a>-<b>.md`
  (from `tools/sources/db2diff.py`), `orchestration/audit/ref/consumables/builddiff_*.txt`, the talent
  files in `tools/forever_talents/client/`, and `assets/db_inputs/basestats/octregenhp.txt`
  and `regenhpperspt.txt` (below).
- **Readers:** `orchestration/update_ledger.md`, the audit findings, hand edits to the sim,
  `ui/core/talents/trees/*.json` through `tools/forever_talents/import_client.py` and
  `import_talents.py`.
- **Refresh:** each new build. The audit loop runs every 6 hours and its `wago_build`
  external wakes the auditors that read these tables. The daily routine logs "Pull
  wago.tools DB2 data for build N" in `TASKS.md`. Then the `dps-report` skill's full pass.
- **Behind when:** `stale.py` compares the newest build we know (the installed beta, and
  `TASKS.md`) with `foreverClientBuild` in `.agents/skills/dps-report/state.json`.
  `tools/audit/gate.py` compares `wago.py latest --probe 300` with each auditor's state.
- **Note:** the passes say not to reach wago.tools any other way when it refuses.

### wago.tools ItemSparse, Classic 1.15.3.55646

- **What:** item set IDs and factions (`gen_db`), required levels (rotopt), a cross-check.
  It's Classic's file, so it has none of Forever's own items.
- **URL:** `https://wago.tools/db2/ItemSparse/csv?build=1.15.3.55646`
- **Fetcher:** `go run ./tools/database/gen_db -outDir=assets -gen=wago-db2-items`, through
  `ReadWeb` in `tools/io_utils.go`. No cache.
- **Input:** `assets/db_inputs/wago_db2_items.csv` (2024).
- **Generator:** `go run ./tools/database/gen_db -outDir=assets -gen=db`.
- **Readers:** `assets/database/db.*`, `tools/rotopt` (`-item-levels`),
  `tools/dps_report/wagocheck.py`, `orchestration/tools/b30_era.py`.
- **Refresh:** never. It's pinned, and phase 3 replaces it with the Forever client's ItemSparse.

### Client gametables and cmangos

- **What:** crit per agility and intellect, base mana, the regen ratios, player base stats
  and HP by level, and open-world normal mobs.
- **Inputs:**
  - `assets/db_inputs/basestats/chancetomeleecrit.txt`,
    `assets/db_inputs/basestats/chancetospellcrit.txt`,
    `assets/db_inputs/basestats/octbasempbyclass.txt`: client gametables from upstream (2023).
    No fetcher.
  - `assets/db_inputs/basestats/octregenhp.txt`, `assets/db_inputs/basestats/regenhpperspt.txt`:
    the gametables of
    wago.tools' Titan client 3.80.2.70315, downloaded by hand (2026-10-15).
  - `assets/db_inputs/basestats/player_classlevelstats.csv`,
    `assets/db_inputs/basestats/player_levelstats.csv`:
    exported by hand from the cmangos classic-db dump (`github.com/cmangos/classic-db`,
    `Full_DB/ClassicDB_1_12_1_z2815.sql.gz`).
  - `assets/db_inputs/mobs/normal_mobs.csv`: `python3 tools/export_normal_mobs.py
    <ClassicDB_1_12_1_z2815.sql.gz> <FactionTemplate.csv>`, the FactionTemplate of wago
    build 1.15.7.61582.
- **Generators:** `tools/gen_level_base_stats.py` → `sim/core/base_stats_levels_auto_gen.go`,
  `tools/gen_regen_tables.py` → `sim/core/regen_tables_auto_gen.go`,
  `tools/gen_mob_stats.py` → `sim/core/mob_stats_auto_gen.go`. `sim/core/base_stats_auto_gen.go`
  has no generator since `base_stats_parser.py` went (phase 0), so edit it by hand.
- **Refresh:** by hand, when the game changes base stats. `logcheck -mobs` checks the mob
  health model against our logs.

## Wowhead

### www.wowhead.com Forever pages

- **What:** `scrape.py` takes every wearable item up to `maxReqLevel` (its stats and icon)
  and every quest's class and side restriction. `recipes.py` takes the recipe items of each
  profession: vendors (stock, cost, map pins), drops with chances, objects, containers,
  fishing, pickpocket and quests.
- **URLs:** `https://www.wowhead.com/forever/items/slot:<s>/min-req-level:<lo>/max-req-level:<hi>`,
  `/forever/quests/...`, `/forever/items/recipes/<prof>`, `/forever/item=<id>`, `/forever/npc=<id>`.
- **Fetchers:** `uv run tools/database/forever_wowhead/scrape.py [max required level]` and
  `uv run tools/database/forever_wowhead/recipes.py`, through
  `tools/database/forever_wowhead/wowhead.py`, which earns the bot check's token in headless
  Chrome and fetches with curl_cffi. `scrape.py` also asks the tooltips below whether a
  crafted item binds on pickup. The `scrape-wowhead` skill has the procedure and the rate limit.
- **Cache:** `~/.cache/wow-data/wowhead/pages` (gzipped pages, `WOWHEAD_CACHE_DAYS`, default 7).
- **Inputs:** `assets/db_inputs/forever_wowhead_items.json`,
  `assets/db_inputs/forever_wowhead_recipes.json`, both with a `meta` block (dataEnv 16).
- **Generator:** `gen_db -gen=db` (`tools/database/forever_wowhead.go`) for the items.
  Nothing here reads the recipes: bayovui's `recipesources.py` does.
- **Readers:** `assets/database/db.*`, `tools/rotopt` (`-forever-items`),
  `tools/database/foreverchanges/suffixes.py` (required levels), `orchestration/tools/`
  (`b30_data.py`, `bis_data.py`, `fit.py`, `primary.py`, `single.py`), bayovui's `bis.py`
  and `recipesources.py`.
- **Refresh:** weekly, locally (never in the cloud routine). `stale.py` says when a scrape
  is older than 7 days.
- **Note:** plain curl gets a 403 from www.wowhead.com. `tools/audit/consumables/src.py` and
  `tools/audit/consumables/srcid.py` still try it, and phase 2 moves them onto `wowhead.py`.

### nether.wowhead.com Forever tooltips (dataEnv 16)

- **What:** each item's and spell's tooltip in Forever's database (name, quality, icon,
  tooltip HTML).
- **URL:** `https://nether.wowhead.com/forever/tooltip/{item|spell}/{id}?dataEnv=16&locale=0`,
  and Classic's `/classic/...?dataEnv=4` as the fallback.
- **Fetchers:** `python3 tools/dps_report/whcache.py [--retry] <ids.json>` (the id lists
  `tools/dps_report/fortt_ids.json` and `talent_ids.json` come from `whids.py` and
  `talent_links.py ids`), `forever_wowhead/scrape.py` (bind on pickup, no cache), and the
  daily routine's `tools/forever_update/build.py`.
- **Cache:** `tools/dps_report/forever_tt_cache/{forever,classic}/{item,spell}/<id>.json`,
  committed (`DASH_TT_CACHE` overrides it). The routine uses `~/.cache/forever_update`.
- **Readers:** the DPS report (`tools/dps_report/build_data.py`, `tooltips.py`). The browser
  UI also asks nether.wowhead.com live (`ui/core/constants/wowhead.ts`, dataEnv 16).
- **Refresh:** when the report is rebuilt. `whcache.py --retry` asks again for the ids
  Forever's Wowhead didn't know yet. The cache has no fetch dates.
- **Hand-typed from it:** `tools/database/forever_items.go` (items we enter by hand).

### nether.wowhead.com Classic snapshots (2024)

- **What:** Classic item and spell tooltips, our main source of Classic item stats and
  effects, and the Classic gear planner dump (items, item sets, random suffixes).
- **URLs:** `https://nether.wowhead.com/classic/tooltip/item/<id>?lvl=60`,
  `/classic/tooltip/spell/<id>`, `/classic/data/gear-planner?dv=100`.
- **Fetchers:** `go run ./tools/database/gen_db -outDir=assets -gen=wowhead-items [-id= |
  -minid= -maxid=]`, `-gen=wowhead-spells`, `-gen=wowhead-gearplannerdb`, through
  `tools/io_utils.go`. The tooltip fetchers only add missing ids. `tools/audit/gear/tt.py`
  asks for single Classic tooltips.
- **Inputs:** `assets/db_inputs/wowhead_item_tooltips.csv`, `wowhead_spell_tooltips.csv`,
  `wowhead_gearplannerdb.txt` (November 2024, the planner's items say up to 1.15.5).
- **Generator:** `gen_db -gen=db`.
- **Readers:** `assets/database/db.*`, `tools/database/foreverchanges/suffixes.py` (the
  planner), the DPS report, `orchestration/tools/b30_data.py` and `bis_data.py`.
- **Refresh:** never. Fetching again would mix in today's Classic data. Phase 3 replaces them.

## foreverchanges.pro

- **What:** `scrape.py` takes each dungeon's boss loot (item, required level, new or changed,
  tooltip lines). `suffixes.py` takes every random suffix an item can roll, with its stats,
  for the items that roll suffixes. Hidden item reveals and patch notes are read by the daily
  routine and the client_update audit pass.
- **URLs:** `https://foreverchanges.pro/dungeons/<slug>`, `https://foreverchanges.pro/item/<id>`,
  `/hidden-items`, `/patch-notes`.
- **Fetchers:** `python3 tools/database/foreverchanges/scrape.py`,
  `python3 tools/database/foreverchanges/suffixes.py <max level> <cache dir>`. Both fetch
  pages and read their Next.js payload through `tools/sources/foreverchanges.py`, and so does
  the daily routine for the hidden items page.
- **Cache:** suffixes keeps `<cache dir>/<id>.html` (404s too). scrape.py has none.
- **Inputs:** `assets/db_inputs/foreverchanges_loot.json`,
  `assets/db_inputs/foreverchanges_suffixes.json`, both with a `meta` block.
- **Generator:** `gen_db -gen=db` (`tools/database/foreverchanges.go`).
- **Readers:** `assets/database/db.*`. Suffix IDs from 990000 stay the same from run to run,
  because `ui/*/bis/*.bis.json` and gear saved in the sim store them.
- **Refresh:** after reveals. `stale.py` says when one is older than 14 days.

## AtlasLoot

- **What:** item sources (boss drops, crafts, PvP and faction rewards), zones, NPCs and
  factions, from the AtlasLoot Forever addon (CurseForge 1422985) installed in the beta
  client. Its loot tables already list Forever's new dungeons. Its Forever bosses have
  names but no NPC IDs yet, so their drops carry the boss's name only.
- **Read from:** the client's `Interface/AddOns`: `AtlasLootClassic_Data/source.lua`,
  `AtlasLootClassic_DungeonsAndRaids/data.lua`, `AtlasLootClassic_PvP/data.lua` and
  `AtlasLootClassic_Factions/data.lua`. Zone names come from
  `nether.wowhead.com/classic/tooltip/zone/<id>`, and faction names from
  `assets/db_inputs/factions.json`, a hand list from upstream.
- **Fetcher:** `go run ./tools/database/gen_db -outDir=assets -gen=atlasloot -atlasloot=<the
  beta's Interface/AddOns>` (`tools/database/atlasloot.go`). It only runs locally, since the
  addon is on our machine. The daily routine and the audit loop read the committed input.
- **Input:** `assets/db_inputs/atlasloot_db.json`, and `assets/db_inputs/atlasloot_db.meta.json`
  next to it with the addon's version and a hash of the four files. The input is read as a
  proto, which has no room for a meta block.
- **Generator:** `gen_db -gen=db`.
- **Cross-check:** `python3 -I tools/sources/crosscheck.py` lists the items where AtlasLoot and
  Wowhead's gear listing disagree on a drop, or only one of them knows it.
- **Refresh:** when CurseForge updates the addon. The `update-sources` skill's AtlasLoot step.
- **Behind when:** `stale.py --addons <AddOns>` hashes the installed files and compares the hash
  with the meta file. bayovui's `stale.py` passes the beta's folder.
- **History:** until 2026-10-17 we read a 2024 Season of Discovery fork from GitHub
  (HiKwonko/AtlasLootClassic_SoD). Switching dropped 1165 SoD-only items and added 116 of
  Forever's, and 914 items gained drop sources.

## Talents, racials and enchants (hand-vendored)

- **hyjal.cc:** the talent calculator's bundle for 7 classes
  (`tools/forever_talents/client/*.json`, build 1.60.1.69876, brought up to later builds from
  wago.tools talent by talent), racial text (`ui/core/forever/racials.json`), and enchant and
  armor kit values (`tools/database/forever_enchants.go`). Fetched by hand, no fetcher.
- **wago.tools Trait tables:** the druid and warrior `client/*.json` (70338).
- **BlizzCon transcription (Deradon/wow-forever-talent-calc, MIT, see
  `tools/forever_talents/LICENSE.upstream`):** the hand corrections in
  `tools/forever_talents/overrides/*.json`. Every class's `data/*.json` now says
  `dataSource: client`.
- **Generator:** `tools/forever_talents/import_client.py <class>`, then
  `import_talents.py <class> --write` → `ui/core/talents/trees/*.json` and the protos.
- **Behind when:** `stale.py` shows the newest build each class's file names. Phase 3 replaces
  all of this with an importer from the client's Trait tables.

## wowforevertalents.com

- **What:** `forever-wiki/data/abilities.json` and `talents.json` (2026-09-14, pre-beta), and
  the quest names, levels and factions typed into `tools/database/forever_items.go`.
- **Fetcher:** none in the repo. `tools/dps_report/build_data.py` and `tooltips.py` read
  `abilities.json`. Nothing reads `talents.json`. Phase 3 removes both.

## Icons: wow.zamimg.com

- **URLs:** `https://wow.zamimg.com/images/wow/icons/{medium,large}/<icon>.jpg`, and the talent
  tree backgrounds named in `ui/core/talents/trees/*.json`.
- **Fetchers:** `tools/dps_report/build_data.py` (fetches any icon it lacks),
  `tools/dps_report/events.py`, the routine's `build.py`.
- **Cache:** `tools/dps_report/icons/` (not tracked). The report embeds them.

## Warcraft Logs API

- **What:** public reports, for proc and glancing statistics. Nothing from it goes into the
  sim yet.
- **URLs:** `https://www.warcraftlogs.com/oauth/token`, `https://<site>.warcraftlogs.com/api/v2/client`.
- **Fetcher:** `python3 tools/warcraftlogs/wclcli.py --site <s> ...` through
  `tools/warcraftlogs/wcl.py`. Credentials in `~/.config/forever-sim/wcl.json`.
- **Cache:** `~/.cache/forever-sim/wcl/`. 3600 API points an hour.

## Our own game client

- **What:** combat logs, BayovCore's character history from the saved variables, `/bayov
  export` text.
- **Fetchers:** `tools/combat_log/archive.sh` (rsync into `~/forever-logs/`),
  `python3 tools/combat_log/history.py`, logcheck's `-archive`.
- **Readers:** `tools/logcheck` (`-audit`, `-mobs`), `tools/combat_log/*.py`, bayovui's
  `statweights.py` (the history).

## News and official notes

- **The daily routine** (`tools/forever_update/ROUTINE.md`, whose instructions its trigger
  prompt points at) reads Blizzard's news and forums, Wowhead news, other news
  sites, foreverchanges.pro's reveals, Forever tooltips and wago.tools builds each day. It
  opens a `forever-update/<date>` PR and logs what it couldn't finish in `TASKS.md`, each item
  tagged with its source. `stale.py` lists its open PRs and the builds it logged.
- **The audit loop** reads Blizzard's dev notes thread
  (`https://us.forums.blizzard.com/en/wow/t/2360696.json?print=true`, parsed by
  `tools/audit/client_update/devnotes.py`) and foreverchanges.pro's patch notes. Snapshots go
  to `orchestration/audit/ref/client_update/sources/`.
- `forever-wiki/references.md` lists the reference sites we looked at.

## gen_db's modes

`go run ./tools/database/gen_db -outDir=assets -gen=<mode>`:

- `db` (also `make items`): reads every input above and writes `assets/database/db.{bin,json}`
  and `leftover_db.{bin,json}`.
- `atlasloot`, `wowhead-items`, `wowhead-spells`, `wowhead-gearplannerdb`, `wago-db2-items`:
  the fetchers above.

Overrides on top of the inputs live in `tools/database/overrides.go` (allow and deny lists),
`tools/database/enchant_overrides.go`, `forever_enchants.go`, `forever_items.go` and
`forever_synthetic_items*.go`. Never hand-merge `assets/database/db.*`: regenerate it.
