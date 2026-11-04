# summ.py [outdir]: one line per request and variant, B minus A scoring mean +- paired error on seeds 1 and 2,
# marked WIN when both seeds gain by more than twice their error. Run from the repo root.
import glob, os, re, sys
out = sys.argv[1] if len(sys.argv) > 1 else 'orchestration/tank_handoff/shaman/round5/out'
res = {}
for f in sorted(glob.glob(f'{out}/*.txt')):
    m = re.match(r'(.+?)__(.+)_s(\d+)\.txt$', os.path.basename(f))
    if not m:
        continue
    for line in open(f):
        if line.startswith('scoring_mean'):
            d, e = map(float, line.split()[1:3])
            res.setdefault((m.group(1), m.group(2)), {})[m.group(3)] = (d, e)
for (req, v), s in sorted(res.items()):
    a, b = s.get('1'), s.get('2')
    tag = 'WIN' if a and b and a[0] > 2 * a[1] and b[0] > 2 * b[1] else ''
    fmt = lambda x: f'{x[0]:+8.1f} +- {x[1]:5.1f}' if x else '        --       '
    print(f'{req:30s} {v:16s} s1 {fmt(a)}  s2 {fmt(b)}  {tag}')
