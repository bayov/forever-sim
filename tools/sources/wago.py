#!/usr/bin/env python3
"""The client's DB2 tables from wago.tools, through one cache on disk.

    python3 tools/sources/wago.py latest [--probe N]
    python3 tools/sources/wago.py builds
    python3 tools/sources/wago.py fetch <build> [Table ...] [--set NAME]
    python3 tools/sources/wago.py dir <build>
    python3 tools/sources/wago.py path <build> <Table>
    python3 tools/sources/wago.py diff <old> <new> [Table ...] [--set NAME] [--by-spell] [--show N]
    python3 tools/sources/wago.py cached

wago.tools serves every table of every client build as CSV, at
https://wago.tools/db2/<Table>/csv?build=<version>. Every script in this repo
that reads a client table goes through here: the audit loop's passes, the
client diff, gen_db's inputs and logcheck (in Go, wago.go, the same layout).

We keep each table once per build, at ~/.cache/wow-data/wago/<version>/<Table>.csv.
bayovui's build scripts and logcheck read and write the same folder, so a
patch downloads each table once for both repos. WOW_DATA_CACHE moves the
whole cache (tools/audit/setup.sh points it into $AUDIT_WORK, since a cloud
session's home doesn't last).

A build is a full version like 1.60.1.70450 or 1.15.9.69722 (the Era client
the mechanics pass compares with). A bare build number like 70450 means the
cached version that ends in it, or 1.60.1.<n> when none is cached.

`latest` asks wago.tools' build list for the newest Forever (1.60) build. That
list has lagged behind the tables before, so --probe N also asks for SpellName
at each of the next N build numbers, 20 at a time, and takes the newest that
answers. It prints "unknown" and exits 2 when wago.tools can't be reached.

`fetch` downloads the tables of a build that aren't cached yet and prints the
build's folder. With no tables it takes a named set (--set), and with neither
it fetches every set. Each set is the tables one reader needs, in SETS below.

`diff` compares two builds row by row, by ID: per table the rows added,
removed and changed, then the changed rows field by field. --by-spell prints
one line per changed spell instead, across the spell tables (what the
mechanics pass reads). tools/sources/db2diff.py writes the longer report
grouped by class that the client_update pass keeps.

Run anything that reads the CSVs with python3 -I. They're downloaded,
untrusted data.
"""
import argparse
import concurrent.futures
import csv
import json
import os
import re
import sys
import time
import urllib.error
import urllib.request

ROOT = os.environ.get('WOW_DATA_CACHE') or os.path.join(os.path.expanduser('~'), '.cache', 'wow-data')
CACHE = os.path.join(ROOT, 'wago')
URL = 'https://wago.tools/db2/%s/csv?build=%s'
BUILDS = 'https://wago.tools/api/builds'
AGENT = 'forever-sim tools/sources/wago.py'
FOREVER = '1.60.1'

csv.field_size_limit(10 ** 9)

# The tables each reader needs. A reader that needs one more table passes it
# by name, or adds it here.
SETS = {
    # The client_update pass and the update ledger's balance diff (db2diff.py):
    # the ledger's "How it was checked" list, plus ItemXItemEffect,
    # SpellXSpellVisual, SpellRadius and SpellRange. wago.tools doesn't serve
    # SpellProcsPerMinuteMod or SpellScaling for 1.60 builds.
    'client_update': [
        'Spell', 'SpellName', 'SpellEffect', 'SpellAuraOptions', 'SpellCooldowns',
        'SpellCategories', 'SpellCategory', 'SpellEquippedItems', 'SpellLevels', 'SpellMisc',
        'SpellPower', 'SpellClassOptions', 'SpellDuration', 'SpellCastTimes',
        'SpellProcsPerMinute', 'SpellAuraRestrictions',
        'SpellTargetRestrictions', 'SpellInterrupts', 'SpellShapeshift', 'SkillLineAbility',
        'TraitDefinition', 'TraitDefinitionEffectPoints', 'TraitNode', 'TraitNodeEntry',
        'TraitNodeXTraitNodeEntry', 'TraitCond', 'TraitEdge', 'TraitTree', 'TraitNodeGroup',
        'ItemSparse', 'ItemEffect', 'ItemXItemEffect', 'SpellItemEnchantment', 'ChrRaces',
        'SpellLabel', 'SpellXSpellVisual', 'SpellRadius', 'SpellRange', 'SpellReagents',
    ],
    'consumables': [
        'Item', 'ItemSparse', 'ItemEffect', 'ItemXItemEffect', 'SpellEffect', 'SpellName',
        'Spell', 'SpellMisc', 'SpellLevels', 'SpellItemEnchantment', 'SpellCategories',
        'SpellCategory', 'SpellDuration', 'SpellAuraOptions',
    ],
    'mechanics': [
        'SpellEffect', 'SpellName', 'SpellMisc', 'SpellAuraOptions', 'SpellCooldowns',
        'SpellDuration', 'SpellLevels', 'SpellPower', 'SpellCastTimes', 'SpellCategories',
        'SpellClassOptions', 'Spell', 'SkillLineAbility', 'ItemEffect', 'ItemXItemEffect',
        'ItemSparse', 'SpellItemEnchantment', 'SpellShapeshiftForm', 'SpellShapeshift',
        'ChrRaces', 'CharBaseInfo', 'TraitDefinition', 'TraitDefinitionEffectPoints',
        'TraitNodeEntry', 'CurvePoint',
    ],
    'rank_tuning': [
        'CurvePoint', 'SkillLineAbility', 'SpellCastTimes', 'SpellCooldowns', 'SpellDuration',
        'SpellEffect', 'SpellItemEnchantment', 'SpellLevels', 'SpellMisc', 'SpellName',
        'SpellPower', 'TraitCond', 'TraitDefinition', 'TraitDefinitionEffectPoints', 'TraitNode',
        'TraitNodeEntry', 'TraitNodeGroupXTraitCond', 'TraitNodeGroupXTraitNode',
        'TraitNodeXTraitNodeEntry',
    ],
    # tools/sources/spelltext.py (spell tooltips, cost lines, icons), and the
    # talent tree importer tools/forever_talents/import_db2.py on top of it.
    'spelltext': [
        'Spell', 'SpellName', 'SpellEffect', 'SpellMisc', 'SpellLevels', 'SpellDuration', 'SpellRange',
        'SpellRadius', 'SpellCastTimes', 'SpellCooldowns', 'SpellPower', 'SpellAuraOptions',
        'SpellTargetRestrictions', 'SpellXDescriptionVariables', 'SpellDescriptionVariables',
        'SpellShapeshift', 'SpellShapeshiftForm', 'SpellEquippedItems', 'ItemSubClass',
        'ManifestInterfaceData',
    ],
    'talents': [
        'TraitTree', 'TraitNode', 'TraitNodeEntry', 'TraitNodeXTraitNodeEntry', 'TraitDefinition',
        'TraitDefinitionEffectPoints', 'TraitEdge', 'CurvePoint', 'TalentTab', 'ChrClasses',
        'SpellClassOptions', 'SkillLine', 'SkillLineAbility',
    ],
}

# The spell tables --by-spell reads, keyed by their SpellID column.
SPELL_TABLES = ['SpellEffect', 'SpellMisc', 'SpellAuraOptions', 'SpellCooldowns', 'SpellLevels',
                'SpellCategories']


def key(version):
    return tuple(int(x) for x in version.split('.'))


def cached():
    """The versions we have tables of, oldest first."""
    if not os.path.isdir(CACHE):
        return []
    return sorted((v for v in os.listdir(CACHE) if re.fullmatch(r'\d+(\.\d+){3}', v)), key=key)


def version(build):
    """The full version for a build: 70450 becomes 1.60.1.70450."""
    build = str(build).strip()
    if '.' in build:
        return build
    for v in reversed(cached()):
        if v.endswith('.' + build):
            return v
    return '%s.%s' % (FOREVER, build)


def download(url, dest):
    """Download url into dest, trying four times, and through a .part file.

    A run stopped halfway leaves no short table behind. wago.tools sometimes
    answers 503 (Cloudflare) or an HTML page instead of a CSV, so we wait and
    try again rather than keep either.
    """
    last = None
    for attempt in range(4):
        try:
            req = urllib.request.Request(url, headers={'User-Agent': AGENT})
            with urllib.request.urlopen(req, timeout=300) as r:
                body = r.read()
            if body[:1] == b'<':
                raise ValueError('an HTML page instead of a CSV')
            os.makedirs(os.path.dirname(dest), exist_ok=True)
            with open(dest + '.part', 'wb') as f:
                f.write(body)
            os.replace(dest + '.part', dest)
            return
        except urllib.error.HTTPError as e:
            last = e
            # 400 is how wago.tools answers for a table the build doesn't have.
            if e.code in (400, 404):
                break
        except (OSError, ValueError) as e:
            last = e
        time.sleep(3 * (attempt + 1))
    raise OSError('%s: %s' % (url, last))


def path(table, build, fetch=True):
    """The cached CSV of a table at a build, downloaded first when we don't have it."""
    v = version(build)
    p = os.path.join(CACHE, v, table + '.csv')
    if fetch and not os.path.exists(p):
        print('downloading %s %s' % (v, table), file=sys.stderr)
        download(URL % (table, v), p)
    return p


def rows(table, build):
    """A table at a build, as a list of rows by column name."""
    with open(path(table, build), newline='', encoding='utf-8') as f:
        return list(csv.DictReader(f))


def by_id(table, build):
    """A table at a build as {ID: row}, or None when wago.tools has no such table."""
    try:
        return {r['ID']: r for r in rows(table, build)}
    except OSError:
        return None


def fetch(build, tables):
    """Download each table of a build we don't have yet. Answers the tables that failed."""
    failed = []
    for t in tables:
        try:
            path(t, build)
        except OSError as e:
            print('failed: %s' % e, file=sys.stderr)
            failed.append(t)
    return failed


def builds():
    """Every Forever (1.60) version in wago.tools' build list, oldest first.

    The beta is listed under several products (wow_classic_beta, wow_cn_beta,
    wowf), and some lag, so we take them all.
    """
    req = urllib.request.Request(BUILDS, headers={'User-Agent': AGENT})
    with urllib.request.urlopen(req, timeout=60) as r:
        d = json.load(r)
    found = {b['version'] for bs in d.values() for b in bs if str(b.get('version', '')).startswith('1.60.')}
    return sorted(found, key=key)


def answers(v):
    """Whether wago.tools has SpellName at version v (it's small)."""
    req = urllib.request.Request(URL % ('SpellName', v), headers={'User-Agent': AGENT, 'Range': 'bytes=0-0'})
    try:
        with urllib.request.urlopen(req, timeout=20) as r:
            return r.status in (200, 206)
    except (OSError, ValueError):
        return False


def latest(probe=0):
    """The newest Forever version wago.tools has, or None when we can't reach it."""
    try:
        best = builds()[-1]
    except (OSError, ValueError, IndexError):
        return None
    if probe:
        head, last = best.rsplit('.', 1)
        cands = ['%s.%d' % (head, n) for n in range(int(last) + 1, int(last) + probe + 1)]
        with concurrent.futures.ThreadPoolExecutor(20) as ex:
            hits = [v for v, ok in zip(cands, ex.map(answers, cands)) if ok]
        if hits:
            best = max(hits, key=key)
    return best


def diff(old, new, tables, show=40):
    """Per table, the rows added, removed and changed between two builds, by ID."""
    names = by_id('SpellName', new) or {}

    def spell(r):
        return next((r[k] for k in ('SpellID', 'Spell') if k in r), None)

    for t in tables:
        a, b = by_id(t, old), by_id(t, new)
        if a is None or b is None:
            print('%s: missing in %s' % (t, old if a is None else new))
            continue
        add, rem = sorted(set(b) - set(a), key=int), sorted(set(a) - set(b), key=int)
        ch = sorted((i for i in set(a) & set(b)
                     if any(a[i][k] != b[i].get(k) for k in a[i] if k != 'VerifiedBuild')), key=int)
        print('%s: +%d -%d ~%d' % (t, len(add), len(rem), len(ch)))
        for i in ch[:show]:
            s = spell(b[i]) or (i if t == 'SpellName' else None)
            name = names.get(s, {}).get('Name_lang', '') if s else ''
            print('   ~%s %s' % (i, name), {k: (a[i][k], b[i].get(k)) for k in a[i]
                                           if a[i][k] != b[i].get(k) and k != 'VerifiedBuild'})
        if t in ('SpellEffect', 'SpellName', 'SkillLineAbility'):
            added = sorted({names.get(spell(b[i]) or i, {}).get('Name_lang', '?') for i in add})
            if added:
                print('   added spells:', ', '.join(added[:show]))


def diff_by_spell(old, new, tables):
    """One line per changed spell, with each changed field, across the spell tables."""
    names = {}
    for b in (old, new):
        names.update({k: r['Name_lang'] for k, r in (by_id('SpellName', b) or {}).items()})
    changed = {}
    for t in tables:
        a, b = {}, {}
        for build, m in ((old, a), (new, b)):
            for r in rows(t, build):
                m.setdefault(r['SpellID'], []).append({k: v for k, v in r.items() if k != 'ID'})
        for sid in set(a) | set(b):
            ra = sorted(a.get(sid, []), key=lambda r: sorted(r.items()))
            rb = sorted(b.get(sid, []), key=lambda r: sorted(r.items()))
            if ra == rb:
                continue
            if t == 'SpellEffect':
                ia = {x['EffectIndex']: x for x in ra}
                ib = {x['EffectIndex']: x for x in rb}
                pairs = [('%s#%s' % (t, i), ia.get(i, {}), ib.get(i, {})) for i in sorted(set(ia) | set(ib))]
            else:
                pairs = [(t, ra[0] if ra else {}, rb[0] if rb else {})]
            for label, x, y in pairs:
                d = {k: (x.get(k), y.get(k)) for k in set(x) | set(y) if x.get(k) != y.get(k)}
                if d:
                    changed.setdefault(sid, []).append(
                        label + ' ' + ' '.join('%s:%s->%s' % (k, v[0], v[1]) for k, v in sorted(d.items())))
    for sid in sorted(changed, key=int):
        print(sid, names.get(sid, '?'), '|', ' ; '.join(changed[sid]))


def tables_for(args):
    if args.tables:
        return args.tables
    if args.set:
        return SETS[args.set]
    return None


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n\n')[0])
    sub = ap.add_subparsers(dest='cmd', required=True)
    p = sub.add_parser('latest')
    p.add_argument('--probe', type=int, default=0)
    sub.add_parser('builds')
    sub.add_parser('cached')
    p = sub.add_parser('fetch')
    p.add_argument('build')
    p.add_argument('tables', nargs='*')
    p.add_argument('--set', choices=sorted(SETS))
    p = sub.add_parser('dir')
    p.add_argument('build')
    p = sub.add_parser('path')
    p.add_argument('build')
    p.add_argument('table')
    p = sub.add_parser('diff')
    p.add_argument('old')
    p.add_argument('new')
    p.add_argument('tables', nargs='*')
    p.add_argument('--set', choices=sorted(SETS))
    p.add_argument('--by-spell', action='store_true')
    p.add_argument('--show', type=int, default=40)
    a = ap.parse_args()

    if a.cmd == 'latest':
        v = latest(a.probe)
        print(v or 'unknown')
        return 0 if v else 2
    if a.cmd == 'builds':
        print('\n'.join(builds()))
    elif a.cmd == 'cached':
        print('\n'.join(cached()))
    elif a.cmd == 'fetch':
        tables = tables_for(a) or sorted({t for s in SETS.values() for t in s})
        failed = fetch(a.build, tables)
        print(os.path.join(CACHE, version(a.build)))
        return 1 if failed else 0
    elif a.cmd == 'dir':
        print(os.path.join(CACHE, version(a.build)))
    elif a.cmd == 'path':
        print(path(a.table, a.build))
    elif a.cmd == 'diff':
        if a.by_spell:
            diff_by_spell(a.old, a.new, tables_for(a) or SPELL_TABLES)
        else:
            tables = tables_for(a)
            if not tables:
                both = [set(os.listdir(os.path.join(CACHE, version(b)))) for b in (a.old, a.new)
                        if os.path.isdir(os.path.join(CACHE, version(b)))]
                tables = sorted(t[:-4] for t in set.intersection(*both) if t.endswith('.csv')) if len(both) == 2 else []
            diff(a.old, a.new, tables, a.show)
    return 0


if __name__ == '__main__':
    sys.exit(main())
