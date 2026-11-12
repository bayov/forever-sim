#!/usr/bin/env python3
"""tank_fairness: summarize bench.sh output. usage: summary.py <dir> [--tsv]
Score at the request name's own weight (max_threat 1/3, balanced 1, max_survival 3; else balanced)."""
import glob,os,re,sys
W={'max_threat':1,'balanced':2,'max_survival':3}
def num(s):
    m=re.match(r'\s*(-?[\d.]+)(?:\s*±\s*([\d.]+))?',s); return (float(m.group(1)),float(m.group(2)) if m.group(2) else 0) if m else (None,None)
d=sys.argv[1]; rows={}
for f in sorted(glob.glob(d+'/*__*.txt')):
    req,boss=os.path.basename(f)[:-4].split('__')
    L=[l for l in open(f) if l.startswith('| ') and not l.startswith('| Build') and not l.startswith('|---')]
    if not L: continue
    c=[x.strip() for x in L[0].strip().strip('|').split('|')]
    k=next((k for k in W if k in req),'balanced'); i=W[k]
    sc,er=num(c[i]); rows.setdefault(req,{})[boss]=dict(score=sc,err=er,tps=num(c[7])[0],dtps=num(c[9])[0],hn=num(c[5])[0])
tsv='--tsv' in sys.argv
print(('request\tboss\tscore\tscore_err\ttps\tdtps\thealing_needed') if tsv else '%-34s %9s %6s %7s %6s %7s'%('request','score','±','tps','dtps','heal'))
for req,bb in rows.items():
    for b,r in bb.items():
        if tsv: print('\t'.join([req,b]+[str(r[x]) for x in('score','err','tps','dtps','hn')]))
    n=len(bb); m={x:sum(r[x] for r in bb.values() if r[x] is not None)/n for x in('score','tps','dtps','hn')}
    e=(sum(r['err']**2 for r in bb.values())**.5)/n
    if tsv: print('\t'.join([req,'scoring_mean','%.1f'%m['score'],'%.1f'%e,'%.1f'%m['tps'],'%.1f'%m['dtps'],'%.1f'%m['hn']]))
    else: print('%-34s %9.1f %6.1f %7.1f %6.1f %7.1f  (%d bosses)'%(req,m['score'],e,m['tps'],m['dtps'],m['hn'],n))
