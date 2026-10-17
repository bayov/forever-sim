# Usage: AUDIT_DB=<db2 dir> python3 mask.py <spellID> ...
# For each effect with a SpellClassMask, lists the class spells it covers (SpellClassOptions).
import csv,sys,os
D=os.environ['AUDIT_DB']
def load(t):
    with open(os.path.join(D,t+'.csv'),newline='',encoding='utf-8') as f: return list(csv.DictReader(f))
names={r['ID']:r['Name_lang'] for r in load('SpellName')}
sco=load('SpellClassOptions')
cls={r['SpellID']:r for r in sco}
lv={r['SpellID']:r['BaseLevel'] for r in load('SpellLevels')}
eff=[e for e in load('SpellEffect') if e['SpellID'] in sys.argv[1:]]
for e in eff:
    m=[int(e['EffectSpellClassMask_%d'%i])&0xffffffff for i in range(4)]
    if not any(m): continue
    setid=cls.get(e['SpellID'],{}).get('SpellClassSet')
    hits=set()
    for r in sco:
        if r['SpellClassSet']!=setid: continue
        mm=[int(r['SpellClassMask_%d'%i])&0xffffffff for i in range(4)]
        if any(mm[i]&m[i] for i in range(4)): hits.add(names.get(r['SpellID'],'?'))
    print(e['SpellID'],names.get(e['SpellID']),'#'+e['EffectIndex'],'aura',e['EffectAura'],'misc',e['EffectMiscValue_0'],'->',sorted(hits))
