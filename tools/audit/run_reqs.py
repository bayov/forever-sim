#!/usr/bin/env python3
"""Sim the DPS report's requests (tools/dps_report/req/<id>.json) with one CLI and write a summary.

usage: run_reqs.py <wowsimcli> <out_dir> [--iterations N] [--jobs J] [--only REGEX] [--duration S]

Runs each request at its own encounter (180 s unless --duration), seed 1, N iterations (default
20000, what the report publishes), J at a time (default: CPU count). Writes <out_dir>/<id>.json
(the CLI's result) and <out_dir>/summary.tsv (id, dps, stdev, published dps). Requests are read
from the CLI's own tree (<cli dir>/tree) when it has one, so old and new builds use their own
requests. Prints only a one-line summary; read summary.tsv or compare_runs.py for the numbers.
"""
import argparse, concurrent.futures as cf, json, os, re, subprocess, sys, tempfile

ap = argparse.ArgumentParser()
ap.add_argument('cli'); ap.add_argument('out')
ap.add_argument('--iterations', type=int, default=20000)
ap.add_argument('--jobs', type=int, default=os.cpu_count() or 4)
ap.add_argument('--only', default='')
ap.add_argument('--duration', type=int)
a = ap.parse_args()

cli = os.path.abspath(a.cli)
tree = os.path.join(os.path.dirname(cli), 'tree')
if not os.path.isdir(tree):
    tree = os.environ.get('AUDIT_REPO') or subprocess.check_output(['git', 'rev-parse', '--show-toplevel'], text=True).strip()
reqdir = os.path.join(tree, 'tools/dps_report/req')
ids = sorted(f[:-5] for f in os.listdir(reqdir) if f.endswith('.json') and not f.endswith('.meta.json'))
ids = [i for i in ids if re.search(a.only, i)]
os.makedirs(a.out, exist_ok=True)


def one(i):
    req = json.load(open(os.path.join(reqdir, i + '.json')))
    req['simOptions'] = {'iterations': a.iterations, 'randomSeed': '1'}
    if a.duration:
        req['encounter']['duration'] = a.duration
    o = os.path.join(a.out, i + '.json')
    with tempfile.NamedTemporaryFile('w', suffix='.json', delete=False) as f:
        json.dump(req, f)
    try:
        subprocess.run([cli, 'sim', '--infile', f.name, '--outfile', o], check=True, capture_output=True, cwd=tree)
    finally:
        os.unlink(f.name)
    r = json.load(open(o))
    if r.get('error'):
        return i, None, None, r['error'].get('message', '')[:200]
    dps = r['raidMetrics']['parties'][0]['players'][0]['dps']  # the player's DPS, as the report publishes
    return i, dps['avg'], dps.get('stdev', 0), None


meta = {}
for i in ids:
    p = os.path.join(reqdir, i + '.meta.json')
    if os.path.exists(p):
        meta[i] = (json.load(open(p)).get('published') or {}).get('dps')
rows, errs = [], []
with cf.ThreadPoolExecutor(a.jobs) as ex:
    for i, avg, sd, err in ex.map(one, ids):
        if err:
            errs.append(f'{i}: {err}')
        else:
            rows.append((i, avg, sd))
with open(os.path.join(a.out, 'summary.tsv'), 'w') as f:
    f.write('id\tdps\tstdev\tpublished\n')
    for i, avg, sd in rows:
        f.write(f'{i}\t{avg:.2f}\t{sd:.2f}\t{meta.get(i) if meta.get(i) is not None else ""}\n')
print(f'run_reqs: {len(rows)} sims ok, {len(errs)} errors -> {a.out}/summary.tsv')
for e in errs:
    print('  error', e)
sys.exit(1 if errs else 0)
