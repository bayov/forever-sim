"""Write a variant of a paladin tankbench request: mkvar.py <base.json> <out.json> [item=enchant ...] [race=X]
[talents=S] [apl=path]. item=enchant sets that item's enchant (0 removes it)."""
import json
import sys

base, out = sys.argv[1:3]
d = json.load(open(base))
p = d['raid']['parties'][0]['players'][0]
for arg in sys.argv[3:]:
    k, v = arg.split('=', 1)
    if k == 'race':
        p['race'] = v
    elif k == 'talents':
        p['talentsString'] = v
    elif k == 'apl':
        p['rotation'] = json.load(open(v))
    else:
        it = [i for i in p['equipment']['items'] if i['id'] == int(k)]
        assert it, k
        if int(v):
            it[0]['enchant'] = int(v)
        else:
            it[0].pop('enchant', None)
json.dump(d, open(out, 'w'), indent=2)
