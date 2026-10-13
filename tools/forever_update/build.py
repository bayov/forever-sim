#!/usr/bin/env python3
"""build.py: turn a day's report.json into the single-file Forever Update page.

    python3 tools/forever_update/build.py report.json -o forever-update.html [--cache DIR] [--offline] [--strict] [--standalone]

The output is an Artifact page (no doctype or <head>; the Artifact tool adds them when it publishes). Use --standalone
for a file to open locally.

It validates report.json (see README.md for the format), fetches the Wowhead Forever tooltip of every item and spell the
report names (items.list[].id, and every "ref": "i:<id>" / "s:<id>"), downloads their icons and the class icons from
wow.zamimg.com, inlines all of it as data, and writes template.html with the data filled in. The page needs no network.

Problems that would make a wrong or empty page (an unknown status or kind, a missing id, an item with no tooltip and no
stats of its own) are listed at the end; with --strict they fail the build.
"""
import argparse, base64, html, json, os, re, subprocess, sys
from html.parser import HTMLParser

HERE = os.path.dirname(os.path.abspath(__file__))
TOOLTIP_URL = 'https://nether.wowhead.com/forever/tooltip/{kind}/{id}?dataEnv=17&locale=0'
ICON_URL = 'https://wow.zamimg.com/images/wow/icons/large/{name}.jpg'
CLASSES = ['Shaman', 'Paladin', 'Rogue', 'Warrior', 'Hunter', 'Mage', 'Priest', 'Warlock', 'Druid']
STATUSES = {'official', 'hotfix', 'blue', 'interview', 'client', 'datamined', 'tracker', 'tooltip', 'unclear'}
KINDS = {'buff', 'nerf', 'change', 'fix', 'new'}
PR_STATES = {'open', 'draft', 'none', 'blocked'}


# ---------- Wowhead tooltip HTML -> lines (same parser as tools/dps_report/tooltips.py) ----------
# lines are [style, text] or ['2', left, right]: t plain, m muted, y gold, g green, 2 two columns.
SKIP_CLASSES = ('whtt-sellprice', 'whtt-extra', 'wowhead-tooltip-requirements', 'whtt-name', 'wowhead-tooltip-item-classes', 'indent', 'whtt-ilvl')
COLOR = {'q': 'y', 'q0': 'm', 'q1': 't', 'q2': 'g', 'q3': 't', 'q4': 't', 'q5': 't'}
VOID = {'br', 'img', 'hr', 'input', 'meta', 'link'}


class _TT(HTMLParser):
    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.lines, self.cur, self.stack, self.tables = [], [], [], []

    def _skip(self):
        return any(s for _, s, _ in self.stack)

    def _color(self):
        for _, _, c in reversed(self.stack):
            if c:
                return c
        return None

    def flush(self):
        txt = re.sub(r'\s+', ' ', ''.join(t for t, _ in self.cur)).strip()
        col = next((c for t, c in self.cur if t.strip() and c), None)
        if txt:
            self.lines.append([col or 't', txt])
        self.cur = []

    def handle_starttag(self, tag, attrs):
        a = dict(attrs)
        cls = (a.get('class') or '').split()
        if tag == 'br':
            if self.tables and self.tables[-1][0] and self.tables[-1][3]:
                self.tables[-1][3].append((' ', None))
            else:
                self.flush()
            return
        if tag in VOID:
            return
        if tag in ('div', 'p', 'table', 'tr'):
            self.flush()
        if tag == 'table':
            self.tables.append([a.get('width') == '100%', [], [], None])
        if tag in ('td', 'th') and self.tables and self.tables[-1][0]:
            t = self.tables[-1]
            t[3] = t[1] if tag == 'td' else t[2]
        skip = any(c in SKIP_CLASSES for c in cls)
        col = next((COLOR[c] for c in cls if c in COLOR), None)
        self.stack.append((tag, skip, col))

    def handle_startendtag(self, tag, attrs):
        self.handle_starttag(tag, attrs)

    def handle_endtag(self, tag):
        if tag in VOID:
            return
        if tag == 'tr' and self.tables and self.tables[-1][0]:
            t = self.tables[-1]
            left = re.sub(r'\s+', ' ', ''.join(x for x, _ in t[1])).strip()
            right = re.sub(r'\s+', ' ', ''.join(x for x, _ in t[2])).strip()
            col = next((c for x, c in t[1] + t[2] if x.strip() and c), None)
            if left or right:
                self.lines.append(['2', left, right] if left and right else [col or 't', left or right])
            t[1], t[2], t[3] = [], [], None
        if tag in ('div', 'p', 'tr') and not (self.tables and self.tables[-1][0] and tag != 'tr'):
            self.flush()
        if tag == 'table' and self.tables:
            self.tables.pop()
            self.flush()
        for i in range(len(self.stack) - 1, -1, -1):
            if self.stack[i][0] == tag:
                del self.stack[i:]
                break

    def handle_data(self, data):
        if self._skip():
            return
        seg = (data, self._color())
        if self.tables and self.tables[-1][0] and self.tables[-1][3] is not None:
            self.tables[-1][3].append(seg)
        else:
            self.cur.append(seg)

    def close(self):
        super().close()
        self.flush()


def tidy_formula(s):
    """Forever wowhead leaves some numbers as formulas: "[48 + (151 * (1 * 1))]" -> "199", "(246 * 1)" -> "246"."""
    if '[' not in s and '*' not in s:
        return s
    prev = None
    while prev != s:
        prev = s
        s = re.sub(r'\(\s*(\d+(?:\.\d+)?)\s*\)', r'\1', s)
        s = re.sub(r'\s*\*\s*1(?![\d.])', '', s)

    def ev(m):
        body = m.group(1)
        if re.fullmatch(r'[\d\s+\-*/().]+', body):
            try:
                v = eval(compile(body, '<f>', 'eval'), {'__builtins__': {}}, {})
                return str(int(v)) if float(v).is_integer() else ('%.1f' % v).rstrip('0').rstrip('.')
            except Exception:
                pass
        return body
    return re.sub(r'\[([^\[\]]*)\]', ev, s)


def html_lines(h):
    p = _TT()
    p.feed(h or '')
    p.close()
    return [[l[0]] + [tidy_formula(x) for x in l[1:]] for l in p.lines]


ITEM_DROP = re.compile(r'^(Sell Price.*|Max Stack.*|Dropped by.*|Drop Chance.*|Durability .*|<.*>)$')
SPELL_DROP = re.compile(r'^(Requires .*|Tools:|Reagents:.*)$')


def item_entry(j):
    title = j.get('name') or ''
    lines = []
    for ln in html_lines(j.get('tooltip')):
        if ln[0] != '2' and (ITEM_DROP.match(ln[1]) or ln[1] == title):
            continue
        if ln[0] != '2' and ln[1].startswith('Item Level'):
            ln = ['y', ln[1]]
        lines.append(ln)
    return {'name': title, 'quality': j.get('quality'), 'icon': j.get('icon'), 'lines': lines}


def spell_entry(j):
    title = j.get('name') or ''
    lines, sub = [], None
    for ln in html_lines(j.get('tooltip')):
        if ln[0] != '2':
            if ln[1] == title or SPELL_DROP.match(ln[1]):
                continue
            m = re.fullmatch(r'(Rank \d+|Talent|Racial|Passive)', ln[1])
            if m and not sub:
                sub = m.group(1)
                continue
        lines.append(ln)
    return {'name': title, 'quality': None, 'icon': j.get('icon'), 'lines': lines, 'sub': sub}


# ---------- fetching ----------
def curl(url, binary=False):
    r = subprocess.run(['curl', '-sS', '-f', '-L', '--max-time', '30', '-A', 'Mozilla/5.0', url], capture_output=True)
    if r.returncode != 0:
        raise RuntimeError(f'{url}: {r.stderr.decode(errors="replace").strip() or "curl exit " + str(r.returncode)}')
    return r.stdout if binary else r.stdout.decode('utf-8')


def cached(cache, name, fetch):
    path = os.path.join(cache, name) if cache else None
    if path and os.path.exists(path):
        with open(path, 'rb') as f:
            return f.read()
    data = fetch()
    if path:
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with open(path, 'wb') as f:
            f.write(data)
    return data


def fetch_tooltip(kind, id_, cache, offline, problems):
    try:
        raw = cached(cache, f'{kind}/{id_}.json', (lambda: (_ for _ in ()).throw(RuntimeError('offline'))) if offline else
                     (lambda: curl(TOOLTIP_URL.format(kind=kind, id=id_), binary=True)))
        j = json.loads(raw)
    except Exception as e:
        problems.append(f'no Wowhead Forever tooltip for {kind} {id_} ({e})')
        return None
    if not j.get('name') or j.get('error'):
        problems.append(f'Wowhead Forever has no {kind} {id_} yet')
        return None
    return item_entry(j) if kind == 'item' else spell_entry(j)


def fetch_icon(name, cache, offline, problems):
    name = re.sub(r'[^a-z0-9_\-]', '', str(name).lower())
    if not name:
        return None
    try:
        data = cached(cache, f'icons/{name}.jpg', (lambda: (_ for _ in ()).throw(RuntimeError('offline'))) if offline else
                      (lambda: curl(ICON_URL.format(name=name), binary=True)))
    except Exception as e:
        problems.append(f'icon {name}: {e}')
        return None
    if not data.startswith(b'\xff\xd8'):
        problems.append(f'icon {name}: not a jpg')
        return None
    return 'data:image/jpeg;base64,' + base64.b64encode(data).decode()


# ---------- validation ----------
def walk(x, path=''):
    if isinstance(x, dict):
        yield path, x
        for k, v in x.items():
            yield from walk(v, f'{path}.{k}' if path else k)
    elif isinstance(x, list):
        for i, v in enumerate(x):
            yield from walk(v, f'{path}[{i}]')


def validate(r, problems):
    for key in ('date',):
        if not r.get(key):
            problems.append(f'report.{key} is missing')
    for path, d in walk(r):
        if 'status' in d and d['status'] not in STATUSES and not path.startswith('sim.pr'):
            problems.append(f'{path}.status "{d["status"]}" is not one of {sorted(STATUSES)}')
        if 'kind' in d and d['kind'] not in KINDS:
            problems.append(f'{path}.kind "{d["kind"]}" is not one of {sorted(KINDS)}')
        if 'ref' in d and d['ref'] and not re.fullmatch(r'[is]:\d+', str(d['ref'])):
            problems.append(f'{path}.ref "{d["ref"]}" should be "i:<item id>" or "s:<spell id>"')
    for grp in ('classes', 'other'):
        for c in (r.get(grp) or {}):
            if c not in CLASSES:
                problems.append(f'{grp}.{c} is not a class name ({", ".join(CLASSES)})')
    for i, it in enumerate((r.get('items') or {}).get('list') or []):
        if not isinstance(it.get('id'), int):
            problems.append(f'items.list[{i}].id must be the numeric item id')
        if not it.get('who'):
            problems.append(f'items.list[{i}] ({it.get("id")}) has no "who"')
    st = ((r.get('sim') or {}).get('pr') or {}).get('state')
    if st and st not in PR_STATES:
        problems.append(f'sim.pr.state "{st}" is not one of {sorted(PR_STATES)}')


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument('report')
    ap.add_argument('-o', '--out', required=True)
    ap.add_argument('--cache', default=os.path.join(os.path.expanduser('~'), '.cache', 'forever_update'), help='tooltip and icon cache')
    ap.add_argument('--offline', action='store_true', help='use only the cache')
    ap.add_argument('--strict', action='store_true', help='fail on any problem')
    ap.add_argument('--standalone', action='store_true', help='add the doctype and head that the Artifact tool adds when it publishes, for viewing the file locally')
    a = ap.parse_args()

    with open(a.report) as f:
        r = json.load(f)
    problems = []
    validate(r, problems)

    refs = set()
    for it in (r.get('items') or {}).get('list') or []:
        if isinstance(it.get('id'), int):
            refs.add(f'i:{it["id"]}')
    for _, d in walk(r):
        if d.get('ref') and re.fullmatch(r'[is]:\d+', str(d['ref'])):
            refs.add(d['ref'])

    tooltips = {}
    for ref in sorted(refs):
        kind, id_ = ref.split(':')
        t = fetch_tooltip('item' if kind == 'i' else 'spell', int(id_), a.cache, a.offline, problems)
        if t:
            tooltips[ref] = t
    for it in (r.get('items') or {}).get('list') or []:
        if f'i:{it.get("id")}' not in tooltips and not it.get('lines'):
            problems.append(f'item {it.get("id")} has no Wowhead tooltip; give it "name", "quality", "icon" and "lines" in report.json')

    icon_names = {t['icon'] for t in tooltips.values() if t.get('icon')}
    icon_names |= {'classicon_' + c.lower() for c in CLASSES}
    for _, d in walk(r):
        if isinstance(d.get('icon'), str):
            icon_names.add(d['icon'])
    icons = {}
    for n in sorted(icon_names):
        u = fetch_icon(n, a.cache, a.offline, problems)
        if u:
            icons[n] = u

    data = json.dumps({'report': r, 'tooltips': tooltips, 'icons': icons}, ensure_ascii=False, separators=(',', ':'))
    data = data.replace('<', '\\u003c')  # nothing in the data can end or confuse the <script> block
    with open(os.path.join(HERE, 'template.html')) as f:
        page = f.read()
    title = r.get('pageTitle') or 'Forever Update'
    page = page.replace('__TITLE__', html.escape(title), 1).replace('__DATA__', data, 1)
    if a.standalone:
        page = '<!doctype html><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover">\n' + page
    with open(a.out, 'w') as f:
        f.write(page)

    print(f'wrote {a.out}: {len(page) / 1024:.0f} KB, {len(tooltips)} tooltips, {len(icons)} icons')
    if problems:
        print(f'{len(problems)} problem(s):', file=sys.stderr)
        for p in problems:
            print('  - ' + p, file=sys.stderr)
        if a.strict:
            sys.exit(1)


if __name__ == '__main__':
    main()
