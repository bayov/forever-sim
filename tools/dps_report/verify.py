#!/usr/bin/env python3
"""verify.py REQ.json [iters] [seed] [race] -> prints avg dps, stdev, stderr for the request (180 s as in the file)."""
import json, subprocess, sys, tempfile, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import BIN, TREE  # env DASH_BIN / DASH_TREE
req = json.load(open(sys.argv[1]))
iters = int(sys.argv[2]) if len(sys.argv) > 2 else 20000
seed = sys.argv[3] if len(sys.argv) > 3 else "1"
if len(sys.argv) > 4:
    req['raid']['parties'][0]['players'][0]['race'] = 'Race' + sys.argv[4]
req['simOptions'] = {'iterations': iters, 'randomSeed': seed}
with tempfile.TemporaryDirectory() as d:
    i, o = os.path.join(d, 'i.json'), os.path.join(d, 'o.json')
    json.dump(req, open(i, 'w'))
    subprocess.run([os.path.join(BIN, 'wowsimcli'), 'sim', '--infile', i, '--outfile', o], check=True, capture_output=True, cwd=TREE)
    r = json.load(open(o))
    if r.get('error'):
        print('ERROR', r['error']['message'][:500]); sys.exit(1)
    p = r['raidMetrics']['parties'][0]['players'][0]
    print(f"{p['dps']['avg']:.2f} dps  stdev {p['dps']['stdev']:.2f}  se {p['dps']['stdev']/iters**0.5:.2f}  race {p.get('name','')} {req['raid']['parties'][0]['players'][0].get('race')}")
