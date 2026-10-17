# Usage: AUDIT_DB=<db2 dir> python3 books.py <spellID> ...
# Shows which item books and teacher spells teach each spell (SpellEffect 36, ItemEffect -> ItemXItemEffect -> ItemSparse) and its SkillLineAbility rows; used for the Ahn'Qiraj book ranks.
import csv,os,sys
D=os.environ['AUDIT_DB']
L=lambda t: list(csv.DictReader(open(os.path.join(D,t+'.csv'),newline='',encoding='utf-8')))
names={r['ID']:r['Name_lang'] for r in L('SpellName')}
teach={}
for e in L('SpellEffect'):
    if e['Effect']=='36': teach.setdefault(e['EffectTriggerSpell'],[]).append(e['SpellID'])
ie={r['ID']:r for r in L('ItemEffect')}
item_of={}
for x in L('ItemXItemEffect'):
    r=ie.get(x['ItemEffectID'])
    if r: item_of.setdefault(r['SpellID'],[]).append((x['ItemID'],r['TriggerType']))
inames={r['ID']:r['Display_lang'] for r in L('ItemSparse')}
sla={}
for r in L('SkillLineAbility'): sla.setdefault(r['Spell'],[]).append(r)
def info(s):
    books=[]
    for t in teach.get(s,[]):
        for it,tt in item_of.get(t,[]): books.append(f"{it} {inames.get(it,'?')} (trig{tt}, via {t})")
    sl=[f"skill{r['SkillLine']} acq{r['AcquireMethod']} supersedes{r['SupercedesSpell']} flags{r['Flags']}" for r in sla.get(s,[])]
    print(s, names.get(s,'?'), '| books:', books or '-', '| teachers(spell36):', teach.get(s,[]), '|', sl)
for a in sys.argv[1:]: info(a)
