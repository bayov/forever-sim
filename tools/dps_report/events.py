#!/usr/bin/env python3
"""events.py [--data DIR] [--out DIR] [-j N] [--lookups-only] [build ids...] : the sample fights for the Specs page's events view.

The Specs page shows each preset's typical, unlucky and lucky fight in the sim UI's own events view (timeline, chart,
log and replay). That view reads a whole RaidSimResult with the fight's debug log, not the compact timeline that
build_data.py writes. So we sim each sample fight again, with the same seed and one iteration, and keep its log.

The seeds come from the timeline_<class>.json datasets in --data (default: the report's own data folder). For each
build we write <out>/<build>.json.gz (default out: <data>/events/), a gzipped JSON object:

  request  the build's RaidSimRequest, as simmed
  result   the typical fight's RaidSimResult without its log, plus typicalIteration, unluckyIteration and
           luckyIteration with each fight's seed, DPS, duration and debug log

The page puts the typical log back into result.logs when it loads it. A fight whose DPS doesn't match the dataset's
is reported: the binary isn't the one the report's numbers came from.

The events view names and draws every spell and item the way the sim UI does: from the item database first, then
from wowhead. The single-file report can't reach wowhead or its icons, so we also write two lookup datasets to --data:

  event_names  the name, icon and rank of every spell and item in the fights (and of each build's weapons)
  event_icons  the icons (36 px, as data URLs) that icons.json and icons_more.json don't have yet
"""
import argparse, base64, copy, gzip, json, os, re, subprocess, sys, tempfile, urllib.request
from concurrent.futures import ThreadPoolExecutor

import whcache
from dashpaths import D, BIN, TREE

HERE = os.path.dirname(os.path.abspath(__file__))
REPORT_DATA = os.path.join(HERE, '..', '..', 'ui', 'dps_report', 'data')
CLI = os.path.join(BIN, 'wowsimcli')
# The report's percentile of each sample fight, as the events view's iterations.
KINDS = {50: 'typicalIteration', 10: 'unluckyIteration', 90: 'luckyIteration'}
ROOT = os.path.join(HERE, '..', '..')
# The main hand, off hand and ranged slots: the events view draws melee and shots with the weapon's icon.
WEAPON_SLOTS = (14, 15, 16)


def sim(req):
    with tempfile.TemporaryDirectory() as d:
        i, o = os.path.join(d, 'i.json'), os.path.join(d, 'o.json')
        json.dump(req, open(i, 'w'))
        subprocess.run([CLI, 'sim', '--infile', i, '--outfile', o], check=True, capture_output=True, cwd=TREE if os.path.isdir(TREE) else None)
        r = json.load(open(o))
    if r.get('error'):
        raise RuntimeError(r['error']['message'][:800])
    return r


def fight(req, seed):
    req = copy.deepcopy(req)
    # The events view finds each unit's lines by its name in the log, like "[Player (#1)]". The requests leave the
    # name empty, which the view can't read, so we give it the sim UI's name.
    for party in req['raid']['parties']:
        for player in party['players']:
            if player.get('class') and not player.get('name'):
                player['name'] = 'Player'
    req['simOptions'] = {'iterations': 1, 'randomSeed': str(seed), 'debugFirstIteration': True}
    return req, sim(req)


def build_events(bid, samples, out_dir):
    req = json.load(open(os.path.join(D, 'req', bid + '.json')))
    result, request, warnings = None, None, []
    for s in samples:
        kind = KINDS.get(int(s['pct']))
        if not kind:
            continue
        r_req, r = fight(req, s['seed'])
        dps = r['raidMetrics']['dps']['avg']
        if abs(dps - float(s['dps'])) > 0.05:
            warnings.append(f"{bid} seed {s['seed']}: {dps:.1f} DPS here, {float(s['dps']):.1f} in the dataset")
        iteration = {'seed': str(s['seed']), 'dps': dps, 'duration': r['firstIterationDuration'], 'logs': r['logs']}
        if kind == 'typicalIteration':
            request = r_req
            result = {k: v for k, v in r.items() if k not in ('logs', 'typicalIteration', 'luckyIteration', 'unluckyIteration', 'error')}
            result['iterationsDone'] = len(samples)
        else:
            result = result or {}
        result[kind] = iteration
    if request is None:
        return bid, 0, [f'{bid}: no typical fight in the dataset']
    payload = json.dumps({'request': request, 'result': result}, separators=(',', ':')).encode()
    path = os.path.join(out_dir, bid + '.json.gz')
    with open(path, 'wb') as f:
        # mtime=0 so the same fights give the same bytes.
        f.write(gzip.compress(payload, 9, mtime=0))
    return bid, os.path.getsize(path), warnings


def fight_ids(path):
    """Every ('spell' | 'item', id) in a build's fights and their metrics, and its weapons."""
    d = json.loads(gzip.open(path).read())
    ids = set()
    for kind in KINDS.values():
        for m in re.finditer(r'\{(SpellID|ItemID): (\d+)', d['result'].get(kind, {}).get('logs', '')):
            ids.add(('spell' if m.group(1) == 'SpellID' else 'item', int(m.group(2))))
    # The metrics list every spell a unit has, also the ones it never cast.
    metrics = json.dumps({k: v for k, v in d['result'].items() if k not in KINDS.values()})
    for kind, i in re.findall(r'"(spell|item)Id": (\d+)', metrics):
        ids.add((kind, int(i)))
    for party in d['request']['raid']['parties']:
        for player in party['players']:
            items = player.get('equipment', {}).get('items', [])
            for slot in WEAPON_SLOTS:
                if slot < len(items) and items[slot].get('id'):
                    ids.add(('item', items[slot]['id']))
    return ids


def lookups(ids, data_dir):
    """The event_names and event_icons datasets for these ids, like the sim UI's Database would name them."""
    db = json.load(open(os.path.join(ROOT, 'assets', 'database', 'db.json')))
    known = {('item', i['id']): (i.get('name'), i.get('icon'), 0) for i in db['items']}
    known.update({('item', i['id']): (i.get('name'), i.get('icon'), i.get('rank', 0)) for i in db.get('itemIcons', [])})
    known.update({('spell', i['id']): (i.get('name'), i.get('icon'), i.get('rank', 0)) for i in db.get('spellIcons', [])})
    names = []
    for kind, i in sorted(ids):
        name, icon, rank = known.get((kind, i), (None, None, 0))
        if not name:
            j = whcache.fetch('forever', kind, i)
            if not whcache.ok(j):
                j = whcache.fetch(whcache.FALLBACK, kind, i)
            if whcache.ok(j):
                name, icon = j['name'], j.get('icon')
                m = re.search(r'Rank (\d+)', j.get('tooltip', '')) if kind == 'spell' else None
                rank = int(m.group(1)) if m else 0
        names.append({'key': kind[0] + str(i), 'name': name or '', 'icon': (icon or '').lower() or None, 'rank': rank or None})

    # The icons the sim UI draws without asking the database: melee, mana ticks, resources and the like.
    action_id = open(os.path.join(ROOT, 'ui', 'core', 'proto_utils', 'action_id.ts')).read()
    fixed = {m.lower() for m in re.findall(r'wow\.zamimg\.com/images/wow/icons/\w+/([\w-]+)\.jpg', action_id)}
    have = set()
    for f in ('icons.json', 'icons_more.json'):
        p = os.path.join(data_dir, f)
        if os.path.exists(p):
            have |= {r['icon'] for r in json.load(open(p))}
    want = sorted(({n['icon'] for n in names if n['icon']} | fixed | {'inv_misc_questionmark'}) - have)
    os.makedirs(os.path.join(D, 'icons'), exist_ok=True)

    def fetch_icon(ic):
        path = os.path.join(D, 'icons', ic + '.jpg')
        if not os.path.exists(path):
            try:
                with urllib.request.urlopen(f'https://wow.zamimg.com/images/wow/icons/medium/{ic}.jpg', timeout=20) as r:
                    open(path, 'wb').write(r.read())
            except Exception:
                return ic, None
        return ic, 'data:image/jpeg;base64,' + base64.b64encode(open(path, 'rb').read()).decode()

    with ThreadPoolExecutor(16) as ex:
        icons = [{'icon': ic, 'url': url} for ic, url in ex.map(fetch_icon, want) if url]
    missing = sorted(set(want) - {r['icon'] for r in icons})
    return names, icons, missing


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--data', default=REPORT_DATA)
    ap.add_argument('--out')
    ap.add_argument('-j', type=int, default=os.cpu_count() or 4)
    ap.add_argument('--lookups-only', action='store_true', help="sim no fights, only rewrite event_names and event_icons from the fights in --out")
    ap.add_argument('builds', nargs='*')
    a = ap.parse_args()
    out_dir = a.out or os.path.join(a.data, 'events')
    os.makedirs(out_dir, exist_ok=True)

    samples = {}
    for f in sorted(os.listdir(a.data)):
        if f.startswith('timeline_') and f.endswith('.json'):
            for row in json.load(open(os.path.join(a.data, f))):
                samples.setdefault(row['build'], []).append(row)
    # The typical fight first: its result holds the metrics the others share.
    for rows in samples.values():
        rows.sort(key=lambda s: int(s['pct']) != 50)
    builds = [] if a.lookups_only else a.builds or sorted(samples)

    total, warnings = 0, []
    with ThreadPoolExecutor(a.j) as ex:
        for bid, size, warn in ex.map(lambda b: build_events(b, samples.get(b, []), out_dir), builds):
            total += size
            warnings += warn
            print(f'{bid}: {size / 1024:.0f} KB', flush=True)
    print(f'{len(builds)} builds, {total / 1024 / 1024:.1f} MB')

    # The lookups cover every build in the folder, so a partial run still writes whole datasets.
    ids = set()
    for f in sorted(os.listdir(out_dir)):
        if f.endswith('.json.gz'):
            ids |= fight_ids(os.path.join(out_dir, f))
    names, icons, missing = lookups(ids, a.data)
    for name, rows in (('event_names', names), ('event_icons', icons)):
        p = os.path.join(a.data, name + '.json')
        json.dump(rows, open(p, 'w'), separators=(',', ':'))
        print(f'{name}: {len(rows)} rows, {os.path.getsize(p) / 1024:.0f} KB')
    unnamed = [n['key'] for n in names if not n['name']]
    if unnamed:
        print('No name for:', ' '.join(unnamed), file=sys.stderr)
    if missing:
        print('No icon for:', ' '.join(missing), file=sys.stderr)
    for w in warnings:
        print('DPS mismatch:', w, file=sys.stderr)


if __name__ == '__main__':
    main()
