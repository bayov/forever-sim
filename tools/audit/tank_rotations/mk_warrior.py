import json,sys,copy
# usage: mk.py A.json out.json op args...
A,out,op=sys.argv[1:4]; a=sys.argv[4:]
r=json.load(open(A)); rot=r['raid']['parties'][0]['players'][0]['rotation']; pl=rot['priorityList']
if op=='del': del pl[int(a[0])]
elif op=='move': x=pl.pop(int(a[0])); pl.insert(int(a[1]),x)
elif op=='setcond': pl[int(a[0])]['action']['condition']=json.loads(a[1])
elif op=='nocond': pl[int(a[0])]['action'].pop('condition',None)
elif op=='insert': pl.insert(int(a[0]),json.loads(a[1]))
elif op=='prepull': rot['prepullActions'].insert(0,json.loads(a[0]))
elif op=='py': exec(a[0])
json.dump(r,open(out,'w'),indent=1)
