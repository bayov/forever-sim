"""mkcons.py <base.json> <out.json> key=Value ... : set a consumes field (key=- removes it; key=true/false for booleans)."""
import json, sys
base, out = sys.argv[1:3]
d = json.load(open(base))
c = d['raid']['parties'][0]['players'][0]['consumes']
for a in sys.argv[3:]:
    k, v = a.split('=', 1)
    if v == '-': c.pop(k, None)
    elif v in ('true', 'false'): c[k] = v == 'true'
    else: c[k] = v
json.dump(d, open(out, 'w'), indent=1)
