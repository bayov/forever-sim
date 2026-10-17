#!/usr/bin/env python3
"""Compare two run_reqs.py summaries (or one against the published numbers).

usage: compare_runs.py <old summary.tsv | --published> <new summary.tsv> [--tol DPS]

Prints only the rows that moved by more than --tol DPS (default 0.5, about 2.5 standard errors at
20k iterations), sorted by the size of the move, then a one-line count. Exit code 1 when anything
moved, so a pass can branch on it without reading the table.
"""
import sys

args = sys.argv[1:]
tol = 0.5
if '--tol' in args:
    k = args.index('--tol'); tol = float(args[k + 1]); del args[k:k + 2]


def load(p, col='dps'):
    rows = {}
    lines = open(p).read().splitlines()
    head = lines[0].split('\t')
    for l in lines[1:]:
        r = dict(zip(head, l.split('\t')))
        if r.get(col):
            rows[r['id']] = float(r[col])
    return rows


new = load(args[1])
old = load(args[1], 'published') if args[0] == '--published' else load(args[0])
moved = sorted(((n - old[i], i, old[i], n) for i, n in new.items() if i in old and abs(n - old[i]) > tol), key=lambda x: -abs(x[0]))
for d, i, o, n in moved:
    print(f'{i:45s} {o:8.1f} -> {n:8.1f}  ({d:+.1f}, {d / o:+.2%})')
only_old, only_new = sorted(set(old) - set(new)), sorted(set(new) - set(old))
print(f'compare_runs: {len(moved)} of {len(set(old) & set(new))} moved more than {tol} DPS'
      + (f'; only in old: {", ".join(only_old)}' if only_old else '') + (f'; only in new: {", ".join(only_new)}' if only_new else ''))
sys.exit(1 if moved else 0)
