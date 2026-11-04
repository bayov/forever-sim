# usage: mk.py preset out.json  (edit via EDIT code in argv[3], with r = rotation dict, pl = list)
import json,sys
p,out,code=sys.argv[1],sys.argv[2],sys.argv[3]
d=json.load(open(f'/home/user/forever-sim/tools/tankbench/req/druid_bear_{p}.json'))
r=d['raid']['parties'][0]['players'][0]['rotation']
pl=r['priorityList']; pre=r['prepullActions']
def cast(i,cond=None,notes=''):
    a={"action":{"castSpell":{"spellId":{"spellId":i}}}}
    if cond: a["action"]["condition"]=cond
    return a
def rage(op,v): return {"cmp":{"lhs":{"currentRage":{}},"op":op,"rhs":{"const":{"val":str(v)}}}}
def late(i,t):
    c={"cmp":{"lhs":{"currentTime":{}},"op":"OpGt","rhs":{"const":{"val":t}}}}
    old=pl[i]["action"].get("condition")
    pl[i]["action"]["condition"]={"and":{"vals":[c,old]}} if old else c
exec(code)
json.dump(d,open(out,'w'),indent=1)
