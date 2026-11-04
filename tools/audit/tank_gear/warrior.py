#!/usr/bin/env python3
"""Audit of the Protection warrior synthetic set. Run from the repo root: python3 tools/audit/tank_gear/warrior.py"""
import re, json
SRC='tools/database/forever_synthetic_items_warrior_tank.go'
txt=open(SRC).read()
W=dict(Sta=1,Str=1,Agi=1,AP=.5,Hit=1,Crit=1,Def=1.5,BV=.4)
L=lambda i:1.125*i+2
F={'T1':.950,'T2':.959,'T3':.911}; MOD={'T1':1,'T2':.75,'T3':.5625}
tier={'Head':'T1','Chest':'T1','Legs':'T1','Shoulder':'T2','Hands':'T2','Feet':'T2','Waist':'T2','Neck':'T3','Back':'T3','Wrist':'T3','Finger':'T3','Finger II':'T3'}
# parse header table
rows=[]
for m in re.finditer(r'^//\t(\w[\w ]*?)\s+(\d+)\s+(T\d|shield|onehand|ranged)\s+([\d.]+)\s+([\d.]+)\s+([\d.-]+)\s+([\d.-]+)\s+(.*)$',txt,re.M):
    rows.append(m.groups())
# parse actual item stats from Go
items=re.findall(r'warriorTankItem\((\d+), "([^"]+)".*?proto\.ItemType_ItemType(\w+), proto\.ArmorType_ArmorType(\w+), (\d+),\s*\n\s*stats\.Stats\{([^}]*)\}, (\d+), (\d+)',txt)
KEY={'Stamina':'Sta','Strength':'Str','Agility':'Agi','AttackPower':'AP','Defense':'Def','BlockValue':'BV','Armor':'Armor'}
def parse(s):
    d={}
    for k,v in re.findall(r'stats\.(\w+): (\d+)',s): d[KEY[k]]=float(v)
    return d
def val(d,hit=0,crit=0):
    d=dict(d); d.pop('Armor',None); d['Hit']=hit; d['Crit']=crit
    lin=sum(W[k]*v for k,v in d.items()); nrm=sum((W[k]*v)**1.7 for k,v in d.items() if v)**(1/1.7)
    return lin,nrm,d
tot=dict(); linset=0
print('piece | ilvl | lin | target | %  | norm | target | % | armor type')
armortype={}
pieces=[]
for id_,name,slot,at,ilvl,s,hit,crit in items:
    d=parse(s); lin,nrm,dd=val(d,float(hit),float(crit))
    sl=name.split('Synthetic ')[1].split(' (')[0]
    t=tier[sl]; ilvl=int(ilvl)
    tl=MOD[t]*L(ilvl)*F[t]; tn=MOD[t]*(0.661*ilvl+2.92)
    print(f'{sl:10} {ilvl} {lin:6.1f} {tl:6.1f} {100*lin/tl:6.1f}% {nrm:5.1f} {tn:5.1f} {100*nrm/tn:6.1f}%  {at} lines={sum(1 for v in dd.values() if v)}')
    pieces.append(dd)
# non-armor pieces
def ex(pat):
    return parse(re.search(pat,txt,re.S).group(1))
rng=dict(Sta=9,Agi=5,AP=1); rng_hit=5
mh=dict(Sta=9,Str=9,Def=6); mh_crit=5
sh=dict(Sta=21,Def=13,BV=46)
for n,d,h,c,tgt in (('Ranged',rng,5,0,.24*L(71)),('MainHand',mh,0,5,.39*L(71)),('Shield',sh,0,0,.72*L(71))):
    lin,nrm,dd=val(d,h,c); print(f'{n:10} 71 {lin:6.1f} {tgt:6.1f} {100*lin/tgt:6.1f}%'); pieces.append(dd)
tr=dict(Def=24,BV=25); lin,nrm,_=val(tr); print(f'Trinket    lin {lin:.1f} (45.9)  norm {nrm:.1f} (36.8) {100*nrm/36.8:.1f}%')
# set totals and shares
S={}
for p in pieces:
    for k,v in p.items(): S[k]=S.get(k,0)+v*W[k]
T=sum(S.values()); print('set total',round(T,1),'(target 761.0)')
row=dict(Sta=31.4,Str=15.5,Agi=6.0,AP=3.4,Hit=9.4,Crit=7.7,Def=20.9,BV=5.8)
print('stat | share | row | diff')
for k in row: print(f'{k:5} {100*S.get(k,0)/T:5.1f} {row[k]:5.1f} {100*S.get(k,0)/T-row[k]:+5.1f}')
print('hit rating',S['Hit'],' Def',S['Def']/1.5,' BV',S['BV']/.4,' Sta',S['Sta'],'  top non-Sta share',max(S[k] for k in S if k!='Sta')/(T-S['Sta']))
armor=sum(MOD[tier[n]]*L(i)*F[tier[n]] for n,i in [('Head',66),('Neck',71),('Shoulder',66),('Back',71),('Chest',66),('Wrist',71),('Hands',66),('Waist',71),('Legs',66),('Feet',66),('Finger',71),('Finger II',71)])
print('armor target total',round(armor,1))
