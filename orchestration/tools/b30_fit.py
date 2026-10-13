"""Q2-Q4 of the ilvl <= 30 analysis: stat weights, slot modifiers and the budget per item level.

Model: an item's value (sum_k (w_k v_k)^p)^(1/p) = slot_mod x V_q(ilvl), V_q(ilvl) = a_q ilvl + b_q per
quality; primaries weigh 1, chest's slot mod is 1. Fitted by Levenberg-Marquardt on log residuals,
then refitted without items more than 3 robust sigmas off (procs the listing does not show, stale
rows). 90% intervals from 200 bootstrap resamples of items.

Usage: python3 b30_fit.py [p]   (default 1.5)
"""
import collections, math, os, statistics, sys
import numpy as np
sys.path.insert(0, os.path.dirname(__file__))
from b30_data import forever, era, clean, kind_of, ARMOR, WEAPON, PRIM, STD_MOD

INIT_W = {'sp': 0.86, 'heal': 0.45, 'school': 0.7, 'ap': 0.5, 'rap': 0.4, 'crit': 1.0, 'hit': 1.0, 'mp5': 2.5}


class Model:
    def __init__(self, rows, p, fix_w=None, fix_mods=None, quals=None):
        self.rows, self.p = rows, p
        self.fix_w = fix_w or {}
        self.fix_mods = fix_mods
        terms = collections.Counter(kind_of(k) for r in rows for k in r['v'])
        self.kinds = [k for k in INIT_W if terms.get(k) and k not in self.fix_w]
        self.n_items = {k: sum(1 for r in rows if any(kind_of(s) == k for s in r['v'])) for k in INIT_W}
        self.slots = sorted({r['slot'] for r in rows})
        self.ref = next(s for s in ('chest', 'two-hand') + tuple(self.slots) if s in self.slots)
        self.free_slots = [] if fix_mods else [s for s in self.slots if s != self.ref]
        self.quals = quals or sorted({r['q'] for r in rows})
        # dense matrix of terms: column per (item term); each term has a kind index
        allk = ['prim'] + list(INIT_W)
        self.cols = sorted({k for r in rows for k in r['v']})
        self.X = np.array([[r['v'].get(c, 0.0) for c in self.cols] for r in rows], float)
        self.colkind = [('prim' if c in PRIM else kind_of(c)) for c in self.cols]
        self.slot_idx = np.array([self.slots.index(r['slot']) for r in rows])
        self.q_idx = np.array([self.quals.index(r['q']) for r in rows])
        self.ilvl = np.array([r['ilvl'] for r in rows], float)

    def unpack(self, th):
        i = 0
        w = {'prim': 1.0}
        w.update({k: v for k, v in INIT_W.items()})
        w.update(self.fix_w)
        for k in self.kinds:
            w[k] = math.exp(th[i]); i += 1
        m = {s: 1.0 for s in self.slots}
        if self.fix_mods:
            m.update({s: self.fix_mods[s] for s in self.slots})
        for s in self.free_slots:
            m[s] = math.exp(th[i]); i += 1
        ab = {}
        for q in self.quals:
            ab[q] = (th[i], th[i + 1]); i += 2
        return w, m, ab

    def theta0(self):
        th = [math.log(INIT_W[k]) for k in self.kinds]
        th += [math.log(STD_MOD.get(s, {'two-hand': 1.0, 'one-hand': .42, 'main-hand': .42, 'off-hand weapon': .42,
                                         'shield': .5625, 'held off-hand': .5625, 'ranged': .32, 'thrown': .32}.get(s, .5)))
               for s in self.free_slots]
        for q in self.quals:
            th += [0.8, 0.0]
        return np.array(th, float)

    def cost(self, w):
        wv = np.array([w[k] for k in self.colkind])
        return ((self.X * wv) ** self.p).sum(1) ** (1 / self.p)

    def resid(self, th):
        w, m, ab = self.unpack(th)
        A = np.array([ab[q][0] for q in self.quals])[self.q_idx]
        B = np.array([ab[q][1] for q in self.quals])[self.q_idx]
        V = np.maximum(A * self.ilvl + B, 1e-3)
        mm = np.array([m[s] for s in self.slots])[self.slot_idx]
        return np.log(self.cost(w)) - np.log(mm * V)

    def fit(self, th=None, mask=None, iters=60):
        th = self.theta0() if th is None else th.copy()
        mask = np.ones(len(self.rows), bool) if mask is None else mask
        lam = 1e-2
        r = self.resid(th)[mask]
        for _ in range(iters):
            J = np.empty((mask.sum(), len(th)))
            for j in range(len(th)):
                d = np.zeros_like(th); d[j] = 1e-5
                J[:, j] = (self.resid(th + d)[mask] - r) / 1e-5
            g = J.T @ r
            H = J.T @ J
            while True:
                step = np.linalg.solve(H + lam * np.diag(np.diag(H) + 1e-9), -g)
                r2 = self.resid(th + step)[mask]
                if (r2 ** 2).sum() < (r ** 2).sum():
                    th, r, lam = th + step, r2, lam / 3
                    break
                lam *= 4
                if lam > 1e8:
                    return th
            if np.abs(step).max() < 1e-7:
                break
        return th

    def robust(self):
        th = self.fit()
        r = self.resid(th)
        mad = 1.4826 * np.median(np.abs(r - np.median(r)))
        mask = np.abs(r) < 3 * mad
        th = self.fit(th, mask)
        return th, mask


def summarize(label, rows, p, boot=200, fix_w=None, fix_mods=None, seed=1):
    M = Model(rows, p, fix_w, fix_mods)
    th, mask = M.robust()
    w, m, ab = M.unpack(th)
    r = M.resid(th)
    rng = np.random.default_rng(seed)
    bs = []
    idx_all = np.arange(len(rows))
    for _ in range(boot):
        idx = rng.choice(idx_all, len(rows))
        mk = np.zeros(len(rows), int)
        np.add.at(mk, idx, 1)
        # weighted resample: replicate rows by repeating their residual (sqrt weight trick)
        sw = np.sqrt(mk * mask)
        def res_w(t, M=M, sw=sw):
            return M.resid(t) * sw
        thb = th.copy()
        lam = 1e-2
        rb = res_w(thb)
        for _ in range(15):
            J = np.empty((len(rows), len(thb)))
            for j in range(len(thb)):
                d = np.zeros_like(thb); d[j] = 1e-5
                J[:, j] = (res_w(thb + d) - rb) / 1e-5
            step = np.linalg.solve(J.T @ J + lam * np.diag(np.diag(J.T @ J) + 1e-9), -J.T @ rb)
            r2 = res_w(thb + step)
            if (r2 ** 2).sum() < (rb ** 2).sum():
                thb, rb, lam = thb + step, r2, lam / 3
            else:
                lam *= 4
        bs.append(M.unpack(thb))
    ci = lambda vals: (np.percentile(vals, 5), np.percentile(vals, 95))
    print(f'\n== {label}, p={p}: {len(rows)} items, {mask.sum()} kept after the 3-sigma trim, rms log error {100*np.sqrt(np.mean(r[mask]**2)):.1f}%')
    print('  weights (relative to a primary point; n = items carrying the stat, kept):')
    for k in INIT_W:
        if k in M.kinds:
            n = sum(1 for i, rr in enumerate(rows) if mask[i] and any(kind_of(s) == k for s in rr['v']))
            lo, hi = ci([b[0][k] for b in bs])
            band = f'90% {lo:.2f}-{hi:.2f}' if n >= 10 else 'too few items for an interval'
            print(f'    {k:6} {w[k]:.2f}  ({band}, n={n})')
    print(f'  slot mods ({M.ref} 1):')
    for s in M.slots:
        if s in M.free_slots:
            lo, hi = ci([b[1][s] for b in bs])
            n = sum(1 for i, rr in enumerate(rows) if mask[i] and rr['slot'] == s)
            print(f'    {s:16} {m[s]:.3f}  (90% {lo:.3f}-{hi:.3f}, n={n})')
    for q in M.quals:
        a, b = ab[q]
        n = sum(1 for i, rr in enumerate(rows) if mask[i] and rr['q'] == q)
        print(f'  quality {q}: V(ilvl) = {a:.3f} ilvl {b:+.2f}  (n={n}); V(20)={a*20+b:.1f} V(25)={a*25+b:.1f} V(30)={a*30+b:.1f}')
    return M, th, mask, w, m, ab


def linear_totals(rows, w, mods, label):
    """Q4: weighted linear stat total per slot mod, fitted against ilvl per quality."""
    print(f'\n== linear totals, {label} (weights {", ".join(f"{k} {w[k]:.2f}" for k in INIT_W)})')
    out = {}
    for q in sorted({r['q'] for r in rows}):
        rq = [r for r in rows if r['q'] == q]
        if len(rq) < 8:
            continue
        lin = np.array([sum((1.0 if k in PRIM else w[kind_of(k)]) * x for k, x in r['v'].items()) / mods[r['slot']] for r in rq])
        X = np.array([[r['ilvl'], 1.0] for r in rq])
        a, b = np.linalg.lstsq(X, lin, rcond=None)[0]
        big = X[:, 0] >= 15
        res = (lin / (X @ [a, b]) - 1)[big]
        nst = statistics.mean(len(r['v']) for r in rq)
        out[q] = (a, b)
        print(f'  quality {q}: linear total / slot mod = {a:.3f} ilvl {b:+.2f}  (n={len(rq)}, rms {100*np.sqrt(np.mean(res**2)):.0f}% at ilvl 15-30, '
              f'{nst:.2f} stats per item); at ilvl 20: {a*20+b:.1f}, 25: {a*25+b:.1f}, 30: {a*30+b:.1f}')
    return out


if __name__ == '__main__':
    p = float(sys.argv[1]) if len(sys.argv) > 1 else 1.5
    F = forever()
    FA = clean(F, ARMOR)
    print('Forever ilvl 5-30 clean armor items:', len(FA), collections.Counter(r['src'] for r in FA))
    # exponent check with every stat in play
    for pp in (1.0, 1.25, 1.5, 1.75, 2.0):
        M = Model(FA, pp)
        th, mask = M.robust()
        r = M.resid(th)
        print(f'  all-stat fit p={pp}: rms log error {100*np.sqrt(np.mean(r**2)):.2f}% (all items), {100*np.sqrt(np.mean(r[mask]**2)):.2f}% (trimmed, {mask.sum()} kept)')
    M, th, mask, w, m, ab = summarize('Forever armor ilvl 5-30', FA, p)
    summarize('Forever armor ilvl 5-30, Classic item ids only', [r for r in FA if r['src'] == 'classic-id'], p, boot=100)
    summarize('Forever armor ilvl 5-30, Forever-new item ids only', [r for r in FA if r['src'] == 'forever-new'], p, boot=100)
    summarize('Forever armor ilvl 15-30', [r for r in FA if r['ilvl'] >= 15], p, boot=100)
    FW = clean(F, ARMOR + WEAPON + ['trinket'])
    summarize('Forever armor + weapons + trinkets ilvl 5-30', FW, p, boot=100)
    summarize('Forever weapons, shields, off-hands ilvl 5-30', [r for r in clean(F, WEAPON) if r['slot'] != 'thrown'], p, boot=100)
    # linear totals with the fitted weights and the standards' slot mods / the fitted ones
    kept = [r for i, r in enumerate(FA) if mask[i]]
    linear_totals(kept, w, STD_MOD, 'Forever armor ilvl 5-30, standards slot mods')
    linear_totals(kept, w, m, 'Forever armor ilvl 5-30, fitted slot mods')
    lin = np.array([sum((1.0 if k in PRIM else w[kind_of(k)]) * x for k, x in r['v'].items()) for r in kept])
    val = M.cost(w)[mask]
    ratio = lin / val
    print(f'\n  linear total / {p}-norm value per item: median {np.median(ratio):.3f}, mean {ratio.mean():.3f}; by stat count: ' +
          ', '.join(f'{n}: {np.median(ratio[[len(r["v"]) == n for r in kept]]):.2f} (n={sum(len(r["v"]) == n for r in kept)})'
                    for n in sorted({len(r['v']) for r in kept})))
