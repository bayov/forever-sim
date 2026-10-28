#!/usr/bin/env python3
"""changes.py: the DPS Report's Changes tab, the report's results before and after each update. It runs no sims.

An update is a game update (a new client build or development notes) or a sim/preset update (our own changes after
a game update was done). Small changes in between join the latest update. tools/dps_report/changes/ holds:

- updates.json, hand-kept: one entry per update, oldest first. `lines` is the public change list, in plain game terms
  (no commit ids, dates, session or audit names, ruling references or file paths: the page ships it). Each line may
  say what it touches (`classes`, `builds`, `races`, racial `parts`), which the page uses to show its effect, and
  `commits`, which never ship and only tell `check` which commits it covers. `internal` lists commits left out of
  the public list on purpose ({commit, why}). `baselineData` is the commit whose report data the update starts from,
  `endData` the one it ended with (null while it is the live update, which ends at the current data). `check` looks
  at the commits after baselineData.
- genesis.json: the results at the first update's baseline, in full.
- <id>.json: a finished update's end results, only the rows that differ from its baseline (the previous update's end).

A snapshot keeps, from the report's datasets: each build's DPS, error, ranks, race and best race (builds.json); each
build on each race (races.json); each racial part's DPS against the generic boss (race_breakdown.json, scenario
`all`); and how often each race is best and its average gap, per role (race_summary.json, baseline over all
presets). The page rebuilds any update's before and after from genesis plus the chain.

Usage (from anywhere):
  changes.py build                    write the page's datasets changes and changes_snap (pipeline.sh publish runs it)
  changes.py check                    list the live update's commits that change build inputs and no public line covers
  changes.py new game|sim <id> "<title>" [--date YYYY-MM-DD] [--source-label L --source-url U]
                                      finish the live update at the current data and start a new one
  changes.py diff <old ref> [<new ref>]
                                      print what moved between two commits' report data (default new: the work tree)
  changes.py freeze                   rewrite genesis.json and every finished update's file from git (genesis from the
                                      first update's baselineData, each update from its endData), for a backfill or repair
"""
import json, os, subprocess, sys
from datetime import datetime, timezone

T = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(T, '..', '..'))
CH = os.path.join(T, 'changes')
REPORT = os.path.join(ROOT, 'ui', 'dps_report')
DATA_FILES = ('builds', 'races', 'race_breakdown', 'race_summary')
# Fields of an update and of a line the page gets. Everything else (commits, internal, the data commits) stays here.
PUBLIC_UPDATE = ('id', 'type', 'title', 'date', 'source', 'summary')
PUBLIC_LINE = ('title', 'kind', 'note', 'classes', 'builds', 'races', 'parts', 'source')
KINDS = ('game', 'sim', 'rotation', 'gear', 'talents', 'consumes', 'race', 'method')
CLASSES = ('warrior', 'paladin', 'hunter', 'rogue', 'priest', 'shaman', 'mage', 'warlock', 'druid')
# Paths whose changes can move a build's numbers: stale.mjs's inputs, plus the report's own sim scripts (method).
INPUT_PATHS = ['sim', 'proto', 'assets/database', 'tools/rotopt', 'tools/database', 'tools/dps_report/patches',
               'tools/dps_report/req', 'tools/dps_report/asm_apls', 'tools/dps_report/run_build.py',
               'tools/dps_report/run_races.py', 'tools/dps_report/build_data.py', 'tools/dps_report/race_override.py',
               'tools/dps_report/race_gear.py']


def git(*a, check=True):
    r = subprocess.run(['git', *a], cwd=ROOT, capture_output=True, text=True)
    if check and r.returncode:
        sys.exit(f'git {" ".join(a)}: {r.stderr.strip()}')
    return r.stdout


def load_data(ref=None):
    """The four datasets at a commit, or in the work tree when ref is None."""
    out = {}
    for f in DATA_FILES:
        p = f'ui/dps_report/data/{f}.json'
        out[f] = json.loads(git('show', f'{ref}:{p}') if ref else open(os.path.join(ROOT, p)).read())
    return out


def num(v, nd=2):
    return None if v is None else round(float(v), nd)


def snapshot(d):
    """The compact snapshot of one set of report data. Rows are lists, so a changed number repeats only its row:
    builds [dps, se, rank, rankAll, race, best race]; meta [class, name, encounter, duty, base build]; races
    [dps, se]; parts [dps, se, active]; labels a racial part's name; summary [times best, average gap, presets]."""
    b = {r['id']: [num(r['dps']), num(r['se']), r.get('rank'), r.get('rankAll'), r.get('race'), r.get('bestRace')] for r in d['builds']}
    meta = {r['id']: [r['class'], r.get('build'), r.get('encounter') or None, r.get('duty') or None, r.get('baseBuild') or None] for r in d['builds']}
    races = {f"{r['build']}|{r['race']}": [num(r['dps']), num(r['se'])] for r in d['races']}
    rb = [r for r in d['race_breakdown'] if r.get('scenario') == 'all']
    parts = {f"{r['build']}|{r['race']}|{r['part']}": [num(r['dps']), num(r['se']), 1 if r.get('active') else 0] for r in rb}
    labels = {f"{r['race']}|{r['part']}": r.get('label') for r in rb}
    summ = {f"{r['race']}|{r['role']}": [r.get('timesBest'), num(r.get('avgGap')), r.get('presets')]
            for r in d['race_summary'] if r.get('variant') == 'baseline' and r.get('scope') == 'All'}
    return {'builds': b, 'meta': meta, 'races': races, 'parts': parts, 'labels': labels, 'summary': summ}


SECTIONS = ('builds', 'meta', 'races', 'parts', 'labels', 'summary')


def sparse(new, old):
    """The rows of new that differ from old, and the keys old has that new lacks."""
    out = {s: {k: v for k, v in new[s].items() if old[s].get(k) != v} for s in SECTIONS}
    gone = {s: sorted(k for k in old[s] if k not in new[s]) for s in SECTIONS}
    out['gone'] = {s: v for s, v in gone.items() if v}
    return out


def apply(state, delta):
    st = {s: dict(state[s]) for s in SECTIONS}
    for s in SECTIONS:
        st[s].update(delta.get(s, {}))
        for k in delta.get('gone', {}).get(s, []):
            st[s].pop(k, None)
    return st


def read_json(p):
    return json.load(open(p))


def write_json(p, v, indent=None):
    with open(p, 'w') as f:
        f.write(json.dumps(v, indent=indent, ensure_ascii=False, separators=None if indent else (',', ':')) + '\n')


def updates():
    return read_json(os.path.join(CH, 'updates.json'))


def save_updates(u):
    write_json(os.path.join(CH, 'updates.json'), u, indent=1)


def live(u):
    lv = [x for x in u['updates'] if not x.get('endData')]
    if len(lv) != 1:
        sys.exit(f'updates.json needs exactly one live update (endData null), has {len(lv)}')
    return lv[0]


def validate(u):
    ids = set()
    for x in u['updates']:
        if x['id'] in ids:
            sys.exit(f'duplicate update id {x["id"]}')
        ids.add(x['id'])
        if x['type'] not in ('game', 'sim'):
            sys.exit(f'{x["id"]}: type must be game or sim')
        for ln in x.get('lines', []):
            if ln.get('kind') not in KINDS:
                sys.exit(f'{x["id"]}: line "{ln.get("title")}" has kind {ln.get("kind")!r}, not one of {", ".join(KINDS)}')
    live(u)


def chain(u):
    """Genesis, then each update's end delta: the closed ones from their files, the live one from the work tree."""
    g = read_json(os.path.join(CH, 'genesis.json'))
    state, ends = g, {}
    for x in u['updates']:
        if x.get('endData'):
            d = read_json(os.path.join(CH, x['id'] + '.json'))
        else:
            d = sparse(snapshot(load_data()), state)
        ends[x['id']] = d
        state = apply(state, d)
    return g, ends


def stamp(ds, changed):
    mp = os.path.join(REPORT, 'datasets.json')
    m = read_json(mp)
    if ds not in m:
        sys.exit(f'{ds} is not in ui/dps_report/datasets.json')
    if changed:
        m[ds]['updated'] = {'at': datetime.now(timezone.utc).isoformat(timespec='milliseconds').replace('+00:00', 'Z'), 'by': 'Claude (forever-sim)'}
        open(mp, 'w').write(json.dumps(m, indent=1, ensure_ascii=False) + '\n')  # as rawstore.py writes it


def write_dataset(ds, value):
    p = os.path.join(REPORT, 'data', ds + '.json')
    new = json.dumps(value, ensure_ascii=False, separators=(',', ':')) + '\n'
    changed = not os.path.exists(p) or open(p).read() != new
    if changed:
        open(p, 'w').write(new)
    stamp(ds, changed)
    return len(new.encode()), changed


def cmd_build():
    u = updates()
    validate(u)
    g, ends = chain(u)
    pub = []
    for x in u['updates']:
        e = {k: x[k] for k in PUBLIC_UPDATE if k in x}
        e['live'] = not x.get('endData')
        e['lines'] = [{k: ln[k] for k in PUBLIC_LINE if k in ln} for ln in x.get('lines', [])]
        pub.append(e)
    n1, c1 = write_dataset('changes', pub[::-1])  # newest first
    n2, c2 = write_dataset('changes_snap', {'genesis': g, 'ends': ends})
    print(f'changes: {len(pub)} updates, {n1 / 1024:.1f} KB{" (changed)" if c1 else ""}; '
          f'changes_snap: {n2 / 1024:.1f} KB{" (changed)" if c2 else ""}')
    uncovered = check(u, quiet=True)
    if uncovered:
        print(f'warning: {len(uncovered)} commit(s) in the live update change build inputs and no public line covers them '
              f'(changes.py check):', file=sys.stderr)
        for c in uncovered:
            print('  ' + c, file=sys.stderr)


def class_dirs():
    return [f'ui/{n}' for n in sorted(os.listdir(os.path.join(ROOT, 'ui')))
            if os.path.isdir(os.path.join(ROOT, 'ui', n)) and any(c in n.split('_') for c in CLASSES)]


def check(u, quiet=False):
    lv = live(u)
    covered = [c for ln in lv.get('lines', []) for c in ln.get('commits', [])] + [i['commit'] for i in lv.get('internal', [])]
    base = lv['baselineData']
    if not git('rev-parse', '--verify', '-q', base + '^{commit}', check=False).strip():
        sys.exit(f'{base} is not in this clone: deepen it (git fetch --shallow-since=<date>) to check the live update')
    # every commit the baseline's data doesn't include, whatever branch it came in on
    log = git('log', '--no-merges', '--format=%H %s', f'{base}..HEAD', '--', *INPUT_PATHS, *class_dirs()).splitlines()
    out = [l for l in log if not any(l.startswith(c) for c in covered)]
    if not quiet:
        print(f'live update {lv["id"]}: {len(log)} commit(s) change build inputs, {len(out)} not covered')
        for l in out:
            print('  ' + l)
    return out


def cmd_new(typ, uid, title, date=None, src_label=None, src_url=None):
    if typ not in ('game', 'sim'):
        sys.exit('type must be game or sim')
    u = updates()
    validate(u)
    lv = live(u)
    if any(x['id'] == uid for x in u['updates']):
        sys.exit(f'update {uid} exists already')
    g, ends = chain(u)
    write_json(os.path.join(CH, lv['id'] + '.json'), ends[lv['id']])
    end = git('log', '-1', '--format=%H', '--', 'ui/dps_report/data').strip()
    if git('status', '--porcelain', '--', 'ui/dps_report/data/builds.json', 'ui/dps_report/data/races.json',
           'ui/dps_report/data/race_breakdown.json', 'ui/dps_report/data/race_summary.json').strip():
        sys.exit('the report data has uncommitted changes: commit them before finishing the update')
    lv['endData'] = end
    x = {'id': uid, 'type': typ, 'title': title, 'date': date or datetime.now(timezone.utc).strftime('%Y-%m-%d'),
         'baselineData': end, 'endData': None, 'lines': [], 'internal': []}
    if src_label or src_url:
        x['source'] = {'label': src_label or src_url, 'url': src_url}
    u['updates'].append(x)
    save_updates(u)
    print(f'finished {lv["id"]} at {end[:10]}; started {typ} update {uid}')
    cmd_build()


def cmd_freeze():
    u = updates()
    validate(u)
    state = snapshot(load_data(u['updates'][0]['baselineData']))
    write_json(os.path.join(CH, 'genesis.json'), state)
    for x in u['updates']:
        if not x.get('endData'):
            break
        end = snapshot(load_data(x['endData']))
        write_json(os.path.join(CH, x['id'] + '.json'), sparse(end, state))
        state = end
        print(f'froze {x["id"]} at {x["endData"][:10]}')
    cmd_build()


def cmd_diff(old, new=None):
    a, b = snapshot(load_data(old)), snapshot(load_data(new))
    print(f'builds ({old} -> {new or "work tree"}):')
    for k in sorted(set(a['builds']) | set(b['builds'])):
        x, y = a['builds'].get(k), b['builds'].get(k)
        if not x or not y:
            print(f'  {k}: {"added" if y else "retired"}')
            continue
        d = y[0] - x[0]
        if abs(d) >= 0.05 or x[4] != y[4]:
            race = f'  race {x[4]} -> {y[4]}' if x[4] != y[4] else ''
            print(f'  {k}: {x[0]} -> {y[0]} ({d:+.1f}){race}')
    moved = {}
    for k in set(a['races']) & set(b['races']):
        d = b['races'][k][0] - a['races'][k][0]
        if abs(d) >= 0.05:
            moved.setdefault(k.split('|')[1], []).append(d)
    print('race rows moved, per race: ' + ', '.join(f'{r} {len(v)} ({min(v):+.1f} to {max(v):+.1f})' for r, v in sorted(moved.items())))
    pm = {}
    for k in set(a['parts']) | set(b['parts']):
        x, y = a['parts'].get(k), b['parts'].get(k)
        if (x and x[2]) or (y and y[2]):
            d = (y[0] if y else 0) - (x[0] if x else 0)
            if abs(d) >= 0.05:
                pm.setdefault(k.split('|', 1)[1], []).append(d)
    print('racial parts moved: ' + ', '.join(f'{p} {len(v)} ({min(v):+.2f} to {max(v):+.2f})' for p, v in sorted(pm.items())))


if __name__ == '__main__':
    a = sys.argv[1:]
    opt = lambda n: a[a.index(n) + 1] if n in a else None
    if a[:1] == ['build']:
        cmd_build()
    elif a[:1] == ['check']:
        check(updates())
    elif a[:1] == ['new'] and len(a) >= 4:
        cmd_new(a[1], a[2], a[3], opt('--date'), opt('--source-label'), opt('--source-url'))
    elif a[:1] == ['diff'] and len(a) >= 2:
        cmd_diff(a[1], a[2] if len(a) > 2 and not a[2].startswith('--') else None)
    elif a[:1] == ['freeze']:
        cmd_freeze()
    else:
        sys.exit(__doc__)
