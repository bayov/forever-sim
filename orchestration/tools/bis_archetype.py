"""Q1 and Q2 of the level 60 BiS budget analysis (orchestration/bis_budget_rules.md).

Q1. Do caster, physical, hybrid and tank pieces sit on the same p = 1.5 budget line?
  A. Era level 60 armor (blues ilvl 45-70 + epics 55-80, no effects): free fit of every weight (tank stats and
     spell crit / hit separately included) -> the measured weights.
  B. Weights fixed at the measured set (spell power 0.86), slot mods and V per quality fitted: residual of each
     epic by archetype and slot tier. Sensitivity to the spell power weight.
  C. Same, with an archetype factor in the fit and the weights free: does the factor survive?
  D. The weighted linear total against one common linear trend (what standards 1a prices).
  E. Season of Discovery level 60 epics (phase 4-6; 7-8 lack tooltips), same steps on SoD's own V line.
Q2. Stats per piece and budget split by archetype and slot tier: Era Phase 1 epics (Molten Core, Onyxia,
  Azuregos, Kazzak), Era pre-raid dungeon blues and epics at ilvl 60-66, all Era epics 55-80, SoD phase 4-6 epics.

Usage: python3 bis_archetype.py
"""
import collections, os, sys
import numpy as np
sys.path.insert(0, os.path.dirname(__file__))
from bis_data import era60, sod60, clean, kind, ARMOR, TIERS, TIER
from bis_model import PModel, W_MEASURED, value, linear, ci, med, P

ARCH = ['caster', 'physical', 'hybrid', 'tank']
SUBS = ['caster dps', 'healer', 'caster int/spi', 'caster mp5', 'physical', 'physical+int', 'hybrid', 'tank']
FREE = ['sp', 'heal', 'school', 'ap', 'rap', 'crit', 'hit', 'spcrit', 'sphit', 'mp5', 'def', 'dodge', 'parry', 'block', 'blockv',
        'fap', 'haste', 'mhaste', 'sphaste', 'exp']


def nstats(v):
    """Stat lines on the piece: healing on top of spell power is one tooltip line, AP+RAP one."""
    ks = set(v)
    n = len(ks)
    if 'sp' in ks and 'heal' in ks:
        n -= 1
    return n


def split(v, w=W_MEASURED):
    g = collections.Counter()
    for k, x in v.items():
        kk = kind(k)
        grp = ('primary' if k in ('str', 'agi', 'int', 'spi') else 'Stamina' if k == 'sta' else
               'spell power' if kk in ('sp', 'heal', 'school') else 'hit' if kk in ('hit', 'sphit') else
               'crit' if kk in ('crit', 'spcrit') else 'AP' if kk in ('ap', 'rap', 'fap') else 'MP5' if kk == 'mp5' else
               'tank' if kk in ('def', 'dodge', 'parry', 'block', 'blockv') else 'haste/exp')
        g[grp] += w[kk] * x
    return g


SPLIT_COLS = ['primary', 'Stamina', 'spell power', 'hit', 'crit', 'AP', 'MP5', 'tank', 'haste/exp']


def fitted(rows, w, free_w=(), factor=None, label='', boot=0, verbose=True, p=P):
    M = PModel(rows, p, w=w, free_w=free_w, factor=factor, factor_ref='physical' if factor else None)
    th, mask = M.robust()
    r = M.resid(th)
    wf, m, ab, f = M.unpack(th)
    if verbose:
        print(f'\n== {label}: {len(rows)} items, {mask.sum()} kept, rms log error {100*np.sqrt(np.mean(r[mask]**2)):.1f}%')
    bs = M.boot(th, mask, boot) if boot else []
    return M, th, mask, r, wf, m, ab, f, bs


def resid_table(rows, r, mask, key, cols, title, q=4):
    """median residual (%, value / budget - 1) per key x tier, epics only, with n and the standard error."""
    print(f'\n{title}')
    print('| archetype | ' + ' | '.join(cols) + ' | all armor |')
    print('|---|' + '---|' * (len(cols) + 1))
    for a in key:
        cells = []
        for c in cols + ['all']:
            x = np.array([r[i] for i, rr in enumerate(rows) if mask[i] and rr['q'] == q and (rr['sub'] if a in SUBS and a not in ARCH else rr['arch']) == a
                          and (c == 'all' or rr['tier'] == c)])
            if len(x) == 0:
                cells.append('-')
                continue
            se = x.std() / np.sqrt(len(x)) if len(x) > 1 else float('nan')
            cells.append(f'{100*(np.exp(np.median(x))-1):+.1f}% (mean {100*(np.exp(x.mean())-1):+.1f} +/- {100*se:.1f}, n={len(x)})')
        print(f'| {a} | ' + ' | '.join(cells) + ' |')


def sub_resid(rows, r, mask, q=4, title=''):
    print(f'\n{title}')
    print('| sub-archetype | median | mean +/- se | n | stats per piece |')
    print('|---|---|---|---|---|')
    for s in SUBS:
        idx = [i for i, rr in enumerate(rows) if mask[i] and rr['q'] == q and rr['sub'] == s]
        if not idx:
            continue
        x = r[idx]
        print(f'| {s} | {100*(np.exp(np.median(x))-1):+.1f}% | {100*(np.exp(x.mean())-1):+.1f} +/- {100*x.std()/np.sqrt(len(x)):.1f} | {len(x)} | '
              f'{np.mean([nstats(rows[i]["v"]) for i in idx]):.1f} |')


def linear_trend(rows, w, mods):
    X = np.array([[rr['ilvl'], 1.0] for rr in rows])
    y = np.array([linear(rr['v'], w) / mods[rr['slot']] for rr in rows])
    a, b = np.linalg.lstsq(X, y, rcond=None)[0]
    return a, b, y / (X @ [a, b])


def q2(rows, label):
    print(f'\n### Q2 {label}: {len(rows)} pieces (all pieces, effects included; priced stats only)')
    print('| archetype | tier | n | mean stats | ' + ' | '.join(f'{k} stats' for k in range(1, 8)) + ' |')
    print('|---|---|---|---|' + '---|' * 7)
    for a in SUBS:
        for t in TIERS + ['all armor']:
            x = [nstats(rr['v']) for rr in rows if rr['sub'] == a and (rr['tier'] == t or (t == 'all armor' and rr['tier'] in TIERS))]
            if len(x) < 3:
                continue
            c = collections.Counter(x)
            print(f'| {a} | {t} | {len(x)} | {np.mean(x):.1f} | ' + ' | '.join(f'{100*c.get(k,0)/len(x):.0f}%' for k in range(1, 8)) + ' |')
    print(f'\nbudget split (weighted linear shares, measured weights), armor pieces of {label}:')
    print('| archetype | n | ' + ' | '.join(SPLIT_COLS) + ' |')
    print('|---|---|' + '---|' * len(SPLIT_COLS))
    for a in SUBS:
        rs = [rr for rr in rows if rr['sub'] == a and rr['tier'] in TIERS]
        if len(rs) < 3:
            continue
        tot = collections.Counter()
        for rr in rs:
            tot.update(split(rr['v']))
        s = sum(tot.values())
        print(f'| {a} | {len(rs)} | ' + ' | '.join(f'{100*tot[c]/s:.1f}' for c in SPLIT_COLS) + ' |')


if __name__ == '__main__':
    E = era60()
    EA = clean(E, ARMOR)
    lvl60 = [r for r in EA if (r['q'] == 3 and 45 <= r['ilvl'] <= 70) or (r['q'] == 4 and 55 <= r['ilvl'] <= 80)]
    lvl60 = [r for r in lvl60 if r['arch'] != 'stamina']
    print(f'Era clean armor, blues 45-70 + epics 55-80: {len(lvl60)} ({collections.Counter(r["arch"] for r in lvl60 if r["q"] == 4)} epics)')

    # A. free weights
    M, th, mask, r, wA, m, ab, f, bs = fitted(lvl60, W_MEASURED, FREE, label='A. Era level 60 armor, every weight free', boot=60)
    print('| stat | weight | 90% interval | items |')
    print('|---|---|---|---|')
    for k in FREE:
        if k in M.free_w:
            lo, hi = ci([b[0][k] for b in bs])
            n = sum(1 for i, rr in enumerate(lvl60) if mask[i] and any(kind(s) == k for s in rr['v']))
            print(f'| {k} | {wA[k]:.2f} | {lo:.2f}-{hi:.2f} | {n} |')
    print('slot mods:', {s: round(x, 3) for s, x in m.items()})
    print('V:', {q: f'{a:.3f} ilvl {b:+.2f}' for q, (a, b) in ab.items()})

    # B. weights fixed at the measured set; tank stats at their free-fit values
    WM = dict(W_MEASURED)
    for k in ('def', 'dodge', 'parry', 'block', 'blockv', 'fap'):
        if k in wA:
            WM[k] = wA[k]
    print('\nmeasured weights used from here on:', {k: round(x, 3) for k, x in WM.items() if k not in ('haste', 'mhaste', 'sphaste', 'exp')})
    for pB in (1.7, 1.5):
        M, th, mask, r, wB, mB, abB, f, _ = fitted(lvl60, WM, (), label=f'B. Era level 60 armor, weights fixed (sp 0.86), p {pB}', p=pB)
        print('slot mods:', {s_: round(x, 3) for s_, x in mB.items()})
        print('V:', {q: f'{a:.3f} ilvl {b:+.2f} (66: {a*66+b:.1f}, 71: {a*71+b:.1f}, 76: {a*76+b:.1f})' for q, (a, b) in abB.items()})
        resid_table(lvl60, r, mask, ARCH, TIERS, f'B1. Era epics, p {pB}: p-norm value / budget - 1, by archetype and slot tier (median; mean +/- standard error)')
        sub_resid(lvl60, r, mask, title=f'B2. Era epics by sub-archetype, p {pB}')
        byn = {n: [r[i] for i, rr in enumerate(lvl60) if mask[i] and nstats(rr['v']) == n] for n in range(2, 8)}
        print('   residual by stat count (blues and epics): ' + ', '.join(f'{n} stats {100*np.median(x):+.1f}% (n={len(x)})' for n, x in byn.items() if x))
    # p 1.5 is the last fit, used below
    p1 = [i for i, rr in enumerate(lvl60) if rr['p1'] in ('Molten Core', 'Onyxia', 'world boss') and mask[i]]
    print('\nB4. Era Phase 1 raid epics (MC, Onyxia, world bosses) on the same model:')
    for a in ARCH:
        x = r[[i for i in p1 if lvl60[i]['arch'] == a]]
        if len(x):
            print(f'   {a}: median {100*(np.exp(np.median(x))-1):+.1f}%, mean {100*(np.exp(x.mean())-1):+.1f}% +/- {100*x.std()/np.sqrt(len(x)):.1f}, n={len(x)}')
    pr = [i for i, rr in enumerate(lvl60) if rr['p1'] == 'pre-raid dungeon' and mask[i]]
    for a in ARCH:
        x = r[[i for i in pr if lvl60[i]['arch'] == a]]
        if len(x):
            print(f'   pre-raid dungeon (blue + epic) {a}: median {100*(np.exp(np.median(x))-1):+.1f}%, n={len(x)}')

    print('\nB5. sensitivity: median epic residual by archetype as the spell power weight moves')
    for sp in (0.55, 0.70, 0.80, 0.84, 0.86, 0.93, 1.0):
        W2 = dict(WM, sp=sp, school=0.71 * sp / 0.86)
        M2, th2, mask2, r2, *_ = fitted(lvl60, W2, (), verbose=False)
        out = []
        for a in ARCH:
            x = r2[[i for i, rr in enumerate(lvl60) if mask2[i] and rr['q'] == 4 and rr['arch'] == a]]
            out.append(f'{a} {100*(np.exp(np.median(x))-1):+.1f}%')
        cd = r2[[i for i, rr in enumerate(lvl60) if mask2[i] and rr['q'] == 4 and rr['sub'] == 'caster dps']]
        print(f'   sp {sp:.2f}: rms {100*np.sqrt(np.mean(r2[mask2]**2)):.2f}%; ' + ', '.join(out) + f'; caster dps {100*(np.exp(np.median(cd))-1):+.1f}%')

    # C0. exponent scan: every weight free, with archetype factors
    print('\nC0. exponent scan, Era level 60 armor, every weight free (rms after trim; archetype factors vs physical from a second fit)')
    for pp in (1.5, 1.6, 1.7, 1.75, 1.8, 2.0):
        Mx, thx, mkx, rx, wx, mx, abx, fx, _ = fitted(lvl60, WM, FREE, verbose=False, p=pp)
        Mf, thf, mkf, rf, wf, mf, abf, ff, _ = fitted(lvl60, WM, FREE, factor=lambda rr: rr['arch'], verbose=False, p=pp)
        print(f'   p {pp}: rms {100*np.sqrt(np.mean(rx[mkx]**2)):.2f}% ({mkx.sum()} kept); sp {wx["sp"]:.2f}, hit {wx["hit"]:.2f}, sphit {wx["sphit"]:.2f}, '
              f'crit {wx["crit"]:.2f}, spcrit {wx["spcrit"]:.2f}, ap {wx["ap"]:.2f}, mp5 {wx["mp5"]:.2f}; epic V {abx[4][0]:.3f} ilvl {abx[4][1]:+.2f}; '
              'factors ' + ', '.join(f'{a} {ff[a]:.3f}' for a in ARCH if a != 'physical'))

    # C. archetype factor in the fit, weights free
    for lab, free, pC in (('C1. archetype factors, weights fixed (sp 0.86), p 1.5', (), 1.5), ('C2. archetype factors, every weight free, p 1.5', FREE, 1.5),
                          ('C3. archetype factors, weights fixed (sp 0.86), p 1.7', (), 1.7)):
        M, th, mask, r, wC, mC, abC, fC, bs = fitted(lvl60, WM, free, factor=lambda rr: rr['arch'], label=lab, boot=40, p=pC)
        print('   factor vs physical: ' + ', '.join(f'{a} {fC[a]:.3f} (90% {ci([b[3][a] for b in bs])[0]:.3f}-{ci([b[3][a] for b in bs])[1]:.3f})' for a in ARCH if a in fC))
        if free:
            print('   weights: ' + ', '.join(f'{k} {wC[k]:.2f}' for k in M.free_w))

    # D. linear totals against one common linear trend (epics only)
    ep = [rr for i, rr in enumerate(lvl60) if rr['q'] == 4]
    a, b, ratio = linear_trend(ep, WM, mB)
    print(f'\nD. Era epics, weighted linear total / fitted slot mod: trend {a:.3f} ilvl {b:+.2f}; median ratio to the trend by archetype and tier:')
    print('| archetype | ' + ' | '.join(TIERS) + ' | all | stats per piece | linear / p-norm value |')
    print('|---|---|---|---|---|---|---|')
    for ar in ARCH:
        cells = []
        for t in TIERS + ['all']:
            x = [ratio[i] for i, rr in enumerate(ep) if rr['arch'] == ar and (t == 'all' or rr['tier'] == t)]
            cells.append(f'{med(x):.3f} (n={len(x)})' if x else '-')
        rr_ = [rr for rr in ep if rr['arch'] == ar]
        print(f'| {ar} | ' + ' | '.join(cells) + f' | {np.mean([nstats(x["v"]) for x in rr_]):.1f} | {med([linear(x["v"], WM)/value(x["v"], WM) for x in rr_]):.3f} |')

    # E. SoD phase 4-6 epics
    S = sod60()
    SA = [rr for rr in clean(S, ARMOR) if rr['phase'] <= 6 and rr['arch'] != 'stamina']
    print(f'\n\nE. SoD phase 4-6 clean armor epics: {len(SA)} {collections.Counter(rr["arch"] for rr in SA)}; by phase {collections.Counter(rr["phase"] for rr in SA)}')
    for pp in (1.5, 1.7):
        Mx, thx, mkx, rx, wx, mx, abx, fx, _ = fitted(SA, WM, FREE, verbose=False, p=pp)
        Mf = PModel(SA, pp, w=WM, free_w=FREE, group=lambda rr: rr['phase'], factor=lambda rr: rr['arch'], factor_ref='physical')
        thf, mkf = Mf.robust()
        rf = Mf.resid(thf)
        bsf = Mf.boot(thf, mkf, 30)
        wf, mf, abf, ff = Mf.unpack(thf)
        print(f'   SoD p {pp}, V per phase, weights free, archetype factors: rms {100*np.sqrt(np.mean(rf[mkf]**2)):.1f}%; sp {wf["sp"]:.2f}, crit {wf["crit"]:.2f}, hit {wf["hit"]:.2f}; factors ' +
              ', '.join(f'{a} {ff[a]:.3f} ({ci([b_[3][a] for b_ in bsf])[0]:.3f}-{ci([b_[3][a] for b_ in bsf])[1]:.3f})' for a in ARCH if a != 'physical' and a in ff))
    M, th, mask, r, wS, mS, abS, fS, bs = fitted(SA, WM, FREE, label='E1. SoD armor, every weight free', boot=40)
    print('| stat | weight | 90% interval | items |')
    print('|---|---|---|---|')
    for k in M.free_w:
        lo, hi = ci([b_[0][k] for b_ in bs])
        n = sum(1 for i, rr in enumerate(SA) if mask[i] and any(kind(s) == k for s in rr['v']))
        print(f'| {k} | {wS[k]:.2f} | {lo:.2f}-{hi:.2f} | {n} |')
    print('slot mods:', {s: round(x, 3) for s, x in mS.items()})
    print('V:', {q: f'{a_:.3f} ilvl {b_:+.2f}' for q, (a_, b_) in abS.items()})
    WS = dict(WM)
    for k in ('haste', 'exp', 'def', 'dodge', 'parry', 'block', 'blockv', 'fap', 'heal'):
        if k in wS:
            WS[k] = wS[k]
    M, th, mask, r, _, mS2, abS2, _, _ = fitted(SA, WS, (), label='E2. SoD armor, Era measured weights (sp 0.86; SoD-only stats and healing at the E1 values)')
    print('V:', {q: f'{a_:.3f} ilvl {b_:+.2f} (66: {a_*66+b_:.1f}, 71: {a_*71+b_:.1f}, 76: {a_*76+b_:.1f})' for q, (a_, b_) in abS2.items()})
    resid_table(SA, r, mask, ARCH, TIERS, 'E3. SoD epics: value / budget - 1 by archetype and tier')
    sub_resid(SA, r, mask, title='E4. SoD epics by sub-archetype')
    for ph in (4, 5, 6):
        out = []
        for a_ in ARCH:
            x = r[[i for i, rr in enumerate(SA) if mask[i] and rr['phase'] == ph and rr['arch'] == a_]]
            if len(x):
                out.append(f'{a_} {100*(np.exp(np.median(x))-1):+.1f}% (n={len(x)})')
        print(f'   phase {ph}: ' + ', '.join(out))
    M, th, mask, r, wC, mC, abC, fC, bs = fitted(SA, WS, (), factor=lambda rr: rr['arch'], label='E5. SoD, archetype factors, weights fixed', boot=40)
    print('   factor vs physical: ' + ', '.join(f'{a_} {fC[a_]:.3f} (90% {ci([b_[3][a_] for b_ in bs])[0]:.3f}-{ci([b_[3][a_] for b_ in bs])[1]:.3f})' for a_ in ARCH if a_ in fC))
    a, b, ratio = linear_trend(SA, WS, mS2)
    print(f'   SoD linear trend {a:.3f} ilvl {b:+.2f}; median linear ratio: ' + ', '.join(
        f'{ar} {med([ratio[i] for i, rr in enumerate(SA) if rr["arch"] == ar]):.3f}' for ar in ARCH) +
        '; stats per piece: ' + ', '.join(f'{ar} {np.mean([nstats(rr["v"]) for rr in SA if rr["arch"] == ar]):.1f}' for ar in ARCH))

    # Q2
    allE = [rr for rr in E if not rr['other'] or True]
    q2([rr for rr in allE if rr['q'] == 4 and rr['p1'] in ('Molten Core', 'Onyxia', 'world boss') and rr['v']], 'Era Phase 1 raid epics (MC, Onyxia, Azuregos, Kazzak)')
    q2([rr for rr in allE if rr['p1'] == 'pre-raid dungeon' and rr['v']], 'Era pre-raid dungeon items ilvl 60-66 (BRD, BRS, Strat, Scholo; blue and epic)')
    q2([rr for rr in allE if rr['q'] == 4 and 55 <= rr['ilvl'] <= 80 and rr['v']], 'all Era epics ilvl 55-80')
    q2([rr for rr in S if rr['phase'] <= 6 and rr['v']], 'SoD phase 4-6 epics')
