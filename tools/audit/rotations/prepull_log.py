#!/usr/bin/env python3
# usage: AUDIT_WORK=<scratch> prepull_log.py CLI REQ.json [--until 3]
# One debug iteration; prints every cast up to t=--until seconds. Use it to check pre-pull timings
# (totems, buffs, pre-casts landing at >=0) and opener order.
import json,subprocess,re,os,argparse
ap=argparse.ArgumentParser(); ap.add_argument('cli'); ap.add_argument('req'); ap.add_argument('--until',type=float,default=3); a=ap.parse_args()
W=os.environ['AUDIT_WORK']; os.makedirs(W,exist_ok=True)
d=json.load(open(a.req)); d['simOptions']={'iterations':1,'randomSeed':'5','debug':True}
f=os.path.join(W,'dbg.json'); json.dump(d,open(f,'w'))
subprocess.run([a.cli,'sim','--infile',f,'--outfile',f+'.o'],check=True,capture_output=True)
for l in json.load(open(f+'.o')).get('logs','').splitlines():
    m=re.search(r'\[(-?[0-9.]+)\]',l)
    if m and 'Casting' in l and float(m.group(1))<a.until: print(l[:160])
