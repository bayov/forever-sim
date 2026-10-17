#!/usr/bin/env python3
# usage: AUDIT_WORK=<scratch> ab.py CLI REQ.json APL_A.json APL_B.json [--s0 1000000] [--durs 120,180,300] [--blocks 8] [--iters 20000] [--jobs 4]
# Paired A/B of two rotations on one preset: same seeds for A and B, iterations split into blocks,
# block k uses seed S0 + k*3e6 (seeds iterate as seed+i inside a run, so blocks must be far apart).
# Prints per duration: mean A, B-A and SE (stdev of block diffs / sqrt(blocks)).
# APL_A may be '-' to use the rotation embedded in REQ. Keeps the req's encounter.durationVariation.
# ALWAYS confirm a candidate on a second seed set (--s0 200000000) and pool: the block SE
# understates the spread between seed sets by about 40% (pass 21/22 finding).
import json,subprocess,sys,statistics as st,os,argparse,uuid
from concurrent.futures import ThreadPoolExecutor
ap=argparse.ArgumentParser(); ap.add_argument('cli'); ap.add_argument('req'); ap.add_argument('a'); ap.add_argument('b')
ap.add_argument('--s0',type=int,default=1000000); ap.add_argument('--durs',default='120,180,300'); ap.add_argument('--blocks',type=int,default=8)
ap.add_argument('--iters',type=int,default=20000); ap.add_argument('--jobs',type=int,default=4); a=ap.parse_args()
W=os.environ['AUDIT_WORK']; os.makedirs(W,exist_ok=True)
def run(apl,dur,seed,it):
    d=json.load(open(a.req)); p=d['raid']['parties'][0]['players'][0]
    if apl!='-': p['rotation']=json.load(open(apl))
    d['encounter']['duration']=dur; d['simOptions']={'iterations':it,'randomSeed':str(seed)}
    f=os.path.join(W,f'ab_{uuid.uuid4().hex}.json'); json.dump(d,open(f,'w'))
    subprocess.run([a.cli,'sim','--infile',f,'--outfile',f+'.o'],check=True,capture_output=True)
    v=json.load(open(f+'.o'))['raidMetrics']['dps']['avg']; os.remove(f); os.remove(f+'.o'); return v
per=a.iters//a.blocks
for dur in [int(x) for x in a.durs.split(',')]:
    seeds=[a.s0+k*3000000 for k in range(a.blocks)]
    jobs=[(x,dur,s,per) for s in seeds for x in (a.a,a.b)]
    with ThreadPoolExecutor(a.jobs) as ex: r=list(ex.map(lambda j:run(*j),jobs))
    d=[r[2*k+1]-r[2*k] for k in range(a.blocks)]
    print(f"{os.path.basename(a.req)} s0={a.s0} {dur}s A={st.mean(r[0::2]):.2f} B-A={st.mean(d):+.2f} ±{st.stdev(d)/a.blocks**0.5:.2f}",flush=True)
