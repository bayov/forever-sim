#!/usr/bin/env python3
"""Precision check with every hit step: the final build against it with 1 to 3 Precision points
moved anywhere valid (all single-destination placements, plus random spreads), scored with
-hit-span 6 (every step file) at 3000 iterations, best 4 plus the build at 10000.
usage: precheck.py id template knobs gearbase pins start outdir"""
import sys, os, random, math
src = open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "climb.py")).read()
src = src.split("log = open")[0].replace('ROT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "rotopt")', 'ROT = "%s/rotopt"' % os.path.dirname(os.path.abspath(__file__)))
src = src.replace('"-hit-span", "2"', '"-hit-span", "6"')
exec(src)
b = parse(start); cur = fmt(b); pt, pi = prec
have = b[pt][pi]
cands = set()
for k in range(1, have + 1):
    base = [t[:] for t in b]; base[pt][pi] -= k
    freed = k
    for j, q in enumerate(prq[pt]):  # talents that need Precision maxed go too
        if q == pi and base[pt][j] > 0:
            freed += base[pt][j]; base[pt][j] = 0
    dests = [x for x in flat if not nosim[x[0]][x[1]] and x != prec]
    for (ti, i) in dests:
        c = [t[:] for t in base]; c[ti][i] += freed
        if valid(c): cands.add(fmt(c))
    n0 = len(cands)
    for _ in range(20000):
        c = [t[:] for t in base]
        for _ in range(freed):
            d = random.choice(dests); c[d[0]][d[1]] += 1
        if valid(c): cands.add(fmt(c))
        if len(cands) - n0 >= 25: break
out = open(os.path.join(outdir, f"{bid}_precheck.log"), "w")
r = score(sorted(cands), 3000, "p1")
top = sorted(r.items(), key=lambda kv: -kv[1][0])[:4]
print(f"{len(r)} candidates; top3000 " + "; ".join(f"{s} {d:.1f}" for s, (d, e) in top), file=out, flush=True)
r2 = score([cur] + [s for s, _ in top], 10000, "p1c")
for s, v in sorted(r2.items(), key=lambda kv: -kv[1][0]):
    print(f"  {s} {v}{'  <- current' if s == cur else ''}", file=out, flush=True)
print("done", file=out, flush=True)
