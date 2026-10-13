"""p-norm budget model for the level 60 BiS analysis (orchestration/bis_budget_rules.md).

value(item) = (sum_k (w_kind(k) x amount_k)^p)^(1/p) = slot_mod(slot) x V_g(ilvl) x factor(f)
  - w: stat weights (primaries 1); any kind can be fixed or fitted;
  - slot_mod: chest 1 (or the first slot present), fixed or fitted;
  - V_g(ilvl) = a_g ilvl + b_g per group g (usually quality);
  - factor(f): optional multiplicative factor per class f (archetype dummies), first class fixed at 1.
Fitted by Levenberg-Marquardt on log residuals, then refitted without items more than 3 robust sigmas off.
"""
import math
import numpy as np
from bis_data import kind

P = 1.5

# Measured weights (forever_item_budget.md): spell power 0.86 (Forever ilvl <= 30; Era level 60 0.84), the rest
# from the Era level 60 fit. Tank stats and SoD-only stats are fitted here (bis_archetype.py prints them).
W_MEASURED = {'prim': 1.0, 'sp': 0.86, 'heal': 0.45, 'school': 0.71, 'ap': 0.52, 'rap': 0.40, 'fap': 0.30,
              'crit': 1.01, 'hit': 1.09, 'spcrit': 1.01, 'sphit': 1.09, 'mp5': 2.39,
              'def': 1.2, 'dodge': 12.0, 'parry': 15.0, 'block': 5.0, 'blockv': 0.15,
              'haste': 14.0, 'mhaste': 14.0, 'sphaste': 14.0, 'exp': 1.0}

STD_SLOT = {'head': 1, 'chest': 1, 'legs': 1, 'shoulder': .75, 'hands': .75, 'waist': .75, 'feet': .75,
            'neck': .5625, 'back': .5625, 'wrist': .5625, 'finger': .5625, 'trinket': .5625,
            'two-hand': 1.0, 'one-hand': .42, 'main-hand': .42, 'off-hand weapon': .42, 'held off-hand': .5625,
            'shield': .5625, 'ranged': .32, 'wand': .32, 'thrown': .32, 'relic': .32}


def wt(k, w):
    return w[kind(k)]


def value(v, w=W_MEASURED, p=P):
    return sum((wt(k, w) * x) ** p for k, x in v.items()) ** (1 / p)


def linear(v, w=W_MEASURED):
    return sum(wt(k, w) * x for k, x in v.items())


class PModel:
    def __init__(self, rows, p=P, w=None, free_w=(), slot_mods=None, slot_key=lambda r: r['slot'],
                 group=lambda r: r['q'], factor=None, factor_ref=None):
        self.rows, self.p = rows, p
        self.w0 = dict(W_MEASURED if w is None else w)
        self.free_w = [k for k in free_w if any(kind(s) == k for r in rows for s in r['v'])]
        self.slot_key = slot_key
        self.slots = sorted({slot_key(r) for r in rows})
        self.ref = next((s for s in ('chest', 'two-hand') if s in self.slots), self.slots[0])
        self.fixed_slots = dict(slot_mods or {})
        self.free_slots = [s for s in self.slots if s != self.ref and s not in self.fixed_slots]
        self.groups = sorted({group(r) for r in rows}, key=str)
        self.g_idx = np.array([self.groups.index(group(r)) for r in rows])
        self.factor = factor
        if factor:
            fs = sorted({factor(r) for r in rows}, key=str)
            self.fref = factor_ref if factor_ref in fs else fs[0]
            self.fclasses = [f for f in fs if f != self.fref]
            self.f_idx = np.array([(self.fclasses.index(factor(r)) if factor(r) != self.fref else -1) for r in rows])
        else:
            self.fclasses = []
        self.cols = sorted({k for r in rows for k in r['v']})
        self.colkind = [kind(c) for c in self.cols]
        self.X = np.array([[r['v'].get(c, 0.0) for c in self.cols] for r in rows], float)
        self.s_idx = np.array([self.slots.index(slot_key(r)) for r in rows])
        self.ilvl = np.array([r['ilvl'] for r in rows], float)

    def unpack(self, th):
        i = 0
        w = dict(self.w0)
        for k in self.free_w:
            w[k] = math.exp(th[i]); i += 1
        m = {s: self.fixed_slots.get(s, 1.0) for s in self.slots}
        for s in self.free_slots:
            m[s] = math.exp(th[i]); i += 1
        ab = {}
        for g in self.groups:
            ab[g] = (th[i], th[i + 1]); i += 2
        f = {}
        if self.factor:
            f[self.fref] = 1.0
            for c in self.fclasses:
                f[c] = math.exp(th[i]); i += 1
        return w, m, ab, f

    def theta0(self):
        th = [math.log(self.w0[k]) for k in self.free_w]
        th += [math.log(STD_SLOT.get(s, .6)) for s in self.free_slots]
        for g in self.groups:
            th += [0.73, 3.0]
        th += [0.0] * len(self.fclasses)
        return np.array(th, float)

    def cost(self, w):
        wv = np.array([w[k] for k in self.colkind])
        return ((self.X * wv) ** self.p).sum(1) ** (1 / self.p)

    def lin(self, w):
        wv = np.array([w[k] for k in self.colkind])
        return (self.X * wv).sum(1)

    def budget(self, th):
        w, m, ab, f = self.unpack(th)
        A = np.array([ab[g][0] for g in self.groups])[self.g_idx]
        B = np.array([ab[g][1] for g in self.groups])[self.g_idx]
        V = np.maximum(A * self.ilvl + B, 1e-3)
        mm = np.array([m[s] for s in self.slots])[self.s_idx]
        out = mm * V
        if self.factor:
            fv = np.array([f[c] for c in self.fclasses] + [1.0])
            out = out * fv[self.f_idx]
        return out

    def resid(self, th):
        w = self.unpack(th)[0]
        return np.log(self.cost(w)) - np.log(self.budget(th))

    def fit(self, th=None, mask=None, iters=80):
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

    def boot(self, th, mask, n=100, seed=1):
        """Bootstrap the fitted parameters (resampling items, a few LM steps from th)."""
        rng = np.random.default_rng(seed)
        out = []
        N = len(self.rows)
        for _ in range(n):
            mk = np.zeros(N)
            np.add.at(mk, rng.choice(N, N), 1)
            sw = np.sqrt(mk * mask)
            tb = th.copy()
            lam = 1e-2
            rb = self.resid(tb) * sw
            for _ in range(12):
                J = np.empty((N, len(tb)))
                for j in range(len(tb)):
                    d = np.zeros_like(tb); d[j] = 1e-5
                    J[:, j] = (self.resid(tb + d) * sw - rb) / 1e-5
                H = J.T @ J
                step = np.linalg.solve(H + lam * np.diag(np.diag(H) + 1e-9), -J.T @ rb)
                r2 = self.resid(tb + step) * sw
                if (r2 ** 2).sum() < (rb ** 2).sum():
                    tb, rb, lam = tb + step, r2, lam / 3
                else:
                    lam *= 4
            out.append(self.unpack(tb))
        return out


def ci(vals, lo=5, hi=95):
    return np.percentile(vals, lo), np.percentile(vals, hi)


def med(x):
    return float(np.median(x)) if len(x) else float('nan')
