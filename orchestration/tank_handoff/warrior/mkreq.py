"""Make a warrior tank request on the synthetic set: mkreq.py <base req> <gear set> <talents> <race> <out> [apl]."""
import json
import sys

base, gearset, talents, race, out = sys.argv[1:6]
apl = sys.argv[6] if len(sys.argv) > 6 else None
d = json.load(open(base))
p = d['raid']['parties'][0]['players'][0]
p['equipment'] = {'items': json.load(open(gearset))['items']}
p['talentsString'] = talents
p['race'] = race
if apl:
    p['rotation'] = json.load(open(apl))
json.dump(d, open(out, 'w'), indent=2)
