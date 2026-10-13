"""wagocheck.py: item level and stats of the dashboard's real items, Forever wowhead vs the Forever client (wago ItemSparse) vs the sim's db."""
import csv, json, os, re, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D, TREE
import whcache
csv.field_size_limit(10**9)
ids = [i for k, i in json.load(open(os.path.join(D, 'fortt_ids.json'))) if k == 'item' and i < 990000]
wago = {}
with open(os.path.join(TREE, 'assets/db_inputs/wago_db2_items.csv')) as f:
    r = csv.reader(f)
    hdr = next(r)
    ix = {h: n for n, h in enumerate(hdr)}
    for row in r:
        if row and row[0].isdigit() and int(row[0]) in ids:
            wago[int(row[0])] = row
db = {i['id']: i for i in json.load(open(os.path.join(TREE, 'assets/database/db.json')))['items']}
STAT = {3: 'Agility', 4: 'Strength', 5: 'Intellect', 6: 'Spirit', 7: 'Stamina'}
bad = 0
for i in ids:
    j = whcache.cached('forever', 'item', i)
    if not whcache.ok(j):
        continue
    m = re.search(r'Item Level (?:<!--ilvl-->)?(\d+)', j['tooltip'])
    wl = int(m.group(1)) if m else None
    w = wago.get(i)
    cl = int(w[ix['ItemLevel']]) if w else None
    stats_wh = dict((int(a), int(b)) for a, b in re.findall(r'<!--stat(\d+)-->\+(\d+)', j['tooltip']))
    stats_cl = {}
    if w:
        for n in range(10):
            st, amt = int(w[ix[f'StatModifier_bonusStat_{n}']]), int(w[ix[f'StatModifier_bonusAmount_{n}']])
            if st > 0 and amt:
                stats_cl[st] = amt
    ok = (wl == cl or cl is None) and (not w or {k: v for k, v in stats_cl.items() if k in STAT} == {k: v for k, v in stats_wh.items() if k in STAT})
    bad += not ok
    if not ok or i in db:
        print(i, j['name'], 'ilvl wowhead', wl, 'client', cl, 'db', db.get(i, {}).get('ilvl'), 'stats wowhead', stats_wh, 'client', stats_cl, 'OK' if ok else 'MISMATCH')
print('items checked', len(ids), 'in client table', len(wago), 'mismatches', bad)
