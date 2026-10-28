---
name: update-sources
description: Pull forever-sim's game data inputs from their sources and regenerate the item database. A new client build's wago.tools tables and their diff, the talent trees, racials, abilities, enchant values and item sets read from them, the foreverchanges.pro scrapes, the Wowhead scrapes (through scrape-wowhead), the installed AtlasLoot Forever and its cross-check against Wowhead, then gen_db, the tests and a commit. Use when tools/sources/stale.py (or bayovui's) says an input is behind, when a new Forever client build is out, when AtlasLoot updated on CurseForge, when the dps-report skill's full pass starts, when bayovui's /update-data runs its forever-sim step, or when the user asks to refresh, pull or update the sim's sources, or types /update-sources.
---

# Update forever-sim's sources

This is the one procedure for pulling game data into forever-sim. `docs/data_sources.md`
is the registry: every source, its URL, its fetcher, its cache, the input it writes and
who reads it. This skill says in which order to run them. When a step here and the
registry disagree, the registry is right, so fix this file.

It runs locally. The daily cloud routine (`tools/forever_update/ROUTINE.md`) only detects
new builds and reveals and prepares PRs, and AtlasLoot is only on our machine.

## 1. What's behind

```bash
python3 tools/sources/stale.py --client-build <beta version> --addons "<beta client>/Interface/AddOns"
```

The beta's version is in the WoW folder's `.build.info`. bayovui's `build/stale.py` runs
this for you with both filled in. Do only the steps it names, plus whatever the user asked
for. Open PRs from the daily routine come first: the user decides each merge.

## 2. A new client build

```bash
python3 tools/sources/wago.py latest --probe 300
python3 tools/sources/wago.py fetch <old> --set client_update   # usually cached already
python3 tools/sources/wago.py fetch <new> --set client_update
python3 -I tools/audit/client_update/table_summary.py <old> <new>
python3 -I tools/sources/db2diff.py <old> <new> /tmp/<old>-<new>.md
```

`<old>` is the build the sim was audited against (`foreverClientBuild` in
`.agents/skills/dps-report/state.json`). The table summary says which tables changed at
all. A build where only rows were added (pets, mounts, ensembles) is cosmetic. Otherwise
read the diff, which groups every change by class.

Processing a real balance build is the dps-report skill's full pass, which is hours of
agents, so tell the user what changed and ask before starting it. When it goes ahead, start the report's game update
first, so its changes land in it: `python3 tools/dps_report/changes.py new game <id> "<title>"` with the
announcement as `--source-label`/`--source-url` (standards 30). A cosmetic build gets none. The audit loop's
client_update auditor writes the ledger (`orchestration/update_ledger.md`) on its own
schedule.

## 3. foreverchanges.pro

```bash
python3 tools/database/foreverchanges/scrape.py
python3 tools/database/foreverchanges/suffixes.py <max level> <cache dir>
```

`scrape.py` writes the loot table (`foreverchanges_loot.json`) and `suffixes.py` the
random suffixes. Both write a `meta` block, and `stale.py` reads their age from it.

## 4. Wowhead

Follow the `scrape-wowhead` skill. It knows the bot check and the rate limit, and the
scrapes take a while.

## 5. AtlasLoot

When `stale.py` says AtlasLoot Forever changed (CurseForge updated it):

```bash
go run ./tools/database/gen_db -outDir=assets -gen=atlasloot -atlasloot="<beta client>/Interface/AddOns"
python3 -I tools/sources/crosscheck.py
```

The cross-check lists where AtlasLoot and Wowhead disagree on a drop, or only one of them
knows it. Tell the user about new conflicts and about Forever items only Wowhead knows,
because those are loot AtlasLoot hasn't caught up with. AtlasLoot lists Zul'Gurub's
quest rewards as drops, so those show up every time.

## 6. What we read out of the client's tables

After a new client build is fetched (step 2), these follow it. `stale.py` names the ones
behind.

```bash
python3 tools/sources/dbcache.py "<beta client>" --write       # our client's hotfix cache
python3 tools/sources/itemsparse.py <new>                      # item sets, factions, PvP ranks
python3 tools/sources/enchants.py <new>                        # enchant values
python3 tools/sources/racials.py <new> --write                 # racial text for the UI
python3 tools/sources/abilities.py <new>                       # ability text for the DPS report
python3 tools/forever_talents/import_db2.py <new>              # talent trees, prints what changes
python3 tools/forever_talents/import_db2.py <new> --write
```

Read what each prints. The talent importer lists every talent whose structure, numbers or
wording changed, and a renamed talent changes the sim's talent strings once its tree is
regenerated (`import_client.py` and `import_talents.py`), so that part waits for the
`dps-report` skill's full pass. The enchant and item inputs only change the database,
which step 7 regenerates. A racial whose numbers changed needs `sim/core/racials_forever*.go`
too.

## 7. Regenerate, test, commit

```bash
go run ./tools/database/gen_db -outDir=assets -gen=db
python3 tools/sources/hidden.py
make test
python3 tools/audit/gate.py
```

`hidden.py` lists the items Forever still hides and says which were revealed since the
last run. Those are worth a Wowhead scrape for their stats. `dbcache.py` reads our own
client, so it runs whenever the user has played since the last run (`stale.py --addons`
says when it was sent hidden items), not only on a new build.

Before committing, compare the new `assets/database/db.json` with the old one by item:
how many items were added or removed, and which fields changed. A refreshed scrape
should mostly add sources. Explain anything that removes items or changes stats.
`gate.py` shows which auditors the commit wakes. Commit with the `cm` skill, one commit
per source, saying what changed in the data.

Then bayovui's `/update-data` picks the new data up (BiS lists, recipe sources, weights).
