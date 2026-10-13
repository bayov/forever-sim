"""Q5 of the level 60 BiS budget analysis (orchestration/bis_budget_rules.md): the proposed pricing and its table.

Budget line (user ruling): weighted linear total = slot mod x L(ilvl), L = 1.125 ilvl + 2.0, weights sp 0.86,
healing on top 0.45, school 0.70, AP 0.5, RAP 0.4, crit and hit rating 1.0, MP5 2.4, slot mods 1 / 0.75 / 0.5625.
L is the average Era epic armor piece (3.9 stats). This script measures how far each archetype's real pieces sit
from it (factor f = linear total / (slot mod x L)), per slot tier and for weapons, on:
  - Era epics ilvl 55-80 (the pieces L was fitted on) and Era Phase 1 raid epics;
  - Forever ilvl 5-30 (against Forever's own linear line per quality);
  - SoD phase 4-6 epics (against SoD's own line);
how well f predicts real pieces compared with the p-norm, and builds the per-archetype table against standards 1a.

Usage: python3 bis_proposal.py
"""
import collections, os, sys
import numpy as np
sys.path.insert(0, os.path.dirname(__file__))
from bis_data import era60, sod60, forever30, clean, ARMOR, TIERS
from bis_model import PModel, W_MEASURED, value, linear, med
from bis_archetype import nstats

W_REC = dict(W_MEASURED, sp=0.86, heal=0.45, school=0.70, ap=0.5, rap=0.4, crit=1.0, hit=1.0, spcrit=1.0, sphit=1.0, mp5=2.4,
             **{'def': 1.56, 'dodge': 12.4, 'parry': 20.5, 'block': 4.5})
MOD = {'head': 1, 'chest': 1, 'legs': 1, 'shoulder': .75, 'hands': .75, 'waist': .75, 'feet': .75,
       'neck': .5625, 'back': .5625, 'wrist': .5625, 'finger': .5625, 'trinket': .5625}
L = lambda il: 1.125 * il + 2.0
SUBS = ['caster dps', 'healer', 'caster int/spi', 'caster mp5', 'physical', 'physical+int', 'hybrid', 'tank']
rng = np.random.default_rng(7)


def ratio(r, line=L, w=W_REC):
    return linear(r['v'], w) / (MOD[r['slot']] * line(r['ilvl']))


def med_ci(x, n=400):
    x = np.asarray(x)
    if len(x) < 3:
        return med(x), float('nan'), float('nan')
    b = [np.median(rng.choice(x, len(x))) for _ in range(n)]
    return float(np.median(x)), np.percentile(b, 5), np.percentile(b, 95)


def factor_table(rows, label, line=L, key='sub', cats=SUBS):
    print(f'\n{label}')
    print('| archetype | ' + ' | '.join(TIERS) + ' | all armor | stats per piece |')
    print('|---|---|---|---|---|---|')
    out = {}
    for a in cats:
        cells = []
        for t in TIERS + ['all']:
            x = [ratio(r, line) for r in rows if r[key] == a and (t == 'all' or r['tier'] == t)]
            if not x:
                cells.append('-')
                continue
            m, lo, hi = med_ci(x)
            out[(a, t)] = (m, lo, hi, len(x))
            cells.append(f'{m:.3f} ({lo:.2f}-{hi:.2f}, n={len(x)})')
        rs = [r for r in rows if r[key] == a]
        if rs:
            print(f'| {a} | ' + ' | '.join(cells) + f' | {np.mean([nstats(r["v"]) for r in rs]):.1f} |')
    return out


if __name__ == '__main__':
    E = era60()
    EA = [r for r in clean(E, ARMOR) if r['q'] == 4 and 55 <= r['ilvl'] <= 80 and r['arch'] != 'stamina']
    X = np.array([[r['ilvl'], 1.0] for r in EA if r['arch'] != 'tank'])
    y = np.array([linear(r['v'], W_REC) / MOD[r['slot']] for r in EA if r['arch'] != 'tank'])
    a, b = np.linalg.lstsq(X, y, rcond=None)[0]
    print(f'Era clean epic armor ilvl 55-80 without tank pieces: n={len(y)}, linear line {a:.3f} ilvl {b:+.2f} (the ruling: 1.125 ilvl + 2.0)')
    fE = factor_table(EA, '1. Era epics ilvl 55-80: linear total / (slot mod x (1.125 ilvl + 2.0)), median (90% interval of the median, n)')
    P1 = [r for r in EA if r['p1'] in ('Molten Core', 'Onyxia', 'world boss')]
    fP = factor_table(P1, '2. Era Phase 1 raid epics only (Molten Core, Onyxia, Azuregos, Kazzak)')

    # accuracy: per-cell factor vs the p-norm, same pieces (Era clean epic armor), same 3-sigma trim rule
    def rms(x):
        x = np.asarray(x) - np.median(x)
        mad = 1.4826 * np.median(np.abs(x))
        return 100 * np.sqrt(np.mean(x ** 2)), 100 * np.sqrt(np.mean(x[np.abs(x) < 3 * mad] ** 2)), int((np.abs(x) >= 3 * mad).sum())
    res_c = [np.log(ratio(r)) for r in EA]
    res_f = [np.log(ratio(r) / fE[(r['sub'], r['tier'])][0]) for r in EA]
    M = PModel(EA, 1.7, w=dict(W_MEASURED, **{'def': 1.56, 'dodge': 12.4, 'parry': 20.5, 'block': 4.5}))
    th, mask = M.robust()
    rp = M.resid(th)
    ab17 = M.unpack(th)[2][4]
    print(f'\n3. How well each rule reproduces the {len(EA)} clean Era epic armor pieces (rms log error: all pieces / without 3-sigma outliers (n out)):')
    print('   one linear total for every piece (standards 1a style): %.1f%% / %.1f%% (%d)' % rms(res_c))
    print('   linear total x archetype-and-tier factor: %.1f%% / %.1f%% (%d)' % rms(res_f))
    print('   p-norm, p 1.7, measured weights, value against V(ilvl) x slot mod: %.1f%% / %.1f%% (%d)' % rms(rp) +
          f'; V_epic(ilvl) = {ab17[0]:.3f} ilvl {ab17[1]:+.2f}')

    # linear / value spread by stat count at p 1.7 (what a stat-count cap controls)
    rs = [r for r in EA]
    print('\n4. Era epics: linear total / p-norm value (p 1.7) by stat count (the spread effect a stat-count cap bounds):')
    for n in range(2, 8):
        x = [linear(r['v'], W_REC) / value(r['v'], W_REC, 1.7) for r in rs if nstats(r['v']) == n]
        if x:
            print(f'   {n} stats: median {med(x):.2f}, max {max(x):.2f}, even split would give {n ** (1 - 1/1.7):.2f} (n={len(x)})')

    # Forever ilvl 5-30: linear factors against Forever's own line per quality
    F = forever30()
    FA = [r for r in clean(F, ARMOR) if r['arch'] != 'stamina' and r['ilvl'] >= 10]
    lines = {}
    for q in (2, 3):
        rq = [r for r in FA if r['q'] == q]
        X = np.array([[r['ilvl'], 1.0] for r in rq])
        y = np.array([linear(r['v'], W_REC) / MOD[r['slot']] for r in rq])
        lines[q] = np.linalg.lstsq(X, y, rcond=None)[0]
    LF = {}
    for r in FA:
        r['_lf'] = linear(r['v'], W_REC) / (MOD[r['slot']] * (lines[r['q']][0] * r['ilvl'] + lines[r['q']][1]))
    print('\n5. Forever ilvl 10-30 armor: linear total / (slot mod x Forever line of its quality), median (90%), n, stats per piece')
    for a_ in ['caster dps', 'healer', 'caster int/spi', 'physical', 'physical+int', 'hybrid']:
        for src in ('all', 'forever-new'):
            x = [r['_lf'] for r in FA if r['sub'] == a_ and (src == 'all' or r['src'] == src)]
            if len(x) >= 3:
                m, lo, hi = med_ci(x)
                print(f'   {a_:14} {src:11}: {m:.3f} ({lo:.2f}-{hi:.2f}), n={len(x)}, stats {np.mean([nstats(r["v"]) for r in FA if r["sub"] == a_ and (src == "all" or r["src"] == src)]):.1f}')

    # SoD phase 4-6 epics against SoD's own line
    S = sod60()
    SA = [r for r in clean(S, ARMOR) if r['phase'] <= 6 and r['arch'] != 'stamina']
    X = np.array([[r['ilvl'], 1.0] for r in SA if r['arch'] != 'tank'])
    y = np.array([linear(r['v'], W_REC) / MOD[r['slot']] for r in SA if r['arch'] != 'tank'])
    sa, sb = np.linalg.lstsq(X, y, rcond=None)[0]
    factor_table(SA, f'6. SoD phase 4-6 epics against SoD\'s own line ({sa:.3f} ilvl {sb:+.2f})', line=lambda il: sa * il + sb)

    # weapons: Era epics against L; Forever ilvl 10-30 against Forever's line (chest equivalents of linear)
    print('\n7. Weapons: linear total / L(ilvl) (Era epics ilvl 55-80) and / Forever line (Forever ilvl 10-30), median, n')
    WT = ['two-hand', 'one-hand', 'held off-hand', 'ranged', 'wand']
    EW = [r for r in clean(E) if r['tier'] in WT and r['q'] == 4 and 55 <= r['ilvl'] <= 80]
    FW = [r for r in clean(F) if r['tier'] in WT and r['ilvl'] >= 10]
    wsub = lambda r: 'caster' if r['sub'] in ('caster dps', 'healer', 'caster mp5') else 'int/spi only' if r['sub'] == 'caster int/spi' else r['arch']
    for t in WT:
        for a_ in ('caster', 'physical', 'hybrid', 'int/spi only'):
            xe = [linear(r['v'], W_REC) / L(r['ilvl']) for r in EW if r['tier'] == t and wsub(r) == a_]
            xf = [linear(r['v'], W_REC) / (lines[r['q']][0] * r['ilvl'] + lines[r['q']][1]) for r in FW if r['tier'] == t and wsub(r) == a_ and r['q'] in lines]
            if xe or xf:
                print(f'   {t:14} {a_:13}: Era {med(xe):.3f} (n={len(xe)})   Forever {med(xf):.3f} (n={len(xf)})')

    # 8. the proposal table against standards 1a
    print('\n8. Proposal vs standards 1a, weighted linear total per piece (proposal weights: sp 0.86, AP 0.5, hit / crit 1.0, MP5 2.4)')
    T1, T2, T3 = TIERS
    fac = {'pure caster': {T1: fE[('caster dps', T1)][0], T2: fE[('caster dps', T2)][0], T3: fE[('caster dps', T3)][0]},
           'physical melee': {t: fE[('physical', t)][0] for t in TIERS},
           'hunter': {t: fE[('physical+int', t)][0] for t in TIERS},
           'hybrid (Ret / Enh)': {t: fE[('hybrid', t)][0] for t in TIERS},
           'feral': {t: fE[('physical', t)][0] for t in TIERS}}
    # weapons, as multiples of L(ilvl): Era physical; caster from Forever's caster / physical ratio at ilvl <= 30
    med_w = lambda rows, t, a_: med([linear(r['v'], W_REC) / L(r['ilvl']) for r in rows if r['tier'] == t and wsub(r) == a_])
    medF = lambda t, a_: med([linear(r['v'], W_REC) / (lines[r['q']][0] * r['ilvl'] + lines[r['q']][1]) for r in FW if r['tier'] == t and wsub(r) == a_ and r['q'] in lines])
    p2h, p1h, oh, rng_, wand = med_w(EW, 'two-hand', 'physical'), med_w(EW, 'one-hand', 'physical'), med_w(EW, 'held off-hand', 'caster'), med_w(EW, 'ranged', 'physical'), med_w(EW, 'wand', 'caster')
    staff = p2h * medF('two-hand', 'caster') / medF('two-hand', 'physical')
    mh = p1h * medF('one-hand', 'caster') / medF('one-hand', 'physical')
    hyb_ratio = fE[('hybrid', 'all')][0] / fE[('physical', 'all')][0]
    hun_ratio = fE[('physical+int', 'all')][0] / fE[('physical', 'all')][0]
    print(f'   weapon multiples of L: physical 2H {p2h:.3f}, physical 1H {p1h:.3f}, caster held off-hand {oh:.3f}, physical ranged {rng_:.3f}, wand {wand:.3f}; '
          f'caster staff {staff:.3f} (Forever ratio {medF("two-hand", "caster")/medF("two-hand", "physical"):.2f} x physical), caster main hand {mh:.3f} '
          f'(ratio {medF("one-hand", "caster")/medF("one-hand", "physical"):.2f}); hybrid / physical armor {hyb_ratio:.2f}, hunter / physical {hun_ratio:.2f}')
    V1a = lambda il: 1.2 * il - 36.4
    today = {'T1 66': 1.73 * V1a(66), 'T2 66': 1.73 * .77 * V1a(66), 'waist 71': 1.73 * .77 * V1a(71), 'T3 71': 1.73 * .5625 * V1a(71),
             '2H 76': 1.73 * V1a(76), '1H 71': 1.73 * .42 * V1a(71), 'caster weapons': 1.73 * 73, 'ranged 71': 1.73 * .32 * V1a(71)}
    sp_share = {'pure caster': 0.282, 'hybrid (Ret / Enh)': 0.212}   # gear_targets 5a caster mean; Ret / Enh keep today's ~20-22%
    stat_trinket = 0.56 * L(71)
    rows_out = []
    for arch, f in fac.items():
        prop = {'T1 66': L(66) * f[T1], 'T2 66': .75 * L(66) * f[T2], 'waist 71': .75 * L(71) * f[T2], 'T3 71': .5625 * L(71) * f[T3]}
        armor_p = 3 * prop['T1 66'] + 3 * prop['T2 66'] + prop['waist 71'] + 5 * prop['T3 71']
        armor_t = 3 * today['T1 66'] + 3 * today['T2 66'] + today['waist 71'] + 5 * today['T3 71']
        if arch == 'pure caster':
            confs = [('staff + wand', staff * L(76) + wand * L(71), today['caster weapons'] + today['ranged 71']),
                     ('staff + relic (Balance, Ele)', staff * L(76), today['caster weapons']),
                     ('main hand + off-hand + wand', (mh + oh) * L(71) + wand * L(71), today['caster weapons'] + today['ranged 71'])]
        elif arch == 'physical melee':
            confs = [('two-hander + ranged', p2h * L(76) + rng_ * L(71), today['2H 76'] + today['ranged 71']),
                     ('two one-handers + ranged', 2 * p1h * L(71) + rng_ * L(71), 2 * today['1H 71'] + today['ranged 71'])]
        elif arch == 'hunter':
            confs = [('two-hander + bow', hun_ratio * p2h * L(76) + rng_ * L(71), today['2H 76'] + today['ranged 71'])]
        elif arch.startswith('hybrid'):
            confs = [('two-hander + relic', hyb_ratio * p2h * L(76), today['2H 76'])]
        else:
            confs = [('two-hander + idol', p2h * L(76), today['2H 76'])]
        rp = 1 + sp_share.get(arch, 0) * (0.86 / 0.55 - 1)
        print(f'\n   {arch}: per piece, proposal (today 1a in its own weights; today re-priced at sp 0.86 x{rp:.3f})')
        for k in prop:
            print(f'      {k:9} {prop[k]:6.1f}   (today {today[k]:6.1f}; re-priced {today[k]*rp:6.1f})')
        print(f'      armor (12 pieces) {armor_p:6.1f}   (today {armor_t:6.1f}; re-priced {armor_t*rp:6.1f}; change {100*(armor_p/(armor_t*rp)-1):+.1f}%)')
        for lab, wp, wt_ in confs:
            tot_p, tot_t = armor_p + wp, (armor_t + wt_) * rp
            print(f'      weapons {lab}: {wp:6.1f} (today {wt_:6.1f}, re-priced {wt_*rp:6.1f}); set {tot_p:6.1f} vs {tot_t:6.1f} re-priced: {100*(tot_p/tot_t-1):+.1f}%; '
                  f'with two stat trinkets (+{2*stat_trinket:.1f}): {tot_p + 2*stat_trinket:6.1f} ({100*((tot_p + 2*stat_trinket)/tot_t-1):+.1f}%)')
            rows_out.append((arch, lab, prop, armor_p, armor_t, rp, wp, wt_, tot_p, tot_t))
    print(f'\n   trinket (ilvl 71): stat trinket {stat_trinket:.1f} linear (0.56 x L(71)); today none')
