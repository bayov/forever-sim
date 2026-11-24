#!/usr/bin/env python3
# /// script
# dependencies = ["curl_cffi", "websocket-client"]
# ///
# Scrape what fishing catches in each Forever zone into
# assets/db_inputs/forever_wowhead_fishing.json.
#
# The sim doesn't read this. bayovui's fishing.py does, for BayovCore's Level Fishing goal,
# and it lives here because this is where we scrape wowhead. cmangos' classic-db knows
# Classic's fishing loot, but not Forever's own zones (Riverglades, Shen'dralas, Zephras
# Isle ...) or what Forever changed. A zone page's "Fishing" tab counts what players
# caught there: `_totalCount` catches in all, and for each item how many catches gave it
# (`count`) and how many a catch gave (`stack`, min and max).
#
# Which zones: every area of the Forever client with a zone map of its own (AreaTable,
# UiMap and UiMapAssignment from wago.tools), plus the dungeons and subzones in
# BayovWikiProfessions' fishing bands. That's about 63 pages, so a run takes about 13
# minutes at wowhead.py's pace, longer if wowhead asks for a captcha. Pages are cached, so
# a stopped run picks up where it stopped.
#
#   uv run tools/database/forever_wowhead/fishing.py [--build 1.60.1.71037]
#
# Output: { "meta": {...}, "zones": { areaId: { "name", "map", "continent", "total",
#   "items": [ { "id", "name", "count", "pct", "stack": [min, max] } ] } } }
# `pct` is count / total, in percent: the chance a catch gives the item.
import csv, io, json, os, sys, time, urllib.request

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from wowhead import CACHE, Wowhead, listview  # noqa: E402

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..', '..'))
OUT = os.path.join(ROOT, 'assets', 'db_inputs', 'forever_wowhead_fishing.json')
BUILD = '1.60.1.71037'
WAGO = 'https://wago.tools/db2/%s/csv?build=%s'

# Fishable dungeons and subzones from BayovWikiProfessions' bands: Blackfathom Deeps, The
# Deadmines, Wailing Caverns, Scarlet Monastery, Jaguero Isle, Maraudon, Sunken Temple, Bay
# of Storms, Hetaera's Clutch, Jademir Lake, Scalebeard's Cave, Scholomance, Stratholme.
EXTRA = [719, 1581, 718, 796, 297, 2100, 1417, 1227, 1222, 1112, 3140, 2057, 2017]


def wago(name, build):
    path = os.path.join(CACHE, '..', 'wago', '%s-%s.csv' % (name, build))
    if not os.path.exists(path):
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with urllib.request.urlopen(urllib.request.Request(
                WAGO % (name, build), headers={'User-Agent': 'forever-sim fishing.py'}), timeout=120) as r:
            data = r.read()
        with open(path + '.part', 'wb') as f:
            f.write(data)
        os.replace(path + '.part', path)
    with open(path, encoding='utf-8') as f:
        return list(csv.DictReader(f))


def zones(build):
    """{area id: (name, zone map or None, continent)} for the zones to scrape."""
    areas = {int(r['ID']): r for r in wago('AreaTable', build)}
    kind = {int(r['ID']): int(r['Type'] or 0) for r in wago('UiMap', build)}
    area_map = {}
    for r in wago('UiMapAssignment', build):
        a, m = int(r['AreaID'] or 0), int(r['UiMapID'] or 0)
        if a and m and (a not in area_map or (kind.get(m) == 3 and kind.get(area_map[a]) != 3)):
            area_map[a] = m
    out = {}
    for a, m in area_map.items():
        r = areas.get(a)
        if r and kind.get(m) == 3 and not int(r['ParentAreaID'] or 0):
            out[a] = (r['AreaName_lang'], m, int(r['ContinentID'] or 0))
    for a in EXTRA:
        r = areas.get(a)
        if a not in out and r:
            out[a] = (r['AreaName_lang'], None, int(r['ContinentID'] or 0))
    return out


def main(argv):
    build = argv[argv.index('--build') + 1] if '--build' in argv else BUILD
    todo = zones(build)
    print('%d zones to read, about %d minutes if none is cached' % (len(todo), len(todo) * 13 // 60 + 1))
    wh = Wowhead()
    result, empty = {}, []
    t0 = time.time()
    for n, (area, (name, umap, cont)) in enumerate(sorted(todo.items(), key=lambda x: x[1][0]), 1):
        html = wh.page('/forever/zone=%d' % area)
        rows = listview(html, 'fishing') or [] if html else []
        total = 0
        if html and "id: 'fishing'" in html:
            i = html.find("id: 'fishing'")
            j = html.find('_totalCount:', i)
            if j > 0 and j - i < 2000:
                total = int(''.join(c for c in html[j + 12:j + 30].split(',')[0] if c.isdigit()) or 0)
        items = []
        for r in rows:
            c = int(r.get('count') or 0)
            if c <= 0:
                continue
            items.append({'id': r['id'], 'name': r.get('name'), 'count': c,
                          'pct': round(100.0 * c / total, 3) if total else None,
                          'stack': r.get('stack') or [1, 1]})
        items.sort(key=lambda x: (-x['count'], x['id']))
        if not items:
            empty.append(name)
        else:
            result[str(area)] = {'name': name, 'map': umap, 'continent': cont, 'total': total, 'items': items}
        print('[%d/%d %ds] %s (%d): %d catches, %d items'
              % (n, len(todo), time.time() - t0, name, area, total, len(items)), flush=True)
    meta = {'source': 'https://www.wowhead.com/forever/zone=<id>, Fishing tab', 'build': build,
            'scraped': time.strftime('%Y-%m-%d'), 'zones': len(result), 'noFishing': sorted(empty)}
    with open(OUT, 'w', encoding='utf-8', newline='\n') as f:
        json.dump({'meta': meta, 'zones': result}, f, indent=1, sort_keys=True)
        f.write('\n')
    print('wrote %s: %d zones with catches, %d without (%s)'
          % (os.path.relpath(OUT, ROOT), len(result), len(empty), ', '.join(sorted(empty))))
    return 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
