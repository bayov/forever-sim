# means.py [outdir]: the six-boss own-weight means (Ragnaros left out, as in summary.tsv) of abtest's scoreA and scoreB.
import glob, os, sys

d = sys.argv[1] if len(sys.argv) > 1 else 'orchestration/tank_handoff/shaman/round5/out5'
for f in sorted(glob.glob(os.path.join(d, '*.txt'))):
    a, b = [], []
    for line in open(f):
        w = line.split()
        if len(w) == 7 and w[0] not in ('boss', 'ragnaros') and w[1].lstrip('-').isdigit():
            a.append(int(w[1]))
            b.append(int(w[2]))
    if a:
        print(f'{os.path.basename(f):45s} {sum(a) / len(a):8.1f} -> {sum(b) / len(b):8.1f}')
