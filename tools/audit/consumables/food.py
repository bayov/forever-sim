# usage: python3 -I food.py <db2_dir>   Lists every 'Nutritious Food' (Well Fed) item with its buff text, points and item level.
import csv,sys,os,re
D=sys.argv[1]
def load(t):
    with open(os.path.join(D,t+'.csv'),newline='',encoding='utf-8') as f: return list(csv.DictReader(f))
items={r['ID']:r for r in load('ItemSparse')}
ie={r['ID']:r for r in load('ItemEffect')}
sn={r['ID']:r['Name_lang'] for r in load('SpellName')}
sp={r['ID']:r for r in load('Spell')}
se={}
for r in load('SpellEffect'): se.setdefault(r['SpellID'],[]).append(r)
rows=[]
for r in load('ItemXItemEffect'):
    e=ie.get(r['ItemEffectID']); it=items.get(r['ItemID'])
    if not e or not it: continue
    if sn.get(e['SpellID'])!='Nutritious Food': continue
    d=sp.get(e['SpellID'],{}).get('Description_lang','')
    m=re.search(r'gain (.*?) for',d)
    bp=[x['EffectBasePointsF'] for x in se.get(e['SpellID'],[]) if x['EffectIndex']=='1']
    trig=[x['EffectTriggerSpell'] for x in se.get(e['SpellID'],[]) if x['EffectIndex']=='1']
    rows.append((m.group(1) if m else d[:60], bp, it['ID'], it['Display_lang'], it['ItemLevel'], trig))
for x in sorted(rows,key=lambda x:(x[0],x[4])): print(x)
