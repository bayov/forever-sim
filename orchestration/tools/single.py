import json, collections, statistics
d = json.load(open('/home/user/w-dash/assets/db_inputs/forever_wowhead_items.json'))['items']
IGNORE = {'slotbak', 'reqlevel', 'armor', 'classes', 'itemset', 'races', 'reqfaction', 'reqrep', 'reqskill', 'reqskillrank'}
SLOT = {1: 'head', 2: 'neck', 3: 'shoulder', 5: 'chest', 20: 'chest', 6: 'waist', 7: 'legs', 8: 'feet', 9: 'wrist', 10: 'hands', 11: 'finger', 16: 'back'}
PRIM = {'str', 'agi', 'sta', 'int', 'spi'}
single = collections.defaultdict(list)  # (q, slot, ilvl) -> [(stat, value, name)]
for it in d:
    eq = it.get('eq', {})
    if it.get('slot') not in SLOT or it.get('quality') not in (2, 3, 4):
        continue
    st = {k: v for k, v in eq.items() if k not in IGNORE}
    if 'splheal' in st and 'splpwr' in st:
        st.pop('splheal')
    if len(st) != 1:
        continue
    (k, v), = st.items()
    if v <= 0:
        continue
    single[(it['quality'], SLOT[it['slot']], it['level'])].append(('prim' if k in PRIM else k, v, it['name']))
# For each non-primary stat: in groups that also have a single-primary item, ratio = primary value / stat value
ratios = collections.defaultdict(list)
for key, lst in single.items():
    prim = [v for k, v, n in lst if k == 'prim']
    if not prim:
        continue
    pv = statistics.mean(prim)
    for k, v, n in lst:
        if k != 'prim':
            ratios[k].append((pv / v, key, n, v, pv))
for k, rs in sorted(ratios.items()):
    vals = [r[0] for r in rs]
    print(f'{k}: n={len(vals)} median weight {statistics.median(vals):.2f}  (range {min(vals):.2f}-{max(vals):.2f})')
    for r in sorted(rs, key=lambda r: r[1][2])[:6]:
        print('    ', r[1], r[2], f'{r[3]} vs primary {r[4]:.1f} -> {r[0]:.2f}')
# Also regression approach: single-stat items of each kind across all ilvls, value per slot-mod vs ilvl, compared with single-primary trend
SM = {'head': 1, 'chest': 1, 'legs': 1, 'shoulder': .75, 'hands': .75, 'waist': .75, 'feet': .75, 'neck': .5625, 'back': .5625, 'wrist': .5625, 'finger': .5625}
import numpy as np
trend = {}
for kind in ('prim', 'splpwr', 'critstrkrtng', 'hitrtng', 'atkpwr', 'manargn', 'splheal'):
    for q in (2, 3, 4):
        pts = [(key[2], v / SM[key[1]]) for key, lst in single.items() if key[0] == q for k, v, n in lst if k == kind]
        if len(pts) >= 6:
            X = np.array([[x, 1] for x, _ in pts]); y = np.array([y for _, y in pts])
            a, b = np.linalg.lstsq(X, y, rcond=None)[0]
            trend[(kind, q)] = (a, b, len(pts))
for q in (2, 3, 4):
    if ('prim', q) not in trend:
        continue
    a0, b0, n0 = trend[('prim', q)]
    print(f'\nquality {q}: single primary stat per slot-mod = {a0:.3f}*ilvl {b0:+.2f} (n={n0})')
    for kind in ('splpwr', 'critstrkrtng', 'hitrtng', 'atkpwr', 'manargn', 'splheal'):
        if (kind, q) in trend:
            a, b, n = trend[(kind, q)]
            for L in (30, 60, 66):
                pass
            print(f'  {kind}: {a:.3f}*ilvl {b:+.2f} (n={n}); implied weight at ilvl 30/50/66: ' + ', '.join(f'{(a0*L+b0)/(a*L+b):.2f}' for L in (30, 50, 66)))
