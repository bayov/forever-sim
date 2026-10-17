# usage: AUDIT_WORK=<scratch> python3 -I srcid.py <item id>...   Same as src.py, by item ID. A nonexistent Forever item returns the 'Forever Items' listing page (shows NO SOURCE).
import sys,os,re,json,subprocess,time
W=os.path.join(os.environ["AUDIT_WORK"],"wh"); os.makedirs(W,exist_ok=True)
RAID={2717:'MC',1977:'ZG',2677:'BWL',3428:'AQ40',3429:'AQ20',3456:'Naxx',2159:'Ony'}
for i in sys.argv[1:]:
    f=os.path.join(W,f'{i}.html')
    if not os.path.exists(f) or os.path.getsize(f)<1000:
        subprocess.run(['curl','-sSL','-A','Mozilla/5.0','-o',f,f'https://www.wowhead.com/forever/item={i}']); time.sleep(0.4)
    s=open(f,encoding='utf-8',errors='ignore').read()
    t=re.search(r'<title>(.*?)</title>',s); name=t.group(1).split(' - ')[0] if t else '?'
    out=[]
    for m in re.finditer(r"id: '([a-z-]+)',(.*?)data: (\[.*?\])\s*,?\s*\}\);",s,re.S):
        lid=m.group(1)
        if lid not in ('dropped-by','sold-by','reward-from-q','contained-in-item','contained-in-object'): continue
        try: data=json.loads(m.group(3))
        except Exception: data=[]
        locs=set(); names=[]
        for d in data:
            for l in d.get('location',[]) or []: locs.add(l)
            nm=d.get('name') or d.get('displayName') or ''
            if d.get('react') and lid=='sold-by' and d.get('tag'): nm+=f"<{d['tag']}>"
            names.append(nm)
        raid=sorted({RAID[l] for l in locs if l in RAID})
        out.append(f"{lid}[{len(data)}]{' RAID:'+','.join(raid) if raid else ''} {', '.join(names[:4])} locs={sorted(locs)[:5]}")
    print(f"{i} {name}: "+(' | '.join(out) if out else 'NO SOURCE LISTED'))
