# usage: python3 mk.py preset name 'python expr mutating pl (priority list) / pp (prepull)'
import json,sys
p,name,code=sys.argv[1:4]
src=f'/home/user/forever-sim/tools/tankbench/req/shaman_warden_{p}.json'
d=json.load(open(src))
pl0=d['raid']['parties'][0]['players'][0]
rot=pl0['rotation']; isstr=isinstance(rot,str)
if isstr: rot=json.loads(rot)
pl=rot['priorityList']; pp=rot.get('prepullActions',[])
exec(code)
rot['priorityList']=pl; rot['prepullActions']=pp
pl0['rotation']=json.dumps(rot) if isstr else rot
out=f'/tmp/audit-work/tank_rotations/shaman/v/{p}__{name}_{p}.json'
json.dump(d,open(out,'w'))
print(out)
