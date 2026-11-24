import json,sys,glob
# mk_race_req.py <req.json> <Race> <out.json> [apl.json]: the request as another race (e.g. Orc, NightElf),
# with its rotation replaced by a rotopt-written APL when given (rotopt -eval -iters 10 -out <dir> -knobs ...,racial=2,racialAt=30).
req,race,out=sys.argv[1:4]
r=json.load(open(req)); pl=r['raid']['parties'][0]['players'][0]
pl['race']='Race'+race
if len(sys.argv)>4: pl['rotation']=json.load(open(sys.argv[4]))
json.dump(r,open(out,'w'))
