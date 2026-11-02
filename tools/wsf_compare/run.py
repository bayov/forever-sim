#!/usr/bin/env python3
"""Sims one of our requests in both sims, matched by convert.py, and prints the DPS side by side.

    run.py <report id | req.json> <out dir> [convert.py args...]

The binaries come from $WSF_WORK/bin (setup.sh): ours_cli (ours_aq_cli for a hunter, unless OURS_AQ=0) and
theirs_cmpsim. Env ALIGN and NOSPELLS go through to convert.py. OURS and THEIRS name other binaries (a flipped rule). Results: <out>/res_ours.json and res_theirs.json.
"""
import json, os, subprocess, sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, '..', '..'))
BIN = os.path.join(os.environ.get('WSF_WORK', os.path.expanduser('~/wsf_compare')), 'bin')

def request_path(name):
    return name if name.endswith('.json') else os.path.join(ROOT, 'tools/dps_report/req', name + '.json')

def main():
    name, out, rest = sys.argv[1], sys.argv[2], sys.argv[3:]
    src = request_path(name)
    conv = subprocess.run([sys.executable, os.path.join(HERE, 'convert.py'), src, out] + rest, capture_output=True, text=True)
    if conv.returncode:
        sys.exit(conv.stdout + conv.stderr)
    klass = json.load(open(src))['raid']['parties'][0]['players'][0]['class']
    ours_bin = 'ours_aq_cli' if klass == 'ClassHunter' and os.environ.get('OURS_AQ', '1') == '1' else 'ours_cli'
    ours = subprocess.run([os.environ.get('OURS') or os.path.join(BIN, ours_bin), 'sim', '--infile', f'{out}/req_ours.json', '--outfile', f'{out}/res_ours.json'],
                          capture_output=True, text=True)
    theirs = subprocess.run([os.path.join(BIN, os.environ.get('THEIRS', 'theirs_cmpsim')), f'{out}/items.json', f'{out}/req_theirs.json', f'{out}/res_theirs.json'],
                            capture_output=True, text=True)
    for who, r in (('ours', ours), ('theirs', theirs)):
        if r.returncode:
            print(f'{who} failed:', (r.stdout + r.stderr).strip()[-600:], file=sys.stderr)
    a, t = dps(f'{out}/res_ours.json'), dps(f'{out}/res_theirs.json')
    label = os.path.basename(name).removesuffix('.json')
    print(f'{label:36s} ours {a[0]:7.1f} (pet {a[1]:6.1f})  theirs {t[0]:7.1f} (pet {t[1]:6.1f})  '
          f'theirs - ours {t[0] - a[0]:+6.1f} (±{(a[2] ** 2 + t[2] ** 2) ** 0.5 * 2:.1f})')

def dps(path):
    """Player DPS, pet DPS and the standard error of the mean."""
    try:
        r = json.load(open(path))
        pl = r['raidMetrics']['parties'][0]['players'][0]
        n = r.get('iterationsDone') or 1
        return pl['dps']['avg'], sum(x['dps']['avg'] for x in pl.get('pets', [])), r['raidMetrics']['dps']['stdev'] / n ** 0.5
    except (OSError, KeyError, ValueError):
        return (float('nan'),) * 3

if __name__ == '__main__':
    main()
