#!/usr/bin/env python3
# Usage: python3 summarize.py <AUDIT_WORK dir> [--write ref.csv] [--compare ref.csv]
# Reads m2_<preset>_<tag>.out.json, prints each sensitivity (base -> rescaled, %, place before -> after
# among all bases), optionally writes the reduced baseline table, and compares against a reference
# table: flags bases that moved > 0.4% and sensitivities whose % moved > 0.2 points or whose place shift changed.
import csv, glob, json, os, re, sys
work = sys.argv[1]; args = sys.argv[2:]
rows = []
for f in sorted(glob.glob(os.path.join(work, 'm2_*.out.json'))):
    n, t = re.match(r'.*m2_(.+)_([a-z0-9]+)\.out\.json$', f).groups()
    if t == 'prep': continue
    d = json.load(open(f))
    if d.get('error'): print('ERROR', n, t); continue
    m = d['raidMetrics']['dps']; rows.append((n, t, m['avg'], m.get('stdev', 0)))
def analyse(rows):
    base = {n: v for n, t, v, _ in rows if t == 'base'}
    out = {}
    for n, t, v, _ in rows:
        if t == 'base' or n not in base: continue
        p0 = 1 + sum(1 for k in base if base[k] > base[n])
        p1 = 1 + sum(1 for k in base if k != n and base[k] > v)
        out[(n, t)] = (base[n], v, (v / base[n] - 1) * 100, p0, p1)
    return base, out
base, sens = analyse(rows)
for (n, t), (b, v, pct, p0, p1) in sorted(sens.items()):
    print(f'{n} {t}: {b:.1f} -> {v:.1f} ({pct:+.1f}%) place {p0} -> {p1} of {len(base)}')
if '--write' in args:
    with open(args[args.index('--write') + 1], 'w', newline='') as fh:
        w = csv.writer(fh); w.writerow(['preset', 'tag', 'dps', 'stdev'])
        for r in rows: w.writerow([r[0], r[1], f'{r[2]:.2f}', f'{r[3]:.2f}'])
if '--compare' in args:
    ref = [(r['preset'], r['tag'], float(r['dps']), float(r['stdev'])) for r in csv.DictReader(open(args[args.index('--compare') + 1]))]
    rb, rs = analyse(ref)
    print('\n== vs reference ==')
    for n in sorted(set(base) | set(rb)):
        if n not in base or n not in rb: print('preset added/removed:', n); continue
        if abs(base[n] / rb[n] - 1) > 0.004: print(f'base moved: {n} {rb[n]:.1f} -> {base[n]:.1f} ({(base[n]/rb[n]-1)*100:+.1f}%)')
    for k in sorted(set(sens) | set(rs)):
        if k not in sens or k not in rs: print('sensitivity added/removed:', k); continue
        a, b = rs[k], sens[k]
        if abs(a[2] - b[2]) > 0.2 or (a[4] - a[3]) != (b[4] - b[3]) or a[4] != b[4]:
            print(f'sensitivity changed: {k[0]} {k[1]} {a[2]:+.1f}% place {a[3]}->{a[4]}  now {b[2]:+.1f}% place {b[3]}->{b[4]}')
