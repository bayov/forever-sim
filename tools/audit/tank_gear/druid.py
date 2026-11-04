# Audit of the Feral tank (bear) synthetic set. Run from repo root: python3 tools/audit/tank_gear/druid.py
import re,json
G='tools/database/forever_synthetic_items_druid_tank.go'
src=open(G).read()
W={'Strength':1,'Agility':1,'Stamina':1,'Intellect':1,'Spirit':1,'SpellPower':.86,'AttackPower':.5,'Defense':1.5,'Dodge':12,'Parry':20,'Block':5,'BlockValue':.4,'BonusArmor':.1,'MeleeCrit':1}
L=lambda il:1.125*il+2.0
N=lambda il:0.661*il+2.92
F={1:.950,.75:.959,.5625:.911}
SM={'Head':1,'Chest':1,'Legs':1,'Shoulder':.75,'Hands':.75,'Waist':.75,'Feet':.75,'Neck':.5625,'Back':.5625,'Wrist':.5625,'Finger':.5625,'Finger II':.5625}
ILV={'Head':66,'Shoulder':66,'Chest':66,'Hands':66,'Legs':66,'Feet':66}
items=[]
for m in re.finditer(r'(?:feralTankPiece\(\d+|feralTankPiece\((\d+)),\s*"([^"]+)",\s*"[^"]*",\s*proto\.ItemType_ItemType(\w+),\s*(\d+),\s*\n?\s*stats\.Stats\{([^}]*)\},\s*([\d.]+)\)',src):
    pass
pat=re.compile(r'feralTankPiece\((\d+),\s*"([^"]+)",\s*"[^"]*",\s*proto\.ItemType_ItemType(\w+),\s*(\d+),\s*stats\.Stats\{([^}]*)\},\s*([\d.]+)\)',re.S)
for m in pat.finditer(src):
    i,name,typ,il,st,hit=m.groups()
    s={k:float(v) for k,v in re.findall(r'stats\.(\w+):\s*([\d.]+)',st)}
    items.append((int(i),name,int(il),s,float(hit)))
tot={}
def lin(s,hit):
    d={k:v*W[k] for k,v in s.items() if k!='Armor'}
    if hit: d['Hit']=hit
    return d
print('%-10s %4s %7s %7s %6s %7s %7s %6s  arm'%('piece','ilvl','lin','tgt','%','norm','tgt','%'))
shares={}
for i,name,il,s,hit in items:
    d=lin(s,hit)
    t=sum(d.values()); n=sum(v**1.7 for v in d.values())**(1/1.7)
    if name=='Two-Hand':
        tg=.9*L(il); nt=None
    elif name.startswith('Tank Trinket'):
        tg=.56*L(71); nt=36.8
    elif name=='Armor Trinket': continue
    else:
        sm=SM[name]; tier=sm; tg=sm*L(il)*F[sm]; nt=sm*N(il)
    print('%-10s %4d %7.2f %7.2f %+5.1f%% %7.2f %7s %s'%(name,il,t,tg,(t/tg-1)*100,n,'%.2f'%nt if nt else '-', ('%+.1f%%'%((n/nt-1)*100)) if nt else ''))
    if 'Trinket' in name: continue
    for k,v in d.items(): shares[k]=shares.get(k,0)+v
    tot[name]=t
T=sum(shares.values()); print('set total (no trinkets) %.1f vs 729.3'%T)
print('shares %:',{k:round(v/T*100,1) for k,v in sorted(shares.items(),key=lambda x:-x[1])})
row={'Sta':33.4,'Str':16.0,'Agi':19.1,'Int':2.3,'AP':.5,'SP':.3,'Hit':5.9,'Crit':1.7,'Def':20.8}
nm={'Stamina':'Sta','Strength':'Str','Agility':'Agi','Defense':'Def','Hit':'Hit'}
act={nm[k]:v/T*100 for k,v in shares.items()}
h=act['Hit']; others={k:v for k,v in row.items() if k not in('Sta','Hit')}
room=100-row['Sta']-h; sc=room/sum(others.values())
print('hit-adjusted target vs actual (tol Sta 1.5, others 3):')
for k in ['Sta','Str','Agi','Int','AP','SP','Crit','Def','Hit']:
    tgt=row['Sta'] if k=='Sta' else (h if k=='Hit' else row[k]*sc)
    a=act.get(k,0); print('  %-4s act %5.1f tgt %5.1f diff %+5.1f'%(k,a,tgt,a-tgt))
nonSta=T-shares['Stamina']; print('guardrail top non-Sta: %s %.1f%%'%max(((k,v/nonSta*100) for k,v in shares.items() if k!='Stamina'),key=lambda x:x[1]))
# raw: hit rating, whole set
hr=sum(h for *_,h in items); print('hit rating in set',hr,'= %.2f%%'%(hr/10))
# trinket
d=[34.5,12]; print('trinket lin 46.5 vs 45.9 %+.1f%%'%((46.5/45.9-1)*100),'norm %.2f vs 36.8 %+.1f%%'%(sum(x**1.7 for x in d)**(1/1.7),((sum(x**1.7 for x in d)**(1/1.7))/36.8-1)*100))
print('armor on-use avg: 1665*0.1*40/180 = %.1f'%(1665*.1*40/180))
# JSON check
gj=json.load(open('ui/feral_tank_druid/gear_sets/level60_phase1_synthetic.gear.json'))['items']
ids=[x.get('id') for x in gj]; print('json ids',ids)
print('set ids not in json:',[i for i,*_ in items if i not in ids])
