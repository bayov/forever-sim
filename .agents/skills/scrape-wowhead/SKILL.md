---
name: scrape-wowhead
description: Scrape Wowhead's Forever database into db_inputs (gear, quests, recipe sources) past its bot check and rate limit, then regenerate what reads them. Use for any www.wowhead.com page, when Wowhead has new Forever data, when stale.py says the scrapes are old, after a 403 from Wowhead, or on /scrape-wowhead.
---

# Scrape Wowhead's Forever database

Wowhead's Forever database (www.wowhead.com/forever) reads the beta client, so it knows
the items, vendors and quests Forever added or changed before anything else does. We
scrape three things from it into `assets/db_inputs/`:

| Script | Writes | Read by |
|---|---|---|
| `tools/database/forever_wowhead/scrape.py` | `forever_wowhead_items.json`: every wearable item up to a required level, with stats, sources, and the class or side of the quest it comes from | `gen_db` (the item database), `rotopt` (items with no known source), bayovui's `bis.py` (BiS sources), `orchestration/tools/*` |
| `tools/database/forever_wowhead/recipes.py` | `forever_wowhead_recipes.json`: every recipe item, with its vendors (stock, price, map pins), drops with chances, chests, quests, pickpocketing, fishing and containers. Also what a trainer asks for each recipe, and who sells each reagent and whether in limited stock | bayovui's `recipesources.py` (BayovCore's recipe planner), nothing in the sim |
| `tools/database/forever_wowhead/class_spells.py` | `forever_wowhead_class_spells.json`: every class's ability ranks, with what a trainer asks for each and where it comes from (a trainer, a quest or a drop). Nine pages | bayovui's `classspells.py` (BayovCore's Training prices), nothing in the sim |

All three use `tools/database/forever_wowhead/wowhead.py`. Its header has the details.

## Getting past Wowhead's defences

Wowhead sits behind AWS WAF, with two separate checks:

1. **A bot check.** curl gets a 403 for every page, even `robots.txt`. A browser runs a
   JavaScript challenge and earns an `aws-waf-token` cookie that lasts about four days.
   `wowhead.py` earns it once in headless Chrome (`google-chrome` must be installed),
   keeps it in `~/.cache/wow-data/wowhead/token.json`, and fetches every page with
   curl_cffi, which looks like Chrome.
2. **A rate limit.** About 110 requests in a minute or two gets the address a 403 for
   everything, browser or not, for some minutes, and a new token doesn't help.
   `wowhead.py` asks once every 3.5 seconds (`DELAY`). On a 403 it waits five minutes,
   then longer, and slows down for the rest of the run. Don't run two scrapes at once,
   and don't loop over Wowhead pages from your own code: use `Wowhead.page()`.

Every page is kept, gzipped, in `~/.cache/wow-data/wowhead/pages`, so a stopped run picks
up where it left off. A page younger than `WOWHEAD_CACHE_DAYS` (7 by default) isn't
fetched again. Use `WOWHEAD_CACHE_DAYS=0` when you need something Wowhead changed today,
and for `recipes.py` after the game patched. Its meta block's `oldestPage` gives the age
of the oldest page it used, and BayovCore lets that data replace what it saw at a vendor
or trainer only when the scrape is newer than its look, so cached pages from before the
patch can't.

The tooltip API (`nether.wowhead.com/forever/tooltip/item/<id>?dataEnv=16`) has neither
check, and plain urllib works for it. `scrape.py` uses it for binding.

## Steps

Run everything from `/home/bayov/projects/forever-sim`. The scripts declare their own
dependencies, so `uv run` them.

1. **Scrape.** They are long, so run them in the background:

   ```bash
   uv run tools/database/forever_wowhead/recipes.py   # ~2,900 pages, a couple of hours fresh
   uv run tools/database/forever_wowhead/class_spells.py   # nine pages, a minute
   uv run tools/database/forever_wowhead/scrape.py    # a few hundred pages
   ```

   `scrape.py` takes an optional highest required level, by default the one the current
   file was scraped with (its `maxReqLevel`). Each prints what it found per listing and a
   total at the end.

2. **Check the diff before using it.** Run `jj diff --stat assets/db_inputs` and compare
   the totals with the last run. Far fewer items, or a profession listing with 0 recipes,
   means Wowhead changed its pages, not that Forever removed things. Fix the parser in
   `wowhead.py` and run again (the page cache makes that quick). A listing at 1,000 rows
   is cut short, and `scrape.py` splits those by level by itself.

3. **Regenerate the item database** after `scrape.py`. Never hand-merge
   `assets/database/db.*`.

   ```bash
   go run ./tools/database/gen_db -outDir=assets -gen=db
   make test
   ```

4. **Commit** with the `cm` skill: the scrape files, the regenerated database, and any
   parser fix, each as its own commit.

5. **bayovui.** The WoW folder (`/mnt/d/Gaming/World of Warcraft`) reads these files. Its
   `/update-data` skill picks them up (`bayovui/build/stale.py` lists what's behind), or
   run from there:

   ```bash
   uv run bayovui/build/recipesources.py beta   # recipe sources
   uv run bayovui/build/bis.py                  # BiS sources
   uv run bayovui/test/run_addon.py core
   ```

   Then commit and deploy BayovCore there.

## One page, by hand

To look something up on Wowhead in a script or a check, use the same client, never curl:

```python
import sys; sys.path.insert(0, 'tools/database/forever_wowhead')
from wowhead import Wowhead, listview, listview_items, gatherer, mapper
wh = Wowhead()
html = wh.page('forever/item=6270')
listview(html, 'sold-by')     # also 'dropped-by', 'contained-in-object', 'reward-from-q', ...
mapper(html)                  # map pins by zone ID (AreaTable IDs)
```

Run it with `uv run --with curl_cffi --with websocket-client python`. `listview_items`
reads a listing page's rows, and `gatherer(html, 3)` reads its item stats (`jsonequip`).

## Not covered here

- The daily cloud routine doesn't scrape Wowhead, since it can't count on Chrome. So this
  skill runs locally, and `tools/sources/stale.py` flags scrapes older than a week (by the
  fetch date in each file's `meta` block). `docs/data_sources.md` lists every source the
  scrapes feed and who reads them.
