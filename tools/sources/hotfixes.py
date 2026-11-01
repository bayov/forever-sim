#!/usr/bin/env python3
"""Server hotfix rows from wago.tools, laid over a client build's tables.

    python3 tools/sources/hotfixes.py fetch <build> [Table ...] [--set NAME] [--search TEXT ...] [--refresh]
    python3 tools/sources/hotfixes.py pushes <build> [Table ...] [--region N]
    python3 tools/sources/hotfixes.py rows <build> <Table> [--record ID ...] [--region N]
    python3 tools/sources/hotfixes.py check <build> [Table ...] [--region N]
    python3 tools/sources/hotfixes.py apply <build> [Table ...] [--set NAME] [--region N]

The realms send every client the rows they changed since its build was made, as
hotfixes, with no new build. Talent reworks reach the game that way days before
a client build has them (the paladin and druid trees of the 2026-10-30 hotfix,
push 113283 and its neighbours, in region 91 on build 70707). wago.tools logs the
rows its users' clients were sent at https://wago.tools/hotfixes?product=wowf,
and docs/data_sources.md ranks them above the client tables for the rows they
change.

`apply` writes a build's tables with the hotfix rows applied, as the cached
pseudo-version `<version>+hf` (like 1.60.1.70707+hf) next to the build's own
folder in wago.py's cache. Everything that reads tables through wago.py reads it
like any build:

    python3 tools/sources/hotfixes.py fetch 1.60.1.70707 --set talents --set spelltext
    python3 tools/sources/hotfixes.py apply 1.60.1.70707 --set talents --set spelltext
    python3 tools/forever_talents/import_db2.py 1.60.1.70707+hf paladin druid

wago.py builds a table of a `+hf` version it doesn't have yet from the hotfix
rows on disk (it never downloads them), so `apply` is only needed to see the
counts or after a `fetch`.

How the rows are found. The page's `data-page` attribute (HTML-escaped JSON)
has `props.hotfixes`, a Laravel page of 25 rows, newest push first. Each row has
push_id, record_id, status, build, table_name, data (the row's fields, or null),
created_at, region_id, locale, wago's own row id and a search_text like
"enus 113283 1328613 70707 spellname [1328613,...]". `search` is a plain
substring match on that text, so "70707 trait" finds every row of build 70707 in
a table whose name starts with Trait, and nothing else. Paging through every
row without a search covers only about a day in 200 pages (item reveals,
mostly), so `fetch` asks once per table family: "<build number> <table name>",
dropping a search another one already covers ("70707 spell" covers SpellName,
SpellEffect and the rest). That is complete for the tables asked for, as long as
wago.tools' users were sent them: it only has what someone's client reported.
`--search` adds any other text (a push id, a spell id, a name).

The site's page order is unstable among rows logged in the same second, which is
most rows of a push, so paging a big search loses rows. gather() splits a search
by the record id's last digits until each part fits in a few pages and checks
every part against the site's count (see there). On 2026-11-01 the paladin and
druid trees' tables took 6,668 rows in about 470 requests, all of them region 91
and build 70707.

Searches are cached under ~/.cache/wow-data/wago/hotfixes/<product>/search/, one
JSON file per search with every row it found, the site's count and when. A search
read completely isn't asked again; `--refresh` reads it again, for new pushes.
Requests go one at a time, a second and a half apart.

Which rows count. Only the build's own rows (the build column) of one region,
91 by default: the Forever launch realms. Other regions run other pushes (region
3 includes retail's), and rows logged for other builds are other clients' data.
Only enUS rows, since our tables are enUS. For each (table, record id) the
highest push id wins, as in the client. Status 1 is a row with data, which
replaces or adds the record; status 2 has no data and removes the record (the
client's "record removed"), which is how a hotfix takes a talent node, edge or
effect out of a tree. Statuses 3 (invalid) and 4 (not public yet) keep the
build's row and are counted in `apply`'s output.

The field order. A row's data lists the table's fields in the order of wago.tools'
CSV header for the same table, with array fields as nested lists where the CSV
expands them into _0, _1 columns. `apply` flattens the arrays and refuses a table
whose rows have another field count. `check` compares each hotfixed row with the
build's row of the same record field by field: a wrong layout shifts every field
after the first wrong one, so most rows differ in most fields, where a real
hotfix changes a field or two. Floats are written the CSV's way, as float32 to 11
significant digits.

Run anything that reads the cached rows with python3 -I. They're downloaded,
untrusted data.
"""
import argparse
import collections
import csv
import datetime
import html
import json
import os
import re
import shutil
import struct
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import wago  # noqa: E402

PRODUCT = 'wowf'
REGION = 91
SUFFIX = '+hf'
URL = 'https://wago.tools/hotfixes?%s'
AGENT = 'forever-sim tools/sources/hotfixes.py'
PAUSE = 1.5
# A search with more pages than this is split rather than paged (gather), and splitting
# stops this many digits deep.
MAX_PAGES = 4
MAX_DEPTH = 5
PAGE_DATA = re.compile(r'data-page="([^"]*)"')

STATUS_DATA, STATUS_REMOVED = 1, 2


def cache_dir(product=PRODUCT):
    return os.path.join(wago.CACHE, 'hotfixes', product)


def slug(text):
    return re.sub(r'[^a-z0-9]+', '_', text.lower()).strip('_') or 'all'


def base_version(v):
    """The client build a version is laid over: 1.60.1.70707+hf is 1.60.1.70707."""
    return v.split('+', 1)[0]


def derived(build):
    return wago.version(build).split('+', 1)[0] + SUFFIX


# Fetching.

def page(search, number, product=PRODUCT):
    """One page of wago.tools' hotfix list for a search, as props.hotfixes."""
    q = urllib.parse.urlencode({'product': product, 'search': search, 'page': number})
    last = None
    for attempt in range(4):
        try:
            req = urllib.request.Request(URL % q, headers={'User-Agent': AGENT})
            with urllib.request.urlopen(req, timeout=120) as r:
                body = r.read().decode('utf-8')
            m = PAGE_DATA.search(body)
            if not m:
                raise ValueError('no data-page attribute')
            return json.loads(html.unescape(m.group(1)))['props']['hotfixes']
        except urllib.error.HTTPError as e:
            last = e
            if e.code in (400, 404):
                break
        except (OSError, ValueError, KeyError) as e:
            last = e
        time.sleep(10 * (attempt + 1))
    raise OSError('%s: %s' % (URL % q, last))


def search_path(search, product=PRODUCT):
    return os.path.join(cache_dir(product), 'search', slug(search) + '.json')


def load_search(search, product=PRODUCT):
    p = search_path(search, product)
    if not os.path.exists(p):
        return None
    with open(p, encoding='utf-8') as f:
        return json.load(f)


def gather(search, product=PRODUCT, depth=0):
    """Every row a search finds, as ({wago row id: row}, total, requests).

    The site's order is unstable among rows logged in the same second (most of a push), so
    paging a big search returns some rows twice and never shows others: "70707 spell" gave
    5,292 distinct rows of 6,134 in 246 pages. So we page a search only when it has at most
    MAX_PAGES pages, and keep the result only when it has as many distinct rows as the site
    counts. Otherwise we split it by the record id's last digits: in "enus 113283 1328613
    70707 spell [...]" the text "3 70707 spell" matches the records whose id ends in 3, and
    "13 70707 spell" those ending in 13. The ten children of a search share none of its
    rows, so they cover it, except a record whose whole id is the digits so far (the
    search can't ask for the space before it); `missing` counts those.
    """
    rows = {}
    h = page(search, 1, product)
    total, last, requests = int(h.get('total') or 0), int(h.get('last_page') or 1), 1

    def add(h):
        for r in h.get('data') or []:
            rows[str(r['id'])] = {k: v for k, v in r.items() if k != 'search_text'}

    add(h)
    if last <= MAX_PAGES or depth >= MAX_DEPTH:
        for n in range(2, last + 1):
            time.sleep(PAUSE)
            add(page(search, n, product))
            requests += 1
        print('  %s: %d of %d rows in %d page%s' % (search, len(rows), total, last, '' if last == 1 else 's'),
              file=sys.stderr)
        if len(rows) >= total or depth >= MAX_DEPTH:
            return rows, total, requests
    for d in range(10):
        time.sleep(PAUSE)
        child = ('%d%s' if depth else '%d %s') % (d, search)
        got, _, n = gather(child, product, depth + 1)
        rows.update(got)
        requests += n
    return rows, total, requests


def fetch_search(search, refresh=False, product=PRODUCT):
    """Every row a search finds, into its cache file. Answers (rows, total, missing).

    A search read completely before isn't asked again unless refresh is set.
    """
    have = load_search(search, product)
    if have and have.get('complete') and not refresh:
        return len(have['rows']), have['total'], 0
    rows, total, requests = gather(search, product)
    missing = max(0, total - len(rows))
    out = {
        'search': search,
        'product': product,
        'url': URL % urllib.parse.urlencode({'product': product, 'search': search}),
        'fetched': datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ'),
        'total': total,
        'requests': requests,
        'missing': missing,
        'complete': missing == 0,
        'rows': sorted(rows.values(), key=lambda r: (-int(r['push_id']), -int(r['id']))),
    }
    os.makedirs(os.path.dirname(search_path(search, product)), exist_ok=True)
    tmp = search_path(search, product) + '.part'
    with open(tmp, 'w', encoding='utf-8') as f:
        json.dump(out, f)
    os.replace(tmp, search_path(search, product))
    return len(rows), total, missing


def searches_for(build, tables):
    """One search per table family: "70707 spell" covers SpellName and SpellEffect."""
    number = wago.version(build).split('+', 1)[0].rsplit('.', 1)[-1]
    wanted = sorted({'%s %s' % (number, t.lower()) for t in tables}, key=len)
    out = []
    for s in wanted:
        if not any(s.startswith(o) for o in out):
            out.append(s)
    return out


# Reading the cached rows.

def cached_rows(product=PRODUCT):
    """Every row of every cached search, once each (by wago.tools' row id)."""
    d = os.path.join(cache_dir(product), 'search')
    rows = {}
    if os.path.isdir(d):
        for name in sorted(os.listdir(d)):
            if name.endswith('.json'):
                with open(os.path.join(d, name), encoding='utf-8') as f:
                    for r in json.load(f).get('rows', []):
                        rows[str(r['id'])] = r
    return list(rows.values())


def covered(build, table, product=PRODUCT):
    """The cached search that covers a table at a build, or None."""
    want = searches_for(build, [table])[0]
    d = os.path.join(cache_dir(product), 'search')
    for name in sorted(os.listdir(d)) if os.path.isdir(d) else []:
        if name.endswith('.json'):
            s = load_search(name[:-5], product)
            if s and s.get('complete') and want.startswith(s['search']):
                return s['search']
    return None


def latest(build, region=REGION, product=PRODUCT):
    """{table: {record id: row}}, the newest push's row of each record, for one build and region."""
    number = int(wago.version(build).split('+', 1)[0].rsplit('.', 1)[-1])
    best = collections.defaultdict(dict)
    for r in cached_rows(product):
        if int(r['build']) != number or int(r['region_id']) != region or r.get('locale', 'enUS') != 'enUS':
            continue
        rec = str(r['record_id'])
        have = best[r['table_name']].get(rec)
        if have is None or (int(r['push_id']), int(r['id'])) > (int(have['push_id']), int(have['id'])):
            best[r['table_name']][rec] = r
    return best


def flatten(data):
    out = []
    for v in data:
        if isinstance(v, list):
            out.extend(flatten(v))
        else:
            out.append(v)
    return out


def cell(v):
    """A hotfix value the way wago.tools' CSV writes it."""
    if v is None:
        return ''
    if isinstance(v, bool):
        return '1' if v else '0'
    if isinstance(v, float):
        if v.is_integer() and abs(v) < 2 ** 53:
            return str(int(v))
        return '%.11g' % struct.unpack('<f', struct.pack('<f', v))[0]
    return str(v)


def header(table, build):
    with open(wago.path(table, base_version(wago.version(build))), newline='', encoding='utf-8') as f:
        return next(csv.reader(f), [])


def as_row(table, cols, r):
    """A hotfix row with data as {column: value}, or raises when its field count isn't the CSV's."""
    values = flatten(r['data'])
    if len(values) != len(cols):
        raise ValueError('%s %s (push %s): %d fields where the CSV has %d columns'
                         % (table, r['record_id'], r['push_id'], len(values), len(cols)))
    return dict(zip(cols, (cell(v) for v in values)))


def id_column(cols):
    return 'ID' if 'ID' in cols else None


def overlay(table, build, region=REGION, product=PRODUCT):
    """A table at a build with its hotfix rows applied: (header, rows, counts)."""
    base = base_version(wago.version(build))
    cols = header(table, base)
    rows = wago.rows(table, base) if cols else []
    counts = collections.Counter()
    hf = latest(base, region, product).get(table, {})
    if not hf:
        return cols, rows, counts
    if not cols:
        # wago.tools serves no rows (and so no header) for some tables, like
        # TraitNodeGroupXTraitNode at 70707. Without a header we can't name the fields.
        counts['no header'] = len(hf)
        return cols, rows, counts
    key = id_column(cols)
    if key is None:
        counts['no ID column'] = len(hf)
        return cols, rows, counts
    by_id = collections.OrderedDict((r[key], r) for r in rows)
    for rec, r in sorted(hf.items(), key=lambda kv: int(kv[0])):
        status = int(r['status'])
        if status == STATUS_DATA and r['data'] is not None:
            new = as_row(table, cols, r)
            new[key] = rec
            counts['changed' if rec in by_id else 'added'] += 1
            if rec in by_id and by_id[rec] == new:
                counts['changed'] -= 1
                counts['same'] += 1
            by_id[rec] = new
        elif status == STATUS_REMOVED:
            counts['removed' if by_id.pop(rec, None) is not None else 'removed, not in build'] += 1
        else:
            counts['status %d kept' % status] += 1
    return cols, list(by_id.values()), counts


def write_table(table, build, region=REGION, product=PRODUCT):
    """Write one table of the +hf version into wago.py's cache. Answers its counts."""
    cols, rows, counts = overlay(table, build, region, product)
    dest = os.path.join(wago.CACHE, derived(build), table + '.csv')
    os.makedirs(os.path.dirname(dest), exist_ok=True)
    if not counts:
        shutil.copyfile(wago.path(table, base_version(derived(build))), dest + '.part')
    else:
        with open(dest + '.part', 'w', newline='', encoding='utf-8') as f:
            w = csv.DictWriter(f, fieldnames=cols, lineterminator='\n')
            w.writeheader()
            w.writerows(rows)
    os.replace(dest + '.part', dest)
    return counts


def build_table(table, version):
    """wago.py's hook: make a missing table of a +hf version from the cached rows."""
    if not version.endswith(SUFFIX):
        raise OSError('%s is not a hotfix version' % version)
    base = base_version(version)
    if covered(base, table) is None:
        print('hotfixes: no cached search covers %s at %s, so %s is the build\'s own table '
              '(hotfixes.py fetch %s %s)' % (table, base, table, base, table), file=sys.stderr)
    write_table(table, base)


def check(table, build, region=REGION, product=PRODUCT):
    """Per column, how many hotfixed rows agree with the build's row of the same record."""
    base = base_version(wago.version(build))
    cols = header(table, base)
    hf = latest(base, region, product).get(table, {})
    if not cols or not hf:
        return None
    key = id_column(cols)
    shipped = {r[key]: r for r in wago.rows(table, base)} if key else {}
    agree, both, bad = collections.Counter(), 0, []
    for rec, r in hf.items():
        if int(r['status']) != STATUS_DATA or rec not in shipped:
            continue
        try:
            new = as_row(table, cols, r)
        except ValueError as e:
            bad.append(str(e))
            continue
        both += 1
        for c in cols:
            if new[c] == shipped[rec][c] or _same_float(new[c], shipped[rec][c]):
                agree[c] += 1
    return cols, both, agree, bad


def _same_float(a, b):
    try:
        return abs(float(a) - float(b)) <= 1e-6 * max(1.0, abs(float(b)))
    except ValueError:
        return False


def tables_for(args):
    tables = list(args.tables)
    for s in args.set or []:
        tables += wago.SETS[s]
    return sorted(set(tables))


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n\n')[0])
    sub = ap.add_subparsers(dest='cmd', required=True)
    for name in ('fetch', 'pushes', 'check', 'apply', 'rows'):
        p = sub.add_parser(name)
        p.add_argument('build')
        p.add_argument('tables', nargs='*')
        p.add_argument('--set', action='append', choices=sorted(wago.SETS))
        p.add_argument('--region', type=int, default=REGION)
        if name == 'fetch':
            p.add_argument('--search', action='append', default=[])
            p.add_argument('--refresh', action='store_true')
        if name == 'rows':
            p.add_argument('--record', action='append', default=[])
    a = ap.parse_args()
    build = base_version(wago.version(a.build))
    tables = tables_for(a)

    if a.cmd == 'fetch':
        searches = searches_for(build, tables) + a.search if tables else a.search
        if not searches:
            ap.error('fetch needs tables, a --set or a --search')
        short = 0
        for s in searches:
            got, total, missing = fetch_search(s, a.refresh)
            print('%s: %d of %d rows%s' % (s, got, total, ', %d missing' % missing if missing else ''))
            short += missing
        return 1 if short else 0

    best = latest(build, a.region)
    if a.cmd == 'pushes':
        pushes = collections.defaultdict(collections.Counter)
        when = {}
        for t, recs in best.items():
            if tables and t not in tables:
                continue
            for r in recs.values():
                pushes[int(r['push_id'])][t] += 1
                when[int(r['push_id'])] = r['created_at']
        for p in sorted(pushes):
            print(p, when[p], ', '.join('%s %d' % kv for kv in sorted(pushes[p].items())))
    elif a.cmd == 'rows':
        for t in tables:
            cols = header(t, build)
            for rec, r in sorted(best.get(t, {}).items(), key=lambda kv: int(kv[0])):
                if a.record and rec not in a.record:
                    continue
                shown = as_row(t, cols, r) if r['data'] is not None and cols else None
                print(t, rec, 'push', r['push_id'], 'status', r['status'], json.dumps(shown))
    elif a.cmd == 'check':
        for t in tables or sorted(best):
            res = check(t, build, a.region)
            if res is None:
                print('%s: nothing to compare' % t)
                continue
            cols, both, agree, bad = res
            print('%s: %d hotfixed rows the build also has' % (t, both))
            for c in cols:
                print('   %-40s %d/%d agree' % (c, agree[c], both))
            for b in bad[:5]:
                print('   ' + b)
    elif a.cmd == 'apply':
        for t in tables or sorted(best):
            counts = write_table(t, build, a.region)
            if counts:
                print('%s: %s' % (t, ', '.join('%s %d' % kv for kv in sorted(counts.items()) if kv[1])))
            if not covered(build, t):
                print('%s: no cached search covers it, copied as the build has it' % t)
        print(os.path.join(wago.CACHE, derived(build)))
    return 0


if __name__ == '__main__':
    sys.exit(main())
