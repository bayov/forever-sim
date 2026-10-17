#!/usr/bin/env python3
"""Print the newest WoW Forever (1.60.x) client build wago.tools knows, e.g. 1.60.1.70450.

usage: wago_latest.py [--probe-from N --probe-to M]

Reads https://wago.tools/api/builds. That list can lag behind the db2 exports, so --probe-from/--probe-to
also try db2 csv URLs for builds 1.60.1.<N..M> (one small request each) and report the newest that
answers. Prints "unknown" and exits 2 when wago.tools can't be reached; never retries another way.
"""
import json, re, subprocess, sys


def curl(url, head=False):
    cmd = ['curl', '-sS', '-f', '-m', '20', '-A', 'Mozilla/5.0']
    if head:
        cmd += ['-o', '/dev/null', '-w', '%{http_code}', '-r', '0-0']
    return subprocess.run(cmd + [url], capture_output=True, text=True)


r = curl('https://wago.tools/api/builds')
best = None
if r.returncode == 0:
    try:
        data = json.loads(r.stdout)
        vers = re.findall(r'1\.60\.\d+\.\d+', json.dumps(data))
        if vers:
            best = max(vers, key=lambda v: tuple(int(x) for x in v.split('.')))
    except ValueError:
        pass
args = sys.argv[1:]
if '--probe-from' in args:
    lo = int(args[args.index('--probe-from') + 1]); hi = int(args[args.index('--probe-to') + 1])
    start = int(best.split('.')[-1]) + 1 if best else lo
    for n in range(max(lo, start), hi + 1):
        if curl(f'https://wago.tools/db2/SpellName/csv?build=1.60.1.{n}', head=True).stdout.strip() in ('200', '206'):
            best = f'1.60.1.{n}'
if not best:
    print('unknown'); sys.exit(2)
print(best)
