# Phase 0 gold line: price every list, one roller on drops (2026-11-25)

The live DPS report (master 86fbcf456, Phase 0 gear) lets through items over the gold line. Elemental
Mage Staff (944) is the Fire, Frost and Frostfire mainhand: one auction at 2000g (BayovAuctionData,
13 Nov, count 1), +12.5 dps over Headmaster's Charge for Fire, so 160g per DPS against the 30g line.
`phase0_content.json` has no gold for it, so the search treated it as free. `tools/rotopt/phase0_prices.py`
prices only the list it's given (default Enhancement), and it was never run on most lists.

## 1. One roller on dungeon drops (user, 2026-11-25)

Assume only we roll on a dungeon drop: roll share 1.0, not 0.5. That halves a drop's farming effort,
so drops get double the budget. Change the default in phase0_prices.py, its docstring and
test_phase0_prices.py, and any note that says 0.5 or "one other player".

## 2. Farming inputs

The effort price needs BayovWikiGear `Data*.lua`, BayovWikiDungeons `Data.lua` and AtlasLoot Classic's
`AtlasLootClassic_DungeonsAndRaids/droprate.lua` (see `newest_file` and `main`). None are in a cloud
container by default. Add bayov/wow with add_repo (read). Build the BayovWiki data with bayovui's own
build (see its AGENTS.md, README.md, `bayovui/build/`), and get droprate.lua from AtlasLoot Classic's
public source at the version in the wow repo's tracked `_forever_/Interface/AddOns/AtlasLootClassic*/*.toc`.
The auction data is the wow repo's `_forever_/WTF/Account/*/SavedVariables/BayovAuctionData.lua`, which is
the newest the user has synced. Say its date in the notes. If an input really can't be had, stop and
tell the parent which one, before any rerun.

## 3. Price every list

Make phase0_prices.py price every spec list in `tools/rotopt/phase0_specs.json` in one run (an item on
several lists takes its own goldPerDps per list, so store what the line needs; read how bis_phase0.sh
reads goldPerDps and keep the two in step). Add a test. Run it with `--write`.

## 4. Re-rank and re-sim

For each spec, find slots where the new prices change what the list may pick: a pick now over the line,
or a ranked item now under it (drops got cheaper). Re-rank those slots (`SLOTS=`, see
`briefs/phase0_step3_1122.md`), write the gear set (`GEAR_ONLY=1 SPEC=<spec>`), update the requests and
race rows that wear it (`tools/dps_report/race_gear.py`), and check each request sims to its list.
Delegate per class group to subagents when it helps.

## 5. Publish

`stale.mjs`, `pipeline.sh hydrate/run/publish` for the changed ids, `build_data.py`, `changes.py` and a
Changes tab entry (gold line now prices every list, one roller on drops, which builds moved). Run the
gear, results and rotations auditors on the changed builds, fix round, merge to master (fetch and merge
first, never force-push), push your branch too, `make dps-report-deploy`, and check the live page.

## Hand back

One message to the parent, under 300 words: master commit, deploy check, the auction data date, per build
what changed (item, dps before and after), and a table of total gold per gear set with the full
pricing (gear only, say if enchants are priced). Then stop.
