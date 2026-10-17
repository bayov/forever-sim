# Usage: AUDIT_DB=<db2 dir> python3 cost.py <spellID> ...
# Prints each spell's cast time, power costs, levels and speed.
import csv,sys,os
D=os.environ['AUDIT_DB']
def load(t):
    with open(os.path.join(D,t+'.csv'),newline='',encoding='utf-8') as f: return list(csv.DictReader(f))
names={r['ID']:r['Name_lang'] for r in load('SpellName')}
pw={}
for r in load('SpellPower'): pw.setdefault(r['SpellID'],[]).append(r)
misc={r['SpellID']:r for r in load('SpellMisc')}
ct={r['ID']:r['Base'] for r in load('SpellCastTimes')}
lv={r['SpellID']:r for r in load('SpellLevels')}
for i in sys.argv[1:]:
    m=misc.get(i,{}); l=lv.get(i,{})
    print(i,names.get(i),'cast',ct.get(m.get('CastingTimeIndex')),'cost',[(p['ManaCost'],p['PowerCostPct'],p['PowerType']) for p in pw.get(i,[])],'lvl',l.get('BaseLevel'),l.get('SpellLevel'),l.get('MaxLevel'),'speed',m.get('Speed'))
