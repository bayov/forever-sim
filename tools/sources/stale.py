#!/usr/bin/env python3
"""Say which of forever-sim's inputs are behind their source, and what to run.

    python3 tools/sources/stale.py [--client-build 1.60.1.70450] [--addons <client's Interface/AddOns>] [--json]
    python3 tools/sources/stale.py --registry

It only reads, and it never goes online (gh aside). docs/data_sources.md is
the registry of every source it checks.

- The newest client build we know of, against the build the DPS report was
  audited against (`foreverClientBuild` in .agents/skills/dps-report/state.json).
  We know a build from --client-build (bayovui's stale.py passes the installed
  beta's) and from the daily update routine's open TASKS.md items tagged
  [source: wago build N] (or, before the tags, worded "Pull wago.tools DB2
  data for build N"). A newer one means forever-sim has catching up
  to do, and the weights and BiS lists bayovui takes from us with it.
- The inputs read out of the client's tables (CLIENT_INPUTS), against the
  newest client build we know. Each names its build in its meta block, and a
  newer client means running its script again.
- The items Forever still hides (assets/database/hidden_items.json, from
  tools/sources/hidden.py): how many, and with --addons, which of them our
  client has since been sent (its hotfix cache, read with
  tools/sources/dbcache.py). Those were revealed, so their stats are worth
  fetching.
- The talent trees' build per class: the newest build tools/forever_talents/
  client/<class>.json names. Only shown, since a build that changed no talent
  doesn't touch the file.
- How old each scrape is, from its `meta` block (tools/sources/meta.py), not
  the file's modification time, which a checkout resets.
- The daily update routine's open `forever-update/<date>` PRs (with gh).
- With --addons, whether the installed AtlasLoot Forever changed since we read
  it: the hash of its four data files against assets/db_inputs/atlasloot_db.meta.json.
  bayovui's stale.py passes the beta client's AddOns folder.

--json prints {"lines": [...], "todo": [...]} for bayovui's stale.py, which
runs this first and prints both.

--registry checks that docs/data_sources.md names every file in
assets/db_inputs/ and every script that downloads something, and exits 1 when
it doesn't. `make test` runs it, so the registry can't fall behind the code.
"""
import argparse
import glob
import hashlib
import json
import os
import re
import subprocess
import sys
import datetime

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..'))
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from meta import read as read_meta  # noqa: E402

REGISTRY = os.path.join(ROOT, 'docs', 'data_sources.md')
INPUTS = os.path.join(ROOT, 'assets', 'db_inputs')
STATE = os.path.join(ROOT, '.agents', 'skills', 'dps-report', 'state.json')
TALENTS = os.path.join(ROOT, 'tools', 'forever_talents', 'client')

# Each scrape, how many days old it may get, and what refreshes it.
SCRAPES = [
    ('forever_wowhead_items.json', 7, 'the scrape-wowhead skill (items)'),
    ('forever_wowhead_recipes.json', 7, 'the scrape-wowhead skill (recipes)'),
    ('foreverchanges_loot.json', 14, 'python3 tools/database/foreverchanges/scrape.py'),
    ('foreverchanges_suffixes.json', 14, 'python3 tools/database/foreverchanges/suffixes.py <max level> <cache dir>'),
]

# The inputs we read out of the client's tables, with the build in their meta
# block, and the script that reads them again.
CLIENT_INPUTS = [
    ('assets/db_inputs/forever_client_items.meta.json', 'python3 tools/sources/itemsparse.py %s'),
    ('assets/db_inputs/forever_enchants.json', 'python3 tools/sources/enchants.py %s'),
    ('ui/core/forever/racials.meta.json', 'python3 tools/sources/racials.py %s --write'),
    ('assets/db_inputs/forever_client_cache.json', 'python3 tools/sources/dbcache.py <beta client> --write (%s)'),
]

HIDDEN = os.path.join(ROOT, 'assets', 'database', 'hidden_items.json')
CLIENT_CACHE = os.path.join(INPUTS, 'forever_client_cache.json')

# The AtlasLoot files gen_db -gen=atlasloot reads, in the order it hashes them
# (AtlasLootFiles in tools/database/atlasloot.go).
ATLASLOOT_META = os.path.join(INPUTS, 'atlasloot_db.meta.json')
ATLASLOOT_FILES = [
    ('AtlasLootClassic_Data', 'source.lua'),
    ('AtlasLootClassic_DungeonsAndRaids', 'data.lua'),
    ('AtlasLootClassic_PvP', 'data.lua'),
    ('AtlasLootClassic_Factions', 'data.lua'),
]

# Where scripts that download something live, and how we spot one.
FETCHER_DIRS = ['tools', 'orchestration']
FETCHER_RE = re.compile(
    r'\burllib\.request\b|^\s*import requests\b|\bcurl_cffi\b|\bhttp\.client\b'  # Python
    r'|\bhttp\.(?:Get|NewRequest|Client)\b'  # Go
    r'|^[^#\n]*\b(?:curl|wget)\s+-'  # shell, or a curl command line in Python
    r'|[\'"]curl[\'"]'  # curl run from Python, as subprocess.run(['curl', ...])
    , re.M)


def build_no(build):
    """The build number of a version like 1.60.1.70425, or of a bare 70425.

    state.json sometimes adds a note after it, like "70450 (spells ... identical to 70425)".
    """
    return int(re.match(r'(?:[\d.]*\.)?(\d+)', str(build).strip()).group(1))


def audited():
    if not os.path.exists(STATE):
        return None
    with open(STATE, encoding='utf-8') as f:
        return json.load(f).get('foreverClientBuild')


def build_tasks():
    """The builds of the open client build items in TASKS.md."""
    path = os.path.join(ROOT, 'TASKS.md')
    if not os.path.exists(path):
        return []
    with open(path, encoding='utf-8') as f:
        text = f.read()
    tagged = re.findall(r'^- \[ \] .*\[source: wago build ([\d.]+)\]', text, re.M)
    worded = re.findall(r'^- \[ \] .*?Pull wago\.tools DB2 data for build ([\d.]+)', text, re.M)
    return list(dict.fromkeys(tagged + worded))


def talent_builds():
    """{class: the newest build its client talent file names}."""
    out = {}
    for path in sorted(glob.glob(os.path.join(TALENTS, '*.json'))):
        with open(path, encoding='utf-8') as f:
            builds = re.findall(r'1\.60\.\d+\.(\d+)', f.read())
        out[os.path.splitext(os.path.basename(path))[0]] = max(map(int, builds)) if builds else None
    return out


def repo():
    """forever-sim's GitHub repo as owner/name, from its origin remote."""
    r = subprocess.run(['jj', 'git', 'remote', 'list', '--ignore-working-copy'],
                       capture_output=True, text=True, cwd=ROOT)
    if r.returncode != 0:
        r = subprocess.run(['git', 'remote', '-v'], capture_output=True, text=True, cwd=ROOT)
    m = re.search(r'^origin\s+\S*github\.com[:/]([^/\s]+/[^/\s]+?)(?:\.git)?(?:\s|$)', r.stdout, re.M)
    return m.group(1) if m else None


def routine_prs(name):
    """The daily update routine's open PRs, as (number, title). None when gh can't answer."""
    try:
        r = subprocess.run(['gh', 'api', 'repos/%s/pulls?state=open' % name, '--jq',
                            '.[] | select(.head.ref | startswith("forever-update/"))'
                            ' | [.number, .title] | @tsv'],
                           capture_output=True, text=True)
    except FileNotFoundError:
        return None
    if r.returncode != 0:
        return None
    return [tuple(ln.split('\t', 1)) for ln in r.stdout.splitlines() if ln.strip()]


def atlasloot_hash(addons):
    """The installed AtlasLoot's hash, as gen_db writes it, or None when it isn't installed."""
    h = hashlib.sha256()
    for addon, name in ATLASLOOT_FILES:
        path = os.path.join(addons, addon, name)
        if not os.path.exists(path):
            return None
        with open(path, 'rb') as f:
            h.update(f.read())
    return h.hexdigest()[:12]


def check(client_build, addons=None):
    lines, todo = [], []
    say = lines.append

    prs = routine_prs(repo() or '')
    if prs is None:
        say('  ?     the daily update routine\'s PRs, gh could not list them')
    for number, title in prs or []:
        say('  OPEN  PR #%s %s' % (number, title))
        todo.append('review PR #%s from the daily update, then merge or close it (the user decides)'
                    % number)

    done = audited()
    newest = [done] if done else []
    if client_build:
        newest.append(client_build)
    for build in build_tasks():
        if not done or build_no(build) > build_no(done):
            say('  NEW   build %s is in TASKS.md, not processed yet' % build)
            newest.append(build)
    if done:
        latest = max(newest, key=build_no)
        ok = build_no(latest) <= build_no(done)
        say('  %s audited against client build %s' % ('ok   ' if ok else 'STALE', done))
        if not ok:
            todo.append('catch up to build %s (the dps-report skill\'s "full pass", which is '
                        'hours of agents, so ask)' % latest)
    else:
        say('  ?     no audited client build in %s' % os.path.relpath(STATE, ROOT))

    if len(newest) > 0:
        latest = max(newest, key=build_no)
        for rel, refresh in CLIENT_INPUTS:
            path = os.path.join(ROOT, rel)
            m = read_meta(path) if os.path.exists(path) else None
            if not m or 'build' not in m:
                say('  ?     %s has no client build in its meta block' % rel)
                todo.append(refresh % latest)
            elif build_no(m['build']) < build_no(latest):
                say('  OLD   %s is from client build %s, %s is out' % (rel, m['build'], latest))
                todo.append(refresh % latest)
            else:
                say('  ok    %s, client build %s' % (rel, m['build']))

    hidden_lines, hidden_todo = check_hidden(addons)
    lines += hidden_lines
    todo += hidden_todo

    talents = talent_builds()
    if talents:
        say('  info  talent trees by class, newest build each names: %s' % ', '.join(
            '%s %s' % (c, b or '?') for c, b in talents.items()))

    if addons:
        rel = os.path.relpath(ATLASLOOT_META, ROOT)
        m = read_meta(ATLASLOOT_META) if os.path.exists(ATLASLOOT_META) else None
        installed = atlasloot_hash(addons)
        if installed is None:
            say('  ?     AtlasLoot Forever isn\'t installed in %s' % addons)
        elif not m:
            say('  ?     %s is missing, so we don\'t know which AtlasLoot we read' % rel)
            todo.append('the update-sources skill\'s AtlasLoot step')
        elif m.get('sha256') != installed:
            say('  NEW   AtlasLoot Forever changed since %s (%s)' % (m.get('fetched'), m.get('version')))
            todo.append('the update-sources skill\'s AtlasLoot step')
        else:
            say('  ok    AtlasLoot Forever, read %s (%s)' % (m.get('fetched'), m.get('version')))

    now = datetime.datetime.now(datetime.timezone.utc)
    for name, days, refresh in SCRAPES:
        path = os.path.join(INPUTS, name)
        rel = os.path.relpath(path, ROOT)
        if not os.path.exists(path):
            say('  ?     %s, never scraped' % rel)
            todo.append(refresh)
            continue
        m = read_meta(path)
        if not m or 'fetched' not in m:
            say('  ?     %s has no meta block, so no fetch date' % rel)
            todo.append(refresh)
            continue
        fetched = datetime.datetime.strptime(m['fetched'], '%Y-%m-%dT%H:%MZ').replace(
            tzinfo=datetime.timezone.utc)
        age = (now - fetched).days
        ok = age <= days
        say('  %s %s, fetched %s (%d days ago)' % ('ok   ' if ok else 'OLD  ', rel, m['fetched'], age))
        if not ok:
            todo.append(refresh)
    return lines, todo


def check_hidden(addons):
    """Lines and to-dos about the hidden items, and the reveals in our client."""
    lines, todo = [], []
    rel = os.path.relpath(HIDDEN, ROOT)
    if not os.path.exists(HIDDEN):
        return ['  ?     %s is missing' % rel], ['python3 tools/sources/hidden.py']
    with open(HIDDEN, encoding='utf-8') as f:
        hidden = json.load(f)
    c = hidden['counts']
    lines.append('  info  %d items still hidden, %d of them gear: %d in the sim with Classic stats, '
                 '%d partly, %d with Wowhead\'s unconfirmed stats, %d of Forever\'s not in it, '
                 'and %d pieces of gear revealed (%s)' % (
                     c['hidden'], c['hiddenGear'], c['inSimWithClassicStats'], c['inSimPartly'],
                     c.get('inSimFromWowhead', 0), c['foreverNotInSim'], c.get('revealedGear', 0), rel))
    if not addons:
        return lines, todo
    client = os.path.dirname(os.path.dirname(os.path.abspath(addons)))
    try:
        import dbcache
        _build, rows = dbcache.entries(os.path.join(client, dbcache.CACHE))
    except (OSError, ValueError) as e:
        lines.append('  ?     our client\'s hotfix cache: %s' % e)
        return lines, todo
    sent = {rid for table, rid, status, data, _push in rows
            if table == dbcache.ITEM_SPARSE_HASH and status == dbcache.VALID and data}
    revealed = [i for i, e in hidden['items'].items() if e['state'] == 'hidden' and int(i) in sent]
    if revealed:
        names = [hidden['items'][i].get('name') or i for i in revealed]
        lines.append('  NEW   our client was sent %d items hidden_items.json calls hidden: %s%s' % (
            len(revealed), ', '.join(names[:8]), ' ...' if len(names) > 8 else ''))
        todo.append('python3 tools/sources/dbcache.py "%s" --write, then python3 tools/sources/hidden.py'
                    % client)
        todo.append('the scrape-wowhead skill, for the revealed items\' stats')
    else:
        lines.append('  ok    no hidden item revealed in our client since %s' % hidden['meta']['fetched'])
    return lines, todo


def fetchers():
    """Repo paths of the scripts that download something."""
    out = []
    for top in FETCHER_DIRS:
        for path in glob.glob(os.path.join(ROOT, top, '**', '*'), recursive=True):
            if not path.endswith(('.py', '.go', '.sh')) or '/node_modules/' in path:
                continue
            if os.path.abspath(path) == os.path.abspath(__file__):
                continue
            with open(path, encoding='utf-8', errors='replace') as f:
                if FETCHER_RE.search(f.read()):
                    out.append(os.path.relpath(path, ROOT))
    return sorted(out)


def unlisted():
    """The inputs and fetchers docs/data_sources.md doesn't name."""
    doc = ''
    if os.path.exists(REGISTRY):
        with open(REGISTRY, encoding='utf-8') as f:
            doc = f.read()
    missing = []
    for path in sorted(glob.glob(os.path.join(INPUTS, '**', '*'), recursive=True)):
        if os.path.isfile(path):
            rel = os.path.relpath(path, INPUTS)
            if rel not in doc:
                missing.append(os.path.relpath(path, ROOT))
    for rel in fetchers():
        if rel not in doc:
            missing.append(rel)
    return missing


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n\n')[0])
    ap.add_argument('--client-build', help='the installed client\'s version, like 1.60.1.70450')
    ap.add_argument('--addons', help='the beta client\'s Interface/AddOns, to check AtlasLoot')
    ap.add_argument('--json', action='store_true')
    ap.add_argument('--registry', action='store_true',
                    help='only check that docs/data_sources.md names every input and fetcher')
    a = ap.parse_args()

    missing = unlisted()
    if a.registry:
        for m in missing:
            print('%s is missing from docs/data_sources.md' % m)
        return 1 if missing else 0

    lines, todo = check(a.client_build, a.addons)
    if missing:
        lines.append('  MISSING from docs/data_sources.md: %s' % ', '.join(missing))
        todo.append('add %s to docs/data_sources.md' % ', '.join(missing))
    if a.json:
        print(json.dumps({'lines': lines, 'todo': todo}))
        return 0
    print('\n'.join(lines))
    print()
    if todo:
        print('To do, in this order:')
        for t in dict.fromkeys(todo):
            print('  - ' + t)
    else:
        print('Nothing is behind.')
    return 0


if __name__ == '__main__':
    sys.exit(main())
