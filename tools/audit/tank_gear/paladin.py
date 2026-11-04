#!/usr/bin/env python3
"""Audit of the Protection paladin synthetic Phase 1 set. Run from the repo root: python3 tools/audit/tank_gear/paladin.py"""
import re, json, ast
GO='tools/database/forever_synthetic_items_paladin_tank.go'
JS='ui/protection_paladin/gear_sets/level60_phase1_synthetic.gear.json'
W={'Strength':1,'Agility':1,'Stamina':1,'Intellect':1,'Spirit':1,'SpellPower':.86,'AttackPower':.5,
   'Defense':1.5,'Block':5,'BlockValue':.4,'Dodge':12,'Parry':20,'HitRating':1,'CritRating':1}
SLOT={'Head':1,'Chest':1,'Legs':1,'Shoulder':.75,'Hands':.75,'Waist':.75,'Feet':.75,'Neck':.5625,'Back':.5625,'Wrist':.5625,'Finger':.5625}
F={1:1.093,.75:1.103,.5625:1.048}
L=lambda i:1.125*i+2.0
txt=open(GO).read()
items=[]
for m in re.finditer(r'tank\(&proto\.UIItem\{(.*?)\}\),\n',txt,re.S):
    b=m.group(1)
    g=lambda p:(re.search(p,b) or [None,None])[1]
    st={k:float(v) for k,v in re.findall(r'stats\.(\w+): ([\d.]+)',b)}
    for k in('HitRating','CritRating'):
        v=g(k+r':\s+([\d.]+)')
        if v: st[k]=float(v)
    items.append(dict(id=int(g(r'Id: (\d+)')),name=g(r'Name: "([^"]+)"'),type=g(r'ItemType_ItemType(\w+)'),
        armor=g(r'ArmorType: (?:proto\.)?(?:ArmorType_ArmorType|plate)(\w*)') or ('Plate' if 'ArmorType: plate' in b else None),
        ilvl=int(g(r'Ilvl:\s+(\d+)')),st=st))
def vals(st): return {k:W[k]*v for k,v in st.items() if k in W}
tot={}
print('%-9s %4s %6s %6s %6s | %6s %6s %6s  %s'%('piece','ilvl','lin','tgt','%','norm','tgt','%','armor'))
for it in items:
    v=vals(it['st']); lin=sum(v.values()); norm=sum(x**1.7 for x in v.values())**(1/1.7)
    t=it['type']
    if t in SLOT:
        m=SLOT[t]; tl=m*L(it['ilvl'])*F[m]; tn=m*(.661*it['ilvl']+2.92)
    elif t=='Weapon':
        sh='Shield' in it['name']
        tl=(.72 if sh else .45)*L(71); tn=None
    elif t=='Trinket': tl=.56*L(71) if v else None; tn=36.8 if v else None
    else: tl=tn=None
    it['lin']=lin
    f=lambda a,b:'%6.1f'%a if b is None else ''
    print('%-9s %4d %6.1f %6s %6s | %6.1f %6s %6s  %s'%(t+('(S)' if 'Shield' in it['name'] else ''),it['ilvl'],lin,'%.1f'%tl if tl else '-','%.1f'%(100*lin/tl) if tl else '-',norm,'%.1f'%tn if tn else '-','%.1f'%(100*norm/tn) if tn else '-',it['armor']))
    if t not in('Trinket',):
        for k,x in v.items(): tot[k]=tot.get(k,0)+x
T=sum(tot.values()); print('set total %.1f (target 844.1)'%T)
ROW=dict(Stamina=32.4,Strength=16.7,Agility=3.5,Spirit=10,SpellPower=1.1,HitRating=3.2,CritRating=3.1,Defense=19.5,Block=3.2,BlockValue=7.2)
hit=100*tot['HitRating']/T
others=[k for k in ROW if k not in('Stamina','Spirit','HitRating')]
room=100-ROW['Stamina']-ROW['Spirit']-hit; sc=room/sum(ROW[k] for k in others)
print('\nstat share row  hit-adj  diff tol')
for k in ROW:
    adj=ROW[k] if k in('Stamina','Spirit') else (hit if k=='HitRating' else ROW[k]*sc)
    s=100*tot.get(k,0)/T; tol=1.5 if k=='Stamina' else 3
    print('%-11s %5.1f %5.1f %5.1f %+5.1f %s %s'%(k,s,ROW[k],adj,s-adj,tol,'OUT' if abs(s-adj)>tol else ''))
ns=T-tot['Stamina']; print('top non-Sta share: %.1f%%'%(100*max(v for k,v in tot.items() if k!='Stamina')/ns))
raw={k:sum(i['st'].get(k,0) for i in items if i['type']!='Trinket') for k in('Defense','Block','BlockValue','HitRating')}
rawt={k:sum(i['st'].get(k,0) for i in items) for k in('Defense','Block','BlockValue','HitRating')}
print('raw (no trinket)',raw,'with trinket',rawt,' hit%% %.1f +3 Precision = %.1f'%(raw['HitRating']/10,raw['HitRating']/10+3))
print('Defense +7 from 2pc set; defense total',raw['Defense']+7,'; set pieces:',[i['name'] for i in items if 'Head' in i['name'] or 'Chest' in i['name']])
# enchants
enc=json.load(open(JS))['items']
print('\nGear json:',[(e['id'],e.get('enchant')) for e in enc])
