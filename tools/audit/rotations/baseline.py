#!/usr/bin/env python3
# usage: AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> baseline.py CLI [--iters 3000] [--seed 424242] > baseline.tsv
# DPS of every req preset at its own encounter settings, fixed seed. Compare against
# orchestration/audit/ref/rotations/baseline_*.tsv to see what moved since the last audited master.
import json,subprocess,os,argparse
from concurrent.futures import ThreadPoolExecutor
ap=argparse.ArgumentParser(); ap.add_argument('cli'); ap.add_argument('--iters',type=int,default=3000); ap.add_argument('--seed',default='424242'); ap.add_argument('--jobs',type=int,default=4); a=ap.parse_args()
R=os.path.join(os.environ['AUDIT_TREE'],'tools/dps_report/req/'); W=os.environ['AUDIT_WORK']; os.makedirs(W,exist_ok=True)
names=sorted(f[:-5] for f in os.listdir(R) if f.endswith('.json') and not f.endswith('.meta.json'))
def one(n):
    d=json.load(open(R+n+'.json')); d['simOptions']={'iterations':a.iters,'randomSeed':a.seed}
    f=os.path.join(W,f'base_{n}.json'); json.dump(d,open(f,'w'))
    subprocess.run([a.cli,'sim','--infile',f,'--outfile',f+'.o'],capture_output=True)
    v=json.load(open(f+'.o'))['raidMetrics']['dps']['avg']; os.remove(f); os.remove(f+'.o'); return n,v
with ThreadPoolExecutor(a.jobs) as ex:
    for n,v in ex.map(one,names): print(f"{n}\t{v:.2f}",flush=True)
