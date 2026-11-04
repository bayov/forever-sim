"""mkvar.py <req> <out> [slot=item:enchant ...] [apl=<file>] [race=<Race>] [talents=<str>]: a variant of a tankbench
request. slot is the equipment index (3 back, 5 wrist, 13 trinket 2, 15 off hand); item or enchant may be empty to keep."""
import json
import sys

src, out = sys.argv[1:3]
d = json.load(open(src))
p = d['raid']['parties'][0]['players'][0]
for a in sys.argv[3:]:
    k, v = a.split('=', 1)
    if k == 'apl':
        p['rotation'] = json.load(open(v))
    elif k == 'race':
        p['race'] = 'Race' + v
    elif k == 'talents':
        p['talentsString'] = v
    else:
        item, ench = v.split(':')
        it = p['equipment']['items'][int(k)]
        if item:
            it['id'] = int(item)
        if ench:
            it['enchant'] = int(ench)
json.dump(d, open(out, 'w'), indent=2)
