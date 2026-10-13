import json, collections, statistics, math
d = json.load(open('/home/user/w-dash/assets/db_inputs/forever_wowhead_items.json'))['items']
PRIM = ['str', 'agi', 'sta', 'int', 'spi']
ALLOWED = set(PRIM) | {'slotbak', 'reqlevel', 'armor', 'classes', 'itemset', 'races', 'reqfaction', 'reqrep', 'reqskill', 'reqskillrank', 'dmgrange', 'speed', 'dps', 'dmgmax1', 'dmgmin1', 'dmgtype1', 'mledps', 'mlespeed', 'mledmgmax', 'mledmgmin', 'rgddmgmax', 'rgddmgmin', 'rgddps', 'rgdspeed'}
ARMOR_SLOTS = {1, 2, 3, 5, 6, 7, 8, 9, 10, 11, 16, 20}
items = []
for it in d:
    eq = it.get('eq', {})
    if it.get('slot') not in ARMOR_SLOTS or it.get('quality') not in (2, 3, 4):
        continue
    if any(k not in ALLOWED for k in eq):
        continue
    vals = [eq.get(k, 0) for k in PRIM if eq.get(k, 0) > 0]
    if not vals or any(eq.get(k, 0) < 0 for k in PRIM):
        continue
    slot = 5 if it['slot'] == 20 else it['slot']
    items.append({'id': it['id'], 'name': it['name'], 'ilvl': it['level'], 'q': it['quality'], 'slot': slot, 'vals': sorted(vals, reverse=True)})
print('primary-only armor items:', len(items))
groups = collections.defaultdict(list)
for x in items:
    groups[(x['q'], x['slot'], x['ilvl'])].append(x)
mixed = {k: v for k, v in groups.items() if len({len(x['vals']) for x in v}) > 1}
print('groups with mixed stat counts:', len(mixed), 'items in them:', sum(len(v) for v in mixed.values()))
def norm(v, p):
    return sum(a ** p for a in v) ** (1 / p)
for band in ((1, 35), (36, 60), (61, 100), (1, 100)):
    print(f'\nilvl {band}:')
    for p in (1.0, 1.25, 1.4, 1.5, 1.6, 1.75, 2.0):
        cvs = []
        for k, v in mixed.items():
            if not band[0] <= k[2] <= band[1]:
                continue
            ns = [norm(x['vals'], p) for x in v]
            m = statistics.mean(ns)
            cvs += [(n / m - 1) for n in ns]
        if cvs:
            rms = math.sqrt(sum(c * c for c in cvs) / len(cvs))
            print(f'  p={p:4}: rms deviation within groups {100*rms:5.2f}%  (n={len(cvs)})')
# show some example groups
ex = [v for k, v in sorted(mixed.items()) if 20 <= k[2] <= 30][:8]
for v in ex:
    print([(x['name'], x['ilvl'], x['vals'], round(norm(x['vals'], 1.5), 1)) for x in v][:5])
