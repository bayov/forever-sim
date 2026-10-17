# Usage: AUDIT_WORK=<scratch> python3 bdiff2.py <oldBuildDir> <newBuildDir> > diff.txt
#   e.g. python3 bdiff2.py db2/1.60.1.70425 db2/1.60.1.70450 (dirs relative to $AUDIT_WORK)
# Diffs SpellEffect/SpellMisc/SpellAuraOptions/SpellCooldowns/SpellLevels/SpellCategories field by field, one line per changed spell.
import csv,sys,collections,os
S=os.environ.get('AUDIT_WORK','.'); A=sys.argv[1]; B=sys.argv[2]
def load(d,t,key):
    m=collections.defaultdict(list)
    for r in csv.DictReader(open(f'{S}/{d}/{t}.csv')): m[r[key]].append(r)
    return m
names={}
for d in (A,B):
    for r in csv.DictReader(open(f'{S}/{d}/SpellName.csv')): names[r['ID']]=r['Name_lang']
changed=collections.defaultdict(list)
for t,key in [('SpellEffect','SpellID'),('SpellMisc','SpellID'),('SpellAuraOptions','SpellID'),('SpellCooldowns','SpellID'),('SpellLevels','SpellID'),('SpellCategories','SpellID')]:
    a=load(A,t,key); b=load(B,t,key)
    for sid in set(a)|set(b):
        ra=sorted([tuple(sorted((k,v) for k,v in r.items() if k!='ID')) for r in a.get(sid,[])])
        rb=sorted([tuple(sorted((k,v) for k,v in r.items() if k!='ID')) for r in b.get(sid,[])])
        if ra!=rb:
            # field-level diffs
            da=[dict(x) for x in ra]; db=[dict(x) for x in rb]
            if t=='SpellEffect':
                ia={x['EffectIndex']:x for x in da}; ib={x['EffectIndex']:x for x in db}
                for i in sorted(set(ia)|set(ib)):
                    x,y=ia.get(i,{}),ib.get(i,{})
                    diffs={k:(x.get(k),y.get(k)) for k in set(x)|set(y) if x.get(k)!=y.get(k)}
                    if diffs: changed[sid].append(f'{t}#{i} '+' '.join(f'{k}:{v[0]}->{v[1]}' for k,v in sorted(diffs.items())))
            else:
                x=da[0] if da else {}; y=db[0] if db else {}
                diffs={k:(x.get(k),y.get(k)) for k in set(x)|set(y) if x.get(k)!=y.get(k)}
                changed[sid].append(f'{t} '+' '.join(f'{k}:{v[0]}->{v[1]}' for k,v in sorted(diffs.items())))
for sid in sorted(changed,key=int):
    print(sid,names.get(sid,'?'),'|',' ; '.join(changed[sid]))
