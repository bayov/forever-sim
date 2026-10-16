"""talent_links.py [ids|refresh]: the wowhead link of each talent rank, for the talenttrees dataset.

Forever talents use the trait system. Forever's client (and so Forever's wowhead) keeps only one spell per talent,
the rank 1 spell of the Forever client tree (tools/forever_talents/client/<class>.json). Our talent trees list a
spell per rank, which are Classic's ids. So we link every rank to the client's rank 1 spell on Forever's wowhead
when it knows it. Else we fall back to each rank's own spell, on Forever's wowhead and then on Classic's.

A link is "<domain>:<spell id>", or "" when neither wowhead knows the spell. build_data.py calls links() for each
talent. The two commands work without the rest of the pipeline:

  ids      writes talent_ids.json ([["spell", id], ...]) with every spell we might link, for whcache.py
  refresh  rewrites the links in the shipped talenttrees.json (ui/dps_report/data) from the cache, so a cache
           refresh reaches the report without a full build_data.py run
"""
import json, os, re, sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D
import whcache

HERE = os.path.dirname(os.path.abspath(__file__))
CLIENT = os.path.join(HERE, '..', 'forever_talents', 'client')
TREES = os.path.join(HERE, '..', '..', 'ui', 'core', 'talents', 'trees')
REPORT_DATA = os.path.join(HERE, '..', '..', 'ui', 'dps_report', 'data')

_client = {}


def _norm(s):
    return re.sub(r'[^a-z0-9]', '', str(s).lower())


def client_spell(cls, tree, name):
    """The Forever client's rank 1 spell id of a talent, by its name, in its tree first. None when the client
    tree doesn't have it."""
    if cls not in _client:
        p = os.path.join(CLIENT, cls.lower() + '.json')
        _client[cls] = json.load(open(p))['trees'] if os.path.exists(p) else []
    trees = _client[cls]
    order = ([trees[tree]] if tree < len(trees) else []) + [t for i, t in enumerate(trees) if i != tree]
    for t in order:
        for tal in t['talents']:
            if _norm(tal['name']) == _norm(name):
                return (tal.get('tooltip') or {}).get('spellId')
    return None


def links(cls, tree, tal):
    """One link per rank of a talent from a tree JSON (ui/core/talents/trees)."""
    ranks = tal.get('spellIds') or []
    first = client_spell(cls, tree, tal['name'])
    if first and whcache.ok(whcache.cached('forever', 'spell', first)):
        return [f'forever:{first}'] * max(len(ranks), tal['maxPoints'])
    out = [next((f'{dm}:{sid}' for dm in ('forever', 'classic') if whcache.ok(whcache.cached(dm, 'spell', sid))), '') for sid in ranks]
    # Forever's wowhead knows one rank of a talent the client trees don't have (like Call of Thunder's rank 5).
    # That's the talent's one Forever spell, so every rank links to it.
    on_forever = next((w for w in out if w.startswith('forever:')), None)
    return [on_forever] * len(out) if on_forever else out


def _all_trees():
    for f in sorted(os.listdir(TREES)):
        if f.endswith('.json'):
            cls = f[:-5].capitalize()
            for ti, tr in enumerate(json.load(open(os.path.join(TREES, f)))):
                for i, tal in enumerate(tr['talents']):
                    yield cls, ti, i, tal


if __name__ == '__main__':
    cmd = sys.argv[1] if len(sys.argv) > 1 else ''
    if cmd == 'ids':
        ids = set()
        for cls, ti, _, tal in _all_trees():
            first = client_spell(cls, ti, tal['name'])
            ids |= {('spell', s) for s in ([first] if first else []) + list(tal.get('spellIds') or []) if s}
        p = os.path.join(D, 'talent_ids.json')
        json.dump(sorted(ids), open(p, 'w'))
        print(len(ids), 'spells ->', p)
    elif cmd == 'refresh':
        by = {(cls, ti, i, tal['name']): tal for cls, ti, i, tal in _all_trees()}
        for d in (REPORT_DATA, os.path.join(D, 'data')):
            p = os.path.join(d, 'talenttrees.json')
            if not os.path.exists(p):
                continue
            rows = json.load(open(p))
            n = 0
            for r in rows:
                # A talent the trees no longer have where the dataset has it (the dataset can be older than the
                # trees) keeps its per rank spells from its current links.
                tal = by.get((r['class'], r['tree'], r['idx'], r['name'])) or {
                    'name': r['name'], 'maxPoints': r['maxPoints'], 'spellIds': [int(w.split(':')[1]) for w in r['wh'] if w]}
                wh = links(r['class'], r['tree'], tal)
                n += wh != r['wh']
                r['wh'] = wh
            json.dump(rows, open(p, 'w'), separators=(',', ':'))
            count = {}
            for r in rows:
                for w in r['wh']:
                    k = w.split(':')[0] or 'none'
                    count[k] = count.get(k, 0) + 1
            print(f'{p}: {n} of {len(rows)} talents changed, ranks by link: {count}')
    else:
        print(__doc__)
