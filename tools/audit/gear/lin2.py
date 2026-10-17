# Standards 1b + 5a checker: per-piece budgets (linear and 1.7-norm), weapon multiples of L(ilvl), stat trinket, hit-adjusted stat shares.
# Usage: AUDIT_TREE=<checkout> [AUDIT_GITREF=origin/master] python3 lin2.py DIR  (targets from the ref when set or when the file is absent)   (DIR from extract.sh; lines without '!' and without 'linear .. vs' / '1.7-norm .. vs' messages pass)
import json,glob,os,sys,re
sys.path.insert(0,os.path.dirname(__file__))
from audit import parse
D=sys.argv[1]
items={}
for f in glob.glob(D+'/*.go'):
    for it in parse(f): items[it['id']]=it
WT={'Strength':1,'Agility':1,'Stamina':1,'Intellect':1,'Spirit':1,'SpellPower':.86,'SpellDamage':.86,
    'FirePower':.7,'FrostPower':.7,'ShadowPower':.7,'ArcanePower':.7,'NaturePower':.7,'HolyPower':.7,
    'HealingPower':.45,'AttackPower':.5,'RangedAttackPower':.4,'MP5':2.4}
MAP={'Strength':'Str','Agility':'Agi','Stamina':'Sta','Intellect':'Int','Spirit':'Spi','AttackPower':'AP','RangedAttackPower':'AP','MP5':'MP5'}
L=lambda il:1.125*il+2.0
N17=lambda il:0.661*il+2.92
TIER={'Head':(1,'T1'),'Chest':(1,'T1'),'Legs':(1,'T1'),'Shoulder':(.75,'T2'),'Hands':(.75,'T2'),'Waist':(.75,'T2'),'Feet':(.75,'T2'),
      'Neck':(.5625,'T3'),'Back':(.5625,'T3'),'Wrist':(.5625,'T3'),'Finger':(.5625,'T3')}
F={'caster':{'T1':1.053,'T2':1.026,'T3':0.995},'phys':{'T1':.95,'T2':.959,'T3':.911},'hunter':{'T1':1.028,'T2':1.056,'T3':.977},'hybrid':{'T1':1.093,'T2':1.103,'T3':1.048}}
def arche(n):
    n=n.lower()
    if 'hunter' in n: return 'hunter'
    if any(x in n for x in ('retribution','enhancement')): return 'hybrid'
    if any(x in n for x in ('warrior','rogue','feral')): return 'phys'
    return 'caster'
def comps(it):
    st=it['stats']; d={}
    for k,v in st.items():
        if k in('Armor',) or k not in WT: continue
        if k=='RangedAttackPower' and st.get('AttackPower'): continue
        kk=MAP.get(k,'SP'); d[kk]=d.get(kk,0)+v*WT[k]
    if it['HitRating']: d['Hit']=d.get('Hit',0)+it['HitRating']
    if it['CritRating']: d['Crit']=d.get('Crit',0)+it['CritRating']
    return d
def n17(d): return sum(v**1.7 for v in d.values())**(1/1.7)
cols=['Sta','Str','Agi','Int','Spi','AP','SP','Hit','Crit','MP5']
T={}
_tp=os.path.join(os.environ.get('AUDIT_TREE','.'),'orchestration','gear_targets.md')
if os.environ.get('AUDIT_GITREF') or not os.path.exists(_tp):
    import subprocess; _tl=subprocess.run(['git','-C',os.environ.get('AUDIT_TREE','.'),'show',os.environ.get('AUDIT_GITREF','origin/master')+':orchestration/gear_targets.md'],capture_output=True,text=True,check=True).stdout.splitlines()
else: _tl=open(_tp).read().splitlines()
for l in _tl:
    if l.startswith('|') and not l.startswith('|---') and 'SoD lists' not in l:
        c=[x.strip() for x in l.strip().strip('|').split('|')]
        T[c[0]]={k:(0 if v in('-','') else float(v)) for k,v in zip(cols,c[2:12])}
def target_for(name):
    n=name.lower()
    for key,pat in [('Warrior Fury','fury'),('Warrior Arms / Fury 2H','arms'),('Rogue daggers','dagger'),('Rogue daggers','mutilate'),('Rogue swords','sword'),('Smite priest','smite'),('Shadow priest','shadow_priest'),('Mage Arcane','arcane'),('Mage (Frostfire / Frost / Fire)','mage'),('Warlock Demonology','demonol'),('Warlock (Destruction / Affliction / Fire)','warlock'),('Balance druid (and Wrath)','balance'),('Feral druid','feral'),('Elemental shaman','elemental'),('Enhancement shaman','enhancement'),('Ret paladin','paladin'),('Hunter Survival','survival'),('Hunter ranged (BM / MM)','ranged')]:
        if pat in n: return key
SL=['head','neck','shoulder','back','chest','wrist','hands','waist','legs','feet','fing1','fing2','trk1','trk2','mh','oh','rng']
for g in sorted(glob.glob(D+'/*.gear.json')):
    base=os.path.basename(g); A=arche(base.split("--",1)[-1]); j=json.load(open(g))['items']; tot={}; msgs=[]; wtot=0
    for i,e in enumerate(j):
        it=items.get(e.get('id')); s=SL[i] if i<len(SL) else str(i)
        if not it:
            if s in('trk1','trk2') and e.get('id'): msgs.append(f"{s}: real item {e['id']} (synthetic trinket expected)")
            if s in('trk1','trk2') and not e.get('id'): msgs.append(f"{s}: empty")
            continue
        d=comps(it); Lt=sum(d.values()); il=it['Ilvl']; t=it['type']
        if t in TIER:
            m,tr=TIER[t]; want=m*L(il)*F[A][tr]; sh=m*N17(il); nv=n17(d)
            if abs(Lt/want-1)>0.01: msgs.append(f"{s} id{it['id']} il{int(il)}: linear {Lt:.1f} vs {want:.1f} ({100*(Lt/want-1):+.1f}%)")
            if abs(nv/sh-1)>0.06: msgs.append(f"{s} id{it['id']}: 1.7-norm {nv:.1f} vs {sh:.1f} ({100*(nv/sh-1):+.1f}%), {len(d)} stats")
            for k,v in d.items(): tot[k]=tot.get(k,0)+v
        elif t=='Trinket':
            if Lt==0: msgs.append(f"[{s} on-use trinket id{it['id']} (check sim code)]"); continue
            if abs(Lt/45.9-1)>0.01: msgs.append(f"{s} trinket id{it['id']}: linear {Lt:.1f} vs 45.9")
            if abs(n17(d)/36.8-1)>0.06: msgs.append(f"{s} trinket: 1.7-norm {n17(d):.1f} vs 36.8")
            if d.get('Sta'): msgs.append(f"{s} trinket carries Stamina")
            msgs.append(f"[{s} trinket id{it['id']} {it['name'][:40]} L{Lt:.1f} {d}]")
        else:
            wtot+=Lt
            for k,v in d.items(): tot[k]=tot.get(k,0)+v
            msgs.append(f"[{s} id{it['id']} {t} {it['HandType'] or it['RangedWeaponType'] or ''} il{int(il)} L {Lt:.1f} = {Lt/L(il):.3f} x L({int(il)})]")
    T_all=sum(tot.values()); key=target_for(base); tg=T.get(key)
    print(f"{base[:62]:62} arch={A} target={key} total L {T_all:.0f}")
    for m in msgs: print('    ',m)
    if tg and T_all:
        a={k:100*tot.get(k,0)/T_all for k in cols}
        held=['SP','Int'] if A=='hybrid' else []
        room=100-a['Sta']-a['Spi']-a['Hit']-sum(tg[k] for k in held); rest=[k for k in cols if k not in('Sta','Spi','Hit')+tuple(held)]; ts=sum(tg[k] for k in rest)
        out=[f"Sta {a['Sta']:.1f}/{tg['Sta']}{'!' if abs(a['Sta']-tg['Sta'])>1.5 else ''}", f"Spi {a['Spi']:.1f}/{tg['Spi']}{'!' if abs(a['Spi']-tg['Spi'])>3 else ''}", f"Hit {a['Hit']:.1f}"]
        if False: pass
        for k in held: out.append(f"{k}(held) {a[k]:.1f}/{tg[k]}{'!' if abs(a[k]-tg[k])>3 else ''}")
        for k in rest:
            exp=tg[k]/ts*room if ts else 0
            if a[k] or exp: out.append(f"{k} {a[k]:.1f}/{exp:.1f}{'!' if abs(a[k]-exp)>3 else ''}")
        nonsta=100-a['Sta']; top=max((a[k],k) for k in cols if k!='Sta')
        out.append(f"top {top[1]} {100*top[0]/nonsta:.0f}%{'!' if top[0]/nonsta>0.55 else ''}")
        print('     shares (actual/target, non-Sta/Spi/Hit rescaled to the room left):',' '.join(out))
