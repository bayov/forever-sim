"""whids.py: every spell and item id the dashboard shows a tooltip for -> fortt_ids.json ([[kind, id], ...]) for whcache.py."""
import glob, json, os, re, sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D, TREE
import tooltips

ids = set()
for r in json.load(open(os.path.join(D, 'data/names.json'))):
    m = re.match(r'^([si])(\d+)', r['key'])
    if m:
        ids.add(('spell' if m.group(1) == 's' else 'item', int(m.group(2))))
for g in json.load(open(os.path.join(D, 'data/gear.json'))):
    if g['itemId'] < 990000:
        ids.add(('item', g['itemId']))
db = json.load(open(os.path.join(TREE, 'assets/database/db.json')))
ench = {e['effectId']: e for e in db['enchants']}
for p in glob.glob(os.path.join(D, 'out/*.json')):
    o = json.load(open(p))
    for it in o['setup']['equipment']:
        e = ench.get((it or {}).get('enchant'))
        if e:
            if e.get('spellId'):
                ids.add(('spell', e['spellId']))
            if e.get('itemId'):
                ids.add(('item', e['itemId']))
for s in json.load(open(os.path.join(D, 'data/setup.json'))):
    v = tooltips.setup_value(s['item'], s['category'])
    if v in tooltips.SETUP_IDS:
        ids.add(tooltips.SETUP_IDS[v])
    elif s['category'] != 'Professions':
        print('no id for setup chip', s['item'], file=sys.stderr)
for race, lst in json.load(open(os.path.join(TREE, 'ui/core/forever/racials.json'))).items():
    for r in lst:
        for v in r.get('variants', []):
            ids.add(('spell', v['spellId']))
for p in glob.glob(os.path.join(TREE, 'ui/core/talents/trees/*.json')):
    for tr in json.load(open(p)):
        for t in tr['talents']:
            for sid in t.get('spellIds') or []:
                ids.add(('spell', sid))
out = sorted(ids)
json.dump(out, open(os.path.join(D, 'fortt_ids.json'), 'w'))
print(len(out), 'ids:', sum(1 for k, _ in out if k == 'spell'), 'spells,', sum(1 for k, _ in out if k == 'item'), 'items')
