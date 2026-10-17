# usage: python3 -I q.py <db2_dir> item|spell <regex-or-spellid>...   (db2_dir: `python3 tools/sources/wago.py dir <build>`)
# Prints items (with on-use effects) or spells (with SpellEffect rows) from a wago.tools db2 CSV dump.
import csv,sys,re,os
D=sys.argv[1]
def load(t):
    with open(os.path.join(D,t+'.csv'),newline='',encoding='utf-8') as f: return list(csv.DictReader(f))
items=load('ItemSparse'); ie={r['ID']:r for r in load('ItemEffect')}; ix=load('ItemXItemEffect'); sn={r['ID']:r['Name_lang'] for r in load('SpellName')}
se={}
for r in load('SpellEffect'): se.setdefault(r['SpellID'],[]).append(r)
sp={r['ID']:r for r in load('Spell')}
cat={r['SpellID']:r for r in load('SpellCategories')}
catn={r['ID']:r['Name_lang'] for r in load('SpellCategory')}
iebyitem={}
for r in ix:
    if r['ItemEffectID'] in ie: iebyitem.setdefault(r['ItemID'],[]).append(ie[r['ItemEffectID']])
def spell(sid,ind=''):
    print(ind+f"spell {sid} {sn.get(sid)} | {sp.get(sid,{}).get('Description_lang','')[:200]!r} cat={cat.get(sid,{}).get('Category')}:{catn.get(cat.get(sid,{}).get('Category'))}")
    for e in se.get(sid,[]):
        print(ind+f"  eff{e['EffectIndex']} E={e['Effect']} A={e['EffectAura']} bp={e['EffectBasePointsF']} misc={e['EffectMiscValue_0']} trig={e['EffectTriggerSpell']} coef={e['EffectBonusCoefficient']}")
mode=sys.argv[2]
for pat in sys.argv[3:]:
    if mode=='item':
        for it in items:
            if re.search(pat,it['Display_lang'],re.I):
                print(f"ITEM {it['ID']} {it['Display_lang']} ilvl={it['ItemLevel']} q={it['OverallQualityID']} reqskill={it['RequiredSkill']}/{it['RequiredSkillRank']} desc={it['Description_lang']!r}")
                for r in iebyitem.get(it['ID'],[]):
                    print('   cd',r['CoolDownMSec'],r['CategoryCoolDownMSec'],'scat',r['SpellCategoryID']); spell(r['SpellID'],'   ')
    else:
        for sid,n in sn.items():
            if re.search(pat,n,re.I) if not pat.isdigit() else sid==pat: spell(sid)
