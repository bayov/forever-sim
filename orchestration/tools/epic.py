import json, math, collections, statistics
import numpy as np
exec(open('/tmp/claude-0/-home-user-forever-sim/ca4b4203-ff20-5bf4-98f0-d157160fd32d/scratchpad/ilvl/fit.py').read().split("for band, flt in")[0])
SM = {'head': 1, 'chest': 1, 'legs': 1, 'shoulder': .77, 'hands': .77, 'waist': .77, 'feet': .77, 'neck': .5625, 'back': .5625, 'wrist': .5625, 'finger': .5625}
def lin(r, w):
    return sum(w[k] * x for k, x in r['v'].items())
for spw in (0.55, 0.86):
    w = {'str': 1, 'agi': 1, 'sta': 1, 'int': 1, 'spi': 1, 'critstrkrtng': 1.0, 'hitrtng': 1.0, 'splpwr': spw, 'atkpwr': 0.5, 'manargn': 2.5, 'splheal': 0.45, 'spldmg': spw}
    print(f'\n=== spell power weight {spw} ===')
    for q in (3, 4):
        for p in (1.5,):
            rq = [r for r in rows if r['q'] == q and r['ilvl'] >= 40]
            X = np.array([[r['ilvl'], 1.0] for r in rq]); y = np.array([cost(r, w, p) / SM[r['slot']] for r in rq])
            a, b = np.linalg.lstsq(X, y, rcond=None)[0]
            res = y / (X @ np.array([a, b])) - 1
            print(f'quality {q}, p={p}: V(ilvl) = {a:.3f}*ilvl {b:+.2f}  (n={len(rq)}, rms {100*np.sqrt(np.mean(res**2)):.1f}%)  V(66)={a*66+b:.1f} V(71)={a*71+b:.1f} V(76)={a*76+b:.1f}   standards: V(66)=42.8 V(71)=48.8 V(76)=54.8')
            # caster vs physical split
            for lab, f in (('caster items (have splpwr)', lambda r: 'splpwr' in r['v']), ('physical items (have str/agi/atkpwr, no splpwr)', lambda r: 'splpwr' not in r['v'] and any(k in r['v'] for k in ('str', 'agi', 'atkpwr')))):
                sub = [r for r in rq if f(r)]
                if sub:
                    rr = [cost(r, w, p) / SM[r['slot']] / (a * r['ilvl'] + b) for r in sub]
                    print(f'    {lab}: n={len(sub)} median budget ratio {statistics.median(rr):.3f}')
            # linear multiplier: L1 / Lp per item
            ratios = [lin(r, w) / cost(r, w, p) for r in rq]
            nst = collections.Counter(len(r['v']) for r in rq)
            print(f'    real items: linear total / 1.5-norm value: median {statistics.median(ratios):.3f}, mean {statistics.mean(ratios):.3f}, p25-p75 {np.percentile(ratios,25):.3f}-{np.percentile(ratios,75):.3f}; stats per item {dict(sorted(nst.items()))}')
