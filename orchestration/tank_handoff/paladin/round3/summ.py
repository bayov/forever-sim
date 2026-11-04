"""Summarize abtest outputs: summ.py <files...> -> file, six-boss mean diff ± paired error, the five without
Broodlord, and Vaelastrasz."""
import math
import sys

SCORED = ['magmadar', 'golemagg', 'onyxia', 'broodlord', 'vaelastrasz', 'chromaggus']
for f in sys.argv[1:]:
    d = {}
    for line in open(f):
        c = line.split()
        if c and c[0] in SCORED:
            d[c[0]] = (float(c[3]), float(c[4]))
    if len(d) < 6:
        print(f'{f}: incomplete')
        continue
    m6 = sum(x for x, _ in d.values()) / 6
    e6 = math.sqrt(sum(s * s for _, s in d.values())) / 6
    five = [d[b] for b in SCORED if b != 'broodlord']
    m5 = sum(x for x, _ in five) / 5
    e5 = math.sqrt(sum(s * s for _, s in five)) / 5
    flag = ' *' if abs(m6) > 2 * e6 else ''
    print(f'{f.split("/")[-1]:42s} six {m6:7.1f} ± {e6:4.1f}{flag:2s} | five (no Broodlord) {m5:7.1f} ± {e5:4.1f} | brood {d["broodlord"][0]:7.1f} | vael {d["vaelastrasz"][0]:7.1f} ± {d["vaelastrasz"][1]:.1f}')
