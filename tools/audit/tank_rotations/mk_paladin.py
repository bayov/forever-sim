import json,sys,copy
# usage: mk.py A.json B.json variant
a=json.load(open(sys.argv[1])); v=sys.argv[3]
r=a['raid']['parties'][0]['players'][0]['rotation']; pl=r['priorityList']
def sid(it): return it['action'].get('castSpell',{}).get('spellId',{}).get('spellId')
def idx(s): return [i for i,it in enumerate(pl) if sid(it)==s][0]
def cast(s,cond=None):
    d={'action':{'castSpell':{'spellId':{'spellId':s}}}}
    if cond: d['action']['condition']=cond
    return d
const=lambda x:{'const':{'val':x}}
if v=='autocast_top':
    x=pl.pop(); pl.insert(0,x)
elif v.startswith('bulwark_'):
    pl[idx(1311015)]['action']['condition']['cmp']['rhs']=const(v.split('_')[1])
elif v=='bulwark_prepull':
    pass
elif v.startswith('consmana_'):
    pl[idx(20924)]['action']['condition']['cmp']['rhs']=const(v.split('_')[1])
elif v=='hammer_exec':
    pl.insert(idx(20924),cast(24239,{'cmp':{'lhs':{'currentHealthPercent':{'sourceUnit':{'type':'Target'}}},'op':'OpLt','rhs':const('0.2')}}))
elif v=='hammer_exec_last':
    pl.insert(len(pl)-1,cast(24239,{'cmp':{'lhs':{'currentHealthPercent':{'sourceUnit':{'type':'Target'}}},'op':'OpLt','rhs':const('0.2')}}))
elif v=='prepull_judge':
    r['prepullActions'].append({'action':{'castSpell':{'spellId':{'spellId':20271}}},'doAtValue':const('-0.1s')})
elif v=='no_prepull_hs':
    r['prepullActions']=[p for p in r['prepullActions'] if p['action']['castSpell']['spellId']['spellId']!=1239570]
elif v=='cons_before_hstrike':
    c=pl.pop(idx(20924)); pl.insert(idx(10333),c)
elif v=='hstrike_before_hshield':
    c=pl.pop(idx(10333)); pl.insert(idx(1239570),c)
elif v=='seal_refresh':
    # reseal when seal has < 1.5s left
    pl[idx(20423)]['action']['condition']={'cmp':{'lhs':{'auraRemainingTime':{'auraId':{'spellId':20423}}},'op':'OpLt','rhs':const('1.5s')}}
elif v.startswith('consrank_'):
    pl[idx(20924)]['action']['castSpell']['spellId']['spellId']=int(v.split('_')[1])
else: raise SystemExit('unknown '+v)
json.dump(a,open(sys.argv[2],'w'))
