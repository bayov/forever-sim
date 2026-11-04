"""Run tankbench over the Anniversary encounters and print a compact table and the six's mean.

bench.py <tankbench binary> <repo dir> <iters> <out prefix> <request.json>...
Each request is one build (its file name is the label). Raw output goes to <out prefix>_<enc>.txt.
"""
import re
import subprocess
import sys

binary, repo, iters, prefix = sys.argv[1:5]
reqs = sys.argv[5:]
SCORED = ['magmadar', 'golemagg', 'onyxia', 'broodlord', 'vaelastrasz', 'chromaggus']
ALL = SCORED + ['ragnaros']
COLS = ['Max threat', 'Balanced', 'Max survival', 'SafeDPS', 'Healing needed', 'Death floor', 'TPS', 'DPS', 'DTPS', 'Death']


def num(s):
    m = re.match(r'\s*(-?[\d.]+)', s)
    return float(m.group(1)) if m else float('nan')


rows = {}  # (req, enc) -> {col: value}
for enc in ALL:
    out = subprocess.run([binary, '-iters', iters, '-encounter', f'tools/encounters/{enc}.json'] + reqs,
                         cwd=repo, capture_output=True, text=True)
    text = out.stdout + out.stderr
    open(f'{prefix}_{enc}.txt', 'w').write(text)
    header = None
    for line in text.splitlines():
        if line.startswith('| Build |'):
            header = [c.strip() for c in line.strip('|').split('|')]
        elif header and line.startswith('| ') and not line.startswith('|---'):
            cells = [c.strip() for c in line.strip('|').split('|')]
            label = cells[0]
            rows[(label, enc)] = {c: cells[header.index(c)] for c in COLS if c in header}

labels = sorted({k[0] for k in rows})
for label in labels:
    print(f'== {label}')
    print('| Encounter | ' + ' | '.join(COLS) + ' |')
    for enc in ALL:
        r = rows.get((label, enc))
        if r:
            print(f'| {enc} | ' + ' | '.join(r.get(c, '') for c in COLS) + ' |')
    means = []
    for c in COLS:
        vals = [num(rows[(label, e)][c]) for e in SCORED if (label, e) in rows and c in rows[(label, e)]]
        means.append(f'{sum(vals) / len(vals):.1f}' if vals else '')
    print('| mean of six | ' + ' | '.join(means) + ' |')
