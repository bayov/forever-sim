# Shared parser for the synthetic item files (tools/database/forever_synthetic_items*.go).
# Usage: imported by lin2.py, std.py and wepcheck.py; "python3 audit.py FILE.go..." prints items with the OBSOLETE pre-1b cost (ignore the %).
import re,sys,json,glob,os
SLOT={'Head':1,'Chest':1,'Legs':1,'Shoulder':.77,'Hands':.77,'Waist':.77,'Feet':.77,'Neck':.5625,'Back':.5625,'Wrist':.5625,'Finger':.5625,'Ranged':.32}
W={'Strength':1,'Agility':1,'Stamina':1,'Intellect':1,'Spirit':1,'SpellPower':.55,'AttackPower':.5,'RangedAttackPower':.4,'Hit':1,'Crit':1}
def parse(path):
    src=open(path).read()
    items=[]
    pos=[m for m in re.finditer(r'\bId:\s*(99\d{4})',src)]
    for i,m in enumerate(pos):
        end=pos[i+1].start() if i+1<len(pos) else len(src)
        body=src[m.end():end]; id=int(m.group(1))
        pre=src[:m.start()].rstrip().split('\n')
        cm=[l.strip()[2:].strip() for l in pre[-4:] if l.strip().startswith('//')]
        comment=cm[-1] if cm else ''
        it={'id':id,'comment':comment.strip()}
        def g(k,cast=str):
            mm=re.search(r'\b'+k+r':\s*([^,\n}]+)',body); return cast(mm.group(1).strip()) if mm else None
        it['name']=re.search(r'Name:\s*"([^"]*)"',body).group(1)
        it['type']=g('Type').split('ItemType')[-1]
        for k in ['ArmorType','WeaponType','HandType','RangedWeaponType']:
            v=g(k); it[k]=v.split(k)[-1] if v else None
        for k in ['HitRating','CritRating','Ilvl','WeaponDamageMin','WeaponDamageMax','WeaponSpeed']:
            v=g(k); it[k]=float(v) if v else 0
        sm=re.search(r'stats\.Stats\{(.*?)\}',body)
        st={}
        if sm:
            for k,v in re.findall(r'stats\.(\w+):\s*([\d.]+)',sm.group(1)): st[k]=float(v)
        it['stats']=st
        sn=re.search(r'SetName:\s*"([^"]*)"',body); it['set']=sn.group(1) if sn else ('const' if re.search(r'SetName:',body) else None)
        items.append(it)
    return items
def slot(it):
    t=it['type']
    if t=='Weapon':
        if it['HandType']=='TwoHand': return 1.0
        if it['HandType'] in ('OffHand',) and it['WeaponType'] in ('Shield','OffHand'): return .5625
        return .42
    if t=='Ranged': return .32
    return SLOT.get(t)
def cost(it):
    st=dict(it['stats']); ap=st.get('AttackPower',0); rap=st.get('RangedAttackPower',0)
    parts=[]
    shared=min(ap,rap)
    if shared: parts.append(shared*.5); ap-=shared; rap-=shared
    if ap: parts.append(ap*.5)
    if rap: parts.append(rap*.4)
    for k in ['Strength','Agility','Stamina','Intellect','Spirit','SpellPower']:
        if st.get(k): parts.append(st[k]*W[k])
    for k in ['HitRating','CritRating']:
        if it[k]: parts.append(it[k])
    return sum(p**1.5 for p in parts)**(2/3)
def main():
  for f in sys.argv[1:]:
    print('=====',os.path.basename(f))
    for it in parse(f):
        s=slot(it); b=(1.2*it['Ilvl']-36.4)*s; c=cost(it)
        flag='' if abs(c/b-1)<=.02 else ' <<<<'
        st=' '.join(f"{k[:3]}{int(v)}" for k,v in it['stats'].items() if k!='Armor')
        print(f"{it['id']} {it['name'][:42]:42} {it['type'][:8]:8} {it['ArmorType'] or '':7} il{int(it['Ilvl'])} arm{int(it['stats'].get('Armor',0)):4} bud{b:5.1f} cost{c:5.1f} {100*(c/b-1):+5.1f}%{flag} | {st} hit{int(it['HitRating'])} crit{int(it['CritRating'])} {('W '+str(it['WeaponDamageMin'])+'-'+str(it['WeaponDamageMax'])+'@'+str(it['WeaponSpeed'])+' '+str(it['WeaponType'] or it['RangedWeaponType'])+' '+str(it['HandType'])) if it['WeaponSpeed'] else ''} set={'Y' if it['set'] else '-'}")
        print('      //',it['comment'])

if __name__=='__main__': main()
