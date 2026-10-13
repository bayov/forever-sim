"""Q6 of the ilvl <= 30 analysis: a level 60 epic budget for the synthetic gear.

Forever's ilvl <= 30 items price stats the way Era does (b30_era.py), so Era's own level 60 items
are the reference for the level 60 budget. This script measures, on Era items:
  - the 1.5-norm V(ilvl) and the stat weights of Era epics (and blues), free fit;
  - the weighted linear total per slot mod of Era epics against ilvl, at 66 / 71 / 76;
  - whether Era's own low-level line (ilvl <= 30) extrapolates to its level 60 items, which is
    what a Forever-from-ilvl-30 extrapolation would have to assume;
and compares them with standards 1a and the all-items proposal.
Crit and hit percentages are rating points (14 per 1% crit, 10 per 1% hit), as in the standards.
"""
import collections, os, statistics, sys
import numpy as np
sys.path.insert(0, os.path.dirname(__file__))
from b30_data import era, forever, clean, kind_of, ARMOR, PRIM, STD_MOD
from b30_fit import Model, summarize, linear_totals

P = 1.5


def lin(r, w):
    return sum((1.0 if k in PRIM else w[kind_of(k)]) * x for k, x in r['v'].items())


def line(rows, w, mods):
    X = np.array([[r['ilvl'], 1.0] for r in rows])
    y = np.array([lin(r, w) / mods[r['slot']] for r in rows])
    a, b = np.linalg.lstsq(X, y, rcond=None)[0]
    res = y / (X @ [a, b]) - 1
    return a, b, np.sqrt(np.mean(res ** 2)), y / (X @ [a, b])


if __name__ == '__main__':
    E = era()
    EA = clean(E, ARMOR)
    epics = [r for r in EA if r['q'] == 4 and 55 <= r['ilvl'] <= 80]
    blues = [r for r in EA if r['q'] == 3 and 45 <= r['ilvl'] <= 70]
    print(f'Era clean armor: epics ilvl 55-80 n={len(epics)}, blues ilvl 45-70 n={len(blues)}')
    print('stats per item: epics', round(statistics.mean(len(r['v']) for r in epics), 2), 'blues', round(statistics.mean(len(r['v']) for r in blues), 2))
    print('stat kinds on epics:', dict(collections.Counter(kind_of(k) for r in epics for k in r['v'])))

    # A. free fit on Era level 60 gear: weights, slot mods, V per quality
    M, th, mask, w_era, m_era, ab_era = summarize('Era armor, blues ilvl 45-70 + epics ilvl 55-80', blues + epics, P, boot=100)

    # B. linear totals of Era epics per slot mod at 66 / 71 / 76
    W_STD = {'sp': 0.55, 'heal': 0.45, 'school': 0.45, 'ap': 0.5, 'rap': 0.4, 'crit': 1.0, 'hit': 1.0, 'mp5': 2.5}
    W_LOW = {'sp': 0.86, 'heal': 0.40, 'school': 0.60, 'ap': 0.5, 'rap': 0.4, 'crit': 1.0, 'hit': 1.0, 'mp5': 2.0}
    # proposal: sp from the ilvl <= 30 fit; the stats ilvl <= 30 cannot price (crit, hit, MP5, healing, school) from Era level 60
    W_REC = {'sp': 0.86, 'heal': 0.45, 'school': 0.70, 'ap': 0.5, 'rap': 0.4, 'crit': 1.0, 'hit': 1.0, 'mp5': 2.4}
    std = lambda L: 1.73 * (1.2 * L - 36.4)
    allp = lambda L: 1.28 * L - 10.9
    print('\nB. Era epics (ilvl 55-80), weighted linear total / slot mod (standards slot mods):')
    print('   weights                           fit                        rms   ilvl 66   71     76    | standards 1a 66/71/76: '
          f'{std(66):.1f} {std(71):.1f} {std(76):.1f}; all-items proposal: {allp(66):.1f} {allp(71):.1f} {allp(76):.1f}')
    kept = [r for i, r in enumerate(blues + epics) if mask[i] and r['q'] == 4]
    for lab, w in (('ilvl<=30 Forever fit (sp 0.86)', W_LOW), ('proposal (sp 0.86, MP5 2.4)', W_REC), ('standards (sp 0.55)', W_STD), ('Era level 60 fit', w_era)):
        for rows, tag in ((epics, 'all'), (kept, 'trimmed')):
            a, b, rms, _ = line(rows, w, STD_MOD)
            print(f'   {lab:32}  {a:.3f} ilvl {b:+6.2f}  ({tag:7}) {100*rms:4.0f}%  {a*66+b:6.1f} {a*71+b:6.1f} {a*76+b:6.1f}')
    # caster vs physical epics on the trend with sp 0.86 and 0.55
    for lab, w in (('sp 0.86', W_LOW), ('sp 0.55', W_STD)):
        a, b, rms, ratio = line(kept, w, STD_MOD)
        cas = [x for x, r in zip(ratio, kept) if 'sp' in r['v'] or any(k.startswith('school') for k in r['v'])]
        phy = [x for x, r in zip(ratio, kept) if not ('sp' in r['v'] or any(k.startswith('school') for k in r['v'])) and any(k in r['v'] for k in ('str', 'agi', 'ap'))]
        heal = [x for x, r in zip(ratio, kept) if 'heal' in r['v']]
        print(f'   {lab}: median linear total vs the trend, caster epics {statistics.median(cas):.3f} (n={len(cas)}), physical {statistics.median(phy):.3f} (n={len(phy)}), healer {statistics.median(heal):.3f} (n={len(heal)})')
    # same for the 1.5-norm budget
    print('   1.5-norm V of Era epics (fit A): ' + ', '.join(f'ilvl {L}: {ab_era[4][0]*L+ab_era[4][1]:.1f} (standards V {1.2*L-36.4:.1f})' for L in (66, 71, 76)))
    rat = [lin(r, w_era) / (M.cost(w_era)[i]) for i, r in enumerate(blues + epics) if mask[i] and r['q'] == 4]
    print(f'   Era epics: linear total / 1.5-norm value median {statistics.median(rat):.3f}, mean {statistics.mean(rat):.3f} (standards 1a uses 1.73)')

    # C. does a low-level line extrapolate? Era blues: line from ilvl 15-30 vs actual at 45-65
    print('\nC. Extrapolating from ilvl <= 30 (Era blues and greens, weights sp 0.86):')
    for q in (2, 3):
        low = [r for r in EA if r['q'] == q and 15 <= r['ilvl'] <= 30]
        hi = [r for r in EA if r['q'] == q and 50 <= r['ilvl'] <= 66]
        a, b, rms, _ = line(low, W_LOW, STD_MOD)
        a2, b2, rms2, _ = line(hi, W_LOW, STD_MOD)
        print(f'   q{q}: Era ilvl 15-30 line {a:.3f} ilvl {b:+.2f} (n={len(low)}) predicts {a*60+b:.1f} at ilvl 60; Era ilvl 50-66 items: {a2:.3f} ilvl {b2:+.2f} '
              f'(n={len(hi)}) -> {a2*60+b2:.1f} at 60 (ratio {(a*60+b)/(a2*60+b2):.2f})')
    F = clean(forever(), ARMOR)
    for q in (2, 3):
        low = [r for r in F if r['q'] == q and 15 <= r['ilvl'] <= 30]
        a, b, rms, _ = line(low, W_LOW, STD_MOD)
        print(f'   Forever q{q} ilvl 15-30 line {a:.3f} ilvl {b:+.2f} (n={len(low)}) -> ilvl 60: {a*60+b:.1f}, 66: {a*66+b:.1f}')
    # epic over blue at the same ilvl (Era), the quality step a blue extrapolation would need
    for L in (60, 66, 71):
        eb = [r for r in EA if r['q'] == 3 and abs(r['ilvl'] - L) <= 3]
        ee = [r for r in EA if r['q'] == 4 and abs(r['ilvl'] - L) <= 3]
        if eb and ee:
            mb = statistics.median(lin(r, W_LOW) / STD_MOD[r['slot']] for r in eb)
            me = statistics.median(lin(r, W_LOW) / STD_MOD[r['slot']] for r in ee)
            print(f'   Era ilvl {L}+/-3: median linear total / slot mod, blues {mb:.1f} (n={len(eb)}), epics {me:.1f} (n={len(ee)}), epic/blue {me/mb:.2f}')

    # D. slot mods of Era level 60 gear (from fit A) for comparison with the ilvl <= 30 ones
    print('\nD. Era level 60 slot mods (fit A): ' + ', '.join(f'{s} {m_era[s]:.3f}' for s in sorted(m_era)))

    # E. Era epics by ilvl bin (linear total / slot mod, sp 0.86), and the fit on narrower ranges
    print('\nE. Era epic linear totals / slot mod by ilvl (sp 0.86):')
    for lo, hi in ((55, 62), (63, 65), (66, 66), (67, 70), (71, 71), (72, 75), (76, 80)):
        b_ = [lin(r, W_LOW) / STD_MOD[r['slot']] for r in epics if lo <= r['ilvl'] <= hi]
        if b_:
            print(f'   ilvl {lo}-{hi}: n={len(b_):3}, median {statistics.median(b_):.1f}, mean {statistics.mean(b_):.1f}')
    for lo, hi in ((55, 71), (58, 76), (60, 80), (55, 80)):
        rows = [r for r in epics if lo <= r['ilvl'] <= hi]
        a, b, rms, _ = line(rows, W_LOW, STD_MOD)
        print(f'   fit on ilvl {lo}-{hi} (n={len(rows)}): {a:.3f} ilvl {b:+.2f} -> 66: {a*66+b:.1f}, 71: {a*71+b:.1f}, 76: {a*76+b:.1f}')

    # F. trinkets: Era stat-only trinkets (no use / proc effect), value against the armor V curve
    tr = [r for r in clean(E, ['trinket']) if r['q'] in (3, 4) and r['ilvl'] >= 45]
    print(f'\nF. Era stat-only trinkets, blue/epic ilvl 45+: n={len(tr)}')
    for r in tr:
        a, b = ab_era[r['q']]
        value = sum(((1.0 if k in PRIM else w_era[kind_of(k)]) * x) ** P for k, x in r['v'].items()) ** (1 / P)
        ea_, eb_, *_ = line(epics, W_LOW, STD_MOD)
        print(f'   {r["name"]} (q{r["q"]}, ilvl {r["ilvl"]}): {r["v"]} -> value / V = {value / (a * r["ilvl"] + b):.3f}; '
              f'linear total (sp 0.86, mp5 2.4, healing 0.45) / epic linear trend at its ilvl = {lin(r, dict(W_LOW, mp5=2.4, heal=0.45)) / (ea_ * r["ilvl"] + eb_):.2f}')

    # G. caster vs physical Era epics against one linear trend, by spell power weight
    cas = lambda r: ('sp' in r['v'] or any(k.startswith('school') for k in r['v'])) and 'heal' not in r['v']
    phy = lambda r: not ('sp' in r['v'] or any(k.startswith('school') for k in r['v']) or 'heal' in r['v']) and any(k in r['v'] for k in ('str', 'agi', 'ap'))
    print(f'\nG. Era epics ilvl 55-80: caster pieces n={sum(map(cas, epics))} ({statistics.mean(len(r["v"]) for r in epics if cas(r)):.1f} stats each), '
          f'physical n={sum(map(phy, epics))} ({statistics.mean(len(r["v"]) for r in epics if phy(r)):.1f} stats each); linear total vs the common trend:')
    for sp in (0.5, 0.55, 0.65, 0.75, 0.86, 1.0):
        w = dict(W_REC, sp=sp, school=0.70 * sp / 0.86)
        a, b, rms, ratio = line(epics, w, STD_MOD)
        c = statistics.median([x for x, r in zip(ratio, epics) if cas(r)])
        ph = statistics.median([x for x, r in zip(ratio, epics) if phy(r)])
        print(f'   sp {sp:4}: trend {a:.3f} ilvl {b:+.2f} (66: {a*66+b:.1f}, 71: {a*71+b:.1f}, 76: {a*76+b:.1f}); caster {c:.3f}, physical {ph:.3f}, caster/physical {c/ph:.3f}')
