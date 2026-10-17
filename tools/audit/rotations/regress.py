#!/usr/bin/env python3
# usage: AUDIT_TREE=<checkout whose req files to use> AUDIT_WORK=<scratch> regress.py OLD_CLI NEW_CLI [--iters 3000] [--seed 424242] [--min 1.0]
# Runs every tools/dps_report/req/*.json (not .meta) on both CLIs with identical settings and seed,
# prints presets whose DPS moved by more than --min. Same req + same seed => any diff is a sim change.
# Caveat: an old CLI can't run req files that use spell ids it doesn't know (e.g. after a client data
# update); such rows show nonsense. Read the table as "what moved", then attribute via git log.
import json,subprocess,os,sys,argparse
from concurrent.futures import ThreadPoolExecutor
ap=argparse.ArgumentParser(); ap.add_argument('old'); ap.add_argument('new')
ap.add_argument('--iters',type=int,default=3000); ap.add_argument('--seed',default='424242'); ap.add_argument('--min',type=float,default=1.0)
ap.add_argument('--jobs',type=int,default=2); a=ap.parse_args()
R=os.path.join(os.environ['AUDIT_TREE'],'tools/dps_report/req/'); W=os.environ['AUDIT_WORK']; os.makedirs(W,exist_ok=True)
names=sorted(f[:-5] for f in os.listdir(R) if f.endswith('.json') and not f.endswith('.meta.json'))
def run(cli,n,tag):
    d=json.load(open(R+n+'.json')); d['simOptions']={'iterations':a.iters,'randomSeed':a.seed}
    f=os.path.join(W,f'reg_{n}.{tag}.{os.getpid()}.json'); json.dump(d,open(f,'w'))
    subprocess.run([cli,'sim','--infile',f,'--outfile',f+'.o'],capture_output=True)
    try: v=json.load(open(f+'.o'))['raidMetrics']['dps']['avg']
    except Exception: v=float('nan')
    for p in (f,f+'.o'):
        if os.path.exists(p): os.remove(p)
    return v
def one(n): return n,run(a.old,n,'o'),run(a.new,n,'n')
with ThreadPoolExecutor(a.jobs) as ex:
    for n,o,v in ex.map(one,names):
        d=v-o
        if not (abs(d)<=a.min): print(f"{n:48s} old {o:8.1f} new {v:8.1f} diff {d:+7.1f}",flush=True)
