#!/usr/bin/env python3
"""Tank stat prices (standards 36) from Era level 60 items, the way bis_budget_rules.md priced the rest.

Era items (Wowhead's Classic gear planner, versionNum 11300), quality 3 at ilvl 45-70 and 4 at 55-80, armor and
shields with no Use:, proc or non-stat Equip: line. The p-norm model of bis_model.py (p 1.7) with every non-tank
weight fixed at standards 1b and the tank weights free:
  def     Defense skill, a point
  dodge   dodge, parry, block chance, in %
  parry
  block
  blockv  block value, a point
  barmor  bonus armor (Wowhead's armorbonus: armor above the slot's base), a point
  res     one school's resistance, a point (Era items with all five at the same amount count as allres, one line)
  allres  "+N All Resistances", a point
Slot mods (chest 1) and V per quality are fitted too. 90% intervals from 60 bootstrap resamples. Also:
  - the same fit at p 1.5 and on epics only (alternatives);
  - the archetype factor of tank pieces (does a tank piece sit on the common line?);
  - f for tank pieces: linear total / (slot mod x L(ilvl)) on Era epics, as standards 1b's f table, at the fitted
    prices, and the shield's linear total as a multiple of L(ilvl);
  - the cross-check against TBC's level 60 rating conversions (each rating point costs 1 in Blizzard's budget,
    and 1.5 Defense rating, 12 dodge, 20 parry and 5 block rating make a point / 1% at level 60);
Writes tank_prices.json next to this script (gear_targets_tanks.py reads it).
Usage: python3 tank_prices.py
"""
import collections, json, os, sys
import numpy as np
sys.path.insert(0, os.path.dirname(__file__))
import bis_data, bis_model
from bis_data import era60, ARMOR, TIERS

RES = {'arcres': 'res', 'firres': 'res', 'frores': 'res', 'natres': 'res', 'shares': 'res', 'holres': 'res'}
for k in RES:
    bis_data._TANKMAP[k] = k
bis_data._TANKMAP['armorbonus'] = 'barmor'


def kind(stat):
    if stat in RES:
        return 'res'
    return bis_data.kind(stat)


bis_model.kind = kind  # one weight for every school's resistance

TANK = ['def', 'dodge', 'parry', 'block', 'blockv', 'barmor', 'res', 'allres']
# standards 1b (non-tank stats fixed at the ruling's weights)
W1B = dict(bis_model.W_MEASURED, sp=0.86, heal=0.45, school=0.70, ap=0.5, rap=0.4, crit=1.0, hit=1.0, spcrit=1.0, sphit=1.0,
           mp5=2.4, fap=0.30, barmor=0.05, res=1.0, allres=3.0)
W1B.update({'def': 1.5, 'dodge': 12.0, 'parry': 20.0, 'block': 5.0, 'blockv': 0.5})
TBC = {'def': 1.5, 'dodge': 12.0, 'parry': 20.0, 'block': 5.0}  # rating points per unit at level 60, each costing 1
MOD = {'head': 1, 'chest': 1, 'legs': 1, 'shoulder': .75, 'hands': .75, 'waist': .75, 'feet': .75,
       'neck': .5625, 'back': .5625, 'wrist': .5625, 'finger': .5625}
L = lambda il: 1.125 * il + 2.0
SLOTS = ARMOR + ['shield']
TANKSTATS = {'def', 'dodge', 'parry', 'block', 'blockv'}


def rows():
    E = era60()
    out = []
    for r in E:
        if r['slot'] not in SLOTS or not r['v'] or r['effect'] is not False:
            continue
        other = {k for k in r['other'] if k not in RES and k != 'armorbonus'}
        if other:
            continue
        if not ((r['q'] == 3 and 45 <= r['ilvl'] <= 70) or (r['q'] == 4 and 55 <= r['ilvl'] <= 80)):
            continue
        v = dict(r['v'])
        rs = {k: x for k, x in v.items() if k in RES}
        if len(rs) >= 5 and len(set(rs.values())) == 1:
            x = next(iter(rs.values()))
            for k in rs:
                del v[k]
            v['allres'] = x
        r = dict(r, v=v)
        ks = set(v)
        r['tankish'] = bool(ks & TANKSTATS)
        r['resish'] = bool(ks & (set(RES) | {'allres', 'barmor'}))
        r['arch2'] = 'tank' if r['tankish'] else ('resist / armor' if r['resish'] else r['arch'])
        r['plain'] = not r['resish']
        if r['arch2'] == 'stamina':
            continue
        out.append(r)
    return out


def fit(R, p, boot=60, factor=False):
    M = bis_model.PModel(R, p, w=W1B, free_w=TANK, factor=(lambda r: r['arch2']) if factor else None,
                         factor_ref='physical' if factor else None)
    th, mask = M.robust()
    r = M.resid(th)
    w, m, ab, f = M.unpack(th)
    bs = M.boot(th, mask, boot) if boot else []
    return M, th, mask, r, w, m, ab, f, bs


def ci(vals):
    return np.percentile(vals, 5), np.percentile(vals, 95)


def main():
    R = rows()
    print(f'Era clean armor + shields, blues 45-70 + epics 55-80: {len(R)} pieces; '
          f'{collections.Counter(r["arch2"] for r in R)}')
    n_with = {k: sum(1 for r in R if any(kind(s) == k for s in r['v'])) for k in TANK}
    print('pieces carrying each tank stat:', n_with)
    res = {}
    for lab, RR, p in (('p 1.7, blues and epics (the standard)', R, 1.7), ('p 1.5, blues and epics', R, 1.5),
                       ('p 1.7, epics only', [r for r in R if r['q'] == 4], 1.7)):
        M, th, mask, r, w, m, ab, f, bs = fit(RR, p)
        rms = 100 * np.sqrt(np.mean(r[mask] ** 2))
        tk = [i for i, rr in enumerate(RR) if mask[i] and rr['arch2'] == 'tank']
        rt = 100 * np.sqrt(np.mean(r[tk] ** 2))
        print(f'\n== {lab}: {len(RR)} pieces, {mask.sum()} kept, rms {rms:.2f}% (tank pieces {rt:.2f}%, n={len(tk)})')
        print('| stat | price | 90% interval | pieces |')
        print('|---|---|---|---|')
        out = {}
        for k in TANK:
            if k in M.free_w:
                lo, hi = ci([b[0][k] for b in bs])
                out[k] = (w[k], lo, hi, n_with[k])
                print(f'| {k} | {w[k]:.3f} | {lo:.3f}-{hi:.3f} | {n_with[k]} |')
        print('slot mods:', {s: round(x, 3) for s, x in m.items()})
        print('V:', {q: f'{a:.3f} ilvl {b:+.2f}' for q, (a, b) in ab.items()})
        res[lab] = {'w': {k: v[0] for k, v in out.items()}, 'ci': {k: v[1:3] for k, v in out.items()}, 'rms': rms, 'rms_tank': rt,
                    'mods': m}
    std = res['p 1.7, blues and epics (the standard)']
    W = dict(W1B, **std['w'])

    # archetype factor at p 1.7 with the prices fitted
    M = bis_model.PModel(R, 1.7, w=W, free_w=(), factor=lambda r: r['arch2'], factor_ref='physical')
    th, mask = M.robust()
    _, _, _, fac = M.unpack(th)
    bs = M.boot(th, mask, 40)
    print('\narchetype factor vs physical at p 1.7, tank prices fixed at the fit: ' +
          ', '.join(f'{a} {fac[a]:.3f} ({ci([b[3][a] for b in bs])[0]:.3f}-{ci([b[3][a] for b in bs])[1]:.3f})' for a in fac if a != 'physical'))

    # f for tank pieces (linear / (slot mod x L)), Era epics 55-80, as 1b's table
    EP = [r for r in R if r['q'] == 4 and r['slot'] in ARMOR]
    lin = lambda v, w=W: sum(w[kind(k)] * x for k, x in v.items())
    print('\nf = linear total / (slot mod x L(ilvl)) on Era epics 55-80 (median, n), at the fitted prices:')
    fT = {}
    for a in ('tank', 'tank, no resistance or bonus armor', 'physical', 'caster', 'hybrid'):
        cells = []
        for t in TIERS + ['all']:
            x = [lin(r['v']) / (MOD[r['slot']] * L(r['ilvl'])) for r in EP if r['arch2'] == a.split(',')[0]
                 and (',' not in a or r['plain']) and (t == 'all' or r['tier'] == t)]
            fT[(a, t)] = (float(np.median(x)) if x else None, len(x))
            cells.append(f'{np.median(x):.3f} (n={len(x)})' if x else '-')
        print(f'   {a:9}: ' + ' | '.join(cells))
    stats_per = np.mean([len(r['v']) for r in EP if r['arch2'] == 'tank'])
    stats_pl = np.mean([len(r['v']) for r in EP if r['arch2'] == 'tank' and r['plain']])
    print(f'   tank epics carry {stats_per:.1f} stat lines a piece, {stats_pl:.1f} without resistance or bonus armor')
    P1 = [r for r in EP if r['arch2'] == 'tank' and r['p1']]
    print(f'   Phase 1 tank epics (MC, Onyxia, world bosses, pre-raid dungeons): f median '
          f'{np.median([lin(r["v"]) / (MOD[r["slot"]] * L(r["ilvl"])) for r in P1]):.3f} (n={len(P1)})')
    SH = [r for r in R if r['slot'] == 'shield' and r['q'] == 4]
    shx = [lin(r['v']) / L(r['ilvl']) for r in SH]
    shv = [bis_model.value(r['v'], W, 1.7) / (0.661 * r['ilvl'] + 2.92) for r in SH]
    print(f'   shields (epic, ilvl 55-80): linear / L(ilvl) median {np.median(shx):.3f} (n={len(shx)}); 1.7-norm / V median {np.median(shv):.3f}')
    for r in SH:
        print(f'      {r["name"]} ({r["ilvl"]}): {r["v"]}  linear {lin(r["v"]):.1f}  / L {lin(r["v"]) / L(r["ilvl"]):.3f}')

    # Phase 1 tank epics, how their budget splits
    print('\nbudget split of Era tank epics (armor, linear shares at the fitted prices):')
    tot = collections.Counter()
    for r in [r for r in EP if r['arch2'] == 'tank']:
        for k, x in r['v'].items():
            tot[kind(k) if kind(k) in TANK or kind(k) == 'prim' and k != 'sta' else ('sta' if k == 'sta' else kind(k))] += W[kind(k)] * x
    s = sum(tot.values())
    print('   ' + ', '.join(f'{k} {100 * x / s:.1f}%' for k, x in tot.most_common()))

    print('\ncross-check, TBC level 60 rating conversions at 1 a rating point: ' +
          ', '.join(f'{k} {TBC[k]} (fit {std["w"][k]:.2f}, {std["ci"][k][0]:.2f}-{std["ci"][k][1]:.2f})' for k in TBC))
    json.dump({'prices': std['w'], 'ci': std['ci'], 'rms': std['rms'], 'rms_tank': std['rms_tank'],
               'alternatives': {k: v['w'] for k, v in res.items()},
               'f_tank': {t: fT[('tank', t)][0] for t in TIERS + ['all']},
               'f_tank_plain': {t: fT[('tank, no resistance or bonus armor', t)][0] for t in TIERS + ['all']}, 'f_physical': {t: fT[('physical', t)][0] for t in TIERS + ['all']},
               'shield_L': float(np.median(shx)), 'tbc': TBC},
              open(os.path.join(os.path.dirname(os.path.abspath(__file__)), 'tank_prices.json'), 'w'), indent=1)


if __name__ == '__main__':
    main()
