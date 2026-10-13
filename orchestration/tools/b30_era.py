"""Q5 of the ilvl <= 30 analysis: did Forever change the low-level budget against Classic Era?

1. Item by item: every Forever ilvl 5-30 item with a Classic id that Era also has (gear planner,
   versionNum 11300), both clean. Unchanged share, and the value ratio Forever / Era.
2. Curves: V_q(ilvl) and the linear totals fitted on all of Era's ilvl <= 30 items and on Forever's,
   same weights, slot mods and exponent.
3. Swaps: items whose Forever version traded one stat for another at the same ilvl and quality
   (mostly Spirit -> spell power); the exchange rate that keeps the item's value (p-norm and linear).
"""
import collections, math, os, statistics, sys
import numpy as np
sys.path.insert(0, os.path.dirname(__file__))
from b30_data import forever, era, clean, kind_of, ARMOR, WEAPON, PRIM, STD_MOD
from b30_fit import Model, summarize, linear_totals

P = 1.5
W = {'prim': 1.0, 'sp': 0.86, 'heal': 0.40, 'school': 0.60, 'ap': 0.5, 'rap': 0.4, 'crit': 1.0, 'hit': 1.0, 'mp5': 2.0}


def wt(k):
    return 1.0 if k in PRIM else W[kind_of(k)]


def val(v, p=P):
    return sum((wt(k) * x) ** p for k, x in v.items()) ** (1 / p)


def lin(v):
    return sum(wt(k) * x for k, x in v.items())


if __name__ == '__main__':
    F = forever()
    E = {r['id']: r for r in era()}
    slots = ARMOR + WEAPON
    pairs = [(f, E[f['id']]) for f in F if f['src'] == 'classic-id' and f['id'] in E]
    pairs = [(f, e) for f, e in pairs if f['slot'] in slots and not f['other'] and not e['other'] and not e['effect'] and (f['v'] or e['v'])]
    print(f'Forever ilvl 5-30 items with a Classic id that Era has, clean on both sides: {len(pairs)}')
    same = [(f, e) for f, e in pairs if f['v'] == e['v'] and f['ilvl'] == e['ilvl'] and f['q'] == e['q']]
    print(f'  identical stats, ilvl and quality: {len(same)} ({100*len(same)/len(pairs):.0f}%)')
    dil = collections.Counter(f['ilvl'] - e['ilvl'] for f, e in pairs)
    dq = collections.Counter((e['q'], f['q']) for f, e in pairs if f['q'] != e['q'])
    print(f'  item level changed on {sum(v for k, v in dil.items() if k)} items: {dict(sorted((k, v) for k, v in dil.items() if k))}')
    print(f'  quality changed on {sum(dq.values())} items: {dict(dq)}')
    stat_changed = [(f, e) for f, e in pairs if f['v'] != e['v']]
    print(f'  stats changed on {len(stat_changed)} items')

    # 0. wago_db2_items.csv (ItemSparse, build 1.15.3.55646) against both listings, primaries only
    import pandas as pd
    from b30_data import INPUTS
    wg = pd.read_csv(os.path.join(INPUTS, 'wago_db2_items.csv'), low_memory=False).set_index('ID')
    ENUM = {3: 'agi', 4: 'str', 5: 'int', 6: 'spi', 7: 'sta'}
    agree_e = agree_f = n = 0
    for f, e in pairs:
        if f['id'] not in wg.index:
            continue
        row = wg.loc[f['id']]
        wp = {ENUM[int(row[f'StatModifier_bonusStat_{i}'])]: int(row[f'StatModifier_bonusAmount_{i}']) for i in range(10)
              if int(row[f'StatModifier_bonusStat_{i}']) in ENUM and row[f'StatModifier_bonusAmount_{i}']}
        n += 1
        agree_e += wp == {k: x for k, x in e['v'].items() if k in PRIM}
        agree_f += wp == {k: x for k, x in f['v'].items() if k in PRIM}
    print(f'  wago ItemSparse (build 1.15.3.55646) primaries match Era on {agree_e} of {n} items, Forever on {agree_f}: it is the Era client, not Forever')

    # 1. ratio Forever / Era for items with stats on both sides, same ilvl and quality
    print('\n1. Value ratio Forever / Era, same item id, same ilvl and quality, both with stats '
          f'(weights {W}, p={P}):')
    print('   band        quality  n   unchanged  median p-norm ratio  mean  median linear ratio   changed items only: n, median p-norm ratio')
    for q in (2, 3):
        for lo, hi in ((5, 15), (16, 20), (21, 25), (26, 30), (5, 30)):
            sub = [(f, e) for f, e in pairs if f['q'] == e['q'] == q and f['ilvl'] == e['ilvl'] and lo <= f['ilvl'] <= hi and f['v'] and e['v']]
            if not sub:
                continue
            rp = [val(f['v']) / val(e['v']) for f, e in sub]
            rl = [lin(f['v']) / lin(e['v']) for f, e in sub]
            ch = [val(f['v']) / val(e['v']) for f, e in sub if f['v'] != e['v']]
            print(f'   ilvl {lo:2}-{hi:2}  q{q}  {len(sub):4}  {sum(f["v"] == e["v"] for f, e in sub):4}      {statistics.median(rp):.3f}               {statistics.mean(rp):.3f}  {statistics.median(rl):.3f}'
                  f'                  {len(ch):3}, {statistics.median(ch) if ch else float("nan"):.3f}')
    # items that moved ilvl or quality: value relative to a common curve
    moved = [(f, e) for f, e in pairs if (f['ilvl'] != e['ilvl'] or f['q'] != e['q']) and f['v'] and e['v']]
    print(f'   items that moved ilvl or quality: {len(moved)}; value ratio median {statistics.median([val(f["v"])/val(e["v"]) for f, e in moved]):.3f}, '
          f'ilvl change median {statistics.median([f["ilvl"]-e["ilvl"] for f, e in moved]):+.0f}')
    stripped = [(f, e) for f, e in pairs if e['v'] and not f['v']]
    gained = [(f, e) for f, e in pairs if f['v'] and not e['v']]
    print(f'   Era items with stats that have none in Forever: {len(stripped)}; stat-less Era items that gained stats: {len(gained)}')

    # 2. curves: fit both sets with the same slot mods (standards) and weights, V per quality
    print('\n2. Budget curves, same model fitted to each set separately (armor, slot mods and weights fixed):')
    FA = clean(F, ARMOR)
    EA = clean(era(30), ARMOR)
    fixw = {k: v for k, v in W.items() if k != 'prim'}
    res = {}
    for label, rows in (('Forever ilvl 5-30', FA), ('Era ilvl 5-30', [r for r in EA if r['ilvl'] >= 5]),
                        ('Forever ilvl 5-30, Classic ids', [r for r in FA if r['src'] == 'classic-id']),
                        ('Forever ilvl 5-30, Forever-new ids', [r for r in FA if r['src'] == 'forever-new'])):
        M = Model(rows, P, fix_w=fixw, fix_mods=STD_MOD)
        th, mask = M.robust()
        w, m, ab = M.unpack(th)
        res[label] = ab
        print(f'   {label}: n={len(rows)} ({mask.sum()} kept): ' + '; '.join(f'q{q} V = {a:.3f} ilvl {b:+.2f} (V(20) {a*20+b:.1f}, V(30) {a*30+b:.1f})' for q, (a, b) in ab.items()))
    print('   Forever / Era V ratio:')
    for q in (2, 3):
        fa, fb = res['Forever ilvl 5-30'][q]
        ea, eb = res['Era ilvl 5-30'][q]
        print(f'     q{q}: ' + ', '.join(f'ilvl {L}: {(fa*L+fb)/(ea*L+eb):.3f}' for L in (15, 20, 25, 30)))
    # items Forever moved in ilvl or quality: value / (slot mod x V) on Era's curves, before and after
    ea = res['Era ilvl 5-30']
    util = lambda r: val(r['v']) / (STD_MOD[r['slot']] * (ea[r['q']][0] * r['ilvl'] + ea[r['q']][1]))
    mv = [(f, e) for f, e in pairs if f['slot'] in ARMOR and (f['ilvl'] != e['ilvl'] or f['q'] != e['q']) and f['v'] and e['v'] and e['ilvl'] >= 10]
    uf = [util(f) for f, e in mv]; ue = [util(e) for f, e in mv]
    print(f'   armor items Forever moved in ilvl or quality (n={len(mv)}): value / budget on Era curves, Era version median {statistics.median(ue):.2f}, '
          f'Forever version median {statistics.median(uf):.2f} (1.00 = on budget)')
    print('   linear totals per slot mod:')
    lt_f = linear_totals([r for r in FA], {k: v for k, v in W.items()}, STD_MOD, 'Forever ilvl 5-30 (untrimmed)')
    lt_e = linear_totals([r for r in EA if r['ilvl'] >= 5], {k: v for k, v in W.items()}, STD_MOD, 'Era ilvl 5-30 (untrimmed)')
    for q in (2, 3):
        print(f'     q{q} Forever / Era linear: ' + ', '.join(f'ilvl {L}: {(lt_f[q][0]*L+lt_f[q][1])/(lt_e[q][0]*L+lt_e[q][1]):.3f}' for L in (15, 20, 25, 30)))

    # 3. swaps: one or more Era stats replaced, the rest kept
    print('\n3. Stat swaps (same ilvl and quality, Forever dropped stats D and added stats A, the rest unchanged):')
    kinds = collections.Counter()
    rates = collections.defaultdict(list)
    for f, e in pairs:
        if f['ilvl'] != e['ilvl'] or f['q'] != e['q'] or f['v'] == e['v']:
            continue
        common = {k for k in f['v'] if k in e['v'] and f['v'][k] == e['v'][k]}
        A = {k: x for k, x in f['v'].items() if k not in common}
        D = {k: x for k, x in e['v'].items() if k not in common}
        key = ('+'.join(sorted(map(kind_of, D))) or '-') + ' -> ' + ('+'.join(sorted(map(kind_of, A))) or '-')
        kinds[key] += 1
        # exchange rate for sp when A is only sp (+heal) and D is primaries only
        if f['slot'] in ARMOR and D and A and set(map(kind_of, A)) <= {'sp', 'heal'} and 'sp' in A and set(D) <= set(PRIM):
            keep = {k: x for k, x in e['v'].items() if k in common}
            target_p = val(e['v'])
            # solve w_sp so the Forever item keeps the Era item's p-norm value (heal fixed at W['heal'])
            def fval(ws):
                vv = sum((wt(k) * x) ** P for k, x in keep.items()) + (ws * A['sp']) ** P + ((W['heal'] * A.get('heal', 0)) ** P)
                return vv ** (1 / P)
            lo, hi = 0.0, 5.0
            for _ in range(60):
                mid = (lo + hi) / 2
                lo, hi = (mid, hi) if fval(mid) < target_p else (lo, mid)
            w_p = (lo + hi) / 2 if fval(0) < target_p else 0.0
            w_lin = (sum(D.values()) - W['heal'] * A.get('heal', 0)) / A['sp']
            rates['heal' in A].append((w_p, w_lin, e['name'], e['ilvl'], D, A))
    for k, n in kinds.most_common(15):
        print(f'   {n:4}  {k}')
    for has_heal, lst in sorted(rates.items()):
        wp = [x[0] for x in lst]; wl = [x[1] for x in lst]
        print(f'   armor, primaries -> spell power{" + extra healing (healing at 0.40)" if has_heal else ""}: n={len(lst)}, implied sp weight '
              f'p-norm median {statistics.median(wp):.2f} (p25-p75 {np.percentile(wp,25):.2f}-{np.percentile(wp,75):.2f}), '
              f'linear median {statistics.median(wl):.2f} (p25-p75 {np.percentile(wl,25):.2f}-{np.percentile(wl,75):.2f})')
        for x in lst[:6]:
            print(f'      {x[2]} (ilvl {x[3]}): {x[4]} -> {x[5]}: {x[0]:.2f} / {x[1]:.2f}')

    # single-term swaps (one primary amount d became sp amount a): the value is kept for any p iff
    # w_sp x a = d; with rounding to whole points, w_sp lies in [d/(a+0.5), d/(a-0.5)].
    print('\n   single-term armor swaps (one primary -> spell power, nothing else changed): w_sp consistent with rounding')
    iv = []
    for f, e in pairs:
        if f['slot'] not in ARMOR or f['ilvl'] != e['ilvl'] or f['q'] != e['q']:
            continue
        common = {k for k in f['v'] if k in e['v'] and f['v'][k] == e['v'][k]}
        A = {k: x for k, x in f['v'].items() if k not in common}
        D = {k: x for k, x in e['v'].items() if k not in common}
        if list(A) == ['sp'] and len(D) == 1 and set(D) <= set(PRIM):
            d, a = list(D.values())[0], A['sp']
            iv.append((d / (a + 0.5), d / (a - 0.5), d, a, e['name']))
    for cand in (0.55, 0.7, 0.8, 0.86, 0.93, 1.0, 1.1):
        n = sum(lo <= cand <= hi for lo, hi, *_ in iv)
        print(f'     w_sp = {cand}: consistent with {n} of {len(iv)} swaps')
    big = [x for x in iv if x[2] >= 4]
    print(f'     swaps of 4+ points ({len(big)}): ' + '; '.join(f'{x[4]} {x[2]}->{x[3]}' for x in big))
