"""The `meta` block each scraped input in assets/db_inputs/ carries.

A scraper writes it as the input's first key:

    {"meta": {"dataEnv": 16, "fetched": "2026-10-17T15:26Z", "source": "www.wowhead.com/forever",
              "tool": "tools/database/forever_wowhead/scrape.py"},
     "items": [...]}

- source: where the data came from (a site, or a client build for game data)
- tool: the script that wrote the file
- fetched: when, in UTC. tools/sources/stale.py reads this for a scrape's age,
  not the file's modification time, which a checkout or a copy resets.
- anything that pins the data further, like Wowhead's dataEnv or a client
  build

Every reader decodes these files by named keys (Go structs, Python dicts), so
the extra key changes nothing for them. docs/data_sources.md lists every input
and says which carry a meta block.
"""
import datetime
import json


def meta(source, tool, **pins):
    return dict(pins, source=source, tool=tool,
                fetched=datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%MZ'))


def line(m):
    """The block as the first line of a file we write by hand, one entry per line."""
    return '"meta": %s,\n' % json.dumps(m, sort_keys=True)


def read(path):
    """The meta block of an input, or None when it has none."""
    with open(path, encoding='utf-8') as f:
        return json.load(f).get('meta')
