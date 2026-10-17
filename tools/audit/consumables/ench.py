# usage: python3 -I ench.py <db2_dir> <SpellItemEnchantment id>...   Prints the enchant, the spells that apply it and the items that cast them.
import csv,sys,os
D=sys.argv[1]
def load(t): return list(csv.DictReader(open(os.path.join(D,t+'.csv'),encoding='utf-8')))
ench={r['ID']:r for r in load('SpellItemEnchantment')}
sn={r['ID']:r['Name_lang'] for r in load('SpellName')}
items={r['ID']:r for r in load('ItemSparse')}
ie={r['ID']:r for r in load('ItemEffect')}
spell2items={}
for r in load('ItemXItemEffect'):
    e=ie.get(r['ItemEffectID'])
    if e: spell2items.setdefault(e['SpellID'],[]).append(r['ItemID'])
ench2spell={}
for r in load('SpellEffect'):
    if r['Effect'] in ('53','54','92') and r['EffectMiscValue_0'] in ench: ench2spell.setdefault(r['EffectMiscValue_0'],[]).append(r['SpellID'])
for eid in sys.argv[2:]:
    e=ench.get(eid)
    print(f"ENCH {eid} {e['Name_lang'] if e else '??'}  eff={e and (e['Effect_0'],e['EffectPointsMin_0'],e['EffectArg_0'],e['Effect_1'],e['EffectPointsMin_1'],e['EffectArg_1'])}")
    for s in ench2spell.get(eid,[]):
        its=[(i,items[i]['Display_lang']) for i in spell2items.get(s,[]) if i in items]
        print(f"   spell {s} {sn.get(s)} items={its}")
