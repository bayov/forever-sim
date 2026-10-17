# usage: AUDIT_WORK=<scratch> python3 -I src.py <db2_dir> <item name>...   Wowhead Forever sources (drop/vendor/quest) of an item and its recipe; flags raid-only (MC/ZG/BWL/AQ/Naxx/Ony).
# Wowhead pages are cached in $AUDIT_WORK/wh. Wowhead sometimes returns 403 to the container; then say so instead of guessing.
import csv,sys,os,re,json,subprocess,time
D=sys.argv[1]; W=os.path.join(os.environ["AUDIT_WORK"],"wh"); os.makedirs(W,exist_ok=True)
items=list(csv.DictReader(open(os.path.join(D,'ItemSparse.csv'),encoding='utf-8')))
byname={}
for it in items: byname.setdefault(it['Display_lang'],[]).append(it['ID'])
RAID={2717:'MC',1977:'ZG',2677:'BWL',3428:'AQ40',3429:'AQ20',3456:'Naxx',2159:'Ony'}
def fetch(i):
    f=os.path.join(W,f'{i}.html')
    if not os.path.exists(f) or os.path.getsize(f)<1000:
        subprocess.run(['curl','-sSL','-A','Mozilla/5.0','-o',f,f'https://www.wowhead.com/forever/item={i}'])
        time.sleep(0.5)
    return open(f,encoding='utf-8',errors='ignore').read()
def sources(i):
    s=fetch(i); out=[]
    for m in re.finditer(r"id: '([a-z-]+)',(.*?)data: (\[.*?\])\s*,?\s*\}\);",s,re.S):
        lid=m.group(1)
        if lid not in ('dropped-by','sold-by','reward-from-q','contained-in-item','contained-in-object','created-by-spell','created-by'): continue
        try: data=json.loads(m.group(3))
        except Exception: data=[]
        locs=set(); names=[]
        for d in data:
            for l in d.get('location',[]) or []: locs.add(l)
            names.append(d.get('name') or d.get('displayName') or '')
        raid=sorted({RAID[l] for l in locs if l in RAID})
        out.append(f"{lid}[{len(data)}]{'RAID:'+','.join(raid) if raid else ''} {', '.join(names[:4])} locs={sorted(locs)[:6]}")
    return out
for name in sys.argv[2:]:
    for nm in [name]+[p+name for p in ('Recipe: ','Plans: ','Formula: ','Schematic: ')]:
        for i in byname.get(nm,[]):
            print(f"{nm} ({i}):"); [print('   ',x) for x in sources(i)]
