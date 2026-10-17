import json, os, statistics
import numpy as np
exec(open(os.path.join(os.path.dirname(os.path.abspath(__file__)), 'fit.py')).read().split("for band, flt in")[0])
w = {'str': 1, 'agi': 1, 'sta': 1, 'int': 1, 'spi': 1, 'critstrkrtng': 1.0, 'hitrtng': 1.0, 'splpwr': 0.86, 'atkpwr': 0.5, 'manargn': 2.5, 'splheal': 0.45, 'spldmg': 0.86}
SM = {'head': 1, 'chest': 1, 'legs': 1, 'shoulder': .77, 'hands': .77, 'waist': .77, 'feet': .77, 'neck': .5625, 'back': .5625, 'wrist': .5625, 'finger': .5625}
lin = lambda r: sum(w[k] * x for k, x in r['v'].items())
for q, lab in ((3, 'rare'), (4, 'epic')):
    for band in ((40, 100), (55, 80)):
        rq = [r for r in rows if r['q'] == q and band[0] <= r['ilvl'] <= band[1]]
        X = np.array([[r['ilvl'], 1.0] for r in rq]); y = np.array([lin(r) / SM[r['slot']] for r in rq])
        a, b = np.linalg.lstsq(X, y, rcond=None)[0]
        res = y / (X @ np.array([a, b])) - 1
        print(f'{lab} ilvl {band}: linear total per slot-mod = {a:.3f}*ilvl {b:+.2f} (n={len(rq)}, rms {100*np.sqrt(np.mean(res**2)):.1f}%)  -> ilvl 66: {a*66+b:.1f}, 71: {a*71+b:.1f}, 76: {a*76+b:.1f}')
        for lab2, f in (('caster', lambda r: 'splpwr' in r['v']), ('physical', lambda r: 'splpwr' not in r['v'] and any(k in r['v'] for k in ('str', 'agi', 'atkpwr')))):
            sub = [r for r in rq if f(r)]
            print(f'     {lab2}: median ratio to the trend {statistics.median([lin(r)/SM[r["slot"]]/(a*r["ilvl"]+b) for r in sub]):.3f} (n={len(sub)})')
print('\nOurs (standards 1a): 1.73 x (1.2*ilvl - 36.4) -> ilvl 66: %.1f, 71: %.1f, 76: %.1f' % (1.73*(1.2*66-36.4), 1.73*(1.2*71-36.4), 1.73*(1.2*76-36.4)))
# sample: real epic chest pieces near ilvl 66 and their linear totals
ex = sorted([r for r in rows if r['q'] == 4 and r['slot'] == 'chest' and 62 <= r['ilvl'] <= 72], key=lambda r: r['ilvl'])[:12]
for r in ex:
    print(f"   {r['name']:40} ilvl {r['ilvl']}  {r['v']}  linear {lin(r):.1f}")
