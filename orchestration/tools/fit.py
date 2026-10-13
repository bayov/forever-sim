import json, math, collections
import numpy as np
d = json.load(open('/home/user/w-dash/assets/db_inputs/forever_wowhead_items.json'))['items']
STATS = ['str', 'agi', 'sta', 'int', 'spi', 'critstrkrtng', 'hitrtng', 'splpwr', 'atkpwr', 'manargn', 'splheal', 'spldmg']
FIXED = {'str': 1, 'agi': 1, 'sta': 1, 'int': 1, 'spi': 1}
IGNORE = {'slotbak', 'reqlevel', 'armor', 'classes', 'itemset', 'races', 'reqfaction', 'reqrep', 'reqskill', 'reqskillrank'}
SLOT = {1: 'head', 2: 'neck', 3: 'shoulder', 5: 'chest', 20: 'chest', 6: 'waist', 7: 'legs', 8: 'feet', 9: 'wrist', 10: 'hands', 11: 'finger', 16: 'back'}
rows = []
for it in d:
    eq = it.get('eq', {})
    if it.get('slot') not in SLOT or it.get('quality') not in (2, 3, 4):
        continue
    if any(k not in IGNORE and k not in STATS for k in eq):
        continue
    if eq.get('splheal') and not eq.get('splpwr'):
        continue
    v = {k: eq.get(k, 0) for k in STATS if eq.get(k, 0)}
    if 'splheal' in v and 'splpwr' in v:
        v.pop('splheal')  # healing listed alongside spell power is the same stat
    if not v or any(x < 0 for x in v.values()):
        continue
    rows.append({'id': it['id'], 'name': it['name'], 'ilvl': it['level'], 'q': it['quality'], 'slot': SLOT[it['slot']], 'v': v})
print('items:', len(rows), collections.Counter(r['q'] for r in rows))
slots = sorted({r['slot'] for r in rows})
stats_free = [s for s in STATS if s not in FIXED and any(s in r['v'] for r in rows)]
print('free weights:', stats_free, {s: sum(1 for r in rows if s in r['v']) for s in stats_free})

def cost(r, w, p):
    return sum((w[k] * x) ** p for k, x in r['v'].items()) ** (1 / p)

def fit(p, rows, iters=40, verbose=False):
    w = dict(FIXED); w.update({'critstrkrtng': 1.0, 'hitrtng': 1.0, 'splpwr': 0.55, 'atkpwr': 0.5, 'manargn': 2.5, 'splheal': 0.3, 'spldmg': 0.55})
    m = {s: 1.0 for s in slots}
    ab = {q: (1.0, 0.0) for q in (2, 3, 4)}
    for it in range(iters):
        # V per quality: linear fit of cost/m vs ilvl
        for q in (2, 3, 4):
            rq = [r for r in rows if r['q'] == q]
            if len(rq) < 8:
                ab[q] = (0.0, 0.0)
                continue
            X = np.array([[r['ilvl'], 1.0] for r in rq]); y = np.array([cost(r, w, p) / m[r['slot']] for r in rq])
            ab[q] = tuple(np.linalg.lstsq(X, y, rcond=None)[0])
        V = lambda r: ab[r['q']][0] * r['ilvl'] + ab[r['q']][1]
        # slot mods: median ratio, chest = 1
        for s in slots:
            rs = [cost(r, w, p) / V(r) for r in rows if r['slot'] == s and V(r) > 0]
            m[s] = float(np.median(rs)) if rs else 1.0
        base = m['chest']
        for s in slots:
            m[s] /= base
        ab = {q: (a * base, b * base) for q, (a, b) in ab.items()}
        # weights: 1-D search each, minimizing log error over items containing the stat
        for k in stats_free:
            rk = [r for r in rows if k in r['v']]
            best = None
            for cand in np.linspace(0.05, 4.0, 160):
                w[k] = cand
                err = sum(math.log(cost(r, w, p) / (m[r['slot']] * V(r))) ** 2 for r in rk if V(r) > 0)
                if best is None or err < best[0]:
                    best = (err, cand)
            w[k] = best[1]
    V = lambda r: ab[r['q']][0] * r['ilvl'] + ab[r['q']][1]
    res = [math.log(cost(r, w, p) / (m[r['slot']] * V(r))) for r in rows if V(r) > 0]
    return math.sqrt(np.mean(np.square(res))), w, m, ab

for band, flt in (('ilvl 1-35 (beta-tuned range)', lambda r: r['ilvl'] <= 35), ('ilvl 36-100', lambda r: r['ilvl'] > 35), ('all', lambda r: True)):
    sub = [r for r in rows if flt(r)]
    print(f'\n== {band}: {len(sub)} items')
    out = {}
    for p in (1.0, 1.25, 1.5, 1.75, 2.0):
        e, w, m, ab = fit(p, sub, iters=12)
        out[p] = (e, w, m, ab)
        print(f'  p={p}: rms log error {100*e:.2f}%')
    pb = min(out, key=lambda p: out[p][0])
    e, w, m, ab = out[pb]
    print(f'  best p={pb}: weights', {k: round(w[k], 2) for k in stats_free})
    print('  slot mods', {s: round(m[s], 3) for s in slots})
    print('  V(ilvl) = a*ilvl + b per quality', {q: (round(a, 3), round(b, 2)) for q, (a, b) in ab.items()})
    json.dump({'p': pb, 'w': w, 'm': m, 'ab': {str(q): v for q, v in ab.items()}}, open(f'/tmp/claude-0/-home-user-forever-sim/ca4b4203-ff20-5bf4-98f0-d157160fd32d/scratchpad/ilvl/fit_{band.split()[1]}.json', 'w'))
