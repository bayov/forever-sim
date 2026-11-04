import glob,os,re,statistics as st
B='orchestration/tank_handoff/warrior/round5/bench/'
W={'max_threat':('0.333',1),'balanced':('1',2),'max_survival':('3',3)}
def num(s):
    m=re.match(r'\s*(-?[\d.]+)(?:\s*±\s*([\d.]+))?',s)
    return (float(m.group(1)),float(m.group(2)) if m.group(2) else None) if m else (None,None)
rows=[];bosses=[]
for f in sorted(glob.glob(B+'*.txt')):
    req,boss=os.path.basename(f)[:-4].split('__')
    line=[l for l in open(f) if l.startswith('| '+req+' ')][0]
    c=[x.strip() for x in line.strip().strip('|').split('|')]
    k=next((k for k in W if req.endswith(k)),None)
    w,i=W[k] if k else ('1',2)
    sc,err=num(c[i]) if c[i]!='none' else (None,None)
    sd=num(c[4])[0]; hn=num(c[5])[0]; df=num(c[6])[0]
    rows.append(dict(request=req,boss=boss,weight=w+('' if k else '*'),score=sc,score_err=err,safe_dps=sd,tps=num(c[7])[0],dtps=num(c[9])[0],healing_needed=hn,death_floor=df,tmi=num(c[10])[0],death_pct=num(c[11])[0],biggest_hit=num(c[14])[0],dps=num(c[8])[0]))
cols=['request','boss','weight','score','score_err','safe_dps','tps','dtps','healing_needed','death_floor','tmi','death_pct','biggest_hit','dps']
def fmt(v): return '' if v is None else (f'{v:.1f}' if isinstance(v,float) else str(v))
out=['# tank-fixes6 warrior round 5, talents and knobs re-searched on the full raid consumables, 1000 iterations, default seed, date 2026-11-04. tankbench -iters 1000 -encounter tools/encounters/<boss>.json tools/tankbench/req/<request>.json','# weight = the request name\'s preset (Max threat 1/3, Balanced 1, Max survival 3); score = tankbench\'s column at that weight, raid DPS; * = legacy request, Balanced assumed. scoring_mean = mean of the 6 bosses without ragnaros; "none" = no healing saves it.','\t'.join(cols)]
order=['broodlord','chromaggus','golemagg','magmadar','onyxia','vaelastrasz','ragnaros']
for req in sorted({r['request'] for r in rows}):
    rr=sorted([r for r in rows if r['request']==req],key=lambda r:order.index(r['boss']) if r['boss'] in order else 9)
    for r in rr: out.append('\t'.join(fmt(r[c]) for c in cols))
    s=[r for r in rr if r['boss']!='ragnaros']; m=dict(rr[0]); m['boss']='scoring_mean'
    for c in cols[3:]:
        v=[r[c] for r in s if r[c] is not None]
        m[c]=sum(v)/len(v) if len(v)==len(s) or c=='score_err' else None
    if m['score_err'] is not None: m['score_err']=None
    out.append('\t'.join(fmt(m[c]) for c in cols))
open('orchestration/tank_handoff/warrior/round5/summary_rows.tsv','w').write('\n'.join(out)+'\n')
