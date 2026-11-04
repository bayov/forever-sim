# usage: gen.py base.json out.json "python code" ; r=rotation pl=priorityList p=player c=consumes eq=items d=request
import json,sys
d=json.load(open(sys.argv[1])); p=d['raid']['parties'][0]['players'][0]
r=p['rotation']; pl=r['priorityList']; pre=r.setdefault('prepullActions',[]); c=p['consumes']; eq=p['equipment']['items']
def sid(a): return a['action'].get('castSpell',{}).get('spellId',{}).get('spellId')
def find(spell,uncond=False):
    for a in pl:
        if sid(a)==spell and (not uncond or 'condition' not in a['action']): return a
def setconst(spell,val):
    a=find(spell); s=json.dumps(a['action']['condition'])
    import re
    a['action']['condition']=json.loads(re.sub(r'"val": "[^"]*"','"val": "%s"'%val,s,count=1))
def ench(slot,e): eq[slot]['enchant']=e
exec(sys.argv[3])
json.dump(d,open(sys.argv[2],'w'),indent=1)
