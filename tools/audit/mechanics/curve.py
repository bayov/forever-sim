# Usage: AUDIT_DB=<db2 dir> python3 curve.py <spellID> ...
# Prints a talent spell's per-rank values from TraitDefinitionEffectPoints and CurvePoint.
import csv,sys,os
D=os.environ['AUDIT_DB']
def load(t):
    with open(os.path.join(D,t+'.csv'),newline='',encoding='utf-8') as f: return list(csv.DictReader(f))
names={r['ID']:r['Name_lang'] for r in load('SpellName')}
td={}
for r in load('TraitDefinition'): td.setdefault(r['SpellID'],[]).append(r['ID'])
ep={}
for r in load('TraitDefinitionEffectPoints'): ep.setdefault(r['TraitDefinitionID'],[]).append(r)
cp={}
for r in load('CurvePoint'): cp.setdefault(r['CurveID'],[]).append(r)
for s in sys.argv[1:]:
    for d in td.get(s,[]):
        for e in ep.get(d,[]):
            pts=sorted(cp.get(e['CurveID'],[]),key=lambda r:int(r['OrderIndex']))
            print(s,names.get(s),'eff',e['EffectIndex'],'op',e['OperationType'],[(p['Pos_0'],p['Pos_1']) for p in pts])
