#!/usr/bin/env python3
"""Print the newest WoW Forever (1.60.x) client build wago.tools has, e.g. 1.60.1.70450.

usage: wago_latest.py [--probe N]

Reads https://wago.tools/api/builds. That list can lag behind the db2 exports, so --probe N also
asks for a tiny db2 csv (SpellName) at each of the next N build numbers above the newest it found,
20 at a time, and reports the newest that answers. Prints "unknown" and exits 2 when wago.tools
can't be reached; never retries another way.
"""
import concurrent.futures as cf, json, re, subprocess, sys


def curl(url, head=False):
    cmd = ['curl', '-sS', '-f', '-m', '20', '-A', 'Mozilla/5.0']
    if head:
        cmd += ['-o', '/dev/null', '-w', '%{http_code}', '-r', '0-0']
    return subprocess.run(cmd + [url], capture_output=True, text=True)


def key(v):
    return tuple(int(x) for x in v.split('.'))


r = curl('https://wago.tools/api/builds')
best = None
if r.returncode == 0:
    vers = re.findall(r'1\.60\.\d+\.\d+', r.stdout)
    if vers:
        best = max(vers, key=key)
args = sys.argv[1:]
if best and '--probe' in args:
    n = int(args[args.index('--probe') + 1])
    head, last = best.rsplit('.', 1)
    cands = [f'{head}.{b}' for b in range(int(last) + 1, int(last) + n + 1)]
    with cf.ThreadPoolExecutor(20) as ex:
        hits = [v for v, rr in zip(cands, ex.map(lambda v: curl(f'https://wago.tools/db2/SpellName/csv?build={v}', head=True), cands))
                if rr.stdout.strip() in ('200', '206')]
    if hits:
        best = max(hits, key=key)
if not best:
    print('unknown'); sys.exit(2)
print(best)
