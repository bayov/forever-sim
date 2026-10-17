# Usage: AUDIT_DB=<forever db2 dir> AUDIT_ERA_DB=<era db2 dir> python3 cmp.py <spellID> ...
# Prints SpellEffect rows side by side for Forever and Era (1.15.9.69722).
import csv,sys,os
D=os.environ['AUDIT_DB']; ERA=os.environ.get('AUDIT_ERA_DB','')
def load(p):
    with open(p,newline='',encoding='utf-8') as f: return list(csv.DictReader(f))
ids=set(sys.argv[1:])
names={r['ID']:r['Name_lang'] for r in load(D+'/SpellName.csv')}
for tag,p in (('FOR',D+'/SpellEffect.csv'),('ERA',ERA+'/SpellEffect.csv')):
    for e in load(p):
        if e['SpellID'] in ids:
            print(tag,e['SpellID'],names.get(e['SpellID'],'?')[:28],'#'+e['EffectIndex'],'eff',e['Effect'],'aura',e['EffectAura'],'bp',e.get('EffectBasePoints'),'bpF',e.get('EffectBasePointsF'),'die',e.get('EffectDieSides'),'var',e.get('Variance'),'coef',e['EffectBonusCoefficient'],'ap',e.get('BonusCoefficientFromAP'),'rpl',e['EffectRealPointsPerLevel'],'trig',e['EffectTriggerSpell'],'misc',e['EffectMiscValue_0'],'amp',e['EffectAmplitude'],'per',e['EffectAuraPeriod'])
