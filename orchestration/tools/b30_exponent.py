"""Q1 of the ilvl <= 30 analysis: the combining exponent p, from primary-only items.

Items whose only stats are Str/Agi/Sta/Int/Spi (all weight 1) need no weights. Within a group of
the same quality, slot and item level the right p gives every item the same value
(sum v^p)^(1/p) whether it has one stat or four. We report the RMS spread within groups that mix
stat counts, the spread integer rounding alone would cause (each amount +/-0.5), and the
'multi-stat premium': linear total of an n-stat item / mean total of the single-stat items in its
group (linear pricing predicts 1.00 for every n).
"""
import collections, math, statistics, sys, os
sys.path.insert(0, os.path.dirname(__file__))
from b30_data import forever, era, clean, ARMOR, WEAPON, PRIM

PS = [1.0, 1.25, 1.4, 1.5, 1.6, 1.75, 2.0, 2.5]


def prim_only(rows, slots):
    return [r for r in clean(rows, slots) if set(r['v']) <= set(PRIM)]


def norm(vals, p):
    return sum(a ** p for a in vals) ** (1 / p)


def groups(rows):
    g = collections.defaultdict(list)
    for r in rows:
        g[(r['q'], r['slot'], r['ilvl'])].append(sorted(r['v'].values(), reverse=True))
    return {k: v for k, v in g.items() if len({len(x) for x in v}) > 1}


def spread(gr, p):
    dev, floor = [], []
    for v in gr.values():
        ns = [norm(x, p) for x in v]
        m = statistics.mean(ns)
        dev += [n / m - 1 for n in ns]
        # rounding: dN/dv_i = (v_i/N)^(p-1); var of uniform(-0.5,0.5) = 1/12
        floor += [math.sqrt(sum((a / n) ** (2 * (p - 1)) / 12 for a in x)) / n for x, n in zip(v, ns)]
    return math.sqrt(statistics.mean(d * d for d in dev)), math.sqrt(statistics.mean(f * f for f in floor)), len(dev)


def best_p(gr):
    grid = [1 + i / 20 for i in range(41)]
    return min(grid, key=lambda p: spread(gr, p)[0])


def premium(gr):
    out = collections.defaultdict(list)
    for v in gr.values():
        singles = [sum(x) for x in v if len(x) == 1]
        if not singles:
            continue
        s = statistics.mean(singles)
        for x in v:
            if len(x) > 1:
                out[len(x)].append(sum(x) / s)
    return out


def report(label, rows, slots):
    gr = groups(prim_only(rows, slots))
    n_items = sum(len(v) for v in gr.values())
    print(f'\n== {label}: {len(gr)} mixed groups, {n_items} items')
    for p in PS:
        s, f, n = spread(gr, p)
        print(f'  p={p:4}: rms spread {100*s:5.2f}%   (rounding alone {100*f:4.2f}%)')
    print(f'  best p on a 0.05 grid: {best_p(gr)}')
    pr = premium(gr)
    for k in sorted(pr):
        v = pr[k]
        lin_pred = 1.0
        p15 = statistics.mean([k ** (1 - 1 / 1.5)])  # an evenly split k-stat item under p=1.5
        print(f'  {k}-stat items vs single-stat items in the same group: linear total ratio median {statistics.median(v):.2f}, '
              f'mean {statistics.mean(v):.2f} (n={len(v)}); linear predicts {lin_pred:.2f}, p=1.5 even split {p15:.2f}, p=2 {k ** 0.5:.2f}')
    return gr


if __name__ == '__main__':
    F = forever()
    E30 = era(30)
    report('Forever ilvl 5-30, armor', F, ARMOR)
    report('Forever ilvl 5-30, armor, Classic item ids', [r for r in F if r['src'] == 'classic-id'], ARMOR)
    report('Forever ilvl 5-30, armor, Forever-new item ids', [r for r in F if r['src'] == 'forever-new'], ARMOR)
    report('Forever ilvl 20-30, armor', [r for r in F if r['ilvl'] >= 20], ARMOR)
    report('Forever ilvl 5-30, weapons/shields/off-hands', F, WEAPON)
    report('Era ilvl 1-30, armor (comparison)', E30, ARMOR)
