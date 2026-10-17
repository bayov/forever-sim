# Wowhead Era tooltip fetcher, cached in $AUDIT_WORK/tt. Usage: python3 tt.py ITEM_ID...  (use it to check a real item's ilvl/damage/speed; e.g. db.json had Onyxia Tooth Pendant at ilvl 65, wowhead says 74).
import json,os,re,subprocess,html
D=os.path.join(os.environ.get('AUDIT_WORK','/tmp/audit-gear'),'tt'); os.makedirs(D,exist_ok=True)
def tip(i):
    p=os.path.join(D,f'{i}.json')
    if not os.path.exists(p):
        out=subprocess.run(['curl','-sS',f'https://nether.wowhead.com/classic/tooltip/item/{i}'],capture_output=True,text=True).stdout
        open(p,'w').write(out)
    try: j=json.load(open(p))
    except: return None
    t=j.get('tooltip','')
    txt=re.sub(r'<[^>]+>',' ',t); txt=html.unescape(re.sub(r'\s+',' ',txt))
    r={'name':j.get('name'),'text':txt}
    m=re.search(r'Item Level (\d+)',txt); r['ilvl']=int(m.group(1)) if m else None
    m=re.search(r'(\d+) Armor',txt); r['armor']=int(m.group(1)) if m else None
    m=re.search(r'(\d+) - (\d+)\s+(?:\w+ )?Damage',txt); r['dmg']=(int(m.group(1)),int(m.group(2))) if m else None
    m=re.search(r'Speed ([\d.]+)',txt); r['speed']=float(m.group(1)) if m else None
    return r
if __name__=='__main__':
    import sys
    for a in sys.argv[1:]: print(tip(int(a)))
