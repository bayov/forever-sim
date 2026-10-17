# Usage: AUDIT_DB=$(python3 tools/sources/wago.py dir 1.60.1.70450) python3 sp.py <spellID|name regex> ...
# Prints each spell's duration, proc chance/ICD/charges, cooldown, GCD and every effect (aura, base points, misc, coefficient, class mask).
import csv,sys,re,os
D=os.environ['AUDIT_DB']
def load(t):
    with open(os.path.join(D,t+'.csv'),newline='',encoding='utf-8') as f: return list(csv.DictReader(f))
names={r['ID']:r['Name_lang'] for r in load('SpellName')}
eff={}
for r in load('SpellEffect'): eff.setdefault(r['SpellID'],[]).append(r)
dur={r['ID']:r['Duration'] for r in load('SpellDuration')}
misc={r['SpellID']:r for r in load('SpellMisc')}
aopt={r['SpellID']:r for r in load('SpellAuraOptions')}
cd={r['SpellID']:r for r in load('SpellCooldowns')}
def show(i):
    m=misc.get(i,{})
    a=aopt.get(i,{}); c=cd.get(i,{})
    print(f"== {i} {names.get(i,'?')} dur={dur.get(m.get('DurationIndex',''),'')} proc={a.get('ProcChance','')} ppmid={a.get('SpellProcsPerMinuteID','')} icd={a.get('ProcCategoryRecovery','')} charges={a.get('ProcCharges','')} cd={c.get('RecoveryTime','')}/{c.get('CategoryRecoveryTime','')} gcd={c.get('StartRecoveryTime','')}")
    for e in sorted(eff.get(i,[]),key=lambda e:int(e['EffectIndex'])):
        print(f"  #{e['EffectIndex']} eff={e['Effect']} aura={e['EffectAura']} bp={e['EffectBasePointsF']} die={e.get('EffectDieSides','')} misc={e['EffectMiscValue_0']}/{e['EffectMiscValue_1']} coef={e['EffectBonusCoefficient']} apcoef={e.get('BonusCoefficientFromAP','')} trig={e['EffectTriggerSpell']} amp={e['EffectAmplitude']} per={e['EffectAuraPeriod']} mask={e['EffectSpellClassMask_0']},{e['EffectSpellClassMask_1']},{e['EffectSpellClassMask_2']} rpl={e.get('EffectRealPointsPerLevel','')}")
args=sys.argv[1:]
for a in args:
    if a.isdigit(): show(a)
    else:
        for i,n in names.items():
            if re.fullmatch(a,n,re.I): show(i)
