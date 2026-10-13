import json, collections, statistics, math
d = json.load(open('/home/user/wowsims/sod/assets/database/db.json'))['items']
# stats array: 0 Str 1 Agi 2 Sta 3 Int 4 Spi; anything else nonzero (except armor 26, bonus armor 40) -> skip
items = []
for it in d:
    s = it.get('stats') or []
    s = s + [0] * (44 - len(s))
    if it.get('quality') not in (2, 3, 4) or it.get('type') not in range(1, 17):
        continue
    other = [i for i, v in enumerate(s) if v and i not in (0, 1, 2, 3, 4, 26, 40)]
    if other or it.get('weaponType'):
        continue
    vals = sorted([v for v in s[:5] if v > 0], reverse=True)
    if not vals:
        continue
    items.append({'id': it['id'], 'ilvl': it.get('ilvl', 0), 'q': it['quality'], 'slot': it['type'], 'vals': vals, 'sod': it['id'] >= 200000})
print('SoD db primary-only items:', len(items), 'of which new SoD items (id>=200000):', sum(x['sod'] for x in items))
def norm(v, p):
    return sum(a ** p for a in v) ** (1 / p)
for label, flt in (('new SoD items', lambda x: x['sod']), ('classic items in the SoD db', lambda x: not x['sod'])):
    groups = collections.defaultdict(list)
    for x in items:
        if flt(x):
            groups[(x['q'], x['slot'], x['ilvl'])].append(x)
    mixed = {k: v for k, v in groups.items() if len({len(x['vals']) for x in v}) > 1}
    print(f'\n{label}: groups {len(mixed)}, items {sum(len(v) for v in mixed.values())}')
    for p in (1.0, 1.25, 1.5, 1.75, 2.0):
        cvs = []
        for k, v in mixed.items():
            ns = [norm(x['vals'], p) for x in v]
            m = statistics.mean(ns)
            cvs += [(n / m - 1) for n in ns]
        if cvs:
            print(f'  p={p:4}: rms {100*math.sqrt(sum(c*c for c in cvs)/len(cvs)):5.2f}%')
