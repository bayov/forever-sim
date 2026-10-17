# usage: python3 -I builddiff.py <old_db2_dir> <new_db2_dir> <forever_enchants.go>
# Diffs a new Forever client build against the last audited one, for what this audit owns:
#  1. every consumable item's (Item.ClassID 0) on-use spells, their trigger spells (2 levels) and the spells their
#     imbue enchants call (SpellEffect 53/54/92/360 -> SpellItemEnchantment EffectArg);
#  2. every enchant in tools/database/forever_enchants.go (spell -> enchant id -> enchant points).
# Armor auras re-encoded 22 -> 674 with the same points are reported as "armor re-encode" and are not findings.
# Item.csv: the new dir's if present, else the old dir's (ClassID doesn't move between builds).
import csv,os,re,sys
O,N,FE=sys.argv[1:4]
def L(d,t): return list(csv.DictReader(open(os.path.join(d,t+'.csv'),newline='',encoding='utf-8')))
def se(d):
    s={}
    for r in L(d,'SpellEffect'):
        s.setdefault(r['SpellID'],[]).append((r['EffectIndex'],r['Effect'],r['EffectAura'],r['EffectBasePointsF'],r['EffectMiscValue_0'],r['EffectTriggerSpell'],r.get('Variance'),r['EffectRealPointsPerLevel'],r['EffectBonusCoefficient']))
    for k in s: s[k].sort()
    return s
o,n=se(O),se(N)
sn={r['ID']:r['Name_lang'] for r in L(N,'SpellName')}
names={r['ID']:r['Display_lang'] for r in L(N,'ItemSparse')}
item={r['ID']:r for r in L(N if os.path.exists(os.path.join(N,'Item.csv')) else O,'Item')}
ie={r['ID']:r for r in L(N,'ItemEffect')}
oe={r['ID']:r for r in L(O,'SpellItemEnchantment')}; ne={r['ID']:r for r in L(N,'SpellItemEnchantment')}
spells=set(); owner={}
for r in L(N,'ItemXItemEffect'):
    if item.get(r['ItemID'],{}).get('ClassID')=='0' and r['ItemEffectID'] in ie:
        s=ie[r['ItemEffectID']]['SpellID']; spells.add(s); owner.setdefault(s,r['ItemID'])
for _ in range(3):
    for s in list(spells):
        for e in n.get(s,[]):
            if e[5]!='0': spells.add(e[5])
            if e[1] in ('53','54','92','360') and e[4] in ne:
                for k in ('EffectArg_0','EffectArg_1','EffectArg_2'):
                    if ne[e[4]].get(k,'0')!='0': spells.add(ne[e[4]][k])
def strip_armor(rows): return [tuple(x[:2])+('ARMOR' if x[2] in ('22','674') and x[4]=='1' else x[2],)+tuple(x[3:]) for x in rows or []]
print('== consumable spells'); c=0
for s in sorted(spells,key=int):
    if o.get(s)==n.get(s): continue
    tag='armor re-encode' if strip_armor(o.get(s))==strip_armor(n.get(s)) else 'CHANGED'
    print(f"{tag} {s} {sn.get(s)} item={owner.get(s)} {names.get(owner.get(s),'')}\n   old {o.get(s)}\n   new {n.get(s)}"); c+=tag=='CHANGED'
print('consumable spell changes:',c)
ids=sorted(set(re.findall(r'forever[A-Za-z]*\((\d+)',open(FE).read())),key=int)
def m(d): return {r['SpellID']:r['EffectMiscValue_0'] for r in L(d,'SpellEffect') if r['Effect'] in ('53','54','92')}
mo,mn=m(O),m(N)
ks=['Effect_0','EffectPointsMin_0','EffectArg_0','Effect_1','EffectPointsMin_1','EffectArg_1','Effect_2','EffectPointsMin_2','EffectArg_2']
print('== forever_enchants.go'); c=0
for i in ids:
    a=oe.get(mo.get(i)); b=ne.get(mn.get(i))
    va=[a.get(k) for k in ks] if a else None; vb=[b.get(k) for k in ks] if b else None
    if (mo.get(i),va)!=(mn.get(i),vb): print('CHANGED',i,sn.get(i),'enchant',mo.get(i),'->',mn.get(i),va,'->',vb); c+=1
print(f'enchants checked {len(ids)}, changed {c}')
