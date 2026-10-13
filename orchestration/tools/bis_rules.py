"""Q3 of the level 60 BiS budget analysis (orchestration/bis_budget_rules.md): special pricing rules.

1. Weapons, wands, off-hands, shields: slot modifier per (slot, archetype), fitted together with the armor so the
   weapon modifiers are relative to a chest of the same item level. Era level 60 (blues 45-70, epics 55-80),
   SoD phase 4-6 epics, Forever ilvl 5-30. Gives the caster weapon spell power rate, two-hand vs one-hand, wand
   vs bow / gun, held off-hand vs one-hander.
2. Healing vs damage, hit and crit on caster vs physical items: from bis_archetype.py's free fits (repeated here
   at p 1.5 and 1.7 for the record).
3. SoD hybrid epics: every SoD phase 4-6 epic that carries physical and spell stats.
4. Forever ilvl 5-30: pieces that carry both physical and spell stats; residual of each archetype on the Forever
   budget line.
5. Forever's stat-to-spell-power swaps (same id, ilvl and quality as Era): the implied spell power weight, by the
   stat it replaced, for single-term and partial swaps.

Usage: python3 bis_rules.py
"""
import collections, os, sys
import numpy as np
sys.path.insert(0, os.path.dirname(__file__))
from bis_data import era60, sod60, forever30, clean, kind, ARMOR, TIERS, PRIM
from bis_model import PModel, W_MEASURED, value, linear, ci, med
from bis_archetype import nstats, ARCH, FREE

WM = dict(W_MEASURED, **{'def': 1.56, 'dodge': 12.4, 'parry': 20.5, 'block': 4.5})
WEAP = ['two-hand', 'one-hand', 'held off-hand', 'shield', 'ranged', 'wand']


def wkey(r):
    if r['tier'] in TIERS:
        return r['slot']
    a = r['arch'] if r['arch'] != 'stamina' else 'physical'
    if r['sub'] == 'caster int/spi':
        a = 'int/spi only'
    return f"{r['tier']} | {a}"


def weapon_mods(rows, label, p, group=lambda r: r['q'], boot=30):
    M = PModel(rows, p, w=WM, slot_key=wkey, group=group)
    th, mask = M.robust()
    w, m, ab, f = M.unpack(th)
    r = M.resid(th)
    bs = M.boot(th, mask, boot) if boot else []
    print(f'\n== {label}, p={p}: {len(rows)} items ({mask.sum()} kept), rms {100*np.sqrt(np.mean(r[mask]**2)):.1f}%; chest = 1')
    print('| slot, archetype | modifier | 90% interval | n (kept) | stats per piece | spell power per budget point (sp x 0.86 / budget) |')
    print('|---|---|---|---|---|---|')
    out = {}
    for s in sorted(m):
        if s in ARMOR:
            continue
        idx = [i for i, rr in enumerate(rows) if wkey(rr) == s and mask[i]]
        if not idx:
            continue
        lo, hi = ci([b[1][s] for b in bs]) if bs else (float('nan'),) * 2
        bud = M.budget(th)
        spr = med([0.86 * (rows[i]['v'].get('sp', 0)) / bud[i] for i in idx])
        out[s] = (m[s], len(idx))
        print(f'| {s} | {m[s]:.3f} | {lo:.3f}-{hi:.3f} | {len(idx)} | {np.mean([nstats(rows[i]["v"]) for i in idx]):.1f} | {spr:.2f} |')
    print('armor slot mods: ' + ', '.join(f'{s} {m[s]:.3f}' for s in ARMOR if s in m))
    return out, M, th, mask


def show_ratio(out, a, b, label):
    if a in out and b in out:
        print(f'   {label}: {out[a][0]:.3f} / {out[b][0]:.3f} = {out[a][0]/out[b][0]:.2f}')


if __name__ == '__main__':
    E = era60()
    lvl = lambda r: (r['q'] == 3 and 45 <= r['ilvl'] <= 70) or (r['q'] == 4 and 55 <= r['ilvl'] <= 80)
    EAW = [r for r in clean(E, ARMOR + ['two-hand', 'one-hand', 'main-hand', 'off-hand weapon', 'held off-hand', 'shield', 'ranged', 'wand'])
           if lvl(r) and r['arch'] != 'stamina']
    print('### 1. Weapons')
    for p in (1.5, 1.7):
        out, *_ = weapon_mods(EAW, 'Era level 60 armor + weapons', p)
        show_ratio(out, 'one-hand | caster', 'one-hand | physical', 'caster one-hander / physical one-hander')
        show_ratio(out, 'two-hand | caster', 'two-hand | physical', 'caster two-hander (staff) / physical two-hander')
        show_ratio(out, 'two-hand | physical', 'one-hand | physical', 'physical two-hander / physical one-hander')
        show_ratio(out, 'two-hand | caster', 'one-hand | caster', 'staff / caster one-hander')
        if 'one-hand | caster' in out and 'held off-hand | caster' in out and 'two-hand | caster' in out:
            mh, oh, st = out['one-hand | caster'][0], out['held off-hand | caster'][0], out['two-hand | caster'][0]
            print(f'   caster main hand + held off-hand = {mh + oh:.3f} chest budgets; staff {st:.3f}')
        show_ratio(out, 'wand | caster', 'ranged | physical', 'wand / bow, gun, crossbow')
    # spell power share of caster weapons vs caster armor
    cw = [r for r in EAW if r['arch'] == 'caster' and r['tier'] in ('one-hand', 'two-hand', 'held off-hand', 'wand') and r['q'] == 4]
    ca = [r for r in EAW if r['arch'] == 'caster' and r['tier'] in TIERS and r['q'] == 4 and 'sp' in r['v']]
    sh = lambda rs: med([0.86 * r['v'].get('sp', 0) / linear(r['v'], WM) for r in rs if 'sp' in r['v']])
    print(f'   spell power share of the linear total: Era caster epic weapons {sh(cw):.2f} (n={sum("sp" in r["v"] for r in cw)}), caster epic armor with sp {sh(ca):.2f} (n={len(ca)})')
    for r in sorted(cw, key=lambda r: (r['tier'], r['ilvl'])):
        print(f'      {r["tier"]:14} {r["name"]:32} ilvl {r["ilvl"]} {r["v"]}')

    S = sod60()
    SAW = [r for r in clean(S) if r['phase'] <= 6 and r['arch'] != 'stamina' and (r['tier'] in TIERS or r['tier'] in WEAP)]
    out, *_ = weapon_mods(SAW, 'SoD phase 4-6 epics, armor + weapons (V per phase)', 1.5, group=lambda r: r['phase'])
    show_ratio(out, 'one-hand | caster', 'one-hand | physical', 'SoD caster one-hander / physical one-hander')
    show_ratio(out, 'two-hand | physical', 'one-hand | physical', 'SoD physical two-hander / physical one-hander')
    show_ratio(out, 'two-hand | caster', 'one-hand | caster', 'SoD staff / caster one-hander')

    F = forever30()
    FAW = [r for r in clean(F) if r['arch'] != 'stamina' and (r['tier'] in TIERS or r['tier'] in WEAP or r['slot'] in ('main-hand', 'off-hand weapon'))]
    WL = dict(WM, heal=0.40, mp5=2.0, school=0.60)
    WM_save = dict(WM)
    WM.update(WL)
    out, *_ = weapon_mods(FAW, 'Forever ilvl 5-30, armor + weapons (heal 0.40, MP5 2.0, school 0.60)', 1.5, boot=30)
    show_ratio(out, 'one-hand | caster', 'one-hand | physical', 'Forever caster one-hander / physical one-hander')
    show_ratio(out, 'two-hand | caster', 'two-hand | physical', 'Forever caster two-hander / physical two-hander')
    WM.clear(); WM.update(WM_save)

    # 2. hit / crit / healing weights at p 1.5 and 1.7 (Era level 60 armor)
    print('\n### 2. Era level 60 armor, every weight free')
    EA = [r for r in EAW if r['tier'] in TIERS]
    for p in (1.5, 1.7):
        M = PModel(EA, p, w=WM, free_w=FREE)
        th, mask = M.robust()
        bs = M.boot(th, mask, 40)
        w = M.unpack(th)[0]
        print(f'   p={p}: ' + ', '.join(f'{k} {w[k]:.2f} ({ci([b[0][k] for b in bs])[0]:.2f}-{ci([b[0][k] for b in bs])[1]:.2f})'
                                       for k in ('sp', 'heal', 'school', 'ap', 'crit', 'spcrit', 'hit', 'sphit', 'mp5') if k in M.free_w))
        print(f'        healing-only point / spell damage+healing point = {w["heal"]/w["sp"]:.2f}')

    # 3. SoD hybrid epics
    print('\n### 3. SoD phase 4-6 epics with physical and spell stats (all, effects included)')
    hy = [r for r in S if r['phase'] <= 6 and r['arch'] == 'hybrid']
    print(f'   {len(hy)} of {sum(1 for r in S if r["phase"] <= 6 and r["v"])} pieces; stats per piece {np.mean([nstats(r["v"]) for r in hy]):.1f}')
    for r in sorted(hy, key=lambda r: (r['phase'], r['slot'])):
        lt = linear(r['v'], WM)
        shares = {k: round(100 * WM[kind(k)] * x / lt) for k, x in r['v'].items()}
        print(f'      P{r["phase"]} {r["slot"]:10} {r["name"]:36} ilvl {r["ilvl"]} classes {r["classes"]} effect {r["effect"]} {shares}')
    # SoD pieces for hybrid classes (paladin 4, shaman 7, druid 1 in SoD's Class enum) carrying Str/Agi
    print('   SoD phase 4-6 pieces restricted to paladin, shaman or druid that carry Str or Agi, by archetype:')
    cls = {4: 'paladin', 7: 'shaman', 1: 'druid'}
    for c, nm in cls.items():
        rs = [r for r in S if r['phase'] <= 6 and c in r['classes'] and ({'str', 'agi', 'ap'} & set(r['v']))]
        print(f'      {nm}: {len(rs)} pieces, ' + ', '.join(f'{k} {n}' for k, n in collections.Counter(r['sub'] for r in rs).most_common()) +
              f'; with spell power: {sum(1 for r in rs if "sp" in r["v"])}')

    # 4. Forever ilvl 5-30 hybrid pieces
    print('\n### 4. Forever ilvl 5-30 armor by archetype (weights: sp 0.86, heal 0.40, MP5 2.0, school 0.60; V per quality)')
    FA = [r for r in clean(F, ARMOR) if r['arch'] != 'stamina']
    allF = [r for r in F if r['tier'] in TIERS and r['v']]
    print(f'   all Forever ilvl 5-30 armor with stats: {len(allF)}; hybrid {sum(r["arch"] == "hybrid" for r in allF)} '
          f'({collections.Counter(r["src"] for r in allF if r["arch"] == "hybrid")}); stats per piece by archetype: ' +
          ', '.join(f'{a} {np.mean([nstats(r["v"]) for r in allF if r["arch"] == a]):.2f} (n={sum(r["arch"] == a for r in allF)})' for a in ARCH))
    for p in (1.5, 1.7):
        M = PModel(FA, p, w=dict(WM, heal=0.40, mp5=2.0, school=0.60))
        th, mask = M.robust()
        r = M.resid(th)
        print(f'   p={p}: rms {100*np.sqrt(np.mean(r[mask]**2)):.1f}%, residual by archetype (median, mean +/- se, n): ' + '; '.join(
            f'{a} {100*np.median(r[ix]):+.1f}% ({100*r[ix].mean():+.1f} +/- {100*r[ix].std()/np.sqrt(len(ix)):.1f}, n={len(ix)})'
            for a in ARCH for ix in [[i for i, x in enumerate(FA) if mask[i] and x['arch'] == a]] if ix))
        for src in ('classic-id', 'forever-new'):
            print(f'        {src}: ' + '; '.join(f'{a} {100*np.median(r[ix]):+.1f}% (n={len(ix)})' for a in ARCH
                                               for ix in [[i for i, x in enumerate(FA) if mask[i] and x['arch'] == a and x['src'] == src]] if ix))
        for lab, sub in (('all', FA), ('Forever-new ids', [x for x in FA if x['src'] == 'forever-new'])):
            Mf = PModel(sub, p, w=dict(WM, heal=0.40, mp5=2.0, school=0.60), factor=lambda x: x['arch'], factor_ref='physical')
            thf, mkf = Mf.robust()
            bsf = Mf.boot(thf, mkf, 30)
            ff = Mf.unpack(thf)[3]
            print(f'        archetype factors ({lab}, vs physical): ' + ', '.join(
                f'{a} {ff[a]:.3f} ({ci([b[3][a] for b in bsf])[0]:.3f}-{ci([b[3][a] for b in bsf])[1]:.3f})' for a in ARCH if a in ff and a != 'physical'))
        hyb = [i for i, x in enumerate(FA) if x['arch'] == 'hybrid']
        if p == 1.5:
            for i in hyb[:40]:
                x = FA[i]
                print(f'        {x["src"]:11} q{x["q"]} ilvl {x["ilvl"]:2} {x["slot"]:8} {x["name"]:32} {x["v"]} resid {100*r[i]:+.0f}%')

    # 5. swaps
    print('\n### 5. Forever stat -> spell power swaps (same Classic id, ilvl and quality as Era; armor and weapons)')
    E30 = {r['id']: r for r in era60(1, 30, quals=(2, 3, 4))}
    pairs = [(f, E30[f['id']]) for f in F if f['src'] == 'classic-id' and f['id'] in E30]
    pairs = [(f, e) for f, e in pairs if not f['other'] and not e['other'] and f['ilvl'] == e['ilvl'] and f['q'] == e['q'] and f['v'] != e['v']]
    res = collections.defaultdict(list)
    for f, e in pairs:
        keys = set(f['v']) | set(e['v'])
        dlt = {k: f['v'].get(k, 0) - e['v'].get(k, 0) for k in keys}
        up = {k: d for k, d in dlt.items() if d > 0}
        dn = {k: -d for k, d in dlt.items() if d < 0}
        if list(up) != ['sp'] or not dn or not set(dn) <= set(PRIM):
            continue
        a = up['sp']
        single = all(e['v'].get(k, 0) == d for k, d in dn.items()) and len(dn) == 1
        src_stat = '+'.join(sorted(dn))
        for p in (1.5, 1.7):
            def solve(aa):
                lo, hi = 0.0, 4.0
                tgt = value(e['v'], WM, p)
                fv = lambda ws: value({k: (x if k != 'sp' else x) for k, x in f['v'].items()}, dict(WM, sp=ws), p) if 'sp' in f['v'] else 0
                ff = dict(f['v']); ff['sp'] = f['v']['sp'] - a + aa
                g = lambda ws: value(ff, dict(WM, sp=ws), p)
                if g(hi) < tgt or g(lo) > tgt:
                    return float('nan')
                for _ in range(60):
                    mid = (lo + hi) / 2
                    lo, hi = (mid, hi) if g(mid) < tgt else (lo, mid)
                return (lo + hi) / 2
            res[(p, 'single' if single else 'partial', src_stat)].append((solve(a), solve(a + 0.5), solve(a - 0.5), sum(dn.values()) / a, e['name'], dn, a))
    for (p, kind_, src), lst in sorted(res.items()):
        w = [x[0] for x in lst if x[0] == x[0]]
        lin_ = [x[3] for x in lst]
        n86 = sum(1 for x in lst if x[1] <= 0.86 <= x[2] or (x[2] != x[2] and x[1] <= 0.86))
        n100 = sum(1 for x in lst if x[1] <= 1.0 <= x[2] or (x[2] != x[2] and x[1] <= 1.0))
        print(f'   p={p} {kind_:7} {src:9} n={len(lst):3}: implied sp weight median {med(w):.2f} (p25-p75 {np.percentile(w,25):.2f}-{np.percentile(w,75):.2f}); '
              f'points removed / sp added median {med(lin_):.2f}; consistent with rounding: 0.86 {n86}, 1.00 {n100}')
    singles = sorted(((x[5], x[6], x[4]) for (p, k, s), lst in res.items() if p == 1.5 and k == 'single' for x in lst), key=lambda t: (list(t[0].values())[0], t[1]))
    big = [(list(d.values())[0], a, n) for d, a, n in singles if list(d.values())[0] >= 4]
    print(f'   single swaps of 4+ points ({len(big)}): ' + '; '.join(f'{n} {d}->{a}' for d, a, n in big))
    print(f'     of these, point for point (rate 1.0): {sum(1 for d, a, n in big if a == d)}; more spell power than points removed: '
          f'{sum(1 for d, a, n in big if a > d)} (rates ' + ', '.join(f'{d/a:.2f}' for d, a, n in big if a > d) + f'); fewer: {sum(1 for d, a, n in big if a < d)}')
    added = [(f, e) for f, e in [(f, E30[f['id']]) for f in F if f['src'] == 'classic-id' and f['id'] in E30]
             if f['ilvl'] == e['ilvl'] and f['q'] == e['q'] and not f['other'] and not e['other'] and f['v'].get('sp', 0) > e['v'].get('sp', 0)
             and all(f['v'].get(k, 0) >= x for k, x in e['v'].items())]
    print(f'   items that gained spell power with nothing removed: {len(added)}, by slot ' + str(collections.Counter(f['slot'] for f, e in added)) +
          '; e.g. ' + '; '.join(f'{f["name"]} {e["v"]} -> {f["v"]}' for f, e in added[:8]))
    print('   examples (p 1.5): ' + '; '.join(f'{x[4]}: {x[5]} -> sp {x[6]}' for (p, k, s), lst in res.items() if p == 1.5 and k == 'partial' for x in lst[:3]))
