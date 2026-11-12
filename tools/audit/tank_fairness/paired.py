#!/usr/bin/env python3
"""tank_fairness: each variant (<preset>@<tag>) against its base preset at the same seed, six-boss mean,
at the preset's own weight. usage: paired.py <dir> [min_gain]   (dir holds bench.sh output of bases and variants)"""
import glob,os,re,sys
W={'max_threat':1,'balanced':2,'max_survival':3}
def num(s):
    m=re.match(r'\s*(-?[\d.]+)',s); return float(m.group(1)) if m else None
d=sys.argv[1]; mn=float(sys.argv[2]) if len(sys.argv)>2 else -1e9; S={}
for f in glob.glob(d+'/*__*.txt'):
    req,boss=os.path.basename(f)[:-4].rsplit('__',1)
    L=[l for l in open(f) if l.startswith('| ') and not l.startswith('| Build') and not l.startswith('|---')]
    if not L: continue
    c=[x.strip() for x in L[0].strip().strip('|').split('|')]
    S.setdefault(req,{})[boss]=[num(c[1]),num(c[2]),num(c[3]),num(c[7]),num(c[9])]
out=[]
for req,bb in S.items():
    if '@' not in req: continue
    base,tag=req.split('@',1); B=S.get(base)
    if not B: continue
    k=next(k for k in W if k in base); i=W[k]-1
    bosses=[b for b in bb if b in B]
    if len(bosses)<6: continue
    g=sum(bb[b][i]-B[b][i] for b in bosses)/len(bosses)
    alls=[sum(bb[b][j]-B[b][j] for b in bosses)/len(bosses) for j in range(5)]
    if g>=mn: out.append((base,g,tag,alls))
for base,g,tag,a in sorted(out,key=lambda x:(x[0],-x[1])):
    print('%-32s %-38s own %+7.1f | MT %+7.1f Bal %+7.1f MS %+7.1f TPS %+6.1f DTPS %+6.1f'%(base,tag,g,*a))
