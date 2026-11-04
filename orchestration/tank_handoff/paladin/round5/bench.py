"""Run tankbench over the seven Anniversary encounters; print each build's per-boss scores and the six scoring
bosses' mean at each weight, with its error (sqrt of the summed squares / 6, x1.25 for the shared seed,
tankscore.SharedSeedSD).

bench.py <tankbench binary> <iters> <seed> <request.json>...   (run from the repo root)
"""
import math
import re
import subprocess
import sys

binary, iters, seed = sys.argv[1:4]
reqs = sys.argv[4:]
SCORED = ['magmadar', 'golemagg', 'onyxia', 'broodlord', 'vaelastrasz', 'chromaggus']
ALL = SCORED + ['ragnaros']
W = ['Max threat', 'Balanced', 'Max survival']
COLS = W + ['SafeDPS', 'Healing needed', 'TPS', 'DTPS']


def val(s):
    m = re.match(r'\s*(-?[\d.]+)(?:\s*±\s*([\d.]+))?', s)
    return (float(m.group(1)), float(m.group(2) or 0)) if m else (float('nan'), 0)


rows = {}
for enc in ALL:
    out = subprocess.run([binary, '-iters', iters, '-seed', seed, '-encounter', f'tools/encounters/{enc}.json'] + reqs,
                         capture_output=True, text=True)
    header = None
    for line in (out.stdout + out.stderr).splitlines():
        if line.startswith('| Build |'):
            header = [c.strip() for c in line.strip('|').split('|')]
        elif header and line.startswith('| ') and not line.startswith('|---'):
            cells = [c.strip() for c in line.strip('|').split('|')]
            rows[(cells[0], enc)] = {c: cells[header.index(c)] for c in COLS}
print(f'# tankbench -iters {iters} -seed {seed}; mean = the six scoring bosses (Ragnaros benched, not scored)')
for label in dict.fromkeys(k[0] for k in rows):
    print(f'== {label}')
    print('| Encounter | ' + ' | '.join(COLS) + ' |')
    for enc in ALL:
        print(f'| {enc} | ' + ' | '.join(rows[(label, enc)][c] for c in COLS) + ' |')
    means = []
    for c in COLS:
        v = [val(rows[(label, e)][c]) for e in SCORED]
        m = sum(x for x, _ in v) / 6
        se = math.sqrt(sum(s * s for _, s in v)) / 6 * 1.25
        means.append(f'{m:.1f} ± {se:.1f}' if se else f'{m:.1f}')
    print('| mean of six | ' + ' | '.join(means) + ' |')
