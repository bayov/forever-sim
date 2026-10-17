"""Pages of https://foreverchanges.pro, and the data in them.

The site is Next.js, server rendered. What a page shows (a dungeon's bosses and
their loot, the hidden items list, an item's tooltip) sits in its React payload:
chunks of `self.__next_f.push([1, "..."])` in the HTML, each a JSON string.
`payload()` joins them back into one text, and the readers pick their JSON
objects out of it.

tools/database/foreverchanges/scrape.py (dungeon loot) and suffixes.py (item
pages) read through here, and so does the daily routine for the hidden items
page (tools/forever_update/ROUTINE.md).
"""
import json
import os
import re
import urllib.error
import urllib.request

SITE = 'https://foreverchanges.pro'
CHUNK = re.compile(r'self\.__next_f\.push\(\[1,(".*?")\]\)</script>', re.S)


def fetch(path, cache=None, missing_ok=False):
    """The HTML of a page, like fetch('/dungeons/the-deadmines').

    With cache (a file path), a page we already have isn't fetched again. With
    missing_ok, a 404 comes back as an empty page (and is cached as one), for
    the item pages the site doesn't have.
    """
    if cache and os.path.exists(cache) and os.path.getsize(cache) > 0:
        with open(cache, encoding='utf-8') as f:
            return f.read()
    req = urllib.request.Request(SITE + path, headers={'User-Agent': 'Mozilla/5.0'})
    try:
        with urllib.request.urlopen(req, timeout=60) as r:
            page = r.read().decode()
    except urllib.error.HTTPError as e:
        if not (missing_ok and e.code == 404):
            raise
        page = '<!-- 404 -->'
    if cache:
        os.makedirs(os.path.dirname(cache) or '.', exist_ok=True)
        with open(cache, 'w', encoding='utf-8') as f:
            f.write(page)
    return page


def payload(page):
    """The page's React payload as one text."""
    return ''.join(json.loads(c) for c in CHUNK.findall(page))
