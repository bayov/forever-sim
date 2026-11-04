# Run from repo root: python3 tools/audit/tank_gear/shaman.py
import re,json
SRC='tools/database/forever_synthetic_items_shaman_tank.go'
txt=open(SRC).read()
W={'Strength':1,'Agility':1,'Stamina':1,'Intellect':1,'Spirit':1,'SpellPower':.86,'AttackPower':.5,'Defense':1.5,'BlockValue':.4}
L=lambda il:1.125*il+2.0
N=lambda il:0.661*il+2.92
F={'T1':1.093,'T2':1.103,'T3':1.048}
SLOT={'Head':(1,'T1'),'Chest':(1,'T1'),'Legs':(1,'T1'),'Shoulder':(.75,'T2'),'Hands':(.75,'T2'),'Waist':(.75,'T2'),'Feet':(.75,'T2'),'Neck':(.5625,'T3'),'Back':(.5625,'T3'),'Wrist':(.5625,'T3'),'Finger':(.5625,'T3'),'Finger II':(.5625,'T3')}
ROW={'Sta':23.9,'Str':14.3,'Int':15.5,'Spi':10.0,'AP':6.4,'SP':7.2,'Hit':6.6,'Crit':9.3,'Def':4.6,'BV':2.4}
KM={'Stamina':'Sta','Strength':'Str','Intellect':'Int','Spirit':'Spi','AttackPower':'AP','SpellPower':'SP','Defense':'Def','BlockValue':'BV'}
# parse items: each wardenArmor(...) call or struct literal block
items=[]
for m in re.finditer(r'wardenArmor\((\d+), "([^"]+)", "[^"]*", proto\.ItemType_\w+, \w+, (\d+),\s*stats\.Stats\{([^}]*)\}, (\d+), (\d+)\)',txt):
    items.append((m[2],int(m[3]),m[4],int(m[5]),int(m[6])))
def st(s):
    return {k:float(v) for k,v in re.findall(r'stats\.(\w+): (\d+)',s)}
tot={}
def add(d,mult=1):
    for k,v in d.items(): tot[k]=tot.get(k,0)+v*mult
print(f"{'piece':10}{'il':>3}{'lin':>7}{'tgt':>7}{'lin%':>7}{'n17':>7}{'tgt':>7}{'n%':>7}")
def comps(s,h,c):
    d={}
    for k,v in s.items():
        if k in W: d[KM[k]]=d.get(KM[k],0)+v*W[k]
    if h: d['Hit']=h
    if c: d['Crit']=c
    return d
for n,il,s,h,c in items:
    d=comps(st(s),h,c); lin=sum(d.values()); nm=sum(v**1.7 for v in d.values())**(1/1.7)
    sm,t=SLOT[n]; tg=sm*L(il)*F[t]; nt=sm*N(il)
    print(f"{n:10}{il:>3}{lin:7.1f}{tg:7.1f}{100*lin/tg:7.1f}{nm:7.1f}{nt:7.1f}{100*nm/nt:7.1f}  {len(d)} lines")
    add(d)
# weapons
mh=comps({'Strength':13,'Stamina':8},8,8); sh=comps({'Stamina':14,'Intellect':13,'Defense':8,'BlockValue':50},0,0)
for n,d,tg in(('MainHand',mh,.45*L(71)),('Shield',sh,.72*L(71))):
    print(f"{n:10}{71:>3}{sum(d.values()):7.1f}{tg:7.1f}{100*sum(d.values())/tg:7.1f}")
    add(d)
T=sum(tot.values()); print('set total',round(T,1),'target',844.1)
print('shares (set / row / diff, tol 3, Sta 1.5):')
for k,r in ROW.items():
    s=100*tot.get(k,0)/T; print(f"  {k:5}{tot.get(k,0):7.1f}{s:7.1f}{r:7.1f}{s-r:+6.1f}")
nonsta=T-tot['Sta']; print('top non-Sta share:',{k:round(100*v/nonsta,1) for k,v in sorted(tot.items(),key=lambda x:-x[1])[:3]})
print('raw: hit',sum(i[3] for i in items)+8,'rating')
# trinkets
tr=comps({'Defense':20,'BlockValue':40},0,0); lin=sum(tr.values()); nm=sum(v**1.7 for v in tr.values())**(1/1.7)
print('tank trinket lin',round(lin,1),'tgt 45.9 ->',round(100*lin/45.9,1),'% ; n17',round(nm,1),'tgt 36.8 ->',round(100*nm/36.8,1),'%; split Def/BV',round(30/lin*100,1),round(16/lin*100,1))
print('armor trinket avg: 1665 BA*0.1*20/120*? ->',1665*.1*(2*20)/180,'(avg over 180s, 2 uses); target 37')
print('AP trinket 335*.5*40/180 =',335*.5*40/180)
